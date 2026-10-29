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
	if req.Title == "" {
		return nil, ErrGoalTitleRequired
	}
	if req.WorkspaceID <= 0 {
		return nil, ErrGoalWorkspaceRequired
	}
	if req.Timeframe == "" {
		return nil, ErrGoalTimeframeRequired
	}

	if req.ParentGoalID != nil && *req.ParentGoalID > 0 {
		_, err := s.repo.GetGoalByID(*req.ParentGoalID)
		if err != nil {
			if errors.Is(err, repository.ErrGoalNotFound) {
				return nil, ErrGoalNotFound
			}
			return nil, err
		}
	}

	goal := &model.Goal{
		WorkspaceID:        req.WorkspaceID,
		OwnerUserID:        userID,
		ParentGoalID:       req.ParentGoalID,
		Title:              req.Title,
		Description:        req.Description,
		Timeframe:          req.Timeframe,
		StartDate:          req.StartDate,
		EndDate:            req.EndDate,
		Status:             model.GoalStatusDraft,
		ProgressPercentage: 0.0,
	}

	id, err := s.repo.CreateGoal(goal)
	if err != nil {
		return nil, err
	}
	goal.ID = id
	return goal, nil
}

func (s *goalServiceImpl) GetGoal(id int64) (*model.GoalDetail, error) {
	g, err := s.repo.GetGoalByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrGoalNotFound) {
			return nil, ErrGoalNotFound
		}
		return nil, err
	}

	krs, err := s.repo.ListKeyResults(id)
	if err != nil {
		return nil, err
	}

	checkIns, _ := s.repo.ListGoalCheckIns(id)
	var latestCheckIn *model.GoalCheckIn
	if len(checkIns) > 0 {
		latestCheckIn = &checkIns[0]
	}

	return &model.GoalDetail{
		Goal:          g,
		KeyResults:    krs,
		LatestCheckIn: latestCheckIn,
	}, nil
}

func (s *goalServiceImpl) ListGoals(filter model.GoalFilter) ([]*model.GoalSummary, error) {
	return s.repo.ListGoals(filter)
}

func (s *goalServiceImpl) UpdateGoal(id int64, req model.UpdateGoalRequest) (*model.Goal, error) {
	goal, err := s.repo.GetGoalByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrGoalNotFound) {
			return nil, ErrGoalNotFound
		}
		return nil, err
	}

	if req.Title != nil {
		if *req.Title == "" {
			return nil, ErrGoalTitleRequired
		}
		goal.Title = *req.Title
	}
	if req.Description != nil {
		goal.Description = *req.Description
	}
	if req.Timeframe != nil {
		if *req.Timeframe == "" {
			return nil, ErrGoalTimeframeRequired
		}
		goal.Timeframe = *req.Timeframe
	}
	if req.OwnerUserID != nil && *req.OwnerUserID != "" {
		goal.OwnerUserID = *req.OwnerUserID
	}
	if req.ParentGoalID != nil {
		if *req.ParentGoalID > 0 && *req.ParentGoalID != id {
			if _, err := s.repo.GetGoalByID(*req.ParentGoalID); err != nil {
				return nil, ErrGoalNotFound
			}
			goal.ParentGoalID = req.ParentGoalID
		} else if *req.ParentGoalID <= 0 {
			goal.ParentGoalID = nil
		}
	}
	if req.StartDate != nil {
		goal.StartDate = req.StartDate
	}
	if req.EndDate != nil {
		goal.EndDate = req.EndDate
	}
	if req.Status != nil && *req.Status != "" {
		goal.Status = *req.Status
	}
	if req.ProgressPercentage != nil {
		prog := *req.ProgressPercentage
		if prog < 0 {
			prog = 0
		}
		if prog > 100 {
			prog = 100
		}
		goal.ProgressPercentage = prog
	}

	if err := s.repo.UpdateGoal(goal); err != nil {
		return nil, err
	}
	return goal, nil
}

func (s *goalServiceImpl) DeleteGoal(id int64) error {
	err := s.repo.DeleteGoal(id)
	if err != nil && errors.Is(err, repository.ErrGoalNotFound) {
		return ErrGoalNotFound
	}
	return err
}

func (s *goalServiceImpl) recalculateGoalProgress(goalID int64) error {
	goal, err := s.repo.GetGoalByID(goalID)
	if err != nil {
		return err
	}

	krs, err := s.repo.ListKeyResults(goalID)
	if err != nil {
		return err
	}

	aggProgress := calculateGoalAggregatedProgress(krs)
	goal.ProgressPercentage = aggProgress
	goal.Status = inferGoalStatus(aggProgress, goal.Status)

	return s.repo.UpdateGoal(goal)
}

