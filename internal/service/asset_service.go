package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/JohnMalugu/tsk-mgr-api/internal/model"
	"github.com/JohnMalugu/tsk-mgr-api/internal/repository"
)

var (
	ErrAssetNotFound             = model.ErrAssetNotFound
	ErrAssetNameRequired         = model.ErrAssetNameRequired
	ErrAssetTagRequired          = model.ErrAssetTagRequired
	ErrInvalidAssetCategory      = model.ErrInvalidAssetCategory
	ErrInvalidAssetStatus        = model.ErrInvalidAssetStatus
	ErrInvalidDepreciation       = model.ErrInvalidDepreciation
	ErrAssetTagConflict          = model.ErrAssetTagConflict
	ErrAssetAlreadyAssigned      = model.ErrAssetAlreadyAssigned
	ErrAssetNotAssigned          = model.ErrAssetNotAssigned
	ErrAssetRetired              = model.ErrAssetRetired
	ErrLicenseNotFound           = model.ErrLicenseNotFound
	ErrLicenseNameRequired       = model.ErrLicenseNameRequired
	ErrInvalidLicenseType        = model.ErrInvalidLicenseType
	ErrNoSeatsAvailable          = model.ErrNoSeatsAvailable
	ErrLicenseAllocNotFound      = repository.ErrLicenseAllocNotFound
	ErrAlreadyAllocated          = model.ErrAlreadyAllocated
	ErrAllocationTargetReq       = model.ErrAllocationTargetReq
	ErrMaintenanceRecordNotFound = model.ErrMaintenanceRecordNotFound
	ErrInvalidMaintenanceType    = model.ErrInvalidMaintenanceType
	ErrInvalidMaintRecordStatus  = model.ErrInvalidMaintRecordStatus
	ErrAuditScanNotFound         = model.ErrAuditScanNotFound
	ErrInvalidAuditStatus        = model.ErrInvalidAuditStatus
	ErrAuditScanAssetRequired    = model.ErrAuditScanAssetRequired
)

// ITAssetService defines business logic for IT asset management, software licenses, maintenance, and audits
type ITAssetService interface {
	// Asset management
	CreateAsset(ctx context.Context, req *model.CreateAssetRequest) (*model.ITAsset, error)
	GetAssetByID(ctx context.Context, id int64) (*model.ITAsset, error)
	GetAssetByTag(ctx context.Context, workspaceID int64, assetTag string) (*model.ITAsset, error)
	UpdateAsset(ctx context.Context, id int64, req *model.UpdateAssetRequest) (*model.ITAsset, error)
	DeleteAsset(ctx context.Context, id int64) error
	ListAssets(ctx context.Context, filter model.AssetFilterCriteria) ([]model.ITAsset, int, error)

	// Assignment and lifecycle transitions
	AssignAsset(ctx context.Context, id int64, req *model.AssignAssetRequest) (*model.ITAsset, error)
	CheckinAsset(ctx context.Context, id int64, location, notes string) (*model.ITAsset, error)
	TransitionAssetStatus(ctx context.Context, id int64, req *model.AssetStatusTransitionRequest) (*model.ITAsset, error)
	CalculateDepreciation(ctx context.Context, id int64) (*model.AssetDepreciationSchedule, error)

	// Maintenance
	CreateMaintenanceRecord(ctx context.Context, assetID int64, req *model.CreateMaintenanceRecordRequest) (*model.AssetMaintenanceRecord, error)
	GetMaintenanceRecordByID(ctx context.Context, id int64) (*model.AssetMaintenanceRecord, error)
	UpdateMaintenanceRecord(ctx context.Context, id int64, req *model.UpdateMaintenanceRecordRequest) (*model.AssetMaintenanceRecord, error)
	ListMaintenanceRecords(ctx context.Context, assetID int64) ([]model.AssetMaintenanceRecord, error)
	DeleteMaintenanceRecord(ctx context.Context, id int64) error

	// Software Licenses
	CreateSoftwareLicense(ctx context.Context, req *model.CreateSoftwareLicenseRequest) (*model.SoftwareLicense, error)
	GetSoftwareLicenseByID(ctx context.Context, id int64) (*model.SoftwareLicense, error)
	UpdateSoftwareLicense(ctx context.Context, id int64, req *model.UpdateSoftwareLicenseRequest) (*model.SoftwareLicense, error)
	DeleteSoftwareLicense(ctx context.Context, id int64) error
	ListSoftwareLicenses(ctx context.Context, filter model.LicenseFilterCriteria) ([]model.SoftwareLicense, int, error)

	// License Allocations
	AllocateLicense(ctx context.Context, licenseID int64, req *model.AllocateLicenseRequest, allocatedBy string) (*model.LicenseAllocation, error)
	GetLicenseAllocationByID(ctx context.Context, id int64) (*model.LicenseAllocation, error)
	DeallocateLicense(ctx context.Context, id int64) error
	ListLicenseAllocations(ctx context.Context, licenseID int64) ([]model.LicenseAllocation, error)

	// Audits and Analytics
	RecordAuditScan(ctx context.Context, workspaceID int64, scannedBy string, req *model.AssetAuditScanRequest) (*model.AssetAuditLog, error)
	ListAuditLogs(ctx context.Context, workspaceID int64, assetID *int64, limit, offset int) ([]model.AssetAuditLog, error)
	GetAssetMetrics(ctx context.Context, workspaceID int64) (*model.AssetMetrics, error)
}

