package model

import "time"

type Tag struct {
	ID        int       `json:"id"`
	UserID    string    `json:"user_id"`
	Name      string    `json:"name"`
	Color     string    `json:"color"`
	CreatedAt time.Time `json:"created_at"`
}

type TaskTag struct {
	TaskID    int       `json:"task_id"`
	TagID     int       `json:"tag_id"`
	CreatedAt time.Time `json:"created_at"`
}
