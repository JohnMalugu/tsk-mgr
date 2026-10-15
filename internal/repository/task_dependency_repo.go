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
