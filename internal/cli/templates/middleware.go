package templates

// MiddlewareInfo holds information about supported middleware types
type MiddlewareInfo struct {
	Type        string
	Description string
}

// GetSupportedMiddleware returns all supported middleware types
func GetSupportedMiddleware() []MiddlewareInfo {
	return []MiddlewareInfo{
		{
			Type:        "auth",
			Description: "Authentication middleware",
		},
		{
			Type:        "ratelimit",
			Description: "Rate limiting middleware",
		},
		{
			Type:        "cors",
			Description: "CORS middleware",
		},
		{
			Type:        "logging",
			Description: "Request/response logging middleware",
		},
		{
			Type:        "recovery",
			Description: "Panic recovery middleware",
		},
		{
			Type:        "timeout",
			Description: "Request timeout middleware",
		},
		{
			Type:        "compression",
			Description: "Response compression middleware",
		},
	}
}

// IsSupportedMiddleware checks if a middleware type is supported
func IsSupportedMiddleware(middlewareType string) bool {
	for _, mw := range GetSupportedMiddleware() {
		if mw.Type == middlewareType {
			return true
		}
	}
	return false
}

// GetMiddlewareInfo returns info for a specific middleware type
func GetMiddlewareInfo(middlewareType string) (MiddlewareInfo, bool) {
	for _, mw := range GetSupportedMiddleware() {
		if mw.Type == middlewareType {
			return mw, true
		}
	}
	return MiddlewareInfo{}, false
}

// GetMiddlewareTemplate returns the Go template for a middleware type
func GetMiddlewareTemplate(middlewareType string) string {
	switch middlewareType {
	case "auth":
		return AuthMiddlewareTemplate
	case "ratelimit":
		return RateLimitMiddlewareTemplate
	case "cors":
		return CORSMiddlewareTemplate
	case "logging":
		return LoggingMiddlewareTemplate
	case "recovery":
		return RecoveryMiddlewareTemplate
	case "timeout":
		return TimeoutMiddlewareTemplate
	case "compression":
		return CompressionMiddlewareTemplate
	default:
		return DefaultMiddlewareTemplate
	}
}

// GetMiddlewareUsage returns usage instructions for a middleware type
func GetMiddlewareUsage(middlewareType string) string {
	switch middlewareType {
	case "auth":
		return AuthMiddlewareUsage
	case "ratelimit":
		return RateLimitMiddlewareUsage
	case "cors":
		return CORSMiddlewareUsage
	case "logging":
		return LoggingMiddlewareUsage
	case "recovery":
		return RecoveryMiddlewareUsage
	case "timeout":
		return TimeoutMiddlewareUsage
	case "compression":
		return CompressionMiddlewareUsage
	default:
		return DefaultMiddlewareUsage
	}
}
