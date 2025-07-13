package layouts

import (
	"fmt"
	"github.com/shapestone/foundry/internal/templating"
)

// ComponentType represents different types of components that can be generated
type ComponentType string

const (
	ComponentHandler    ComponentType = "handler"
	ComponentModel      ComponentType = "model"
	ComponentMiddleware ComponentType = "middleware"
	ComponentService    ComponentType = "service"
	ComponentRepository ComponentType = "repository"
	ComponentDomain     ComponentType = "domain"
)

// ComponentPlacement defines where a component should be placed in a layout
type ComponentPlacement struct {
	// Directory path relative to project root
	Directory string

	// Filename pattern (can include placeholders like {name})
	FilenamePattern string

	// Package name for the Go package
	PackageName string

	// Import path pattern for the package
	ImportPath string

	// Whether this component type supports multiple files per domain/area
	SupportsMultiple bool
}

// LayoutMetadata contains information about a layout
type LayoutMetadata struct {
	Name        string
	DisplayName string
	Description string
	UseCase     string
	TeamSize    string
	Complexity  string
	Features    []string
}

// DirectoryStructure represents the directory structure for a layout
type DirectoryStructure struct {
	Path        string   // Directory path
	Description string   // What this directory is for
	Required    bool     // Whether this directory is required
	Children    []string // Subdirectories
}

// Layout defines how components are organized in a specific project structure
type Layout interface {
	// Metadata returns layout information
	Metadata() LayoutMetadata

	// GetComponentPlacement returns where a component should be placed
	GetComponentPlacement(component ComponentType) (*ComponentPlacement, error)

	// GetDirectoryStructure returns the complete directory structure
	GetDirectoryStructure() []DirectoryStructure

	// IsComponentSupported checks if a component type is supported
	IsComponentSupported(component ComponentType) bool

	// GetTemplateData returns layout-specific template data
	GetTemplateData(projectName string) templating.TemplateData

	// ValidateProjectStructure validates if a project follows this layout
	ValidateProjectStructure(projectPath string) error

	// GetProjectFiles returns files that should be created during project initialization
	GetProjectFiles() []ProjectFile
}

// ProjectFile represents a file to be created during project initialization
type ProjectFile struct {
	Path         string // Relative path from project root
	TemplateName string // Template to use for this file
	Required     bool   // Whether this file is required
}

// LayoutManager manages available layouts and layout operations
type LayoutManager interface {
	// GetLayout retrieves a layout by name
	GetLayout(name string) (Layout, error)

	// ListLayouts returns all available layouts
	ListLayouts() []Layout

	// DetectLayout attempts to detect the layout of an existing project
	DetectLayout(projectPath string) (Layout, error)

	// ValidateLayoutName checks if a layout name is valid
	ValidateLayoutName(name string) error

	// GetDefaultLayout returns the default layout
	GetDefaultLayout() Layout
}

// LayoutConfig represents layout configuration stored in .foundry.yaml
type LayoutConfig struct {
	// Layout name
	Layout string `yaml:"layout"`

	// Custom component placements (overrides)
	ComponentPaths map[ComponentType]string `yaml:"component_paths,omitempty"`

	// Custom template directory
	TemplateDir string `yaml:"template_dir,omitempty"`

	// Layout-specific configuration
	Config map[string]interface{} `yaml:"config,omitempty"`
}

// LayoutDetectionRule defines rules for detecting layouts
type LayoutDetectionRule struct {
	// Required directories that must exist
	RequiredDirs []string

	// Required files that must exist
	RequiredFiles []string

	// Prohibited paths that shouldn't exist
	ProhibitedPaths []string

	// Layout name if this rule matches
	LayoutName string

	// Confidence score (0-100) for this detection
	Confidence int
}

// ComponentGenerator generates components within a specific layout
type ComponentGenerator interface {
	// GenerateComponent creates a component in the appropriate location for this layout
	GenerateComponent(component ComponentType, name string, data templating.TemplateData) error

	// GetComponentPath returns the expected path for a component
	GetComponentPath(component ComponentType, name string) string

	// ValidateComponentName validates a component name for this layout
	ValidateComponentName(component ComponentType, name string) error
}

// DomainLayout represents layouts that support domain-based organization (DDD, etc.)
type DomainLayout interface {
	Layout

	// GetDomainPlacement returns where domain-specific components should be placed
	GetDomainPlacement(domain string, component ComponentType) (*ComponentPlacement, error)

	// ListDomains returns existing domains in the project
	ListDomains(projectPath string) ([]string, error)

	// CreateDomain creates a new domain structure
	CreateDomain(domain string) error

	// IsDomainRequired checks if a domain must be specified for this component
	IsDomainRequired(component ComponentType) bool
}

// LayoutError represents layout-related errors
type LayoutError struct {
	Type   string // "not_found", "validation", "detection"
	Layout string
	Cause  error
}

func (e *LayoutError) Error() string {
	if e.Layout != "" {
		return fmt.Sprintf("layout error (%s) for layout '%s': %v", e.Type, e.Layout, e.Cause)
	}
	return fmt.Sprintf("layout error (%s): %v", e.Type, e.Cause)
}

// NewLayoutError creates a new layout error
func NewLayoutError(errorType, layout string, cause error) *LayoutError {
	return &LayoutError{
		Type:   errorType,
		Layout: layout,
		Cause:  cause,
	}
}
