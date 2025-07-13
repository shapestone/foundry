// internal/parser/adapters.go
package parser

import (
	"context"
	"fmt"
	"github.com/shapestone/foundry/internal/scaffolder"
	"os"
	"path/filepath"
	"strings"

	"github.com/shapestone/foundry/internal/project"
)

// Adapters to integrate existing components with the parser interfaces

// ProjectAnalyzerAdapter adapts the existing project package
type ProjectAnalyzerAdapter struct{}

func NewProjectAnalyzerAdapter() ProjectAnalyzer {
	return &ProjectAnalyzerAdapter{}
}

func (p *ProjectAnalyzerAdapter) GetProjectRoot() (string, error) {
	// Start from current directory and walk up to find go.mod
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		goModPath := filepath.Join(dir, "go.mod")
		if _, err := os.Stat(goModPath); err == nil {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached root directory
			break
		}
		dir = parent
	}

	// Fallback to current directory
	if cwd, err := os.Getwd(); err == nil {
		return cwd, nil
	}

	return "", fmt.Errorf("could not determine project root")
}

func (p *ProjectAnalyzerAdapter) GetModuleName(projectRoot string) (string, error) {
	// Change to project root temporarily
	originalDir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	defer os.Chdir(originalDir)

	if err := os.Chdir(projectRoot); err != nil {
		return "", err
	}

	return project.GetCurrentModule(), nil
}

func (p *ProjectAnalyzerAdapter) GetProjectName(projectRoot string) (string, error) {
	// Change to project root temporarily
	originalDir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	defer os.Chdir(originalDir)

	if err := os.Chdir(projectRoot); err != nil {
		return "", err
	}

	return project.GetProjectName(), nil
}

func (p *ProjectAnalyzerAdapter) IsGoProject(projectRoot string) bool {
	goModPath := filepath.Join(projectRoot, "go.mod")
	_, err := os.Stat(goModPath)
	return err == nil
}

func (p *ProjectAnalyzerAdapter) GetExistingComponents(projectRoot string) (*ProjectComponents, error) {
	components := &ProjectComponents{
		Handlers:   []string{},
		Models:     []string{},
		Middleware: []string{},
		Routes:     []string{},
		Databases:  []string{},
	}

	// Get existing handlers
	handlerDir := filepath.Join(projectRoot, "internal", "handlers")
	if handlers, err := p.getGoFilesInDir(handlerDir); err == nil {
		components.Handlers = handlers
	}

	// Get existing models
	modelDir := filepath.Join(projectRoot, "internal", "models")
	if models, err := p.getGoFilesInDir(modelDir); err == nil {
		components.Models = models
	}

	// Get existing middleware
	middlewareDir := filepath.Join(projectRoot, "internal", "middleware")
	if middleware, err := p.getGoFilesInDir(middlewareDir); err == nil {
		components.Middleware = middleware
	}

	// Get existing routes
	routeDir := filepath.Join(projectRoot, "internal", "routes")
	if routes, err := p.getGoFilesInDir(routeDir); err == nil {
		components.Routes = routes
	}

	// Get existing database configurations
	dbDir := filepath.Join(projectRoot, "internal", "database")
	if databases, err := p.getGoFilesInDir(dbDir); err == nil {
		components.Databases = databases
	}

	return components, nil
}

func (p *ProjectAnalyzerAdapter) getGoFilesInDir(dir string) ([]string, error) {
	files := []string{}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return files, err
	}

	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".go") {
			// Remove .go extension to get component name
			name := strings.TrimSuffix(entry.Name(), ".go")
			files = append(files, name)
		}
	}

	return files, nil
}

// FlagExtractorAdapter provides flag extraction with type safety
type FlagExtractorAdapter struct{}

func NewFlagExtractorAdapter() FlagExtractor {
	return &FlagExtractorAdapter{}
}

func (f *FlagExtractorAdapter) ExtractString(flags map[string]interface{}, key string, defaultValue string) string {
	if value, exists := flags[key]; exists {
		if str, ok := value.(string); ok {
			return str
		}
	}
	return defaultValue
}

func (f *FlagExtractorAdapter) ExtractBool(flags map[string]interface{}, key string, defaultValue bool) bool {
	if value, exists := flags[key]; exists {
		if b, ok := value.(bool); ok {
			return b
		}
	}
	return defaultValue
}

func (f *FlagExtractorAdapter) ExtractInt(flags map[string]interface{}, key string, defaultValue int) int {
	if value, exists := flags[key]; exists {
		if i, ok := value.(int); ok {
			return i
		}
		// Try to convert from int64 (common in CLI parsing)
		if i64, ok := value.(int64); ok {
			return int(i64)
		}
	}
	return defaultValue
}

func (f *FlagExtractorAdapter) ExtractStringSlice(flags map[string]interface{}, key string, defaultValue []string) []string {
	if value, exists := flags[key]; exists {
		if slice, ok := value.([]string); ok {
			return slice
		}
		// Try to convert from string with comma separation
		if str, ok := value.(string); ok {
			if str == "" {
				return defaultValue
			}
			return strings.Split(str, ",")
		}
	}
	return defaultValue
}

func (f *FlagExtractorAdapter) ValidateRequiredFlags(flags map[string]interface{}, required []string) error {
	missing := []string{}

	for _, flag := range required {
		if _, exists := flags[flag]; !exists {
			missing = append(missing, flag)
		}
	}

	if len(missing) > 0 {
		return fmt.Errorf("missing required flags: %s", strings.Join(missing, ", "))
	}

	return nil
}

// Factory functions

