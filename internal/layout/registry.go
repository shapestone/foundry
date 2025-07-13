package layout

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"gopkg.in/yaml.v3"
)

// Registry manages layout sources and configurations
type Registry struct {
	configPath string
	config     *LayoutRegistry
	layouts    map[string]LayoutListEntry
	mu         sync.RWMutex
}

// NewRegistry creates a new registry instance
func NewRegistry(configPath string) (*Registry, error) {
	registry := &Registry{
		configPath: configPath,
		layouts:    make(map[string]LayoutListEntry),
	}

	// Load or create default config
	if err := registry.loadConfig(); err != nil {
		// Create default config if it doesn't exist
		registry.config = registry.defaultConfig()
		if err := registry.saveConfig(); err != nil {
			return nil, fmt.Errorf("failed to save default config: %w", err)
		}
	}

	// Load layout index
	if err := registry.loadLayouts(); err != nil {
		// Non-fatal, start with empty index
		fmt.Printf("Warning: failed to load layout index: %v\n", err)
	}

	return registry, nil
}

// defaultConfig returns the default registry configuration
func (r *Registry) defaultConfig() *LayoutRegistry {
	homeDir, _ := os.UserHomeDir()
	cacheDir := filepath.Join(homeDir, ".foundry", "cache", "layouts")

	return &LayoutRegistry{
		Version: "1.0",
		Registries: map[string]RegistryConfig{
			"official": {
				URL:     "https://registry.foundry.dev/layouts",
				Trusted: true,
			},
			"community": {
				URL:     "https://community.foundry.dev/layouts",
				Trusted: false,
			},
		},
		LocalPaths: []string{
			filepath.Join(homeDir, ".foundry", "layouts"),
			"/usr/share/foundry/layouts",
			"./templates/layouts", // For backward compatibility
		},
		Cache: CacheConfig{
			Directory: cacheDir,
			TTL:       24 * time.Hour,
		},
		Security: SecurityConfig{
			VerifyChecksums: true,
			AllowUntrusted:  false,
		},
	}
}

// loadConfig loads the registry configuration
func (r *Registry) loadConfig() error {
	data, err := os.ReadFile(r.configPath)
	if err != nil {
		return err
	}

	var config LayoutRegistry
	if err := yaml.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("failed to parse config: %w", err)
	}

	r.config = &config
	return nil
}

// saveConfig saves the registry configuration
func (r *Registry) saveConfig() error {
	// Ensure directory exists
	dir := filepath.Dir(r.configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := yaml.Marshal(r.config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(r.configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}

// loadLayouts loads the layout index
func (r *Registry) loadLayouts() error {
	// Load from local paths
	for _, path := range r.config.LocalPaths {
		if err := r.scanLocalPath(path); err != nil {
			// Non-fatal, continue with other paths
			continue
		}
	}

	// Load cached remote layouts
	indexPath := filepath.Join(r.config.Cache.Directory, "index.json")
	if data, err := os.ReadFile(indexPath); err == nil {
		var remoteLayouts map[string]LayoutListEntry
		if err := json.Unmarshal(data, &remoteLayouts); err == nil {
			r.mu.Lock()
			for name, entry := range remoteLayouts {
				r.layouts[name] = entry
			}
			r.mu.Unlock()
		}
	}

	return nil
}

// scanLocalPath scans a local directory for layouts
func (r *Registry) scanLocalPath(basePath string) error {
	// Expand home directory
	if basePath[0] == '~' {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		basePath = filepath.Join(homeDir, basePath[1:])
	}

	// Check if path exists
	if _, err := os.Stat(basePath); os.IsNotExist(err) {
		return nil // Not an error, just skip
	}

	// Scan for layout directories
	entries, err := os.ReadDir(basePath)
	if err != nil {
		return fmt.Errorf("failed to read directory %s: %w", basePath, err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		layoutPath := filepath.Join(basePath, entry.Name())
		manifestPath := filepath.Join(layoutPath, "layout.manifest.yaml")

		// Check if manifest exists
		if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
			continue
		}

		// Load manifest
		data, err := os.ReadFile(manifestPath)
		if err != nil {
			continue
		}

		var manifest LayoutManifest
		if err := yaml.Unmarshal(data, &manifest); err != nil {
			continue
		}

		// Add to registry
		r.mu.Lock()
		r.layouts[entry.Name()] = LayoutListEntry{
			Name:        entry.Name(),
			Version:     manifest.Version,
			Description: manifest.Description,
			Source: LayoutSource{
				Type:     "local",
				Location: layoutPath,
			},
			Installed: true,
			UpdatedAt: time.Now(),
		}
		r.mu.Unlock()
	}

	return nil
}

// GetLayoutSource returns the source information for a layout
func (r *Registry) GetLayoutSource(name string) (LayoutSource, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	entry, exists := r.layouts[name]
	if !exists {
		return LayoutSource{}, fmt.Errorf("layout '%s' not found", name)
	}

	return entry.Source, nil
}

// ListLayouts returns all available layouts
func (r *Registry) ListLayouts() []LayoutListEntry {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var list []LayoutListEntry
	for _, entry := range r.layouts {
		list = append(list, entry)
	}

	return list
}

// AddLayout adds a new layout to the registry
func (r *Registry) AddLayout(entry LayoutListEntry) error {
	r.mu.Lock()
	r.layouts[entry.Name] = entry
	r.mu.Unlock()

	// Save updated index
	return r.saveIndex()
}

// RemoveLayout removes a layout from the registry
func (r *Registry) RemoveLayout(name string) error {
	r.mu.Lock()
	delete(r.layouts, name)
	r.mu.Unlock()

	// Save updated index
	return r.saveIndex()
}

// UpdateLayout updates a layout in the registry
func (r *Registry) UpdateLayout(name string, entry LayoutListEntry) error {
	r.mu.Lock()
	r.layouts[name] = entry
	r.mu.Unlock()

	// Save updated index
	return r.saveIndex()
}

// saveIndex saves the layout index to disk
func (r *Registry) saveIndex() error {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Filter out local layouts (they're scanned dynamically)
	remoteLayouts := make(map[string]LayoutListEntry)
	for name, entry := range r.layouts {
		if entry.Source.Type != "local" {
			remoteLayouts[name] = entry
		}
	}

	indexPath := filepath.Join(r.config.Cache.Directory, "index.json")
	dir := filepath.Dir(indexPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create index directory: %w", err)
	}

	data, err := json.MarshalIndent(remoteLayouts, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal index: %w", err)
	}

	if err := os.WriteFile(indexPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write index: %w", err)
	}

	return nil
}

// RefreshRemoteRegistries fetches latest layout lists from remote registries
func (r *Registry) RefreshRemoteRegistries() error {
	// TODO: Implement fetching from remote registries
	// This would:
	// 1. Fetch layout lists from each configured registry
	// 2. Merge with local layouts
	// 3. Update the index
	return nil
}

// GetConfig returns the current registry configuration
func (r *Registry) GetConfig() *LayoutRegistry {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.config
}

// HasLayout checks if a layout exists in the registry
func (r *Registry) HasLayout(name string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, exists := r.layouts[name]
	return exists
}

// GetLayout returns a specific layout entry
func (r *Registry) GetLayout(name string) (LayoutListEntry, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	entry, exists := r.layouts[name]
	if !exists {
		return LayoutListEntry{}, fmt.Errorf("layout '%s' not found", name)
	}

	return entry, nil
}
