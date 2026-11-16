package service

import (
	"context"
	"errors"
	"strings"

	"github.com/JohnMalugu/tsk-mgr-api/internal/model"
	"github.com/JohnMalugu/tsk-mgr-api/internal/repository"
)

var (
	ErrVendorNotFound            = model.ErrVendorNotFound
	ErrVendorNameRequired        = model.ErrVendorNameRequired
	ErrVendorSlugRequired        = model.ErrVendorSlugRequired
	ErrVendorSlugConflict        = model.ErrVendorSlugConflict
	ErrInvalidVendorCategory     = model.ErrInvalidVendorCategory
	ErrInvalidVendorTier         = model.ErrInvalidVendorTier
	ErrInvalidVendorStatus       = model.ErrInvalidVendorStatus
	ErrInvalidVendorRisk         = model.ErrInvalidVendorRisk
	ErrContractNotFound          = model.ErrContractNotFound
	ErrContractTitleRequired     = model.ErrContractTitleRequired
	ErrContractNumberRequired    = model.ErrContractNumberRequired
	ErrInvalidContractType       = model.ErrInvalidContractType
	ErrInvalidContractStatus     = model.ErrInvalidContractStatus
	ErrInvalidContractDates      = model.ErrInvalidContractDates
	ErrContractNumberConflict    = model.ErrContractNumberConflict
	ErrEvaluationNotFound        = model.ErrEvaluationNotFound
	ErrContractMilestoneNotFound = repository.ErrContractMilestoneNotFound
	ErrInvalidScore              = model.ErrInvalidScore
	ErrInvalidMilestoneStatus    = model.ErrInvalidMilestoneStatus
	ErrVendorCILinkNotFound      = model.ErrVendorCILinkNotFound
	ErrVendorCILinkConflict      = model.ErrVendorCILinkConflict
)

// VendorService orchestrates supplier onboarding, contract management, evaluations, and spend analytics
type VendorService interface {
	// Vendor operations
	CreateVendor(ctx context.Context, req *model.CreateVendorRequest) (*model.Vendor, error)
	GetVendorByID(ctx context.Context, id int64) (*model.Vendor, error)
	GetVendorBySlug(ctx context.Context, workspaceID int64, slug string) (*model.Vendor, error)
	UpdateVendor(ctx context.Context, id int64, req *model.UpdateVendorRequest) (*model.Vendor, error)
	DeleteVendor(ctx context.Context, id int64) error
	ListVendors(ctx context.Context, filter model.VendorFilterCriteria) ([]model.Vendor, int, error)

	// Contract operations
	CreateContract(ctx context.Context, req *model.CreateVendorContractRequest) (*model.VendorContract, error)
	GetContractByID(ctx context.Context, id int64) (*model.VendorContract, error)
	GetContractByNumber(ctx context.Context, workspaceID int64, contractNumber string) (*model.VendorContract, error)
	UpdateContract(ctx context.Context, id int64, req *model.UpdateVendorContractRequest) (*model.VendorContract, error)
	DeleteContract(ctx context.Context, id int64) error
	ListContracts(ctx context.Context, filter model.ContractFilterCriteria) ([]model.VendorContract, int, error)
	RenewContract(ctx context.Context, id int64, req *model.RenewContractRequest) (*model.VendorContract, error)

	// Evaluation operations
	CreateEvaluation(ctx context.Context, vendorID int64, evaluatorID string, req *model.CreateVendorEvaluationRequest) (*model.VendorEvaluation, error)
	ListEvaluations(ctx context.Context, vendorID int64) ([]model.VendorEvaluation, error)
	DeleteEvaluation(ctx context.Context, id int64) error

	// Milestone operations
	CreateMilestone(ctx context.Context, contractID int64, req *model.CreateContractMilestoneRequest) (*model.ContractMilestone, error)
	GetMilestoneByID(ctx context.Context, id int64) (*model.ContractMilestone, error)
	UpdateMilestone(ctx context.Context, id int64, req *model.UpdateContractMilestoneRequest) (*model.ContractMilestone, error)
	DeleteMilestone(ctx context.Context, id int64) error
	ListMilestones(ctx context.Context, contractID int64) ([]model.ContractMilestone, error)

	// CI Link operations
	LinkCI(ctx context.Context, vendorID int64, req *model.LinkVendorCIRequest) (*model.VendorCILink, error)
	UnlinkCI(ctx context.Context, vendorID, ciID int64) error
	ListCILinks(ctx context.Context, vendorID int64) ([]model.VendorCILink, error)

	// Analytics
	GetVendorMetrics(ctx context.Context, workspaceID int64) (*model.VendorMetrics, error)
}

