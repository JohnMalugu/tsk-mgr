package model

import (
	"errors"
	"strings"
	"time"
)

// Measurement window constants
const (
	MeasurementWindowDaily     = "daily"
	MeasurementWindowWeekly    = "weekly"
	MeasurementWindowMonthly   = "monthly"
	MeasurementWindowQuarterly = "quarterly"
	MeasurementWindowAnnual    = "annual"
)

// Service hours constants
const (
	ServiceHours24x7   = "24x7"
	ServiceHours9x5    = "9x5_business_hours"
	ServiceHoursCustom = "custom"
)

// Target status constants
const (
	TargetStatusDraft      = "draft"
	TargetStatusActive     = "active"
	TargetStatusDeprecated = "deprecated"
)

// Outage type constants
const (
	OutageTypeUnplanned            = "unplanned"
	OutageTypePlannedMaintenance   = "planned_maintenance"
	OutageTypeThirdPartyDependency = "third_party_dependency"
)

// Outage severity constants
const (
	OutageSeverityCritical = "critical"
	OutageSeverityMajor    = "major"
	OutageSeverityMinor    = "minor"
)

// Redundancy mechanism constants
const (
	RedundancyMechanismNone          = "none"
	RedundancyMechanismActiveActive  = "active_active"
	RedundancyMechanismActivePassive = "active_passive"
	RedundancyMechanismNPlusOne      = "n_plus_one"
)

// SPOF risk level constants
const (
	SPOFRiskLow      = "low"
	SPOFRiskMedium   = "medium"
	SPOFRiskHigh     = "high"
	SPOFRiskCritical = "critical"
)

// SPOF status constants
const (
	SPOFStatusOpen         = "open"
	SPOFStatusMitigated    = "mitigated"
	SPOFStatusAcceptedRisk = "accepted_risk"
)

// Availability domain errors for targets
var (
	ErrAvailabilityTargetNotFound     = errors.New("availability target not found")
	ErrAvailabilityTargetNameRequired = errors.New("availability target name is required")
	ErrInvalidTargetPercentage        = errors.New("target percentage must be between 0.0 and 100.0")
	ErrInvalidMeasurementWindow       = errors.New("invalid measurement window")
	ErrInvalidServiceHours            = errors.New("invalid service hours")
	ErrInvalidTargetStatus            = errors.New("invalid target status")
	ErrOutageEventNotFound            = errors.New("outage event not found")
	ErrOutageTitleRequired            = errors.New("outage title is required")
	ErrInvalidOutageType              = errors.New("invalid outage type")
	ErrInvalidOutageSeverity          = errors.New("invalid outage severity")
	ErrInvalidOutageTimestamps        = errors.New("outage ended_at cannot be before started_at")
	ErrAvailabilityMeasurementNotFound = errors.New("availability measurement not found")
	ErrPeriodRequired                 = errors.New("measurement period is required")
	ErrInvalidOperationalMinutes      = errors.New("total operational minutes must be greater than zero")
	ErrSPOFRecordNotFound             = errors.New("single point of failure record not found")
	ErrSPOFComponentNameRequired      = errors.New("SPOF component name is required")
	ErrInvalidRedundancyMechanism     = errors.New("invalid redundancy mechanism")
	ErrInvalidSPOFRiskLevel           = errors.New("invalid SPOF risk level")
	ErrInvalidSPOFStatus              = errors.New("invalid SPOF status")
)

// ValidateMeasurementWindow verifies if a window string is recognized
func ValidateMeasurementWindow(w string) error {
	switch w {
	case MeasurementWindowDaily, MeasurementWindowWeekly, MeasurementWindowMonthly,
		MeasurementWindowQuarterly, MeasurementWindowAnnual:
		return nil
	default:
		return ErrInvalidMeasurementWindow
	}
}

// ValidateServiceHours verifies if service hours are supported
func ValidateServiceHours(sh string) error {
	switch sh {
	case ServiceHours24x7, ServiceHours9x5, ServiceHoursCustom:
		return nil
	default:
		return ErrInvalidServiceHours
	}
}

// ValidateTargetStatus verifies if target status is valid
func ValidateTargetStatus(s string) error {
	switch s {
	case TargetStatusDraft, TargetStatusActive, TargetStatusDeprecated:
		return nil
	default:
		return ErrInvalidTargetStatus
	}
}

