package commands

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/shapestone/foundry/internal/layout"
	"github.com/spf13/cobra"
)

// BuildNewCommand creates the new command using the adapter pattern
func BuildNewCommand(adapter *CLIAdapter) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "new [project-name]",
		Short: "Create a new Go project IN A NEW DIRECTORY",
		Long: `Create a new Go project in a new directory with the specified layout.

IMPORTANT: The new command creates a NEW DIRECTORY for your project.
This is different from 'foundry init' which initializes in the current directory.

This follows the same pattern as 'git clone' - it creates a new directory.
Use 'foundry init' if you want to initialize in the current directory.`,
		Example: `  # Create new project directory (like 'git clone')
  foundry new myproject         # Creates ./myproject/ directory
  foundry new myapi --layout=microservice
  foundry new myapp --module=github.com/user/myapp
  foundry new mysvc --layout=hexagonal --author="John Doe"

  # Compare with 'init' command (current directory):
  mkdir myproject && cd myproject
  foundry init                  # Initializes in current directory

  # List available layouts
  foundry new --list-layouts`,
		Args: func(cmd *cobra.Command, args []string) error {
			// Allow no arguments if --list-layouts flag is set
			if listLayouts, _ := cmd.Flags().GetBool("list-layouts"); listLayouts {
				return nil
			}
			// Otherwise require exactly 1 argument
			return cobra.ExactArgs(1)(cmd, args)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runNew(cmd, args, adapter)
		},
	}

	// Project configuration flags
	cmd.Flags().StringP("module", "m", "", "Go module name (default: project name)")
	cmd.Flags().StringP("layout", "l", "standard", "Project layout to use")
	cmd.Flags().StringP("author", "a", "", "Project author name")
	cmd.Flags().StringP("license", "", "MIT", "Project license")
	cmd.Flags().StringP("description", "d", "", "Project description")
	cmd.Flags().StringP("github", "g", "", "GitHub username")
	cmd.Flags().BoolP("force", "f", false, "Overwrite existing directory")
	cmd.Flags().Bool("no-git", false, "Skip git initialization")
	cmd.Flags().Bool("list-layouts", false, "List available layouts and exit")

	return cmd
}

