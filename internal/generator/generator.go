package generator

import (
	"fmt"
	"os"
	"path/filepath"
	"text/template"
)

// Generator handles file generation from templates
type Generator interface {
	Generate(path string, tmplContent string, data interface{}) error
}

// FileGenerator implements the Generator interface for file system operations
type FileGenerator struct{}

// NewFileGenerator creates a new file generator
func NewFileGenerator() *FileGenerator {
	return &FileGenerator{}
}

// Generate creates a file from a template
func (g *FileGenerator) Generate(path string, tmplContent string, data interface{}) error {
	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("creating directory %s: %w", dir, err)
	}

	// Parse template
	tmpl, err := template.New(filepath.Base(path)).Parse(tmplContent)
	if err != nil {
		return fmt.Errorf("parsing template: %w", err)
	}

	// Create file
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("creating file %s: %w", path, err)
	}
	defer file.Close()

	// Execute template
	if err := tmpl.Execute(file, data); err != nil {
		return fmt.Errorf("executing template: %w", err)
	}

	return nil
}
