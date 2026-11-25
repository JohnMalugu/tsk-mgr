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

// CreateCase inserts a new validation test case linked to a test plan
func (r *sqlTestingRepository) CreateCase(ctx context.Context, c *model.ValidationTestCase) error {
	if c.ID == "" {
		c.ID = uuid.New().String()
	}
	now := time.Now().UTC()
	if c.CreatedAt.IsZero() {
		c.CreatedAt = now
	}
	c.UpdatedAt = now
	if c.Severity == "" {
		c.Severity = model.TestCaseSeverityMedium
	}
	if c.Category == "" {
		c.Category = model.TestCaseCategoryIntegration
	}
	if c.Status == "" {
		c.Status = model.TestCaseStatusReady
	}

	// Check case code uniqueness within workspace
	var count int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM validation_test_cases WHERE workspace_id = ? AND case_code = ?", c.WorkspaceID, c.CaseCode).Scan(&count)
	if err == nil && count > 0 {
		return ErrTestCaseCodeConflict
	}

	query := `
		INSERT INTO validation_test_cases (
			id, workspace_id, plan_id, case_code, title,
			description, preconditions, test_steps, expected_result,
			severity, category, status, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err = r.db.ExecContext(ctx, query,
		c.ID,
		c.WorkspaceID,
		c.PlanID,
		c.CaseCode,
		c.Title,
		c.Description,
		c.Preconditions,
		c.TestSteps,
		c.ExpectedResult,
		c.Severity,
		c.Category,
		c.Status,
		c.CreatedAt,
		c.UpdatedAt,
	)
	return err
}

// GetCaseByID retrieves a test case by ID
func (r *sqlTestingRepository) GetCaseByID(ctx context.Context, id string) (*model.ValidationTestCase, error) {
	query := `
		SELECT id, workspace_id, plan_id, case_code, title,
		       description, preconditions, test_steps, expected_result,
		       severity, category, status, created_at, updated_at
		FROM validation_test_cases
		WHERE id = ?`

	var c model.ValidationTestCase
	var desc, pre sql.NullString
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&c.ID,
		&c.WorkspaceID,
		&c.PlanID,
		&c.CaseCode,
		&c.Title,
		&desc,
		&pre,
		&c.TestSteps,
		&c.ExpectedResult,
		&c.Severity,
		&c.Category,
		&c.Status,
		&c.CreatedAt,
		&c.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrTestCaseNotFound
		}
		return nil, err
	}
	if desc.Valid {
		c.Description = desc.String
	}
	if pre.Valid {
		c.Preconditions = pre.String
	}
	return &c, nil
}

// GetCaseByCode retrieves a test case by case_code within a workspace
func (r *sqlTestingRepository) GetCaseByCode(ctx context.Context, workspaceID, code string) (*model.ValidationTestCase, error) {
	query := `
		SELECT id, workspace_id, plan_id, case_code, title,
		       description, preconditions, test_steps, expected_result,
		       severity, category, status, created_at, updated_at
		FROM validation_test_cases
		WHERE workspace_id = ? AND case_code = ?`

	var c model.ValidationTestCase
	var desc, pre sql.NullString
	err := r.db.QueryRowContext(ctx, query, workspaceID, code).Scan(
		&c.ID,
		&c.WorkspaceID,
		&c.PlanID,
		&c.CaseCode,
		&c.Title,
		&desc,
		&pre,
		&c.TestSteps,
		&c.ExpectedResult,
		&c.Severity,
		&c.Category,
		&c.Status,
		&c.CreatedAt,
		&c.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrTestCaseNotFound
		}
		return nil, err
	}
	if desc.Valid {
		c.Description = desc.String
	}
	if pre.Valid {
		c.Preconditions = pre.String
	}
	return &c, nil
}

// UpdateCase updates an existing test case
func (r *sqlTestingRepository) UpdateCase(ctx context.Context, c *model.ValidationTestCase) error {
	c.UpdatedAt = time.Now().UTC()

	query := `
		UPDATE validation_test_cases
		SET title = ?, description = ?, preconditions = ?, test_steps = ?,
		    expected_result = ?, severity = ?, category = ?, status = ?, updated_at = ?
		WHERE id = ?`

	res, err := r.db.ExecContext(ctx, query,
		c.Title,
		c.Description,
		c.Preconditions,
		c.TestSteps,
		c.ExpectedResult,
		c.Severity,
		c.Category,
		c.Status,
		c.UpdatedAt,
		c.ID,
	)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrTestCaseNotFound
	}
	return nil
}

// DeleteCase removes a test case
func (r *sqlTestingRepository) DeleteCase(ctx context.Context, id string) error {
	res, err := r.db.ExecContext(ctx, "DELETE FROM validation_test_cases WHERE id = ?", id)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrTestCaseNotFound
	}
	return nil
}

// ListCases queries test cases based on filter criteria
func (r *sqlTestingRepository) ListCases(ctx context.Context, filter model.TestCaseFilter) ([]model.ValidationTestCase, int, error) {
	var conditions []string
	var args []interface{}

	if filter.WorkspaceID != "" {
		conditions = append(conditions, "workspace_id = ?")
		args = append(args, filter.WorkspaceID)
	}
	if filter.PlanID != "" {
		conditions = append(conditions, "plan_id = ?")
		args = append(args, filter.PlanID)
	}
	if filter.Category != "" {
		conditions = append(conditions, "category = ?")
		args = append(args, filter.Category)
	}
	if filter.Severity != "" {
		conditions = append(conditions, "severity = ?")
		args = append(args, filter.Severity)
	}
	if filter.Status != "" {
		conditions = append(conditions, "status = ?")
		args = append(args, filter.Status)
	}
	if filter.Search != "" {
		conditions = append(conditions, "(title LIKE ? OR case_code LIKE ? OR description LIKE ?)")
		s := "%" + filter.Search + "%"
		args = append(args, s, s, s)
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM validation_test_cases %s", whereClause)
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
		SELECT id, workspace_id, plan_id, case_code, title,
		       description, preconditions, test_steps, expected_result,
		       severity, category, status, created_at, updated_at
		FROM validation_test_cases
		%s
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?`, whereClause)

	queryArgs := append(args, limit, offset)
	rows, err := r.db.QueryContext(ctx, selectQuery, queryArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var cases []model.ValidationTestCase
	for rows.Next() {
		var c model.ValidationTestCase
		var desc, pre sql.NullString
		if err := rows.Scan(
			&c.ID,
			&c.WorkspaceID,
			&c.PlanID,
			&c.CaseCode,
			&c.Title,
			&desc,
			&pre,
			&c.TestSteps,
			&c.ExpectedResult,
			&c.Severity,
			&c.Category,
			&c.Status,
			&c.CreatedAt,
			&c.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		if desc.Valid {
			c.Description = desc.String
		}
		if pre.Valid {
			c.Preconditions = pre.String
		}
		cases = append(cases, c)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return cases, total, nil
}

// CreateRun records a new test execution run
func (r *sqlTestingRepository) CreateRun(ctx context.Context, run *model.ValidationTestRun) error {
	if run.ID == "" {
		run.ID = uuid.New().String()
	}
	now := time.Now().UTC()
	if run.ExecutedAt.IsZero() {
		run.ExecutedAt = now
	}
	if run.CreatedAt.IsZero() {
		run.CreatedAt = now
	}
	if run.Status == "" {
		run.Status = model.TestRunStatusPassed
	}

	query := `
		INSERT INTO validation_test_runs (
			id, workspace_id, plan_id, case_id, run_code,
			executed_by, status, actual_result, execution_time_ms,
			error_message, stack_trace, evidence_url, defect_id,
			executed_at, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := r.db.ExecContext(ctx, query,
		run.ID,
		run.WorkspaceID,
		run.PlanID,
		run.CaseID,
		run.RunCode,
		run.ExecutedBy,
		run.Status,
		run.ActualResult,
		run.ExecutionTimeMs,
		run.ErrorMessage,
		run.StackTrace,
		run.EvidenceURL,
		run.DefectID,
		run.ExecutedAt,
		run.CreatedAt,
	)
	return err
}

// GetRunByID retrieves a test run execution record by ID
func (r *sqlTestingRepository) GetRunByID(ctx context.Context, id string) (*model.ValidationTestRun, error) {
	query := `
		SELECT id, workspace_id, plan_id, case_id, run_code,
		       executed_by, status, actual_result, execution_time_ms,
		       error_message, stack_trace, evidence_url, defect_id,
		       executed_at, created_at
		FROM validation_test_runs
		WHERE id = ?`

	var run model.ValidationTestRun
	var actRes, errMsg, stTrace, evURL, defID sql.NullString
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&run.ID,
		&run.WorkspaceID,
		&run.PlanID,
		&run.CaseID,
		&run.RunCode,
		&run.ExecutedBy,
		&run.Status,
		&actRes,
		&run.ExecutionTimeMs,
		&errMsg,
		&stTrace,
		&evURL,
		&defID,
		&run.ExecutedAt,
		&run.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrTestRunNotFound
		}
		return nil, err
	}
	if actRes.Valid {
		run.ActualResult = actRes.String
	}
	if errMsg.Valid {
		run.ErrorMessage = errMsg.String
	}
	if stTrace.Valid {
		run.StackTrace = stTrace.String
	}
	if evURL.Valid {
		run.EvidenceURL = evURL.String
	}
	if defID.Valid {
		run.DefectID = defID.String
	}
	return &run, nil
}

// UpdateRun updates an existing test run execution record
func (r *sqlTestingRepository) UpdateRun(ctx context.Context, run *model.ValidationTestRun) error {
	query := `
		UPDATE validation_test_runs
		SET status = ?, actual_result = ?, execution_time_ms = ?,
		    error_message = ?, stack_trace = ?, evidence_url = ?, defect_id = ?
		WHERE id = ?`

	res, err := r.db.ExecContext(ctx, query,
		run.Status,
		run.ActualResult,
		run.ExecutionTimeMs,
		run.ErrorMessage,
		run.StackTrace,
		run.EvidenceURL,
		run.DefectID,
		run.ID,
	)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrTestRunNotFound
	}
	return nil
}

// ListRuns queries test run execution logs based on filter criteria
func (r *sqlTestingRepository) ListRuns(ctx context.Context, filter model.TestRunFilter) ([]model.ValidationTestRun, int, error) {
	var conditions []string
	var args []interface{}

	if filter.WorkspaceID != "" {
		conditions = append(conditions, "workspace_id = ?")
		args = append(args, filter.WorkspaceID)
	}
	if filter.PlanID != "" {
		conditions = append(conditions, "plan_id = ?")
		args = append(args, filter.PlanID)
	}
	if filter.CaseID != "" {
		conditions = append(conditions, "case_id = ?")
		args = append(args, filter.CaseID)
	}
	if filter.Status != "" {
		conditions = append(conditions, "status = ?")
		args = append(args, filter.Status)
	}
	if filter.ExecutedBy != "" {
		conditions = append(conditions, "executed_by = ?")
		args = append(args, filter.ExecutedBy)
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM validation_test_runs %s", whereClause)
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
		SELECT id, workspace_id, plan_id, case_id, run_code,
		       executed_by, status, actual_result, execution_time_ms,
		       error_message, stack_trace, evidence_url, defect_id,
		       executed_at, created_at
		FROM validation_test_runs
		%s
		ORDER BY executed_at DESC
		LIMIT ? OFFSET ?`, whereClause)

	queryArgs := append(args, limit, offset)
	rows, err := r.db.QueryContext(ctx, selectQuery, queryArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var runs []model.ValidationTestRun
	for rows.Next() {
		var run model.ValidationTestRun
		var actRes, errMsg, stTrace, evURL, defID sql.NullString
		if err := rows.Scan(
			&run.ID,
			&run.WorkspaceID,
			&run.PlanID,
			&run.CaseID,
			&run.RunCode,
			&run.ExecutedBy,
			&run.Status,
			&actRes,
			&run.ExecutionTimeMs,
			&errMsg,
			&stTrace,
			&evURL,
			&defID,
			&run.ExecutedAt,
			&run.CreatedAt,
		); err != nil {
			return nil, 0, err
		}
		if actRes.Valid {
			run.ActualResult = actRes.String
		}
		if errMsg.Valid {
			run.ErrorMessage = errMsg.String
		}
		if stTrace.Valid {
			run.StackTrace = stTrace.String
		}
		if evURL.Valid {
			run.EvidenceURL = evURL.String
		}
		if defID.Valid {
			run.DefectID = defID.String
		}
		runs = append(runs, run)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return runs, total, nil
}

// CreateQualityGate inserts a new quality gate policy
func (r *sqlTestingRepository) CreateQualityGate(ctx context.Context, g *model.QualityGate) error {
	if g.ID == "" {
		g.ID = uuid.New().String()
	}
	now := time.Now().UTC()
	if g.CreatedAt.IsZero() {
		g.CreatedAt = now
	}
	g.UpdatedAt = now
	if g.Status == "" {
		g.Status = model.QualityGateStatusPending
	}
	if g.MinPassRate <= 0 {
		g.MinPassRate = 95.0
	}

	// Check name uniqueness within workspace
	var count int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM quality_gates WHERE workspace_id = ? AND name = ?", g.WorkspaceID, g.Name).Scan(&count)
	if err == nil && count > 0 {
		return ErrQualityGateNameConflict
	}

	query := `
		INSERT INTO quality_gates (
			id, workspace_id, name, release_id, change_id,
			min_pass_rate, require_security_signoff, require_performance_signoff,
			status, waived_by, waived_reason, decided_at, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err = r.db.ExecContext(ctx, query,
		g.ID,
		g.WorkspaceID,
		g.Name,
		g.ReleaseID,
		g.ChangeID,
		g.MinPassRate,
		g.RequireSecuritySignoff,
		g.RequirePerformanceSignoff,
		g.Status,
		g.WaivedBy,
		g.WaivedReason,
		g.DecidedAt,
		g.CreatedAt,
		g.UpdatedAt,
	)
	return err
}

// GetQualityGateByID retrieves a quality gate by ID
func (r *sqlTestingRepository) GetQualityGateByID(ctx context.Context, id string) (*model.QualityGate, error) {
	query := `
		SELECT id, workspace_id, name, release_id, change_id,
		       min_pass_rate, require_security_signoff, require_performance_signoff,
		       status, waived_by, waived_reason, decided_at, created_at, updated_at
		FROM quality_gates
		WHERE id = ?`

	var g model.QualityGate
	var relID, chgID, waivedBy, waivedReason sql.NullString
	var decidedAt sql.NullTime
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&g.ID,
		&g.WorkspaceID,
		&g.Name,
		&relID,
		&chgID,
		&g.MinPassRate,
		&g.RequireSecuritySignoff,
		&g.RequirePerformanceSignoff,
		&g.Status,
		&waivedBy,
		&waivedReason,
		&decidedAt,
		&g.CreatedAt,
		&g.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrQualityGateNotFound
		}
		return nil, err
	}
	if relID.Valid {
		g.ReleaseID = relID.String
	}
	if chgID.Valid {
		g.ChangeID = chgID.String
	}
	if waivedBy.Valid {
		g.WaivedBy = waivedBy.String
	}
	if waivedReason.Valid {
		g.WaivedReason = waivedReason.String
	}
	if decidedAt.Valid {
		g.DecidedAt = &decidedAt.Time
	}
	return &g, nil
}

// UpdateQualityGate updates an existing quality gate
func (r *sqlTestingRepository) UpdateQualityGate(ctx context.Context, g *model.QualityGate) error {
	g.UpdatedAt = time.Now().UTC()

	query := `
		UPDATE quality_gates
		SET name = ?, release_id = ?, change_id = ?, min_pass_rate = ?,
		    require_security_signoff = ?, require_performance_signoff = ?,
		    status = ?, waived_by = ?, waived_reason = ?, decided_at = ?, updated_at = ?
		WHERE id = ?`

	res, err := r.db.ExecContext(ctx, query,
		g.Name,
		g.ReleaseID,
		g.ChangeID,
		g.MinPassRate,
		g.RequireSecuritySignoff,
		g.RequirePerformanceSignoff,
		g.Status,
		g.WaivedBy,
		g.WaivedReason,
		g.DecidedAt,
		g.UpdatedAt,
		g.ID,
	)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrQualityGateNotFound
	}
	return nil
}

// DeleteQualityGate removes a quality gate
func (r *sqlTestingRepository) DeleteQualityGate(ctx context.Context, id string) error {
	res, err := r.db.ExecContext(ctx, "DELETE FROM quality_gates WHERE id = ?", id)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrQualityGateNotFound
	}
	return nil
}

// ListQualityGates queries quality gates based on filter criteria
func (r *sqlTestingRepository) ListQualityGates(ctx context.Context, filter model.QualityGateFilter) ([]model.QualityGate, int, error) {
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
	if filter.Status != "" {
		conditions = append(conditions, "status = ?")
		args = append(args, filter.Status)
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM quality_gates %s", whereClause)
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
		SELECT id, workspace_id, name, release_id, change_id,
		       min_pass_rate, require_security_signoff, require_performance_signoff,
		       status, waived_by, waived_reason, decided_at, created_at, updated_at
		FROM quality_gates
		%s
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?`, whereClause)

	queryArgs := append(args, limit, offset)
	rows, err := r.db.QueryContext(ctx, selectQuery, queryArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var gates []model.QualityGate
	for rows.Next() {
		var g model.QualityGate
		var relID, chgID, waivedBy, waivedReason sql.NullString
		var decidedAt sql.NullTime
		if err := rows.Scan(
			&g.ID,
			&g.WorkspaceID,
			&g.Name,
			&relID,
			&chgID,
			&g.MinPassRate,
			&g.RequireSecuritySignoff,
			&g.RequirePerformanceSignoff,
			&g.Status,
			&waivedBy,
			&waivedReason,
			&decidedAt,
			&g.CreatedAt,
			&g.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		if relID.Valid {
			g.ReleaseID = relID.String
		}
		if chgID.Valid {
			g.ChangeID = chgID.String
		}
		if waivedBy.Valid {
			g.WaivedBy = waivedBy.String
		}
		if waivedReason.Valid {
			g.WaivedReason = waivedReason.String
		}
		if decidedAt.Valid {
			g.DecidedAt = &decidedAt.Time
		}
		gates = append(gates, g)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return gates, total, nil
}

// CreateDefectLink records a defect linkage for a test run
func (r *sqlTestingRepository) CreateDefectLink(ctx context.Context, d *model.ValidationDefect) error {
	if d.ID == "" {
		d.ID = uuid.New().String()
	}
	now := time.Now().UTC()
	if d.LoggedAt.IsZero() {
		d.LoggedAt = now
	}
	if d.Status == "" {
		d.Status = model.DefectStatusOpen
	}
	if d.Severity == "" {
		d.Severity = model.TestCaseSeverityMedium
	}

	query := `
		INSERT INTO validation_defects (
			id, workspace_id, run_id, incident_id, problem_id,
			severity, title, status, logged_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := r.db.ExecContext(ctx, query,
		d.ID,
		d.WorkspaceID,
		d.RunID,
		d.IncidentID,
		d.ProblemID,
		d.Severity,
		d.Title,
		d.Status,
		d.LoggedAt,
	)
	return err
}

// ListDefectLinks queries validation defects based on filter criteria
func (r *sqlTestingRepository) ListDefectLinks(ctx context.Context, filter model.DefectLinkFilter) ([]model.ValidationDefect, int, error) {
	var conditions []string
	var args []interface{}

	if filter.WorkspaceID != "" {
		conditions = append(conditions, "workspace_id = ?")
		args = append(args, filter.WorkspaceID)
	}
	if filter.RunID != "" {
		conditions = append(conditions, "run_id = ?")
		args = append(args, filter.RunID)
	}
	if filter.IncidentID != "" {
		conditions = append(conditions, "incident_id = ?")
		args = append(args, filter.IncidentID)
	}
	if filter.ProblemID != "" {
		conditions = append(conditions, "problem_id = ?")
		args = append(args, filter.ProblemID)
	}
	if filter.Status != "" {
		conditions = append(conditions, "status = ?")
		args = append(args, filter.Status)
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM validation_defects %s", whereClause)
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
		SELECT id, workspace_id, run_id, incident_id, problem_id,
		       severity, title, status, logged_at
		FROM validation_defects
		%s
		ORDER BY logged_at DESC
		LIMIT ? OFFSET ?`, whereClause)

	queryArgs := append(args, limit, offset)
	rows, err := r.db.QueryContext(ctx, selectQuery, queryArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var defects []model.ValidationDefect
	for rows.Next() {
		var d model.ValidationDefect
		var incID, probID sql.NullString
		if err := rows.Scan(
			&d.ID,
			&d.WorkspaceID,
			&d.RunID,
			&incID,
			&probID,
			&d.Severity,
			&d.Title,
			&d.Status,
			&d.LoggedAt,
		); err != nil {
			return nil, 0, err
		}
		if incID.Valid {
			d.IncidentID = incID.String
		}
		if probID.Valid {
			d.ProblemID = probID.String
		}
		defects = append(defects, d)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return defects, total, nil
}

// GetTestingAnalytics computes quality posture, pass rates, and validation KPIs
func (r *sqlTestingRepository) GetTestingAnalytics(ctx context.Context, workspaceID string) (*model.TestingAnalytics, error) {
	analytics := &model.TestingAnalytics{
		TestTypeDistribution:    make(map[string]int),
		EnvironmentDistribution: make(map[string]int),
	}

	// 1. Total plans
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM validation_test_plans WHERE workspace_id = ?", workspaceID).Scan(&analytics.TotalPlans)
	if err != nil {
		return nil, err
	}

	// 2. Total cases
	err = r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM validation_test_cases WHERE workspace_id = ?", workspaceID).Scan(&analytics.TotalCases)
	if err != nil {
		return nil, err
	}

	// 3. Total runs
	err = r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM validation_test_runs WHERE workspace_id = ?", workspaceID).Scan(&analytics.TotalRuns)
	if err != nil {
		return nil, err
	}

	// 4. Test run breakdown by status
	runRows, err := r.db.QueryContext(ctx, "SELECT status, COUNT(*) FROM validation_test_runs WHERE workspace_id = ?", workspaceID)
	if err == nil {
		defer runRows.Close()
		for runRows.Next() {
			var st string
			var count int
			if err := runRows.Scan(&st, &count); err == nil {
				switch st {
				case model.TestRunStatusPassed:
					analytics.PassedRuns = count
				case model.TestRunStatusFailed:
					analytics.FailedRuns = count
				case model.TestRunStatusBlocked:
					analytics.BlockedRuns = count
				case model.TestRunStatusSkipped:
					analytics.SkippedRuns = count
				}
			}
		}
	}

	// 5. Overall pass rate
	if analytics.TotalRuns > 0 {
		analytics.OverallPassRate = (float64(analytics.PassedRuns) / float64(analytics.TotalRuns)) * 100.0
	} else {
		analytics.OverallPassRate = 100.0
	}

	// 6. Defect density
	var totalDefects int
	_ = r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM validation_defects WHERE workspace_id = ?", workspaceID).Scan(&totalDefects)
	if analytics.TotalRuns > 0 {
		analytics.DefectDensity = float64(totalDefects) / float64(analytics.TotalRuns)
	}

	// 7. Quality gate compliance
	var totalGates, compliantGates int
	_ = r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM quality_gates WHERE workspace_id = ?", workspaceID).Scan(&totalGates)
	_ = r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM quality_gates WHERE workspace_id = ? AND status IN ('passed', 'waived')", workspaceID).Scan(&compliantGates)
	if totalGates > 0 {
		analytics.QualityGateCompliance = (float64(compliantGates) / float64(totalGates)) * 100.0
	} else {
		analytics.QualityGateCompliance = 100.0
	}

	// 8. Test type distribution
	typeRows, err := r.db.QueryContext(ctx, "SELECT test_type, COUNT(*) FROM validation_test_plans WHERE workspace_id = ? GROUP BY test_type", workspaceID)
	if err == nil {
		defer typeRows.Close()
		for typeRows.Next() {
			var tt string
			var count int
			if err := typeRows.Scan(&tt, &count); err == nil {
				analytics.TestTypeDistribution[tt] = count
			}
		}
	}

	// 9. Environment distribution
	envRows, err := r.db.QueryContext(ctx, "SELECT environment, COUNT(*) FROM validation_test_plans WHERE workspace_id = ? GROUP BY environment", workspaceID)
	if err == nil {
		defer envRows.Close()
		for envRows.Next() {
			var env string
			var count int
			if err := envRows.Scan(&env, &count); err == nil {
				analytics.EnvironmentDistribution[env] = count
			}
		}
	}

	// 10. Composite Quality Health Index (0 - 100)
	health := (analytics.OverallPassRate * 0.7) + (analytics.QualityGateCompliance * 0.3)
	health -= float64(analytics.FailedRuns) * 3.0
	if health < 0.0 {
		health = 0.0
	}
	if health > 100.0 {
		health = 100.0
	}
	analytics.QualityHealthIndex = health

	return analytics, nil
}
