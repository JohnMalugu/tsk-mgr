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
	ErrCSIInitiativeNotFound          = model.ErrCSIInitiativeNotFound
	ErrCSIInitiativeTitleRequired     = model.ErrCSIInitiativeTitleRequired
	ErrCSIInitiativeCodeConflict      = errors.New("initiative with this code already exists in workspace")
	ErrCSIBenchmarkNotFound           = model.ErrCSIBenchmarkNotFound
	ErrCSIBenchmarkMetricNameRequired = model.ErrCSIBenchmarkMetricNameRequired
	ErrCSIBenefitNotFound             = model.ErrCSIBenefitNotFound
	ErrCSIActionNotFound              = model.ErrCSIActionNotFound
	ErrCSIReviewNotFound              = model.ErrCSIReviewNotFound
)

// CSIRepository defines data access methods for ITIL Continual Service Improvement
type CSIRepository interface {
	// Initiatives (CIR)
	CreateInitiative(ctx context.Context, initiative *model.CSIInitiative) error
	GetInitiativeByID(ctx context.Context, id string) (*model.CSIInitiative, error)
	GetInitiativeByCode(ctx context.Context, workspaceID string, code string) (*model.CSIInitiative, error)
	UpdateInitiative(ctx context.Context, initiative *model.CSIInitiative) error
	DeleteInitiative(ctx context.Context, id string) error
	ListInitiatives(ctx context.Context, filter model.CSIFilterCriteria) ([]model.CSIInitiative, int, error)

	// Benchmarks
	CreateBenchmark(ctx context.Context, benchmark *model.CSIBenchmark) error
	GetBenchmarkByID(ctx context.Context, id string) (*model.CSIBenchmark, error)
	UpdateBenchmark(ctx context.Context, benchmark *model.CSIBenchmark) error
	DeleteBenchmark(ctx context.Context, id string) error
	ListBenchmarks(ctx context.Context, initiativeID string) ([]model.CSIBenchmark, error)

	// Benefits / ROI
	SaveBenefit(ctx context.Context, benefit *model.CSIBenefit) error
	GetBenefitByInitiativeID(ctx context.Context, initiativeID string) (*model.CSIBenefit, error)
	UpdateBenefit(ctx context.Context, benefit *model.CSIBenefit) error

	// Actions
	CreateAction(ctx context.Context, action *model.CSIAction) error
	GetActionByID(ctx context.Context, id string) (*model.CSIAction, error)
	UpdateAction(ctx context.Context, action *model.CSIAction) error
	DeleteAction(ctx context.Context, id string) error
	ListActions(ctx context.Context, initiativeID string) ([]model.CSIAction, error)

	// Reviews
	CreateReview(ctx context.Context, review *model.CSIReview) error
	GetReviewByID(ctx context.Context, id string) (*model.CSIReview, error)
	ListReviews(ctx context.Context, initiativeID string) ([]model.CSIReview, error)

	// Posture Analytics
	GetCSIAnalytics(ctx context.Context, workspaceID string) (*model.CSIAnalytics, error)
}

type csiRepository struct {
	db *sql.DB
}

// NewCSIRepository creates a new SQLite repository for ITIL Continual Service Improvement
func NewCSIRepository(db *sql.DB) CSIRepository {
	return &csiRepository{db: db}
}

