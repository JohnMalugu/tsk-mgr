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
	ErrTestPlanNotFound         = errors.New("test plan not found")
	ErrTestPlanCodeConflict     = errors.New("test plan with this code already exists in workspace")
	ErrTestCaseNotFound         = errors.New("test case not found")
	ErrTestCaseCodeConflict     = errors.New("test case with this code already exists in workspace")
	ErrTestRunNotFound          = errors.New("test run not found")
	ErrQualityGateNotFound      = errors.New("quality gate not found")
	ErrQualityGateNameConflict  = errors.New("quality gate with this name already exists in workspace")
	ErrDefectLinkNotFound       = errors.New("defect link not found")
)

// TestingRepository defines data access methods for ITIL Service Validation & Testing
type TestingRepository interface {
	// Test Plans
	CreatePlan(ctx context.Context, p *model.ValidationTestPlan) error
	GetPlanByID(ctx context.Context, id string) (*model.ValidationTestPlan, error)
	GetPlanByCode(ctx context.Context, workspaceID, code string) (*model.ValidationTestPlan, error)
	UpdatePlan(ctx context.Context, p *model.ValidationTestPlan) error
	DeletePlan(ctx context.Context, id string) error
	ListPlans(ctx context.Context, filter model.TestPlanFilter) ([]model.ValidationTestPlan, int, error)

	// Test Cases
	CreateCase(ctx context.Context, c *model.ValidationTestCase) error
	GetCaseByID(ctx context.Context, id string) (*model.ValidationTestCase, error)
	GetCaseByCode(ctx context.Context, workspaceID, code string) (*model.ValidationTestCase, error)
	UpdateCase(ctx context.Context, c *model.ValidationTestCase) error
	DeleteCase(ctx context.Context, id string) error
	ListCases(ctx context.Context, filter model.TestCaseFilter) ([]model.ValidationTestCase, int, error)

	// Test Runs
	CreateRun(ctx context.Context, r *model.ValidationTestRun) error
	GetRunByID(ctx context.Context, id string) (*model.ValidationTestRun, error)
	UpdateRun(ctx context.Context, r *model.ValidationTestRun) error
	ListRuns(ctx context.Context, filter model.TestRunFilter) ([]model.ValidationTestRun, int, error)

	// Quality Gates
	CreateQualityGate(ctx context.Context, g *model.QualityGate) error
	GetQualityGateByID(ctx context.Context, id string) (*model.QualityGate, error)
	UpdateQualityGate(ctx context.Context, g *model.QualityGate) error
	DeleteQualityGate(ctx context.Context, id string) error
	ListQualityGates(ctx context.Context, filter model.QualityGateFilter) ([]model.QualityGate, int, error)

	// Defect Links
	CreateDefectLink(ctx context.Context, d *model.ValidationDefect) error
	ListDefectLinks(ctx context.Context, filter model.DefectLinkFilter) ([]model.ValidationDefect, int, error)

	// Posture Analytics
	GetTestingAnalytics(ctx context.Context, workspaceID string) (*model.TestingAnalytics, error)
}

type sqlTestingRepository struct {
	db *sql.DB
}

// NewTestingRepository initializes a SQLite testing repository
func NewTestingRepository(db *sql.DB) TestingRepository {
	return &sqlTestingRepository{db: db}
}

