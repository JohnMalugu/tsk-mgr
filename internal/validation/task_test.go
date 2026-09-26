package validation

import (
	"strings"
	"testing"

	"github.com/JohnMalugu/tsk-mgr-api/internal/model"
)

func TestValidateTaskRejectsOversizedDescription(t *testing.T) {
	task := model.Task{Title: "Task", Description: strings.Repeat("a", 5001)}
	errors := ValidateTask(&task)
	for _, validationError := range errors {
		if validationError.Field == "description" {
			return
		}
	}
	t.Fatal("expected a description length validation error")
}
