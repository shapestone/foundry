package generator

import (
	"fmt"
	"os"
	"strings"
)

// Public API functions for the CLI commands
// These maintain backward compatibility while using the new template system

var (
	// Global component generator instance
	componentGen *ComponentGenerator
)

// init initializes the global component generator
func init() {
	componentGen = NewComponentGenerator()
}

// CreateHandler creates a new HTTP handler
func CreateHandler(name string) error {
	// Get current working directory as the project path
	projectPath, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	err = componentGen.GenerateHandler(name, projectPath)
	if err == nil {
		// Show the actual filename that was created
		filename := fmt.Sprintf("%s_handler.go", toSnakeCase(name))
		fmt.Printf("✅ Handler %s created successfully!\n", name)
		fmt.Printf("📁 File created: %s\n", filename)
		fmt.Printf("💡 To wire this handler into your routes, run:\n")
		fmt.Printf("   foundry wire handler %s\n", name)
	}
	return err
}

// CreateModel creates a new data model
func CreateModel(name string) error {
	// Get current working directory as the project path
	projectPath, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	err = componentGen.GenerateModel(name, projectPath)
	if err == nil {
		// Show the actual filename that was created
		filename := fmt.Sprintf("%s_model.go", toSnakeCase(name))
		fmt.Printf("✅ Model %s created successfully!\n", name)
		fmt.Printf("📁 File created: %s\n", filename)
		fmt.Printf("💡 To wire this model into your project, run:\n")
		fmt.Printf("   foundry wire model %s\n", name)
	}
	return err
}

// CreateMiddleware creates a new middleware
func CreateMiddleware(name string) error {
	// Get current working directory as the project path
	projectPath, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	err = componentGen.GenerateMiddleware(name, projectPath)
	if err == nil {
		// Show the actual filename that was created
		filename := fmt.Sprintf("%s_middleware.go", toSnakeCase(name))
		fmt.Printf("✅ Middleware %s created successfully!\n", name)
		fmt.Printf("📁 File created: %s\n", filename)
		fmt.Printf("💡 To wire this middleware into your routes, run:\n")
		fmt.Printf("   foundry wire middleware %s\n", name)
	}
	return err
}

// CreateService creates a new service
func CreateService(name string) error {
	// Get current working directory as the project path
	projectPath, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	err = componentGen.GenerateService(name, projectPath)
	if err == nil {
		// Show the actual filename that was created
		filename := fmt.Sprintf("%s_service.go", toSnakeCase(name))
		fmt.Printf("✅ Service %s created successfully!\n", name)
		fmt.Printf("📁 File created: %s\n", filename)
		fmt.Printf("💡 To wire this service into your project, run:\n")
		fmt.Printf("   foundry wire service %s\n", name)
	}
	return err
}

// CreateProject creates a new project
func CreateProject(name, path string) error {
	integration := NewTemplateIntegration()

	// Determine module path from project name and path
	modulePath := fmt.Sprintf("github.com/example/%s", name)
	if path == "" {
		path = name
	}

	err := integration.GenerateProjectWithNewSystem(name, modulePath, path)
	if err == nil {
		fmt.Printf("✅ Project %s created successfully!\n", name)
		fmt.Printf("📁 Project created in: %s\n", path)
		fmt.Printf("💡 To get started:\n")
		fmt.Printf("   cd %s\n", path)
		fmt.Printf("   go mod tidy\n")
		fmt.Printf("   go run main.go\n")
	}
	return err
}

// SetProjectPath allows setting a custom project path (for testing)
func SetProjectPath(path string) {
	// This could be used for testing or when working with projects in different directories
	// For now, we'll just validate the path exists
	if _, err := os.Stat(path); err != nil {
		fmt.Printf("Warning: Project path %s may not exist: %v\n", path, err)
	}
}

// GetSupportedComponents returns a list of supported component types
func GetSupportedComponents() []string {
	return []string{"handler", "model", "middleware", "service"}
}

// ValidateComponentType checks if a component type is supported
func ValidateComponentType(componentType string) bool {
	supported := GetSupportedComponents()
	for _, t := range supported {
		if t == componentType {
			return true
		}
	}
	return false
}

// Helper function to convert to snake_case (matching template_integration.go)
func toSnakeCase(s string) string {
	if s == "" {
		return s
	}

	var result strings.Builder
	for i, r := range s {
		if i > 0 && (r >= 'A' && r <= 'Z') {
			result.WriteRune('_')
		}
		result.WriteRune(r)
	}

	return strings.ToLower(result.String())
}
