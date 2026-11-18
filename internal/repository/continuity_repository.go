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
	if rb.ID == "" {
		rb.ID = uuid.New().String()
	}
	now := time.Now().UTC()
	if rb.CreatedAt.IsZero() {
		rb.CreatedAt = now
	}
	rb.UpdatedAt = now

	query := `
	INSERT INTO dr_runbooks (
		id, workspace_id, plan_id, runbook_code, title,
		description, execution_mode, estimated_duration_minutes,
		created_at, updated_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := r.db.ExecContext(ctx, query,
		rb.ID, rb.WorkspaceID, rb.PlanID, rb.RunbookCode, rb.Title,
		rb.Description, rb.ExecutionMode, rb.EstimatedDurationMinutes,
		rb.CreatedAt, rb.UpdatedAt,
	)
	return err
}

func (r *sqliteContinuityRepository) GetRunbookByID(ctx context.Context, id string) (*model.RecoveryRunbook, error) {
	query := `
	SELECT id, workspace_id, plan_id, runbook_code, title,
	       description, execution_mode, estimated_duration_minutes,
	       created_at, updated_at
	FROM dr_runbooks WHERE id = ?`

	var rb model.RecoveryRunbook
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&rb.ID, &rb.WorkspaceID, &rb.PlanID, &rb.RunbookCode, &rb.Title,
		&rb.Description, &rb.ExecutionMode, &rb.EstimatedDurationMinutes,
		&rb.CreatedAt, &rb.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrRunbookNotFound
		}
		return nil, err
	}

	steps, err := r.ListRunbookSteps(ctx, rb.ID)
	if err != nil {
		return nil, err
	}
	rb.Steps = steps
	return &rb, nil
}

func (r *sqliteContinuityRepository) UpdateRunbook(ctx context.Context, rb *model.RecoveryRunbook) error {
	rb.UpdatedAt = time.Now().UTC()

	query := `
	UPDATE dr_runbooks SET
		title = ?, description = ?, execution_mode = ?,
		estimated_duration_minutes = ?, updated_at = ?
	WHERE id = ?`

	res, err := r.db.ExecContext(ctx, query,
		rb.Title, rb.Description, rb.ExecutionMode,
		rb.EstimatedDurationMinutes, rb.UpdatedAt,
		rb.ID,
	)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrRunbookNotFound
	}
	return nil
}

func (r *sqliteContinuityRepository) DeleteRunbook(ctx context.Context, id string) error {
	res, err := r.db.ExecContext(ctx, "DELETE FROM dr_runbooks WHERE id = ?", id)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrRunbookNotFound
	}
	return nil
}

func (r *sqliteContinuityRepository) ListRunbooksByPlan(ctx context.Context, planID string) ([]model.RecoveryRunbook, error) {
	query := `
	SELECT id, workspace_id, plan_id, runbook_code, title,
	       description, execution_mode, estimated_duration_minutes,
	       created_at, updated_at
	FROM dr_runbooks WHERE plan_id = ? ORDER BY runbook_code ASC`

	rows, err := r.db.QueryContext(ctx, query, planID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []model.RecoveryRunbook
	for rows.Next() {
		var rb model.RecoveryRunbook
		if err := rows.Scan(
			&rb.ID, &rb.WorkspaceID, &rb.PlanID, &rb.RunbookCode, &rb.Title,
			&rb.Description, &rb.ExecutionMode, &rb.EstimatedDurationMinutes,
			&rb.CreatedAt, &rb.UpdatedAt,
		); err != nil {
			return nil, err
		}
		list = append(list, rb)
	}
	return list, rows.Err()
}

func (r *sqliteContinuityRepository) CreateRunbookStep(ctx context.Context, step *model.RunbookStep) error {
	if step.ID == "" {
		step.ID = uuid.New().String()
	}
	now := time.Now().UTC()
	if step.CreatedAt.IsZero() {
		step.CreatedAt = now
	}
	step.UpdatedAt = now

	isCrit := 0
	if step.IsCritical {
		isCrit = 1
	}

	query := `
	INSERT INTO dr_runbook_steps (
		id, runbook_id, step_order, name, action_type,
		execution_role, command_or_script, estimated_duration_minutes,
		is_critical, fallback_procedure, created_at, updated_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := r.db.ExecContext(ctx, query,
		step.ID, step.RunbookID, step.StepOrder, step.Name, step.ActionType,
		step.ExecutionRole, step.CommandOrScript, step.EstimatedDurationMinutes,
		isCrit, step.FallbackProcedure, step.CreatedAt, step.UpdatedAt,
	)
	return err
}

