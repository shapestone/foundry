package commands

import (
	"io"
)

// CLIAdapter adapts the main CLI struct to the commands interface
type CLIAdapter struct {
	stdout io.Writer
	stderr io.Writer
	config *Config
}

// NewCLIAdapter creates a new CLI adapter
func NewCLIAdapter(cli interface {
	GetStdout() io.Writer
	GetStderr() io.Writer
	GetConfig() interface{}
}) *CLIAdapter {
	// Convert config interface{} to our Config type
	var config *Config
	if rawConfig := cli.GetConfig(); rawConfig != nil {
		// Handle different config types
		switch cfg := rawConfig.(type) {
		case *Config:
			config = cfg
		default:
			// Try to extract fields via reflection or create default
			config = &Config{
				Verbose:    false,
				ConfigFile: "",
				Author:     "",
				GitHub:     "",
			}
		}
	} else {
		// Create default config
		config = &Config{
			Verbose:    false,
			ConfigFile: "",
			Author:     "",
			GitHub:     "",
		}
	}

	return &CLIAdapter{
		stdout: cli.GetStdout(),
		stderr: cli.GetStderr(),
		config: config,
	}
}

// GetStdout returns stdout writer
func (a *CLIAdapter) GetStdout() io.Writer {
	return a.stdout
}

// GetStderr returns stderr writer
func (a *CLIAdapter) GetStderr() io.Writer {
	return a.stderr
}

// GetConfig returns configuration
func (a *CLIAdapter) GetConfig() *Config {
	return a.config
}
