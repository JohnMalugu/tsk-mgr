package service

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/JohnMalugu/tsk-mgr-api/internal/model"
	"github.com/JohnMalugu/tsk-mgr-api/internal/repository"
)

var (
	ErrScheduleNotFound         = errors.New("on-call schedule not found")
	ErrScheduleNameRequired     = errors.New("on-call schedule name is required")
	ErrInvalidRotationType      = errors.New("invalid rotation type: must be daily, weekly, or custom")
	ErrInvalidTimezone          = errors.New("invalid timezone specification")
	ErrEscalationPolicyNotFound = errors.New("escalation policy not found")
	ErrEscalationPolicyNameReq  = errors.New("escalation policy name is required")
	ErrInvalidEscalationTarget  = errors.New("invalid escalation target type: must be user, schedule, or webhook")
	ErrOnCallOverrideNotFound   = errors.New("on-call override not found")
	ErrInvalidOverrideTimeRange = errors.New("override start time must be before end time")
	ErrOverrideUserRequired     = errors.New("override user id is required")
	ErrOnCallPageNotFound       = errors.New("on-call page not found")
	ErrPageTitleRequired        = errors.New("page title is required")
	ErrNoOnCallResponders       = errors.New("no active on-call responders available")
)

// OnCallService coordinates schedules, rotation calculation, overrides, escalation policies, and incident pages
type OnCallService interface {
	// Schedules & Rotations
	CreateSchedule(creatorID string, req model.CreateScheduleRequest) (*model.OnCallSchedule, error)
	GetSchedule(id int64) (*model.OnCallSchedule, error)
	ListSchedules(filter model.OnCallFilter) ([]model.OnCallSchedule, int, error)
	UpdateSchedule(id int64, req model.UpdateScheduleRequest) (*model.OnCallSchedule, error)
	DeleteSchedule(id int64) error
	SetParticipants(scheduleID int64, userIDs []string) error
	ListParticipants(scheduleID int64) ([]model.OnCallParticipant, error)
	GenerateShifts(scheduleID int64, horizonDays int) ([]model.OnCallShift, error)
	ListShifts(scheduleID int64, startTime, endTime time.Time) ([]model.OnCallShift, error)
	GetActiveResponder(scheduleID int64) (*model.ActiveResponder, error)

	// Escalation Policies
	CreateEscalationPolicy(req model.CreateEscalationPolicyRequest) (*model.EscalationPolicy, error)
	GetEscalationPolicy(id int64) (*model.EscalationPolicy, error)
	GetDefaultEscalationPolicy(workspaceID int64) (*model.EscalationPolicy, error)
	ListEscalationPolicies(workspaceID int64) ([]model.EscalationPolicy, error)
	UpdateEscalationPolicy(id int64, req model.CreateEscalationPolicyRequest) (*model.EscalationPolicy, error)
	DeleteEscalationPolicy(id int64) error

	// Overrides
	CreateOverride(creatorID string, scheduleID int64, req model.CreateOverrideRequest) (*model.OnCallOverride, error)
	GetOverride(id int64) (*model.OnCallOverride, error)
	ListOverrides(scheduleID int64) ([]model.OnCallOverride, error)
	DeleteOverride(id int64) error

	// Pages & Alert Workflows
	TriggerPage(creatorID string, req model.TriggerPageRequest) (*model.OnCallPage, error)
	AcknowledgePage(pageID int64, userID string, req model.AcknowledgePageRequest) (*model.OnCallPage, error)
	ResolvePage(pageID int64, req model.ResolvePageRequest) (*model.OnCallPage, error)
	EscalatePage(pageID int64) (*model.OnCallPage, error)
	GetPage(id int64) (*model.OnCallPage, error)
	ListPages(workspaceID int64, status string, limit, offset int) ([]model.OnCallPage, int, error)

	// Operational Metrics
	GetOnCallMetrics(workspaceID int64) (*model.OnCallMetrics, error)
}

type oncallServiceImpl struct {
	repo        repository.OnCallRepository
	projectRepo repository.ProjectRepository
	userRepo    repository.UserRepository
}

