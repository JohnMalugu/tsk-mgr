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

// Contract milestone status constants
const (
	ContractMilestonePending   = "pending"
	ContractMilestoneCompleted = "completed"
	ContractMilestoneCancelled = "cancelled"
)

// Vendor CI relationship types
const (
	VendorCIRelProvides  = "provides"
	VendorCIRelMaintains = "maintains"
	VendorCIRelSupports  = "supports"
	VendorCIRelLicenses  = "licenses"
)

var (
	ErrEvaluationNotFound     = errors.New("vendor evaluation not found")
	ErrMilestoneNotFound      = errors.New("contract milestone not found")
	ErrInvalidScore           = errors.New("score must be between 1.0 and 5.0")
	ErrInvalidMilestoneStatus = errors.New("invalid contract milestone status")
	ErrVendorCILinkNotFound   = errors.New("vendor CI link not found")
	ErrVendorCILinkConflict   = errors.New("vendor CI link already exists")
)

// ValidateScore validates that an evaluation rating is between 1.0 and 5.0
func ValidateScore(s float64) error {
	if s < 1.0 || s > 5.0 {
		return ErrInvalidScore
	}
	return nil
}

// ValidateMilestoneStatus validates milestone status
func ValidateMilestoneStatus(s string) error {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case ContractMilestonePending, ContractMilestoneCompleted, ContractMilestoneCancelled:
		return nil
	default:
		return ErrInvalidMilestoneStatus
	}
}

// VendorEvaluation records periodic performance scorecards and reviews for a vendor
type VendorEvaluation struct {
	ID                  int64     `json:"id"`
	VendorID            int64     `json:"vendor_id"`
	EvaluatorID         string    `json:"evaluator_id"`
	EvaluationPeriod    string    `json:"evaluation_period"`
	QualityScore        float64   `json:"quality_score"`
	ResponsivenessScore float64   `json:"responsiveness_score"`
	SLAScore            float64   `json:"sla_score"`
	PricingScore        float64   `json:"pricing_score"`
	OverallScore        float64   `json:"overall_score"`
	Feedback            string    `json:"feedback,omitempty"`
	CreatedAt           time.Time `json:"created_at"`

	// Enriched fields
	EvaluatorName string `json:"evaluator_name,omitempty"`
	VendorName    string `json:"vendor_name,omitempty"`
}

