package layouts

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/shapestone/foundry/internal/templating"
)

// StandardLayout implements the standard Go project layout
type StandardLayout struct{}

// NewStandardLayout creates a new standard layout instance
func NewStandardLayout() Layout {
	return &StandardLayout{}
}

// Metadata returns layout information
func (s *StandardLayout) Metadata() LayoutMetadata {
	return LayoutMetadata{
		Name:        "standard",
		DisplayName: "Standard Go Project",
		Description: "The most common Go project layout used by kubernetes, docker, prometheus",
		UseCase:     "Most Go applications",
		TeamSize:    "2-10 developers",
		Complexity:  "Medium",
		Features: []string{
			"internal/ for private packages",
			"cmd/ for application entry points",
			"pkg/ for public libraries",
			"Clear separation of concerns",
		},
	}
}

// GetComponentPlacement returns where a component should be placed
func (s *StandardLayout) GetComponentPlacement(component ComponentType) (*ComponentPlacement, error) {
	switch component {
	case ComponentHandler:
		return &ComponentPlacement{
			Directory:        "internal/handlers",
			FilenamePattern:  "{name}.go",
			PackageName:      "handlers",
			ImportPath:       "{module}/internal/handlers",
			SupportsMultiple: true,
		}, nil

	case ComponentModel:
		return &ComponentPlacement{
			Directory:        "internal/models",
			FilenamePattern:  "{name}.go",
			PackageName:      "models",
			ImportPath:       "{module}/internal/models",
			SupportsMultiple: true,
		}, nil

	case ComponentMiddleware:
		return &ComponentPlacement{
			Directory:        "internal/middleware",
			FilenamePattern:  "{name}.go",
			PackageName:      "middleware",
			ImportPath:       "{module}/internal/middleware",
			SupportsMultiple: true,
		}, nil

	case ComponentService:
		return &ComponentPlacement{
			Directory:        "internal/services",
			FilenamePattern:  "{name}.go",
			PackageName:      "services",
			ImportPath:       "{module}/internal/services",
			SupportsMultiple: true,
		}, nil

	case ComponentRepository:
		return &ComponentPlacement{
			Directory:        "internal/repository",
			FilenamePattern:  "{name}.go",
			PackageName:      "repository",
			ImportPath:       "{module}/internal/repository",
			SupportsMultiple: true,
		}, nil

	default:
		return nil, NewLayoutError("validation", "standard",
			fmt.Errorf("component type %s not supported in standard layout", component))
	}
}

// GetDirectoryStructure returns the complete directory structure
func (s *StandardLayout) GetDirectoryStructure() []DirectoryStructure {
	return []DirectoryStructure{
		{
			Path:        "cmd",
			Description: "Main applications for this project",
			Required:    true,
			Children:    []string{"{projectname}"},
		},
		{
			Path:        "internal",
			Description: "Private application and library code",
			Required:    true,
			Children:    []string{"handlers", "models", "middleware", "services", "repository", "routes"},
		},
		{
			Path:        "internal/handlers",
			Description: "HTTP request handlers",
			Required:    true,
			Children:    []string{},
		},
		{
			Path:        "internal/models",
			Description: "Data models and structures",
			Required:    false,
			Children:    []string{},
		},
		{
			Path:        "internal/middleware",
			Description: "HTTP middleware components",
			Required:    true,
			Children:    []string{},
		},
		{
			Path:        "internal/services",
			Description: "Business logic layer",
			Required:    false,
			Children:    []string{},
		},
		{
			Path:        "internal/repository",
			Description: "Data access layer",
			Required:    false,
			Children:    []string{},
		},
		{
			Path:        "internal/routes",
			Description: "Route configuration",
			Required:    true,
			Children:    []string{},
		},
		{
			Path:        "pkg",
			Description: "Public library code",
			Required:    false,
			Children:    []string{"utils"},
		},
		{
			Path:        "api",
			Description: "API definitions (OpenAPI, Protocol Buffers)",
			Required:    false,
			Children:    []string{},
		},
		{
			Path:        "configs",
			Description: "Configuration files",
			Required:    false,
			Children:    []string{},
		},
		{
			Path:        "scripts",
			Description: "Build and install scripts",
			Required:    false,
			Children:    []string{},
		},
		{
			Path:        "test",
			Description: "Additional test data",
			Required:    false,
			Children:    []string{},
		},
	}
}

