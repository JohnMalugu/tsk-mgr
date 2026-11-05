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
	OwnerUsername       string `json:"owner_username,omitempty"`
	ProjectName         string `json:"project_name,omitempty"`
	TaskLinksCount      int    `json:"task_links_count,omitempty"`
	OpenTaskLinksCount  int    `json:"open_task_links_count,omitempty"`
}
