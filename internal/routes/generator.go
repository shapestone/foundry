package routes

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Update represents a file modification
type Update struct {
	Path     string
	Original []byte
	Modified []byte
	Changes  []string
}

// Generator handles route file updates for auto-wiring
type Generator interface {
	UpdateRoutes(handlerName string, moduleName string) (*Update, error)
	ValidateGoFile(path string) error
}

// FileGenerator implements Generator for file system operations
type FileGenerator struct{}

// NewFileGenerator creates a new file generator
func NewFileGenerator() *FileGenerator {
	return &FileGenerator{}
}

// UpdateRoutes calculates the changes needed to add a handler to routes.go
func (g *FileGenerator) UpdateRoutes(handlerName string, moduleName string) (*Update, error) {
	routesPath := filepath.Join("internal", "routes", "routes.go")

	// Read current file
	original, err := os.ReadFile(routesPath)
	if err != nil {
		return nil, fmt.Errorf("reading routes.go: %w", err)
	}

	modified := string(original)
	changes := []string{}

	// Add import if needed
	importLine := fmt.Sprintf(`"%s/internal/handlers"`, moduleName)
	if !strings.Contains(modified, importLine) {
		// Find import block and add handlers import
		if strings.Contains(modified, "import (") {
			modified = strings.Replace(
				modified,
				"import (",
				fmt.Sprintf("import (\n\t%s", importLine),
				1,
			)
		} else {
			// Add import block if it doesn't exist
			packageLine := "package routes"
			modified = strings.Replace(
				modified,
				packageLine,
				fmt.Sprintf("%s\n\nimport (\n\t%s\n)", packageLine, importLine),
				1,
			)
		}
		changes = append(changes, fmt.Sprintf("Add import: %s", importLine))
	}

	// Create handler registration code
	handlerVar := strings.ToLower(handlerName) + "Handler"
	handlerType := strings.Title(handlerName) + "Handler"
	routePath := "/" + strings.ToLower(handlerName) + "s"

	handlerCode := fmt.Sprintf(
		"\n\t\t// %s routes\n\t\t%s := handlers.New%s()\n\t\tr.Mount(\"%s\", %s.Routes())",
		strings.Title(handlerName),
		handlerVar,
		handlerType,
		routePath,
		handlerVar,
	)

	// Find the API v1 route block and insert handler
	lines := strings.Split(modified, "\n")
	inserted := false

	for i, line := range lines {
		// Look for the API v1 route block
		if strings.Contains(line, `r.Route("/api/v1"`) {
			// Find the comment "// Handler routes will be auto-generated here"
			for j := i; j < len(lines); j++ {
				if strings.Contains(lines[j], "// Handler routes will be auto-generated here") {
					// Insert after the comment
					lines[j] = lines[j] + handlerCode
					changes = append(changes, fmt.Sprintf("Add %s handler registration", handlerName))
					inserted = true
					break
				}
			}
			break
		}
	}

	if !inserted {
		// Fallback: try to find the closing of the API v1 route block
		for i, line := range lines {
			if strings.Contains(line, `r.Route("/api/v1"`) {
				braceCount := 0
				for j := i; j < len(lines); j++ {
					for _, char := range lines[j] {
						if char == '{' {
							braceCount++
						} else if char == '}' {
							braceCount--
							if braceCount == 0 {
								// Insert before the closing brace
								lines[j] = handlerCode + "\n" + lines[j]
								changes = append(changes, fmt.Sprintf("Add %s handler registration", handlerName))
								inserted = true
								break
							}
						}
					}
					if inserted {
						break
					}
				}
				break
			}
		}
	}

	if !inserted {
		return nil, fmt.Errorf("could not find appropriate location to insert handler routes")
	}

	modified = strings.Join(lines, "\n")

	return &Update{
		Path:     routesPath,
		Original: original,
		Modified: []byte(modified),
		Changes:  changes,
	}, nil
}

// ValidateGoFile validates Go syntax using gofmt
func (g *FileGenerator) ValidateGoFile(path string) error {
	cmd := exec.Command("gofmt", "-e", path)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("invalid Go syntax: %s", output)
	}
	return nil
}

// ApplyUpdate applies a file update with rollback support
func ApplyUpdate(update *Update, validator Generator) error {
	// Create backup
	backupPath := update.Path + ".backup"
	if err := os.WriteFile(backupPath, update.Original, 0644); err != nil {
		return fmt.Errorf("creating backup: %w", err)
	}

	// Cleanup function
	cleanup := func(success bool) {
		if success {
			os.Remove(backupPath)
		} else {
			// Restore from backup
			os.WriteFile(update.Path, update.Original, 0644)
			os.Remove(backupPath)
		}
	}

	// Write new content
	if err := os.WriteFile(update.Path, update.Modified, 0644); err != nil {
		cleanup(false)
		return fmt.Errorf("writing file: %w", err)
	}

	// Validate syntax
	if err := validator.ValidateGoFile(update.Path); err != nil {
		cleanup(false)
		return err
	}

	cleanup(true)
	return nil
}

// UpdateHandlerRoutes is a convenience function for updating handler routes
func UpdateHandlerRoutes(handlerName, moduleName string) error {
	generator := NewFileGenerator()

	update, err := generator.UpdateRoutes(handlerName, moduleName)
	if err != nil {
		return fmt.Errorf("calculating route updates: %w", err)
	}

	return ApplyUpdate(update, generator)
}

// RemoveHandlerRoutes removes a handler from routes (for cleanup/undo operations)
func RemoveHandlerRoutes(handlerName string) error {
	routesPath := filepath.Join("internal", "routes", "routes.go")

	// Read current file
	content, err := os.ReadFile(routesPath)
	if err != nil {
		return fmt.Errorf("reading routes.go: %w", err)
	}

	modified := string(content)

	// Remove handler registration lines
	lines := strings.Split(modified, "\n")
	var filteredLines []string

	handlerVar := strings.ToLower(handlerName) + "Handler"

	for _, line := range lines {
		// Skip lines related to this handler
		if strings.Contains(line, handlerVar) ||
			strings.Contains(line, fmt.Sprintf("// %s routes", strings.Title(handlerName))) {
			continue
		}
		filteredLines = append(filteredLines, line)
	}

	modified = strings.Join(filteredLines, "\n")

	return os.WriteFile(routesPath, []byte(modified), 0644)
}
