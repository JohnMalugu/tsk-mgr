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
	ErrDeploymentPipelineNotFound     = errors.New("deployment pipeline not found")
	ErrDeploymentPipelineCodeConflict = errors.New("deployment pipeline with this code already exists in workspace")
	ErrDeploymentRunNotFound          = errors.New("deployment run not found")
	ErrDeploymentGateNotFound         = errors.New("deployment gate not found")
	ErrDeploymentStepNotFound         = errors.New("deployment step not found")
	ErrRollbackPolicyNotFound         = errors.New("rollback policy not found")
)

// DeploymentRepository defines data access methods for ITIL Release & Deployment Management
type DeploymentRepository interface {
	// Pipelines
	CreatePipeline(ctx context.Context, p *model.DeploymentPipeline) error
	GetPipelineByID(ctx context.Context, id string) (*model.DeploymentPipeline, error)
	GetPipelineByCode(ctx context.Context, workspaceID, code string) (*model.DeploymentPipeline, error)
	UpdatePipeline(ctx context.Context, p *model.DeploymentPipeline) error
	DeletePipeline(ctx context.Context, id string) error
	ListPipelines(ctx context.Context, filter model.DeploymentPipelineFilter) ([]model.DeploymentPipeline, int, error)

	// Runs
	CreateRun(ctx context.Context, r *model.DeploymentRun) error
	GetRunByID(ctx context.Context, id string) (*model.DeploymentRun, error)
	UpdateRun(ctx context.Context, r *model.DeploymentRun) error
	ListRuns(ctx context.Context, filter model.DeploymentRunFilter) ([]model.DeploymentRun, int, error)

	// Gates
	CreateGate(ctx context.Context, g *model.DeploymentGate) error
	GetGateByID(ctx context.Context, id string) (*model.DeploymentGate, error)
	UpdateGate(ctx context.Context, g *model.DeploymentGate) error
	ListGates(ctx context.Context, filter model.DeploymentGateFilter) ([]model.DeploymentGate, int, error)

	// Steps
	CreateStep(ctx context.Context, s *model.DeploymentStep) error
	GetStepByID(ctx context.Context, id string) (*model.DeploymentStep, error)
	UpdateStep(ctx context.Context, s *model.DeploymentStep) error
	ListSteps(ctx context.Context, filter model.DeploymentStepFilter) ([]model.DeploymentStep, int, error)

	// Rollback Policies
	CreateRollbackPolicy(ctx context.Context, p *model.RollbackPolicy) error
	GetRollbackPolicyByID(ctx context.Context, id string) (*model.RollbackPolicy, error)
	UpdateRollbackPolicy(ctx context.Context, p *model.RollbackPolicy) error
	ListRollbackPolicies(ctx context.Context, filter model.RollbackPolicyFilter) ([]model.RollbackPolicy, int, error)

	// Posture Analytics
	GetDeploymentAnalytics(ctx context.Context, workspaceID string) (*model.DeploymentAnalytics, error)
}

type sqlDeploymentRepository struct {
	db *sql.DB
}

// NewDeploymentRepository initializes a SQLite deployment repository
func NewDeploymentRepository(db *sql.DB) DeploymentRepository {
	return &sqlDeploymentRepository{db: db}
}

