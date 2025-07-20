// File: internal/cli/testing.go
package cli

import (
	"bytes"
)

// NewTestCLI creates a CLI instance configured for testing
func NewTestCLI(opts ...Option) *CLI {
	var stdout, stderr bytes.Buffer

	// Default test options
	testOpts := []Option{
		WithOutput(&stdout, &stderr),
		WithConfig(&Config{}),
	}

	// Append any additional options
	testOpts = append(testOpts, opts...)

	return New(testOpts...)
}

// GetOutput returns the captured stdout and stderr from the CLI
func (c *CLI) GetOutput() (stdout, stderr string) {
	if buf, ok := c.stdout.(*bytes.Buffer); ok {
		stdout = buf.String()
	}
	if buf, ok := c.stderr.(*bytes.Buffer); ok {
		stderr = buf.String()
	}
	return
}

// GetStdoutString returns just the captured stdout as a string
func (c *CLI) GetStdoutString() string {
	if buf, ok := c.stdout.(*bytes.Buffer); ok {
		return buf.String()
	}
	return ""
}

// GetStderrString returns just the captured stderr as a string
func (c *CLI) GetStderrString() string {
	if buf, ok := c.stderr.(*bytes.Buffer); ok {
		return buf.String()
	}
	return ""
}

// ResetOutput clears the captured output buffers
func (c *CLI) ResetOutput() {
	if buf, ok := c.stdout.(*bytes.Buffer); ok {
		buf.Reset()
	}
	if buf, ok := c.stderr.(*bytes.Buffer); ok {
		buf.Reset()
	}
}

// SetInput sets the input for the CLI (useful for testing interactive commands)
func (c *CLI) SetInput(input string) {
	c.stdin = bytes.NewBufferString(input)
}
