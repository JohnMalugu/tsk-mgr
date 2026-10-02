package validation

import (
	"strings"
	"testing"
	"time"

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

func TestValidateTaskRejectsUnknownStatus(t *testing.T) {
	task := model.Task{Title: "Task", Status: "blocked"}
	for _, validationError := range ValidateTask(&task) {
		if validationError.Field == "status" {
			return
		}
	}
	t.Fatal("expected a status validation error")
}

func TestValidateTaskRejectsInvalidRecurrence(t *testing.T) {
	task := model.Task{Title: "Task", DueDate: time.Date(2030, 1, 2, 9, 0, 0, 0, time.UTC), Recurrence: &model.RecurrenceRule{Frequency: "yearly", Interval: 0}}
	fields := make(map[string]bool)
	for _, validationError := range ValidateTask(&task) {
		fields[validationError.Field] = true
	}
	if !fields["recurrence.frequency"] || !fields["recurrence.interval"] {
		t.Fatalf("expected recurrence-specific errors, got %#v", fields)
	}
}