// Pipelines implementation
func (r *sqlDeploymentRepository) CreatePipeline(ctx context.Context, p *model.DeploymentPipeline) error {
	if p.ID == "" {
		p.ID = uuid.New().String()
	}
	now := time.Now().UTC()
	if p.CreatedAt.IsZero() {
		p.CreatedAt = now
	}
	p.UpdatedAt = now

	// Check for unique code in workspace
	var count int
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM deployment_pipelines WHERE workspace_id = ? AND pipeline_code = ?`, p.WorkspaceID, p.PipelineCode).Scan(&count)
	if err == nil && count > 0 {
		return ErrDeploymentPipelineCodeConflict
	}

	query := `
		INSERT INTO deployment_pipelines (
			id, workspace_id, pipeline_code, name, description,
			target_environment, deployment_strategy, auto_rollback_enabled,
			linked_ci_id, status, created_by_user_id, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err = r.db.ExecContext(ctx, query,
		p.ID, p.WorkspaceID, p.PipelineCode, p.Name, p.Description,
		p.TargetEnvironment, p.DeploymentStrategy, p.AutoRollbackEnabled,
		p.LinkedCIID, p.Status, p.CreatedByUserID, p.CreatedAt, p.UpdatedAt,
	)
	return err
}

func (r *sqlDeploymentRepository) GetPipelineByID(ctx context.Context, id string) (*model.DeploymentPipeline, error) {
	query := `
		SELECT id, workspace_id, pipeline_code, name, description,
		       target_environment, deployment_strategy, auto_rollback_enabled,
		       linked_ci_id, status, created_by_user_id, created_at, updated_at
		FROM deployment_pipelines
		WHERE id = ?
	`
	var p model.DeploymentPipeline
	var desc, ciID, user sql.NullString
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&p.ID, &p.WorkspaceID, &p.PipelineCode, &p.Name, &desc,
		&p.TargetEnvironment, &p.DeploymentStrategy, &p.AutoRollbackEnabled,
		&ciID, &p.Status, &user, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrDeploymentPipelineNotFound
		}
		return nil, err
	}
	p.Description = desc.String
	p.LinkedCIID = ciID.String
	p.CreatedByUserID = user.String
	return &p, nil
}

func (r *sqlDeploymentRepository) GetPipelineByCode(ctx context.Context, workspaceID, code string) (*model.DeploymentPipeline, error) {
	query := `
		SELECT id, workspace_id, pipeline_code, name, description,
		       target_environment, deployment_strategy, auto_rollback_enabled,
		       linked_ci_id, status, created_by_user_id, created_at, updated_at
		FROM deployment_pipelines
		WHERE workspace_id = ? AND pipeline_code = ?
	`
	var p model.DeploymentPipeline
	var desc, ciID, user sql.NullString
	err := r.db.QueryRowContext(ctx, query, workspaceID, code).Scan(
		&p.ID, &p.WorkspaceID, &p.PipelineCode, &p.Name, &desc,
		&p.TargetEnvironment, &p.DeploymentStrategy, &p.AutoRollbackEnabled,
		&ciID, &p.Status, &user, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrDeploymentPipelineNotFound
		}
		return nil, err
	}
	p.Description = desc.String
	p.LinkedCIID = ciID.String
	p.CreatedByUserID = user.String
	return &p, nil
}

func (r *sqlDeploymentRepository) UpdatePipeline(ctx context.Context, p *model.DeploymentPipeline) error {
	p.UpdatedAt = time.Now().UTC()
	query := `
		UPDATE deployment_pipelines
		SET name = ?, description = ?, target_environment = ?, deployment_strategy = ?,
		    auto_rollback_enabled = ?, linked_ci_id = ?, status = ?, updated_at = ?
		WHERE id = ?
	`
	res, err := r.db.ExecContext(ctx, query,
		p.Name, p.Description, p.TargetEnvironment, p.DeploymentStrategy,
		p.AutoRollbackEnabled, p.LinkedCIID, p.Status, p.UpdatedAt, p.ID,
	)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrDeploymentPipelineNotFound
	}
	return nil
}

