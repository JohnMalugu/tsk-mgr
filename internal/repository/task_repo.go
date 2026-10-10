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

func (r *taskRepositoryImpl) Create(task *model.Task) error {
	query := `
	INSERT INTO tasks (user_id, title, description, created_at, updated_at, due_date, estimate_minutes, completed, status, priority)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	result, err := r.db.Exec(query,
		task.UserID, task.Title, task.Description,
		task.CreatedAt, task.UpdatedAt, task.DueDate,
		task.EstimateMinutes, task.Completed, task.Status, task.Priority)

	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err == nil {
		task.ID = int(id)
	}

	return err
}

func (r *taskRepositoryImpl) GetByID(id int) (*model.Task, error) {
	query := `
	SELECT id, user_id, title, description, created_at, updated_at, due_date, estimate_minutes, completed, status, priority
	FROM tasks WHERE id = ?`

	row := r.db.QueryRow(query, id)
	var task model.Task

	err := row.Scan(&task.ID, &task.UserID, &task.Title, &task.Description,
		&task.CreatedAt, &task.UpdatedAt, &task.DueDate,
		&task.EstimateMinutes, &task.Completed, &task.Status, &task.Priority)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &task, nil
}

func (r *taskRepositoryImpl) GetByUserID(userID string) ([]*model.Task, error) {
	query := `
	SELECT id, user_id, title, description, created_at, updated_at, due_date, estimate_minutes, completed, status, priority
	FROM tasks WHERE user_id = ? ORDER BY created_at DESC`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []*model.Task
	for rows.Next() {
		var task model.Task
		err := rows.Scan(&task.ID, &task.UserID, &task.Title, &task.Description,
			&task.CreatedAt, &task.UpdatedAt, &task.DueDate,
			&task.EstimateMinutes, &task.Completed, &task.Status, &task.Priority)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, &task)
	}

	return tasks, rows.Err()
}

func (r *taskRepositoryImpl) Update(task *model.Task) error {
	query := `
	UPDATE tasks 
	SET title = ?, description = ?, updated_at = ?, due_date = ?, 
		estimate_minutes = ?, completed = ?, status = ?, priority = ?
	WHERE id = ? AND user_id = ?`

	_, err := r.db.Exec(query,
		task.Title, task.Description, task.UpdatedAt, task.DueDate,
		task.EstimateMinutes, task.Completed, task.Status, task.Priority,
		task.ID, task.UserID)

	return err
}

func (r *taskRepositoryImpl) Delete(id int) error {
	query := `DELETE FROM tasks WHERE id = ?`
	_, err := r.db.Exec(query, id)
	return err
}