// CreateInitiative registers a new continual improvement initiative
func (r *csiRepository) CreateInitiative(ctx context.Context, initiative *model.CSIInitiative) error {
	if err := initiative.Validate(); err != nil {
		return err
	}
	if initiative.ID == "" {
		initiative.ID = uuid.New().String()
	}
	now := time.Now().UTC()
	if initiative.CreatedAt.IsZero() {
		initiative.CreatedAt = now
	}
	initiative.UpdatedAt = now

	if initiative.InitiativeCode == "" {
		var count int
		_ = r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM csi_initiatives WHERE workspace_id = ?", initiative.WorkspaceID).Scan(&count)
		initiative.InitiativeCode = fmt.Sprintf("CSI-%04d", count+1)
	}

	// Check code conflict
	existing, err := r.GetInitiativeByCode(ctx, initiative.WorkspaceID, initiative.InitiativeCode)
	if err == nil && existing != nil {
		return ErrCSIInitiativeCodeConflict
	}

	query := `
	INSERT INTO csi_initiatives (
		id, workspace_id, initiative_code, title, description,
		strategic_alignment, priority, status, source,
		source_reference_id, ci_id, owner_id,
		target_completion_date, completed_at, created_at, updated_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err = r.db.ExecContext(ctx, query,
		initiative.ID, initiative.WorkspaceID, initiative.InitiativeCode, initiative.Title, initiative.Description,
		initiative.StrategicAlignment, initiative.Priority, initiative.Status, initiative.Source,
		initiative.SourceReferenceID, initiative.CIID, initiative.OwnerID,
		initiative.TargetCompletionDate, initiative.CompletedAt, initiative.CreatedAt, initiative.UpdatedAt,
	)
	return err
}

// GetInitiativeByID retrieves an initiative by its primary ID
func (r *csiRepository) GetInitiativeByID(ctx context.Context, id string) (*model.CSIInitiative, error) {
	query := `
	SELECT id, workspace_id, initiative_code, title, description,
	       strategic_alignment, priority, status, source,
	       source_reference_id, ci_id, owner_id,
	       target_completion_date, completed_at, created_at, updated_at
	FROM csi_initiatives WHERE id = ?`

	var i model.CSIInitiative
	var desc, sourceRef, ciID, ownerID sql.NullString
	var completedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&i.ID, &i.WorkspaceID, &i.InitiativeCode, &i.Title, &desc,
		&i.StrategicAlignment, &i.Priority, &i.Status, &i.Source,
		&sourceRef, &ciID, &ownerID,
		&i.TargetCompletionDate, &completedAt, &i.CreatedAt, &i.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrCSIInitiativeNotFound
		}
		return nil, err
	}
	if desc.Valid {
		i.Description = desc.String
	}
	if sourceRef.Valid {
		i.SourceReferenceID = sourceRef.String
	}
	if ciID.Valid {
		i.CIID = ciID.String
	}
	if ownerID.Valid {
		i.OwnerID = ownerID.String
	}
	if completedAt.Valid {
		t := completedAt.Time
		i.CompletedAt = &t
	}
	return &i, nil
}

// GetInitiativeByCode retrieves an initiative by code within a workspace
func (r *csiRepository) GetInitiativeByCode(ctx context.Context, workspaceID string, code string) (*model.CSIInitiative, error) {
	query := `
	SELECT id, workspace_id, initiative_code, title, description,
	       strategic_alignment, priority, status, source,
	       source_reference_id, ci_id, owner_id,
	       target_completion_date, completed_at, created_at, updated_at
	FROM csi_initiatives WHERE workspace_id = ? AND initiative_code = ?`

	var i model.CSIInitiative
	var desc, sourceRef, ciID, ownerID sql.NullString
	var completedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, workspaceID, code).Scan(
		&i.ID, &i.WorkspaceID, &i.InitiativeCode, &i.Title, &desc,
		&i.StrategicAlignment, &i.Priority, &i.Status, &i.Source,
		&sourceRef, &ciID, &ownerID,
		&i.TargetCompletionDate, &completedAt, &i.CreatedAt, &i.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrCSIInitiativeNotFound
		}
		return nil, err
	}
	if desc.Valid {
		i.Description = desc.String
	}
	if sourceRef.Valid {
		i.SourceReferenceID = sourceRef.String
	}
	if ciID.Valid {
		i.CIID = ciID.String
	}
	if ownerID.Valid {
		i.OwnerID = ownerID.String
	}
	if completedAt.Valid {
		t := completedAt.Time
		i.CompletedAt = &t
	}
	return &i, nil
}

// UpdateInitiative modifies details of an existing initiative
func (r *csiRepository) UpdateInitiative(ctx context.Context, initiative *model.CSIInitiative) error {
	if err := initiative.Validate(); err != nil {
		return err
	}
	initiative.UpdatedAt = time.Now().UTC()

	query := `
	UPDATE csi_initiatives SET
		title = ?, description = ?, strategic_alignment = ?,
		priority = ?, status = ?, source = ?, source_reference_id = ?,
		ci_id = ?, owner_id = ?, target_completion_date = ?,
		completed_at = ?, updated_at = ?
	WHERE id = ?`

	res, err := r.db.ExecContext(ctx, query,
		initiative.Title, initiative.Description, initiative.StrategicAlignment,
		initiative.Priority, initiative.Status, initiative.Source, initiative.SourceReferenceID,
		initiative.CIID, initiative.OwnerID, initiative.TargetCompletionDate,
		initiative.CompletedAt, initiative.UpdatedAt,
		initiative.ID,
	)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrCSIInitiativeNotFound
	}
	return nil
}

// DeleteInitiative removes an initiative record and cascades
func (r *csiRepository) DeleteInitiative(ctx context.Context, id string) error {
	query := `DELETE FROM csi_initiatives WHERE id = ?`
	res, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrCSIInitiativeNotFound
	}
	return nil
}

// ListInitiatives queries initiatives matching filter criteria
func (r *csiRepository) ListInitiatives(ctx context.Context, filter model.CSIFilterCriteria) ([]model.CSIInitiative, int, error) {
	var conditions []string
	var args []interface{}

	if filter.WorkspaceID != "" {
		conditions = append(conditions, "workspace_id = ?")
		args = append(args, filter.WorkspaceID)
	}
	if filter.StrategicAlignment != "" {
		conditions = append(conditions, "strategic_alignment = ?")
		args = append(args, filter.StrategicAlignment)
	}
	if filter.Priority != "" {
		conditions = append(conditions, "priority = ?")
		args = append(args, filter.Priority)
	}
	if filter.Status != "" {
		conditions = append(conditions, "status = ?")
		args = append(args, filter.Status)
	}
	if filter.Source != "" {
		conditions = append(conditions, "source = ?")
		args = append(args, filter.Source)
	}
	if filter.CIID != "" {
		conditions = append(conditions, "ci_id = ?")
		args = append(args, filter.CIID)
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM csi_initiatives %s", whereClause)
	var total int
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
	SELECT id, workspace_id, initiative_code, title, description,
	       strategic_alignment, priority, status, source,
	       source_reference_id, ci_id, owner_id,
	       target_completion_date, completed_at, created_at, updated_at
	FROM csi_initiatives %s
	ORDER BY created_at DESC
	LIMIT ? OFFSET ?`, whereClause)

	queryArgs := append(args, limit, offset)
	rows, err := r.db.QueryContext(ctx, query, queryArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var initiatives []model.CSIInitiative
	for rows.Next() {
		var i model.CSIInitiative
		var desc, sourceRef, ciID, ownerID sql.NullString
		var completedAt sql.NullTime

		err := rows.Scan(
			&i.ID, &i.WorkspaceID, &i.InitiativeCode, &i.Title, &desc,
			&i.StrategicAlignment, &i.Priority, &i.Status, &i.Source,
			&sourceRef, &ciID, &ownerID,
			&i.TargetCompletionDate, &completedAt, &i.CreatedAt, &i.UpdatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		if desc.Valid {
			i.Description = desc.String
		}
		if sourceRef.Valid {
			i.SourceReferenceID = sourceRef.String
		}
		if ciID.Valid {
			i.CIID = ciID.String
		}
		if ownerID.Valid {
			i.OwnerID = ownerID.String
		}
		if completedAt.Valid {
			t := completedAt.Time
			i.CompletedAt = &t
		}
		initiatives = append(initiatives, i)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	return initiatives, total, nil
}

// CreateBenchmark adds a baseline/target metric to an initiative
func (r *csiRepository) CreateBenchmark(ctx context.Context, benchmark *model.CSIBenchmark) error {
	if err := benchmark.Validate(); err != nil {
		return err
	}
	if benchmark.ID == "" {
		benchmark.ID = uuid.New().String()
	}
	now := time.Now().UTC()
	if benchmark.CreatedAt.IsZero() {
		benchmark.CreatedAt = now
	}
	benchmark.UpdatedAt = now

	query := `
	INSERT INTO csi_benchmarks (
		id, workspace_id, initiative_id, metric_name, unit_of_measure,
		baseline_value, current_value, target_value, verified_at,
		notes, created_at, updated_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := r.db.ExecContext(ctx, query,
		benchmark.ID, benchmark.WorkspaceID, benchmark.InitiativeID, benchmark.MetricName, benchmark.UnitOfMeasure,
		benchmark.BaselineValue, benchmark.CurrentValue, benchmark.TargetValue, benchmark.VerifiedAt,
		benchmark.Notes, benchmark.CreatedAt, benchmark.UpdatedAt,
	)
	return err
}

// GetBenchmarkByID retrieves a benchmark metric by ID
func (r *csiRepository) GetBenchmarkByID(ctx context.Context, id string) (*model.CSIBenchmark, error) {
	query := `
	SELECT id, workspace_id, initiative_id, metric_name, unit_of_measure,
	       baseline_value, current_value, target_value, verified_at,
	       notes, created_at, updated_at
	FROM csi_benchmarks WHERE id = ?`

	var b model.CSIBenchmark
	var notes sql.NullString
	var verifiedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&b.ID, &b.WorkspaceID, &b.InitiativeID, &b.MetricName, &b.UnitOfMeasure,
		&b.BaselineValue, &b.CurrentValue, &b.TargetValue, &verifiedAt,
		&notes, &b.CreatedAt, &b.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrCSIBenchmarkNotFound
		}
		return nil, err
	}
	if notes.Valid {
		b.Notes = notes.String
	}
	if verifiedAt.Valid {
		t := verifiedAt.Time
		b.VerifiedAt = &t
	}
	return &b, nil
}

// UpdateBenchmark updates current value, target, and verification timestamp of a benchmark
func (r *csiRepository) UpdateBenchmark(ctx context.Context, benchmark *model.CSIBenchmark) error {
	if err := benchmark.Validate(); err != nil {
		return err
	}
	benchmark.UpdatedAt = time.Now().UTC()

	query := `
	UPDATE csi_benchmarks SET
		metric_name = ?, unit_of_measure = ?, baseline_value = ?,
		current_value = ?, target_value = ?, verified_at = ?,
		notes = ?, updated_at = ?
	WHERE id = ?`

	res, err := r.db.ExecContext(ctx, query,
		benchmark.MetricName, benchmark.UnitOfMeasure, benchmark.BaselineValue,
		benchmark.CurrentValue, benchmark.TargetValue, benchmark.VerifiedAt,
		benchmark.Notes, benchmark.UpdatedAt,
		benchmark.ID,
	)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrCSIBenchmarkNotFound
	}
	return nil
}

// DeleteBenchmark removes a benchmark metric
func (r *csiRepository) DeleteBenchmark(ctx context.Context, id string) error {
	query := `DELETE FROM csi_benchmarks WHERE id = ?`
	res, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrCSIBenchmarkNotFound
	}
	return nil
}

// ListBenchmarks lists all benchmark metrics associated with an initiative
func (r *csiRepository) ListBenchmarks(ctx context.Context, initiativeID string) ([]model.CSIBenchmark, error) {
	query := `
	SELECT id, workspace_id, initiative_id, metric_name, unit_of_measure,
	       baseline_value, current_value, target_value, verified_at,
	       notes, created_at, updated_at
	FROM csi_benchmarks WHERE initiative_id = ?
	ORDER BY created_at ASC`

	rows, err := r.db.QueryContext(ctx, query, initiativeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var benchmarks []model.CSIBenchmark
	for rows.Next() {
		var b model.CSIBenchmark
		var notes sql.NullString
		var verifiedAt sql.NullTime

		err := rows.Scan(
			&b.ID, &b.WorkspaceID, &b.InitiativeID, &b.MetricName, &b.UnitOfMeasure,
			&b.BaselineValue, &b.CurrentValue, &b.TargetValue, &verifiedAt,
			&notes, &b.CreatedAt, &b.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		if notes.Valid {
			b.Notes = notes.String
		}
		if verifiedAt.Valid {
			t := verifiedAt.Time
			b.VerifiedAt = &t
		}
		benchmarks = append(benchmarks, b)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return benchmarks, nil
}

// SaveBenefit records or updates benefit realization parameters for an initiative
func (r *csiRepository) SaveBenefit(ctx context.Context, benefit *model.CSIBenefit) error {
	if err := benefit.Validate(); err != nil {
		return err
	}
	if benefit.ID == "" {
		benefit.ID = uuid.New().String()
	}
	now := time.Now().UTC()
	if benefit.CreatedAt.IsZero() {
		benefit.CreatedAt = now
	}
	benefit.UpdatedAt = now

	existing, err := r.GetBenefitByInitiativeID(ctx, benefit.InitiativeID)
	if err == nil && existing != nil {
		benefit.ID = existing.ID
		benefit.CreatedAt = existing.CreatedAt
		updateQuery := `
		UPDATE csi_benefits SET
			projected_financial_savings = ?, realized_financial_savings = ?,
			effort_hours_invested = ?, hours_saved_per_month = ?,
			risk_reduction_percentage = ?, status = ?,
			realized_at = ?, updated_at = ?
		WHERE id = ?`

		_, err = r.db.ExecContext(ctx, updateQuery,
			benefit.ProjectedFinancialSavings, benefit.RealizedFinancialSavings,
			benefit.EffortHoursInvested, benefit.HoursSavedPerMonth,
			benefit.RiskReductionPercentage, benefit.Status,
			benefit.RealizedAt, benefit.UpdatedAt,
			benefit.ID,
		)
		return err
	}

	insertQuery := `
	INSERT INTO csi_benefits (
		id, workspace_id, initiative_id, projected_financial_savings,
		realized_financial_savings, effort_hours_invested, hours_saved_per_month,
		risk_reduction_percentage, status, realized_at, created_at, updated_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err = r.db.ExecContext(ctx, insertQuery,
		benefit.ID, benefit.WorkspaceID, benefit.InitiativeID, benefit.ProjectedFinancialSavings,
		benefit.RealizedFinancialSavings, benefit.EffortHoursInvested, benefit.HoursSavedPerMonth,
		benefit.RiskReductionPercentage, benefit.Status, benefit.RealizedAt, benefit.CreatedAt, benefit.UpdatedAt,
	)
	return err
}

// GetBenefitByInitiativeID retrieves benefit tracking for an initiative
func (r *csiRepository) GetBenefitByInitiativeID(ctx context.Context, initiativeID string) (*model.CSIBenefit, error) {
	query := `
	SELECT id, workspace_id, initiative_id, projected_financial_savings,
	       realized_financial_savings, effort_hours_invested, hours_saved_per_month,
	       risk_reduction_percentage, status, realized_at, created_at, updated_at
	FROM csi_benefits WHERE initiative_id = ?`

	var b model.CSIBenefit
	var realizedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, initiativeID).Scan(
		&b.ID, &b.WorkspaceID, &b.InitiativeID, &b.ProjectedFinancialSavings,
		&b.RealizedFinancialSavings, &b.EffortHoursInvested, &b.HoursSavedPerMonth,
		&b.RiskReductionPercentage, &b.Status, &realizedAt, &b.CreatedAt, &b.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrCSIBenefitNotFound
		}
		return nil, err
	}
	if realizedAt.Valid {
		t := realizedAt.Time
		b.RealizedAt = &t
	}
	return &b, nil
}

// UpdateBenefit updates benefit metrics
func (r *csiRepository) UpdateBenefit(ctx context.Context, benefit *model.CSIBenefit) error {
	return r.SaveBenefit(ctx, benefit)
}

// CreateAction creates an action item under an initiative
func (r *csiRepository) CreateAction(ctx context.Context, action *model.CSIAction) error {
	if err := action.Validate(); err != nil {
		return err
	}
	if action.ID == "" {
		action.ID = uuid.New().String()
	}
	now := time.Now().UTC()
	if action.CreatedAt.IsZero() {
		action.CreatedAt = now
	}
	action.UpdatedAt = now

	if action.StepNumber <= 0 {
		var maxStep int
		_ = r.db.QueryRowContext(ctx, "SELECT COALESCE(MAX(step_number), 0) FROM csi_actions WHERE initiative_id = ?", action.InitiativeID).Scan(&maxStep)
		action.StepNumber = maxStep + 1
	}

	query := `
	INSERT INTO csi_actions (
		id, workspace_id, initiative_id, step_number,
		description, assignee_id, due_date, status,
		verification_evidence, completed_at, created_at, updated_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := r.db.ExecContext(ctx, query,
		action.ID, action.WorkspaceID, action.InitiativeID, action.StepNumber,
		action.Description, action.AssigneeID, action.DueDate, action.Status,
		action.VerificationEvidence, action.CompletedAt, action.CreatedAt, action.UpdatedAt,
	)
	return err
}

// GetActionByID retrieves an action item by ID
func (r *csiRepository) GetActionByID(ctx context.Context, id string) (*model.CSIAction, error) {
	query := `
	SELECT id, workspace_id, initiative_id, step_number,
	       description, assignee_id, due_date, status,
	       verification_evidence, completed_at, created_at, updated_at
	FROM csi_actions WHERE id = ?`

	var a model.CSIAction
	var assigneeID, evidence sql.NullString
	var completedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&a.ID, &a.WorkspaceID, &a.InitiativeID, &a.StepNumber,
		&a.Description, &assigneeID, &a.DueDate, &a.Status,
		&evidence, &completedAt, &a.CreatedAt, &a.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrCSIActionNotFound
		}
		return nil, err
	}
	if assigneeID.Valid {
		a.AssigneeID = assigneeID.String
	}
	if evidence.Valid {
		a.VerificationEvidence = evidence.String
	}
	if completedAt.Valid {
		t := completedAt.Time
		a.CompletedAt = &t
	}
	return &a, nil
}

// UpdateAction updates fields of an action step
func (r *csiRepository) UpdateAction(ctx context.Context, action *model.CSIAction) error {
	if err := action.Validate(); err != nil {
		return err
	}
	action.UpdatedAt = time.Now().UTC()

	query := `
	UPDATE csi_actions SET
		step_number = ?, description = ?, assignee_id = ?,
		due_date = ?, status = ?, verification_evidence = ?,
		completed_at = ?, updated_at = ?
	WHERE id = ?`

	res, err := r.db.ExecContext(ctx, query,
		action.StepNumber, action.Description, action.AssigneeID,
		action.DueDate, action.Status, action.VerificationEvidence,
		action.CompletedAt, action.UpdatedAt,
		action.ID,
	)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrCSIActionNotFound
	}
	return nil
}

// DeleteAction removes an action item
func (r *csiRepository) DeleteAction(ctx context.Context, id string) error {
	query := `DELETE FROM csi_actions WHERE id = ?`
	res, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrCSIActionNotFound
	}
	return nil
}

// ListActions returns action items for an initiative ordered by step number
func (r *csiRepository) ListActions(ctx context.Context, initiativeID string) ([]model.CSIAction, error) {
	query := `
	SELECT id, workspace_id, initiative_id, step_number,
	       description, assignee_id, due_date, status,
	       verification_evidence, completed_at, created_at, updated_at
	FROM csi_actions WHERE initiative_id = ?
	ORDER BY step_number ASC`

	rows, err := r.db.QueryContext(ctx, query, initiativeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var actions []model.CSIAction
	for rows.Next() {
		var a model.CSIAction
		var assigneeID, evidence sql.NullString
		var completedAt sql.NullTime

		err := rows.Scan(
			&a.ID, &a.WorkspaceID, &a.InitiativeID, &a.StepNumber,
			&a.Description, &assigneeID, &a.DueDate, &a.Status,
			&evidence, &completedAt, &a.CreatedAt, &a.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		if assigneeID.Valid {
			a.AssigneeID = assigneeID.String
		}
		if evidence.Valid {
			a.VerificationEvidence = evidence.String
		}
		if completedAt.Valid {
			t := completedAt.Time
			a.CompletedAt = &t
		}
		actions = append(actions, a)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return actions, nil
}

// CreateReview logs a governance or post-implementation review
func (r *csiRepository) CreateReview(ctx context.Context, review *model.CSIReview) error {
	if err := review.Validate(); err != nil {
		return err
	}
	if review.ID == "" {
		review.ID = uuid.New().String()
	}
	now := time.Now().UTC()
	if review.CreatedAt.IsZero() {
		review.CreatedAt = now
	}
	review.UpdatedAt = now

	query := `
	INSERT INTO csi_reviews (
		id, workspace_id, initiative_id, review_date,
		reviewer_id, outcome, findings, lessons_learned,
		next_review_date, created_at, updated_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := r.db.ExecContext(ctx, query,
		review.ID, review.WorkspaceID, review.InitiativeID, review.ReviewDate,
		review.ReviewerID, review.Outcome, review.Findings, review.LessonsLearned,
		review.NextReviewDate, review.CreatedAt, review.UpdatedAt,
	)
	return err
}

// GetReviewByID retrieves a review entry by ID
func (r *csiRepository) GetReviewByID(ctx context.Context, id string) (*model.CSIReview, error) {
	query := `
	SELECT id, workspace_id, initiative_id, review_date,
	       reviewer_id, outcome, findings, lessons_learned,
	       next_review_date, created_at, updated_at
	FROM csi_reviews WHERE id = ?`

	var rv model.CSIReview
	var lessons sql.NullString
	var nextRev sql.NullTime

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&rv.ID, &rv.WorkspaceID, &rv.InitiativeID, &rv.ReviewDate,
		&rv.ReviewerID, &rv.Outcome, &rv.Findings, &lessons,
		&nextRev, &rv.CreatedAt, &rv.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrCSIReviewNotFound
		}
		return nil, err
	}
	if lessons.Valid {
		rv.LessonsLearned = lessons.String
	}
	if nextRev.Valid {
		t := nextRev.Time
		rv.NextReviewDate = &t
	}
	return &rv, nil
}

