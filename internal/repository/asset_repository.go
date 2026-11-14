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
	ErrAssetNotFound             = model.ErrAssetNotFound
	ErrAssetTagConflict          = model.ErrAssetTagConflict
	ErrLicenseNotFound           = model.ErrLicenseNotFound
	ErrMaintenanceRecordNotFound = model.ErrMaintenanceRecordNotFound
	ErrLicenseAllocNotFound      = model.ErrAllocationNotFound
	ErrNoSeatsAvailable          = model.ErrNoSeatsAvailable
	ErrAlreadyAllocated          = model.ErrAlreadyAllocated
)

// ITAssetRepository handles persistence and queries for assets, maintenance, licenses, and audits
type ITAssetRepository interface {
	// IT Asset operations
	CreateAsset(ctx context.Context, asset *model.ITAsset) error
	GetAssetByID(ctx context.Context, id int64) (*model.ITAsset, error)
	GetAssetByTag(ctx context.Context, workspaceID int64, assetTag string) (*model.ITAsset, error)
	UpdateAsset(ctx context.Context, asset *model.ITAsset) error
	DeleteAsset(ctx context.Context, id int64) error
	ListAssets(ctx context.Context, filter model.AssetFilterCriteria) ([]model.ITAsset, int, error)
	GetNextAssetTagSequence(ctx context.Context, workspaceID int64) (int, error)

	// Assignment and status
	AssignAsset(ctx context.Context, assetID int64, userID *string, dept, location string, assignedAt time.Time, notes string) error
	CheckinAsset(ctx context.Context, assetID int64, location string, notes string) error
	UpdateAssetStatus(ctx context.Context, assetID int64, status string, notes string) error

	// Maintenance
	CreateMaintenanceRecord(ctx context.Context, record *model.AssetMaintenanceRecord) error
	GetMaintenanceRecordByID(ctx context.Context, id int64) (*model.AssetMaintenanceRecord, error)
	UpdateMaintenanceRecord(ctx context.Context, record *model.AssetMaintenanceRecord) error
	ListMaintenanceRecords(ctx context.Context, assetID int64) ([]model.AssetMaintenanceRecord, error)
	DeleteMaintenanceRecord(ctx context.Context, id int64) error

	// Software Licenses
	CreateSoftwareLicense(ctx context.Context, license *model.SoftwareLicense) error
	GetSoftwareLicenseByID(ctx context.Context, id int64) (*model.SoftwareLicense, error)
	UpdateSoftwareLicense(ctx context.Context, license *model.SoftwareLicense) error
	DeleteSoftwareLicense(ctx context.Context, id int64) error
	ListSoftwareLicenses(ctx context.Context, filter model.LicenseFilterCriteria) ([]model.SoftwareLicense, int, error)

	// License Allocations
	CreateLicenseAllocation(ctx context.Context, alloc *model.LicenseAllocation) error
	GetLicenseAllocationByID(ctx context.Context, id int64) (*model.LicenseAllocation, error)
	DeleteLicenseAllocation(ctx context.Context, id int64) error
	ListLicenseAllocations(ctx context.Context, licenseID int64) ([]model.LicenseAllocation, error)

	// Audits and Metrics
	CreateAuditLog(ctx context.Context, audit *model.AssetAuditLog) error
	ListAuditLogs(ctx context.Context, workspaceID int64, assetID *int64, limit, offset int) ([]model.AssetAuditLog, error)
	GetAssetMetrics(ctx context.Context, workspaceID int64) (*model.AssetMetrics, error)
}

type sqlITAssetRepository struct {
	db *sql.DB
}

// NewITAssetRepository creates a new instance of ITAssetRepository
func NewITAssetRepository(db *sql.DB) ITAssetRepository {
	return &sqlITAssetRepository{db: db}
}

var errNotImplemented = errors.New("not implemented")