type itAssetService struct {
	repo repository.ITAssetRepository
}

// NewITAssetService creates a new instance of ITAssetService
func NewITAssetService(repo repository.ITAssetRepository) ITAssetService {
	return &itAssetService{repo: repo}
}

var errNotImpl = errors.New("not implemented")

func computeAssetBookValue(asset *model.ITAsset) {
	if asset == nil {
		return
	}
	if asset.PurchaseDate == nil || asset.DepreciationMethod == model.DepreciationNone || asset.UsefulLifeMonths <= 0 {
		asset.CurrentBookValue = asset.PurchaseCost
		asset.AgeMonths = 0
		return
	}

	now := time.Now().UTC()
	elapsedMonths := int(now.Sub(*asset.PurchaseDate).Hours() / (24 * 30.4375))
	if elapsedMonths < 0 {
		elapsedMonths = 0
	}
	asset.AgeMonths = elapsedMonths

	depreciable := asset.PurchaseCost - asset.ResidualValue
	if depreciable <= 0 {
		asset.CurrentBookValue = asset.PurchaseCost
		return
	}

	monthlyDepr := depreciable / float64(asset.UsefulLifeMonths)
	accumDepr := monthlyDepr * float64(elapsedMonths)
	currentBook := asset.PurchaseCost - accumDepr
	if currentBook < asset.ResidualValue {
		currentBook = asset.ResidualValue
	}
	asset.CurrentBookValue = currentBook
}

func (s *itAssetService) CreateAsset(ctx context.Context, req *model.CreateAssetRequest) (*model.ITAsset, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	status := strings.ToLower(strings.TrimSpace(req.Status))
	if status == "" {
		status = model.AssetStatusInStock
	}

	currency := strings.ToUpper(strings.TrimSpace(req.Currency))
	if currency == "" {
		currency = "USD"
	}

	deprMethod := strings.ToLower(strings.TrimSpace(req.DepreciationMethod))
	if deprMethod == "" {
		deprMethod = model.DepreciationStraightLine
	}

	usefulLife := req.UsefulLifeMonths
	if usefulLife <= 0 {
		usefulLife = 36
	}

	tag := strings.TrimSpace(req.AssetTag)
	if tag == "" {
		seq, err := s.repo.GetNextAssetTagSequence(ctx, req.WorkspaceID)
		if err != nil {
			return nil, err
		}
		tag = fmt.Sprintf("AST-%04d", seq)
	}

	var assignedAt *time.Time
	if req.AssignedToUserID != nil && strings.TrimSpace(*req.AssignedToUserID) != "" {
		now := time.Now().UTC()
		assignedAt = &now
		status = model.AssetStatusInUse
	}

	asset := &model.ITAsset{
		WorkspaceID:        req.WorkspaceID,
		AssetTag:           tag,
		Name:               strings.TrimSpace(req.Name),
		Category:           strings.ToLower(strings.TrimSpace(req.Category)),
		Status:             status,
		Make:               strings.TrimSpace(req.Make),
		Model:              strings.TrimSpace(req.Model),
		SerialNumber:       strings.TrimSpace(req.SerialNumber),
		PurchaseDate:       req.PurchaseDate,
		PurchaseCost:       req.PurchaseCost,
		Currency:           currency,
		WarrantyExpiry:     req.WarrantyExpiry,
		ResidualValue:      req.ResidualValue,
		UsefulLifeMonths:   usefulLife,
		DepreciationMethod: deprMethod,
		Location:           strings.TrimSpace(req.Location),
		AssignedToUserID:   req.AssignedToUserID,
		AssignedToDept:     strings.TrimSpace(req.AssignedToDept),
		AssignedAt:         assignedAt,
		CIID:               req.CIID,
		Notes:              strings.TrimSpace(req.Notes),
	}

	if err := s.repo.CreateAsset(ctx, asset); err != nil {
		return nil, err
	}

	created, err := s.repo.GetAssetByID(ctx, asset.ID)
	if err != nil {
		return asset, nil
	}
	computeAssetBookValue(created)
	return created, nil
}

