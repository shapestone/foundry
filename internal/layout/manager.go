package layout

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"text/template"
	"time"
)

// Manager handles layout operations
type Manager struct {
	registry *Registry
	cache    *Cache
	loader   *Loader
}

// NewManager creates a new layout manager
func NewManager(configPath string) (*Manager, error) {
	// Create registry
	registry, err := NewRegistry(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create registry: %w", err)
	}

	// Create cache
	config := registry.GetConfig()
	cache, err := NewCache(config.Cache.Directory, config.Cache.TTL)
	if err != nil {
		return nil, fmt.Errorf("failed to create cache: %w", err)
	}

	// Create loader
	loader := NewLoader(registry, cache)

	return &Manager{
		registry: registry,
		cache:    cache,
		loader:   loader,
	}, nil
}

// ListLayouts returns all available layouts
func (m *Manager) ListLayouts() []LayoutListEntry {
	return m.registry.ListLayouts()
}

// GetLayout retrieves a specific layout
func (m *Manager) GetLayout(ctx context.Context, name string) (*Layout, error) {
	return m.loader.Load(ctx, name)
}

// AddRemoteLayout adds a remote layout to the registry
func (m *Manager) AddRemoteLayout(url string) error {
	// Parse URL to determine source type
	source := LayoutSource{
		Type:     "remote",
		Location: url,
	}

	// TODO: Determine layout name from manifest
	// For now, extract from URL
	name := filepath.Base(url)
	if ext := filepath.Ext(name); ext != "" {
		name = name[:len(name)-len(ext)]
	}

	entry := LayoutListEntry{
		Name:        name,
		Description: "Remote layout",
		Source:      source,
		Installed:   false,
		UpdatedAt:   time.Now(),
	}

	return m.registry.AddLayout(entry)
}

// RefreshLayouts updates the layout registry
func (m *Manager) RefreshLayouts() error {
	return m.registry.RefreshRemoteRegistries()
}

// ProjectData represents template variables for project generation
type ProjectData struct {
	ProjectName     string
	ModuleName      string
	Author          string
	License         string
	Description     string
	GitHubUsername  string
	Year            int
	GoVersion       string
	CustomVariables map[string]string
}

// GenerateProject generates a project using the specified layout
func (m *Manager) GenerateProject(ctx context.Context, layoutName string, projectPath string, data ProjectData) error {
	// Load layout
	layout, err := m.GetLayout(ctx, layoutName)
	if err != nil {
		return fmt.Errorf("failed to load layout: %w", err)
	}

	// Validate project data against layout variables
	if err := m.validateProjectData(layout, &data); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Create project directory structure
	if err := m.createDirectories(layout, projectPath, data); err != nil {
		return fmt.Errorf("failed to create directories: %w", err)
	}

	// Generate files from templates
	if err := m.generateFiles(layout, projectPath, data); err != nil {
		return fmt.Errorf("failed to generate files: %w", err)
	}

	return nil
}

// validateProjectData validates and fills in default values
func (m *Manager) validateProjectData(layout *Layout, data *ProjectData) error {
	// Set defaults
	if data.Year == 0 {
		data.Year = time.Now().Year()
	}

	if data.CustomVariables == nil {
		data.CustomVariables = make(map[string]string)
	}

	// Apply layout variable defaults
	for _, variable := range layout.Manifest.Variables {
		if _, exists := data.CustomVariables[variable.Name]; !exists && variable.Default != "" {
			data.CustomVariables[variable.Name] = variable.Default
		}

		// Check required variables
		if variable.Required {
			if _, exists := data.CustomVariables[variable.Name]; !exists {
				return fmt.Errorf("required variable '%s' not provided", variable.Name)
			}
		}
	}

	return nil
}

// createDirectories creates the project directory structure
func (m *Manager) createDirectories(layout *Layout, projectPath string, data ProjectData) error {
	// Create project root
	if err := os.MkdirAll(projectPath, 0755); err != nil {
		return fmt.Errorf("failed to create project directory: %w", err)
	}

	// Create directories from layout manifest
	for _, dir := range layout.Manifest.Structure.Directories {
		// Process template in directory path
		dirPath := m.processTemplatePath(dir.Path, data)
		fullPath := filepath.Join(projectPath, dirPath)

		if err := os.MkdirAll(fullPath, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dirPath, err)
		}
	}

	return nil
}

