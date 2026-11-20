package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"
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

// Metrics & Telemetry
func (r *sqliteCapacityRepository) RecordMetric(ctx context.Context, metric *model.CapacityMetric) error {
	if metric.ID == "" {
		metric.ID = uuid.New().String()
	}
	now := time.Now().UTC()
	if metric.CreatedAt.IsZero() {
		metric.CreatedAt = now
	}
	if metric.SampledAt.IsZero() {
		metric.SampledAt = now
	}
	if metric.PeakValue < metric.AverageValue {
		metric.PeakValue = metric.AverageValue
	}

	isAnomalyInt := 0
	if metric.IsAnomaly {
		isAnomalyInt = 1
	}

	query := `
	INSERT INTO capacity_metrics (
		id, workspace_id, pool_id, metric_name, sampled_at,
		average_value, peak_value, p95_value, p99_value,
		is_anomaly, anomaly_score, created_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := r.db.ExecContext(ctx, query,
		metric.ID, metric.WorkspaceID, metric.PoolID, metric.MetricName, metric.SampledAt,
		metric.AverageValue, metric.PeakValue, metric.P95Value, metric.P99Value,
		isAnomalyInt, metric.AnomalyScore, metric.CreatedAt,
	)
	return err
}

func (r *sqliteCapacityRepository) GetMetricByID(ctx context.Context, id string) (*model.CapacityMetric, error) {
	query := `
	SELECT id, workspace_id, pool_id, metric_name, sampled_at,
	       average_value, peak_value, p95_value, p99_value,
	       is_anomaly, anomaly_score, created_at
	FROM capacity_metrics WHERE id = ?`

	var m model.CapacityMetric
	var isAnomalyInt int
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&m.ID, &m.WorkspaceID, &m.PoolID, &m.MetricName, &m.SampledAt,
		&m.AverageValue, &m.PeakValue, &m.P95Value, &m.P99Value,
		&isAnomalyInt, &m.AnomalyScore, &m.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrCapacityMetricNotFound
		}
		return nil, err
	}
	m.IsAnomaly = (isAnomalyInt == 1)
	return &m, nil
}

func (r *sqliteCapacityRepository) ListMetrics(ctx context.Context, filter model.CapacityMetricFilterCriteria) ([]model.CapacityMetric, int, error) {
	whereClauses := []string{"1=1"}
	args := []interface{}{}

	if filter.WorkspaceID != "" {
		whereClauses = append(whereClauses, "workspace_id = ?")
		args = append(args, filter.WorkspaceID)
	}
	if filter.PoolID != "" {
		whereClauses = append(whereClauses, "pool_id = ?")
		args = append(args, filter.PoolID)
	}
	if filter.MetricName != "" {
		whereClauses = append(whereClauses, "metric_name = ?")
		args = append(args, filter.MetricName)
	}
	if filter.IsAnomaly != nil {
		whereClauses = append(whereClauses, "is_anomaly = ?")
		val := 0
		if *filter.IsAnomaly {
			val = 1
		}
		args = append(args, val)
	}
	if filter.FromDate != nil {
		whereClauses = append(whereClauses, "sampled_at >= ?")
		args = append(args, *filter.FromDate)
	}
	if filter.ToDate != nil {
		whereClauses = append(whereClauses, "sampled_at <= ?")
		args = append(args, *filter.ToDate)
	}

	whereSQL := strings.Join(whereClauses, " AND ")

	var total int
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM capacity_metrics WHERE %s", whereSQL)
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
	SELECT id, workspace_id, pool_id, metric_name, sampled_at,
	       average_value, peak_value, p95_value, p99_value,
	       is_anomaly, anomaly_score, created_at
	FROM capacity_metrics WHERE %s ORDER BY sampled_at DESC LIMIT ? OFFSET ?`, whereSQL)

	queryArgs := append(args, limit, offset)
	rows, err := r.db.QueryContext(ctx, query, queryArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	list := []model.CapacityMetric{}
	for rows.Next() {
		var m model.CapacityMetric
		var isAnomalyInt int
		if err := rows.Scan(
			&m.ID, &m.WorkspaceID, &m.PoolID, &m.MetricName, &m.SampledAt,
			&m.AverageValue, &m.PeakValue, &m.P95Value, &m.P99Value,
			&isAnomalyInt, &m.AnomalyScore, &m.CreatedAt,
		); err != nil {
			return nil, 0, err
		}
		m.IsAnomaly = (isAnomalyInt == 1)
		list = append(list, m)
	}
	return list, total, rows.Err()
}

