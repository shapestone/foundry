package project

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// ModuleReader reads module information from go.mod
type ModuleReader interface {
	GetModuleName() (string, error)
}

// GoModReader implements ModuleReader for go.mod files
type GoModReader struct{}

// NewGoModReader creates a new go.mod reader
func NewGoModReader() *GoModReader {
	return &GoModReader{}
}

// GetModuleName reads the module name from go.mod
func (r *GoModReader) GetModuleName() (string, error) {
	content, err := os.ReadFile("go.mod")
	if err != nil {
		return "", fmt.Errorf("reading go.mod: %w", err)
	}

	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "module ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "module ")), nil
		}
	}

	return "", fmt.Errorf("module declaration not found in go.mod")
}

// GetCurrentModule reads the module name from go.mod file using bufio for better compatibility
func GetCurrentModule() string {
	// Try to read go.mod file
	goModPath := "go.mod"
	file, err := os.Open(goModPath)
	if err != nil {
		return "myapp"
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "module ") {
			// Extract module name
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				return parts[1]
			}
		}
	}

	return "myapp"
}
