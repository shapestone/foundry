// test/integration/core_commands_test.go
package integration

import (
	"testing"
)

// TestFoundryInit tests the init command functionality
func TestFoundryInit(t *testing.T) {
	tests := []struct {
		name       string
		args       []string
		wantErr    bool
		checkFiles []string
	}{
		{
			name:    "init with project name",
			args:    []string{"init", "test-project", "--force"},
			wantErr: false,
			checkFiles: []string{
				"go.mod",
				"main.go",
				"Makefile",
			},
		},
		{
			name:    "init without project name uses directory",
			args:    []string{"init", "--force"},
			wantErr: false,
			checkFiles: []string{
				"go.mod",
				"main.go",
			},
		},
		{
			name:    "init with custom module",
			args:    []string{"init", "test-project", "--module", "github.com/user/test", "--force"},
			wantErr: false,
			checkFiles: []string{
				"go.mod",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewTestHelper(t)

			output, err := h.RunFoundry(tt.args...)

			if tt.wantErr {
				h.AssertError(err, "")
			} else {
				h.AssertNoError(err)
				h.AssertOutputContains(output, "initialized successfully")

				// Check expected files exist
				for _, file := range tt.checkFiles {
					h.AssertFileExists(file)
				}

				// Verify go.mod content if init succeeded
				if !tt.wantErr {
					h.AssertFileContains("go.mod", "module")
				}
			}
		})
	}
}

// TestFoundryNew tests the new command functionality
func TestFoundryNew(t *testing.T) {
	tests := []struct {
		name       string
		args       []string
		wantErr    bool
		projectDir string
		checkFiles []string
	}{
		{
			name:       "new with project name",
			args:       []string{"new", "myproject"},
			wantErr:    false,
			projectDir: "myproject",
			checkFiles: []string{
				"myproject/go.mod",
				"myproject/main.go",
				"myproject/Makefile",
			},
		},
		{
			name:       "new with custom layout",
			args:       []string{"new", "myapi", "--layout", "microservice"},
			wantErr:    false,
			projectDir: "myapi",
			checkFiles: []string{
				"myapi/go.mod",
				"myapi/main.go",
			},
		},
		{
			name:    "new without project name",
			args:    []string{"new"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewTestHelper(t)

			output, err := h.RunFoundry(tt.args...)

			if tt.wantErr {
				h.AssertError(err, "")
			} else {
				h.AssertNoError(err)
				h.AssertOutputContains(output, "created successfully")

				// Check expected files exist
				for _, file := range tt.checkFiles {
					h.AssertFileExists(file)
				}
			}
		})
	}
}

// TestFoundryAddHandler tests adding handlers
func TestFoundryAddHandler(t *testing.T) {
	h := NewTestHelper(t)

	// First create a project
	output, err := h.RunFoundry("init", "test-project", "--force")
	h.AssertNoError(err)
	h.AssertOutputContains(output, "initialized successfully")

	// Test adding a handler
	output, err = h.RunFoundry("add", "handler", "user")
	h.AssertNoError(err)
	h.AssertOutputContains(output, "Handler created successfully")

	// Check that handler file was created
	h.AssertFileExists("internal/handlers/user.go")

	// Check handler content
	h.AssertFileContains("internal/handlers/user.go", "type UserHandler struct")
	h.AssertFileContains("internal/handlers/user.go", "func (h *UserHandler) ListUsers")
	h.AssertFileContains("internal/handlers/user.go", "func (h *UserHandler) CreateUser")
}

// TestFoundryAddMiddleware tests adding middleware
func TestFoundryAddMiddleware(t *testing.T) {
	h := NewTestHelper(t)

	// First create a project
	output, err := h.RunFoundry("init", "test-project", "--force")
	h.AssertNoError(err)

	// Test adding middleware
	output, err = h.RunFoundry("add", "middleware", "auth")
	h.AssertNoError(err)
	h.AssertOutputContains(output, "Middleware created successfully")

	// Check that middleware file was created
	h.AssertFileExists("internal/middleware/auth.go")

	// Check middleware content
	h.AssertFileContains("internal/middleware/auth.go", "func AuthMiddleware")
}

// TestFoundryAddModel tests adding models
func TestFoundryAddModel(t *testing.T) {
	h := NewTestHelper(t)

	// First create a project
	output, err := h.RunFoundry("init", "test-project", "--force")
	h.AssertNoError(err)

	// Test adding a model
	output, err = h.RunFoundry("add", "model", "user")
	h.AssertNoError(err)
	h.AssertOutputContains(output, "Model created successfully")

	// Check that model file was created
	h.AssertFileExists("internal/models/user.go")

	// Check model content
	h.AssertFileContains("internal/models/user.go", "type User struct")
	h.AssertFileContains("internal/models/user.go", "func NewUser()")
}

// TestFoundryAddDatabase tests adding database support
func TestFoundryAddDatabase(t *testing.T) {
	h := NewTestHelper(t)

	// First create a project
	output, err := h.RunFoundry("init", "test-project", "--force")
	h.AssertNoError(err)

	// Test adding postgres database
	output, err = h.RunFoundry("add", "db", "postgres")
	h.AssertNoError(err)
	h.AssertOutputContains(output, "Database support added successfully")

	// Check that database files were created
	h.AssertFileExists("internal/database/database.go")
	h.AssertFileExists("internal/database/config.go")
	h.AssertFileExists(".env.example")

	// Check database content
	h.AssertFileContains("internal/database/database.go", "NewConnection")
	h.AssertFileContains(".env.example", "DB_HOST")
}