func (r *sqlITAssetRepository) scanAsset(row interface{ Scan(...any) error }) (*model.ITAsset, error) {
	var a model.ITAsset
	var purchaseDate, warrantyExpiry, assignedAt sql.NullTime
	var assignedUserID sql.NullString
	var ciID sql.NullInt64
	var assignedUserName, assignedUserEmail, ciName sql.NullString

	err := row.Scan(
		&a.ID,
		&a.WorkspaceID,
		&a.AssetTag,
		&a.Name,
		&a.Category,
		&a.Status,
		&a.Make,
		&a.Model,
		&a.SerialNumber,
		&purchaseDate,
		&a.PurchaseCost,
		&a.Currency,
		&warrantyExpiry,
		&a.ResidualValue,
		&a.UsefulLifeMonths,
		&a.DepreciationMethod,
		&a.Location,
		&assignedUserID,
		&a.AssignedToDept,
		&assignedAt,
		&ciID,
		&a.Notes,
		&a.CreatedAt,
		&a.UpdatedAt,
		&assignedUserName,
		&assignedUserEmail,
		&ciName,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrAssetNotFound
		}
		return nil, err
	}
	if purchaseDate.Valid {
		t := purchaseDate.Time
		a.PurchaseDate = &t
	}
	if warrantyExpiry.Valid {
		t := warrantyExpiry.Time
		a.WarrantyExpiry = &t
	}
	if assignedAt.Valid {
		t := assignedAt.Time
		a.AssignedAt = &t
	}
	if assignedUserID.Valid {
		u := assignedUserID.String
		a.AssignedToUserID = &u
	}
	if ciID.Valid {
		id := ciID.Int64
		a.CIID = &id
	}
	if assignedUserName.Valid {
		a.AssignedUserName = assignedUserName.String
	}
	if assignedUserEmail.Valid {
		a.AssignedUserEmail = assignedUserEmail.String
	}
	if ciName.Valid {
		a.CIName = ciName.String
	}

	return &a, nil
}

const assetSelectFields = `
	a.id, a.workspace_id, a.asset_tag, a.name, a.category, a.status, a.make, a.model, a.serial_number,
	a.purchase_date, a.purchase_cost, a.currency, a.warranty_expiry, a.residual_value, a.useful_life_months,
	a.depreciation_method, a.location, a.assigned_to_user_id, a.assigned_to_dept, a.assigned_at, a.ci_id,
	a.notes, a.created_at, a.updated_at,
	COALESCE(u.username, '') AS assigned_user_name,
	COALESCE(u.email, '') AS assigned_user_email,
	COALESCE(ci.name, '') AS ci_name
`

const assetFromClause = `
	FROM it_assets a
	LEFT JOIN users u ON a.assigned_to_user_id = u.id
	LEFT JOIN configuration_items ci ON a.ci_id = ci.id
`

func (r *sqlITAssetRepository) CreateAsset(ctx context.Context, asset *model.ITAsset) error {
	now := time.Now().UTC()
	asset.CreatedAt = now
	asset.UpdatedAt = now

	// Check asset_tag uniqueness in workspace
	var exists int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM it_assets WHERE workspace_id = ? AND asset_tag = ?", asset.WorkspaceID, asset.AssetTag).Scan(&exists)
	if err == nil && exists > 0 {
		return ErrAssetTagConflict
	}

	query := `
	INSERT INTO it_assets (
		workspace_id, asset_tag, name, category, status, make, model, serial_number,
		purchase_date, purchase_cost, currency, warranty_expiry, residual_value, useful_life_months,
		depreciation_method, location, assigned_to_user_id, assigned_to_dept, assigned_at, ci_id,
		notes, created_at, updated_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);`

	res, err := r.db.ExecContext(ctx, query,
		asset.WorkspaceID, asset.AssetTag, asset.Name, asset.Category, asset.Status, asset.Make, asset.Model, asset.SerialNumber,
		asset.PurchaseDate, asset.PurchaseCost, asset.Currency, asset.WarrantyExpiry, asset.ResidualValue, asset.UsefulLifeMonths,
		asset.DepreciationMethod, asset.Location, asset.AssignedToUserID, asset.AssignedToDept, asset.AssignedAt, asset.CIID,
		asset.Notes, asset.CreatedAt, asset.UpdatedAt,
	)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE") || strings.Contains(err.Error(), "constraint") {
			return ErrAssetTagConflict
		}
		return err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return err
	}
	asset.ID = id
	return nil
}

