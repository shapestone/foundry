// internal/parser/handler_parser.go
package parser

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/shapestone/foundry/internal/scaffolder"
)

// handlerParser implements parsing for handler commands
type handlerParser struct {
	validator       EnhancedValidator
	projectAnalyzer ProjectAnalyzer
	flagExtractor   FlagExtractor
}

// ParseHandlerCommand parses handler command arguments and flags into a HandlerSpec
func (p *handlerParser) ParseHandlerCommand(ctx context.Context, args []string, flags map[string]interface{}) (*scaffolder.HandlerSpec, error) {
	// Validate basic arguments
	if len(args) == 0 {
		return nil, &ParseError{
			Command:     "handler",
			Args:        args,
			Flags:       flags,
			Message:     "handler name is required",
			Code:        string(CodeArgsInsufficient),
			Suggestions: []string{"Provide a handler name", "Example: foundry add handler user"},
		}
	}

	if len(args) > 1 {
		return nil, &ParseError{
			Command:     "handler",
			Args:        args,
			Flags:       flags,
			Message:     "too many arguments provided",
			Code:        string(CodeArgsExcess),
			Suggestions: []string{"Use only one handler name", "Example: foundry add handler user"},
		}
	}

	handlerName := args[0]

	// Create validation context
	validationCtx := &ValidationContext{
		Command:     "add handler",
		Args:        args,
		Flags:       flags,
		ProjectRoot: p.getProjectRoot(),
		WorkingDir:  p.getWorkingDir(),
	}

	// Validate with enhanced validator
	validationResult := p.validator.ValidateWithContext(validationCtx)
	if !validationResult.Valid {
		return nil, p.createParseErrorFromValidation(validationResult)
	}

	// Extract flags with defaults
	handlerType := p.flagExtractor.ExtractString(flags, "type", "REST")
	autoWire := p.flagExtractor.ExtractBool(flags, "auto-wire", false)
	dryRun := p.flagExtractor.ExtractBool(flags, "dry-run", false)

	// Get project information
	projectRoot := validationCtx.ProjectRoot
	moduleName, err := p.projectAnalyzer.GetModuleName(projectRoot)
	if err != nil {
		return nil, &ParseError{
			Command:     "handler",
			Args:        args,
			Flags:       flags,
			Message:     "failed to determine module name",
			Code:        string(CodeProjectInvalid),
			Cause:       err,
			Suggestions: []string{"Ensure you're in a valid Go project", "Check that go.mod exists"},
		}
	}

	// Check for existing handler conflicts
	existingComponents, err := p.projectAnalyzer.GetExistingComponents(projectRoot)
	if err == nil { // Don't fail if we can't get existing components
		for _, existingHandler := range existingComponents.Handlers {
			if existingHandler == handlerName {
				return nil, &ParseError{
					Command: "handler",
					Args:    args,
					Flags:   flags,
					Message: fmt.Sprintf("handler '%s' already exists", handlerName),
					Code:    string(CodeNameExists),
					Suggestions: []string{
						"Choose a different handler name",
						"Use --force to overwrite (if implemented)",
						fmt.Sprintf("Try: foundry add handler %s2", handlerName),
					},
				}
			}
		}
	}

	// Create metadata from context
	metadata := make(map[string]string)
	metadata["command"] = "add handler"
	metadata["parser_version"] = "2.0"
	if validationResult.HasWarnings() {
		metadata["warnings_count"] = fmt.Sprintf("%d", len(validationResult.Warnings))
	}

	// Build the handler specification
	spec := &scaffolder.HandlerSpec{
		Name:        handlerName,
		Type:        handlerType,
		AutoWire:    autoWire,
		DryRun:      dryRun,
		ProjectRoot: projectRoot,
		Module:      moduleName,
		Metadata:    metadata,
	}

	return spec, nil
}

// Helper methods

func (p *handlerParser) getProjectRoot() string {
	if root, err := p.projectAnalyzer.GetProjectRoot(); err == nil {
		return root
	}
	// Fallback to current directory
	if cwd, err := os.Getwd(); err == nil {
		return cwd
	}
	return "."
}

func (p *handlerParser) getWorkingDir() string {
	if cwd, err := os.Getwd(); err == nil {
		return cwd
	}
	return "."
}

func (p *handlerParser) createParseErrorFromValidation(result *ValidationResult) error {
	if len(result.Errors) == 0 {
		return nil
	}

	// Use the first error as the primary error
	primaryError := result.Errors[0]

	suggestions := primaryError.Suggestions
	if len(suggestions) == 0 && len(result.Suggestions) > 0 {
		// Add general suggestions if no specific ones
		for _, suggestion := range result.Suggestions {
			suggestions = append(suggestions, suggestion.Suggestion)
		}
	}

	parseError := &ParseError{
		Command:     result.Context.Command,
		Args:        result.Context.Args,
		Flags:       result.Context.Flags,
		Message:     primaryError.Message,
		Code:        primaryError.Code,
		Suggestions: suggestions,
	}

	// If there are multiple errors, wrap them
	if len(result.Errors) > 1 {
		var allErrors ParseErrors
		for _, err := range result.Errors {
			allErrors = append(allErrors, &ParseError{
				Command:     result.Context.Command,
				Args:        result.Context.Args,
				Flags:       result.Context.Flags,
				Message:     err.Message,
				Code:        err.Code,
				Suggestions: err.Suggestions,
			})
		}
		return allErrors
	}

	return parseError
}

