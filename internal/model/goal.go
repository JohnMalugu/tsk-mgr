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

// Key result metric type constants
const (
	KeyResultMetricNumerical  = "numerical"
	KeyResultMetricPercentage = "percentage"
	KeyResultMetricCurrency   = "currency"
	KeyResultMetricBoolean    = "boolean"
)

// KeyResult represents a measurable outcome contributing to a Goal
type KeyResult struct {
	ID                 int64     `json:"id"`
	GoalID             int64     `json:"goal_id"`
	Title              string    `json:"title"`
	Description        string    `json:"description,omitempty"`
	MetricType         string    `json:"metric_type"`    // numerical, percentage, currency, boolean
	Unit               string    `json:"unit,omitempty"` // e.g. "%", "$", "users", "ms"
	StartValue         float64   `json:"start_value"`
	TargetValue        float64   `json:"target_value"`
	CurrentValue       float64   `json:"current_value"`
	Weight             float64   `json:"weight"` // default 1.0
	ProgressPercentage float64   `json:"progress_percentage"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

// KeyResultTaskLink represents a bidirectional connection between a key result and a task
type KeyResultTaskLink struct {
	ID          int64     `json:"id"`
	KeyResultID int64     `json:"key_result_id"`
	TaskID      int       `json:"task_id"`
	CreatedAt   time.Time `json:"created_at"`
}

// LinkedTaskInfo provides enriched task summary inside key result views
type LinkedTaskInfo struct {
	TaskID    int       `json:"task_id"`
	Title     string    `json:"title"`
	Status    string    `json:"status"`
	Priority  string    `json:"priority"`
	Completed bool      `json:"completed"`
	LinkedAt  time.Time `json:"linked_at"`
}

// Goal check-in confidence ratings
const (
	GoalConfidenceHigh   = "high"
	GoalConfidenceMedium = "medium"
	GoalConfidenceLow    = "low"
)

// GoalCheckIn records a periodic status update, confidence level, and progress note
type GoalCheckIn struct {
	ID                 int64     `json:"id"`
	GoalID             int64     `json:"goal_id"`
	UserID             string    `json:"user_id"`
	ConfidenceRating   string    `json:"confidence_rating"` // high, medium, low
	Status             string    `json:"status"`            // on_track, behind, at_risk
	ProgressPercentage float64   `json:"progress_percentage"`
	Note               string    `json:"note"`
	CreatedAt          time.Time `json:"created_at"`
}

// CreateGoalRequest represents payload for provisioning a new goal
type CreateGoalRequest struct {
	WorkspaceID  int64      `json:"workspace_id"`
	ParentGoalID *int64     `json:"parent_goal_id,omitempty"`
	Title        string     `json:"title"`
	Description  string     `json:"description,omitempty"`
	Timeframe    string     `json:"timeframe"` // e.g. "Q1 2026"
	StartDate    *time.Time `json:"start_date,omitempty"`
	EndDate      *time.Time `json:"end_date,omitempty"`
}

// UpdateGoalRequest represents payload for modifying goal attributes
type UpdateGoalRequest struct {
	ParentGoalID       *int64     `json:"parent_goal_id,omitempty"`
	OwnerUserID        *string    `json:"owner_user_id,omitempty"`
	Title              *string    `json:"title,omitempty"`
	Description        *string    `json:"description,omitempty"`
	Timeframe          *string    `json:"timeframe,omitempty"`
	StartDate          *time.Time `json:"start_date,omitempty"`
	EndDate            *time.Time `json:"end_date,omitempty"`
	Status             *string    `json:"status,omitempty"`
	ProgressPercentage *float64   `json:"progress_percentage,omitempty"`
}

// CreateKeyResultRequest represents payload for adding a measurable key result to a goal
type CreateKeyResultRequest struct {
	Title        string   `json:"title"`
	Description  string   `json:"description,omitempty"`
	MetricType   string   `json:"metric_type"` // numerical, percentage, currency, boolean
	Unit         string   `json:"unit,omitempty"`
	StartValue   float64  `json:"start_value"`
	TargetValue  float64  `json:"target_value"`
	CurrentValue *float64 `json:"current_value,omitempty"`
	Weight       *float64 `json:"weight,omitempty"`
}

// UpdateKeyResultRequest represents payload for updating key result metrics and values
type UpdateKeyResultRequest struct {
	Title        *string  `json:"title,omitempty"`
	Description  *string  `json:"description,omitempty"`
	MetricType   *string  `json:"metric_type,omitempty"`
	Unit         *string  `json:"unit,omitempty"`
	StartValue   *float64 `json:"start_value,omitempty"`
	TargetValue  *float64 `json:"target_value,omitempty"`
	CurrentValue *float64 `json:"current_value,omitempty"`
	Weight       *float64 `json:"weight,omitempty"`
}

// GoalCheckInRequest represents payload for submitting a progress check-in update
type GoalCheckInRequest struct {
	ConfidenceRating   string   `json:"confidence_rating"` // high, medium, low
	Status             string   `json:"status"`            // on_track, behind, at_risk
	ProgressPercentage *float64 `json:"progress_percentage,omitempty"`
	Note               string   `json:"note"`
}

// LinkTaskToKeyResultRequest represents payload for linking a task to a key result
type LinkTaskToKeyResultRequest struct {
	TaskID int `json:"task_id"`
}

// GoalFilter specifies search and listing criteria for goals
type GoalFilter struct {
	WorkspaceID  *int64 `json:"workspace_id,omitempty"`
	OwnerUserID  string `json:"owner_user_id,omitempty"`
	Status       string `json:"status,omitempty"`
	Timeframe    string `json:"timeframe,omitempty"`
	ParentGoalID *int64 `json:"parent_goal_id,omitempty"`
	Limit        int    `json:"limit,omitempty"`
	Offset       int    `json:"offset,omitempty"`
}

// GoalTree provides hierarchical representation of an objective and its sub-goals
type GoalTree struct {
	Goal       *Goal       `json:"goal"`
	KeyResults []KeyResult `json:"key_results"`
	SubGoals   []*GoalTree `json:"sub_goals,omitempty"`
}

// GoalDetail represents a goal with embedded key results and the latest check-in
type GoalDetail struct {
	Goal          *Goal        `json:"goal"`
	KeyResults    []KeyResult  `json:"key_results"`
	LatestCheckIn *GoalCheckIn `json:"latest_check_in,omitempty"`
}

// GoalSummary summarizes high-level goal progress and status count
type GoalSummary struct {
	ID                 int64   `json:"id"`
	Title              string  `json:"title"`
	Timeframe          string  `json:"timeframe"`
	Status             string  `json:"status"`
	ProgressPercentage float64 `json:"progress_percentage"`
	KeyResultCount     int     `json:"key_result_count"`
	OwnerUserID        string  `json:"owner_user_id"`
}

// GoalMetrics aggregates OKR compliance, status distribution, and average progress
type GoalMetrics struct {
	WorkspaceID        int64          `json:"workspace_id"`
	TotalGoals         int            `json:"total_goals"`
	AchievedGoals      int            `json:"achieved_goals"`
	OnTrackGoals       int            `json:"on_track_goals"`
	BehindGoals        int            `json:"behind_goals"`
	AtRiskGoals        int            `json:"at_risk_goals"`
	DraftGoals         int            `json:"draft_goals"`
	AverageProgress    float64        `json:"average_progress"`
	AchievementRate    float64        `json:"achievement_rate"` // percentage achieved
	StatusDistribution map[string]int `json:"status_distribution"`
}
