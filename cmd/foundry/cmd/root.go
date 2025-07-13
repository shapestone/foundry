package cmd

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"

	// Global flags
	verbose    bool
	configFile string
)

// rootCmd represents the base command
var rootCmd = &cobra.Command{
	Use:   "foundry",
	Short: "A powerful Go project scaffolding and code generation tool",
	Long: `Foundry is a modern CLI tool for Go developers that helps you:
- Quickly scaffold new Go projects with production-ready layouts
- Generate boilerplate code for common patterns
- Maintain consistent project structure across teams
- Support multiple project layouts (standard, microservice, DDD, etc.)`,
	SilenceErrors: true,
	SilenceUsage:  true,
}

// Execute runs the root command
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")
	rootCmd.PersistentFlags().StringVarP(&configFile, "config", "c", "", "Config file (default: $HOME/.foundry/config.yaml)")

	// Add version command
	rootCmd.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("Foundry %s\n", version)
			fmt.Printf("Commit: %s\n", commit)
			fmt.Printf("Built: %s\n", date)
			fmt.Printf("Go version: %s\n", runtime.Version())
		},
	})
}

// SetVersionInfo sets the version information
func SetVersionInfo(v, c, d string) {
	version = v
	commit = c
	date = d
}
