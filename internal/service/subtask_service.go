package service

import (
	"errors"

	"github.com/JohnMalugu/tsk-mgr-api/internal/model"
	"github.com/JohnMalugu/tsk-mgr-api/internal/repository"
)

var (
	ErrSubtaskNotFound     = errors.New("subtask not found")
	ErrSubtaskTitleRequired = errors.New("subtask title is required")
)

type SubtaskService interface {
	CreateSubtask(taskID int, userID, title string) (*model.Subtask, error)
	GetSubtasks(taskID int, userID string) ([]*model.Subtask, error)
	ToggleSubtask(subtaskID int, userID string, completed bool) (*model.Subtask, error)
	UpdateSubtask(subtaskID int, userID, title string) (*model.Subtask, error)
	DeleteSubtask(subtaskID int, userID string) error
	ReorderSubtasks(taskID int, userID string, subtaskIDs []int) error
}

type subtaskServiceImpl struct {
	subtaskRepo repository.SubtaskRepository
	taskRepo    repository.TaskRepository
}

func NewSubtaskService(subtaskRepo repository.SubtaskRepository, taskRepo repository.TaskRepository) SubtaskService {
	return &subtaskServiceImpl{
		subtaskRepo: subtaskRepo,
		taskRepo:    taskRepo,
	}
}
