package layout

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Cache manages layout caching
type Cache struct {
	dir      string
	ttl      time.Duration
	mu       sync.RWMutex
	layouts  map[string]*Layout
	metadata map[string]*CacheMetadata
}

// CacheMetadata stores cache metadata
type CacheMetadata struct {
	Name      string    `json:"name"`
	Version   string    `json:"version"`
	CachedAt  time.Time `json:"cached_at"`
	ExpiresAt time.Time `json:"expires_at"`
	Source    string    `json:"source"`
}

// NewCache creates a new cache instance
func NewCache(dir string, ttl time.Duration) (*Cache, error) {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create cache directory: %w", err)
	}

	cache := &Cache{
		dir:      dir,
		ttl:      ttl,
		layouts:  make(map[string]*Layout),
		metadata: make(map[string]*CacheMetadata),
	}

	// Load existing cache metadata
	if err := cache.loadMetadata(); err != nil {
		// Non-fatal error, just start with empty cache
		fmt.Printf("Warning: failed to load cache metadata: %v\n", err)
	}

	return cache, nil
}

// Get retrieves a layout from cache
func (c *Cache) Get(name string) (*Layout, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	layout, exists := c.layouts[name]
	if !exists {
		return nil, false
	}

	// Check if cache is expired
	meta, hasMeta := c.metadata[name]
	if hasMeta && time.Now().After(meta.ExpiresAt) {
		return nil, false
	}

	return layout, true
}

// Set stores a layout in cache
func (c *Cache) Set(name string, layout *Layout) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.layouts[name] = layout

	// Create metadata
	now := time.Now()
	c.metadata[name] = &CacheMetadata{
		Name:      name,
		Version:   layout.Version,
		CachedAt:  now,
		ExpiresAt: now.Add(c.ttl),
		Source:    layout.Source.Type,
	}

	// Save metadata to disk
	return c.saveMetadata()
}

// Remove removes a layout from cache
func (c *Cache) Remove(name string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.layouts, name)
	delete(c.metadata, name)

	// Remove from disk
	layoutDir := filepath.Join(c.dir, "layouts", name)
	if err := os.RemoveAll(layoutDir); err != nil {
		return fmt.Errorf("failed to remove cached layout: %w", err)
	}

	return c.saveMetadata()
}

// Clear removes all cached layouts
func (c *Cache) Clear() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.layouts = make(map[string]*Layout)
	c.metadata = make(map[string]*CacheMetadata)

	// Remove all cached files
	if err := os.RemoveAll(c.dir); err != nil {
		return fmt.Errorf("failed to clear cache: %w", err)
	}

	// Recreate cache directory
	if err := os.MkdirAll(c.dir, 0755); err != nil {
		return fmt.Errorf("failed to recreate cache directory: %w", err)
	}

	return nil
}

// List returns all cached layouts
func (c *Cache) List() []CacheMetadata {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var list []CacheMetadata
	for _, meta := range c.metadata {
		list = append(list, *meta)
	}
	return list
}

// Refresh reloads cache from disk
func (c *Cache) Refresh() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Clear in-memory cache
	c.layouts = make(map[string]*Layout)

	// Reload metadata
	return c.loadMetadata()
}

// loadMetadata loads cache metadata from disk
func (c *Cache) loadMetadata() error {
	metaPath := filepath.Join(c.dir, "metadata.json")

	data, err := os.ReadFile(metaPath)
	if err != nil {
		if os.IsNotExist(err) {
			// No metadata file yet, that's okay
			return nil
		}
		return err
	}

	var metadata map[string]*CacheMetadata
	if err := json.Unmarshal(data, &metadata); err != nil {
		return fmt.Errorf("failed to parse cache metadata: %w", err)
	}

	c.metadata = metadata

	// Remove expired entries
	now := time.Now()
	for name, meta := range c.metadata {
		if now.After(meta.ExpiresAt) {
			delete(c.metadata, name)
		}
	}

	return nil
}

// saveMetadata saves cache metadata to disk
func (c *Cache) saveMetadata() error {
	metaPath := filepath.Join(c.dir, "metadata.json")

	data, err := json.MarshalIndent(c.metadata, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal cache metadata: %w", err)
	}

	if err := os.WriteFile(metaPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write cache metadata: %w", err)
	}

	return nil
}

// SaveLayout saves a layout to disk cache
func (c *Cache) SaveLayout(layout *Layout) error {
	layoutDir := filepath.Join(c.dir, "layouts", layout.Name)
	if err := os.MkdirAll(layoutDir, 0755); err != nil {
		return fmt.Errorf("failed to create layout cache directory: %w", err)
	}

	// Save manifest
	manifestPath := filepath.Join(layoutDir, "manifest.json")
	manifestData, err := json.MarshalIndent(layout.Manifest, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal manifest: %w", err)
	}
	if err := os.WriteFile(manifestPath, manifestData, 0644); err != nil {
		return fmt.Errorf("failed to write manifest: %w", err)
	}

	// Save templates
	templatesDir := filepath.Join(layoutDir, "templates")
	if err := os.MkdirAll(templatesDir, 0755); err != nil {
		return fmt.Errorf("failed to create templates directory: %w", err)
	}

	for path, content := range layout.Templates {
		templatePath := filepath.Join(templatesDir, path)
		templateDir := filepath.Dir(templatePath)

		if err := os.MkdirAll(templateDir, 0755); err != nil {
			return fmt.Errorf("failed to create template directory: %w", err)
		}

		if err := os.WriteFile(templatePath, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to write template %s: %w", path, err)
		}
	}

	// Save source info
	sourcePath := filepath.Join(layoutDir, "source.json")
	sourceData, err := json.MarshalIndent(layout.Source, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal source: %w", err)
	}
	if err := os.WriteFile(sourcePath, sourceData, 0644); err != nil {
		return fmt.Errorf("failed to write source: %w", err)
	}

	return nil
}

// LoadLayout loads a layout from disk cache
func (c *Cache) LoadLayout(name string) (*Layout, error) {
	layoutDir := filepath.Join(c.dir, "layouts", name)

	// Check if directory exists
	if _, err := os.Stat(layoutDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("layout not found in cache")
	}

	// Load manifest
	manifestPath := filepath.Join(layoutDir, "manifest.json")
	manifestData, err := os.ReadFile(manifestPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read manifest: %w", err)
	}

	var manifest LayoutManifest
	if err := json.Unmarshal(manifestData, &manifest); err != nil {
		return nil, fmt.Errorf("failed to parse manifest: %w", err)
	}

	// Load source
	sourcePath := filepath.Join(layoutDir, "source.json")
	sourceData, err := os.ReadFile(sourcePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read source: %w", err)
	}

	var source LayoutSource
	if err := json.Unmarshal(sourceData, &source); err != nil {
		return nil, fmt.Errorf("failed to parse source: %w", err)
	}

	// Load templates
	templates := make(map[string]string)
	templatesDir := filepath.Join(layoutDir, "templates")

	err = filepath.Walk(templatesDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(templatesDir, path)
		if err != nil {
			return err
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read template %s: %w", relPath, err)
		}

		templates[filepath.ToSlash(relPath)] = string(content)
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to load templates: %w", err)
	}

	return &Layout{
		Name:      name,
		Version:   manifest.Version,
		Source:    source,
		Manifest:  &manifest,
		Templates: templates,
		LoadedAt:  time.Now(),
	}, nil
}
