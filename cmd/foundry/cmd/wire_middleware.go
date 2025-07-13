package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/shapestone/foundry/internal/middleware"
	"github.com/spf13/cobra"
)

var wireMiddlewareCmd = &cobra.Command{
	Use:   "middleware [type]",
	Short: "Wire existing middleware into your application",
	Args:  cobra.ExactArgs(1),
	Example: `  foundry wire middleware auth
  foundry wire middleware cors
  foundry wire middleware ratelimit`,
	RunE: func(cmd *cobra.Command, args []string) error {
		dryRun, _ := cmd.Flags().GetBool("dry-run")
		return wireMiddleware(args[0], dryRun)
	},
}

func init() {
	wireMiddlewareCmd.Flags().Bool("dry-run", false, "Preview changes without applying them")
}

func wireMiddleware(middlewareType string, dryRun bool) error {
	// Check if we're in a Go project
	if _, err := os.Stat("go.mod"); os.IsNotExist(err) {
		return fmt.Errorf("go.mod not found. Please run this command from your project root")
	}

	// Validate middleware type
	description, ok := supportedMiddleware[middlewareType]
	if !ok {
		return fmt.Errorf("unsupported middleware type '%s'. Supported types: %v",
			middlewareType, getSupportedMiddlewareList())
	}

	// Check if middleware file exists
	middlewarePath := filepath.Join("internal", "middleware", fmt.Sprintf("%s.go", middlewareType))
	if _, err := os.Stat(middlewarePath); os.IsNotExist(err) {
		return fmt.Errorf("middleware %s not found at %s. Did you mean to run: foundry add middleware %s",
			middlewareType, middlewarePath, middlewareType)
	}

	if !dryRun {
		fmt.Printf("🔌 Wiring middleware: %s\n", middlewareType)
	}

	// Create auto-wirer and wire the middleware
	autoWirer := middleware.NewAutoWirer(".")

	if err := autoWirer.WireMiddleware(middlewareType, dryRun); err != nil {
		// If auto-wiring fails, show manual instructions
		if !dryRun {
			fmt.Printf("❌ Auto-wiring failed: %v\n", err)
			fmt.Printf("💡 Manual wiring instructions for %s middleware:\n", middlewareType)
			showManualWiringInstructions(middlewareType, description)
		}
		return err
	}

	if !dryRun {
		fmt.Printf("✅ Middleware %s wired successfully!\n", middlewareType)
		fmt.Printf("📝 %s\n", description)

		// Show middleware-specific usage tips
		showMiddlewareUsageTips(middlewareType)
	}

	return nil
}

// showManualWiringInstructions provides fallback manual instructions
func showManualWiringInstructions(middlewareType, description string) {
	middlewareName := strings.Title(middlewareType) + "Middleware"
	fmt.Printf(`
Manual wiring steps for %s middleware:

1. Add import to your main.go:
   import "yourmodule/internal/middleware"

2. Add middleware to your router:
   r.Use(middleware.%s)

3. Make sure it's positioned correctly:
   - Recovery and CORS: First
   - Logging: After recovery
   - Auth and rate limiting: Before routes

Example router setup:
   r := chi.NewRouter()
   r.Use(middleware.RecoveryMiddleware)
   r.Use(middleware.LoggingMiddleware)
   r.Use(middleware.%s)
   r.Route("/api/v1", routes.RegisterAPIRoutes)
`, middlewareType, middlewareName, middlewareName)
}

// showMiddlewareUsageTips provides specific usage information
func showMiddlewareUsageTips(middlewareType string) {
	switch middlewareType {
	case "auth":
		fmt.Println(`
💡 Usage Tips:
  - Update validateToken() in internal/middleware/auth.go with your auth logic
  - Access user info in handlers: userID := r.Context().Value("user_id").(string)
  - Test with: curl -H "Authorization: Bearer your-token" http://localhost:8080/api/v1/endpoint`)

	case "ratelimit":
		fmt.Println(`
💡 Usage Tips:
  - Default: 100 requests/minute per IP
  - Customize limits in your middleware call: RateLimitMiddleware(50, time.Minute)
  - Returns 429 Too Many Requests when limit exceeded`)

	case "cors":
		fmt.Println(`
💡 Usage Tips:
  - Default allows all origins (*) - change for production!
  - Customize origins in internal/middleware/cors.go
  - Test with: curl -H "Origin: https://yoursite.com" -X OPTIONS http://localhost:8080/api/v1/endpoint`)

	case "logging":
		fmt.Println(`
💡 Usage Tips:
  - Logs all HTTP requests with method, path, status, duration
  - Customize log format in internal/middleware/logging.go
  - Use with request ID middleware for better tracing`)

	case "recovery":
		fmt.Println(`
💡 Usage Tips:
  - Should be the first middleware in your chain
  - Catches panics and returns 500 Internal Server Error
  - Check logs for panic details and stack traces`)

	case "timeout":
		fmt.Println(`
💡 Usage Tips:
  - Default: 30 second timeout
  - Customize: TimeoutMiddleware(60 * time.Second)
  - Returns 408 Request Timeout for slow requests`)

	case "compression":
		fmt.Println(`
💡 Usage Tips:
  - Automatically compresses responses with gzip
  - Only compresses if client sends Accept-Encoding: gzip
  - Test with: curl -H "Accept-Encoding: gzip" http://localhost:8080/api/v1/endpoint -v`)
	}
}
