package model

import "time"

// Risk severity tiers based on risk matrix score (Probability * Impact)
const (
	RiskSeverityLow      = "low"      // Score 1-5
	RiskSeverityMedium   = "medium"   // Score 6-11
	RiskSeverityHigh     = "high"     // Score 12-19
	RiskSeverityCritical = "critical" // Score 20-25
)

// Risk lifecycle statuses
const (
	RiskStatusIdentified = "identified"
	RiskStatusAssessing  = "assessing"
	RiskStatusMitigating = "mitigating"
	RiskStatusMonitored  = "monitored"
	RiskStatusAccepted   = "accepted"
	RiskStatusClosed     = "closed"
)

// Risk categories
const (
	RiskCategoryTechnical   = "technical"
	RiskCategorySchedule    = "schedule"
	RiskCategoryFinancial   = "financial"
	RiskCategoryOperational = "operational"
	RiskCategorySecurity    = "security"
	RiskCategoryCompliance  = "compliance"
	RiskCategoryExternal    = "external"
)

// Risk response strategies
const (
	RiskStrategyMitigate = "mitigate"
	RiskStrategyAvoid    = "avoid"
	RiskStrategyTransfer = "transfer"
	RiskStrategyAccept   = "accept"
)

// Risk represents an operational, project, or organizational risk item in the risk register
type Risk struct {
	ID                   int64      `json:"id"`
	WorkspaceID          int64      `json:"workspace_id"`
	ProjectID            *int       `json:"project_id,omitempty"`
	RiskNumber           int        `json:"risk_number"`
	RiskCode             string     `json:"risk_code"`
	Title                string     `json:"title"`
	Description          string     `json:"description"`
	Category             string     `json:"category"`
	Probability          int        `json:"probability"` // 1-5 scale
	Impact               int        `json:"impact"`      // 1-5 scale
	Score                int        `json:"score"`       // probability * impact (1-25)
	Severity             string     `json:"severity"`    // low, medium, high, critical
	Status               string     `json:"status"`      // identified, assessing, mitigating, monitored, accepted, closed
	Strategy             string     `json:"strategy"`    // mitigate, avoid, transfer, accept
	MitigationPlan       string     `json:"mitigation_plan,omitempty"`
	ContingencyPlan      string     `json:"contingency_plan,omitempty"`
	OwnerID              *string    `json:"owner_id,omitempty"`
	TargetResolutionDate *time.Time `json:"target_resolution_date,omitempty"`
	ClosedAt             *time.Time `json:"closed_at,omitempty"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`

	// Enriched fields
	OwnerUsername      string `json:"owner_username,omitempty"`
	ProjectName        string `json:"project_name,omitempty"`
	TaskLinksCount     int    `json:"task_links_count,omitempty"`
	OpenTaskLinksCount int    `json:"open_task_links_count,omitempty"`
}

// Risk task link action types
const (
	RiskActionMitigation    = "mitigation"
	RiskActionContingency   = "contingency"
	RiskActionInvestigation = "investigation"
)

// RiskTaskLink associates a remediation, preventative, or investigation task with a risk
type RiskTaskLink struct {
	ID         int64     `json:"id"`
	RiskID     int64     `json:"risk_id"`
	TaskID     int       `json:"task_id"`
	ActionType string    `json:"action_type"` // mitigation, contingency, investigation
	Notes      string    `json:"notes,omitempty"`
	CreatedAt  time.Time `json:"created_at"`

	// Enriched task fields
	TaskTitle     string `json:"task_title,omitempty"`
	TaskStatus    string `json:"task_status,omitempty"`
	TaskPriority  string `json:"task_priority,omitempty"`
	TaskCompleted bool   `json:"task_completed"`
}

// RiskReview represents a formal periodic evaluation or reassessment of a risk
type RiskReview struct {
	ID                  int64     `json:"id"`
	RiskID              int64     `json:"risk_id"`
	ReviewerID          string    `json:"reviewer_id"`
	PreviousProbability int       `json:"previous_probability"`
	PreviousImpact      int       `json:"previous_impact"`
	PreviousScore       int       `json:"previous_score"`
	NewProbability      int       `json:"new_probability"`
	NewImpact           int       `json:"new_impact"`
	NewScore            int       `json:"new_score"`
	PreviousStatus      string    `json:"previous_status"`
	NewStatus           string    `json:"new_status"`
	Comments            string    `json:"comments"`
	CreatedAt           time.Time `json:"created_at"`

	// Enriched fields
	ReviewerUsername string `json:"reviewer_username,omitempty"`
}

