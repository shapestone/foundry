package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var addCmd = &cobra.Command{
	Use:   "add [component-type] [name]",
	Short: "Add a new component to the current project",
	Long: `Add a new component to the current project using the project's layout templates.
	
The add command generates boilerplate code for common components like handlers, models,
middleware, services, and more. The available component types depend on the project layout.`,
	Example: `  foundry add handler users
  foundry add model product
  foundry add middleware auth
  foundry add service payment
  foundry add repository user`,
	Args: cobra.ExactArgs(2),
	RunE: runAdd,
}

func init() {
	// Component configuration flags
	addCmd.Flags().StringP("output", "o", "", "Custom output directory")
	addCmd.Flags().BoolP("force", "f", false, "Overwrite existing files")
	addCmd.Flags().Bool("dry-run", false, "Show what would be generated without creating files")

	// Add subcommands
	addCmd.AddCommand(handlerCmd)
	addCmd.AddCommand(modelCmd)
	addCmd.AddCommand(middlewareCmd)
	addCmd.AddCommand(databaseCmd)

	// Register with root command
	rootCmd.AddCommand(addCmd)
}

func runAdd(cmd *cobra.Command, args []string) error {
	componentType := args[0]
	componentName := args[1]

	// Validate component name
	if err := ValidateComponentName(componentName); err != nil {
		return fmt.Errorf("invalid component name: %v", err)
	}

	// Get flags
	outputDir, _ := cmd.Flags().GetString("output")
	force, _ := cmd.Flags().GetBool("force")
	dryRun, _ := cmd.Flags().GetBool("dry-run")

	// Check if we're in a Foundry project
	projectLayout, err := detectProjectLayout()
	if err != nil {
		return fmt.Errorf("not in a Foundry project directory: %w", err)
	}

	// Create layout manager
	manager, err := createLayoutManager()
	if err != nil {
		return fmt.Errorf("failed to create layout manager: %w", err)
	}

	// Get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	// Load the layout to check available components
	ctx := context.Background()
	layout, err := manager.GetLayout(ctx, projectLayout)
	if err != nil {
		return fmt.Errorf("failed to load layout '%s': %w", projectLayout, err)
	}

	// Check if component type is available
	component, exists := layout.Manifest.Components[componentType]
	if !exists {
		// List available components
		var available []string
		for name := range layout.Manifest.Components {
			available = append(available, name)
		}

		if len(available) == 0 {
			return fmt.Errorf("layout '%s' does not define any components", projectLayout)
		}

		return fmt.Errorf("unknown component type '%s'. Available types: %s",
			componentType, strings.Join(available, ", "))
	}

	// Determine output directory
	if outputDir == "" {
		outputDir = component.TargetDir
	}

	// Check if file already exists
	targetFile := filepath.Join(outputDir, toFileName(componentName)+".go")
	if !force && !dryRun {
		if _, err := os.Stat(targetFile); err == nil {
			return fmt.Errorf("file already exists: %s (use --force to overwrite)", targetFile)
		}
	}

	// Show what would be done in dry-run mode
	if dryRun {
		fmt.Printf("Would generate %s '%s' at: %s\n", componentType, componentName, targetFile)
		fmt.Println("\nTemplate that would be used:")
		fmt.Printf("  %s\n", component.Template)
		fmt.Println("\nVariables available in template:")
		fmt.Printf("  Name: %s\n", componentName)
		fmt.Printf("  PackageName: %s\n", toPackageName(filepath.Base(outputDir)))
		fmt.Printf("  Type: %s\n", componentType)
		return nil
	}

	// Generate the component
	if err := manager.GenerateComponent(ctx, projectLayout, componentType, componentName, cwd); err != nil {
		return fmt.Errorf("failed to generate component: %w", err)
	}

	// Success message
	fmt.Printf("✓ Generated %s '%s' at: %s\n", componentType, componentName, targetFile)

	// Provide next steps based on component type
	printNextSteps(componentType, componentName)

	return nil
}

func detectProjectLayout() (string, error) {
	// Look for foundry.yaml or .foundry.yaml
	configFiles := []string{"foundry.yaml", ".foundry.yaml", "foundry.yml", ".foundry.yml"}

	for _, configFile := range configFiles {
		if _, err := os.Stat(configFile); err == nil {
			// Read the config file
			data, err := os.ReadFile(configFile)
			if err != nil {
				return "", fmt.Errorf("failed to read config file: %w", err)
			}

			// Parse the config
			var config struct {
				Layout string `yaml:"layout"`
			}
			if err := yaml.Unmarshal(data, &config); err != nil {
				return "", fmt.Errorf("failed to parse config file: %w", err)
			}

			if config.Layout != "" {
				return config.Layout, nil
			}

			// Default to standard if not specified
			return "standard", nil
		}
	}

	// Check for go.mod to confirm we're in a Go project
	if _, err := os.Stat("go.mod"); err != nil {
		return "", fmt.Errorf("no go.mod found")
	}

	// Default to standard layout
	return "standard", nil
}

func isValidComponentName(name string) bool {
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

func toFileName(name string) string {
	// Convert to snake_case for file names
	return toSnakeCase(name)
}

func toPackageName(name string) string {
	// Package names should be lowercase without underscores
	return strings.ToLower(strings.ReplaceAll(name, "_", ""))
}

func toSnakeCase(s string) string {
	var result []rune
	for i, r := range s {
		if r >= 'A' && r <= 'Z' && i > 0 {
			// Add underscore before uppercase letters (except first)
			if i > 0 && s[i-1] >= 'a' && s[i-1] <= 'z' {
				result = append(result, '_')
			}
		}
		result = append(result, toLowerRune(r))
	}
	return string(result)
}

func toLowerRune(r rune) rune {
	if r >= 'A' && r <= 'Z' {
		return r + ('a' - 'A')
	}
	return r
}

func printNextSteps(componentType, componentName string) {
	fmt.Println("\nNext steps:")

	switch componentType {
	case "handler":
		fmt.Printf("  1. Add routes for %s in your router configuration\n", componentName)
		fmt.Println("  2. Implement the handler methods based on your requirements")
		fmt.Println("  3. Add tests for the handler")

	case "model":
		fmt.Printf("  1. Run migrations to create the %s table\n", componentName)
		fmt.Println("  2. Implement any custom methods needed")
		fmt.Println("  3. Add validation rules as required")

	case "middleware":
		fmt.Printf("  1. Add the %s middleware to your router or specific routes\n", componentName)
		fmt.Println("  2. Configure the middleware options as needed")
		fmt.Println("  3. Add tests for the middleware")

	case "service":
		fmt.Printf("  1. Inject the %s service where needed\n", componentName)
		fmt.Println("  2. Implement the service methods")
		fmt.Println("  3. Add unit tests for the service")

	case "repository":
		fmt.Printf("  1. Inject the %s repository into your services\n", componentName)
		fmt.Println("  2. Implement the data access methods")
		fmt.Println("  3. Add integration tests for the repository")

	default:
		fmt.Println("  1. Review the generated code and customize as needed")
		fmt.Println("  2. Add appropriate tests")
		fmt.Println("  3. Update documentation")
	}
}
