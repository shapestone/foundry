package commands

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

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
		Args: cobra.ExactArgs(1),
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
	cmd.Flags().StringP("vars", "", "", "Comma-separated layout variables (key=value)")
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
		return listAvailableLayouts(stdout)
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
		CustomVariables: customVars,
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

// listAvailableLayouts lists all available layouts
func listAvailableLayouts(stdout io.Writer) error {
	// TODO: Integrate with layout manager when available
	// For now, show basic layouts

	fmt.Fprintln(stdout, "📋 Available layouts:")
	fmt.Fprintln(stdout)

	layouts := []struct {
		name        string
		description string
		features    string
	}{
		{
			name:        "standard",
			description: "Standard Go project layout with cmd, internal, and pkg",
			features:    "Basic structure, Makefile, Docker support",
		},
		{
			name:        "microservice",
			description: "Microservice layout with API, gRPC, and Docker",
			features:    "REST/gRPC APIs, Docker, middleware, health checks",
		},
		{
			name:        "hexagonal",
			description: "Hexagonal architecture with domain-driven design",
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
	fmt.Fprintln(stdout, "")
	fmt.Fprintln(stdout, "🔄 Command comparison:")
	fmt.Fprintln(stdout, "  foundry new myproject    # Creates ./myproject/ directory")
	fmt.Fprintln(stdout, "  foundry init myproject   # Initializes in current directory")

	return nil
}

// Helper functions (shared with init command)
// Note: initGitRepo and isValidProjectName are defined in init.go to avoid duplication
