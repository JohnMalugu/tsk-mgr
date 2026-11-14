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