// NewOnCallService creates a new OnCallService instance
func NewOnCallService(
	repo repository.OnCallRepository,
	projectRepo repository.ProjectRepository,
	userRepo repository.UserRepository,
) OnCallService {
	return &oncallServiceImpl{
		repo:        repo,
		projectRepo: projectRepo,
		userRepo:    userRepo,
	}
}

func parseTimeOfDay(timeStr string) (int, int) {
	parts := strings.Split(strings.TrimSpace(timeStr), ":")
	if len(parts) != 2 {
		return 9, 0
	}
	h, err1 := strconv.Atoi(parts[0])
	m, err2 := strconv.Atoi(parts[1])
	if err1 != nil || err2 != nil || h < 0 || h > 23 || m < 0 || m > 59 {
		return 9, 0
	}
	return h, m
}

func (s *oncallServiceImpl) CreateSchedule(creatorID string, req model.CreateScheduleRequest) (*model.OnCallSchedule, error) {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, ErrScheduleNameRequired
	}

	if req.RotationType == "" {
		req.RotationType = model.RotationWeekly
	}
	if err := model.ValidateRotationType(req.RotationType); err != nil {
		return nil, ErrInvalidRotationType
	}

	if err := model.ValidateTimezone(req.Timezone); err != nil {
		return nil, ErrInvalidTimezone
	}
	if req.Timezone == "" {
		req.Timezone = "UTC"
	}

	if req.RotationIntervalDays <= 0 {
		switch strings.ToLower(req.RotationType) {
		case model.RotationDaily:
			req.RotationIntervalDays = 1
		case model.RotationWeekly:
			req.RotationIntervalDays = 7
		default:
			req.RotationIntervalDays = 7
		}
	}

	if strings.TrimSpace(req.ShiftStartTime) == "" {
		req.ShiftStartTime = "09:00"
	}
	if req.ShiftDurationHours <= 0 {
		req.ShiftDurationHours = 24
	}

	if req.StartDate.IsZero() {
		req.StartDate = time.Now().Truncate(24 * time.Hour)
	}

	schedule := &model.OnCallSchedule{
		WorkspaceID:          req.WorkspaceID,
		ProjectID:            req.ProjectID,
		Name:                 name,
		Description:          strings.TrimSpace(req.Description),
		Timezone:             req.Timezone,
		RotationType:         req.RotationType,
		RotationIntervalDays: req.RotationIntervalDays,
		ShiftStartTime:       req.ShiftStartTime,
		ShiftDurationHours:   req.ShiftDurationHours,
		HandoverDayOfWeek:    req.HandoverDayOfWeek,
		StartDate:            req.StartDate,
		EndDate:              req.EndDate,
		IsActive:             true,
		CreatedByID:          creatorID,
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
	}

	scheduleID, err := s.repo.CreateSchedule(schedule, req.ParticipantUserIDs)
	if err != nil {
		return nil, err
	}

	if len(req.ParticipantUserIDs) > 0 {
		_, _ = s.GenerateShifts(scheduleID, 30)
	}

	return s.GetSchedule(scheduleID)
}

func (s *oncallServiceImpl) GetSchedule(id int64) (*model.OnCallSchedule, error) {
	sched, err := s.repo.GetScheduleByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrScheduleNotFound) {
			return nil, ErrScheduleNotFound
		}
		return nil, err
	}

	if resp, err := s.GetActiveResponder(id); err == nil && resp != nil {
		sched.CurrentOnCallUserID = &resp.UserID
		sched.CurrentOnCallUserName = resp.UserName
	}

	return sched, nil
}

func (s *oncallServiceImpl) ListSchedules(filter model.OnCallFilter) ([]model.OnCallSchedule, int, error) {
	schedules, total, err := s.repo.ListSchedules(filter)
	if err != nil {
		return nil, 0, err
	}

	for i := range schedules {
		if resp, err := s.GetActiveResponder(schedules[i].ID); err == nil && resp != nil {
			schedules[i].CurrentOnCallUserID = &resp.UserID
			schedules[i].CurrentOnCallUserName = resp.UserName
		}
	}

	return schedules, total, nil
}

