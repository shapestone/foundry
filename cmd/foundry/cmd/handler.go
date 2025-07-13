package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/shapestone/foundry"
	"github.com/shapestone/foundry/internal/generator"
	"github.com/shapestone/foundry/internal/project"
	"github.com/shapestone/foundry/internal/utils"
	"github.com/spf13/cobra"
)

var handlerCmd = &cobra.Command{
	Use:   "handler [name]",
	Short: "Add a new REST handler",
	Args:  cobra.ExactArgs(1),
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if len(args) > 0 {
			if err := ValidateComponentName(args[0]); err != nil {
				return fmt.Errorf("invalid handler name: %v", err)
			}
		}
		return nil
	},
	Example: `  foundry add handler user
  foundry add handler product
  foundry add handler order --dry-run
  foundry add handler user --auto-wire`,
	Run: func(cmd *cobra.Command, args []string) {
		dryRun, _ := cmd.Flags().GetBool("dry-run")
		autoWire, _ := cmd.Flags().GetBool("auto-wire")
		addHandler(args[0], dryRun, autoWire)
	},
}

func init() {
	handlerCmd.Flags().Bool("dry-run", false, "Preview changes without applying them")
	handlerCmd.Flags().Bool("auto-wire", false, "Automatically wire the handler into routes")
}

func addHandler(name string, dryRun bool, autoWire bool) {
	// Check if we're in a Go project
	if _, err := os.Stat("go.mod"); os.IsNotExist(err) {
		fmt.Println("❌ Error: go.mod not found. Please run this command from your project root")
		os.Exit(1)
	}

	// Convert name to different cases
	handlerName := strings.Title(name)
	resourceName := strings.ToLower(name)
	resourcePath := strings.ToLower(name) + "s" // simple pluralization
	resourceNamePlural := resourcePath

	fmt.Printf("🔨 Adding handler: %s\n", name)

	// Create handler file
	handlerDir := filepath.Join("internal", "handlers")
	handlerPath := filepath.Join(handlerDir, fmt.Sprintf("%s.go", resourceName))

	// Check if handler already exists
	if _, err := os.Stat(handlerPath); err == nil {
		fmt.Printf("❌ Error: handler %s already exists\n", handlerPath)
		os.Exit(1)
	}

	// Data for template
	data := struct {
		HandlerName        string
		ResourceName       string
		ResourcePath       string
		ResourceNamePlural string
	}{
		HandlerName:        handlerName,
		ResourceName:       resourceName,
		ResourcePath:       resourcePath,
		ResourceNamePlural: resourceNamePlural,
	}

	if !dryRun {
		// Read handler template
		tmplContent, err := foundry.Templates.ReadFile("templates/handler.go.tmpl")
		if err != nil {
			fmt.Printf("❌ Error reading handler template: %v\n", err)
			os.Exit(1)
		}

		// Use the new generator
		gen := generator.NewFileGenerator()
		if err := gen.Generate(handlerPath, string(tmplContent), data); err != nil {
			fmt.Printf("❌ Error creating handler: %v\n", err)
			os.Exit(1)
		}
	}

	// Handle auto-wiring
	if autoWire {
		fmt.Println("\n🔄 Auto-wiring handler...")
		if err := utils.UpdateRoutesFile(name, dryRun); err != nil {
			fmt.Printf("❌ Error auto-wiring handler: %v\n", err)
			if !dryRun {
				fmt.Println("💡 Your handler was created but you'll need to manually wire it up")
				fmt.Println("   You can try: foundry wire handler " + name)
			}
		} else {
			if !dryRun {
				showHandlerSuccess(handlerPath, resourcePath, resourceNamePlural, resourceName, true)
			}
		}
	} else {
		// Show manual wiring instructions
		if !dryRun {
			showHandlerSuccess(handlerPath, resourcePath, resourceNamePlural, resourceName, false)
		}
	}
}

func showHandlerSuccess(handlerPath, resourcePath, resourceNamePlural, resourceName string, autoWired bool) {
	moduleName := project.GetCurrentModule()

	wireStatus := ""
	if autoWired {
		wireStatus = `📝 Routes updated:
  internal/routes/routes.go

`
	} else {
		wireStatus = `📌 Manual wiring required:
  Run: foundry wire handler ` + resourceName + `
  Or manually update internal/routes/routes.go

`
	}

	fmt.Printf(`✅ Handler created successfully!

📁 Created:
  %s

%s🚀 Available endpoints:
  GET    /api/v1/%s       - List all %s
  POST   /api/v1/%s       - Create a new %s
  GET    /api/v1/%s/{id}  - Get %s by ID
  PUT    /api/v1/%s/{id}  - Update %s by ID
  DELETE /api/v1/%s/{id}  - Delete %s by ID

💡 Next steps:
  - Import the handlers package in your routes file:
    import "%s/internal/handlers"
  - Implement your business logic in %s
  - Add validation and error handling
  - Connect to your database or service layer
`, handlerPath,
		wireStatus,
		resourcePath, resourceNamePlural,
		resourcePath, resourceName,
		resourcePath, resourceName,
		resourcePath, resourceName,
		resourcePath, resourceName,
		moduleName,
		handlerPath)
}
