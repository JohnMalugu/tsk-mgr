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