type vendorService struct {
	repo repository.VendorRepository
}

// NewVendorService creates a new instance of VendorService
func NewVendorService(repo repository.VendorRepository) VendorService {
	return &vendorService{repo: repo}
}

var errVndNotImplemented = errors.New("not implemented")

func (s *vendorService) CreateVendor(ctx context.Context, req *model.CreateVendorRequest) (*model.Vendor, error) {
	if req == nil {
		return nil, ErrVendorNameRequired
	}
	if err := req.Validate(); err != nil {
		return nil, err
	}

	slug := model.NormalizeSlug(req.Slug)
	if slug == "" {
		slug = model.NormalizeSlug(req.Name)
	}
	if slug == "" {
		return nil, ErrVendorSlugRequired
	}

	// Verify slug uniqueness within workspace
	existing, err := s.repo.GetVendorBySlug(ctx, req.WorkspaceID, slug)
	if err == nil && existing != nil {
		return nil, ErrVendorSlugConflict
	}

	category := strings.ToLower(strings.TrimSpace(req.Category))
	if category == "" {
		category = model.VendorCategoryOther
	}
	tier := strings.ToLower(strings.TrimSpace(req.Tier))
	if tier == "" {
		tier = model.VendorTierCommodity
	}
	status := strings.ToLower(strings.TrimSpace(req.Status))
	if status == "" {
		status = model.VendorStatusActive
	}
	riskLevel := strings.ToLower(strings.TrimSpace(req.RiskLevel))
	if riskLevel == "" {
		riskLevel = model.VendorRiskLow
	}

	vendor := &model.Vendor{
		WorkspaceID:            req.WorkspaceID,
		Name:                   strings.TrimSpace(req.Name),
		Slug:                   slug,
		Category:               category,
		Tier:                   tier,
		Status:                 status,
		RiskLevel:              riskLevel,
		Website:                strings.TrimSpace(req.Website),
		ContactPerson:          strings.TrimSpace(req.ContactPerson),
		ContactEmail:           strings.TrimSpace(req.ContactEmail),
		ContactPhone:           strings.TrimSpace(req.ContactPhone),
		SupportPortalURL:       strings.TrimSpace(req.SupportPortalURL),
		SecurityCertifications: strings.TrimSpace(req.SecurityCertifications),
		Notes:                  strings.TrimSpace(req.Notes),
	}

	if err := s.repo.CreateVendor(ctx, vendor); err != nil {
		return nil, err
	}
	return s.repo.GetVendorByID(ctx, vendor.ID)
}

func (s *vendorService) GetVendorByID(ctx context.Context, id int64) (*model.Vendor, error) {
	if id <= 0 {
		return nil, ErrVendorNotFound
	}
	return s.repo.GetVendorByID(ctx, id)
}

func (s *vendorService) GetVendorBySlug(ctx context.Context, workspaceID int64, slug string) (*model.Vendor, error) {
	cleanSlug := model.NormalizeSlug(slug)
	if cleanSlug == "" {
		return nil, ErrVendorSlugRequired
	}
	return s.repo.GetVendorBySlug(ctx, workspaceID, cleanSlug)
}

func (s *vendorService) UpdateVendor(ctx context.Context, id int64, req *model.UpdateVendorRequest) (*model.Vendor, error) {
	if id <= 0 || req == nil {
		return nil, ErrVendorNotFound
	}

	v, err := s.repo.GetVendorByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)
		if name == "" {
			return nil, ErrVendorNameRequired
		}
		v.Name = name
	}
	if req.Category != nil {
		cat := strings.ToLower(strings.TrimSpace(*req.Category))
		if err := model.ValidateVendorCategory(cat); err != nil {
			return nil, err
		}
		v.Category = cat
	}
	if req.Tier != nil {
		tier := strings.ToLower(strings.TrimSpace(*req.Tier))
		if err := model.ValidateVendorTier(tier); err != nil {
			return nil, err
		}
		v.Tier = tier
	}
	if req.Status != nil {
		status := strings.ToLower(strings.TrimSpace(*req.Status))
		if err := model.ValidateVendorStatus(status); err != nil {
			return nil, err
		}
		v.Status = status
	}
	if req.RiskLevel != nil {
		risk := strings.ToLower(strings.TrimSpace(*req.RiskLevel))
		if err := model.ValidateVendorRisk(risk); err != nil {
			return nil, err
		}
		v.RiskLevel = risk
	}
	if req.Website != nil {
		v.Website = strings.TrimSpace(*req.Website)
	}
	if req.ContactPerson != nil {
		v.ContactPerson = strings.TrimSpace(*req.ContactPerson)
	}
	if req.ContactEmail != nil {
		v.ContactEmail = strings.TrimSpace(*req.ContactEmail)
	}
	if req.ContactPhone != nil {
		v.ContactPhone = strings.TrimSpace(*req.ContactPhone)
	}
	if req.SupportPortalURL != nil {
		v.SupportPortalURL = strings.TrimSpace(*req.SupportPortalURL)
	}
	if req.SecurityCertifications != nil {
		v.SecurityCertifications = strings.TrimSpace(*req.SecurityCertifications)
	}
	if req.Notes != nil {
		v.Notes = strings.TrimSpace(*req.Notes)
	}

	if err := s.repo.UpdateVendor(ctx, v); err != nil {
		return nil, err
	}
	return s.repo.GetVendorByID(ctx, id)
}

