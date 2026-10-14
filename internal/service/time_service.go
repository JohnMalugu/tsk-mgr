package service

import (
	"errors"
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
