// internal/middleware/autowiring.go
package middleware

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/shapestone/foundry/internal/project"
)

// RouterPattern represents different router frameworks
type RouterPattern string

const (
	RouterChi     RouterPattern = "chi"
	RouterGin     RouterPattern = "gin"
	RouterGorilla RouterPattern = "gorilla"
	RouterHTTP    RouterPattern = "http"
)

// MiddlewarePosition defines where middleware should be inserted
type MiddlewarePosition int

const (
	PositionEarly  MiddlewarePosition = iota // Recovery, CORS
	PositionMiddle                           // Logging, Request ID
	PositionLate                             // Auth, Rate limiting
)

// AutoWirer handles automatic middleware wiring
type AutoWirer struct {
	projectPath string
	moduleName  string
}

// NewAutoWirer creates a new middleware auto-wirer
func NewAutoWirer(projectPath string) *AutoWirer {
	return &AutoWirer{
		projectPath: projectPath,
		moduleName:  project.GetCurrentModule(),
	}
}

// WireMiddleware automatically wires middleware into the project
func (aw *AutoWirer) WireMiddleware(middlewareType string, dryRun bool) error {
	// Find main.go file
	mainFile, err := aw.findMainFile()
	if err != nil {
		return fmt.Errorf("failed to find main.go: %w", err)
	}

	// Detect router pattern
	pattern, err := aw.detectRouterPattern(mainFile)
	if err != nil {
		return fmt.Errorf("failed to detect router pattern: %w", err)
	}

	// Read current content
	content, err := os.ReadFile(mainFile)
	if err != nil {
		return fmt.Errorf("failed to read %s: %w", mainFile, err)
	}

	originalContent := string(content)

	// Check if middleware is already wired
	if aw.isMiddlewareAlreadyWired(originalContent, middlewareType) {
		return fmt.Errorf("middleware %s appears to already be wired", middlewareType)
	}

	// Generate the wired content
	newContent, err := aw.wireMiddlewareInContent(originalContent, middlewareType, pattern)
	if err != nil {
		return fmt.Errorf("failed to wire middleware: %w", err)
	}

	// Show preview
	if !aw.showPreview(mainFile, originalContent, newContent, middlewareType) {
		return fmt.Errorf("wiring cancelled by user")
	}

	if dryRun {
		fmt.Println("🔍 Dry run complete - no changes made")
		return nil
	}

	// Write the updated content
	if err := os.WriteFile(mainFile, []byte(newContent), 0644); err != nil {
		return fmt.Errorf("failed to write %s: %w", mainFile, err)
	}

	fmt.Printf("✅ Middleware %s wired successfully!\n", middlewareType)
	return nil
}

// findMainFile locates the main.go file in the project
func (aw *AutoWirer) findMainFile() (string, error) {
	// Common locations for main.go (check root first since it's most common)
	candidates := []string{
		"main.go", // Root (most common)
		filepath.Join("cmd", project.GetProjectName(), "main.go"), // Standard layout
		"cmd/main.go", // Generic cmd
	}

	for _, candidate := range candidates {
		fullPath := filepath.Join(aw.projectPath, candidate)
		if _, err := os.Stat(fullPath); err == nil {
			return fullPath, nil
		}
	}

	return "", fmt.Errorf("main.go not found in expected locations")
}

// detectRouterPattern analyzes the main.go file to detect the router pattern
func (aw *AutoWirer) detectRouterPattern(mainFile string) (RouterPattern, error) {
	content, err := os.ReadFile(mainFile)
	if err != nil {
		return "", err
	}

	contentStr := string(content)

	// Check imports and usage patterns
	patterns := map[RouterPattern][]string{
		RouterChi: {
			"github.com/go-chi/chi",
			"chi.NewRouter",
			".Use(",
		},
		RouterGin: {
			"github.com/gin-gonic/gin",
			"gin.Default",
			"gin.New",
		},
		RouterGorilla: {
			"github.com/gorilla/mux",
			"mux.NewRouter",
		},
		RouterHTTP: {
			"net/http",
			"http.HandleFunc",
			"http.ListenAndServe",
		},
	}

	for pattern, indicators := range patterns {
		matches := 0
		for _, indicator := range indicators {
			if strings.Contains(contentStr, indicator) {
				matches++
			}
		}
		// If we find multiple indicators, this is likely the pattern
		if matches >= 2 || (pattern == RouterHTTP && matches >= 1) {
			return pattern, nil
		}
	}

	// Default to Chi if no clear pattern
	return RouterChi, nil
}