// generateFiles generates files from templates
func (m *Manager) generateFiles(layout *Layout, projectPath string, data ProjectData) error {
	// Create template functions
	funcMap := template.FuncMap{
		"lower":      toLower,
		"upper":      toUpper,
		"capitalize": capitalize,
		"snake":      toSnakeCase,
		"camel":      toCamelCase,
		"pascal":     toPascalCase,
		"kebab":      toKebabCase,
		"default":    defaultString,
		"snake_case": toSnakeCase,
		"title":      toPascalCase,
		"plural":     pluralize,
	}

	// Generate each file
	for _, file := range layout.Manifest.Structure.Files {
		// Get template content
		templateContent, exists := layout.Templates[file.Template]
		if !exists {
			return fmt.Errorf("template not found: %s", file.Template)
		}

		// Parse template
		tmpl, err := template.New(filepath.Base(file.Template)).Funcs(funcMap).Parse(templateContent)
		if err != nil {
			return fmt.Errorf("failed to parse template %s: %w", file.Template, err)
		}

		// Process target path
		targetPath := m.processTemplatePath(file.Target, data)
		fullPath := filepath.Join(projectPath, targetPath)

		// Ensure target directory exists
		targetDir := filepath.Dir(fullPath)
		if err := os.MkdirAll(targetDir, 0755); err != nil {
			return fmt.Errorf("failed to create directory for %s: %w", targetPath, err)
		}

		// Create target file
		outFile, err := os.Create(fullPath)
		if err != nil {
			return fmt.Errorf("failed to create file %s: %w", targetPath, err)
		}
		defer outFile.Close()

		// Execute template
		if err := tmpl.Execute(outFile, data); err != nil {
			return fmt.Errorf("failed to execute template for %s: %w", targetPath, err)
		}

		// Set appropriate permissions
		mode := os.FileMode(0644)
		if filepath.Base(fullPath) == "main.go" || filepath.Ext(fullPath) == ".sh" {
			mode = 0755
		}
		if err := os.Chmod(fullPath, mode); err != nil {
			return fmt.Errorf("failed to set permissions for %s: %w", targetPath, err)
		}
	}

	return nil
}

// processTemplatePath processes template variables in paths
func (m *Manager) processTemplatePath(path string, data ProjectData) string {
	// Simple variable replacement for paths
	result := path
	result = replaceVar(result, "ProjectName", data.ProjectName)
	result = replaceVar(result, "ModuleName", data.ModuleName)

	// Also check custom variables
	for key, value := range data.CustomVariables {
		result = replaceVar(result, key, value)
	}

	return result
}

// GenerateComponent generates a component using the layout's component templates
func (m *Manager) GenerateComponent(ctx context.Context, layoutName string, componentType string, componentName string, projectPath string) error {
	// Load layout
	layout, err := m.GetLayout(ctx, layoutName)
	if err != nil {
		return fmt.Errorf("failed to load layout: %w", err)
	}

	// Check if component type exists
	component, exists := layout.Manifest.Components[componentType]
	if !exists {
		return fmt.Errorf("component type '%s' not found in layout", componentType)
	}

	// Get template content
	templateContent, exists := layout.Templates[component.Template]
	if !exists {
		return fmt.Errorf("template not found: %s", component.Template)
	}

	// Create template data
	data := struct {
		ComponentName string
		ModuleName    string
		Name          string
		PackageName   string
		Type          string
	}{
		ComponentName: componentName,
		ModuleName:    "example.com/project", // TODO: Get from project
		Name:          componentName,
		PackageName:   toLower(componentName),
		Type:          componentType,
	}

	// Parse and execute template
	funcMap := template.FuncMap{
		"lower":      toLower,
		"upper":      toUpper,
		"capitalize": capitalize,
		"snake":      toSnakeCase,
		"camel":      toCamelCase,
		"pascal":     toPascalCase,
		"kebab":      toKebabCase,
		"default":    defaultString,
		"snake_case": toSnakeCase,
		"title":      toPascalCase,
		"plural":     pluralize,
	}

	tmpl, err := template.New(componentName).Funcs(funcMap).Parse(templateContent)
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	// Determine target path
	targetDir := filepath.Join(projectPath, component.TargetDir)
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("failed to create target directory: %w", err)
	}

	targetFile := filepath.Join(targetDir, toLower(componentName)+".go")

	// Create file
	outFile, err := os.Create(targetFile)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer outFile.Close()

	// Execute template
	if err := tmpl.Execute(outFile, data); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	fmt.Printf("Generated %s: %s\n", componentType, targetFile)
	return nil
}

// replaceVar replaces {{.Var}} with value in string
func replaceVar(s, varName, value string) string {
	// Handle both {{.Var}} and {{Var}} formats
	s = replaceAll(s, "{{."+varName+"}}", value)
	s = replaceAll(s, "{{"+varName+"}}", value)
	return s
}

// Simple string replacement helper
func replaceAll(s, old, new string) string {
	result := s
	for {
		index := indexOf(result, old)
		if index == -1 {
			break
		}
		result = result[:index] + new + result[index+len(old):]
	}
	return result
}

// indexOf returns the index of substr in s, or -1 if not found
func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
