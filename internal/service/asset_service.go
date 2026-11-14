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
	asset, err := s.repo.GetAssetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if asset.Status == model.AssetStatusRetired || asset.Status == model.AssetStatusDisposed {
		return nil, ErrAssetRetired
	}

	dept := strings.TrimSpace(req.Department)
	var userID *string
	if req.UserID != nil && strings.TrimSpace(*req.UserID) != "" {
		trimmed := strings.TrimSpace(*req.UserID)
		userID = &trimmed
	}

	if userID == nil && dept == "" {
		return nil, errors.New("user ID or department is required for assignment")
	}

	now := time.Now().UTC()
	if err := s.repo.AssignAsset(ctx, id, userID, dept, strings.TrimSpace(req.Location), now, strings.TrimSpace(req.Notes)); err != nil {
		return nil, err
	}

	updated, err := s.repo.GetAssetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	computeAssetBookValue(updated)
	return updated, nil
}

func (s *itAssetService) CheckinAsset(ctx context.Context, id int64, location, notes string) (*model.ITAsset, error) {
	asset, err := s.repo.GetAssetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if asset.AssignedToUserID == nil && asset.AssignedToDept == "" {
		return nil, ErrAssetNotAssigned
	}

	if err := s.repo.CheckinAsset(ctx, id, strings.TrimSpace(location), strings.TrimSpace(notes)); err != nil {
		return nil, err
	}

	updated, err := s.repo.GetAssetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	computeAssetBookValue(updated)
	return updated, nil
}

func (s *itAssetService) TransitionAssetStatus(ctx context.Context, id int64, req *model.AssetStatusTransitionRequest) (*model.ITAsset, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	asset, err := s.repo.GetAssetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if asset.Status == model.AssetStatusDisposed {
		return nil, ErrAssetRetired
	}

	status := strings.ToLower(strings.TrimSpace(req.Status))
	if err := s.repo.UpdateAssetStatus(ctx, id, status, strings.TrimSpace(req.Notes)); err != nil {
		return nil, err
	}

	updated, err := s.repo.GetAssetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	computeAssetBookValue(updated)
	return updated, nil
}

func (s *itAssetService) CalculateDepreciation(ctx context.Context, id int64) (*model.AssetDepreciationSchedule, error) {
	asset, err := s.repo.GetAssetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	sched := &model.AssetDepreciationSchedule{
		AssetID:          asset.ID,
		AssetTag:         asset.AssetTag,
		AssetName:        asset.Name,
		PurchaseCost:     asset.PurchaseCost,
		ResidualValue:    asset.ResidualValue,
		UsefulLifeMonths: asset.UsefulLifeMonths,
	}

	if asset.PurchaseDate == nil || asset.DepreciationMethod == model.DepreciationNone || asset.UsefulLifeMonths <= 0 {
		sched.CurrentBookValue = asset.PurchaseCost
		return sched, nil
	}

	now := time.Now().UTC()
	elapsedMonths := int(now.Sub(*asset.PurchaseDate).Hours() / (24 * 30.4375))
	if elapsedMonths < 0 {
		elapsedMonths = 0
	}
	sched.AgeMonths = elapsedMonths

	depreciable := asset.PurchaseCost - asset.ResidualValue
	if depreciable <= 0 {
		sched.CurrentBookValue = asset.PurchaseCost
		return sched, nil
	}

	monthlyDepr := depreciable / float64(asset.UsefulLifeMonths)
	sched.MonthlyDepreciation = monthlyDepr
	accumDepr := monthlyDepr * float64(elapsedMonths)
	sched.AccumulatedDepreciation = accumDepr

	currentBook := asset.PurchaseCost - accumDepr
	if currentBook < asset.ResidualValue {
		currentBook = asset.ResidualValue
	}
	sched.CurrentBookValue = currentBook

	return sched, nil
}
func (s *itAssetService) CreateMaintenanceRecord(ctx context.Context, assetID int64, req *model.CreateMaintenanceRecordRequest) (*model.AssetMaintenanceRecord, error) {
	if _, err := s.repo.GetAssetByID(ctx, assetID); err != nil {
		return nil, err
	}
	if err := req.Validate(); err != nil {
		return nil, err
	}

	status := strings.ToLower(strings.TrimSpace(req.Status))
	if status == "" {
		status = model.MaintRecordScheduled
	}

	record := &model.AssetMaintenanceRecord{
		AssetID:         assetID,
		MaintenanceType: strings.ToLower(strings.TrimSpace(req.MaintenanceType)),
		Status:          status,
		Cost:            req.Cost,
		PerformedBy:     strings.TrimSpace(req.PerformedBy),
		Description:     strings.TrimSpace(req.Description),
		ScheduledDate:   req.ScheduledDate,
		CompletedDate:   req.CompletedDate,
	}

	if err := s.repo.CreateMaintenanceRecord(ctx, record); err != nil {
		return nil, err
	}

	return s.repo.GetMaintenanceRecordByID(ctx, record.ID)
}

