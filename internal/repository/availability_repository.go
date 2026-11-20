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
	ErrAvailabilityTargetNotFound      = model.ErrAvailabilityTargetNotFound
	ErrAvailabilityTargetNameRequired  = model.ErrAvailabilityTargetNameRequired
	ErrAvailabilityTargetNameConflict  = errors.New("availability target with this name already exists in workspace")
	ErrOutageEventNotFound             = model.ErrOutageEventNotFound
	ErrAvailabilityMeasurementNotFound = model.ErrAvailabilityMeasurementNotFound
	ErrSPOFRecordNotFound              = model.ErrSPOFRecordNotFound
	ErrAvailabilityTestNotFound        = model.ErrAvailabilityTestNotFound
)

// AvailabilityRepository defines data access methods for ITIL Availability Management
type AvailabilityRepository interface {
	// Targets
	CreateTarget(ctx context.Context, target *model.AvailabilityTarget) error
	GetTargetByID(ctx context.Context, id string) (*model.AvailabilityTarget, error)
	GetTargetByName(ctx context.Context, workspaceID string, name string) (*model.AvailabilityTarget, error)
	UpdateTarget(ctx context.Context, target *model.AvailabilityTarget) error
	DeleteTarget(ctx context.Context, id string) error
	ListTargets(ctx context.Context, filter model.AvailabilityFilterCriteria) ([]model.AvailabilityTarget, int, error)

	// Outages
	CreateOutage(ctx context.Context, outage *model.OutageEvent) error
	GetOutageByID(ctx context.Context, id string) (*model.OutageEvent, error)
	UpdateOutage(ctx context.Context, outage *model.OutageEvent) error
	DeleteOutage(ctx context.Context, id string) error
	ListOutages(ctx context.Context, filter model.AvailabilityFilterCriteria) ([]model.OutageEvent, int, error)

	// Measurements
	SaveMeasurement(ctx context.Context, m *model.AvailabilityMeasurement) error
	GetMeasurementByID(ctx context.Context, id string) (*model.AvailabilityMeasurement, error)
	GetMeasurementByPeriod(ctx context.Context, targetID string, period string) (*model.AvailabilityMeasurement, error)
	ListMeasurements(ctx context.Context, filter model.AvailabilityFilterCriteria) ([]model.AvailabilityMeasurement, int, error)

	// SPOFs
	CreateSPOF(ctx context.Context, spof *model.SPOFRecord) error
	GetSPOFByID(ctx context.Context, id string) (*model.SPOFRecord, error)
	UpdateSPOF(ctx context.Context, spof *model.SPOFRecord) error
	DeleteSPOF(ctx context.Context, id string) error
	ListSPOFs(ctx context.Context, filter model.AvailabilityFilterCriteria) ([]model.SPOFRecord, int, error)

	// Availability Tests
	CreateTest(ctx context.Context, test *model.AvailabilityTest) error
	GetTestByID(ctx context.Context, id string) (*model.AvailabilityTest, error)
	UpdateTest(ctx context.Context, test *model.AvailabilityTest) error
	DeleteTest(ctx context.Context, id string) error
	ListTests(ctx context.Context, filter model.AvailabilityFilterCriteria) ([]model.AvailabilityTest, int, error)

	// Posture Analytics
	GetAvailabilityAnalytics(ctx context.Context, workspaceID string) (*model.AvailabilityAnalytics, error)
}

type availabilityRepository struct {
	db *sql.DB
}

// NewAvailabilityRepository creates a new SQLite repository for ITIL Availability Management
func NewAvailabilityRepository(db *sql.DB) AvailabilityRepository {
	return &availabilityRepository{db: db}
}