func (r *sqliteContinuityRepository) GetRunbookStepByID(ctx context.Context, id string) (*model.RunbookStep, error) {
	query := `
	SELECT id, runbook_id, step_order, name, action_type,
	       execution_role, command_or_script, estimated_duration_minutes,
	       is_critical, fallback_procedure, created_at, updated_at
	FROM dr_runbook_steps WHERE id = ?`

	var s model.RunbookStep
	var isCrit int
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&s.ID, &s.RunbookID, &s.StepOrder, &s.Name, &s.ActionType,
		&s.ExecutionRole, &s.CommandOrScript, &s.EstimatedDurationMinutes,
		&isCrit, &s.FallbackProcedure, &s.CreatedAt, &s.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrRunbookStepNotFound
		}
		return nil, err
	}
	s.IsCritical = (isCrit == 1)
	return &s, nil
}

func (r *sqliteContinuityRepository) UpdateRunbookStep(ctx context.Context, step *model.RunbookStep) error {
	step.UpdatedAt = time.Now().UTC()
	isCrit := 0
	if step.IsCritical {
		isCrit = 1
	}

	query := `
	UPDATE dr_runbook_steps SET
		step_order = ?, name = ?, action_type = ?,
		execution_role = ?, command_or_script = ?,
		estimated_duration_minutes = ?, is_critical = ?,
		fallback_procedure = ?, updated_at = ?
	WHERE id = ?`

	res, err := r.db.ExecContext(ctx, query,
		step.StepOrder, step.Name, step.ActionType,
		step.ExecutionRole, step.CommandOrScript,
		step.EstimatedDurationMinutes, isCrit,
		step.FallbackProcedure, step.UpdatedAt,
		step.ID,
	)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrRunbookStepNotFound
	}
	return nil
}

func (r *sqliteContinuityRepository) DeleteRunbookStep(ctx context.Context, id string) error {
	res, err := r.db.ExecContext(ctx, "DELETE FROM dr_runbook_steps WHERE id = ?", id)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrRunbookStepNotFound
	}
	return nil
}

