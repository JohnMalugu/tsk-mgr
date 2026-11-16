package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/JohnMalugu/tsk-mgr-api/internal/model"
)

var (
	ErrVendorNotFound            = model.ErrVendorNotFound
	ErrVendorSlugConflict        = model.ErrVendorSlugConflict
	ErrContractNotFound          = model.ErrContractNotFound
	ErrContractNumberConflict     = model.ErrContractNumberConflict
	ErrEvaluationNotFound        = model.ErrEvaluationNotFound
	ErrContractMilestoneNotFound = model.ErrMilestoneNotFound
	ErrVendorCILinkNotFound      = model.ErrVendorCILinkNotFound
	ErrVendorCILinkConflict      = model.ErrVendorCILinkConflict
)

// VendorRepository defines data access methods for vendors, contracts, evaluations, and milestones
type VendorRepository interface {
	// Vendor operations
	CreateVendor(ctx context.Context, vendor *model.Vendor) error
	GetVendorByID(ctx context.Context, id int64) (*model.Vendor, error)
	GetVendorBySlug(ctx context.Context, workspaceID int64, slug string) (*model.Vendor, error)
	UpdateVendor(ctx context.Context, vendor *model.Vendor) error
	DeleteVendor(ctx context.Context, id int64) error
	ListVendors(ctx context.Context, filter model.VendorFilterCriteria) ([]model.Vendor, int, error)

	// Contract operations
	CreateContract(ctx context.Context, contract *model.VendorContract) error
	GetContractByID(ctx context.Context, id int64) (*model.VendorContract, error)
	GetContractByNumber(ctx context.Context, workspaceID int64, contractNumber string) (*model.VendorContract, error)
	UpdateContract(ctx context.Context, contract *model.VendorContract) error
	DeleteContract(ctx context.Context, id int64) error
	ListContracts(ctx context.Context, filter model.ContractFilterCriteria) ([]model.VendorContract, int, error)
	RenewContract(ctx context.Context, id int64, newEndDate time.Time, newTotalVal, newAnnualCost *float64, autoRenew *bool, noticeDays *int, notes string) error
	GetNextContractNumberSequence(ctx context.Context, workspaceID int64) (int, error)

	// Evaluation operations
	CreateEvaluation(ctx context.Context, eval *model.VendorEvaluation) error
	ListEvaluations(ctx context.Context, vendorID int64) ([]model.VendorEvaluation, error)
	DeleteEvaluation(ctx context.Context, id int64) error

	// Milestone operations
	CreateMilestone(ctx context.Context, milestone *model.ContractMilestone) error
	GetMilestoneByID(ctx context.Context, id int64) (*model.ContractMilestone, error)
	UpdateMilestone(ctx context.Context, milestone *model.ContractMilestone) error
	DeleteMilestone(ctx context.Context, id int64) error
	ListMilestones(ctx context.Context, contractID int64) ([]model.ContractMilestone, error)

	// CI Link operations
	LinkCI(ctx context.Context, link *model.VendorCILink) error
	UnlinkCI(ctx context.Context, vendorID, ciID int64) error
	ListCILinks(ctx context.Context, vendorID int64) ([]model.VendorCILink, error)

	// Analytics
	GetVendorMetrics(ctx context.Context, workspaceID int64) (*model.VendorMetrics, error)
}

type sqlVendorRepository struct {
	db *sql.DB
}

// NewVendorRepository creates a new instance of VendorRepository
func NewVendorRepository(db *sql.DB) VendorRepository {
	return &sqlVendorRepository{db: db}
}

var errVendorNotImplemented = errors.New("not implemented")

