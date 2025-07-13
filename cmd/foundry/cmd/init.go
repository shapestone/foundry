package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/shapestone/foundry/internal/layout"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init [project-name]",
	Short: "Initialize a new Go project in the current directory",
	Long: `Initialize a new Go project in the current directory with the specified layout.
	
The init command scaffolds a Go project in the current directory. If no project name
is provided, it uses the current directory name.`,
	Example: `  foundry init
  foundry init myproject
  foundry init --layout=microservice
  foundry init --module=github.com/user/myproject`,
	Args: cobra.MaximumNArgs(1),
	RunE: runInit,
}

func init() {
	// Project configuration flags
	initCmd.Flags().StringP("module", "m", "", "Go module name (default: project name)")
	initCmd.Flags().StringP("layout", "l", "standard", "Project layout to use")
	initCmd.Flags().StringP("author", "a", "", "Project author name")
	initCmd.Flags().StringP("license", "", "MIT", "Project license")
	initCmd.Flags().StringP("description", "d", "", "Project description")
	initCmd.Flags().StringP("github", "g", "", "GitHub username")

	// Layout-specific variables
	initCmd.Flags().StringP("vars", "", "", "Comma-separated layout variables (key=value)")

	// Behavior flags
	initCmd.Flags().BoolP("force", "f", false, "Overwrite existing files")
	initCmd.Flags().Bool("no-git", false, "Skip git initialization")

	// Register with root command
	rootCmd.AddCommand(initCmd)
}

func runInit(cmd *cobra.Command, args []string) error {
	// Determine project name
	var projectName string
	if len(args) > 0 {
		projectName = args[0]
	} else {
		// Use current directory name
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current directory: %w", err)
		}
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
	varsStr, _ := cmd.Flags().GetString("vars")
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

	// Parse custom variables
	customVars := make(map[string]string)
	if varsStr != "" {
		pairs := strings.Split(varsStr, ",")
		for _, pair := range pairs {
			parts := strings.SplitN(pair, "=", 2)
			if len(parts) == 2 {
				customVars[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
			}
		}
	}

	// Check if directory is empty (except for .git)
	if !force {
		empty, err := isDirEmpty(".")
		if err != nil {
			return fmt.Errorf("failed to check directory: %w", err)
		}
		if !empty {
			return fmt.Errorf("current directory is not empty (use --force to overwrite)")
		}
	}

	// Create layout manager
	manager, err := createLayoutManager()
	if err != nil {
		return fmt.Errorf("failed to create layout manager: %w", err)
	}

	// Create project data
	projectData := layout.ProjectData{
		ProjectName:     projectName,
		ModuleName:      moduleName,
		Author:          author,
		License:         license,
		Description:     description,
		GitHubUsername:  githubUsername,
		GoVersion:       "1.21",
		Year:            time.Now().Year(),
		CustomVariables: customVars,
	}

	// Generate project
	fmt.Printf("Initializing project '%s' with layout '%s'...\n", projectName, layoutName)

	ctx := context.Background()
	if err := manager.GenerateProject(ctx, layoutName, ".", projectData); err != nil {
		return fmt.Errorf("failed to generate project: %w", err)
	}

	// Initialize git repository
	if !noGit && !isGitRepo(".") {
		if err := initGitRepo("."); err != nil {
			fmt.Printf("Warning: failed to initialize git repository: %v\n", err)
		} else {
			fmt.Println("✓ Initialized git repository")

			// Create initial commit
			if err := createInitialCommit(); err != nil {
				fmt.Printf("Warning: failed to create initial commit: %v\n", err)
			}
		}
	}

	// Save project configuration
	if err := saveProjectConfig(layoutName); err != nil {
		fmt.Printf("Warning: failed to save project configuration: %v\n", err)
	}

	// Success message
	fmt.Printf("\n✨ Project '%s' initialized successfully!\n\n", projectName)
	fmt.Println("Next steps:")
	fmt.Println("  go mod tidy")
	fmt.Println("  make build")
	fmt.Println("  make run")
	fmt.Println()
	fmt.Println("To add components:")
	fmt.Printf("  foundry add handler users\n")
	fmt.Printf("  foundry add model product\n")
	fmt.Printf("  foundry add middleware auth\n")

	return nil
}

func isDirEmpty(dir string) (bool, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false, err
	}

	for _, entry := range entries {
		// Ignore .git directory
		if entry.Name() == ".git" {
			continue
		}
		// Ignore .DS_Store on macOS
		if entry.Name() == ".DS_Store" {
			continue
		}
		return false, nil
	}

	return true, nil
}

func isGitRepo(dir string) bool {
	gitDir := filepath.Join(dir, ".git")
	info, err := os.Stat(gitDir)
	return err == nil && info.IsDir()
}

func createInitialCommit() error {
	// In a real implementation, this would use go-git or exec git commands
	// For now, just return nil
	return nil
}

func saveProjectConfig(layoutName string) error {
	// Save the layout name in foundry.yaml so we know which layout was used
	// This is already done by the layout system if foundry.yaml.tmpl exists
	return nil
}
