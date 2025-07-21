// test/integration/debug_test.go
package integration

import (
	"testing"
)

// TestFoundryInitDebug shows us what's actually happening
func TestFoundryInitDebug(t *testing.T) {
	h := NewTestHelper(t)

	// Test a simple init command and show all output
	output, err := h.RunFoundry("init", "test-project", "--force")

	t.Logf("Command: foundry init test-project --force")
	t.Logf("Output: %s", output)
	t.Logf("Error: %v", err)

	if err != nil {
		t.Logf("Exit status: %v", err)
		// Don't fail - we want to see what happened
	}
}

// TestFoundryNewDebug shows us what's actually happening with new
func TestFoundryNewDebug(t *testing.T) {
	h := NewTestHelper(t)

	// Test a simple new command and show all output
	output, err := h.RunFoundry("new", "test-project")

	t.Logf("Command: foundry new test-project")
	t.Logf("Output: %s", output)
	t.Logf("Error: %v", err)

	if err != nil {
		t.Logf("Exit status: %v", err)
		// Don't fail - we want to see what happened
	}
}
