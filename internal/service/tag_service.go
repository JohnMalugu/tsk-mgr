package service

import (
	"errors"

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
