package service

import (
	"errors"

	"github.com/JohnMalugu/tsk-mgr-api/internal/model"
	"github.com/JohnMalugu/tsk-mgr-api/internal/repository"
)

var (
	ErrDependencySelfRefDB  = errors.New("a task cannot depend on itself")
	ErrDependencyExistsDB   = errors.New("dependency already exists")
	ErrDependencyNotFoundDB = errors.New("dependency not found")
	ErrDependencyCycleDB    = errors.New("adding dependency would create a circular dependency")
	ErrPrerequisiteNotMet   = errors.New("task cannot be completed because prerequisites are incomplete")
)

type TaskDependencyService interface {
	AddDependency(taskID, dependsOnTaskID int, userID string) (*model.TaskDependencyInfo, error)
	RemoveDependency(taskID, dependsOnTaskID int, userID string) error
	GetDependencies(taskID int, userID string) ([]*model.TaskDependencyInfo, error)
	GetDependents(taskID int, userID string) ([]*model.TaskDependencyInfo, error)
	GetBlockedTasks(userID string) ([]*model.Task, error)
	GetReadyTasks(userID string) ([]*model.Task, error)
	GetDAGSummary(taskID int, userID string) (*model.TaskDAGSummary, error)
}

type taskDependencyServiceImpl struct {
	depRepo  repository.TaskDependencyRepository
	taskRepo repository.TaskRepository
}

func NewTaskDependencyService(depRepo repository.TaskDependencyRepository, taskRepo repository.TaskRepository) TaskDependencyService {
	return &taskDependencyServiceImpl{
		depRepo:  depRepo,
		taskRepo: taskRepo,
	}
}

// wouldCreateCycle checks if adding a dependency where taskID depends on dependsOnTaskID would introduce a cycle.
func (s *taskDependencyServiceImpl) wouldCreateCycle(taskID, dependsOnTaskID int, userID string) (bool, error) {
	allDeps, err := s.depRepo.GetAllDependenciesForUser(userID)
	if err != nil {
		return false, err
	}

	adj := make(map[int][]int)
	for _, d := range allDeps {
		adj[d.TaskID] = append(adj[d.TaskID], d.DependsOnTaskID)
	}

	visited := make(map[int]bool)
	var dfs func(current int) bool
	dfs = func(current int) bool {
		if current == taskID {
			return true
		}
		if visited[current] {
			return false
		}
		visited[current] = true
		for _, next := range adj[current] {
			if dfs(next) {
				return true
			}
		}
		return false
	}

	return dfs(dependsOnTaskID), nil
}
