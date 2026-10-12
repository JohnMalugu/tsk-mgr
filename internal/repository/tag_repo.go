package repository

import (
	"database/sql"

	"github.com/JohnMalugu/tsk-mgr-api/internal/model"
)

type TagRepository interface {
	Create(tag *model.Tag) error
	GetByID(id int) (*model.Tag, error)
	GetByUserID(userID string) ([]*model.Tag, error)
	Update(tag *model.Tag) error
	Delete(id int) error
	AttachTagToTask(taskID, tagID int) error
	DetachTagFromTask(taskID, tagID int) error
	GetTagsForTask(taskID int) ([]*model.Tag, error)
	GetTasksForTag(tagID int) ([]*model.Task, error)
}

type tagRepositoryImpl struct {
	db *sql.DB
}

func NewTagRepository(db *sql.DB) TagRepository {
	return &tagRepositoryImpl{db: db}
}
