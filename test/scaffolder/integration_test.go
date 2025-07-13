// test/scaffolder/integration_test.go
package scaffolder_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/shapestone/foundry/internal/scaffolder"
)

// Integration tests that test the scaffolder with real dependencies
// These tests create temporary directories and files

func TestScaffolder_Integration_CreateHandler_WithRealFileSystem(t *testing.T) {
	// Skip integration tests in short mode
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Create temporary directory for test project
	tempDir, err := os.MkdirTemp("", "foundry-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a minimal go.mod file to make it a valid Go project
	goModPath := filepath.Join(tempDir, "go.mod")
	goModContent := `module github.com/test/project

go 1.21
`
	if err := os.WriteFile(goModPath, []byte(goModContent), 0644); err != nil {
		t.Fatalf("Failed to create go.mod: %v", err)
	}

	// Create the scaffolder with real adapters
	scaffolderInstance := scaffolder.NewScaffolderWithAdapters()

	spec := &scaffolder.HandlerSpec{
		Name:        "user",
		Type:        "REST",
		AutoWire:    false,
		DryRun:      false,
		ProjectRoot: tempDir,
	}

	// Execute the scaffolding
	result, err := scaffolderInstance.CreateHandler(context.Background(), spec)

	// Verify results
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if !result.Success {
		t.Fatalf("Expected success, got failure: %s", result.Message)
	}

	// Verify file was created
	expectedPath := filepath.Join(tempDir, "internal", "handlers", "user.go")
	if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
		t.Fatalf("Handler file was not created: %s", expectedPath)
	}

	// Verify file content
	content, err := os.ReadFile(expectedPath)
	if err != nil {
		t.Fatalf("Failed to read created file: %v", err)
	}

	contentStr := string(content)
	if len(contentStr) == 0 {
		t.Fatalf("Created file is empty")
	}

	// The actual content will depend on your template implementation
	t.Logf("Created file content: %s", contentStr)
}

func TestScaffolder_Integration_CreateHandler_DryRun(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Create temporary directory for test project
	tempDir, err := os.MkdirTemp("", "foundry-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a minimal go.mod file
	goModPath := filepath.Join(tempDir, "go.mod")
	goModContent := `module github.com/test/project

go 1.21
`
	if err := os.WriteFile(goModPath, []byte(goModContent), 0644); err != nil {
		t.Fatalf("Failed to create go.mod: %v", err)
	}

	// Create the scaffolder with real adapters
	scaffolderInstance := scaffolder.NewScaffolderWithAdapters()

	spec := &scaffolder.HandlerSpec{
		Name:        "user",
		Type:        "REST",
		AutoWire:    false,
		DryRun:      true, // Dry run - should not create files
		ProjectRoot: tempDir,
	}

	// Execute the scaffolding
	result, err := scaffolderInstance.CreateHandler(context.Background(), spec)

	// Verify results
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if !result.Success {
		t.Fatalf("Expected success, got failure: %s", result.Message)
	}

	// Verify no file was created
	expectedPath := filepath.Join(tempDir, "internal", "handlers", "user.go")
	if _, err := os.Stat(expectedPath); !os.IsNotExist(err) {
		t.Fatalf("Handler file should not be created in dry run mode: %s", expectedPath)
	}

	// Verify changes were reported
	if len(result.Changes) == 0 {
		t.Fatalf("Expected changes to be reported in dry run")
	}
}

func TestScaffolder_Integration_CreateHandler_InvalidProject(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Create temporary directory without go.mod (invalid Go project)
	tempDir, err := os.MkdirTemp("", "foundry-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create the scaffolder with real adapters
	scaffolderInstance := scaffolder.NewScaffolderWithAdapters()

	spec := &scaffolder.HandlerSpec{
		Name:        "user",
		Type:        "REST",
		ProjectRoot: tempDir,
	}

	// Execute the scaffolding
	_, err = scaffolderInstance.CreateHandler(context.Background(), spec)

	// Should fail because it's not a Go project
	if err == nil {
		t.Fatalf("Expected error for invalid Go project")
	}

	if !contains(err.Error(), "not a Go project") {
		t.Fatalf("Expected 'not a Go project' error, got: %s", err.Error())
	}
}

func TestScaffolder_Integration_CreateHandler_HandlerAlreadyExists(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Create temporary directory for test project
	tempDir, err := os.MkdirTemp("", "foundry-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a minimal go.mod file
	goModPath := filepath.Join(tempDir, "go.mod")
	goModContent := `module github.com/test/project

go 1.21
`
	if err := os.WriteFile(goModPath, []byte(goModContent), 0644); err != nil {
		t.Fatalf("Failed to create go.mod: %v", err)
	}

	// Create the handler directory and file (to simulate existing handler)
	handlerDir := filepath.Join(tempDir, "internal", "handlers")
	if err := os.MkdirAll(handlerDir, 0755); err != nil {
		t.Fatalf("Failed to create handler directory: %v", err)
	}

	existingHandlerPath := filepath.Join(handlerDir, "user.go")
	existingContent := "package handlers\n\n// Existing handler"
	if err := os.WriteFile(existingHandlerPath, []byte(existingContent), 0644); err != nil {
		t.Fatalf("Failed to create existing handler: %v", err)
	}

	// Create the scaffolder with real adapters
	scaffolderInstance := scaffolder.NewScaffolderWithAdapters()

	spec := &scaffolder.HandlerSpec{
		Name:        "user",
		Type:        "REST",
		ProjectRoot: tempDir,
	}

	// Execute the scaffolding
	_, err = scaffolderInstance.CreateHandler(context.Background(), spec)

	// Should fail because handler already exists
	if err == nil {
		t.Fatalf("Expected error for existing handler")
	}

	if !contains(err.Error(), "already exists") {
		t.Fatalf("Expected 'already exists' error, got: %s", err.Error())
	}
}

// Helper function for string contains check
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(substr) > 0 && someContains(s, substr)))
}

func someContains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
