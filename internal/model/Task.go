package model

import "time"

const (
	TaskStatusTodo       = "todo"
	TaskStatusInProgress = "in_progress"
	TaskStatusCompleted  = "completed"
	TaskStatusCanceled   = "canceled"
)

type ChecklistItem struct {
	ID        int    `json:"id"`
	Text      string `json:"text"`
	Completed bool   `json:"completed"`
}

type TaskComment struct {
	ID        int       `json:"id"`
	TaskID    int       `json:"taskId"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type RecurrenceRule struct {
	Frequency  string     `json:"frequency"`
	Interval   int        `json:"interval"`
	DayOfMonth int        `json:"dayOfMonth,omitempty"`
	Until      *time.Time `json:"until,omitempty"`
}

type Task struct {
	ID                   int             `json:"id"`
	UserID               string          `json:"userId"`
	Title                string          `json:"title"`
	Description          string          `json:"description,omitempty"`
	CreatedAt            time.Time       `json:"createdAt"`
	UpdatedAt            time.Time       `json:"updatedAt"`
	DueDate              time.Time       `json:"dueDate"`
	EstimateMinutes      int             `json:"estimateMinutes,omitempty"`
	Completed            bool            `json:"completed"`
	Status               string          `json:"status"`
	Priority             string          `json:"priority"`
	Tags                 []string        `json:"tags"`
	DependsOn            []int           `json:"dependsOn"`
	Checklist            []ChecklistItem `json:"checklist,omitempty"`
	Recurrence           *RecurrenceRule `json:"recurrence,omitempty"`
	RecurrenceSeriesID   int             `json:"recurrenceSeriesId,omitempty"`
	RecurrenceOccurrence int             `json:"recurrenceOccurrence,omitempty"`
}