// ListReviews returns all reviews associated with an initiative
func (r *csiRepository) ListReviews(ctx context.Context, initiativeID string) ([]model.CSIReview, error) {
	query := `
	SELECT id, workspace_id, initiative_id, review_date,
	       reviewer_id, outcome, findings, lessons_learned,
	       next_review_date, created_at, updated_at
	FROM csi_reviews WHERE initiative_id = ?
	ORDER BY review_date DESC`

	rows, err := r.db.QueryContext(ctx, query, initiativeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reviews []model.CSIReview
	for rows.Next() {
		var rv model.CSIReview
		var lessons sql.NullString
		var nextRev sql.NullTime

		err := rows.Scan(
			&rv.ID, &rv.WorkspaceID, &rv.InitiativeID, &rv.ReviewDate,
			&rv.ReviewerID, &rv.Outcome, &rv.Findings, &lessons,
			&nextRev, &rv.CreatedAt, &rv.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		if lessons.Valid {
			rv.LessonsLearned = lessons.String
		}
		if nextRev.Valid {
			t := nextRev.Time
			rv.NextReviewDate = &t
		}
		reviews = append(reviews, rv)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return reviews, nil
}

// GetCSIAnalytics computes executive continual improvement metrics, realized financial savings, and maturity index
func (r *csiRepository) GetCSIAnalytics(ctx context.Context, workspaceID string) (*model.CSIAnalytics, error) {
	analytics := &model.CSIAnalytics{
		InitiativesByAlignment: []model.AlignmentSummary{},
	}

	// 1. Initiative counts
	_ = r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM csi_initiatives WHERE workspace_id = ?", workspaceID).Scan(&analytics.TotalInitiatives)
	_ = r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM csi_initiatives WHERE workspace_id = ?", workspaceID).Scan(&analytics.TotalInitiatives)
	_ = r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM csi_initiatives WHERE workspace_id = ? AND status IN ('identified', 'approved', 'in_progress', 'on_hold')", workspaceID).Scan(&analytics.ActiveInitiatives)
	_ = r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM csi_initiatives WHERE workspace_id = ? AND status = 'completed'", workspaceID).Scan(&analytics.CompletedInitiatives)

	if analytics.TotalInitiatives > 0 {
		analytics.CompletionRate = (float64(analytics.CompletedInitiatives) / float64(analytics.TotalInitiatives)) * 100.0
	}

	// 2. Financial and Effort Benefits
	var projSavings, realSavings, hoursSaved, effortHours sql.NullFloat64
	_ = r.db.QueryRowContext(ctx, `
		SELECT SUM(projected_financial_savings), SUM(realized_financial_savings),
		       SUM(hours_saved_per_month), SUM(effort_hours_invested)
		FROM csi_benefits WHERE workspace_id = ?`, workspaceID).Scan(&projSavings, &realSavings, &hoursSaved, &effortHours)

	if projSavings.Valid {
		analytics.TotalProjectedSavings = projSavings.Float64
	}
	if realSavings.Valid {
		analytics.TotalRealizedSavings = realSavings.Float64
	}
	if hoursSaved.Valid {
		analytics.TotalHoursSavedPerMonth = hoursSaved.Float64
	}
	if effortHours.Valid {
		analytics.TotalEffortHoursInvested = effortHours.Float64
	}

	// Calculate Average ROI
	benefitRows, err := r.db.QueryContext(ctx, "SELECT effort_hours_invested, realized_financial_savings, hours_saved_per_month FROM csi_benefits WHERE workspace_id = ?", workspaceID)
	if err == nil && benefitRows != nil {
		defer benefitRows.Close()
		var totalROI float64
		var count int
		for benefitRows.Next() {
			var b model.CSIBenefit
			if scanErr := benefitRows.Scan(&b.EffortHoursInvested, &b.RealizedFinancialSavings, &b.HoursSavedPerMonth); scanErr == nil {
				roi := b.CalculateROI(75.0)
				totalROI += roi
				count++
			}
		}
		if count > 0 {
			analytics.AverageROI = totalROI / float64(count)
		}
	}

	// 3. Alignment Breakdown
	alignRows, err := r.db.QueryContext(ctx, `
		SELECT strategic_alignment, COUNT(*)
		FROM csi_initiatives
		WHERE workspace_id = ?
		GROUP BY strategic_alignment
		ORDER BY COUNT(*) DESC`, workspaceID)
	if err == nil && alignRows != nil {
		defer alignRows.Close()
		for alignRows.Next() {
			var s model.AlignmentSummary
			if scanErr := alignRows.Scan(&s.Alignment, &s.Count); scanErr == nil {
				analytics.InitiativesByAlignment = append(analytics.InitiativesByAlignment, s)
			}
		}
	}

	analytics.CalculateMaturityScore()
	return analytics, nil
}
