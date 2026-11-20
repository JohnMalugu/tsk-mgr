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
	ErrCapacityPoolNotFound      = model.ErrCapacityPoolNotFound
	ErrCapacityPoolNameRequired  = model.ErrCapacityPoolNameRequired
	ErrCapacityPoolNameConflict  = errors.New("capacity pool with this name already exists in workspace")
	ErrInvalidResourceType       = model.ErrInvalidResourceType
	ErrInvalidPoolStatus         = model.ErrInvalidPoolStatus
	ErrInvalidTotalCapacity      = model.ErrInvalidTotalCapacity
	ErrInvalidThresholdPercent   = model.ErrInvalidThresholdPercent
	ErrCapacityMetricNotFound    = model.ErrCapacityMetricNotFound
	ErrCapacityPlanNotFound      = model.ErrCapacityPlanNotFound
	ErrInvalidPlanHorizon        = model.ErrInvalidPlanHorizon
	ErrInvalidPlanStatus         = model.ErrInvalidPlanStatus
	ErrBottleneckNotFound        = model.ErrBottleneckNotFound
	ErrInvalidBottleneckSeverity = model.ErrInvalidBottleneckSeverity
	ErrInvalidBottleneckStatus   = model.ErrInvalidBottleneckStatus
)

// CapacityRepository defines data access methods for ITIL Capacity and Performance Management
type CapacityRepository interface {
	// Pools
	CreatePool(ctx context.Context, pool *model.CapacityPool) error
	GetPoolByID(ctx context.Context, id string) (*model.CapacityPool, error)
	GetPoolByName(ctx context.Context, workspaceID string, name string) (*model.CapacityPool, error)
	UpdatePool(ctx context.Context, pool *model.CapacityPool) error
	DeletePool(ctx context.Context, id string) error
	ListPools(ctx context.Context, filter model.CapacityPoolFilterCriteria) ([]model.CapacityPool, int, error)

	// Metrics & Telemetry
	RecordMetric(ctx context.Context, metric *model.CapacityMetric) error
	GetMetricByID(ctx context.Context, id string) (*model.CapacityMetric, error)
	ListMetrics(ctx context.Context, filter model.CapacityMetricFilterCriteria) ([]model.CapacityMetric, int, error)

	// Forecasts
	SaveForecast(ctx context.Context, forecast *model.CapacityForecast) error
	GetLatestForecastByPoolID(ctx context.Context, poolID string) (*model.CapacityForecast, error)

	// Capacity Plans
	CreatePlan(ctx context.Context, plan *model.CapacityPlan) error
	GetPlanByID(ctx context.Context, id string) (*model.CapacityPlan, error)
	UpdatePlan(ctx context.Context, plan *model.CapacityPlan) error
	DeletePlan(ctx context.Context, id string) error
	ListPlans(ctx context.Context, filter model.CapacityPlanFilterCriteria) ([]model.CapacityPlan, int, error)

	// Bottlenecks
	CreateBottleneck(ctx context.Context, b *model.BottleneckRecord) error
	GetBottleneckByID(ctx context.Context, id string) (*model.BottleneckRecord, error)
	UpdateBottleneck(ctx context.Context, b *model.BottleneckRecord) error
	DeleteBottleneck(ctx context.Context, id string) error
	ListBottlenecks(ctx context.Context, filter model.BottleneckFilterCriteria) ([]model.BottleneckRecord, int, error)

	// Posture Analytics
	GetCapacityAnalytics(ctx context.Context, workspaceID string) (*model.CapacityAnalyticsReport, error)
}

type sqliteCapacityRepository struct {
	db *sql.DB
}

// NewCapacityRepository creates a new CapacityRepository instance
func NewCapacityRepository(db *sql.DB) CapacityRepository {
	return &sqliteCapacityRepository{db: db}
}

