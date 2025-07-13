// internal/parser/interfaces.go
package parser

import (
	"context"

	"github.com/shapestone/foundry/internal/scaffolder"
)

// Parser is the main interface for parsing CLI input into scaffolder specifications
type Parser interface {
	ParseHandlerCommand(ctx context.Context, args []string, flags map[string]interface{}) (*scaffolder.HandlerSpec, error)
	ParseModelCommand(ctx context.Context, args []string, flags map[string]interface{}) (*scaffolder.ModelSpec, error)
	ParseMiddlewareCommand(ctx context.Context, args []string, flags map[string]interface{}) (*scaffolder.MiddlewareSpec, error)
	ParseDatabaseCommand(ctx context.Context, args []string, flags map[string]interface{}) (*scaffolder.DatabaseSpec, error)
	ParseWireCommand(ctx context.Context, args []string, flags map[string]interface{}) (*scaffolder.WireSpec, error)
}

// Validator provides validation for CLI inputs
type Validator interface {
	ValidateComponentName(name string) error
	ValidateComponentType(componentType string) error
	ValidateProjectStructure(projectRoot string) error
	ValidateAddCommandArgs(args []string) error
}

// EnhancedValidator provides rich validation with context and suggestions
type EnhancedValidator interface {
	Validator
	ValidateWithContext(ctx *ValidationContext) *ValidationResult
	SuggestCorrections(input string) []string
}

// ValidationContext provides context for validation
type ValidationContext struct {
	Command     string                 `json:"command"`
	Args        []string               `json:"args"`
	Flags       map[string]interface{} `json:"flags"`
	ProjectRoot string                 `json:"project_root"`
	WorkingDir  string                 `json:"working_dir"`
	Environment map[string]string      `json:"environment"`
	UserContext map[string]interface{} `json:"user_context"`
}

// ValidationResult represents the result of validation
type ValidationResult struct {
	Valid       bool                   `json:"valid"`
	Errors      []ValidationError      `json:"errors"`
	Warnings    []ValidationWarning    `json:"warnings"`
	Suggestions []ValidationSuggestion `json:"suggestions"`
	Context     *ValidationContext     `json:"context"`
}

// ValidationError represents a validation error with rich context
type ValidationError struct {
	Field       string                 `json:"field"`
	Value       interface{}            `json:"value"`
	Message     string                 `json:"message"`
	Code        string                 `json:"code"`
	Severity    ValidationSeverity     `json:"severity"`
	Suggestions []string               `json:"suggestions"`
	Context     map[string]interface{} `json:"context"`
}

// ValidationWarning represents a validation warning
type ValidationWarning struct {
	Field       string                 `json:"field"`
	Value       interface{}            `json:"value"`
	Message     string                 `json:"message"`
	Code        string                 `json:"code"`
	Suggestions []string               `json:"suggestions"`
	Context     map[string]interface{} `json:"context"`
}

// ValidationSuggestion represents a suggestion for fixing validation issues
type ValidationSuggestion struct {
	Field      string  `json:"field"`
	Suggestion string  `json:"suggestion"`
	Reason     string  `json:"reason"`
	Confidence float64 `json:"confidence"` // 0.0 to 1.0
	AutoFix    bool    `json:"auto_fix"`   // Can this be automatically fixed?
	FixCommand string  `json:"fix_command,omitempty"`
}

// ValidationSeverity defines the severity level of validation issues
type ValidationSeverity string

const (
	SeverityError   ValidationSeverity = "error"
	SeverityWarning ValidationSeverity = "warning"
	SeverityInfo    ValidationSeverity = "info"
	SeverityHint    ValidationSeverity = "hint"
)

// ValidationCode represents specific validation error codes
type ValidationCode string

