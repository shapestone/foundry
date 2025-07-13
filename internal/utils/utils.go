package utils

import (
	"fmt"
	"github.com/shapestone/foundry/internal/generator"
	"github.com/shapestone/foundry/internal/project"
)

// GenerateFile creates a file from a template
// Deprecated: Use generator.FileGenerator instead
func GenerateFile(tmplContent, path string, data interface{}) error {
	// Add validation to prevent empty path issues
	if path == "" {
		return fmt.Errorf("file path cannot be empty")
	}

	gen := generator.NewFileGenerator()
	return gen.Generate(path, tmplContent, data)
}

// GetCurrentModule reads the module name from go.mod
// Deprecated: Use project.GetCurrentModule instead
func GetCurrentModule() string {
	return project.GetCurrentModule()
}
