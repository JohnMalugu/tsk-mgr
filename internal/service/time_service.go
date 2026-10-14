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

func (s *timeTrackingServiceImpl) GetActiveTimer(userID string) (*model.TimeEntryDB, error) {
	active, err := s.timeRepo.GetActiveTimerByUser(userID)
	if err != nil {
		return nil, err
	}
	if active == nil {
		return nil, nil
	}

	// Calculate live elapsed seconds
	active.DurationSeconds = int64(time.Now().UTC().Sub(active.StartedAt).Seconds())
	return active, nil
}

func (s *timeTrackingServiceImpl) GetTaskTimeEntries(taskID int, userID string) ([]*model.TimeEntryDB, error) {
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

	entries, err := s.timeRepo.GetByTaskID(taskID)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	for _, e := range entries {
		if e.EndedAt == nil {
			e.DurationSeconds = int64(now.Sub(e.StartedAt).Seconds())
		}
	}

	return entries, nil
}

func (s *timeTrackingServiceImpl) LogManualTime(taskID int, userID string, startedAt, endedAt time.Time, note string) (*model.TimeEntryDB, error) {
	if !endedAt.After(startedAt) {
		return nil, ErrInvalidTimeRange
	}

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

	duration := int64(endedAt.Sub(startedAt).Seconds())
	now := time.Now().UTC()
	entry := &model.TimeEntryDB{
		TaskID:          taskID,
		UserID:          userID,
		StartedAt:       startedAt.UTC(),
		EndedAt:         &endedAt,
		DurationSeconds: duration,
		Note:            strings.TrimSpace(note),
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	if err := s.timeRepo.Create(entry); err != nil {
		return nil, err
	}

	return entry, nil
}

func (s *timeTrackingServiceImpl) DeleteTimeEntry(entryID int, userID string) error {
	entry, err := s.timeRepo.GetByID(entryID)
	if err != nil {
		return err
	}
	if entry == nil {
		return ErrTimeEntryNotFoundDB
	}
	if entry.UserID != userID {
		return ErrUnauthorizedTimeEntry
	}

	return s.timeRepo.Delete(entryID)
}

func (s *timeTrackingServiceImpl) GetTaskTimeSummary(taskID int, userID string) (*model.TaskTimeSummaryDB, error) {
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

	entries, err := s.timeRepo.GetByTaskID(taskID)
	if err != nil {
		return nil, err
	}

	var actualSeconds int64
	now := time.Now().UTC()
	for _, e := range entries {
		if e.EndedAt == nil {
			actualSeconds += int64(now.Sub(e.StartedAt).Seconds())
		} else {
			actualSeconds += e.DurationSeconds
		}
	}

	estimatedSeconds := int64(task.EstimateMinutes) * 60
	varianceSeconds := actualSeconds - estimatedSeconds

	return &model.TaskTimeSummaryDB{
		TaskID:           taskID,
		EstimatedSeconds: estimatedSeconds,
		ActualSeconds:    actualSeconds,
		VarianceSeconds:  varianceSeconds,
	}, nil
}

func (s *timeTrackingServiceImpl) GetTimeReport(userID string, from, to *time.Time) (*model.TimeReportDB, error) {
	entries, err := s.timeRepo.GetUserTimeEntriesInRange(userID, from, to)
	if err != nil {
		return nil, err
	}

	tasks, err := s.taskRepo.GetByUserID(userID)
	if err != nil {
		return nil, err
	}
	taskTitleMap := make(map[int]string)
	for _, t := range tasks {
		taskTitleMap[t.ID] = t.Title
	}

	var totalSeconds int64
	var activeCount int
	var completedCount int
	taskDurationMap := make(map[int]int64)

	now := time.Now().UTC()
	for _, e := range entries {
		var d int64
		if e.EndedAt == nil {
			activeCount++
			d = int64(now.Sub(e.StartedAt).Seconds())
		} else {
			completedCount++
			d = e.DurationSeconds
		}
		totalSeconds += d
		taskDurationMap[e.TaskID] += d
	}

	var byTask []model.TaskTimeTotalDB
	for tID, d := range taskDurationMap {
		title := taskTitleMap[tID]
		if title == "" {
			title = "Unknown Task"
		}
		byTask = append(byTask, model.TaskTimeTotalDB{
			TaskID:       tID,
			TaskTitle:    title,
			TotalSeconds: d,
		})
	}

	return &model.TimeReportDB{
		TotalSeconds: totalSeconds,
		EntryCount:   completedCount,
		ActiveCount:  activeCount,
		ByTask:       byTask,
	}, nil
}





