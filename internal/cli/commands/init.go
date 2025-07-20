package commands

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec" // <-- ADD THIS IMPORT
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/shapestone/foundry/internal/layout"
	"github.com/spf13/cobra"
)

// ProjectData holds project initialization data
type ProjectData struct {
	ProjectName     string
	ModuleName      string
	Author          string
	License         string
	Description     string
	GitHubUsername  string
	GoVersion       string
	Year            int
	CustomVariables map[string]string
}

// BuildInitCommand creates the init command using the adapter pattern
func BuildInitCommand(adapter *CLIAdapter) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init [project-name]",
		Short: "Initialize a new Go project IN THE CURRENT DIRECTORY",
		Long: `Initialize a new Go project in the current directory with the specified layout.

IMPORTANT: The init command scaffolds a Go project IN THE CURRENT DIRECTORY.
If no project name is provided, it uses the current directory name.

This follows the same pattern as 'git init' - it initializes the project where you are.
Use 'foundry new' if you want to create a new directory for your project.`,
		Example: `  # Initialize project in current directory (like 'git init')
  mkdir myproject && cd myproject
  foundry init

  # Initialize with specific name in current directory  
  foundry init myproject

  # Initialize with custom layout and module
  foundry init --layout=microservice --module=github.com/user/myproject

  # Compare with 'new' command (creates subdirectory):
  foundry new myproject    # Creates ./myproject/ directory`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runInit(cmd, args, adapter)
		},
	}

	// Project configuration flags
	cmd.Flags().StringP("module", "m", "", "Go module name (default: project name)")
	cmd.Flags().StringP("layout", "l", "standard", "Project layout to use")
	cmd.Flags().StringP("author", "a", "", "Project author name")
	cmd.Flags().StringP("license", "", "MIT", "Project license")
	cmd.Flags().StringP("description", "d", "", "Project description")
	cmd.Flags().StringP("github", "g", "", "GitHub username")
	cmd.Flags().BoolP("force", "f", false, "Overwrite existing files")
	cmd.Flags().Bool("no-git", false, "Skip git initialization")

	return cmd
}

// runInit executes the init command
func runInit(cmd *cobra.Command, args []string, adapter *CLIAdapter) error {
	stdout := adapter.GetStdout()
	stderr := adapter.GetStderr()

	// Get current working directory for clear messaging
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	// Determine project name
	var projectName string
	if len(args) > 0 {
		projectName = args[0]
	} else {
		// Use current directory name
		projectName = filepath.Base(cwd)
	}

	// Validate project name
	if !isValidProjectName(projectName) {
		return fmt.Errorf("invalid project name: %s (use only letters, numbers, hyphens, and underscores)", projectName)
	}

	// Get flags
	moduleName, _ := cmd.Flags().GetString("module")
	layoutName, _ := cmd.Flags().GetString("layout")
	author, _ := cmd.Flags().GetString("author")
	license, _ := cmd.Flags().GetString("license")
	description, _ := cmd.Flags().GetString("description")
	githubUsername, _ := cmd.Flags().GetString("github")
	force, _ := cmd.Flags().GetBool("force")
	noGit, _ := cmd.Flags().GetBool("no-git")

	// Default module name
	if moduleName == "" {
		if githubUsername != "" {
			moduleName = fmt.Sprintf("github.com/%s/%s", githubUsername, projectName)
		} else {
			moduleName = projectName
		}
	}

	// Default description
	if description == "" {
		description = fmt.Sprintf("A Go project created with Foundry using the %s layout", layoutName)
	}

	// Enhanced safety check with clearer messaging
	if !force {
		empty, err := isDirEmpty(".")
		if err != nil {
			return fmt.Errorf("failed to check directory: %w", err)
		}
		if !empty {
			fmt.Fprintf(stderr, "⚠️  Current directory is not empty: %s\n", cwd)
			fmt.Fprintln(stderr, "Found non-hidden files that would be mixed with the new project.")
			fmt.Fprintln(stderr, "")
			fmt.Fprintln(stderr, "Options:")
			fmt.Fprintln(stderr, "  1. Use --force to initialize anyway (may overwrite files)")
			fmt.Fprintln(stderr, "  2. Use 'foundry new myproject' to create a new directory instead")
			fmt.Fprintln(stderr, "  3. Create an empty directory first: mkdir myproject && cd myproject")
			return fmt.Errorf("directory not empty (use --force to overwrite or 'foundry new' to create subdirectory)")
		}
	}

	// Create project data
	projectData := ProjectData{
		ProjectName:     projectName,
		ModuleName:      moduleName,
		Author:          author,
		License:         license,
		Description:     description,
		GitHubUsername:  githubUsername,
		GoVersion:       "1.21",
		Year:            time.Now().Year(),
		CustomVariables: make(map[string]string),
	}

	// Clear initialization message showing current location
	fmt.Fprintf(stdout, "🚀 Initializing project '%s' in current directory...\n", projectName)
	fmt.Fprintf(stdout, "📁 Location: %s\n", cwd)
	fmt.Fprintf(stdout, "🏗️  Layout: %s\n", layoutName)
	fmt.Fprintln(stdout, "")

	if err := generateProject(layoutName, ".", projectData, stdout, stderr); err != nil {
		return fmt.Errorf("failed to generate project: %w", err)
	}

	// Initialize git repository
	if !noGit && !isGitRepo(".") {
		if err := initGitRepo("."); err != nil {
			fmt.Fprintf(stderr, "Warning: failed to initialize git repository: %v\n", err)
		} else {
			fmt.Fprintln(stdout, "✓ Initialized git repository")

			// Create initial commit
			if err := createInitialCommit(); err != nil {
				fmt.Fprintf(stderr, "Warning: failed to create initial commit: %v\n", err)
			}
		}
	}

	// Enhanced success message with clear location info
	fmt.Fprintln(stdout, "")
	fmt.Fprintf(stdout, "✨ Project '%s' initialized successfully!\n", projectName)
	fmt.Fprintf(stdout, "📍 Project location: %s\n", cwd)
	fmt.Fprintln(stdout, "")
	fmt.Fprintln(stdout, "🎯 Next steps:")
	fmt.Fprintln(stdout, "  go mod tidy        # Download dependencies")
	fmt.Fprintln(stdout, "  make build         # Build the project")
	fmt.Fprintln(stdout, "  make run           # Run the project")
	fmt.Fprintln(stdout, "")
	fmt.Fprintln(stdout, "🔧 To add components:")
	fmt.Fprintln(stdout, "  foundry add handler users")
	fmt.Fprintln(stdout, "  foundry add model product")
	fmt.Fprintln(stdout, "  foundry add middleware auth")
	fmt.Fprintln(stdout, "")
	fmt.Fprintln(stdout, "💡 Tip: Use 'foundry new <name>' next time to create a new directory")

	return nil
}

