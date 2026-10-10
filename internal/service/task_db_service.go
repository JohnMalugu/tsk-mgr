package service

import (
	"errors"
	"time"

	"github.com/JohnMalugu/tsk-mgr-api/internal/model"
	"github.com/JohnMalugu/tsk-mgr-api/internal/repository"
)

var (
	ErrTaskNotFoundDB = errors.New("task not found in database")
	ErrUnauthorizedTask = errors.New("unauthorized to access this task")
)

type TaskDBService interface {
	CreateTask(userID, title, description string, dueDate time.Time, estimateMinutes int, priority string) (*model.Task, error)
	GetTask(id int, userID string) (*model.Task, error)
	GetUserTasks(userID string) ([]*model.Task, error)
	UpdateTask(id int, userID, title, description string, dueDate time.Time, estimateMinutes int, completed bool, status, priority string) (*model.Task, error)
	DeleteTask(id int, userID string) error
}

type taskDBServiceImpl struct {
	repo repository.TaskRepository
}

func NewTaskDBService(repo repository.TaskRepository) TaskDBService {
	return &taskDBServiceImpl{repo: repo}
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