func (r *sqlDeploymentRepository) DeletePipeline(ctx context.Context, id string) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM deployment_pipelines WHERE id = ?`, id)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrDeploymentPipelineNotFound
	}
	return nil
}

func (r *sqlDeploymentRepository) ListPipelines(ctx context.Context, filter model.DeploymentPipelineFilter) ([]model.DeploymentPipeline, int, error) {
	var conditions []string
	var args []interface{}

	if filter.WorkspaceID != "" {
		conditions = append(conditions, "workspace_id = ?")
		args = append(args, filter.WorkspaceID)
	}
	if filter.TargetEnvironment != "" {
		conditions = append(conditions, "target_environment = ?")
		args = append(args, filter.TargetEnvironment)
	}
	if filter.DeploymentStrategy != "" {
		conditions = append(conditions, "deployment_strategy = ?")
		args = append(args, filter.DeploymentStrategy)
	}
	if filter.Status != "" {
		conditions = append(conditions, "status = ?")
		args = append(args, filter.Status)
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM deployment_pipelines %s", whereClause)
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

	query := fmt.Sprintf(`
		SELECT id, workspace_id, pipeline_code, name, description,
		       target_environment, deployment_strategy, auto_rollback_enabled,
		       linked_ci_id, status, created_by_user_id, created_at, updated_at
		FROM deployment_pipelines
		%s
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`, whereClause)

	argsWithPaging := append(args, limit, offset)
	rows, err := r.db.QueryContext(ctx, query, argsWithPaging...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var pipelines []model.DeploymentPipeline
	for rows.Next() {
		var p model.DeploymentPipeline
		var desc, ciID, user sql.NullString
		err := rows.Scan(
			&p.ID, &p.WorkspaceID, &p.PipelineCode, &p.Name, &desc,
			&p.TargetEnvironment, &p.DeploymentStrategy, &p.AutoRollbackEnabled,
			&ciID, &p.Status, &user, &p.CreatedAt, &p.UpdatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		p.Description = desc.String
		p.LinkedCIID = ciID.String
		p.CreatedByUserID = user.String
		pipelines = append(pipelines, p)
	}
	return pipelines, total, rows.Err()
}
// Runs implementation
func (r *sqlDeploymentRepository) CreateRun(ctx context.Context, run *model.DeploymentRun) error {
	if run.ID == "" {
		run.ID = uuid.New().String()
	}
	now := time.Now().UTC()
	if run.CreatedAt.IsZero() {
		run.CreatedAt = now
	}
	run.UpdatedAt = now
	if run.StartedAt.IsZero() {
		run.StartedAt = now
	}

	// Auto-increment run_number within pipeline
	var maxRun int64
	err := r.db.QueryRowContext(ctx, `SELECT COALESCE(MAX(run_number), 0) FROM deployment_runs WHERE pipeline_id = ?`, run.PipelineID).Scan(&maxRun)
	if err == nil {
		run.RunNumber = maxRun + 1
	} else {
		run.RunNumber = 1
	}

	query := `
		INSERT INTO deployment_runs (
			id, workspace_id, pipeline_id, run_number, release_version,
			environment, trigger_type, status, started_at, finished_at,
			duration_seconds, deployed_by_user_id, rollback_reason, commit_hash,
			changelog, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err = r.db.ExecContext(ctx, query,
		run.ID, run.WorkspaceID, run.PipelineID, run.RunNumber, run.ReleaseVersion,
		run.Environment, run.TriggerType, run.Status, run.StartedAt, run.FinishedAt,
		run.DurationSeconds, run.DeployedByUserID, run.RollbackReason, run.CommitHash,
		run.Changelog, run.CreatedAt, run.UpdatedAt,
	)
	return err
}

func (r *sqlDeploymentRepository) GetRunByID(ctx context.Context, id string) (*model.DeploymentRun, error) {
	query := `
		SELECT id, workspace_id, pipeline_id, run_number, release_version,
		       environment, trigger_type, status, started_at, finished_at,
		       duration_seconds, deployed_by_user_id, rollback_reason, commit_hash,
		       changelog, created_at, updated_at
		FROM deployment_runs
		WHERE id = ?
	`
	var run model.DeploymentRun
	var deployedBy, rollback, commit, changelog sql.NullString
	var finishedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&run.ID, &run.WorkspaceID, &run.PipelineID, &run.RunNumber, &run.ReleaseVersion,
		&run.Environment, &run.TriggerType, &run.Status, &run.StartedAt, &finishedAt,
		&run.DurationSeconds, &deployedBy, &rollback, &commit,
		&changelog, &run.CreatedAt, &run.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrDeploymentRunNotFound
		}
		return nil, err
	}
	if finishedAt.Valid {
		run.FinishedAt = &finishedAt.Time
	}
	run.DeployedByUserID = deployedBy.String
	run.RollbackReason = rollback.String
	run.CommitHash = commit.String
	run.Changelog = changelog.String
	return &run, nil
}

func (r *sqlDeploymentRepository) UpdateRun(ctx context.Context, run *model.DeploymentRun) error {
	run.UpdatedAt = time.Now().UTC()
	query := `
		UPDATE deployment_runs
		SET status = ?, finished_at = ?, duration_seconds = ?, rollback_reason = ?, updated_at = ?
		WHERE id = ?
	`
	res, err := r.db.ExecContext(ctx, query,
		run.Status, run.FinishedAt, run.DurationSeconds, run.RollbackReason, run.UpdatedAt, run.ID,
	)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrDeploymentRunNotFound
	}
	return nil
}

