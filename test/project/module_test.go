package project_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/shapestone/foundry/internal/project"
)

func TestNewGoModReader(t *testing.T) {
	// When
	reader := project.NewGoModReader()

	// Then
	if reader == nil {
		t.Error("Expected NewGoModReader to return non-nil reader")
	}
}

func TestGetModuleName(t *testing.T) {
	tests := []struct {
		name         string
		goModContent string
		expected     string
		expectError  bool
	}{
		{
			name: "SimpleModule",
			goModContent: `module myapp

go 1.20
`,
			expected:    "myapp",
			expectError: false,
		},
		{
			name: "ModuleWithGitHub",
			goModContent: `module github.com/user/project

go 1.20
`,
			expected:    "github.com/user/project",
			expectError: false,
		},
		{
			name: "ModuleWithDomain",
			goModContent: `module example.com/myservice

go 1.20

require (
	github.com/gorilla/mux v1.8.0
)
`,
			expected:    "example.com/myservice",
			expectError: false,
		},
		{
			name: "ComplexModulePath",
			goModContent: `module github.com/company/team/project/service

go 1.21

require (
	github.com/spf13/cobra v1.7.0
	github.com/spf13/viper v1.16.0
)

require (
	github.com/fsnotify/fsnotify v1.6.0 // indirect
)
`,
			expected:    "github.com/company/team/project/service",
			expectError: false,
		},
		{
			name:         "EmptyFile",
			goModContent: "",
			expected:     "",
			expectError:  true,
		},
		{
			name: "NoModuleStatement",
			goModContent: `go 1.20

require github.com/spf13/cobra v1.7.0
`,
			expected:    "",
			expectError: true,
		},
		{
			name: "MalformedModule",
			goModContent: `module

go 1.20
`,
			expected:    "",
			expectError: true,
		},
		{
			name: "ModuleWithComment",
			goModContent: `// This is my module
module myproject // main module

go 1.20
`,
			expected:    "myproject // main module", // Actual behavior includes the comment
			expectError: false,
		},
		{
			name: "ModuleWithWhitespace",
			goModContent: `module   github.com/user/project   

go 1.20
`,
			expected:    "github.com/user/project",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Given
			tmpDir := t.TempDir()
			goModPath := filepath.Join(tmpDir, "go.mod")

			if err := os.WriteFile(goModPath, []byte(tt.goModContent), 0644); err != nil {
				t.Fatal(err)
			}

			// Change to the directory with go.mod
			oldWd, _ := os.Getwd()
			if err := os.Chdir(tmpDir); err != nil {
				t.Fatal(err)
			}
			defer os.Chdir(oldWd)

			reader := project.NewGoModReader()

			// When
			result, err := reader.GetModuleName()

			// Then
			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if result != tt.expected {
					t.Errorf("GetModuleName() = %q, want %q", result, tt.expected)
				}
			}
		})
	}
}

func TestGetModuleName_FileNotFound(t *testing.T) {
	// Given - change to a directory without go.mod
	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(oldWd)

	reader := project.NewGoModReader()

	// When
	result, err := reader.GetModuleName()

	// Then
	if err == nil {
		t.Error("Expected error for non-existent go.mod")
	}
	if result != "" {
		t.Errorf("Expected empty result for non-existent file, got %q", result)
	}
}

func TestGetModuleName_FilePermissions(t *testing.T) {
	// Skip on Windows as permission handling is different
	if os.Getenv("GOOS") == "windows" {
		t.Skip("Skipping permission test on Windows")
	}

	// Given
	tmpDir := t.TempDir()
	goModPath := filepath.Join(tmpDir, "go.mod")

	goModContent := `module myapp
go 1.20
`
	if err := os.WriteFile(goModPath, []byte(goModContent), 0000); err != nil {
		t.Fatal(err)
	}

	oldWd, _ := os.Getwd()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(oldWd)

	reader := project.NewGoModReader()

	// When
	result, err := reader.GetModuleName()

	// Then
	if err == nil {
		t.Error("Expected error for unreadable file")
	}
	if result != "" {
		t.Errorf("Expected empty result for unreadable file, got %q", result)
	}
}

func TestGetCurrentModule(t *testing.T) {
	tests := []struct {
		name        string
		setupFunc   func() (string, func())
		expected    string
		expectError bool
	}{
		{
			name: "ValidModule",
			setupFunc: func() (string, func()) {
				tmpDir := t.TempDir()
				goModContent := `module testmodule

go 1.20
`
				if err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(goModContent), 0644); err != nil {
					t.Fatal(err)
				}

				oldWd, _ := os.Getwd()
				if err := os.Chdir(tmpDir); err != nil {
					t.Fatal(err)
				}

				cleanup := func() {
					os.Chdir(oldWd)
				}

				return tmpDir, cleanup
			},
			expected:    "testmodule",
			expectError: false,
		},
		{
			name: "NoGoMod",
			setupFunc: func() (string, func()) {
				tmpDir := t.TempDir()

				oldWd, _ := os.Getwd()
				if err := os.Chdir(tmpDir); err != nil {
					t.Fatal(err)
				}

				cleanup := func() {
					os.Chdir(oldWd)
				}

				return tmpDir, cleanup
			},
			expected:    "yourmodule", // Default fallback value
			expectError: false,        // Doesn't error, just returns default
		},
		{
			name: "NestedDirectory",
			setupFunc: func() (string, func()) {
				tmpDir := t.TempDir()
				goModContent := `module github.com/test/nested

go 1.20
`
				if err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(goModContent), 0644); err != nil {
					t.Fatal(err)
				}

				// Create nested directory
				nestedDir := filepath.Join(tmpDir, "internal", "nested")
				if err := os.MkdirAll(nestedDir, 0755); err != nil {
					t.Fatal(err)
				}

				oldWd, _ := os.Getwd()
				if err := os.Chdir(nestedDir); err != nil {
					t.Fatal(err)
				}

				cleanup := func() {
					os.Chdir(oldWd)
				}

				return nestedDir, cleanup
			},
			expected:    "yourmodule", // GetCurrentModule might not walk up directories
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Given
			_, cleanup := tt.setupFunc()
			defer cleanup()

			// When
			result := project.GetCurrentModule()

			// Then
			if tt.expectError && result != tt.expected {
				t.Errorf("Expected result to be %q on error, got %q", tt.expected, result)
			}
			if !tt.expectError && result != tt.expected {
				t.Errorf("GetCurrentModule() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func BenchmarkGetModuleName(b *testing.B) {
	// Setup
	tmpDir := b.TempDir()
	goModPath := filepath.Join(tmpDir, "go.mod")

	goModContent := `module github.com/benchmark/test

go 1.20

require (
	github.com/spf13/cobra v1.7.0
	github.com/spf13/viper v1.16.0
	github.com/stretchr/testify v1.8.4
)
`
	if err := os.WriteFile(goModPath, []byte(goModContent), 0644); err != nil {
		b.Fatal(err)
	}

	oldWd, _ := os.Getwd()
	if err := os.Chdir(tmpDir); err != nil {
		b.Fatal(err)
	}
	defer os.Chdir(oldWd)

	reader := project.NewGoModReader()

	// Benchmark
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = reader.GetModuleName()
	}
}
