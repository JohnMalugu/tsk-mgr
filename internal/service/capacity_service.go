package service

import (
	"context"
	"strings"
	"time"

	"github.com/JohnMalugu/tsk-mgr-api/internal/model"
	"github.com/JohnMalugu/tsk-mgr-api/internal/repository"
)

var (
	_ = strings.TrimSpace
	_ = time.Now
)

var (
	ErrCapacityPoolNotFound      = repository.ErrCapacityPoolNotFound
	ErrCapacityPoolNameRequired  = repository.ErrCapacityPoolNameRequired
	ErrCapacityPoolNameConflict  = repository.ErrCapacityPoolNameConflict
	ErrInvalidResourceType       = repository.ErrInvalidResourceType
	ErrInvalidPoolStatus         = repository.ErrInvalidPoolStatus
	ErrInvalidTotalCapacity      = repository.ErrInvalidTotalCapacity
	ErrInvalidThresholdPercent   = repository.ErrInvalidThresholdPercent
	ErrCapacityMetricNotFound    = repository.ErrCapacityMetricNotFound
	ErrCapacityPlanNotFound      = repository.ErrCapacityPlanNotFound
	ErrInvalidPlanHorizon        = repository.ErrInvalidPlanHorizon
	ErrInvalidPlanStatus         = repository.ErrInvalidPlanStatus
	ErrBottleneckNotFound        = repository.ErrBottleneckNotFound
	ErrInvalidBottleneckSeverity = repository.ErrInvalidBottleneckSeverity
	ErrInvalidBottleneckStatus   = repository.ErrInvalidBottleneckStatus
)

// CapacityService defines business logic for ITIL Capacity and Performance Management
type CapacityService interface {
	// Pools
	CreatePool(ctx context.Context, workspaceID string, req model.CreateCapacityPoolRequest) (*model.CapacityPool, error)
	GetPoolByID(ctx context.Context, id string) (*model.CapacityPool, error)
	GetPoolByName(ctx context.Context, workspaceID string, name string) (*model.CapacityPool, error)
	UpdatePool(ctx context.Context, id string, req model.UpdateCapacityPoolRequest) (*model.CapacityPool, error)
	DeletePool(ctx context.Context, id string) error
	ListPools(ctx context.Context, filter model.CapacityPoolFilterCriteria) ([]model.CapacityPool, int, error)

	// Metrics & Telemetry
	RecordMetric(ctx context.Context, workspaceID string, req model.RecordCapacityMetricRequest) (*model.CapacityMetric, error)
	GetMetricByID(ctx context.Context, id string) (*model.CapacityMetric, error)
	ListMetrics(ctx context.Context, filter model.CapacityMetricFilterCriteria) ([]model.CapacityMetric, int, error)

	// Forecasting
	GenerateForecast(ctx context.Context, workspaceID string, poolID string, horizonDays int) (*model.CapacityForecast, error)
	GetLatestForecast(ctx context.Context, poolID string) (*model.CapacityForecast, error)

	// Capacity Plans
	CreatePlan(ctx context.Context, workspaceID string, req model.CreateCapacityPlanRequest) (*model.CapacityPlan, error)
	GetPlanByID(ctx context.Context, id string) (*model.CapacityPlan, error)
	UpdatePlan(ctx context.Context, id string, req model.UpdateCapacityPlanRequest) (*model.CapacityPlan, error)
	ApprovePlan(ctx context.Context, id string, approverID string) (*model.CapacityPlan, error)
	DeletePlan(ctx context.Context, id string) error
	ListPlans(ctx context.Context, filter model.CapacityPlanFilterCriteria) ([]model.CapacityPlan, int, error)

	// Bottlenecks
	CreateBottleneck(ctx context.Context, workspaceID string, req model.CreateBottleneckRequest) (*model.BottleneckRecord, error)
	GetBottleneckByID(ctx context.Context, id string) (*model.BottleneckRecord, error)
	UpdateBottleneck(ctx context.Context, id string, req model.UpdateBottleneckRequest) (*model.BottleneckRecord, error)
	ResolveBottleneck(ctx context.Context, id string, mitigationPlan string) (*model.BottleneckRecord, error)
	DeleteBottleneck(ctx context.Context, id string) error
	ListBottlenecks(ctx context.Context, filter model.BottleneckFilterCriteria) ([]model.BottleneckRecord, int, error)

	// Posture Analytics
	GetCapacityAnalytics(ctx context.Context, workspaceID string) (*model.CapacityAnalyticsReport, error)
}

