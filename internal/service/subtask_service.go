package service

import (
	"errors"
	"strings"
	"time"

	"github.com/JohnMalugu/tsk-mgr-api/internal/model"
	"github.com/JohnMalugu/tsk-mgr-api/internal/repository"
)

var (
	ErrSubtaskNotFound      = errors.New("subtask not found")
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

func (s *subtaskServiceImpl) CreateSubtask(taskID int, userID, title string) (*model.Subtask, error) {
	title = strings.TrimSpace(title)
	if title == "" {
		return nil, ErrSubtaskTitleRequired
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

	existing, err := s.subtaskRepo.GetByTaskID(taskID)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	subtask := &model.Subtask{
		TaskID:    taskID,
		Title:     title,
		Completed: false,
		Position:  len(existing),
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := s.subtaskRepo.Create(subtask); err != nil {
		return nil, err
	}

	return subtask, nil
}

func (s *subtaskServiceImpl) GetSubtasks(taskID int, userID string) ([]*model.Subtask, error) {
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

	return s.subtaskRepo.GetByTaskID(taskID)
}

func (s *subtaskServiceImpl) ToggleSubtask(subtaskID int, userID string, completed bool) (*model.Subtask, error) {
	subtask, err := s.subtaskRepo.GetByID(subtaskID)
	if err != nil {
		return nil, err
	}
	if subtask == nil {
		return nil, ErrSubtaskNotFound
	}

	task, err := s.taskRepo.GetByID(subtask.TaskID)
	if err != nil {
		return nil, err
	}
	if task == nil {
		return nil, ErrTaskNotFoundDB
	}
	if task.UserID != userID {
		return nil, ErrUnauthorizedTask
	}

	subtask.Completed = completed
	subtask.UpdatedAt = time.Now().UTC()

	if err := s.subtaskRepo.Update(subtask); err != nil {
		return nil, err
	}

	return subtask, nil
}

func (s *subtaskServiceImpl) UpdateSubtask(subtaskID int, userID, title string) (*model.Subtask, error) {
	title = strings.TrimSpace(title)
	if title == "" {
		return nil, ErrSubtaskTitleRequired
	}

	subtask, err := s.subtaskRepo.GetByID(subtaskID)
	if err != nil {
		return nil, err
	}
	if subtask == nil {
		return nil, ErrSubtaskNotFound
	}

	task, err := s.taskRepo.GetByID(subtask.TaskID)
	if err != nil {
		return nil, err
	}
	if task == nil {
		return nil, ErrTaskNotFoundDB
	}
	if task.UserID != userID {
		return nil, ErrUnauthorizedTask
	}

	subtask.Title = title
	subtask.UpdatedAt = time.Now().UTC()

	if err := s.subtaskRepo.Update(subtask); err != nil {
		return nil, err
	}

	return subtask, nil
}

func (s *subtaskServiceImpl) DeleteSubtask(subtaskID int, userID string) error {
	subtask, err := s.subtaskRepo.GetByID(subtaskID)
	if err != nil {
		return err
	}
	if subtask == nil {
		return ErrSubtaskNotFound
	}

	task, err := s.taskRepo.GetByID(subtask.TaskID)
	if err != nil {
		return err
	}
	if task == nil {
		return ErrTaskNotFoundDB
	}
	if task.UserID != userID {
		return ErrUnauthorizedTask
	}

	return s.subtaskRepo.Delete(subtaskID)
}

func (s *subtaskServiceImpl) ReorderSubtasks(taskID int, userID string, subtaskIDs []int) error {
	task, err := s.taskRepo.GetByID(taskID)
	if err != nil {
		return err
	}
	if task == nil {
		return ErrTaskNotFoundDB
	}
	if task.UserID != userID {
		return ErrUnauthorizedTask
	}

	return s.subtaskRepo.ReorderSubtasks(taskID, subtaskIDs)
}