// ContractMilestone tracks key deliverable phases and payment releases for a contract
type ContractMilestone struct {
	ID          int64      `json:"id"`
	ContractID  int64      `json:"contract_id"`
	Title       string     `json:"title"`
	Description string     `json:"description,omitempty"`
	DueDate     time.Time  `json:"due_date"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	Amount      float64    `json:"amount"`
	Status      string     `json:"status"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// VendorCILink connects a vendor to a CMDB Configuration Item
type VendorCILink struct {
	ID           int64     `json:"id"`
	VendorID     int64     `json:"vendor_id"`
	CIID         int64     `json:"ci_id"`
	Relationship string    `json:"relationship"`
	CreatedAt    time.Time `json:"created_at"`

	// Enriched fields
	CIName string `json:"ci_name,omitempty"`
	CIType string `json:"ci_type,omitempty"`
}

// CreateVendorRequest represents payload to register a new vendor
type CreateVendorRequest struct {
	WorkspaceID            int64  `json:"workspace_id"`
	Name                   string `json:"name"`
	Slug                   string `json:"slug,omitempty"`
	Category               string `json:"category,omitempty"`
	Tier                   string `json:"tier,omitempty"`
	Status                 string `json:"status,omitempty"`
	RiskLevel              string `json:"risk_level,omitempty"`
	Website                string `json:"website,omitempty"`
	ContactPerson          string `json:"contact_person,omitempty"`
	ContactEmail           string `json:"contact_email,omitempty"`
	ContactPhone           string `json:"contact_phone,omitempty"`
	SupportPortalURL       string `json:"support_portal_url,omitempty"`
	SecurityCertifications string `json:"security_certifications,omitempty"`
	Notes                  string `json:"notes,omitempty"`
}

// Validate validates CreateVendorRequest parameters
func (r *CreateVendorRequest) Validate() error {
	if strings.TrimSpace(r.Name) == "" {
		return ErrVendorNameRequired
	}
	if r.Category != "" {
		if err := ValidateVendorCategory(r.Category); err != nil {
			return err
		}
	}
	if r.Tier != "" {
		if err := ValidateVendorTier(r.Tier); err != nil {
			return err
		}
	}
	if r.Status != "" {
		if err := ValidateVendorStatus(r.Status); err != nil {
			return err
		}
	}
	if r.RiskLevel != "" {
		if err := ValidateVendorRisk(r.RiskLevel); err != nil {
			return err
		}
	}
	return nil
}

// UpdateVendorRequest represents payload to modify existing vendor properties
type UpdateVendorRequest struct {
	Name                   *string `json:"name,omitempty"`
	Category               *string `json:"category,omitempty"`
	Tier                   *string `json:"tier,omitempty"`
	Status                 *string `json:"status,omitempty"`
	RiskLevel              *string `json:"risk_level,omitempty"`
	Website                *string `json:"website,omitempty"`
	ContactPerson          *string `json:"contact_person,omitempty"`
	ContactEmail           *string `json:"contact_email,omitempty"`
	ContactPhone           *string `json:"contact_phone,omitempty"`
	SupportPortalURL       *string `json:"support_portal_url,omitempty"`
	SecurityCertifications *string `json:"security_certifications,omitempty"`
	Notes                  *string `json:"notes,omitempty"`
}

// CreateVendorContractRequest represents payload to create a new vendor contract
type CreateVendorContractRequest struct {
	WorkspaceID         int64     `json:"workspace_id"`
	VendorID            int64     `json:"vendor_id"`
	ContractNumber      string    `json:"contract_number,omitempty"`
	Title               string    `json:"title"`
	ContractType        string    `json:"contract_type,omitempty"`
	Status              string    `json:"status,omitempty"`
	TotalValue          float64   `json:"total_value,omitempty"`
	AnnualRecurringCost float64   `json:"annual_recurring_cost,omitempty"`
	Currency            string    `json:"currency,omitempty"`
	StartDate           time.Time `json:"start_date"`
	EndDate             time.Time `json:"end_date"`
	AutoRenew           bool      `json:"auto_renew,omitempty"`
	RenewalNoticeDays   int       `json:"renewal_notice_days,omitempty"`
	OwnerUserID         string    `json:"owner_user_id,omitempty"`
	PaymentTerms        string    `json:"payment_terms,omitempty"`
	DocumentURL         string    `json:"document_url,omitempty"`
	Notes               string    `json:"notes,omitempty"`
}

// Validate validates CreateVendorContractRequest
func (r *CreateVendorContractRequest) Validate() error {
	if strings.TrimSpace(r.Title) == "" {
		return ErrContractTitleRequired
	}
	if r.VendorID <= 0 {
		return ErrVendorNotFound
	}
	if r.ContractType != "" {
		if err := ValidateContractType(r.ContractType); err != nil {
			return err
		}
	}
	if r.Status != "" {
		if err := ValidateContractStatus(r.Status); err != nil {
			return err
		}
	}
	if !r.StartDate.IsZero() && !r.EndDate.IsZero() && !r.EndDate.After(r.StartDate) {
		return ErrInvalidContractDates
	}
	return nil
}

// UpdateVendorContractRequest represents payload to update contract terms
type UpdateVendorContractRequest struct {
	Title               *string    `json:"title,omitempty"`
	ContractType        *string    `json:"contract_type,omitempty"`
	Status              *string    `json:"status,omitempty"`
	TotalValue          *float64   `json:"total_value,omitempty"`
	AnnualRecurringCost *float64   `json:"annual_recurring_cost,omitempty"`
	Currency            *string    `json:"currency,omitempty"`
	StartDate           *time.Time `json:"start_date,omitempty"`
	EndDate             *time.Time `json:"end_date,omitempty"`
	AutoRenew           *bool      `json:"auto_renew,omitempty"`
	RenewalNoticeDays   *int       `json:"renewal_notice_days,omitempty"`
	OwnerUserID         *string    `json:"owner_user_id,omitempty"`
	PaymentTerms        *string    `json:"payment_terms,omitempty"`
	DocumentURL         *string    `json:"document_url,omitempty"`
	Notes               *string    `json:"notes,omitempty"`
}

// CreateVendorEvaluationRequest represents payload to submit a vendor review scorecard
type CreateVendorEvaluationRequest struct {
	EvaluationPeriod    string  `json:"evaluation_period"`
	QualityScore        float64 `json:"quality_score"`
	ResponsivenessScore float64 `json:"responsiveness_score"`
	SLAScore            float64 `json:"sla_score"`
	PricingScore        float64 `json:"pricing_score"`
	Feedback            string  `json:"feedback,omitempty"`
}

// Validate validates CreateVendorEvaluationRequest
func (r *CreateVendorEvaluationRequest) Validate() error {
	if strings.TrimSpace(r.EvaluationPeriod) == "" {
		return errors.New("evaluation period is required")
	}
	if err := ValidateScore(r.QualityScore); err != nil {
		return err
	}
	if err := ValidateScore(r.ResponsivenessScore); err != nil {
		return err
	}
	if err := ValidateScore(r.SLAScore); err != nil {
		return err
	}
	if err := ValidateScore(r.PricingScore); err != nil {
		return err
	}
	return nil
}

// RenewContractRequest represents payload to extend and renew an existing contract
type RenewContractRequest struct {
	NewEndDate      time.Time `json:"new_end_date"`
	NewTotalValue   *float64  `json:"new_total_value,omitempty"`
	NewAnnualCost   *float64  `json:"new_annual_recurring_cost,omitempty"`
	AutoRenew       *bool     `json:"auto_renew,omitempty"`
	NoticeDays      *int      `json:"renewal_notice_days,omitempty"`
	Notes           string    `json:"notes,omitempty"`
}

// Validate validates RenewContractRequest
func (r *RenewContractRequest) Validate() error {
	if r.NewEndDate.IsZero() {
		return errors.New("new end date is required")
	}
	return nil
}

// CreateContractMilestoneRequest represents payload to add a milestone to a contract
type CreateContractMilestoneRequest struct {
	Title       string    `json:"title"`
	Description string    `json:"description,omitempty"`
	DueDate     time.Time `json:"due_date"`
	Amount      float64   `json:"amount,omitempty"`
	Status      string    `json:"status,omitempty"`
}

// Validate validates CreateContractMilestoneRequest
func (r *CreateContractMilestoneRequest) Validate() error {
	if strings.TrimSpace(r.Title) == "" {
		return errors.New("milestone title is required")
	}
	if r.DueDate.IsZero() {
		return errors.New("milestone due date is required")
	}
	if r.Status != "" {
		if err := ValidateMilestoneStatus(r.Status); err != nil {
			return err
		}
	}
	return nil
}

// UpdateContractMilestoneRequest represents payload to modify milestone progress
type UpdateContractMilestoneRequest struct {
	Title       *string    `json:"title,omitempty"`
	Description *string    `json:"description,omitempty"`
	DueDate     *time.Time `json:"due_date,omitempty"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	Amount      *float64   `json:"amount,omitempty"`
	Status      *string    `json:"status,omitempty"`
}

// LinkVendorCIRequest represents payload to associate a vendor with a CMDB CI
type LinkVendorCIRequest struct {
	CIID         int64  `json:"ci_id"`
	Relationship string `json:"relationship,omitempty"`
}

// Validate validates LinkVendorCIRequest
func (r *LinkVendorCIRequest) Validate() error {
	if r.CIID <= 0 {
		return errors.New("valid configuration item ID is required")
	}
	return nil
}
