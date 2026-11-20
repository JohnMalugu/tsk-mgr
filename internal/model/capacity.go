package model

import (
	"errors"
	"math"
	"strings"
	"time"
)

// Resource type constants
const (
	ResourceTypeComputeCPU        = "compute_cpu"
	ResourceTypeMemoryRAM         = "memory_ram"
	ResourceTypeStorageDisk       = "storage_disk"
	ResourceTypeNetworkBandwidth  = "network_bandwidth"
	ResourceTypeDatabaseIOPS      = "database_iops"
	ResourceTypeLicenseSeats      = "license_seats"
)

// Capacity pool status constants
const (
	PoolStatusOptimal   = "optimal"
	PoolStatusWarning   = "warning"
	PoolStatusCritical  = "critical"
	PoolStatusExhausted = "exhausted"
)

// Unit of measure constants
const (
	UnitCores = "cores"
	UnitGB    = "gb"
	UnitTB    = "tb"
	UnitMbps  = "mbps"
	UnitIOPS  = "iops"
	UnitSeats = "seats"
)

// Capacity domain errors
var (
	ErrCapacityPoolNotFound      = errors.New("capacity pool not found")
	ErrCapacityPoolNameRequired  = errors.New("capacity pool name is required")
	ErrInvalidResourceType       = errors.New("invalid resource type")
	ErrInvalidPoolStatus         = errors.New("invalid pool status")
	ErrInvalidTotalCapacity      = errors.New("total capacity must be greater than zero")
	ErrInvalidThresholdPercent   = errors.New("threshold percent must be between 1 and 100")
	ErrCapacityMetricNotFound    = errors.New("capacity metric not found")
	ErrCapacityPlanNotFound      = errors.New("capacity plan not found")
	ErrInvalidPlanHorizon        = errors.New("invalid capacity plan horizon")
	ErrInvalidPlanStatus         = errors.New("invalid capacity plan status")
	ErrBottleneckNotFound        = errors.New("bottleneck record not found")
	ErrInvalidBottleneckSeverity = errors.New("invalid bottleneck severity")
	ErrInvalidBottleneckStatus   = errors.New("invalid bottleneck status")
)

// ValidateResourceType verifies if a resource type is supported
func ValidateResourceType(rt string) error {
	switch rt {
	case ResourceTypeComputeCPU, ResourceTypeMemoryRAM, ResourceTypeStorageDisk,
		ResourceTypeNetworkBandwidth, ResourceTypeDatabaseIOPS, ResourceTypeLicenseSeats:
		return nil
	default:
		return ErrInvalidResourceType
	}
}

// ValidatePoolStatus verifies if a capacity pool status is valid
func ValidatePoolStatus(status string) error {
	switch status {
	case PoolStatusOptimal, PoolStatusWarning, PoolStatusCritical, PoolStatusExhausted:
		return nil
	default:
		return ErrInvalidPoolStatus
	}
}

// CapacityPool represents an infrastructure or software resource pool managed under ITIL Capacity Management
type CapacityPool struct {
	ID                       string    `json:"id"`
	WorkspaceID              string    `json:"workspace_id"`
	Name                     string    `json:"name"`
	ResourceType             string    `json:"resource_type"`
	CIID                     string    `json:"ci_id,omitempty"`
	TotalCapacity            float64   `json:"total_capacity"`
	AllocatedCapacity        float64   `json:"allocated_capacity"`
	UtilizedCapacity         float64   `json:"utilized_capacity"`
	UnitOfMeasure            string    `json:"unit_of_measure"`
	WarningThresholdPercent  float64   `json:"warning_threshold_percent"`
	CriticalThresholdPercent float64   `json:"critical_threshold_percent"`
	Status                   string    `json:"status"`
	Description              string    `json:"description,omitempty"`
	CreatedAt                time.Time `json:"created_at"`
	UpdatedAt                time.Time `json:"updated_at"`
}

