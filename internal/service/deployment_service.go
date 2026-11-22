package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/JohnMalugu/tsk-mgr-api/internal/model"
	"github.com/JohnMalugu/tsk-mgr-api/internal/repository"
)

var (
	_ = strings.TrimSpace
	_ = time.Now
	_ = errors.New
)

var (
	ErrDeploymentPipelineNotFound     = repository.ErrDeploymentPipelineNotFound
	ErrDeploymentPipelineCodeConflict = repository.ErrDeploymentPipelineCodeConflict
	ErrDeploymentRunNotFound          = repository.ErrDeploymentRunNotFound
	ErrDeploymentGateNotFound         = repository.ErrDeploymentGateNotFound
	ErrDeploymentStepNotFound         = repository.ErrDeploymentStepNotFound
	ErrRollbackPolicyNotFound         = repository.ErrRollbackPolicyNotFound
	ErrInvalidPipelineStrategy        = errors.New("invalid deployment pipeline strategy")
	ErrInvalidRunStatusTransition     = errors.New("invalid deployment run status transition")
)

// DeploymentService defines business logic and orchestration for ITIL Release & Deployment Management
type DeploymentService interface {
	// Pipelines
	CreatePipeline(ctx context.Context, workspaceID string, req model.CreateDeploymentPipelineRequest) (*model.DeploymentPipeline, error)
	GetPipelineByID(ctx context.Context, id string) (*model.DeploymentPipeline, error)
	GetPipelineByCode(ctx context.Context, workspaceID, code string) (*model.DeploymentPipeline, error)
	UpdatePipeline(ctx context.Context, id string, req model.UpdateDeploymentPipelineRequest) (*model.DeploymentPipeline, error)
	DeletePipeline(ctx context.Context, id string) error
	ListPipelines(ctx context.Context, filter model.DeploymentPipelineFilter) ([]model.DeploymentPipeline, int, error)

	// Runs
	CreateRun(ctx context.Context, workspaceID string, req model.CreateDeploymentRunRequest) (*model.DeploymentRun, error)
	GetRunByID(ctx context.Context, id string) (*model.DeploymentRun, error)
	UpdateRunStatus(ctx context.Context, id string, req model.UpdateDeploymentRunStatusRequest) (*model.DeploymentRun, error)
	ListRuns(ctx context.Context, filter model.DeploymentRunFilter) ([]model.DeploymentRun, int, error)

	// Gates
	CreateGate(ctx context.Context, workspaceID string, req model.CreateDeploymentGateRequest) (*model.DeploymentGate, error)
	GetGateByID(ctx context.Context, id string) (*model.DeploymentGate, error)
	EvaluateGate(ctx context.Context, id string, req model.EvaluateDeploymentGateRequest) (*model.DeploymentGate, error)
	ListGates(ctx context.Context, filter model.DeploymentGateFilter) ([]model.DeploymentGate, int, error)

	// Steps
	CreateStep(ctx context.Context, workspaceID string, req model.CreateDeploymentStepRequest) (*model.DeploymentStep, error)
	GetStepByID(ctx context.Context, id string) (*model.DeploymentStep, error)
	PromoteStep(ctx context.Context, id string, req model.PromoteDeploymentStepRequest) (*model.DeploymentStep, error)
	ListSteps(ctx context.Context, filter model.DeploymentStepFilter) ([]model.DeploymentStep, int, error)

	// Rollback Policies
	CreateRollbackPolicy(ctx context.Context, workspaceID string, req model.CreateRollbackPolicyRequest) (*model.RollbackPolicy, error)
	GetRollbackPolicyByID(ctx context.Context, id string) (*model.RollbackPolicy, error)
	ExecuteRollback(ctx context.Context, id string, req model.ExecuteRollbackRequest) (*model.RollbackPolicy, error)
	ListRollbackPolicies(ctx context.Context, filter model.RollbackPolicyFilter) ([]model.RollbackPolicy, int, error)

	// Posture Analytics
	GetDeploymentAnalytics(ctx context.Context, workspaceID string) (*model.DeploymentAnalytics, error)
}

