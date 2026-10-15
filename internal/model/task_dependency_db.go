package model

import "time"

// TaskDependencyDB represents a dependency link where TaskID depends on DependsOnTaskID.
type TaskDependencyDB struct {
	ID              int       `json:"id"`
	TaskID          int       `json:"taskId"`
	DependsOnTaskID int       `json:"dependsOnTaskId"`
	CreatedAt       time.Time `json:"createdAt"`
}
