package model

import "time"

// Workload capacity defaults and limits
const (
	DefaultWeeklyCapacityHours = 40.0
	DefaultDailyCapacityHours  = 8.0
	MaxDailyCapacityHours      = 24.0
	MaxWeeklyCapacityHours     = 168.0
)

// ResourceProfile defines a team member's working capacity, role, and skills within a workspace
type ResourceProfile struct {
	ID                  int64     `json:"id"`
	WorkspaceID         int64     `json:"workspace_id"`
	UserID              string    `json:"user_id"`
	JobTitle            string    `json:"job_title,omitempty"`
	WeeklyCapacityHours float64   `json:"weekly_capacity_hours"`
	DailyCapacityHours  float64   `json:"daily_capacity_hours"`
	Skills              []string  `json:"skills,omitempty"`
	HourlyRate          float64   `json:"hourly_rate"`
	Currency            string    `json:"currency"` // e.g. "USD", "EUR"
	IsActive            bool      `json:"is_active"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`

	// Enriched user details
	Username string `json:"username,omitempty"`
	Email    string `json:"email,omitempty"`
}