// CreatePlan inserts a new validation test plan
func (r *sqlTestingRepository) CreatePlan(ctx context.Context, p *model.ValidationTestPlan) error {
	if p.ID == "" {
		p.ID = uuid.New().String()
	}
	now := time.Now().UTC()
	if p.CreatedAt.IsZero() {
		p.CreatedAt = now
	}
	p.UpdatedAt = now
	if p.Status == "" {
		p.Status = model.TestPlanStatusDraft
	}
	if p.TargetPassRate <= 0 {
		p.TargetPassRate = 95.0
	}

	// Check plan code uniqueness within workspace
	var count int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM validation_test_plans WHERE workspace_id = ? AND plan_code = ?", p.WorkspaceID, p.PlanCode).Scan(&count)
	if err == nil && count > 0 {
		return ErrTestPlanCodeConflict
	}

	query := `
		INSERT INTO validation_test_plans (
			id, workspace_id, plan_code, title, description,
			release_id, change_id, ci_id, test_type, environment,
			status, target_pass_rate, created_by, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err = r.db.ExecContext(ctx, query,
		p.ID,
		p.WorkspaceID,
		p.PlanCode,
		p.Title,
		p.Description,
		p.ReleaseID,
		p.ChangeID,
		p.CIID,
		p.TestType,
		p.Environment,
		p.Status,
		p.TargetPassRate,
		p.CreatedBy,
		p.CreatedAt,
		p.UpdatedAt,
	)
	return err
}

// GetPlanByID retrieves a validation test plan by ID
func (r *sqlTestingRepository) GetPlanByID(ctx context.Context, id string) (*model.ValidationTestPlan, error) {
	query := `
		SELECT id, workspace_id, plan_code, title, description,
		       release_id, change_id, ci_id, test_type, environment,
		       status, target_pass_rate, created_by, created_at, updated_at
		FROM validation_test_plans
		WHERE id = ?`

	var p model.ValidationTestPlan
	var desc, relID, chgID, ciID, createdBy sql.NullString
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&p.ID,
		&p.WorkspaceID,
		&p.PlanCode,
		&p.Title,
		&desc,
		&relID,
		&chgID,
		&ciID,
		&p.TestType,
		&p.Environment,
		&p.Status,
		&p.TargetPassRate,
		&createdBy,
		&p.CreatedAt,
		&p.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrTestPlanNotFound
		}
		return nil, err
	}
	if desc.Valid {
		p.Description = desc.String
	}
	if relID.Valid {
		p.ReleaseID = relID.String
	}
	if chgID.Valid {
		p.ChangeID = chgID.String
	}
	if ciID.Valid {
		p.CIID = ciID.String
	}
	if createdBy.Valid {
		p.CreatedBy = createdBy.String
	}
	return &p, nil
}

// GetPlanByCode retrieves a validation test plan by plan_code within a workspace
func (r *sqlTestingRepository) GetPlanByCode(ctx context.Context, workspaceID, code string) (*model.ValidationTestPlan, error) {
	query := `
		SELECT id, workspace_id, plan_code, title, description,
		       release_id, change_id, ci_id, test_type, environment,
		       status, target_pass_rate, created_by, created_at, updated_at
		FROM validation_test_plans
		WHERE workspace_id = ? AND plan_code = ?`

	var p model.ValidationTestPlan
	var desc, relID, chgID, ciID, createdBy sql.NullString
	err := r.db.QueryRowContext(ctx, query, workspaceID, code).Scan(
		&p.ID,
		&p.WorkspaceID,
		&p.PlanCode,
		&p.Title,
		&desc,
		&relID,
		&chgID,
		&ciID,
		&p.TestType,
		&p.Environment,
		&p.Status,
		&p.TargetPassRate,
		&createdBy,
		&p.CreatedAt,
		&p.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrTestPlanNotFound
		}
		return nil, err
	}
	if desc.Valid {
		p.Description = desc.String
	}
	if relID.Valid {
		p.ReleaseID = relID.String
	}
	if chgID.Valid {
		p.ChangeID = chgID.String
	}
	if ciID.Valid {
		p.CIID = ciID.String
	}
	if createdBy.Valid {
		p.CreatedBy = createdBy.String
	}
	return &p, nil
}

// UpdatePlan updates an existing test plan
func (r *sqlTestingRepository) UpdatePlan(ctx context.Context, p *model.ValidationTestPlan) error {
	p.UpdatedAt = time.Now().UTC()

	query := `
		UPDATE validation_test_plans
		SET title = ?, description = ?, release_id = ?, change_id = ?,
		    ci_id = ?, test_type = ?, environment = ?, status = ?,
		    target_pass_rate = ?, updated_at = ?
		WHERE id = ?`

	res, err := r.db.ExecContext(ctx, query,
		p.Title,
		p.Description,
		p.ReleaseID,
		p.ChangeID,
		p.CIID,
		p.TestType,
		p.Environment,
		p.Status,
		p.TargetPassRate,
		p.UpdatedAt,
		p.ID,
	)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrTestPlanNotFound
	}
	return nil
}

// DeletePlan removes a validation test plan
func (r *sqlTestingRepository) DeletePlan(ctx context.Context, id string) error {
	res, err := r.db.ExecContext(ctx, "DELETE FROM validation_test_plans WHERE id = ?", id)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrTestPlanNotFound
	}
	return nil
}

// ListPlans queries validation test plans based on filter criteria
func (r *sqlTestingRepository) ListPlans(ctx context.Context, filter model.TestPlanFilter) ([]model.ValidationTestPlan, int, error) {
	var conditions []string
	var args []interface{}

	if filter.WorkspaceID != "" {
		conditions = append(conditions, "workspace_id = ?")
		args = append(args, filter.WorkspaceID)
	}
	if filter.ReleaseID != "" {
		conditions = append(conditions, "release_id = ?")
		args = append(args, filter.ReleaseID)
	}
	if filter.ChangeID != "" {
		conditions = append(conditions, "change_id = ?")
		args = append(args, filter.ChangeID)
	}
	if filter.TestType != "" {
		conditions = append(conditions, "test_type = ?")
		args = append(args, filter.TestType)
	}
	if filter.Environment != "" {
		conditions = append(conditions, "environment = ?")
		args = append(args, filter.Environment)
	}
	if filter.Status != "" {
		conditions = append(conditions, "status = ?")
		args = append(args, filter.Status)
	}
	if filter.Search != "" {
		conditions = append(conditions, "(title LIKE ? OR plan_code LIKE ? OR description LIKE ?)")
		s := "%" + filter.Search + "%"
		args = append(args, s, s, s)
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM validation_test_plans %s", whereClause)
	var total int
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
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
		SELECT id, workspace_id, plan_code, title, description,
		       release_id, change_id, ci_id, test_type, environment,
		       status, target_pass_rate, created_by, created_at, updated_at
		FROM validation_test_plans
		%s
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?`, whereClause)

	queryArgs := append(args, limit, offset)
	rows, err := r.db.QueryContext(ctx, selectQuery, queryArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var plans []model.ValidationTestPlan
	for rows.Next() {
		var p model.ValidationTestPlan
		var desc, relID, chgID, ciID, createdBy sql.NullString
		if err := rows.Scan(
			&p.ID,
			&p.WorkspaceID,
			&p.PlanCode,
			&p.Title,
			&desc,
			&relID,
			&chgID,
			&ciID,
			&p.TestType,
			&p.Environment,
			&p.Status,
			&p.TargetPassRate,
			&createdBy,
			&p.CreatedAt,
			&p.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		if desc.Valid {
			p.Description = desc.String
		}
		if relID.Valid {
			p.ReleaseID = relID.String
		}
		if chgID.Valid {
			p.ChangeID = chgID.String
		}
		if ciID.Valid {
			p.CIID = ciID.String
		}
		if createdBy.Valid {
			p.CreatedBy = createdBy.String
		}
		plans = append(plans, p)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return plans, total, nil
}

// Cases stubs
func (r *sqlTestingRepository) CreateCase(ctx context.Context, c *model.ValidationTestCase) error {
	return errors.New("not implemented")
}
func (r *sqlTestingRepository) GetCaseByID(ctx context.Context, id string) (*model.ValidationTestCase, error) {
	return nil, errors.New("not implemented")
}
func (r *sqlTestingRepository) GetCaseByCode(ctx context.Context, workspaceID, code string) (*model.ValidationTestCase, error) {
	return nil, errors.New("not implemented")
}
func (r *sqlTestingRepository) UpdateCase(ctx context.Context, c *model.ValidationTestCase) error {
	return errors.New("not implemented")
}
func (r *sqlTestingRepository) DeleteCase(ctx context.Context, id string) error {
	return errors.New("not implemented")
}
func (r *sqlTestingRepository) ListCases(ctx context.Context, filter model.TestCaseFilter) ([]model.ValidationTestCase, int, error) {
	return nil, 0, errors.New("not implemented")
}

// Runs stubs
func (r *sqlTestingRepository) CreateRun(ctx context.Context, run *model.ValidationTestRun) error {
	return errors.New("not implemented")
}
func (r *sqlTestingRepository) GetRunByID(ctx context.Context, id string) (*model.ValidationTestRun, error) {
	return nil, errors.New("not implemented")
}
func (r *sqlTestingRepository) UpdateRun(ctx context.Context, run *model.ValidationTestRun) error {
	return errors.New("not implemented")
}
func (r *sqlTestingRepository) ListRuns(ctx context.Context, filter model.TestRunFilter) ([]model.ValidationTestRun, int, error) {
	return nil, 0, errors.New("not implemented")
}

// Quality Gates stubs
func (r *sqlTestingRepository) CreateQualityGate(ctx context.Context, g *model.QualityGate) error {
	return errors.New("not implemented")
}
func (r *sqlTestingRepository) GetQualityGateByID(ctx context.Context, id string) (*model.QualityGate, error) {
	return nil, errors.New("not implemented")
}
func (r *sqlTestingRepository) UpdateQualityGate(ctx context.Context, g *model.QualityGate) error {
	return errors.New("not implemented")
}
func (r *sqlTestingRepository) DeleteQualityGate(ctx context.Context, id string) error {
	return errors.New("not implemented")
}
func (r *sqlTestingRepository) ListQualityGates(ctx context.Context, filter model.QualityGateFilter) ([]model.QualityGate, int, error) {
	return nil, 0, errors.New("not implemented")
}

// Defects stubs
func (r *sqlTestingRepository) CreateDefectLink(ctx context.Context, d *model.ValidationDefect) error {
	return errors.New("not implemented")
}
func (r *sqlTestingRepository) ListDefectLinks(ctx context.Context, filter model.DefectLinkFilter) ([]model.ValidationDefect, int, error) {
	return nil, 0, errors.New("not implemented")
}

// Posture Analytics stub
func (r *sqlTestingRepository) GetTestingAnalytics(ctx context.Context, workspaceID string) (*model.TestingAnalytics, error) {
	return nil, errors.New("not implemented")
}
