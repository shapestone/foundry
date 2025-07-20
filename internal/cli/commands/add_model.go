package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/shapestone/foundry/internal/cli/generators"
	"github.com/spf13/cobra"
)

// BuildAddModelCommand creates the add model subcommand
func BuildAddModelCommand(c CLI) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "model [name]",
		Short: "Add a new data model",
		Args:  cobra.ExactArgs(1),
		Example: `  foundry add model user
  foundry add model product
  foundry add model order`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runAddModel(c, cmd, args)
		},
	}

	return cmd
}

// runAddModel executes the add model subcommand
func runAddModel(c CLI, cmd *cobra.Command, args []string) error {
	name := args[0]

	// Check if we're in a Go project
	if _, err := os.Stat("go.mod"); os.IsNotExist(err) {
		return fmt.Errorf("go.mod not found. Please run this command from your project root")
	}

	fmt.Fprintf(c.GetStdout(), "🔨 Adding model: %s\n", name)

	// Create models directory
	modelsDir := filepath.Join("internal", "models")
	modelPath := filepath.Join(modelsDir, fmt.Sprintf("%s.go", name))

	// Check if model already exists
	if _, err := os.Stat(modelPath); err == nil {
		return fmt.Errorf("model %s already exists", modelPath)
	}

	// Create directory if it doesn't exist
	if err := os.MkdirAll(modelsDir, 0755); err != nil {
		return fmt.Errorf("failed to create models directory: %w", err)
	}

	// Create model generator
	generator := generators.NewModelGenerator(c.GetStdout(), c.GetStderr())

	// Generate model files
	options := generators.ModelOptions{
		Name:      name,
		OutputDir: modelsDir,
	}

	if err := generator.Generate(options); err != nil {
		return fmt.Errorf("failed to generate model files: %w", err)
	}

	return nil
}