func (r *sqlITAssetRepository) GetAssetByID(ctx context.Context, id int64) (*model.ITAsset, error) {
	query := fmt.Sprintf("SELECT %s %s WHERE a.id = ?", assetSelectFields, assetFromClause)
	return r.scanAsset(r.db.QueryRowContext(ctx, query, id))
}

func (r *sqlITAssetRepository) GetAssetByTag(ctx context.Context, workspaceID int64, assetTag string) (*model.ITAsset, error) {
	query := fmt.Sprintf("SELECT %s %s WHERE a.workspace_id = ? AND a.asset_tag = ?", assetSelectFields, assetFromClause)
	return r.scanAsset(r.db.QueryRowContext(ctx, query, workspaceID, assetTag))
}

func (r *sqlITAssetRepository) UpdateAsset(ctx context.Context, asset *model.ITAsset) error {
	asset.UpdatedAt = time.Now().UTC()

	query := `
	UPDATE it_assets SET
		name = ?, category = ?, status = ?, make = ?, model = ?, serial_number = ?,
		purchase_date = ?, purchase_cost = ?, currency = ?, warranty_expiry = ?,
		residual_value = ?, useful_life_months = ?, depreciation_method = ?, location = ?,
		assigned_to_user_id = ?, assigned_to_dept = ?, assigned_at = ?, ci_id = ?,
		notes = ?, updated_at = ?
	WHERE id = ?;`

	res, err := r.db.ExecContext(ctx, query,
		asset.Name, asset.Category, asset.Status, asset.Make, asset.Model, asset.SerialNumber,
		asset.PurchaseDate, asset.PurchaseCost, asset.Currency, asset.WarrantyExpiry,
		asset.ResidualValue, asset.UsefulLifeMonths, asset.DepreciationMethod, asset.Location,
		asset.AssignedToUserID, asset.AssignedToDept, asset.AssignedAt, asset.CIID,
		asset.Notes, asset.UpdatedAt, asset.ID,
	)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrAssetNotFound
	}
	return nil
}

func (r *sqlITAssetRepository) DeleteAsset(ctx context.Context, id int64) error {
	res, err := r.db.ExecContext(ctx, "DELETE FROM it_assets WHERE id = ?", id)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrAssetNotFound
	}
	return nil
}