func (s *goalServiceImpl) CreateKeyResult(goalID int64, req model.CreateKeyResultRequest) (*model.KeyResult, error) {
	if _, err := s.repo.GetGoalByID(goalID); err != nil {
		if errors.Is(err, repository.ErrGoalNotFound) {
			return nil, ErrGoalNotFound
		}
		return nil, err
	}

	if req.Title == "" {
		return nil, ErrKeyResultTitleRequired
	}

	mType := req.MetricType
	if mType == "" {
		mType = model.KeyResultMetricNumerical
	}
	if mType != model.KeyResultMetricNumerical &&
		mType != model.KeyResultMetricPercentage &&
		mType != model.KeyResultMetricCurrency &&
		mType != model.KeyResultMetricBoolean {
		return nil, ErrInvalidMetricType
	}

	if mType != model.KeyResultMetricBoolean && req.StartValue == req.TargetValue {
		return nil, ErrInvalidTargetValue
	}

	var curVal float64
	if req.CurrentValue != nil {
		curVal = *req.CurrentValue
	} else {
		curVal = req.StartValue
	}

	weight := 1.0
	if req.Weight != nil && *req.Weight > 0 {
		weight = *req.Weight
	}

	progress := calculateKeyResultProgress(mType, req.StartValue, req.TargetValue, curVal)

	kr := &model.KeyResult{
		GoalID:             goalID,
		Title:              req.Title,
		Description:        req.Description,
		MetricType:         mType,
		Unit:               req.Unit,
		StartValue:         req.StartValue,
		TargetValue:        req.TargetValue,
		CurrentValue:       curVal,
		Weight:             weight,
		ProgressPercentage: progress,
	}

	id, err := s.repo.CreateKeyResult(kr)
	if err != nil {
		return nil, err
	}
	kr.ID = id

	_ = s.recalculateGoalProgress(goalID)

	return kr, nil
}

func (s *goalServiceImpl) GetKeyResult(id int64) (*model.KeyResult, error) {
	kr, err := s.repo.GetKeyResultByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrKeyResultNotFound) {
			return nil, ErrKeyResultNotFound
		}
		return nil, err
	}
	return kr, nil
}

func (s *goalServiceImpl) ListKeyResults(goalID int64) ([]model.KeyResult, error) {
	if _, err := s.repo.GetGoalByID(goalID); err != nil {
		if errors.Is(err, repository.ErrGoalNotFound) {
			return nil, ErrGoalNotFound
		}
		return nil, err
	}
	return s.repo.ListKeyResults(goalID)
}

func (s *goalServiceImpl) UpdateKeyResult(id int64, req model.UpdateKeyResultRequest) (*model.KeyResult, error) {
	kr, err := s.repo.GetKeyResultByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrKeyResultNotFound) {
			return nil, ErrKeyResultNotFound
		}
		return nil, err
	}

	if req.Title != nil {
		if *req.Title == "" {
			return nil, ErrKeyResultTitleRequired
		}
		kr.Title = *req.Title
	}
	if req.Description != nil {
		kr.Description = *req.Description
	}
	if req.MetricType != nil {
		mType := *req.MetricType
		if mType != model.KeyResultMetricNumerical &&
			mType != model.KeyResultMetricPercentage &&
			mType != model.KeyResultMetricCurrency &&
			mType != model.KeyResultMetricBoolean {
			return nil, ErrInvalidMetricType
		}
		kr.MetricType = mType
	}
	if req.Unit != nil {
		kr.Unit = *req.Unit
	}
	if req.StartValue != nil {
		kr.StartValue = *req.StartValue
	}
	if req.TargetValue != nil {
		kr.TargetValue = *req.TargetValue
	}
	if req.CurrentValue != nil {
		kr.CurrentValue = *req.CurrentValue
	}
	if req.Weight != nil && *req.Weight > 0 {
		kr.Weight = *req.Weight
	}

	if kr.MetricType != model.KeyResultMetricBoolean && kr.StartValue == kr.TargetValue {
		return nil, ErrInvalidTargetValue
	}

	kr.ProgressPercentage = calculateKeyResultProgress(kr.MetricType, kr.StartValue, kr.TargetValue, kr.CurrentValue)

	if err := s.repo.UpdateKeyResult(kr); err != nil {
		return nil, err
	}

	_ = s.recalculateGoalProgress(kr.GoalID)

	return kr, nil
}

func (s *goalServiceImpl) DeleteKeyResult(id int64) error {
	kr, err := s.repo.GetKeyResultByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrKeyResultNotFound) {
			return ErrKeyResultNotFound
		}
		return err
	}

	if err := s.repo.DeleteKeyResult(id); err != nil {
		return err
	}

	_ = s.recalculateGoalProgress(kr.GoalID)
	return nil
}

