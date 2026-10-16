package service

import (
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/JohnMalugu/tsk-mgr-api/internal/model"
	"github.com/JohnMalugu/tsk-mgr-api/internal/repository"
)

var (
	ErrActivityActionRequired  = errors.New("activity action is required")
	ErrActivitySummaryRequired = errors.New("activity summary is required")
	ErrActivityNotFoundDB      = errors.New("activity not found")
	ErrInvalidRetentionCutoff  = errors.New("retention days must be greater than zero")
)

type ActivityService interface {
	RecordActivity(userID string, taskID *int, action, summary string, details interface{}) (*model.ActivityDB, error)
	GetUserActivity(userID string, filter *model.ActivityFilter) (*model.PaginatedActivities, error)
	GetTaskActivity(taskID int, userID string, offset, limit int) ([]*model.ActivityDB, error)
	GetActivitySummary(userID string, from, to *time.Time) (*model.ActivitySummary, error)
	PurgeOldActivities(retentionDays int) (int64, error)
}

type activityServiceImpl struct {
	activityRepo repository.ActivityRepository
	taskRepo     repository.TaskRepository
}

func NewActivityService(activityRepo repository.ActivityRepository, taskRepo repository.TaskRepository) ActivityService {
	return &activityServiceImpl{
		activityRepo: activityRepo,
		taskRepo:     taskRepo,
	}
}

func (s *activityServiceImpl) RecordActivity(userID string, taskID *int, action, summary string, details interface{}) (*model.ActivityDB, error) {
	action = strings.TrimSpace(action)
	if action == "" {
		return nil, ErrActivityActionRequired
	}

	summary = strings.TrimSpace(summary)
	if summary == "" {
		return nil, ErrActivitySummaryRequired
	}

	if taskID != nil {
		task, err := s.taskRepo.GetByID(*taskID)
		if err != nil {
			return nil, err
		}
		if task == nil {
			return nil, ErrTaskNotFoundDB
		}
		if task.UserID != userID {
			return nil, ErrUnauthorizedTask
		}
	}

	var detailsJSON string
	if details != nil {
		b, err := json.Marshal(details)
		if err == nil {
			detailsJSON = string(b)
		}
	}

	entry := &model.ActivityDB{
		UserID:      userID,
		TaskID:      taskID,
		Action:      action,
		Summary:     summary,
		DetailsJSON: detailsJSON,
		CreatedAt:   time.Now().UTC(),
	}

	if err := s.activityRepo.Create(entry); err != nil {
		return nil, err
	}

	return entry, nil
}
