package layout

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"io/fs"
)

// EmbeddedTemplateProvider defines interface for accessing embedded templates
type EmbeddedTemplateProvider interface {
	ReadDir(name string) ([]fs.DirEntry, error)
	ReadFile(name string) ([]byte, error)
}

// embeddedTemplates holds the reference to the actual embedded filesystem
var embeddedTemplates EmbeddedTemplateProvider

// SetEmbeddedTemplates sets the embedded template provider
// This will be called from main.go to inject the foundry.Templates
func SetEmbeddedTemplates(provider EmbeddedTemplateProvider) {
	embeddedTemplates = provider
}

// GetEmbeddedLayouts returns all embedded layout names
func GetEmbeddedLayouts() []string {
	if embeddedTemplates == nil {
		return []string{}
	}

	entries, err := embeddedTemplates.ReadDir("templates")
	if err != nil {
		return []string{}
	}

	var layouts []string
	for _, entry := range entries {
		if entry.IsDir() {
			// Check if it has a layout.manifest.yaml
			manifestPath := "templates/" + entry.Name() + "/layout.manifest.yaml"
			if _, err := embeddedTemplates.ReadFile(manifestPath); err == nil {
				layouts = append(layouts, entry.Name())
			}
		}
	}
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
	var layouts []LayoutListEntry

	embeddedLayouts := GetEmbeddedLayouts()
	for _, name := range embeddedLayouts {
		manifest, err := ParseEmbeddedManifest(name)
		if err != nil {
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
		})
	}

	return layouts
}
