package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
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

func (r *sqlVendorRepository) scanVendor(row interface{ Scan(...any) error }) (*model.Vendor, error) {
	var v model.Vendor
	var activeContracts sql.NullInt64
	var totalSpend, avgScore sql.NullFloat64

	err := row.Scan(
		&v.ID,
		&v.WorkspaceID,
		&v.Name,
		&v.Slug,
		&v.Category,
		&v.Tier,
		&v.Status,
		&v.RiskLevel,
		&v.Website,
		&v.ContactPerson,
		&v.ContactEmail,
		&v.ContactPhone,
		&v.SupportPortalURL,
		&v.SecurityCertifications,
		&v.Notes,
		&v.CreatedAt,
		&v.UpdatedAt,
		&activeContracts,
		&totalSpend,
		&avgScore,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrVendorNotFound
		}
		return nil, err
	}
	if activeContracts.Valid {
		v.ActiveContractsCount = int(activeContracts.Int64)
	}
	if totalSpend.Valid {
		v.TotalAnnualSpend = totalSpend.Float64
	}
	if avgScore.Valid {
		v.AverageEvaluationScore = avgScore.Float64
	}
	return &v, nil
}

const vendorSelectFields = `
	v.id, v.workspace_id, v.name, v.slug, v.category, v.tier, v.status, v.risk_level,
	v.website, v.contact_person, v.contact_email, v.contact_phone, v.support_portal_url,
	v.security_certifications, v.notes, v.created_at, v.updated_at,
	(SELECT COUNT(*) FROM vendor_contracts c WHERE c.vendor_id = v.id AND c.status = 'active') AS active_contracts_count,
	(SELECT COALESCE(SUM(c.annual_recurring_cost), 0.0) FROM vendor_contracts c WHERE c.vendor_id = v.id AND c.status = 'active') AS total_annual_spend,
	(SELECT COALESCE(AVG(e.overall_score), 0.0) FROM vendor_evaluations e WHERE e.vendor_id = v.id) AS average_evaluation_score
`

const vendorFromClause = `FROM vendors v`

func (r *sqlVendorRepository) CreateVendor(ctx context.Context, vendor *model.Vendor) error {
	now := time.Now().UTC()
	vendor.CreatedAt = now
	vendor.UpdatedAt = now

	// Check slug uniqueness
	var count int
	_ = r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM vendors WHERE workspace_id = ? AND slug = ?", vendor.WorkspaceID, vendor.Slug).Scan(&count)
	if count > 0 {
		return ErrVendorSlugConflict
	}

	query := `
	INSERT INTO vendors (
		workspace_id, name, slug, category, tier, status, risk_level,
		website, contact_person, contact_email, contact_phone, support_portal_url,
		security_certifications, notes, created_at, updated_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);`

	res, err := r.db.ExecContext(ctx, query,
		vendor.WorkspaceID, vendor.Name, vendor.Slug, vendor.Category, vendor.Tier, vendor.Status, vendor.RiskLevel,
		vendor.Website, vendor.ContactPerson, vendor.ContactEmail, vendor.ContactPhone, vendor.SupportPortalURL,
		vendor.SecurityCertifications, vendor.Notes, vendor.CreatedAt, vendor.UpdatedAt,
	)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE") || strings.Contains(err.Error(), "constraint") {
			return ErrVendorSlugConflict
		}
		return err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return err
	}
	vendor.ID = id
	return nil
}

func (r *sqlVendorRepository) GetVendorByID(ctx context.Context, id int64) (*model.Vendor, error) {
	query := fmt.Sprintf("SELECT %s %s WHERE v.id = ?", vendorSelectFields, vendorFromClause)
	return r.scanVendor(r.db.QueryRowContext(ctx, query, id))
}

func (r *sqlVendorRepository) GetVendorBySlug(ctx context.Context, workspaceID int64, slug string) (*model.Vendor, error) {
	query := fmt.Sprintf("SELECT %s %s WHERE v.workspace_id = ? AND v.slug = ?", vendorSelectFields, vendorFromClause)
	return r.scanVendor(r.db.QueryRowContext(ctx, query, workspaceID, slug))
}