type capacityServiceImpl struct {
	repo repository.CapacityRepository
}

// NewCapacityService creates a new CapacityService instance
func NewCapacityService(repo repository.CapacityRepository) CapacityService {
	return &capacityServiceImpl{repo: repo}
}

// Pools
func (s *capacityServiceImpl) CreatePool(ctx context.Context, workspaceID string, req model.CreateCapacityPoolRequest) (*model.CapacityPool, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	warnThresh := req.WarningThresholdPercent
	if warnThresh <= 0 {
		warnThresh = 75.0
	}
	critThresh := req.CriticalThresholdPercent
	if critThresh <= 0 {
		critThresh = 90.0
	}
	uom := req.UnitOfMeasure
	if uom == "" {
		switch req.ResourceType {
		case model.ResourceTypeComputeCPU:
			uom = model.UnitCores
		case model.ResourceTypeMemoryRAM:
			uom = model.UnitGB
		case model.ResourceTypeStorageDisk:
			uom = model.UnitGB
		case model.ResourceTypeNetworkBandwidth:
			uom = model.UnitMbps
		case model.ResourceTypeDatabaseIOPS:
			uom = model.UnitIOPS
		case model.ResourceTypeLicenseSeats:
			uom = model.UnitSeats
		default:
			uom = "units"
		}
	}

	pool := &model.CapacityPool{
		WorkspaceID:              workspaceID,
		Name:                     strings.TrimSpace(req.Name),
		ResourceType:             req.ResourceType,
		CIID:                     req.CIID,
		TotalCapacity:            req.TotalCapacity,
		AllocatedCapacity:        req.AllocatedCapacity,
		UtilizedCapacity:         req.UtilizedCapacity,
		UnitOfMeasure:            uom,
		WarningThresholdPercent:  warnThresh,
		CriticalThresholdPercent: critThresh,
		Description:              req.Description,
	}
	pool.RecalculateStatus()

	if err := s.repo.CreatePool(ctx, pool); err != nil {
		return nil, err
	}
	return pool, nil
}

func (s *capacityServiceImpl) GetPoolByID(ctx context.Context, id string) (*model.CapacityPool, error) {
	return s.repo.GetPoolByID(ctx, id)
}

func (s *capacityServiceImpl) GetPoolByName(ctx context.Context, workspaceID string, name string) (*model.CapacityPool, error) {
	return s.repo.GetPoolByName(ctx, workspaceID, name)
}

func (s *capacityServiceImpl) UpdatePool(ctx context.Context, id string, req model.UpdateCapacityPoolRequest) (*model.CapacityPool, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	pool, err := s.repo.GetPoolByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.Name != nil {
		pool.Name = strings.TrimSpace(*req.Name)
	}
	if req.TotalCapacity != nil {
		pool.TotalCapacity = *req.TotalCapacity
	}
	if req.AllocatedCapacity != nil {
		pool.AllocatedCapacity = *req.AllocatedCapacity
	}
	if req.UtilizedCapacity != nil {
		pool.UtilizedCapacity = *req.UtilizedCapacity
	}
	if req.WarningThresholdPercent != nil {
		pool.WarningThresholdPercent = *req.WarningThresholdPercent
	}
	if req.CriticalThresholdPercent != nil {
		pool.CriticalThresholdPercent = *req.CriticalThresholdPercent
	}
	if req.Description != nil {
		pool.Description = *req.Description
	}
	pool.RecalculateStatus()

	if err := s.repo.UpdatePool(ctx, pool); err != nil {
		return nil, err
	}
	return pool, nil
}

func (s *capacityServiceImpl) DeletePool(ctx context.Context, id string) error {
	return s.repo.DeletePool(ctx, id)
}

func (s *capacityServiceImpl) ListPools(ctx context.Context, filter model.CapacityPoolFilterCriteria) ([]model.CapacityPool, int, error) {
	return s.repo.ListPools(ctx, filter)
}

