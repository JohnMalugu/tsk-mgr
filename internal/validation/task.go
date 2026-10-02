package validation

import (
	"strings"
	"time"
	"unicode/utf8"

	"github.com/JohnMalugu/tsk-mgr-api/internal/model"
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

	if utf8.RuneCountInString(task.Title) > 255 {
		errors = append(errors, ValidationError{
			Field:   "title",
			Message: "Title must be less than 255 characters",
		})
	}
	if utf8.RuneCountInString(task.Description) > 5000 {
		errors = append(errors, ValidationError{
			Field:   "description",
			Message: "Description must be 5000 characters or less",
		})
	}
	if task.EstimateMinutes < 0 {
		errors = append(errors, ValidationError{
			Field:   "estimateMinutes",
			Message: "Estimate must be non-negative",
		})
	}
	if task.Recurrence != nil {
		rule := task.Recurrence
		if rule.Frequency != "daily" && rule.Frequency != "weekly" && rule.Frequency != "monthly" {
			errors = append(errors, ValidationError{Field: "recurrence.frequency", Message: "Frequency must be daily, weekly, or monthly"})
		}
		if rule.Interval < 1 || rule.Interval > 365 {
			errors = append(errors, ValidationError{Field: "recurrence.interval", Message: "Interval must be between 1 and 365"})
		}
		if rule.Until != nil && !task.DueDate.IsZero() && rule.Until.Before(task.DueDate) {
			errors = append(errors, ValidationError{Field: "recurrence.until", Message: "Until cannot be before the first due date"})
		}
	}

	if task.Priority != "" {
		switch strings.ToLower(task.Priority) {
		case "low", "medium", "high":
		default:
			errors = append(errors, ValidationError{
				Field:   "priority",
				Message: "Priority must be low, medium, or high",
			})
		}
	}

	for _, tag := range task.Tags {
		if strings.TrimSpace(tag) == "" {
			errors = append(errors, ValidationError{
				Field:   "tags",
				Message: "Tags cannot contain empty values",
			})
		}
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
