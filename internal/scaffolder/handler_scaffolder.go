// internal/scaffolder/handler_scaffolder.go
package scaffolder

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
)

// handlerScaffolder implements handler creation logic
type handlerScaffolder struct {
	fileSystem       FileSystem
	templateRenderer TemplateRenderer
	projectAnalyzer  ProjectAnalyzer
	userInteraction  UserInteraction
}

// CreateHandler creates a new handler based on the specification
func (s *handlerScaffolder) CreateHandler(ctx context.Context, spec *HandlerSpec) (*Result, error) {
	// Validate the specification
	if err := s.validateHandlerSpec(spec); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Check if we're in a Go project
	if !s.projectAnalyzer.IsGoProject(spec.ProjectRoot) {
		return nil, fmt.Errorf("not a Go project: go.mod not found in %s", spec.ProjectRoot)
	}

	// Get module information if not provided
	if spec.Module == "" {
		module, err := s.projectAnalyzer.GetModuleName(spec.ProjectRoot)
		if err != nil {
			return nil, fmt.Errorf("failed to get module name: %w", err)
		}
		spec.Module = module
	}

	result := &Result{
		FilesCreated: []string{},
		FilesUpdated: []string{},
		Changes:      []string{},
		Warnings:     []string{},
		Success:      true,
		Metadata:     make(map[string]string),
	}

	// Prepare handler data
	handlerData := s.prepareHandlerData(spec)

	// Generate handler file path
	handlerPath := s.generateHandlerPath(spec)

	// Check if handler already exists
	if s.fileSystem.Exists(handlerPath) {
		return nil, fmt.Errorf("handler already exists: %s", handlerPath)
	}

	// Load and render template
	handlerContent, err := s.renderHandlerTemplate(handlerData)
	if err != nil {
		return nil, fmt.Errorf("failed to render handler template: %w", err)
	}

	// If dry run, show preview and return
	if spec.DryRun {
		changes := []string{
			fmt.Sprintf("+ Create %s", handlerPath),
			fmt.Sprintf("+ Add handler: %s", spec.Name),
		}

		if spec.AutoWire {
			changes = append(changes, "+ Update routes.go")
		}

		result.Changes = changes
		result.Message = "Dry run completed - no files were created"
		return result, nil
	}

	// Create the handler file
	if err := s.createHandlerFile(handlerPath, handlerContent); err != nil {
		return nil, fmt.Errorf("failed to create handler file: %w", err)
	}

	result.FilesCreated = append(result.FilesCreated, handlerPath)
	result.Changes = append(result.Changes, fmt.Sprintf("Created handler: %s", handlerPath))

	// Handle auto-wiring if requested
	if spec.AutoWire {
		wireResult, err := s.wireHandler(ctx, spec)
		if err != nil {
			result.Warnings = append(result.Warnings,
				fmt.Sprintf("Handler created but auto-wiring failed: %v", err))
		} else {
			result.FilesUpdated = append(result.FilesUpdated, wireResult.FilesUpdated...)
			result.Changes = append(result.Changes, wireResult.Changes...)
		}
	}

	result.Message = fmt.Sprintf("Handler '%s' created successfully", spec.Name)
	result.Metadata["handler_path"] = handlerPath
	result.Metadata["resource_path"] = strings.ToLower(spec.Name) + "s"

	return result, nil
}

// validateHandlerSpec validates the handler specification
func (s *handlerScaffolder) validateHandlerSpec(spec *HandlerSpec) error {
	var errors ValidationErrors

	if spec.Name == "" {
		errors = append(errors, ValidationError{
			Field:   "name",
			Message: "handler name is required",
		})
	}

	if len(spec.Name) < 2 {
		errors = append(errors, ValidationError{
			Field:   "name",
			Message: "handler name must be at least 2 characters",
		})
	}

	if len(spec.Name) > 50 {
		errors = append(errors, ValidationError{
			Field:   "name",
			Message: "handler name must be less than 50 characters",
		})
	}

	// Validate name format (basic Go identifier rules)
	if !isValidGoIdentifier(spec.Name) {
		errors = append(errors, ValidationError{
			Field:   "name",
			Message: "handler name must be a valid Go identifier",
		})
	}

	if spec.ProjectRoot == "" {
		errors = append(errors, ValidationError{
			Field:   "project_root",
			Message: "project root is required",
		})
	}

	if errors.HasErrors() {
		return errors
	}

	return nil
}