func (r *sqlVendorRepository) UpdateVendor(ctx context.Context, vendor *model.Vendor) error {
	vendor.UpdatedAt = time.Now().UTC()

	query := `
	UPDATE vendors SET
		name = ?, category = ?, tier = ?, status = ?, risk_level = ?,
		website = ?, contact_person = ?, contact_email = ?, contact_phone = ?,
		support_portal_url = ?, security_certifications = ?, notes = ?, updated_at = ?
	WHERE id = ?;`

	res, err := r.db.ExecContext(ctx, query,
		vendor.Name, vendor.Category, vendor.Tier, vendor.Status, vendor.RiskLevel,
		vendor.Website, vendor.ContactPerson, vendor.ContactEmail, vendor.ContactPhone,
		vendor.SupportPortalURL, vendor.SecurityCertifications, vendor.Notes, vendor.UpdatedAt, vendor.ID,
	)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrVendorNotFound
	}
	return nil
}

func (r *sqlVendorRepository) DeleteVendor(ctx context.Context, id int64) error {
	res, err := r.db.ExecContext(ctx, "DELETE FROM vendors WHERE id = ?", id)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrVendorNotFound
	}
	return nil
}

func (r *sqlVendorRepository) ListVendors(ctx context.Context, filter model.VendorFilterCriteria) ([]model.Vendor, int, error) {
	var whereClauses []string
	var args []interface{}

	if filter.WorkspaceID > 0 {
		whereClauses = append(whereClauses, "v.workspace_id = ?")
		args = append(args, filter.WorkspaceID)
	}
	if filter.Category != "" {
		whereClauses = append(whereClauses, "v.category = ?")
		args = append(args, filter.Category)
	}
	if filter.Tier != "" {
		whereClauses = append(whereClauses, "v.tier = ?")
		args = append(args, filter.Tier)
	}
	if filter.Status != "" {
		whereClauses = append(whereClauses, "v.status = ?")
		args = append(args, filter.Status)
	}
	if filter.RiskLevel != "" {
		whereClauses = append(whereClauses, "v.risk_level = ?")
		args = append(args, filter.RiskLevel)
	}
	if filter.Search != "" {
		searchPattern := "%" + filter.Search + "%"
		whereClauses = append(whereClauses, "(v.name LIKE ? OR v.slug LIKE ? OR v.contact_person LIKE ? OR v.contact_email LIKE ?)")
		args = append(args, searchPattern, searchPattern, searchPattern, searchPattern)
	}

	whereSQL := ""
	if len(whereClauses) > 0 {
		whereSQL = "WHERE " + strings.Join(whereClauses, " AND ")
	}

	countQuery := fmt.Sprintf("SELECT COUNT(*) %s %s", vendorFromClause, whereSQL)
	var total int
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := fmt.Sprintf("SELECT %s %s %s ORDER BY v.id DESC", vendorSelectFields, vendorFromClause, whereSQL)
	if filter.Limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", filter.Limit)
		if filter.Offset > 0 {
			query += fmt.Sprintf(" OFFSET %d", filter.Offset)
		}
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var vendors []model.Vendor
	for rows.Next() {
		v, err := r.scanVendor(rows)
		if err != nil {
			return nil, 0, err
		}
		vendors = append(vendors, *v)
	}
	return vendors, total, rows.Err()
}
func (r *sqlVendorRepository) scanContract(row interface{ Scan(...any) error }) (*model.VendorContract, error) {
	var c model.VendorContract
	var autoRenewInt int
	var vendorName, ownerUserName sql.NullString
	var milestonesCount sql.NullInt64

	err := row.Scan(
		&c.ID,
		&c.WorkspaceID,
		&c.VendorID,
		&c.ContractNumber,
		&c.Title,
		&c.ContractType,
		&c.Status,
		&c.TotalValue,
		&c.AnnualRecurringCost,
		&c.Currency,
		&c.StartDate,
		&c.EndDate,
		&autoRenewInt,
		&c.RenewalNoticeDays,
		&c.OwnerUserID,
		&c.PaymentTerms,
		&c.DocumentURL,
		&c.Notes,
		&c.CreatedAt,
		&c.UpdatedAt,
		&vendorName,
		&ownerUserName,
		&milestonesCount,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrContractNotFound
		}
		return nil, err
	}
	c.AutoRenew = (autoRenewInt == 1)
	if vendorName.Valid {
		c.VendorName = vendorName.String
	}
	if ownerUserName.Valid {
		c.OwnerUserName = ownerUserName.String
	}
	if milestonesCount.Valid {
		c.MilestonesCount = int(milestonesCount.Int64)
	}

	now := time.Now().UTC()
	days := int(c.EndDate.Sub(now).Hours() / 24)
	c.DaysUntilExpiration = days

	return &c, nil
}