// Pools
func (r *sqliteCapacityRepository) CreatePool(ctx context.Context, pool *model.CapacityPool) error {
	if pool.ID == "" {
		pool.ID = uuid.New().String()
	}
	now := time.Now().UTC()
	if pool.CreatedAt.IsZero() {
		pool.CreatedAt = now
	}
	pool.UpdatedAt = now
	if pool.WarningThresholdPercent <= 0 {
		pool.WarningThresholdPercent = 75.0
	}
	if pool.CriticalThresholdPercent <= 0 {
		pool.CriticalThresholdPercent = 90.0
	}
	pool.RecalculateStatus()

	query := `
	INSERT INTO capacity_pools (
		id, workspace_id, name, resource_type, ci_id,
		total_capacity, allocated_capacity, utilized_capacity,
		unit_of_measure, warning_threshold_percent, critical_threshold_percent,
		status, description, created_at, updated_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := r.db.ExecContext(ctx, query,
		pool.ID, pool.WorkspaceID, pool.Name, pool.ResourceType, pool.CIID,
		pool.TotalCapacity, pool.AllocatedCapacity, pool.UtilizedCapacity,
		pool.UnitOfMeasure, pool.WarningThresholdPercent, pool.CriticalThresholdPercent,
		pool.Status, pool.Description, pool.CreatedAt, pool.UpdatedAt,
	)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return ErrCapacityPoolNameConflict
		}
		return err
	}
	return nil
}

func (r *sqliteCapacityRepository) GetPoolByID(ctx context.Context, id string) (*model.CapacityPool, error) {
	query := `
	SELECT id, workspace_id, name, resource_type, ci_id,
	       total_capacity, allocated_capacity, utilized_capacity,
	       unit_of_measure, warning_threshold_percent, critical_threshold_percent,
	       status, description, created_at, updated_at
	FROM capacity_pools WHERE id = ?`

	var p model.CapacityPool
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&p.ID, &p.WorkspaceID, &p.Name, &p.ResourceType, &p.CIID,
		&p.TotalCapacity, &p.AllocatedCapacity, &p.UtilizedCapacity,
		&p.UnitOfMeasure, &p.WarningThresholdPercent, &p.CriticalThresholdPercent,
		&p.Status, &p.Description, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrCapacityPoolNotFound
		}
		return nil, err
	}
	return &p, nil
}

func (r *sqliteCapacityRepository) GetPoolByName(ctx context.Context, workspaceID string, name string) (*model.CapacityPool, error) {
	query := `
	SELECT id, workspace_id, name, resource_type, ci_id,
	       total_capacity, allocated_capacity, utilized_capacity,
	       unit_of_measure, warning_threshold_percent, critical_threshold_percent,
	       status, description, created_at, updated_at
	FROM capacity_pools WHERE workspace_id = ? AND name = ?`

	var p model.CapacityPool
	err := r.db.QueryRowContext(ctx, query, workspaceID, name).Scan(
		&p.ID, &p.WorkspaceID, &p.Name, &p.ResourceType, &p.CIID,
		&p.TotalCapacity, &p.AllocatedCapacity, &p.UtilizedCapacity,
		&p.UnitOfMeasure, &p.WarningThresholdPercent, &p.CriticalThresholdPercent,
		&p.Status, &p.Description, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrCapacityPoolNotFound
		}
		return nil, err
	}
	return &p, nil
}

func (r *sqliteCapacityRepository) UpdatePool(ctx context.Context, pool *model.CapacityPool) error {
	pool.UpdatedAt = time.Now().UTC()
	pool.RecalculateStatus()

	query := `
	UPDATE capacity_pools SET
		name = ?, resource_type = ?, ci_id = ?,
		total_capacity = ?, allocated_capacity = ?, utilized_capacity = ?,
		unit_of_measure = ?, warning_threshold_percent = ?, critical_threshold_percent = ?,
		status = ?, description = ?, updated_at = ?
	WHERE id = ?`

	res, err := r.db.ExecContext(ctx, query,
		pool.Name, pool.ResourceType, pool.CIID,
		pool.TotalCapacity, pool.AllocatedCapacity, pool.UtilizedCapacity,
		pool.UnitOfMeasure, pool.WarningThresholdPercent, pool.CriticalThresholdPercent,
		pool.Status, pool.Description, pool.UpdatedAt,
		pool.ID,
	)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return ErrCapacityPoolNameConflict
		}
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrCapacityPoolNotFound
	}
	return nil
}

func (r *sqliteCapacityRepository) DeletePool(ctx context.Context, id string) error {
	res, err := r.db.ExecContext(ctx, "DELETE FROM capacity_pools WHERE id = ?", id)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrCapacityPoolNotFound
	}
	return nil
}

func (r *sqliteCapacityRepository) ListPools(ctx context.Context, filter model.CapacityPoolFilterCriteria) ([]model.CapacityPool, int, error) {
	whereClauses := []string{"1=1"}
	args := []interface{}{}

	if filter.WorkspaceID != "" {
		whereClauses = append(whereClauses, "workspace_id = ?")
		args = append(args, filter.WorkspaceID)
	}
	if filter.ResourceType != "" {
		whereClauses = append(whereClauses, "resource_type = ?")
		args = append(args, filter.ResourceType)
	}
	if filter.Status != "" {
		whereClauses = append(whereClauses, "status = ?")
		args = append(args, filter.Status)
	}
	if filter.CIID != "" {
		whereClauses = append(whereClauses, "ci_id = ?")
		args = append(args, filter.CIID)
	}
	if filter.Search != "" {
		whereClauses = append(whereClauses, "(name LIKE ? OR description LIKE ?)")
		s := "%" + filter.Search + "%"
		args = append(args, s, s)
	}

	whereSQL := strings.Join(whereClauses, " AND ")

	var total int
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM capacity_pools WHERE %s", whereSQL)
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
	SELECT id, workspace_id, name, resource_type, ci_id,
	       total_capacity, allocated_capacity, utilized_capacity,
	       unit_of_measure, warning_threshold_percent, critical_threshold_percent,
	       status, description, created_at, updated_at
	FROM capacity_pools WHERE %s ORDER BY name ASC LIMIT ? OFFSET ?`, whereSQL)

	queryArgs := append(args, limit, offset)
	rows, err := r.db.QueryContext(ctx, query, queryArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	list := []model.CapacityPool{}
	for rows.Next() {
		var p model.CapacityPool
		if err := rows.Scan(
			&p.ID, &p.WorkspaceID, &p.Name, &p.ResourceType, &p.CIID,
			&p.TotalCapacity, &p.AllocatedCapacity, &p.UtilizedCapacity,
			&p.UnitOfMeasure, &p.WarningThresholdPercent, &p.CriticalThresholdPercent,
			&p.Status, &p.Description, &p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		list = append(list, p)
	}
	return list, total, rows.Err()
}

// Metric Stubs
func (r *sqliteCapacityRepository) RecordMetric(ctx context.Context, metric *model.CapacityMetric) error {
	return nil
}
func (r *sqliteCapacityRepository) GetMetricByID(ctx context.Context, id string) (*model.CapacityMetric, error) {
	return nil, ErrCapacityMetricNotFound
}
func (r *sqliteCapacityRepository) ListMetrics(ctx context.Context, filter model.CapacityMetricFilterCriteria) ([]model.CapacityMetric, int, error) {
	return nil, 0, nil
}

// Forecast Stubs
func (r *sqliteCapacityRepository) SaveForecast(ctx context.Context, forecast *model.CapacityForecast) error {
	return nil
}
func (r *sqliteCapacityRepository) GetLatestForecastByPoolID(ctx context.Context, poolID string) (*model.CapacityForecast, error) {
	return nil, nil
}

// Plan Stubs
func (r *sqliteCapacityRepository) CreatePlan(ctx context.Context, plan *model.CapacityPlan) error {
	return nil
}
func (r *sqliteCapacityRepository) GetPlanByID(ctx context.Context, id string) (*model.CapacityPlan, error) {
	return nil, ErrCapacityPlanNotFound
}
func (r *sqliteCapacityRepository) UpdatePlan(ctx context.Context, plan *model.CapacityPlan) error {
	return nil
}
func (r *sqliteCapacityRepository) DeletePlan(ctx context.Context, id string) error {
	return nil
}
func (r *sqliteCapacityRepository) ListPlans(ctx context.Context, filter model.CapacityPlanFilterCriteria) ([]model.CapacityPlan, int, error) {
	return nil, 0, nil
}

// Bottleneck Stubs
func (r *sqliteCapacityRepository) CreateBottleneck(ctx context.Context, b *model.BottleneckRecord) error {
	return nil
}
func (r *sqliteCapacityRepository) GetBottleneckByID(ctx context.Context, id string) (*model.BottleneckRecord, error) {
	return nil, ErrBottleneckNotFound
}
func (r *sqliteCapacityRepository) UpdateBottleneck(ctx context.Context, b *model.BottleneckRecord) error {
	return nil
}
func (r *sqliteCapacityRepository) DeleteBottleneck(ctx context.Context, id string) error {
	return nil
}
func (r *sqliteCapacityRepository) ListBottlenecks(ctx context.Context, filter model.BottleneckFilterCriteria) ([]model.BottleneckRecord, int, error) {
	return nil, 0, nil
}

// Analytics Stub
func (r *sqliteCapacityRepository) GetCapacityAnalytics(ctx context.Context, workspaceID string) (*model.CapacityAnalyticsReport, error) {
	return &model.CapacityAnalyticsReport{WorkspaceID: workspaceID}, nil
}
