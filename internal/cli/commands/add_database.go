package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/shapestone/foundry/internal/cli/generators"
	"github.com/shapestone/foundry/internal/cli/templates"
	"github.com/spf13/cobra"
)

// BuildAddDatabaseCommand creates the add database subcommand
func BuildAddDatabaseCommand(c CLI) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "db [type]",
		Short: "Add database support to your project",
		Args:  cobra.ExactArgs(1),
		Example: `  foundry add db postgres
  foundry add db mysql
  foundry add db sqlite
  foundry add db mongodb`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runAddDatabase(c, cmd, args)
		},
	}

	cmd.Flags().Bool("with-migrations", false, "Include migration setup")
	cmd.Flags().Bool("with-docker", false, "Add docker-compose configuration")

	return cmd
}

// runAddDatabase executes the add database subcommand
func runAddDatabase(c CLI, cmd *cobra.Command, args []string) error {
	dbType := args[0]

	// Get flags
	withMigrations, _ := cmd.Flags().GetBool("with-migrations")
	withDocker, _ := cmd.Flags().GetBool("with-docker")

	// Check if we're in a Go project
	if _, err := os.Stat("go.mod"); os.IsNotExist(err) {
		return fmt.Errorf("go.mod not found. Please run this command from your project root")
	}

	// Validate database type
	if !templates.IsSupportedDatabase(dbType) {
		fmt.Fprintf(c.GetStderr(), "❌ Error: unsupported database type '%s'\n", dbType)
		fmt.Fprintln(c.GetStderr(), "Supported types:")
		for _, db := range templates.GetSupportedDatabases() {
			fmt.Fprintf(c.GetStderr(), "  - %s: %s\n", db.Type, db.Description)
		}
		return fmt.Errorf("unsupported database type: %s", dbType)
	}

	fmt.Fprintf(c.GetStdout(), "🗄️  Adding database: %s\n", dbType)

	// Create database directory
	dbDir := filepath.Join("internal", "database")
	dbPath := filepath.Join(dbDir, "database.go")

	// Check if database config already exists
	if _, err := os.Stat(dbPath); err == nil {
		return fmt.Errorf("database configuration already exists at %s", dbPath)
	}

	// Create directory if it doesn't exist
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		return fmt.Errorf("failed to create database directory: %w", err)
	}

	// Create database generator
	generator := generators.NewDatabaseGenerator(c.GetStdout(), c.GetStderr())

	// Generate database files
	options := generators.DatabaseOptions{
		Type:           dbType,
		WithMigrations: withMigrations,
		WithDocker:     withDocker,
		OutputDir:      dbDir,
	}

	if err := generator.Generate(options); err != nil {
		return fmt.Errorf("failed to generate database files: %w", err)
	}

	return nil
}
