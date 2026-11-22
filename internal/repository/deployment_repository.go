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
// Gates implementation
func (r *sqlDeploymentRepository) CreateGate(ctx context.Context, g *model.DeploymentGate) error {
	if g.ID == "" {
		g.ID = uuid.New().String()
	}
	now := time.Now().UTC()
	if g.CreatedAt.IsZero() {
		g.CreatedAt = now
	}
	g.UpdatedAt = now
	if g.Status == "" {
		g.Status = model.GateStatusPending
	}

	query := `
		INSERT INTO deployment_gates (
			id, workspace_id, run_id, gate_type, gate_name, status,
			evaluator_id, evaluated_at, waiver_reason, evidence_details,
			created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err := r.db.ExecContext(ctx, query,
		g.ID, g.WorkspaceID, g.RunID, g.GateType, g.GateName, g.Status,
		g.EvaluatorID, g.EvaluatedAt, g.WaiverReason, g.EvidenceDetails,
		g.CreatedAt, g.UpdatedAt,
	)
	return err
}

func (r *sqlDeploymentRepository) GetGateByID(ctx context.Context, id string) (*model.DeploymentGate, error) {
	query := `
		SELECT id, workspace_id, run_id, gate_type, gate_name, status,
		       evaluator_id, evaluated_at, waiver_reason, evidence_details,
		       created_at, updated_at
		FROM deployment_gates
		WHERE id = ?
	`
	var g model.DeploymentGate
	var evaluator, waiver, evidence sql.NullString
	var evaluatedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&g.ID, &g.WorkspaceID, &g.RunID, &g.GateType, &g.GateName, &g.Status,
		&evaluator, &evaluatedAt, &waiver, &evidence,
		&g.CreatedAt, &g.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrDeploymentGateNotFound
		}
		return nil, err
	}
	if evaluatedAt.Valid {
		g.EvaluatedAt = &evaluatedAt.Time
	}
	g.EvaluatorID = evaluator.String
	g.WaiverReason = waiver.String
	g.EvidenceDetails = evidence.String
	return &g, nil
}

func (r *sqlDeploymentRepository) UpdateGate(ctx context.Context, g *model.DeploymentGate) error {
	g.UpdatedAt = time.Now().UTC()
	query := `
		UPDATE deployment_gates
		SET status = ?, evaluator_id = ?, evaluated_at = ?, waiver_reason = ?,
		    evidence_details = ?, updated_at = ?
		WHERE id = ?
	`
	res, err := r.db.ExecContext(ctx, query,
		g.Status, g.EvaluatorID, g.EvaluatedAt, g.WaiverReason,
		g.EvidenceDetails, g.UpdatedAt, g.ID,
	)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrDeploymentGateNotFound
	}
	return nil
}

func (r *sqlDeploymentRepository) ListGates(ctx context.Context, filter model.DeploymentGateFilter) ([]model.DeploymentGate, int, error) {
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
	if filter.GateType != "" {
		conditions = append(conditions, "gate_type = ?")
		args = append(args, filter.GateType)
	}
	if filter.Status != "" {
		conditions = append(conditions, "status = ?")
		args = append(args, filter.Status)
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM deployment_gates %s", whereClause)
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
		SELECT id, workspace_id, run_id, gate_type, gate_name, status,
		       evaluator_id, evaluated_at, waiver_reason, evidence_details,
		       created_at, updated_at
		FROM deployment_gates
		%s
		ORDER BY created_at ASC
		LIMIT ? OFFSET ?
	`, whereClause)

	argsWithPaging := append(args, limit, offset)
	rows, err := r.db.QueryContext(ctx, query, argsWithPaging...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var gates []model.DeploymentGate
	for rows.Next() {
		var g model.DeploymentGate
		var evaluator, waiver, evidence sql.NullString
		var evaluatedAt sql.NullTime

		err := rows.Scan(
			&g.ID, &g.WorkspaceID, &g.RunID, &g.GateType, &g.GateName, &g.Status,
			&evaluator, &evaluatedAt, &waiver, &evidence,
			&g.CreatedAt, &g.UpdatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		if evaluatedAt.Valid {
			g.EvaluatedAt = &evaluatedAt.Time
		}
		g.EvaluatorID = evaluator.String
		g.WaiverReason = waiver.String
		g.EvidenceDetails = evidence.String
		gates = append(gates, g)
	}
	return gates, total, rows.Err()
}

// Steps implementation
func (r *sqlDeploymentRepository) CreateStep(ctx context.Context, s *model.DeploymentStep) error {
	if s.ID == "" {
		s.ID = uuid.New().String()
	}
	now := time.Now().UTC()
	if s.CreatedAt.IsZero() {
		s.CreatedAt = now
	}
	s.UpdatedAt = now
	if s.Status == "" {
		s.Status = model.StepStatusPending
	}

	query := `
		INSERT INTO deployment_steps (
			id, workspace_id, run_id, step_order, step_name, traffic_percentage,
			soak_time_minutes, error_rate_threshold, actual_error_rate, status,
			started_at, completed_at, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err := r.db.ExecContext(ctx, query,
		s.ID, s.WorkspaceID, s.RunID, s.StepOrder, s.StepName, s.TrafficPercentage,
		s.SoakTimeMinutes, s.ErrorRateThreshold, s.ActualErrorRate, s.Status,
		s.StartedAt, s.CompletedAt, s.CreatedAt, s.UpdatedAt,
	)
	return err
}

func (r *sqlDeploymentRepository) GetStepByID(ctx context.Context, id string) (*model.DeploymentStep, error) {
	query := `
		SELECT id, workspace_id, run_id, step_order, step_name, traffic_percentage,
		       soak_time_minutes, error_rate_threshold, actual_error_rate, status,
		       started_at, completed_at, created_at, updated_at
		FROM deployment_steps
		WHERE id = ?
	`
	var s model.DeploymentStep
	var startedAt, completedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&s.ID, &s.WorkspaceID, &s.RunID, &s.StepOrder, &s.StepName, &s.TrafficPercentage,
		&s.SoakTimeMinutes, &s.ErrorRateThreshold, &s.ActualErrorRate, &s.Status,
		&startedAt, &completedAt, &s.CreatedAt, &s.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrDeploymentStepNotFound
		}
		return nil, err
	}
	if startedAt.Valid {
		s.StartedAt = &startedAt.Time
	}
	if completedAt.Valid {
		s.CompletedAt = &completedAt.Time
	}
	return &s, nil
}

func (r *sqlDeploymentRepository) UpdateStep(ctx context.Context, s *model.DeploymentStep) error {
	s.UpdatedAt = time.Now().UTC()
	query := `
		UPDATE deployment_steps
		SET step_order = ?, step_name = ?, traffic_percentage = ?, soak_time_minutes = ?,
		    error_rate_threshold = ?, actual_error_rate = ?, status = ?,
		    started_at = ?, completed_at = ?, updated_at = ?
		WHERE id = ?
	`
	res, err := r.db.ExecContext(ctx, query,
		s.StepOrder, s.StepName, s.TrafficPercentage, s.SoakTimeMinutes,
		s.ErrorRateThreshold, s.ActualErrorRate, s.Status,
		s.StartedAt, s.CompletedAt, s.UpdatedAt, s.ID,
	)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrDeploymentStepNotFound
	}
	return nil
}

func (r *sqlDeploymentRepository) ListSteps(ctx context.Context, filter model.DeploymentStepFilter) ([]model.DeploymentStep, int, error) {
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
	if filter.Status != "" {
		conditions = append(conditions, "status = ?")
		args = append(args, filter.Status)
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM deployment_steps %s", whereClause)
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
		SELECT id, workspace_id, run_id, step_order, step_name, traffic_percentage,
		       soak_time_minutes, error_rate_threshold, actual_error_rate, status,
		       started_at, completed_at, created_at, updated_at
		FROM deployment_steps
		%s
		ORDER BY step_order ASC
		LIMIT ? OFFSET ?
	`, whereClause)

	argsWithPaging := append(args, limit, offset)
	rows, err := r.db.QueryContext(ctx, query, argsWithPaging...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var steps []model.DeploymentStep
	for rows.Next() {
		var s model.DeploymentStep
		var startedAt, completedAt sql.NullTime

		err := rows.Scan(
			&s.ID, &s.WorkspaceID, &s.RunID, &s.StepOrder, &s.StepName, &s.TrafficPercentage,
			&s.SoakTimeMinutes, &s.ErrorRateThreshold, &s.ActualErrorRate, &s.Status,
			&startedAt, &completedAt, &s.CreatedAt, &s.UpdatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		if startedAt.Valid {
			s.StartedAt = &startedAt.Time
		}
		if completedAt.Valid {
			s.CompletedAt = &completedAt.Time
		}
		steps = append(steps, s)
	}
	return steps, total, rows.Err()
}
// Rollback Policies implementation
func (r *sqlDeploymentRepository) CreateRollbackPolicy(ctx context.Context, p *model.RollbackPolicy) error {
	if p.ID == "" {
		p.ID = uuid.New().String()
	}
	now := time.Now().UTC()
	if p.CreatedAt.IsZero() {
		p.CreatedAt = now
	}
	p.UpdatedAt = now
	if p.ExecutionStatus == "" {
		p.ExecutionStatus = model.RollbackStatusTriggered
	}

	query := `
		INSERT INTO deployment_rollbacks (
			id, workspace_id, pipeline_id, run_id, trigger_condition,
			threshold_value, target_rollback_version, execution_status,
			triggered_at, completed_at, rollback_logs, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err := r.db.ExecContext(ctx, query,
		p.ID, p.WorkspaceID, p.PipelineID, p.RunID, p.TriggerCondition,
		p.ThresholdValue, p.TargetRollbackVersion, p.ExecutionStatus,
		p.TriggeredAt, p.CompletedAt, p.RollbackLogs, p.CreatedAt, p.UpdatedAt,
	)
	return err
}

func (r *sqlDeploymentRepository) GetRollbackPolicyByID(ctx context.Context, id string) (*model.RollbackPolicy, error) {
	query := `
		SELECT id, workspace_id, pipeline_id, run_id, trigger_condition,
		       threshold_value, target_rollback_version, execution_status,
		       triggered_at, completed_at, rollback_logs, created_at, updated_at
		FROM deployment_rollbacks
		WHERE id = ?
	`
	var p model.RollbackPolicy
	var runID, logs sql.NullString
	var triggeredAt, completedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&p.ID, &p.WorkspaceID, &p.PipelineID, &runID, &p.TriggerCondition,
		&p.ThresholdValue, &p.TargetRollbackVersion, &p.ExecutionStatus,
		&triggeredAt, &completedAt, &logs, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrRollbackPolicyNotFound
		}
		return nil, err
	}
	p.RunID = runID.String
	p.RollbackLogs = logs.String
	if triggeredAt.Valid {
		p.TriggeredAt = &triggeredAt.Time
	}
	if completedAt.Valid {
		p.CompletedAt = &completedAt.Time
	}
	return &p, nil
}

func (r *sqlDeploymentRepository) UpdateRollbackPolicy(ctx context.Context, p *model.RollbackPolicy) error {
	p.UpdatedAt = time.Now().UTC()
	query := `
		UPDATE deployment_rollbacks
		SET execution_status = ?, triggered_at = ?, completed_at = ?,
		    rollback_logs = ?, updated_at = ?
		WHERE id = ?
	`
	res, err := r.db.ExecContext(ctx, query,
		p.ExecutionStatus, p.TriggeredAt, p.CompletedAt,
		p.RollbackLogs, p.UpdatedAt, p.ID,
	)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrRollbackPolicyNotFound
	}
	return nil
}

func (r *sqlDeploymentRepository) ListRollbackPolicies(ctx context.Context, filter model.RollbackPolicyFilter) ([]model.RollbackPolicy, int, error) {
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
	if filter.ExecutionStatus != "" {
		conditions = append(conditions, "execution_status = ?")
		args = append(args, filter.ExecutionStatus)
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM deployment_rollbacks %s", whereClause)
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
		SELECT id, workspace_id, pipeline_id, run_id, trigger_condition,
		       threshold_value, target_rollback_version, execution_status,
		       triggered_at, completed_at, rollback_logs, created_at, updated_at
		FROM deployment_rollbacks
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

	var policies []model.RollbackPolicy
	for rows.Next() {
		var p model.RollbackPolicy
		var runID, logs sql.NullString
		var triggeredAt, completedAt sql.NullTime

		err := rows.Scan(
			&p.ID, &p.WorkspaceID, &p.PipelineID, &runID, &p.TriggerCondition,
			&p.ThresholdValue, &p.TargetRollbackVersion, &p.ExecutionStatus,
			&triggeredAt, &completedAt, &logs, &p.CreatedAt, &p.UpdatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		p.RunID = runID.String
		p.RollbackLogs = logs.String
		if triggeredAt.Valid {
			p.TriggeredAt = &triggeredAt.Time
		}
		if completedAt.Valid {
			p.CompletedAt = &completedAt.Time
		}
		policies = append(policies, p)
	}
	return policies, total, rows.Err()
}
// Posture Analytics implementation
func (r *sqlDeploymentRepository) GetDeploymentAnalytics(ctx context.Context, workspaceID string) (*model.DeploymentAnalytics, error) {
	analytics := &model.DeploymentAnalytics{
		EnvironmentDistribution: make(map[string]int),
	}

	// 1. Pipelines counts
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM deployment_pipelines WHERE workspace_id = ?`, workspaceID).Scan(&analytics.TotalPipelines)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	_ = r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM deployment_pipelines WHERE workspace_id = ? AND status = ?`, workspaceID, model.PipelineStatusActive).Scan(&analytics.ActivePipelines)

	// 2. Runs counts by status
	_ = r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM deployment_runs WHERE workspace_id = ?`, workspaceID).Scan(&analytics.TotalRuns)
	_ = r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM deployment_runs WHERE workspace_id = ? AND status = ?`, workspaceID, model.RunStatusSuccessful).Scan(&analytics.SuccessfulRuns)
	_ = r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM deployment_runs WHERE workspace_id = ? AND status = ?`, workspaceID, model.RunStatusFailed).Scan(&analytics.FailedRuns)
	_ = r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM deployment_runs WHERE workspace_id = ? AND status = ?`, workspaceID, model.RunStatusRolledBack).Scan(&analytics.RolledBackRuns)

	// 3. Rollbacks count
	_ = r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM deployment_rollbacks WHERE workspace_id = ?`, workspaceID).Scan(&analytics.RollbacksCount)

	// 4. Environment distribution
	envRows, err := r.db.QueryContext(ctx, `SELECT environment, COUNT(*) FROM deployment_runs WHERE workspace_id = ? GROUP BY environment`, workspaceID)
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

	// 5. DORA Metrics computation
	// Deployment frequency per week (runs in the last 30 days / 4.28 weeks)
	thirtyDaysAgo := time.Now().UTC().AddDate(0, 0, -30)
	var recentRuns int
	_ = r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM deployment_runs WHERE workspace_id = ? AND started_at >= ?`, workspaceID, thirtyDaysAgo).Scan(&recentRuns)
	analytics.DeploymentFrequencyPerWeek = float64(recentRuns) / (30.0 / 7.0)

	// Lead time for changes (average duration_seconds / 3600 for successful runs)
	var avgDuration sql.NullFloat64
	_ = r.db.QueryRowContext(ctx, `SELECT AVG(duration_seconds) FROM deployment_runs WHERE workspace_id = ? AND status = ?`, workspaceID, model.RunStatusSuccessful).Scan(&avgDuration)
	if avgDuration.Valid {
		analytics.LeadTimeForChangesHours = avgDuration.Float64 / 3600.0
	}

	// Change failure rate: (failed + rolled_back) / total * 100
	if analytics.TotalRuns > 0 {
		unsuccessful := analytics.FailedRuns + analytics.RolledBackRuns
		analytics.ChangeFailureRate = (float64(unsuccessful) / float64(analytics.TotalRuns)) * 100.0
	}

	// MTTR (Mean time to recover): average duration of rolled back runs in minutes
	var avgRecoveryDuration sql.NullFloat64
	_ = r.db.QueryRowContext(ctx, `SELECT AVG(duration_seconds) FROM deployment_runs WHERE workspace_id = ? AND status = ?`, workspaceID, model.RunStatusRolledBack).Scan(&avgRecoveryDuration)
	if avgRecoveryDuration.Valid {
		analytics.MTTRMinutes = avgRecoveryDuration.Float64 / 60.0
	}

	// 6. Composite Deployment Health Index (0 - 100)
	health := 100.0
	if analytics.TotalRuns > 0 {
		// Penalty for change failure rate (up to 50 pts)
		health -= (analytics.ChangeFailureRate * 0.5)
	}
	if analytics.TotalPipelines > 0 {
		// Pipeline active ratio bonus (up to 20 pts)
		activeRatio := float64(analytics.ActivePipelines) / float64(analytics.TotalPipelines)
		health -= (1.0 - activeRatio) * 20.0
	}
	if analytics.RollbacksCount > 0 {
		// Penalty for excessive rollbacks
		penalty := float64(analytics.RollbacksCount) * 2.0
		if penalty > 20.0 {
			penalty = 20.0
		}
		health -= penalty
	}
	if health < 0.0 {
		health = 0.0
	}
	if health > 100.0 {
		health = 100.0
	}
	analytics.DeploymentHealthIndex = health

	return analytics, nil
}
