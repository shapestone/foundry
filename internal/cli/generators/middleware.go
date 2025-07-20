package generators

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/shapestone/foundry/internal/layout"
)

// MiddlewareGenerator handles middleware file generation
type MiddlewareGenerator struct {
	stdout io.Writer
	stderr io.Writer
}

// NewMiddlewareGenerator creates a new middleware generator
func NewMiddlewareGenerator(stdout, stderr io.Writer) *MiddlewareGenerator {
	return &MiddlewareGenerator{
		stdout: stdout,
		stderr: stderr,
	}
}

// MiddlewareOptions holds options for middleware generation
type MiddlewareOptions struct {
	Type      string
	AutoWire  bool
	OutputDir string
}

// Generate creates middleware files based on options
func (g *MiddlewareGenerator) Generate(options MiddlewareOptions) error {
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
		return g.generateLegacyMiddleware(options)
	}

	// Generate component using layout system
	fmt.Fprintf(g.stdout, "🔨 Generating middleware using '%s' layout...\n", layoutName)

	ctx := context.Background()
	err = manager.GenerateComponent(ctx, layoutName, "middleware", options.Type, ".")
	if err != nil {
		return fmt.Errorf("failed to generate middleware using layout system: %w", err)
	}

	// Handle auto-wiring
	if options.AutoWire {
		fmt.Fprintln(g.stdout, "\n🔄 Auto-wiring middleware...")
		if err := g.wireMiddleware(options.Type); err != nil {
			fmt.Fprintf(g.stderr, "❌ Error auto-wiring middleware: %v\n", err)
			fmt.Fprintln(g.stdout, "💡 Your middleware was created but you'll need to manually wire it up")
			fmt.Fprintf(g.stdout, "   You can try: foundry wire middleware %s\n", options.Type)
			g.showSuccess(options, getMiddlewareInfo(options.Type), false)
		} else {
			g.showSuccess(options, getMiddlewareInfo(options.Type), true)
		}
	} else {
		g.showSuccess(options, getMiddlewareInfo(options.Type), false)
	}

	return nil
}

// generateLegacyMiddleware falls back to legacy generation when layout system is unavailable
func (g *MiddlewareGenerator) generateLegacyMiddleware(options MiddlewareOptions) error {
	fmt.Fprintln(g.stdout, "🔧 Using legacy middleware generation...")

	// Create middleware file using legacy template
	middlewarePath := filepath.Join(options.OutputDir, fmt.Sprintf("%s.go", options.Type))
	if err := g.createLegacyMiddlewareFile(middlewarePath, options.Type); err != nil {
		return fmt.Errorf("failed to create middleware file: %w", err)
	}

	// Handle auto-wiring
	if options.AutoWire {
		fmt.Fprintln(g.stdout, "\n🔄 Auto-wiring middleware...")
		if err := g.wireMiddleware(options.Type); err != nil {
			fmt.Fprintf(g.stderr, "❌ Error auto-wiring middleware: %v\n", err)
			fmt.Fprintln(g.stdout, "💡 Your middleware was created but you'll need to manually wire it up")
			fmt.Fprintf(g.stdout, "   You can try: foundry wire middleware %s\n", options.Type)
			g.showSuccess(options, getMiddlewareInfo(options.Type), false)
		} else {
			g.showSuccess(options, getMiddlewareInfo(options.Type), true)
		}
	} else {
		g.showSuccess(options, getMiddlewareInfo(options.Type), false)
	}

	return nil
}

// createLegacyMiddlewareFile creates a middleware file using legacy templates
func (g *MiddlewareGenerator) createLegacyMiddlewareFile(middlewarePath, middlewareType string) error {
	template := getLegacyMiddlewareTemplate(middlewareType)
	return writeFile(middlewarePath, template)
}

// getLayoutManager gets the layout manager instance
func (g *MiddlewareGenerator) getLayoutManager() (*layout.Manager, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	configPath := filepath.Join(homeDir, ".foundry", "layouts.yaml")
	return layout.NewManager(configPath)
}

// wireMiddleware attempts to auto-wire middleware into the application
func (g *MiddlewareGenerator) wireMiddleware(middlewareType string) error {
	// TODO: Implement auto-wiring logic
	// For now, return a placeholder implementation
	fmt.Fprintf(g.stdout, "⚠️  Auto-wiring not yet implemented - use manual instructions\n")
	return fmt.Errorf("auto-wiring not implemented")
}

// showSuccess displays success message with instructions
func (g *MiddlewareGenerator) showSuccess(options MiddlewareOptions, middlewareInfo MiddlewareInfo, autoWired bool) {
	middlewarePath := filepath.Join(options.OutputDir, fmt.Sprintf("%s.go", options.Type))

	wireStatus := ""
	if autoWired {
		wireStatus = `
🔌 Middleware automatically wired into your router!

`
	} else {
		wireStatus = `
📌 Manual wiring required:
  Run: foundry wire middleware ` + options.Type + `
  Or manually update your main.go file

`
	}

	usage := getMiddlewareUsage(options.Type)

	fmt.Fprintf(g.stdout, `✅ Middleware created successfully!

📁 Created:
  %s

📝 Description:
  %s
%s%s
`, middlewarePath, middlewareInfo.Description, wireStatus, usage)
}

// MiddlewareInfo holds information about middleware types
type MiddlewareInfo struct {
	Type        string
	Description string
}

