// internal/parser/validator.go
package parser

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// enhancedValidator implements rich validation with context and suggestions
type enhancedValidator struct {
	projectAnalyzer ProjectAnalyzer
}

// NewEnhancedValidator creates a new enhanced validator
func NewEnhancedValidator(projectAnalyzer ProjectAnalyzer) EnhancedValidator {
	return &enhancedValidator{
		projectAnalyzer: projectAnalyzer,
	}
}

// ValidateComponentName validates a component name (legacy interface)
func (v *enhancedValidator) ValidateComponentName(name string) error {
	result := v.validateComponentNameWithContext(name, nil)
	if !result.Valid {
		return fmt.Errorf(result.Errors[0].Message)
	}
	return nil
}

// ValidateComponentType validates a component type (legacy interface)
func (v *enhancedValidator) ValidateComponentType(componentType string) error {
	validTypes := map[string]bool{
		"handler":    true,
		"model":      true,
		"middleware": true,
		"database":   true,
	}

	if !validTypes[componentType] {
		return fmt.Errorf("unsupported component type: %s (valid types: handler, model, middleware, database)", componentType)
	}
	return nil
}

// ValidateProjectStructure validates project structure (legacy interface)
func (v *enhancedValidator) ValidateProjectStructure(projectRoot string) error {
	if !v.projectAnalyzer.IsGoProject(projectRoot) {
		return fmt.Errorf("not a valid Go project (go.mod not found)")
	}
	return nil
}

// ValidateAddCommandArgs validates add command arguments (legacy interface)
func (v *enhancedValidator) ValidateAddCommandArgs(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("add command requires component type and name")
	}

	componentType := args[0]
	componentName := args[1]

	if err := v.ValidateComponentType(componentType); err != nil {
		return err
	}

	if err := v.ValidateComponentName(componentName); err != nil {
		return err
	}

	return nil
}

// ValidateWithContext performs comprehensive validation with rich context
func (v *enhancedValidator) ValidateWithContext(ctx *ValidationContext) *ValidationResult {
	result := NewValidationResult()
	result.Context = ctx

	// Validate based on command type
	switch ctx.Command {
	case "add handler":
		v.validateHandlerCommand(ctx, result)
	case "add model":
		v.validateModelCommand(ctx, result)
	case "add middleware":
		v.validateMiddlewareCommand(ctx, result)
	case "add database":
		v.validateDatabaseCommand(ctx, result)
	case "wire handler":
		v.validateWireHandlerCommand(ctx, result)
	default:
		result.AddWarning("command", "Unknown command type", CodeArgsInvalid)
	}

	// Common validations
	v.validateProjectContext(ctx, result)
	v.validateEnvironment(ctx, result)

	return result
}

// SuggestCorrections provides suggestions for input corrections
func (v *enhancedValidator) SuggestCorrections(input string) []string {
	suggestions := []string{}

	// Common typos and corrections
	corrections := map[string][]string{
		"handelr":    {"handler"},
		"handlr":     {"handler"},
		"handeler":   {"handler"},
		"midleware":  {"middleware"},
		"middlware":  {"middleware"},
		"databse":    {"database"},
		"databas":    {"database"},
		"postgre":    {"postgres"},
		"postgresql": {"postgres"},
		"mongod":     {"mongodb"},
		"mongo":      {"mongodb"},
	}

	if corrected, exists := corrections[strings.ToLower(input)]; exists {
		suggestions = append(suggestions, corrected...)
	}

	// Suggest similar valid component types
	if isComponentTypeLike(input) {
		validTypes := []string{"handler", "model", "middleware", "database"}
		for _, validType := range validTypes {
			if similarity(input, validType) > 0.6 {
				suggestions = append(suggestions, validType)
			}
		}
	}

	// Suggest name improvements
	if len(input) > 0 {
		if !isValidGoIdentifier(input) {
			sanitized := sanitizeGoIdentifier(input)
			if sanitized != input {
				suggestions = append(suggestions, sanitized)
			}
		}

		if isGoReservedWord(input) {
			suggestions = append(suggestions, input+"Handler", "my"+capitalize(input))
		}
	}

	return suggestions
}

// Command-specific validation methods

func (v *enhancedValidator) validateHandlerCommand(ctx *ValidationContext, result *ValidationResult) {
	if len(ctx.Args) == 0 {
		result.AddError("args", "Handler name is required", CodeArgsInsufficient)
		result.AddSuggestion("args", "user", "A common handler name", 0.9)
		return
	}

	handlerName := ctx.Args[0]
	nameResult := v.validateComponentNameWithContext(handlerName, ctx)

	// Merge validation results
	result.Errors = append(result.Errors, nameResult.Errors...)
	result.Warnings = append(result.Warnings, nameResult.Warnings...)
	result.Suggestions = append(result.Suggestions, nameResult.Suggestions...)

	if len(nameResult.Errors) > 0 {
		result.Valid = false
	}

	// Handler-specific validations
	v.validateHandlerFlags(ctx, result)
	v.validateHandlerConflicts(ctx, result)
}

