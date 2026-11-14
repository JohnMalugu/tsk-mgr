package model

import (
	"errors"
	"strings"
	"time"
)

// Hardware and Asset category constants
const (
	AssetCategoryLaptop      = "laptop"
	AssetCategoryDesktop     = "desktop"
	AssetCategoryServer      = "server"
	AssetCategoryNetwork     = "network"
	AssetCategoryMobile      = "mobile_device"
	AssetCategoryPeripheral  = "peripheral"
	AssetCategoryCloudVM     = "cloud_vm"
	AssetCategoryOther       = "other"
)

// Asset lifecycle status constants
const (
	AssetStatusOrdered   = "ordered"
	AssetStatusReceived  = "received"
	AssetStatusInStock   = "in_stock"
	AssetStatusInUse     = "in_use"
	AssetStatusInRepair  = "in_repair"
	AssetStatusRetired   = "retired"
	AssetStatusDisposed  = "disposed"
)

// Asset depreciation methods
const (
	DepreciationStraightLine = "straight_line"
	DepreciationNone         = "none"
)

var (
	ErrAssetNotFound       = errors.New("asset not found")
	ErrAssetNameRequired   = errors.New("asset name is required")
	ErrAssetTagRequired    = errors.New("asset tag is required")
	ErrInvalidAssetCategory = errors.New("invalid asset category")
	ErrInvalidAssetStatus   = errors.New("invalid asset status")
	ErrInvalidDepreciation  = errors.New("invalid depreciation method")
	ErrAssetTagConflict    = errors.New("asset tag already exists")
	ErrAssetAlreadyAssigned = errors.New("asset is already assigned")
	ErrAssetNotAssigned     = errors.New("asset is not assigned")
	ErrAssetRetired         = errors.New("cannot perform operation on retired or disposed asset")
)

// ValidateAssetCategory checks if category is valid
func ValidateAssetCategory(c string) error {
	switch strings.ToLower(strings.TrimSpace(c)) {
	case AssetCategoryLaptop, AssetCategoryDesktop, AssetCategoryServer, AssetCategoryNetwork,
		AssetCategoryMobile, AssetCategoryPeripheral, AssetCategoryCloudVM, AssetCategoryOther:
		return nil
	default:
		return ErrInvalidAssetCategory
	}
}

// ValidateAssetStatus checks if status is valid
func ValidateAssetStatus(s string) error {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case AssetStatusOrdered, AssetStatusReceived, AssetStatusInStock, AssetStatusInUse,
		AssetStatusInRepair, AssetStatusRetired, AssetStatusDisposed:
		return nil
	default:
		return ErrInvalidAssetStatus
	}
}

// ValidateDepreciationMethod checks if depreciation method is valid
func ValidateDepreciationMethod(m string) error {
	switch strings.ToLower(strings.TrimSpace(m)) {
	case DepreciationStraightLine, DepreciationNone:
		return nil
	default:
		return ErrInvalidDepreciation
	}
}

// ITAsset represents a physical or virtual enterprise hardware asset
type ITAsset struct {
	ID                 int64      `json:"id"`
	WorkspaceID        int64      `json:"workspace_id"`
	AssetTag           string     `json:"asset_tag"`
	Name               string     `json:"name"`
	Category           string     `json:"category"`
	Status             string     `json:"status"`
	Make               string     `json:"make"`
	Model              string     `json:"model"`
	SerialNumber       string     `json:"serial_number"`
	PurchaseDate       *time.Time `json:"purchase_date,omitempty"`
	PurchaseCost       float64    `json:"purchase_cost"`
	Currency           string     `json:"currency"`
	WarrantyExpiry     *time.Time `json:"warranty_expiry,omitempty"`
	ResidualValue      float64    `json:"residual_value"`
	UsefulLifeMonths   int        `json:"useful_life_months"`
	DepreciationMethod string     `json:"depreciation_method"`
	Location           string     `json:"location"`
	AssignedToUserID   *string    `json:"assigned_to_user_id,omitempty"`
	AssignedToDept     string     `json:"assigned_to_department,omitempty"`
	AssignedAt         *time.Time `json:"assigned_at,omitempty"`
	CIID               *int64     `json:"ci_id,omitempty"`
	Notes              string     `json:"notes,omitempty"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`

	// Enriched fields
	AssignedUserName  string  `json:"assigned_user_name,omitempty"`
	AssignedUserEmail string  `json:"assigned_user_email,omitempty"`
	CIName            string  `json:"ci_name,omitempty"`
	CurrentBookValue  float64 `json:"current_book_value,omitempty"`
	AgeMonths         int     `json:"age_months,omitempty"`
}

// Software License types
const (
	LicenseTypePerpetual    = "perpetual"
	LicenseTypeSubscription = "subscription"
	LicenseTypeSeatBased    = "seat_based"
	LicenseTypeSiteWide     = "site_wide"
)

