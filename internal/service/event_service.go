package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/JohnMalugu/tsk-mgr-api/internal/model"
	"github.com/JohnMalugu/tsk-mgr-api/internal/repository"
)

var (
	_ = fmt.Sprintf
	_ = strings.TrimSpace
	_ = time.Now
	_ = errors.New
)

var (
	ErrEventRecordNotFound         = repository.ErrEventRecordNotFound
	ErrCorrelatedAlertNotFound     = repository.ErrCorrelatedAlertNotFound
	ErrCorrelatedAlertCodeConflict = repository.ErrCorrelatedAlertCodeConflict
	ErrCorrelationRuleNotFound     = repository.ErrCorrelationRuleNotFound
	ErrEventSuppressionNotFound    = repository.ErrEventSuppressionNotFound
	ErrEventActionNotFound         = repository.ErrEventActionNotFound
	ErrInvalidAlertTransition      = errors.New("invalid alert status transition")
)

// EventService defines business logic and orchestration for ITIL Monitoring & Event Management
type EventService interface {
	// Raw Events
	IngestEvent(ctx context.Context, workspaceID string, req model.IngestEventRequest) (*model.EventRecord, *model.CorrelatedAlert, error)
	GetEventByID(ctx context.Context, id string) (*model.EventRecord, error)
	ListEvents(ctx context.Context, filter model.EventRecordFilter) ([]model.EventRecord, int, error)

	// Correlated Alerts
	GetAlertByID(ctx context.Context, id string) (*model.CorrelatedAlert, error)
	UpdateAlertStatus(ctx context.Context, id string, req model.UpdateAlertStatusRequest) (*model.CorrelatedAlert, error)
	ListAlerts(ctx context.Context, filter model.CorrelatedAlertFilter) ([]model.CorrelatedAlert, int, error)

	// Correlation Rules
	CreateRule(ctx context.Context, workspaceID string, req model.CreateCorrelationRuleRequest) (*model.CorrelationRule, error)
	GetRuleByID(ctx context.Context, id string) (*model.CorrelationRule, error)
	UpdateRule(ctx context.Context, id string, req model.UpdateCorrelationRuleRequest) (*model.CorrelationRule, error)
	DeleteRule(ctx context.Context, id string) error
	ListRules(ctx context.Context, filter model.CorrelationRuleFilter) ([]model.CorrelationRule, int, error)

	// Event Suppressions
	CreateSuppression(ctx context.Context, workspaceID string, req model.CreateEventSuppressionRequest) (*model.EventSuppression, error)
	GetSuppressionByID(ctx context.Context, id string) (*model.EventSuppression, error)
	DeleteSuppression(ctx context.Context, id string) error
	ListSuppressions(ctx context.Context, filter model.EventSuppressionFilter) ([]model.EventSuppression, int, error)

	// Event Actions
	TriggerAction(ctx context.Context, workspaceID string, req model.TriggerEventActionRequest) (*model.EventAction, error)
	ListActions(ctx context.Context, filter model.EventActionFilter) ([]model.EventAction, int, error)

	// Posture Analytics
	GetEventAnalytics(ctx context.Context, workspaceID string) (*model.EventAnalytics, error)
}

type eventService struct {
	repo repository.EventRepository
}

// NewEventService constructs an ITIL Monitoring & Event Management service
func NewEventService(repo repository.EventRepository) EventService {
	return &eventService{repo: repo}
}

func severityRank(sev string) int {
	switch sev {
	case model.EventSeverityCritical:
		return 5
	case model.EventSeverityHigh:
		return 4
	case model.EventSeverityMedium:
		return 3
	case model.EventSeverityLow:
		return 2
	case model.EventSeverityInfo:
		return 1
	default:
		return 0
	}
}

func severityToPriority(sev string) string {
	switch sev {
	case model.EventSeverityCritical:
		return model.AlertPriorityP1
	case model.EventSeverityHigh:
		return model.AlertPriorityP2
	case model.EventSeverityMedium:
		return model.AlertPriorityP3
	default:
		return model.AlertPriorityP4
	}
}