func (s *oncallServiceImpl) UpdateSchedule(id int64, req model.UpdateScheduleRequest) (*model.OnCallSchedule, error) {
	existing, err := s.repo.GetScheduleByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrScheduleNotFound) {
			return nil, ErrScheduleNotFound
		}
		return nil, err
	}

	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)
		if name == "" {
			return nil, ErrScheduleNameRequired
		}
		existing.Name = name
	}
	if req.Description != nil {
		existing.Description = strings.TrimSpace(*req.Description)
	}
	if req.Timezone != nil {
		tz := strings.TrimSpace(*req.Timezone)
		if err := model.ValidateTimezone(tz); err != nil {
			return nil, ErrInvalidTimezone
		}
		existing.Timezone = tz
	}
	if req.RotationType != nil {
		rot := strings.TrimSpace(*req.RotationType)
		if err := model.ValidateRotationType(rot); err != nil {
			return nil, ErrInvalidRotationType
		}
		existing.RotationType = rot
	}
	if req.RotationIntervalDays != nil && *req.RotationIntervalDays > 0 {
		existing.RotationIntervalDays = *req.RotationIntervalDays
	}
	if req.ShiftStartTime != nil {
		existing.ShiftStartTime = strings.TrimSpace(*req.ShiftStartTime)
	}
	if req.ShiftDurationHours != nil && *req.ShiftDurationHours > 0 {
		existing.ShiftDurationHours = *req.ShiftDurationHours
	}
	if req.HandoverDayOfWeek != nil {
		existing.HandoverDayOfWeek = *req.HandoverDayOfWeek
	}
	if req.IsActive != nil {
		existing.IsActive = *req.IsActive
	}

	if err := s.repo.UpdateSchedule(existing); err != nil {
		if errors.Is(err, repository.ErrScheduleNotFound) {
			return nil, ErrScheduleNotFound
		}
		return nil, err
	}

	_, _ = s.GenerateShifts(id, 30)

	return s.GetSchedule(id)
}

func (s *oncallServiceImpl) DeleteSchedule(id int64) error {
	err := s.repo.DeleteSchedule(id)
	if err != nil {
		if errors.Is(err, repository.ErrScheduleNotFound) {
			return ErrScheduleNotFound
		}
		return err
	}
	return nil
}

func (s *oncallServiceImpl) SetParticipants(scheduleID int64, userIDs []string) error {
	_, err := s.repo.GetScheduleByID(scheduleID)
	if err != nil {
		if errors.Is(err, repository.ErrScheduleNotFound) {
			return ErrScheduleNotFound
		}
		return err
	}

	if err := s.repo.SetParticipants(scheduleID, userIDs); err != nil {
		return err
	}

	_, _ = s.GenerateShifts(scheduleID, 30)
	return nil
}

func (s *oncallServiceImpl) ListParticipants(scheduleID int64) ([]model.OnCallParticipant, error) {
	_, err := s.repo.GetScheduleByID(scheduleID)
	if err != nil {
		if errors.Is(err, repository.ErrScheduleNotFound) {
			return nil, ErrScheduleNotFound
		}
		return nil, err
	}
	return s.repo.ListParticipants(scheduleID)
}

func (s *oncallServiceImpl) GenerateShifts(scheduleID int64, horizonDays int) ([]model.OnCallShift, error) {
	sched, err := s.repo.GetScheduleByID(scheduleID)
	if err != nil {
		if errors.Is(err, repository.ErrScheduleNotFound) {
			return nil, ErrScheduleNotFound
		}
		return nil, err
	}

	participants, err := s.repo.ListParticipants(scheduleID)
	if err != nil {
		return nil, err
	}
	if len(participants) == 0 {
		return nil, nil
	}

	if horizonDays <= 0 {
		horizonDays = 30
	}

	loc := time.UTC
	if sched.Timezone != "" {
		if l, err := time.LoadLocation(sched.Timezone); err == nil {
			loc = l
		}
	}

	hour, min := parseTimeOfDay(sched.ShiftStartTime)
	startDate := sched.StartDate.In(loc)
	startWindow := time.Date(startDate.Year(), startDate.Month(), startDate.Day(), hour, min, 0, 0, loc)

	intervalDays := sched.RotationIntervalDays
	if intervalDays <= 0 {
		intervalDays = 7
	}
	shiftDuration := time.Duration(sched.ShiftDurationHours) * time.Hour
	if shiftDuration <= 0 {
		shiftDuration = time.Duration(intervalDays*24) * time.Hour
	}
	rotationStep := time.Duration(intervalDays*24) * time.Hour

	cutoff := startWindow.AddDate(0, 0, horizonDays)
	if sched.EndDate != nil && sched.EndDate.Before(cutoff) {
		cutoff = sched.EndDate.In(loc)
	}

	var shifts []model.OnCallShift
	curr := startWindow
	pIndex := 0

	for curr.Before(cutoff) {
		shiftEnd := curr.Add(shiftDuration)
		if sched.EndDate != nil && shiftEnd.After(*sched.EndDate) {
			shiftEnd = sched.EndDate.In(loc)
		}

		p := participants[pIndex%len(participants)]
		shift := model.OnCallShift{
			ScheduleID:  scheduleID,
			UserID:      p.UserID,
			StartTime:   curr.UTC(),
			EndTime:     shiftEnd.UTC(),
			UserName:    p.UserName,
			UserEmail:   p.UserEmail,
			IsActiveNow: time.Now().After(curr.UTC()) && time.Now().Before(shiftEnd.UTC()),
		}
		shifts = append(shifts, shift)

		curr = curr.Add(rotationStep)
		pIndex++
	}

	if err := s.repo.SaveShifts(scheduleID, shifts); err != nil {
		return nil, err
	}

	return shifts, nil
}