// ValidateOutageType verifies if outage type is recognized
func ValidateOutageType(ot string) error {
	switch ot {
	case OutageTypeUnplanned, OutageTypePlannedMaintenance, OutageTypeThirdPartyDependency:
		return nil
	default:
		return ErrInvalidOutageType
	}
}

// ValidateOutageSeverity verifies if outage severity is valid
func ValidateOutageSeverity(sev string) error {
	switch sev {
	case OutageSeverityCritical, OutageSeverityMajor, OutageSeverityMinor:
		return nil
	default:
		return ErrInvalidOutageSeverity
	}
}

// ValidateRedundancyMechanism verifies if redundancy mechanism is supported
func ValidateRedundancyMechanism(rm string) error {
	switch rm {
	case RedundancyMechanismNone, RedundancyMechanismActiveActive,
		RedundancyMechanismActivePassive, RedundancyMechanismNPlusOne:
		return nil
	default:
		return ErrInvalidRedundancyMechanism
	}
}

// ValidateSPOFRiskLevel verifies if SPOF risk level is recognized
func ValidateSPOFRiskLevel(rl string) error {
	switch rl {
	case SPOFRiskLow, SPOFRiskMedium, SPOFRiskHigh, SPOFRiskCritical:
		return nil
	default:
		return ErrInvalidSPOFRiskLevel
	}
}

// ValidateSPOFStatus verifies if SPOF status is valid
func ValidateSPOFStatus(st string) error {
	switch st {
	case SPOFStatusOpen, SPOFStatusMitigated, SPOFStatusAcceptedRisk:
		return nil
	default:
		return ErrInvalidSPOFStatus
	}
}

