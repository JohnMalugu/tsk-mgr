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
	if req.WorkspaceID <= 0 {
		return nil, ErrWorkspaceRequired
	}
	userID := req.UserID
	if userID == "" {
		return nil, ErrUserIDRequired
	}

	weeklyCapacity := model.DefaultWeeklyCapacityHours
	if req.WeeklyCapacityHours != nil {
		if *req.WeeklyCapacityHours <= 0 || *req.WeeklyCapacityHours > model.MaxWeeklyCapacityHours {
			return nil, ErrInvalidCapacity
		}
		weeklyCapacity = *req.WeeklyCapacityHours
	}

	dailyCapacity := model.DefaultDailyCapacityHours
	if req.DailyCapacityHours != nil {
		if *req.DailyCapacityHours <= 0 || *req.DailyCapacityHours > model.MaxDailyCapacityHours {
			return nil, ErrInvalidCapacity
		}
		dailyCapacity = *req.DailyCapacityHours
	}

	var rate float64
	if req.HourlyRate != nil && *req.HourlyRate >= 0 {
		rate = *req.HourlyRate
	}

	curr := req.Currency
	if curr == "" {
		curr = "USD"
	}

	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	skills := req.Skills
	if skills == nil {
		skills = []string{}
	}

	profile := &model.ResourceProfile{
		WorkspaceID:         req.WorkspaceID,
		UserID:              userID,
		JobTitle:            req.JobTitle,
		WeeklyCapacityHours: weeklyCapacity,
		DailyCapacityHours:  dailyCapacity,
		Skills:              skills,
		HourlyRate:          rate,
		Currency:            curr,
		IsActive:            isActive,
	}

	id, err := s.repo.CreateProfile(profile)
	if err != nil {
		if errors.Is(err, repository.ErrProfileExists) {
			return nil, ErrProfileExists
		}
		return nil, err
	}
	profile.ID = id
	return s.repo.GetProfileByID(id)
}

func (s *workloadServiceImpl) GetProfile(id int64) (*model.ResourceProfile, error) {
	p, err := s.repo.GetProfileByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrProfileNotFound) {
			return nil, ErrProfileNotFound
		}
		return nil, err
	}
	return p, nil
}

func (s *workloadServiceImpl) GetProfileByUser(workspaceID int64, userID string) (*model.ResourceProfile, error) {
	if workspaceID <= 0 {
		return nil, ErrWorkspaceRequired
	}
	if userID == "" {
		return nil, ErrUserIDRequired
	}
	p, err := s.repo.GetProfileByUserID(workspaceID, userID)
	if err != nil {
		if errors.Is(err, repository.ErrProfileNotFound) {
			return nil, ErrProfileNotFound
		}
		return nil, err
	}
	return p, nil
}

func (s *workloadServiceImpl) ListProfiles(workspaceID int64, isActive *bool) ([]*model.ResourceProfile, error) {
	if workspaceID <= 0 {
		return nil, ErrWorkspaceRequired
	}
	return s.repo.ListProfiles(workspaceID, isActive)
}

func (s *workloadServiceImpl) UpdateProfile(id int64, req model.UpdateResourceProfileRequest) (*model.ResourceProfile, error) {
	profile, err := s.repo.GetProfileByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrProfileNotFound) {
			return nil, ErrProfileNotFound
		}
		return nil, err
	}

	if req.JobTitle != nil {
		profile.JobTitle = *req.JobTitle
	}
	if req.WeeklyCapacityHours != nil {
		if *req.WeeklyCapacityHours <= 0 || *req.WeeklyCapacityHours > model.MaxWeeklyCapacityHours {
			return nil, ErrInvalidCapacity
		}
		profile.WeeklyCapacityHours = *req.WeeklyCapacityHours
	}
	if req.DailyCapacityHours != nil {
		if *req.DailyCapacityHours <= 0 || *req.DailyCapacityHours > model.MaxDailyCapacityHours {
			return nil, ErrInvalidCapacity
		}
		profile.DailyCapacityHours = *req.DailyCapacityHours
	}
	if req.Skills != nil {
		profile.Skills = *req.Skills
	}
	if req.HourlyRate != nil && *req.HourlyRate >= 0 {
		profile.HourlyRate = *req.HourlyRate
	}
	if req.Currency != nil && *req.Currency != "" {
		profile.Currency = *req.Currency
	}
	if req.IsActive != nil {
		profile.IsActive = *req.IsActive
	}

	if err := s.repo.UpdateProfile(profile); err != nil {
		return nil, err
	}
	return s.repo.GetProfileByID(id)
}

