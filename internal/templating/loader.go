package templating

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/shapestone/foundry"
)

// FileSystemLoader loads templates from the filesystem
type FileSystemLoader struct {
	config *TemplateConfig
	cache  TemplateCache
}

// NewFileSystemLoader creates a new filesystem template loader
func NewFileSystemLoader(config *TemplateConfig, cache TemplateCache) TemplateLoader {
	return &FileSystemLoader{
		config: config,
		cache:  cache,
	}
}

// LoadTemplate loads a specific template by name and layout
func (f *FileSystemLoader) LoadTemplate(layout, category, name string) (*Template, error) {
	// Generate cache key
	cacheKey := f.generateCacheKey(layout, category, name)

	// Check cache first if enabled
	if f.config.EnableCaching {
		if cached, found := f.cache.Get(cacheKey); found {
			return cached, nil
		}
	}

	// Try loading from custom template directory first
	if f.config.CustomTemplateDir != "" {
		if tmpl, err := f.loadFromDirectory(f.config.CustomTemplateDir, layout, category, name); err == nil {
			f.cacheTemplate(cacheKey, tmpl)
			return tmpl, nil
		}
	}

	// Try loading from main template directory
	if tmpl, err := f.loadFromDirectory(f.config.TemplateDir, layout, category, name); err == nil {
		f.cacheTemplate(cacheKey, tmpl)
		return tmpl, nil
	}

	// Fallback to embedded templates if enabled
	if f.config.FallbackToEmbedded {
		if tmpl, err := f.loadEmbeddedTemplate(layout, category, name); err == nil {
			f.cacheTemplate(cacheKey, tmpl)
			return tmpl, nil
		}
	}

	return nil, NewTemplateError("load", name, layout,
		fmt.Errorf("template not found in any source"))
}

// LoadProjectTemplates loads all templates needed for project creation
func (f *FileSystemLoader) LoadProjectTemplates(layout string) ([]*Template, error) {
	var templates []*Template

	// Project template files we expect to find
	projectTemplates := []string{
		"main.go",
		"go.mod",
		"README.md",
		"gitignore",
		"routes.go",
		"foundry.yaml",
	}

	for _, templateName := range projectTemplates {
		tmpl, err := f.LoadTemplate(layout, "project", templateName)
		if err != nil {
			// Some project templates are optional, continue on error
			continue
		}
		templates = append(templates, tmpl)
	}

	if len(templates) == 0 {
		return nil, NewTemplateError("load", "project", layout,
			fmt.Errorf("no project templates found"))
	}

	return templates, nil
}

// LoadComponentTemplate loads template for a specific component type
func (f *FileSystemLoader) LoadComponentTemplate(layout, component string) (*Template, error) {
	return f.LoadTemplate(layout, "components", component+".go")
}

// ListAvailableTemplates returns all available templates for a layout
func (f *FileSystemLoader) ListAvailableTemplates(layout string) ([]string, error) {
	var templates []string

	// Check main template directory
	layoutDir := filepath.Join(f.config.TemplateDir, "layouts", layout)
	if err := f.walkTemplateDirectory(layoutDir, &templates); err == nil {
		// Directory exists, add templates found
	}

	// Check custom template directory
	if f.config.CustomTemplateDir != "" {
		customLayoutDir := filepath.Join(f.config.CustomTemplateDir, "layouts", layout)
		if err := f.walkTemplateDirectory(customLayoutDir, &templates); err == nil {
			// Directory exists, add templates found
		}
	}

	// Add embedded templates if fallback enabled
	if f.config.FallbackToEmbedded {
		embeddedTemplates := f.listEmbeddedTemplates(layout)
		templates = append(templates, embeddedTemplates...)
	}

	// Remove duplicates
	templates = f.removeDuplicates(templates)

	return templates, nil
}

// TemplateExists checks if a template exists
func (f *FileSystemLoader) TemplateExists(layout, category, name string) bool {
	_, err := f.LoadTemplate(layout, category, name)
	return err == nil
}