func (s *goalServiceImpl) syncKeyResultTaskProgress(keyResultID int64) error {
	tasks, err := s.repo.ListLinkedTasks(keyResultID)
	if err != nil {
		return err
	}
	if len(tasks) == 0 {
		return nil
	}

	completedCount := 0
	for _, t := range tasks {
		if t.Completed || t.Status == model.TaskStatusCompleted {
			completedCount++
		}
	}

	kr, err := s.repo.GetKeyResultByID(keyResultID)
	if err != nil {
		return err
	}

	kr.CurrentValue = float64(completedCount)
	kr.TargetValue = float64(len(tasks))
	kr.StartValue = 0.0
	kr.Unit = "tasks"
	kr.ProgressPercentage = math.Round((float64(completedCount)/float64(len(tasks)))*1000) / 10

	if err := s.repo.UpdateKeyResult(kr); err != nil {
		return err
	}

	return s.recalculateGoalProgress(kr.GoalID)
}

func (s *goalServiceImpl) LinkTask(keyResultID int64, taskID int) error {
	kr, err := s.repo.GetKeyResultByID(keyResultID)
	if err != nil {
		if errors.Is(err, repository.ErrKeyResultNotFound) {
			return ErrKeyResultNotFound
		}
		return err
	}

	task, err := s.taskRepo.GetByID(taskID)
	if err != nil {
		return err
	}
	if task == nil {
		return ErrTaskNotFound
	}

	if err := s.repo.LinkTaskToKeyResult(keyResultID, taskID); err != nil {
		return err
	}

	_ = s.syncKeyResultTaskProgress(kr.ID)
	return nil
}

func (s *goalServiceImpl) UnlinkTask(keyResultID int64, taskID int) error {
	kr, err := s.repo.GetKeyResultByID(keyResultID)
	if err != nil {
		if errors.Is(err, repository.ErrKeyResultNotFound) {
			return ErrKeyResultNotFound
		}
		return err
	}

	if err := s.repo.UnlinkTask(keyResultID, taskID); err != nil {
		return err
	}

	_ = s.syncKeyResultTaskProgress(kr.ID)
	return nil
}

func (s *goalServiceImpl) ListLinkedTasks(keyResultID int64) ([]model.LinkedTaskInfo, error) {
	if _, err := s.repo.GetKeyResultByID(keyResultID); err != nil {
		if errors.Is(err, repository.ErrKeyResultNotFound) {
			return nil, ErrKeyResultNotFound
		}
		return nil, err
	}
	return s.repo.ListLinkedTasks(keyResultID)
}

func (s *goalServiceImpl) CreateCheckIn(userID string, goalID int64, req model.GoalCheckInRequest) (*model.GoalCheckIn, error) {
	goal, err := s.repo.GetGoalByID(goalID)
	if err != nil {
		if errors.Is(err, repository.ErrGoalNotFound) {
			return nil, ErrGoalNotFound
		}
		return nil, err
	}

	conf := req.ConfidenceRating
	if conf == "" {
		conf = model.GoalConfidenceMedium
	} else if conf != model.GoalConfidenceHigh && conf != model.GoalConfidenceMedium && conf != model.GoalConfidenceLow {
		return nil, ErrInvalidConfidenceRating
	}

	status := req.Status
	if status == "" {
		status = model.GoalStatusOnTrack
	}

	prog := goal.ProgressPercentage
	if req.ProgressPercentage != nil {
		p := *req.ProgressPercentage
		if p < 0 {
			p = 0
		}
		if p > 100 {
			p = 100
		}
		prog = p
	}

	checkIn := &model.GoalCheckIn{
		GoalID:             goalID,
		UserID:             userID,
		ConfidenceRating:   conf,
		Status:             status,
		ProgressPercentage: prog,
		Note:               req.Note,
	}

	id, err := s.repo.CreateGoalCheckIn(checkIn)
	if err != nil {
		return nil, err
	}
	checkIn.ID = id

	// Update goal status and progress from check-in
	goal.Status = status
	goal.ProgressPercentage = prog
	_ = s.repo.UpdateGoal(goal)

	return checkIn, nil
}

func (s *goalServiceImpl) ListCheckIns(goalID int64) ([]model.GoalCheckIn, error) {
	if _, err := s.repo.GetGoalByID(goalID); err != nil {
		if errors.Is(err, repository.ErrGoalNotFound) {
			return nil, ErrGoalNotFound
		}
		return nil, err
	}
	return s.repo.ListGoalCheckIns(goalID)
}

func (s *goalServiceImpl) GetGoalTree(goalID int64) (*model.GoalTree, error) {
	tree, err := s.repo.GetGoalTree(goalID)
	if err != nil {
		if errors.Is(err, repository.ErrGoalNotFound) {
			return nil, ErrGoalNotFound
		}
		return nil, err
	}
	return tree, nil
}

func (s *goalServiceImpl) GetGoalMetrics(workspaceID int64) (*model.GoalMetrics, error) {
	return s.repo.GetGoalMetrics(workspaceID)
}
