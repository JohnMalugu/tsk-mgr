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
