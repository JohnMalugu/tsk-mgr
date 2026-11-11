package model

import (
	"errors"
	"strings"
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

// ValidateProblemStatus validates whether the problem status is recognized
func ValidateProblemStatus(status string) error {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case ProblemStatusLogged, ProblemStatusInvestigating, ProblemStatusWorkaroundFound,
		ProblemStatusKnownError, ProblemStatusResolved, ProblemStatusClosed:
		return nil
	default:
		return ErrInvalidProblemStatus
	}
}

// ValidateProblemCategory validates whether the problem category is recognized
func ValidateProblemCategory(cat string) error {
	switch strings.ToLower(strings.TrimSpace(cat)) {
	case ProblemCategoryInfrastructure, ProblemCategorySoftware, ProblemCategoryNetwork,
		ProblemCategoryDatabase, ProblemCategorySecurity, ProblemCategoryThirdParty, ProblemCategoryProcess:
		return nil
	default:
		return ErrInvalidProblemCategory
	}
}

// ValidateProblemPriority validates whether the priority level is recognized
func ValidateProblemPriority(p string) error {
	switch strings.ToLower(strings.TrimSpace(p)) {
	case ProblemPriorityCritical, ProblemPriorityHigh, ProblemPriorityMedium, ProblemPriorityLow:
		return nil
	default:
		return ErrInvalidProblemPriority
	}
}

// ValidateProblemUrgency validates whether the urgency level is recognized
func ValidateProblemUrgency(u string) error {
	switch strings.ToLower(strings.TrimSpace(u)) {
	case ProblemUrgencyCritical, ProblemUrgencyHigh, ProblemUrgencyMedium, ProblemUrgencyLow:
		return nil
	default:
		return ErrInvalidProblemUrgency
	}
}

// ValidateProblemImpact validates whether the impact level is recognized
func ValidateProblemImpact(i string) error {
	switch strings.ToLower(strings.TrimSpace(i)) {
	case ProblemImpactCritical, ProblemImpactHigh, ProblemImpactMedium, ProblemImpactLow:
		return nil
	default:
		return ErrInvalidProblemImpact
	}
}

// ValidateRCAMethod validates whether the RCA methodology is recognized
func ValidateRCAMethod(m string) error {
	switch strings.ToLower(strings.TrimSpace(m)) {
	case RCAMethod5Whys, RCAMethodFishbone, RCAMethodFaultTree, RCAMethodTimeline, RCAMethodOther, "":
		return nil
	default:
		return ErrInvalidRCAMethod
	}
}

// ValidateProblemTransition validates whether transitioning from current to target status is permitted by ITIL workflow
func ValidateProblemTransition(current, target string) error {
	current = strings.ToLower(strings.TrimSpace(current))
	target = strings.ToLower(strings.TrimSpace(target))

	if current == target {
		return nil
	}

	if err := ValidateProblemStatus(target); err != nil {
		return err
	}

	switch current {
	case ProblemStatusLogged:
		// From logged can go to investigating or directly closed if duplicate/not a problem
		if target == ProblemStatusInvestigating || target == ProblemStatusClosed {
			return nil
		}
	case ProblemStatusInvestigating:
		// From investigating can identify workaround, publish known error, resolve, or close
		if target == ProblemStatusWorkaroundFound || target == ProblemStatusKnownError ||
			target == ProblemStatusResolved || target == ProblemStatusClosed {
			return nil
		}
	case ProblemStatusWorkaroundFound:
		// Can publish to KEDB, resolve with permanent fix, continue investigating, or close
		if target == ProblemStatusKnownError || target == ProblemStatusResolved ||
			target == ProblemStatusInvestigating || target == ProblemStatusClosed {
			return nil
		}
	case ProblemStatusKnownError:
		// Known error can be resolved when permanent fix is ready, updated in investigation, or closed
		if target == ProblemStatusResolved || target == ProblemStatusInvestigating ||
			target == ProblemStatusWorkaroundFound || target == ProblemStatusClosed {
			return nil
		}
	case ProblemStatusResolved:
		// Resolved can be verified and closed, or re-opened into investigating if recurrence happens
		if target == ProblemStatusClosed || target == ProblemStatusInvestigating {
			return nil
		}
	case ProblemStatusClosed:
		// Closed is terminal; re-opening requires moving back to investigating
		if target == ProblemStatusInvestigating {
			return nil
		}
	}

	return ErrInvalidStatusTransition
}

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

// Problem CI relationship roles
const (
	ProblemCIRelRootCause  = "root_cause_asset"
	ProblemCIRelImpacted   = "impacted_asset"
	ProblemCIRelWorkaround = "workaround_applied"
	ProblemCIRelMonitoring = "monitoring_target"
)

// ValidateProblemCIRelation validates whether the problem-CI relationship role is valid
func ValidateProblemCIRelation(rel string) error {
	switch strings.ToLower(strings.TrimSpace(rel)) {
	case ProblemCIRelRootCause, ProblemCIRelImpacted, ProblemCIRelWorkaround, ProblemCIRelMonitoring, "":
		return nil
	default:
		return errors.New("invalid problem configuration item relationship")
	}
}

// ProblemIncidentLink represents an association between an underlying Problem and an Incident caused by it
type ProblemIncidentLink struct {
	ID         int64     `json:"id"`
	ProblemID  int64     `json:"problem_id"`
	IncidentID int64     `json:"incident_id"`
	Notes      string    `json:"notes,omitempty"`
	LinkedAt   time.Time `json:"linked_at"`

	// Enriched incident fields
	IncidentNumber   int    `json:"incident_number,omitempty"`
	IncidentTitle    string `json:"incident_title,omitempty"`
	IncidentSeverity string `json:"incident_severity,omitempty"`
	IncidentStatus   string `json:"incident_status,omitempty"`
}

// ProblemCILink represents an association between a Problem and an affected or root-cause Configuration Item
type ProblemCILink struct {
	ID           int64     `json:"id"`
	ProblemID    int64     `json:"problem_id"`
	CIID         int64     `json:"ci_id"`
	Relationship string    `json:"relationship"` // root_cause_asset, impacted_asset, workaround_applied, monitoring_target
	Notes        string    `json:"notes,omitempty"`
	LinkedAt     time.Time `json:"linked_at"`

	// Enriched CI fields
	CICode       string `json:"ci_code,omitempty"`
	CIName       string `json:"ci_name,omitempty"`
	CIType       string `json:"ci_type,omitempty"`
	Criticality  string `json:"criticality,omitempty"`
	CIStatus     string `json:"ci_status,omitempty"`
}