// prepareHandlerData prepares template data for handler generation
func (s *handlerScaffolder) prepareHandlerData(spec *HandlerSpec) map[string]interface{} {
	handlerName := toGoIdentifier(spec.Name)
	resourceName := strings.ToLower(spec.Name)
	resourceNamePlural := pluralize(resourceName)
	resourcePath := resourceName + "s"

	return map[string]interface{}{
		"HandlerName":        handlerName,
		"ResourceName":       resourceName,
		"ResourceNamePlural": resourceNamePlural,
		"ResourcePath":       resourcePath,
		"Module":             spec.Module,
		"Type":               spec.Type,
		"Metadata":           spec.Metadata,
	}
}

// generateHandlerPath generates the file path for the handler
func (s *handlerScaffolder) generateHandlerPath(spec *HandlerSpec) string {
	fileName := strings.ToLower(spec.Name) + ".go"
	return filepath.Join(spec.ProjectRoot, "internal", "handlers", fileName)
}

// renderHandlerTemplate renders the handler template with data
func (s *handlerScaffolder) renderHandlerTemplate(data map[string]interface{}) (string, error) {
	template, err := s.templateRenderer.LoadTemplate("handler.go.tmpl")
	if err != nil {
		return "", fmt.Errorf("failed to load handler template: %w", err)
	}

	content, err := s.templateRenderer.RenderTemplate(template, data)
	if err != nil {
		return "", fmt.Errorf("failed to render handler template: %w", err)
	}

	return content, nil
}

// createHandlerFile creates the handler file with the given content
func (s *handlerScaffolder) createHandlerFile(path, content string) error {
	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := s.fileSystem.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	// Write the file
	if err := s.fileSystem.WriteFile(path, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write file %s: %w", path, err)
	}

	return nil
}

// wireHandler handles auto-wiring of the handler
func (s *handlerScaffolder) wireHandler(ctx context.Context, spec *HandlerSpec) (*Result, error) {
	// TODO: Implement wiring logic here
	// For now, return a simple success result
	return &Result{
		Success: true,
		Message: "Auto-wiring not yet implemented",
		Changes: []string{"Would update routes.go"},
	}, nil
}

// Helper functions

// toGoIdentifier converts a string to a valid Go identifier
func toGoIdentifier(s string) string {
	// Remove hyphens and underscores, capitalize each word
	parts := strings.FieldsFunc(s, func(r rune) bool {
		return r == '-' || r == '_'
	})

	var result strings.Builder
	for _, part := range parts {
		if len(part) > 0 {
			result.WriteString(strings.ToUpper(part[:1]))
			if len(part) > 1 {
				result.WriteString(strings.ToLower(part[1:]))
			}
		}
	}

	return result.String()
}

// pluralize adds simple pluralization
func pluralize(s string) string {
	if strings.HasSuffix(s, "y") {
		return s[:len(s)-1] + "ies"
	}
	if strings.HasSuffix(s, "s") || strings.HasSuffix(s, "x") || strings.HasSuffix(s, "z") ||
		strings.HasSuffix(s, "ch") || strings.HasSuffix(s, "sh") {
		return s + "es"
	}
	return s + "s"
}

// isValidGoIdentifier checks if a string is a valid Go identifier
func isValidGoIdentifier(s string) bool {
	if s == "" {
		return false
	}

	// Check for Go keywords
	goKeywords := map[string]bool{
		"break": true, "case": true, "chan": true, "const": true, "continue": true,
		"default": true, "defer": true, "else": true, "fallthrough": true, "for": true,
		"func": true, "go": true, "goto": true, "if": true, "import": true,
		"interface": true, "map": true, "package": true, "range": true, "return": true,
		"select": true, "struct": true, "switch": true, "type": true, "var": true,
	}

	if goKeywords[strings.ToLower(s)] {
		return false
	}

	// Basic validation - more comprehensive validation can be added
	if strings.ContainsAny(s, " \t\n\r") {
		return false
	}

	return true
}