// generateProject creates the project structure using the layout system
func generateProject(layoutName, targetDir string, data ProjectData, stdout, stderr io.Writer) error {
	// Get layout manager config path
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}
	configPath := filepath.Join(homeDir, ".foundry", "layouts.yaml")

	// Create layout manager
	manager, err := layout.NewManager(configPath)
	if err != nil {
		return fmt.Errorf("failed to create layout manager: %w", err)
	}

	// Convert ProjectData to layout.ProjectData
	layoutData := layout.ProjectData{
		ProjectName:     data.ProjectName,
		ModuleName:      data.ModuleName,
		Author:          data.Author,
		License:         data.License,
		Description:     data.Description,
		GitHubUsername:  data.GitHubUsername,
		Year:            data.Year,
		GoVersion:       data.GoVersion,
		CustomVariables: data.CustomVariables,
	}

	// Generate project using layout system
	ctx := context.Background()
	err = manager.GenerateProject(ctx, layoutName, targetDir, layoutData)
	if err != nil {
		return fmt.Errorf("layout generation failed: %w", err)
	}

	fmt.Fprintf(stdout, "✓ Generated project using '%s' layout\n", layoutName)
	return nil
}

// Helper functions

// isValidProjectName validates a project name
func isValidProjectName(name string) bool {
	// Allow letters, numbers, hyphens, and underscores
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9_-]+$`, name)
	return matched && len(name) > 0
}

// isDirEmpty checks if a directory is empty (ignoring .git)
func isDirEmpty(dir string) (bool, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false, err
	}

	for _, entry := range entries {
		// Ignore .git directory and hidden files starting with .
		if !strings.HasPrefix(entry.Name(), ".") {
			return false, nil
		}
	}
	return true, nil
}

// isGitRepo checks if the current directory is a git repository
func isGitRepo(dir string) bool {
	gitDir := filepath.Join(dir, ".git")
	if stat, err := os.Stat(gitDir); err == nil {
		return stat.IsDir()
	}
	return false
}

// initGitRepo initializes a git repository
func initGitRepo(dir string) error {
	cmd := exec.Command("git", "init")
	cmd.Dir = dir
	return cmd.Run()
}

// createInitialCommit creates the initial git commit
func createInitialCommit() error {
	// Add all files
	if err := exec.Command("git", "add", ".").Run(); err != nil {
		return err
	}

	// Create initial commit
	return exec.Command("git", "commit", "-m", "Initial commit").Run()
}
