package validation

import (
	"strings"
	"testing"
	"time"

	"github.com/JohnMalugu/tsk-mgr-api/internal/model"
)

func TestValidateTaskCountsTitleCharacters(t *testing.T) {
	task := &model.Task{Title: strings.Repeat("é", 255), DueDate: time.Now()}
	if errors := ValidateTask(task); len(errors) != 0 {
		t.Fatalf("expected 255 Unicode characters to be valid, got %#v", errors)
	}

	task.Title += "é"
	if errors := ValidateTask(task); len(errors) == 0 || errors[0].Field != "title" {
		t.Fatalf("expected title length validation error, got %#v", errors)
	}
}