// Validate checks whether the CapacityPool entity contains valid parameters
func (p *CapacityPool) Validate() error {
	if strings.TrimSpace(p.Name) == "" {
		return ErrCapacityPoolNameRequired
	}
	if err := ValidateResourceType(p.ResourceType); err != nil {
		return err
	}
	if p.TotalCapacity <= 0 {
		return ErrInvalidTotalCapacity
	}
	if p.WarningThresholdPercent <= 0 || p.WarningThresholdPercent > 100 {
		return ErrInvalidThresholdPercent
	}
	if p.CriticalThresholdPercent <= 0 || p.CriticalThresholdPercent > 100 {
		return ErrInvalidThresholdPercent
	}
	if p.CriticalThresholdPercent < p.WarningThresholdPercent {
		return errors.New("critical threshold percent must be greater than or equal to warning threshold")
	}
	if p.Status != "" {
		if err := ValidatePoolStatus(p.Status); err != nil {
			return err
		}
	}
	return nil
}

// UtilizationPercent calculates the percentage of total capacity currently utilized
func (p *CapacityPool) UtilizationPercent() float64 {
	if p.TotalCapacity <= 0 {
		return 0.0
	}
	pct := (p.UtilizedCapacity / p.TotalCapacity) * 100.0
	return math.Round(pct*100) / 100
}

// RecalculateStatus determines the pool status based on utilized capacity and thresholds
func (p *CapacityPool) RecalculateStatus() {
	utilization := p.UtilizationPercent()
	if utilization >= 100.0 {
		p.Status = PoolStatusExhausted
	} else if utilization >= p.CriticalThresholdPercent {
		p.Status = PoolStatusCritical
	} else if utilization >= p.WarningThresholdPercent {
		p.Status = PoolStatusWarning
	} else {
		p.Status = PoolStatusOptimal
	}
}

// CapacityMetric represents time-series utilization telemetry sampled for a capacity pool
type CapacityMetric struct {
	ID           string    `json:"id"`
	WorkspaceID  string    `json:"workspace_id"`
	PoolID       string    `json:"pool_id"`
	MetricName   string    `json:"metric_name"`
	SampledAt    time.Time `json:"sampled_at"`
	AverageValue float64   `json:"average_value"`
	PeakValue    float64   `json:"peak_value"`
	P95Value     float64   `json:"p95_value"`
	P99Value     float64   `json:"p99_value"`
	IsAnomaly    bool      `json:"is_anomaly"`
	AnomalyScore float64   `json:"anomaly_score"`
	CreatedAt    time.Time `json:"created_at"`
}

// Validate checks whether the CapacityMetric fields are valid
func (m *CapacityMetric) Validate() error {
	if strings.TrimSpace(m.PoolID) == "" {
		return ErrCapacityPoolNotFound
	}
	if strings.TrimSpace(m.MetricName) == "" {
		return errors.New("metric_name is required")
	}
	if m.SampledAt.IsZero() {
		return errors.New("sampled_at timestamp is required")
	}
	if m.AverageValue < 0 || m.PeakValue < 0 {
		return errors.New("metric values cannot be negative")
	}
	if m.PeakValue < m.AverageValue {
		m.PeakValue = m.AverageValue
	}
	return nil
}

// Recommended scaling action constants
const (
	ActionScaleUp   = "scale_up"
	ActionScaleOut  = "scale_out"
	ActionScaleDown = "scale_down"
	ActionRebalance = "rebalance"
	ActionOptimize  = "optimize"
)

// ValidateRecommendedAction verifies that a scaling recommendation is valid
func ValidateRecommendedAction(action string) error {
	switch action {
	case ActionScaleUp, ActionScaleOut, ActionScaleDown, ActionRebalance, ActionOptimize:
		return nil
	default:
		return errors.New("invalid recommended scaling action")
	}
}

