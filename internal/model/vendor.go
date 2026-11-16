package model

import (
	"errors"
	"strings"
	"time"
)

// Vendor category constants
const (
	VendorCategorySoftware   = "software"
	VendorCategoryHardware   = "hardware"
	VendorCategoryCloud      = "cloud"
	VendorCategoryConsulting = "consulting"
	VendorCategoryTelecom    = "telecom"
	VendorCategoryFacilities = "facilities"
	VendorCategoryOther      = "other"
)

// Vendor tier constants
const (
	VendorTierStrategic = "strategic"
	VendorTierPreferred = "preferred"
	VendorTierTactical  = "tactical"
	VendorTierCommodity = "commodity"
)

// Vendor lifecycle status constants
const (
	VendorStatusActive      = "active"
	VendorStatusUnderReview = "under_review"
	VendorStatusSuspended   = "suspended"
	VendorStatusTerminated  = "terminated"
)

// Vendor risk levels
const (
	VendorRiskLow      = "low"
	VendorRiskMedium   = "medium"
	VendorRiskHigh     = "high"
	VendorRiskCritical = "critical"
)

var (
	ErrVendorNotFound       = errors.New("vendor not found")
	ErrVendorNameRequired   = errors.New("vendor name is required")
	ErrVendorSlugRequired   = errors.New("vendor slug is required")
	ErrVendorSlugConflict   = errors.New("vendor slug already exists")
	ErrInvalidVendorCategory = errors.New("invalid vendor category")
	ErrInvalidVendorTier     = errors.New("invalid vendor tier")
	ErrInvalidVendorStatus   = errors.New("invalid vendor status")
	ErrInvalidVendorRisk     = errors.New("invalid vendor risk level")
)

// ValidateVendorCategory validates vendor category
func ValidateVendorCategory(c string) error {
	switch strings.ToLower(strings.TrimSpace(c)) {
	case VendorCategorySoftware, VendorCategoryHardware, VendorCategoryCloud,
		VendorCategoryConsulting, VendorCategoryTelecom, VendorCategoryFacilities, VendorCategoryOther:
		return nil
	default:
		return ErrInvalidVendorCategory
	}
}

// ValidateVendorTier validates vendor relationship tier
func ValidateVendorTier(t string) error {
	switch strings.ToLower(strings.TrimSpace(t)) {
	case VendorTierStrategic, VendorTierPreferred, VendorTierTactical, VendorTierCommodity:
		return nil
	default:
		return ErrInvalidVendorTier
	}
}

// ValidateVendorStatus validates vendor lifecycle status
func ValidateVendorStatus(s string) error {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case VendorStatusActive, VendorStatusUnderReview, VendorStatusSuspended, VendorStatusTerminated:
		return nil
	default:
		return ErrInvalidVendorStatus
	}
}

// ValidateVendorRisk validates vendor risk rating
func ValidateVendorRisk(r string) error {
	switch strings.ToLower(strings.TrimSpace(r)) {
	case VendorRiskLow, VendorRiskMedium, VendorRiskHigh, VendorRiskCritical:
		return nil
	default:
		return ErrInvalidVendorRisk
	}
}

// Vendor represents a 3rd-party supplier, service provider, or vendor organisation
type Vendor struct {
	ID                     int64     `json:"id"`
	WorkspaceID            int64     `json:"workspace_id"`
	Name                   string    `json:"name"`
	Slug                   string    `json:"slug"`
	Category               string    `json:"category"`
	Tier                   string    `json:"tier"`
	Status                 string    `json:"status"`
	RiskLevel              string    `json:"risk_level"`
	Website                string    `json:"website,omitempty"`
	ContactPerson          string    `json:"contact_person,omitempty"`
	ContactEmail           string    `json:"contact_email,omitempty"`
	ContactPhone           string    `json:"contact_phone,omitempty"`
	SupportPortalURL       string    `json:"support_portal_url,omitempty"`
	SecurityCertifications string    `json:"security_certifications,omitempty"`
	Notes                  string    `json:"notes,omitempty"`
	CreatedAt              time.Time `json:"created_at"`
	UpdatedAt              time.Time `json:"updated_at"`

	// Enriched fields
	ActiveContractsCount   int     `json:"active_contracts_count,omitempty"`
	TotalAnnualSpend       float64 `json:"total_annual_spend,omitempty"`
	AverageEvaluationScore float64 `json:"average_evaluation_score,omitempty"`
}
