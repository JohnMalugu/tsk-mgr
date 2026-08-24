package model

import "time"

type Task struct {
	ID        int       `json:"id"`
	Title     string    `json:"title"`
	DueDate   time.Time `json:"dueDate"`
	Completed bool      `json:"completed"`
}