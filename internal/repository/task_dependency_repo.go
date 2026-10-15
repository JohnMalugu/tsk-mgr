package repository

import (
	"database/sql"

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
