package service

import (
	"errors"
	"strings"
	"time"

	"github.com/JohnMalugu/tsk-mgr-api/internal/model"
	"github.com/JohnMalugu/tsk-mgr-api/internal/repository"
)

var (
	ErrAuditLogNotFoundDB      = errors.New("audit log not found")
	ErrInvalidRetentionPeriod = errors.New("retention period must be at least 30 days")
	ErrAuditAlertNotFoundDB    = errors.New("audit alert not found")
	ErrInvalidAlertStatus      = errors.New("invalid alert resolution status")
	ErrExportFormatUnsupported = errors.New("unsupported audit export format: only 'json' and 'csv' are supported")
)

// AuditService coordinates enterprise security audit event recording, compliance policies, anomaly alerting, and exports
type AuditService interface {
	// Event logging & querying
	RecordEvent(actorID string, req model.CreateAuditLogRequest) (*model.AuditLog, error)
	GetEvent(id int64) (*model.AuditLog, error)
	ListEvents(filter model.AuditLogFilter) ([]*model.AuditLog, int64, error)

	// Retention & Purging
	GetRetentionPolicy(workspaceID int64) (*model.AuditRetentionPolicy, error)
	ConfigureRetentionPolicy(workspaceID int64, req model.RetentionPolicyRequest) (*model.AuditRetentionPolicy, error)
	PurgeExpiredEvents(workspaceID int64) (int64, error)

	// Anomaly alerting
	ListAlerts(filter model.AuditAlertFilter) ([]*model.AuditAlert, error)
	GetAlert(id int64) (*model.AuditAlert, error)
	ResolveAlert(id int64, userID string, req model.ResolveAlertRequest) (*model.AuditAlert, error)

	// Compliance & Exports
	GenerateExport(userID string, req model.ExportAuditLogsRequest) (*model.AuditExport, []byte, error)
	ListExports(workspaceID int64) ([]*model.AuditExport, error)
	GetAuditMetrics(workspaceID int64) (*model.AuditMetrics, error)
	GetComplianceSummary(workspaceID int64) (*model.ComplianceSummary, error)
}

type auditServiceImpl struct {
	repo          repository.AuditRepository
	workspaceRepo repository.WorkspaceRepository
	userRepo      repository.UserRepository
}

// NewAuditService creates a new instance of AuditService
func NewAuditService(
	repo repository.AuditRepository,
	workspaceRepo repository.WorkspaceRepository,
	userRepo repository.UserRepository,
) AuditService {
	return &auditServiceImpl{
		repo:          repo,
		workspaceRepo: workspaceRepo,
		userRepo:      userRepo,
	}
}

func (s *auditServiceImpl) RecordEvent(actorID string, req model.CreateAuditLogRequest) (*model.AuditLog, error) {
	if req.WorkspaceID <= 0 {
		return nil, ErrWorkspaceRequired
	}

	ws, err := s.workspaceRepo.GetByID(int(req.WorkspaceID))
	if err != nil || ws == nil {
		return nil, ErrWorkspaceNotFound
	}

	action := strings.TrimSpace(req.Action)
	if action == "" {
		return nil, model.ErrAuditActionRequired
	}

	sev := strings.ToLower(strings.TrimSpace(req.Severity))
	if sev == "" {
		sev = model.AuditSeverityInfo
	}
	if err := model.ValidateAuditSeverity(sev); err != nil {
		return nil, err
	}

	cat := strings.ToLower(strings.TrimSpace(req.Category))
	if cat == "" {
		cat = model.AuditCategorySystem
	}

	actor := strings.TrimSpace(actorID)
	if actor == "" {
		actor = "system"
	}

	log := &model.AuditLog{
		WorkspaceID: req.WorkspaceID,
		ActorID:     actor,
		Action:      action,
		Category:    cat,
		Severity:    sev,
		EntityType:  strings.TrimSpace(req.EntityType),
		EntityID:    strings.TrimSpace(req.EntityID),
		EntityName:  strings.TrimSpace(req.EntityName),
		Description: strings.TrimSpace(req.Description),
		BeforeState: strings.TrimSpace(req.BeforeState),
		AfterState:  strings.TrimSpace(req.AfterState),
		Metadata:    strings.TrimSpace(req.Metadata),
		IPAddress:   strings.TrimSpace(req.IPAddress),
		UserAgent:   strings.TrimSpace(req.UserAgent),
	}

	id, err := s.repo.CreateAuditLog(log)
	if err != nil {
		return nil, err
	}

	// Automated anomaly detection: if critical severity or failed login, trigger alert check
	if sev == model.AuditSeverityCritical || action == model.AuditActionFailedLogin {
		_ = s.evaluateAnomalyAlert(log)
	}

	return s.repo.GetAuditLogByID(id)
}

func (s *auditServiceImpl) evaluateAnomalyAlert(log *model.AuditLog) error {
	title := "Security Anomaly Detected"
	desc := log.Description
	if log.Action == model.AuditActionFailedLogin {
		title = "Repeated or Suspicious Failed Login Attempt"
		desc = "Failed authentication event observed from IP: " + log.IPAddress
	} else if log.Severity == model.AuditSeverityCritical {
		title = "Critical Security Event: " + log.Action
	}

	alert := &model.AuditAlert{
		WorkspaceID:    log.WorkspaceID,
		Title:          title,
		Description:    desc,
		Severity:       log.Severity,
		Status:         model.AuditAlertStatusOpen,
		TriggerEventID: &log.ID,
	}
	_, err := s.repo.CreateAlert(alert)
	return err
}

