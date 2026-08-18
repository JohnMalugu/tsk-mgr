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

// NewAppError creates a new AppError
func NewAppError(code int, message string, internalErr error) *AppError {
	return &AppError{
		Code:        code,
		Message:     message,
		InternalErr: internalErr,
	}
}


// RespondWithError sends an error response to the client
func RespondWithError(w http.ResponseWriter, r *http.Request, appErr *AppError) {
	// Log the error (for debugging)
	if appErr.InternalErr != nil {
		log.Printf("[ERROR] %s %s - %d: %v", r.Method, r.URL.Path, appErr.Code, appErr.InternalErr)
	} else {
		log.Printf("[ERROR] %s %s - %d: %s", r.Method, r.URL.Path, appErr.Code, appErr.Message)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(appErr.Code)

	errorResponse := ErrorResponse{
		Status:    appErr.Code,
		Error:     http.StatusText(appErr.Code),
		Message:   appErr.Message,
		Timestamp: time.Now().UTC(),
		Path:      r.URL.Path,
	}

	json.NewEncoder(w).Encode(errorResponse)
}