// isMiddlewareAlreadyWired checks if middleware is already configured
func (aw *AutoWirer) isMiddlewareAlreadyWired(content, middlewareType string) bool {
	middlewareName := strings.Title(middlewareType) + "Middleware"
	return strings.Contains(content, middlewareName)
}

// wireMiddlewareInContent adds middleware to the router configuration
func (aw *AutoWirer) wireMiddlewareInContent(content, middlewareType string, pattern RouterPattern) (string, error) {
	lines := strings.Split(content, "\n")
	result := make([]string, 0, len(lines)+5) // Pre-allocate with extra space

	// Add import if not present
	importAdded := false
	middlewareAdded := false

	middlewareImport := fmt.Sprintf("\"%s/internal/middleware\"", aw.moduleName)

	// Get middleware position
	position := aw.getMiddlewarePosition(middlewareType)

	for i, line := range lines {
		result = append(result, line)

		// Add import after other imports
		if !importAdded && aw.isImportSection(line, i, lines) {
			// Look for the closing of import block or add after current import
			if strings.Contains(line, ")") && strings.TrimSpace(line) == ")" {
				// Insert before closing parenthesis
				result[len(result)-1] = fmt.Sprintf("\t%s", middlewareImport)
				result = append(result, line)
			} else if strings.Contains(line, "import") && !strings.Contains(line, "(") {
				// Single import line, add after it
				result = append(result, fmt.Sprintf("import %s", middlewareImport))
			}
			importAdded = true
		}

		// Add middleware use statement
		if !middlewareAdded && aw.isRouterSetupLocation(line, pattern, position, lines, i) {
			middlewareCall := aw.generateMiddlewareCall(middlewareType, pattern)
			result = append(result, middlewareCall)
			middlewareAdded = true
		}
	}

	if !middlewareAdded {
		return "", fmt.Errorf("could not find suitable location to add middleware")
	}

	return strings.Join(result, "\n"), nil
}

// isImportSection determines if we're in the import section
func (aw *AutoWirer) isImportSection(line string, index int, lines []string) bool {
	trimmed := strings.TrimSpace(line)

	// Single import line
	if strings.HasPrefix(trimmed, "import ") && !strings.Contains(trimmed, "(") {
		return true
	}

	// Multi-line import block - look for the import statement
	if strings.Contains(trimmed, "import (") {
		return true
	}

	// We're inside an import block - check if we haven't hit the closing )
	if index > 0 {
		// Look backwards to find the import block start
		for i := index - 1; i >= 0; i-- {
			prevLine := strings.TrimSpace(lines[i])
			if strings.Contains(prevLine, "import (") {
				// We're in an import block, check if current line is before the closing )
				return !strings.Contains(trimmed, ")") || (strings.Contains(trimmed, ")") && strings.Contains(trimmed, "\""))
			}
			// If we find a non-empty, non-comment line that's not an import, we're not in imports
			if prevLine != "" && !strings.HasPrefix(prevLine, "//") && !strings.Contains(prevLine, "import") {
				break
			}
		}
	}

	return false
}

// isRouterSetupLocation determines if this is where we should add middleware
func (aw *AutoWirer) isRouterSetupLocation(line string, pattern RouterPattern, position MiddlewarePosition, lines []string, index int) bool {
	trimmed := strings.TrimSpace(line)

	switch pattern {
	case RouterChi:
		// Look for router creation or existing middleware
		if strings.Contains(trimmed, "chi.NewRouter()") || strings.Contains(trimmed, ":= chi.NewRouter()") {
			return position == PositionEarly
		}

		// Insert based on position relative to existing middleware
		if strings.Contains(trimmed, ".Use(") {
			return aw.shouldInsertAtPosition(trimmed, position)
		}

		// Before route definitions
		if strings.Contains(trimmed, ".Route(") || strings.Contains(trimmed, ".Get(") ||
			strings.Contains(trimmed, ".Post(") || strings.Contains(trimmed, ".Mount(") {
			return true
		}

	case RouterGin:
		if strings.Contains(trimmed, "gin.Default()") || strings.Contains(trimmed, "gin.New()") {
			return position == PositionEarly
		}

		if strings.Contains(trimmed, ".Use(") {
			return aw.shouldInsertAtPosition(trimmed, position)
		}

	case RouterGorilla:
		// Look for router creation - insert immediately after for early middleware
		if strings.Contains(trimmed, "mux.NewRouter()") || strings.Contains(trimmed, ":= mux.NewRouter()") {
			return position == PositionEarly
		}

		// Look for existing middleware usage - insert based on position
		if strings.Contains(trimmed, "router.Use(") {
			return aw.shouldInsertAtPosition(trimmed, position)
		}

		// Insert before route setup for early/middle middleware if no existing middleware
		if position <= PositionMiddle {
			if strings.Contains(trimmed, "router.HandleFunc(") ||
				strings.Contains(trimmed, ".PathPrefix(") ||
				strings.Contains(trimmed, "setupRoutes(") {
				return true
			}
		}

		// Insert before server configuration for late middleware
		if strings.Contains(trimmed, "&http.Server{") || strings.Contains(trimmed, "srv := &http.Server") {
			return true
		}

	case RouterHTTP:
		// For net/http, add before http.ListenAndServe
		if strings.Contains(trimmed, "http.ListenAndServe") {
			return true
		}
	}

	return false
}