// Metrics & Telemetry
func (s *capacityServiceImpl) RecordMetric(ctx context.Context, workspaceID string, req model.RecordCapacityMetricRequest) (*model.CapacityMetric, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	pool, err := s.repo.GetPoolByID(ctx, req.PoolID)
	if err != nil {
		return nil, err
	}

	sampledAt := time.Now().UTC()
	if req.SampledAt != nil && !req.SampledAt.IsZero() {
		sampledAt = *req.SampledAt
	}
	peak := req.PeakValue
	if peak < req.AverageValue {
		peak = req.AverageValue
	}

	metric := &model.CapacityMetric{
		WorkspaceID:  workspaceID,
		PoolID:       req.PoolID,
		MetricName:   strings.TrimSpace(req.MetricName),
		SampledAt:    sampledAt,
		AverageValue: req.AverageValue,
		PeakValue:    peak,
		P95Value:     req.P95Value,
		P99Value:     req.P99Value,
		IsAnomaly:    req.IsAnomaly,
		AnomalyScore: req.AnomalyScore,
	}

	if err := s.repo.RecordMetric(ctx, metric); err != nil {
		return nil, err
	}

	// If metric measures pool resource utilization, update pool's utilized capacity
	if strings.Contains(metric.MetricName, "utilization") || strings.Contains(metric.MetricName, "used") {
		if req.AverageValue <= pool.TotalCapacity {
			pool.UtilizedCapacity = req.AverageValue
			pool.RecalculateStatus()
			_ = s.repo.UpdatePool(ctx, pool)
		}
	}

	return metric, nil
}

func (s *capacityServiceImpl) GetMetricByID(ctx context.Context, id string) (*model.CapacityMetric, error) {
	return s.repo.GetMetricByID(ctx, id)
}

func (s *capacityServiceImpl) ListMetrics(ctx context.Context, filter model.CapacityMetricFilterCriteria) ([]model.CapacityMetric, int, error) {
	return s.repo.ListMetrics(ctx, filter)
}

// Forecasting
func (s *capacityServiceImpl) GenerateForecast(ctx context.Context, workspaceID string, poolID string, horizonDays int) (*model.CapacityForecast, error) {
	pool, err := s.repo.GetPoolByID(ctx, poolID)
	if err != nil {
		return nil, err
	}

	if horizonDays <= 0 {
		horizonDays = 90
	}

	currentUtil := pool.UtilizationPercent()
	growthRate := 5.0 // baseline 5% monthly growth
	if currentUtil > 80.0 {
		growthRate = 8.5
	}

	months := float64(horizonDays) / 30.0
	projectedUtil := currentUtil + (growthRate * months)

	daysToExhaustion := -1
	if currentUtil < 100.0 && growthRate > 0.0 {
		dailyGrowth := growthRate / 30.0
		daysToExhaustion = int((100.0 - currentUtil) / dailyGrowth)
	}

	recAction := model.ActionOptimize
	recNotes := "Current capacity is sufficient under projected growth rate"
	if daysToExhaustion > 0 && daysToExhaustion <= 30 {
		recAction = model.ActionScaleUp
		recNotes = "Immediate scale-up required: exhaustion projected within 30 days"
	} else if daysToExhaustion > 30 && daysToExhaustion <= 90 {
		recAction = model.ActionScaleOut
		recNotes = "Plan horizontal scale-out expansion within current quarter"
	} else if currentUtil < 20.0 {
		recAction = model.ActionScaleDown
		recNotes = "Pool is substantially under-utilized; consider right-sizing down"
	}

	forecast := &model.CapacityForecast{
		WorkspaceID:                 workspaceID,
		PoolID:                      poolID,
		ForecastHorizonDays:         horizonDays,
		CurrentUtilizationPercent:   currentUtil,
		ProjectedUtilizationPercent: projectedUtil,
		GrowthRateMonthlyPercent:    growthRate,
		DaysToExhaustion:            daysToExhaustion,
		ConfidenceScore:             0.90,
		RecommendedAction:           recAction,
		RecommendationNotes:         recNotes,
	}

	if err := s.repo.SaveForecast(ctx, forecast); err != nil {
		return nil, err
	}
	return forecast, nil
}

func (s *capacityServiceImpl) GetLatestForecast(ctx context.Context, poolID string) (*model.CapacityForecast, error) {
	return s.repo.GetLatestForecastByPoolID(ctx, poolID)
}

