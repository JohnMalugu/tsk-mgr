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

// TimesheetLineRequest payload for adding or updating a timesheet line entry
type TimesheetLineRequest struct {
	TaskID          *int64 `json:"task_id,omitempty"`
	ProjectID       *int64 `json:"project_id,omitempty"`
	WorkDate        string `json:"work_date"` // YYYY-MM-DD
	DurationMinutes int    `json:"duration_minutes"`
	IsBillable      *bool  `json:"is_billable,omitempty"`
	Description     string `json:"description,omitempty"`
}

// SubmitTimesheetRequest payload for submitting a timesheet for manager review
type SubmitTimesheetRequest struct {
	Comments string `json:"comments,omitempty"`
}

// ReviewTimesheetRequest payload for approving or rejecting a submitted timesheet
type ReviewTimesheetRequest struct {
	Action          string `json:"action"` // "approve" or "reject"
	Comments        string `json:"comments,omitempty"`
	RejectionReason string `json:"rejection_reason,omitempty"`
}

// TimesheetFilter provides querying and filtering options across timesheets
type TimesheetFilter struct {
	WorkspaceID int64   `json:"workspace_id"`
	UserID      *int64  `json:"user_id,omitempty"`
	Status      *string `json:"status,omitempty"`
	StartDate   *string `json:"start_date,omitempty"` // YYYY-MM-DD
	EndDate     *string `json:"end_date,omitempty"`   // YYYY-MM-DD
	Limit       int     `json:"limit,omitempty"`
	Offset      int     `json:"offset,omitempty"`
}

// TimesheetMetrics aggregates utilization and compliance statistics
type TimesheetMetrics struct {
	WorkspaceID          int64   `json:"workspace_id"`
	TotalTimesheets      int     `json:"total_timesheets"`
	DraftCount           int     `json:"draft_count"`
	SubmittedCount       int     `json:"submitted_count"`
	ApprovedCount        int     `json:"approved_count"`
	RejectedCount        int     `json:"rejected_count"`
	TotalLoggedMinutes   int     `json:"total_logged_minutes"`
	TotalBillableMinutes int     `json:"total_billable_minutes"`
	TotalLoggedHours     float64 `json:"total_logged_hours"`
	TotalBillableHours   float64 `json:"total_billable_hours"`
	BillablePercentage   float64 `json:"billable_percentage"`
}

