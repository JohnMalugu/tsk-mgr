package service

import (
	"errors"
	"strings"
	"time"

	"github.com/JohnMalugu/tsk-mgr-api/internal/model"
	"github.com/JohnMalugu/tsk-mgr-api/internal/repository"
)

var (
	ErrTagNotFound     = errors.New("tag not found")
	ErrUnauthorizedTag = errors.New("unauthorized to access this tag")
	ErrTagNameRequired = errors.New("tag name is required")
)

type TagService interface {
	CreateTag(userID, name, color string) (*model.Tag, error)
	GetTag(id int, userID string) (*model.Tag, error)
	GetUserTags(userID string) ([]*model.Tag, error)
	UpdateTag(id int, userID, name, color string) (*model.Tag, error)
	DeleteTag(id int, userID string) error
	AddTagToTask(taskID, tagID int, userID string) error
	RemoveTagFromTask(taskID, tagID int, userID string) error
	GetTaskTags(taskID int, userID string) ([]*model.Tag, error)
}

type tagServiceImpl struct {
	tagRepo  repository.TagRepository
	taskRepo repository.TaskRepository
}

func NewTagService(tagRepo repository.TagRepository, taskRepo repository.TaskRepository) TagService {
	return &tagServiceImpl{
		tagRepo:  tagRepo,
		taskRepo: taskRepo,
	}
}

func (s *tagServiceImpl) CreateTag(userID, name, color string) (*model.Tag, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, ErrTagNameRequired
	}

	if color == "" {
		color = "#10b981"
	}

	tag := &model.Tag{
		UserID:    userID,
		Name:      name,
		Color:     color,
		CreatedAt: time.Now().UTC(),
	}

	if err := s.tagRepo.Create(tag); err != nil {
		return nil, err
	}

	return tag, nil
}

func (s *tagServiceImpl) GetTag(id int, userID string) (*model.Tag, error) {
	tag, err := s.tagRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if tag == nil {
		return nil, ErrTagNotFound
	}
	if tag.UserID != userID {
		return nil, ErrUnauthorizedTag
	}
	return tag, nil
}

func (s *tagServiceImpl) GetUserTags(userID string) ([]*model.Tag, error) {
	return s.tagRepo.GetByUserID(userID)
}

func (s *tagServiceImpl) UpdateTag(id int, userID, name, color string) (*model.Tag, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, ErrTagNameRequired
	}

	tag, err := s.GetTag(id, userID)
	if err != nil {
		return nil, err
	}

	tag.Name = name
	if color != "" {
		tag.Color = color
	}

	if err := s.tagRepo.Update(tag); err != nil {
		return nil, err
	}

	return tag, nil
}

func (s *tagServiceImpl) DeleteTag(id int, userID string) error {
	_, err := s.GetTag(id, userID)
	if err != nil {
		return err
	}

	return s.tagRepo.Delete(id)
}

func (s *tagServiceImpl) AddTagToTask(taskID, tagID int, userID string) error {
	_, err := s.GetTag(tagID, userID)
	if err != nil {
		return err
	}

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

	return s.tagRepo.AttachTagToTask(taskID, tagID)
}




