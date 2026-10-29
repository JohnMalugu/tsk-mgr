package model

import "time"

// Goal status constants
const (
	GoalStatusDraft    = "draft"
	GoalStatusActive   = "active"
	GoalStatusOnTrack  = "on_track"
	GoalStatusBehind   = "behind"
	GoalStatusAtRisk   = "at_risk"
	GoalStatusAchieved = "achieved"
	GoalStatusCanceled = "canceled"
)

// Goal represents a high-level strategic objective or OKR
type Goal struct {
	ID                 int64      `json:"id"`
	WorkspaceID        int64      `json:"workspace_id"`
	OwnerUserID        string     `json:"owner_user_id"`
	ParentGoalID       *int64     `json:"parent_goal_id,omitempty"`
	Title              string     `json:"title"`
	Description        string     `json:"description,omitempty"`
	Timeframe          string     `json:"timeframe"` // e.g. "Q1 2026", "2026 Annual"
	StartDate          *time.Time `json:"start_date,omitempty"`
	EndDate            *time.Time `json:"end_date,omitempty"`
	Status             string     `json:"status"`              // draft, active, on_track, behind, at_risk, achieved, canceled
	ProgressPercentage float64    `json:"progress_percentage"` // 0.0 - 100.0
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
}
