package templates

// Middleware Templates

const AuthMiddlewareTemplate = `package middleware

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
	// In a real application, you would:
	// 1. Parse and validate the JWT token
	// 2. Check expiration
	// 3. Verify signature
	// 4. Extract user ID from claims
	
	if token == "valid-token" {
		return "user123", nil
	}
	
	return "", fmt.Errorf("invalid token")
}
`

const RateLimitMiddlewareTemplate = `package middleware

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"
)

// RateLimiter tracks requests per IP
type RateLimiter struct {
	requests map[string][]time.Time
	mu       sync.RWMutex
	limit    int
	window   time.Duration
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		requests: make(map[string][]time.Time),
		limit:    limit,
		window:   window,
	}
	
	// Clean up old entries periodically
	go rl.cleanup()
	
	return rl
}

// RateLimitMiddleware creates rate limiting middleware
func RateLimitMiddleware(limit int, window time.Duration) func(http.Handler) http.Handler {
	limiter := NewRateLimiter(limit, window)
	
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := getClientIP(r)
			
			if !limiter.Allow(ip) {
				w.Header().Set("Retry-After", fmt.Sprintf("%.0f", window.Seconds()))
				http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
				return
			}
			
			next.ServeHTTP(w, r)
		})
	}
}

// Allow checks if a request from the given IP is allowed
func (rl *RateLimiter) Allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	
	now := time.Now()
	cutoff := now.Add(-rl.window)
	
	// Get existing requests for this IP
	requests := rl.requests[ip]
	
	// Filter out old requests
	var validRequests []time.Time
	for _, reqTime := range requests {
		if reqTime.After(cutoff) {
			validRequests = append(validRequests, reqTime)
		}
	}
	
	// Check if under limit
	if len(validRequests) >= rl.limit {
		rl.requests[ip] = validRequests
		return false
	}
	
	// Add current request
	validRequests = append(validRequests, now)
	rl.requests[ip] = validRequests
	
	return true
}

// cleanup removes old entries periodically
func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(rl.window)
	defer ticker.Stop()
	
	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		cutoff := now.Add(-rl.window)
		
		for ip, requests := range rl.requests {
			var validRequests []time.Time
			for _, reqTime := range requests {
				if reqTime.After(cutoff) {
					validRequests = append(validRequests, reqTime)
				}
			}
			
			if len(validRequests) == 0 {
				delete(rl.requests, ip)
			} else {
				rl.requests[ip] = validRequests
			}
		}
		rl.mu.Unlock()
	}
}

// getClientIP extracts the client IP from the request
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return strings.Split(xff, ",")[0]
	}
	
	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}
	
	// Fall back to remote address
	return strings.Split(r.RemoteAddr, ":")[0]
}
`

const CORSMiddlewareTemplate = `package middleware

import (
	"net/http"
	"strconv"
	"strings"
)

// CORSConfig holds CORS configuration
type CORSConfig struct {
	AllowedOrigins   []string
	AllowedMethods   []string
	AllowedHeaders   []string
	ExposedHeaders   []string
	AllowCredentials bool
	MaxAge           int
}

// DefaultCORSConfig returns a default CORS configuration
func DefaultCORSConfig() CORSConfig {
	return CORSConfig{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"},
		AllowedHeaders: []string{"Origin", "Content-Length", "Content-Type", "Authorization"},
		ExposedHeaders: []string{},
		AllowCredentials: false,
		MaxAge: 86400, // 24 hours
	}
}

// CORSMiddleware creates CORS middleware with default configuration
func CORSMiddleware(next http.Handler) http.Handler {
	return CORSWithConfig(DefaultCORSConfig())(next)
}

// CORSWithConfig creates CORS middleware with custom configuration
func CORSWithConfig(config CORSConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			
			// Set Access-Control-Allow-Origin
			if len(config.AllowedOrigins) == 1 && config.AllowedOrigins[0] == "*" {
				w.Header().Set("Access-Control-Allow-Origin", "*")
			} else if isOriginAllowed(origin, config.AllowedOrigins) {
				w.Header().Set("Access-Control-Allow-Origin", origin)
			}
			
			// Set other CORS headers
			if len(config.AllowedMethods) > 0 {
				w.Header().Set("Access-Control-Allow-Methods", strings.Join(config.AllowedMethods, ", "))
			}
			
			if len(config.AllowedHeaders) > 0 {
				w.Header().Set("Access-Control-Allow-Headers", strings.Join(config.AllowedHeaders, ", "))
			}
			
			if len(config.ExposedHeaders) > 0 {
				w.Header().Set("Access-Control-Expose-Headers", strings.Join(config.ExposedHeaders, ", "))
			}
			
			if config.AllowCredentials {
				w.Header().Set("Access-Control-Allow-Credentials", "true")
			}
			
			if config.MaxAge > 0 {
				w.Header().Set("Access-Control-Max-Age", strconv.Itoa(config.MaxAge))
			}
			
			// Handle preflight requests
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			
			next.ServeHTTP(w, r)
		})
	}
}

// isOriginAllowed checks if an origin is in the allowed list
func isOriginAllowed(origin string, allowedOrigins []string) bool {
	for _, allowedOrigin := range allowedOrigins {
		if allowedOrigin == "*" || allowedOrigin == origin {
			return true
		}
	}
	return false
}
`

