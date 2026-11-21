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

func (r *csiRepository) CreateBenchmark(ctx context.Context, benchmark *model.CSIBenchmark) error {
	return errors.New("not implemented")
}

func (r *csiRepository) GetBenchmarkByID(ctx context.Context, id string) (*model.CSIBenchmark, error) {
	return nil, errors.New("not implemented")
}

func (r *csiRepository) UpdateBenchmark(ctx context.Context, benchmark *model.CSIBenchmark) error {
	return errors.New("not implemented")
}

func (r *csiRepository) DeleteBenchmark(ctx context.Context, id string) error {
	return errors.New("not implemented")
}

func (r *csiRepository) ListBenchmarks(ctx context.Context, initiativeID string) ([]model.CSIBenchmark, error) {
	return nil, errors.New("not implemented")
}

func (r *csiRepository) SaveBenefit(ctx context.Context, benefit *model.CSIBenefit) error {
	return errors.New("not implemented")
}

func (r *csiRepository) GetBenefitByInitiativeID(ctx context.Context, initiativeID string) (*model.CSIBenefit, error) {
	return nil, errors.New("not implemented")
}

func (r *csiRepository) UpdateBenefit(ctx context.Context, benefit *model.CSIBenefit) error {
	return errors.New("not implemented")
}

func (r *csiRepository) CreateAction(ctx context.Context, action *model.CSIAction) error {
	return errors.New("not implemented")
}

func (r *csiRepository) GetActionByID(ctx context.Context, id string) (*model.CSIAction, error) {
	return nil, errors.New("not implemented")
}

func (r *csiRepository) UpdateAction(ctx context.Context, action *model.CSIAction) error {
	return errors.New("not implemented")
}

func (r *csiRepository) DeleteAction(ctx context.Context, id string) error {
	return errors.New("not implemented")
}

func (r *csiRepository) ListActions(ctx context.Context, initiativeID string) ([]model.CSIAction, error) {
	return nil, errors.New("not implemented")
}

func (r *csiRepository) CreateReview(ctx context.Context, review *model.CSIReview) error {
	return errors.New("not implemented")
}

func (r *csiRepository) GetReviewByID(ctx context.Context, id string) (*model.CSIReview, error) {
	return nil, errors.New("not implemented")
}

func (r *csiRepository) ListReviews(ctx context.Context, initiativeID string) ([]model.CSIReview, error) {
	return nil, errors.New("not implemented")
}

func (r *csiRepository) GetCSIAnalytics(ctx context.Context, workspaceID string) (*model.CSIAnalytics, error) {
	return nil, errors.New("not implemented")
}
