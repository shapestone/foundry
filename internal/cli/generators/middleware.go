package generators

import (
	"fmt"
	"io"
	"path/filepath"

	"github.com/shapestone/foundry/internal/cli/templates"
)

// MiddlewareGenerator handles middleware file generation
type MiddlewareGenerator struct {
	stdout io.Writer
	stderr io.Writer
}

// NewMiddlewareGenerator creates a new middleware generator
func NewMiddlewareGenerator(stdout, stderr io.Writer) *MiddlewareGenerator {
	return &MiddlewareGenerator{
		stdout: stdout,
		stderr: stderr,
	}
}

// MiddlewareOptions holds options for middleware generation
type MiddlewareOptions struct {
	Type      string
	AutoWire  bool
	OutputDir string
}

// Generate creates middleware files based on options
func (g *MiddlewareGenerator) Generate(options MiddlewareOptions) error {
	middlewareInfo, ok := templates.GetMiddlewareInfo(options.Type)
	if !ok {
		return fmt.Errorf("unsupported middleware type: %s", options.Type)
	}

	// Create middleware file
	middlewarePath := filepath.Join(options.OutputDir, fmt.Sprintf("%s.go", options.Type))
	if err := g.createMiddlewareFile(middlewarePath, options.Type); err != nil {
		return fmt.Errorf("failed to create middleware file: %w", err)
	}

	// Handle auto-wiring
	if options.AutoWire {
		fmt.Fprintln(g.stdout, "\n🔄 Auto-wiring middleware...")
		if err := g.wireMiddleware(options.Type); err != nil {
			fmt.Fprintf(g.stderr, "❌ Error auto-wiring middleware: %v\n", err)
			fmt.Fprintln(g.stdout, "💡 Your middleware was created but you'll need to manually wire it up")
			fmt.Fprintf(g.stdout, "   You can try: foundry wire middleware %s\n", options.Type)
			g.showSuccess(options, middlewareInfo, false)
		} else {
			g.showSuccess(options, middlewareInfo, true)
		}
	} else {
		g.showSuccess(options, middlewareInfo, false)
	}

	return nil
}

// createMiddlewareFile creates a middleware file based on type
func (g *MiddlewareGenerator) createMiddlewareFile(middlewarePath, middlewareType string) error {
	template := templates.GetMiddlewareTemplate(middlewareType)
	return writeFile(middlewarePath, template)
}

// wireMiddleware attempts to auto-wire middleware into the application
func (g *MiddlewareGenerator) wireMiddleware(middlewareType string) error {
	// TODO: Implement auto-wiring logic
	// For now, return a placeholder implementation

	// Placeholder - in a real implementation, this would:
	// 1. Find main.go or router files
	// 2. Parse the Go code
	// 3. Add middleware to the middleware chain
	// 4. Update imports if needed
	fmt.Fprintf(g.stdout, "⚠️  Auto-wiring not yet implemented - use manual instructions\n")
	return fmt.Errorf("auto-wiring not implemented")
}

// showSuccess displays success message with instructions
func (g *MiddlewareGenerator) showSuccess(options MiddlewareOptions, middlewareInfo templates.MiddlewareInfo, autoWired bool) {
	middlewarePath := filepath.Join(options.OutputDir, fmt.Sprintf("%s.go", options.Type))

	wireStatus := ""
	if autoWired {
		wireStatus = `
🔌 Middleware automatically wired into your router!

`
	} else {
		wireStatus = `
📌 Manual wiring required:
  Run: foundry wire middleware ` + options.Type + `
  Or manually update your main.go file

`
	}

	usage := templates.GetMiddlewareUsage(options.Type)

	fmt.Fprintf(g.stdout, `✅ Middleware created successfully!

📁 Created:
  %s

📝 Description:
  %s
%s%s
`, middlewarePath, middlewareInfo.Description, wireStatus, usage)
}
