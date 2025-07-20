package generators

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/shapestone/foundry/internal/layout"
	"github.com/shapestone/foundry/internal/routes"
)

// HandlerGenerator handles handler file generation
type HandlerGenerator struct {
	stdout io.Writer
	stderr io.Writer
}

// NewHandlerGenerator creates a new handler generator
func NewHandlerGenerator(stdout, stderr io.Writer) *HandlerGenerator {
	return &HandlerGenerator{
		stdout: stdout,
		stderr: stderr,
	}
}

// HandlerOptions holds options for handler generation
type HandlerOptions struct {
	Name      string
	AutoWire  bool
	OutputDir string
}

// Generate creates handler files based on options
func (g *HandlerGenerator) Generate(options HandlerOptions) error {
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
		return g.generateLegacyHandler(options)
	}

	// Generate component using layout system
	fmt.Fprintf(g.stdout, "🔨 Generating handler using '%s' layout...\n", layoutName)

	ctx := context.Background()
	err = manager.GenerateComponent(ctx, layoutName, "handler", options.Name, ".")
	if err != nil {
		return fmt.Errorf("failed to generate handler using layout system: %w", err)
	}

	// Handle auto-wiring
	if options.AutoWire {
		fmt.Fprintln(g.stdout, "\n🔄 Auto-wiring handler...")
		if err := g.wireHandler(options.Name); err != nil {
			fmt.Fprintf(g.stderr, "❌ Error auto-wiring handler: %v\n", err)
			fmt.Fprintln(g.stdout, "💡 Your handler was created but you'll need to manually wire it up")
			fmt.Fprintf(g.stdout, "   You can try: foundry wire handler %s\n", options.Name)
			g.showSuccess(options, false)
		} else {
			g.showSuccess(options, true)
		}
	} else {
		g.showSuccess(options, false)
	}

	return nil
}

// generateLegacyHandler falls back to legacy generation when layout system is unavailable
func (g *HandlerGenerator) generateLegacyHandler(options HandlerOptions) error {
	fmt.Fprintln(g.stdout, "🔧 Using legacy handler generation...")

	// Create handler file using legacy template
	handlerPath := filepath.Join(options.OutputDir, fmt.Sprintf("%s.go", strings.ToLower(options.Name)))
	if err := g.createLegacyHandlerFile(handlerPath, options.Name); err != nil {
		return fmt.Errorf("failed to create handler file: %w", err)
	}

	// Handle auto-wiring
	if options.AutoWire {
		fmt.Fprintln(g.stdout, "\n🔄 Auto-wiring handler...")
		if err := g.wireHandler(options.Name); err != nil {
			fmt.Fprintf(g.stderr, "❌ Error auto-wiring handler: %v\n", err)
			fmt.Fprintln(g.stdout, "💡 Your handler was created but you'll need to manually wire it up")
			fmt.Fprintf(g.stdout, "   You can try: foundry wire handler %s\n", options.Name)
			g.showSuccess(options, false)
		} else {
			g.showSuccess(options, true)
		}
	} else {
		g.showSuccess(options, false)
	}

	return nil
}

// createLegacyHandlerFile creates a handler file using legacy templates
func (g *HandlerGenerator) createLegacyHandlerFile(handlerPath, name string) error {
	// Get legacy template (you can move this to a separate legacy templates file)
	template := getLegacyHandlerTemplate(name)
	return writeFile(handlerPath, template)
}

// getLayoutManager gets the layout manager instance
func (g *HandlerGenerator) getLayoutManager() (*layout.Manager, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	configPath := filepath.Join(homeDir, ".foundry", "layouts.yaml")
	return layout.NewManager(configPath)
}

// detectProjectLayout detects the current project layout
func detectProjectLayout() (string, error) {
	// Check for foundry.yaml file
	if data, err := os.ReadFile("foundry.yaml"); err == nil {
		// Parse the foundry.yaml to get layout name
		// For now, just return standard, but you can implement YAML parsing
		_ = data
		return "standard", nil
	}

	// Check for .foundry.yaml file
	if data, err := os.ReadFile(".foundry.yaml"); err == nil {
		// Parse the .foundry.yaml to get layout name
		_ = data
		return "standard", nil
	}

	// Default to standard layout
	return "standard", nil
}

// wireHandler attempts to auto-wire handler into routes
func (g *HandlerGenerator) wireHandler(name string) error {
	// Get current module name
	moduleName := getCurrentModule()
	if moduleName == "" {
		return fmt.Errorf("could not determine module name")
	}

	// Create route generator
	generator := routes.NewFileGenerator()

	// Calculate the required changes
	update, err := generator.UpdateRoutes(strings.ToLower(name), moduleName)
	if err != nil {
		return fmt.Errorf("calculating route updates: %w", err)
	}

	// Show what changes will be made
	fmt.Fprintf(g.stdout, "📝 Updating routes file: %s\n", update.Path)
	for _, change := range update.Changes {
		fmt.Fprintf(g.stdout, "  - %s\n", change)
	}

	// Apply the changes
	if err := routes.ApplyUpdate(update, generator); err != nil {
		return fmt.Errorf("applying route updates: %w", err)
	}

	fmt.Fprintf(g.stdout, "✅ Routes updated successfully!\n")
	return nil
}