// Forecasts
func (r *sqliteCapacityRepository) SaveForecast(ctx context.Context, forecast *model.CapacityForecast) error {
	if forecast.ID == "" {
		forecast.ID = uuid.New().String()
	}
	now := time.Now().UTC()
	if forecast.CreatedAt.IsZero() {
		forecast.CreatedAt = now
	}
	forecast.UpdatedAt = now
	if forecast.RecommendedAction == "" {
		forecast.RecommendedAction = model.ActionScaleUp
	}

	query := `
	INSERT OR REPLACE INTO capacity_forecasts (
		id, workspace_id, pool_id, forecast_horizon_days,
		current_utilization_percent, projected_utilization_percent,
		growth_rate_monthly_percent, days_to_exhaustion, confidence_score,
		recommended_action, recommendation_notes, created_at, updated_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := r.db.ExecContext(ctx, query,
		forecast.ID, forecast.WorkspaceID, forecast.PoolID, forecast.ForecastHorizonDays,
		forecast.CurrentUtilizationPercent, forecast.ProjectedUtilizationPercent,
		forecast.GrowthRateMonthlyPercent, forecast.DaysToExhaustion, forecast.ConfidenceScore,
		forecast.RecommendedAction, forecast.RecommendationNotes, forecast.CreatedAt, forecast.UpdatedAt,
	)
	return err
}

func (r *sqliteCapacityRepository) GetLatestForecastByPoolID(ctx context.Context, poolID string) (*model.CapacityForecast, error) {
	query := `
	SELECT id, workspace_id, pool_id, forecast_horizon_days,
	       current_utilization_percent, projected_utilization_percent,
	       growth_rate_monthly_percent, days_to_exhaustion, confidence_score,
	       recommended_action, recommendation_notes, created_at, updated_at
	FROM capacity_forecasts WHERE pool_id = ? ORDER BY created_at DESC LIMIT 1`

	var f model.CapacityForecast
	err := r.db.QueryRowContext(ctx, query, poolID).Scan(
		&f.ID, &f.WorkspaceID, &f.PoolID, &f.ForecastHorizonDays,
		&f.CurrentUtilizationPercent, &f.ProjectedUtilizationPercent,
		&f.GrowthRateMonthlyPercent, &f.DaysToExhaustion, &f.ConfidenceScore,
		&f.RecommendedAction, &f.RecommendationNotes, &f.CreatedAt, &f.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &f, nil
}

// Capacity Plans
func (r *sqliteCapacityRepository) CreatePlan(ctx context.Context, plan *model.CapacityPlan) error {
	if plan.ID == "" {
		plan.ID = uuid.New().String()
	}
	now := time.Now().UTC()
	if plan.CreatedAt.IsZero() {
		plan.CreatedAt = now
	}
	plan.UpdatedAt = now
	if plan.PlanningHorizon == "" {
		plan.PlanningHorizon = model.PlanHorizonShortTerm
	}
	if plan.Status == "" {
		plan.Status = model.PlanStatusDraft
	}
	if plan.Currency == "" {
		plan.Currency = "USD"
	}

	query := `
	INSERT INTO capacity_plans (
		id, workspace_id, pool_id, title, planning_horizon,
		current_capacity, required_capacity, estimated_cost,
		currency, justification, status, target_date, approved_by,
		created_at, updated_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := r.db.ExecContext(ctx, query,
		plan.ID, plan.WorkspaceID, plan.PoolID, plan.Title, plan.PlanningHorizon,
		plan.CurrentCapacity, plan.RequiredCapacity, plan.EstimatedCost,
		plan.Currency, plan.Justification, plan.Status, plan.TargetDate, plan.ApprovedBy,
		plan.CreatedAt, plan.UpdatedAt,
	)
	return err
}

func (r *sqliteCapacityRepository) GetPlanByID(ctx context.Context, id string) (*model.CapacityPlan, error) {
	query := `
	SELECT id, workspace_id, pool_id, title, planning_horizon,
	       current_capacity, required_capacity, estimated_cost,
	       currency, justification, status, target_date, approved_by,
	       created_at, updated_at
	FROM capacity_plans WHERE id = ?`

	var plan model.CapacityPlan
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&plan.ID, &plan.WorkspaceID, &plan.PoolID, &plan.Title, &plan.PlanningHorizon,
		&plan.CurrentCapacity, &plan.RequiredCapacity, &plan.EstimatedCost,
		&plan.Currency, &plan.Justification, &plan.Status, &plan.TargetDate, &plan.ApprovedBy,
		&plan.CreatedAt, &plan.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrCapacityPlanNotFound
		}
		return nil, err
	}
	return &plan, nil
}