var (
	ErrLicenseNotFound     = errors.New("software license not found")
	ErrLicenseNameRequired = errors.New("software name is required")
	ErrInvalidLicenseType  = errors.New("invalid software license type")
	ErrNoSeatsAvailable    = errors.New("no license seats available")
	ErrAllocationNotFound  = errors.New("license allocation not found")
	ErrAlreadyAllocated    = errors.New("license is already allocated to this target")
	ErrAllocationTargetReq = errors.New("allocation target (user or asset) is required")
)

// ValidateLicenseType validates software license type
func ValidateLicenseType(t string) error {
	switch strings.ToLower(strings.TrimSpace(t)) {
	case LicenseTypePerpetual, LicenseTypeSubscription, LicenseTypeSeatBased, LicenseTypeSiteWide:
		return nil
	default:
		return ErrInvalidLicenseType
	}
}

// SoftwareLicense represents a managed enterprise software license or subscription
type SoftwareLicense struct {
	ID             int64      `json:"id"`
	WorkspaceID    int64      `json:"workspace_id"`
	SoftwareName   string     `json:"software_name"`
	Version        string     `json:"version"`
	Publisher      string     `json:"publisher"`
	LicenseKey     string     `json:"license_key,omitempty"`
	LicenseType    string     `json:"license_type"`
	TotalSeats     int        `json:"total_seats"`
	AllocatedSeats int        `json:"allocated_seats"`
	PurchaseCost   float64    `json:"purchase_cost"`
	Currency       string     `json:"currency"`
	ExpirationDate *time.Time `json:"expiration_date,omitempty"`
	Vendor         string     `json:"vendor,omitempty"`
	Notes          string     `json:"notes,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`

	// Enriched fields
	AvailableSeats int `json:"available_seats"`
}

// LicenseAllocation maps a software license seat to a user or hardware asset
type LicenseAllocation struct {
	ID          int64     `json:"id"`
	LicenseID   int64     `json:"license_id"`
	AssetID     *int64    `json:"asset_id,omitempty"`
	UserID      *string   `json:"user_id,omitempty"`
	AllocatedBy string    `json:"allocated_by"`
	AllocatedAt time.Time `json:"allocated_at"`
	Notes       string    `json:"notes,omitempty"`

	// Enriched fields
	AssetTag     string `json:"asset_tag,omitempty"`
	AssetName    string `json:"asset_name,omitempty"`
	UserName     string `json:"user_name,omitempty"`
	UserEmail    string `json:"user_email,omitempty"`
	SoftwareName string `json:"software_name,omitempty"`
}

// Maintenance types for IT assets
const (
	MaintTypePreventative = "preventative"
	MaintTypeRepair       = "repair"
	MaintTypeUpgrade      = "upgrade"
	MaintTypeInspection   = "inspection"
)

// Maintenance statuses for IT assets
const (
	MaintRecordScheduled  = "scheduled"
	MaintRecordInProgress = "in_progress"
	MaintRecordCompleted  = "completed"
	MaintRecordCancelled  = "cancelled"
)

// Physical audit scan status constants
const (
	AuditStatusVerified  = "verified"
	AuditStatusMisplaced = "misplaced"
	AuditStatusMissing   = "missing"
)

var (
	ErrMaintenanceRecordNotFound = errors.New("asset maintenance record not found")
	ErrInvalidMaintenanceType    = errors.New("invalid maintenance type")
	ErrInvalidMaintRecordStatus  = errors.New("invalid maintenance record status")
	ErrAuditScanNotFound         = errors.New("asset audit scan record not found")
	ErrInvalidAuditStatus        = errors.New("invalid asset audit status")
	ErrAuditScanAssetRequired    = errors.New("asset ID or tag is required for audit scan")
)

// ValidateMaintenanceType validates maintenance record type
func ValidateMaintenanceType(t string) error {
	switch strings.ToLower(strings.TrimSpace(t)) {
	case MaintTypePreventative, MaintTypeRepair, MaintTypeUpgrade, MaintTypeInspection:
		return nil
	default:
		return ErrInvalidMaintenanceType
	}
}

// ValidateMaintenanceRecordStatus validates maintenance record status
func ValidateMaintenanceRecordStatus(s string) error {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case MaintRecordScheduled, MaintRecordInProgress, MaintRecordCompleted, MaintRecordCancelled:
		return nil
	default:
		return ErrInvalidMaintRecordStatus
	}
}

// ValidateAuditStatus validates physical inventory audit status
func ValidateAuditStatus(s string) error {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case AuditStatusVerified, AuditStatusMisplaced, AuditStatusMissing:
		return nil
	default:
		return ErrInvalidAuditStatus
	}
}

