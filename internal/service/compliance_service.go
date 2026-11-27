package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/JohnMalugu/tsk-mgr-api/internal/model"
	"github.com/JohnMalugu/tsk-mgr-api/internal/repository"
)

var (
	_ = uuid.New
	_ = fmt.Sprintf
	_ = strings.TrimSpace
	_ = time.Now
	_ = errors.New
)

var (
	ErrGRCFrameworkNotFound     = repository.ErrGRCFrameworkNotFound
	ErrGRCFrameworkCodeConflict = repository.ErrGRCFrameworkCodeConflict
	ErrGRCControlNotFound       = repository.ErrGRCControlNotFound
	ErrGRCControlCodeConflict   = repository.ErrGRCControlCodeConflict
	ErrGRCEvidenceNotFound      = repository.ErrGRCEvidenceNotFound
	ErrGRCAssessmentNotFound    = repository.ErrGRCAssessmentNotFound
	ErrGRCFindingNotFound       = repository.ErrGRCFindingNotFound
)

// ComplianceService defines business logic and orchestration for ITIL GRC & Compliance Management
type ComplianceService interface {
	// Frameworks
	CreateFramework(ctx context.Context, workspaceID string, req model.CreateGRCFrameworkRequest) (*model.ComplianceFramework, error)
	GetFrameworkByID(ctx context.Context, id string) (*model.ComplianceFramework, error)
	GetFrameworkByCode(ctx context.Context, workspaceID, code string) (*model.ComplianceFramework, error)
	UpdateFramework(ctx context.Context, id string, req model.UpdateGRCFrameworkRequest) (*model.ComplianceFramework, error)
	DeleteFramework(ctx context.Context, id string) error
	ListFrameworks(ctx context.Context, filter model.GRCFrameworkFilter) ([]model.ComplianceFramework, int, error)

	// Controls
	CreateControl(ctx context.Context, workspaceID string, req model.CreateGRCControlRequest) (*model.ComplianceControl, error)
	GetControlByID(ctx context.Context, id string) (*model.ComplianceControl, error)
	GetControlByCode(ctx context.Context, workspaceID, code string) (*model.ComplianceControl, error)
	UpdateControl(ctx context.Context, id string, req model.UpdateGRCControlRequest) (*model.ComplianceControl, error)
	DeleteControl(ctx context.Context, id string) error
	ListControls(ctx context.Context, filter model.GRCControlFilter) ([]model.ComplianceControl, int, error)

	// Evidence
	SubmitEvidence(ctx context.Context, workspaceID string, req model.SubmitGRCEvidenceRequest) (*model.ComplianceEvidence, error)
	GetEvidenceByID(ctx context.Context, id string) (*model.ComplianceEvidence, error)
	ReviewEvidence(ctx context.Context, id string, req model.ReviewGRCEvidenceRequest) (*model.ComplianceEvidence, error)
	DeleteEvidence(ctx context.Context, id string) error
	ListEvidence(ctx context.Context, filter model.GRCEvidenceFilter) ([]model.ComplianceEvidence, int, error)

	// Assessments
	CreateAssessment(ctx context.Context, workspaceID string, req model.CreateGRCAssessmentRequest) (*model.ComplianceAssessment, error)
	GetAssessmentByID(ctx context.Context, id string) (*model.ComplianceAssessment, error)
	CompleteAssessment(ctx context.Context, id string, req model.CompleteGRCAssessmentRequest) (*model.ComplianceAssessment, error)
	DeleteAssessment(ctx context.Context, id string) error
	ListAssessments(ctx context.Context, filter model.GRCAssessmentFilter) ([]model.ComplianceAssessment, int, error)

	// Findings
	CreateFinding(ctx context.Context, workspaceID string, req model.CreateGRCFindingRequest) (*model.ComplianceFinding, error)
	GetFindingByID(ctx context.Context, id string) (*model.ComplianceFinding, error)
	UpdateFinding(ctx context.Context, id string, req model.UpdateGRCFindingRequest) (*model.ComplianceFinding, error)
	DeleteFinding(ctx context.Context, id string) error
	ListFindings(ctx context.Context, filter model.GRCFindingFilter) ([]model.ComplianceFinding, int, error)

	// Analytics
	GetComplianceAnalytics(ctx context.Context, workspaceID string) (*model.ComplianceAnalytics, error)
}