const contractSelectFields = `
	c.id, c.workspace_id, c.vendor_id, c.contract_number, c.title, c.contract_type,
	c.status, c.total_value, c.annual_recurring_cost, c.currency, c.start_date, c.end_date,
	c.auto_renew, c.renewal_notice_days, c.owner_user_id, c.payment_terms, c.document_url,
	c.notes, c.created_at, c.updated_at,
	COALESCE(v.name, '') AS vendor_name,
	COALESCE(u.username, '') AS owner_user_name,
	(SELECT COUNT(*) FROM contract_milestones m WHERE m.contract_id = c.id) AS milestones_count
`

const contractFromClause = `
	FROM vendor_contracts c
	JOIN vendors v ON c.vendor_id = v.id
	LEFT JOIN users u ON c.owner_user_id = u.id
`

func (r *sqlVendorRepository) CreateContract(ctx context.Context, contract *model.VendorContract) error {
	now := time.Now().UTC()
	contract.CreatedAt = now
	contract.UpdatedAt = now

	// Check contract number uniqueness
	var count int
	_ = r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM vendor_contracts WHERE workspace_id = ? AND contract_number = ?", contract.WorkspaceID, contract.ContractNumber).Scan(&count)
	if count > 0 {
		return ErrContractNumberConflict
	}

	autoRenewInt := 0
	if contract.AutoRenew {
		autoRenewInt = 1
	}

	query := `
	INSERT INTO vendor_contracts (
		workspace_id, vendor_id, contract_number, title, contract_type, status,
		total_value, annual_recurring_cost, currency, start_date, end_date,
		auto_renew, renewal_notice_days, owner_user_id, payment_terms, document_url,
		notes, created_at, updated_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);`

	res, err := r.db.ExecContext(ctx, query,
		contract.WorkspaceID, contract.VendorID, contract.ContractNumber, contract.Title, contract.ContractType, contract.Status,
		contract.TotalValue, contract.AnnualRecurringCost, contract.Currency, contract.StartDate, contract.EndDate,
		autoRenewInt, contract.RenewalNoticeDays, contract.OwnerUserID, contract.PaymentTerms, contract.DocumentURL,
		contract.Notes, contract.CreatedAt, contract.UpdatedAt,
	)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE") || strings.Contains(err.Error(), "constraint") {
			return ErrContractNumberConflict
		}
		return err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return err
	}
	contract.ID = id
	return nil
}

func (r *sqlVendorRepository) GetContractByID(ctx context.Context, id int64) (*model.VendorContract, error) {
	query := fmt.Sprintf("SELECT %s %s WHERE c.id = ?", contractSelectFields, contractFromClause)
	return r.scanContract(r.db.QueryRowContext(ctx, query, id))
}

func (r *sqlVendorRepository) GetContractByNumber(ctx context.Context, workspaceID int64, contractNumber string) (*model.VendorContract, error) {
	query := fmt.Sprintf("SELECT %s %s WHERE c.workspace_id = ? AND c.contract_number = ?", contractSelectFields, contractFromClause)
	return r.scanContract(r.db.QueryRowContext(ctx, query, workspaceID, contractNumber))
}

func (r *sqlVendorRepository) UpdateContract(ctx context.Context, contract *model.VendorContract) error {
	contract.UpdatedAt = time.Now().UTC()
	autoRenewInt := 0
	if contract.AutoRenew {
		autoRenewInt = 1
	}

	query := `
	UPDATE vendor_contracts SET
		title = ?, contract_type = ?, status = ?, total_value = ?,
		annual_recurring_cost = ?, currency = ?, start_date = ?, end_date = ?,
		auto_renew = ?, renewal_notice_days = ?, owner_user_id = ?,
		payment_terms = ?, document_url = ?, notes = ?, updated_at = ?
	WHERE id = ?;`

	res, err := r.db.ExecContext(ctx, query,
		contract.Title, contract.ContractType, contract.Status, contract.TotalValue,
		contract.AnnualRecurringCost, contract.Currency, contract.StartDate, contract.EndDate,
		autoRenewInt, contract.RenewalNoticeDays, contract.OwnerUserID,
		contract.PaymentTerms, contract.DocumentURL, contract.Notes, contract.UpdatedAt, contract.ID,
	)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrContractNotFound
	}
	return nil
}

