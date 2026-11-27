package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/JohnMalugu/tsk-mgr-api/internal/model"
)

var (
	_ = uuid.New
	_ = fmt.Sprintf
	_ = strings.Join
	_ = time.Now
	_ = errors.New
)

var (
	ErrGRCFrameworkNotFound     = errors.New("compliance framework not found")
	ErrGRCFrameworkCodeConflict = errors.New("compliance framework with this code already exists in workspace")
	ErrGRCControlNotFound       = errors.New("compliance control not found")
	ErrGRCControlCodeConflict   = errors.New("compliance control with this code already exists in workspace")
	ErrGRCEvidenceNotFound      = errors.New("compliance evidence not found")
	ErrGRCAssessmentNotFound    = errors.New("compliance assessment not found")
	ErrGRCFindingNotFound       = errors.New("compliance finding not found")
)

// ComplianceRepository defines data access methods for ITIL GRC & Compliance Management
type ComplianceRepository interface {
	// Frameworks
	CreateFramework(ctx context.Context, f *model.ComplianceFramework) error
	GetFrameworkByID(ctx context.Context, id string) (*model.ComplianceFramework, error)
	GetFrameworkByCode(ctx context.Context, workspaceID, code string) (*model.ComplianceFramework, error)
	UpdateFramework(ctx context.Context, f *model.ComplianceFramework) error
	DeleteFramework(ctx context.Context, id string) error
	ListFrameworks(ctx context.Context, filter model.GRCFrameworkFilter) ([]model.ComplianceFramework, int, error)

	// Controls
	CreateControl(ctx context.Context, c *model.ComplianceControl) error
	GetControlByID(ctx context.Context, id string) (*model.ComplianceControl, error)
	GetControlByCode(ctx context.Context, workspaceID, code string) (*model.ComplianceControl, error)
	UpdateControl(ctx context.Context, c *model.ComplianceControl) error
	DeleteControl(ctx context.Context, id string) error
	ListControls(ctx context.Context, filter model.GRCControlFilter) ([]model.ComplianceControl, int, error)

	// Evidence
	CreateEvidence(ctx context.Context, e *model.ComplianceEvidence) error
	GetEvidenceByID(ctx context.Context, id string) (*model.ComplianceEvidence, error)
	UpdateEvidence(ctx context.Context, e *model.ComplianceEvidence) error
	DeleteEvidence(ctx context.Context, id string) error
	ListEvidence(ctx context.Context, filter model.GRCEvidenceFilter) ([]model.ComplianceEvidence, int, error)

	// Assessments
	CreateAssessment(ctx context.Context, a *model.ComplianceAssessment) error
	GetAssessmentByID(ctx context.Context, id string) (*model.ComplianceAssessment, error)
	UpdateAssessment(ctx context.Context, a *model.ComplianceAssessment) error
	DeleteAssessment(ctx context.Context, id string) error
	ListAssessments(ctx context.Context, filter model.GRCAssessmentFilter) ([]model.ComplianceAssessment, int, error)

	// Findings
	CreateFinding(ctx context.Context, f *model.ComplianceFinding) error
	GetFindingByID(ctx context.Context, id string) (*model.ComplianceFinding, error)
	UpdateFinding(ctx context.Context, f *model.ComplianceFinding) error
	DeleteFinding(ctx context.Context, id string) error
	ListFindings(ctx context.Context, filter model.GRCFindingFilter) ([]model.ComplianceFinding, int, error)

	// Posture Analytics
	GetComplianceAnalytics(ctx context.Context, workspaceID string) (*model.ComplianceAnalytics, error)
}

type sqlComplianceRepository struct {
	db *sql.DB
}

// NewComplianceRepository initializes a SQLite compliance repository
func NewComplianceRepository(db *sql.DB) ComplianceRepository {
	return &sqlComplianceRepository{db: db}
}

