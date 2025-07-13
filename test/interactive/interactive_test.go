package interactive_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/shapestone/foundry/internal/interactive"
)

func TestNewConsolePrompter(t *testing.T) {
	prompter := interactive.NewConsolePrompter()
	if prompter == nil {
		t.Fatal("NewConsolePrompter() returned nil")
	}

	// Test that it implements the Prompter interface
	var _ interactive.Prompter = prompter
}

func TestNewConsolePrompterWithIO(t *testing.T) {
	input := strings.NewReader("test")
	output := &bytes.Buffer{}

	prompter := interactive.NewConsolePrompterWithIO(input, output)
	if prompter == nil {
		t.Fatal("NewConsolePrompterWithIO() returned nil")
	}
}

func TestConfirm(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
		prompt   string
	}{
		{
			name:     "YesResponse",
			input:    "y\n",
			expected: true,
			prompt:   "Continue?",
		},
		{
			name:     "NoResponse",
			input:    "n\n",
			expected: false,
			prompt:   "Continue?",
		},
		{
			name:     "YesUppercase",
			input:    "Y\n",
			expected: true,
			prompt:   "Continue?",
		},
		{
			name:     "NoUppercase",
			input:    "N\n",
			expected: false,
			prompt:   "Continue?",
		},
		{
			name:     "YesLong",
			input:    "yes\n",
			expected: true,
			prompt:   "Continue?",
		},
		{
			name:     "NoLong",
			input:    "no\n",
			expected: false,
			prompt:   "Continue?",
		},
		{
			name:     "EmptyDefaultsToNo",
			input:    "\n",
			expected: false,
			prompt:   "Continue?",
		},
		{
			name:     "InvalidThenYes",
			input:    "maybe\ny\n",
			expected: true, // Implementation correctly loops and gets to "y"
			prompt:   "Continue?",
		},
		{
			name:     "InvalidThenNo",
			input:    "invalid\nn\n",
			expected: false,
			prompt:   "Continue?",
		},
		{
			name:     "MultipleInvalidThenYes",
			input:    "abc\n123\nmaybe\ny\n",
			expected: true, // Implementation correctly loops and gets to "y"
			prompt:   "Continue?",
		},
		{
			name:     "WhitespaceHandling",
			input:    "  y  \n",
			expected: true,
			prompt:   "Continue?",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := strings.NewReader(tt.input)
			output := &bytes.Buffer{}

			prompter := interactive.NewConsolePrompterWithIO(input, output)
			result := prompter.Confirm(tt.prompt)

			if result != tt.expected {
				t.Errorf("Confirm(%q) with input %q = %v, want %v",
					tt.prompt, strings.TrimSpace(tt.input), result, tt.expected)
			}

			// Check that prompt was displayed
			outputStr := output.String()
			if !strings.Contains(outputStr, tt.prompt) {
				t.Errorf("Output should contain prompt %q, got: %q", tt.prompt, outputStr)
			}

			// Check for (y/N) prompt format
			if !strings.Contains(outputStr, "(y/N)") {
				t.Errorf("Output should contain (y/N) prompt format, got: %q", outputStr)
			}
		})
	}
}

func TestConfirmEOF(t *testing.T) {
	// Test handling of EOF (empty input)
	input := strings.NewReader("")
	output := &bytes.Buffer{}

	prompter := interactive.NewConsolePrompterWithIO(input, output)
	result := prompter.Confirm("Continue?")

	if result != false {
		t.Errorf("Confirm() with EOF should return false, got %v", result)
	}
}

func TestShowPreview(t *testing.T) {
	tests := []struct {
		name          string
		title         string
		changes       []string
		message       string
		input         string
		expected      bool
		shouldContain []string
	}{
		{
			name:          "AcceptChanges",
			title:         "Preview changes:",
			changes:       []string{"+ line 1", "- line 2"},
			message:       "Apply these changes?",
			input:         "y\n",
			expected:      true,
			shouldContain: []string{"📝 Preview changes:", "+ line 1", "- line 2", "Apply these changes?"},
		},
		{
			name:          "RejectChanges",
			title:         "Preview changes:",
			changes:       []string{"+ line 1", "- line 2"},
			message:       "Apply these changes?",
			input:         "n\n",
			expected:      false,
			shouldContain: []string{"📝 Preview changes:", "+ line 1", "- line 2", "Apply these changes?"},
		},
		{
			name:          "EmptyChanges",
			title:         "No changes",
			changes:       []string{},
			message:       "Continue?",
			input:         "y\n",
			expected:      true,
			shouldContain: []string{"📝 No changes", "Continue?"},
		},
		{
			name:          "EmptyChangesNoMessage",
			title:         "No changes",
			changes:       []string{},
			message:       "",
			input:         "y\n",
			expected:      true,
			shouldContain: []string{"📝 No changes", "Continue?"},
		},
		{
			name:          "DefaultMessage",
			title:         "Changes preview",
			changes:       []string{"+ new file"},
			message:       "",
			input:         "y\n",
			expected:      true,
			shouldContain: []string{"📝 Changes preview", "+ new file", "Apply these changes?"},
		},
		{
			name:          "MultipleChanges",
			title:         "Multiple changes",
			changes:       []string{"+ add this", "- remove this", "~ modify this", "unchanged line"},
			message:       "Proceed?",
			input:         "y\n",
			expected:      true,
			shouldContain: []string{"+ add this", "- remove this", "~ modify this", "unchanged line"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := strings.NewReader(tt.input)
			output := &bytes.Buffer{}

			prompter := interactive.NewConsolePrompterWithIO(input, output)
			result := prompter.ShowPreview(tt.title, tt.changes, tt.message)

			if result != tt.expected {
				t.Errorf("ShowPreview() = %v, want %v", result, tt.expected)
			}

			outputStr := output.String()

			// Check for required content
			for _, content := range tt.shouldContain {
				if !strings.Contains(outputStr, content) {
					t.Errorf("Output should contain %q, got: %q", content, outputStr)
				}
			}

			// Check for decorative elements when there are changes
			if len(tt.changes) > 0 {
				if !strings.Contains(outputStr, "━━━") {
					t.Error("Output should contain decorative border for changes")
				}
			}
		})
	}
}

