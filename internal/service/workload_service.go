package service

import (
	"errors"
	"math"
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
	if req.WorkspaceID <= 0 {
		return nil, ErrWorkspaceRequired
	}
	if req.UserID == "" {
		return nil, ErrUserIDRequired
	}
	if req.ProjectID <= 0 {
		return nil, ErrProjectNotFound
	}
	if s.projectRepo != nil {
		p, err := s.projectRepo.GetByID(req.ProjectID)
		if err != nil || p == nil {
			return nil, ErrProjectNotFound
		}
	}
	if req.EndDate.Before(req.StartDate) {
		return nil, ErrInvalidDateRange
	}
	if req.AllocatedHoursPerWeek <= 0 || req.AllocatedHoursPerWeek > model.MaxWeeklyCapacityHours {
		return nil, ErrInvalidCapacity
	}

	alloc := &model.ProjectResourceAllocation{
		WorkspaceID:           req.WorkspaceID,
		ProjectID:             req.ProjectID,
		UserID:                req.UserID,
		AllocatedHoursPerWeek: req.AllocatedHoursPerWeek,
		StartDate:             req.StartDate,
		EndDate:               req.EndDate,
		Status:                model.AllocationStatusPlanned,
		Notes:                 req.Notes,
	}

	id, err := s.repo.CreateAllocation(alloc)
	if err != nil {
		return nil, err
	}
	alloc.ID = id
	return s.repo.GetAllocationByID(id)
}

func (s *workloadServiceImpl) GetAllocation(id int64) (*model.ProjectResourceAllocation, error) {
	a, err := s.repo.GetAllocationByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrAllocationNotFound) {
			return nil, ErrAllocationNotFound
		}
		return nil, err
	}
	return a, nil
}

func (s *workloadServiceImpl) ListAllocations(workspaceID int64, projectID int, userID string, status string) ([]model.ProjectResourceAllocation, error) {
	if workspaceID <= 0 {
		return nil, ErrWorkspaceRequired
	}
	return s.repo.ListAllocations(workspaceID, projectID, userID, status)
}

func (s *workloadServiceImpl) UpdateAllocation(id int64, req model.UpdateResourceAllocationRequest) (*model.ProjectResourceAllocation, error) {
	alloc, err := s.repo.GetAllocationByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrAllocationNotFound) {
			return nil, ErrAllocationNotFound
		}
		return nil, err
	}

	if req.AllocatedHoursPerWeek != nil {
		if *req.AllocatedHoursPerWeek <= 0 || *req.AllocatedHoursPerWeek > model.MaxWeeklyCapacityHours {
			return nil, ErrInvalidCapacity
		}
		alloc.AllocatedHoursPerWeek = *req.AllocatedHoursPerWeek
	}
	if req.StartDate != nil {
		alloc.StartDate = *req.StartDate
	}
	if req.EndDate != nil {
		alloc.EndDate = *req.EndDate
	}
	if alloc.EndDate.Before(alloc.StartDate) {
		return nil, ErrInvalidDateRange
	}

	if req.Status != nil {
		st := *req.Status
		if st != model.AllocationStatusPlanned &&
			st != model.AllocationStatusActive &&
			st != model.AllocationStatusCompleted &&
			st != model.AllocationStatusCanceled {
			return nil, ErrInvalidAllocationStatus
		}
		alloc.Status = st
	}
	if req.Notes != nil {
		alloc.Notes = *req.Notes
	}

	if err := s.repo.UpdateAllocation(alloc); err != nil {
		return nil, err
	}
	return s.repo.GetAllocationByID(id)
}

func (s *workloadServiceImpl) DeleteAllocation(id int64) error {
	err := s.repo.DeleteAllocation(id)
	if err != nil && errors.Is(err, repository.ErrAllocationNotFound) {
		return ErrAllocationNotFound
	}
	return err
}