// IngestEvent processes an incoming event, applies suppression windows, deduplicates, and clusters into correlated alerts
func (s *eventService) IngestEvent(ctx context.Context, workspaceID string, req model.IngestEventRequest) (*model.EventRecord, *model.CorrelatedAlert, error) {
	if req.WorkspaceID == "" {
		req.WorkspaceID = workspaceID
	}
	if err := req.Validate(); err != nil {
		return nil, nil, err
	}

	if req.Severity == "" {
		req.Severity = model.EventSeverityInfo
	}
	if req.EventType == "" {
		req.EventType = model.EventTypeInformational
	}

	eventTime := time.Now().UTC()
	if req.Timestamp != nil && !req.Timestamp.IsZero() {
		eventTime = req.Timestamp.UTC()
	}

	fingerprint := strings.TrimSpace(req.Fingerprint)
	if fingerprint == "" {
		fingerprint = fmt.Sprintf("%s:%s:%s", req.Source, req.Resource, req.Severity)
	}

	dedupKey := strings.TrimSpace(req.DeduplicationKey)
	if dedupKey == "" {
		dedupKey = fingerprint
	}

	record := &model.EventRecord{
		WorkspaceID:      req.WorkspaceID,
		Source:           req.Source,
		EventType:        req.EventType,
		Severity:         req.Severity,
		CIID:             req.CIID,
		Resource:         req.Resource,
		Message:          req.Message,
		Fingerprint:      fingerprint,
		DeduplicationKey: dedupKey,
		RawPayload:       req.RawPayload,
		Timestamp:        eventTime,
		CreatedAt:        time.Now().UTC(),
	}

	if err := s.repo.IngestEvent(ctx, record); err != nil {
		return nil, nil, err
	}

	// Check active suppression / blackout windows
	activeTrue := true
	suppressions, _, _ := s.repo.ListSuppressions(ctx, model.EventSuppressionFilter{
		WorkspaceID: req.WorkspaceID,
		IsActive:    &activeTrue,
	})

	isSuppressed := false
	for _, sup := range suppressions {
		if (record.Timestamp.Equal(sup.StartTime) || record.Timestamp.After(sup.StartTime)) &&
			(record.Timestamp.Equal(sup.EndTime) || record.Timestamp.Before(sup.EndTime)) {
			if sup.CIID != "" && sup.CIID == record.CIID {
				isSuppressed = true
				break
			}
			if sup.ResourcePattern != "" && strings.Contains(strings.ToLower(record.Resource), strings.ToLower(sup.ResourcePattern)) {
				isSuppressed = true
				break
			}
		}
	}

	// Alert correlation and deduplication clustering
	existingAlert, err := s.repo.GetAlertByFingerprint(ctx, req.WorkspaceID, record.Fingerprint)
	if err == nil && existingAlert != nil {
		// Existing active alert: aggregate and update
		existingAlert.EventCount++
		existingAlert.LastSeenAt = record.Timestamp
		if severityRank(record.Severity) > severityRank(existingAlert.Severity) {
			existingAlert.Severity = record.Severity
			existingAlert.Priority = severityToPriority(record.Severity)
		}
		if isSuppressed && existingAlert.Status == model.AlertStatusOpen {
			existingAlert.Status = model.AlertStatusSuppressed
		}

		if err := s.repo.UpdateAlert(ctx, existingAlert); err != nil {
			return nil, nil, err
		}
		return record, existingAlert, nil
	}

	// New alert clustering
	initialStatus := model.AlertStatusOpen
	if isSuppressed {
		initialStatus = model.AlertStatusSuppressed
	}

	alertCode := fmt.Sprintf("ALT-%05d", time.Now().UnixNano()%100000)
	title := fmt.Sprintf("[%s] %s alert on %s", strings.ToUpper(record.Severity), record.Source, record.Resource)

	newAlert := &model.CorrelatedAlert{
		WorkspaceID: req.WorkspaceID,
		AlertCode:   alertCode,
		Title:       title,
		Description: record.Message,
		Status:      initialStatus,
		Severity:    record.Severity,
		Priority:    severityToPriority(record.Severity),
		CIID:        record.CIID,
		Resource:    record.Resource,
		Fingerprint: record.Fingerprint,
		EventCount:  1,
		FirstSeenAt: record.Timestamp,
		LastSeenAt:  record.Timestamp,
	}

	// Evaluate correlation rules
	rules, _, _ := s.repo.ListRules(ctx, model.CorrelationRuleFilter{
		WorkspaceID: req.WorkspaceID,
		IsActive:    &activeTrue,
	})
	for _, rule := range rules {
		matchedSource := rule.SourcePattern == "" || rule.SourcePattern == "*" || strings.EqualFold(rule.SourcePattern, record.Source)
		matchedSev := rule.SeverityPattern == "" || rule.SeverityPattern == "*" || strings.EqualFold(rule.SeverityPattern, record.Severity)
		if matchedSource && matchedSev {
			newAlert.RuleID = rule.ID
			if rule.ActionType == model.RuleActionEscalateIncident {
				newAlert.Status = model.AlertStatusEscalated
				newAlert.IncidentID = fmt.Sprintf("INC-%05d", time.Now().UnixNano()%100000)
			}
			break
		}
	}

	if err := s.repo.CreateAlert(ctx, newAlert); err != nil {
		return nil, nil, err
	}

	// If alert was escalated by rule, record auto-action
	if newAlert.RuleID != "" {
		_ = s.repo.CreateAction(ctx, &model.EventAction{
			WorkspaceID: req.WorkspaceID,
			AlertID:     newAlert.ID,
			ActionType:  model.ActionTriggerIncident,
			Status:      model.ActionStatusExecuted,
			OutputLogs:  fmt.Sprintf("Auto-triggered incident %s from correlation rule %s", newAlert.IncidentID, newAlert.RuleID),
		})
	}

	return record, newAlert, nil
}

