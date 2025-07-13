package cmd

import (
	"fmt"
	"go/token"
	"regexp"
	"strings"
	"unicode"
)

// Go reserved keywords that cannot be used as component names
var goKeywords = map[string]bool{
	"break": true, "case": true, "chan": true, "const": true, "continue": true,
	"default": true, "defer": true, "else": true, "fallthrough": true, "for": true,
	"func": true, "go": true, "goto": true, "if": true, "import": true,
	"interface": true, "map": true, "package": true, "range": true, "return": true,
	"select": true, "struct": true, "switch": true, "type": true, "var": true,
}

// Common invalid patterns for component names
var invalidPatterns = []*regexp.Regexp{
	regexp.MustCompile(`^\d`),            // Cannot start with number
	regexp.MustCompile(`[^a-zA-Z0-9_-]`), // Only letters, numbers, underscore, hyphen
	regexp.MustCompile(`--+`),            // Multiple consecutive hyphens
	regexp.MustCompile(`__+`),            // Multiple consecutive underscores
	regexp.MustCompile(`^-`),             // Cannot start with hyphen
	regexp.MustCompile(`-$`),             // Cannot end with hyphen
	regexp.MustCompile(`^_`),             // Cannot start with underscore
	regexp.MustCompile(`_$`),             // Cannot end with underscore
}

// ValidateComponentName validates a component name according to Go and CLI conventions
func ValidateComponentName(name string) error {
	if name == "" {
		return fmt.Errorf("component name cannot be empty")
	}

	// Check length
	if len(name) > 50 {
		return fmt.Errorf("component name too long (max 50 characters)")
	}

	if len(name) < 2 {
		return fmt.Errorf("component name too short (min 2 characters)")
	}

	// Check for spaces
	if strings.Contains(name, " ") {
		return fmt.Errorf("component name cannot contain spaces")
	}

	// Check for tabs or other whitespace
	for _, r := range name {
		if unicode.IsSpace(r) {
			return fmt.Errorf("component name cannot contain whitespace characters")
		}
	}

	// Check Go reserved keywords
	if goKeywords[strings.ToLower(name)] {
		return fmt.Errorf("component name cannot be a Go reserved keyword: %s", name)
	}

	// Check invalid patterns
	for _, pattern := range invalidPatterns {
		if pattern.MatchString(name) {
			switch pattern.String() {
			case `^\d`:
				return fmt.Errorf("component name cannot start with a number")
			case `[^a-zA-Z0-9_-]`:
				return fmt.Errorf("component name can only contain letters, numbers, underscores, and hyphens")
			case `--+`:
				return fmt.Errorf("component name cannot contain consecutive hyphens")
			case `__+`:
				return fmt.Errorf("component name cannot contain consecutive underscores")
			case `^-`:
				return fmt.Errorf("component name cannot start with a hyphen")
			case `-$`:
				return fmt.Errorf("component name cannot end with a hyphen")
			case `^_`:
				return fmt.Errorf("component name cannot start with an underscore")
			case `_$`:
				return fmt.Errorf("component name cannot end with an underscore")
			}
		}
	}

	// Check if it's a valid Go identifier when converted
	goIdentifier := ToGoIdentifier(name)
	if !token.IsIdentifier(goIdentifier) {
		return fmt.Errorf("component name %q would not generate a valid Go identifier", name)
	}

	// Check for common problematic names
	problematicNames := map[string]string{
		"test":    "conflicts with Go testing",
		"main":    "conflicts with main package",
		"init":    "conflicts with init function",
		"new":     "conflicts with built-in new function",
		"make":    "conflicts with built-in make function",
		"len":     "conflicts with built-in len function",
		"cap":     "conflicts with built-in cap function",
		"append":  "conflicts with built-in append function",
		"copy":    "conflicts with built-in copy function",
		"delete":  "conflicts with built-in delete function",
		"close":   "conflicts with built-in close function",
		"panic":   "conflicts with built-in panic function",
		"recover": "conflicts with built-in recover function",
		"print":   "conflicts with built-in print function",
		"println": "conflicts with built-in println function",
		"error":   "conflicts with built-in error type",
		"string":  "conflicts with built-in string type",
		"int":     "conflicts with built-in int type",
		"float64": "conflicts with built-in float64 type",
		"bool":    "conflicts with built-in bool type",
		"byte":    "conflicts with built-in byte type",
		"rune":    "conflicts with built-in rune type",
	}

	if reason, exists := problematicNames[strings.ToLower(name)]; exists {
		return fmt.Errorf("component name %q is not recommended: %s", name, reason)
	}

	return nil
}

// ToGoIdentifier converts a component name to a valid Go identifier
// This is used to check if the name will work in generated code
func ToGoIdentifier(name string) string {
	// Replace hyphens with underscores for Go compatibility
	identifier := strings.ReplaceAll(name, "-", "_")

	// Capitalize first letter for exported identifiers
	if len(identifier) > 0 {
		identifier = strings.ToUpper(string(identifier[0])) + identifier[1:]
	}

	return identifier
}

// ValidateComponentType validates that the component type is supported
func ValidateComponentType(componentType string) error {
	validTypes := map[string]bool{
		"handler":    true,
		"model":      true,
		"middleware": true,
	}

	if !validTypes[componentType] {
		return fmt.Errorf("unsupported component type: %s (valid types: handler, model, middleware)", componentType)
	}

	return nil
}

// ValidateAddCommandArgs validates arguments for the add command
func ValidateAddCommandArgs(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("add command requires component type and name")
	}

	componentType := args[0]
	componentName := args[1]

	if err := ValidateComponentType(componentType); err != nil {
		return err
	}

	if err := ValidateComponentName(componentName); err != nil {
		return err
	}

	return nil
}
