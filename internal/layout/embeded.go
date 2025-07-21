package layout

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
	"time"
)

// EmbeddedTemplateProvider interface for embedded template access
type EmbeddedTemplateProvider interface {
	ReadFile(name string) ([]byte, error)
	ReadDir(name string) ([]os.DirEntry, error)
}

// Global embedded templates provider
var embeddedTemplates EmbeddedTemplateProvider

// SetEmbeddedTemplates sets the embedded templates provider
func SetEmbeddedTemplates(provider EmbeddedTemplateProvider) {
	embeddedTemplates = provider
	fmt.Printf("DEBUG: SetEmbeddedTemplates called with provider: %T\n", provider)
}

// GetEmbeddedLayouts returns all embedded layout names
func GetEmbeddedLayouts() []string {
	fmt.Printf("DEBUG: GetEmbeddedLayouts called\n")

	if embeddedTemplates == nil {
		fmt.Printf("DEBUG: embeddedTemplates is nil, returning empty\n")
		return []string{}
	}

	fmt.Printf("DEBUG: embeddedTemplates is not nil, reading templates dir\n")
	entries, err := embeddedTemplates.ReadDir("templates")
	if err != nil {
		fmt.Printf("DEBUG: Error reading templates dir: %v\n", err)
		return []string{}
	}

	fmt.Printf("DEBUG: Found %d entries in templates dir\n", len(entries))
	var layouts []string
	for _, entry := range entries {
		fmt.Printf("DEBUG: Checking entry: %s (isDir: %v)\n", entry.Name(), entry.IsDir())
		if entry.IsDir() {
			// Check if it has a layout.manifest.yaml
			manifestPath := "templates/" + entry.Name() + "/layout.manifest.yaml"
			fmt.Printf("DEBUG: Looking for manifest: %s\n", manifestPath)
			if _, err := embeddedTemplates.ReadFile(manifestPath); err == nil {
				fmt.Printf("DEBUG: ✅ Found manifest for: %s\n", entry.Name())
				layouts = append(layouts, entry.Name())
			} else {
				fmt.Printf("DEBUG: ❌ No manifest for %s: %v\n", entry.Name(), err)
			}
		}
	}
	fmt.Printf("DEBUG: Returning %d layouts: %v\n", len(layouts), layouts)
	return layouts
}

// GetEmbeddedLayoutManifest reads a layout manifest from embedded templates
func GetEmbeddedLayoutManifest(layoutName string) ([]byte, error) {
	if embeddedTemplates == nil {
		return nil, fmt.Errorf("embedded templates not initialized")
	}

	manifestPath := "templates/" + layoutName + "/layout.manifest.yaml"
	return embeddedTemplates.ReadFile(manifestPath)
}

// GetEmbeddedTemplateFile reads a template file from embedded templates
func GetEmbeddedTemplateFile(layoutName, templatePath string) ([]byte, error) {
	if embeddedTemplates == nil {
		return nil, fmt.Errorf("embedded templates not initialized")
	}

	fullPath := "templates/" + layoutName + "/" + templatePath
	return embeddedTemplates.ReadFile(fullPath)
}

// ParseEmbeddedManifest parses a layout manifest from embedded templates
func ParseEmbeddedManifest(layoutName string) (*LayoutManifest, error) {
	manifestData, err := GetEmbeddedLayoutManifest(layoutName)
	if err != nil {
		return nil, fmt.Errorf("failed to read manifest for layout %s: %w", layoutName, err)
	}

	var manifest LayoutManifest
	if err := yaml.Unmarshal(manifestData, &manifest); err != nil {
		return nil, fmt.Errorf("failed to parse manifest for layout %s: %w", layoutName, err)
	}

	return &manifest, nil
}

// GetEmbeddedLayoutList returns layout list entries for all embedded layouts
func GetEmbeddedLayoutList() []LayoutListEntry {
	fmt.Printf("DEBUG: GetEmbeddedLayoutList called\n")
	var layouts []LayoutListEntry

	embeddedLayouts := GetEmbeddedLayouts()
	fmt.Printf("DEBUG: GetEmbeddedLayoutList got %d layouts from GetEmbeddedLayouts: %v\n", len(embeddedLayouts), embeddedLayouts)

	for _, name := range embeddedLayouts {
		manifest, err := ParseEmbeddedManifest(name)
		if err != nil {
			fmt.Printf("DEBUG: Failed to parse manifest for %s: %v\n", name, err)
			continue
		}

		layouts = append(layouts, LayoutListEntry{
			Name:        name,
			Version:     manifest.Version,
			Description: manifest.Description,
			Source: LayoutSource{
				Type:     "embedded",
				Location: "built-in",
			},
			Installed: true,
			UpdatedAt: time.Now(),
		})
	}

	fmt.Printf("DEBUG: GetEmbeddedLayoutList returning %d layouts\n", len(layouts))
	return layouts
}
