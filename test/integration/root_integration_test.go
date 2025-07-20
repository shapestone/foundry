// File: test/integration/root_integration_test.go
//go:build integration

package integration

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/shapestone/foundry/internal/cli"
)

// Test data for integration tests
var (
	testVersion = "1.0.0"
	testCommit  = "abc123def"
	testDate    = "2024-01-15T10:30:00Z"
)

// TestExecute_IntegrationHappyPath tests the Execute function with valid scenarios
func TestExecute_IntegrationHappyPath(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		wantErr  bool
		contains string
	}{
		{
			name:     "help command",
			args:     []string{"--help"},
			wantErr:  false,
			contains: "Foundry is a modern CLI tool for Go developers",
		},
		{
			name:     "version command",
			args:     []string{"version"},
			wantErr:  false,
			contains: "Foundry",
		},
		{
			name:     "verbose flag",
			args:     []string{"--verbose", "--help"},
			wantErr:  false,
			contains: "verbose",
		},
		{
			name:     "config flag",
			args:     []string{"--config", "/tmp/test.yaml", "--help"},
			wantErr:  false,
			contains: "Config file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean, isolated test - no global state manipulation needed
			app := cli.NewTestCLI()

			// Execute command with isolated CLI instance
			err := app.Execute(tt.args)

			// Verify: Check results
			if (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.contains != "" {
				stdout := app.GetStdout()
				if !strings.Contains(stdout, tt.contains) {
					t.Errorf("Execute() output = %q, want to contain %q", stdout, tt.contains)
				}
			}
		})
	}
}