// CreateFramework inserts a new compliance framework
func (r *sqlComplianceRepository) CreateFramework(ctx context.Context, f *model.ComplianceFramework) error {
	if f.ID == "" {
		f.ID = uuid.New().String()
	}
	now := time.Now().UTC()
	if f.CreatedAt.IsZero() {
		f.CreatedAt = now
	}
	f.UpdatedAt = now
	if f.StandardType == "" {
		f.StandardType = model.GRCStandardCustom
	}
	if f.Version == "" {
		f.Version = "1.0"
	}
	if f.Status == "" {
		f.Status = model.GRCFrameworkStatusActive
	}

	var count int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM compliance_frameworks WHERE workspace_id = ? AND framework_code = ?", f.WorkspaceID, f.FrameworkCode).Scan(&count)
	if err == nil && count > 0 {
		return ErrGRCFrameworkCodeConflict
	}

	query := `
		INSERT INTO compliance_frameworks (
			id, workspace_id, framework_code, name, standard_type,
			version, description, owner_id, status, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err = r.db.ExecContext(ctx, query,
		f.ID, f.WorkspaceID, f.FrameworkCode, f.Name, f.StandardType,
		f.Version, f.Description, f.OwnerID, f.Status, f.CreatedAt, f.UpdatedAt,
	)
	return err
}

func scanGRCFramework(scanner interface{ Scan(dest ...interface{}) error }) (*model.ComplianceFramework, error) {
	var f model.ComplianceFramework
	var desc, ownerID sql.NullString
	err := scanner.Scan(
		&f.ID, &f.WorkspaceID, &f.FrameworkCode, &f.Name, &f.StandardType,
		&f.Version, &desc, &ownerID, &f.Status, &f.CreatedAt, &f.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	if desc.Valid {
		f.Description = desc.String
	}
	if ownerID.Valid {
		f.OwnerID = ownerID.String
	}
	return &f, nil
}

// GetFrameworkByID retrieves a compliance framework by ID
func (r *sqlComplianceRepository) GetFrameworkByID(ctx context.Context, id string) (*model.ComplianceFramework, error) {
	query := `
		SELECT id, workspace_id, framework_code, name, standard_type,
		       version, description, owner_id, status, created_at, updated_at
		FROM compliance_frameworks
		WHERE id = ?`

	f, err := scanGRCFramework(r.db.QueryRowContext(ctx, query, id))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrGRCFrameworkNotFound
		}
		return nil, err
	}
	return f, nil
}

// GetFrameworkByCode retrieves a compliance framework by framework_code within a workspace
func (r *sqlComplianceRepository) GetFrameworkByCode(ctx context.Context, workspaceID, code string) (*model.ComplianceFramework, error) {
	query := `
		SELECT id, workspace_id, framework_code, name, standard_type,
		       version, description, owner_id, status, created_at, updated_at
		FROM compliance_frameworks
		WHERE workspace_id = ? AND framework_code = ?`

	f, err := scanGRCFramework(r.db.QueryRowContext(ctx, query, workspaceID, code))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrGRCFrameworkNotFound
		}
		return nil, err
	}
	return f, nil
}

// UpdateFramework updates an existing compliance framework
func (r *sqlComplianceRepository) UpdateFramework(ctx context.Context, f *model.ComplianceFramework) error {
	f.UpdatedAt = time.Now().UTC()

	query := `
		UPDATE compliance_frameworks
		SET name = ?, version = ?, description = ?, owner_id = ?, status = ?, updated_at = ?
		WHERE id = ?`

	res, err := r.db.ExecContext(ctx, query,
		f.Name, f.Version, f.Description, f.OwnerID, f.Status, f.UpdatedAt, f.ID,
	)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrGRCFrameworkNotFound
	}
	return nil
}

// DeleteFramework removes a compliance framework
func (r *sqlComplianceRepository) DeleteFramework(ctx context.Context, id string) error {
	res, err := r.db.ExecContext(ctx, "DELETE FROM compliance_frameworks WHERE id = ?", id)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrGRCFrameworkNotFound
	}
	return nil
}

// ListFrameworks queries compliance frameworks based on filter criteria
func (r *sqlComplianceRepository) ListFrameworks(ctx context.Context, filter model.GRCFrameworkFilter) ([]model.ComplianceFramework, int, error) {
	var conditions []string
	var args []interface{}

	if filter.WorkspaceID != "" {
		conditions = append(conditions, "workspace_id = ?")
		args = append(args, filter.WorkspaceID)
	}
	if filter.StandardType != "" {
		conditions = append(conditions, "standard_type = ?")
		args = append(args, filter.StandardType)
	}
	if filter.Status != "" {
		conditions = append(conditions, "status = ?")
		args = append(args, filter.Status)
	}
	if filter.Search != "" {
		conditions = append(conditions, "(name LIKE ? OR framework_code LIKE ? OR description LIKE ?)")
		s := "%" + filter.Search + "%"
		args = append(args, s, s, s)
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	var total int
	err := r.db.QueryRowContext(ctx, fmt.Sprintf("SELECT COUNT(*) FROM compliance_frameworks %s", whereClause), args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	limit := filter.Limit
	if limit <= 0 {
		limit = 50
	}
	offset := filter.Offset
	if offset < 0 {
		offset = 0
	}

	selectQuery := fmt.Sprintf(`
		SELECT id, workspace_id, framework_code, name, standard_type,
		       version, description, owner_id, status, created_at, updated_at
		FROM compliance_frameworks
		%s
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?`, whereClause)

	rows, err := r.db.QueryContext(ctx, selectQuery, append(args, limit, offset)...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var frameworks []model.ComplianceFramework
	for rows.Next() {
		f, err := scanGRCFramework(rows)
		if err != nil {
			return nil, 0, err
		}
		frameworks = append(frameworks, *f)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return frameworks, total, nil
}

// CreateControl inserts a new compliance control under a framework
func (r *sqlComplianceRepository) CreateControl(ctx context.Context, c *model.ComplianceControl) error {
	if c.ID == "" {
		c.ID = uuid.New().String()
	}
	now := time.Now().UTC()
	if c.CreatedAt.IsZero() {
		c.CreatedAt = now
	}
	c.UpdatedAt = now
	if c.Domain == "" {
		c.Domain = model.GRCDomainAccessControl
	}
	if c.ImplementationStatus == "" {
		c.ImplementationStatus = model.GRCControlNotImplemented
	}
	if c.AutomationLevel == "" {
		c.AutomationLevel = model.GRCAutomationManual
	}

	var count int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM compliance_controls WHERE workspace_id = ? AND control_code = ?", c.WorkspaceID, c.ControlCode).Scan(&count)
	if err == nil && count > 0 {
		return ErrGRCControlCodeConflict
	}

	query := `
		INSERT INTO compliance_controls (
			id, workspace_id, framework_id, control_code, title, description,
			domain, implementation_status, automation_level, owner_id, ci_id,
			last_tested_at, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err = r.db.ExecContext(ctx, query,
		c.ID, c.WorkspaceID, c.FrameworkID, c.ControlCode, c.Title, c.Description,
		c.Domain, c.ImplementationStatus, c.AutomationLevel, c.OwnerID, c.CIID,
		c.LastTestedAt, c.CreatedAt, c.UpdatedAt,
	)
	return err
}

func scanGRCControl(scanner interface{ Scan(dest ...interface{}) error }) (*model.ComplianceControl, error) {
	var c model.ComplianceControl
	var desc, ownerID, ciID sql.NullString
	var lastTestedAt sql.NullTime
	err := scanner.Scan(
		&c.ID, &c.WorkspaceID, &c.FrameworkID, &c.ControlCode, &c.Title, &desc,
		&c.Domain, &c.ImplementationStatus, &c.AutomationLevel, &ownerID, &ciID,
		&lastTestedAt, &c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	if desc.Valid {
		c.Description = desc.String
	}
	if ownerID.Valid {
		c.OwnerID = ownerID.String
	}
	if ciID.Valid {
		c.CIID = ciID.String
	}
	if lastTestedAt.Valid {
		c.LastTestedAt = &lastTestedAt.Time
	}
	return &c, nil
}

// GetControlByID retrieves a compliance control by ID
func (r *sqlComplianceRepository) GetControlByID(ctx context.Context, id string) (*model.ComplianceControl, error) {
	query := `
		SELECT id, workspace_id, framework_id, control_code, title, description,
		       domain, implementation_status, automation_level, owner_id, ci_id,
		       last_tested_at, created_at, updated_at
		FROM compliance_controls
		WHERE id = ?`

	c, err := scanGRCControl(r.db.QueryRowContext(ctx, query, id))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrGRCControlNotFound
		}
		return nil, err
	}
	return c, nil
}

