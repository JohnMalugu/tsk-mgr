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

// Availability domain errors for targets
var (
	ErrAvailabilityTargetNotFound     = errors.New("availability target not found")
	ErrAvailabilityTargetNameRequired = errors.New("availability target name is required")
	ErrInvalidTargetPercentage        = errors.New("target percentage must be between 0.0 and 100.0")
	ErrInvalidMeasurementWindow       = errors.New("invalid measurement window")
	ErrInvalidServiceHours            = errors.New("invalid service hours")
	ErrInvalidTargetStatus            = errors.New("invalid target status")
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
