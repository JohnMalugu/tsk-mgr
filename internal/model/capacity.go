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