func (v *enhancedValidator) validateModelCommand(ctx *ValidationContext, result *ValidationResult) {
	if len(ctx.Args) == 0 {
		result.AddError("args", "Model name is required", CodeArgsInsufficient)
		result.AddSuggestion("args", "User", "A common model name", 0.9)
		return
	}

	modelName := ctx.Args[0]
	nameResult := v.validateComponentNameWithContext(modelName, ctx)

	// Merge validation results
	result.Errors = append(result.Errors, nameResult.Errors...)
	result.Warnings = append(result.Warnings, nameResult.Warnings...)
	result.Suggestions = append(result.Suggestions, nameResult.Suggestions...)

	if len(nameResult.Errors) > 0 {
		result.Valid = false
	}

	// Model-specific suggestions
	if !strings.HasSuffix(strings.ToLower(modelName), "model") && len(modelName) < 10 {
		result.AddSuggestion("name", modelName+"Model", "Consider adding 'Model' suffix", 0.6)
	}
}

func (v *enhancedValidator) validateMiddlewareCommand(ctx *ValidationContext, result *ValidationResult) {
	if len(ctx.Args) == 0 {
		result.AddError("args", "Middleware type is required", CodeArgsInsufficient)
		result.AddSuggestion("args", "auth", "Authentication middleware", 0.9)
		result.AddSuggestion("args", "cors", "CORS middleware", 0.8)
		return
	}

	middlewareType := ctx.Args[0]

	// Validate middleware type
	validMiddlewareTypes := map[string]bool{
		"auth":        true,
		"cors":        true,
		"ratelimit":   true,
		"logging":     true,
		"recovery":    true,
		"timeout":     true,
		"compression": true,
	}

	if !validMiddlewareTypes[middlewareType] {
		result.AddError("type", fmt.Sprintf("Unsupported middleware type: %s", middlewareType), CodeTypeUnsupported)

		// Suggest similar types
		for validType := range validMiddlewareTypes {
			if similarity(middlewareType, validType) > 0.6 {
				result.AddSuggestion("type", validType, fmt.Sprintf("Did you mean '%s'?", validType), 0.8)
			}
		}
	}
}

func (v *enhancedValidator) validateDatabaseCommand(ctx *ValidationContext, result *ValidationResult) {
	if len(ctx.Args) == 0 {
		result.AddError("args", "Database type is required", CodeArgsInsufficient)
		result.AddSuggestion("args", "postgres", "PostgreSQL database", 0.9)
		result.AddSuggestion("args", "mysql", "MySQL database", 0.8)
		return
	}

	dbType := ctx.Args[0]

	// Validate database type
	validDatabaseTypes := map[string]bool{
		"postgres": true,
		"mysql":    true,
		"sqlite":   true,
		"mongodb":  true,
	}

	if !validDatabaseTypes[dbType] {
		result.AddError("type", fmt.Sprintf("Unsupported database type: %s", dbType), CodeTypeUnsupported)

		// Suggest similar types
		for validType := range validDatabaseTypes {
			if similarity(dbType, validType) > 0.6 {
				result.AddSuggestion("type", validType, fmt.Sprintf("Did you mean '%s'?", validType), 0.8)
			}
		}
	}
}

func (v *enhancedValidator) validateWireHandlerCommand(ctx *ValidationContext, result *ValidationResult) {
	if len(ctx.Args) == 0 {
		result.AddError("args", "Handler name is required for wiring", CodeArgsInsufficient)
		return
	}

	handlerName := ctx.Args[0]

	// Check if handler exists
	if ctx.ProjectRoot != "" {
		handlerPath := filepath.Join(ctx.ProjectRoot, "internal", "handlers", handlerName+".go")
		if _, err := os.Stat(handlerPath); os.IsNotExist(err) {
			result.AddError("handler", fmt.Sprintf("Handler '%s' does not exist", handlerName), CodeNameExists)
			result.AddSuggestion("handler", "", fmt.Sprintf("Create handler first: foundry add handler %s", handlerName), 0.9)
		}
	}
}

// Helper validation methods