func (r *sqlVendorRepository) DeleteContract(ctx context.Context, id int64) error {
	res, err := r.db.ExecContext(ctx, "DELETE FROM vendor_contracts WHERE id = ?", id)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrContractNotFound
	}
	return nil
}

func (r *sqlVendorRepository) ListContracts(ctx context.Context, filter model.ContractFilterCriteria) ([]model.VendorContract, int, error) {
	var whereClauses []string
	var args []interface{}

	if filter.WorkspaceID > 0 {
		whereClauses = append(whereClauses, "c.workspace_id = ?")
		args = append(args, filter.WorkspaceID)
	}
	if filter.VendorID > 0 {
		whereClauses = append(whereClauses, "c.vendor_id = ?")
		args = append(args, filter.VendorID)
	}
	if filter.ContractType != "" {
		whereClauses = append(whereClauses, "c.contract_type = ?")
		args = append(args, filter.ContractType)
	}
	if filter.Status != "" {
		whereClauses = append(whereClauses, "c.status = ?")
		args = append(args, filter.Status)
	}
	if filter.OwnerUserID != "" {
		whereClauses = append(whereClauses, "c.owner_user_id = ?")
		args = append(args, filter.OwnerUserID)
	}
	if filter.Search != "" {
		searchPattern := "%" + filter.Search + "%"
		whereClauses = append(whereClauses, "(c.title LIKE ? OR c.contract_number LIKE ? OR v.name LIKE ?)")
		args = append(args, searchPattern, searchPattern, searchPattern)
	}
	if filter.ExpiringDays != nil && *filter.ExpiringDays > 0 {
		now := time.Now().UTC()
		future := now.AddDate(0, 0, *filter.ExpiringDays)
		whereClauses = append(whereClauses, "c.end_date >= ? AND c.end_date <= ?")
		args = append(args, now, future)
	}

	whereSQL := ""
	if len(whereClauses) > 0 {
		whereSQL = "WHERE " + strings.Join(whereClauses, " AND ")
	}

	countQuery := fmt.Sprintf("SELECT COUNT(*) %s %s", contractFromClause, whereSQL)
	var total int
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := fmt.Sprintf("SELECT %s %s %s ORDER BY c.id DESC", contractSelectFields, contractFromClause, whereSQL)
	if filter.Limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", filter.Limit)
		if filter.Offset > 0 {
			query += fmt.Sprintf(" OFFSET %d", filter.Offset)
		}
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var contracts []model.VendorContract
	for rows.Next() {
		c, err := r.scanContract(rows)
		if err != nil {
			return nil, 0, err
		}
		contracts = append(contracts, *c)
	}
	return contracts, total, rows.Err()
}

func (r *sqlVendorRepository) RenewContract(ctx context.Context, id int64, newEndDate time.Time, newTotalVal, newAnnualCost *float64, autoRenew *bool, noticeDays *int, notes string) error {
	now := time.Now().UTC()

	query := `
	UPDATE vendor_contracts SET
		end_date = ?,
		status = 'active',
		total_value = CASE WHEN ? >= 0.0 THEN ? ELSE total_value END,
		annual_recurring_cost = CASE WHEN ? >= 0.0 THEN ? ELSE annual_recurring_cost END,
		auto_renew = CASE WHEN ? >= 0 THEN ? ELSE auto_renew END,
		renewal_notice_days = CASE WHEN ? > 0 THEN ? ELSE renewal_notice_days END,
		notes = CASE WHEN ? != '' THEN ? ELSE notes END,
		updated_at = ?
	WHERE id = ?;`

	valParam := -1.0
	if newTotalVal != nil {
		valParam = *newTotalVal
	}
	costParam := -1.0
	if newAnnualCost != nil {
		costParam = *newAnnualCost
	}
	autoRenewParam := -1
	if autoRenew != nil {
		if *autoRenew {
			autoRenewParam = 1
		} else {
			autoRenewParam = 0
		}
	}
	noticeParam := -1
	if noticeDays != nil {
		noticeParam = *noticeDays
	}

	res, err := r.db.ExecContext(ctx, query,
		newEndDate,
		valParam, valParam,
		costParam, costParam,
		autoRenewParam, autoRenewParam,
		noticeParam, noticeParam,
		notes, notes,
		now, id,
	)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrContractNotFound
	}
	return nil
}

func (r *sqlVendorRepository) GetNextContractNumberSequence(ctx context.Context, workspaceID int64) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM vendor_contracts WHERE workspace_id = ?", workspaceID).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count + 1, nil
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
