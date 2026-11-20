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

// Availability test type constants
const (
	TestTypeFailover              = "failover"
	TestTypeLoadBalancerIsolation = "load_balancer_isolation"
	TestTypePowerSimulation       = "power_simulation"
	TestTypeChaosDrill            = "chaos_drill"
)

// Availability test frequency constants
const (
	TestFrequencyWeekly     = "weekly"
	TestFrequencyMonthly    = "monthly"
	TestFrequencyQuarterly  = "quarterly"
	TestFrequencySemiAnnual = "semi_annual"
	TestFrequencyAnnual     = "annual"
)

// Availability test outcome constants
const (
	TestOutcomePassed  = "passed"
	TestOutcomeFailed  = "failed"
	TestOutcomePartial = "partial"
	TestOutcomePending = "pending"
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
	ErrAvailabilityTestNotFound       = errors.New("availability test not found")
	ErrAvailabilityTestNameRequired   = errors.New("availability test name is required")
	ErrInvalidTestType                = errors.New("invalid availability test type")
	ErrInvalidTestFrequency           = errors.New("invalid availability test frequency")
	ErrInvalidTestOutcome             = errors.New("invalid availability test outcome")
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

// ValidateTestType verifies if availability test type is recognized
func ValidateTestType(tt string) error {
	switch tt {
	case TestTypeFailover, TestTypeLoadBalancerIsolation, TestTypePowerSimulation, TestTypeChaosDrill:
		return nil
	default:
		return ErrInvalidTestType
	}
}

// ValidateTestFrequency verifies if availability test frequency is supported
func ValidateTestFrequency(tf string) error {
	switch tf {
	case TestFrequencyWeekly, TestFrequencyMonthly, TestFrequencyQuarterly,
		TestFrequencySemiAnnual, TestFrequencyAnnual:
		return nil
	default:
		return ErrInvalidTestFrequency
	}
}

// ValidateTestOutcome verifies if availability test outcome is valid
func ValidateTestOutcome(to string) error {
	switch to {
	case TestOutcomePassed, TestOutcomeFailed, TestOutcomePartial, TestOutcomePending:
		return nil
	default:
		return ErrInvalidTestOutcome
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

// AvailabilityTest records planned and historical failover, load balancing, and disaster drills
type AvailabilityTest struct {
	ID          string     `json:"id"`
	WorkspaceID string     `json:"workspace_id"`
	TargetID    string     `json:"target_id,omitempty"`
	Name        string     `json:"name"`
	TestType    string     `json:"test_type"`
	Frequency   string     `json:"frequency"`
	LastTestedAt *time.Time `json:"last_tested_at,omitempty"`
	NextTestDue time.Time  `json:"next_test_due"`
	Outcome     string     `json:"outcome"`
	Findings    string     `json:"findings,omitempty"`
	ConductedBy string     `json:"conducted_by,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// Validate validates the AvailabilityTest entity
func (t *AvailabilityTest) Validate() error {
	if strings.TrimSpace(t.Name) == "" {
		return ErrAvailabilityTestNameRequired
	}
	if err := ValidateTestType(t.TestType); err != nil {
		return err
	}
	if err := ValidateTestFrequency(t.Frequency); err != nil {
		return err
	}
	if err := ValidateTestOutcome(t.Outcome); err != nil {
		return err
	}
	return nil
}

// CreateAvailabilityTargetRequest contains parameters for registering a new availability SLA target
type CreateAvailabilityTargetRequest struct {
	Name               string  `json:"name"`
	TargetPercentage   float64 `json:"target_percentage"`
	MeasurementWindow  string  `json:"measurement_window"`
	ServiceHours       string  `json:"service_hours"`
	CommittedMTBFHours float64 `json:"committed_mtbf_hours"`
	TargetMTTRMinutes  float64 `json:"target_mttr_minutes"`
	CIID               string  `json:"ci_id,omitempty"`
	Status             string  `json:"status,omitempty"`
	Description        string  `json:"description,omitempty"`
}

// Validate validates CreateAvailabilityTargetRequest
func (r *CreateAvailabilityTargetRequest) Validate() error {
	if strings.TrimSpace(r.Name) == "" {
		return ErrAvailabilityTargetNameRequired
	}
	if r.TargetPercentage <= 0 || r.TargetPercentage > 100.0 {
		return ErrInvalidTargetPercentage
	}
	if r.MeasurementWindow == "" {
		r.MeasurementWindow = MeasurementWindowMonthly
	} else if err := ValidateMeasurementWindow(r.MeasurementWindow); err != nil {
		return err
	}
	if r.ServiceHours == "" {
		r.ServiceHours = ServiceHours24x7
	} else if err := ValidateServiceHours(r.ServiceHours); err != nil {
		return err
	}
	if r.Status == "" {
		r.Status = TargetStatusActive
	} else if err := ValidateTargetStatus(r.Status); err != nil {
		return err
	}
	return nil
}

// UpdateAvailabilityTargetRequest contains fields to update on an existing target
type UpdateAvailabilityTargetRequest struct {
	Name               *string  `json:"name,omitempty"`
	TargetPercentage   *float64 `json:"target_percentage,omitempty"`
	MeasurementWindow  *string  `json:"measurement_window,omitempty"`
	ServiceHours       *string  `json:"service_hours,omitempty"`
	CommittedMTBFHours *float64 `json:"committed_mtbf_hours,omitempty"`
	TargetMTTRMinutes  *float64 `json:"target_mttr_minutes,omitempty"`
	CIID               *string  `json:"ci_id,omitempty"`
	Status             *string  `json:"status,omitempty"`
	Description        *string  `json:"description,omitempty"`
}

// Validate validates UpdateAvailabilityTargetRequest
func (r *UpdateAvailabilityTargetRequest) Validate() error {
	if r.Name != nil && strings.TrimSpace(*r.Name) == "" {
		return ErrAvailabilityTargetNameRequired
	}
	if r.TargetPercentage != nil && (*r.TargetPercentage <= 0 || *r.TargetPercentage > 100.0) {
		return ErrInvalidTargetPercentage
	}
	if r.MeasurementWindow != nil {
		if err := ValidateMeasurementWindow(*r.MeasurementWindow); err != nil {
			return err
		}
	}
	if r.ServiceHours != nil {
		if err := ValidateServiceHours(*r.ServiceHours); err != nil {
			return err
		}
	}
	if r.Status != nil {
		if err := ValidateTargetStatus(*r.Status); err != nil {
			return err
		}
	}
	return nil
}

// CreateOutageEventRequest contains parameters for creating an outage event
type CreateOutageEventRequest struct {
	TargetID        string     `json:"target_id,omitempty"`
	CIID            string     `json:"ci_id,omitempty"`
	Title           string     `json:"title"`
	OutageType      string     `json:"outage_type"`
	Severity        string     `json:"severity"`
	StartedAt       time.Time  `json:"started_at"`
	EndedAt         *time.Time `json:"ended_at,omitempty"`
	RootCause       string     `json:"root_cause,omitempty"`
	IsSLAExempt     bool       `json:"is_sla_exempt"`
	ExemptionReason string     `json:"exemption_reason,omitempty"`
	ReportedBy      string     `json:"reported_by,omitempty"`
}

// Validate validates CreateOutageEventRequest
func (r *CreateOutageEventRequest) Validate() error {
	if strings.TrimSpace(r.Title) == "" {
		return ErrOutageTitleRequired
	}
	if r.OutageType == "" {
		r.OutageType = OutageTypeUnplanned
	} else if err := ValidateOutageType(r.OutageType); err != nil {
		return err
	}
	if r.Severity == "" {
		r.Severity = OutageSeverityMajor
	} else if err := ValidateOutageSeverity(r.Severity); err != nil {
		return err
	}
	if r.StartedAt.IsZero() {
		r.StartedAt = time.Now().UTC()
	}
	if r.EndedAt != nil && r.EndedAt.Before(r.StartedAt) {
		return ErrInvalidOutageTimestamps
	}
	return nil
}

// UpdateOutageEventRequest contains fields to modify an existing outage event
type UpdateOutageEventRequest struct {
	Title           *string    `json:"title,omitempty"`
	OutageType      *string    `json:"outage_type,omitempty"`
	Severity        *string    `json:"severity,omitempty"`
	StartedAt       *time.Time `json:"started_at,omitempty"`
	EndedAt         *time.Time `json:"ended_at,omitempty"`
	RootCause       *string    `json:"root_cause,omitempty"`
	IsSLAExempt     *bool      `json:"is_sla_exempt,omitempty"`
	ExemptionReason *string    `json:"exemption_reason,omitempty"`
}

// Validate validates UpdateOutageEventRequest
func (r *UpdateOutageEventRequest) Validate() error {
	if r.Title != nil && strings.TrimSpace(*r.Title) == "" {
		return ErrOutageTitleRequired
	}
	if r.OutageType != nil {
		if err := ValidateOutageType(*r.OutageType); err != nil {
			return err
		}
	}
	if r.Severity != nil {
		if err := ValidateOutageSeverity(*r.Severity); err != nil {
			return err
		}
	}
	return nil
}

// ResolveOutageRequest contains parameters to resolve an ongoing outage
type ResolveOutageRequest struct {
	EndedAt         time.Time `json:"ended_at"`
	RootCause       string    `json:"root_cause,omitempty"`
	IsSLAExempt     *bool     `json:"is_sla_exempt,omitempty"`
	ExemptionReason string    `json:"exemption_reason,omitempty"`
}

// Validate validates ResolveOutageRequest
func (r *ResolveOutageRequest) Validate() error {
	if r.EndedAt.IsZero() {
		r.EndedAt = time.Now().UTC()
	}
	return nil
}

// CreateMeasurementRequest contains parameters to record or recalculate availability measurement
type CreateMeasurementRequest struct {
	TargetID                string  `json:"target_id"`
	Period                  string  `json:"period"`
	TotalOperationalMinutes float64 `json:"total_operational_minutes"`
	TotalDowntimeMinutes    float64 `json:"total_downtime_minutes"`
	ExemptDowntimeMinutes   float64 `json:"exempt_downtime_minutes"`
	MTBFHours               float64 `json:"mtbf_hours"`
	MTTRMinutes             float64 `json:"mttr_minutes"`
	OutageCount             int     `json:"outage_count"`
}

// Validate validates CreateMeasurementRequest
func (r *CreateMeasurementRequest) Validate() error {
	if strings.TrimSpace(r.TargetID) == "" {
		return ErrAvailabilityTargetNotFound
	}
	if strings.TrimSpace(r.Period) == "" {
		return ErrPeriodRequired
	}
	if r.TotalOperationalMinutes <= 0 {
		return ErrInvalidOperationalMinutes
	}
	return nil
}

// CreateSPOFRequest contains parameters for logging a single point of failure
type CreateSPOFRequest struct {
	ComponentName       string    `json:"component_name"`
	CIID                string    `json:"ci_id,omitempty"`
	RiskLevel           string    `json:"risk_level"`
	RedundancyMechanism string    `json:"redundancy_mechanism"`
	MitigationPlan      string    `json:"mitigation_plan,omitempty"`
	Status              string    `json:"status,omitempty"`
	ReviewDate          time.Time `json:"review_date"`
	OwnerID             string    `json:"owner_id,omitempty"`
}

// Validate validates CreateSPOFRequest
func (r *CreateSPOFRequest) Validate() error {
	if strings.TrimSpace(r.ComponentName) == "" {
		return ErrSPOFComponentNameRequired
	}
	if r.RiskLevel == "" {
		r.RiskLevel = SPOFRiskMedium
	} else if err := ValidateSPOFRiskLevel(r.RiskLevel); err != nil {
		return err
	}
	if r.RedundancyMechanism == "" {
		r.RedundancyMechanism = RedundancyMechanismNone
	} else if err := ValidateRedundancyMechanism(r.RedundancyMechanism); err != nil {
		return err
	}
	if r.Status == "" {
		r.Status = SPOFStatusOpen
	} else if err := ValidateSPOFStatus(r.Status); err != nil {
		return err
	}
	if r.ReviewDate.IsZero() {
		r.ReviewDate = time.Now().UTC().AddDate(0, 3, 0)
	}
	return nil
}

// UpdateSPOFRequest contains fields to update a SPOF entry
type UpdateSPOFRequest struct {
	ComponentName       *string    `json:"component_name,omitempty"`
	RiskLevel           *string    `json:"risk_level,omitempty"`
	RedundancyMechanism *string    `json:"redundancy_mechanism,omitempty"`
	MitigationPlan      *string    `json:"mitigation_plan,omitempty"`
	Status              *string    `json:"status,omitempty"`
	ReviewDate          *time.Time `json:"review_date,omitempty"`
	OwnerID             *string    `json:"owner_id,omitempty"`
}

// Validate validates UpdateSPOFRequest
func (r *UpdateSPOFRequest) Validate() error {
	if r.ComponentName != nil && strings.TrimSpace(*r.ComponentName) == "" {
		return ErrSPOFComponentNameRequired
	}
	if r.RiskLevel != nil {
		if err := ValidateSPOFRiskLevel(*r.RiskLevel); err != nil {
			return err
		}
	}
	if r.RedundancyMechanism != nil {
		if err := ValidateRedundancyMechanism(*r.RedundancyMechanism); err != nil {
			return err
		}
	}
	if r.Status != nil {
		if err := ValidateSPOFStatus(*r.Status); err != nil {
			return err
		}
	}
	return nil
}

// CreateAvailabilityTestRequest contains parameters for creating an availability test schedule
type CreateAvailabilityTestRequest struct {
	TargetID    string    `json:"target_id,omitempty"`
	Name        string    `json:"name"`
	TestType    string    `json:"test_type"`
	Frequency   string    `json:"frequency"`
	NextTestDue time.Time `json:"next_test_due"`
	Findings    string    `json:"findings,omitempty"`
	ConductedBy string    `json:"conducted_by,omitempty"`
}

// Validate validates CreateAvailabilityTestRequest
func (r *CreateAvailabilityTestRequest) Validate() error {
	if strings.TrimSpace(r.Name) == "" {
		return ErrAvailabilityTestNameRequired
	}
	if r.TestType == "" {
		r.TestType = TestTypeFailover
	} else if err := ValidateTestType(r.TestType); err != nil {
		return err
	}
	if r.Frequency == "" {
		r.Frequency = TestFrequencyQuarterly
	} else if err := ValidateTestFrequency(r.Frequency); err != nil {
		return err
	}
	if r.NextTestDue.IsZero() {
		r.NextTestDue = time.Now().UTC().AddDate(0, 1, 0)
	}
	return nil
}

// UpdateAvailabilityTestRequest contains fields to update an availability test schedule
type UpdateAvailabilityTestRequest struct {
	Name        *string    `json:"name,omitempty"`
	TestType    *string    `json:"test_type,omitempty"`
	Frequency   *string    `json:"frequency,omitempty"`
	NextTestDue *time.Time `json:"next_test_due,omitempty"`
	Findings    *string    `json:"findings,omitempty"`
	ConductedBy *string    `json:"conducted_by,omitempty"`
}

// Validate validates UpdateAvailabilityTestRequest
func (r *UpdateAvailabilityTestRequest) Validate() error {
	if r.Name != nil && strings.TrimSpace(*r.Name) == "" {
		return ErrAvailabilityTestNameRequired
	}
	if r.TestType != nil {
		if err := ValidateTestType(*r.TestType); err != nil {
			return err
		}
	}
	if r.Frequency != nil {
		if err := ValidateTestFrequency(*r.Frequency); err != nil {
			return err
		}
	}
	return nil
}

// RecordTestResultRequest captures the outcome and findings after conducting an availability test
type RecordTestResultRequest struct {
	TestedAt    time.Time `json:"tested_at"`
	Outcome     string    `json:"outcome"`
	Findings    string    `json:"findings,omitempty"`
	NextTestDue time.Time `json:"next_test_due"`
	ConductedBy string    `json:"conducted_by,omitempty"`
}

// Validate validates RecordTestResultRequest
func (r *RecordTestResultRequest) Validate() error {
	if r.TestedAt.IsZero() {
		r.TestedAt = time.Now().UTC()
	}
	if err := ValidateTestOutcome(r.Outcome); err != nil {
		return err
	}
	if r.NextTestDue.IsZero() {
		r.NextTestDue = r.TestedAt.AddDate(0, 3, 0)
	}
	return nil
}

// AvailabilityFilterCriteria defines search, filter, and pagination options for availability resources
type AvailabilityFilterCriteria struct {
	WorkspaceID string `json:"workspace_id"`
	TargetID    string `json:"target_id,omitempty"`
	CIID        string `json:"ci_id,omitempty"`
	Status      string `json:"status,omitempty"`
	OutageType  string `json:"outage_type,omitempty"`
	Severity    string `json:"severity,omitempty"`
	RiskLevel   string `json:"risk_level,omitempty"`
	Outcome     string `json:"outcome,omitempty"`
	Period      string `json:"period,omitempty"`
	Limit       int    `json:"limit,omitempty"`
	Offset      int    `json:"offset,omitempty"`
}

// TopOutageSummary represents outage frequency and downtime statistics grouped by CI
type TopOutageSummary struct {
	CIID                 string  `json:"ci_id"`
	OutageCount          int     `json:"outage_count"`
	TotalDowntimeMinutes float64 `json:"total_downtime_minutes"`
}

// AvailabilityAnalytics represents executive posture, SLA compliance rate, and operational resilience
type AvailabilityAnalytics struct {
	TotalTargets                  int                `json:"total_targets"`
	ActiveTargets                 int                `json:"active_targets"`
	OverallAvailabilityPercentage float64            `json:"overall_availability_percentage"`
	SLABreachCount                int                `json:"sla_breach_count"`
	SLAComplianceRate             float64            `json:"sla_compliance_rate"`
	AverageMTBFHours              float64            `json:"average_mtbf_hours"`
	AverageMTTRMinutes            float64            `json:"average_mttr_minutes"`
	TotalOutagesRecorded          int                `json:"total_outages_recorded"`
	ActiveOutages                 int                `json:"active_outages"`
	CriticalSPOFCount             int                `json:"critical_spof_count"`
	TotalSPOFCount                int                `json:"total_spof_count"`
	UpcomingTestsCount            int                `json:"upcoming_tests_count"`
	HealthIndexScore              float64            `json:"health_index_score"`
	TopOutagesByCI                []TopOutageSummary `json:"top_outages_by_ci"`
}

// CalculateHealthIndex derives a composite resilience score (0-100) based on availability, SLA compliance, active outages, and SPOFs
func (a *AvailabilityAnalytics) CalculateHealthIndex() float64 {
	score := 100.0

	// Deduct for active outages (15 points each, up to 30)
	activePenalty := float64(a.ActiveOutages) * 15.0
	if activePenalty > 30.0 {
		activePenalty = 30.0
	}
	score -= activePenalty

	// Deduct for critical SPOFs (10 points each, up to 30)
	spofPenalty := float64(a.CriticalSPOFCount) * 10.0
	if spofPenalty > 30.0 {
		spofPenalty = 30.0
	}
	score -= spofPenalty

	// Deduct for SLA compliance deficit
	if a.SLAComplianceRate < 100.0 {
		deficit := (100.0 - a.SLAComplianceRate) * 0.4
		score -= deficit
	}

	if score < 0.0 {
		score = 0.0
	}
	if score > 100.0 {
		score = 100.0
	}
	a.HealthIndexScore = score
	return score
}