func (v *enhancedValidator) validateComponentNameWithContext(name string, ctx *ValidationContext) *ValidationResult {
	result := NewValidationResult()

	if name == "" {
		result.AddError("name", "Component name cannot be empty", CodeNameEmpty)
		return result
	}

	if len(name) < 2 {
		result.AddError("name", "Component name must be at least 2 characters", CodeNameTooShort)
		return result
	}

	if len(name) > 50 {
		result.AddError("name", "Component name must be less than 50 characters", CodeNameTooLong)
		return result
	}

	if !isValidGoIdentifier(name) {
		result.AddError("name", "Component name must be a valid Go identifier", CodeNameInvalidChars)
		sanitized := sanitizeGoIdentifier(name)
		if sanitized != name {
			result.AddSuggestion("name", sanitized, "Use valid Go identifier", 0.9)
		}
	}

	if isGoReservedWord(name) {
		result.AddError("name", fmt.Sprintf("'%s' is a Go reserved word", name), CodeNameReservedWord)
		result.AddSuggestion("name", name+"Component", "Add 'Component' suffix", 0.8)
	}

	return result
}

func (v *enhancedValidator) validateHandlerFlags(ctx *ValidationContext, result *ValidationResult) {
	// Validate type flag if present
	if handlerType, exists := ctx.Flags["type"]; exists {
		if typeStr, ok := handlerType.(string); ok {
			validTypes := []string{"REST", "GraphQL", "gRPC"}
			isValid := false
			for _, validType := range validTypes {
				if strings.EqualFold(typeStr, validType) {
					isValid = true
					break
				}
			}
			if !isValid {
				result.AddWarning("type", fmt.Sprintf("Unknown handler type: %s", typeStr), CodeTypeInvalid)
				for _, validType := range validTypes {
					result.AddSuggestion("type", validType, fmt.Sprintf("Use %s handler type", validType), 0.7)
				}
			}
		}
	}

	// Validate flag combinations
	if autoWire, exists := ctx.Flags["auto-wire"]; exists {
		if dryRun, dryExists := ctx.Flags["dry-run"]; dryExists {
			if autoWire.(bool) && dryRun.(bool) {
				result.AddWarning("flags", "Auto-wire has no effect in dry-run mode", CodeFlagConflict)
				result.AddSuggestion("flags", "Remove --auto-wire or --dry-run", "Choose one option", 0.8)
			}
		}
	}
}

func (v *enhancedValidator) validateHandlerConflicts(ctx *ValidationContext, result *ValidationResult) {
	if len(ctx.Args) == 0 || ctx.ProjectRoot == "" {
		return
	}

	handlerName := ctx.Args[0]
	handlerPath := filepath.Join(ctx.ProjectRoot, "internal", "handlers", handlerName+".go")

	if _, err := os.Stat(handlerPath); err == nil {
		result.AddError("conflict", fmt.Sprintf("Handler '%s' already exists", handlerName), CodeNameExists)
		result.AddSuggestion("name", handlerName+"2", "Try with number suffix", 0.7)
		result.AddSuggestion("name", handlerName+"New", "Try with 'New' suffix", 0.6)
	}
}

func (v *enhancedValidator) validateProjectContext(ctx *ValidationContext, result *ValidationResult) {
	if ctx.ProjectRoot == "" {
		result.AddWarning("project", "Could not determine project root", CodeProjectNotFound)
		return
	}

	if !v.projectAnalyzer.IsGoProject(ctx.ProjectRoot) {
		result.AddError("project", "Not a valid Go project (go.mod not found)", CodeProjectNotGo)
		result.AddSuggestion("project", "Run 'go mod init <module-name>' first", "Initialize Go module", 0.9)
	}
}

func (v *enhancedValidator) validateEnvironment(ctx *ValidationContext, result *ValidationResult) {
	// Check for common environment issues
	if gopath := ctx.Environment["GOPATH"]; gopath == "" {
		result.AddWarning("environment", "GOPATH not set (may be intentional with Go modules)", CodeProjectInvalid)
	}

	if goroot := ctx.Environment["GOROOT"]; goroot == "" {
		result.AddWarning("environment", "GOROOT not set", CodeProjectInvalid)
	}
}

// Utility functions

func isComponentTypeLike(input string) bool {
	componentWords := []string{"handle", "model", "middle", "data", "wire"}
	inputLower := strings.ToLower(input)

	for _, word := range componentWords {
		if strings.Contains(inputLower, word) {
			return true
		}
	}
	return false
}

func similarity(a, b string) float64 {
	// Simple similarity calculation based on common characters
	if len(a) == 0 || len(b) == 0 {
		return 0.0
	}

	aLower := strings.ToLower(a)
	bLower := strings.ToLower(b)

	// Count common characters
	common := 0
	maxLen := len(aLower)
	if len(bLower) > maxLen {
		maxLen = len(bLower)
	}

	for i := 0; i < len(aLower) && i < len(bLower); i++ {
		if aLower[i] == bLower[i] {
			common++
		}
	}

	return float64(common) / float64(maxLen)
}