type deploymentService struct {
	repo repository.DeploymentRepository
}

// NewDeploymentService constructs an ITIL Release & Deployment Management service
func NewDeploymentService(repo repository.DeploymentRepository) DeploymentService {
	return &deploymentService{repo: repo}
}

// Pipelines implementation
func (s *deploymentService) CreatePipeline(ctx context.Context, workspaceID string, req model.CreateDeploymentPipelineRequest) (*model.DeploymentPipeline, error) {
	if req.WorkspaceID == "" {
		req.WorkspaceID = workspaceID
	}
	if err := req.Validate(); err != nil {
		return nil, err
	}

	pipeline := &model.DeploymentPipeline{
		WorkspaceID:         req.WorkspaceID,
		PipelineCode:        strings.TrimSpace(req.PipelineCode),
		Name:                strings.TrimSpace(req.Name),
		Description:         strings.TrimSpace(req.Description),
		TargetEnvironment:   req.TargetEnvironment,
		DeploymentStrategy:  req.DeploymentStrategy,
		AutoRollbackEnabled: req.AutoRollbackEnabled,
		LinkedCIID:          strings.TrimSpace(req.LinkedCIID),
		Status:              model.PipelineStatusActive,
		CreatedByUserID:     strings.TrimSpace(req.CreatedByUserID),
	}

	if err := s.repo.CreatePipeline(ctx, pipeline); err != nil {
		return nil, err
	}
	return pipeline, nil
}

func (s *deploymentService) GetPipelineByID(ctx context.Context, id string) (*model.DeploymentPipeline, error) {
	return s.repo.GetPipelineByID(ctx, id)
}

func (s *deploymentService) GetPipelineByCode(ctx context.Context, workspaceID, code string) (*model.DeploymentPipeline, error) {
	return s.repo.GetPipelineByCode(ctx, workspaceID, code)
}

func (s *deploymentService) UpdatePipeline(ctx context.Context, id string, req model.UpdateDeploymentPipelineRequest) (*model.DeploymentPipeline, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	existing, err := s.repo.GetPipelineByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.Name != nil {
		existing.Name = strings.TrimSpace(*req.Name)
	}
	if req.Description != nil {
		existing.Description = strings.TrimSpace(*req.Description)
	}
	if req.TargetEnvironment != nil {
		existing.TargetEnvironment = *req.TargetEnvironment
	}
	if req.DeploymentStrategy != nil {
		existing.DeploymentStrategy = *req.DeploymentStrategy
	}
	if req.AutoRollbackEnabled != nil {
		existing.AutoRollbackEnabled = *req.AutoRollbackEnabled
	}
	if req.LinkedCIID != nil {
		existing.LinkedCIID = strings.TrimSpace(*req.LinkedCIID)
	}
	if req.Status != nil {
		existing.Status = *req.Status
	}

	if err := s.repo.UpdatePipeline(ctx, existing); err != nil {
		return nil, err
	}
	return existing, nil
}

func (s *deploymentService) DeletePipeline(ctx context.Context, id string) error {
	return s.repo.DeletePipeline(ctx, id)
}

func (s *deploymentService) ListPipelines(ctx context.Context, filter model.DeploymentPipelineFilter) ([]model.DeploymentPipeline, int, error) {
	return s.repo.ListPipelines(ctx, filter)
}
// Runs implementation
func (s *deploymentService) CreateRun(ctx context.Context, workspaceID string, req model.CreateDeploymentRunRequest) (*model.DeploymentRun, error) {
	if req.WorkspaceID == "" {
		req.WorkspaceID = workspaceID
	}
	if err := req.Validate(); err != nil {
		return nil, err
	}

	pipeline, err := s.repo.GetPipelineByID(ctx, req.PipelineID)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	run := &model.DeploymentRun{
		WorkspaceID:      req.WorkspaceID,
		PipelineID:       pipeline.ID,
		ReleaseVersion:   strings.TrimSpace(req.ReleaseVersion),
		Environment:      req.Environment,
		TriggerType:      req.TriggerType,
		Status:           model.RunStatusPending,
		StartedAt:        now,
		DeployedByUserID: strings.TrimSpace(req.DeployedByUserID),
		CommitHash:       strings.TrimSpace(req.CommitHash),
		Changelog:        strings.TrimSpace(req.Changelog),
	}

	if err := s.repo.CreateRun(ctx, run); err != nil {
		return nil, err
	}
	return run, nil
}