func (s *oncallServiceImpl) ListShifts(scheduleID int64, startTime, endTime time.Time) ([]model.OnCallShift, error) {
	_, err := s.repo.GetScheduleByID(scheduleID)
	if err != nil {
		if errors.Is(err, repository.ErrScheduleNotFound) {
			return nil, ErrScheduleNotFound
		}
		return nil, err
	}
	if startTime.IsZero() {
		startTime = time.Now().AddDate(0, 0, -7)
	}
	if endTime.IsZero() {
		endTime = time.Now().AddDate(0, 0, 30)
	}
	return s.repo.ListShifts(scheduleID, startTime, endTime)
}

func (s *oncallServiceImpl) GetActiveResponder(scheduleID int64) (*model.ActiveResponder, error) {
	sched, err := s.repo.GetScheduleByID(scheduleID)
	if err != nil {
		if errors.Is(err, repository.ErrScheduleNotFound) {
			return nil, ErrScheduleNotFound
		}
		return nil, err
	}

	now := time.Now()

	// 1. Check active override
	activeOverride, err := s.repo.GetActiveOverride(scheduleID, now)
	if err == nil && activeOverride != nil {
		resp := &model.ActiveResponder{
			UserID:         activeOverride.OverrideUserID,
			UserName:       activeOverride.OverrideUserName,
			ScheduleID:     scheduleID,
			ScheduleName:   sched.Name,
			IsOverride:     true,
			ShiftStartTime: activeOverride.StartTime,
			ShiftEndTime:   activeOverride.EndTime,
		}
		if resp.UserName == "" && s.userRepo != nil {
			if u, err := s.userRepo.GetByID(resp.UserID); err == nil && u != nil {
				resp.UserName = u.Username
				resp.UserEmail = u.Email
			}
		}
		return resp, nil
	}

	// 2. Check current computed shift
	currentShift, err := s.repo.GetCurrentShift(scheduleID, now)
	if err == nil && currentShift != nil {
		resp := &model.ActiveResponder{
			UserID:         currentShift.UserID,
			UserName:       currentShift.UserName,
			UserEmail:      currentShift.UserEmail,
			ScheduleID:     scheduleID,
			ScheduleName:   sched.Name,
			IsOverride:     false,
			ShiftStartTime: currentShift.StartTime,
			ShiftEndTime:   currentShift.EndTime,
		}
		if resp.UserName == "" && s.userRepo != nil {
			if u, err := s.userRepo.GetByID(resp.UserID); err == nil && u != nil {
				resp.UserName = u.Username
				resp.UserEmail = u.Email
			}
		}
		return resp, nil
	}

	// 3. Fallback to first participant in schedule
	participants, err := s.repo.ListParticipants(scheduleID)
	if err == nil && len(participants) > 0 {
		p := participants[0]
		resp := &model.ActiveResponder{
			UserID:         p.UserID,
			UserName:       p.UserName,
			UserEmail:      p.UserEmail,
			ScheduleID:     scheduleID,
			ScheduleName:   sched.Name,
			IsOverride:     false,
			ShiftStartTime: now.Truncate(24 * time.Hour),
			ShiftEndTime:   now.Truncate(24 * time.Hour).Add(24 * time.Hour),
		}
		if resp.UserName == "" && s.userRepo != nil {
			if u, err := s.userRepo.GetByID(resp.UserID); err == nil && u != nil {
				resp.UserName = u.Username
				resp.UserEmail = u.Email
			}
		}
		return resp, nil
	}

	return nil, ErrNoOnCallResponders
}

