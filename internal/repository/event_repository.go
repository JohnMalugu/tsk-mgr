package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/JohnMalugu/tsk-mgr-api/internal/model"
)

var (
	_ = uuid.New
	_ = fmt.Sprintf
	_ = strings.Join
	_ = time.Now
	_ = errors.New
)

var (
	ErrEventRecordNotFound         = errors.New("event record not found")
	ErrCorrelatedAlertNotFound     = errors.New("correlated alert not found")
	ErrCorrelatedAlertCodeConflict = errors.New("correlated alert with this code already exists in workspace")
	ErrCorrelationRuleNotFound     = errors.New("correlation rule not found")
	ErrEventSuppressionNotFound    = errors.New("event suppression not found")
	ErrEventActionNotFound         = errors.New("event action not found")
)

// EventRepository defines data access methods for ITIL Monitoring & Event Management
type EventRepository interface {
	// Raw Events
	IngestEvent(ctx context.Context, e *model.EventRecord) error
	GetEventByID(ctx context.Context, id string) (*model.EventRecord, error)
	ListEvents(ctx context.Context, filter model.EventRecordFilter) ([]model.EventRecord, int, error)

	// Correlated Alerts
	CreateAlert(ctx context.Context, a *model.CorrelatedAlert) error
	GetAlertByID(ctx context.Context, id string) (*model.CorrelatedAlert, error)
	GetAlertByFingerprint(ctx context.Context, workspaceID, fingerprint string) (*model.CorrelatedAlert, error)
	UpdateAlert(ctx context.Context, a *model.CorrelatedAlert) error
	ListAlerts(ctx context.Context, filter model.CorrelatedAlertFilter) ([]model.CorrelatedAlert, int, error)

	// Correlation Rules
	CreateRule(ctx context.Context, r *model.CorrelationRule) error
	GetRuleByID(ctx context.Context, id string) (*model.CorrelationRule, error)
	UpdateRule(ctx context.Context, r *model.CorrelationRule) error
	DeleteRule(ctx context.Context, id string) error
	ListRules(ctx context.Context, filter model.CorrelationRuleFilter) ([]model.CorrelationRule, int, error)

	// Event Suppressions
	CreateSuppression(ctx context.Context, s *model.EventSuppression) error
	GetSuppressionByID(ctx context.Context, id string) (*model.EventSuppression, error)
	UpdateSuppression(ctx context.Context, s *model.EventSuppression) error
	DeleteSuppression(ctx context.Context, id string) error
	ListSuppressions(ctx context.Context, filter model.EventSuppressionFilter) ([]model.EventSuppression, int, error)

	// Event Actions
	CreateAction(ctx context.Context, a *model.EventAction) error
	ListActions(ctx context.Context, filter model.EventActionFilter) ([]model.EventAction, int, error)

	// Posture Analytics
	GetEventAnalytics(ctx context.Context, workspaceID string) (*model.EventAnalytics, error)
}

type sqlEventRepository struct {
	db *sql.DB
}

// NewEventRepository initializes a SQLite event repository
func NewEventRepository(db *sql.DB) EventRepository {
	return &sqlEventRepository{db: db}
}

