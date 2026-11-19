package model

import (
	"errors"
	"strings"
	"time"
)

// Department constants for IT Cost Centers
const (
	DeptEngineering    = "engineering"
	DeptInfrastructure = "infrastructure"
	DeptSecurity       = "security"
	DeptProduct        = "product"
	DeptOperations     = "operations"
	DeptFinance        = "finance"
)

// Cost Center Status constants
const (
	CostCenterStatusActive   = "active"
	CostCenterStatusInactive = "inactive"
	CostCenterStatusArchived = "archived"
)

// FinOps domain errors
var (
	ErrCostCenterNotFound     = errors.New("cost center not found")
	ErrCostCenterCodeRequired = errors.New("cost center code is required")
	ErrCostCenterCodeConflict = errors.New("cost center with this code already exists in workspace")
	ErrCostCenterNameRequired = errors.New("cost center name is required")
	ErrInvalidCostCenterStatus = errors.New("invalid cost center status")
	ErrInvalidDepartment       = errors.New("invalid department")
)

// ValidateCostCenterStatus validates whether the given status is recognized
func ValidateCostCenterStatus(status string) error {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case CostCenterStatusActive, CostCenterStatusInactive, CostCenterStatusArchived:
		return nil
	default:
		return ErrInvalidCostCenterStatus
	}
}

// ValidateDepartment validates whether the given department is recognized
func ValidateDepartment(dept string) error {
	switch strings.ToLower(strings.TrimSpace(dept)) {
	case DeptEngineering, DeptInfrastructure, DeptSecurity, DeptProduct, DeptOperations, DeptFinance:
		return nil
	default:
		return ErrInvalidDepartment
	}
}

// CostCenter represents an accounting and operational unit for IT spending and cost attribution
type CostCenter struct {
	ID          string    `json:"id"`
	WorkspaceID string    `json:"workspace_id"`
	Code        string    `json:"code"`
	Name        string    `json:"name"`
	Department  string    `json:"department"`
	ManagerID   string    `json:"manager_id"`
	Status      string    `json:"status"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
