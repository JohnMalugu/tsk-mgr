package service

import (
	"errors"
	"strings"
	"time"

	"github.com/JohnMalugu/tsk-mgr-api/internal/model"
	"github.com/JohnMalugu/tsk-mgr-api/internal/repository"
)

var (
	ErrActiveTimerExists     = errors.New("an active timer is already running")
	ErrNoActiveTimerFound     = errors.New("no active timer is currently running")
	ErrInvalidTimeRange       = errors.New("ended time must be after started time")
	ErrTimeEntryNotFoundDB    = errors.New("time entry not found")
	ErrUnauthorizedTimeEntry  = errors.New("unauthorized to access this time entry")
)

type TimeTrackingService interface {
	StartTimer(taskID int, userID, note string) (*model.TimeEntryDB, error)
	StopTimer(userID string) (*model.TimeEntryDB, error)
	GetActiveTimer(userID string) (*model.TimeEntryDB, error)
	GetTaskTimeEntries(taskID int, userID string) ([]*model.TimeEntryDB, error)
	LogManualTime(taskID int, userID string, startedAt, endedAt time.Time, note string) (*model.TimeEntryDB, error)
	DeleteTimeEntry(entryID int, userID string) error
	GetTaskTimeSummary(taskID int, userID string) (*model.TaskTimeSummaryDB, error)
	GetTimeReport(userID string, from, to *time.Time) (*model.TimeReportDB, error)
}

type timeTrackingServiceImpl struct {
	timeRepo repository.TimeEntryRepository
	taskRepo repository.TaskRepository
}

func NewTimeTrackingService(timeRepo repository.TimeEntryRepository, taskRepo repository.TaskRepository) TimeTrackingService {
	return &timeTrackingServiceImpl{
		timeRepo: timeRepo,
		taskRepo: taskRepo,
	}
}

func (s *timeTrackingServiceImpl) StartTimer(taskID int, userID, note string) (*model.TimeEntryDB, error) {
	task, err := s.taskRepo.GetByID(taskID)
	if err != nil {
		return nil, err
	}
	if task == nil {
		return nil, ErrTaskNotFoundDB
	}
	if task.UserID != userID {
		return nil, ErrUnauthorizedTask
	}

	active, err := s.timeRepo.GetActiveTimerByUser(userID)
	if err != nil {
		return nil, err
	}
	if active != nil {
		return nil, ErrActiveTimerExists
	}

	now := time.Now().UTC()
	entry := &model.TimeEntryDB{
		TaskID:          taskID,
		UserID:          userID,
		StartedAt:       now,
		EndedAt:         nil,
		DurationSeconds: 0,
		Note:            strings.TrimSpace(note),
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	if err := s.timeRepo.Create(entry); err != nil {
		return nil, err
	}

	return entry, nil
}

func (s *timeTrackingServiceImpl) StopTimer(userID string) (*model.TimeEntryDB, error) {
	active, err := s.timeRepo.GetActiveTimerByUser(userID)
	if err != nil {
		return nil, err
	}
	if active == nil {
		return nil, ErrNoActiveTimerFound
	}

	endedAt := time.Now().UTC()
	duration := int64(endedAt.Sub(active.StartedAt).Seconds())
	if duration < 0 {
		duration = 0
	}

	active.EndedAt = &endedAt
	active.DurationSeconds = duration
	active.UpdatedAt = endedAt

	if err := s.timeRepo.Update(active); err != nil {
		return nil, err
	}

	return active, nil
}


