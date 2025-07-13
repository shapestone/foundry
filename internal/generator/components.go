package generator

import (
	"path/filepath"
	"strings"
)

// ComponentGenerator handles component generation
type ComponentGenerator struct {
	integration   *TemplateIntegration
	fileGenerator *FileGenerator
}

// NewComponentGenerator creates a new component generator
func NewComponentGenerator() *ComponentGenerator {
	return &ComponentGenerator{
		integration:   NewTemplateIntegration(),
		fileGenerator: NewFileGenerator(),
	}
}

// NewBackwardCompatibilityAdapter creates a backward compatibility adapter
// This maintains the existing API while using the new system
func NewBackwardCompatibilityAdapter() *ComponentGenerator {
	return NewComponentGenerator()
}

// GenerateHandler creates a new HTTP handler
func (cg *ComponentGenerator) GenerateHandler(name, outputPath string) error {
	return cg.integration.GenerateComponentWithNewSystem("handler", name, outputPath)
}

// GenerateModel creates a new data model
func (cg *ComponentGenerator) GenerateModel(name, outputPath string) error {
	return cg.integration.GenerateComponentWithNewSystem("model", name, outputPath)
}

// GenerateMiddleware creates a new middleware
func (cg *ComponentGenerator) GenerateMiddleware(name, outputPath string) error {
	// Try new template system first
	err := cg.integration.GenerateComponentWithNewSystem("middleware", name, outputPath)
	if err != nil {
		// Fallback to simple embedded template for middleware
		return cg.generateMiddlewareFallback(name, outputPath)
	}
	return nil
}

// GenerateService creates a new service
func (cg *ComponentGenerator) GenerateService(name, outputPath string) error {
	return cg.integration.GenerateComponentWithNewSystem("service", name, outputPath)
}

// generateMiddlewareFallback creates middleware using a simple embedded template
func (cg *ComponentGenerator) generateMiddlewareFallback(name, outputPath string) error {
	template := getMiddlewareTemplateForType(name)

	data := map[string]interface{}{
		"Name": strings.Title(name),
	}

	filename := strings.ToLower(name) + ".go"
	targetPath := filepath.Join(outputPath, "internal", "middleware", filename)

	return cg.fileGenerator.Generate(targetPath, template, data)
}

// getMiddlewareTemplateForType returns the appropriate template for the middleware type
func getMiddlewareTemplateForType(middlewareType string) string {
	switch middlewareType {
	case "auth":
		return `package middleware

import (
	"context"
	"net/http"
	"strings"
)

// AuthMiddleware validates authentication tokens
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get the Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Authorization header required", http.StatusUnauthorized)
			return
		}

		// Check for Bearer token
		if !strings.HasPrefix(authHeader, "Bearer ") {
			http.Error(w, "Invalid authorization format", http.StatusUnauthorized)
			return
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token == "" {
			http.Error(w, "Token required", http.StatusUnauthorized)
			return
		}

		// TODO: Validate token (JWT, database lookup, etc.)
		if !validateToken(token) {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		// TODO: Add user info to context
		ctx := context.WithValue(r.Context(), "user_id", getUserIDFromToken(token))
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// validateToken validates the authentication token
// TODO: Implement proper token validation
func validateToken(token string) bool {
	// Placeholder implementation
	return token == "valid-token"
}

// getUserIDFromToken extracts user ID from token
// TODO: Implement proper user ID extraction
func getUserIDFromToken(token string) string {
	// Placeholder implementation
	return "user123"
}
`

	case "cors":
		return `package middleware

import (
	"net/http"
)

// CorsMiddleware adds CORS headers to responses
func CorsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*") // TODO: Configure for production
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Max-Age", "86400") // 24 hours

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
`

	case "logging":
		return `package middleware

import (
	"log"
	"net/http"
	"time"
)

// LoggingMiddleware logs HTTP requests
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Wrap the response writer to capture status code
		lrw := &loggingResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(lrw, r)

		duration := time.Since(start)
		log.Printf("[%s] %s %s %d %v",
			r.Method,
			r.RequestURI,
			r.RemoteAddr,
			lrw.statusCode,
			duration,
		)
	})
}

type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}
`

	default:
		return `package middleware

import (
	"net/http"
)

// {{.Name}}Middleware provides {{.Name}} functionality
func {{.Name}}Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO: Implement {{.Name}} middleware logic
		
		// Call the next handler
		next.ServeHTTP(w, r)
	})
}
`
	}
}
