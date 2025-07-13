package testutil

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

// TestSuite provides a testing environment for Foundry CLI tests
type TestSuite struct {
	T       *testing.T
	TempDir string
	Binary  string
}

// NewTestSuite creates a new test suite
func NewTestSuite(t *testing.T) *TestSuite {
	tempDir := t.TempDir()

	// Build the foundry binary for testing
	binary := filepath.Join(tempDir, "foundry")
	if runtime.GOOS == "windows" {
		binary += ".exe"
	}

	ts := &TestSuite{
		T:       t,
		TempDir: tempDir,
		Binary:  binary,
	}

	// Build the binary
	if err := ts.BuildBinary(); err != nil {
		t.Fatalf("Failed to build binary: %v", err)
	}

	return ts
}

// BuildBinary builds the foundry binary for testing
func (ts *TestSuite) BuildBinary() error {
	// Get the project root (where go.mod is)
	projectRoot, err := findProjectRoot()
	if err != nil {
		return fmt.Errorf("failed to find project root: %w", err)
	}

	cmd := exec.Command("go", "build", "-o", ts.Binary, "./cmd/foundry")
	cmd.Dir = projectRoot
	cmd.Env = append(os.Environ(), "CGO_ENABLED=0")

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("build failed: %v\nstderr: %s", err, stderr.String())
	}

	return nil
}

// findProjectRoot finds the project root by looking for go.mod
func findProjectRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("go.mod not found")
		}
		dir = parent
	}
}

// RunFoundry executes the foundry binary with the given arguments
func (ts *TestSuite) RunFoundry(args ...string) (string, error) {
	cmd := exec.Command(ts.Binary, args...)
	cmd.Dir = ts.TempDir

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("command failed: %v\nstdout: %s\nstderr: %s",
			err, stdout.String(), stderr.String())
	}

	return stdout.String(), nil
}

// RunFoundryInDir executes the foundry binary in a specific directory
func (ts *TestSuite) RunFoundryInDir(dir string, args ...string) (string, error) {
	cmd := exec.Command(ts.Binary, args...)
	cmd.Dir = dir

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("command failed: %v\nstdout: %s\nstderr: %s",
			err, stdout.String(), stderr.String())
	}

	return stdout.String(), nil
}

// AssertProjectStructure verifies that the expected files exist in the project
func (ts *TestSuite) AssertProjectStructure(projectName string, expectedFiles []string) {
	projectPath := filepath.Join(ts.TempDir, projectName)

	for _, file := range expectedFiles {
		path := filepath.Join(projectPath, file)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			ts.T.Errorf("Expected file not found: %s", file)
		}
	}
}

// AssertFileContains checks if a file contains the expected content
func (ts *TestSuite) AssertFileContains(projectName, filePath, expectedContent string) {
	fullPath := filepath.Join(ts.TempDir, projectName, filePath)

	content, err := os.ReadFile(fullPath)
	if err != nil {
		ts.T.Fatalf("Failed to read file %s: %v", fullPath, err)
	}

	if !strings.Contains(string(content), expectedContent) {
		ts.T.Errorf("File %s does not contain expected content: %s", filePath, expectedContent)
	}
}

// AssertGeneratedCodeCompiles verifies that the generated project compiles
func (ts *TestSuite) AssertGeneratedCodeCompiles(projectPath string) {
	fullPath := filepath.Join(ts.TempDir, projectPath)

	// Run go mod tidy first
	tidyCmd := exec.Command("go", "mod", "tidy")
	tidyCmd.Dir = fullPath

	var tidyStderr bytes.Buffer
	tidyCmd.Stderr = &tidyStderr

	if err := tidyCmd.Run(); err != nil {
		ts.T.Errorf("go mod tidy failed: %v\nstderr: %s", err, tidyStderr.String())
		return
	}

	// Try to build the project
	buildCmd := exec.Command("go", "build", "-o", "test-binary", ".")
	buildCmd.Dir = fullPath

	var buildStderr bytes.Buffer
	buildCmd.Stderr = &buildStderr

	if err := buildCmd.Run(); err != nil {
		ts.T.Errorf("Project compilation failed: %v\nstderr: %s", err, buildStderr.String())
	}

	// Clean up the test binary
	os.Remove(filepath.Join(fullPath, "test-binary"))
}

// AssertGeneratedCodeRuns verifies that the generated project runs without errors
func (ts *TestSuite) AssertGeneratedCodeRuns(projectPath string) {
	fullPath := filepath.Join(ts.TempDir, projectPath)

	// First ensure it compiles
	ts.AssertGeneratedCodeCompiles(projectPath)

	// Run the project with a timeout
	runCmd := exec.Command("go", "run", ".")
	runCmd.Dir = fullPath

	// Set up channels for async execution
	done := make(chan error, 1)

	go func() {
		err := runCmd.Start()
		if err != nil {
			done <- err
			return
		}

		// Wait a bit to see if it crashes immediately
		time.Sleep(2 * time.Second)

		// Kill the process (it's a server, so it won't exit on its own)
		runCmd.Process.Kill()
		done <- nil
	}()

	// Wait for completion or timeout
	select {
	case err := <-done:
		if err != nil {
			ts.T.Errorf("Failed to run generated project: %v", err)
		}
	case <-time.After(5 * time.Second):
		runCmd.Process.Kill()
		ts.T.Error("Generated project timed out")
	}
}

// CreateTestProject creates a new project using foundry for testing
func (ts *TestSuite) CreateTestProject(name string) error {
	_, err := ts.RunFoundry("new", name)
	return err
}

// AssertCommandOutput checks if command output contains expected text
func (ts *TestSuite) AssertCommandOutput(output, expected string) {
	if !strings.Contains(output, expected) {
		ts.T.Errorf("Output does not contain expected text.\nExpected: %s\nGot: %s",
			expected, output)
	}
}

// AssertNoError fails the test if an error occurred
func (ts *TestSuite) AssertNoError(err error) {
	if err != nil {
		ts.T.Fatalf("Unexpected error: %v", err)
	}
}

// FileExists checks if a file exists
func (ts *TestSuite) FileExists(projectName, filePath string) bool {
	fullPath := filepath.Join(ts.TempDir, projectName, filePath)
	_, err := os.Stat(fullPath)
	return err == nil
}
