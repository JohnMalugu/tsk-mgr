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

func (r *subtaskRepositoryImpl) GetByID(id int) (*model.Subtask, error) {
	query := `
	SELECT id, task_id, title, completed, position, created_at, updated_at
	FROM subtasks WHERE id = ?`

	row := r.db.QueryRow(query, id)
	var s model.Subtask

	err := row.Scan(&s.ID, &s.TaskID, &s.Title, &s.Completed, &s.Position, &s.CreatedAt, &s.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &s, nil
}

func (r *subtaskRepositoryImpl) GetByTaskID(taskID int) ([]*model.Subtask, error) {
	query := `
	SELECT id, task_id, title, completed, position, created_at, updated_at
	FROM subtasks WHERE task_id = ? ORDER BY position ASC, created_at ASC`

	rows, err := r.db.Query(query, taskID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*model.Subtask
	for rows.Next() {
		var s model.Subtask
		err := rows.Scan(&s.ID, &s.TaskID, &s.Title, &s.Completed, &s.Position, &s.CreatedAt, &s.UpdatedAt)
		if err != nil {
			return nil, err
		}
		list = append(list, &s)
	}

	return list, rows.Err()
}

func (r *subtaskRepositoryImpl) Update(subtask *model.Subtask) error {
	query := `
	UPDATE subtasks
	SET title = ?, completed = ?, position = ?, updated_at = ?
	WHERE id = ?`

	_, err := r.db.Exec(query,
		subtask.Title, subtask.Completed, subtask.Position, subtask.UpdatedAt, subtask.ID)
	return err
}




