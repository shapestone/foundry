package cli

import (
	"fmt"
	"github.com/shapestone/foundry/internal/cli/commands"
	"io"
	"os"
	"runtime"

	"github.com/spf13/cobra"
)

// CLI represents the command line interface with injected dependencies
type CLI struct {
	// Configuration
	config *Config

	// I/O interfaces (injectable for testing)
	stdout io.Writer
	stderr io.Writer
	stdin  io.Reader

	// Command tree
	rootCmd *cobra.Command

	// Version information
	version VersionInfo
}

// Config holds all configuration options
type Config struct {
	Verbose    bool   `yaml:"verbose"`
	ConfigFile string `yaml:"config_file"`
	Author     string `yaml:"author"`
	GitHub     string `yaml:"github"`
}

// VersionInfo holds version-related information
type VersionInfo struct {
	Version string
	Commit  string
	Date    string
}

// Option represents a functional option for CLI configuration
type Option func(*CLI)

// New creates a new CLI instance with the given options
func New(opts ...Option) *CLI {
	cli := &CLI{
		config: &Config{},
		stdout: os.Stdout,
		stderr: os.Stderr,
		stdin:  os.Stdin,
		version: VersionInfo{
			Version: "dev",
			Commit:  "none",
			Date:    "unknown",
		},
	}

	// Apply options
	for _, opt := range opts {
		opt(cli)
	}

	// Build command tree with CLI context
	cli.buildCommands()

	return cli
}

// Option functions for dependency injection

// WithConfig sets the configuration
func WithConfig(config *Config) Option {
	return func(c *CLI) { c.config = config }
}

// WithOutput sets the stdout and stderr writers
func WithOutput(stdout, stderr io.Writer) Option {
	return func(c *CLI) {
		c.stdout = stdout
		c.stderr = stderr
	}
}

// WithInput sets the stdin reader
func WithInput(stdin io.Reader) Option {
	return func(c *CLI) { c.stdin = stdin }
}

// WithVersionInfo sets the version information
func WithVersionInfo(version, commit, date string) Option {
	return func(c *CLI) {
		c.version = VersionInfo{
			Version: version,
			Commit:  commit,
			Date:    date,
		}
	}
}

// Execute runs the CLI with the given arguments
func (c *CLI) Execute(args []string) error {
	// Set arguments for this execution (no global state)
	c.rootCmd.SetArgs(args)

	// Execute with isolated context
	return c.rootCmd.Execute()
}

// buildCommands constructs the command tree with CLI context
func (c *CLI) buildCommands() {
	c.rootCmd = &cobra.Command{
		Use:   "foundry",
		Short: "A powerful Go project scaffolding and code generation tool",
		Long: `Foundry is a modern CLI tool for Go developers that helps you:
- Quickly scaffold new Go projects with production-ready layouts
- Generate boilerplate code for common patterns
- Maintain consistent project structure across teams
- Support multiple project layouts (standard, microservice, DDD, etc.)`,
		SilenceErrors:     true,
		SilenceUsage:      true,
		PersistentPreRunE: c.initializeConfig,
	}

	// Set I/O for the command
	c.rootCmd.SetOut(c.stdout)
	c.rootCmd.SetErr(c.stderr)
	c.rootCmd.SetIn(c.stdin)

	// Add persistent flags (bound to config, not globals)
	c.addPersistentFlags()

	// Add subcommands using the new command builders
	adapter := commands.NewCLIAdapter(c)
	c.rootCmd.AddCommand(c.buildVersionCommand())
	c.rootCmd.AddCommand(commands.BuildInitCommand(adapter))
	c.rootCmd.AddCommand(commands.BuildNewCommand(adapter))
	c.rootCmd.AddCommand(commands.BuildAddCommand(adapter))
	c.rootCmd.AddCommand(commands.BuildLayoutCommand(adapter))
	c.rootCmd.AddCommand(commands.BuildWireCommand(adapter))
}

// addPersistentFlags adds flags that are available to all commands
func (c *CLI) addPersistentFlags() {
	flags := c.rootCmd.PersistentFlags()

	// Bind flags to config struct, not global variables
	flags.BoolVarP(&c.config.Verbose, "verbose", "v", false, "Enable verbose output")
	flags.StringVarP(&c.config.ConfigFile, "config", "c", "", "Config file (default: $HOME/.foundry/config.yaml)")
	flags.StringVar(&c.config.Author, "author", "", "Author name for generated code")
	flags.StringVar(&c.config.GitHub, "github", "", "GitHub username")
}

// initializeConfig is called before each command execution
func (c *CLI) initializeConfig(cmd *cobra.Command, args []string) error {
	// Load config file if specified
	if c.config.ConfigFile != "" {
		if err := c.loadConfigFile(c.config.ConfigFile); err != nil {
			return fmt.Errorf("failed to load config file: %w", err)
		}
	}

	// Additional initialization can be added here
	return nil
}

// loadConfigFile loads configuration from a file
func (c *CLI) loadConfigFile(filename string) error {
	// TODO: Implement config file loading
	// For now, just validate the file exists
	if _, err := os.Stat(filename); err != nil {
		return fmt.Errorf("config file not found: %s", filename)
	}
	return nil
}

// buildVersionCommand creates the version command
func (c *CLI) buildVersionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Fprintf(c.stdout, "Foundry %s\n", c.version.Version)
			fmt.Fprintf(c.stdout, "Commit: %s\n", c.version.Commit)
			fmt.Fprintf(c.stdout, "Built: %s\n", c.version.Date)
			fmt.Fprintf(c.stdout, "Go version: %s\n", runtime.Version())
		},
	}
}

// GetConfig returns the current configuration
func (c *CLI) GetConfig() interface{} {
	return c.config
}

// GetStdout returns the stdout writer
func (c *CLI) GetStdout() io.Writer {
	return c.stdout
}

// GetStderr returns the stderr writer
func (c *CLI) GetStderr() io.Writer {
	return c.stderr
}

// GetVersionInfo returns the current version information
func (c *CLI) GetVersionInfo() VersionInfo {
	return c.version
}

// Command builder methods are now implemented in their respective command files:
// - buildInitCommand() -> init_command.go
// - buildNewCommand() -> new_command.go
// - buildAddCommand() -> add_command.go
// - buildLayoutCommand() and subcommands -> layout_command.go
// - buildWireCommand() -> wire_command.go