func (s *oncallServiceImpl) CreateEscalationPolicy(req model.CreateEscalationPolicyRequest) (*model.EscalationPolicy, error) {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, ErrEscalationPolicyNameReq
	}

	var steps []model.EscalationStep
	for i, stepReq := range req.Steps {
		if err := model.ValidateTargetType(stepReq.TargetType); err != nil {
			return nil, ErrInvalidEscalationTarget
		}
		targetID := strings.TrimSpace(stepReq.TargetID)
		if targetID == "" {
			return nil, errors.New("escalation step target ID is required")
		}
		delay := stepReq.DelayMinutes
		if delay <= 0 {
			delay = 5
		}
		steps = append(steps, model.EscalationStep{
			StepOrder:    i + 1,
			TargetType:   strings.ToLower(strings.TrimSpace(stepReq.TargetType)),
			TargetID:     targetID,
			DelayMinutes: delay,
		})
	}

	policy := &model.EscalationPolicy{
		WorkspaceID: req.WorkspaceID,
		ProjectID:   req.ProjectID,
		Name:        name,
		Description: strings.TrimSpace(req.Description),
		RepeatCount: req.RepeatCount,
		IsDefault:   req.IsDefault,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	policyID, err := s.repo.CreateEscalationPolicy(policy, steps)
	if err != nil {
		return nil, err
	}

	return s.GetEscalationPolicy(policyID)
}

func (s *oncallServiceImpl) GetEscalationPolicy(id int64) (*model.EscalationPolicy, error) {
	policy, err := s.repo.GetEscalationPolicyByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrEscalationPolicyNotFound) {
			return nil, ErrEscalationPolicyNotFound
		}
		return nil, err
	}
	return policy, nil
}

func (s *oncallServiceImpl) GetDefaultEscalationPolicy(workspaceID int64) (*model.EscalationPolicy, error) {
	policy, err := s.repo.GetDefaultEscalationPolicy(workspaceID)
	if err != nil {
		if errors.Is(err, repository.ErrEscalationPolicyNotFound) {
			return nil, ErrEscalationPolicyNotFound
		}
		return nil, err
	}
	return policy, nil
}

func (s *oncallServiceImpl) ListEscalationPolicies(workspaceID int64) ([]model.EscalationPolicy, error) {
	return s.repo.ListEscalationPolicies(workspaceID)
}

func (s *oncallServiceImpl) UpdateEscalationPolicy(id int64, req model.CreateEscalationPolicyRequest) (*model.EscalationPolicy, error) {
	existing, err := s.repo.GetEscalationPolicyByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrEscalationPolicyNotFound) {
			return nil, ErrEscalationPolicyNotFound
		}
		return nil, err
	}

	if name := strings.TrimSpace(req.Name); name != "" {
		existing.Name = name
	}
	existing.Description = strings.TrimSpace(req.Description)
	existing.RepeatCount = req.RepeatCount
	existing.IsDefault = req.IsDefault
	if req.ProjectID != nil {
		existing.ProjectID = req.ProjectID
	}

	if err := s.repo.UpdateEscalationPolicy(existing); err != nil {
		if errors.Is(err, repository.ErrEscalationPolicyNotFound) {
			return nil, ErrEscalationPolicyNotFound
		}
		return nil, err
	}

	if len(req.Steps) > 0 {
		var steps []model.EscalationStep
		for i, stepReq := range req.Steps {
			if err := model.ValidateTargetType(stepReq.TargetType); err != nil {
				return nil, ErrInvalidEscalationTarget
			}
			targetID := strings.TrimSpace(stepReq.TargetID)
			if targetID == "" {
				return nil, errors.New("escalation step target ID is required")
			}
			delay := stepReq.DelayMinutes
			if delay <= 0 {
				delay = 5
			}
			steps = append(steps, model.EscalationStep{
				PolicyID:     id,
				StepOrder:    i + 1,
				TargetType:   strings.ToLower(strings.TrimSpace(stepReq.TargetType)),
				TargetID:     targetID,
				DelayMinutes: delay,
			})
		}
		if err := s.repo.SetEscalationSteps(id, steps); err != nil {
			return nil, err
		}
	}

	return s.GetEscalationPolicy(id)
}