// TestExecute_IntegrationErrorConditions tests error scenarios
func TestExecute_IntegrationErrorConditions(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "invalid command",
			args:    []string{"nonexistent-command"},
			wantErr: true,
		},
		{
			name:    "invalid flag",
			args:    []string{"--invalid-flag"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean, isolated test
			app := cli.NewTestCLI()

			// Execute command
			err := app.Execute(tt.args)

			// Verify: Check error condition
			if (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestExecute_IntegrationConfigFile tests config file handling
func TestExecute_IntegrationConfigFile(t *testing.T) {
	// Setup: Create temporary config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test-config.yaml")

	configContent := `# Test config file
verbose: true
`
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	// Test with config file
	app := cli.NewTestCLI()
	err = app.Execute([]string{"--config", configPath, "--help"})

	// Verify: Check that config file path was set
	config := app.GetConfig()
	if config.ConfigFile != configPath {
		t.Errorf("Expected configFile to be %q, got %q", configPath, config.ConfigFile)
	}
}

// TestSetVersionInfo_IntegrationStateModification tests version info setting
func TestSetVersionInfo_IntegrationStateModification(t *testing.T) {
	tests := []struct {
		name        string
		inputVer    string
		inputCommit string
		inputDate   string
	}{
		{
			name:        "standard version info",
			inputVer:    testVersion,
			inputCommit: testCommit,
			inputDate:   testDate,
		},
		{
			name:        "empty values",
			inputVer:    "",
			inputCommit: "",
			inputDate:   "",
		},
		{
			name:        "special characters",
			inputVer:    "v1.0.0-beta+build.123",
			inputCommit: "abc123def456789",
			inputDate:   time.Now().Format(time.RFC3339),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create CLI with specific version info
			app := cli.NewTestCLI(
				cli.WithVersionInfo(tt.inputVer, tt.inputCommit, tt.inputDate),
			)

			// Verify: Check that version info was set correctly
			versionInfo := app.GetVersionInfo()
			if versionInfo.Version != tt.inputVer {
				t.Errorf("Version = %q, want %q", versionInfo.Version, tt.inputVer)
			}
			if versionInfo.Commit != tt.inputCommit {
				t.Errorf("Commit = %q, want %q", versionInfo.Commit, tt.inputCommit)
			}
			if versionInfo.Date != tt.inputDate {
				t.Errorf("Date = %q, want %q", versionInfo.Date, tt.inputDate)
			}
		})
	}
}

// TestSetVersionInfo_IntegrationVersionCommand tests version command output after setting version info
func TestSetVersionInfo_IntegrationVersionCommand(t *testing.T) {
	// Create CLI with test version info
	app := cli.NewTestCLI(
		cli.WithVersionInfo(testVersion, testCommit, testDate),
	)

	// Execute version command
	err := app.Execute([]string{"version"})
	if err != nil {
		t.Fatalf("Version command failed: %v", err)
	}

	// Verify: Check that version info appears in output
	output := app.GetStdout()
	expectedStrings := []string{
		fmt.Sprintf("Foundry %s", testVersion),
		fmt.Sprintf("Commit: %s", testCommit),
		fmt.Sprintf("Built: %s", testDate),
		fmt.Sprintf("Go version: %s", runtime.Version()),
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(output, expected) {
			t.Errorf("Version command output missing expected string: %q\nFull output: %q", expected, output)
		}
	}
}

// TestExecute_IntegrationConcurrentAccess tests concurrent command execution
func TestExecute_IntegrationConcurrentAccess(t *testing.T) {
	const numGoroutines = 10

	// Channel to collect results
	results := make(chan error, numGoroutines)

	// Execute: Run multiple goroutines concurrently with isolated CLI instances
	for i := 0; i < numGoroutines; i++ {
		go func(goroutineID int) {
			// Each goroutine gets its own CLI instance - no shared state!
			app := cli.NewTestCLI()

			var err error
			if goroutineID%2 == 0 {
				err = app.Execute([]string{"version"})
			} else {
				err = app.Execute([]string{"--help"})
			}

			results <- err
		}(i)
	}

	// Verify: Collect all results
	var errors []error
	for i := 0; i < numGoroutines; i++ {
		if err := <-results; err != nil {
			errors = append(errors, err)
		}
	}

	// Check that no errors occurred during concurrent execution
	if len(errors) > 0 {
		t.Errorf("Concurrent execution produced errors: %v", errors)
	}
}

// BenchmarkExecute_Integration benchmarks the Execute function
func BenchmarkExecute_Integration(b *testing.B) {
	app := cli.NewTestCLI()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		app.Execute([]string{"version"})
		app.ResetOutput() // Clear output between runs
	}
}

// TestExecute_IntegrationRealCommandLineExecution tests actual command line execution
func TestExecute_IntegrationRealCommandLineExecution(t *testing.T) {
	// This test requires the binary to be built first
	// Skip if we can't find or build the binary

	// Try to build the binary
	tmpDir := t.TempDir()
	binaryPath := filepath.Join(tmpDir, "foundry-test")

	// Build from the project root (two levels up from test/integration)
	buildCmd := exec.Command("go", "build", "-o", binaryPath, "./cmd/foundry")
	buildCmd.Dir = "../.."

	if err := buildCmd.Run(); err != nil {
		t.Skipf("Skipping real execution test: could not build binary: %v", err)
	}

	// Make the binary executable
	if err := os.Chmod(binaryPath, 0755); err != nil {
		t.Skipf("Skipping real execution test: could not make binary executable: %v", err)
	}

	tests := []struct {
		name     string
		args     []string
		wantErr  bool
		contains string
	}{
		{
			name:     "version command real execution",
			args:     []string{"version"},
			wantErr:  false,
			contains: "Foundry",
		},
		{
			name:     "help command real execution",
			args:     []string{"--help"},
			wantErr:  false,
			contains: "Foundry is a modern CLI tool for Go developers",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Execute: Run the actual binary
			cmd := exec.Command(binaryPath, tt.args...)
			output, err := cmd.CombinedOutput()

			// Verify: Check results
			if (err != nil) != tt.wantErr {
				t.Errorf("Real execution error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.contains != "" && !strings.Contains(string(output), tt.contains) {
				t.Errorf("Real execution output = %q, want to contain %q", string(output), tt.contains)
			}
		})
	}
}