func (s *auditServiceImpl) GetEvent(id int64) (*model.AuditLog, error) {
	if id <= 0 {
		return nil, ErrAuditLogNotFoundDB
	}
	log, err := s.repo.GetAuditLogByID(id)
	if err != nil || log == nil {
		return nil, ErrAuditLogNotFoundDB
	}
	return log, nil
}

func (s *auditServiceImpl) ListEvents(filter model.AuditLogFilter) ([]*model.AuditLog, int64, error) {
	if filter.WorkspaceID <= 0 {
		return nil, 0, ErrWorkspaceRequired
	}

	ws, err := s.workspaceRepo.GetByID(int(filter.WorkspaceID))
	if err != nil || ws == nil {
		return nil, 0, ErrWorkspaceNotFound
	}

	if filter.Severity != nil && strings.TrimSpace(*filter.Severity) != "" {
		if err := model.ValidateAuditSeverity(*filter.Severity); err != nil {
			return nil, 0, err
		}
	}

	return s.repo.ListAuditLogs(filter)
}

func (s *auditServiceImpl) GetRetentionPolicy(workspaceID int64) (*model.AuditRetentionPolicy, error) {
	if workspaceID <= 0 {
		return nil, ErrWorkspaceRequired
	}
	ws, err := s.workspaceRepo.GetByID(int(workspaceID))
	if err != nil || ws == nil {
		return nil, ErrWorkspaceNotFound
	}

	policy, err := s.repo.GetRetentionPolicy(workspaceID)
	if err != nil {
		// Return standard default retention policy
		return &model.AuditRetentionPolicy{
			WorkspaceID:      workspaceID,
			RetentionDays:    365,
			Frameworks:       "soc2,gdpr",
			AutoPurgeEnabled: false,
		}, nil
	}
	return policy, nil
}

func (s *auditServiceImpl) ConfigureRetentionPolicy(workspaceID int64, req model.RetentionPolicyRequest) (*model.AuditRetentionPolicy, error) {
	if workspaceID <= 0 {
		return nil, ErrWorkspaceRequired
	}
	ws, err := s.workspaceRepo.GetByID(int(workspaceID))
	if err != nil || ws == nil {
		return nil, ErrWorkspaceNotFound
	}

	if req.RetentionDays < 30 {
		return nil, ErrInvalidRetentionPeriod
	}

	frameworks := "soc2,gdpr"
	if len(req.Frameworks) > 0 {
		frameworks = strings.Join(req.Frameworks, ",")
	}

	policy := &model.AuditRetentionPolicy{
		WorkspaceID:      workspaceID,
		RetentionDays:    req.RetentionDays,
		Frameworks:       frameworks,
		AutoPurgeEnabled: req.AutoPurgeEnabled,
		UpdatedAt:        time.Now(),
	}

	if err := s.repo.UpsertRetentionPolicy(policy); err != nil {
		return nil, err
	}

	if req.AutoPurgeEnabled {
		_, _ = s.PurgeExpiredEvents(workspaceID)
	}

	return s.GetRetentionPolicy(workspaceID)
}

func (s *auditServiceImpl) PurgeExpiredEvents(workspaceID int64) (int64, error) {
	if workspaceID <= 0 {
		return 0, ErrWorkspaceRequired
	}
	policy, err := s.GetRetentionPolicy(workspaceID)
	if err != nil {
		return 0, err
	}

	cutoff := time.Now().AddDate(0, 0, -policy.RetentionDays)
	return s.repo.PurgeExpiredAuditLogs(workspaceID, cutoff)
}

func (s *auditServiceImpl) ListAlerts(filter model.AuditAlertFilter) ([]*model.AuditAlert, error) {
	if filter.WorkspaceID <= 0 {
		return nil, ErrWorkspaceRequired
	}
	ws, err := s.workspaceRepo.GetByID(int(filter.WorkspaceID))
	if err != nil || ws == nil {
		return nil, ErrWorkspaceNotFound
	}

	return s.repo.ListAlerts(filter)
}

func (s *auditServiceImpl) GetAlert(id int64) (*model.AuditAlert, error) {
	if id <= 0 {
		return nil, ErrAuditAlertNotFoundDB
	}
	alert, err := s.repo.GetAlertByID(id)
	if err != nil || alert == nil {
		return nil, ErrAuditAlertNotFoundDB
	}
	return alert, nil
}

func (s *auditServiceImpl) ResolveAlert(id int64, userID string, req model.ResolveAlertRequest) (*model.AuditAlert, error) {
	if id <= 0 {
		return nil, ErrAuditAlertNotFoundDB
	}
	alert, err := s.repo.GetAlertByID(id)
	if err != nil || alert == nil {
		return nil, ErrAuditAlertNotFoundDB
	}

	target := strings.ToLower(strings.TrimSpace(req.Status))
	switch target {
	case model.AuditAlertStatusAcknowledged, model.AuditAlertStatusResolved, model.AuditAlertStatusDismissed:
		// valid
	default:
		return nil, ErrInvalidAlertStatus
	}

	if err := s.repo.ResolveAlert(id, target, strings.TrimSpace(req.ResolutionNotes), userID); err != nil {
		return nil, err
	}
	return s.GetAlert(id)
}

func (s *auditServiceImpl) GenerateExport(userID string, req model.ExportAuditLogsRequest) (*model.AuditExport, []byte, error) {
	return nil, nil, nil
}

func (s *auditServiceImpl) ListExports(workspaceID int64) ([]*model.AuditExport, error) {
	return nil, nil
}

func (s *auditServiceImpl) GetAuditMetrics(workspaceID int64) (*model.AuditMetrics, error) {
	return nil, nil
}

func (s *auditServiceImpl) GetComplianceSummary(workspaceID int64) (*model.ComplianceSummary, error) {
	return nil, nil
}

// Ensure unused import doesn't fail
var _ = time.Now
