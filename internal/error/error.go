package error

package apierror

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
)

// ErrorResponse represents a standard error response
type ErrorResponse struct {
	Status    int       `json:"status"`
	Error     string    `json:"error"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
	Path      string    `json:"path"`
	TraceID   string    `json:"traceId,omitempty"`
}