func (s *itAssetService) GetAssetByID(ctx context.Context, id int64) (*model.ITAsset, error) {
	asset, err := s.repo.GetAssetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	computeAssetBookValue(asset)
	return asset, nil
}

func (s *itAssetService) GetAssetByTag(ctx context.Context, workspaceID int64, assetTag string) (*model.ITAsset, error) {
	asset, err := s.repo.GetAssetByTag(ctx, workspaceID, assetTag)
	if err != nil {
		return nil, err
	}
	computeAssetBookValue(asset)
	return asset, nil
}

func (s *itAssetService) UpdateAsset(ctx context.Context, id int64, req *model.UpdateAssetRequest) (*model.ITAsset, error) {
	asset, err := s.repo.GetAssetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)
		if name == "" {
			return nil, ErrAssetNameRequired
		}
		asset.Name = name
	}
	if req.Category != nil {
		cat := strings.ToLower(strings.TrimSpace(*req.Category))
		if err := model.ValidateAssetCategory(cat); err != nil {
			return nil, err
		}
		asset.Category = cat
	}
	if req.Make != nil {
		asset.Make = strings.TrimSpace(*req.Make)
	}
	if req.Model != nil {
		asset.Model = strings.TrimSpace(*req.Model)
	}
	if req.SerialNumber != nil {
		asset.SerialNumber = strings.TrimSpace(*req.SerialNumber)
	}
	if req.PurchaseDate != nil {
		asset.PurchaseDate = req.PurchaseDate
	}
	if req.PurchaseCost != nil {
		asset.PurchaseCost = *req.PurchaseCost
	}
	if req.Currency != nil {
		asset.Currency = strings.ToUpper(strings.TrimSpace(*req.Currency))
	}
	if req.WarrantyExpiry != nil {
		asset.WarrantyExpiry = req.WarrantyExpiry
	}
	if req.ResidualValue != nil {
		asset.ResidualValue = *req.ResidualValue
	}
	if req.UsefulLifeMonths != nil && *req.UsefulLifeMonths > 0 {
		asset.UsefulLifeMonths = *req.UsefulLifeMonths
	}
	if req.DepreciationMethod != nil {
		method := strings.ToLower(strings.TrimSpace(*req.DepreciationMethod))
		if err := model.ValidateDepreciationMethod(method); err != nil {
			return nil, err
		}
		asset.DepreciationMethod = method
	}
	if req.Location != nil {
		asset.Location = strings.TrimSpace(*req.Location)
	}
	if req.AssignedToDept != nil {
		asset.AssignedToDept = strings.TrimSpace(*req.AssignedToDept)
	}
	if req.CIID != nil {
		asset.CIID = req.CIID
	}
	if req.Notes != nil {
		asset.Notes = strings.TrimSpace(*req.Notes)
	}

	if err := s.repo.UpdateAsset(ctx, asset); err != nil {
		return nil, err
	}

	updated, err := s.repo.GetAssetByID(ctx, id)
	if err != nil {
		return asset, nil
	}
	computeAssetBookValue(updated)
	return updated, nil
}

func (s *itAssetService) DeleteAsset(ctx context.Context, id int64) error {
	return s.repo.DeleteAsset(ctx, id)
}

