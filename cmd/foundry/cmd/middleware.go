package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/shapestone/foundry/internal/generator"
	"github.com/shapestone/foundry/internal/middleware"
	"github.com/spf13/cobra"
)

var middlewareCmd = &cobra.Command{
	Use:   "middleware [type]",
	Short: "Add middleware to your project",
	Args:  cobra.ExactArgs(1),
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if len(args) > 0 {
			if err := ValidateComponentName(args[0]); err != nil {
				return fmt.Errorf("invalid middleware name: %v", err)
			}
		}
		return nil
	},
	Example: `  foundry add middleware auth
  foundry add middleware ratelimit
  foundry add middleware cors
  foundry add middleware logging
  foundry add middleware recovery
  foundry add middleware timeout
  foundry add middleware compression`,
	RunE: func(cmd *cobra.Command, args []string) error {
		autoWire, _ := cmd.Flags().GetBool("auto-wire")
		dryRun, _ := cmd.Flags().GetBool("dry-run")
		return addMiddleware(args[0], autoWire, dryRun)
	},
}

var supportedMiddleware = map[string]string{
	"auth":        "Authentication middleware",
	"ratelimit":   "Rate limiting middleware",
	"cors":        "CORS middleware",
	"logging":     "Request/response logging middleware",
	"recovery":    "Panic recovery middleware",
	"timeout":     "Request timeout middleware",
	"compression": "Response compression middleware",
}

func init() {
	middlewareCmd.Flags().Bool("auto-wire", false, "Automatically wire the middleware into your router")
	middlewareCmd.Flags().Bool("dry-run", false, "Preview changes without applying them")
}

func addMiddleware(middlewareType string, autoWire, dryRun bool) error {
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

	if !dryRun {
		fmt.Printf("🔨 Adding middleware: %s\n", middlewareType)
	}

	// Create middleware directory
	middlewareDir := filepath.Join("internal", "middleware")
	middlewarePath := filepath.Join(middlewareDir, fmt.Sprintf("%s.go", middlewareType))

	// Check if middleware already exists
	if _, err := os.Stat(middlewarePath); err == nil && !dryRun {
		return fmt.Errorf("middleware %s already exists", middlewarePath)
	}

	if dryRun {
		fmt.Printf("Would create middleware: %s\n", middlewarePath)
		if autoWire {
			fmt.Printf("Would auto-wire middleware into router\n")
		}
		return nil
	}

	// Generate the middleware file
	if err := generator.CreateMiddleware(middlewareType); err != nil {
		return fmt.Errorf("failed to create middleware: %w", err)
	}

	// Handle auto-wiring
	if autoWire {
		fmt.Println("\n🔄 Auto-wiring middleware...")

		// Create auto-wirer
		autoWirer := middleware.NewAutoWirer(".")

		// Wire the middleware
		if err := autoWirer.WireMiddleware(middlewareType, false); err != nil {
			fmt.Printf("❌ Error auto-wiring middleware: %v\n", err)
			fmt.Println("💡 Your middleware was created but you'll need to manually wire it up")
			fmt.Printf("   You can try: foundry wire middleware %s\n", middlewareType)
			showMiddlewareSuccess(middlewareType, middlewarePath, description, false)
		} else {
			showMiddlewareSuccess(middlewareType, middlewarePath, description, true)
		}
	} else {
		// Show manual wiring instructions
		showMiddlewareSuccess(middlewareType, middlewarePath, description, false)
	}

	return nil
}

type MiddlewareTemplateData struct {
	MiddlewareType string
	ModuleName     string
}

func getSupportedMiddlewareList() []string {
	var types []string
	for t := range supportedMiddleware {
		types = append(types, t)
	}
	return types
}

func showMiddlewareSuccess(middlewareType, path, description string, autoWired bool) {
	var usage string

	// Add note about auto-wiring status
	wireStatus := ""
	if autoWired {
		wireStatus = `
🔌 Middleware automatically wired into your router!

`
	} else {
		wireStatus = `
📌 Manual wiring required:
  Run: foundry wire middleware ` + middlewareType + `
  Or manually update your main.go file

`
	}

	switch middlewareType {
	case "auth":
		usage = fmt.Sprintf(`
📌 Usage in your handlers:

1. The middleware is now active and will protect your routes
2. Get user info from context in your handlers:

   func (h *Handler) ProtectedEndpoint(w http.ResponseWriter, r *http.Request) {
       userID := r.Context().Value("user_id").(string)
       // Use userID...
   }

3. Test your protected endpoints:
   # This should return 401 Unauthorized:
   curl http://localhost:8080/api/v1/users
   
   # This should work:
   curl -H "Authorization: Bearer valid-token" http://localhost:8080/api/v1/users`)

	case "ratelimit":
		usage = `
📌 Configuration:

The middleware is configured with default limits (100 requests/minute).
To customize, edit the middleware call in main.go:

   r.Use(middleware.RateLimitMiddleware(50, time.Minute))  // 50 req/min
   r.Use(middleware.RateLimitMiddleware(1000, time.Hour))  // 1000 req/hour`

	case "cors":
		usage = `
📌 Configuration:

The CORS middleware is configured with permissive defaults.
For production, edit internal/middleware/cors.go:

   - Change allowed origins from "*" to your domain
   - Customize allowed methods and headers
   - Set appropriate max age for preflight requests`

	case "logging":
		usage = `
📌 Usage:

The logging middleware will now log all HTTP requests.
Log format: [METHOD] /path remote_addr status bytes duration

To customize logging, edit internal/middleware/logging.go`

	case "recovery":
		usage = `
📌 Usage:

The recovery middleware will catch panics and return 500 errors.
Your server will stay running even if handlers panic.

Check your logs for panic details when they occur.`

	case "timeout":
		usage = `
📌 Configuration:

Default timeout is set to 30 seconds.
To customize, edit the middleware call in main.go:

   r.Use(middleware.TimeoutMiddleware(60 * time.Second))  // 60 second timeout`

	case "compression":
		usage = `
📌 Usage:

Responses will now be automatically compressed with gzip.
The middleware only compresses if the client supports it.

Test with: curl -H "Accept-Encoding: gzip" http://localhost:8080/api/v1/users -v`

	default:
		usage = `
📌 Next steps:
  - Review the generated middleware code
  - Customize the middleware logic as needed
  - Add tests for the middleware`
	}

	fmt.Printf(`✅ Middleware created successfully!

📁 Created:
  %s

📝 Description:
  %s
%s%s
`, path, description, wireStatus, usage)
}