func (r *sqliteCapacityRepository) UpdatePlan(ctx context.Context, plan *model.CapacityPlan) error {
	plan.UpdatedAt = time.Now().UTC()

	query := `
	UPDATE capacity_plans SET
		title = ?, planning_horizon = ?, required_capacity = ?,
		estimated_cost = ?, currency = ?, justification = ?,
		status = ?, target_date = ?, approved_by = ?, updated_at = ?
	WHERE id = ?`

	res, err := r.db.ExecContext(ctx, query,
		plan.Title, plan.PlanningHorizon, plan.RequiredCapacity,
		plan.EstimatedCost, plan.Currency, plan.Justification,
		plan.Status, plan.TargetDate, plan.ApprovedBy, plan.UpdatedAt,
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
		return ErrCapacityPlanNotFound
	}
	return nil
}

func (r *sqliteCapacityRepository) DeletePlan(ctx context.Context, id string) error {
	res, err := r.db.ExecContext(ctx, "DELETE FROM capacity_plans WHERE id = ?", id)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrCapacityPlanNotFound
	}
	return nil
}

func (r *sqliteCapacityRepository) ListPlans(ctx context.Context, filter model.CapacityPlanFilterCriteria) ([]model.CapacityPlan, int, error) {
	whereClauses := []string{"1=1"}
	args := []interface{}{}

	if filter.WorkspaceID != "" {
		whereClauses = append(whereClauses, "workspace_id = ?")
		args = append(args, filter.WorkspaceID)
	}
	if filter.PoolID != "" {
		whereClauses = append(whereClauses, "pool_id = ?")
		args = append(args, filter.PoolID)
	}
	if filter.PlanningHorizon != "" {
		whereClauses = append(whereClauses, "planning_horizon = ?")
		args = append(args, filter.PlanningHorizon)
	}
	if filter.Status != "" {
		whereClauses = append(whereClauses, "status = ?")
		args = append(args, filter.Status)
	}

	whereSQL := strings.Join(whereClauses, " AND ")

	var total int
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM capacity_plans WHERE %s", whereSQL)
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
	SELECT id, workspace_id, pool_id, title, planning_horizon,
	       current_capacity, required_capacity, estimated_cost,
	       currency, justification, status, target_date, approved_by,
	       created_at, updated_at
	FROM capacity_plans WHERE %s ORDER BY created_at DESC LIMIT ? OFFSET ?`, whereSQL)

	queryArgs := append(args, limit, offset)
	rows, err := r.db.QueryContext(ctx, query, queryArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	list := []model.CapacityPlan{}
	for rows.Next() {
		var plan model.CapacityPlan
		if err := rows.Scan(
			&plan.ID, &plan.WorkspaceID, &plan.PoolID, &plan.Title, &plan.PlanningHorizon,
			&plan.CurrentCapacity, &plan.RequiredCapacity, &plan.EstimatedCost,
			&plan.Currency, &plan.Justification, &plan.Status, &plan.TargetDate, &plan.ApprovedBy,
			&plan.CreatedAt, &plan.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		list = append(list, plan)
	}
	return list, total, rows.Err()
}

// Bottlenecks
func (r *sqliteCapacityRepository) CreateBottleneck(ctx context.Context, b *model.BottleneckRecord) error {
	if b.ID == "" {
		b.ID = uuid.New().String()
	}
	now := time.Now().UTC()
	if b.CreatedAt.IsZero() {
		b.CreatedAt = now
	}
	b.UpdatedAt = now
	if b.DetectedAt.IsZero() {
		b.DetectedAt = now
	}
	if b.Severity == "" {
		b.Severity = model.BottleneckSeverityMedium
	}
	if b.Status == "" {
		b.Status = model.BottleneckStatusOpen
	}

	query := `
	INSERT INTO capacity_bottlenecks (
		id, workspace_id, pool_id, affected_ci_id, title,
		description, severity, saturation_percent, status,
		mitigation_plan, detected_at, resolved_at, created_at, updated_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := r.db.ExecContext(ctx, query,
		b.ID, b.WorkspaceID, b.PoolID, b.AffectedCIID, b.Title,
		b.Description, b.Severity, b.SaturationPercent, b.Status,
		b.MitigationPlan, b.DetectedAt, b.ResolvedAt, b.CreatedAt, b.UpdatedAt,
	)
	return err
}

func (r *sqliteCapacityRepository) GetBottleneckByID(ctx context.Context, id string) (*model.BottleneckRecord, error) {
	query := `
	SELECT id, workspace_id, pool_id, affected_ci_id, title,
	       description, severity, saturation_percent, status,
	       mitigation_plan, detected_at, resolved_at, created_at, updated_at
	FROM capacity_bottlenecks WHERE id = ?`

	var b model.BottleneckRecord
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&b.ID, &b.WorkspaceID, &b.PoolID, &b.AffectedCIID, &b.Title,
		&b.Description, &b.Severity, &b.SaturationPercent, &b.Status,
		&b.MitigationPlan, &b.DetectedAt, &b.ResolvedAt, &b.CreatedAt, &b.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrBottleneckNotFound
		}
		return nil, err
	}
	return &b, nil
}

