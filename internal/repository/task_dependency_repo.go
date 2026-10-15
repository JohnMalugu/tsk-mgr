package repository

import (
	"database/sql"
	"time"

	"github.com/JohnMalugu/tsk-mgr-api/internal/model"
)

type TaskDependencyRepository interface {
	AddDependency(taskID, dependsOnTaskID int) (*model.TaskDependencyDB, error)
	RemoveDependency(taskID, dependsOnTaskID int) error
	GetDependenciesByTaskID(taskID int) ([]*model.TaskDependencyDB, error)
	GetDependentsByTaskID(taskID int) ([]*model.TaskDependencyDB, error)
	Exists(taskID, dependsOnTaskID int) (bool, error)
	GetAllDependenciesForUser(userID string) ([]*model.TaskDependencyDB, error)
}

type taskDependencyRepositoryImpl struct {
	db *sql.DB
}

func NewTaskDependencyRepository(db *sql.DB) TaskDependencyRepository {
	return &taskDependencyRepositoryImpl{db: db}
}

func (r *taskDependencyRepositoryImpl) AddDependency(taskID, dependsOnTaskID int) (*model.TaskDependencyDB, error) {
	now := time.Now().UTC()
	query := `INSERT INTO task_dependencies (task_id, depends_on_task_id, created_at) VALUES (?, ?, ?)`
	res, err := r.db.Exec(query, taskID, dependsOnTaskID, now)
	if err != nil {
		return nil, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}
	return &model.TaskDependencyDB{
		ID:              int(id),
		TaskID:          taskID,
		DependsOnTaskID: dependsOnTaskID,
		CreatedAt:       now,
	}, nil
}

func (r *taskDependencyRepositoryImpl) RemoveDependency(taskID, dependsOnTaskID int) error {
	query := `DELETE FROM task_dependencies WHERE task_id = ? AND depends_on_task_id = ?`
	res, err := r.db.Exec(query, taskID, dependsOnTaskID)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *taskDependencyRepositoryImpl) GetDependenciesByTaskID(taskID int) ([]*model.TaskDependencyDB, error) {
	query := `SELECT id, task_id, depends_on_task_id, created_at FROM task_dependencies WHERE task_id = ? ORDER BY id ASC`
	rows, err := r.db.Query(query, taskID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*model.TaskDependencyDB
	for rows.Next() {
		var d model.TaskDependencyDB
		if err := rows.Scan(&d.ID, &d.TaskID, &d.DependsOnTaskID, &d.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, &d)
	}
	return list, rows.Err()
}

func (r *taskDependencyRepositoryImpl) GetDependentsByTaskID(taskID int) ([]*model.TaskDependencyDB, error) {
	query := `SELECT id, task_id, depends_on_task_id, created_at FROM task_dependencies WHERE depends_on_task_id = ? ORDER BY id ASC`
	rows, err := r.db.Query(query, taskID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*model.TaskDependencyDB
	for rows.Next() {
		var d model.TaskDependencyDB
		if err := rows.Scan(&d.ID, &d.TaskID, &d.DependsOnTaskID, &d.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, &d)
	}
	return list, rows.Err()
}

func (r *taskDependencyRepositoryImpl) Exists(taskID, dependsOnTaskID int) (bool, error) {
	query := `SELECT COUNT(1) FROM task_dependencies WHERE task_id = ? AND depends_on_task_id = ?`
	var count int
	err := r.db.QueryRow(query, taskID, dependsOnTaskID).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