// shouldInsertAtPosition determines if middleware should be inserted at this position
func (aw *AutoWirer) shouldInsertAtPosition(line string, position MiddlewarePosition) bool {
	line = strings.ToLower(line)

	switch position {
	case PositionEarly:
		// Insert before any middleware, or at the very beginning
		return true

	case PositionMiddle:
		// Insert after early middleware, before auth/rate limiting
		hasEarly := strings.Contains(line, "recovery") || strings.Contains(line, "cors")
		hasLate := strings.Contains(line, "auth") || strings.Contains(line, "ratelimit")
		return hasEarly || !hasLate

	case PositionLate:
		// Insert after all other middleware
		return true
	}

	return false
}

// generateMiddlewareCall creates the appropriate middleware call for the router pattern
func (aw *AutoWirer) generateMiddlewareCall(middlewareType string, pattern RouterPattern) string {
	middlewareName := strings.Title(middlewareType) + "Middleware"

	switch pattern {
	case RouterChi, RouterGin:
		if middlewareType == "ratelimit" {
			return fmt.Sprintf("\tr.Use(middleware.RateLimitMiddleware(100, time.Minute))")
		}
		if middlewareType == "timeout" {
			return fmt.Sprintf("\tr.Use(middleware.TimeoutMiddleware(30 * time.Second))")
		}
		return fmt.Sprintf("\tr.Use(middleware.%s)", middlewareName)

	case RouterGorilla:
		if middlewareType == "ratelimit" {
			return fmt.Sprintf("\trouter.Use(middleware.RateLimitMiddleware(100, time.Minute))")
		}
		if middlewareType == "timeout" {
			return fmt.Sprintf("\trouter.Use(middleware.TimeoutMiddleware(30 * time.Second))")
		}
		return fmt.Sprintf("\trouter.Use(middleware.%s)", middlewareName)

	case RouterHTTP:
		// For net/http, we'd need to wrap handlers
		return fmt.Sprintf("\t// Add middleware.%s to your handler chain", middlewareName)
	}

	return fmt.Sprintf("\trouter.Use(middleware.%s)", middlewareName)
}

// getMiddlewarePosition returns the recommended position for middleware type
func (aw *AutoWirer) getMiddlewarePosition(middlewareType string) MiddlewarePosition {
	switch middlewareType {
	case "recovery", "cors":
		return PositionEarly
	case "logging", "compression":
		return PositionMiddle
	case "auth", "ratelimit", "timeout":
		return PositionLate
	default:
		return PositionMiddle
	}
}

// showPreview displays the changes and asks for confirmation
func (aw *AutoWirer) showPreview(filename, oldContent, newContent, middlewareType string) bool {
	fmt.Printf("🔍 Preview changes to %s:\n\n", filename)

	// Show diff-like output
	oldLines := strings.Split(oldContent, "\n")
	newLines := strings.Split(newContent, "\n")

	// Simple diff - find added lines
	addedLines := []string{}
	for _, newLine := range newLines {
		found := false
		for _, oldLine := range oldLines {
			if newLine == oldLine {
				found = true
				break
			}
		}
		if !found && strings.TrimSpace(newLine) != "" {
			addedLines = append(addedLines, newLine)
		}
	}

	for _, line := range addedLines {
		fmt.Printf("+ %s\n", line)
	}

	fmt.Printf("\nThis will add the %s middleware to your router.\n", middlewareType)
	fmt.Print("Apply these changes? (y/n): ")

	scanner := bufio.NewScanner(os.Stdin)
	if scanner.Scan() {
		response := strings.ToLower(strings.TrimSpace(scanner.Text()))
		return response == "y" || response == "yes"
	}

	return false
}
