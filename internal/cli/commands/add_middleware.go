package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/shapestone/foundry/internal/cli/generators"
	"github.com/shapestone/foundry/internal/cli/templates"
	"github.com/spf13/cobra"
)

// BuildAddMiddlewareCommand creates the add middleware subcommand
func BuildAddMiddlewareCommand(c CLI) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "middleware [type]",
		Short: "Add middleware to your project",
		Args:  cobra.ExactArgs(1),
		Example: `  foundry add middleware auth
  foundry add middleware ratelimit
  foundry add middleware cors
  foundry add middleware logging
  foundry add middleware recovery
  foundry add middleware timeout
  foundry add middleware compression`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runAddMiddleware(c, cmd, args)
		},
	}

	cmd.Flags().Bool("auto-wire", false, "Automatically wire the middleware into your router")
	cmd.Flags().Bool("dry-run", false, "Preview changes without applying them")

	return cmd
}

// runAddMiddleware executes the add middleware subcommand
func runAddMiddleware(c CLI, cmd *cobra.Command, args []string) error {
	middlewareType := args[0]

	// Get flags
	autoWire, _ := cmd.Flags().GetBool("auto-wire")
	dryRun, _ := cmd.Flags().GetBool("dry-run")

	// Check if we're in a Go project
	if _, err := os.Stat("go.mod"); os.IsNotExist(err) {
		return fmt.Errorf("go.mod not found. Please run this command from your project root")
	}

	// Validate middleware type
	if !templates.IsSupportedMiddleware(middlewareType) {
		fmt.Fprintf(c.GetStderr(), "❌ Error: unsupported middleware type '%s'\n", middlewareType)
		fmt.Fprintln(c.GetStderr(), "Supported types:")
		for _, mw := range templates.GetSupportedMiddleware() {
			fmt.Fprintf(c.GetStderr(), "  - %s: %s\n", mw.Type, mw.Description)
		}
		return fmt.Errorf("unsupported middleware type: %s", middlewareType)
	}

	fmt.Fprintf(c.GetStdout(), "🔨 Adding middleware: %s\n", middlewareType)

	// Create middleware directory
	middlewareDir := filepath.Join("internal", "middleware")
	middlewarePath := filepath.Join(middlewareDir, fmt.Sprintf("%s.go", middlewareType))

	// Check if middleware already exists
	if _, err := os.Stat(middlewarePath); err == nil && !dryRun {
		return fmt.Errorf("middleware %s already exists", middlewarePath)
	}

	if dryRun {
		fmt.Fprintf(c.GetStdout(), "Would create middleware: %s\n", middlewarePath)
		if autoWire {
			fmt.Fprintf(c.GetStdout(), "Would auto-wire middleware into router\n")
		}
		return nil
	}

	// Create directory if it doesn't exist
	if err := os.MkdirAll(middlewareDir, 0755); err != nil {
		return fmt.Errorf("failed to create middleware directory: %w", err)
	}

	// Create middleware generator
	generator := generators.NewMiddlewareGenerator(c.GetStdout(), c.GetStderr())

	// Generate middleware files
	options := generators.MiddlewareOptions{
		Type:      middlewareType,
		AutoWire:  autoWire,
		OutputDir: middlewareDir,
	}

	if err := generator.Generate(options); err != nil {
		return fmt.Errorf("failed to generate middleware files: %w", err)
	}

	return nil
}