func (r *sqliteCapacityRepository) UpdateBottleneck(ctx context.Context, b *model.BottleneckRecord) error {
	b.UpdatedAt = time.Now().UTC()

	query := `
	UPDATE capacity_bottlenecks SET
		title = ?, description = ?, severity = ?,
		saturation_percent = ?, status = ?, mitigation_plan = ?,
		resolved_at = ?, updated_at = ?
	WHERE id = ?`

	res, err := r.db.ExecContext(ctx, query,
		b.Title, b.Description, b.Severity,
		b.SaturationPercent, b.Status, b.MitigationPlan,
		b.ResolvedAt, b.UpdatedAt, b.ID,
	)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrBottleneckNotFound
	}
	return nil
}

func (r *sqliteCapacityRepository) DeleteBottleneck(ctx context.Context, id string) error {
	res, err := r.db.ExecContext(ctx, "DELETE FROM capacity_bottlenecks WHERE id = ?", id)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrBottleneckNotFound
	}
	return nil
}

func (r *sqliteCapacityRepository) ListBottlenecks(ctx context.Context, filter model.BottleneckFilterCriteria) ([]model.BottleneckRecord, int, error) {
	whereClauses := []string{"1=1"}
	args := []interface{}{}

	if filter.WorkspaceID != "" {
		whereClauses = append(whereClauses, "workspace_id = ?")
		args = append(args, filter.WorkspaceID)
	}
	if filter.PoolID != "" {
		whereClauses = append(whereClauses, "pool_id = ?")
		args = append(args, filter.PoolID)
	}
	if filter.AffectedCIID != "" {
		whereClauses = append(whereClauses, "affected_ci_id = ?")
		args = append(args, filter.AffectedCIID)
	}
	if filter.Severity != "" {
		whereClauses = append(whereClauses, "severity = ?")
		args = append(args, filter.Severity)
	}
	if filter.Status != "" {
		whereClauses = append(whereClauses, "status = ?")
		args = append(args, filter.Status)
	}

	whereSQL := strings.Join(whereClauses, " AND ")

	var total int
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM capacity_bottlenecks WHERE %s", whereSQL)
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
	SELECT id, workspace_id, pool_id, affected_ci_id, title,
	       description, severity, saturation_percent, status,
	       mitigation_plan, detected_at, resolved_at, created_at, updated_at
	FROM capacity_bottlenecks WHERE %s ORDER BY detected_at DESC LIMIT ? OFFSET ?`, whereSQL)

	queryArgs := append(args, limit, offset)
	rows, err := r.db.QueryContext(ctx, query, queryArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	list := []model.BottleneckRecord{}
	for rows.Next() {
		var b model.BottleneckRecord
		if err := rows.Scan(
			&b.ID, &b.WorkspaceID, &b.PoolID, &b.AffectedCIID, &b.Title,
			&b.Description, &b.Severity, &b.SaturationPercent, &b.Status,
			&b.MitigationPlan, &b.DetectedAt, &b.ResolvedAt, &b.CreatedAt, &b.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		list = append(list, b)
	}
	return list, total, rows.Err()
}

// Posture Analytics
func (r *sqliteCapacityRepository) GetCapacityAnalytics(ctx context.Context, workspaceID string) (*model.CapacityAnalyticsReport, error) {
	report := &model.CapacityAnalyticsReport{
		WorkspaceID:               workspaceID,
		UtilizationByResourceType: make(map[string]float64),
		TopSaturatedPools:         make([]model.CapacityPool, 0),
	}

	// 1. Pool Status Counts
	poolCountsQuery := `
	SELECT
		COUNT(*),
		COALESCE(SUM(CASE WHEN status = 'optimal' THEN 1 ELSE 0 END), 0),
		COALESCE(SUM(CASE WHEN status = 'warning' THEN 1 ELSE 0 END), 0),
		COALESCE(SUM(CASE WHEN status = 'critical' THEN 1 ELSE 0 END), 0),
		COALESCE(SUM(CASE WHEN status = 'exhausted' THEN 1 ELSE 0 END), 0)
	FROM capacity_pools WHERE workspace_id = ?`

	err := r.db.QueryRowContext(ctx, poolCountsQuery, workspaceID).Scan(
		&report.TotalPools,
		&report.OptimalPoolsCount,
		&report.WarningPoolsCount,
		&report.CriticalPoolsCount,
		&report.ExhaustedPoolsCount,
	)
	if err != nil {
		return nil, err
	}

	// 2. Average Utilization & Per-Resource-Type Utilization
	rows, err := r.db.QueryContext(ctx, `
		SELECT resource_type, COALESCE(SUM(total_capacity), 0.0), COALESCE(SUM(utilized_capacity), 0.0)
		FROM capacity_pools WHERE workspace_id = ?
		GROUP BY resource_type`, workspaceID,
	)
	if err == nil {
		defer rows.Close()
		var globalTotal, globalUtilized float64
		for rows.Next() {
			var rType string
			var tot, util float64
			if err := rows.Scan(&rType, &tot, &util); err == nil {
				globalTotal += tot
				globalUtilized += util
				if tot > 0 {
					report.UtilizationByResourceType[rType] = math.Round((util/tot)*10000) / 100
				}
			}
		}
		if globalTotal > 0 {
			report.AverageUtilizationPercent = math.Round((globalUtilized/globalTotal)*10000) / 100
		}
	}

	// 3. Top Saturated Pools
	topRows, err := r.db.QueryContext(ctx, `
		SELECT id, workspace_id, name, resource_type, ci_id,
		       total_capacity, allocated_capacity, utilized_capacity,
		       unit_of_measure, warning_threshold_percent, critical_threshold_percent,
		       status, description, created_at, updated_at
		FROM capacity_pools WHERE workspace_id = ?
		ORDER BY (utilized_capacity / total_capacity) DESC LIMIT 5`, workspaceID,
	)
	if err == nil {
		defer topRows.Close()
		for topRows.Next() {
			var p model.CapacityPool
			if err := topRows.Scan(
				&p.ID, &p.WorkspaceID, &p.Name, &p.ResourceType, &p.CIID,
				&p.TotalCapacity, &p.AllocatedCapacity, &p.UtilizedCapacity,
				&p.UnitOfMeasure, &p.WarningThresholdPercent, &p.CriticalThresholdPercent,
				&p.Status, &p.Description, &p.CreatedAt, &p.UpdatedAt,
			); err == nil {
				report.TopSaturatedPools = append(report.TopSaturatedPools, p)
			}
		}
	}

	// 4. Anomalies
	_ = r.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM capacity_metrics WHERE workspace_id = ? AND is_anomaly = 1`,
		workspaceID,
	).Scan(&report.TotalAnomaliesDetected)

	// 5. Bottlenecks
	_ = r.db.QueryRowContext(ctx, `
		SELECT
			COALESCE(SUM(CASE WHEN status IN ('open', 'investigating') THEN 1 ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN status IN ('open', 'investigating') AND severity = 'critical' THEN 1 ELSE 0 END), 0)
		FROM capacity_bottlenecks WHERE workspace_id = ?`,
		workspaceID,
	).Scan(&report.OpenBottlenecksCount, &report.CriticalBottlenecksCount)

	// 6. Active Capacity Plans & Cost
	_ = r.db.QueryRowContext(ctx, `
		SELECT
			COUNT(*),
			COALESCE(SUM(estimated_cost), 0.0)
		FROM capacity_plans WHERE workspace_id = ? AND status IN ('draft', 'in_review', 'approved')`,
		workspaceID,
	).Scan(&report.TotalPlansActive, &report.TotalEstimatedPlanCost)

	// 7. Capacity Health Index (0-100)
	score := 100.0
	score -= float64(report.ExhaustedPoolsCount) * 25.0
	score -= float64(report.CriticalPoolsCount) * 15.0
	score -= float64(report.WarningPoolsCount) * 5.0
	score -= float64(report.CriticalBottlenecksCount) * 10.0
	score -= float64(report.OpenBottlenecksCount) * 2.0
	score -= float64(report.TotalAnomaliesDetected) * 1.0

	if score < 0.0 {
		score = 0.0
	}
	if score > 100.0 {
		score = 100.0
	}
	report.CapacityHealthIndex = math.Round(score*10) / 10

	return report, nil
}
