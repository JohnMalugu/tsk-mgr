package model

import "time"

type TimeEntry struct {
	ID              int        `json:"id"`
	TaskID          int        `json:"taskId"`
	StartedAt       time.Time  `json:"startedAt"`
	EndedAt         *time.Time `json:"endedAt,omitempty"`
	DurationSeconds int64      `json:"durationSeconds"`
	Note            string     `json:"note,omitempty"`
}