func (s *itAssetService) ListAssets(ctx context.Context, filter model.AssetFilterCriteria) ([]model.ITAsset, int, error) {
	assets, total, err := s.repo.ListAssets(ctx, filter)
	if err != nil {
		return nil, 0, err
	}
	for i := range assets {
		computeAssetBookValue(&assets[i])
	}
	return assets, total, nil
}
func (s *itAssetService) AssignAsset(ctx context.Context, id int64, req *model.AssignAssetRequest) (*model.ITAsset, error) {
	return nil, errNotImpl
}
func (s *itAssetService) CheckinAsset(ctx context.Context, id int64, location, notes string) (*model.ITAsset, error) {
	return nil, errNotImpl
}
func (s *itAssetService) TransitionAssetStatus(ctx context.Context, id int64, req *model.AssetStatusTransitionRequest) (*model.ITAsset, error) {
	return nil, errNotImpl
}
func (s *itAssetService) CalculateDepreciation(ctx context.Context, id int64) (*model.AssetDepreciationSchedule, error) {
	return nil, errNotImpl
}
func (s *itAssetService) CreateMaintenanceRecord(ctx context.Context, assetID int64, req *model.CreateMaintenanceRecordRequest) (*model.AssetMaintenanceRecord, error) {
	return nil, errNotImpl
}
func (s *itAssetService) GetMaintenanceRecordByID(ctx context.Context, id int64) (*model.AssetMaintenanceRecord, error) {
	return nil, errNotImpl
}
func (s *itAssetService) UpdateMaintenanceRecord(ctx context.Context, id int64, req *model.UpdateMaintenanceRecordRequest) (*model.AssetMaintenanceRecord, error) {
	return nil, errNotImpl
}
func (s *itAssetService) ListMaintenanceRecords(ctx context.Context, assetID int64) ([]model.AssetMaintenanceRecord, error) {
	return nil, errNotImpl
}
func (s *itAssetService) DeleteMaintenanceRecord(ctx context.Context, id int64) error {
	return errNotImpl
}
func (s *itAssetService) CreateSoftwareLicense(ctx context.Context, req *model.CreateSoftwareLicenseRequest) (*model.SoftwareLicense, error) {
	return nil, errNotImpl
}
func (s *itAssetService) GetSoftwareLicenseByID(ctx context.Context, id int64) (*model.SoftwareLicense, error) {
	return nil, errNotImpl
}
func (s *itAssetService) UpdateSoftwareLicense(ctx context.Context, id int64, req *model.UpdateSoftwareLicenseRequest) (*model.SoftwareLicense, error) {
	return nil, errNotImpl
}
func (s *itAssetService) DeleteSoftwareLicense(ctx context.Context, id int64) error {
	return errNotImpl
}
func (s *itAssetService) ListSoftwareLicenses(ctx context.Context, filter model.LicenseFilterCriteria) ([]model.SoftwareLicense, int, error) {
	return nil, 0, errNotImpl
}
func (s *itAssetService) AllocateLicense(ctx context.Context, licenseID int64, req *model.AllocateLicenseRequest, allocatedBy string) (*model.LicenseAllocation, error) {
	return nil, errNotImpl
}
func (s *itAssetService) GetLicenseAllocationByID(ctx context.Context, id int64) (*model.LicenseAllocation, error) {
	return nil, errNotImpl
}
func (s *itAssetService) DeallocateLicense(ctx context.Context, id int64) error {
	return errNotImpl
}
func (s *itAssetService) ListLicenseAllocations(ctx context.Context, licenseID int64) ([]model.LicenseAllocation, error) {
	return nil, errNotImpl
}
func (s *itAssetService) RecordAuditScan(ctx context.Context, workspaceID int64, scannedBy string, req *model.AssetAuditScanRequest) (*model.AssetAuditLog, error) {
	return nil, errNotImpl
}
func (s *itAssetService) ListAuditLogs(ctx context.Context, workspaceID int64, assetID *int64, limit, offset int) ([]model.AssetAuditLog, error) {
	return nil, errNotImpl
}
func (s *itAssetService) GetAssetMetrics(ctx context.Context, workspaceID int64) (*model.AssetMetrics, error) {
	return nil, errNotImpl
}