type complianceService struct {
	repo repository.ComplianceRepository
}

// NewComplianceService constructs an ITIL GRC & Compliance Management service
func NewComplianceService(repo repository.ComplianceRepository) ComplianceService {
	return &complianceService{repo: repo}
}

// CreateFramework validates and registers a new compliance framework
func (s *complianceService) CreateFramework(ctx context.Context, workspaceID string, req model.CreateGRCFrameworkRequest) (*model.ComplianceFramework, error) {
	if req.WorkspaceID == "" {
		req.WorkspaceID = workspaceID
	}
	if err := req.Validate(); err != nil {
		return nil, err
	}

	code := strings.TrimSpace(req.FrameworkCode)
	if code == "" {
		code = fmt.Sprintf("FW-%s", strings.ToUpper(uuid.New().String()[:8]))
	}
	std := req.StandardType
	if std == "" {
		std = model.GRCStandardCustom
	}
	ver := strings.TrimSpace(req.Version)
	if ver == "" {
		ver = "1.0"
	}

	fw := &model.ComplianceFramework{
		WorkspaceID:   req.WorkspaceID,
		FrameworkCode: code,
		Name:          req.Name,
		StandardType:  std,
		Version:       ver,
		Description:   req.Description,
		OwnerID:       req.OwnerID,
		Status:        model.GRCFrameworkStatusActive,
	}

	if err := s.repo.CreateFramework(ctx, fw); err != nil {
		return nil, err
	}
	return fw, nil
}

// GetFrameworkByID retrieves a compliance framework by ID
func (s *complianceService) GetFrameworkByID(ctx context.Context, id string) (*model.ComplianceFramework, error) {
	if strings.TrimSpace(id) == "" {
		return nil, ErrGRCFrameworkNotFound
	}
	return s.repo.GetFrameworkByID(ctx, id)
}

// GetFrameworkByCode retrieves a compliance framework by code within a workspace
func (s *complianceService) GetFrameworkByCode(ctx context.Context, workspaceID, code string) (*model.ComplianceFramework, error) {
	if strings.TrimSpace(code) == "" {
		return nil, ErrGRCFrameworkNotFound
	}
	return s.repo.GetFrameworkByCode(ctx, workspaceID, code)
}

// UpdateFramework modifies an existing compliance framework
func (s *complianceService) UpdateFramework(ctx context.Context, id string, req model.UpdateGRCFrameworkRequest) (*model.ComplianceFramework, error) {
	if strings.TrimSpace(id) == "" {
		return nil, ErrGRCFrameworkNotFound
	}
	if err := req.Validate(); err != nil {
		return nil, err
	}

	fw, err := s.repo.GetFrameworkByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.Name != nil {
		fw.Name = *req.Name
	}
	if req.Version != nil {
		fw.Version = *req.Version
	}
	if req.Description != nil {
		fw.Description = *req.Description
	}
	if req.OwnerID != nil {
		fw.OwnerID = *req.OwnerID
	}
	if req.Status != nil {
		fw.Status = *req.Status
	}

	if err := s.repo.UpdateFramework(ctx, fw); err != nil {
		return nil, err
	}
	return fw, nil
}

// DeleteFramework removes a compliance framework by ID
func (s *complianceService) DeleteFramework(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return ErrGRCFrameworkNotFound
	}
	return s.repo.DeleteFramework(ctx, id)
}

// ListFrameworks queries compliance frameworks matching filter criteria
func (s *complianceService) ListFrameworks(ctx context.Context, filter model.GRCFrameworkFilter) ([]model.ComplianceFramework, int, error) {
	return s.repo.ListFrameworks(ctx, filter)
}

