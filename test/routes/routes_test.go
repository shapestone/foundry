package routes_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/shapestone/foundry/internal/routes"
)

// Test data - sample routes.go file content
const sampleRoutesFile = `package routes

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func RegisterAPIRoutes(r *chi.Mux) {
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
}`

const sampleRoutesWithImport = `package routes

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"myapp/internal/handlers"
)

func RegisterAPIRoutes(r *chi.Mux) {
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
}`

const invalidGoFile = `package routes

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

func RegisterAPIRoutes(r *chi.Mux) {
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	// Missing closing brace
`

func TestNewFileUpdater(t *testing.T) {
	updater := routes.NewFileUpdater()
	if updater == nil {
		t.Fatal("NewFileUpdater() returned nil")
	}

	// Test that it implements the Updater interface
	var _ routes.Updater = updater
}

func TestFileUpdater_UpdateRoutes(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()

	// Create internal/routes directory structure
	routesDir := filepath.Join(tempDir, "internal", "routes")
	if err := os.MkdirAll(routesDir, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	routesPath := filepath.Join(routesDir, "routes.go")

	tests := []struct {
		name            string
		handlerName     string
		moduleName      string
		initialContent  string
		expectedChanges int
		expectError     bool
		setupFunc       func() error
		validateFunc    func(*routes.Update) error
	}{
		{
			name:            "AddHandlerToCleanFile",
			handlerName:     "user",
			moduleName:      "myapp",
			initialContent:  sampleRoutesFile,
			expectedChanges: 2, // import + handler registration
			expectError:     false,
			setupFunc: func() error {
				return os.WriteFile(routesPath, []byte(sampleRoutesFile), 0644)
			},
			validateFunc: func(update *routes.Update) error {
				modified := string(update.Modified)
				if !strings.Contains(modified, `"myapp/internal/handlers"`) {
					return fmt.Errorf("missing import statement")
				}
				if !strings.Contains(modified, "userHandler := handlers.NewUserHandler()") {
					return fmt.Errorf("missing handler initialization")
				}
				if !strings.Contains(modified, `r.Mount("/users", userHandler.Routes())`) {
					return fmt.Errorf("missing route mounting")
				}
				return nil
			},
		},
		{
			name:            "AddHandlerWithExistingImport",
			handlerName:     "product",
			moduleName:      "myapp",
			initialContent:  sampleRoutesWithImport,
			expectedChanges: 1, // only handler registration, import exists
			expectError:     false,
			setupFunc: func() error {
				return os.WriteFile(routesPath, []byte(sampleRoutesWithImport), 0644)
			},
			validateFunc: func(update *routes.Update) error {
				modified := string(update.Modified)
				if !strings.Contains(modified, "productHandler := handlers.NewProductHandler()") {
					return fmt.Errorf("missing handler initialization")
				}
				if !strings.Contains(modified, `r.Mount("/products", productHandler.Routes())`) {
					return fmt.Errorf("missing route mounting")
				}
				// Should not add duplicate import
				importCount := strings.Count(modified, `"myapp/internal/handlers"`)
				if importCount != 1 {
					return fmt.Errorf("import added when it already existed")
				}
				return nil
			},
		},
		{
			name:            "FileNotFound",
			handlerName:     "user",
			moduleName:      "myapp",
			initialContent:  "",
			expectedChanges: 0,
			expectError:     true,
			setupFunc: func() error {
				// Don't create the file
				return nil
			},
			validateFunc: func(update *routes.Update) error {
				return nil // Should not reach here
			},
		},
		{
			name:            "EmptyHandlerName",
			handlerName:     "",
			moduleName:      "myapp",
			initialContent:  sampleRoutesFile,
			expectedChanges: 2, // Implementation still adds import and handler (even with empty name)
			expectError:     false,
			setupFunc: func() error {
				return os.WriteFile(routesPath, []byte(sampleRoutesFile), 0644)
			},
			validateFunc: func(update *routes.Update) error {
				// Should handle empty handler name gracefully
				return nil
			},
		},
		{
			name:            "EmptyModuleName",
			handlerName:     "user",
			moduleName:      "",
			initialContent:  sampleRoutesFile,
			expectedChanges: 2, // Implementation still adds import and handler (even with empty module)
			expectError:     false,
			setupFunc: func() error {
				return os.WriteFile(routesPath, []byte(sampleRoutesFile), 0644)
			},
			validateFunc: func(update *routes.Update) error {
				// Should handle empty module name gracefully
				return nil
			},
		},
		{
			name:            "ComplexHandlerName",
			handlerName:     "user-profile",
			moduleName:      "myapp",
			initialContent:  sampleRoutesFile,
			expectedChanges: 2,
			expectError:     false,
			setupFunc: func() error {
				return os.WriteFile(routesPath, []byte(sampleRoutesFile), 0644)
			},
			validateFunc: func(update *routes.Update) error {
				modified := string(update.Modified)
				if !strings.Contains(modified, "user-profileHandler") {
					return fmt.Errorf("handler name not preserved correctly")
				}
				return nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Change to temp directory for this test
			oldWd, _ := os.Getwd()
			defer os.Chdir(oldWd)
			os.Chdir(tempDir)

			// Setup
			if err := tt.setupFunc(); err != nil {
				t.Fatalf("Setup failed: %v", err)
			}

			// Test
			updater := routes.NewFileUpdater()
			update, err := updater.UpdateRoutes(tt.handlerName, tt.moduleName)

			// Verify error expectation
			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			} else if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			// Verify update structure
			if update == nil {
				t.Fatal("Update should not be nil")
			}

			if len(update.Changes) != tt.expectedChanges {
				t.Errorf("Expected %d changes, got %d: %v",
					tt.expectedChanges, len(update.Changes), update.Changes)
			}

			if update.Path == "" {
				t.Error("Update path should not be empty")
			}

			if len(update.Original) == 0 && tt.initialContent != "" {
				t.Error("Original content should not be empty")
			}

			if len(update.Modified) == 0 {
				t.Error("Modified content should not be empty")
			}

			// Run custom validation
			if err := tt.validateFunc(update); err != nil {
				t.Errorf("Validation failed: %v", err)
			}

			// Cleanup
			os.Remove(routesPath)
		})
	}
}