func (s *vendorService) DeleteVendor(ctx context.Context, id int64) error {
	if id <= 0 {
		return ErrVendorNotFound
	}
	if _, err := s.repo.GetVendorByID(ctx, id); err != nil {
		return err
	}
	return s.repo.DeleteVendor(ctx, id)
}

func (s *vendorService) ListVendors(ctx context.Context, filter model.VendorFilterCriteria) ([]model.Vendor, int, error) {
	if filter.Limit <= 0 {
		filter.Limit = 50
	}
	if filter.Offset < 0 {
		filter.Offset = 0
	}
	return s.repo.ListVendors(ctx, filter)
}
func (s *vendorService) CreateContract(ctx context.Context, req *model.CreateVendorContractRequest) (*model.VendorContract, error) {
	return nil, errVndNotImplemented
}
func (s *vendorService) GetContractByID(ctx context.Context, id int64) (*model.VendorContract, error) {
	return nil, errVndNotImplemented
}
func (s *vendorService) GetContractByNumber(ctx context.Context, workspaceID int64, contractNumber string) (*model.VendorContract, error) {
	return nil, errVndNotImplemented
}
func (s *vendorService) UpdateContract(ctx context.Context, id int64, req *model.UpdateVendorContractRequest) (*model.VendorContract, error) {
	return nil, errVndNotImplemented
}
func (s *vendorService) DeleteContract(ctx context.Context, id int64) error {
	return errVndNotImplemented
}
func (s *vendorService) ListContracts(ctx context.Context, filter model.ContractFilterCriteria) ([]model.VendorContract, int, error) {
	return nil, 0, errVndNotImplemented
}
func (s *vendorService) RenewContract(ctx context.Context, id int64, req *model.RenewContractRequest) (*model.VendorContract, error) {
	return nil, errVndNotImplemented
}
func (s *vendorService) CreateEvaluation(ctx context.Context, vendorID int64, evaluatorID string, req *model.CreateVendorEvaluationRequest) (*model.VendorEvaluation, error) {
	return nil, errVndNotImplemented
}
func (s *vendorService) ListEvaluations(ctx context.Context, vendorID int64) ([]model.VendorEvaluation, error) {
	return nil, errVndNotImplemented
}
func (s *vendorService) DeleteEvaluation(ctx context.Context, id int64) error {
	return errVndNotImplemented
}
func (s *vendorService) CreateMilestone(ctx context.Context, contractID int64, req *model.CreateContractMilestoneRequest) (*model.ContractMilestone, error) {
	return nil, errVndNotImplemented
}
func (s *vendorService) GetMilestoneByID(ctx context.Context, id int64) (*model.ContractMilestone, error) {
	return nil, errVndNotImplemented
}
func (s *vendorService) UpdateMilestone(ctx context.Context, id int64, req *model.UpdateContractMilestoneRequest) (*model.ContractMilestone, error) {
	return nil, errVndNotImplemented
}
func (s *vendorService) DeleteMilestone(ctx context.Context, id int64) error {
	return errVndNotImplemented
}
func (s *vendorService) ListMilestones(ctx context.Context, contractID int64) ([]model.ContractMilestone, error) {
	return nil, errVndNotImplemented
}
func (s *vendorService) LinkCI(ctx context.Context, vendorID int64, req *model.LinkVendorCIRequest) (*model.VendorCILink, error) {
	return nil, errVndNotImplemented
}
func (s *vendorService) UnlinkCI(ctx context.Context, vendorID, ciID int64) error {
	return errVndNotImplemented
}
func (s *vendorService) ListCILinks(ctx context.Context, vendorID int64) ([]model.VendorCILink, error) {
	return nil, errVndNotImplemented
}
func (s *vendorService) GetVendorMetrics(ctx context.Context, workspaceID int64) (*model.VendorMetrics, error) {
	return nil, errVndNotImplemented
}
