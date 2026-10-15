package model

import "time"

// TaskDependencyDB represents a dependency link where TaskID depends on DependsOnTaskID.
type TaskDependencyDB struct {
	ID              int       `json:"id"`
	TaskID          int       `json:"taskId"`
	DependsOnTaskID int       `json:"dependsOnTaskId"`
	CreatedAt       time.Time `json:"createdAt"`
}

// TaskDependencyInfo provides details about a dependency relationship along with the target task details.
type TaskDependencyInfo struct {
	DependencyID    int   `json:"dependencyId"`
	TaskID          int   `json:"taskId"`
	DependsOnTaskID int   `json:"dependsOnTaskId"`
	Task            *Task `json:"task,omitempty"`
}

// TaskDAGSummary provides an overview of a task's dependency topology and blocked state.
type TaskDAGSummary struct {
	TaskID          int   `json:"taskId"`
	PrerequisiteIDs []int `json:"prerequisiteIds"`
	DependentIDs    []int `json:"dependentIds"`
	IsBlocked       bool  `json:"isBlocked"`
	BlockingTaskIDs []int `json:"blockingTaskIds"`
}
