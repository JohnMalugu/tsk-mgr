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

func (r *subtaskRepositoryImpl) Create(subtask *model.Subtask) error {
	query := `
	INSERT INTO subtasks (task_id, title, completed, position, created_at, updated_at)
	VALUES (?, ?, ?, ?, ?, ?)`

	result, err := r.db.Exec(query,
		subtask.TaskID, subtask.Title, subtask.Completed,
		subtask.Position, subtask.CreatedAt, subtask.UpdatedAt)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err == nil {
		subtask.ID = int(id)
	}

	return err
}