// IsComponentSupported checks if a component type is supported
func (s *StandardLayout) IsComponentSupported(component ComponentType) bool {
	supportedComponents := []ComponentType{
		ComponentHandler,
		ComponentModel,
		ComponentMiddleware,
		ComponentService,
		ComponentRepository,
	}

	for _, supported := range supportedComponents {
		if component == supported {
			return true
		}
	}
	return false
}

// GetTemplateData returns layout-specific template data
func (s *StandardLayout) GetTemplateData(projectName string) templating.TemplateData {
	return templating.TemplateData{
		"Layout":      "standard",
		"ProjectName": projectName,
		"Structure": map[string]string{
			"HandlersPath":   "internal/handlers",
			"ModelsPath":     "internal/models",
			"MiddlewarePath": "internal/middleware",
			"ServicesPath":   "internal/services",
			"RepositoryPath": "internal/repository",
			"RoutesPath":     "internal/routes",
		},
		"Imports": map[string]string{
			"HandlersImport":   "{module}/internal/handlers",
			"ModelsImport":     "{module}/internal/models",
			"MiddlewareImport": "{module}/internal/middleware",
			"ServicesImport":   "{module}/internal/services",
			"RepositoryImport": "{module}/internal/repository",
		},
	}
}

// ValidateProjectStructure validates if a project follows this layout
func (s *StandardLayout) ValidateProjectStructure(projectPath string) error {
	requiredDirs := []string{
		"cmd",
		"internal",
		"internal/handlers",
		"internal/middleware",
		"internal/routes",
	}

	for _, dir := range requiredDirs {
		dirPath := filepath.Join(projectPath, dir)
		if _, err := os.Stat(dirPath); os.IsNotExist(err) {
			return NewLayoutError("validation", "standard",
				fmt.Errorf("required directory %s does not exist", dir))
		}
	}

	// Check for go.mod in project root
	goModPath := filepath.Join(projectPath, "go.mod")
	if _, err := os.Stat(goModPath); os.IsNotExist(err) {
		return NewLayoutError("validation", "standard",
			fmt.Errorf("go.mod file not found in project root"))
	}

	return nil
}

// GetProjectFiles returns files that should be created during project initialization
func (s *StandardLayout) GetProjectFiles() []ProjectFile {
	return []ProjectFile{
		{
			Path:         "cmd/{projectname}/main.go",
			TemplateName: "main.go",
			Required:     true,
		},
		{
			Path:         "go.mod",
			TemplateName: "go.mod",
			Required:     true,
		},
		{
			Path:         "README.md",
			TemplateName: "README.md",
			Required:     true,
		},
		{
			Path:         ".gitignore",
			TemplateName: "gitignore",
			Required:     true,
		},
		{
			Path:         "internal/routes/routes.go",
			TemplateName: "routes.go",
			Required:     true,
		},
		{
			Path:         ".foundry.yaml",
			TemplateName: "foundry.yaml",
			Required:     true,
		},
	}
}

// GetDetectionRules returns rules for detecting if a project uses standard layout
func (s *StandardLayout) GetDetectionRules() []LayoutDetectionRule {
	return []LayoutDetectionRule{
		{
			RequiredDirs: []string{
				"cmd",
				"internal",
				"internal/handlers",
			},
			RequiredFiles: []string{
				"go.mod",
			},
			ProhibitedPaths: []string{
				"internal/core",         // hexagonal
				"internal/domain",       // clean
				"internal/usecase",      // clean
				"internal/delivery",     // clean
				"internal/api/handlers", // microservice
			},
			LayoutName: "standard",
			Confidence: 90,
		},
	}
}
