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

// SprintTaskDB represents the association between a sprint and a task with story points
type SprintTaskDB struct {
	ID           int64     `json:"id"`
	SprintID     int64     `json:"sprint_id"`
	TaskID       int64     `json:"task_id"`
	StoryPoints  int       `json:"story_points"`
	AddedAt      time.Time `json:"added_at"`
	TaskTitle    string    `json:"task_title,omitempty"`
	TaskStatus   string    `json:"task_status,omitempty"`
	TaskPriority string    `json:"task_priority,omitempty"`
}

// SprintFilter defines criteria for querying sprints
type SprintFilter struct {
	WorkspaceID int64        `json:"workspace_id"`
	ProjectID   int64        `json:"project_id"`
	Status      SprintStatus `json:"status"`
	Offset      int          `json:"offset"`
	Limit       int          `json:"limit"`
}

// SprintSummary represents a paginated list of sprints
type SprintSummary struct {
	Sprints []SprintDB `json:"sprints"`
	Total   int        `json:"total"`
	Offset  int        `json:"offset"`
	Limit   int        `json:"limit"`
}

// SprintMetrics contains calculated burndown and velocity statistics for a sprint
type SprintMetrics struct {
	SprintID             int64        `json:"sprint_id"`
	SprintName           string       `json:"sprint_name"`
	Status               SprintStatus `json:"status"`
	TotalTasks           int          `json:"total_tasks"`
	CompletedTasks       int          `json:"completed_tasks"`
	InProgressTasks      int          `json:"in_progress_tasks"`
	TodoTasks            int          `json:"todo_tasks"`
	TotalStoryPoints     int          `json:"total_story_points"`
	CompletedStoryPoints int          `json:"completed_story_points"`
	RemainingStoryPoints int          `json:"remaining_story_points"`
	CompletionPercentage float64      `json:"completion_percentage"`
}

// CreateSprintRequest represents payload to create a new sprint
type CreateSprintRequest struct {
	WorkspaceID int64      `json:"workspace_id"`
	ProjectID   int64      `json:"project_id"`
	Name        string     `json:"name"`
	Goal        string     `json:"goal,omitempty"`
	StartDate   *time.Time `json:"start_date,omitempty"`
	EndDate     *time.Time `json:"end_date,omitempty"`
}

// UpdateSprintRequest represents payload to modify sprint metadata
type UpdateSprintRequest struct {
	Name      *string       `json:"name,omitempty"`
	Goal      *string       `json:"goal,omitempty"`
	Status    *SprintStatus `json:"status,omitempty"`
	StartDate *time.Time    `json:"start_date,omitempty"`
	EndDate   *time.Time    `json:"end_date,omitempty"`
}

// AddTaskToSprintRequest represents payload to assign a task to a sprint
type AddTaskToSprintRequest struct {
	TaskID      int64 `json:"task_id"`
	StoryPoints int   `json:"story_points,omitempty"`
}

// UpdateStoryPointsRequest represents payload to adjust story points estimation on a task
type UpdateStoryPointsRequest struct {
	StoryPoints int `json:"story_points"`
}






