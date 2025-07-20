package generators

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/shapestone/foundry/internal/layout"
)

// ModelGenerator handles model file generation
type ModelGenerator struct {
	stdout io.Writer
	stderr io.Writer
}

// NewModelGenerator creates a new model generator
func NewModelGenerator(stdout, stderr io.Writer) *ModelGenerator {
	return &ModelGenerator{
		stdout: stdout,
		stderr: stderr,
	}
}

// ModelOptions holds options for model generation
type ModelOptions struct {
	Name      string
	OutputDir string
}

// Generate creates model files based on options
func (g *ModelGenerator) Generate(options ModelOptions) error {
	// Detect current layout
	layoutName, err := detectProjectLayout()
	if err != nil {
		fmt.Fprintf(g.stderr, "Warning: could not detect project layout, using standard: %v\n", err)
		layoutName = "standard"
	}

	// Get layout manager
	manager, err := g.getLayoutManager()
	if err != nil {
		fmt.Fprintf(g.stderr, "Warning: layout manager unavailable, falling back to legacy generation: %v\n", err)
		return g.generateLegacyModel(options)
	}

	// Generate component using layout system
	fmt.Fprintf(g.stdout, "🔨 Generating model using '%s' layout...\n", layoutName)

	ctx := context.Background()
	err = manager.GenerateComponent(ctx, layoutName, "model", options.Name, ".")
	if err != nil {
		return fmt.Errorf("failed to generate model using layout system: %w", err)
	}

	// Show success message
	g.showSuccess(options)
	return nil
}

// generateLegacyModel falls back to legacy generation when layout system is unavailable
func (g *ModelGenerator) generateLegacyModel(options ModelOptions) error {
	fmt.Fprintln(g.stdout, "🔧 Using legacy model generation...")

	// Create model file using legacy template
	modelPath := filepath.Join(options.OutputDir, fmt.Sprintf("%s.go", strings.ToLower(options.Name)))
	if err := g.createLegacyModelFile(modelPath, options.Name); err != nil {
		return fmt.Errorf("failed to create model file: %w", err)
	}

	// Show success message
	g.showSuccess(options)
	return nil
}

// createLegacyModelFile creates a model file using legacy templates
func (g *ModelGenerator) createLegacyModelFile(modelPath, name string) error {
	template := getLegacyModelTemplate(name)
	return writeFile(modelPath, template)
}

// getLayoutManager gets the layout manager instance
func (g *ModelGenerator) getLayoutManager() (*layout.Manager, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	configPath := filepath.Join(homeDir, ".foundry", "layouts.yaml")
	return layout.NewManager(configPath)
}

// showSuccess displays success message with instructions
func (g *ModelGenerator) showSuccess(options ModelOptions) {
	modelPath := filepath.Join(options.OutputDir, fmt.Sprintf("%s.go", strings.ToLower(options.Name)))
	fields := getModelFields(strings.ToLower(options.Name))
	fieldsList := g.getFieldsList(fields)

	fmt.Fprintf(g.stdout, `✅ Model created successfully!

📁 Created:
  %s

📝 Model includes:
  - ID field (int64)
  - Timestamps (created_at, updated_at)
  %s
  - Create/Update request types
  - Filter and response types
  - Validation methods
  - Helper functions

💡 Next steps:
  - Add your specific fields to the model
  - Implement validation logic
  - Use in your handlers and services:
    
    import "%s/internal/models"
    
    %s := models.New%s()
    if err := %s.Validate(); err != nil {
        // handle error
    }

🔧 Generate related components:
  foundry add handler %s     # Create HTTP handler
  foundry add service %s     # Create business logic layer
  foundry add repository %s  # Create data access layer
`, modelPath, fieldsList, getCurrentModule(), strings.ToLower(options.Name),
		capitalize(options.Name), strings.ToLower(options.Name),
		strings.ToLower(options.Name), strings.ToLower(options.Name), strings.ToLower(options.Name))
}

// getFieldsList returns a formatted list of included fields
func (g *ModelGenerator) getFieldsList(fields ModelFields) string {
	var fieldsList []string
	if fields.IncludeNameField {
		fieldsList = append(fieldsList, "- Name field (string)")
	}
	if fields.IncludeEmailField {
		fieldsList = append(fieldsList, "- Email field (string)")
	}
	if fields.IncludeTitleField {
		fieldsList = append(fieldsList, "- Title field (string)")
	}
	if fields.IncludeDescriptionField {
		fieldsList = append(fieldsList, "- Description field (string)")
	}
	if len(fieldsList) == 0 {
		return "- Base model structure"
	}
	return strings.Join(fieldsList, "\n  ")
}

