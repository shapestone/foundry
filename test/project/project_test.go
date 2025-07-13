package project_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/shapestone/foundry/internal/project"
)

// TestGetProjectName_BasicFunctionality tests the basic behavior
func TestGetProjectName_BasicFunctionality(t *testing.T) {
	// Given - we're in a directory
	tmpDir := t.TempDir()
	testDir := filepath.Join(tmpDir, "test-project-name")
	if err := os.MkdirAll(testDir, 0755); err != nil {
		t.Fatal(err)
	}

	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)

	if err := os.Chdir(testDir); err != nil {
		t.Fatal(err)
	}

	// When
	result := project.GetProjectName()

	// Then - based on the current implementation in project.go
	// it should return the directory name
	expected := "test-project-name"
	if result != expected {
		t.Errorf("GetProjectName() = %q, want %q", result, expected)
	}
}

// TestGetProjectName_Fallback tests the fallback behavior
func TestGetProjectName_Fallback(t *testing.T) {
	// The function should never return empty
	result := project.GetProjectName()

	if result == "" {
		t.Error("GetProjectName() should never return empty string")
	}

	// It should return something reasonable
	if result == "." || result == "/" || result == "" {
		t.Errorf("GetProjectName() returned invalid name: %q", result)
	}
}
