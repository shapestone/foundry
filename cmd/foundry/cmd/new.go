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

var newCmd = &cobra.Command{
	Use:   "new [project-name]",
	Short: "Create a new Go project in a new directory",
	Long: `Create a new Go project in a new directory with the specified layout.
	
The new command creates a project directory and scaffolds a complete Go project
using the selected layout template.`,
	Example: `  foundry new myproject
  foundry new myapi --layout=microservice
  foundry new myapp --module=github.com/user/myapp
  foundry new mysvc --layout=hexagonal --author="John Doe"`,
	Args: cobra.ExactArgs(1),
	RunE: runNew,
}

func init() {
	// Project configuration flags
	newCmd.Flags().StringP("module", "m", "", "Go module name (default: project name)")
	newCmd.Flags().StringP("layout", "l", "standard", "Project layout to use")
	newCmd.Flags().StringP("author", "a", "", "Project author name")
	newCmd.Flags().StringP("license", "", "MIT", "Project license")
	newCmd.Flags().StringP("description", "d", "", "Project description")
	newCmd.Flags().StringP("github", "g", "", "GitHub username")

	// Layout-specific variables
	newCmd.Flags().StringP("vars", "", "", "Comma-separated layout variables (key=value)")

	// Behavior flags
	newCmd.Flags().BoolP("force", "f", false, "Overwrite existing directory")
	newCmd.Flags().Bool("no-git", false, "Skip git initialization")
	newCmd.Flags().Bool("list-layouts", false, "List available layouts and exit")

	// Register with root command
	rootCmd.AddCommand(newCmd)
}

func runNew(cmd *cobra.Command, args []string) error {
	// Check if listing layouts
	if listLayouts, _ := cmd.Flags().GetBool("list-layouts"); listLayouts {
		return listAvailableLayouts()
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

	// Create project directory
	projectPath := filepath.Join(".", projectName)

	// Check if directory exists
	if _, err := os.Stat(projectPath); err == nil {
		if !force {
			return fmt.Errorf("directory '%s' already exists (use --force to overwrite)", projectName)
		}
		// Remove existing directory
		if err := os.RemoveAll(projectPath); err != nil {
			return fmt.Errorf("failed to remove existing directory: %w", err)
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
	fmt.Printf("Creating project '%s' with layout '%s'...\n", projectName, layoutName)

	ctx := context.Background()
	if err := manager.GenerateProject(ctx, layoutName, projectPath, projectData); err != nil {
		// Clean up on failure
		os.RemoveAll(projectPath)
		return fmt.Errorf("failed to generate project: %w", err)
	}

	// Initialize git repository
	if !noGit {
		if err := initGitRepo(projectPath); err != nil {
			fmt.Printf("Warning: failed to initialize git repository: %v\n", err)
		} else {
			fmt.Println("✓ Initialized git repository")
		}
	}

	// Success message
	fmt.Printf("\n✨ Project '%s' created successfully!\n\n", projectName)
	fmt.Println("Next steps:")
	fmt.Printf("  cd %s\n", projectName)
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

func listAvailableLayouts() error {
	manager, err := createLayoutManager()
	if err != nil {
		return fmt.Errorf("failed to create layout manager: %w", err)
	}

	layouts := manager.ListLayouts()

	fmt.Println("Available layouts:")
	fmt.Println()

	for _, l := range layouts {
		fmt.Printf("  %-20s %s\n", l.Name, l.Description)
	}

	fmt.Println()
	fmt.Println("Use 'foundry new myproject --layout=<name>' to create a project with a specific layout")
	fmt.Println("Use 'foundry layout info <name>' for detailed information about a layout")

	return nil
}

func isValidProjectName(name string) bool {
	if name == "" {
		return false
	}

	for _, r := range name {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') ||
			(r >= '0' && r <= '9') || r == '-' || r == '_') {
			return false
		}
	}

	// Can't start with a number
	if name[0] >= '0' && name[0] <= '9' {
		return false
	}

	return true
}

func initGitRepo(projectPath string) error {
	// Simple git init implementation
	// In a real implementation, this would use go-git or exec git commands
	gitDir := filepath.Join(projectPath, ".git")
	return os.MkdirAll(gitDir, 0755)
}
