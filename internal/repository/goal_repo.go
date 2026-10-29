package repository

import (
	"database/sql"
	"errors"

	"github.com/JohnMalugu/tsk-mgr-api/internal/model"
)

var (
	ErrGoalNotFound        = errors.New("goal not found")
	ErrKeyResultNotFound   = errors.New("key result not found")
	ErrTaskAlreadyLinked   = errors.New("task is already linked to this key result")
	ErrKeyResultTaskLinkNF = errors.New("task link not found for key result")
)

// GoalRepository defines data access operations for Goals, Key Results, task links, check-ins, and metrics
type GoalRepository interface {
	CreateGoal(goal *model.Goal) (int64, error)
	GetGoalByID(id int64) (*model.Goal, error)
	ListGoals(filter model.GoalFilter) ([]*model.GoalSummary, error)
	UpdateGoal(goal *model.Goal) error
	DeleteGoal(id int64) error

	CreateKeyResult(kr *model.KeyResult) (int64, error)
	GetKeyResultByID(id int64) (*model.KeyResult, error)
	ListKeyResults(goalID int64) ([]model.KeyResult, error)
	UpdateKeyResult(kr *model.KeyResult) error
	DeleteKeyResult(id int64) error

	LinkTaskToKeyResult(keyResultID int64, taskID int) error
	UnlinkTask(keyResultID int64, taskID int) error
	ListLinkedTasks(keyResultID int64) ([]model.LinkedTaskInfo, error)

	CreateGoalCheckIn(checkIn *model.GoalCheckIn) (int64, error)
	ListGoalCheckIns(goalID int64) ([]model.GoalCheckIn, error)

	GetGoalTree(goalID int64) (*model.GoalTree, error)
	GetGoalMetrics(workspaceID int64) (*model.GoalMetrics, error)
}

type goalRepositoryImpl struct {
	db *sql.DB
}

// NewGoalRepository constructs a new SQLite-backed GoalRepository instance
func NewGoalRepository(db *sql.DB) GoalRepository {
	return &goalRepositoryImpl{db: db}
}

func (r *goalRepositoryImpl) CreateGoal(goal *model.Goal) (int64, error) {
	return 0, nil
}

func (r *goalRepositoryImpl) GetGoalByID(id int64) (*model.Goal, error) {
	return nil, nil
}

func (r *goalRepositoryImpl) ListGoals(filter model.GoalFilter) ([]*model.GoalSummary, error) {
	return nil, nil
}

func (r *goalRepositoryImpl) UpdateGoal(goal *model.Goal) error {
	return nil
}

func (r *goalRepositoryImpl) DeleteGoal(id int64) error {
	return nil
}

func (r *goalRepositoryImpl) CreateKeyResult(kr *model.KeyResult) (int64, error) {
	return 0, nil
}

func (r *goalRepositoryImpl) GetKeyResultByID(id int64) (*model.KeyResult, error) {
	return nil, nil
}

func (r *goalRepositoryImpl) ListKeyResults(goalID int64) ([]model.KeyResult, error) {
	return nil, nil
}

func (r *goalRepositoryImpl) UpdateKeyResult(kr *model.KeyResult) error {
	return nil
}

func (r *goalRepositoryImpl) DeleteKeyResult(id int64) error {
	return nil
}

func (r *goalRepositoryImpl) LinkTaskToKeyResult(keyResultID int64, taskID int) error {
	return nil
}

func (r *goalRepositoryImpl) UnlinkTask(keyResultID int64, taskID int) error {
	return nil
}

func (r *goalRepositoryImpl) ListLinkedTasks(keyResultID int64) ([]model.LinkedTaskInfo, error) {
	return nil, nil
}

func (r *goalRepositoryImpl) CreateGoalCheckIn(checkIn *model.GoalCheckIn) (int64, error) {
	return 0, nil
}

func (r *goalRepositoryImpl) ListGoalCheckIns(goalID int64) ([]model.GoalCheckIn, error) {
	return nil, nil
}

func (r *goalRepositoryImpl) GetGoalTree(goalID int64) (*model.GoalTree, error) {
	return nil, nil
}

func (r *goalRepositoryImpl) GetGoalMetrics(workspaceID int64) (*model.GoalMetrics, error) {
	return nil, nil
}