// ModelFields represents which fields to include in a model
type ModelFields struct {
	IncludeNameField        bool
	IncludeEmailField       bool
	IncludeTitleField       bool
	IncludeDescriptionField bool
}

// getModelFields returns smart default fields based on model name
func getModelFields(name string) ModelFields {
	return ModelFields{
		IncludeNameField:        name == "user" || name == "customer" || name == "person",
		IncludeEmailField:       name == "user" || name == "customer" || name == "person",
		IncludeTitleField:       name == "post" || name == "article" || name == "product",
		IncludeDescriptionField: name == "post" || name == "article" || name == "product" || name == "project",
	}
}

// getLegacyModelTemplate returns the legacy model template
func getLegacyModelTemplate(name string) string {
	nameField := ""
	emailField := ""
	titleField := ""
	descriptionField := ""

	fields := getModelFields(strings.ToLower(name))

	if fields.IncludeNameField {
		nameField = "\tName      string    `json:\"name\" db:\"name\"`\n"
	}
	if fields.IncludeEmailField {
		emailField = "\tEmail     string    `json:\"email\" db:\"email\"`\n"
	}
	if fields.IncludeTitleField {
		titleField = "\tTitle     string    `json:\"title\" db:\"title\"`\n"
	}
	if fields.IncludeDescriptionField {
		descriptionField = "\tDescription string  `json:\"description\" db:\"description\"`\n"
	}

	nameInitial := strings.ToLower(string(name[0]))

	return `package models

import (
	"fmt"
	"time"
)

// ` + capitalize(name) + ` represents a ` + strings.ToLower(name) + ` in the system
type ` + capitalize(name) + ` struct {
	ID        int64     ` + "`json:\"id\" db:\"id\"`" + `
` + nameField + emailField + titleField + descriptionField + `	CreatedAt time.Time ` + "`json:\"created_at\" db:\"created_at\"`" + `
	UpdatedAt time.Time ` + "`json:\"updated_at\" db:\"updated_at\"`" + `
}

// New` + capitalize(name) + ` creates a new ` + strings.ToLower(name) + ` instance
func New` + capitalize(name) + `() *` + capitalize(name) + ` {
	now := time.Now()
	return &` + capitalize(name) + `{
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// Validate validates the ` + strings.ToLower(name) + ` data
func (` + nameInitial + ` *` + capitalize(name) + `) Validate() error {
	if ` + nameInitial + `.ID < 0 {
		return fmt.Errorf("invalid ID")
	}

	return nil
}

// Update updates the ` + strings.ToLower(name) + ` with new data
func (` + nameInitial + ` *` + capitalize(name) + `) Update(data map[string]interface{}) error {
	` + nameInitial + `.UpdatedAt = time.Now()
	return ` + nameInitial + `.Validate()
}

// ToMap converts the ` + strings.ToLower(name) + ` to a map
func (` + nameInitial + ` *` + capitalize(name) + `) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id":         ` + nameInitial + `.ID,
		"created_at": ` + nameInitial + `.CreatedAt,
		"updated_at": ` + nameInitial + `.UpdatedAt,
	}
}

// FromMap populates the ` + strings.ToLower(name) + ` from a map
func (` + nameInitial + ` *` + capitalize(name) + `) FromMap(data map[string]interface{}) error {
	if id, ok := data["id"].(int64); ok {
		` + nameInitial + `.ID = id
	}

	if createdAt, ok := data["created_at"].(time.Time); ok {
		` + nameInitial + `.CreatedAt = createdAt
	}

	if updatedAt, ok := data["updated_at"].(time.Time); ok {
		` + nameInitial + `.UpdatedAt = updatedAt
	}

	return ` + nameInitial + `.Validate()
}

// ` + capitalize(name) + `Repository interface defines data access methods
type ` + capitalize(name) + `Repository interface {
	Create(` + nameInitial + ` *` + capitalize(name) + `) error
	GetByID(id int64) (*` + capitalize(name) + `, error)
	Update(` + nameInitial + ` *` + capitalize(name) + `) error
	Delete(id int64) error
	List(limit, offset int) ([]*` + capitalize(name) + `, error)
}

// ` + capitalize(name) + `Service interface defines business logic methods
type ` + capitalize(name) + `Service interface {
	Create` + capitalize(name) + `(data map[string]interface{}) (*` + capitalize(name) + `, error)
	Get` + capitalize(name) + `ByID(id int64) (*` + capitalize(name) + `, error)
	Update` + capitalize(name) + `(id int64, data map[string]interface{}) (*` + capitalize(name) + `, error)
	Delete` + capitalize(name) + `(id int64) error
	List` + capitalize(name) + `s(limit, offset int) ([]*` + capitalize(name) + `, error)
}
`
}
