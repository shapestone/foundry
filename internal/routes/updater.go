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

// Updater handles route file updates
type Updater interface {
	UpdateRoutes(handlerName string, moduleName string) (*Update, error)
	ValidateGoFile(path string) error
}

// FileUpdater implements Updater for file system operations
type FileUpdater struct{}

// NewFileUpdater creates a new file updater
func NewFileUpdater() *FileUpdater {
	return &FileUpdater{}
}

// UpdateRoutes calculates the changes needed to add a handler to routes.go
func (u *FileUpdater) UpdateRoutes(handlerName string, moduleName string) (*Update, error) {
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
		modified = strings.Replace(
			modified,
			"import (",
			fmt.Sprintf("import (\n\t%s", importLine),
			1,
		)
		changes = append(changes, fmt.Sprintf("Add import: %s", importLine))
	}

	// Add handler registration
	handlerCode := fmt.Sprintf(
		"\n\t// %s routes\n\t%sHandler := handlers.New%sHandler()\n\tr.Mount(\"/%ss\", %sHandler.Routes())",
		strings.Title(handlerName),
		handlerName,
		strings.Title(handlerName),
		handlerName,
		handlerName,
	)

	// Find the right place to insert the handler
	lines := strings.Split(modified, "\n")
	inserted := false
	for i := len(lines) - 1; i >= 0; i-- {
		if strings.Contains(lines[i], "func RegisterAPIRoutes") {
			// Find the closing brace of the function
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

	if !inserted {
		return nil, fmt.Errorf("could not find appropriate location to insert handler")
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
func (u *FileUpdater) ValidateGoFile(path string) error {
	cmd := exec.Command("gofmt", "-e", path)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("invalid Go syntax: %s", output)
	}
	return nil
}

// ApplyUpdate applies a file update with rollback support
func ApplyUpdate(update *Update, validator Updater) error {
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
