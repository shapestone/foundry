package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/shapestone/foundry/internal/cli/generators"
	"github.com/spf13/cobra"
)

// BuildAddHandlerCommand creates the add handler subcommand
func BuildAddHandlerCommand(c CLI) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "handler [name]",
		Short: "Add a new REST handler",
		Args:  cobra.ExactArgs(1),
		Example: `  foundry add handler user
  foundry add handler product
  foundry add handler order --dry-run
  foundry add handler user --auto-wire`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runAddHandler(c, cmd, args)
		},
	}

	cmd.Flags().Bool("dry-run", false, "Preview changes without applying them")
	cmd.Flags().Bool("auto-wire", false, "Automatically wire the handler into routes")

	return cmd
}

// runAddHandler executes the add handler subcommand
func runAddHandler(c CLI, cmd *cobra.Command, args []string) error {
	name := args[0]

	// Get flags
	dryRun, _ := cmd.Flags().GetBool("dry-run")
	autoWire, _ := cmd.Flags().GetBool("auto-wire")

	// Check if we're in a Go project
	if _, err := os.Stat("go.mod"); os.IsNotExist(err) {
		return fmt.Errorf("go.mod not found. Please run this command from your project root")
	}

	fmt.Fprintf(c.GetStdout(), "🔨 Adding handler: %s\n", name)

	// Create handlers directory
	handlersDir := filepath.Join("internal", "handlers")
	handlerPath := filepath.Join(handlersDir, fmt.Sprintf("%s.go", name))

	// Check if handler already exists
	if _, err := os.Stat(handlerPath); err == nil && !dryRun {
		return fmt.Errorf("handler %s already exists", handlerPath)
	}

	if dryRun {
		fmt.Fprintf(c.GetStdout(), "Would create handler: %s\n", handlerPath)
		if autoWire {
			fmt.Fprintf(c.GetStdout(), "Would auto-wire handler into routes\n")
		}
		return nil
	}

	// Create directory if it doesn't exist
	if err := os.MkdirAll(handlersDir, 0755); err != nil {
		return fmt.Errorf("failed to create handlers directory: %w", err)
	}

	// Create handler generator
	generator := generators.NewHandlerGenerator(c.GetStdout(), c.GetStderr())

	// Generate handler files
	options := generators.HandlerOptions{
		Name:      name,
		AutoWire:  autoWire,
		OutputDir: handlersDir,
	}

	if err := generator.Generate(options); err != nil {
		return fmt.Errorf("failed to generate handler files: %w", err)
	}

	return nil
}
