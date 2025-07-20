package generators

import (
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/shapestone/foundry/internal/cli/templates"
	"github.com/shapestone/foundry/internal/routes"
)

// HandlerGenerator handles handler file generation
type HandlerGenerator struct {
	stdout io.Writer
	stderr io.Writer
}

// NewHandlerGenerator creates a new handler generator
func NewHandlerGenerator(stdout, stderr io.Writer) *HandlerGenerator {
	return &HandlerGenerator{
		stdout: stdout,
		stderr: stderr,
	}
}

// HandlerOptions holds options for handler generation
type HandlerOptions struct {
	Name      string
	AutoWire  bool
	OutputDir string
}

// Generate creates handler files based on options
func (g *HandlerGenerator) Generate(options HandlerOptions) error {
	// Create handler file
	handlerPath := filepath.Join(options.OutputDir, fmt.Sprintf("%s.go", strings.ToLower(options.Name)))
	if err := g.createHandlerFile(handlerPath, options.Name); err != nil {
		return fmt.Errorf("failed to create handler file: %w", err)
	}

	// Handle auto-wiring
	if options.AutoWire {
		fmt.Fprintln(g.stdout, "\n🔄 Auto-wiring handler...")
		if err := g.wireHandler(options.Name); err != nil {
			fmt.Fprintf(g.stderr, "❌ Error auto-wiring handler: %v\n", err)
			fmt.Fprintln(g.stdout, "💡 Your handler was created but you'll need to manually wire it up")
			fmt.Fprintf(g.stdout, "   You can try: foundry wire handler %s\n", options.Name)
			g.showSuccess(options, false)
		} else {
			g.showSuccess(options, true)
		}
	} else {
		g.showSuccess(options, false)
	}

	return nil
}

// createHandlerFile creates a handler file
func (g *HandlerGenerator) createHandlerFile(handlerPath, name string) error {
	template := templates.GetHandlerTemplate(name)
	return writeFile(handlerPath, template)
}

// wireHandler attempts to auto-wire handler into routes
func (g *HandlerGenerator) wireHandler(name string) error {
	// Get current module name
	moduleName := getCurrentModule()
	if moduleName == "" {
		return fmt.Errorf("could not determine module name")
	}

	// Create route generator
	generator := routes.NewFileGenerator()

	// Calculate the required changes
	update, err := generator.UpdateRoutes(strings.ToLower(name), moduleName)
	if err != nil {
		return fmt.Errorf("calculating route updates: %w", err)
	}

	// Show what changes will be made
	fmt.Fprintf(g.stdout, "📝 Updating routes file: %s\n", update.Path)
	for _, change := range update.Changes {
		fmt.Fprintf(g.stdout, "  - %s\n", change)
	}

	// Apply the changes
	if err := routes.ApplyUpdate(update, generator); err != nil {
		return fmt.Errorf("applying route updates: %w", err)
	}

	fmt.Fprintf(g.stdout, "✅ Routes updated successfully!\n")
	return nil
}

// showSuccess displays success message with instructions
func (g *HandlerGenerator) showSuccess(options HandlerOptions, autoWired bool) {
	handlerPath := filepath.Join(options.OutputDir, fmt.Sprintf("%s.go", strings.ToLower(options.Name)))
	resourcePath := strings.ToLower(options.Name) + "s" // simple pluralization

	wireStatus := ""
	if autoWired {
		wireStatus = `📝 Routes updated:
  internal/routes/routes.go

`
	} else {
		wireStatus = `📌 Manual wiring required:
  Run: foundry wire handler ` + options.Name + `
  Or manually update internal/routes/routes.go

`
	}

	usage := templates.GetHandlerUsage(options.Name)

	fmt.Fprintf(g.stdout, `✅ Handler created successfully!

📁 Created:
  %s

%s🚀 Available endpoints:
  GET    /api/v1/%s       - List all %s
  POST   /api/v1/%s       - Create a new %s
  GET    /api/v1/%s/{id}  - Get %s by ID
  PUT    /api/v1/%s/{id}  - Update %s by ID
  DELETE /api/v1/%s/{id}  - Delete %s by ID

%s
`, handlerPath,
		wireStatus,
		resourcePath, resourcePath,
		resourcePath, strings.ToLower(options.Name),
		resourcePath, strings.ToLower(options.Name),
		resourcePath, strings.ToLower(options.Name),
		resourcePath, strings.ToLower(options.Name),
		usage)
}
