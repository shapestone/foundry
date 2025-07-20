package templates

import (
	"strings"
	"unicode"
)

// GetHandlerTemplate returns the Go template for a handler
func GetHandlerTemplate(name string) string {
	template := HandlerTemplate

	// Process all template variables with proper case conversion
	template = strings.ReplaceAll(template, "{{.Name}}", capitalize(name))
	template = strings.ReplaceAll(template, "{{.name}}", strings.ToLower(name))

	return template
}

// GetHandlerUsage returns usage instructions for handlers
func GetHandlerUsage(name string) string {
	template := HandlerUsage

	// Process template variables in usage text
	template = strings.ReplaceAll(template, "{{.Name}}", capitalize(name))
	template = strings.ReplaceAll(template, "{{.name}}", strings.ToLower(name))

	return template
}

// capitalize returns the string with the first letter capitalized
func capitalize(s string) string {
	if s == "" {
		return s
	}

	runes := []rune(s)
	runes[0] = unicode.ToUpper(runes[0])
	return string(runes)
}
