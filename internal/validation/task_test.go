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

func TestValidateTaskRejectsNegativeEstimate(t *testing.T) {
	task := model.Task{Title: "Task", EstimateMinutes: -1}
	for _, validationError := range ValidateTask(&task) {
		if validationError.Field == "estimateMinutes" {
			return
		}
	}
	t.Fatal("expected estimate validation error")
}