func (s *oncallServiceImpl) DeleteEscalationPolicy(id int64) error {
	err := s.repo.DeleteEscalationPolicy(id)
	if err != nil {
		if errors.Is(err, repository.ErrEscalationPolicyNotFound) {
			return ErrEscalationPolicyNotFound
		}
		return err
	}
	return nil
}

func (s *oncallServiceImpl) CreateOverride(creatorID string, scheduleID int64, req model.CreateOverrideRequest) (*model.OnCallOverride, error) {
	_, err := s.repo.GetScheduleByID(scheduleID)
	if err != nil {
		if errors.Is(err, repository.ErrScheduleNotFound) {
			return nil, ErrScheduleNotFound
		}
		return nil, err
	}

	overrideUserID := strings.TrimSpace(req.OverrideUserID)
	if overrideUserID == "" {
		return nil, ErrOverrideUserRequired
	}

	if req.StartTime.IsZero() || req.EndTime.IsZero() || !req.StartTime.Before(req.EndTime) {
		return nil, ErrInvalidOverrideTimeRange
	}

	override := &model.OnCallOverride{
		ScheduleID:     scheduleID,
		OriginalUserID: req.OriginalUserID,
		OverrideUserID: overrideUserID,
		StartTime:      req.StartTime.UTC(),
		EndTime:        req.EndTime.UTC(),
		Reason:         strings.TrimSpace(req.Reason),
		CreatedByID:    creatorID,
		CreatedAt:      time.Now(),
	}

	overrideID, err := s.repo.CreateOverride(override)
	if err != nil {
		return nil, err
	}

	return s.GetOverride(overrideID)
}

func (s *oncallServiceImpl) GetOverride(id int64) (*model.OnCallOverride, error) {
	override, err := s.repo.GetOverrideByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrOnCallOverrideNotFound) {
			return nil, ErrOnCallOverrideNotFound
		}
		return nil, err
	}
	return override, nil
}

func (s *oncallServiceImpl) ListOverrides(scheduleID int64) ([]model.OnCallOverride, error) {
	_, err := s.repo.GetScheduleByID(scheduleID)
	if err != nil {
		if errors.Is(err, repository.ErrScheduleNotFound) {
			return nil, ErrScheduleNotFound
		}
		return nil, err
	}
	return s.repo.ListOverrides(scheduleID)
}

func (s *oncallServiceImpl) DeleteOverride(id int64) error {
	err := s.repo.DeleteOverride(id)
	if err != nil {
		if errors.Is(err, repository.ErrOnCallOverrideNotFound) {
			return ErrOnCallOverrideNotFound
		}
		return err
	}
	return nil
}

