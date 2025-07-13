package templating

import (
	"fmt"
	"strings"
	"text/template"
)

// DefaultTemplateManager implements TemplateManager interface
type DefaultTemplateManager struct {
	config *TemplateConfig
	cache  map[string]*Template
}

// NewDefaultTemplateManager creates a new default template manager
func NewDefaultTemplateManager(config *TemplateConfig) TemplateManager {
	return &DefaultTemplateManager{
		config: config,
		cache:  make(map[string]*Template),
	}
}

// LoadAndRender loads and renders a template
func (tm *DefaultTemplateManager) LoadAndRender(layoutName, category, templateName string, data TemplateData) (string, error) {
	content := getEmbeddedTemplate(templateName)
	if content == "" {
		return "", fmt.Errorf("template not found: %s", templateName)
	}

	tmpl, err := template.New(templateName).Parse(content)
	if err != nil {
		return "", fmt.Errorf("parsing template: %w", err)
	}

	var buf strings.Builder
	err = tmpl.Execute(&buf, data)
	if err != nil {
		return "", fmt.Errorf("executing template: %w", err)
	}

	return buf.String(), nil
}

// GetTemplate gets a parsed template - returns *Template not *template.Template
func (tm *DefaultTemplateManager) GetTemplate(layoutName, category, templateName string) (*Template, error) {
	content := getEmbeddedTemplate(templateName)
	if content == "" {
		return nil, fmt.Errorf("template not found: %s", templateName)
	}

	parsed, err := template.New(templateName).Parse(content)
	if err != nil {
		return nil, fmt.Errorf("parsing template: %w", err)
	}

	return &Template{
		Name:     templateName,
		Content:  content,
		Parsed:   parsed,
		Layout:   layoutName,
		Category: category,
	}, nil
}

// RenderTemplate renders a loaded template
func (tm *DefaultTemplateManager) RenderTemplate(tmpl *Template, data TemplateData) (string, error) {
	if tmpl.Parsed == nil {
		return "", fmt.Errorf("template not parsed")
	}

	var buf strings.Builder
	err := tmpl.Parsed.Execute(&buf, data)
	if err != nil {
		return "", fmt.Errorf("executing template: %w", err)
	}

	return buf.String(), nil
}

// ClearCache clears the template cache
func (tm *DefaultTemplateManager) ClearCache() {
	tm.cache = make(map[string]*Template)
}

// getEmbeddedTemplate returns embedded template content
func getEmbeddedTemplate(templateName string) string {
	templates := map[string]string{
		"handler.go": `package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
)

// {{.HandlerName}}Handler handles {{.ResourceName}} operations
type {{.HandlerName}}Handler struct{}

// New{{.HandlerName}}Handler creates a new {{.ResourceName}} handler
func New{{.HandlerName}}Handler() *{{.HandlerName}}Handler {
	return &{{.HandlerName}}Handler{}
}

// Routes returns the router for {{.ResourceName}} endpoints
func (h *{{.HandlerName}}Handler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.List{{.HandlerName}}s)
	r.Post("/", h.Create{{.HandlerName}})
	r.Get("/{id}", h.Get{{.HandlerName}})
	r.Put("/{id}", h.Update{{.HandlerName}})
	r.Delete("/{id}", h.Delete{{.HandlerName}})
	return r
}

// List{{.HandlerName}}s handles GET /{{.ResourcePath}}
func (h *{{.HandlerName}}Handler) List{{.HandlerName}}s(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"{{.ResourceNamePlural}}": []interface{}{},
		"total": 0,
	})
}

// Create{{.HandlerName}} handles POST /{{.ResourcePath}}
func (h *{{.HandlerName}}Handler) Create{{.HandlerName}}(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "{{.HandlerName}} created successfully",
	})
}

// Get{{.HandlerName}} handles GET /{{.ResourcePath}}/{id}
func (h *{{.HandlerName}}Handler) Get{{.HandlerName}}(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id": id,
		"message": "{{.HandlerName}} found",
	})
}

// Update{{.HandlerName}} handles PUT /{{.ResourcePath}}/{id}
func (h *{{.HandlerName}}Handler) Update{{.HandlerName}}(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id": id,
		"message": "{{.HandlerName}} updated successfully",
	})
}

// Delete{{.HandlerName}} handles DELETE /{{.ResourcePath}}/{id}
func (h *{{.HandlerName}}Handler) Delete{{.HandlerName}}(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id": id,
		"message": "{{.HandlerName}} deleted successfully",
	})
}`,
		"model.go": `package models

import "time"

// {{.ModelName}} represents a {{.ResourceName}} entity
type {{.ModelName}} struct {
	ID        string    ` + "`json:\"id\"`" + `
	CreatedAt time.Time ` + "`json:\"created_at\"`" + `
	UpdatedAt time.Time ` + "`json:\"updated_at\"`" + `
}

// New{{.ModelName}} creates a new {{.ResourceName}}
func New{{.ModelName}}() *{{.ModelName}} {
	now := time.Now()
	return &{{.ModelName}}{
		ID:        generateID(),
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// Validate validates the {{.ResourceName}} data
func (m *{{.ModelName}}) Validate() error {
	return nil
}

// Update updates the {{.ResourceName}} timestamp
func (m *{{.ModelName}}) Update() {
	m.UpdatedAt = time.Now()
}

// generateID generates a unique ID
func generateID() string {
	return "generated-id"
}`,
	}

	return templates[templateName]
}
