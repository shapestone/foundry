// internal/scaffolder/interfaces.go
package scaffolder

import (
	"context"
	"io"
)

// Scaffolder is the main interface for code generation operations
type Scaffolder interface {
	CreateHandler(ctx context.Context, spec *HandlerSpec) (*Result, error)
	CreateModel(ctx context.Context, spec *ModelSpec) (*Result, error)
	CreateMiddleware(ctx context.Context, spec *MiddlewareSpec) (*Result, error)
	CreateDatabase(ctx context.Context, spec *DatabaseSpec) (*Result, error)
	WireHandler(ctx context.Context, spec *WireSpec) (*Result, error)
}

// HandlerSpec defines the specification for creating a handler
type HandlerSpec struct {
	Name        string            `json:"name"`
	Type        string            `json:"type"` // "REST", "GraphQL", etc.
	AutoWire    bool              `json:"auto_wire"`
	DryRun      bool              `json:"dry_run"`
	ProjectRoot string            `json:"project_root"`
	Module      string            `json:"module"`
	Metadata    map[string]string `json:"metadata"`
}

// ModelSpec defines the specification for creating a model
type ModelSpec struct {
	Name              string            `json:"name"`
	Fields            []FieldSpec       `json:"fields"`
	IncludeTimestamps bool              `json:"include_timestamps"`
	IncludeValidation bool              `json:"include_validation"`
	ProjectRoot       string            `json:"project_root"`
	Module            string            `json:"module"`
	Metadata          map[string]string `json:"metadata"`
}

// FieldSpec defines a field in a model
type FieldSpec struct {
	Name     string            `json:"name"`
	Type     string            `json:"type"`
	Tags     map[string]string `json:"tags"`
	Required bool              `json:"required"`
}

// MiddlewareSpec defines the specification for creating middleware
type MiddlewareSpec struct {
	Name        string            `json:"name"`
	Type        string            `json:"type"` // "auth", "cors", "ratelimit", etc.
	AutoWire    bool              `json:"auto_wire"`
	ProjectRoot string            `json:"project_root"`
	Module      string            `json:"module"`
	Metadata    map[string]string `json:"metadata"`
}

// DatabaseSpec defines the specification for creating database support
type DatabaseSpec struct {
	Type           string            `json:"type"` // "postgres", "mysql", "sqlite", "mongodb"
	WithMigrations bool              `json:"with_migrations"`
	WithDocker     bool              `json:"with_docker"`
	ProjectRoot    string            `json:"project_root"`
	Module         string            `json:"module"`
	Metadata       map[string]string `json:"metadata"`
}

// WireSpec defines the specification for wiring components
type WireSpec struct {
	ComponentType string            `json:"component_type"` // "handler", "middleware"
	ComponentName string            `json:"component_name"`
	ProjectRoot   string            `json:"project_root"`
	Module        string            `json:"module"`
	Metadata      map[string]string `json:"metadata"`
}

// Result represents the outcome of a scaffolding operation
type Result struct {
	FilesCreated []string          `json:"files_created"`
	FilesUpdated []string          `json:"files_updated"`
	Changes      []string          `json:"changes"`
	Warnings     []string          `json:"warnings"`
	Success      bool              `json:"success"`
	Message      string            `json:"message"`
	Metadata     map[string]string `json:"metadata"`
}

// FileOperation represents a file system operation
type FileOperation struct {
	Type    OperationType `json:"type"`
	Path    string        `json:"path"`
	Content []byte        `json:"content"`
	Mode    uint32        `json:"mode"`
}

// OperationType defines the type of file operation
type OperationType string

const (
	OperationCreate OperationType = "create"
	OperationUpdate OperationType = "update"
	OperationDelete OperationType = "delete"
)

// Dependencies that can be injected for testing

// FileSystem abstracts file system operations
type FileSystem interface {
	Exists(path string) bool
	ReadFile(path string) ([]byte, error)
	WriteFile(path string, data []byte, perm uint32) error
	MkdirAll(path string, perm uint32) error
	Remove(path string) error
	Stat(path string) (FileInfo, error)
}

// FileInfo represents file information
type FileInfo interface {
	Name() string
	Size() int64
	Mode() uint32
	IsDir() bool
}

// TemplateRenderer handles template operations
type TemplateRenderer interface {
	LoadTemplate(name string) (Template, error)
	RenderTemplate(tmpl Template, data interface{}) (string, error)
}

// Template represents a loaded template
type Template interface {
	Name() string
	Execute(wr io.Writer, data interface{}) error
}

// ProjectAnalyzer analyzes project structure and configuration
type ProjectAnalyzer interface {
	GetModuleName(projectRoot string) (string, error)
	GetProjectName(projectRoot string) (string, error)
	ValidateProjectStructure(projectRoot string) error
	IsGoProject(projectRoot string) bool
}

// UserInteraction handles user prompts and confirmations
type UserInteraction interface {
	Confirm(message string) bool
	ShowPreview(title string, changes []string, message string) bool
}

// ValidationError represents a validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func (e *ValidationError) Error() string {
	return e.Field + ": " + e.Message
}

// ValidationErrors represents multiple validation errors
type ValidationErrors []ValidationError

func (e ValidationErrors) Error() string {
	if len(e) == 0 {
		return ""
	}
	if len(e) == 1 {
		return e[0].Error()
	}
	return e[0].Error() + " (and " + string(rune(len(e)-1)) + " more errors)"
}

// HasErrors returns true if there are validation errors
func (e ValidationErrors) HasErrors() bool {
	return len(e) > 0
}