// showSuccess displays success message with instructions
func (g *HandlerGenerator) showSuccess(options HandlerOptions, autoWired bool) {
	handlerPath := filepath.Join(options.OutputDir, fmt.Sprintf("%s.go", strings.ToLower(options.Name)))
	resourcePath := strings.ToLower(options.Name) + "s" // simple pluralization

	wireStatus := ""
	if autoWired {
		wireStatus = `📝 Routes updated:
  internal/routes/routes.go

`
	} else {
		wireStatus = `📌 Manual wiring required:
  Run: foundry wire handler ` + options.Name + `
  Or manually update internal/routes/routes.go

`
	}

	fmt.Fprintf(g.stdout, `✅ Handler created successfully!

📁 Created:
  %s

%s🚀 Available endpoints:
  GET    /api/v1/%s       - List all %s
  POST   /api/v1/%s       - Create a new %s
  GET    /api/v1/%s/{id}  - Get %s by ID
  PUT    /api/v1/%s/{id}  - Update %s by ID
  DELETE /api/v1/%s/{id}  - Delete %s by ID

💡 Next steps:
  - Import the handlers package in your routes file
  - Implement your business logic in %sHandler methods
  - Add validation and error handling
  - Connect to your database or service layer
  - Add tests for the handler
  
  Example usage:
    %sHandler := handlers.New%sHandler()
    %sHandler.RegisterRoutes(apiRouter)
`, handlerPath,
		wireStatus,
		resourcePath, resourcePath,
		resourcePath, strings.ToLower(options.Name),
		resourcePath, strings.ToLower(options.Name),
		resourcePath, strings.ToLower(options.Name),
		resourcePath, strings.ToLower(options.Name),
		options.Name,
		strings.ToLower(options.Name), options.Name,
		strings.ToLower(options.Name))
}

// getLegacyHandlerTemplate returns the legacy handler template
func getLegacyHandlerTemplate(name string) string {
	// This is your existing handler template from templates package
	return `package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

// ` + capitalize(name) + `Handler handles ` + capitalize(name) + ` related requests
type ` + capitalize(name) + `Handler struct {
	// Add dependencies here (database, services, etc.)
}

// New` + capitalize(name) + `Handler creates a new ` + name + ` handler
func New` + capitalize(name) + `Handler() *` + capitalize(name) + `Handler {
	return &` + capitalize(name) + `Handler{
		// Initialize dependencies
	}
}

// List` + capitalize(name) + `s handles GET /` + strings.ToLower(name) + `s
func (h *` + capitalize(name) + `Handler) List` + capitalize(name) + `s(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement list ` + strings.ToLower(name) + `s logic
	` + strings.ToLower(name) + `s := []map[string]interface{}{
		{
			"id":   1,
			"name": "Example ` + capitalize(name) + `",
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"data": ` + strings.ToLower(name) + `s,
	})
}

// Get` + capitalize(name) + ` handles GET /` + strings.ToLower(name) + `s/{id}
func (h *` + capitalize(name) + `Handler) Get` + capitalize(name) + `(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	// TODO: Implement get ` + strings.ToLower(name) + ` logic
	` + strings.ToLower(name) + ` := map[string]interface{}{
		"id":   id,
		"name": "Example ` + capitalize(name) + `",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"data": ` + strings.ToLower(name) + `,
	})
}

// Create` + capitalize(name) + ` handles POST /` + strings.ToLower(name) + `s
func (h *` + capitalize(name) + `Handler) Create` + capitalize(name) + `(w http.ResponseWriter, r *http.Request) {
	var req map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// TODO: Implement create ` + strings.ToLower(name) + ` logic
	` + strings.ToLower(name) + ` := map[string]interface{}{
		"id":   1,
		"name": req["name"],
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"data": ` + strings.ToLower(name) + `,
	})
}

// Update` + capitalize(name) + ` handles PUT /` + strings.ToLower(name) + `s/{id}
func (h *` + capitalize(name) + `Handler) Update` + capitalize(name) + `(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	var req map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// TODO: Implement update ` + strings.ToLower(name) + ` logic
	` + strings.ToLower(name) + ` := map[string]interface{}{
		"id":   id,
		"name": req["name"],
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"data": ` + strings.ToLower(name) + `,
	})
}

// Delete` + capitalize(name) + ` handles DELETE /` + strings.ToLower(name) + `s/{id}
func (h *` + capitalize(name) + `Handler) Delete` + capitalize(name) + `(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	// TODO: Implement delete ` + strings.ToLower(name) + ` logic
	w.WriteHeader(http.StatusNoContent)
}

// RegisterRoutes registers all ` + strings.ToLower(name) + ` routes
func (h *` + capitalize(name) + `Handler) RegisterRoutes(r *mux.Router) {
	` + strings.ToLower(name) + `Router := r.PathPrefix("/` + strings.ToLower(name) + `s").Subrouter()
	
	` + strings.ToLower(name) + `Router.HandleFunc("", h.List` + capitalize(name) + `s).Methods("GET")
	` + strings.ToLower(name) + `Router.HandleFunc("", h.Create` + capitalize(name) + `).Methods("POST")
	` + strings.ToLower(name) + `Router.HandleFunc("/{id}", h.Get` + capitalize(name) + `).Methods("GET")
	` + strings.ToLower(name) + `Router.HandleFunc("/{id}", h.Update` + capitalize(name) + `).Methods("PUT")
	` + strings.ToLower(name) + `Router.HandleFunc("/{id}", h.Delete` + capitalize(name) + `).Methods("DELETE")
}
`
}

// capitalize helper function
func capitalize(s string) string {
	if len(s) == 0 {
		return s
	}
	return strings.ToUpper(s[:1]) + strings.ToLower(s[1:])
}
