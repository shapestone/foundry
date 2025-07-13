package integration_test

import (
	"os"
	"path/filepath"
	"testing"
)

func TestProjectStructure(t *testing.T) {
	t.Run("NoUnwantedHandlerFiles", func(t *testing.T) {
		// Check that internal/handlers doesn't contain any Go files
		handlersDir := filepath.Join("internal", "handlers")

		entries, err := os.ReadDir(handlersDir)
		if os.IsNotExist(err) {
			// Directory doesn't exist, which is fine
			return
		}

		if err != nil {
			t.Fatalf("Failed to read handlers directory: %v", err)
		}

		for _, entry := range entries {
			if !entry.IsDir() && filepath.Ext(entry.Name()) == ".go" {
				t.Errorf("Unexpected Go file in handlers directory: %s", entry.Name())
				t.Log("The handlers directory should be empty until 'foundry add handler' is used")
			}
		}
	})
}
