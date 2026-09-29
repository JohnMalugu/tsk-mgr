package model

import "time"

type Activity struct {
	ID         int       `json:"id"`
	TaskID     int       `json:"taskId"`
	Action     string    `json:"action"`
	Summary    string    `json:"summary"`
	OccurredAt time.Time `json:"occurredAt"`
}