func TestFileUpdater_ValidateGoFile(t *testing.T) {
	tempDir := t.TempDir()

	tests := []struct {
		name          string
		fileContent   string
		expectError   bool
		errorContains string
	}{
		{
			name:        "ValidGoFile",
			fileContent: sampleRoutesFile,
			expectError: false,
		},
		{
			name:          "InvalidGoFile",
			fileContent:   invalidGoFile,
			expectError:   true,
			errorContains: "invalid Go syntax",
		},
		{
			name:          "EmptyFile",
			fileContent:   "",
			expectError:   true,
			errorContains: "invalid Go syntax",
		},
		{
			name:        "ValidMinimalFile",
			fileContent: "package main\n",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test file
			testFile := filepath.Join(tempDir, fmt.Sprintf("test_%s.go", tt.name))
			if err := os.WriteFile(testFile, []byte(tt.fileContent), 0644); err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}
			defer os.Remove(testFile)

			// Test validation
			updater := routes.NewFileUpdater()
			err := updater.ValidateGoFile(testFile)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if tt.errorContains != "" && !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("Expected error to contain %q, got: %v", tt.errorContains, err)
				}
			} else if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestFileUpdater_ValidateGoFile_FileNotFound(t *testing.T) {
	updater := routes.NewFileUpdater()
	err := updater.ValidateGoFile("/nonexistent/file.go")

	if err == nil {
		t.Error("Expected error for nonexistent file")
	}
}

func TestApplyUpdate(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.go")

	tests := []struct {
		name            string
		originalContent string
		modifiedContent string
		validContent    bool
		expectError     bool
		errorContains   string
	}{
		{
			name:            "SuccessfulUpdate",
			originalContent: sampleRoutesFile,
			modifiedContent: sampleRoutesWithImport,
			validContent:    true,
			expectError:     false,
		},
		{
			name:            "UpdateWithInvalidSyntax",
			originalContent: sampleRoutesFile,
			modifiedContent: invalidGoFile,
			validContent:    false,
			expectError:     true,
			errorContains:   "invalid Go syntax",
		},
		{
			name:            "UpdateEmptyToValid",
			originalContent: "",
			modifiedContent: "package main\n",
			validContent:    true,
			expectError:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create original file
			if err := os.WriteFile(testFile, []byte(tt.originalContent), 0644); err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}

			// Create update
			update := &routes.Update{
				Path:     testFile,
				Original: []byte(tt.originalContent),
				Modified: []byte(tt.modifiedContent),
				Changes:  []string{"test change"},
			}

			// Create validator
			validator := routes.NewFileUpdater()

			// Apply update
			err := routes.ApplyUpdate(update, validator)

			// Check error expectation
			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if tt.errorContains != "" && !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("Expected error to contain %q, got: %v", tt.errorContains, err)
				}

				// Verify rollback - file should be restored to original
				content, readErr := os.ReadFile(testFile)
				if readErr == nil && string(content) != tt.originalContent {
					t.Errorf("File was not rolled back to original content")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}

				// Verify file was updated
				content, readErr := os.ReadFile(testFile)
				if readErr != nil {
					t.Fatalf("Failed to read updated file: %v", readErr)
				}

				if string(content) != tt.modifiedContent {
					t.Errorf("File content not updated correctly")
				}

				// Verify backup was cleaned up
				backupPath := testFile + ".backup"
				if _, err := os.Stat(backupPath); !os.IsNotExist(err) {
					t.Errorf("Backup file was not cleaned up")
				}
			}

			// Cleanup
			os.Remove(testFile)
			os.Remove(testFile + ".backup")
		})
	}
}

