package generators

import (
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/shapestone/foundry/internal/cli/templates"
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
	// Create model file
	modelPath := filepath.Join(options.OutputDir, fmt.Sprintf("%s.go", strings.ToLower(options.Name)))
	if err := g.createModelFile(modelPath, options.Name); err != nil {
		return fmt.Errorf("failed to create model file: %w", err)
	}

	// Show success message
	g.showSuccess(options)

	return nil
}

// createModelFile creates a model file
func (g *ModelGenerator) createModelFile(modelPath, name string) error {
	template := templates.GetModelTemplate(name)
	return writeFile(modelPath, template)
}

// showSuccess displays success message with instructions
func (g *ModelGenerator) showSuccess(options ModelOptions) {
	modelPath := filepath.Join(options.OutputDir, fmt.Sprintf("%s.go", strings.ToLower(options.Name)))
	fields := templates.GetModelFields(strings.ToLower(options.Name))
	usage := templates.GetModelUsage(options.Name)

	fieldsList := g.getFieldsList(fields)

	fmt.Fprintf(g.stdout, `✅ Model created successfully!

📁 Created:
  %s

📝 Model includes:
  - ID field (string)
  - Timestamps (created_at, updated_at)
  %s
  - Validate() method
  - Update() method
  - Constructor function

%s
`, modelPath, fieldsList, usage)
}

// getFieldsList returns a formatted list of included fields
func (g *ModelGenerator) getFieldsList(fields templates.ModelFields) string {
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
	return strings.Join(fieldsList, "\n  ")
}