func (r *sqliteContinuityRepository) ListRunbookSteps(ctx context.Context, runbookID string) ([]model.RunbookStep, error) {
	query := `
	SELECT id, runbook_id, step_order, name, action_type,
	       execution_role, command_or_script, estimated_duration_minutes,
	       is_critical, fallback_procedure, created_at, updated_at
	FROM dr_runbook_steps WHERE runbook_id = ? ORDER BY step_order ASC`

	rows, err := r.db.QueryContext(ctx, query, runbookID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var steps []model.RunbookStep
	for rows.Next() {
		var s model.RunbookStep
		var isCrit int
		if err := rows.Scan(
			&s.ID, &s.RunbookID, &s.StepOrder, &s.Name, &s.ActionType,
			&s.ExecutionRole, &s.CommandOrScript, &s.EstimatedDurationMinutes,
			&isCrit, &s.FallbackProcedure, &s.CreatedAt, &s.UpdatedAt,
		); err != nil {
			return nil, err
		}
		s.IsCritical = (isCrit == 1)
		steps = append(steps, s)
	}
	return steps, rows.Err()
}

// Drills
func (r *sqliteContinuityRepository) CreateDrill(ctx context.Context, drill *model.ContinuityDrill) error {
	if drill.ID == "" {
		drill.ID = uuid.New().String()
	}
	now := time.Now().UTC()
	if drill.CreatedAt.IsZero() {
		drill.CreatedAt = now
	}
	drill.UpdatedAt = now
	if drill.Status == "" {
		drill.Status = model.DrillStatusScheduled
	}

	successInt := 0
	if drill.Success {
		successInt = 1
	}

	query := `
	INSERT INTO continuity_drills (
		id, workspace_id, plan_id, drill_code, drill_type,
		scenario, scheduled_date, status, actual_recovery_duration_minutes,
		success, conducted_by, post_mortem_summary, created_at, updated_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := r.db.ExecContext(ctx, query,
		drill.ID, drill.WorkspaceID, drill.PlanID, drill.DrillCode, drill.DrillType,
		drill.Scenario, drill.ScheduledDate, drill.Status, drill.ActualRecoveryDurationMinutes,
		successInt, drill.ConductedBy, drill.PostMortemSummary, drill.CreatedAt, drill.UpdatedAt,
	)
	return err
}

func (r *sqliteContinuityRepository) GetDrillByID(ctx context.Context, id string) (*model.ContinuityDrill, error) {
	query := `
	SELECT id, workspace_id, plan_id, drill_code, drill_type,
	       scenario, scheduled_date, status, actual_recovery_duration_minutes,
	       success, conducted_by, post_mortem_summary, created_at, updated_at
	FROM continuity_drills WHERE id = ?`

	var d model.ContinuityDrill
	var successInt int
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&d.ID, &d.WorkspaceID, &d.PlanID, &d.DrillCode, &d.DrillType,
		&d.Scenario, &d.ScheduledDate, &d.Status, &d.ActualRecoveryDurationMinutes,
		&successInt, &d.ConductedBy, &d.PostMortemSummary, &d.CreatedAt, &d.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrDrillNotFound
		}
		return nil, err
	}
	d.Success = (successInt == 1)

	findings, _, err := r.ListDrillFindings(ctx, model.DrillFindingFilterCriteria{
		DrillID: d.ID,
		Limit:   100,
	})
	if err != nil {
		return nil, err
	}
	d.Findings = findings
	return &d, nil
}

func (r *sqliteContinuityRepository) UpdateDrill(ctx context.Context, drill *model.ContinuityDrill) error {
	drill.UpdatedAt = time.Now().UTC()
	successInt := 0
	if drill.Success {
		successInt = 1
	}

	query := `
	UPDATE continuity_drills SET
		scenario = ?, scheduled_date = ?, status = ?,
		actual_recovery_duration_minutes = ?, success = ?,
		conducted_by = ?, post_mortem_summary = ?, updated_at = ?
	WHERE id = ?`

	res, err := r.db.ExecContext(ctx, query,
		drill.Scenario, drill.ScheduledDate, drill.Status,
		drill.ActualRecoveryDurationMinutes, successInt,
		drill.ConductedBy, drill.PostMortemSummary, drill.UpdatedAt,
		drill.ID,
	)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrDrillNotFound
	}
	return nil
}

func (r *sqliteContinuityRepository) DeleteDrill(ctx context.Context, id string) error {
	res, err := r.db.ExecContext(ctx, "DELETE FROM continuity_drills WHERE id = ?", id)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrDrillNotFound
	}
	return nil
}

func (r *sqliteContinuityRepository) ListDrills(ctx context.Context, filter model.DrillFilterCriteria) ([]model.ContinuityDrill, int, error) {
	whereClauses := []string{"1=1"}
	args := []interface{}{}

	if filter.WorkspaceID != "" {
		whereClauses = append(whereClauses, "workspace_id = ?")
		args = append(args, filter.WorkspaceID)
	}
	if filter.PlanID != "" {
		whereClauses = append(whereClauses, "plan_id = ?")
		args = append(args, filter.PlanID)
	}
	if filter.DrillType != "" {
		whereClauses = append(whereClauses, "drill_type = ?")
		args = append(args, filter.DrillType)
	}
	if filter.Status != "" {
		whereClauses = append(whereClauses, "status = ?")
		args = append(args, filter.Status)
	}

	whereSQL := strings.Join(whereClauses, " AND ")

	var total int
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM continuity_drills WHERE %s", whereSQL)
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
	SELECT id, workspace_id, plan_id, drill_code, drill_type,
	       scenario, scheduled_date, status, actual_recovery_duration_minutes,
	       success, conducted_by, post_mortem_summary, created_at, updated_at
	FROM continuity_drills WHERE %s ORDER BY scheduled_date DESC LIMIT ? OFFSET ?`, whereSQL)

	queryArgs := append(args, limit, offset)
	rows, err := r.db.QueryContext(ctx, query, queryArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	drills := []model.ContinuityDrill{}
	for rows.Next() {
		var d model.ContinuityDrill
		var successInt int
		if err := rows.Scan(
			&d.ID, &d.WorkspaceID, &d.PlanID, &d.DrillCode, &d.DrillType,
			&d.Scenario, &d.ScheduledDate, &d.Status, &d.ActualRecoveryDurationMinutes,
			&successInt, &d.ConductedBy, &d.PostMortemSummary, &d.CreatedAt, &d.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		d.Success = (successInt == 1)
		drills = append(drills, d)
	}
	return drills, total, rows.Err()
}

func (r *sqliteContinuityRepository) CreateDrillFinding(ctx context.Context, finding *model.DrillFinding) error {
	if finding.ID == "" {
		finding.ID = uuid.New().String()
	}
	now := time.Now().UTC()
	if finding.CreatedAt.IsZero() {
		finding.CreatedAt = now
	}
	finding.UpdatedAt = now
	if finding.Status == "" {
		finding.Status = model.DrillFindingStatusOpen
	}

	query := `
	INSERT INTO continuity_drill_findings (
		id, workspace_id, drill_id, severity, category,
		description, remediation_action, assigned_to, status,
		resolved_at, created_at, updated_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := r.db.ExecContext(ctx, query,
		finding.ID, finding.WorkspaceID, finding.DrillID, finding.Severity, finding.Category,
		finding.Description, finding.RemediationAction, finding.AssignedTo, finding.Status,
		finding.ResolvedAt, finding.CreatedAt, finding.UpdatedAt,
	)
	return err
}

func (r *sqliteContinuityRepository) GetDrillFindingByID(ctx context.Context, id string) (*model.DrillFinding, error) {
	query := `
	SELECT id, workspace_id, drill_id, severity, category,
	       description, remediation_action, assigned_to, status,
	       resolved_at, created_at, updated_at
	FROM continuity_drill_findings WHERE id = ?`

	var f model.DrillFinding
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&f.ID, &f.WorkspaceID, &f.DrillID, &f.Severity, &f.Category,
		&f.Description, &f.RemediationAction, &f.AssignedTo, &f.Status,
		&f.ResolvedAt, &f.CreatedAt, &f.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrDrillFindingNotFound
		}
		return nil, err
	}
	return &f, nil
}

func (r *sqliteContinuityRepository) UpdateDrillFinding(ctx context.Context, finding *model.DrillFinding) error {
	finding.UpdatedAt = time.Now().UTC()
	if finding.Status == model.DrillFindingStatusResolved && finding.ResolvedAt == nil {
		now := time.Now().UTC()
		finding.ResolvedAt = &now
	}

	query := `
	UPDATE continuity_drill_findings SET
		severity = ?, category = ?, description = ?,
		remediation_action = ?, assigned_to = ?, status = ?,
		resolved_at = ?, updated_at = ?
	WHERE id = ?`

	res, err := r.db.ExecContext(ctx, query,
		finding.Severity, finding.Category, finding.Description,
		finding.RemediationAction, finding.AssignedTo, finding.Status,
		finding.ResolvedAt, finding.UpdatedAt,
		finding.ID,
	)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrDrillFindingNotFound
	}
	return nil
}

func (r *sqliteContinuityRepository) DeleteDrillFinding(ctx context.Context, id string) error {
	res, err := r.db.ExecContext(ctx, "DELETE FROM continuity_drill_findings WHERE id = ?", id)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrDrillFindingNotFound
	}
	return nil
}

func (r *sqliteContinuityRepository) ListDrillFindings(ctx context.Context, filter model.DrillFindingFilterCriteria) ([]model.DrillFinding, int, error) {
	whereClauses := []string{"1=1"}
	args := []interface{}{}

	if filter.WorkspaceID != "" {
		whereClauses = append(whereClauses, "workspace_id = ?")
		args = append(args, filter.WorkspaceID)
	}
	if filter.DrillID != "" {
		whereClauses = append(whereClauses, "drill_id = ?")
		args = append(args, filter.DrillID)
	}
	if filter.Severity != "" {
		whereClauses = append(whereClauses, "severity = ?")
		args = append(args, filter.Severity)
	}
	if filter.Category != "" {
		whereClauses = append(whereClauses, "category = ?")
		args = append(args, filter.Category)
	}
	if filter.Status != "" {
		whereClauses = append(whereClauses, "status = ?")
		args = append(args, filter.Status)
	}

	whereSQL := strings.Join(whereClauses, " AND ")

	var total int
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM continuity_drill_findings WHERE %s", whereSQL)
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
	SELECT id, workspace_id, drill_id, severity, category,
	       description, remediation_action, assigned_to, status,
	       resolved_at, created_at, updated_at
	FROM continuity_drill_findings WHERE %s ORDER BY created_at DESC LIMIT ? OFFSET ?`, whereSQL)

	queryArgs := append(args, limit, offset)
	rows, err := r.db.QueryContext(ctx, query, queryArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	findings := []model.DrillFinding{}
	for rows.Next() {
		var f model.DrillFinding
		if err := rows.Scan(
			&f.ID, &f.WorkspaceID, &f.DrillID, &f.Severity, &f.Category,
			&f.Description, &f.RemediationAction, &f.AssignedTo, &f.Status,
			&f.ResolvedAt, &f.CreatedAt, &f.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		findings = append(findings, f)
	}
	return findings, total, rows.Err()
}

// Analytics
func (r *sqliteContinuityRepository) GetContinuityReadinessReport(ctx context.Context, workspaceID string) (*model.ContinuityReadinessReport, error) {
	report := &model.ContinuityReadinessReport{
		WorkspaceID:        workspaceID,
		PlansByTier:        make(map[string]int),
		FindingsBySeverity: make(map[string]int),
	}

	// Plans stats
	planRows, err := r.db.QueryContext(ctx, `
		SELECT recovery_tier, status, COUNT(*)
		FROM dr_plans WHERE workspace_id = ?
		GROUP BY recovery_tier, status`, workspaceID)
	if err != nil {
		return nil, err
	}
	defer planRows.Close()

	for planRows.Next() {
		var tier, status string
		var count int
		if err := planRows.Scan(&tier, &status, &count); err != nil {
			return nil, err
		}
		report.TotalPlans += count
		report.PlansByTier[tier] += count
		if status == model.DRPlanStatusActive {
			report.ActivePlans += count
		}
		if status == model.DRPlanStatusApproved {
			report.ApprovedPlans += count
		}
	}

	// Objectives stats
	var avgRTO, avgRPO sql.NullFloat64
	var totalObjectives int
	err = r.db.QueryRowContext(ctx, `
		SELECT COUNT(*), AVG(rto_value_minutes), AVG(rpo_value_minutes)
		FROM recovery_objectives WHERE workspace_id = ?`, workspaceID).Scan(
		&totalObjectives, &avgRTO, &avgRPO,
	)
	if err == nil {
		report.TotalObjectives = totalObjectives
		if avgRTO.Valid {
			report.AverageRTOMinutes = avgRTO.Float64
		}
		if avgRPO.Valid {
			report.AverageRPOMinutes = avgRPO.Float64
		}
	}

	// Drills stats
	var totalDrills, completedDrills, successfulDrills int
	var avgRecoveryDuration sql.NullFloat64
	err = r.db.QueryRowContext(ctx, `
		SELECT COUNT(*),
		       COALESCE(SUM(CASE WHEN status = 'completed' THEN 1 ELSE 0 END), 0),
		       COALESCE(SUM(CASE WHEN status = 'completed' AND success = 1 THEN 1 ELSE 0 END), 0),
		       AVG(CASE WHEN status = 'completed' THEN actual_recovery_duration_minutes ELSE NULL END)
		FROM continuity_drills WHERE workspace_id = ?`, workspaceID).Scan(
		&totalDrills, &completedDrills, &successfulDrills, &avgRecoveryDuration,
	)
	if err == nil {
		report.TotalDrills = totalDrills
		report.CompletedDrills = completedDrills
		if completedDrills > 0 {
			report.DrillPassRate = (float64(successfulDrills) / float64(completedDrills)) * 100.0
			report.RTOComplianceRate = report.DrillPassRate
		}
		if avgRecoveryDuration.Valid {
			report.AverageRecoveryDuration = avgRecoveryDuration.Float64
		}
	}

	// Findings stats
	findingRows, err := r.db.QueryContext(ctx, `
		SELECT severity, status, COUNT(*)
		FROM continuity_drill_findings WHERE workspace_id = ?
		GROUP BY severity, status`, workspaceID)
	if err != nil {
		return nil, err
	}
	defer findingRows.Close()

	var openCritical, openHigh, openOther int
	for findingRows.Next() {
		var sev, st string
		var cnt int
		if err := findingRows.Scan(&sev, &st, &cnt); err != nil {
			return nil, err
		}
		report.TotalFindings += cnt
		report.FindingsBySeverity[sev] += cnt
		if st != model.DrillFindingStatusResolved && st != model.DrillFindingStatusWaived {
			report.OpenFindingsCount += cnt
			switch sev {
			case model.DrillFindingSeverityCritical:
				openCritical += cnt
			case model.DrillFindingSeverityHigh:
				openHigh += cnt
			default:
				openOther += cnt
			}
		}
	}

	// Calculate composite Readiness Score (0 - 100)
	score := 50.0 // Baseline
	if report.TotalPlans > 0 {
		activeRatio := float64(report.ActivePlans+report.ApprovedPlans) / float64(report.TotalPlans)
		score += activeRatio * 20.0
	}
	if report.CompletedDrills > 0 {
		score += (report.DrillPassRate / 100.0) * 20.0
	} else if report.TotalPlans > 0 {
		score -= 10.0 // Penalty for no drills
	}
	// Deduct for open findings
	score -= float64(openCritical * 15)
	score -= float64(openHigh * 8)
	score -= float64(openOther * 2)

	if score < 0 {
		score = 0
	}
	if score > 100 {
		score = 100
	}
	report.ReadinessScore = score

	return report, nil
}
