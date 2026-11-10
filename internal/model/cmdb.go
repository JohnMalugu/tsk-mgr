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

// Configuration Item relationship types
const (
	CIRelationDependsOn  = "depends_on"
	CIRelationRunsOn     = "runs_on"
	CIRelationConnectsTo = "connects_to"
	CIRelationHosts      = "hosts"
	CIRelationContains   = "contains"
	CIRelationBackedBy   = "backed_by"
	CIRelationManagedBy  = "managed_by"
)

// ValidateCIRelationType validates if a CI relationship type is valid
func ValidateCIRelationType(r string) error {
	switch strings.ToLower(strings.TrimSpace(r)) {
	case CIRelationDependsOn, CIRelationRunsOn, CIRelationConnectsTo, CIRelationHosts, CIRelationContains, CIRelationBackedBy, CIRelationManagedBy:
		return nil
	default:
		return ErrInvalidRelationship
	}
}

// CIRelationship represents a directional dependency or association between two Configuration Items
type CIRelationship struct {
	ID           int64     `json:"id"`
	SourceCIID   int64     `json:"source_ci_id"` // parent/dependent CI
	TargetCIID   int64     `json:"target_ci_id"` // prerequisite/underlying CI
	RelationType string    `json:"relation_type"`
	Description  string    `json:"description,omitempty"`
	CreatedAt    time.Time `json:"created_at"`

	// Enriched fields
	SourceCIName string `json:"source_ci_name,omitempty"`
	SourceCICode string `json:"source_ci_code,omitempty"`
	SourceCIType string `json:"source_ci_type,omitempty"`
	TargetCIName string `json:"target_ci_name,omitempty"`
	TargetCICode string `json:"target_ci_code,omitempty"`
	TargetCIType string `json:"target_ci_type,omitempty"`
}

// CI Link target entity types
const (
	CILinkEntityTask     = "task"
	CILinkEntityChange   = "change"
	CILinkEntityIncident = "incident"
)

// CI Link impact levels
const (
	CIImpactDirect     = "direct"
	CIImpactIndirect   = "indirect"
	CIImpactUnaffected = "unaffected"
)

// ValidateCILinkEntityType validates if an entity type is supported for CI linkage
func ValidateCILinkEntityType(t string) error {
	switch strings.ToLower(strings.TrimSpace(t)) {
	case CILinkEntityTask, CILinkEntityChange, CILinkEntityIncident:
		return nil
	default:
		return errors.New("invalid CI entity link type: must be task, change, or incident")
	}
}

// ValidateCIImpactLevel validates if an impact level is valid
func ValidateCIImpactLevel(l string) error {
	switch strings.ToLower(strings.TrimSpace(l)) {
	case CIImpactDirect, CIImpactIndirect, CIImpactUnaffected:
		return nil
	default:
		return ErrInvalidImpactLevel
	}
}

// CIEntityLink represents a link between a Configuration Item and a Task, Change Request, or Incident
type CIEntityLink struct {
	ID          int64     `json:"id"`
	CIID        int64     `json:"ci_id"`
	EntityType  string    `json:"entity_type"` // task, change, incident
	EntityID    int64     `json:"entity_id"`
	ImpactLevel string    `json:"impact_level"` // direct, indirect, unaffected
	Notes       string    `json:"notes,omitempty"`
	CreatedAt   time.Time `json:"created_at"`

	// Enriched fields
	CIName       string `json:"ci_name,omitempty"`
	CICode       string `json:"ci_code,omitempty"`
	EntityTitle  string `json:"entity_title,omitempty"`
	EntityStatus string `json:"entity_status,omitempty"`
}

// CI Maintenance statuses
const (
	CIMaintStatusScheduled  = "scheduled"
	CIMaintStatusInProgress = "in_progress"
	CIMaintStatusCompleted  = "completed"
	CIMaintStatusCancelled  = "cancelled"
)

// ValidateCIMaintStatus validates if a maintenance schedule status is supported
func ValidateCIMaintStatus(s string) error {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case CIMaintStatusScheduled, CIMaintStatusInProgress, CIMaintStatusCompleted, CIMaintStatusCancelled:
		return nil
	default:
		return ErrInvalidMaintStatus
	}
}

// CIMaintenanceSchedule represents a planned or active maintenance window for a Configuration Item
type CIMaintenanceSchedule struct {
	ID                      int64     `json:"id"`
	CIID                    int64     `json:"ci_id"`
	Title                   string    `json:"title"`
	Description             string    `json:"description,omitempty"`
	WindowStartAt           time.Time `json:"window_start_at"`
	WindowEndAt             time.Time `json:"window_end_at"`
	ExpectedDowntimeMinutes int       `json:"expected_downtime_minutes"`
	Status                  string    `json:"status"` // scheduled, in_progress, completed, cancelled
	CreatedByUserID         string    `json:"created_by_user_id"`
	CreatedAt               time.Time `json:"created_at"`
	UpdatedAt               time.Time `json:"updated_at"`

	// Enriched fields
	CIName            string `json:"ci_name,omitempty"`
	CICode            string `json:"ci_code,omitempty"`
	CreatedByUsername string `json:"created_by_username,omitempty"`
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
