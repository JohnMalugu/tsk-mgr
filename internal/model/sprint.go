package model

import (
	"strings"
)

// SprintStatus represents the current lifecycle state of a sprint or iteration
type SprintStatus string

const (
	SprintStatusPlanned   SprintStatus = "planned"
	SprintStatusActive    SprintStatus = "active"
	SprintStatusCompleted SprintStatus = "completed"
	SprintStatusArchived  SprintStatus = "archived"
)

var ValidSprintStatuses = []SprintStatus{
	SprintStatusPlanned,
	SprintStatusActive,
	SprintStatusCompleted,
	SprintStatusArchived,
}

// IsValidSprintStatus checks if the given status string is valid
func IsValidSprintStatus(status string) bool {
	trimmed := strings.ToLower(strings.TrimSpace(status))
	for _, valid := range ValidSprintStatuses {
		if string(valid) == trimmed {
			return true
		}
	}
	return false
}
