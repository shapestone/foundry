package testutil

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// CreateTestProject creates a minimal foundry project structure for testing
func CreateTestProject(t *testing.T, dir, name string) string {
	t.Helper()

	projectPath := filepath.Join(dir, name)

	// Create directory structure
	dirs := []string{
		filepath.Join(projectPath, "cmd", name),
		filepath.Join(projectPath, "internal", "handlers"),
		filepath.Join(projectPath, "internal", "models"),
		filepath.Join(projectPath, "internal", "middleware"),
		filepath.Join(projectPath, "internal", "routes"),
		filepath.Join(projectPath, "internal", "utils"),
	}

	for _, d := range dirs {
		if err := os.MkdirAll(d, 0755); err != nil {
			t.Fatalf("Failed to create directory %s: %v", d, err)
		}
	}

	// Create go.mod
	goModContent := `module ` + name + `

go 1.20

require (
	github.com/gorilla/mux v1.8.0
)
`
	if err := os.WriteFile(filepath.Join(projectPath, "go.mod"), []byte(goModContent), 0644); err != nil {
		t.Fatalf("Failed to create go.mod: %v", err)
	}

	// Create main.go
	mainContent := `package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	fmt.Println("Starting ` + name + ` server...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
`
	mainPath := filepath.Join(projectPath, "cmd", name, "main.go")
	if err := os.WriteFile(mainPath, []byte(mainContent), 0644); err != nil {
		t.Fatalf("Failed to create main.go: %v", err)
	}

	// Create .gitignore
	gitignoreContent := `# Binaries
*.exe
*.dll
*.so
*.dylib

# Test binary
*.test

# Output of go coverage tools
*.out

# Dependency directories
vendor/

# IDE files
.idea/
.vscode/
*.swp
*.swo
*~

# OS files
.DS_Store
Thumbs.db
`
	if err := os.WriteFile(filepath.Join(projectPath, ".gitignore"), []byte(gitignoreContent), 0644); err != nil {
		t.Fatalf("Failed to create .gitignore: %v", err)
	}

	// Create README.md
	readmeContent := `# ` + name + `

A Foundry-generated Go project.

## Getting Started

` + "```bash" + `
go mod download
go run cmd/` + name + `/main.go
` + "```" + `
`
	if err := os.WriteFile(filepath.Join(projectPath, "README.md"), []byte(readmeContent), 0644); err != nil {
		t.Fatalf("Failed to create README.md: %v", err)
	}

	// Create Makefile
	makefileContent := `.PHONY: build run test clean

build:
	go build -o bin/` + name + ` cmd/` + name + `/main.go

run:
	go run cmd/` + name + `/main.go

test:
	go test ./...

clean:
	rm -rf bin/
`
	if err := os.WriteFile(filepath.Join(projectPath, "Makefile"), []byte(makefileContent), 0644); err != nil {
		t.Fatalf("Failed to create Makefile: %v", err)
	}

	return projectPath
}

// AssertFileExists checks if a file exists and fails the test if it doesn't
func AssertFileExists(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Errorf("Expected file to exist: %s", path)
	}
}

// AssertFileNotExists checks if a file doesn't exist and fails the test if it does
func AssertFileNotExists(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); err == nil {
		t.Errorf("Expected file to not exist: %s", path)
	}
}

// AssertFileContains checks if a file contains expected content
func AssertFileContains(t *testing.T, path string, expected ...string) {
	t.Helper()

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("Failed to read file %s: %v", path, err)
	}

	contentStr := string(content)
	for _, exp := range expected {
		if !strings.Contains(contentStr, exp) {
			t.Errorf("File %s doesn't contain expected content: %q", path, exp)
		}
	}
}

// AssertFileNotContains checks if a file doesn't contain unexpected content
func AssertFileNotContains(t *testing.T, path string, unexpected ...string) {
	t.Helper()

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("Failed to read file %s: %v", path, err)
	}

	contentStr := string(content)
	for _, unexp := range unexpected {
		if strings.Contains(contentStr, unexp) {
			t.Errorf("File %s contains unexpected content: %q", path, unexp)
		}
	}
}

// AssertDirExists checks if a directory exists and fails the test if it doesn't
func AssertDirExists(t *testing.T, path string) {
	t.Helper()

	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		t.Errorf("Expected directory to exist: %s", path)
		return
	}

	if err != nil {
		t.Errorf("Error checking directory: %v", err)
		return
	}

	if !info.IsDir() {
		t.Errorf("Expected %s to be a directory, but it's a file", path)
	}
}

// AssertProjectCompiles checks if a project compiles successfully
func AssertProjectCompiles(t *testing.T, projectPath string) {
	t.Helper()

	cmd := exec.Command("go", "build", "./...")
	cmd.Dir = projectPath

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Errorf("Project doesn't compile: %v\nOutput: %s", err, output)
	}
}

// BuildFoundryBinary builds the foundry binary for testing
func BuildFoundryBinary(t *testing.T) string {
	t.Helper()

	binPath := filepath.Join(t.TempDir(), "foundry")
	if runtime.GOOS == "windows" {
		binPath += ".exe"
	}

	cmd := exec.Command("go", "build", "-o", binPath, "../../cmd/foundry")

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to build foundry binary: %v\nOutput: %s", err, output)
	}

	return binPath
}

// RunFoundryCommand runs a foundry command and returns output and error
func RunFoundryCommand(t *testing.T, binPath, workDir string, args ...string) (string, error) {
	t.Helper()

	cmd := exec.Command(binPath, args...)
	cmd.Dir = workDir

	output, err := cmd.CombinedOutput()
	return string(output), err
}

// CreateGoMod creates a go.mod file with the given content
func CreateGoMod(t *testing.T, dir, module string, deps ...string) {
	t.Helper()

	content := "module " + module + "\n\ngo 1.20\n"

	if len(deps) > 0 {
		content += "\nrequire (\n"
		for _, dep := range deps {
			content += "\t" + dep + "\n"
		}
		content += ")\n"
	}

	goModPath := filepath.Join(dir, "go.mod")
	if err := os.WriteFile(goModPath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create go.mod: %v", err)
	}
}

// CleanupProject removes test artifacts and restores working directory
func CleanupProject(t *testing.T, originalWd string) {
	t.Helper()

	if err := os.Chdir(originalWd); err != nil {
		t.Errorf("Failed to restore working directory: %v", err)
	}
}
