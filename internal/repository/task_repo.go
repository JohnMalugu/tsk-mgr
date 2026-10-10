package repository

import (
	"database/sql"
	"github.com/JohnMalugu/tsk-mgr-api/internal/model"
)

type TaskRepository interface {
	Create(task *model.Task) error
	GetByID(id int) (*model.Task, error)
	GetByUserID(userID string) ([]*model.Task, error)
	Update(task *model.Task) error
	Delete(id int) error
}

type taskRepositoryImpl struct {
	db *sql.DB
}

func NewTaskRepository(db *sql.DB) TaskRepository {
	return &taskRepositoryImpl{db: db}
}