func (r *sqlITAssetRepository) ListAssets(ctx context.Context, filter model.AssetFilterCriteria) ([]model.ITAsset, int, error) {
	var whereClauses []string
	var args []interface{}

	if filter.WorkspaceID > 0 {
		whereClauses = append(whereClauses, "a.workspace_id = ?")
		args = append(args, filter.WorkspaceID)
	}
	if filter.Category != "" {
		whereClauses = append(whereClauses, "a.category = ?")
		args = append(args, filter.Category)
	}
	if filter.Status != "" {
		whereClauses = append(whereClauses, "a.status = ?")
		args = append(args, filter.Status)
	}
	if filter.AssignedToUserID != "" {
		whereClauses = append(whereClauses, "a.assigned_to_user_id = ?")
		args = append(args, filter.AssignedToUserID)
	}
	if filter.Location != "" {
		whereClauses = append(whereClauses, "a.location LIKE ?")
		args = append(args, "%"+filter.Location+"%")
	}
	if filter.Search != "" {
		searchPattern := "%" + filter.Search + "%"
		whereClauses = append(whereClauses, "(a.name LIKE ? OR a.asset_tag LIKE ? OR a.serial_number LIKE ? OR a.make LIKE ? OR a.model LIKE ?)")
		args = append(args, searchPattern, searchPattern, searchPattern, searchPattern, searchPattern)
	}
	if filter.WarrantyExpiringDays != nil && *filter.WarrantyExpiringDays > 0 {
		now := time.Now().UTC()
		future := now.AddDate(0, 0, *filter.WarrantyExpiringDays)
		whereClauses = append(whereClauses, "a.warranty_expiry IS NOT NULL AND a.warranty_expiry >= ? AND a.warranty_expiry <= ?")
		args = append(args, now, future)
	}

	whereSQL := ""
	if len(whereClauses) > 0 {
		whereSQL = "WHERE " + strings.Join(whereClauses, " AND ")
	}

	countQuery := fmt.Sprintf("SELECT COUNT(*) %s %s", assetFromClause, whereSQL)
	var total int
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := fmt.Sprintf("SELECT %s %s %s ORDER BY a.id DESC", assetSelectFields, assetFromClause, whereSQL)
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

	var assets []model.ITAsset
	for rows.Next() {
		a, err := r.scanAsset(rows)
		if err != nil {
			return nil, 0, err
		}
		assets = append(assets, *a)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return assets, total, nil
}

func (r *sqlITAssetRepository) GetNextAssetTagSequence(ctx context.Context, workspaceID int64) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM it_assets WHERE workspace_id = ?", workspaceID).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count + 1, nil
}
func (r *sqlITAssetRepository) AssignAsset(ctx context.Context, assetID int64, userID *string, dept, location string, assignedAt time.Time, notes string) error {
	now := time.Now().UTC()
	query := `
	UPDATE it_assets SET
		assigned_to_user_id = ?,
		assigned_to_dept = ?,
		location = CASE WHEN ? != '' THEN ? ELSE location END,
		assigned_at = ?,
		status = 'in_use',
		notes = CASE WHEN ? != '' THEN ? ELSE notes END,
		updated_at = ?
	WHERE id = ?;`

	res, err := r.db.ExecContext(ctx, query, userID, dept, location, location, assignedAt, notes, notes, now, assetID)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrAssetNotFound
	}
	return nil
}

func (r *sqlITAssetRepository) CheckinAsset(ctx context.Context, assetID int64, location string, notes string) error {
	now := time.Now().UTC()
	query := `
	UPDATE it_assets SET
		assigned_to_user_id = NULL,
		assigned_to_dept = '',
		assigned_at = NULL,
		location = CASE WHEN ? != '' THEN ? ELSE location END,
		status = 'in_stock',
		notes = CASE WHEN ? != '' THEN ? ELSE notes END,
		updated_at = ?
	WHERE id = ?;`

	res, err := r.db.ExecContext(ctx, query, location, location, notes, notes, now, assetID)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrAssetNotFound
	}
	return nil
}

func (r *sqlITAssetRepository) UpdateAssetStatus(ctx context.Context, assetID int64, status string, notes string) error {
	now := time.Now().UTC()
	query := `
	UPDATE it_assets SET
		status = ?,
		notes = CASE WHEN ? != '' THEN ? ELSE notes END,
		updated_at = ?
	WHERE id = ?;`

	res, err := r.db.ExecContext(ctx, query, status, notes, notes, now, assetID)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrAssetNotFound
	}
	return nil
}
func (r *sqlITAssetRepository) scanMaintenanceRecord(row interface{ Scan(...any) error }) (*model.AssetMaintenanceRecord, error) {
	var m model.AssetMaintenanceRecord
	var schedDate, compDate sql.NullTime
	var assetTag, assetName sql.NullString

	err := row.Scan(
		&m.ID,
		&m.AssetID,
		&m.MaintenanceType,
		&m.Status,
		&m.Cost,
		&m.PerformedBy,
		&m.Description,
		&schedDate,
		&compDate,
		&m.CreatedAt,
		&m.UpdatedAt,
		&assetTag,
		&assetName,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrMaintenanceRecordNotFound
		}
		return nil, err
	}
	if schedDate.Valid {
		t := schedDate.Time
		m.ScheduledDate = &t
	}
	if compDate.Valid {
		t := compDate.Time
		m.CompletedDate = &t
	}
	if assetTag.Valid {
		m.AssetTag = assetTag.String
	}
	if assetName.Valid {
		m.AssetName = assetName.String
	}
	return &m, nil
}

