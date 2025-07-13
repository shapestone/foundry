package layout

import (
	"time"
)

// LayoutManifest represents the configuration for a layout
type LayoutManifest struct {
	Name              string                       `yaml:"name" json:"name"`
	Version           string                       `yaml:"version" json:"version"`
	Author            string                       `yaml:"author,omitempty" json:"author,omitempty"`
	Description       string                       `yaml:"description" json:"description"`
	MinFoundryVersion string                       `yaml:"min_foundry_version,omitempty" json:"min_foundry_version,omitempty"`
	Structure         LayoutStructure              `yaml:"structure" json:"structure"`
	Dependencies      []string                     `yaml:"dependencies,omitempty" json:"dependencies,omitempty"`
	Features          []string                     `yaml:"features,omitempty" json:"features,omitempty"`
	Variables         []LayoutVariable             `yaml:"variables,omitempty" json:"variables,omitempty"`
	Components        map[string]ComponentTemplate `yaml:"components,omitempty" json:"components,omitempty"`
}

// LayoutStructure defines the directory and file structure
type LayoutStructure struct {
	Directories []DirectorySpec `yaml:"directories" json:"directories"`
	Files       []FileSpec      `yaml:"files" json:"files"`
}

// DirectorySpec defines a directory to create
type DirectorySpec struct {
	Path        string `yaml:"path" json:"path"`
	Description string `yaml:"description,omitempty" json:"description,omitempty"`
}

// FileSpec defines a file to generate from a template
type FileSpec struct {
	Template string `yaml:"template" json:"template"`
	Target   string `yaml:"target" json:"target"`
}

// LayoutVariable defines a configurable variable for the layout
type LayoutVariable struct {
	Name        string `yaml:"name" json:"name"`
	Default     string `yaml:"default,omitempty" json:"default,omitempty"`
	Description string `yaml:"description,omitempty" json:"description,omitempty"`
	Required    bool   `yaml:"required,omitempty" json:"required,omitempty"`
}

// ComponentTemplate defines a component that can be added to the project
type ComponentTemplate struct {
	Template  string `yaml:"template" json:"template"`
	TargetDir string `yaml:"target_dir" json:"target_dir"`
}

// LayoutSource represents where a layout comes from
type LayoutSource struct {
	Type     string `yaml:"type" json:"type"`                   // local, remote, github
	Location string `yaml:"location" json:"location"`           // path, URL, or repo
	Ref      string `yaml:"ref,omitempty" json:"ref,omitempty"` // version, tag, or branch
	Checksum string `yaml:"checksum,omitempty" json:"checksum,omitempty"`
}

// LayoutRegistry represents the registry configuration
type LayoutRegistry struct {
	Version    string                    `yaml:"version" json:"version"`
	Registries map[string]RegistryConfig `yaml:"registries" json:"registries"`
	LocalPaths []string                  `yaml:"local_paths" json:"local_paths"`
	Cache      CacheConfig               `yaml:"cache" json:"cache"`
	Security   SecurityConfig            `yaml:"security" json:"security"`
}

// RegistryConfig represents a remote registry
type RegistryConfig struct {
	URL     string `yaml:"url" json:"url"`
	Trusted bool   `yaml:"trusted" json:"trusted"`
}

// CacheConfig represents cache settings
type CacheConfig struct {
	Directory string        `yaml:"directory" json:"directory"`
	TTL       time.Duration `yaml:"ttl" json:"ttl"`
}

// SecurityConfig represents security settings
type SecurityConfig struct {
	VerifyChecksums bool `yaml:"verify_checksums" json:"verify_checksums"`
	AllowUntrusted  bool `yaml:"allow_untrusted" json:"allow_untrusted"`
}

// Layout represents a loaded layout with its templates
type Layout struct {
	Name      string
	Version   string
	Source    LayoutSource
	Manifest  *LayoutManifest
	Templates map[string]string // template path -> content
	LoadedAt  time.Time
}

// LayoutListEntry represents a layout in the list
type LayoutListEntry struct {
	Name        string       `yaml:"name" json:"name"`
	Version     string       `yaml:"version" json:"version"`
	Description string       `yaml:"description" json:"description"`
	Source      LayoutSource `yaml:"source" json:"source"`
	Installed   bool         `yaml:"installed" json:"installed"`
	UpdatedAt   time.Time    `yaml:"updated_at,omitempty" json:"updated_at,omitempty"`
}