// runNew executes the new command
func runNew(cmd *cobra.Command, args []string, adapter *CLIAdapter) error {
	stdout := adapter.GetStdout()
	stderr := adapter.GetStderr()

	// Check if listing layouts
	if listLayouts, _ := cmd.Flags().GetBool("list-layouts"); listLayouts {
		return listAvailableLayouts(stdout, adapter)
	}

	projectName := args[0]

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

	// Create project directory path
	currentDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}
	projectPath := filepath.Join(currentDir, projectName)

	// Enhanced directory existence check with clearer messaging
	if _, err := os.Stat(projectPath); err == nil {
		if !force {
			fmt.Fprintf(stderr, "⚠️  Directory already exists: %s\n", projectPath)
			fmt.Fprintln(stderr, "")
			fmt.Fprintln(stderr, "Options:")
			fmt.Fprintln(stderr, "  1. Use --force to overwrite the existing directory")
			fmt.Fprintln(stderr, "  2. Choose a different project name")
			fmt.Fprintln(stderr, "  3. Use 'foundry init' to initialize in the existing directory")
			return fmt.Errorf("directory '%s' already exists (use --force to overwrite)", projectName)
		}
		// Remove existing directory
		fmt.Fprintf(stdout, "🗑️  Removing existing directory: %s\n", projectPath)
		if err := os.RemoveAll(projectPath); err != nil {
			return fmt.Errorf("failed to remove existing directory: %w", err)
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

	// Clear creation message showing new directory location
	fmt.Fprintf(stdout, "🚀 Creating new project '%s'...\n", projectName)
	fmt.Fprintf(stdout, "📁 Creating directory: %s\n", projectPath)
	fmt.Fprintf(stdout, "🏗️  Layout: %s\n", layoutName)
	fmt.Fprintln(stdout, "")

	if err := generateProject(layoutName, projectPath, projectData, stdout, stderr); err != nil {
		// Clean up on failure
		os.RemoveAll(projectPath)
		return fmt.Errorf("failed to generate project: %w", err)
	}

	// Initialize git repository
	if !noGit {
		if err := initGitRepo(projectPath); err != nil {
			fmt.Fprintf(stderr, "Warning: failed to initialize git repository: %v\n", err)
		} else {
			fmt.Fprintln(stdout, "✓ Initialized git repository")

			// Create initial commit in the project directory
			if err := createInitialCommitInDir(projectPath); err != nil {
				fmt.Fprintf(stderr, "Warning: failed to create initial commit: %v\n", err)
			}
		}
	}

	// Enhanced success message with clear location and next steps
	fmt.Fprintln(stdout, "")
	fmt.Fprintf(stdout, "✨ Project '%s' created successfully!\n", projectName)
	fmt.Fprintf(stdout, "📍 Project location: %s\n", projectPath)
	fmt.Fprintln(stdout, "")
	fmt.Fprintln(stdout, "🎯 Next steps:")
	fmt.Fprintf(stdout, "  cd %s           # Enter project directory\n", projectName)
	fmt.Fprintln(stdout, "  go mod tidy        # Download dependencies")
	fmt.Fprintln(stdout, "  make build         # Build the project")
	fmt.Fprintln(stdout, "  make run           # Run the project")
	fmt.Fprintln(stdout, "")
	fmt.Fprintln(stdout, "🔧 To add components (after cd):")
	fmt.Fprintln(stdout, "  foundry add handler users")
	fmt.Fprintln(stdout, "  foundry add model product")
	fmt.Fprintln(stdout, "  foundry add middleware auth")
	fmt.Fprintln(stdout, "")
	fmt.Fprintln(stdout, "💡 Tip: Use 'foundry init' next time to initialize in current directory")

	return nil
}

// listAvailableLayouts lists all available layouts using the layout manager
func listAvailableLayouts(stdout io.Writer, adapter *CLIAdapter) error {
	// Get layout manager
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}
	configPath := filepath.Join(homeDir, ".foundry", "layouts.yaml")

	manager, err := layout.NewManager(configPath)
	if err != nil {
		// If layout manager fails, show basic layouts
		return showBasicLayouts(stdout)
	}

	// Get layouts from manager
	layouts := manager.ListLayouts()

	fmt.Fprintln(stdout, "📋 Available layouts:")
	fmt.Fprintln(stdout)

	if len(layouts) == 0 {
		return showBasicLayouts(stdout)
	}

	for _, l := range layouts {
		fmt.Fprintf(stdout, "  🏗️  %-15s %s\n", l.Name, l.Description)
		if l.Source.Type != "local" {
			fmt.Fprintf(stdout, "      Source: %s (%s)\n", l.Source.Type, l.Source.Location)
		}
		fmt.Fprintln(stdout)
	}

	fmt.Fprintln(stdout, "📖 Usage:")
	fmt.Fprintln(stdout, "  foundry new myproject --layout=<name>     # Create with specific layout")
	fmt.Fprintln(stdout, "  foundry layout info <name>               # Detailed layout information")
	fmt.Fprintln(stdout, "  foundry layout list                      # List all layouts with details")
	fmt.Fprintln(stdout, "")
	fmt.Fprintln(stdout, "🔄 Command comparison:")
	fmt.Fprintln(stdout, "  foundry new myproject    # Creates ./myproject/ directory")
	fmt.Fprintln(stdout, "  foundry init myproject   # Initializes in current directory")

	return nil
}

// showBasicLayouts shows fallback layout list when layout manager is unavailable
func showBasicLayouts(stdout io.Writer) error {
	fmt.Fprintln(stdout, "📋 Available layouts:")
	fmt.Fprintln(stdout)

	layouts := []struct {
		name        string
		description string
		features    string
	}{
		{
			name:        "standard",
			description: "Standard Go project layout with Gorilla Mux router",
			features:    "HTTP server, middleware, Docker, Makefile",
		},
		{
			name:        "microservice",
			description: "Microservice layout with API, gRPC, and Docker",
			features:    "REST/gRPC APIs, Docker, middleware, health checks",
		},
		{
			name:        "hexagonal",
			description: "hexagonal architecture with domain-driven design",
			features:    "Clean architecture, DDD patterns, dependency injection",
		},
		{
			name:        "minimal",
			description: "Minimal Go project with basic structure",
			features:    "Lightweight, essential files only",
		},
	}

	for _, l := range layouts {
		fmt.Fprintf(stdout, "  🏗️  %-15s %s\n", l.name, l.description)
		fmt.Fprintf(stdout, "      Features: %s\n", l.features)
		fmt.Fprintln(stdout)
	}

	fmt.Fprintln(stdout, "📖 Usage:")
	fmt.Fprintln(stdout, "  foundry new myproject --layout=<name>     # Create with specific layout")
	fmt.Fprintln(stdout, "  foundry layout info <name>               # Detailed layout information")
	fmt.Fprintln(stdout, "  foundry layout list                      # List all layouts with details")

	return nil
}

// createInitialCommitInDir creates the initial git commit in a specific directory
func createInitialCommitInDir(dir string) error {
	// Add all files
	cmd := exec.Command("git", "add", ".")
	cmd.Dir = dir
	if err := cmd.Run(); err != nil {
		return err
	}

	// Create initial commit
	cmd = exec.Command("git", "commit", "-m", "Initial commit")
	cmd.Dir = dir
	return cmd.Run()
}

// NOTE: initGitRepo function is defined in init.go - removed duplicate
// isValidProjectName is also defined in init.go - will use that one
// generateProject is also defined in init.go - will use that one