// IngestEvent inserts an ingested raw telemetry event into the database
func (r *sqlEventRepository) IngestEvent(ctx context.Context, e *model.EventRecord) error {
	if e.ID == "" {
		e.ID = uuid.New().String()
	}
	now := time.Now().UTC()
	if e.Timestamp.IsZero() {
		e.Timestamp = now
	}
	if e.CreatedAt.IsZero() {
		e.CreatedAt = now
	}

	query := `
		INSERT INTO event_records (
			id, workspace_id, source, event_type, severity, ci_id,
			resource, message, fingerprint, deduplication_key, raw_payload,
			timestamp, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := r.db.ExecContext(ctx, query,
		e.ID,
		e.WorkspaceID,
		e.Source,
		e.EventType,
		e.Severity,
		e.CIID,
		e.Resource,
		e.Message,
		e.Fingerprint,
		e.DeduplicationKey,
		e.RawPayload,
		e.Timestamp,
		e.CreatedAt,
	)
	return err
}

// GetEventByID retrieves an event record by ID
func (r *sqlEventRepository) GetEventByID(ctx context.Context, id string) (*model.EventRecord, error) {
	query := `
		SELECT id, workspace_id, source, event_type, severity, ci_id,
		       resource, message, fingerprint, deduplication_key, raw_payload,
		       timestamp, created_at
		FROM event_records
		WHERE id = ?`

	var e model.EventRecord
	var ciID, rawPayload sql.NullString
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&e.ID,
		&e.WorkspaceID,
		&e.Source,
		&e.EventType,
		&e.Severity,
		&ciID,
		&e.Resource,
		&e.Message,
		&e.Fingerprint,
		&e.DeduplicationKey,
		&rawPayload,
		&e.Timestamp,
		&e.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrEventRecordNotFound
		}
		return nil, err
	}
	if ciID.Valid {
		e.CIID = ciID.String
	}
	if rawPayload.Valid {
		e.RawPayload = rawPayload.String
	}
	return &e, nil
}

// ListEvents queries event records matching filter criteria
func (r *sqlEventRepository) ListEvents(ctx context.Context, filter model.EventRecordFilter) ([]model.EventRecord, int, error) {
	var conditions []string
	var args []interface{}

	if filter.WorkspaceID != "" {
		conditions = append(conditions, "workspace_id = ?")
		args = append(args, filter.WorkspaceID)
	}
	if filter.Source != "" {
		conditions = append(conditions, "source = ?")
		args = append(args, filter.Source)
	}
	if filter.Severity != "" {
		conditions = append(conditions, "severity = ?")
		args = append(args, filter.Severity)
	}
	if filter.CIID != "" {
		conditions = append(conditions, "ci_id = ?")
		args = append(args, filter.CIID)
	}
	if filter.Resource != "" {
		conditions = append(conditions, "resource LIKE ?")
		args = append(args, "%"+filter.Resource+"%")
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM event_records %s", whereClause)
	var total int
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	limit := filter.Limit
	if limit <= 0 {
		limit = 50
	}
	offset := filter.Offset
	if offset < 0 {
		offset = 0
	}

	selectQuery := fmt.Sprintf(`
		SELECT id, workspace_id, source, event_type, severity, ci_id,
		       resource, message, fingerprint, deduplication_key, raw_payload,
		       timestamp, created_at
		FROM event_records
		%s
		ORDER BY timestamp DESC
		LIMIT ? OFFSET ?`, whereClause)

	queryArgs := append(args, limit, offset)
	rows, err := r.db.QueryContext(ctx, selectQuery, queryArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var events []model.EventRecord
	for rows.Next() {
		var e model.EventRecord
		var ciID, rawPayload sql.NullString
		if err := rows.Scan(
			&e.ID,
			&e.WorkspaceID,
			&e.Source,
			&e.EventType,
			&e.Severity,
			&ciID,
			&e.Resource,
			&e.Message,
			&e.Fingerprint,
			&e.DeduplicationKey,
			&rawPayload,
			&e.Timestamp,
			&e.CreatedAt,
		); err != nil {
			return nil, 0, err
		}
		if ciID.Valid {
			e.CIID = ciID.String
		}
		if rawPayload.Valid {
			e.RawPayload = rawPayload.String
		}
		events = append(events, e)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return events, total, nil
}

// CreateAlert inserts a new correlated alert record
func (r *sqlEventRepository) CreateAlert(ctx context.Context, a *model.CorrelatedAlert) error {
	if a.ID == "" {
		a.ID = uuid.New().String()
	}
	now := time.Now().UTC()
	if a.FirstSeenAt.IsZero() {
		a.FirstSeenAt = now
	}
	if a.LastSeenAt.IsZero() {
		a.LastSeenAt = now
	}
	if a.CreatedAt.IsZero() {
		a.CreatedAt = now
	}
	a.UpdatedAt = now
	if a.EventCount <= 0 {
		a.EventCount = 1
	}

	// Check alert code uniqueness within workspace
	var count int
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM correlated_alerts WHERE workspace_id = ? AND alert_code = ?`, a.WorkspaceID, a.AlertCode).Scan(&count)
	if err == nil && count > 0 {
		return ErrCorrelatedAlertCodeConflict
	}

	query := `
		INSERT INTO correlated_alerts (
			id, workspace_id, alert_code, title, description,
			status, severity, priority, ci_id, resource,
			fingerprint, event_count, first_seen_at, last_seen_at,
			acknowledged_by, acknowledged_at, resolved_at,
			incident_id, rule_id, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err = r.db.ExecContext(ctx, query,
		a.ID,
		a.WorkspaceID,
		a.AlertCode,
		a.Title,
		a.Description,
		a.Status,
		a.Severity,
		a.Priority,
		a.CIID,
		a.Resource,
		a.Fingerprint,
		a.EventCount,
		a.FirstSeenAt,
		a.LastSeenAt,
		a.AcknowledgedBy,
		a.AcknowledgedAt,
		a.ResolvedAt,
		a.IncidentID,
		a.RuleID,
		a.CreatedAt,
		a.UpdatedAt,
	)
	return err
}

// GetAlertByID retrieves a correlated alert by ID
func (r *sqlEventRepository) GetAlertByID(ctx context.Context, id string) (*model.CorrelatedAlert, error) {
	query := `
		SELECT id, workspace_id, alert_code, title, description,
		       status, severity, priority, ci_id, resource,
		       fingerprint, event_count, first_seen_at, last_seen_at,
		       acknowledged_by, acknowledged_at, resolved_at,
		       incident_id, rule_id, created_at, updated_at
		FROM correlated_alerts
		WHERE id = ?`

	var a model.CorrelatedAlert
	var desc, ciID, ackBy, incID, ruleID sql.NullString
	var ackAt, resAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&a.ID,
		&a.WorkspaceID,
		&a.AlertCode,
		&a.Title,
		&desc,
		&a.Status,
		&a.Severity,
		&a.Priority,
		&ciID,
		&a.Resource,
		&a.Fingerprint,
		&a.EventCount,
		&a.FirstSeenAt,
		&a.LastSeenAt,
		&ackBy,
		&ackAt,
		&resAt,
		&incID,
		&ruleID,
		&a.CreatedAt,
		&a.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrCorrelatedAlertNotFound
		}
		return nil, err
	}
	if desc.Valid {
		a.Description = desc.String
	}
	if ciID.Valid {
		a.CIID = ciID.String
	}
	if ackBy.Valid {
		a.AcknowledgedBy = ackBy.String
	}
	if ackAt.Valid {
		a.AcknowledgedAt = &ackAt.Time
	}
	if resAt.Valid {
		a.ResolvedAt = &resAt.Time
	}
	if incID.Valid {
		a.IncidentID = incID.String
	}
	if ruleID.Valid {
		a.RuleID = ruleID.String
	}
	return &a, nil
}

// GetAlertByFingerprint retrieves the active (non-resolved) alert for a given fingerprint in workspace
func (r *sqlEventRepository) GetAlertByFingerprint(ctx context.Context, workspaceID, fingerprint string) (*model.CorrelatedAlert, error) {
	query := `
		SELECT id, workspace_id, alert_code, title, description,
		       status, severity, priority, ci_id, resource,
		       fingerprint, event_count, first_seen_at, last_seen_at,
		       acknowledged_by, acknowledged_at, resolved_at,
		       incident_id, rule_id, created_at, updated_at
		FROM correlated_alerts
		WHERE workspace_id = ? AND fingerprint = ? AND status != 'resolved'
		ORDER BY last_seen_at DESC
		LIMIT 1`

	var a model.CorrelatedAlert
	var desc, ciID, ackBy, incID, ruleID sql.NullString
	var ackAt, resAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, workspaceID, fingerprint).Scan(
		&a.ID,
		&a.WorkspaceID,
		&a.AlertCode,
		&a.Title,
		&desc,
		&a.Status,
		&a.Severity,
		&a.Priority,
		&ciID,
		&a.Resource,
		&a.Fingerprint,
		&a.EventCount,
		&a.FirstSeenAt,
		&a.LastSeenAt,
		&ackBy,
		&ackAt,
		&resAt,
		&incID,
		&ruleID,
		&a.CreatedAt,
		&a.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrCorrelatedAlertNotFound
		}
		return nil, err
	}
	if desc.Valid {
		a.Description = desc.String
	}
	if ciID.Valid {
		a.CIID = ciID.String
	}
	if ackBy.Valid {
		a.AcknowledgedBy = ackBy.String
	}
	if ackAt.Valid {
		a.AcknowledgedAt = &ackAt.Time
	}
	if resAt.Valid {
		a.ResolvedAt = &resAt.Time
	}
	if incID.Valid {
		a.IncidentID = incID.String
	}
	if ruleID.Valid {
		a.RuleID = ruleID.String
	}
	return &a, nil
}

// UpdateAlert updates an existing correlated alert
func (r *sqlEventRepository) UpdateAlert(ctx context.Context, a *model.CorrelatedAlert) error {
	a.UpdatedAt = time.Now().UTC()

	query := `
		UPDATE correlated_alerts
		SET title = ?, description = ?, status = ?, severity = ?,
		    priority = ?, ci_id = ?, resource = ?, event_count = ?,
		    last_seen_at = ?, acknowledged_by = ?, acknowledged_at = ?,
		    resolved_at = ?, incident_id = ?, rule_id = ?, updated_at = ?
		WHERE id = ?`

	res, err := r.db.ExecContext(ctx, query,
		a.Title,
		a.Description,
		a.Status,
		a.Severity,
		a.Priority,
		a.CIID,
		a.Resource,
		a.EventCount,
		a.LastSeenAt,
		a.AcknowledgedBy,
		a.AcknowledgedAt,
		a.ResolvedAt,
		a.IncidentID,
		a.RuleID,
		a.UpdatedAt,
		a.ID,
	)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrCorrelatedAlertNotFound
	}
	return nil
}

// ListAlerts queries correlated alerts based on filter criteria
func (r *sqlEventRepository) ListAlerts(ctx context.Context, filter model.CorrelatedAlertFilter) ([]model.CorrelatedAlert, int, error) {
	var conditions []string
	var args []interface{}

	if filter.WorkspaceID != "" {
		conditions = append(conditions, "workspace_id = ?")
		args = append(args, filter.WorkspaceID)
	}
	if filter.Status != "" {
		conditions = append(conditions, "status = ?")
		args = append(args, filter.Status)
	}
	if filter.Severity != "" {
		conditions = append(conditions, "severity = ?")
		args = append(args, filter.Severity)
	}
	if filter.Priority != "" {
		conditions = append(conditions, "priority = ?")
		args = append(args, filter.Priority)
	}
	if filter.CIID != "" {
		conditions = append(conditions, "ci_id = ?")
		args = append(args, filter.CIID)
	}
	if filter.Search != "" {
		conditions = append(conditions, "(title LIKE ? OR resource LIKE ? OR alert_code LIKE ?)")
		searchTerm := "%" + filter.Search + "%"
		args = append(args, searchTerm, searchTerm, searchTerm)
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM correlated_alerts %s", whereClause)
	var total int
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	limit := filter.Limit
	if limit <= 0 {
		limit = 50
	}
	offset := filter.Offset
	if offset < 0 {
		offset = 0
	}

	selectQuery := fmt.Sprintf(`
		SELECT id, workspace_id, alert_code, title, description,
		       status, severity, priority, ci_id, resource,
		       fingerprint, event_count, first_seen_at, last_seen_at,
		       acknowledged_by, acknowledged_at, resolved_at,
		       incident_id, rule_id, created_at, updated_at
		FROM correlated_alerts
		%s
		ORDER BY last_seen_at DESC
		LIMIT ? OFFSET ?`, whereClause)

	queryArgs := append(args, limit, offset)
	rows, err := r.db.QueryContext(ctx, selectQuery, queryArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var alerts []model.CorrelatedAlert
	for rows.Next() {
		var a model.CorrelatedAlert
		var desc, ciID, ackBy, incID, ruleID sql.NullString
		var ackAt, resAt sql.NullTime

		if err := rows.Scan(
			&a.ID,
			&a.WorkspaceID,
			&a.AlertCode,
			&a.Title,
			&desc,
			&a.Status,
			&a.Severity,
			&a.Priority,
			&ciID,
			&a.Resource,
			&a.Fingerprint,
			&a.EventCount,
			&a.FirstSeenAt,
			&a.LastSeenAt,
			&ackBy,
			&ackAt,
			&resAt,
			&incID,
			&ruleID,
			&a.CreatedAt,
			&a.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		if desc.Valid {
			a.Description = desc.String
		}
		if ciID.Valid {
			a.CIID = ciID.String
		}
		if ackBy.Valid {
			a.AcknowledgedBy = ackBy.String
		}
		if ackAt.Valid {
			a.AcknowledgedAt = &ackAt.Time
		}
		if resAt.Valid {
			a.ResolvedAt = &resAt.Time
		}
		if incID.Valid {
			a.IncidentID = incID.String
		}
		if ruleID.Valid {
			a.RuleID = ruleID.String
		}
		alerts = append(alerts, a)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return alerts, total, nil
}

// CreateRule inserts a new event correlation rule
func (r *sqlEventRepository) CreateRule(ctx context.Context, rule *model.CorrelationRule) error {
	if rule.ID == "" {
		rule.ID = uuid.New().String()
	}
	now := time.Now().UTC()
	if rule.CreatedAt.IsZero() {
		rule.CreatedAt = now
	}
	rule.UpdatedAt = now
	if rule.TimeWindowSeconds <= 0 {
		rule.TimeWindowSeconds = 300
	}
	if rule.ThresholdCount <= 0 {
		rule.ThresholdCount = 1
	}
	if rule.ActionType == "" {
		rule.ActionType = model.RuleActionCreateAlert
	}

	query := `
		INSERT INTO correlation_rules (
			id, workspace_id, name, description, source_pattern,
			severity_pattern, time_window_seconds, threshold_count,
			grouping_fields, action_type, is_active, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := r.db.ExecContext(ctx, query,
		rule.ID,
		rule.WorkspaceID,
		rule.Name,
		rule.Description,
		rule.SourcePattern,
		rule.SeverityPattern,
		rule.TimeWindowSeconds,
		rule.ThresholdCount,
		rule.GroupingFields,
		rule.ActionType,
		rule.IsActive,
		rule.CreatedAt,
		rule.UpdatedAt,
	)
	return err
}