func (r *sqlVendorRepository) CreateVendor(ctx context.Context, vendor *model.Vendor) error {
	return errVendorNotImplemented
}
func (r *sqlVendorRepository) GetVendorByID(ctx context.Context, id int64) (*model.Vendor, error) {
	return nil, errVendorNotImplemented
}
func (r *sqlVendorRepository) GetVendorBySlug(ctx context.Context, workspaceID int64, slug string) (*model.Vendor, error) {
	return nil, errVendorNotImplemented
}
func (r *sqlVendorRepository) UpdateVendor(ctx context.Context, vendor *model.Vendor) error {
	return errVendorNotImplemented
}
func (r *sqlVendorRepository) DeleteVendor(ctx context.Context, id int64) error {
	return errVendorNotImplemented
}
func (r *sqlVendorRepository) ListVendors(ctx context.Context, filter model.VendorFilterCriteria) ([]model.Vendor, int, error) {
	return nil, 0, errVendorNotImplemented
}
func (r *sqlVendorRepository) CreateContract(ctx context.Context, contract *model.VendorContract) error {
	return errVendorNotImplemented
}
func (r *sqlVendorRepository) GetContractByID(ctx context.Context, id int64) (*model.VendorContract, error) {
	return nil, errVendorNotImplemented
}
func (r *sqlVendorRepository) GetContractByNumber(ctx context.Context, workspaceID int64, contractNumber string) (*model.VendorContract, error) {
	return nil, errVendorNotImplemented
}
func (r *sqlVendorRepository) UpdateContract(ctx context.Context, contract *model.VendorContract) error {
	return errVendorNotImplemented
}
func (r *sqlVendorRepository) DeleteContract(ctx context.Context, id int64) error {
	return errVendorNotImplemented
}
func (r *sqlVendorRepository) ListContracts(ctx context.Context, filter model.ContractFilterCriteria) ([]model.VendorContract, int, error) {
	return nil, 0, errVendorNotImplemented
}
func (r *sqlVendorRepository) RenewContract(ctx context.Context, id int64, newEndDate time.Time, newTotalVal, newAnnualCost *float64, autoRenew *bool, noticeDays *int, notes string) error {
	return errVendorNotImplemented
}
func (r *sqlVendorRepository) GetNextContractNumberSequence(ctx context.Context, workspaceID int64) (int, error) {
	return 0, errVendorNotImplemented
}
func (r *sqlVendorRepository) CreateEvaluation(ctx context.Context, eval *model.VendorEvaluation) error {
	return errVendorNotImplemented
}
func (r *sqlVendorRepository) ListEvaluations(ctx context.Context, vendorID int64) ([]model.VendorEvaluation, error) {
	return nil, errVendorNotImplemented
}
func (r *sqlVendorRepository) DeleteEvaluation(ctx context.Context, id int64) error {
	return errVendorNotImplemented
}
func (r *sqlVendorRepository) CreateMilestone(ctx context.Context, milestone *model.ContractMilestone) error {
	return errVendorNotImplemented
}
func (r *sqlVendorRepository) GetMilestoneByID(ctx context.Context, id int64) (*model.ContractMilestone, error) {
	return nil, errVendorNotImplemented
}
func (r *sqlVendorRepository) UpdateMilestone(ctx context.Context, milestone *model.ContractMilestone) error {
	return errVendorNotImplemented
}
func (r *sqlVendorRepository) DeleteMilestone(ctx context.Context, id int64) error {
	return errVendorNotImplemented
}
func (r *sqlVendorRepository) ListMilestones(ctx context.Context, contractID int64) ([]model.ContractMilestone, error) {
	return nil, errVendorNotImplemented
}
func (r *sqlVendorRepository) LinkCI(ctx context.Context, link *model.VendorCILink) error {
	return errVendorNotImplemented
}
func (r *sqlVendorRepository) UnlinkCI(ctx context.Context, vendorID, ciID int64) error {
	return errVendorNotImplemented
}
func (r *sqlVendorRepository) ListCILinks(ctx context.Context, vendorID int64) ([]model.VendorCILink, error) {
	return nil, errVendorNotImplemented
}
func (r *sqlVendorRepository) GetVendorMetrics(ctx context.Context, workspaceID int64) (*model.VendorMetrics, error) {
	return nil, errVendorNotImplemented
}