func TestApplyUpdate_BackupCreationFailure(t *testing.T) {
	// Test backup creation failure by using an invalid backup path
	update := &routes.Update{
		Path:     "/nonexistent/directory/file.go",
		Original: []byte("original"),
		Modified: []byte("modified"),
		Changes:  []string{"test change"},
	}

	validator := routes.NewFileUpdater()
	err := routes.ApplyUpdate(update, validator)

	if err == nil {
		t.Error("Expected error for backup creation failure")
	}

	if !strings.Contains(err.Error(), "creating backup") {
		t.Errorf("Expected backup creation error, got: %v", err)
	}
}

func TestUpdateStruct(t *testing.T) {
	// Test the Update struct
	update := &routes.Update{
		Path:     "/test/path.go",
		Original: []byte("original content"),
		Modified: []byte("modified content"),
		Changes:  []string{"change1", "change2"},
	}

	if update.Path != "/test/path.go" {
		t.Errorf("Path not set correctly")
	}

	if string(update.Original) != "original content" {
		t.Errorf("Original content not set correctly")
	}

	if string(update.Modified) != "modified content" {
		t.Errorf("Modified content not set correctly")
	}

	if len(update.Changes) != 2 {
		t.Errorf("Changes not set correctly")
	}
}

func TestUpdaterInterface(t *testing.T) {
	// Test that FileUpdater implements Updater interface
	var updater routes.Updater = routes.NewFileUpdater()

	if updater == nil {
		t.Error("FileUpdater should implement Updater interface")
	}

	// Test interface methods exist
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.go")
	os.WriteFile(testFile, []byte("package main\n"), 0644)

	// Should not panic
	_, err := updater.UpdateRoutes("test", "module")
	if err == nil {
		t.Log("UpdateRoutes method accessible through interface")
	}

	err = updater.ValidateGoFile(testFile)
	if err == nil {
		t.Log("ValidateGoFile method accessible through interface")
	}
}

// Integration tests
func TestFileUpdater_Integration(t *testing.T) {
	tempDir := t.TempDir()
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	os.Chdir(tempDir)

	// Create project structure
	routesDir := filepath.Join("internal", "routes")
	if err := os.MkdirAll(routesDir, 0755); err != nil {
		t.Fatalf("Failed to create directory structure: %v", err)
	}

	routesPath := filepath.Join(routesDir, "routes.go")
	if err := os.WriteFile(routesPath, []byte(sampleRoutesFile), 0644); err != nil {
		t.Fatalf("Failed to create routes file: %v", err)
	}

	// Test complete workflow
	updater := routes.NewFileUpdater()

	// Step 1: Calculate update
	update, err := updater.UpdateRoutes("user", "myapp")
	if err != nil {
		t.Fatalf("UpdateRoutes failed: %v", err)
	}

	// Step 2: Apply update
	err = routes.ApplyUpdate(update, updater)
	if err != nil {
		t.Fatalf("ApplyUpdate failed: %v", err)
	}

	// Step 3: Verify final file
	finalContent, err := os.ReadFile(routesPath)
	if err != nil {
		t.Fatalf("Failed to read final file: %v", err)
	}

	finalStr := string(finalContent)
	if !strings.Contains(finalStr, `"myapp/internal/handlers"`) {
		t.Error("Import not added correctly")
	}

	if !strings.Contains(finalStr, "userHandler := handlers.NewUserHandler()") {
		t.Error("Handler not added correctly")
	}

	// Step 4: Validate syntax
	err = updater.ValidateGoFile(routesPath)
	if err != nil {
		t.Errorf("Final file has invalid syntax: %v", err)
	}
}

// Benchmark tests
func BenchmarkUpdateRoutes(b *testing.B) {
	tempDir := b.TempDir()
	routesDir := filepath.Join(tempDir, "internal", "routes")
	os.MkdirAll(routesDir, 0755)

	routesPath := filepath.Join(routesDir, "routes.go")
	os.WriteFile(routesPath, []byte(sampleRoutesFile), 0644)

	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	os.Chdir(tempDir)

	updater := routes.NewFileUpdater()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		updater.UpdateRoutes("user", "myapp")
	}
}

func BenchmarkValidateGoFile(b *testing.B) {
	tempDir := b.TempDir()
	testFile := filepath.Join(tempDir, "test.go")
	os.WriteFile(testFile, []byte(sampleRoutesFile), 0644)

	updater := routes.NewFileUpdater()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		updater.ValidateGoFile(testFile)
	}
}

func BenchmarkApplyUpdate(b *testing.B) {
	tempDir := b.TempDir()

	updater := routes.NewFileUpdater()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		testFile := filepath.Join(tempDir, fmt.Sprintf("test_%d.go", i))
		os.WriteFile(testFile, []byte(sampleRoutesFile), 0644)

		update := &routes.Update{
			Path:     testFile,
			Original: []byte(sampleRoutesFile),
			Modified: []byte(sampleRoutesWithImport),
			Changes:  []string{"test"},
		}
		b.StartTimer()

		routes.ApplyUpdate(update, updater)
	}
}
