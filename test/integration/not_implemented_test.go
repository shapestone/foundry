// test/integration/not_implemented_test.go
// Add missing import
package integration

import (
	"strings"
	"testing"
)

// TestNotImplementedFeatures tests that unimplemented features give honest errors
func TestNotImplementedFeatures(t *testing.T) {
	tests := []struct {
		name             string
		args             []string
		expectedError    string
		shouldContain    []string
		shouldNotContain []string
	}{
		{
			name:          "layout add remote",
			args:          []string{"layout", "add", "github.com/user/repo"},
			expectedError: "not implemented",
			shouldContain: []string{
				"❌ Remote layout support is not yet implemented",
				"Current workarounds:",
				"git clone github.com/user/repo",
			},
			shouldNotContain: []string{
				"Successfully added",
				"✅",
			},
		},
		{
			name:          "layout update",
			args:          []string{"layout", "update"},
			expectedError: "not implemented",
			shouldContain: []string{
				"❌ Layout registry updates are not yet implemented",
				"This feature will refresh remote layout sources",
			},
			shouldNotContain: []string{
				"updated successfully",
				"✅",
			},
		},
		{
			name:          "layout update specific",
			args:          []string{"layout", "update", "some-layout"},
			expectedError: "not yet implemented",
			shouldContain: []string{
				"❌ Updating individual layouts is not yet implemented",
				"Layout 'some-layout' update skipped",
			},
		},
		{
			name:          "layout remove",
			args:          []string{"layout", "remove", "test-layout"},
			expectedError: "not yet implemented",
			shouldContain: []string{
				"Removing layout 'test-layout'",
				"layout removal not yet implemented",
			},
		},
		{
			name:          "wire handler",
			args:          []string{"wire", "handler", "test"},
			expectedError: "", // Wire shows warning but doesn't error
			shouldContain: []string{
				"⚠️  Wire command is not yet fully implemented",
				"This feature will automatically:",
				"Update imports",
				"Register handlers with routers",
			},
		},
		{
			name:          "wire middleware",
			args:          []string{"wire", "middleware", "auth"},
			expectedError: "", // Wire shows warning but doesn't error
			shouldContain: []string{
				"⚠️  Wire command is not yet fully implemented",
				"Update dependency injection configurations",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewTestHelper(t)

			output, err := h.RunFoundry(tt.args...)

			// Check error expectation
			if tt.expectedError != "" {
				h.AssertError(err, tt.expectedError)
			}

			// Check required content
			for _, content := range tt.shouldContain {
				h.AssertOutputContains(output, content)
			}

			// Check forbidden content
			for _, content := range tt.shouldNotContain {
				if len(content) > 0 && output != "" {
					if strings.Contains(output, content) {
						t.Errorf("Output should not contain '%s', but it does.\nOutput: %s", content, output)
					}
				}
			}
		})
	}
}

// TestLayoutInfo tests the layout info command
func TestLayoutInfo(t *testing.T) {
	h := NewTestHelper(t)

	// Test standard layout info
	output, err := h.RunFoundry("layout", "info", "standard")
	h.AssertNoError(err)

	// Should show layout information
	h.AssertOutputContains(output, "Layout: standard")
	h.AssertOutputContains(output, "Version:")
	h.AssertOutputContains(output, "Description:")
	h.AssertOutputContains(output, "Directory Structure:")

	// Test nonexistent layout
	output, err = h.RunFoundry("layout", "info", "nonexistent-layout")
	h.AssertError(err, "not found")
}

// TestLayoutList tests the layout list command
func TestLayoutList(t *testing.T) {
	h := NewTestHelper(t)

	// Test basic layout list
	output, err := h.RunFoundry("layout", "list")
	h.AssertNoError(err)

	// Should show at least the standard layout
	h.AssertOutputContains(output, "standard")

	// Test with filters
	output, err = h.RunFoundry("layout", "list", "--local")
	h.AssertNoError(err)

	output, err = h.RunFoundry("layout", "list", "--remote")
	h.AssertNoError(err)
}

// TestNewProjectListLayouts tests the --list-layouts flag
func TestNewProjectListLayouts(t *testing.T) {
	h := NewTestHelper(t)

	output, err := h.RunFoundry("new", "--list-layouts")
	h.AssertNoError(err)

	// Should show available layouts
	h.AssertOutputContains(output, "Available layouts:")
	h.AssertOutputContains(output, "standard")
	h.AssertOutputContains(output, "microservice")
}

// TestInvalidCommands tests error handling for invalid commands
func TestInvalidCommands(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		expectError bool
	}{
		{
			name:        "invalid command",
			args:        []string{"invalid-command"},
			expectError: true,
		},
		{
			name:        "add without type",
			args:        []string{"add"},
			expectError: true,
		},
		{
			name:        "add handler without name",
			args:        []string{"add", "handler"},
			expectError: true,
		},
		{
			name:        "init with invalid project name",
			args:        []string{"init", "invalid!@#$%"},
			expectError: true,
		},
		{
			name:        "new with invalid project name",
			args:        []string{"new", "invalid spaces"},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewTestHelper(t)

			_, err := h.RunFoundry(tt.args...)

			if tt.expectError {
				if err == nil {
					t.Fatalf("Expected error for command %v, but got none", tt.args)
				}
			} else {
				h.AssertNoError(err)
			}
		})
	}
}
