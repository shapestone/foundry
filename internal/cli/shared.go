package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/shapestone/foundry/internal/layout"
)

// Shared utility functions used across multiple commands

// createLayoutManager creates a layout manager instance
func (c *CLI) createLayoutManager() (*layout.Manager, error) {
	// Determine config path
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	configPath := filepath.Join(homeDir, ".foundry", "layouts.yaml")
	return layout.NewManager(configPath)
}

// Project and component name validation

// isValidProjectName checks if a project name is valid
func isValidProjectName(name string) bool {
	if name == "" {
		return false
	}

	for _, r := range name {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') ||
			(r >= '0' && r <= '9') || r == '-' || r == '_') {
			return false
		}
	}

	// Can't start with a number
	if name[0] >= '0' && name[0] <= '9' {
		return false
	}

	return true
}

// validateComponentName validates component names
func validateComponentName(name string) error {
	if name == "" {
		return fmt.Errorf("component name cannot be empty")
	}

	if !isValidComponentName(name) {
		return fmt.Errorf("invalid characters (use only letters, numbers, hyphens, and underscores)")
	}

	return nil
}

// isValidComponentName checks if a component name is valid
func isValidComponentName(name string) bool {
	if name == "" {
		return false
	}

	for _, r := range name {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') ||
			(r >= '0' && r <= '9') || r == '-' || r == '_') {
			return false
		}
	}

	// Can't start with a number
	if name[0] >= '0' && name[0] <= '9' {
		return false
	}

	return true
}

// File system utilities

// isDirEmpty checks if a directory is empty (ignoring .git and .DS_Store)
func isDirEmpty(dir string) (bool, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false, err
	}

	for _, entry := range entries {
		// Ignore .git directory
		if entry.Name() == ".git" {
			continue
		}
		// Ignore .DS_Store on macOS
		if entry.Name() == ".DS_Store" {
			continue
		}
		return false, nil
	}

	return true, nil
}

// isGitRepo checks if a directory is a git repository
func isGitRepo(dir string) bool {
	gitDir := filepath.Join(dir, ".git")
	info, err := os.Stat(gitDir)
	return err == nil && info.IsDir()
}

// Git operations

// initGitRepo initializes a git repository
func initGitRepo(projectPath string) error {
	// Simple git init implementation
	// In a real implementation, this would use go-git or exec git commands
	gitDir := filepath.Join(projectPath, ".git")
	return os.MkdirAll(gitDir, 0755)
}

// createInitialCommit creates an initial git commit
func createInitialCommit() error {
	// In a real implementation, this would use go-git or exec git commands
	// For now, just return nil
	return nil
}

// Project configuration

// saveProjectConfig saves project configuration
func saveProjectConfig(layoutName string) error {
	// Save the layout name in foundry.yaml so we know which layout was used
	// This is already done by the layout system if foundry.yaml.tmpl exists
	return nil
}

// String manipulation utilities

// toTitle converts string to title case
func toTitle(s string) string {
	if s == "" {
		return s
	}

	// Simple title case - first letter uppercase, rest lowercase
	first := strings.ToUpper(string(s[0]))
	rest := strings.ToLower(s[1:])
	return first + rest
}