func (s *deploymentService) GetRunByID(ctx context.Context, id string) (*model.DeploymentRun, error) {
	return s.repo.GetRunByID(ctx, id)
}

func (s *deploymentService) UpdateRunStatus(ctx context.Context, id string, req model.UpdateDeploymentRunStatusRequest) (*model.DeploymentRun, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	run, err := s.repo.GetRunByID(ctx, id)
	if err != nil {
		return nil, err
	}

	run.Status = req.Status
	now := time.Now().UTC()
	switch req.Status {
	case model.RunStatusSuccessful, model.RunStatusFailed, model.RunStatusRolledBack, model.RunStatusCancelled:
		run.FinishedAt = &now
		if req.DurationSeconds > 0 {
			run.DurationSeconds = req.DurationSeconds
		} else if !run.StartedAt.IsZero() {
			run.DurationSeconds = now.Sub(run.StartedAt).Seconds()
		}
	case model.RunStatusInProgress:
		if run.StartedAt.IsZero() {
			run.StartedAt = now
		}
	}

	if req.RollbackReason != "" {
		run.RollbackReason = strings.TrimSpace(req.RollbackReason)
	}

	if err := s.repo.UpdateRun(ctx, run); err != nil {
		return nil, err
	}
	return run, nil
}

func (s *deploymentService) ListRuns(ctx context.Context, filter model.DeploymentRunFilter) ([]model.DeploymentRun, int, error) {
	return s.repo.ListRuns(ctx, filter)
}
// Gates implementation
func (s *deploymentService) CreateGate(ctx context.Context, workspaceID string, req model.CreateDeploymentGateRequest) (*model.DeploymentGate, error) {
	if req.WorkspaceID == "" {
		req.WorkspaceID = workspaceID
	}
	if err := req.Validate(); err != nil {
		return nil, err
	}

	run, err := s.repo.GetRunByID(ctx, req.RunID)
	if err != nil {
		return nil, err
	}

	gate := &model.DeploymentGate{
		WorkspaceID: req.WorkspaceID,
		RunID:       run.ID,
		GateType:    req.GateType,
		GateName:    strings.TrimSpace(req.GateName),
		Status:      model.GateStatusPending,
	}

	if err := s.repo.CreateGate(ctx, gate); err != nil {
		return nil, err
	}
	return gate, nil
}

func (s *deploymentService) GetGateByID(ctx context.Context, id string) (*model.DeploymentGate, error) {
	return s.repo.GetGateByID(ctx, id)
}

func (s *deploymentService) EvaluateGate(ctx context.Context, id string, req model.EvaluateDeploymentGateRequest) (*model.DeploymentGate, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	gate, err := s.repo.GetGateByID(ctx, id)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	gate.Status = req.Status
	gate.EvaluatorID = strings.TrimSpace(req.EvaluatorID)
	gate.EvaluatedAt = &now
	gate.WaiverReason = strings.TrimSpace(req.WaiverReason)
	gate.EvidenceDetails = strings.TrimSpace(req.EvidenceDetails)

	if err := s.repo.UpdateGate(ctx, gate); err != nil {
		return nil, err
	}
	return gate, nil
}

