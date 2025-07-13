package templating

import (
	"fmt"
	"io"
	"text/template"
)

// TemplateData represents data passed to templates
type TemplateData map[string]interface{}

// Template represents a parsed template
type Template struct {
	Name     string
	Content  string
	Parsed   *template.Template
	Layout   string
	Category string // "project", "component"
}

// TemplateLoader loads templates from various sources
type TemplateLoader interface {
	// LoadTemplate loads a specific template by name and layout
	LoadTemplate(layout, category, name string) (*Template, error)

	// LoadProjectTemplates loads all templates needed for project creation
	LoadProjectTemplates(layout string) ([]*Template, error)

	// LoadComponentTemplate loads template for a specific component type
	LoadComponentTemplate(layout, component string) (*Template, error)

	// ListAvailableTemplates returns all available templates for a layout
	ListAvailableTemplates(layout string) ([]string, error)

	// TemplateExists checks if a template exists
	TemplateExists(layout, category, name string) bool
}

// TemplateRenderer handles template rendering with data
type TemplateRenderer interface {
	// Render renders a template with the provided data
	Render(tmpl *Template, data TemplateData) (string, error)

	// RenderToWriter renders a template directly to a writer
	RenderToWriter(tmpl *Template, data TemplateData, w io.Writer) error

	// ValidateTemplate validates template syntax
	ValidateTemplate(tmpl *Template) error
}

// TemplateCache provides caching for parsed templates
type TemplateCache interface {
	// Get retrieves a cached template
	Get(key string) (*Template, bool)

	// Set stores a template in cache
	Set(key string, tmpl *Template)

	// Clear clears the cache
	Clear()

	// Remove removes a specific template from cache
	Remove(key string)
}

// TemplateManager coordinates template operations
type TemplateManager interface {
	// LoadAndRender loads and renders a template in one operation
	LoadAndRender(layout, category, name string, data TemplateData) (string, error)

	// GetTemplate gets a template (with caching)
	GetTemplate(layout, category, name string) (*Template, error)

	// RenderTemplate renders a loaded template
	RenderTemplate(tmpl *Template, data TemplateData) (string, error)

	// ClearCache clears the template cache
	ClearCache()
}

// TemplateSource represents where templates are loaded from
type TemplateSource string

const (
	SourceFilesystem TemplateSource = "filesystem"
	SourceEmbedded   TemplateSource = "embedded"
	SourceCustom     TemplateSource = "custom"
)

// TemplateConfig configures template loading behavior
type TemplateConfig struct {
	// TemplateDir is the base directory for external templates
	TemplateDir string

	// FallbackToEmbedded enables falling back to embedded templates
	FallbackToEmbedded bool

	// EnableCaching enables template caching
	EnableCaching bool

	// CustomTemplateDir is project-specific template directory
	CustomTemplateDir string
}

// TemplateError represents template-related errors
type TemplateError struct {
	Type     string // "load", "parse", "render"
	Template string
	Layout   string
	Cause    error
}

func (e *TemplateError) Error() string {
	return fmt.Sprintf("template error (%s): %s in layout %s: %v", e.Type, e.Template, e.Layout, e.Cause)
}

// NewTemplateError creates a new template error
func NewTemplateError(errorType, template, layout string, cause error) *TemplateError {
	return &TemplateError{
		Type:     errorType,
		Template: template,
		Layout:   layout,
		Cause:    cause,
	}
}
