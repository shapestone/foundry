package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/shapestone/foundry/internal/project"
	"github.com/shapestone/foundry/internal/utils"
	"github.com/spf13/cobra"
)

var wireHandlerCmd = &cobra.Command{
	Use:   "handler [name]",
	Short: "Wire an existing handler into routes",
	Args:  cobra.ExactArgs(1),
	Example: `  foundry wire handler user
  foundry wire handler product`,
	Run: func(cmd *cobra.Command, args []string) {
		wireHandler(args[0])
	},
}

func wireHandler(name string) {
	// Check if we're in a Go project
	if _, err := os.Stat("go.mod"); os.IsNotExist(err) {
		fmt.Println("❌ Error: go.mod not found. Please run this command from your project root")
		os.Exit(1)
	}

	// Convert name to different cases
	resourceName := strings.ToLower(name)
	resourcePath := strings.ToLower(name) + "s" // simple pluralization

	// Check if handler exists
	handlerPath := filepath.Join("internal", "handlers", fmt.Sprintf("%s.go", resourceName))
	if _, err := os.Stat(handlerPath); os.IsNotExist(err) {
		fmt.Printf("❌ Error: handler %s not found at %s\n", name, handlerPath)
		fmt.Println("💡 Did you mean to run: foundry add handler " + name)
		os.Exit(1)
	}

	// Check if routes file exists
	routesPath := filepath.Join("internal", "routes", "routes.go")
	if _, err := os.Stat(routesPath); os.IsNotExist(err) {
		fmt.Printf("❌ Error: routes.go not found at %s\n", routesPath)
		os.Exit(1)
	}

	fmt.Printf("🔌 Wiring handler: %s\n", name)

	// Update routes file
	if err := utils.UpdateRoutesFile(name, false); err != nil {
		if err.Error() == "update cancelled by user" {
			// User cancelled - this is not an error, exit gracefully
			fmt.Println("ℹ️  Wire operation cancelled")
			return
		}
		fmt.Printf("❌ Error wiring handler: %v\n", err)
		os.Exit(1)
	}

	// Success message
	moduleName := project.GetCurrentModule()
	fmt.Printf(`✅ Handler wired successfully!

📝 Routes updated:
  internal/routes/routes.go

🚀 Available endpoints:
  GET    /api/v1/%s       - List all %s
  POST   /api/v1/%s       - Create new %s
  GET    /api/v1/%s/{id}  - Get by ID
  PUT    /api/v1/%s/{id}  - Update by ID
  DELETE /api/v1/%s/{id}  - Delete by ID

💡 Next steps:
  - Ensure your handler is imported: import "%s/internal/handlers"
  - Test your endpoints: curl http://localhost:8080/api/v1/%s
`, resourcePath, resourcePath,
		resourcePath, resourceName,
		resourcePath,
		resourcePath,
		resourcePath,
		moduleName, resourcePath)
}
