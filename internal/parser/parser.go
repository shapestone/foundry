// internal/parser/parser.go
package parser

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/shapestone/foundry/internal/scaffolder"
)

// parser is the main implementation of the Parser interface
type parser struct {
	handlerParser   *handlerParser
	validator       EnhancedValidator
	projectAnalyzer ProjectAnalyzer
	flagExtractor   FlagExtractor
}

// New creates a new parser with all dependencies
func New(projectAnalyzer ProjectAnalyzer, flagExtractor FlagExtractor) Parser {
	validator := NewEnhancedValidator(projectAnalyzer)

	return &parser{
		handlerParser: &handlerParser{
			validator:       validator,
			projectAnalyzer: projectAnalyzer,
			flagExtractor:   flagExtractor,
		},
		validator:       validator,
		projectAnalyzer: projectAnalyzer,
		flagExtractor:   flagExtractor,
	}
}

// ParseHandlerCommand parses handler command input
func (p *parser) ParseHandlerCommand(ctx context.Context, args []string, flags map[string]interface{}) (*scaffolder.HandlerSpec, error) {
	return p.handlerParser.ParseHandlerCommand(ctx, args, flags)
}

// ParseModelCommand parses model command input
func (p *parser) ParseModelCommand(ctx context.Context, args []string, flags map[string]interface{}) (*scaffolder.ModelSpec, error) {
	// Validate basic arguments
	if len(args) == 0 {
		return nil, &ParseError{
			Command:     "model",
			Args:        args,
			Flags:       flags,
			Message:     "model name is required",
			Code:        string(CodeArgsInsufficient),
			Suggestions: []string{"Provide a model name", "Example: foundry add model User"},
		}
	}

	modelName := args[0]

	// Create validation context
	validationCtx := &ValidationContext{
		Command:     "add model",
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
	includeTimestamps := p.flagExtractor.ExtractBool(flags, "timestamps", true)
	includeValidation := p.flagExtractor.ExtractBool(flags, "validation", true)

	// Get project information
	projectRoot := validationCtx.ProjectRoot
	moduleName, err := p.projectAnalyzer.GetModuleName(projectRoot)
	if err != nil {
		return nil, &ParseError{
			Command:     "model",
			Args:        args,
			Flags:       flags,
			Message:     "failed to determine module name",
			Code:        string(CodeProjectInvalid),
			Cause:       err,
			Suggestions: []string{"Ensure you're in a valid Go project", "Check that go.mod exists"},
		}
	}

	// Parse field specifications from flags if provided
	fields := []scaffolder.FieldSpec{}
	if fieldSpecs := p.flagExtractor.ExtractStringSlice(flags, "fields", []string{}); len(fieldSpecs) > 0 {
		for _, fieldSpec := range fieldSpecs {
			field, err := p.parseFieldSpec(fieldSpec)
			if err != nil {
				return nil, &ParseError{
					Command:     "model",
					Args:        args,
					Flags:       flags,
					Message:     fmt.Sprintf("invalid field specification: %s", err.Error()),
					Code:        string(CodeArgsInvalid),
					Suggestions: []string{"Use format: name:type", "Example: --fields name:string,age:int"},
				}
			}
			fields = append(fields, *field)
		}
	}

	// Create metadata
	metadata := make(map[string]string)
	metadata["command"] = "add model"
	metadata["parser_version"] = "2.0"

	// Build the model specification
	spec := &scaffolder.ModelSpec{
		Name:              modelName,
		Fields:            fields,
		IncludeTimestamps: includeTimestamps,
		IncludeValidation: includeValidation,
		ProjectRoot:       projectRoot,
		Module:            moduleName,
		Metadata:          metadata,
	}

	return spec, nil
}

// ParseMiddlewareCommand parses middleware command input
func (p *parser) ParseMiddlewareCommand(ctx context.Context, args []string, flags map[string]interface{}) (*scaffolder.MiddlewareSpec, error) {
	// Validate basic arguments
	if len(args) == 0 {
		return nil, &ParseError{
			Command:     "middleware",
			Args:        args,
			Flags:       flags,
			Message:     "middleware type is required",
			Code:        string(CodeArgsInsufficient),
			Suggestions: []string{"Provide a middleware type", "Example: foundry add middleware auth"},
		}
	}

	middlewareType := args[0]

	// Create validation context
	validationCtx := &ValidationContext{
		Command:     "add middleware",
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
	autoWire := p.flagExtractor.ExtractBool(flags, "auto-wire", false)

	// Get project information
	projectRoot := validationCtx.ProjectRoot
	moduleName, err := p.projectAnalyzer.GetModuleName(projectRoot)
	if err != nil {
		return nil, &ParseError{
			Command:     "middleware",
			Args:        args,
			Flags:       flags,
			Message:     "failed to determine module name",
			Code:        string(CodeProjectInvalid),
			Cause:       err,
			Suggestions: []string{"Ensure you're in a valid Go project", "Check that go.mod exists"},
		}
	}

	// Create metadata
	metadata := make(map[string]string)
	metadata["command"] = "add middleware"
	metadata["parser_version"] = "2.0"

	// Build the middleware specification
	spec := &scaffolder.MiddlewareSpec{
		Name:        middlewareType,
		Type:        middlewareType,
		AutoWire:    autoWire,
		ProjectRoot: projectRoot,
		Module:      moduleName,
		Metadata:    metadata,
	}

	return spec, nil
}

// ParseDatabaseCommand parses database command input
func (p *parser) ParseDatabaseCommand(ctx context.Context, args []string, flags map[string]interface{}) (*scaffolder.DatabaseSpec, error) {
	// Validate basic arguments
	if len(args) == 0 {
		return nil, &ParseError{
			Command:     "database",
			Args:        args,
			Flags:       flags,
			Message:     "database type is required",
			Code:        string(CodeArgsInsufficient),
			Suggestions: []string{"Provide a database type", "Example: foundry add db postgres"},
		}
	}

	dbType := args[0]

	// Create validation context
	validationCtx := &ValidationContext{
		Command:     "add database",
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
	withMigrations := p.flagExtractor.ExtractBool(flags, "with-migrations", false)
	withDocker := p.flagExtractor.ExtractBool(flags, "with-docker", false)

	// Get project information
	projectRoot := validationCtx.ProjectRoot
	moduleName, err := p.projectAnalyzer.GetModuleName(projectRoot)
	if err != nil {
		return nil, &ParseError{
			Command:     "database",
			Args:        args,
			Flags:       flags,
			Message:     "failed to determine module name",
			Code:        string(CodeProjectInvalid),
			Cause:       err,
			Suggestions: []string{"Ensure you're in a valid Go project", "Check that go.mod exists"},
		}
	}

	// Create metadata
	metadata := make(map[string]string)
	metadata["command"] = "add database"
	metadata["parser_version"] = "2.0"

	// Build the database specification
	spec := &scaffolder.DatabaseSpec{
		Type:           dbType,
		WithMigrations: withMigrations,
		WithDocker:     withDocker,
		ProjectRoot:    projectRoot,
		Module:         moduleName,
		Metadata:       metadata,
	}

	return spec, nil
}

// ParseWireCommand parses wire command input
func (p *parser) ParseWireCommand(ctx context.Context, args []string, flags map[string]interface{}) (*scaffolder.WireSpec, error) {
	// Wire commands need at least component type and name
	if len(args) < 2 {
		return nil, &ParseError{
			Command:     "wire",
			Args:        args,
			Flags:       flags,
			Message:     "wire command requires component type and name",
			Code:        string(CodeArgsInsufficient),
			Suggestions: []string{"Provide component type and name", "Example: foundry wire handler user"},
		}
	}

	componentType := args[0]
	componentName := args[1]

	// Create validation context
	validationCtx := &ValidationContext{
		Command:     fmt.Sprintf("wire %s", componentType),
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

	// Get project information
	projectRoot := validationCtx.ProjectRoot
	moduleName, err := p.projectAnalyzer.GetModuleName(projectRoot)
	if err != nil {
		return nil, &ParseError{
			Command:     "wire",
			Args:        args,
			Flags:       flags,
			Message:     "failed to determine module name",
			Code:        string(CodeProjectInvalid),
			Cause:       err,
			Suggestions: []string{"Ensure you're in a valid Go project", "Check that go.mod exists"},
		}
	}

	// Create metadata
	metadata := make(map[string]string)
	metadata["command"] = "wire"
	metadata["parser_version"] = "2.0"

	// Build the wire specification
	spec := &scaffolder.WireSpec{
		ComponentType: componentType,
		ComponentName: componentName,
		ProjectRoot:   projectRoot,
		Module:        moduleName,
		Metadata:      metadata,
	}

	return spec, nil
}

// Helper methods

func (p *parser) getProjectRoot() string {
	if root, err := p.projectAnalyzer.GetProjectRoot(); err == nil {
		return root
	}
	// Fallback to current directory
	if cwd, err := os.Getwd(); err == nil {
		return cwd
	}
	return "."
}

func (p *parser) getWorkingDir() string {
	if cwd, err := os.Getwd(); err == nil {
		return cwd
	}
	return "."
}

func (p *parser) createParseErrorFromValidation(result *ValidationResult) error {
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

func (p *parser) parseFieldSpec(fieldSpec string) (*scaffolder.FieldSpec, error) {
	// Parse field specification in format "name:type" or "name:type:tag1,tag2"
	parts := strings.Split(fieldSpec, ":")
	if len(parts) < 2 {
		return nil, fmt.Errorf("field specification must be in format 'name:type'")
	}

	field := &scaffolder.FieldSpec{
		Name: strings.TrimSpace(parts[0]),
		Type: strings.TrimSpace(parts[1]),
		Tags: make(map[string]string),
	}

	// Validate field name
	if field.Name == "" {
		return nil, fmt.Errorf("field name cannot be empty")
	}

	if !isValidGoIdentifier(field.Name) {
		return nil, fmt.Errorf("field name must be a valid Go identifier")
	}

	// Validate field type
	if field.Type == "" {
		return nil, fmt.Errorf("field type cannot be empty")
	}

	// Parse tags if provided
	if len(parts) > 2 {
		tagsPart := strings.TrimSpace(parts[2])
		if tagsPart != "" {
			// Simple tag parsing - could be enhanced later
			field.Tags["json"] = field.Name
			if strings.Contains(tagsPart, "required") {
				field.Required = true
			}
		}
	}

	return field, nil
}
