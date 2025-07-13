package templating

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"text/template"
)

// DefaultRenderer implements template rendering with common functionality
type DefaultRenderer struct {
	funcMap template.FuncMap
}

// NewDefaultRenderer creates a new default template renderer
func NewDefaultRenderer() TemplateRenderer {
	return &DefaultRenderer{
		funcMap: createDefaultFuncMap(),
	}
}

// Render renders a template with the provided data
func (r *DefaultRenderer) Render(tmpl *Template, data TemplateData) (string, error) {
	var buf bytes.Buffer

	if err := r.RenderToWriter(tmpl, data, &buf); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// RenderToWriter renders a template directly to a writer
func (r *DefaultRenderer) RenderToWriter(tmpl *Template, data TemplateData, w io.Writer) error {
	if tmpl.Parsed == nil {
		return NewTemplateError("render", tmpl.Name, tmpl.Layout,
			fmt.Errorf("template not parsed"))
	}

	// Add function map to template if not already set
	if tmpl.Parsed.Funcs(r.funcMap) == nil {
		return NewTemplateError("render", tmpl.Name, tmpl.Layout,
			fmt.Errorf("failed to add function map"))
	}

	// Execute template
	if err := tmpl.Parsed.Execute(w, data); err != nil {
		return NewTemplateError("render", tmpl.Name, tmpl.Layout, err)
	}

	return nil
}

// ValidateTemplate validates template syntax
func (r *DefaultRenderer) ValidateTemplate(tmpl *Template) error {
	if tmpl.Parsed == nil {
		// Try to parse the template content
		parsed, err := template.New(tmpl.Name).Funcs(r.funcMap).Parse(tmpl.Content)
		if err != nil {
			return NewTemplateError("parse", tmpl.Name, tmpl.Layout, err)
		}
		tmpl.Parsed = parsed
	}

	// Test render with empty data to validate syntax
	var buf bytes.Buffer
	testData := make(TemplateData)

	if err := tmpl.Parsed.Execute(&buf, testData); err != nil {
		return NewTemplateError("validate", tmpl.Name, tmpl.Layout, err)
	}

	return nil
}

// createDefaultFuncMap creates the default function map for templates
func createDefaultFuncMap() template.FuncMap {
	return template.FuncMap{
		// String manipulation functions
		"title":      strings.Title,
		"lower":      strings.ToLower,
		"upper":      strings.ToUpper,
		"camelCase":  toCamelCase,
		"snakeCase":  toSnakeCase,
		"kebabCase":  toKebabCase,
		"pascalCase": toPascalCase,

		// Pluralization functions
		"plural":   pluralize,
		"singular": singularize,

		// Path and import functions
		"joinPath":   joinPath,
		"cleanPath":  cleanPath,
		"importPath": generateImportPath,

		// Conditional functions
		"eq":  func(a, b interface{}) bool { return a == b },
		"ne":  func(a, b interface{}) bool { return a != b },
		"not": func(b bool) bool { return !b },

		// String testing functions
		"hasPrefix": strings.HasPrefix,
		"hasSuffix": strings.HasSuffix,
		"contains":  strings.Contains,

		// Default value function
		"default": func(defaultVal, val interface{}) interface{} {
			if val == nil || val == "" {
				return defaultVal
			}
			return val
		},

		// Go-specific functions
		"goPackage":  extractGoPackageName,
		"goImport":   formatGoImport,
		"goComment":  formatGoComment,
		"goReceiver": generateGoReceiver,
	}
}

// String case conversion functions

func toCamelCase(s string) string {
	if s == "" {
		return s
	}

	// Handle kebab-case and snake_case
	s = strings.ReplaceAll(s, "-", "_")
	parts := strings.Split(s, "_")

	if len(parts) == 1 {
		return strings.ToLower(parts[0])
	}

	result := strings.ToLower(parts[0])
	for _, part := range parts[1:] {
		if part != "" {
			result += strings.Title(strings.ToLower(part))
		}
	}
	return result
}

func toPascalCase(s string) string {
	camel := toCamelCase(s)
	if camel == "" {
		return camel
	}
	return strings.ToUpper(camel[:1]) + camel[1:]
}

func toSnakeCase(s string) string {
	if s == "" {
		return s
	}

	// Handle kebab-case
	s = strings.ReplaceAll(s, "-", "_")

	// Convert PascalCase/camelCase to snake_case
	var result strings.Builder
	for i, r := range s {
		if i > 0 && 'A' <= r && r <= 'Z' {
			result.WriteRune('_')
		}
		result.WriteRune(r)
	}

	return strings.ToLower(result.String())
}

func toKebabCase(s string) string {
	return strings.ReplaceAll(toSnakeCase(s), "_", "-")
}

// Pluralization functions (simple implementation)

func pluralize(s string) string {
	if s == "" {
		return s
	}

	// Simple pluralization rules
	if strings.HasSuffix(s, "y") {
		return s[:len(s)-1] + "ies"
	}
	if strings.HasSuffix(s, "s") || strings.HasSuffix(s, "sh") || strings.HasSuffix(s, "ch") {
		return s + "es"
	}
	return s + "s"
}

func singularize(s string) string {
	if s == "" {
		return s
	}

	// Simple singularization rules
	if strings.HasSuffix(s, "ies") {
		return s[:len(s)-3] + "y"
	}
	if strings.HasSuffix(s, "es") {
		return s[:len(s)-2]
	}
	if strings.HasSuffix(s, "s") && len(s) > 1 {
		return s[:len(s)-1]
	}
	return s
}

// Path and import functions

func joinPath(parts ...string) string {
	return strings.Join(parts, "/")
}

func cleanPath(path string) string {
	// Remove double slashes and clean up path
	cleaned := strings.ReplaceAll(path, "//", "/")
	cleaned = strings.Trim(cleaned, "/")
	return cleaned
}

func generateImportPath(module, internalPath string) string {
	if module == "" || internalPath == "" {
		return internalPath
	}
	return joinPath(module, internalPath)
}

// Go-specific functions

func extractGoPackageName(path string) string {
	if path == "" {
		return "main"
	}

	// Extract the last part of the path as package name
	parts := strings.Split(path, "/")
	packageName := parts[len(parts)-1]

	// Clean up package name to be valid Go identifier
	packageName = strings.ReplaceAll(packageName, "-", "")
	packageName = strings.ReplaceAll(packageName, "_", "")

	if packageName == "" {
		return "main"
	}

	return strings.ToLower(packageName)
}

func formatGoImport(importPath string) string {
	if importPath == "" {
		return ""
	}
	return `"` + importPath + `"`
}

func formatGoComment(comment string) string {
	if comment == "" {
		return ""
	}

	lines := strings.Split(comment, "\n")
	var result []string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			if !strings.HasPrefix(line, "//") {
				line = "// " + line
			}
			result = append(result, line)
		}
	}

	return strings.Join(result, "\n")
}

func generateGoReceiver(typeName string) string {
	if typeName == "" {
		return "r"
	}

	// Use first letter of type name, lowercased
	receiver := strings.ToLower(string(typeName[0]))

	// Avoid common Go reserved words
	reserved := map[string]string{
		"i": "idx", // interface
		"t": "typ", // type
		"c": "ctx", // context (common)
	}

	if replacement, exists := reserved[receiver]; exists {
		return replacement
	}

	return receiver
}