// GetControlByCode retrieves a compliance control by control_code within a workspace
func (r *sqlComplianceRepository) GetControlByCode(ctx context.Context, workspaceID, code string) (*model.ComplianceControl, error) {
	query := `
		SELECT id, workspace_id, framework_id, control_code, title, description,
		       domain, implementation_status, automation_level, owner_id, ci_id,
		       last_tested_at, created_at, updated_at
		FROM compliance_controls
		WHERE workspace_id = ? AND control_code = ?`

	c, err := scanGRCControl(r.db.QueryRowContext(ctx, query, workspaceID, code))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrGRCControlNotFound
		}
		return nil, err
	}
	return c, nil
}

// UpdateControl updates an existing compliance control
func (r *sqlComplianceRepository) UpdateControl(ctx context.Context, c *model.ComplianceControl) error {
	c.UpdatedAt = time.Now().UTC()

	query := `
		UPDATE compliance_controls
		SET title = ?, description = ?, domain = ?, implementation_status = ?,
		    automation_level = ?, owner_id = ?, ci_id = ?, last_tested_at = ?, updated_at = ?
		WHERE id = ?`

	res, err := r.db.ExecContext(ctx, query,
		c.Title, c.Description, c.Domain, c.ImplementationStatus,
		c.AutomationLevel, c.OwnerID, c.CIID, c.LastTestedAt, c.UpdatedAt, c.ID,
	)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrGRCControlNotFound
	}
	return nil
}