// getMiddlewareInfo returns info for a specific middleware type
func getMiddlewareInfo(middlewareType string) MiddlewareInfo {
	switch middlewareType {
	case "auth":
		return MiddlewareInfo{Type: "auth", Description: "Authentication middleware"}
	case "ratelimit":
		return MiddlewareInfo{Type: "ratelimit", Description: "Rate limiting middleware"}
	case "cors":
		return MiddlewareInfo{Type: "cors", Description: "CORS middleware"}
	case "logging":
		return MiddlewareInfo{Type: "logging", Description: "Request/response logging middleware"}
	case "recovery":
		return MiddlewareInfo{Type: "recovery", Description: "Panic recovery middleware"}
	case "timeout":
		return MiddlewareInfo{Type: "timeout", Description: "Request timeout middleware"}
	case "compression":
		return MiddlewareInfo{Type: "compression", Description: "Response compression middleware"}
	default:
		return MiddlewareInfo{Type: middlewareType, Description: "Custom middleware"}
	}
}

// getMiddlewareUsage returns usage instructions for a middleware type
func getMiddlewareUsage(middlewareType string) string {
	switch middlewareType {
	case "auth":
		return `📌 Usage in your handlers:

1. The middleware is now active and will protect your routes
2. Get user info from context in your handlers:

   func (h *Handler) ProtectedEndpoint(w http.ResponseWriter, r *http.Request) {
       userID := r.Context().Value("userID").(string)
       // Use userID...
   }

3. Test your protected endpoints:
   # This should return 401 Unauthorized:
   curl http://localhost:8080/api/v1/users
   
   # This should work:
   curl -H "Authorization: Bearer valid-token" http://localhost:8080/api/v1/users`

	case "ratelimit":
		return `📌 Configuration:

The middleware is configured with default limits (100 requests/minute).
To customize, edit the middleware call in main.go:

   r.Use(middleware.RateLimitMiddleware(50, time.Minute))  // 50 req/min
   r.Use(middleware.RateLimitMiddleware(1000, time.Hour))  // 1000 req/hour`

	case "cors":
		return `📌 Configuration:

The CORS middleware is configured with permissive defaults.
For production, edit internal/middleware/cors.go:

   - Change allowed origins from "*" to your domain
   - Customize allowed methods and headers
   - Set appropriate max age for preflight requests`

	case "logging":
		return `📌 Usage:

The logging middleware will now log all HTTP requests.
Log format: [METHOD] /path remote_addr status bytes duration

To customize logging, edit internal/middleware/logging.go`

	case "recovery":
		return `📌 Usage:

The recovery middleware will catch panics and return 500 errors.
Your server will stay running even if handlers panic.

Check your logs for panic details when they occur.`

	case "timeout":
		return `📌 Configuration:

Default timeout is set to 30 seconds.
To customize, edit the middleware call in main.go:

   r.Use(middleware.TimeoutMiddleware(60 * time.Second))  // 60 second timeout`

	case "compression":
		return `📌 Usage:

Responses will now be automatically compressed with gzip.
The middleware only compresses if the client supports it.

Test with: curl -H "Accept-Encoding: gzip" http://localhost:8080/api/v1/users -v`

	default:
		return `📌 Next steps:
  - Review the generated middleware code
  - Customize the middleware logic as needed
  - Add tests for the middleware`
	}
}

// getLegacyMiddlewareTemplate returns the legacy middleware template
func getLegacyMiddlewareTemplate(middlewareType string) string {
	switch middlewareType {
	case "auth":
		return `package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"
)

// AuthMiddleware validates JWT tokens and protects routes
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get the authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Missing authorization header", http.StatusUnauthorized)
			return
		}

		// Check if it's a Bearer token
		if !strings.HasPrefix(authHeader, "Bearer ") {
			http.Error(w, "Invalid authorization header format", http.StatusUnauthorized)
			return
		}

		// Extract the token
		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token == "" {
			http.Error(w, "Missing token", http.StatusUnauthorized)
			return
		}

		// Validate the token (implement your validation logic here)
		userID, err := validateToken(token)
		if err != nil {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		// Add user info to context
		ctx := context.WithValue(r.Context(), "user_id", userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// validateToken validates a JWT token and returns the user ID
// TODO: Implement your token validation logic
func validateToken(token string) (string, error) {
	// Placeholder implementation
	if token == "valid-token" {
		return "user123", nil
	}
	
	return "", fmt.Errorf("invalid token")
}
`

	case "logging":
		return `package middleware

import (
	"log"
	"net/http"
	"time"
)

// responseWriter wraps http.ResponseWriter to capture status code and size
type responseWriter struct {
	http.ResponseWriter
	status int
	size   int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	size, err := rw.ResponseWriter.Write(b)
	rw.size += size
	return size, err
}

// LoggingMiddleware logs HTTP requests
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		
		// Wrap the response writer
		wrapped := &responseWriter{
			ResponseWriter: w,
			status:         http.StatusOK,
		}
		
		// Call the next handler
		next.ServeHTTP(wrapped, r)
		
		// Log the request
		duration := time.Since(start)
		log.Printf("[%s] %s %s %d %d %v",
			r.Method,
			r.RequestURI,
			r.RemoteAddr,
			wrapped.status,
			wrapped.size,
			duration,
		)
	})
}
`

	default:
		return `package middleware

import (
	"net/http"
)

// ` + capitalize(middlewareType) + `Middleware implements ` + middlewareType + ` middleware
func ` + capitalize(middlewareType) + `Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO: Implement ` + middlewareType + ` middleware logic
		
		// Call the next handler
		next.ServeHTTP(w, r)
	})
}
`
	}
}
