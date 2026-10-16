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

// ActivityFilter specifies criteria for querying activity feeds.
type ActivityFilter struct {
	Action string     `json:"action,omitempty"`
	TaskID *int       `json:"taskId,omitempty"`
	From   *time.Time `json:"from,omitempty"`
	To     *time.Time `json:"to,omitempty"`
	Offset int        `json:"offset"`
	Limit  int        `json:"limit"`
}

// PaginatedActivities wraps activity items with pagination metadata.
type PaginatedActivities struct {
	Items  []*ActivityDB `json:"items"`
	Total  int           `json:"total"`
	Offset int           `json:"offset"`
	Limit  int           `json:"limit"`
}

// ActionCount holds the occurrence count for a specific activity action.
type ActionCount struct {
	Action string `json:"action"`
	Count  int    `json:"count"`
}

// ActivitySummary aggregates activity metrics for a user over a time window.
type ActivitySummary struct {
	TotalEvents     int           `json:"totalEvents"`
	ActionBreakdown []ActionCount `json:"actionBreakdown"`
	RecentEvents    []*ActivityDB `json:"recentEvents"`
}