// CapacityForecast models trend analysis, capacity runway, and sizing recommendations
type CapacityForecast struct {
	ID                          string    `json:"id"`
	WorkspaceID                 string    `json:"workspace_id"`
	PoolID                      string    `json:"pool_id"`
	ForecastHorizonDays         int       `json:"forecast_horizon_days"`
	CurrentUtilizationPercent   float64   `json:"current_utilization_percent"`
	ProjectedUtilizationPercent float64   `json:"projected_utilization_percent"`
	GrowthRateMonthlyPercent    float64   `json:"growth_rate_monthly_percent"`
	DaysToExhaustion            int       `json:"days_to_exhaustion"`
	ConfidenceScore             float64   `json:"confidence_score"`
	RecommendedAction           string    `json:"recommended_action"`
	RecommendationNotes         string    `json:"recommendation_notes,omitempty"`
	CreatedAt                   time.Time `json:"created_at"`
	UpdatedAt                   time.Time `json:"updated_at"`
}

// Validate checks whether the CapacityForecast parameters are valid
func (f *CapacityForecast) Validate() error {
	if strings.TrimSpace(f.PoolID) == "" {
		return ErrCapacityPoolNotFound
	}
	if f.ForecastHorizonDays <= 0 {
		return errors.New("forecast_horizon_days must be greater than zero")
	}
	if f.ConfidenceScore < 0.0 || f.ConfidenceScore > 1.0 {
		return errors.New("confidence_score must be between 0.0 and 1.0")
	}
	if f.RecommendedAction != "" {
		if err := ValidateRecommendedAction(f.RecommendedAction); err != nil {
			return err
		}
	}
	return nil
}

// Capacity plan horizon constants
const (
	PlanHorizonShortTerm  = "short_term_30d"
	PlanHorizonMediumTerm = "medium_term_90d"
	PlanHorizonLongTerm   = "long_term_annual"
)

// Capacity plan status constants
const (
	PlanStatusDraft       = "draft"
	PlanStatusInReview    = "in_review"
	PlanStatusApproved    = "approved"
	PlanStatusImplemented = "implemented"
	PlanStatusRejected    = "rejected"
)

// ValidatePlanHorizon checks if a plan planning horizon is recognized
func ValidatePlanHorizon(horizon string) error {
	switch horizon {
	case PlanHorizonShortTerm, PlanHorizonMediumTerm, PlanHorizonLongTerm:
		return nil
	default:
		return ErrInvalidPlanHorizon
	}
}

// ValidatePlanStatus checks if a plan lifecycle status is recognized
func ValidatePlanStatus(status string) error {
	switch status {
	case PlanStatusDraft, PlanStatusInReview, PlanStatusApproved, PlanStatusImplemented, PlanStatusRejected:
		return nil
	default:
		return ErrInvalidPlanStatus
	}
}

