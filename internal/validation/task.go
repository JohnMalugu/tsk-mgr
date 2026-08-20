package validation

import (
	"fmt"
	"strings"
	"time"

	"github.com/yourusername/task-manager-api/internal/model"
)

// ValidationError holds validation errors
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// ValidateTask validates a task
func ValidateTask(task *model.Task) []ValidationError {
	var errors []ValidationError

	// Validate title
	if strings.TrimSpace(task.Title) == "" {
		errors = append(errors, ValidationError{
			Field:   "title",
			Message: "Title cannot be empty",
		})
	}

	if len(task.Title) > 255 {
		errors = append(errors, ValidationError{
			Field:   "title",
			Message: "Title must be less than 255 characters",
		})
	}
}