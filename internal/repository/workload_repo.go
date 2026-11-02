package repository

import (
	"database/sql"
	"errors"
	"time"

	"github.com/JohnMalugu/tsk-mgr-api/internal/model"
)

var (
	ErrProfileNotFound    = errors.New("resource profile not found")
	ErrProfileExists      = errors.New("resource profile already exists for user in workspace")
	ErrTimeOffNotFound    = errors.New("resource time-off not found")
	ErrAllocationNotFound = errors.New("project resource allocation not found")
)

// WorkloadRepository manages persistence for capacity profiles, time-offs, project allocations, and workload queries
type WorkloadRepository interface {
	CreateProfile(profile *model.ResourceProfile) (int64, error)
	GetProfileByID(id int64) (*model.ResourceProfile, error)
	GetProfileByUserID(workspaceID int64, userID string) (*model.ResourceProfile, error)
	ListProfiles(workspaceID int64, isActive *bool) ([]*model.ResourceProfile, error)
	UpdateProfile(profile *model.ResourceProfile) error
	DeleteProfile(id int64) error

	CreateTimeOff(timeOff *model.ResourceTimeOff) (int64, error)
	GetTimeOffByID(id int64) (*model.ResourceTimeOff, error)
	ListTimeOffs(workspaceID int64, userID string, from, to *time.Time) ([]model.ResourceTimeOff, error)
	UpdateTimeOff(timeOff *model.ResourceTimeOff) error
	DeleteTimeOff(id int64) error

	CreateAllocation(alloc *model.ProjectResourceAllocation) (int64, error)
	GetAllocationByID(id int64) (*model.ProjectResourceAllocation, error)
	ListAllocations(workspaceID int64, projectID int, userID string, status string) ([]model.ProjectResourceAllocation, error)
	UpdateAllocation(alloc *model.ProjectResourceAllocation) error
	DeleteAllocation(id int64) error

	GetUserTasksForWorkload(workspaceID int64, userID string, from, to time.Time) ([]model.Task, error)
	GetWorkspaceWorkloadMetrics(workspaceID int64) (*model.WorkloadMetrics, error)
}

type workloadRepositoryImpl struct {
	db *sql.DB
}

// NewWorkloadRepository creates a new instance of WorkloadRepository
func NewWorkloadRepository(db *sql.DB) WorkloadRepository {
	return &workloadRepositoryImpl{db: db}
}

// Stubs for incremental implementation
func (r *workloadRepositoryImpl) CreateProfile(profile *model.ResourceProfile) (int64, error) {
	return 0, nil
}
func (r *workloadRepositoryImpl) GetProfileByID(id int64) (*model.ResourceProfile, error) {
	return nil, nil
}
func (r *workloadRepositoryImpl) GetProfileByUserID(workspaceID int64, userID string) (*model.ResourceProfile, error) {
	return nil, nil
}
func (r *workloadRepositoryImpl) ListProfiles(workspaceID int64, isActive *bool) ([]*model.ResourceProfile, error) {
	return nil, nil
}
func (r *workloadRepositoryImpl) UpdateProfile(profile *model.ResourceProfile) error {
	return nil
}
func (r *workloadRepositoryImpl) DeleteProfile(id int64) error {
	return nil
}

func (r *workloadRepositoryImpl) CreateTimeOff(timeOff *model.ResourceTimeOff) (int64, error) {
	return 0, nil
}
func (r *workloadRepositoryImpl) GetTimeOffByID(id int64) (*model.ResourceTimeOff, error) {
	return nil, nil
}
func (r *workloadRepositoryImpl) ListTimeOffs(workspaceID int64, userID string, from, to *time.Time) ([]model.ResourceTimeOff, error) {
	return nil, nil
}
func (r *workloadRepositoryImpl) UpdateTimeOff(timeOff *model.ResourceTimeOff) error {
	return nil
}
func (r *workloadRepositoryImpl) DeleteTimeOff(id int64) error {
	return nil
}

func (r *workloadRepositoryImpl) CreateAllocation(alloc *model.ProjectResourceAllocation) (int64, error) {
	return 0, nil
}
func (r *workloadRepositoryImpl) GetAllocationByID(id int64) (*model.ProjectResourceAllocation, error) {
	return nil, nil
}
func (r *workloadRepositoryImpl) ListAllocations(workspaceID int64, projectID int, userID string, status string) ([]model.ProjectResourceAllocation, error) {
	return nil, nil
}
func (r *workloadRepositoryImpl) UpdateAllocation(alloc *model.ProjectResourceAllocation) error {
	return nil
}
func (r *workloadRepositoryImpl) DeleteAllocation(id int64) error {
	return nil
}

func (r *workloadRepositoryImpl) GetUserTasksForWorkload(workspaceID int64, userID string, from, to time.Time) ([]model.Task, error) {
	return nil, nil
}
func (r *workloadRepositoryImpl) GetWorkspaceWorkloadMetrics(workspaceID int64) (*model.WorkloadMetrics, error) {
	return nil, nil
}