// CalculateRiskSeverity calculates risk score and categorizes it into severity tier
func CalculateRiskSeverity(probability, impact int) (int, string) {
	if probability < 1 {
		probability = 1
	} else if probability > 5 {
		probability = 5
	}
	if impact < 1 {
		impact = 1
	} else if impact > 5 {
		impact = 5
	}

	score := probability * impact
	var severity string
	switch {
	case score >= 20:
		severity = RiskSeverityCritical
	case score >= 12:
		severity = RiskSeverityHigh
	case score >= 6:
		severity = RiskSeverityMedium
	default:
		severity = RiskSeverityLow
	}
	return score, severity
}

// CreateRiskRequest payload for defining a new risk in the register
type CreateRiskRequest struct {
	WorkspaceID          int64      `json:"workspace_id"`
	ProjectID            *int       `json:"project_id,omitempty"`
	Title                string     `json:"title"`
	Description          string     `json:"description"`
	Category             string     `json:"category"`
	Probability          int        `json:"probability"` // 1-5 scale
	Impact               int        `json:"impact"`      // 1-5 scale
	Strategy             string     `json:"strategy"`    // mitigate, avoid, transfer, accept
	MitigationPlan       string     `json:"mitigation_plan,omitempty"`
	ContingencyPlan      string     `json:"contingency_plan,omitempty"`
	OwnerID              *string    `json:"owner_id,omitempty"`
	TargetResolutionDate *time.Time `json:"target_resolution_date,omitempty"`
}

// UpdateRiskRequest payload for updating risk metadata and plans
type UpdateRiskRequest struct {
	Title                *string    `json:"title,omitempty"`
	Description          *string    `json:"description,omitempty"`
	Category             *string    `json:"category,omitempty"`
	Strategy             *string    `json:"strategy,omitempty"`
	MitigationPlan       *string    `json:"mitigation_plan,omitempty"`
	ContingencyPlan      *string    `json:"contingency_plan,omitempty"`
	OwnerID              *string    `json:"owner_id,omitempty"`
	TargetResolutionDate *time.Time `json:"target_resolution_date,omitempty"`
}

// AssessRiskRequest payload for reassessing probability and impact
type AssessRiskRequest struct {
	Probability int    `json:"probability"` // 1-5 scale
	Impact      int    `json:"impact"`      // 1-5 scale
	Comments    string `json:"comments"`
}

// ChangeRiskStatusRequest payload for updating risk lifecycle status
type ChangeRiskStatusRequest struct {
	Status   string `json:"status"` // identified, assessing, mitigating, monitored, accepted, closed
	Comments string `json:"comments,omitempty"`
}

// LinkRiskTaskRequest payload for associating a task with a risk
type LinkRiskTaskRequest struct {
	TaskID     int    `json:"task_id"`
	ActionType string `json:"action_type"` // mitigation, contingency, investigation
	Notes      string `json:"notes,omitempty"`
}

// CreateRiskReviewRequest payload for submitting a periodic risk review
type CreateRiskReviewRequest struct {
	NewProbability *int   `json:"new_probability,omitempty"` // optional updated probability
	NewImpact      *int   `json:"new_impact,omitempty"`      // optional updated impact
	NewStatus      string `json:"new_status,omitempty"`      // optional updated status
	Comments       string `json:"comments"`
}

// RiskFilter specifies query parameters for filtering and paginating risks
type RiskFilter struct {
	WorkspaceID int64
	ProjectID   *int
	Category    string
	Severity    string
	Status      string
	Strategy    string
	OwnerID     string
	Search      string
	MinScore    int
	MaxScore    int
	Limit       int
	Offset      int
}

// RiskMetrics summarizes workspace-level risk exposure, heatmaps, and distribution
type RiskMetrics struct {
	WorkspaceID        int64          `json:"workspace_id"`
	TotalRisks         int            `json:"total_risks"`
	ActiveRisks        int            `json:"active_risks"`
	ClosedRisks        int            `json:"closed_risks"`
	TotalExposureScore int            `json:"total_exposure_score"`
	AverageRiskScore   float64        `json:"average_risk_score"`
	SeverityBreakdown  map[string]int `json:"severity_breakdown"`
	StatusBreakdown    map[string]int `json:"status_breakdown"`
	CategoryBreakdown  map[string]int `json:"category_breakdown"`
	StrategyBreakdown  map[string]int `json:"strategy_breakdown"`
	HeatmapMatrix      [5][5]int      `json:"heatmap_matrix"` // rows: probability 1..5, cols: impact 1..5
	CriticalRisks      []Risk         `json:"critical_risks"`
}