// validateHandlerName validates the handler name with rich context
func (p *handlerParser) validateHandlerName(name string, ctx *ValidationContext) *ValidationResult {
	result := NewValidationResult()
	result.Context = ctx

	// Basic validation
	if name == "" {
		result.AddError("name", "Handler name cannot be empty", CodeNameEmpty)
		result.AddSuggestion("name", "user", "A common handler name", 0.9)
		result.AddSuggestion("name", "api", "For API handlers", 0.8)
		return result
	}

	if len(name) < 2 {
		result.AddError("name", "Handler name must be at least 2 characters long", CodeNameTooShort)
		result.AddSuggestion("name", name+"Handler", "Add 'Handler' suffix", 0.7)
		return result
	}

	if len(name) > 50 {
		result.AddError("name", "Handler name must be less than 50 characters", CodeNameTooLong)
		result.AddSuggestion("name", name[:20], "Truncate to 20 characters", 0.6)
		return result
	}

	// Check for invalid characters
	if !isValidGoIdentifier(name) {
		result.AddError("name", "Handler name must be a valid Go identifier", CodeNameInvalidChars)
		cleanName := sanitizeGoIdentifier(name)
		if cleanName != name {
			result.AddSuggestion("name", cleanName, "Remove invalid characters", 0.9)
		}
	}

	// Check for reserved words
	if isGoReservedWord(name) {
		result.AddError("name", fmt.Sprintf("'%s' is a Go reserved word", name), CodeNameReservedWord)
		result.AddSuggestion("name", name+"Handler", "Add 'Handler' suffix", 0.9)
		result.AddSuggestion("name", "my"+capitalize(name), "Add 'my' prefix", 0.7)
	}

	// Add helpful warnings for common patterns
	if len(name) > 20 {
		result.AddWarning("name", "Long handler names may be harder to work with", CodeNameTooLong)
		result.AddSuggestion("name", abbreviateHandlerName(name), "Use abbreviated form", 0.6)
	}

	if isAllLowercase(name) && len(name) > 3 {
		result.AddWarning("name", "Consider using camelCase for multi-word names", CodeNameInvalidChars)
		result.AddSuggestion("name", toCamelCase(name), "Convert to camelCase", 0.8)
	}

	return result
}

// Helper functions for validation

func isValidGoIdentifier(s string) bool {
	if s == "" {
		return false
	}

	// Check if first character is a letter or underscore
	first := rune(s[0])
	if !((first >= 'a' && first <= 'z') || (first >= 'A' && first <= 'Z') || first == '_') {
		return false
	}

	// Check remaining characters
	for _, r := range s[1:] {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_') {
			return false
		}
	}

	return true
}

func isGoReservedWord(s string) bool {
	reserved := map[string]bool{
		"break": true, "case": true, "chan": true, "const": true, "continue": true,
		"default": true, "defer": true, "else": true, "fallthrough": true, "for": true,
		"func": true, "go": true, "goto": true, "if": true, "import": true,
		"interface": true, "map": true, "package": true, "range": true, "return": true,
		"select": true, "struct": true, "switch": true, "type": true, "var": true,
	}
	return reserved[s]
}

func sanitizeGoIdentifier(s string) string {
	result := ""
	for i, r := range s {
		if i == 0 {
			if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || r == '_' {
				result += string(r)
			} else {
				result += "_"
			}
		} else {
			if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' {
				result += string(r)
			} else if r == '-' {
				result += "_"
			}
			// Skip other invalid characters
		}
	}
	return result
}

func capitalize(s string) string {
	if s == "" {
		return s
	}
	return string(rune(s[0])-32) + s[1:]
}

func abbreviateHandlerName(name string) string {
	if len(name) <= 10 {
		return name
	}
	// Simple abbreviation - take first 10 chars
	return name[:10]
}

func isAllLowercase(s string) bool {
	for _, r := range s {
		if r >= 'A' && r <= 'Z' {
			return false
		}
	}
	return true
}

func toCamelCase(s string) string {
	// Simple camelCase conversion
	parts := strings.Split(s, "_")
	if len(parts) == 1 {
		parts = strings.Split(s, "-")
	}

	result := parts[0]
	for i := 1; i < len(parts); i++ {
		if len(parts[i]) > 0 {
			result += capitalize(parts[i])
		}
	}
	return result
}
