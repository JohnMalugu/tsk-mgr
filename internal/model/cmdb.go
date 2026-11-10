package model

import (
	"errors"
	"strings"
	"time"
)

var (
	ErrInvalidCIType        = errors.New("invalid configuration item type")
	ErrInvalidCICriticality = errors.New("invalid configuration item criticality tier")
	ErrInvalidCIStatus      = errors.New("invalid configuration item status")
	ErrInvalidCIEnvironment = errors.New("invalid configuration item environment")
	ErrInvalidRelationship  = errors.New("invalid CI relationship type")
	ErrInvalidImpactLevel   = errors.New("invalid CI impact level")
	ErrInvalidMaintStatus   = errors.New("invalid CI maintenance status")
)

// Configuration Item types in ITIL CMDB
const (
	CITypeService       = "service"
	CITypeApplication   = "application"
	CITypeDatabase      = "database"
	CITypeServer        = "server"
	CITypeCluster       = "cluster"
	CITypeNetwork       = "network"
	CITypeCloudResource = "cloud_resource"
)

// Configuration Item criticality tiers
const (
	CICriticalityTier1 = "tier_1_mission_critical"
	CICriticalityTier2 = "tier_2_business_critical"
	CICriticalityTier3 = "tier_3_supporting"
	CICriticalityTier4 = "tier_4_non_critical"
)

// Configuration Item lifecycle statuses
const (
	CIStatusPlanning       = "planning"
	CIStatusActive         = "active"
	CIStatusMaintenance    = "maintenance"
	CIStatusDegraded       = "degraded"
	CIStatusDecommissioned = "decommissioned"
	CIStatusRetired        = "retired"
)

// Configuration Item environment types
const (
	CIEnvProduction  = "production"
	CIEnvStaging     = "staging"
	CIEnvTesting     = "testing"
	CIEnvDevelopment = "development"
)

// ValidateCIStatus validates if a CI status string is supported
func ValidateCIStatus(s string) error {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case CIStatusPlanning, CIStatusActive, CIStatusMaintenance, CIStatusDegraded, CIStatusDecommissioned, CIStatusRetired:
		return nil
	default:
		return ErrInvalidCIStatus
	}
}

// ValidateCIEnvironment validates if a CI environment string is supported
func ValidateCIEnvironment(e string) error {
	switch strings.ToLower(strings.TrimSpace(e)) {
	case CIEnvProduction, CIEnvStaging, CIEnvTesting, CIEnvDevelopment:
		return nil
	default:
		return ErrInvalidCIEnvironment
	}
}

// ValidateCIType validates if a CI type string is supported
func ValidateCIType(t string) error {
	switch strings.ToLower(strings.TrimSpace(t)) {
	case CITypeService, CITypeApplication, CITypeDatabase, CITypeServer, CITypeCluster, CITypeNetwork, CITypeCloudResource:
		return nil
	default:
		return ErrInvalidCIType
	}
}

// ValidateCICriticality validates if a CI criticality tier string is supported
func ValidateCICriticality(c string) error {
	switch strings.ToLower(strings.TrimSpace(c)) {
	case CICriticalityTier1, CICriticalityTier2, CICriticalityTier3, CICriticalityTier4:
		return nil
	default:
		return ErrInvalidCICriticality
	}
}

// ConfigurationItem represents an enterprise IT infrastructure or service configuration item in CMDB
type ConfigurationItem struct {
	ID           int64     `json:"id"`
	WorkspaceID  int64     `json:"workspace_id"`
	ProjectID    *int64    `json:"project_id,omitempty"`
	CICode       string    `json:"ci_code"` // e.g. "CI-1"
	Name         string    `json:"name"`
	CIType       string    `json:"ci_type"`
	Status       string    `json:"status"`
	Criticality  string    `json:"criticality"`
	Environment  string    `json:"environment"`
	OwnerUserID  string    `json:"owner_user_id"`
	IPAddress    string    `json:"ip_address,omitempty"`
	Hostname     string    `json:"hostname,omitempty"`
	Version      string    `json:"version,omitempty"`
	CostPerMonth float64   `json:"cost_per_month,omitempty"`
	Attributes   string    `json:"attributes,omitempty"` // JSON metadata key-values
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`

	// Enriched fields
	OwnerName           string `json:"owner_name,omitempty"`
	ProjectName         string `json:"project_name,omitempty"`
	UpstreamCount       int    `json:"upstream_count,omitempty"`
	DownstreamCount     int    `json:"downstream_count,omitempty"`
	LinkedEntitiesCount int    `json:"linked_entities_count,omitempty"`
}
