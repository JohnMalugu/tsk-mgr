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

// Metric & Forecast Stubs
func (s *capacityServiceImpl) RecordMetric(ctx context.Context, workspaceID string, req model.RecordCapacityMetricRequest) (*model.CapacityMetric, error) {
	return nil, nil
}
func (s *capacityServiceImpl) GetMetricByID(ctx context.Context, id string) (*model.CapacityMetric, error) {
	return nil, ErrCapacityMetricNotFound
}
func (s *capacityServiceImpl) ListMetrics(ctx context.Context, filter model.CapacityMetricFilterCriteria) ([]model.CapacityMetric, int, error) {
	return nil, 0, nil
}
func (s *capacityServiceImpl) GenerateForecast(ctx context.Context, workspaceID string, poolID string, horizonDays int) (*model.CapacityForecast, error) {
	return nil, nil
}
func (s *capacityServiceImpl) GetLatestForecast(ctx context.Context, poolID string) (*model.CapacityForecast, error) {
	return nil, nil
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

// Bottleneck Stubs
func (s *capacityServiceImpl) CreateBottleneck(ctx context.Context, workspaceID string, req model.CreateBottleneckRequest) (*model.BottleneckRecord, error) {
	return nil, nil
}
func (s *capacityServiceImpl) GetBottleneckByID(ctx context.Context, id string) (*model.BottleneckRecord, error) {
	return nil, ErrBottleneckNotFound
}
func (s *capacityServiceImpl) UpdateBottleneck(ctx context.Context, id string, req model.UpdateBottleneckRequest) (*model.BottleneckRecord, error) {
	return nil, ErrBottleneckNotFound
}
func (s *capacityServiceImpl) ResolveBottleneck(ctx context.Context, id string, mitigationPlan string) (*model.BottleneckRecord, error) {
	return nil, ErrBottleneckNotFound
}
func (s *capacityServiceImpl) DeleteBottleneck(ctx context.Context, id string) error {
	return nil
}
func (s *capacityServiceImpl) ListBottlenecks(ctx context.Context, filter model.BottleneckFilterCriteria) ([]model.BottleneckRecord, int, error) {
	return nil, 0, nil
}

// Analytics Stub
func (s *capacityServiceImpl) GetCapacityAnalytics(ctx context.Context, workspaceID string) (*model.CapacityAnalyticsReport, error) {
	return &model.CapacityAnalyticsReport{WorkspaceID: workspaceID}, nil
}