// CapacityPlan represents a formal ITIL capacity plan proposing scaling and sizing actions
type CapacityPlan struct {
	ID               string     `json:"id"`
	WorkspaceID      string     `json:"workspace_id"`
	PoolID           string     `json:"pool_id"`
	Title            string     `json:"title"`
	PlanningHorizon  string     `json:"planning_horizon"`
	CurrentCapacity  float64    `json:"current_capacity"`
	RequiredCapacity float64    `json:"required_capacity"`
	EstimatedCost    float64    `json:"estimated_cost"`
	Currency         string     `json:"currency"`
	Justification    string     `json:"justification"`
	Status           string     `json:"status"`
	TargetDate       *time.Time `json:"target_date,omitempty"`
	ApprovedBy       string     `json:"approved_by,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

// Validate checks whether the CapacityPlan fields are valid
func (p *CapacityPlan) Validate() error {
	if strings.TrimSpace(p.PoolID) == "" {
		return ErrCapacityPoolNotFound
	}
	if strings.TrimSpace(p.Title) == "" {
		return errors.New("capacity plan title is required")
	}
	if err := ValidatePlanHorizon(p.PlanningHorizon); err != nil {
		return err
	}
	if p.RequiredCapacity <= 0 {
		return errors.New("required_capacity must be greater than zero")
	}
	if p.EstimatedCost < 0 {
		return errors.New("estimated_cost cannot be negative")
	}
	if p.Status != "" {
		if err := ValidatePlanStatus(p.Status); err != nil {
			return err
		}
	}
	return nil
}

// Bottleneck severity constants
const (
	BottleneckSeverityLow      = "low"
	BottleneckSeverityMedium   = "medium"
	BottleneckSeverityHigh     = "high"
	BottleneckSeverityCritical = "critical"
)

// Bottleneck status constants
const (
	BottleneckStatusOpen          = "open"
	BottleneckStatusInvestigating = "investigating"
	BottleneckStatusMitigated     = "mitigated"
	BottleneckStatusResolved      = "resolved"
)

// ValidateBottleneckSeverity verifies bottleneck severity
func ValidateBottleneckSeverity(sev string) error {
	switch sev {
	case BottleneckSeverityLow, BottleneckSeverityMedium, BottleneckSeverityHigh, BottleneckSeverityCritical:
		return nil
	default:
		return ErrInvalidBottleneckSeverity
	}
}

// ValidateBottleneckStatus verifies bottleneck status
func ValidateBottleneckStatus(status string) error {
	switch status {
	case BottleneckStatusOpen, BottleneckStatusInvestigating, BottleneckStatusMitigated, BottleneckStatusResolved:
		return nil
	default:
		return ErrInvalidBottleneckStatus
	}
}

// BottleneckRecord logs a resource constraint, saturation event, or performance degradation
type BottleneckRecord struct {
	ID                string     `json:"id"`
	WorkspaceID       string     `json:"workspace_id"`
	PoolID            string     `json:"pool_id"`
	AffectedCIID      string     `json:"affected_ci_id,omitempty"`
	Title             string     `json:"title"`
	Description       string     `json:"description"`
	Severity          string     `json:"severity"`
	SaturationPercent float64    `json:"saturation_percent"`
	Status            string     `json:"status"`
	MitigationPlan    string     `json:"mitigation_plan,omitempty"`
	DetectedAt        time.Time  `json:"detected_at"`
	ResolvedAt        *time.Time `json:"resolved_at,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

// Validate checks whether the BottleneckRecord fields are valid
func (b *BottleneckRecord) Validate() error {
	if strings.TrimSpace(b.PoolID) == "" {
		return ErrCapacityPoolNotFound
	}
	if strings.TrimSpace(b.Title) == "" {
		return errors.New("bottleneck title is required")
	}
	if err := ValidateBottleneckSeverity(b.Severity); err != nil {
		return err
	}
	if b.SaturationPercent < 0 {
		return errors.New("saturation_percent cannot be negative")
	}
	if b.Status != "" {
		if err := ValidateBottleneckStatus(b.Status); err != nil {
			return err
		}
	}
	return nil
}

// CreateCapacityPoolRequest defines payload to create a new capacity resource pool
type CreateCapacityPoolRequest struct {
	Name                     string  `json:"name"`
	ResourceType             string  `json:"resource_type"`
	CIID                     string  `json:"ci_id,omitempty"`
	TotalCapacity            float64 `json:"total_capacity"`
	AllocatedCapacity        float64 `json:"allocated_capacity,omitempty"`
	UtilizedCapacity         float64 `json:"utilized_capacity,omitempty"`
	UnitOfMeasure            string  `json:"unit_of_measure"`
	WarningThresholdPercent  float64 `json:"warning_threshold_percent,omitempty"`
	CriticalThresholdPercent float64 `json:"critical_threshold_percent,omitempty"`
	Description              string  `json:"description,omitempty"`
}

// Validate validates CreateCapacityPoolRequest
func (r *CreateCapacityPoolRequest) Validate() error {
	if strings.TrimSpace(r.Name) == "" {
		return ErrCapacityPoolNameRequired
	}
	if err := ValidateResourceType(r.ResourceType); err != nil {
		return err
	}
	if r.TotalCapacity <= 0 {
		return ErrInvalidTotalCapacity
	}
	if r.WarningThresholdPercent != 0.0 && (r.WarningThresholdPercent < 1.0 || r.WarningThresholdPercent > 100.0) {
		return ErrInvalidThresholdPercent
	}
	if r.CriticalThresholdPercent != 0.0 && (r.CriticalThresholdPercent < 1.0 || r.CriticalThresholdPercent > 100.0) {
		return ErrInvalidThresholdPercent
	}
	if r.WarningThresholdPercent > 0 && r.CriticalThresholdPercent > 0 && r.CriticalThresholdPercent < r.WarningThresholdPercent {
		return errors.New("critical threshold percent must be greater than or equal to warning threshold")
	}
	return nil
}

// UpdateCapacityPoolRequest defines payload to modify an existing capacity pool
type UpdateCapacityPoolRequest struct {
	Name                     *string  `json:"name,omitempty"`
	TotalCapacity            *float64 `json:"total_capacity,omitempty"`
	AllocatedCapacity        *float64 `json:"allocated_capacity,omitempty"`
	UtilizedCapacity         *float64 `json:"utilized_capacity,omitempty"`
	WarningThresholdPercent  *float64 `json:"warning_threshold_percent,omitempty"`
	CriticalThresholdPercent *float64 `json:"critical_threshold_percent,omitempty"`
	Description              *string  `json:"description,omitempty"`
}

// Validate validates UpdateCapacityPoolRequest
func (r *UpdateCapacityPoolRequest) Validate() error {
	if r.Name != nil && strings.TrimSpace(*r.Name) == "" {
		return ErrCapacityPoolNameRequired
	}
	if r.TotalCapacity != nil && *r.TotalCapacity <= 0 {
		return ErrInvalidTotalCapacity
	}
	if r.WarningThresholdPercent != nil && (*r.WarningThresholdPercent < 1.0 || *r.WarningThresholdPercent > 100.0) {
		return ErrInvalidThresholdPercent
	}
	if r.CriticalThresholdPercent != nil && (*r.CriticalThresholdPercent < 1.0 || *r.CriticalThresholdPercent > 100.0) {
		return ErrInvalidThresholdPercent
	}
	return nil
}

// RecordCapacityMetricRequest defines payload to ingest a telemetry metric sample
type RecordCapacityMetricRequest struct {
	PoolID       string     `json:"pool_id"`
	MetricName   string     `json:"metric_name"`
	SampledAt    *time.Time `json:"sampled_at,omitempty"`
	AverageValue float64    `json:"average_value"`
	PeakValue    float64    `json:"peak_value"`
	P95Value     float64    `json:"p95_value,omitempty"`
	P99Value     float64    `json:"p99_value,omitempty"`
	IsAnomaly    bool       `json:"is_anomaly,omitempty"`
	AnomalyScore float64    `json:"anomaly_score,omitempty"`
}

// Validate validates RecordCapacityMetricRequest
func (r *RecordCapacityMetricRequest) Validate() error {
	if strings.TrimSpace(r.PoolID) == "" {
		return ErrCapacityPoolNotFound
	}
	if strings.TrimSpace(r.MetricName) == "" {
		return errors.New("metric_name is required")
	}
	if r.AverageValue < 0 || r.PeakValue < 0 {
		return errors.New("metric values cannot be negative")
	}
	return nil
}

// CreateCapacityPlanRequest defines payload to create a capacity expansion plan
type CreateCapacityPlanRequest struct {
	PoolID           string     `json:"pool_id"`
	Title            string     `json:"title"`
	PlanningHorizon  string     `json:"planning_horizon"`
	RequiredCapacity float64    `json:"required_capacity"`
	EstimatedCost    float64    `json:"estimated_cost,omitempty"`
	Currency         string     `json:"currency,omitempty"`
	Justification    string     `json:"justification"`
	TargetDate       *time.Time `json:"target_date,omitempty"`
}

// Validate validates CreateCapacityPlanRequest
func (r *CreateCapacityPlanRequest) Validate() error {
	if strings.TrimSpace(r.PoolID) == "" {
		return ErrCapacityPoolNotFound
	}
	if strings.TrimSpace(r.Title) == "" {
		return errors.New("capacity plan title is required")
	}
	if err := ValidatePlanHorizon(r.PlanningHorizon); err != nil {
		return err
	}
	if r.RequiredCapacity <= 0 {
		return errors.New("required_capacity must be greater than zero")
	}
	if r.EstimatedCost < 0 {
		return errors.New("estimated_cost cannot be negative")
	}
	return nil
}

// UpdateCapacityPlanRequest defines payload to update a capacity plan
type UpdateCapacityPlanRequest struct {
	Title            *string    `json:"title,omitempty"`
	RequiredCapacity *float64   `json:"required_capacity,omitempty"`
	EstimatedCost    *float64   `json:"estimated_cost,omitempty"`
	Justification    *string    `json:"justification,omitempty"`
	Status           *string    `json:"status,omitempty"`
	ApprovedBy       *string    `json:"approved_by,omitempty"`
	TargetDate       *time.Time `json:"target_date,omitempty"`
}

// Validate validates UpdateCapacityPlanRequest
func (r *UpdateCapacityPlanRequest) Validate() error {
	if r.Title != nil && strings.TrimSpace(*r.Title) == "" {
		return errors.New("capacity plan title is required")
	}
	if r.RequiredCapacity != nil && *r.RequiredCapacity <= 0 {
		return errors.New("required_capacity must be greater than zero")
	}
	if r.EstimatedCost != nil && *r.EstimatedCost < 0 {
		return errors.New("estimated_cost cannot be negative")
	}
	if r.Status != nil {
		if err := ValidatePlanStatus(*r.Status); err != nil {
			return err
		}
	}
	return nil
}

// CreateBottleneckRequest defines payload to report a capacity constraint
type CreateBottleneckRequest struct {
	PoolID            string  `json:"pool_id"`
	AffectedCIID      string  `json:"affected_ci_id,omitempty"`
	Title             string  `json:"title"`
	Description       string  `json:"description"`
	Severity          string  `json:"severity"`
	SaturationPercent float64 `json:"saturation_percent"`
	MitigationPlan    string  `json:"mitigation_plan,omitempty"`
}

// Validate validates CreateBottleneckRequest
func (r *CreateBottleneckRequest) Validate() error {
	if strings.TrimSpace(r.PoolID) == "" {
		return ErrCapacityPoolNotFound
	}
	if strings.TrimSpace(r.Title) == "" {
		return errors.New("bottleneck title is required")
	}
	if err := ValidateBottleneckSeverity(r.Severity); err != nil {
		return err
	}
	if r.SaturationPercent < 0 {
		return errors.New("saturation_percent cannot be negative")
	}
	return nil
}

// UpdateBottleneckRequest defines payload to update a bottleneck record
type UpdateBottleneckRequest struct {
	Title             *string  `json:"title,omitempty"`
	Description       *string  `json:"description,omitempty"`
	Severity          *string  `json:"severity,omitempty"`
	SaturationPercent *float64 `json:"saturation_percent,omitempty"`
	Status            *string  `json:"status,omitempty"`
	MitigationPlan    *string  `json:"mitigation_plan,omitempty"`
}

// Validate validates UpdateBottleneckRequest
func (r *UpdateBottleneckRequest) Validate() error {
	if r.Severity != nil {
		if err := ValidateBottleneckSeverity(*r.Severity); err != nil {
			return err
		}
	}
	if r.Status != nil {
		if err := ValidateBottleneckStatus(*r.Status); err != nil {
			return err
		}
	}
	return nil
}