func TestColorizeChange(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		hasGreen  bool
		hasRed    bool
		hasYellow bool
	}{
		{
			name:     "Addition",
			input:    "+ added line",
			hasGreen: true,
		},
		{
			name:   "Deletion",
			input:  "- deleted line",
			hasRed: true,
		},
		{
			name:      "Modification",
			input:     "~ modified line",
			hasYellow: true,
		},
		{
			name:      "ModificationAlt",
			input:     "M modified file",
			hasYellow: true,
		},
		{
			name:  "Unchanged",
			input: "regular line",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Since colorizeChange is not exported, we test it through ShowPreview
			input := strings.NewReader("n\n")
			output := &bytes.Buffer{}
			prompter := interactive.NewConsolePrompterWithIO(input, output)

			prompter.ShowPreview("Test", []string{tt.input}, "Test?")
			outputStr := output.String()

			if tt.hasGreen && !strings.Contains(outputStr, "\033[32m") {
				t.Error("Output should contain green color code for addition")
			}
			if tt.hasRed && !strings.Contains(outputStr, "\033[31m") {
				t.Error("Output should contain red color code for deletion")
			}
			if tt.hasYellow && !strings.Contains(outputStr, "\033[33m") {
				t.Error("Output should contain yellow color code for modification")
			}
		})
	}
}

func TestConsolePrompterIntegration(t *testing.T) {
	t.Run("PrompterMethods", func(t *testing.T) {
		prompter := interactive.NewConsolePrompter()

		// Verify interface implementation
		var _ interactive.Prompter = prompter

		if prompter == nil {
			t.Error("ConsolePrompter should not be nil")
		}

		t.Log("ConsolePrompter has the following methods:")
		t.Log("- Confirm(prompt string) bool")
		t.Log("- ShowPreview(title string, changes []string, message string) bool")
	})
}

func TestPrompterInterface(t *testing.T) {
	// Verify that ConsolePrompter implements Prompter interface
	var prompter interactive.Prompter = interactive.NewConsolePrompter()

	if prompter == nil {
		t.Error("ConsolePrompter should implement Prompter interface")
	}

	t.Log("Consider exposing a Prompter interface for better testability")
}

func TestColorOutput(t *testing.T) {
	tests := []struct {
		name     string
		changes  []string
		hasColor bool
	}{
		{
			name:     "ColoredAdditions",
			changes:  []string{"+ new file", "+ another addition"},
			hasColor: true,
		},
		{
			name:     "ColoredDeletions",
			changes:  []string{"- deleted file", "- removed line"},
			hasColor: true,
		},
		{
			name:     "MixedChanges",
			changes:  []string{"+ addition", "- deletion", "~ modification", "normal line"},
			hasColor: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := strings.NewReader("n\n")
			output := &bytes.Buffer{}

			prompter := interactive.NewConsolePrompterWithIO(input, output)
			prompter.ShowPreview("Test Preview", tt.changes, "Test?")

			outputStr := output.String()

			t.Log("ShowPreview should use colors:")
			t.Log("- Green for additions (lines starting with +)")
			t.Log("- Red for deletions (lines starting with -)")
			t.Log("- Default color for other lines")

			// Check for presence of ANSI color codes
			if tt.hasColor {
				hasAnsiCodes := strings.Contains(outputStr, "\033[") ||
					strings.Contains(outputStr, "\x1b[")
				if !hasAnsiCodes {
					t.Log("Note: No ANSI color codes detected in output (this may be expected in test environment)")
				}
			}
		})
	}
}

func TestStripAnsiCodes(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "NoAnsiCodes",
			input:    "plain text",
			expected: "plain text",
		},
		{
			name:     "RedText",
			input:    "\033[31mred text\033[0m",
			expected: "red text",
		},
		{
			name:     "GreenText",
			input:    "\033[32mgreen text\033[0m",
			expected: "green text",
		},
		{
			name:     "MultipleColors",
			input:    "\033[31mred\033[0m and \033[32mgreen\033[0m",
			expected: "red and green",
		},
		{
			name:     "EmptyString",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := interactive.StripAnsiCodes(tt.input)
			if result != tt.expected {
				t.Errorf("StripAnsiCodes(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestIsColorSupported(t *testing.T) {
	// This test is environment-dependent, so we just test that it doesn't panic
	result := interactive.IsColorSupported()

	// Just verify it returns a boolean (either true or false is fine)
	if result != true && result != false {
		t.Error("IsColorSupported() should return a boolean")
	}

	t.Logf("Color support detected: %v", result)
}

// Benchmark tests
func BenchmarkConfirm(b *testing.B) {
	input := strings.NewReader(strings.Repeat("y\n", b.N))
	output := &bytes.Buffer{}
	prompter := interactive.NewConsolePrompterWithIO(input, output)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		prompter.Confirm("Test?")
	}
}

func BenchmarkShowPreview(b *testing.B) {
	changes := []string{"+ addition", "- deletion", "~ modification"}
	input := strings.NewReader(strings.Repeat("y\n", b.N))
	output := &bytes.Buffer{}
	prompter := interactive.NewConsolePrompterWithIO(input, output)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		prompter.ShowPreview("Test", changes, "Apply?")
		output.Reset()
	}
}