// CreateControl validates and registers a new compliance control under a framework
func (s *complianceService) CreateControl(ctx context.Context, workspaceID string, req model.CreateGRCControlRequest) (*model.ComplianceControl, error) {
	if req.WorkspaceID == "" {
		req.WorkspaceID = workspaceID
	}
	if err := req.Validate(); err != nil {
		return nil, err
	}

	if _, err := s.repo.GetFrameworkByID(ctx, req.FrameworkID); err != nil {
		return nil, err
	}

	code := strings.TrimSpace(req.ControlCode)
	if code == "" {
		code = fmt.Sprintf("CTRL-%s", strings.ToUpper(uuid.New().String()[:8]))
	}
	domain := req.Domain
	if domain == "" {
		domain = model.GRCDomainAccessControl
	}
	implStatus := req.ImplementationStatus
	if implStatus == "" {
		implStatus = model.GRCControlNotImplemented
	}
	autoLevel := req.AutomationLevel
	if autoLevel == "" {
		autoLevel = model.GRCAutomationManual
	}

	ctrl := &model.ComplianceControl{
		WorkspaceID:          req.WorkspaceID,
		FrameworkID:          req.FrameworkID,
		ControlCode:          code,
		Title:                req.Title,
		Description:          req.Description,
		Domain:               domain,
		ImplementationStatus: implStatus,
		AutomationLevel:      autoLevel,
		OwnerID:              req.OwnerID,
		CIID:                 req.CIID,
	}

	if err := s.repo.CreateControl(ctx, ctrl); err != nil {
		return nil, err
	}
	return ctrl, nil
}

// GetControlByID retrieves a compliance control by ID
func (s *complianceService) GetControlByID(ctx context.Context, id string) (*model.ComplianceControl, error) {
	if strings.TrimSpace(id) == "" {
		return nil, ErrGRCControlNotFound
	}
	return s.repo.GetControlByID(ctx, id)
}

// GetControlByCode retrieves a compliance control by code within a workspace
func (s *complianceService) GetControlByCode(ctx context.Context, workspaceID, code string) (*model.ComplianceControl, error) {
	if strings.TrimSpace(code) == "" {
		return nil, ErrGRCControlNotFound
	}
	return s.repo.GetControlByCode(ctx, workspaceID, code)
}

// UpdateControl modifies an existing compliance control
func (s *complianceService) UpdateControl(ctx context.Context, id string, req model.UpdateGRCControlRequest) (*model.ComplianceControl, error) {
	if strings.TrimSpace(id) == "" {
		return nil, ErrGRCControlNotFound
	}
	if err := req.Validate(); err != nil {
		return nil, err
	}

	ctrl, err := s.repo.GetControlByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.Title != nil {
		ctrl.Title = *req.Title
	}
	if req.Description != nil {
		ctrl.Description = *req.Description
	}
	if req.Domain != nil {
		ctrl.Domain = *req.Domain
	}
	if req.ImplementationStatus != nil {
		ctrl.ImplementationStatus = *req.ImplementationStatus
	}
	if req.AutomationLevel != nil {
		ctrl.AutomationLevel = *req.AutomationLevel
	}
	if req.OwnerID != nil {
		ctrl.OwnerID = *req.OwnerID
	}
	if req.CIID != nil {
		ctrl.CIID = *req.CIID
	}

	if err := s.repo.UpdateControl(ctx, ctrl); err != nil {
		return nil, err
	}
	return ctrl, nil
}

// DeleteControl deletes a compliance control by ID
func (s *complianceService) DeleteControl(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return ErrGRCControlNotFound
	}
	return s.repo.DeleteControl(ctx, id)
}

// ListControls queries compliance controls matching filter criteria
func (s *complianceService) ListControls(ctx context.Context, filter model.GRCControlFilter) ([]model.ComplianceControl, int, error) {
	return s.repo.ListControls(ctx, filter)
}

// SubmitEvidence attaches an audit evidence artifact to a compliance control
func (s *complianceService) SubmitEvidence(ctx context.Context, workspaceID string, req model.SubmitGRCEvidenceRequest) (*model.ComplianceEvidence, error) {
	if req.WorkspaceID == "" {
		req.WorkspaceID = workspaceID
	}
	if err := req.Validate(); err != nil {
		return nil, err
	}

	if _, err := s.repo.GetControlByID(ctx, req.ControlID); err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	validFrom := now
	if req.ValidFrom != nil && !req.ValidFrom.IsZero() {
		validFrom = req.ValidFrom.UTC()
	}
	evType := req.EvidenceType
	if evType == "" {
		evType = model.GRCEvidencePolicyDoc
	}
	collectedBy := strings.TrimSpace(req.CollectedBy)
	if collectedBy == "" {
		collectedBy = "user_1"
	}

	ev := &model.ComplianceEvidence{
		WorkspaceID:  req.WorkspaceID,
		ControlID:    req.ControlID,
		Title:        req.Title,
		EvidenceType: evType,
		ArtifactURL:  req.ArtifactURL,
		CollectedBy:  collectedBy,
		ValidFrom:    validFrom,
		ValidUntil:   req.ValidUntil,
		ReviewStatus: model.GRCEvidenceStatusPending,
		CreatedAt:    now,
	}

	if err := s.repo.CreateEvidence(ctx, ev); err != nil {
		return nil, err
	}
	return ev, nil
}

