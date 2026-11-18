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
	ErrDRPlanNotFound            = model.ErrDRPlanNotFound
	ErrDRPlanCodeRequired        = model.ErrDRPlanCodeRequired
	ErrDRPlanCodeConflict        = model.ErrDRPlanCodeConflict
	ErrDRPlanNameRequired        = model.ErrDRPlanNameRequired
	ErrInvalidRecoveryTier       = model.ErrInvalidRecoveryTier
	ErrInvalidDRPlanStatus       = model.ErrInvalidDRPlanStatus
	ErrRecoveryObjectiveNotFound = model.ErrRecoveryObjectiveNotFound
	ErrServiceNameRequired       = model.ErrServiceNameRequired
	ErrInvalidRTOValue           = model.ErrInvalidRTOValue
	ErrInvalidRPOValue           = model.ErrInvalidRPOValue
	ErrInvalidAvailability       = model.ErrInvalidAvailability
	ErrInvalidDataLossTolerance  = model.ErrInvalidDataLossTolerance
	ErrBIANotFound               = model.ErrBIANotFound
	ErrInvalidMTDValue           = model.ErrInvalidMTDValue
	ErrInvalidOperationalImpact  = model.ErrInvalidOperationalImpact
	ErrInvalidRegulatoryImpact   = model.ErrInvalidRegulatoryImpact
	ErrInvalidCriticalWindow     = model.ErrInvalidCriticalWindow
	ErrRunbookNotFound           = model.ErrRunbookNotFound
	ErrRunbookStepNotFound       = model.ErrRunbookStepNotFound
	ErrRunbookTitleRequired      = model.ErrRunbookTitleRequired
	ErrStepNameRequired          = model.ErrStepNameRequired
	ErrInvalidExecutionMode      = model.ErrInvalidExecutionMode
	ErrInvalidActionType         = model.ErrInvalidActionType
	ErrInvalidExecutionRole      = model.ErrInvalidExecutionRole
	ErrDrillNotFound             = model.ErrDrillNotFound
	ErrDrillFindingNotFound      = model.ErrDrillFindingNotFound
	ErrInvalidDrillType          = model.ErrInvalidDrillType
	ErrInvalidDrillStatus        = model.ErrInvalidDrillStatus
	ErrDrillScenarioRequired     = model.ErrDrillScenarioRequired
	ErrFindingDescRequired       = model.ErrFindingDescRequired
	ErrInvalidFindingSeverity    = model.ErrInvalidFindingSeverity
	ErrInvalidFindingCategory    = model.ErrInvalidFindingCategory
	ErrInvalidDrillFindingStatus = model.ErrInvalidDrillFindingStatus
)

