package model

import "time"

type Subtask struct {
	ID        int       `json:"id"`
	TaskID    int       `json:"task_id"`
	Title     string    `json:"title"`
	Completed bool      `json:"completed"`
	Position  int       `json:"position"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