// Plan Stubs
func (s *capacityServiceImpl) CreatePlan(ctx context.Context, workspaceID string, req model.CreateCapacityPlanRequest) (*model.CapacityPlan, error) {
	return nil, nil
}
func (s *capacityServiceImpl) GetPlanByID(ctx context.Context, id string) (*model.CapacityPlan, error) {
	return nil, ErrCapacityPlanNotFound
}
func (s *capacityServiceImpl) UpdatePlan(ctx context.Context, id string, req model.UpdateCapacityPlanRequest) (*model.CapacityPlan, error) {
	return nil, ErrCapacityPlanNotFound
}
func (s *capacityServiceImpl) ApprovePlan(ctx context.Context, id string, approverID string) (*model.CapacityPlan, error) {
	return nil, ErrCapacityPlanNotFound
}
func (s *capacityServiceImpl) DeletePlan(ctx context.Context, id string) error {
	return nil
}
func (s *capacityServiceImpl) ListPlans(ctx context.Context, filter model.CapacityPlanFilterCriteria) ([]model.CapacityPlan, int, error) {
	return nil, 0, nil
}

// Bottlenecks
func (s *capacityServiceImpl) CreateBottleneck(ctx context.Context, workspaceID string, req model.CreateBottleneckRequest) (*model.BottleneckRecord, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	if _, err := s.repo.GetPoolByID(ctx, req.PoolID); err != nil {
		return nil, err
	}

	b := &model.BottleneckRecord{
		WorkspaceID:       workspaceID,
		PoolID:            req.PoolID,
		AffectedCIID:      req.AffectedCIID,
		Title:             strings.TrimSpace(req.Title),
		Description:       req.Description,
		Severity:          req.Severity,
		SaturationPercent: req.SaturationPercent,
		Status:            model.BottleneckStatusOpen,
		MitigationPlan:    req.MitigationPlan,
	}

	if err := s.repo.CreateBottleneck(ctx, b); err != nil {
		return nil, err
	}
	return b, nil
}

func (s *capacityServiceImpl) GetBottleneckByID(ctx context.Context, id string) (*model.BottleneckRecord, error) {
	return s.repo.GetBottleneckByID(ctx, id)
}

func (s *capacityServiceImpl) UpdateBottleneck(ctx context.Context, id string, req model.UpdateBottleneckRequest) (*model.BottleneckRecord, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	b, err := s.repo.GetBottleneckByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.Title != nil {
		b.Title = strings.TrimSpace(*req.Title)
	}
	if req.Description != nil {
		b.Description = *req.Description
	}
	if req.Severity != nil {
		b.Severity = *req.Severity
	}
	if req.SaturationPercent != nil {
		b.SaturationPercent = *req.SaturationPercent
	}
	if req.Status != nil {
		b.Status = *req.Status
		if *req.Status == model.BottleneckStatusResolved && b.ResolvedAt == nil {
			now := time.Now().UTC()
			b.ResolvedAt = &now
		}
	}
	if req.MitigationPlan != nil {
		b.MitigationPlan = *req.MitigationPlan
	}

	if err := s.repo.UpdateBottleneck(ctx, b); err != nil {
		return nil, err
	}
	return b, nil
}

func (s *capacityServiceImpl) ResolveBottleneck(ctx context.Context, id string, mitigationPlan string) (*model.BottleneckRecord, error) {
	b, err := s.repo.GetBottleneckByID(ctx, id)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	b.Status = model.BottleneckStatusResolved
	b.ResolvedAt = &now
	if mitigationPlan != "" {
		b.MitigationPlan = mitigationPlan
	}

	if err := s.repo.UpdateBottleneck(ctx, b); err != nil {
		return nil, err
	}
	return b, nil
}

func (s *capacityServiceImpl) DeleteBottleneck(ctx context.Context, id string) error {
	return s.repo.DeleteBottleneck(ctx, id)
}

func (s *capacityServiceImpl) ListBottlenecks(ctx context.Context, filter model.BottleneckFilterCriteria) ([]model.BottleneckRecord, int, error) {
	return s.repo.ListBottlenecks(ctx, filter)
}

// Analytics Stub
func (s *capacityServiceImpl) GetCapacityAnalytics(ctx context.Context, workspaceID string) (*model.CapacityAnalyticsReport, error) {
	return &model.CapacityAnalyticsReport{WorkspaceID: workspaceID}, nil
}
