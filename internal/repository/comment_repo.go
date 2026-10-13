package repository

import (
	"database/sql"

	"github.com/JohnMalugu/tsk-mgr-api/internal/model"
)

type CommentRepository interface {
	Create(comment *model.Comment) error
	GetByID(id int) (*model.Comment, error)
	GetByTaskID(taskID int) ([]*model.Comment, error)
	Update(comment *model.Comment) error
	Delete(id int) error
}

type commentRepositoryImpl struct {
	db *sql.DB
}

func NewCommentRepository(db *sql.DB) CommentRepository {
	return &commentRepositoryImpl{db: db}
}