func (r *sqlDeploymentRepository) ListRuns(ctx context.Context, filter model.DeploymentRunFilter) ([]model.DeploymentRun, int, error) {
	var conditions []string
	var args []interface{}

	if filter.WorkspaceID != "" {
		conditions = append(conditions, "workspace_id = ?")
		args = append(args, filter.WorkspaceID)
	}
	if filter.PipelineID != "" {
		conditions = append(conditions, "pipeline_id = ?")
		args = append(args, filter.PipelineID)
	}
	if filter.Environment != "" {
		conditions = append(conditions, "environment = ?")
		args = append(args, filter.Environment)
	}
	if filter.Status != "" {
		conditions = append(conditions, "status = ?")
		args = append(args, filter.Status)
	}
	if filter.TriggerType != "" {
		conditions = append(conditions, "trigger_type = ?")
		args = append(args, filter.TriggerType)
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM deployment_runs %s", whereClause)
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

	query := fmt.Sprintf(`
		SELECT id, workspace_id, pipeline_id, run_number, release_version,
		       environment, trigger_type, status, started_at, finished_at,
		       duration_seconds, deployed_by_user_id, rollback_reason, commit_hash,
		       changelog, created_at, updated_at
		FROM deployment_runs
		%s
		ORDER BY started_at DESC
		LIMIT ? OFFSET ?
	`, whereClause)

	argsWithPaging := append(args, limit, offset)
	rows, err := r.db.QueryContext(ctx, query, argsWithPaging...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var runs []model.DeploymentRun
	for rows.Next() {
		var run model.DeploymentRun
		var deployedBy, rollback, commit, changelog sql.NullString
		var finishedAt sql.NullTime

		err := rows.Scan(
			&run.ID, &run.WorkspaceID, &run.PipelineID, &run.RunNumber, &run.ReleaseVersion,
			&run.Environment, &run.TriggerType, &run.Status, &run.StartedAt, &finishedAt,
			&run.DurationSeconds, &deployedBy, &rollback, &commit,
			&changelog, &run.CreatedAt, &run.UpdatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		if finishedAt.Valid {
			run.FinishedAt = &finishedAt.Time
		}
		run.DeployedByUserID = deployedBy.String
		run.RollbackReason = rollback.String
		run.CommitHash = commit.String
		run.Changelog = changelog.String
		runs = append(runs, run)
	}
	return runs, total, rows.Err()
}
func (r *sqlDeploymentRepository) CreateGate(ctx context.Context, g *model.DeploymentGate) error {
	return nil
}
func (r *sqlDeploymentRepository) GetGateByID(ctx context.Context, id string) (*model.DeploymentGate, error) {
	return nil, nil
}
func (r *sqlDeploymentRepository) UpdateGate(ctx context.Context, g *model.DeploymentGate) error {
	return nil
}
func (r *sqlDeploymentRepository) ListGates(ctx context.Context, filter model.DeploymentGateFilter) ([]model.DeploymentGate, int, error) {
	return nil, 0, nil
}
func (r *sqlDeploymentRepository) CreateStep(ctx context.Context, s *model.DeploymentStep) error {
	return nil
}
func (r *sqlDeploymentRepository) GetStepByID(ctx context.Context, id string) (*model.DeploymentStep, error) {
	return nil, nil
}
func (r *sqlDeploymentRepository) UpdateStep(ctx context.Context, s *model.DeploymentStep) error {
	return nil
}
func (r *sqlDeploymentRepository) ListSteps(ctx context.Context, filter model.DeploymentStepFilter) ([]model.DeploymentStep, int, error) {
	return nil, 0, nil
}
func (r *sqlDeploymentRepository) CreateRollbackPolicy(ctx context.Context, p *model.RollbackPolicy) error {
	return nil
}
func (r *sqlDeploymentRepository) GetRollbackPolicyByID(ctx context.Context, id string) (*model.RollbackPolicy, error) {
	return nil, nil
}
func (r *sqlDeploymentRepository) UpdateRollbackPolicy(ctx context.Context, p *model.RollbackPolicy) error {
	return nil
}
func (r *sqlDeploymentRepository) ListRollbackPolicies(ctx context.Context, filter model.RollbackPolicyFilter) ([]model.RollbackPolicy, int, error) {
	return nil, 0, nil
}
func (r *sqlDeploymentRepository) GetDeploymentAnalytics(ctx context.Context, workspaceID string) (*model.DeploymentAnalytics, error) {
	return nil, nil
}
