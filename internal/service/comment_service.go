package service

import (
	"errors"
	"strings"
	"time"

	"github.com/JohnMalugu/tsk-mgr-api/internal/model"
	"github.com/JohnMalugu/tsk-mgr-api/internal/repository"
)

var (
	ErrCommentContentRequired = errors.New("comment content is required")
	ErrUnauthorizedComment     = errors.New("unauthorized to access this comment")
)

type CommentService interface {
	AddComment(taskID int, userID, content string) (*model.Comment, error)
	GetTaskComments(taskID int, userID string) ([]*model.Comment, error)
	UpdateComment(commentID int, userID, content string) (*model.Comment, error)
	DeleteComment(commentID int, userID string) error
}

type commentServiceImpl struct {
	commentRepo repository.CommentRepository
	taskRepo    repository.TaskRepository
}

func NewCommentService(commentRepo repository.CommentRepository, taskRepo repository.TaskRepository) CommentService {
	return &commentServiceImpl{
		commentRepo: commentRepo,
		taskRepo:    taskRepo,
	}
}

func (s *commentServiceImpl) AddComment(taskID int, userID, content string) (*model.Comment, error) {
	content = strings.TrimSpace(content)
	if content == "" {
		return nil, ErrCommentContentRequired
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

	now := time.Now().UTC()
	comment := &model.Comment{
		TaskID:    taskID,
		UserID:    userID,
		Content:   content,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := s.commentRepo.Create(comment); err != nil {
		return nil, err
	}

	return comment, nil
}

func (s *commentServiceImpl) GetTaskComments(taskID int, userID string) ([]*model.Comment, error) {
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

	return s.commentRepo.GetByTaskID(taskID)
}

func (s *commentServiceImpl) UpdateComment(commentID int, userID, content string) (*model.Comment, error) {
	content = strings.TrimSpace(content)
	if content == "" {
		return nil, ErrCommentContentRequired
	}

	comment, err := s.commentRepo.GetByID(commentID)
	if err != nil {
		return nil, err
	}
	if comment == nil {
		return nil, ErrCommentNotFound
	}
	if comment.UserID != userID {
		return nil, ErrUnauthorizedComment
	}

	comment.Content = content
	comment.UpdatedAt = time.Now().UTC()

	if err := s.commentRepo.Update(comment); err != nil {
		return nil, err
	}

	return comment, nil
}

func (s *commentServiceImpl) DeleteComment(commentID int, userID string) error {
	comment, err := s.commentRepo.GetByID(commentID)
	if err != nil {
		return err
	}
	if comment == nil {
		return ErrCommentNotFound
	}
	if comment.UserID != userID {
		return ErrUnauthorizedComment
	}

	return s.commentRepo.Delete(commentID)
}
