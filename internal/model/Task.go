package model

import "time"

type Task struct {
	ID          int       `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description,omitempty"`
	DueDate     time.Time `json:"dueDate"`
	Completed   bool      `json:"completed"`
	Priority    string    `json:"priority"`
	Tags        []string  `json:"tags"`
}
