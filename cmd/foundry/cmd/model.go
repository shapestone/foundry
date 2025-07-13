package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/shapestone/foundry"
	"github.com/shapestone/foundry/internal/generator"
	"github.com/shapestone/foundry/internal/project"
	"github.com/spf13/cobra"
)

var modelCmd = &cobra.Command{
	Use:   "model [name]",
	Short: "Add a new data model",
	Args:  cobra.ExactArgs(1),
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if len(args) > 0 {
			if err := ValidateComponentName(args[0]); err != nil {
				return fmt.Errorf("invalid model name: %v", err)
			}
		}
		return nil
	},
	Example: `  foundry add model user
  foundry add model product
  foundry add model order`,
	Run: func(cmd *cobra.Command, args []string) {
		addModel(args[0])
	},
}

func init() {
	// Add this command to the 'add' parent command
	// This goes in add.go's init function
}

func addModel(name string) {
	// Check if we're in a Go project
	if _, err := os.Stat("go.mod"); os.IsNotExist(err) {
		fmt.Println("❌ Error: go.mod not found. Please run this command from your project root")
		os.Exit(1)
	}

	// Convert name to different cases
	modelName := strings.Title(name)
	resourceName := strings.ToLower(name)

	fmt.Printf("🔨 Adding model: %s\n", name)

	// Create models directory
	modelsDir := filepath.Join("internal", "models")
	modelPath := filepath.Join(modelsDir, fmt.Sprintf("%s.go", resourceName))

	// Check if model already exists
	if _, err := os.Stat(modelPath); err == nil {
		fmt.Printf("❌ Error: model %s already exists\n", modelPath)
		os.Exit(1)
	}

	// Determine which fields to include based on model name
	data := struct {
		ModelName               string
		ResourceName            string
		IncludeNameField        bool
		IncludeEmailField       bool
		IncludeTitleField       bool
		IncludeDescriptionField bool
	}{
		ModelName:    modelName,
		ResourceName: resourceName,
		// Smart defaults based on model name
		IncludeNameField:        name == "user" || name == "customer" || name == "person",
		IncludeEmailField:       name == "user" || name == "customer" || name == "person",
		IncludeTitleField:       name == "post" || name == "article" || name == "product",
		IncludeDescriptionField: name == "post" || name == "article" || name == "product" || name == "project",
	}

	// Read model template
	tmplContent, err := foundry.Templates.ReadFile("templates/model.go.tmpl")
	if err != nil {
		fmt.Printf("❌ Error reading model template: %v\n", err)
		os.Exit(1)
	}

	// Use the new generator
	gen := generator.NewFileGenerator()
	if err := gen.Generate(modelPath, string(tmplContent), data); err != nil {
		fmt.Printf("❌ Error creating model: %v\n", err)
		os.Exit(1)
	}

	// Show success message
	moduleName := project.GetCurrentModule()
	fmt.Printf(`✅ Model created successfully!

📁 Created:
  %s

📝 Model includes:
  - ID field (string)
  - Timestamps (created_at, updated_at)
  %s
  - Validate() method
  - Update() method
  - Constructor function

💡 Next steps:
  - Implement generateID() based on your needs (UUID, ULID, etc.)
  - Add custom fields specific to your %s
  - Add methods for your business logic
  - Use in your handlers:
    
    import "%s/internal/models"
    
    %s := models.New%s()
    if err := %s.Validate(); err != nil {
        // handle error
    }
`, modelPath, getFieldsList(data), resourceName, moduleName, resourceName, modelName, resourceName)
}

func getFieldsList(data struct {
	ModelName               string
	ResourceName            string
	IncludeNameField        bool
	IncludeEmailField       bool
	IncludeTitleField       bool
	IncludeDescriptionField bool
}) string {
	fields := []string{}
	if data.IncludeNameField {
		fields = append(fields, "- Name field (string)")
	}
	if data.IncludeEmailField {
		fields = append(fields, "- Email field (string)")
	}
	if data.IncludeTitleField {
		fields = append(fields, "- Title field (string)")
	}
	if data.IncludeDescriptionField {
		fields = append(fields, "- Description field (string)")
	}
	return strings.Join(fields, "\n  ")
}
