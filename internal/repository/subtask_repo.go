package repository

import (
	"database/sql"

	"github.com/JohnMalugu/tsk-mgr-api/internal/model"
)

type SubtaskRepository interface {
	Create(subtask *model.Subtask) error
	GetByID(id int) (*model.Subtask, error)
	GetByTaskID(taskID int) ([]*model.Subtask, error)
	Update(subtask *model.Subtask) error
	Delete(id int) error
	ReorderSubtasks(taskID int, subtaskIDs []int) error
}

type subtaskRepositoryImpl struct {
	db *sql.DB
}

func NewSubtaskRepository(db *sql.DB) SubtaskRepository {
	return &subtaskRepositoryImpl{db: db}
}
