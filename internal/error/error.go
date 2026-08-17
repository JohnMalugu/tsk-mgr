package error

// ErrorResponse represents a standard error response
type ErrorResponse struct {
	Status    int       `json:"status"`
	Error     string    `json:"error"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
	Path      string    `json:"path"`
	TraceID   string    `json:"traceId,omitempty"`
}

// AppError represents an application error
type AppError struct {
	Code       int    // HTTP status code
	Message    string // Error message
	InternalErr error  // Original error (for logging)
}
