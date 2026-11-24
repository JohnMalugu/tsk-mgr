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

// CreateAlert stub
func (r *sqlEventRepository) CreateAlert(ctx context.Context, a *model.CorrelatedAlert) error {
	return errors.New("not implemented")
}

// GetAlertByID stub
func (r *sqlEventRepository) GetAlertByID(ctx context.Context, id string) (*model.CorrelatedAlert, error) {
	return nil, errors.New("not implemented")
}

// GetAlertByFingerprint stub
func (r *sqlEventRepository) GetAlertByFingerprint(ctx context.Context, workspaceID, fingerprint string) (*model.CorrelatedAlert, error) {
	return nil, errors.New("not implemented")
}

// UpdateAlert stub
func (r *sqlEventRepository) UpdateAlert(ctx context.Context, a *model.CorrelatedAlert) error {
	return errors.New("not implemented")
}

// ListAlerts stub
func (r *sqlEventRepository) ListAlerts(ctx context.Context, filter model.CorrelatedAlertFilter) ([]model.CorrelatedAlert, int, error) {
	return nil, 0, errors.New("not implemented")
}

// CreateRule stub
func (r *sqlEventRepository) CreateRule(ctx context.Context, rule *model.CorrelationRule) error {
	return errors.New("not implemented")
}

// GetRuleByID stub
func (r *sqlEventRepository) GetRuleByID(ctx context.Context, id string) (*model.CorrelationRule, error) {
	return nil, errors.New("not implemented")
}

// UpdateRule stub
func (r *sqlEventRepository) UpdateRule(ctx context.Context, rule *model.CorrelationRule) error {
	return errors.New("not implemented")
}

// DeleteRule stub
func (r *sqlEventRepository) DeleteRule(ctx context.Context, id string) error {
	return errors.New("not implemented")
}

// ListRules stub
func (r *sqlEventRepository) ListRules(ctx context.Context, filter model.CorrelationRuleFilter) ([]model.CorrelationRule, int, error) {
	return nil, 0, errors.New("not implemented")
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
