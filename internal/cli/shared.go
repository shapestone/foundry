// File: internal/cli/shared.go
// FIXED: Changed package from 'generators' to 'cli'

package cli

import (
	"os"
	"path/filepath"
	"strings"
)

// writeFile writes content to a file
func writeFile(path, content string) error {
	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	return os.WriteFile(path, []byte(content), 0644)
}

// getCurrentModule gets the current module name from go.mod
func getCurrentModule() string {
	data, err := os.ReadFile("go.mod")
	if err != nil {
		return "myapp"
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "module ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "module"))
		}
	}
	return "myapp"
}

// getProjectName gets the current project name
func getProjectName() string {
	if cwd, err := os.Getwd(); err == nil {
		return filepath.Base(cwd)
	}
	return "myapp"
}

// detectProjectLayout detects the current project layout from foundry.yaml
func detectProjectLayout() (string, error) {
	// Check for foundry.yaml
	if data, err := os.ReadFile("foundry.yaml"); err == nil {
		// Simple parsing - look for "layout: layoutname"
		lines := strings.Split(string(data), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "layout:") {
				parts := strings.Split(line, ":")
				if len(parts) >= 2 {
					return strings.TrimSpace(parts[1]), nil
				}
			}
		}
	}

	// Check for .foundry.yaml
	if data, err := os.ReadFile(".foundry.yaml"); err == nil {
		// Simple parsing - look for "layout: layoutname"
		lines := strings.Split(string(data), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "layout:") {
				parts := strings.Split(line, ":")
				if len(parts) >= 2 {
					return strings.TrimSpace(parts[1]), nil
				}
			}
		}
	}

	// Default to standard layout
	return "standard", nil
}

// toSnakeCase converts to snake_case
func toSnakeCase(s string) string {
	var result []rune
	for i, r := range s {
		if r >= 'A' && r <= 'Z' && i > 0 {
			// Add underscore before uppercase letters (except first)
			if i > 0 && s[i-1] >= 'a' && s[i-1] <= 'z' {
				result = append(result, '_')
			}
		}
		result = append(result, toLowerRune(r))
	}
	return string(result)
}

// toLowerRune converts rune to lowercase
func toLowerRune(r rune) rune {
	if r >= 'A' && r <= 'Z' {
		return r + ('a' - 'A')
	}
	return r
}

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

// capitalize capitalizes the first letter of a string
func capitalize(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + strings.ToLower(s[1:])
}