func (s *deploymentService) ListGates(ctx context.Context, filter model.DeploymentGateFilter) ([]model.DeploymentGate, int, error) {
	return s.repo.ListGates(ctx, filter)
}
// Steps implementation
func (s *deploymentService) CreateStep(ctx context.Context, workspaceID string, req model.CreateDeploymentStepRequest) (*model.DeploymentStep, error) {
	if req.WorkspaceID == "" {
		req.WorkspaceID = workspaceID
	}
	if err := req.Validate(); err != nil {
		return nil, err
	}

	run, err := s.repo.GetRunByID(ctx, req.RunID)
	if err != nil {
		return nil, err
	}

	step := &model.DeploymentStep{
		WorkspaceID:        req.WorkspaceID,
		RunID:              run.ID,
		StepOrder:          req.StepOrder,
		StepName:           strings.TrimSpace(req.StepName),
		TrafficPercentage:  req.TrafficPercentage,
		SoakTimeMinutes:    req.SoakTimeMinutes,
		ErrorRateThreshold: req.ErrorRateThreshold,
		Status:             model.StepStatusPending,
	}

	if err := s.repo.CreateStep(ctx, step); err != nil {
		return nil, err
	}
	return step, nil
}

func (s *deploymentService) GetStepByID(ctx context.Context, id string) (*model.DeploymentStep, error) {
	return s.repo.GetStepByID(ctx, id)
}

func (s *deploymentService) PromoteStep(ctx context.Context, id string, req model.PromoteDeploymentStepRequest) (*model.DeploymentStep, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	step, err := s.repo.GetStepByID(ctx, id)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	step.Status = req.Status
	step.ActualErrorRate = req.ActualErrorRate

	if req.Status == model.StepStatusSoaking && step.StartedAt == nil {
		step.StartedAt = &now
	}
	if req.Status == model.StepStatusPromoted || req.Status == model.StepStatusAborted {
		step.CompletedAt = &now
	}

	if err := s.repo.UpdateStep(ctx, step); err != nil {
		return nil, err
	}
	return step, nil
}

func (s *deploymentService) ListSteps(ctx context.Context, filter model.DeploymentStepFilter) ([]model.DeploymentStep, int, error) {
	return s.repo.ListSteps(ctx, filter)
}

// Rollback Policies implementation
func (s *deploymentService) CreateRollbackPolicy(ctx context.Context, workspaceID string, req model.CreateRollbackPolicyRequest) (*model.RollbackPolicy, error) {
	if req.WorkspaceID == "" {
		req.WorkspaceID = workspaceID
	}
	if err := req.Validate(); err != nil {
		return nil, err
	}

	pipeline, err := s.repo.GetPipelineByID(ctx, req.PipelineID)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	policy := &model.RollbackPolicy{
		WorkspaceID:           req.WorkspaceID,
		PipelineID:            pipeline.ID,
		RunID:                 strings.TrimSpace(req.RunID),
		TriggerCondition:      req.TriggerCondition,
		ThresholdValue:        req.ThresholdValue,
		TargetRollbackVersion: strings.TrimSpace(req.TargetRollbackVersion),
		ExecutionStatus:       model.RollbackStatusTriggered,
		TriggeredAt:           &now,
	}

	if err := s.repo.CreateRollbackPolicy(ctx, policy); err != nil {
		return nil, err
	}
	return policy, nil
}

func (s *deploymentService) GetRollbackPolicyByID(ctx context.Context, id string) (*model.RollbackPolicy, error) {
	return s.repo.GetRollbackPolicyByID(ctx, id)
}

func (s *deploymentService) ExecuteRollback(ctx context.Context, id string, req model.ExecuteRollbackRequest) (*model.RollbackPolicy, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	policy, err := s.repo.GetRollbackPolicyByID(ctx, id)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	policy.ExecutionStatus = req.ExecutionStatus
	if req.RollbackLogs != "" {
		policy.RollbackLogs = strings.TrimSpace(req.RollbackLogs)
	}
	if req.ExecutionStatus == model.RollbackStatusCompleted || req.ExecutionStatus == model.RollbackStatusFailed {
		policy.CompletedAt = &now
	}

	if err := s.repo.UpdateRollbackPolicy(ctx, policy); err != nil {
		return nil, err
	}
	return policy, nil
}

func (s *deploymentService) ListRollbackPolicies(ctx context.Context, filter model.RollbackPolicyFilter) ([]model.RollbackPolicy, int, error) {
	return s.repo.ListRollbackPolicies(ctx, filter)
}
func (s *deploymentService) GetDeploymentAnalytics(ctx context.Context, workspaceID string) (*model.DeploymentAnalytics, error) {
	return nil, nil
}
