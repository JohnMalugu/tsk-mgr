package model

import (
	"strings"
	"time"
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

// SprintDB represents a sprint or iteration entity stored in the database
type SprintDB struct {
	ID              int64        `json:"id"`
	WorkspaceID     int64        `json:"workspace_id"`
	ProjectID       int64        `json:"project_id"`
	Name            string       `json:"name"`
	Goal            string       `json:"goal"`
	Status          SprintStatus `json:"status"`
	StartDate       *time.Time   `json:"start_date,omitempty"`
	EndDate         *time.Time   `json:"end_date,omitempty"`
	TotalPoints     int          `json:"total_points"`
	CompletedPoints int          `json:"completed_points"`
	CreatedAt       time.Time    `json:"created_at"`
	UpdatedAt       time.Time    `json:"updated_at"`
}