// ContinuityRepository defines the data access contract for ITIL Service Continuity and Disaster Recovery
type ContinuityRepository interface {
	// Disaster Recovery Plans
	CreateDRPlan(ctx context.Context, plan *model.DisasterRecoveryPlan) error
	GetDRPlanByID(ctx context.Context, id string) (*model.DisasterRecoveryPlan, error)
	GetDRPlanByCode(ctx context.Context, workspaceID string, planCode string) (*model.DisasterRecoveryPlan, error)
	UpdateDRPlan(ctx context.Context, plan *model.DisasterRecoveryPlan) error
	DeleteDRPlan(ctx context.Context, id string) error
	ListDRPlans(ctx context.Context, filter model.DRPlanFilterCriteria) ([]model.DisasterRecoveryPlan, int, error)

	// Recovery Objectives (RTO/RPO)
	CreateRecoveryObjective(ctx context.Context, obj *model.RecoveryObjective) error
	GetRecoveryObjectiveByID(ctx context.Context, id string) (*model.RecoveryObjective, error)
	UpdateRecoveryObjective(ctx context.Context, obj *model.RecoveryObjective) error
	DeleteRecoveryObjective(ctx context.Context, id string) error
	ListRecoveryObjectivesByPlan(ctx context.Context, planID string) ([]model.RecoveryObjective, error)
	ListRecoveryObjectives(ctx context.Context, workspaceID string) ([]model.RecoveryObjective, error)

	// Business Impact Analyses (BIA)
	CreateBIA(ctx context.Context, bia *model.BusinessImpactAnalysis) error
	GetBIAByID(ctx context.Context, id string) (*model.BusinessImpactAnalysis, error)
	UpdateBIA(ctx context.Context, bia *model.BusinessImpactAnalysis) error
	DeleteBIA(ctx context.Context, id string) error
	ListBIAsByPlan(ctx context.Context, planID string) ([]model.BusinessImpactAnalysis, error)
	ListBIAs(ctx context.Context, workspaceID string) ([]model.BusinessImpactAnalysis, error)

	// Recovery Runbooks & Steps
	CreateRunbook(ctx context.Context, rb *model.RecoveryRunbook) error
	GetRunbookByID(ctx context.Context, id string) (*model.RecoveryRunbook, error)
	UpdateRunbook(ctx context.Context, rb *model.RecoveryRunbook) error
	DeleteRunbook(ctx context.Context, id string) error
	ListRunbooksByPlan(ctx context.Context, planID string) ([]model.RecoveryRunbook, error)
	CreateRunbookStep(ctx context.Context, step *model.RunbookStep) error
	GetRunbookStepByID(ctx context.Context, id string) (*model.RunbookStep, error)
	UpdateRunbookStep(ctx context.Context, step *model.RunbookStep) error
	DeleteRunbookStep(ctx context.Context, id string) error
	ListRunbookSteps(ctx context.Context, runbookID string) ([]model.RunbookStep, error)

	// Continuity Drills & Findings
	CreateDrill(ctx context.Context, drill *model.ContinuityDrill) error
	GetDrillByID(ctx context.Context, id string) (*model.ContinuityDrill, error)
	UpdateDrill(ctx context.Context, drill *model.ContinuityDrill) error
	DeleteDrill(ctx context.Context, id string) error
	ListDrills(ctx context.Context, filter model.DrillFilterCriteria) ([]model.ContinuityDrill, int, error)
	CreateDrillFinding(ctx context.Context, finding *model.DrillFinding) error
	GetDrillFindingByID(ctx context.Context, id string) (*model.DrillFinding, error)
	UpdateDrillFinding(ctx context.Context, finding *model.DrillFinding) error
	DeleteDrillFinding(ctx context.Context, id string) error
	ListDrillFindings(ctx context.Context, filter model.DrillFindingFilterCriteria) ([]model.DrillFinding, int, error)

	// Continuity Readiness & Compliance Analytics
	GetContinuityReadinessReport(ctx context.Context, workspaceID string) (*model.ContinuityReadinessReport, error)
}

// sqliteContinuityRepository implements ContinuityRepository
type sqliteContinuityRepository struct {
	db *sql.DB
}

// NewContinuityRepository creates a new instance of ContinuityRepository
func NewContinuityRepository(db *sql.DB) ContinuityRepository {
	return &sqliteContinuityRepository{db: db}
}

