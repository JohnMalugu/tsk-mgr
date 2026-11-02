package service

import (
	"errors"
	"time"

	"github.com/JohnMalugu/tsk-mgr-api/internal/model"
	"github.com/JohnMalugu/tsk-mgr-api/internal/repository"
)

var (
	ErrProfileNotFound          = errors.New("resource profile not found")
	ErrProfileExists            = errors.New("resource profile already exists for user in workspace")
	ErrTimeOffNotFound          = errors.New("resource time-off not found")
	ErrAllocationNotFound       = errors.New("project resource allocation not found")
	ErrUserIDRequired           = errors.New("user ID is required")
	ErrInvalidCapacity          = errors.New("capacity hours must be positive and within valid limits")
	ErrInvalidDateRange         = errors.New("start date must be before or equal to end date")
	ErrInvalidTimeOffType       = errors.New("invalid time-off type")
	ErrInvalidAllocationStatus  = errors.New("invalid allocation status")
)

// WorkloadService orchestrates resource capacity, time-offs, project commitments, and team workload analytics
type WorkloadService interface {
	CreateProfile(req model.CreateResourceProfileRequest) (*model.ResourceProfile, error)
	GetProfile(id int64) (*model.ResourceProfile, error)
	GetProfileByUser(workspaceID int64, userID string) (*model.ResourceProfile, error)
	ListProfiles(workspaceID int64, isActive *bool) ([]*model.ResourceProfile, error)
	UpdateProfile(id int64, req model.UpdateResourceProfileRequest) (*model.ResourceProfile, error)
	DeleteProfile(id int64) error

	CreateTimeOff(req model.CreateTimeOffRequest) (*model.ResourceTimeOff, error)
	GetTimeOff(id int64) (*model.ResourceTimeOff, error)
	ListTimeOffs(workspaceID int64, userID string, from, to *time.Time) ([]model.ResourceTimeOff, error)
	UpdateTimeOff(id int64, req model.UpdateTimeOffRequest) (*model.ResourceTimeOff, error)
	DeleteTimeOff(id int64) error

	CreateAllocation(req model.CreateResourceAllocationRequest) (*model.ProjectResourceAllocation, error)
	GetAllocation(id int64) (*model.ProjectResourceAllocation, error)
	ListAllocations(workspaceID int64, projectID int, userID string, status string) ([]model.ProjectResourceAllocation, error)
	UpdateAllocation(id int64, req model.UpdateResourceAllocationRequest) (*model.ProjectResourceAllocation, error)
	DeleteAllocation(id int64) error

	GetUserWorkload(workspaceID int64, userID string, from, to time.Time) (*model.UserWorkloadSummary, error)
	GetTeamHeatmap(workspaceID int64, from, to time.Time) (*model.TeamWorkloadHeatmap, error)
	GetOverallocatedAlerts(workspaceID int64, from, to time.Time) ([]model.OverallocatedAlert, error)
	GetWorkspaceMetrics(workspaceID int64) (*model.WorkloadMetrics, error)
}

type workloadServiceImpl struct {
	repo        repository.WorkloadRepository
	projectRepo repository.ProjectRepository
}

// NewWorkloadService constructs a new WorkloadService instance
func NewWorkloadService(repo repository.WorkloadRepository, projectRepo repository.ProjectRepository) WorkloadService {
	return &workloadServiceImpl{
		repo:        repo,
		projectRepo: projectRepo,
	}
}

// Stubs for incremental implementation
func (s *workloadServiceImpl) CreateProfile(req model.CreateResourceProfileRequest) (*model.ResourceProfile, error) {
	return nil, nil
}
func (s *workloadServiceImpl) GetProfile(id int64) (*model.ResourceProfile, error) {
	return nil, nil
}
func (s *workloadServiceImpl) GetProfileByUser(workspaceID int64, userID string) (*model.ResourceProfile, error) {
	return nil, nil
}
func (s *workloadServiceImpl) ListProfiles(workspaceID int64, isActive *bool) ([]*model.ResourceProfile, error) {
	return nil, nil
}
func (s *workloadServiceImpl) UpdateProfile(id int64, req model.UpdateResourceProfileRequest) (*model.ResourceProfile, error) {
	return nil, nil
}
func (s *workloadServiceImpl) DeleteProfile(id int64) error {
	return nil
}

func (s *workloadServiceImpl) CreateTimeOff(req model.CreateTimeOffRequest) (*model.ResourceTimeOff, error) {
	return nil, nil
}
func (s *workloadServiceImpl) GetTimeOff(id int64) (*model.ResourceTimeOff, error) {
	return nil, nil
}
func (s *workloadServiceImpl) ListTimeOffs(workspaceID int64, userID string, from, to *time.Time) ([]model.ResourceTimeOff, error) {
	return nil, nil
}
func (s *workloadServiceImpl) UpdateTimeOff(id int64, req model.UpdateTimeOffRequest) (*model.ResourceTimeOff, error) {
	return nil, nil
}
func (s *workloadServiceImpl) DeleteTimeOff(id int64) error {
	return nil
}

func (s *workloadServiceImpl) CreateAllocation(req model.CreateResourceAllocationRequest) (*model.ProjectResourceAllocation, error) {
	return nil, nil
}
func (s *workloadServiceImpl) GetAllocation(id int64) (*model.ProjectResourceAllocation, error) {
	return nil, nil
}
func (s *workloadServiceImpl) ListAllocations(workspaceID int64, projectID int, userID string, status string) ([]model.ProjectResourceAllocation, error) {
	return nil, nil
}
func (s *workloadServiceImpl) UpdateAllocation(id int64, req model.UpdateResourceAllocationRequest) (*model.ProjectResourceAllocation, error) {
	return nil, nil
}
func (s *workloadServiceImpl) DeleteAllocation(id int64) error {
	return nil
}

func (s *workloadServiceImpl) GetUserWorkload(workspaceID int64, userID string, from, to time.Time) (*model.UserWorkloadSummary, error) {
	return nil, nil
}
func (s *workloadServiceImpl) GetTeamHeatmap(workspaceID int64, from, to time.Time) (*model.TeamWorkloadHeatmap, error) {
	return nil, nil
}
func (s *workloadServiceImpl) GetOverallocatedAlerts(workspaceID int64, from, to time.Time) ([]model.OverallocatedAlert, error) {
	return nil, nil
}
func (s *workloadServiceImpl) GetWorkspaceMetrics(workspaceID int64) (*model.WorkloadMetrics, error) {
	return nil, nil
}
