package routes

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// RegisterAPIRoutes sets up all API routes
func RegisterAPIRoutes() *chi.Mux {
	r := chi.NewRouter()

	// Middleware stack
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)

	// API versioning
	r.Route("/api/v1", func(r chi.Router) {
		// Health check endpoint
		r.Get("/health", HealthHandler)

		// Handler routes will be auto-generated here
		// Example:
		// userHandler := handlers.NewUserHandler()
		// r.Mount("/users", userHandler.Routes())
	})

	return r
}

// HealthHandler provides a simple health check endpoint
func HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok","message":"API is healthy"}`))
}

// RegisterStaticRoutes sets up static file serving routes
func RegisterStaticRoutes(r *chi.Mux) {
	// Serve static files from public directory
	fileServer := http.FileServer(http.Dir("./public/"))
	r.Handle("/static/*", http.StripPrefix("/static/", fileServer))
}

// RegisterWebRoutes sets up web page routes (if needed)
func RegisterWebRoutes(r *chi.Mux) {
	r.Route("/", func(r chi.Router) {
		r.Get("/", IndexHandler)
		r.Get("/about", AboutHandler)
	})
}

// IndexHandler serves the main page
func IndexHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`
<!DOCTYPE html>
<html>
<head>
    <title>Welcome</title>
</head>
<body>
    <h1>Welcome to Your Application</h1>
    <p>API is available at <a href="/api/v1/health">/api/v1/health</a></p>
</body>
</html>
	`))
}

// AboutHandler serves the about page
func AboutHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`
<!DOCTYPE html>
<html>
<head>
    <title>About</title>
</head>
<body>
    <h1>About This Application</h1>
    <p>Built with Foundry CLI tool</p>
    <p><a href="/">← Back to Home</a></p>
</body>
</html>
	`))
}