// GetRuleByID retrieves a correlation rule by ID
func (r *sqlEventRepository) GetRuleByID(ctx context.Context, id string) (*model.CorrelationRule, error) {
	query := `
		SELECT id, workspace_id, name, description, source_pattern,
		       severity_pattern, time_window_seconds, threshold_count,
		       grouping_fields, action_type, is_active, created_at, updated_at
		FROM correlation_rules
		WHERE id = ?`

	var rule model.CorrelationRule
	var desc sql.NullString
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&rule.ID,
		&rule.WorkspaceID,
		&rule.Name,
		&desc,
		&rule.SourcePattern,
		&rule.SeverityPattern,
		&rule.TimeWindowSeconds,
		&rule.ThresholdCount,
		&rule.GroupingFields,
		&rule.ActionType,
		&rule.IsActive,
		&rule.CreatedAt,
		&rule.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrCorrelationRuleNotFound
		}
		return nil, err
	}
	if desc.Valid {
		rule.Description = desc.String
	}
	return &rule, nil
}

// UpdateRule updates an existing correlation rule
func (r *sqlEventRepository) UpdateRule(ctx context.Context, rule *model.CorrelationRule) error {
	rule.UpdatedAt = time.Now().UTC()

	query := `
		UPDATE correlation_rules
		SET name = ?, description = ?, source_pattern = ?, severity_pattern = ?,
		    time_window_seconds = ?, threshold_count = ?, grouping_fields = ?,
		    action_type = ?, is_active = ?, updated_at = ?
		WHERE id = ?`

	res, err := r.db.ExecContext(ctx, query,
		rule.Name,
		rule.Description,
		rule.SourcePattern,
		rule.SeverityPattern,
		rule.TimeWindowSeconds,
		rule.ThresholdCount,
		rule.GroupingFields,
		rule.ActionType,
		rule.IsActive,
		rule.UpdatedAt,
		rule.ID,
	)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrCorrelationRuleNotFound
	}
	return nil
}

