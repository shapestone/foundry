package project

import (
	"os"
	"path/filepath"
	"strings"
)

// GetProjectName returns the project name from the current directory or module
func GetProjectName() string {
	// First try to get from current directory name
	cwd, err := os.Getwd()
	if err == nil {
		projectName := filepath.Base(cwd)
		if projectName != "" && projectName != "." && projectName != "/" {
			return projectName
		}
	}

	// Fall back to module name
	module := GetCurrentModule()
	parts := strings.Split(module, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}

	return "myapp"
}
