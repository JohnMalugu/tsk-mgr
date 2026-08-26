package middleware

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"runtime/debug"
)

// RecoveryMiddleware recovers from panics and returns a 500 error
type RecoveryMiddleware struct {
	next http.Handler
}

// NewRecoveryMiddleware creates a new recovery middleware
func NewRecoveryMiddleware(next http.Handler) *RecoveryMiddleware {
	return &RecoveryMiddleware{next: next}
}

// ServeHTTP implements the http.Handler interface
func (m *RecoveryMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Use defer with recover to catch panics
	defer func() {
		if err := recover(); err != nil {
			// Log the panic with stack trace
			log.Printf("[PANIC] %v\n", err)
			log.Printf("[STACK TRACE]\n%s", debug.Stack())

			// Send error response to client
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"status":  500,
				"error":   "Internal Server Error",
				"message": "An unexpected error occurred",
			})
		}
	}()

	// Call the next handler
	m.next.ServeHTTP(w, r)
}