const maintSelectFields = `
	m.id, m.asset_id, m.maintenance_type, m.status, m.cost, m.performed_by, m.description,
	m.scheduled_date, m.completed_date, m.created_at, m.updated_at,
	COALESCE(a.asset_tag, '') AS asset_tag,
	COALESCE(a.name, '') AS asset_name
`

const maintFromClause = `
	FROM it_asset_maintenance_records m
	JOIN it_assets a ON m.asset_id = a.id
`

func (r *sqlITAssetRepository) CreateMaintenanceRecord(ctx context.Context, record *model.AssetMaintenanceRecord) error {
	now := time.Now().UTC()
	record.CreatedAt = now
	record.UpdatedAt = now

	query := `
	INSERT INTO it_asset_maintenance_records (
		asset_id, maintenance_type, status, cost, performed_by, description,
		scheduled_date, completed_date, created_at, updated_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?);`

	res, err := r.db.ExecContext(ctx, query,
		record.AssetID, record.MaintenanceType, record.Status, record.Cost, record.PerformedBy, record.Description,
		record.ScheduledDate, record.CompletedDate, record.CreatedAt, record.UpdatedAt,
	)
	if err != nil {
		return err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return err
	}
	record.ID = id
	return nil
}

func (r *sqlITAssetRepository) GetMaintenanceRecordByID(ctx context.Context, id int64) (*model.AssetMaintenanceRecord, error) {
	query := fmt.Sprintf("SELECT %s %s WHERE m.id = ?", maintSelectFields, maintFromClause)
	return r.scanMaintenanceRecord(r.db.QueryRowContext(ctx, query, id))
}

func (r *sqlITAssetRepository) UpdateMaintenanceRecord(ctx context.Context, record *model.AssetMaintenanceRecord) error {
	record.UpdatedAt = time.Now().UTC()

	query := `
	UPDATE it_asset_maintenance_records SET
		maintenance_type = ?,
		status = ?,
		cost = ?,
		performed_by = ?,
		description = ?,
		scheduled_date = ?,
		completed_date = ?,
		updated_at = ?
	WHERE id = ?;`

	res, err := r.db.ExecContext(ctx, query,
		record.MaintenanceType, record.Status, record.Cost, record.PerformedBy, record.Description,
		record.ScheduledDate, record.CompletedDate, record.UpdatedAt, record.ID,
	)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrMaintenanceRecordNotFound
	}
	return nil
}

func (r *sqlITAssetRepository) ListMaintenanceRecords(ctx context.Context, assetID int64) ([]model.AssetMaintenanceRecord, error) {
	query := fmt.Sprintf("SELECT %s %s WHERE m.asset_id = ? ORDER BY m.id DESC", maintSelectFields, maintFromClause)
	rows, err := r.db.QueryContext(ctx, query, assetID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []model.AssetMaintenanceRecord
	for rows.Next() {
		m, err := r.scanMaintenanceRecord(rows)
		if err != nil {
			return nil, err
		}
		records = append(records, *m)
	}
	return records, rows.Err()
}

func (r *sqlITAssetRepository) DeleteMaintenanceRecord(ctx context.Context, id int64) error {
	res, err := r.db.ExecContext(ctx, "DELETE FROM it_asset_maintenance_records WHERE id = ?", id)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrMaintenanceRecordNotFound
	}
	return nil
}
func (r *sqlITAssetRepository) scanSoftwareLicense(row interface{ Scan(...any) error }) (*model.SoftwareLicense, error) {
	var l model.SoftwareLicense
	var expDate sql.NullTime

	err := row.Scan(
		&l.ID,
		&l.WorkspaceID,
		&l.SoftwareName,
		&l.Version,
		&l.Publisher,
		&l.LicenseKey,
		&l.LicenseType,
		&l.TotalSeats,
		&l.AllocatedSeats,
		&l.PurchaseCost,
		&l.Currency,
		&expDate,
		&l.Vendor,
		&l.Notes,
		&l.CreatedAt,
		&l.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrLicenseNotFound
		}
		return nil, err
	}
	if expDate.Valid {
		t := expDate.Time
		l.ExpirationDate = &t
	}
	if l.TotalSeats >= l.AllocatedSeats {
		l.AvailableSeats = l.TotalSeats - l.AllocatedSeats
	} else {
		l.AvailableSeats = 0
	}
	return &l, nil
}

