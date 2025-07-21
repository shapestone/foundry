// test/integration/integration_test.go
package integration

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestHelper provides utilities for integration testing
type TestHelper struct {
	t           *testing.T
	tempDir     string
	foundryPath string
}

// NewTestHelper creates a new test helper
func NewTestHelper(t *testing.T) *TestHelper {
	tempDir := t.TempDir()
	foundryPath := buildFoundryBinary(t)

	return &TestHelper{
		t:           t,
		tempDir:     tempDir,
		foundryPath: foundryPath,
	}
}

// buildFoundryBinary builds the foundry binary for testing
func buildFoundryBinary(t *testing.T) string {
	t.Helper()

	// Get project root (go up from test/integration)
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}

	projectRoot := filepath.Join(wd, "..", "..")
	binaryPath := filepath.Join(projectRoot, "foundry-test")

	// Build the binary
	cmd := exec.Command("go", "build", "-o", binaryPath, "./cmd/foundry")
	cmd.Dir = projectRoot
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to build foundry binary: %v\nOutput: %s", err, output)
	}

	return binaryPath
}

// RunFoundry executes foundry command with given arguments
func (h *TestHelper) RunFoundry(args ...string) (string, error) {
	h.t.Helper()

	cmd := exec.Command(h.foundryPath, args...)
	cmd.Dir = h.tempDir

	output, err := cmd.CombinedOutput()
	return string(output), err
}

// RunFoundryInDir executes foundry command in specific directory
func (h *TestHelper) RunFoundryInDir(dir string, args ...string) (string, error) {
	h.t.Helper()

	cmd := exec.Command(h.foundryPath, args...)
	cmd.Dir = dir

	output, err := cmd.CombinedOutput()
	return string(output), err
}

// AssertFileExists checks if a file exists
func (h *TestHelper) AssertFileExists(path string) {
	h.t.Helper()

	fullPath := filepath.Join(h.tempDir, path)
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		h.t.Fatalf("Expected file to exist: %s", path)
	}
}

// AssertFileContains checks if a file contains expected content
func (h *TestHelper) AssertFileContains(path, content string) {
	h.t.Helper()

	fullPath := filepath.Join(h.tempDir, path)
	data, err := os.ReadFile(fullPath)
	if err != nil {
		h.t.Fatalf("Failed to read file %s: %v", path, err)
	}

	if !strings.Contains(string(data), content) {
		h.t.Fatalf("File %s does not contain expected content: %s\nActual content:\n%s",
			path, content, string(data))
	}
}

// AssertFileNotExists checks if a file does not exist
func (h *TestHelper) AssertFileNotExists(path string) {
	h.t.Helper()

	fullPath := filepath.Join(h.tempDir, path)
	if _, err := os.Stat(fullPath); err == nil {
		h.t.Fatalf("Expected file to not exist: %s", path)
	}
}

// AssertOutputContains checks if command output contains expected text
func (h *TestHelper) AssertOutputContains(output, expected string) {
	h.t.Helper()

	if !strings.Contains(output, expected) {
		h.t.Fatalf("Output does not contain expected text: %s\nActual output:\n%s",
			expected, output)
	}
}

// AssertError checks if error occurred and contains expected message
func (h *TestHelper) AssertError(err error, expectedMsg string) {
	h.t.Helper()

	if err == nil {
		h.t.Fatalf("Expected error containing '%s', but got no error", expectedMsg)
	}

	if !strings.Contains(err.Error(), expectedMsg) {
		h.t.Fatalf("Expected error to contain '%s', but got: %v", expectedMsg, err)
	}
}

// AssertNoError checks that no error occurred
func (h *TestHelper) AssertNoError(err error) {
	h.t.Helper()

	if err != nil {
		h.t.Fatalf("Expected no error, but got: %v", err)
	}
}

// ChangeToTempDir changes to the test temp directory
func (h *TestHelper) ChangeToTempDir() func() {
	h.t.Helper()

	originalDir, err := os.Getwd()
	if err != nil {
		h.t.Fatalf("Failed to get working directory: %v", err)
	}

	if err := os.Chdir(h.tempDir); err != nil {
		h.t.Fatalf("Failed to change to temp directory: %v", err)
	}

	// Return cleanup function
	return func() {
		if err := os.Chdir(originalDir); err != nil {
			h.t.Fatalf("Failed to restore working directory: %v", err)
		}
	}
}

// CreateFile creates a file with given content in temp directory
func (h *TestHelper) CreateFile(path, content string) {
	h.t.Helper()

	fullPath := filepath.Join(h.tempDir, path)
	dir := filepath.Dir(fullPath)

	if err := os.MkdirAll(dir, 0755); err != nil {
		h.t.Fatalf("Failed to create directory for %s: %v", path, err)
	}

	if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
		h.t.Fatalf("Failed to create file %s: %v", path, err)
	}
}

// GetTempDir returns the temp directory path
func (h *TestHelper) GetTempDir() string {
	return h.tempDir
}

// Cleanup is called automatically by t.TempDir(), but can be called manually
func (h *TestHelper) Cleanup() {
	// Remove test binary
	if h.foundryPath != "" {
		os.Remove(h.foundryPath)
	}
}