// AvailabilityTarget defines an agreed SLA uptime commitment, MTBF, and MTTR for a service or CI
type AvailabilityTarget struct {
	ID                 string    `json:"id"`
	WorkspaceID        string    `json:"workspace_id"`
	Name               string    `json:"name"`
	TargetPercentage   float64   `json:"target_percentage"`
	MeasurementWindow  string    `json:"measurement_window"`
	ServiceHours       string    `json:"service_hours"`
	CommittedMTBFHours float64   `json:"committed_mtbf_hours"`
	TargetMTTRMinutes  float64   `json:"target_mttr_minutes"`
	CIID               string    `json:"ci_id,omitempty"`
	Status             string    `json:"status"`
	Description        string    `json:"description,omitempty"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

// Validate validates the AvailabilityTarget entity
func (t *AvailabilityTarget) Validate() error {
	if strings.TrimSpace(t.Name) == "" {
		return ErrAvailabilityTargetNameRequired
	}
	if t.TargetPercentage <= 0 || t.TargetPercentage > 100.0 {
		return ErrInvalidTargetPercentage
	}
	if err := ValidateMeasurementWindow(t.MeasurementWindow); err != nil {
		return err
	}
	if err := ValidateServiceHours(t.ServiceHours); err != nil {
		return err
	}
	if err := ValidateTargetStatus(t.Status); err != nil {
		return err
	}
	return nil
}

// OutageEvent represents an incident, planned maintenance, or third-party outage affecting service availability
type OutageEvent struct {
	ID              string     `json:"id"`
	WorkspaceID     string     `json:"workspace_id"`
	TargetID        string     `json:"target_id,omitempty"`
	CIID            string     `json:"ci_id,omitempty"`
	Title           string     `json:"title"`
	OutageType      string     `json:"outage_type"`
	Severity        string     `json:"severity"`
	StartedAt       time.Time  `json:"started_at"`
	EndedAt         *time.Time `json:"ended_at,omitempty"`
	DurationMinutes float64    `json:"duration_minutes"`
	RootCause       string     `json:"root_cause,omitempty"`
	IsSLAExempt     bool       `json:"is_sla_exempt"`
	ExemptionReason string     `json:"exemption_reason,omitempty"`
	ReportedBy      string     `json:"reported_by,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

// Validate validates the OutageEvent entity
func (o *OutageEvent) Validate() error {
	if strings.TrimSpace(o.Title) == "" {
		return ErrOutageTitleRequired
	}
	if err := ValidateOutageType(o.OutageType); err != nil {
		return err
	}
	if err := ValidateOutageSeverity(o.Severity); err != nil {
		return err
	}
	if o.EndedAt != nil && o.EndedAt.Before(o.StartedAt) {
		return ErrInvalidOutageTimestamps
	}
	return nil
}

// CalculateDuration computes duration in minutes if EndedAt is present
func (o *OutageEvent) CalculateDuration() float64 {
	if o.EndedAt == nil {
		return 0
	}
	dur := o.EndedAt.Sub(o.StartedAt).Minutes()
	if dur < 0 {
		return 0
	}
	return dur
}

// AvailabilityMeasurement tracks periodic uptime calculations, MTBF, and MTTR records
type AvailabilityMeasurement struct {
	ID                     string    `json:"id"`
	WorkspaceID            string    `json:"workspace_id"`
	TargetID               string    `json:"target_id"`
	Period                 string    `json:"period"` // e.g. "2026-M11", "2026-Q4", "2026-W47"
	TotalOperationalMinutes float64   `json:"total_operational_minutes"`
	TotalDowntimeMinutes   float64   `json:"total_downtime_minutes"`
	ExemptDowntimeMinutes  float64   `json:"exempt_downtime_minutes"`
	ActualUptimePercentage float64   `json:"actual_uptime_percentage"`
	MTBFHours              float64   `json:"mtbf_hours"`
	MTTRMinutes            float64   `json:"mttr_minutes"`
	OutageCount            int       `json:"outage_count"`
	IsSLABreached          bool      `json:"is_sla_breached"`
	CalculatedAt           time.Time `json:"calculated_at"`
	CreatedAt              time.Time `json:"created_at"`
	UpdatedAt              time.Time `json:"updated_at"`
}

// Validate validates the AvailabilityMeasurement entity
func (m *AvailabilityMeasurement) Validate() error {
	if strings.TrimSpace(m.TargetID) == "" {
		return ErrAvailabilityTargetNotFound
	}
	if strings.TrimSpace(m.Period) == "" {
		return ErrPeriodRequired
	}
	if m.TotalOperationalMinutes <= 0 {
		return ErrInvalidOperationalMinutes
	}
	return nil
}

// ComputeUptimePercentage computes the actual uptime percentage accounting for SLA exemptions
func (m *AvailabilityMeasurement) ComputeUptimePercentage() float64 {
	if m.TotalOperationalMinutes <= 0 {
		return 100.0
	}
	effectiveDowntime := m.TotalDowntimeMinutes - m.ExemptDowntimeMinutes
	if effectiveDowntime < 0 {
		effectiveDowntime = 0
	}
	if effectiveDowntime > m.TotalOperationalMinutes {
		effectiveDowntime = m.TotalOperationalMinutes
	}
	uptimeMinutes := m.TotalOperationalMinutes - effectiveDowntime
	pct := (uptimeMinutes / m.TotalOperationalMinutes) * 100.0
	if pct < 0 {
		return 0
	}
	if pct > 100.0 {
		return 100.0
	}
	return pct
}

// CheckSLABreach checks whether the actual uptime percentage fell below the target percentage
func (m *AvailabilityMeasurement) CheckSLABreach(targetPercentage float64) bool {
	return m.ActualUptimePercentage < targetPercentage
}

// SPOFRecord represents an identified Single Point of Failure in system architecture
type SPOFRecord struct {
	ID                  string    `json:"id"`
	WorkspaceID         string    `json:"workspace_id"`
	ComponentName       string    `json:"component_name"`
	CIID                string    `json:"ci_id,omitempty"`
	RiskLevel           string    `json:"risk_level"`
	RedundancyMechanism string    `json:"redundancy_mechanism"`
	MitigationPlan      string    `json:"mitigation_plan,omitempty"`
	Status              string    `json:"status"`
	ReviewDate          time.Time `json:"review_date"`
	OwnerID             string    `json:"owner_id,omitempty"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

// Validate validates the SPOFRecord entity
func (s *SPOFRecord) Validate() error {
	if strings.TrimSpace(s.ComponentName) == "" {
		return ErrSPOFComponentNameRequired
	}
	if err := ValidateSPOFRiskLevel(s.RiskLevel); err != nil {
		return err
	}
	if err := ValidateRedundancyMechanism(s.RedundancyMechanism); err != nil {
		return err
	}
	if err := ValidateSPOFStatus(s.Status); err != nil {
		return err
	}
	return nil
}
