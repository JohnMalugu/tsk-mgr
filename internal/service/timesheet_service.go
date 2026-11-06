package service

import (
	"errors"
	"strings"
	"time"

	"github.com/JohnMalugu/tsk-mgr-api/internal/model"
	"github.com/JohnMalugu/tsk-mgr-api/internal/repository"
)

var (
	ErrWorkspaceNotFound       = ErrWorkspaceNotFoundDB
	ErrTimesheetNotFound       = errors.New("timesheet not found")
	ErrTimesheetLineNotFound   = errors.New("timesheet line not found")
	ErrInvalidPeriodDates      = errors.New("invalid timesheet period dates: end date must be on or after start date")
	ErrTimesheetAlreadyExists  = errors.New("timesheet for this workspace, user, and start date already exists")
	ErrTimesheetNotEditable    = errors.New("timesheet is not in draft status and cannot be edited")
	ErrTimesheetNotSubmittable = errors.New("timesheet must be in draft status to submit")
	ErrTimesheetNotReviewable  = errors.New("timesheet must be in submitted status to review")
	ErrUnauthorizedTimesheet   = errors.New("user not authorized to perform action on timesheet")
	ErrInvalidLineDuration     = errors.New("duration must be greater than zero")
	ErrInvalidWorkDate         = errors.New("work date is required and must fall within timesheet period")
	ErrRejectionReasonRequired = errors.New("rejection reason is required when rejecting a timesheet")
)

// TimesheetService manages enterprise timesheet generation, line item entry, manager approval workflows, and metrics
type TimesheetService interface {
	CreateTimesheet(userID string, req model.CreateTimesheetRequest) (*model.Timesheet, error)
	GetTimesheet(id int64) (*model.Timesheet, error)
	ListTimesheets(filter model.TimesheetFilter) ([]*model.Timesheet, error)
	UpdateTimesheet(id int64, userID string, req model.UpdateTimesheetRequest) (*model.Timesheet, error)
	DeleteTimesheet(id int64, userID string) error
	PopulateFromTimeEntries(id int64, userID string, req model.PopulateTimesheetRequest) (*model.Timesheet, error)

	AddLine(timesheetID int64, userID string, req model.TimesheetLineRequest) (*model.TimesheetLine, error)
	UpdateLine(timesheetID int64, lineID int64, userID string, req model.TimesheetLineRequest) (*model.TimesheetLine, error)
	DeleteLine(timesheetID int64, lineID int64, userID string) error
	ListLines(timesheetID int64) ([]model.TimesheetLine, error)

	SubmitTimesheet(id int64, userID string, req model.SubmitTimesheetRequest) (*model.Timesheet, error)
	ApproveTimesheet(id int64, reviewerID string, req model.ReviewTimesheetRequest) (*model.Timesheet, error)
	RejectTimesheet(id int64, reviewerID string, req model.ReviewTimesheetRequest) (*model.Timesheet, error)
	ReopenTimesheet(id int64, reviewerID string, comments string) (*model.Timesheet, error)
	ListApprovals(timesheetID int64) ([]model.TimesheetApproval, error)
	GetTimesheetMetrics(workspaceID int64) (*model.TimesheetMetrics, error)
}

type timesheetServiceImpl struct {
	repo          repository.TimesheetRepository
	timeEntryRepo repository.TimeEntryRepository
	taskRepo      repository.TaskRepository
	projectRepo   repository.ProjectRepository
	workspaceRepo repository.WorkspaceRepository
}

// NewTimesheetService creates a new instance of TimesheetService
func NewTimesheetService(
	repo repository.TimesheetRepository,
	timeEntryRepo repository.TimeEntryRepository,
	taskRepo repository.TaskRepository,
	projectRepo repository.ProjectRepository,
	workspaceRepo repository.WorkspaceRepository,
) TimesheetService {
	return &timesheetServiceImpl{
		repo:          repo,
		timeEntryRepo: timeEntryRepo,
		taskRepo:      taskRepo,
		projectRepo:   projectRepo,
		workspaceRepo: workspaceRepo,
	}
}

