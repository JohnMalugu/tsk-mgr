package service

import (
	"errors"
	"time"

	"github.com/JohnMalugu/tsk-mgr-api/internal/model"
	"github.com/JohnMalugu/tsk-mgr-api/internal/repository"
)

var (
	ErrTaskNotFoundDB   = errors.New("task not found in database")
	ErrUnauthorizedTask = errors.New("unauthorized to access this task")
)

type TaskDBService interface {
	CreateTask(userID, title, description string, dueDate time.Time, estimateMinutes int, priority string) (*model.Task, error)
	GetTask(id int, userID string) (*model.Task, error)
	GetUserTasks(userID string) ([]*model.Task, error)
	UpdateTask(id int, userID, title, description string, dueDate time.Time, estimateMinutes int, completed bool, status, priority string) (*model.Task, error)
	DeleteTask(id int, userID string) error
	CompleteTask(id int, userID string, completed bool) (*model.Task, error)
	SetActivityService(activitySvc ActivityService)
}

type taskDBServiceImpl struct {
	repo        repository.TaskRepository
	activitySvc ActivityService
}

func NewTaskDBService(repo repository.TaskRepository) TaskDBService {
	return &taskDBServiceImpl{repo: repo}
}

func (s *taskDBServiceImpl) SetActivityService(activitySvc ActivityService) {
	s.activitySvc = activitySvc
}

func (s *taskDBServiceImpl) CreateTask(userID, title, description string, dueDate time.Time, estimateMinutes int, priority string) (*model.Task, error) {
	now := time.Now().UTC()
	task := &model.Task{
		UserID:          userID,
		Title:           title,
		Description:     description,
		CreatedAt:       now,
		UpdatedAt:       now,
		DueDate:         dueDate,
		EstimateMinutes: estimateMinutes,
		Completed:       false,
		Status:          model.TaskStatusTodo,
		Priority:        priority,
	}

	if err := s.repo.Create(task); err != nil {
		return nil, err
	}

	if s.activitySvc != nil {
		_, _ = s.activitySvc.RecordActivity(userID, &task.ID, "task_created", "Created task: "+task.Title, nil)
	}

	return task, nil
}

func (s *taskDBServiceImpl) GetTask(id int, userID string) (*model.Task, error) {
	task, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if task == nil {
		return nil, ErrTaskNotFoundDB
	}

	if task.UserID != userID {
		return nil, ErrUnauthorizedTask
	}

	return task, nil
}

func (s *taskDBServiceImpl) GetUserTasks(userID string) ([]*model.Task, error) {
	return s.repo.GetByUserID(userID)
}

func (s *taskDBServiceImpl) UpdateTask(id int, userID, title, description string, dueDate time.Time, estimateMinutes int, completed bool, status, priority string) (*model.Task, error) {
	task, err := s.GetTask(id, userID)
	if err != nil {
		return nil, err
	}

	task.Title = title
	task.Description = description
	task.DueDate = dueDate
	task.EstimateMinutes = estimateMinutes
	task.Completed = completed
	task.Status = status
	task.Priority = priority
	task.UpdatedAt = time.Now().UTC()

	if err := s.repo.Update(task); err != nil {
		return nil, err
	}

	if s.activitySvc != nil {
		_, _ = s.activitySvc.RecordActivity(userID, &task.ID, "task_updated", "Updated task: "+task.Title, nil)
	}

	return task, nil
}

func (s *taskDBServiceImpl) DeleteTask(id int, userID string) error {
	task, err := s.GetTask(id, userID)
	if err != nil {
		return err
	}

	if err := s.repo.Delete(id); err != nil {
		return err
	}

	if s.activitySvc != nil {
		_, _ = s.activitySvc.RecordActivity(userID, &id, "task_deleted", "Deleted task: "+task.Title, nil)
	}

	return nil
}

func (s *taskDBServiceImpl) CompleteTask(id int, userID string, completed bool) (*model.Task, error) {
	task, err := s.GetTask(id, userID)
	if err != nil {
		return nil, err
	}

	task.Completed = completed
	if completed {
		task.Status = model.TaskStatusCompleted
	} else {
		task.Status = model.TaskStatusTodo
	}
	task.UpdatedAt = time.Now().UTC()

	if err := s.repo.Update(task); err != nil {
		return nil, err
	}

	if s.activitySvc != nil {
		action := "task_reopened"
		summary := "Reopened task: " + task.Title
		if completed {
			action = "task_completed"
			summary = "Completed task: " + task.Title
		}
		_, _ = s.activitySvc.RecordActivity(userID, &task.ID, action, summary, nil)
	}

	return task, nil
}