// DR Plans
func (r *sqliteContinuityRepository) CreateDRPlan(ctx context.Context, plan *model.DisasterRecoveryPlan) error {
	if plan.ID == "" {
		plan.ID = uuid.New().String()
	}
	now := time.Now().UTC()
	if plan.CreatedAt.IsZero() {
		plan.CreatedAt = now
	}
	plan.UpdatedAt = now
	if plan.Version <= 0 {
		plan.Version = 1
	}

	query := `
	INSERT INTO dr_plans (
		id, workspace_id, plan_code, name, scope, recovery_tier, status,
		primary_coordinator_id, backup_coordinator_id, invocation_criteria,
		document_url, version, created_at, updated_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := r.db.ExecContext(ctx, query,
		plan.ID, plan.WorkspaceID, plan.PlanCode, plan.Name, plan.Scope,
		plan.RecoveryTier, plan.Status, plan.PrimaryCoordinatorID,
		plan.BackupCoordinatorID, plan.InvocationCriteria, plan.DocumentURL,
		plan.Version, plan.CreatedAt, plan.UpdatedAt,
	)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return ErrDRPlanCodeConflict
		}
		return err
	}
	return nil
}

func (r *sqliteContinuityRepository) GetDRPlanByID(ctx context.Context, id string) (*model.DisasterRecoveryPlan, error) {
	query := `
	SELECT id, workspace_id, plan_code, name, scope, recovery_tier, status,
	       primary_coordinator_id, backup_coordinator_id, invocation_criteria,
	       document_url, version, created_at, updated_at
	FROM dr_plans WHERE id = ?`

	var p model.DisasterRecoveryPlan
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&p.ID, &p.WorkspaceID, &p.PlanCode, &p.Name, &p.Scope, &p.RecoveryTier, &p.Status,
		&p.PrimaryCoordinatorID, &p.BackupCoordinatorID, &p.InvocationCriteria,
		&p.DocumentURL, &p.Version, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrDRPlanNotFound
		}
		return nil, err
	}
	return &p, nil
}

func (r *sqliteContinuityRepository) GetDRPlanByCode(ctx context.Context, workspaceID string, planCode string) (*model.DisasterRecoveryPlan, error) {
	query := `
	SELECT id, workspace_id, plan_code, name, scope, recovery_tier, status,
	       primary_coordinator_id, backup_coordinator_id, invocation_criteria,
	       document_url, version, created_at, updated_at
	FROM dr_plans WHERE workspace_id = ? AND plan_code = ?`

	var p model.DisasterRecoveryPlan
	err := r.db.QueryRowContext(ctx, query, workspaceID, planCode).Scan(
		&p.ID, &p.WorkspaceID, &p.PlanCode, &p.Name, &p.Scope, &p.RecoveryTier, &p.Status,
		&p.PrimaryCoordinatorID, &p.BackupCoordinatorID, &p.InvocationCriteria,
		&p.DocumentURL, &p.Version, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrDRPlanNotFound
		}
		return nil, err
	}
	return &p, nil
}

func (r *sqliteContinuityRepository) UpdateDRPlan(ctx context.Context, plan *model.DisasterRecoveryPlan) error {
	plan.UpdatedAt = time.Now().UTC()
	plan.Version++

	query := `
	UPDATE dr_plans SET
		name = ?, scope = ?, recovery_tier = ?, status = ?,
		primary_coordinator_id = ?, backup_coordinator_id = ?,
		invocation_criteria = ?, document_url = ?, version = ?, updated_at = ?
	WHERE id = ?`

	res, err := r.db.ExecContext(ctx, query,
		plan.Name, plan.Scope, plan.RecoveryTier, plan.Status,
		plan.PrimaryCoordinatorID, plan.BackupCoordinatorID,
		plan.InvocationCriteria, plan.DocumentURL, plan.Version, plan.UpdatedAt,
		plan.ID,
	)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrDRPlanNotFound
	}
	return nil
}

func (r *sqliteContinuityRepository) DeleteDRPlan(ctx context.Context, id string) error {
	res, err := r.db.ExecContext(ctx, "DELETE FROM dr_plans WHERE id = ?", id)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrDRPlanNotFound
	}
	return nil
}

func (r *sqliteContinuityRepository) ListDRPlans(ctx context.Context, filter model.DRPlanFilterCriteria) ([]model.DisasterRecoveryPlan, int, error) {
	whereClauses := []string{"1=1"}
	args := []interface{}{}

	if filter.WorkspaceID != "" {
		whereClauses = append(whereClauses, "workspace_id = ?")
		args = append(args, filter.WorkspaceID)
	}
	if filter.RecoveryTier != "" {
		whereClauses = append(whereClauses, "recovery_tier = ?")
		args = append(args, filter.RecoveryTier)
	}
	if filter.Status != "" {
		whereClauses = append(whereClauses, "status = ?")
		args = append(args, filter.Status)
	}
	if filter.Search != "" {
		whereClauses = append(whereClauses, "(name LIKE ? OR plan_code LIKE ?)")
		s := "%" + filter.Search + "%"
		args = append(args, s, s)
	}

	whereSQL := strings.Join(whereClauses, " AND ")

	var total int
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM dr_plans WHERE %s", whereSQL)
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
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
	SELECT id, workspace_id, plan_code, name, scope, recovery_tier, status,
	       primary_coordinator_id, backup_coordinator_id, invocation_criteria,
	       document_url, version, created_at, updated_at
	FROM dr_plans WHERE %s ORDER BY created_at DESC LIMIT ? OFFSET ?`, whereSQL)

	queryArgs := append(args, limit, offset)
	rows, err := r.db.QueryContext(ctx, query, queryArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	plans := []model.DisasterRecoveryPlan{}
	for rows.Next() {
		var p model.DisasterRecoveryPlan
		if err := rows.Scan(
			&p.ID, &p.WorkspaceID, &p.PlanCode, &p.Name, &p.Scope, &p.RecoveryTier, &p.Status,
			&p.PrimaryCoordinatorID, &p.BackupCoordinatorID, &p.InvocationCriteria,
			&p.DocumentURL, &p.Version, &p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		plans = append(plans, p)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	return plans, total, nil
}

// Recovery Objectives
func (r *sqliteContinuityRepository) CreateRecoveryObjective(ctx context.Context, obj *model.RecoveryObjective) error {
	if obj.ID == "" {
		obj.ID = uuid.New().String()
	}
	now := time.Now().UTC()
	if obj.CreatedAt.IsZero() {
		obj.CreatedAt = now
	}
	obj.UpdatedAt = now

	query := `
	INSERT INTO recovery_objectives (
		id, workspace_id, plan_id, ci_id, service_name,
		rto_value_minutes, rpo_value_minutes, target_availability,
		data_loss_tolerance, notes, created_at, updated_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := r.db.ExecContext(ctx, query,
		obj.ID, obj.WorkspaceID, obj.PlanID, obj.CIID, obj.ServiceName,
		obj.RTOValueMinutes, obj.RPOValueMinutes, obj.TargetAvailability,
		obj.DataLossTolerance, obj.Notes, obj.CreatedAt, obj.UpdatedAt,
	)
	return err
}

func (r *sqliteContinuityRepository) GetRecoveryObjectiveByID(ctx context.Context, id string) (*model.RecoveryObjective, error) {
	query := `
	SELECT id, workspace_id, plan_id, ci_id, service_name,
	       rto_value_minutes, rpo_value_minutes, target_availability,
	       data_loss_tolerance, notes, created_at, updated_at
	FROM recovery_objectives WHERE id = ?`

	var o model.RecoveryObjective
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&o.ID, &o.WorkspaceID, &o.PlanID, &o.CIID, &o.ServiceName,
		&o.RTOValueMinutes, &o.RPOValueMinutes, &o.TargetAvailability,
		&o.DataLossTolerance, &o.Notes, &o.CreatedAt, &o.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrRecoveryObjectiveNotFound
		}
		return nil, err
	}
	return &o, nil
}

func (r *sqliteContinuityRepository) UpdateRecoveryObjective(ctx context.Context, obj *model.RecoveryObjective) error {
	obj.UpdatedAt = time.Now().UTC()

	query := `
	UPDATE recovery_objectives SET
		ci_id = ?, service_name = ?, rto_value_minutes = ?,
		rpo_value_minutes = ?, target_availability = ?,
		data_loss_tolerance = ?, notes = ?, updated_at = ?
	WHERE id = ?`

	res, err := r.db.ExecContext(ctx, query,
		obj.CIID, obj.ServiceName, obj.RTOValueMinutes,
		obj.RPOValueMinutes, obj.TargetAvailability,
		obj.DataLossTolerance, obj.Notes, obj.UpdatedAt,
		obj.ID,
	)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrRecoveryObjectiveNotFound
	}
	return nil
}

func (r *sqliteContinuityRepository) DeleteRecoveryObjective(ctx context.Context, id string) error {
	res, err := r.db.ExecContext(ctx, "DELETE FROM recovery_objectives WHERE id = ?", id)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrRecoveryObjectiveNotFound
	}
	return nil
}

func (r *sqliteContinuityRepository) ListRecoveryObjectivesByPlan(ctx context.Context, planID string) ([]model.RecoveryObjective, error) {
	query := `
	SELECT id, workspace_id, plan_id, ci_id, service_name,
	       rto_value_minutes, rpo_value_minutes, target_availability,
	       data_loss_tolerance, notes, created_at, updated_at
	FROM recovery_objectives WHERE plan_id = ? ORDER BY service_name ASC`

	rows, err := r.db.QueryContext(ctx, query, planID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []model.RecoveryObjective
	for rows.Next() {
		var o model.RecoveryObjective
		if err := rows.Scan(
			&o.ID, &o.WorkspaceID, &o.PlanID, &o.CIID, &o.ServiceName,
			&o.RTOValueMinutes, &o.RPOValueMinutes, &o.TargetAvailability,
			&o.DataLossTolerance, &o.Notes, &o.CreatedAt, &o.UpdatedAt,
		); err != nil {
			return nil, err
		}
		list = append(list, o)
	}
	return list, rows.Err()
}

func (r *sqliteContinuityRepository) ListRecoveryObjectives(ctx context.Context, workspaceID string) ([]model.RecoveryObjective, error) {
	query := `
	SELECT id, workspace_id, plan_id, ci_id, service_name,
	       rto_value_minutes, rpo_value_minutes, target_availability,
	       data_loss_tolerance, notes, created_at, updated_at
	FROM recovery_objectives WHERE workspace_id = ? ORDER BY service_name ASC`

	rows, err := r.db.QueryContext(ctx, query, workspaceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []model.RecoveryObjective
	for rows.Next() {
		var o model.RecoveryObjective
		if err := rows.Scan(
			&o.ID, &o.WorkspaceID, &o.PlanID, &o.CIID, &o.ServiceName,
			&o.RTOValueMinutes, &o.RPOValueMinutes, &o.TargetAvailability,
			&o.DataLossTolerance, &o.Notes, &o.CreatedAt, &o.UpdatedAt,
		); err != nil {
			return nil, err
		}
		list = append(list, o)
	}
	return list, rows.Err()
}

// BIA
func (r *sqliteContinuityRepository) CreateBIA(ctx context.Context, bia *model.BusinessImpactAnalysis) error {
	if bia.ID == "" {
		bia.ID = uuid.New().String()
	}
	now := time.Now().UTC()
	if bia.CreatedAt.IsZero() {
		bia.CreatedAt = now
	}
	bia.UpdatedAt = now
	if bia.AssessmentDate.IsZero() {
		bia.AssessmentDate = now
	}

	query := `
	INSERT INTO business_impact_analyses (
		id, workspace_id, plan_id, ci_id, service_name,
		mtd_minutes, financial_loss_rate_per_hour, operational_impact,
		regulatory_impact, critical_recovery_window, assessment_date,
		assessed_by, created_at, updated_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := r.db.ExecContext(ctx, query,
		bia.ID, bia.WorkspaceID, bia.PlanID, bia.CIID, bia.ServiceName,
		bia.MTDMinutes, bia.FinancialLossRatePerHour, bia.OperationalImpact,
		bia.RegulatoryImpact, bia.CriticalRecoveryWindow, bia.AssessmentDate,
		bia.AssessedBy, bia.CreatedAt, bia.UpdatedAt,
	)
	return err
}

func (r *sqliteContinuityRepository) GetBIAByID(ctx context.Context, id string) (*model.BusinessImpactAnalysis, error) {
	query := `
	SELECT id, workspace_id, plan_id, ci_id, service_name,
	       mtd_minutes, financial_loss_rate_per_hour, operational_impact,
	       regulatory_impact, critical_recovery_window, assessment_date,
	       assessed_by, created_at, updated_at
	FROM business_impact_analyses WHERE id = ?`

	var b model.BusinessImpactAnalysis
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&b.ID, &b.WorkspaceID, &b.PlanID, &b.CIID, &b.ServiceName,
		&b.MTDMinutes, &b.FinancialLossRatePerHour, &b.OperationalImpact,
		&b.RegulatoryImpact, &b.CriticalRecoveryWindow, &b.AssessmentDate,
		&b.AssessedBy, &b.CreatedAt, &b.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrBIANotFound
		}
		return nil, err
	}
	return &b, nil
}

func (r *sqliteContinuityRepository) UpdateBIA(ctx context.Context, bia *model.BusinessImpactAnalysis) error {
	bia.UpdatedAt = time.Now().UTC()

	query := `
	UPDATE business_impact_analyses SET
		ci_id = ?, service_name = ?, mtd_minutes = ?,
		financial_loss_rate_per_hour = ?, operational_impact = ?,
		regulatory_impact = ?, critical_recovery_window = ?,
		assessed_by = ?, updated_at = ?
	WHERE id = ?`

	res, err := r.db.ExecContext(ctx, query,
		bia.CIID, bia.ServiceName, bia.MTDMinutes,
		bia.FinancialLossRatePerHour, bia.OperationalImpact,
		bia.RegulatoryImpact, bia.CriticalRecoveryWindow,
		bia.AssessedBy, bia.UpdatedAt,
		bia.ID,
	)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrBIANotFound
	}
	return nil
}