func (s *timesheetServiceImpl) CreateTimesheet(userID string, req model.CreateTimesheetRequest) (*model.Timesheet, error) {
	if req.WorkspaceID <= 0 {
		return nil, ErrWorkspaceRequired
	}
	if strings.TrimSpace(userID) == "" {
		return nil, ErrUnauthorizedTimesheet
	}

	startDate, err := time.Parse("2006-01-02", req.PeriodStartDate)
	if err != nil {
		return nil, ErrInvalidPeriodDates
	}
	endDate, err := time.Parse("2006-01-02", req.PeriodEndDate)
	if err != nil {
		return nil, ErrInvalidPeriodDates
	}
	if endDate.Before(startDate) {
		return nil, ErrInvalidPeriodDates
	}

	ws, err := s.workspaceRepo.GetByID(int(req.WorkspaceID))
	if err != nil || ws == nil {
		return nil, ErrWorkspaceNotFound
	}

	member, err := s.workspaceRepo.GetMember(int(req.WorkspaceID), userID)
	if err != nil || member == nil {
		return nil, ErrUnauthorizedTimesheet
	}

	existing, err := s.repo.GetTimesheetByPeriod(req.WorkspaceID, userID, req.PeriodStartDate)
	if err == nil && existing != nil {
		return nil, ErrTimesheetAlreadyExists
	}

	ts := &model.Timesheet{
		WorkspaceID:     req.WorkspaceID,
		UserID:          userID,
		PeriodStartDate: req.PeriodStartDate,
		PeriodEndDate:   req.PeriodEndDate,
		Status:          model.TimesheetStatusDraft,
		Notes:           strings.TrimSpace(req.Notes),
	}
	id, err := s.repo.CreateTimesheet(ts)
	if err != nil {
		return nil, err
	}

	if req.AutoPopulate {
		_, _ = s.populateHelper(id, userID, false)
	}

	return s.GetTimesheet(id)
}

func (s *timesheetServiceImpl) GetTimesheet(id int64) (*model.Timesheet, error) {
	ts, err := s.repo.GetTimesheetByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrTimesheetNotFound) {
			return nil, ErrTimesheetNotFound
		}
		return nil, err
	}
	return ts, nil
}

func (s *timesheetServiceImpl) ListTimesheets(filter model.TimesheetFilter) ([]*model.Timesheet, error) {
	if filter.WorkspaceID <= 0 {
		return nil, ErrWorkspaceRequired
	}
	list, err := s.repo.ListTimesheets(filter)
	if err != nil {
		return nil, err
	}
	if list == nil {
		list = []*model.Timesheet{}
	}
	return list, nil
}

func (s *timesheetServiceImpl) populateHelper(timesheetID int64, userID string, includeExisting bool) (*model.Timesheet, error) {
	ts, err := s.repo.GetTimesheetByID(timesheetID)
	if err != nil {
		return nil, ErrTimesheetNotFound
	}
	if ts.UserID != userID {
		return nil, ErrUnauthorizedTimesheet
	}
	if ts.Status != model.TimesheetStatusDraft {
		return nil, ErrTimesheetNotEditable
	}

	startTime, err := time.Parse("2006-01-02", ts.PeriodStartDate)
	if err != nil {
		return nil, ErrInvalidPeriodDates
	}
	endTime, err := time.Parse("2006-01-02", ts.PeriodEndDate)
	if err != nil {
		return nil, ErrInvalidPeriodDates
	}
	endTime = endTime.Add(23*time.Hour + 59*time.Minute + 59*time.Second)

	entries, err := s.timeEntryRepo.GetUserTimeEntriesInRange(userID, &startTime, &endTime)
	if err != nil {
		return nil, err
	}

	var existingLines []model.TimesheetLine
	if !includeExisting {
		existingLines, _ = s.repo.ListTimesheetLines(timesheetID)
	}

	for _, entry := range entries {
		if entry.DurationSeconds <= 0 {
			continue
		}
		durationMins := int(entry.DurationSeconds / 60)
		if durationMins == 0 && entry.DurationSeconds > 0 {
			durationMins = 1
		}
		workDate := entry.StartedAt.Format("2006-01-02")

		if !includeExisting {
			alreadyExists := false
			for _, line := range existingLines {
				if line.WorkDate == workDate && line.TaskID != nil && *line.TaskID == entry.TaskID {
					alreadyExists = true
					break
				}
			}
			if alreadyExists {
				continue
			}
		}

		var projID *int
		var taskID *int
		if entry.TaskID > 0 {
			tid := entry.TaskID
			taskID = &tid
			task, err := s.taskRepo.GetByID(entry.TaskID)
			if err == nil && task != nil && task.ProjectID != nil {
				pid := *task.ProjectID
				projID = &pid
			}
		}

		line := &model.TimesheetLine{
			TimesheetID:     timesheetID,
			TaskID:          taskID,
			ProjectID:       projID,
			WorkDate:        workDate,
			DurationMinutes: durationMins,
			IsBillable:      true,
			Description:     entry.Note,
		}
		_, _ = s.repo.AddTimesheetLine(line)
	}

	if err := s.repo.RecalculateTotals(timesheetID); err != nil {
		return nil, err
	}

	return s.GetTimesheet(timesheetID)
}

