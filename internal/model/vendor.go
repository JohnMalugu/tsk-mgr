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

// Vendor contract types
const (
	ContractTypeMSA         = "msa"
	ContractTypeSOW         = "sow"
	ContractTypeSLA         = "sla"
	ContractTypeLicense     = "license"
	ContractTypeMaintenance = "maintenance"
	ContractTypeNDA         = "nda"
)

// Vendor contract lifecycle statuses
const (
	ContractStatusDraft          = "draft"
	ContractStatusUnderReview    = "under_review"
	ContractStatusActive         = "active"
	ContractStatusPendingRenewal = "pending_renewal"
	ContractStatusExpired        = "expired"
	ContractStatusTerminated     = "terminated"
)

var (
	ErrContractNotFound       = errors.New("vendor contract not found")
	ErrContractTitleRequired   = errors.New("contract title is required")
	ErrContractNumberRequired  = errors.New("contract number is required")
	ErrInvalidContractType     = errors.New("invalid contract type")
	ErrInvalidContractStatus   = errors.New("invalid contract status")
	ErrInvalidContractDates    = errors.New("contract end date must be after start date")
	ErrContractNumberConflict  = errors.New("contract number already exists")
)

// ValidateContractType validates contract type
func ValidateContractType(t string) error {
	switch strings.ToLower(strings.TrimSpace(t)) {
	case ContractTypeMSA, ContractTypeSOW, ContractTypeSLA,
		ContractTypeLicense, ContractTypeMaintenance, ContractTypeNDA:
		return nil
	default:
		return ErrInvalidContractType
	}
}

// ValidateContractStatus validates contract status
func ValidateContractStatus(s string) error {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case ContractStatusDraft, ContractStatusUnderReview, ContractStatusActive,
		ContractStatusPendingRenewal, ContractStatusExpired, ContractStatusTerminated:
		return nil
	default:
		return ErrInvalidContractStatus
	}
}

// VendorContract represents a formal contract, MSA, SOW, or agreement with a vendor
type VendorContract struct {
	ID                 int64     `json:"id"`
	WorkspaceID        int64     `json:"workspace_id"`
	VendorID           int64     `json:"vendor_id"`
	ContractNumber     string    `json:"contract_number"`
	Title              string    `json:"title"`
	ContractType       string    `json:"contract_type"`
	Status             string    `json:"status"`
	TotalValue         float64   `json:"total_value"`
	AnnualRecurringCost float64  `json:"annual_recurring_cost"`
	Currency           string    `json:"currency"`
	StartDate          time.Time `json:"start_date"`
	EndDate            time.Time `json:"end_date"`
	AutoRenew          bool      `json:"auto_renew"`
	RenewalNoticeDays  int       `json:"renewal_notice_days"`
	OwnerUserID        string    `json:"owner_user_id"`
	PaymentTerms       string    `json:"payment_terms,omitempty"`
	DocumentURL        string    `json:"document_url,omitempty"`
	Notes              string    `json:"notes,omitempty"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`

	// Enriched fields
	VendorName          string `json:"vendor_name,omitempty"`
	OwnerUserName       string `json:"owner_user_name,omitempty"`
	DaysUntilExpiration int    `json:"days_until_expiration,omitempty"`
	MilestonesCount     int    `json:"milestones_count,omitempty"`
}