const licenseSelectFields = `
	id, workspace_id, software_name, version, publisher, license_key, license_type,
	total_seats, allocated_seats, purchase_cost, currency, expiration_date, vendor,
	notes, created_at, updated_at
`

func (r *sqlITAssetRepository) CreateSoftwareLicense(ctx context.Context, license *model.SoftwareLicense) error {
	now := time.Now().UTC()
	license.CreatedAt = now
	license.UpdatedAt = now
	license.AllocatedSeats = 0

	query := `
	INSERT INTO software_licenses (
		workspace_id, software_name, version, publisher, license_key, license_type,
		total_seats, allocated_seats, purchase_cost, currency, expiration_date,
		vendor, notes, created_at, updated_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);`

	res, err := r.db.ExecContext(ctx, query,
		license.WorkspaceID, license.SoftwareName, license.Version, license.Publisher, license.LicenseKey, license.LicenseType,
		license.TotalSeats, license.AllocatedSeats, license.PurchaseCost, license.Currency, license.ExpirationDate,
		license.Vendor, license.Notes, license.CreatedAt, license.UpdatedAt,
	)
	if err != nil {
		return err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return err
	}
	license.ID = id
	license.AvailableSeats = license.TotalSeats
	return nil
}

func (r *sqlITAssetRepository) GetSoftwareLicenseByID(ctx context.Context, id int64) (*model.SoftwareLicense, error) {
	query := fmt.Sprintf("SELECT %s FROM software_licenses WHERE id = ?", licenseSelectFields)
	return r.scanSoftwareLicense(r.db.QueryRowContext(ctx, query, id))
}

func (r *sqlITAssetRepository) UpdateSoftwareLicense(ctx context.Context, license *model.SoftwareLicense) error {
	license.UpdatedAt = time.Now().UTC()

	query := `
	UPDATE software_licenses SET
		software_name = ?,
		version = ?,
		publisher = ?,
		license_key = ?,
		license_type = ?,
		total_seats = ?,
		purchase_cost = ?,
		currency = ?,
		expiration_date = ?,
		vendor = ?,
		notes = ?,
		updated_at = ?
	WHERE id = ?;`

	res, err := r.db.ExecContext(ctx, query,
		license.SoftwareName, license.Version, license.Publisher, license.LicenseKey, license.LicenseType,
		license.TotalSeats, license.PurchaseCost, license.Currency, license.ExpirationDate,
		license.Vendor, license.Notes, license.UpdatedAt, license.ID,
	)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrLicenseNotFound
	}
	return nil
}

func (r *sqlITAssetRepository) DeleteSoftwareLicense(ctx context.Context, id int64) error {
	res, err := r.db.ExecContext(ctx, "DELETE FROM software_licenses WHERE id = ?", id)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrLicenseNotFound
	}
	return nil
}