func (r *sqliteContinuityRepository) DeleteBIA(ctx context.Context, id string) error {
	res, err := r.db.ExecContext(ctx, "DELETE FROM business_impact_analyses WHERE id = ?", id)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrBIANotFound
	}
	return nil
}

func (r *sqliteContinuityRepository) ListBIAsByPlan(ctx context.Context, planID string) ([]model.BusinessImpactAnalysis, error) {
	query := `
	SELECT id, workspace_id, plan_id, ci_id, service_name,
	       mtd_minutes, financial_loss_rate_per_hour, operational_impact,
	       regulatory_impact, critical_recovery_window, assessment_date,
	       assessed_by, created_at, updated_at
	FROM business_impact_analyses WHERE plan_id = ? ORDER BY service_name ASC`

	rows, err := r.db.QueryContext(ctx, query, planID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []model.BusinessImpactAnalysis
	for rows.Next() {
		var b model.BusinessImpactAnalysis
		if err := rows.Scan(
			&b.ID, &b.WorkspaceID, &b.PlanID, &b.CIID, &b.ServiceName,
			&b.MTDMinutes, &b.FinancialLossRatePerHour, &b.OperationalImpact,
			&b.RegulatoryImpact, &b.CriticalRecoveryWindow, &b.AssessmentDate,
			&b.AssessedBy, &b.CreatedAt, &b.UpdatedAt,
		); err != nil {
			return nil, err
		}
		list = append(list, b)
	}
	return list, rows.Err()
}

func (r *sqliteContinuityRepository) ListBIAs(ctx context.Context, workspaceID string) ([]model.BusinessImpactAnalysis, error) {
	query := `
	SELECT id, workspace_id, plan_id, ci_id, service_name,
	       mtd_minutes, financial_loss_rate_per_hour, operational_impact,
	       regulatory_impact, critical_recovery_window, assessment_date,
	       assessed_by, created_at, updated_at
	FROM business_impact_analyses WHERE workspace_id = ? ORDER BY service_name ASC`

	rows, err := r.db.QueryContext(ctx, query, workspaceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []model.BusinessImpactAnalysis
	for rows.Next() {
		var b model.BusinessImpactAnalysis
		if err := rows.Scan(
			&b.ID, &b.WorkspaceID, &b.PlanID, &b.CIID, &b.ServiceName,
			&b.MTDMinutes, &b.FinancialLossRatePerHour, &b.OperationalImpact,
			&b.RegulatoryImpact, &b.CriticalRecoveryWindow, &b.AssessmentDate,
			&b.AssessedBy, &b.CreatedAt, &b.UpdatedAt,
		); err != nil {
			return nil, err
		}
		list = append(list, b)
	}
	return list, rows.Err()
}

// Runbooks
func (r *sqliteContinuityRepository) CreateRunbook(ctx context.Context, rb *model.RecoveryRunbook) error {
	return nil
}

func (r *sqliteContinuityRepository) GetRunbookByID(ctx context.Context, id string) (*model.RecoveryRunbook, error) {
	return nil, ErrRunbookNotFound
}

func (r *sqliteContinuityRepository) UpdateRunbook(ctx context.Context, rb *model.RecoveryRunbook) error {
	return nil
}

func (r *sqliteContinuityRepository) DeleteRunbook(ctx context.Context, id string) error {
	return nil
}

func (r *sqliteContinuityRepository) ListRunbooksByPlan(ctx context.Context, planID string) ([]model.RecoveryRunbook, error) {
	return nil, nil
}

func (r *sqliteContinuityRepository) CreateRunbookStep(ctx context.Context, step *model.RunbookStep) error {
	return nil
}

func (r *sqliteContinuityRepository) GetRunbookStepByID(ctx context.Context, id string) (*model.RunbookStep, error) {
	return nil, ErrRunbookStepNotFound
}

func (r *sqliteContinuityRepository) UpdateRunbookStep(ctx context.Context, step *model.RunbookStep) error {
	return nil
}

func (r *sqliteContinuityRepository) DeleteRunbookStep(ctx context.Context, id string) error {
	return nil
}

func (r *sqliteContinuityRepository) ListRunbookSteps(ctx context.Context, runbookID string) ([]model.RunbookStep, error) {
	return nil, nil
}

// Drills
func (r *sqliteContinuityRepository) CreateDrill(ctx context.Context, drill *model.ContinuityDrill) error {
	return nil
}

func (r *sqliteContinuityRepository) GetDrillByID(ctx context.Context, id string) (*model.ContinuityDrill, error) {
	return nil, ErrDrillNotFound
}

func (r *sqliteContinuityRepository) UpdateDrill(ctx context.Context, drill *model.ContinuityDrill) error {
	return nil
}

func (r *sqliteContinuityRepository) DeleteDrill(ctx context.Context, id string) error {
	return nil
}

func (r *sqliteContinuityRepository) ListDrills(ctx context.Context, filter model.DrillFilterCriteria) ([]model.ContinuityDrill, int, error) {
	return nil, 0, nil
}

func (r *sqliteContinuityRepository) CreateDrillFinding(ctx context.Context, finding *model.DrillFinding) error {
	return nil
}

func (r *sqliteContinuityRepository) GetDrillFindingByID(ctx context.Context, id string) (*model.DrillFinding, error) {
	return nil, ErrDrillFindingNotFound
}

func (r *sqliteContinuityRepository) UpdateDrillFinding(ctx context.Context, finding *model.DrillFinding) error {
	return nil
}

func (r *sqliteContinuityRepository) DeleteDrillFinding(ctx context.Context, id string) error {
	return nil
}

func (r *sqliteContinuityRepository) ListDrillFindings(ctx context.Context, filter model.DrillFindingFilterCriteria) ([]model.DrillFinding, int, error) {
	return nil, 0, nil
}

// Analytics
func (r *sqliteContinuityRepository) GetContinuityReadinessReport(ctx context.Context, workspaceID string) (*model.ContinuityReadinessReport, error) {
	return &model.ContinuityReadinessReport{WorkspaceID: workspaceID}, nil
}