// GetEvidenceByID retrieves an audit evidence record by ID
func (s *complianceService) GetEvidenceByID(ctx context.Context, id string) (*model.ComplianceEvidence, error) {
	if strings.TrimSpace(id) == "" {
		return nil, ErrGRCEvidenceNotFound
	}
	return s.repo.GetEvidenceByID(ctx, id)
}

// ReviewEvidence records an auditor's approval or rejection of collected evidence
func (s *complianceService) ReviewEvidence(ctx context.Context, id string, req model.ReviewGRCEvidenceRequest) (*model.ComplianceEvidence, error) {
	if strings.TrimSpace(id) == "" {
		return nil, ErrGRCEvidenceNotFound
	}
	if err := req.Validate(); err != nil {
		return nil, err
	}

	ev, err := s.repo.GetEvidenceByID(ctx, id)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	ev.ReviewStatus = req.ReviewStatus
	ev.ReviewedBy = req.ReviewedBy
	ev.ReviewedAt = &now

	if err := s.repo.UpdateEvidence(ctx, ev); err != nil {
		return nil, err
	}

	// Update parent control's LastTestedAt when evidence is approved
	if req.ReviewStatus == model.GRCEvidenceStatusApproved {
		if ctrl, err := s.repo.GetControlByID(ctx, ev.ControlID); err == nil {
			ctrl.LastTestedAt = &now
			_ = s.repo.UpdateControl(ctx, ctrl)
		}
	}

	return ev, nil
}

// DeleteEvidence deletes an audit evidence artifact by ID
func (s *complianceService) DeleteEvidence(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return ErrGRCEvidenceNotFound
	}
	return s.repo.DeleteEvidence(ctx, id)
}

// ListEvidence queries audit evidence records matching filter criteria
func (s *complianceService) ListEvidence(ctx context.Context, filter model.GRCEvidenceFilter) ([]model.ComplianceEvidence, int, error) {
	return s.repo.ListEvidence(ctx, filter)
}

// Assessments stubs
func (s *complianceService) CreateAssessment(ctx context.Context, workspaceID string, req model.CreateGRCAssessmentRequest) (*model.ComplianceAssessment, error) {
	return nil, errors.New("not implemented")
}
func (s *complianceService) GetAssessmentByID(ctx context.Context, id string) (*model.ComplianceAssessment, error) {
	return nil, errors.New("not implemented")
}
func (s *complianceService) CompleteAssessment(ctx context.Context, id string, req model.CompleteGRCAssessmentRequest) (*model.ComplianceAssessment, error) {
	return nil, errors.New("not implemented")
}
func (s *complianceService) DeleteAssessment(ctx context.Context, id string) error {
	return errors.New("not implemented")
}
func (s *complianceService) ListAssessments(ctx context.Context, filter model.GRCAssessmentFilter) ([]model.ComplianceAssessment, int, error) {
	return nil, 0, errors.New("not implemented")
}

// Findings stubs
func (s *complianceService) CreateFinding(ctx context.Context, workspaceID string, req model.CreateGRCFindingRequest) (*model.ComplianceFinding, error) {
	return nil, errors.New("not implemented")
}
func (s *complianceService) GetFindingByID(ctx context.Context, id string) (*model.ComplianceFinding, error) {
	return nil, errors.New("not implemented")
}
func (s *complianceService) UpdateFinding(ctx context.Context, id string, req model.UpdateGRCFindingRequest) (*model.ComplianceFinding, error) {
	return nil, errors.New("not implemented")
}
func (s *complianceService) DeleteFinding(ctx context.Context, id string) error {
	return errors.New("not implemented")
}
func (s *complianceService) ListFindings(ctx context.Context, filter model.GRCFindingFilter) ([]model.ComplianceFinding, int, error) {
	return nil, 0, errors.New("not implemented")
}

// Analytics stub
func (s *complianceService) GetComplianceAnalytics(ctx context.Context, workspaceID string) (*model.ComplianceAnalytics, error) {
	return nil, errors.New("not implemented")
}