const LoggingMiddlewareTemplate = `package middleware

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

const RecoveryMiddlewareTemplate = `package middleware

import (
	"log"
	"net/http"
	"runtime/debug"
)

// RecoveryMiddleware recovers from panics and returns a 500 error
func RecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				// Log the panic
				log.Printf("Panic recovered: %v\n%s", err, debug.Stack())
				
				// Return 500 Internal Server Error
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}()
		
		next.ServeHTTP(w, r)
	})
}
`

const TimeoutMiddlewareTemplate = `package middleware

import (
	"context"
	"net/http"
	"time"
)

// TimeoutMiddleware creates a middleware that times out requests
func TimeoutMiddleware(timeout time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Create a context with timeout
			ctx, cancel := context.WithTimeout(r.Context(), timeout)
			defer cancel()
			
			// Create a channel to signal completion
			done := make(chan struct{})
			
			// Run the handler in a goroutine
			go func() {
				defer func() {
					// Recover from any panics in the handler
					if err := recover(); err != nil {
						// Let the recovery middleware handle this
						panic(err)
					}
					close(done)
				}()
				
				next.ServeHTTP(w, r.WithContext(ctx))
			}()
			
			// Wait for completion or timeout
			select {
			case <-done:
				// Handler completed normally
				return
			case <-ctx.Done():
				// Request timed out
				if ctx.Err() == context.DeadlineExceeded {
					http.Error(w, "Request Timeout", http.StatusRequestTimeout)
				}
				return
			}
		})
	}
}
`

const CompressionMiddlewareTemplate = `package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

// gzipResponseWriter wraps http.ResponseWriter with gzip compression
type gzipResponseWriter struct {
	io.Writer
	http.ResponseWriter
}

func (w gzipResponseWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

// CompressionMiddleware adds gzip compression to responses
func CompressionMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if the client accepts gzip encoding
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}
		
		// Set the content encoding header
		w.Header().Set("Content-Encoding", "gzip")
		w.Header().Set("Vary", "Accept-Encoding")
		
		// Create a gzip writer
		gz := gzip.NewWriter(w)
		defer gz.Close()
		
		// Wrap the response writer
		gzipWriter := gzipResponseWriter{Writer: gz, ResponseWriter: w}
		
		// Call the next handler with the wrapped writer
		next.ServeHTTP(gzipWriter, r)
	})
}
`

const DefaultMiddlewareTemplate = `package middleware

import (
	"net/http"
)

// {{.Name}}Middleware implements {{.Name}} middleware
func {{.Name}}Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO: Implement {{.Name}} middleware logic
		
		// Call the next handler
		next.ServeHTTP(w, r)
	})
}
`

// Middleware Usage Instructions

const AuthMiddlewareUsage = `
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
   curl -H "Authorization: Bearer valid-token" http://localhost:8080/api/v1/users`

const RateLimitMiddlewareUsage = `
📌 Configuration:

The middleware is configured with default limits (100 requests/minute).
To customize, edit the middleware call in main.go:

   r.Use(middleware.RateLimitMiddleware(50, time.Minute))  // 50 req/min
   r.Use(middleware.RateLimitMiddleware(1000, time.Hour))  // 1000 req/hour`

const CORSMiddlewareUsage = `
📌 Configuration:

The CORS middleware is configured with permissive defaults.
For production, edit internal/middleware/cors.go:

   - Change allowed origins from "*" to your domain
   - Customize allowed methods and headers
   - Set appropriate max age for preflight requests`

const LoggingMiddlewareUsage = `
📌 Usage:

The logging middleware will now log all HTTP requests.
Log format: [METHOD] /path remote_addr status bytes duration

To customize logging, edit internal/middleware/logging.go`

const RecoveryMiddlewareUsage = `
📌 Usage:

The recovery middleware will catch panics and return 500 errors.
Your server will stay running even if handlers panic.

Check your logs for panic details when they occur.`

const TimeoutMiddlewareUsage = `
📌 Configuration:

Default timeout is set to 30 seconds.
To customize, edit the middleware call in main.go:

   r.Use(middleware.TimeoutMiddleware(60 * time.Second))  // 60 second timeout`

const CompressionMiddlewareUsage = `
📌 Usage:

Responses will now be automatically compressed with gzip.
The middleware only compresses if the client supports it.

Test with: curl -H "Accept-Encoding: gzip" http://localhost:8080/api/v1/users -v`

const DefaultMiddlewareUsage = `
📌 Next steps:
  - Review the generated middleware code
  - Customize the middleware logic as needed
  - Add tests for the middleware`
