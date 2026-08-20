package validation

import (
	"fmt"
	"github.com/yourusername/task-manager-api/internal/model"
	"strings"
	"time"
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

	// Validate due date
	if task.DueDate.IsZero() {
		errors = append(errors, ValidationError{
			Field:   "dueDate",
			Message: "Due date cannot be empty",
		})
	}

	// Due date should be in the future (optional, but good practice)
	if !task.DueDate.IsZero() && task.DueDate.Before(time.Now()) {
		// Actually, tasks CAN be due in the past (if overdue)
		// So we just warn, don't error
	}
	return errors
}