// DeleteControl removes a compliance control
func (r *sqlComplianceRepository) DeleteControl(ctx context.Context, id string) error {
	res, err := r.db.ExecContext(ctx, "DELETE FROM compliance_controls WHERE id = ?", id)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrGRCControlNotFound
	}
	return nil
}

// ListControls queries compliance controls based on filter criteria
func (r *sqlComplianceRepository) ListControls(ctx context.Context, filter model.GRCControlFilter) ([]model.ComplianceControl, int, error) {
	var conditions []string
	var args []interface{}

	if filter.WorkspaceID != "" {
		conditions = append(conditions, "workspace_id = ?")
		args = append(args, filter.WorkspaceID)
	}
	if filter.FrameworkID != "" {
		conditions = append(conditions, "framework_id = ?")
		args = append(args, filter.FrameworkID)
	}
	if filter.Domain != "" {
		conditions = append(conditions, "domain = ?")
		args = append(args, filter.Domain)
	}
	if filter.ImplementationStatus != "" {
		conditions = append(conditions, "implementation_status = ?")
		args = append(args, filter.ImplementationStatus)
	}
	if filter.AutomationLevel != "" {
		conditions = append(conditions, "automation_level = ?")
		args = append(args, filter.AutomationLevel)
	}
	if filter.OwnerID != "" {
		conditions = append(conditions, "owner_id = ?")
		args = append(args, filter.OwnerID)
	}
	if filter.Search != "" {
		conditions = append(conditions, "(title LIKE ? OR control_code LIKE ? OR description LIKE ?)")
		s := "%" + filter.Search + "%"
		args = append(args, s, s, s)
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	var total int
	err := r.db.QueryRowContext(ctx, fmt.Sprintf("SELECT COUNT(*) FROM compliance_controls %s", whereClause), args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	limit := filter.Limit
	if limit <= 0 {
		limit = 50
	}
	offset := filter.Offset
	if offset < 0 {
		offset = 0
	}

	selectQuery := fmt.Sprintf(`
		SELECT id, workspace_id, framework_id, control_code, title, description,
		       domain, implementation_status, automation_level, owner_id, ci_id,
		       last_tested_at, created_at, updated_at
		FROM compliance_controls
		%s
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?`, whereClause)

	rows, err := r.db.QueryContext(ctx, selectQuery, append(args, limit, offset)...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var controls []model.ComplianceControl
	for rows.Next() {
		c, err := scanGRCControl(rows)
		if err != nil {
			return nil, 0, err
		}
		controls = append(controls, *c)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return controls, total, nil
}

// Evidence stubs
func (r *sqlComplianceRepository) CreateEvidence(ctx context.Context, e *model.ComplianceEvidence) error {
	return errors.New("not implemented")
}
func (r *sqlComplianceRepository) GetEvidenceByID(ctx context.Context, id string) (*model.ComplianceEvidence, error) {
	return nil, errors.New("not implemented")
}
func (r *sqlComplianceRepository) UpdateEvidence(ctx context.Context, e *model.ComplianceEvidence) error {
	return errors.New("not implemented")
}
func (r *sqlComplianceRepository) DeleteEvidence(ctx context.Context, id string) error {
	return errors.New("not implemented")
}
func (r *sqlComplianceRepository) ListEvidence(ctx context.Context, filter model.GRCEvidenceFilter) ([]model.ComplianceEvidence, int, error) {
	return nil, 0, errors.New("not implemented")
}

// Assessments stubs
func (r *sqlComplianceRepository) CreateAssessment(ctx context.Context, a *model.ComplianceAssessment) error {
	return errors.New("not implemented")
}
func (r *sqlComplianceRepository) GetAssessmentByID(ctx context.Context, id string) (*model.ComplianceAssessment, error) {
	return nil, errors.New("not implemented")
}
func (r *sqlComplianceRepository) UpdateAssessment(ctx context.Context, a *model.ComplianceAssessment) error {
	return errors.New("not implemented")
}
func (r *sqlComplianceRepository) DeleteAssessment(ctx context.Context, id string) error {
	return errors.New("not implemented")
}
func (r *sqlComplianceRepository) ListAssessments(ctx context.Context, filter model.GRCAssessmentFilter) ([]model.ComplianceAssessment, int, error) {
	return nil, 0, errors.New("not implemented")
}

// Findings stubs
func (r *sqlComplianceRepository) CreateFinding(ctx context.Context, f *model.ComplianceFinding) error {
	return errors.New("not implemented")
}
func (r *sqlComplianceRepository) GetFindingByID(ctx context.Context, id string) (*model.ComplianceFinding, error) {
	return nil, errors.New("not implemented")
}
func (r *sqlComplianceRepository) UpdateFinding(ctx context.Context, f *model.ComplianceFinding) error {
	return errors.New("not implemented")
}
func (r *sqlComplianceRepository) DeleteFinding(ctx context.Context, id string) error {
	return errors.New("not implemented")
}
func (r *sqlComplianceRepository) ListFindings(ctx context.Context, filter model.GRCFindingFilter) ([]model.ComplianceFinding, int, error) {
	return nil, 0, errors.New("not implemented")
}

// Analytics stub
func (r *sqlComplianceRepository) GetComplianceAnalytics(ctx context.Context, workspaceID string) (*model.ComplianceAnalytics, error) {
	return nil, errors.New("not implemented")
}