// GetEventByID retrieves an event record by ID
func (s *eventService) GetEventByID(ctx context.Context, id string) (*model.EventRecord, error) {
	if strings.TrimSpace(id) == "" {
		return nil, ErrEventRecordNotFound
	}
	return s.repo.GetEventByID(ctx, id)
}

// ListEvents lists event records according to query filters
func (s *eventService) ListEvents(ctx context.Context, filter model.EventRecordFilter) ([]model.EventRecord, int, error) {
	return s.repo.ListEvents(ctx, filter)
}

// GetAlertByID stub
func (s *eventService) GetAlertByID(ctx context.Context, id string) (*model.CorrelatedAlert, error) {
	return nil, errors.New("not implemented")
}

// UpdateAlertStatus stub
func (s *eventService) UpdateAlertStatus(ctx context.Context, id string, req model.UpdateAlertStatusRequest) (*model.CorrelatedAlert, error) {
	return nil, errors.New("not implemented")
}

// ListAlerts stub
func (s *eventService) ListAlerts(ctx context.Context, filter model.CorrelatedAlertFilter) ([]model.CorrelatedAlert, int, error) {
	return nil, 0, errors.New("not implemented")
}

// CreateRule stub
func (s *eventService) CreateRule(ctx context.Context, workspaceID string, req model.CreateCorrelationRuleRequest) (*model.CorrelationRule, error) {
	return nil, errors.New("not implemented")
}

// GetRuleByID stub
func (s *eventService) GetRuleByID(ctx context.Context, id string) (*model.CorrelationRule, error) {
	return nil, errors.New("not implemented")
}

// UpdateRule stub
func (s *eventService) UpdateRule(ctx context.Context, id string, req model.UpdateCorrelationRuleRequest) (*model.CorrelationRule, error) {
	return nil, errors.New("not implemented")
}

// DeleteRule stub
func (s *eventService) DeleteRule(ctx context.Context, id string) error {
	return errors.New("not implemented")
}

// ListRules stub
func (s *eventService) ListRules(ctx context.Context, filter model.CorrelationRuleFilter) ([]model.CorrelationRule, int, error) {
	return nil, 0, errors.New("not implemented")
}

// CreateSuppression stub
func (s *eventService) CreateSuppression(ctx context.Context, workspaceID string, req model.CreateEventSuppressionRequest) (*model.EventSuppression, error) {
	return nil, errors.New("not implemented")
}

// GetSuppressionByID stub
func (s *eventService) GetSuppressionByID(ctx context.Context, id string) (*model.EventSuppression, error) {
	return nil, errors.New("not implemented")
}

// DeleteSuppression stub
func (s *eventService) DeleteSuppression(ctx context.Context, id string) error {
	return errors.New("not implemented")
}

// ListSuppressions stub
func (s *eventService) ListSuppressions(ctx context.Context, filter model.EventSuppressionFilter) ([]model.EventSuppression, int, error) {
	return nil, 0, errors.New("not implemented")
}

// TriggerAction stub
func (s *eventService) TriggerAction(ctx context.Context, workspaceID string, req model.TriggerEventActionRequest) (*model.EventAction, error) {
	return nil, errors.New("not implemented")
}

// ListActions stub
func (s *eventService) ListActions(ctx context.Context, filter model.EventActionFilter) ([]model.EventAction, int, error) {
	return nil, 0, errors.New("not implemented")
}

// GetEventAnalytics stub
func (s *eventService) GetEventAnalytics(ctx context.Context, workspaceID string) (*model.EventAnalytics, error) {
	return nil, errors.New("not implemented")
}