func (s *itAssetService) GetMaintenanceRecordByID(ctx context.Context, id int64) (*model.AssetMaintenanceRecord, error) {
	return s.repo.GetMaintenanceRecordByID(ctx, id)
}

func (s *itAssetService) UpdateMaintenanceRecord(ctx context.Context, id int64, req *model.UpdateMaintenanceRecordRequest) (*model.AssetMaintenanceRecord, error) {
	record, err := s.repo.GetMaintenanceRecordByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.MaintenanceType != nil {
		mType := strings.ToLower(strings.TrimSpace(*req.MaintenanceType))
		if err := model.ValidateMaintenanceType(mType); err != nil {
			return nil, err
		}
		record.MaintenanceType = mType
	}
	if req.Status != nil {
		mStatus := strings.ToLower(strings.TrimSpace(*req.Status))
		if err := model.ValidateMaintenanceRecordStatus(mStatus); err != nil {
			return nil, err
		}
		record.Status = mStatus
	}
	if req.Cost != nil {
		record.Cost = *req.Cost
	}
	if req.PerformedBy != nil {
		record.PerformedBy = strings.TrimSpace(*req.PerformedBy)
	}
	if req.Description != nil {
		desc := strings.TrimSpace(*req.Description)
		if desc == "" {
			return nil, errors.New("maintenance description cannot be empty")
		}
		record.Description = desc
	}
	if req.ScheduledDate != nil {
		record.ScheduledDate = req.ScheduledDate
	}
	if req.CompletedDate != nil {
		record.CompletedDate = req.CompletedDate
	}

	if err := s.repo.UpdateMaintenanceRecord(ctx, record); err != nil {
		return nil, err
	}

	return s.repo.GetMaintenanceRecordByID(ctx, id)
}

func (s *itAssetService) ListMaintenanceRecords(ctx context.Context, assetID int64) ([]model.AssetMaintenanceRecord, error) {
	if _, err := s.repo.GetAssetByID(ctx, assetID); err != nil {
		return nil, err
	}
	return s.repo.ListMaintenanceRecords(ctx, assetID)
}

func (s *itAssetService) DeleteMaintenanceRecord(ctx context.Context, id int64) error {
	return s.repo.DeleteMaintenanceRecord(ctx, id)
}

func (s *itAssetService) CreateSoftwareLicense(ctx context.Context, req *model.CreateSoftwareLicenseRequest) (*model.SoftwareLicense, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	currency := strings.ToUpper(strings.TrimSpace(req.Currency))
	if currency == "" {
		currency = "USD"
	}

	totalSeats := req.TotalSeats
	if totalSeats <= 0 {
		totalSeats = 1
	}

	license := &model.SoftwareLicense{
		WorkspaceID:    req.WorkspaceID,
		SoftwareName:   strings.TrimSpace(req.SoftwareName),
		Version:        strings.TrimSpace(req.Version),
		Publisher:      strings.TrimSpace(req.Publisher),
		LicenseKey:     strings.TrimSpace(req.LicenseKey),
		LicenseType:    strings.ToLower(strings.TrimSpace(req.LicenseType)),
		TotalSeats:     totalSeats,
		PurchaseCost:   req.PurchaseCost,
		Currency:       currency,
		ExpirationDate: req.ExpirationDate,
		Vendor:         strings.TrimSpace(req.Vendor),
		Notes:          strings.TrimSpace(req.Notes),
	}

	if err := s.repo.CreateSoftwareLicense(ctx, license); err != nil {
		return nil, err
	}

	return s.repo.GetSoftwareLicenseByID(ctx, license.ID)
}

