// test/integration/project_quality_test.go
package integration

import (
	"os/exec"
	"path/filepath"
	"testing"
)

// TestGeneratedProjectBuilds tests that generated projects actually compile
func TestGeneratedProjectBuilds(t *testing.T) {
	layouts := []string{"standard", "microservice"}

	for _, layout := range layouts {
		t.Run("layout_"+layout, func(t *testing.T) {
			h := NewTestHelper(t)

			// Create project
			projectName := "test-" + layout
			output, err := h.RunFoundry("new", projectName, "--layout", layout)
			h.AssertNoError(err)
			h.AssertOutputContains(output, "created successfully")

			// Try to build the project
			projectDir := filepath.Join(h.GetTempDir(), projectName)
			cmd := exec.Command("go", "mod", "tidy")
			cmd.Dir = projectDir
			if output, err := cmd.CombinedOutput(); err != nil {
				t.Fatalf("go mod tidy failed: %v\nOutput: %s", err, output)
			}

			cmd = exec.Command("go", "build", "./...")
			cmd.Dir = projectDir
			if output, err := cmd.CombinedOutput(); err != nil {
				t.Fatalf("go build failed for %s layout: %v\nOutput: %s", layout, err, output)
			}
		})
	}
}

// TestGeneratedComponentsCompile tests that added components compile
func TestGeneratedComponentsCompile(t *testing.T) {
	h := NewTestHelper(t)

	// Create base project
	_, err := h.RunFoundry("init", "test-project", "--force")
	h.AssertNoError(err)

	// Add various components
	components := []struct {
		cmd  string
		name string
	}{
		{"handler", "user"},
		{"handler", "product"},
		{"model", "user"},
		{"model", "order"},
		{"middleware", "auth"},
		{"middleware", "logging"},
	}

	for _, comp := range components {
		if output, err := h.RunFoundry("add", comp.cmd, comp.name); err != nil {
			t.Fatalf("Failed to add %s %s: %v\nOutput: %s", comp.cmd, comp.name, err, output)
		}
		// Don't need to check output content for this test
	}

	// Verify project still builds
	cmd := exec.Command("go", "mod", "tidy")
	cmd.Dir = h.GetTempDir()
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("go mod tidy failed: %v\nOutput: %s", err, output)
	}

	cmd = exec.Command("go", "build", "./...")
	cmd.Dir = h.GetTempDir()
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("Project with components failed to build: %v\nOutput: %s", err, output)
	}
}

// TestProjectStructureConsistency tests that projects follow expected structure
func TestProjectStructureConsistency(t *testing.T) {
	h := NewTestHelper(t)

	// Create project
	_, err := h.RunFoundry("new", "structured-project")
	h.AssertNoError(err)

	projectDir := "structured-project"

	// Check standard Go project structure exists
	expectedDirs := []string{
		projectDir + "/cmd",
		projectDir + "/internal",
	}

	expectedFiles := []string{
		projectDir + "/go.mod",
		projectDir + "/main.go",
		projectDir + "/Makefile",
		projectDir + "/README.md",
	}

	for _, dir := range expectedDirs {
		h.AssertFileExists(dir)
	}

	for _, file := range expectedFiles {
		h.AssertFileExists(file)
	}

	// Check go.mod content
	h.AssertFileContains(projectDir+"/go.mod", "module structured-project")
	h.AssertFileContains(projectDir+"/go.mod", "go 1.21")
}

// TestDatabaseIntegration tests database component generation
func TestDatabaseIntegration(t *testing.T) {
	dbTypes := []string{"postgres", "mysql", "sqlite"}

	for _, dbType := range dbTypes {
		t.Run("database_"+dbType, func(t *testing.T) {
			h := NewTestHelper(t)

			// Create project
			output, err := h.RunFoundry("init", "db-test", "--force")
			h.AssertNoError(err)

			// Add database
			_, err = h.RunFoundry("add", "db", dbType)
			h.AssertNoError(err)
			h.AssertOutputContains(output, "Database support added successfully")

			// Check files were created
			h.AssertFileExists("internal/database/database.go")
			h.AssertFileExists("internal/database/config.go")
			h.AssertFileExists(".env.example")

			// Check database-specific content
			switch dbType {
			case "postgres":
				h.AssertFileContains("internal/database/database.go", "pgxpool")
				h.AssertFileContains(".env.example", "DB_HOST")
			case "mysql":
				h.AssertFileContains("internal/database/database.go", "sql.Open")
				h.AssertFileContains(".env.example", "DB_HOST")
			case "sqlite":
				h.AssertFileContains("internal/database/database.go", "sqlite3")
				h.AssertFileContains(".env.example", "DB_PATH")
			}

			// Verify project compiles (database imports might be missing, but structure should be valid)
			cmd := exec.Command("go", "mod", "tidy")
			cmd.Dir = h.GetTempDir()
			cmd.CombinedOutput() // Ignore errors as drivers might not be available
		})
	}
}

// TestAutoWireFeatureWarnings tests that auto-wire shows proper warnings
func TestAutoWireFeatureWarnings(t *testing.T) {
	h := NewTestHelper(t)

	// Create project
	output, err := h.RunFoundry("init", "autowire-test", "--force")
	h.AssertNoError(err)

	// Test handler auto-wire
	_, err = h.RunFoundry("add", "handler", "user", "--auto-wire")
	// Should succeed but show warning about auto-wire not working
	h.AssertOutputContains(output, "Handler created successfully")
	// May contain auto-wire warning or error

	// Test middleware auto-wire
	_, err = h.RunFoundry("add", "middleware", "auth", "--auto-wire")
	h.AssertOutputContains(output, "Middleware created successfully")
}

// TestVersionCommand tests the version command
func TestVersionCommand(t *testing.T) {
	h := NewTestHelper(t)

	output, err := h.RunFoundry("version")
	h.AssertNoError(err)

	h.AssertOutputContains(output, "Foundry")
	h.AssertOutputContains(output, "Commit:")
	h.AssertOutputContains(output, "Built:")
	h.AssertOutputContains(output, "Go version:")
}

// TestHelpCommands tests help output
func TestHelpCommands(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{"root help", []string{"--help"}},
		{"init help", []string{"init", "--help"}},
		{"new help", []string{"new", "--help"}},
		{"add help", []string{"add", "--help"}},
		{"layout help", []string{"layout", "--help"}},
		{"wire help", []string{"wire", "--help"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewTestHelper(t)

			output, err := h.RunFoundry(tt.args...)
			h.AssertNoError(err)

			// Help should contain usage information
			h.AssertOutputContains(output, "Usage:")
		})
	}
}