// DeleteRule removes a correlation rule by ID
func (r *sqlEventRepository) DeleteRule(ctx context.Context, id string) error {
	res, err := r.db.ExecContext(ctx, "DELETE FROM correlation_rules WHERE id = ?", id)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrCorrelationRuleNotFound
	}
	return nil
}

// ListRules queries correlation rules based on filter criteria
func (r *sqlEventRepository) ListRules(ctx context.Context, filter model.CorrelationRuleFilter) ([]model.CorrelationRule, int, error) {
	var conditions []string
	var args []interface{}

	if filter.WorkspaceID != "" {
		conditions = append(conditions, "workspace_id = ?")
		args = append(args, filter.WorkspaceID)
	}
	if filter.IsActive != nil {
		conditions = append(conditions, "is_active = ?")
		args = append(args, *filter.IsActive)
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM correlation_rules %s", whereClause)
	var total int
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	limit := filter.Limit
	if limit <= 0 {
		limit = 50
	}
	offset := filter.Offset
	if offset < 0 {
		offset = 0
	}

	selectQuery := fmt.Sprintf(`
		SELECT id, workspace_id, name, description, source_pattern,
		       severity_pattern, time_window_seconds, threshold_count,
		       grouping_fields, action_type, is_active, created_at, updated_at
		FROM correlation_rules
		%s
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?`, whereClause)

	queryArgs := append(args, limit, offset)
	rows, err := r.db.QueryContext(ctx, selectQuery, queryArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var rules []model.CorrelationRule
	for rows.Next() {
		var rule model.CorrelationRule
		var desc sql.NullString
		if err := rows.Scan(
			&rule.ID,
			&rule.WorkspaceID,
			&rule.Name,
			&desc,
			&rule.SourcePattern,
			&rule.SeverityPattern,
			&rule.TimeWindowSeconds,
			&rule.ThresholdCount,
			&rule.GroupingFields,
			&rule.ActionType,
			&rule.IsActive,
			&rule.CreatedAt,
			&rule.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		if desc.Valid {
			rule.Description = desc.String
		}
		rules = append(rules, rule)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return rules, total, nil
}

// CreateSuppression stub
func (r *sqlEventRepository) CreateSuppression(ctx context.Context, s *model.EventSuppression) error {
	return errors.New("not implemented")
}

// GetSuppressionByID stub
func (r *sqlEventRepository) GetSuppressionByID(ctx context.Context, id string) (*model.EventSuppression, error) {
	return nil, errors.New("not implemented")
}

// UpdateSuppression stub
func (r *sqlEventRepository) UpdateSuppression(ctx context.Context, s *model.EventSuppression) error {
	return errors.New("not implemented")
}

// DeleteSuppression stub
func (r *sqlEventRepository) DeleteSuppression(ctx context.Context, id string) error {
	return errors.New("not implemented")
}

// ListSuppressions stub
func (r *sqlEventRepository) ListSuppressions(ctx context.Context, filter model.EventSuppressionFilter) ([]model.EventSuppression, int, error) {
	return nil, 0, errors.New("not implemented")
}

// CreateAction stub
func (r *sqlEventRepository) CreateAction(ctx context.Context, a *model.EventAction) error {
	return errors.New("not implemented")
}

// ListActions stub
func (r *sqlEventRepository) ListActions(ctx context.Context, filter model.EventActionFilter) ([]model.EventAction, int, error) {
	return nil, 0, errors.New("not implemented")
}

// GetEventAnalytics stub
func (r *sqlEventRepository) GetEventAnalytics(ctx context.Context, workspaceID string) (*model.EventAnalytics, error) {
	return nil, errors.New("not implemented")
}
