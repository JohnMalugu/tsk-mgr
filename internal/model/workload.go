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

// Time-off type constants
const (
	TimeOffTypeVacation      = "vacation"
	TimeOffTypeSickLeave     = "sick_leave"
	TimeOffTypePublicHoliday = "public_holiday"
	TimeOffTypeTraining      = "training"
	TimeOffTypePersonal      = "personal"
)

// ResourceTimeOff records scheduled unavailability such as vacation, holidays, or leave
type ResourceTimeOff struct {
	ID          int64     `json:"id"`
	WorkspaceID int64     `json:"workspace_id"`
	UserID      string    `json:"user_id"`
	StartDate   time.Time `json:"start_date"`
	EndDate     time.Time `json:"end_date"`
	HoursPerDay float64   `json:"hours_per_day"`
	TimeOffType string    `json:"time_off_type"` // vacation, sick_leave, public_holiday, training, personal
	Description string    `json:"description,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	// Enriched user details
	Username string `json:"username,omitempty"`
}

// Allocation status constants
const (
	AllocationStatusPlanned   = "planned"
	AllocationStatusActive    = "active"
	AllocationStatusCompleted = "completed"
	AllocationStatusCanceled  = "canceled"
)

// ProjectResourceAllocation represents planned commitment of a team member to a project
type ProjectResourceAllocation struct {
	ID                    int64     `json:"id"`
	WorkspaceID           int64     `json:"workspace_id"`
	ProjectID             int       `json:"project_id"`
	UserID                string    `json:"user_id"`
	AllocatedHoursPerWeek float64   `json:"allocated_hours_per_week"`
	StartDate             time.Time `json:"start_date"`
	EndDate               time.Time `json:"end_date"`
	Status                string    `json:"status"` // planned, active, completed, canceled
	Notes                 string    `json:"notes,omitempty"`
	CreatedAt             time.Time `json:"created_at"`
	UpdatedAt             time.Time `json:"updated_at"`

	// Enriched fields
	ProjectName  string `json:"project_name,omitempty"`
	ProjectColor string `json:"project_color,omitempty"`
	Username     string `json:"username,omitempty"`
}

// Workload utilization status constants
const (
	UtilizationStatusUnder    = "underutilized"
	UtilizationStatusOptimal  = "optimal"
	UtilizationStatusCapacity = "at_capacity"
	UtilizationStatusOver     = "overutilized"
)

// WorkloadDailyBucket aggregates daily capacity and demand metrics for a specific calendar day
type WorkloadDailyBucket struct {
	Date                  string  `json:"date"` // YYYY-MM-DD
	AvailableHours        float64 `json:"available_hours"`
	AssignedTaskHours     float64 `json:"assigned_task_hours"`
	AllocatedProjectHours float64 `json:"allocated_project_hours"`
	TimeOffHours          float64 `json:"time_off_hours"`
	TotalDemandHours      float64 `json:"total_demand_hours"`
	UtilizationRate       float64 `json:"utilization_rate"` // percentage e.g. 85.5%
	IsOverCapacity        bool    `json:"is_over_capacity"`
	TaskCount             int     `json:"task_count"`
}

// UserWorkloadSummary summarizes capacity, allocated demand, and workload health for a user over a date range
type UserWorkloadSummary struct {
	UserID                 string                `json:"user_id"`
	Username               string                `json:"username"`
	Email                  string                `json:"email,omitempty"`
	JobTitle               string                `json:"job_title,omitempty"`
	TotalAvailableHours    float64               `json:"total_available_hours"`
	TotalAssignedHours     float64               `json:"total_assigned_hours"`
	TotalAllocatedHours    float64               `json:"total_allocated_hours"`
	TotalTimeOffHours      float64               `json:"total_time_off_hours"`
	OverallUtilizationRate float64               `json:"overall_utilization_rate"`
	UtilizationStatus      string                `json:"utilization_status"` // underutilized, optimal, at_capacity, overutilized
	DailyBreakdown         []WorkloadDailyBucket `json:"daily_breakdown"`
}

// CreateResourceProfileRequest defines parameters to register a team member's capacity profile
type CreateResourceProfileRequest struct {
	WorkspaceID         int64    `json:"workspace_id"`
	UserID              string   `json:"user_id"`
	JobTitle            string   `json:"job_title,omitempty"`
	WeeklyCapacityHours *float64 `json:"weekly_capacity_hours,omitempty"`
	DailyCapacityHours  *float64 `json:"daily_capacity_hours,omitempty"`
	Skills              []string `json:"skills,omitempty"`
	HourlyRate          *float64 `json:"hourly_rate,omitempty"`
	Currency            string   `json:"currency,omitempty"`
	IsActive            *bool    `json:"is_active,omitempty"`
}

// UpdateResourceProfileRequest defines parameters to update a resource capacity profile
type UpdateResourceProfileRequest struct {
	JobTitle            *string   `json:"job_title,omitempty"`
	WeeklyCapacityHours *float64  `json:"weekly_capacity_hours,omitempty"`
	DailyCapacityHours  *float64  `json:"daily_capacity_hours,omitempty"`
	Skills              *[]string `json:"skills,omitempty"`
	HourlyRate          *float64  `json:"hourly_rate,omitempty"`
	Currency            *string   `json:"currency,omitempty"`
	IsActive            *bool     `json:"is_active,omitempty"`
}

// CreateTimeOffRequest defines parameters to schedule a planned absence or holiday
type CreateTimeOffRequest struct {
	WorkspaceID int64     `json:"workspace_id"`
	UserID      string    `json:"user_id"`
	StartDate   time.Time `json:"start_date"`
	EndDate     time.Time `json:"end_date"`
	HoursPerDay *float64  `json:"hours_per_day,omitempty"`
	TimeOffType string    `json:"time_off_type"` // vacation, sick_leave, public_holiday, training, personal
	Description string    `json:"description,omitempty"`
}

// UpdateTimeOffRequest defines parameters to update a scheduled absence
type UpdateTimeOffRequest struct {
	StartDate   *time.Time `json:"start_date,omitempty"`
	EndDate     *time.Time `json:"end_date,omitempty"`
	HoursPerDay *float64   `json:"hours_per_day,omitempty"`
	TimeOffType *string    `json:"time_off_type,omitempty"`
	Description *string    `json:"description,omitempty"`
}

// CreateResourceAllocationRequest defines parameters to book a team member's hours to a project
type CreateResourceAllocationRequest struct {
	WorkspaceID           int64     `json:"workspace_id"`
	ProjectID             int       `json:"project_id"`
	UserID                string    `json:"user_id"`
	AllocatedHoursPerWeek float64   `json:"allocated_hours_per_week"`
	StartDate             time.Time `json:"start_date"`
	EndDate               time.Time `json:"end_date"`
	Notes                 string    `json:"notes,omitempty"`
}

// UpdateResourceAllocationRequest defines parameters to adjust an active or planned project allocation
type UpdateResourceAllocationRequest struct {
	AllocatedHoursPerWeek *float64   `json:"allocated_hours_per_week,omitempty"`
	StartDate             *time.Time `json:"start_date,omitempty"`
	EndDate               *time.Time `json:"end_date,omitempty"`
	Status                *string    `json:"status,omitempty"` // planned, active, completed, canceled
	Notes                 *string    `json:"notes,omitempty"`
}

// WorkloadFilter specifies filtering criteria for team workload and capacity queries
type WorkloadFilter struct {
	WorkspaceID int64      `json:"workspace_id"`
	UserID      string     `json:"user_id,omitempty"`
	Skill       string     `json:"skill,omitempty"`
	From        *time.Time `json:"from,omitempty"`
	To          *time.Time `json:"to,omitempty"`
}

// TeamWorkloadHeatmap aggregates team capacity heatmap matrix across multiple days
type TeamWorkloadHeatmap struct {
	WorkspaceID            int64                 `json:"workspace_id"`
	StartDate              time.Time             `json:"start_date"`
	EndDate                time.Time             `json:"end_date"`
	Days                   []string              `json:"days"` // YYYY-MM-DD
	Users                  []UserWorkloadSummary `json:"users"`
	OverallUtilizationRate float64               `json:"overall_utilization_rate"`
	OverallocatedUserCount int                   `json:"overallocated_user_count"`
}

// OverallocatedAlert highlights a resource exceeding daily or weekly capacity limits
type OverallocatedAlert struct {
	UserID          string  `json:"user_id"`
	Username        string  `json:"username"`
	Date            string  `json:"date"` // YYYY-MM-DD
	CapacityHours   float64 `json:"capacity_hours"`
	DemandHours     float64 `json:"demand_hours"`
	OverHours       float64 `json:"over_hours"`
	UtilizationRate float64 `json:"utilization_rate"`
	ActiveTaskCount int     `json:"active_task_count"`
}

// WorkloadMetrics aggregates high-level capacity, demand, and utilization metrics for a workspace
type WorkloadMetrics struct {
	WorkspaceID                 int64   `json:"workspace_id"`
	TotalProfiles               int     `json:"total_profiles"`
	ActiveProfiles              int     `json:"active_profiles"`
	TotalWeeklyCapacityHours    float64 `json:"total_weekly_capacity_hours"`
	TotalAssignedTaskHours      float64 `json:"total_assigned_task_hours"`
	TotalAllocatedProjectHours  float64 `json:"total_allocated_project_hours"`
	AverageUtilizationRate      float64 `json:"average_utilization_rate"`
	UnderutilizedCount          int     `json:"underutilized_count"`
	OptimalCount                int     `json:"optimal_count"`
	AtCapacityCount             int     `json:"at_capacity_count"`
	OverallocatedCount          int     `json:"overallocated_count"`
	TotalScheduledTimeOffDays   int     `json:"total_scheduled_time_off_days"`
}







