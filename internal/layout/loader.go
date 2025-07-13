package layout

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// Loader handles loading layouts from various sources
type Loader struct {
	registry *Registry
	cache    *Cache
	client   *http.Client
}

// NewLoader creates a new layout loader
func NewLoader(registry *Registry, cache *Cache) *Loader {
	return &Loader{
		registry: registry,
		cache:    cache,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Load loads a layout by name
func (l *Loader) Load(ctx context.Context, name string) (*Layout, error) {
	// Check cache first
	if layout, ok := l.cache.Get(name); ok {
		return layout, nil
	}

	// Find layout source
	source, err := l.registry.GetLayoutSource(name)
	if err != nil {
		return nil, fmt.Errorf("layout not found: %w", err)
	}

	// Load based on source type
	var layout *Layout
	switch source.Type {
	case "local":
		layout, err = l.loadLocal(ctx, name, source)
	case "remote":
		layout, err = l.loadRemote(ctx, name, source)
	case "github":
		layout, err = l.loadGitHub(ctx, name, source)
	default:
		return nil, fmt.Errorf("unsupported source type: %s", source.Type)
	}

	if err != nil {
		return nil, err
	}

	// Cache the loaded layout
	l.cache.Set(name, layout)
	return layout, nil
}

// loadLocal loads a layout from the local filesystem
func (l *Loader) loadLocal(ctx context.Context, name string, source LayoutSource) (*Layout, error) {
	basePath := expandPath(source.Location)

	// Load manifest
	manifestPath := filepath.Join(basePath, "layout.manifest.yaml")
	manifest, err := l.loadManifest(manifestPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load manifest: %w", err)
	}

	// Load all template files
	templates := make(map[string]string)

	// Load project templates
	projectDir := filepath.Join(basePath, "project")
	if err := l.loadTemplatesFromDir(projectDir, "project", templates); err != nil {
		return nil, fmt.Errorf("failed to load project templates: %w", err)
	}

	// Load component templates
	componentsDir := filepath.Join(basePath, "components")
	if err := l.loadTemplatesFromDir(componentsDir, "components", templates); err != nil {
		// Component templates are optional
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("failed to load component templates: %w", err)
		}
	}

	return &Layout{
		Name:      name,
		Version:   manifest.Version,
		Source:    source,
		Manifest:  manifest,
		Templates: templates,
		LoadedAt:  time.Now(),
	}, nil
}

// loadRemote loads a layout from a remote URL
func (l *Loader) loadRemote(ctx context.Context, name string, source LayoutSource) (*Layout, error) {
	// Create cache directory for this layout
	cacheDir := filepath.Join(l.cache.dir, "remote", name, source.Ref)
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create cache directory: %w", err)
	}

	// Check if already downloaded
	manifestPath := filepath.Join(cacheDir, "layout.manifest.yaml")
	if fileExists(manifestPath) {
		// Already downloaded, load from cache
		source.Location = cacheDir
		return l.loadLocal(ctx, name, source)
	}

	// Download layout archive
	archivePath := filepath.Join(cacheDir, "layout.tar.gz")
	if err := l.downloadFile(ctx, source.Location, archivePath); err != nil {
		return nil, fmt.Errorf("failed to download layout: %w", err)
	}

	// Verify checksum if provided
	if source.Checksum != "" {
		if err := verifySHA256(archivePath, source.Checksum); err != nil {
			os.Remove(archivePath)
			return nil, fmt.Errorf("checksum verification failed: %w", err)
		}
	}

	// Extract archive
	if err := extractTarGz(archivePath, cacheDir); err != nil {
		return nil, fmt.Errorf("failed to extract layout: %w", err)
	}

	// Remove archive after extraction
	os.Remove(archivePath)

	// Load from extracted files
	source.Location = cacheDir
	return l.loadLocal(ctx, name, source)
}

// loadGitHub loads a layout from a GitHub repository
func (l *Loader) loadGitHub(ctx context.Context, name string, source LayoutSource) (*Layout, error) {
	// Create cache directory for this layout
	ref := source.Ref
	if ref == "" {
		ref = "main"
	}

	cacheDir := filepath.Join(l.cache.dir, "github", strings.ReplaceAll(source.Location, "/", "_"), ref)

	// Check if already cloned
	manifestPath := filepath.Join(cacheDir, "layout.manifest.yaml")
	if fileExists(manifestPath) {
		// Already cloned, load from cache
		source.Location = cacheDir
		return l.loadLocal(ctx, name, source)
	}

	// Clone the repository
	repoURL := fmt.Sprintf("https://github.com/%s.git", source.Location)
	if err := cloneGitRepository(repoURL, ref, cacheDir); err != nil {
		// Fallback to archive download
		archiveURL := fmt.Sprintf("https://github.com/%s/archive/refs/heads/%s.tar.gz", source.Location, ref)
		source.Location = archiveURL
		source.Type = "remote"
		return l.loadRemote(ctx, name, source)
	}

	// Load from cloned repository
	source.Location = cacheDir
	return l.loadLocal(ctx, name, source)
}

// loadManifest loads and parses a layout manifest file
func (l *Loader) loadManifest(path string) (*LayoutManifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var manifest LayoutManifest
	if err := yaml.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("failed to parse manifest: %w", err)
	}

	// Validate manifest
	if manifest.Name == "" {
		return nil, fmt.Errorf("manifest missing required field: name")
	}
	if manifest.Version == "" {
		return nil, fmt.Errorf("manifest missing required field: version")
	}

	return &manifest, nil
}

// loadTemplatesFromDir recursively loads all template files from a directory
func (l *Loader) loadTemplatesFromDir(dir, prefix string, templates map[string]string) error {
	if !isDir(dir) {
		return os.ErrNotExist
	}

	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() || !strings.HasSuffix(path, ".tmpl") {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read template %s: %w", path, err)
		}

		// Calculate relative path for template key
		relPath, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}

		key := filepath.Join(prefix, relPath)
		key = filepath.ToSlash(key) // Ensure consistent path separators
		templates[key] = string(content)

		return nil
	})
}

// downloadFile downloads a file from a URL
func (l *Loader) downloadFile(ctx context.Context, url, dest string) error {
	// Ensure destination directory exists
	if err := ensureDir(filepath.Dir(dest)); err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return err
	}

	// Add user agent
	req.Header.Set("User-Agent", "Foundry-CLI/1.0")

	resp, err := l.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}