// loadFromDirectory loads a template from a specific directory
func (f *FileSystemLoader) loadFromDirectory(baseDir, layout, category, name string) (*Template, error) {
	templatePath := filepath.Join(baseDir, "layouts", layout, category, name+".tmpl")

	// Check if file exists
	if _, err := os.Stat(templatePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("template file not found: %s", templatePath)
	}

	// Read template content
	content, err := os.ReadFile(templatePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read template file: %v", err)
	}

	// Parse template
	tmpl, err := template.New(name).Parse(string(content))
	if err != nil {
		return nil, NewTemplateError("parse", name, layout, err)
	}

	return &Template{
		Name:     name,
		Content:  string(content),
		Parsed:   tmpl,
		Layout:   layout,
		Category: category,
	}, nil
}

// loadEmbeddedTemplate loads a template from embedded files
func (f *FileSystemLoader) loadEmbeddedTemplate(layout, category, name string) (*Template, error) {
	// Map to current embedded template structure
	var templatePath string

	switch category {
	case "project":
		switch name {
		case "main.go":
			templatePath = "templates/main.go.tmpl"
		case "go.mod":
			templatePath = "templates/go.mod.tmpl"
		case "README.md":
			templatePath = "templates/README.md.tmpl"
		case "gitignore":
			templatePath = "templates/gitignore.tmpl"
		case "routes.go":
			templatePath = "templates/routes.go.tmpl"
		case "foundry.yaml":
			templatePath = "templates/foundry.yaml.tmpl"
		default:
			return nil, fmt.Errorf("unknown project template: %s", name)
		}
	case "components":
		switch name {
		case "handler.go":
			templatePath = "templates/handler.go.tmpl"
		case "model.go":
			templatePath = "templates/model.go.tmpl"
		case "middleware.go":
			templatePath = "templates/middleware.go.tmpl"
		default:
			return nil, fmt.Errorf("unknown component template: %s", name)
		}
	default:
		return nil, fmt.Errorf("unknown template category: %s", category)
	}

	// Load from embedded files
	content, err := foundry.Templates.ReadFile(templatePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read embedded template: %v", err)
	}

	// Parse template
	tmpl, err := template.New(name).Parse(string(content))
	if err != nil {
		return nil, NewTemplateError("parse", name, layout, err)
	}

	return &Template{
		Name:     name,
		Content:  string(content),
		Parsed:   tmpl,
		Layout:   layout,
		Category: category,
	}, nil
}

// walkTemplateDirectory walks a template directory and collects template names
func (f *FileSystemLoader) walkTemplateDirectory(dir string, templates *[]string) error {
	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && strings.HasSuffix(path, ".tmpl") {
			// Extract relative path and remove .tmpl extension
			relPath, err := filepath.Rel(dir, path)
			if err != nil {
				return err
			}

			templateName := strings.TrimSuffix(relPath, ".tmpl")
			*templates = append(*templates, templateName)
		}

		return nil
	})
}

// listEmbeddedTemplates returns list of embedded templates
func (f *FileSystemLoader) listEmbeddedTemplates(layout string) []string {
	// For now, return known embedded templates
	// This could be made more dynamic in the future
	return []string{
		"project/main.go",
		"project/go.mod",
		"project/README.md",
		"project/gitignore",
		"project/routes.go",
		"project/foundry.yaml",
		"components/handler.go",
		"components/model.go",
		"components/middleware.go",
	}
}

// generateCacheKey generates a cache key for a template
func (f *FileSystemLoader) generateCacheKey(layout, category, name string) string {
	return fmt.Sprintf("%s:%s:%s", layout, category, name)
}

// cacheTemplate stores a template in cache if caching is enabled
func (f *FileSystemLoader) cacheTemplate(key string, tmpl *Template) {
	if f.config.EnableCaching && f.cache != nil {
		f.cache.Set(key, tmpl)
	}
}

// removeDuplicates removes duplicate strings from a slice
func (f *FileSystemLoader) removeDuplicates(input []string) []string {
	seen := make(map[string]bool)
	var result []string

	for _, item := range input {
		if !seen[item] {
			seen[item] = true
			result = append(result, item)
		}
	}

	return result
}
