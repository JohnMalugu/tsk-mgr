package model

import "time"

// ActivityDB represents an immutable audit log record stored in SQLite.
type ActivityDB struct {
	ID          int       `json:"id"`
	UserID      string    `json:"userId"`
	TaskID      *int      `json:"taskId,omitempty"`
	Action      string    `json:"action"`
	Summary     string    `json:"summary"`
	DetailsJSON string    `json:"detailsJson,omitempty"`
	CreatedAt   time.Time `json:"createdAt"`
}