// CreateTarget persists a new availability target into the database
func (r *availabilityRepository) CreateTarget(ctx context.Context, target *model.AvailabilityTarget) error {
	if err := target.Validate(); err != nil {
		return err
	}
	if target.ID == "" {
		target.ID = uuid.New().String()
	}
	now := time.Now().UTC()
	if target.CreatedAt.IsZero() {
		target.CreatedAt = now
	}
	target.UpdatedAt = now

	// Check for unique name in workspace
	existing, err := r.GetTargetByName(ctx, target.WorkspaceID, target.Name)
	if err == nil && existing != nil {
		return ErrAvailabilityTargetNameConflict
	}

	query := `
	INSERT INTO availability_targets (
		id, workspace_id, name, target_percentage, measurement_window,
		service_hours, committed_mtbf_hours, target_mttr_minutes,
		ci_id, status, description, created_at, updated_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err = r.db.ExecContext(ctx, query,
		target.ID, target.WorkspaceID, target.Name, target.TargetPercentage, target.MeasurementWindow,
		target.ServiceHours, target.CommittedMTBFHours, target.TargetMTTRMinutes,
		target.CIID, target.Status, target.Description, target.CreatedAt, target.UpdatedAt,
	)
	if err != nil {
		return err
	}
	return nil
}

// GetTargetByID retrieves an availability target by its unique ID
func (r *availabilityRepository) GetTargetByID(ctx context.Context, id string) (*model.AvailabilityTarget, error) {
	query := `
	SELECT id, workspace_id, name, target_percentage, measurement_window,
	       service_hours, committed_mtbf_hours, target_mttr_minutes,
	       ci_id, status, description, created_at, updated_at
	FROM availability_targets WHERE id = ?`

	var t model.AvailabilityTarget
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&t.ID, &t.WorkspaceID, &t.Name, &t.TargetPercentage, &t.MeasurementWindow,
		&t.ServiceHours, &t.CommittedMTBFHours, &t.TargetMTTRMinutes,
		&t.CIID, &t.Status, &t.Description, &t.CreatedAt, &t.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrAvailabilityTargetNotFound
		}
		return nil, err
	}
	return &t, nil
}

// GetTargetByName retrieves an availability target by workspace and name
func (r *availabilityRepository) GetTargetByName(ctx context.Context, workspaceID string, name string) (*model.AvailabilityTarget, error) {
	query := `
	SELECT id, workspace_id, name, target_percentage, measurement_window,
	       service_hours, committed_mtbf_hours, target_mttr_minutes,
	       ci_id, status, description, created_at, updated_at
	FROM availability_targets WHERE workspace_id = ? AND name = ?`

	var t model.AvailabilityTarget
	err := r.db.QueryRowContext(ctx, query, workspaceID, name).Scan(
		&t.ID, &t.WorkspaceID, &t.Name, &t.TargetPercentage, &t.MeasurementWindow,
		&t.ServiceHours, &t.CommittedMTBFHours, &t.TargetMTTRMinutes,
		&t.CIID, &t.Status, &t.Description, &t.CreatedAt, &t.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrAvailabilityTargetNotFound
		}
		return nil, err
	}
	return &t, nil
}

// UpdateTarget updates an existing target in SQLite
func (r *availabilityRepository) UpdateTarget(ctx context.Context, target *model.AvailabilityTarget) error {
	if err := target.Validate(); err != nil {
		return err
	}
	target.UpdatedAt = time.Now().UTC()

	query := `
	UPDATE availability_targets SET
		name = ?, target_percentage = ?, measurement_window = ?,
		service_hours = ?, committed_mtbf_hours = ?, target_mttr_minutes = ?,
		ci_id = ?, status = ?, description = ?, updated_at = ?
	WHERE id = ?`

	res, err := r.db.ExecContext(ctx, query,
		target.Name, target.TargetPercentage, target.MeasurementWindow,
		target.ServiceHours, target.CommittedMTBFHours, target.TargetMTTRMinutes,
		target.CIID, target.Status, target.Description, target.UpdatedAt,
		target.ID,
	)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrAvailabilityTargetNotFound
	}
	return nil
}

// DeleteTarget deletes an availability target by ID
func (r *availabilityRepository) DeleteTarget(ctx context.Context, id string) error {
	query := `DELETE FROM availability_targets WHERE id = ?`
	res, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrAvailabilityTargetNotFound
	}
	return nil
}

// ListTargets returns a filtered list of availability targets
func (r *availabilityRepository) ListTargets(ctx context.Context, filter model.AvailabilityFilterCriteria) ([]model.AvailabilityTarget, int, error) {
	var conditions []string
	var args []interface{}

	if filter.WorkspaceID != "" {
		conditions = append(conditions, "workspace_id = ?")
		args = append(args, filter.WorkspaceID)
	}
	if filter.Status != "" {
		conditions = append(conditions, "status = ?")
		args = append(args, filter.Status)
	}
	if filter.CIID != "" {
		conditions = append(conditions, "ci_id = ?")
		args = append(args, filter.CIID)
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM availability_targets %s", whereClause)
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
	SELECT id, workspace_id, name, target_percentage, measurement_window,
	       service_hours, committed_mtbf_hours, target_mttr_minutes,
	       ci_id, status, description, created_at, updated_at
	FROM availability_targets %s
	ORDER BY created_at DESC
	LIMIT ? OFFSET ?`, whereClause)

	queryArgs := append(args, limit, offset)
	rows, err := r.db.QueryContext(ctx, query, queryArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var targets []model.AvailabilityTarget
	for rows.Next() {
		var t model.AvailabilityTarget
		err := rows.Scan(
			&t.ID, &t.WorkspaceID, &t.Name, &t.TargetPercentage, &t.MeasurementWindow,
			&t.ServiceHours, &t.CommittedMTBFHours, &t.TargetMTTRMinutes,
			&t.CIID, &t.Status, &t.Description, &t.CreatedAt, &t.UpdatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		targets = append(targets, t)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	return targets, total, nil
}

func (r *availabilityRepository) CreateOutage(ctx context.Context, outage *model.OutageEvent) error {
	return errors.New("not implemented")
}

func (r *availabilityRepository) GetOutageByID(ctx context.Context, id string) (*model.OutageEvent, error) {
	return nil, errors.New("not implemented")
}

func (r *availabilityRepository) UpdateOutage(ctx context.Context, outage *model.OutageEvent) error {
	return errors.New("not implemented")
}

func (r *availabilityRepository) DeleteOutage(ctx context.Context, id string) error {
	return errors.New("not implemented")
}

func (r *availabilityRepository) ListOutages(ctx context.Context, filter model.AvailabilityFilterCriteria) ([]model.OutageEvent, int, error) {
	return nil, 0, errors.New("not implemented")
}

func (r *availabilityRepository) SaveMeasurement(ctx context.Context, m *model.AvailabilityMeasurement) error {
	return errors.New("not implemented")
}

func (r *availabilityRepository) GetMeasurementByID(ctx context.Context, id string) (*model.AvailabilityMeasurement, error) {
	return nil, errors.New("not implemented")
}

func (r *availabilityRepository) GetMeasurementByPeriod(ctx context.Context, targetID string, period string) (*model.AvailabilityMeasurement, error) {
	return nil, errors.New("not implemented")
}

func (r *availabilityRepository) ListMeasurements(ctx context.Context, filter model.AvailabilityFilterCriteria) ([]model.AvailabilityMeasurement, int, error) {
	return nil, 0, errors.New("not implemented")
}

func (r *availabilityRepository) CreateSPOF(ctx context.Context, spof *model.SPOFRecord) error {
	return errors.New("not implemented")
}

func (r *availabilityRepository) GetSPOFByID(ctx context.Context, id string) (*model.SPOFRecord, error) {
	return nil, errors.New("not implemented")
}

func (r *availabilityRepository) UpdateSPOF(ctx context.Context, spof *model.SPOFRecord) error {
	return errors.New("not implemented")
}

func (r *availabilityRepository) DeleteSPOF(ctx context.Context, id string) error {
	return errors.New("not implemented")
}

func (r *availabilityRepository) ListSPOFs(ctx context.Context, filter model.AvailabilityFilterCriteria) ([]model.SPOFRecord, int, error) {
	return nil, 0, errors.New("not implemented")
}

func (r *availabilityRepository) CreateTest(ctx context.Context, test *model.AvailabilityTest) error {
	return errors.New("not implemented")
}

func (r *availabilityRepository) GetTestByID(ctx context.Context, id string) (*model.AvailabilityTest, error) {
	return nil, errors.New("not implemented")
}

func (r *availabilityRepository) UpdateTest(ctx context.Context, test *model.AvailabilityTest) error {
	return errors.New("not implemented")
}

func (r *availabilityRepository) DeleteTest(ctx context.Context, id string) error {
	return errors.New("not implemented")
}

func (r *availabilityRepository) ListTests(ctx context.Context, filter model.AvailabilityFilterCriteria) ([]model.AvailabilityTest, int, error) {
	return nil, 0, errors.New("not implemented")
}

func (r *availabilityRepository) GetAvailabilityAnalytics(ctx context.Context, workspaceID string) (*model.AvailabilityAnalytics, error) {
	return nil, errors.New("not implemented")
}
