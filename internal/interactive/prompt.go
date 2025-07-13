package interactive

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

// ANSI color codes
const (
	ColorReset  = "\033[0m"
	ColorRed    = "\033[31m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
	ColorBlue   = "\033[34m"
	ColorPurple = "\033[35m"
	ColorCyan   = "\033[36m"
	ColorWhite  = "\033[37m"
)

// Prompter handles user interaction
type Prompter interface {
	Confirm(message string) bool
	ShowPreview(title string, changes []string, message string) bool
}

// ConsolePrompter implements Prompter for console interaction
type ConsolePrompter struct {
	input  io.Reader
	output io.Writer
}

// NewConsolePrompter creates a new console prompter
func NewConsolePrompter() *ConsolePrompter {
	return &ConsolePrompter{
		input:  os.Stdin,
		output: os.Stdout,
	}
}

// NewConsolePrompterWithIO creates a new console prompter with custom input/output
// This is useful for testing
func NewConsolePrompterWithIO(input io.Reader, output io.Writer) *ConsolePrompter {
	return &ConsolePrompter{
		input:  input,
		output: output,
	}
}

// Confirm asks the user for confirmation
func (p *ConsolePrompter) Confirm(message string) bool {
	scanner := bufio.NewScanner(p.input)

	for {
		fmt.Fprintf(p.output, "%s (y/N): ", message)

		if !scanner.Scan() {
			// Handle EOF or error
			return false
		}

		response := strings.TrimSpace(strings.ToLower(scanner.Text()))

		switch response {
		case "y", "yes":
			return true
		case "n", "no", "":
			return false
		default:
			fmt.Fprintf(p.output, "Please enter 'y' for yes or 'n' for no.\n")
			continue
		}
	}
}

// ShowPreview shows a preview of changes and asks for confirmation
func (p *ConsolePrompter) ShowPreview(title string, changes []string, message string) bool {
	if len(changes) == 0 {
		fmt.Fprintf(p.output, "\n📝 %s\n", "No changes")
		if message != "" {
			return p.Confirm(message)
		}
		return p.Confirm("Continue?")
	}

	fmt.Fprintf(p.output, "\n📝 %s\n", title)
	fmt.Fprintf(p.output, "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")

	for _, change := range changes {
		coloredChange := p.colorizeChange(change)
		fmt.Fprintf(p.output, "  %s\n", coloredChange)
	}

	fmt.Fprintf(p.output, "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")

	if message == "" {
		message = "Apply these changes?"
	}

	return p.Confirm(fmt.Sprintf("\n❓ %s", message))
}

// colorizeChange applies color formatting to changes based on their prefix
func (p *ConsolePrompter) colorizeChange(change string) string {
	change = strings.TrimSpace(change)

	if strings.HasPrefix(change, "+") {
		// Green for additions
		return fmt.Sprintf("%s%s%s", ColorGreen, change, ColorReset)
	} else if strings.HasPrefix(change, "-") {
		// Red for deletions
		return fmt.Sprintf("%s%s%s", ColorRed, change, ColorReset)
	} else if strings.HasPrefix(change, "~") || strings.HasPrefix(change, "M") {
		// Yellow for modifications
		return fmt.Sprintf("%s%s%s", ColorYellow, change, ColorReset)
	}

	// Default color for other lines
	return change
}

// IsColorSupported checks if the current terminal supports color output
func IsColorSupported() bool {
	term := os.Getenv("TERM")
	if term == "" {
		return false
	}

	// Check for common terminals that support color
	colorTerms := []string{"xterm", "xterm-256color", "screen", "tmux", "vt100"}
	for _, colorTerm := range colorTerms {
		if strings.Contains(term, colorTerm) {
			return true
		}
	}

	// Check for explicit color support environment variables
	if os.Getenv("COLORTERM") != "" {
		return true
	}

	return false
}

// Helper function to strip ANSI color codes (useful for testing)
func StripAnsiCodes(s string) string {
	// Simple regex-like replacement for ANSI escape sequences
	result := s
	ansiCodes := []string{
		ColorReset, ColorRed, ColorGreen, ColorYellow,
		ColorBlue, ColorPurple, ColorCyan, ColorWhite,
	}

	for _, code := range ansiCodes {
		result = strings.ReplaceAll(result, code, "")
	}

	return result
}
