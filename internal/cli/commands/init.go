package commands

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

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
	cmd.Flags().StringP("vars", "", "", "Comma-separated layout variables (key=value)")
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
		CustomVariables: customVars,
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

	// Save project configuration
	if err := saveProjectConfig(layoutName); err != nil {
		fmt.Fprintf(stderr, "Warning: failed to save project configuration: %v\n", err)
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

// saveProjectConfig saves project configuration to .foundry.yaml
func saveProjectConfig(layout string) error {
	config := fmt.Sprintf("layout: %s\n", layout)
	return os.WriteFile(".foundry.yaml", []byte(config), 0644)
}

// generateProject creates the project structure based on layout
func generateProject(layoutName, targetDir string, data ProjectData, stdout, stderr io.Writer) error {
	// For now, create a basic Go project structure
	// TODO: Implement full layout system with templates

	if err := createBasicProjectStructure(targetDir, data); err != nil {
		return fmt.Errorf("creating project structure: %w", err)
	}

	fmt.Fprintf(stdout, "✓ Created project structure\n")
	return nil
}

// createBasicProjectStructure creates a minimal Go project
func createBasicProjectStructure(targetDir string, data ProjectData) error {
	// Create directories
	dirs := []string{
		"cmd/" + data.ProjectName,
		"internal",
		"pkg",
		"api",
		"configs",
		"scripts",
		"test",
		"docs",
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(filepath.Join(targetDir, dir), 0755); err != nil {
			return err
		}
	}

	// Create go.mod
	goMod := fmt.Sprintf(`module %s

go %s

require (
	github.com/spf13/cobra v1.8.0
)
`, data.ModuleName, data.GoVersion)

	if err := os.WriteFile(filepath.Join(targetDir, "go.mod"), []byte(goMod), 0644); err != nil {
		return err
	}

	// Create main.go
	mainGo := fmt.Sprintf(`package main

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "%s",
	Short: "%s",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Hello from %s!")
	},
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
`, data.ProjectName, data.Description, data.ProjectName)

	mainPath := filepath.Join(targetDir, "cmd", data.ProjectName, "main.go")
	if err := os.WriteFile(mainPath, []byte(mainGo), 0644); err != nil {
		return err
	}

	// Create README.md
	readme := fmt.Sprintf(`# %s

%s

## Getting Started

### Prerequisites

- Go %s or later

### Installation

`+"```bash"+`
go mod tidy
`+"```"+`

### Usage

`+"```bash"+`
go run ./cmd/%s
`+"```"+`

### Building

`+"```bash"+`
go build -o %s ./cmd/%s
./%s
`+"```"+`

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the %s License - see the LICENSE file for details.
`, data.ProjectName, data.Description, data.GoVersion, data.ProjectName, data.ProjectName, data.ProjectName, data.ProjectName, data.License)

	if err := os.WriteFile(filepath.Join(targetDir, "README.md"), []byte(readme), 0644); err != nil {
		return err
	}

	// Create .gitignore
	gitignore := `# Binaries for programs and plugins
*.exe
*.exe~
*.dll
*.so
*.dylib

# Test binary, built with go test -c
*.test

# Output of the go coverage tool
*.out

# Dependency directories
vendor/

# Go workspace file
go.work

# IDE files
.vscode/
.idea/
*.swp
*.swo

# OS generated files
.DS_Store
.DS_Store?
._*
.Spotlight-V100
.Trashes
ehthumbs.db
Thumbs.db

# Project specific
/tmp/
/logs/
/dist/
`

	return os.WriteFile(filepath.Join(targetDir, ".gitignore"), []byte(gitignore), 0644)
}