func (s *itAssetService) GetSoftwareLicenseByID(ctx context.Context, id int64) (*model.SoftwareLicense, error) {
	return s.repo.GetSoftwareLicenseByID(ctx, id)
}

func (s *itAssetService) UpdateSoftwareLicense(ctx context.Context, id int64, req *model.UpdateSoftwareLicenseRequest) (*model.SoftwareLicense, error) {
	license, err := s.repo.GetSoftwareLicenseByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.SoftwareName != nil {
		name := strings.TrimSpace(*req.SoftwareName)
		if name == "" {
			return nil, ErrLicenseNameRequired
		}
		license.SoftwareName = name
	}
	if req.Version != nil {
		license.Version = strings.TrimSpace(*req.Version)
	}
	if req.Publisher != nil {
		license.Publisher = strings.TrimSpace(*req.Publisher)
	}
	if req.LicenseKey != nil {
		license.LicenseKey = strings.TrimSpace(*req.LicenseKey)
	}
	if req.LicenseType != nil {
		lType := strings.ToLower(strings.TrimSpace(*req.LicenseType))
		if err := model.ValidateLicenseType(lType); err != nil {
			return nil, err
		}
		license.LicenseType = lType
	}
	if req.TotalSeats != nil && *req.TotalSeats >= 0 {
		license.TotalSeats = *req.TotalSeats
	}
	if req.PurchaseCost != nil {
		license.PurchaseCost = *req.PurchaseCost
	}
	if req.Currency != nil {
		license.Currency = strings.ToUpper(strings.TrimSpace(*req.Currency))
	}
	if req.ExpirationDate != nil {
		license.ExpirationDate = req.ExpirationDate
	}
	if req.Vendor != nil {
		license.Vendor = strings.TrimSpace(*req.Vendor)
	}
	if req.Notes != nil {
		license.Notes = strings.TrimSpace(*req.Notes)
	}

	if err := s.repo.UpdateSoftwareLicense(ctx, license); err != nil {
		return nil, err
	}

	return s.repo.GetSoftwareLicenseByID(ctx, id)
}

func (s *itAssetService) DeleteSoftwareLicense(ctx context.Context, id int64) error {
	return s.repo.DeleteSoftwareLicense(ctx, id)
}

func (s *itAssetService) ListSoftwareLicenses(ctx context.Context, filter model.LicenseFilterCriteria) ([]model.SoftwareLicense, int, error) {
	return s.repo.ListSoftwareLicenses(ctx, filter)
}

func (s *itAssetService) AllocateLicense(ctx context.Context, licenseID int64, req *model.AllocateLicenseRequest, allocatedBy string) (*model.LicenseAllocation, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	if _, err := s.repo.GetSoftwareLicenseByID(ctx, licenseID); err != nil {
		return nil, err
	}

	if req.AssetID != nil && *req.AssetID > 0 {
		if _, err := s.repo.GetAssetByID(ctx, *req.AssetID); err != nil {
			return nil, err
		}
	}

	alloc := &model.LicenseAllocation{
		LicenseID:   licenseID,
		AssetID:     req.AssetID,
		UserID:      req.UserID,
		AllocatedBy: allocatedBy,
		Notes:       strings.TrimSpace(req.Notes),
	}

	if err := s.repo.CreateLicenseAllocation(ctx, alloc); err != nil {
		return nil, err
	}

	return s.repo.GetLicenseAllocationByID(ctx, alloc.ID)
}

func (s *itAssetService) GetLicenseAllocationByID(ctx context.Context, id int64) (*model.LicenseAllocation, error) {
	return s.repo.GetLicenseAllocationByID(ctx, id)
}

func (s *itAssetService) DeallocateLicense(ctx context.Context, id int64) error {
	return s.repo.DeleteLicenseAllocation(ctx, id)
}

func (s *itAssetService) ListLicenseAllocations(ctx context.Context, licenseID int64) ([]model.LicenseAllocation, error) {
	if _, err := s.repo.GetSoftwareLicenseByID(ctx, licenseID); err != nil {
		return nil, err
	}
	return s.repo.ListLicenseAllocations(ctx, licenseID)
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
