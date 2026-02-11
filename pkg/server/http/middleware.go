package http

import (
	"net/http"
	"time"

	"github.com/futuretea/localrecall-mcp-server/pkg/core/logging"
)

// RequestMiddleware wraps the handler with logging middleware
func RequestMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Log the request
		logging.Debug("Request: %s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)

		// Call the next handler
		next.ServeHTTP(w, r)

		// Log the completion
		duration := time.Since(start)
		logging.Debug("Request completed in %v", duration)
	})
}
