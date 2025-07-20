package templates

import (
	"strings"
)

// GetModelTemplate returns the Go template for a model
func GetModelTemplate(name string) string {
	return strings.ReplaceAll(ModelTemplate, "{{.Name}}", name)
}

// GetModelUsage returns usage instructions for models
func GetModelUsage(name string) string {
	return strings.ReplaceAll(ModelUsage, "{{.Name}}", name)
}

// GetModelFields returns smart default fields based on model name
func GetModelFields(name string) ModelFields {
	return ModelFields{
		IncludeNameField:        name == "user" || name == "customer" || name == "person",
		IncludeEmailField:       name == "user" || name == "customer" || name == "person",
		IncludeTitleField:       name == "post" || name == "article" || name == "product",
		IncludeDescriptionField: name == "post" || name == "article" || name == "product" || name == "project",
	}
}

// ModelFields represents which fields to include in a model
type ModelFields struct {
	IncludeNameField        bool
	IncludeEmailField       bool
	IncludeTitleField       bool
	IncludeDescriptionField bool
}
