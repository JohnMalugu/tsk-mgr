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

	
}