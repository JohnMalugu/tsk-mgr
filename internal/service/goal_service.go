package service

import (
	"errors"
	"math"

	"github.com/JohnMalugu/tsk-mgr-api/internal/model"
	"github.com/JohnMalugu/tsk-mgr-api/internal/repository"
)

var (
	ErrGoalNotFound            = errors.New("goal not found")
	ErrGoalTitleRequired       = errors.New("goal title is required")
	ErrGoalWorkspaceRequired   = errors.New("workspace ID is required")
	ErrGoalTimeframeRequired   = errors.New("timeframe is required")
	ErrKeyResultNotFound       = errors.New("key result not found")
	ErrKeyResultTitleRequired  = errors.New("key result title is required")
	ErrInvalidMetricType       = errors.New("invalid key result metric type")
	ErrInvalidTargetValue      = errors.New("target value cannot equal start value")
	ErrInvalidConfidenceRating = errors.New("invalid confidence rating")
)

// Helper methods for KeyResult and Goal progress computation

func calculateKeyResultProgress(metricType string, startVal, targetVal, currentVal float64) float64 {
	if metricType == model.KeyResultMetricBoolean {
		if currentVal >= 1.0 {
			return 100.0
		}
		return 0.0
	}

	delta := targetVal - startVal
	if delta == 0 {
		if currentVal >= targetVal {
			return 100.0
		}
		return 0.0
	}

	progress := ((currentVal - startVal) / delta) * 100.0
	if progress < 0.0 {
		progress = 0.0
	}
	if progress > 100.0 {
		progress = 100.0
	}

	return math.Round(progress*10) / 10
}

func calculateGoalAggregatedProgress(krs []model.KeyResult) float64 {
	if len(krs) == 0 {
		return 0.0
	}

	var totalWeight float64
	var weightedProgress float64

	for _, kr := range krs {
		w := kr.Weight
		if w <= 0 {
			w = 1.0
		}
		totalWeight += w
		weightedProgress += kr.ProgressPercentage * w
	}

	if totalWeight <= 0 {
		return 0.0
	}

	return math.Round((weightedProgress/totalWeight)*10) / 10
}

func inferGoalStatus(progress float64, currentStatus string) string {
	if progress >= 100.0 {
		return model.GoalStatusAchieved
	}
	if currentStatus == model.GoalStatusCanceled {
		return model.GoalStatusCanceled
	}
	if currentStatus == model.GoalStatusDraft {
		return model.GoalStatusDraft
	}
	if progress >= 70.0 {
		return model.GoalStatusOnTrack
	}
	if progress >= 40.0 {
		return model.GoalStatusBehind
	}
	return model.GoalStatusAtRisk
}

// GoalService orchestrates goals, key results, task linkages, progress tracking, and check-ins
type GoalService interface {
	CreateGoal(userID string, req model.CreateGoalRequest) (*model.Goal, error)
	GetGoal(id int64) (*model.GoalDetail, error)
	ListGoals(filter model.GoalFilter) ([]*model.GoalSummary, error)
	UpdateGoal(id int64, req model.UpdateGoalRequest) (*model.Goal, error)
	DeleteGoal(id int64) error

	CreateKeyResult(goalID int64, req model.CreateKeyResultRequest) (*model.KeyResult, error)
	GetKeyResult(id int64) (*model.KeyResult, error)
	ListKeyResults(goalID int64) ([]model.KeyResult, error)
	UpdateKeyResult(id int64, req model.UpdateKeyResultRequest) (*model.KeyResult, error)
	DeleteKeyResult(id int64) error

	LinkTask(keyResultID int64, taskID int) error
	UnlinkTask(keyResultID int64, taskID int) error
	ListLinkedTasks(keyResultID int64) ([]model.LinkedTaskInfo, error)

	CreateCheckIn(userID string, goalID int64, req model.GoalCheckInRequest) (*model.GoalCheckIn, error)
	ListCheckIns(goalID int64) ([]model.GoalCheckIn, error)

	GetGoalTree(goalID int64) (*model.GoalTree, error)
	GetGoalMetrics(workspaceID int64) (*model.GoalMetrics, error)
}

type goalServiceImpl struct {
	repo     repository.GoalRepository
	taskRepo repository.TaskRepository
}

// NewGoalService instantiates a new GoalService
func NewGoalService(repo repository.GoalRepository, taskRepo repository.TaskRepository) GoalService {
	return &goalServiceImpl{
		repo:     repo,
		taskRepo: taskRepo,
	}
}

func (s *goalServiceImpl) CreateGoal(userID string, req model.CreateGoalRequest) (*model.Goal, error) {
	return nil, nil
}

func (s *goalServiceImpl) GetGoal(id int64) (*model.GoalDetail, error) {
	return nil, nil
}

func (s *goalServiceImpl) ListGoals(filter model.GoalFilter) ([]*model.GoalSummary, error) {
	return nil, nil
}

func (s *goalServiceImpl) UpdateGoal(id int64, req model.UpdateGoalRequest) (*model.Goal, error) {
	return nil, nil
}

func (s *goalServiceImpl) DeleteGoal(id int64) error {
	return nil
}

func (s *goalServiceImpl) CreateKeyResult(goalID int64, req model.CreateKeyResultRequest) (*model.KeyResult, error) {
	return nil, nil
}

func (s *goalServiceImpl) GetKeyResult(id int64) (*model.KeyResult, error) {
	return nil, nil
}

func (s *goalServiceImpl) ListKeyResults(goalID int64) ([]model.KeyResult, error) {
	return nil, nil
}

func (s *goalServiceImpl) UpdateKeyResult(id int64, req model.UpdateKeyResultRequest) (*model.KeyResult, error) {
	return nil, nil
}

func (s *goalServiceImpl) DeleteKeyResult(id int64) error {
	return nil
}

func (s *goalServiceImpl) LinkTask(keyResultID int64, taskID int) error {
	return nil
}

func (s *goalServiceImpl) UnlinkTask(keyResultID int64, taskID int) error {
	return nil
}

func (s *goalServiceImpl) ListLinkedTasks(keyResultID int64) ([]model.LinkedTaskInfo, error) {
	return nil, nil
}

func (s *goalServiceImpl) CreateCheckIn(userID string, goalID int64, req model.GoalCheckInRequest) (*model.GoalCheckIn, error) {
	return nil, nil
}

func (s *goalServiceImpl) ListCheckIns(goalID int64) ([]model.GoalCheckIn, error) {
	return nil, nil
}

func (s *goalServiceImpl) GetGoalTree(goalID int64) (*model.GoalTree, error) {
	return nil, nil
}

func (s *goalServiceImpl) GetGoalMetrics(workspaceID int64) (*model.GoalMetrics, error) {
	return nil, nil
}