const (
	// Component name validation codes
	CodeNameEmpty        ValidationCode = "NAME_EMPTY"
	CodeNameTooShort     ValidationCode = "NAME_TOO_SHORT"
	CodeNameTooLong      ValidationCode = "NAME_TOO_LONG"
	CodeNameInvalidChars ValidationCode = "NAME_INVALID_CHARS"
	CodeNameReservedWord ValidationCode = "NAME_RESERVED_WORD"
	CodeNameExists       ValidationCode = "NAME_EXISTS"

	// Project validation codes
	CodeProjectNotFound ValidationCode = "PROJECT_NOT_FOUND"
	CodeProjectNotGo    ValidationCode = "PROJECT_NOT_GO"
	CodeProjectInvalid  ValidationCode = "PROJECT_INVALID"

	// Argument validation codes
	CodeArgsInsufficient ValidationCode = "ARGS_INSUFFICIENT"
	CodeArgsExcess       ValidationCode = "ARGS_EXCESS"
	CodeArgsInvalid      ValidationCode = "ARGS_INVALID"

	// Flag validation codes
	CodeFlagInvalid  ValidationCode = "FLAG_INVALID"
	CodeFlagConflict ValidationCode = "FLAG_CONFLICT"
	CodeFlagRequired ValidationCode = "FLAG_REQUIRED"

	// Type validation codes
	CodeTypeUnsupported ValidationCode = "TYPE_UNSUPPORTED"
	CodeTypeInvalid     ValidationCode = "TYPE_INVALID"
)

// ProjectAnalyzer provides project structure analysis
type ProjectAnalyzer interface {
	GetProjectRoot() (string, error)
	GetModuleName(projectRoot string) (string, error)
	GetProjectName(projectRoot string) (string, error)
	IsGoProject(projectRoot string) bool
	GetExistingComponents(projectRoot string) (*ProjectComponents, error)
}

// ProjectComponents represents existing components in a project
type ProjectComponents struct {
	Handlers   []string `json:"handlers"`
	Models     []string `json:"models"`
	Middleware []string `json:"middleware"`
	Routes     []string `json:"routes"`
	Databases  []string `json:"databases"`
}

// FlagExtractor extracts and validates flags from command input
type FlagExtractor interface {
	ExtractString(flags map[string]interface{}, key string, defaultValue string) string
	ExtractBool(flags map[string]interface{}, key string, defaultValue bool) bool
	ExtractInt(flags map[string]interface{}, key string, defaultValue int) int
	ExtractStringSlice(flags map[string]interface{}, key string, defaultValue []string) []string
	ValidateRequiredFlags(flags map[string]interface{}, required []string) error
}

// ParseError represents errors during parsing
type ParseError struct {
	Command     string                 `json:"command"`
	Args        []string               `json:"args"`
	Flags       map[string]interface{} `json:"flags"`
	Message     string                 `json:"message"`
	Code        string                 `json:"code"`
	Suggestions []string               `json:"suggestions"`
	Cause       error                  `json:"cause,omitempty"`
}

func (e *ParseError) Error() string {
	if e.Cause != nil {
		return e.Message + ": " + e.Cause.Error()
	}
	return e.Message
}

func (e *ParseError) Unwrap() error {
	return e.Cause
}

// ParseErrors represents multiple parsing errors
type ParseErrors []*ParseError

func (e ParseErrors) Error() string {
	if len(e) == 0 {
		return ""
	}
	if len(e) == 1 {
		return e[0].Error()
	}
	return e[0].Error() + " (and " + string(rune(len(e)-1)) + " more errors)"
}

// Helper functions for validation results

// NewValidationResult creates a new validation result
func NewValidationResult() *ValidationResult {
	return &ValidationResult{
		Valid:       true,
		Errors:      []ValidationError{},
		Warnings:    []ValidationWarning{},
		Suggestions: []ValidationSuggestion{},
	}
}

// AddError adds a validation error
func (r *ValidationResult) AddError(field, message string, code ValidationCode) {
	r.Valid = false
	r.Errors = append(r.Errors, ValidationError{
		Field:    field,
		Message:  message,
		Code:     string(code),
		Severity: SeverityError,
		Context:  make(map[string]interface{}),
	})
}

// AddWarning adds a validation warning
func (r *ValidationResult) AddWarning(field, message string, code ValidationCode) {
	r.Warnings = append(r.Warnings, ValidationWarning{
		Field:   field,
		Message: message,
		Code:    string(code),
		Context: make(map[string]interface{}),
	})
}

// AddSuggestion adds a validation suggestion
func (r *ValidationResult) AddSuggestion(field, suggestion, reason string, confidence float64) {
	r.Suggestions = append(r.Suggestions, ValidationSuggestion{
		Field:      field,
		Suggestion: suggestion,
		Reason:     reason,
		Confidence: confidence,
		AutoFix:    confidence > 0.8, // High confidence suggestions can be auto-fixed
	})
}

// HasErrors returns true if there are validation errors
func (r *ValidationResult) HasErrors() bool {
	return len(r.Errors) > 0
}

// HasWarnings returns true if there are validation warnings
func (r *ValidationResult) HasWarnings() bool {
	return len(r.Warnings) > 0
}