func (s *oncallServiceImpl) TriggerPage(creatorID string, req model.TriggerPageRequest) (*model.OnCallPage, error) {
	title := strings.TrimSpace(req.Title)
	if title == "" {
		return nil, ErrPageTitleRequired
	}

	var policy *model.EscalationPolicy
	var err error

	if req.PolicyID > 0 {
		policy, err = s.repo.GetEscalationPolicyByID(req.PolicyID)
	}
	if policy == nil || err != nil {
		policy, err = s.repo.GetDefaultEscalationPolicy(req.WorkspaceID)
		if err != nil {
			return nil, ErrEscalationPolicyNotFound
		}
	}

	var pagedUserID *string
	if len(policy.Steps) > 0 {
		step1 := policy.Steps[0]
		switch step1.TargetType {
		case model.TargetUser:
			uid := step1.TargetID
			pagedUserID = &uid
		case model.TargetSchedule:
			schedID, _ := strconv.ParseInt(step1.TargetID, 10, 64)
			if schedID > 0 {
				if resp, err := s.GetActiveResponder(schedID); err == nil && resp != nil {
					pagedUserID = &resp.UserID
				}
			}
		}
	}

	triggerSource := strings.TrimSpace(req.TriggerSource)
	if triggerSource == "" {
		triggerSource = model.PageTriggerManual
	}

	page := &model.OnCallPage{
		WorkspaceID:      req.WorkspaceID,
		PolicyID:         policy.ID,
		IncidentID:       req.IncidentID,
		Title:            title,
		Summary:          strings.TrimSpace(req.Summary),
		Status:           model.PageStatusTriggered,
		CurrentStepOrder: 1,
		PagedUserID:      pagedUserID,
		TriggerSource:    triggerSource,
		CreatedByID:      creatorID,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	pageID, err := s.repo.CreatePage(page)
	if err != nil {
		return nil, err
	}

	return s.GetPage(pageID)
}

func (s *oncallServiceImpl) AcknowledgePage(pageID int64, userID string, req model.AcknowledgePageRequest) (*model.OnCallPage, error) {
	page, err := s.repo.GetPageByID(pageID)
	if err != nil {
		if errors.Is(err, repository.ErrOnCallPageNotFound) {
			return nil, ErrOnCallPageNotFound
		}
		return nil, err
	}

	if page.Status == model.PageStatusResolved {
		return page, nil
	}

	now := time.Now()
	page.Status = model.PageStatusAcknowledged
	page.AcknowledgedBy = &userID
	page.AcknowledgedAt = &now

	if err := s.repo.UpdatePage(page); err != nil {
		return nil, err
	}

	return s.GetPage(pageID)
}

func (s *oncallServiceImpl) ResolvePage(pageID int64, req model.ResolvePageRequest) (*model.OnCallPage, error) {
	page, err := s.repo.GetPageByID(pageID)
	if err != nil {
		if errors.Is(err, repository.ErrOnCallPageNotFound) {
			return nil, ErrOnCallPageNotFound
		}
		return nil, err
	}

	now := time.Now()
	page.Status = model.PageStatusResolved
	page.ResolvedAt = &now

	if err := s.repo.UpdatePage(page); err != nil {
		return nil, err
	}

	return s.GetPage(pageID)
}

func (s *oncallServiceImpl) EscalatePage(pageID int64) (*model.OnCallPage, error) {
	page, err := s.repo.GetPageByID(pageID)
	if err != nil {
		if errors.Is(err, repository.ErrOnCallPageNotFound) {
			return nil, ErrOnCallPageNotFound
		}
		return nil, err
	}

	if page.Status == model.PageStatusResolved {
		return page, nil
	}

	policy, err := s.repo.GetEscalationPolicyByID(page.PolicyID)
	if err == nil && policy != nil {
		nextStepOrder := page.CurrentStepOrder + 1
		if nextStepOrder <= len(policy.Steps) {
			step := policy.Steps[nextStepOrder-1]
			page.CurrentStepOrder = nextStepOrder
			page.Status = model.PageStatusEscalated
			switch step.TargetType {
			case model.TargetUser:
				uid := step.TargetID
				page.PagedUserID = &uid
			case model.TargetSchedule:
				schedID, _ := strconv.ParseInt(step.TargetID, 10, 64)
				if schedID > 0 {
					if resp, err := s.GetActiveResponder(schedID); err == nil && resp != nil {
						page.PagedUserID = &resp.UserID
					}
				}
			}
		} else {
			page.Status = model.PageStatusEscalated
		}
	} else {
		page.Status = model.PageStatusEscalated
	}

	if err := s.repo.UpdatePage(page); err != nil {
		return nil, err
	}

	return s.GetPage(pageID)
}

func (s *oncallServiceImpl) GetPage(id int64) (*model.OnCallPage, error) {
	page, err := s.repo.GetPageByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrOnCallPageNotFound) {
			return nil, ErrOnCallPageNotFound
		}
		return nil, err
	}
	return page, nil
}

func (s *oncallServiceImpl) ListPages(workspaceID int64, status string, limit, offset int) ([]model.OnCallPage, int, error) {
	return s.repo.ListPages(workspaceID, status, limit, offset)
}

func (s *oncallServiceImpl) GetOnCallMetrics(workspaceID int64) (*model.OnCallMetrics, error) {
	return s.repo.GetOnCallMetrics(workspaceID)
}
