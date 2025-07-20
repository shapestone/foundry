package templates

// Handler Templates

const HandlerTemplate = `package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

// {{.Name}}Handler handles {{.Name}} related requests
type {{.Name}}Handler struct {
	// Add dependencies here (database, services, etc.)
}

// New{{.Name}}Handler creates a new {{.Name}} handler
func New{{.Name}}Handler() *{{.Name}}Handler {
	return &{{.Name}}Handler{
		// Initialize dependencies
	}
}

// List{{.Name}}s handles GET /{{.name}}s
func (h *{{.Name}}Handler) List{{.Name}}s(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement list {{.name}}s logic
	{{.name}}s := []map[string]interface{}{
		{
			"id":   1,
			"name": "Example {{.Name}}",
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"data": {{.name}}s,
	})
}

// Get{{.Name}} handles GET /{{.name}}s/{id}
func (h *{{.Name}}Handler) Get{{.Name}}(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	// TODO: Implement get {{.name}} logic
	{{.name}} := map[string]interface{}{
		"id":   id,
		"name": "Example {{.Name}}",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"data": {{.name}},
	})
}

// Create{{.Name}} handles POST /{{.name}}s
func (h *{{.Name}}Handler) Create{{.Name}}(w http.ResponseWriter, r *http.Request) {
	var req map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// TODO: Implement create {{.name}} logic
	// Validate request data
	// Save to database
	// Return created {{.name}}

	{{.name}} := map[string]interface{}{
		"id":   1,
		"name": req["name"],
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"data": {{.name}},
	})
}

// Update{{.Name}} handles PUT /{{.name}}s/{id}
func (h *{{.Name}}Handler) Update{{.Name}}(w http.ResponseWriter, r *http.Request) {
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

	// TODO: Implement update {{.name}} logic
	// Validate request data
	// Update in database
	// Return updated {{.name}}

	{{.name}} := map[string]interface{}{
		"id":   id,
		"name": req["name"],
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"data": {{.name}},
	})
}

// Delete{{.Name}} handles DELETE /{{.name}}s/{id}
func (h *{{.Name}}Handler) Delete{{.Name}}(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	// TODO: Implement delete {{.name}} logic
	// Remove from database

	w.WriteHeader(http.StatusNoContent)
}

// RegisterRoutes registers all {{.name}} routes
func (h *{{.Name}}Handler) RegisterRoutes(r *mux.Router) {
	{{.name}}Router := r.PathPrefix("/{{.name}}s").Subrouter()
	
	{{.name}}Router.HandleFunc("", h.List{{.Name}}s).Methods("GET")
	{{.name}}Router.HandleFunc("", h.Create{{.Name}}).Methods("POST")
	{{.name}}Router.HandleFunc("/{id}", h.Get{{.Name}}).Methods("GET")
	{{.name}}Router.HandleFunc("/{id}", h.Update{{.Name}}).Methods("PUT")
	{{.name}}Router.HandleFunc("/{id}", h.Delete{{.Name}}).Methods("DELETE")
}
`

const HandlerUsage = `
💡 Next steps:
  - Import the handlers package in your routes file:
    import "your-module/internal/handlers"
  - Implement your business logic in {{.Name}}Handler methods
  - Add validation and error handling
  - Connect to your database or service layer
  - Add tests for the handler
  
  Example usage:
    {{.name}}Handler := handlers.New{{.Name}}Handler()
    {{.name}}Handler.RegisterRoutes(apiRouter)
`