func (s *timesheetServiceImpl) UpdateTimesheet(id int64, userID string, req model.UpdateTimesheetRequest) (*model.Timesheet, error) {
	ts, err := s.GetTimesheet(id)
	if err != nil {
		return nil, err
	}
	if ts.UserID != userID {
		return nil, ErrUnauthorizedTimesheet
	}
	if ts.Status != model.TimesheetStatusDraft {
		return nil, ErrTimesheetNotEditable
	}

	if req.Notes != nil {
		ts.Notes = strings.TrimSpace(*req.Notes)
	}

	if err := s.repo.UpdateTimesheet(ts); err != nil {
		return nil, err
	}

	return s.GetTimesheet(id)
}

func (s *timesheetServiceImpl) DeleteTimesheet(id int64, userID string) error {
	ts, err := s.GetTimesheet(id)
	if err != nil {
		return err
	}
	if ts.UserID != userID {
		return ErrUnauthorizedTimesheet
	}
	if ts.Status != model.TimesheetStatusDraft {
		return ErrTimesheetNotEditable
	}

	return s.repo.DeleteTimesheet(id)
}

func (s *timesheetServiceImpl) PopulateFromTimeEntries(id int64, userID string, req model.PopulateTimesheetRequest) (*model.Timesheet, error) {
	return s.populateHelper(id, userID, req.IncludeExisting)
}

func (s *timesheetServiceImpl) AddLine(timesheetID int64, userID string, req model.TimesheetLineRequest) (*model.TimesheetLine, error) {
	return nil, nil
}

func (s *timesheetServiceImpl) UpdateLine(timesheetID int64, lineID int64, userID string, req model.TimesheetLineRequest) (*model.TimesheetLine, error) {
	return nil, nil
}

func (s *timesheetServiceImpl) DeleteLine(timesheetID int64, lineID int64, userID string) error {
	return nil
}

func (s *timesheetServiceImpl) ListLines(timesheetID int64) ([]model.TimesheetLine, error) {
	return nil, nil
}

func (s *timesheetServiceImpl) SubmitTimesheet(id int64, userID string, req model.SubmitTimesheetRequest) (*model.Timesheet, error) {
	return nil, nil
}

func (s *timesheetServiceImpl) ApproveTimesheet(id int64, reviewerID string, req model.ReviewTimesheetRequest) (*model.Timesheet, error) {
	return nil, nil
}

func (s *timesheetServiceImpl) RejectTimesheet(id int64, reviewerID string, req model.ReviewTimesheetRequest) (*model.Timesheet, error) {
	return nil, nil
}

func (s *timesheetServiceImpl) ReopenTimesheet(id int64, reviewerID string, comments string) (*model.Timesheet, error) {
	return nil, nil
}

func (s *timesheetServiceImpl) ListApprovals(timesheetID int64) ([]model.TimesheetApproval, error) {
	return nil, nil
}

func (s *timesheetServiceImpl) GetTimesheetMetrics(workspaceID int64) (*model.TimesheetMetrics, error) {
	return nil, nil
}