// AssetMaintenanceRecord tracks maintenance, servicing, and repair history for an asset
type AssetMaintenanceRecord struct {
	ID              int64      `json:"id"`
	AssetID         int64      `json:"asset_id"`
	MaintenanceType string     `json:"maintenance_type"`
	Status          string     `json:"status"`
	Cost            float64    `json:"cost"`
	PerformedBy     string     `json:"performed_by"`
	Description     string     `json:"description"`
	ScheduledDate   *time.Time `json:"scheduled_date,omitempty"`
	CompletedDate   *time.Time `json:"completed_date,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`

	// Enriched fields
	AssetTag  string `json:"asset_tag,omitempty"`
	AssetName string `json:"asset_name,omitempty"`
}

// AssetAuditLog records a physical or barcode inventory audit scan event
type AssetAuditLog struct {
	ID          int64     `json:"id"`
	WorkspaceID int64     `json:"workspace_id"`
	AssetID     int64     `json:"asset_id"`
	ScannedBy   string    `json:"scanned_by"`
	ScannedAt   time.Time `json:"scanned_at"`
	AuditStatus string    `json:"audit_status"`
	Location    string    `json:"location"`
	Notes       string    `json:"notes,omitempty"`

	// Enriched fields
	AssetTag    string `json:"asset_tag,omitempty"`
	AssetName   string `json:"asset_name,omitempty"`
	ScannerName string `json:"scanner_name,omitempty"`
}

// CreateAssetRequest represents payload to register a new IT asset
type CreateAssetRequest struct {
	WorkspaceID        int64      `json:"workspace_id"`
	AssetTag           string     `json:"asset_tag,omitempty"`
	Name               string     `json:"name"`
	Category           string     `json:"category"`
	Status             string     `json:"status,omitempty"`
	Make               string     `json:"make,omitempty"`
	Model              string     `json:"model,omitempty"`
	SerialNumber       string     `json:"serial_number,omitempty"`
	PurchaseDate       *time.Time `json:"purchase_date,omitempty"`
	PurchaseCost       float64    `json:"purchase_cost,omitempty"`
	Currency           string     `json:"currency,omitempty"`
	WarrantyExpiry     *time.Time `json:"warranty_expiry,omitempty"`
	ResidualValue      float64    `json:"residual_value,omitempty"`
	UsefulLifeMonths   int        `json:"useful_life_months,omitempty"`
	DepreciationMethod string     `json:"depreciation_method,omitempty"`
	Location           string     `json:"location,omitempty"`
	AssignedToUserID   *string    `json:"assigned_to_user_id,omitempty"`
	AssignedToDept     string     `json:"assigned_to_department,omitempty"`
	CIID               *int64     `json:"ci_id,omitempty"`
	Notes              string     `json:"notes,omitempty"`
}

// Validate validates CreateAssetRequest fields
func (r *CreateAssetRequest) Validate() error {
	if strings.TrimSpace(r.Name) == "" {
		return ErrAssetNameRequired
	}
	if err := ValidateAssetCategory(r.Category); err != nil {
		return err
	}
	if r.Status != "" {
		if err := ValidateAssetStatus(r.Status); err != nil {
			return err
		}
	}
	if r.DepreciationMethod != "" {
		if err := ValidateDepreciationMethod(r.DepreciationMethod); err != nil {
			return err
		}
	}
	return nil
}

// UpdateAssetRequest represents payload to update existing asset attributes
type UpdateAssetRequest struct {
	Name               *string    `json:"name,omitempty"`
	Category           *string    `json:"category,omitempty"`
	Make               *string    `json:"make,omitempty"`
	Model              *string    `json:"model,omitempty"`
	SerialNumber       *string    `json:"serial_number,omitempty"`
	PurchaseDate       *time.Time `json:"purchase_date,omitempty"`
	PurchaseCost       *float64   `json:"purchase_cost,omitempty"`
	Currency           *string    `json:"currency,omitempty"`
	WarrantyExpiry     *time.Time `json:"warranty_expiry,omitempty"`
	ResidualValue      *float64   `json:"residual_value,omitempty"`
	UsefulLifeMonths   *int       `json:"useful_life_months,omitempty"`
	DepreciationMethod *string    `json:"depreciation_method,omitempty"`
	Location           *string    `json:"location,omitempty"`
	AssignedToDept     *string    `json:"assigned_to_department,omitempty"`
	CIID               *int64     `json:"ci_id,omitempty"`
	Notes              *string    `json:"notes,omitempty"`
}

// AssignAssetRequest represents payload to assign an asset to a user or department
type AssignAssetRequest struct {
	UserID     *string `json:"user_id,omitempty"`
	Department string  `json:"department,omitempty"`
	Location   string  `json:"location,omitempty"`
	Notes      string  `json:"notes,omitempty"`
}

// AssetStatusTransitionRequest represents payload to transition an asset's lifecycle status
type AssetStatusTransitionRequest struct {
	Status string `json:"status"`
	Notes  string `json:"notes,omitempty"`
}

// Validate validates AssetStatusTransitionRequest
func (r *AssetStatusTransitionRequest) Validate() error {
	return ValidateAssetStatus(r.Status)
}