func (s *workloadServiceImpl) GetUserWorkload(workspaceID int64, userID string, from, to time.Time) (*model.UserWorkloadSummary, error) {
	if workspaceID <= 0 {
		return nil, ErrWorkspaceRequired
	}
	if userID == "" {
		return nil, ErrUserIDRequired
	}
	if !from.IsZero() && !to.IsZero() && to.Before(from) {
		return nil, ErrInvalidDateRange
	}

	if from.IsZero() {
		from = time.Now().Truncate(24 * time.Hour)
	} else {
		from = from.Truncate(24 * time.Hour)
	}
	if to.IsZero() {
		to = from.Add(7 * 24 * time.Hour)
	} else {
		to = to.Truncate(24 * time.Hour)
	}

	// 1. Fetch profile (fallback to defaults if not registered)
	profile, err := s.repo.GetProfileByUserID(workspaceID, userID)
	dailyCap := model.DefaultDailyCapacityHours
	username := userID
	email := ""
	jobTitle := ""
	if err == nil && profile != nil {
		dailyCap = profile.DailyCapacityHours
		username = profile.Username
		email = profile.Email
		jobTitle = profile.JobTitle
	}

	// 2. Fetch tasks
	tasks, err := s.repo.GetUserTasksForWorkload(workspaceID, userID, from, to.Add(24*time.Hour))
	if err != nil {
		return nil, err
	}

	// 3. Fetch time offs
	timeOffs, err := s.repo.ListTimeOffs(workspaceID, userID, &from, &to)
	if err != nil {
		return nil, err
	}

	// 4. Fetch active allocations
	allocs, err := s.repo.ListAllocations(workspaceID, 0, userID, model.AllocationStatusActive)
	if err != nil {
		return nil, err
	}

	summary := &model.UserWorkloadSummary{
		UserID:         userID,
		Username:       username,
		Email:          email,
		JobTitle:       jobTitle,
		DailyBreakdown: []model.WorkloadDailyBucket{},
	}

	var totalAvailable, totalTaskHours, totalAllocHours, totalTimeOff float64

	// Day-by-day iteration
	for d := from; !d.After(to); d = d.Add(24 * time.Hour) {
		dayStr := d.Format("2006-01-02")
		weekday := d.Weekday()
		isWeekend := weekday == time.Saturday || weekday == time.Sunday

		avail := dailyCap
		if isWeekend {
			avail = 0.0
		}

		// Time off for this day
		var dayTimeOff float64
		for _, toItem := range timeOffs {
			toStart := toItem.StartDate.Truncate(24 * time.Hour)
			toEnd := toItem.EndDate.Truncate(24 * time.Hour)
			if (!d.Before(toStart)) && (!d.After(toEnd)) && !isWeekend {
				dayTimeOff += toItem.HoursPerDay
			}
		}
		if dayTimeOff > avail {
			dayTimeOff = avail
		}
		avail -= dayTimeOff

		// Task demand for this day (tasks due on this day)
		var dayTaskHours float64
		var dayTaskCount int
		for _, t := range tasks {
			tDue := t.DueDate.Truncate(24 * time.Hour)
			if tDue.Equal(d) && t.Status != "completed" && !t.Completed {
				hrs := 1.0
				if t.EstimateMinutes > 0 {
					hrs = float64(t.EstimateMinutes) / 60.0
				}
				dayTaskHours += hrs
				dayTaskCount++
			}
		}

		// Allocation demand for this day
		var dayAllocHours float64
		for _, a := range allocs {
			aStart := a.StartDate.Truncate(24 * time.Hour)
			aEnd := a.EndDate.Truncate(24 * time.Hour)
			if (!d.Before(aStart)) && (!d.After(aEnd)) && !isWeekend {
				dayAllocHours += (a.AllocatedHoursPerWeek / 5.0)
			}
		}

		totalDemand := dayTaskHours + dayAllocHours
		var utilRate float64
		if avail > 0 {
			utilRate = math.Round((totalDemand/avail)*1000) / 10
		} else if totalDemand > 0 {
			utilRate = 100.0
		}

		isOver := totalDemand > avail && avail > 0

		bucket := model.WorkloadDailyBucket{
			Date:                  dayStr,
			AvailableHours:        math.Round(avail*10) / 10,
			AssignedTaskHours:     math.Round(dayTaskHours*10) / 10,
			AllocatedProjectHours: math.Round(dayAllocHours*10) / 10,
			TimeOffHours:          math.Round(dayTimeOff*10) / 10,
			TotalDemandHours:      math.Round(totalDemand*10) / 10,
			UtilizationRate:       utilRate,
			IsOverCapacity:        isOver,
			TaskCount:             dayTaskCount,
		}

		summary.DailyBreakdown = append(summary.DailyBreakdown, bucket)

		totalAvailable += avail
		totalTaskHours += dayTaskHours
		totalAllocHours += dayAllocHours
		totalTimeOff += dayTimeOff
	}

	summary.TotalAvailableHours = math.Round(totalAvailable*10) / 10
	summary.TotalAssignedHours = math.Round(totalTaskHours*10) / 10
	summary.TotalAllocatedHours = math.Round(totalAllocHours*10) / 10
	summary.TotalTimeOffHours = math.Round(totalTimeOff*10) / 10

	totDemand := summary.TotalAssignedHours + summary.TotalAllocatedHours
	if summary.TotalAvailableHours > 0 {
		summary.OverallUtilizationRate = math.Round((totDemand/summary.TotalAvailableHours)*1000) / 10
	} else if totDemand > 0 {
		summary.OverallUtilizationRate = 100.0
	}

	if summary.OverallUtilizationRate < 70.0 {
		summary.UtilizationStatus = model.UtilizationStatusUnder
	} else if summary.OverallUtilizationRate < 100.0 {
		summary.UtilizationStatus = model.UtilizationStatusOptimal
	} else if summary.OverallUtilizationRate == 100.0 {
		summary.UtilizationStatus = model.UtilizationStatusCapacity
	} else {
		summary.UtilizationStatus = model.UtilizationStatusOver
	}

	return summary, nil
}
func (s *workloadServiceImpl) GetTeamHeatmap(workspaceID int64, from, to time.Time) (*model.TeamWorkloadHeatmap, error) {
	if workspaceID <= 0 {
		return nil, ErrWorkspaceRequired
	}
	if !from.IsZero() && !to.IsZero() && to.Before(from) {
		return nil, ErrInvalidDateRange
	}

	if from.IsZero() {
		from = time.Now().Truncate(24 * time.Hour)
	} else {
		from = from.Truncate(24 * time.Hour)
	}
	if to.IsZero() {
		to = from.Add(7 * 24 * time.Hour)
	} else {
		to = to.Truncate(24 * time.Hour)
	}

	var days []string
	for d := from; !d.After(to); d = d.Add(24 * time.Hour) {
		days = append(days, d.Format("2006-01-02"))
	}

	isActive := true
	profiles, err := s.repo.ListProfiles(workspaceID, &isActive)
	if err != nil {
		return nil, err
	}

	heatmap := &model.TeamWorkloadHeatmap{
		WorkspaceID: workspaceID,
		StartDate:   from,
		EndDate:     to,
		Days:        days,
		Users:       []model.UserWorkloadSummary{},
	}

	var totalTeamAvailable, totalTeamDemand float64
	overallocatedUsers := make(map[string]bool)

	for _, p := range profiles {
		summary, err := s.GetUserWorkload(workspaceID, p.UserID, from, to)
		if err != nil {
			return nil, err
		}
		heatmap.Users = append(heatmap.Users, *summary)

		totalTeamAvailable += summary.TotalAvailableHours
		totalTeamDemand += (summary.TotalAssignedHours + summary.TotalAllocatedHours)

		if summary.UtilizationStatus == model.UtilizationStatusOver {
			overallocatedUsers[p.UserID] = true
		} else {
			for _, b := range summary.DailyBreakdown {
				if b.IsOverCapacity {
					overallocatedUsers[p.UserID] = true
					break
				}
			}
		}
	}

	if totalTeamAvailable > 0 {
		heatmap.OverallUtilizationRate = math.Round((totalTeamDemand/totalTeamAvailable)*1000) / 10
	} else if totalTeamDemand > 0 {
		heatmap.OverallUtilizationRate = 100.0
	}

	heatmap.OverallocatedUserCount = len(overallocatedUsers)
	return heatmap, nil
}

func (s *workloadServiceImpl) GetOverallocatedAlerts(workspaceID int64, from, to time.Time) ([]model.OverallocatedAlert, error) {
	heatmap, err := s.GetTeamHeatmap(workspaceID, from, to)
	if err != nil {
		return nil, err
	}

	var alerts []model.OverallocatedAlert
	for _, user := range heatmap.Users {
		for _, b := range user.DailyBreakdown {
			if b.IsOverCapacity {
				overHrs := math.Round((b.TotalDemandHours-b.AvailableHours)*10) / 10
				if overHrs < 0 {
					overHrs = 0
				}
				alerts = append(alerts, model.OverallocatedAlert{
					UserID:          user.UserID,
					Username:        user.Username,
					Date:            b.Date,
					CapacityHours:   b.AvailableHours,
					DemandHours:     b.TotalDemandHours,
					OverHours:       overHrs,
					UtilizationRate: b.UtilizationRate,
					ActiveTaskCount: b.TaskCount,
				})
			}
		}
	}

	if alerts == nil {
		alerts = []model.OverallocatedAlert{}
	}
	return alerts, nil
}
func (s *workloadServiceImpl) GetWorkspaceMetrics(workspaceID int64) (*model.WorkloadMetrics, error) {
	return nil, nil
}