func (r *sqlITAssetRepository) ListSoftwareLicenses(ctx context.Context, filter model.LicenseFilterCriteria) ([]model.SoftwareLicense, int, error) {
	var whereClauses []string
	var args []interface{}

	if filter.WorkspaceID > 0 {
		whereClauses = append(whereClauses, "workspace_id = ?")
		args = append(args, filter.WorkspaceID)
	}
	if filter.LicenseType != "" {
		whereClauses = append(whereClauses, "license_type = ?")
		args = append(args, filter.LicenseType)
	}
	if filter.Publisher != "" {
		whereClauses = append(whereClauses, "publisher LIKE ?")
		args = append(args, "%"+filter.Publisher+"%")
	}
	if filter.Search != "" {
		searchPattern := "%" + filter.Search + "%"
		whereClauses = append(whereClauses, "(software_name LIKE ? OR publisher LIKE ? OR vendor LIKE ?)")
		args = append(args, searchPattern, searchPattern, searchPattern)
	}
	if filter.ExpiringDays != nil && *filter.ExpiringDays > 0 {
		now := time.Now().UTC()
		future := now.AddDate(0, 0, *filter.ExpiringDays)
		whereClauses = append(whereClauses, "expiration_date IS NOT NULL AND expiration_date >= ? AND expiration_date <= ?")
		args = append(args, now, future)
	}

	whereSQL := ""
	if len(whereClauses) > 0 {
		whereSQL = "WHERE " + strings.Join(whereClauses, " AND ")
	}

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM software_licenses %s", whereSQL)
	var total int
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := fmt.Sprintf("SELECT %s FROM software_licenses %s ORDER BY id DESC", licenseSelectFields, whereSQL)
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

	var licenses []model.SoftwareLicense
	for rows.Next() {
		l, err := r.scanSoftwareLicense(rows)
		if err != nil {
			return nil, 0, err
		}
		licenses = append(licenses, *l)
	}
	return licenses, total, rows.Err()
}
func (r *sqlITAssetRepository) scanLicenseAllocation(row interface{ Scan(...any) error }) (*model.LicenseAllocation, error) {
	var a model.LicenseAllocation
	var assetID sql.NullInt64
	var userID sql.NullString
	var assetTag, assetName, userName, userEmail, softName sql.NullString

	err := row.Scan(
		&a.ID,
		&a.LicenseID,
		&assetID,
		&userID,
		&a.AllocatedBy,
		&a.AllocatedAt,
		&a.Notes,
		&assetTag,
		&assetName,
		&userName,
		&userEmail,
		&softName,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrLicenseAllocNotFound
		}
		return nil, err
	}
	if assetID.Valid {
		id := assetID.Int64
		a.AssetID = &id
	}
	if userID.Valid {
		u := userID.String
		a.UserID = &u
	}
	if assetTag.Valid {
		a.AssetTag = assetTag.String
	}
	if assetName.Valid {
		a.AssetName = assetName.String
	}
	if userName.Valid {
		a.UserName = userName.String
	}
	if userEmail.Valid {
		a.UserEmail = userEmail.String
	}
	if softName.Valid {
		a.SoftwareName = softName.String
	}
	return &a, nil
}

const allocSelectFields = `
	la.id, la.license_id, la.asset_id, la.user_id, la.allocated_by, la.allocated_at, la.notes,
	COALESCE(ast.asset_tag, '') AS asset_tag,
	COALESCE(ast.name, '') AS asset_name,
	COALESCE(u.username, '') AS user_name,
	COALESCE(u.email, '') AS user_email,
	COALESCE(sl.software_name, '') AS software_name
`

const allocFromClause = `
	FROM license_allocations la
	JOIN software_licenses sl ON la.license_id = sl.id
	LEFT JOIN it_assets ast ON la.asset_id = ast.id
	LEFT JOIN users u ON la.user_id = u.id
`

