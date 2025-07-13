package cmd

import (
	"github.com/spf13/cobra"
)

var wireCmd = &cobra.Command{
	Use:     "wire [resource]",
	Aliases: []string{"w"},
	Short:   "Wire components into your project",
	Long: `Wire automatically connects components to your application.
This includes updating imports, registering handlers, and applying middleware.`,
	Example: `  foundry wire handler user
  foundry wire middleware auth`,
}

func init() {
	// Add subcommands
	wireCmd.AddCommand(wireHandlerCmd)
	wireCmd.AddCommand(wireMiddlewareCmd)

	// Register with root command
	rootCmd.AddCommand(wireCmd)
}
