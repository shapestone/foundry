// File: internal/cli/legacy.go
// This file provides backward compatibility for existing tests
// It should be removed once all tests are migrated to the new architecture

package cli

// Legacy functions for backward compatibility with existing tests
// These work directly with the global variables in root.go

// Execute runs the legacy root command (backward compatibility)
func Execute() error {
	return rootCmd.Execute()
}

// SetVersionInfo sets version info using legacy globals
func SetVersionInfo(v, c, d string) {
	version = v
	commit = c
	date = d
}

// GetVersionInfo returns version info from legacy globals
func GetVersionInfo() (string, string, string) {
	return version, commit, date
}

// GetVerbose returns the verbose flag from legacy globals
func GetVerbose() bool {
	return verbose
}

// SetVerbose sets the verbose flag on legacy globals
func SetVerbose(v bool) {
	verbose = v
}

// GetConfigFile returns the config file path from legacy globals
func GetConfigFile() string {
	return configFile
}

// SetConfigFile sets the config file path on legacy globals
func SetConfigFile(path string) {
	configFile = path
}

// ResetGlobalFlags resets legacy global flags
func ResetGlobalFlags() {
	verbose = false
	configFile = ""
}