func (r *sqlITAssetRepository) CreateLicenseAllocation(ctx context.Context, alloc *model.LicenseAllocation) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Check license capacity
	var totalSeats, allocatedSeats int
	err = tx.QueryRowContext(ctx, "SELECT total_seats, allocated_seats FROM software_licenses WHERE id = ?", alloc.LicenseID).Scan(&totalSeats, &allocatedSeats)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrLicenseNotFound
		}
		return err
	}
	if allocatedSeats >= totalSeats {
		return ErrNoSeatsAvailable
	}

	// Check if target already has this license
	if alloc.AssetID != nil && *alloc.AssetID > 0 {
		var count int
		_ = tx.QueryRowContext(ctx, "SELECT COUNT(*) FROM license_allocations WHERE license_id = ? AND asset_id = ?", alloc.LicenseID, *alloc.AssetID).Scan(&count)
		if count > 0 {
			return ErrAlreadyAllocated
		}
	} else if alloc.UserID != nil && *alloc.UserID != "" {
		var count int
		_ = tx.QueryRowContext(ctx, "SELECT COUNT(*) FROM license_allocations WHERE license_id = ? AND user_id = ?", alloc.LicenseID, *alloc.UserID).Scan(&count)
		if count > 0 {
			return ErrAlreadyAllocated
		}
	}

	alloc.AllocatedAt = time.Now().UTC()
	insertQuery := `
	INSERT INTO license_allocations (
		license_id, asset_id, user_id, allocated_by, allocated_at, notes
	) VALUES (?, ?, ?, ?, ?, ?);`

	res, err := tx.ExecContext(ctx, insertQuery,
		alloc.LicenseID, alloc.AssetID, alloc.UserID, alloc.AllocatedBy, alloc.AllocatedAt, alloc.Notes,
	)
	if err != nil {
		return err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return err
	}
	alloc.ID = id

	// Increment allocated seats
	_, err = tx.ExecContext(ctx, "UPDATE software_licenses SET allocated_seats = allocated_seats + 1, updated_at = ? WHERE id = ?", alloc.AllocatedAt, alloc.LicenseID)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *sqlITAssetRepository) GetLicenseAllocationByID(ctx context.Context, id int64) (*model.LicenseAllocation, error) {
	query := fmt.Sprintf("SELECT %s %s WHERE la.id = ?", allocSelectFields, allocFromClause)
	return r.scanLicenseAllocation(r.db.QueryRowContext(ctx, query, id))
}

func (r *sqlITAssetRepository) DeleteLicenseAllocation(ctx context.Context, id int64) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var licenseID int64
	err = tx.QueryRowContext(ctx, "SELECT license_id FROM license_allocations WHERE id = ?", id).Scan(&licenseID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrLicenseAllocNotFound
		}
		return err
	}

	_, err = tx.ExecContext(ctx, "DELETE FROM license_allocations WHERE id = ?", id)
	if err != nil {
		return err
	}

	now := time.Now().UTC()
	_, err = tx.ExecContext(ctx, "UPDATE software_licenses SET allocated_seats = CASE WHEN allocated_seats > 0 THEN allocated_seats - 1 ELSE 0 END, updated_at = ? WHERE id = ?", now, licenseID)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *sqlITAssetRepository) ListLicenseAllocations(ctx context.Context, licenseID int64) ([]model.LicenseAllocation, error) {
	query := fmt.Sprintf("SELECT %s %s WHERE la.license_id = ? ORDER BY la.id DESC", allocSelectFields, allocFromClause)
	rows, err := r.db.QueryContext(ctx, query, licenseID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var allocations []model.LicenseAllocation
	for rows.Next() {
		a, err := r.scanLicenseAllocation(rows)
		if err != nil {
			return nil, err
		}
		allocations = append(allocations, *a)
	}
	return allocations, rows.Err()
}
func (r *sqlITAssetRepository) CreateAuditLog(ctx context.Context, audit *model.AssetAuditLog) error {
	return errNotImplemented
}
func (r *sqlITAssetRepository) ListAuditLogs(ctx context.Context, workspaceID int64, assetID *int64, limit, offset int) ([]model.AssetAuditLog, error) {
	return nil, errNotImplemented
}
func (r *sqlITAssetRepository) GetAssetMetrics(ctx context.Context, workspaceID int64) (*model.AssetMetrics, error) {
	return nil, errNotImplemented
}
