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
