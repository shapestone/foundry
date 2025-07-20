package templates

// Model Templates

const ModelTemplate = `package models

import (
	"fmt"
	"time"
)

// {{.Name}} represents a {{.name}} in the system
type {{.Name}} struct {
	ID        string    ` + "`json:\"id\" db:\"id\"`" + `
	{{.NameField}}{{.EmailField}}{{.TitleField}}{{.DescriptionField}}CreatedAt time.Time ` + "`json:\"created_at\" db:\"created_at\"`" + `
	UpdatedAt time.Time ` + "`json:\"updated_at\" db:\"updated_at\"`" + `
}

// New{{.Name}} creates a new {{.name}} instance
func New{{.Name}}() *{{.Name}} {
	now := time.Now()
	return &{{.Name}}{
		ID:        generateID(),
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// Validate validates the {{.name}} data
func ({{.nameInitial}} *{{.Name}}) Validate() error {
	if {{.nameInitial}}.ID == "" {
		return fmt.Errorf("ID is required")
	}

	{{.NameValidation}}{{.EmailValidation}}{{.TitleValidation}}{{.DescriptionValidation}}return nil
}

// Update updates the {{.name}} with new data
func ({{.nameInitial}} *{{.Name}}) Update(data map[string]interface{}) error {
	{{.NameUpdate}}{{.EmailUpdate}}{{.TitleUpdate}}{{.DescriptionUpdate}}{{.nameInitial}}.UpdatedAt = time.Now()
	return {{.nameInitial}}.Validate()
}

// ToMap converts the {{.name}} to a map
func ({{.nameInitial}} *{{.Name}}) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id":         {{.nameInitial}}.ID,
		{{.NameToMap}}{{.EmailToMap}}{{.TitleToMap}}{{.DescriptionToMap}}"created_at": {{.nameInitial}}.CreatedAt,
		"updated_at": {{.nameInitial}}.UpdatedAt,
	}
}

// FromMap populates the {{.name}} from a map
func ({{.nameInitial}} *{{.Name}}) FromMap(data map[string]interface{}) error {
	if id, ok := data["id"].(string); ok {
		{{.nameInitial}}.ID = id
	}

	{{.NameFromMap}}{{.EmailFromMap}}{{.TitleFromMap}}{{.DescriptionFromMap}}if createdAt, ok := data["created_at"].(time.Time); ok {
		{{.nameInitial}}.CreatedAt = createdAt
	}

	if updatedAt, ok := data["updated_at"].(time.Time); ok {
		{{.nameInitial}}.UpdatedAt = updatedAt
	}

	return {{.nameInitial}}.Validate()
}

// generateID generates a unique ID for the {{.name}}
// TODO: Implement your preferred ID generation strategy (UUID, ULID, etc.)
func generateID() string {
	// Placeholder implementation - replace with your preferred method
	return fmt.Sprintf("{{.name}}_%d", time.Now().UnixNano())
}

// {{.Name}}Repository interface defines data access methods
type {{.Name}}Repository interface {
	Create({{.nameInitial}} *{{.Name}}) error
	GetByID(id string) (*{{.Name}}, error)
	Update({{.nameInitial}} *{{.Name}}) error
	Delete(id string) error
	List(limit, offset int) ([]*{{.Name}}, error)
}

// {{.Name}}Service interface defines business logic methods
type {{.Name}}Service interface {
	Create{{.Name}}(data map[string]interface{}) (*{{.Name}}, error)
	Get{{.Name}}ByID(id string) (*{{.Name}}, error)
	Update{{.Name}}(id string, data map[string]interface{}) (*{{.Name}}, error)
	Delete{{.Name}}(id string) error
	List{{.Name}}s(limit, offset int) ([]*{{.Name}}, error)
}
`

const ModelUsage = `
💡 Next steps:
  - Implement generateID() based on your needs (UUID, ULID, etc.)
  - Add custom fields specific to your {{.Name}}
  - Add methods for your business logic
  - Implement the {{.Name}}Repository interface for data access
  - Implement the {{.Name}}Service interface for business logic
  - Use in your handlers:
    
    import "your-module/internal/models"
    
    {{.name}} := models.New{{.Name}}()
    if err := {{.name}}.Validate(); err != nil {
        // handle error
    }
`

// Field templates for models

const NameFieldTemplate = `Name      string    ` + "`json:\"name\" db:\"name\"`" + `
	`

const EmailFieldTemplate = `Email     string    ` + "`json:\"email\" db:\"email\"`" + `
	`

const TitleFieldTemplate = `Title     string    ` + "`json:\"title\" db:\"title\"`" + `
	`

const DescriptionFieldTemplate = `Description string  ` + "`json:\"description\" db:\"description\"`" + `
	`

// Validation templates

const NameValidationTemplate = `if {{.nameInitial}}.Name == "" {
		return fmt.Errorf("name is required")
	}

	`

const EmailValidationTemplate = `if {{.nameInitial}}.Email == "" {
		return fmt.Errorf("email is required")
	}
	// Add email format validation here

	`

const TitleValidationTemplate = `if {{.nameInitial}}.Title == "" {
		return fmt.Errorf("title is required")
	}

	`

const DescriptionValidationTemplate = `// Description is optional, no validation needed

	`

// Update templates

const NameUpdateTemplate = `if name, ok := data["name"].(string); ok {
		{{.nameInitial}}.Name = name
	}

	`

const EmailUpdateTemplate = `if email, ok := data["email"].(string); ok {
		{{.nameInitial}}.Email = email
	}

	`

const TitleUpdateTemplate = `if title, ok := data["title"].(string); ok {
		{{.nameInitial}}.Title = title
	}

	`

const DescriptionUpdateTemplate = `if description, ok := data["description"].(string); ok {
		{{.nameInitial}}.Description = description
	}

	`

// ToMap templates

const NameToMapTemplate = `"name":       {{.nameInitial}}.Name,
		`

const EmailToMapTemplate = `"email":      {{.nameInitial}}.Email,
		`

const TitleToMapTemplate = `"title":      {{.nameInitial}}.Title,
		`

const DescriptionToMapTemplate = `"description": {{.nameInitial}}.Description,
		`

// FromMap templates

const NameFromMapTemplate = `if name, ok := data["name"].(string); ok {
		{{.nameInitial}}.Name = name
	}

	`

const EmailFromMapTemplate = `if email, ok := data["email"].(string); ok {
		{{.nameInitial}}.Email = email
	}

	`

const TitleFromMapTemplate = `if title, ok := data["title"].(string); ok {
		{{.nameInitial}}.Title = title
	}

	`

const DescriptionFromMapTemplate = `if description, ok := data["description"].(string); ok {
		{{.nameInitial}}.Description = description
	}

	`