func (s *workloadServiceImpl) DeleteProfile(id int64) error {
	err := s.repo.DeleteProfile(id)
	if err != nil && errors.Is(err, repository.ErrProfileNotFound) {
		return ErrProfileNotFound
	}
	return err
}

func (s *workloadServiceImpl) CreateTimeOff(req model.CreateTimeOffRequest) (*model.ResourceTimeOff, error) {
	if req.WorkspaceID <= 0 {
		return nil, ErrWorkspaceRequired
	}
	if req.UserID == "" {
		return nil, ErrUserIDRequired
	}
	if req.EndDate.Before(req.StartDate) {
		return nil, ErrInvalidDateRange
	}

	timeOffType := req.TimeOffType
	if timeOffType == "" {
		timeOffType = model.TimeOffTypeVacation
	} else if timeOffType != model.TimeOffTypeVacation &&
		timeOffType != model.TimeOffTypeSickLeave &&
		timeOffType != model.TimeOffTypePublicHoliday &&
		timeOffType != model.TimeOffTypeTraining &&
		timeOffType != model.TimeOffTypePersonal {
		return nil, ErrInvalidTimeOffType
	}

	hoursPerDay := model.DefaultDailyCapacityHours
	if req.HoursPerDay != nil {
		if *req.HoursPerDay <= 0 || *req.HoursPerDay > model.MaxDailyCapacityHours {
			return nil, ErrInvalidCapacity
		}
		hoursPerDay = *req.HoursPerDay
	}

	timeOff := &model.ResourceTimeOff{
		WorkspaceID: req.WorkspaceID,
		UserID:      req.UserID,
		StartDate:   req.StartDate,
		EndDate:     req.EndDate,
		HoursPerDay: hoursPerDay,
		TimeOffType: timeOffType,
		Description: req.Description,
	}

	id, err := s.repo.CreateTimeOff(timeOff)
	if err != nil {
		return nil, err
	}
	timeOff.ID = id
	return s.repo.GetTimeOffByID(id)
}

func (s *workloadServiceImpl) GetTimeOff(id int64) (*model.ResourceTimeOff, error) {
	to, err := s.repo.GetTimeOffByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrTimeOffNotFound) {
			return nil, ErrTimeOffNotFound
		}
		return nil, err
	}
	return to, nil
}

func (s *workloadServiceImpl) ListTimeOffs(workspaceID int64, userID string, from, to *time.Time) ([]model.ResourceTimeOff, error) {
	if workspaceID <= 0 {
		return nil, ErrWorkspaceRequired
	}
	if from != nil && to != nil && to.Before(*from) {
		return nil, ErrInvalidDateRange
	}
	return s.repo.ListTimeOffs(workspaceID, userID, from, to)
}

func (s *workloadServiceImpl) UpdateTimeOff(id int64, req model.UpdateTimeOffRequest) (*model.ResourceTimeOff, error) {
	timeOff, err := s.repo.GetTimeOffByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrTimeOffNotFound) {
			return nil, ErrTimeOffNotFound
		}
		return nil, err
	}

	if req.StartDate != nil {
		timeOff.StartDate = *req.StartDate
	}
	if req.EndDate != nil {
		timeOff.EndDate = *req.EndDate
	}
	if timeOff.EndDate.Before(timeOff.StartDate) {
		return nil, ErrInvalidDateRange
	}

	if req.TimeOffType != nil {
		tot := *req.TimeOffType
		if tot != model.TimeOffTypeVacation &&
			tot != model.TimeOffTypeSickLeave &&
			tot != model.TimeOffTypePublicHoliday &&
			tot != model.TimeOffTypeTraining &&
			tot != model.TimeOffTypePersonal {
			return nil, ErrInvalidTimeOffType
		}
		timeOff.TimeOffType = tot
	}

	if req.HoursPerDay != nil {
		if *req.HoursPerDay <= 0 || *req.HoursPerDay > model.MaxDailyCapacityHours {
			return nil, ErrInvalidCapacity
		}
		timeOff.HoursPerDay = *req.HoursPerDay
	}
	if req.Description != nil {
		timeOff.Description = *req.Description
	}

	if err := s.repo.UpdateTimeOff(timeOff); err != nil {
		return nil, err
	}
	return s.repo.GetTimeOffByID(id)
}

func (s *workloadServiceImpl) DeleteTimeOff(id int64) error {
	err := s.repo.DeleteTimeOff(id)
	if err != nil && errors.Is(err, repository.ErrTimeOffNotFound) {
		return ErrTimeOffNotFound
	}
	return err
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
