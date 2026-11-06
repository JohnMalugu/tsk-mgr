package model

import "time"

// Timesheet lifecycle statuses
const (
	TimesheetStatusDraft     = "draft"
	TimesheetStatusSubmitted = "submitted"
	TimesheetStatusApproved  = "approved"
	TimesheetStatusRejected  = "rejected"
)

// Timesheet represents a periodic worklog summary submitted for managerial approval
type Timesheet struct {
	ID              int64      `json:"id"`
	WorkspaceID     int64      `json:"workspace_id"`
	UserID          string     `json:"user_id"`
	PeriodStartDate string     `json:"period_start_date"` // YYYY-MM-DD
	PeriodEndDate   string     `json:"period_end_date"`   // YYYY-MM-DD
	Status          string     `json:"status"`            // draft, submitted, approved, rejected
	TotalMinutes    int        `json:"total_minutes"`
	BillableMinutes int        `json:"billable_minutes"`
	SubmittedAt     *time.Time `json:"submitted_at,omitempty"`
	ReviewedAt      *time.Time `json:"reviewed_at,omitempty"`
	ReviewerID      *string    `json:"reviewer_id,omitempty"`
	RejectionReason string     `json:"rejection_reason,omitempty"`
	Notes           string     `json:"notes,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`

	// Enriched fields
	Username         string `json:"username,omitempty"`
	ReviewerUsername string `json:"reviewer_username,omitempty"`
	LinesCount       int    `json:"lines_count,omitempty"`
}

// Timesheet approval actions
const (
	TimesheetActionSubmit  = "submit"
	TimesheetActionApprove = "approve"
	TimesheetActionReject  = "reject"
	TimesheetActionReopen  = "reopen"
)

// TimesheetLine represents an individual line item of logged work within a timesheet
type TimesheetLine struct {
	ID              int64     `json:"id"`
	TimesheetID     int64     `json:"timesheet_id"`
	TaskID          *int      `json:"task_id,omitempty"`
	ProjectID       *int      `json:"project_id,omitempty"`
	WorkDate        string    `json:"work_date"` // YYYY-MM-DD
	DurationMinutes int       `json:"duration_minutes"`
	IsBillable      bool      `json:"is_billable"`
	Description     string    `json:"description"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`

	// Enriched fields
	TaskTitle   string `json:"task_title,omitempty"`
	ProjectName string `json:"project_name,omitempty"`
}

// TimesheetApproval represents an audit record of submission, approval, rejection, or reopening
type TimesheetApproval struct {
	ID             int64     `json:"id"`
	TimesheetID    int64     `json:"timesheet_id"`
	ActorID        string    `json:"actor_id"`
	Action         string    `json:"action"` // submit, approve, reject, reopen
	PreviousStatus string    `json:"previous_status"`
	NewStatus      string    `json:"new_status"`
	Comments       string    `json:"comments,omitempty"`
	CreatedAt      time.Time `json:"created_at"`

	// Enriched fields
	ActorUsername string `json:"actor_username,omitempty"`
}

// CreateTimesheetRequest payload for initializing a new timesheet
type CreateTimesheetRequest struct {
	WorkspaceID     int64  `json:"workspace_id"`
	PeriodStartDate string `json:"period_start_date"` // YYYY-MM-DD
	PeriodEndDate   string `json:"period_end_date"`   // YYYY-MM-DD
	Notes           string `json:"notes,omitempty"`
	AutoPopulate    bool   `json:"auto_populate,omitempty"` // pull existing time entries into lines
}

// UpdateTimesheetRequest payload for updating editable draft timesheet notes
type UpdateTimesheetRequest struct {
	Notes *string `json:"notes,omitempty"`
}

// PopulateTimesheetRequest payload for importing recorded time_entries into a timesheet
type PopulateTimesheetRequest struct {
	IncludeExisting bool `json:"include_existing,omitempty"`
}

