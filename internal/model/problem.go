package model

import (
	"errors"
	"time"
)

var (
	ErrProblemTitleRequired    = errors.New("problem title is required")
	ErrProblemNotFound         = errors.New("problem not found")
	ErrInvalidProblemStatus    = errors.New("invalid problem status")
	ErrInvalidProblemCategory  = errors.New("invalid problem category")
	ErrInvalidProblemPriority  = errors.New("invalid problem priority")
	ErrInvalidProblemUrgency   = errors.New("invalid problem urgency")
	ErrInvalidProblemImpact    = errors.New("invalid problem impact")
	ErrInvalidRCAMethod        = errors.New("invalid root cause analysis method")
	ErrInvalidStatusTransition = errors.New("illegal problem status transition")
	ErrKnownErrorRequiresWorkaround = errors.New("known error requires an approved workaround or root cause description")
)

// Problem operational lifecycle statuses
const (
	ProblemStatusLogged          = "logged"
	ProblemStatusInvestigating   = "investigating"
	ProblemStatusWorkaroundFound = "workaround_found"
	ProblemStatusKnownError      = "known_error"
	ProblemStatusResolved        = "resolved"
	ProblemStatusClosed          = "closed"
)

// Problem ITIL categories
const (
	ProblemCategoryInfrastructure = "infrastructure"
	ProblemCategorySoftware       = "software"
	ProblemCategoryNetwork        = "network"
	ProblemCategoryDatabase       = "database"
	ProblemCategorySecurity       = "security"
	ProblemCategoryThirdParty     = "third_party"
	ProblemCategoryProcess        = "process"
)

// Problem priority levels
const (
	ProblemPriorityCritical = "critical"
	ProblemPriorityHigh     = "high"
	ProblemPriorityMedium   = "medium"
	ProblemPriorityLow      = "low"
)

// Problem urgency levels
const (
	ProblemUrgencyCritical = "critical"
	ProblemUrgencyHigh     = "high"
	ProblemUrgencyMedium   = "medium"
	ProblemUrgencyLow      = "low"
)

// Problem impact levels
const (
	ProblemImpactCritical = "critical"
	ProblemImpactHigh     = "high"
	ProblemImpactMedium   = "medium"
	ProblemImpactLow      = "low"
)

// Root Cause Analysis (RCA) methodologies
const (
	RCAMethod5Whys      = "5_whys"
	RCAMethodFishbone   = "fishbone"
	RCAMethodFaultTree  = "fault_tree"
	RCAMethodTimeline   = "timeline"
	RCAMethodOther      = "other"
)

// Problem represents an ITIL Problem investigation record for identifying and resolving root causes
type Problem struct {
	ID                     int64      `json:"id"`
	WorkspaceID            int64      `json:"workspace_id"`
	ProjectID              *int64     `json:"project_id,omitempty"`
	ProblemCode            string     `json:"problem_code"`
	Title                  string     `json:"title"`
	Description            string     `json:"description"`
	Status                 string     `json:"status"`
	Category               string     `json:"category"`
	Priority               string     `json:"priority"`
	Urgency                string     `json:"urgency"`
	Impact                 string     `json:"impact"`
	AssignedTo             *string    `json:"assigned_to,omitempty"`
	AssignedTeam           string     `json:"assigned_team,omitempty"`
	RootCauseCategory      string     `json:"root_cause_category,omitempty"`
	RootCauseSummary       string     `json:"root_cause_summary,omitempty"`
	Workaround             string     `json:"workaround,omitempty"`
	PermanentSolution      string     `json:"permanent_solution,omitempty"`
	IsKnownError           bool       `json:"is_known_error"`
	KnownErrorPublishedAt  *time.Time `json:"known_error_published_at,omitempty"`
	RCAMethod              string     `json:"rca_method,omitempty"`
	CreatedByID            string     `json:"created_by_id"`
	StartedAt              *time.Time `json:"started_at,omitempty"`
	ResolvedAt             *time.Time `json:"resolved_at,omitempty"`
	ClosedAt               *time.Time `json:"closed_at,omitempty"`
	CreatedAt              time.Time  `json:"created_at"`
	UpdatedAt              time.Time  `json:"updated_at"`

	// Enriched audit & relation counts
	ProjectName          string `json:"project_name,omitempty"`
	AssigneeName         string `json:"assignee_name,omitempty"`
	CreatorName          string `json:"creator_name,omitempty"`
	LinkedIncidentsCount int    `json:"linked_incidents_count"`
	LinkedCIsCount       int    `json:"linked_cis_count"`
	LinkedChangesCount   int    `json:"linked_changes_count"`
	LinkedTasksCount     int    `json:"linked_tasks_count"`
	RCAStepsCount        int    `json:"rca_steps_count"`
}
