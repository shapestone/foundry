package commands

import (
	"io"

	"github.com/spf13/cobra"
)

// BuildAddCommand creates the main add command with subcommands
func BuildAddCommand(c CLI) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add [component-type] [name]",
		Short: "Add a new component to the current project",
		Long: `Add a new component to the current project using the project's layout templates.
	
The add command generates boilerplate code for common components like handlers, models,
middleware, services, and more. The available component types depend on the project layout.`,
		Example: `  foundry add handler users
  foundry add model product
  foundry add middleware auth
  foundry add service payment
  foundry add repository user`,
	}

	// Component configuration flags
	cmd.Flags().StringP("output", "o", "", "Custom output directory")
	cmd.Flags().BoolP("force", "f", false, "Overwrite existing files")
	cmd.Flags().Bool("dry-run", false, "Show what would be generated without creating files")

	// Add subcommands for specific component types
	cmd.AddCommand(BuildAddDatabaseCommand(c))
	cmd.AddCommand(BuildAddMiddlewareCommand(c))
	cmd.AddCommand(BuildAddHandlerCommand(c))
	cmd.AddCommand(BuildAddModelCommand(c))

	return cmd
}

// CLI interface - defines what we need from the main CLI
type CLI interface {
	GetStdout() io.Writer
	GetStderr() io.Writer
	GetConfig() *Config
}

type Config struct {
	Verbose    bool
	ConfigFile string
	Author     string
	GitHub     string
}