// NewParser creates a new parser with real adapters
func NewParser() Parser {
	projectAnalyzer := NewProjectAnalyzerAdapter()
	flagExtractor := NewFlagExtractorAdapter()
	return New(projectAnalyzer, flagExtractor)
}

// NewParserForTesting creates a new parser with mock dependencies
func NewParserForTesting(projectAnalyzer ProjectAnalyzer, flagExtractor FlagExtractor) Parser {
	return New(projectAnalyzer, flagExtractor)
}

// Helper functions for command integration

// CobraFlagsToMap converts cobra command flags to a map for parser consumption
func CobraFlagsToMap(flags interface{}) map[string]interface{} {
	// This would be implemented based on your specific cobra setup
	// For now, return an empty map as placeholder
	flagMap := make(map[string]interface{})

	// TODO: Implement actual flag extraction from cobra command
	// This would typically involve introspecting the cobra command's flags
	// and converting them to the map format expected by the parser

	return flagMap
}

// ValidateAndParseHandler is a convenience function for handler parsing
func ValidateAndParseHandler(args []string, flags map[string]interface{}) (*scaffolder.HandlerSpec, error) {
	parser := NewParser()
	return parser.ParseHandlerCommand(context.Background(), args, flags)
}

// ValidateAndParseModel is a convenience function for model parsing
func ValidateAndParseModel(args []string, flags map[string]interface{}) (*scaffolder.ModelSpec, error) {
	parser := NewParser()
	return parser.ParseModelCommand(context.Background(), args, flags)
}

// ValidateAndParseMiddleware is a convenience function for middleware parsing
func ValidateAndParseMiddleware(args []string, flags map[string]interface{}) (*scaffolder.MiddlewareSpec, error) {
	parser := NewParser()
	return parser.ParseMiddlewareCommand(context.Background(), args, flags)
}

// ValidateAndParseDatabase is a convenience function for database parsing
func ValidateAndParseDatabase(args []string, flags map[string]interface{}) (*scaffolder.DatabaseSpec, error) {
	parser := NewParser()
	return parser.ParseDatabaseCommand(context.Background(), args, flags)
}

// ValidateAndParseWire is a convenience function for wire parsing
func ValidateAndParseWire(args []string, flags map[string]interface{}) (*scaffolder.WireSpec, error) {
	parser := NewParser()
	return parser.ParseWireCommand(context.Background(), args, flags)
}

// ErrorFormatter provides rich error formatting for CLI output
type ErrorFormatter struct{}

// FormatParseError formats a parse error for user-friendly CLI output
func (f *ErrorFormatter) FormatParseError(err error) string {
	var output strings.Builder

	if parseErr, ok := err.(*ParseError); ok {
		output.WriteString(fmt.Sprintf("❌ Error: %s\n", parseErr.Message))

		if len(parseErr.Suggestions) > 0 {
			output.WriteString("\n💡 Suggestions:\n")
			for _, suggestion := range parseErr.Suggestions {
				output.WriteString(fmt.Sprintf("  • %s\n", suggestion))
			}
		}

		if parseErr.Code != "" {
			output.WriteString(fmt.Sprintf("\n🔍 Error Code: %s\n", parseErr.Code))
		}
	} else if parseErrs, ok := err.(ParseErrors); ok {
		output.WriteString(fmt.Sprintf("❌ Multiple errors found:\n"))

		for i, parseErr := range parseErrs {
			output.WriteString(fmt.Sprintf("\n%d. %s", i+1, parseErr.Message))
			if len(parseErr.Suggestions) > 0 {
				output.WriteString("\n   💡 Suggestions:")
				for _, suggestion := range parseErr.Suggestions {
					output.WriteString(fmt.Sprintf("\n     • %s", suggestion))
				}
			}
		}
	} else {
		output.WriteString(fmt.Sprintf("❌ Error: %s\n", err.Error()))
	}

	return output.String()
}

// FormatValidationResult formats a validation result for user-friendly CLI output
func (f *ErrorFormatter) FormatValidationResult(result *ValidationResult) string {
	var output strings.Builder

	if result.Valid {
		output.WriteString("✅ Validation passed\n")

		if len(result.Warnings) > 0 {
			output.WriteString("\n⚠️  Warnings:\n")
			for _, warning := range result.Warnings {
				output.WriteString(fmt.Sprintf("  • %s\n", warning.Message))
			}
		}

		if len(result.Suggestions) > 0 {
			output.WriteString("\n💡 Suggestions for improvement:\n")
			for _, suggestion := range result.Suggestions {
				confidence := fmt.Sprintf("%.0f%%", suggestion.Confidence*100)
				output.WriteString(fmt.Sprintf("  • %s (%s confidence)\n", suggestion.Suggestion, confidence))
			}
		}
	} else {
		output.WriteString("❌ Validation failed\n")

		if len(result.Errors) > 0 {
			output.WriteString("\n🚫 Errors:\n")
			for _, err := range result.Errors {
				output.WriteString(fmt.Sprintf("  • %s\n", err.Message))
				if len(err.Suggestions) > 0 {
					for _, suggestion := range err.Suggestions {
						output.WriteString(fmt.Sprintf("    💡 %s\n", suggestion))
					}
				}
			}
		}

		if len(result.Warnings) > 0 {
			output.WriteString("\n⚠️  Warnings:\n")
			for _, warning := range result.Warnings {
				output.WriteString(fmt.Sprintf("  • %s\n", warning.Message))
			}
		}
	}

	return output.String()
}

// NewErrorFormatter creates a new error formatter
func NewErrorFormatter() *ErrorFormatter {
	return &ErrorFormatter{}
}
