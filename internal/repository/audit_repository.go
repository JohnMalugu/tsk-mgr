package repository

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/JohnMalugu/tsk-mgr-api/internal/model"
)

var (
	ErrAuditLogNotFound        = errors.New("audit log not found")
	ErrRetentionPolicyNotFound = errors.New("retention policy not found")
	ErrAuditAlertNotFound      = errors.New("audit alert not found")
	ErrAuditExportNotFound     = errors.New("audit export not found")
)

// AuditRepository provides database access for audit logs, policies, alerts, and exports
type AuditRepository interface {
	// Audit logs
	CreateAuditLog(log *model.AuditLog) (int64, error)
	GetAuditLogByID(id int64) (*model.AuditLog, error)
	ListAuditLogs(filter model.AuditLogFilter) ([]*model.AuditLog, int64, error)
	PurgeExpiredAuditLogs(workspaceID int64, olderThan time.Time) (int64, error)

	// Retention policy
	GetRetentionPolicy(workspaceID int64) (*model.AuditRetentionPolicy, error)
	UpsertRetentionPolicy(policy *model.AuditRetentionPolicy) error

	// Alerts
	CreateAlert(alert *model.AuditAlert) (int64, error)
	GetAlertByID(id int64) (*model.AuditAlert, error)
	ListAlerts(filter model.AuditAlertFilter) ([]*model.AuditAlert, error)
	ResolveAlert(id int64, status, notes, resolvedBy string) error

	// Exports
	CreateExport(export *model.AuditExport) (int64, error)
	GetExportByID(id int64) (*model.AuditExport, error)
	ListExports(workspaceID int64) ([]*model.AuditExport, error)

	// Metrics & Compliance
	GetAuditMetrics(workspaceID int64) (*model.AuditMetrics, error)
	GetComplianceSummary(workspaceID int64) (*model.ComplianceSummary, error)
}

type auditRepositoryImpl struct {
	db *sql.DB
}

// NewAuditRepository creates a new instance of AuditRepository
func NewAuditRepository(db *sql.DB) AuditRepository {
	return &auditRepositoryImpl{db: db}
}

// ComputeChecksum generates a SHA256 integrity hash for an audit record
func ComputeChecksum(workspaceID int64, actorID, action, entityType, entityID, description string, t time.Time) string {
	payload := fmt.Sprintf("%d:%s:%s:%s:%s:%s:%d", workspaceID, actorID, action, entityType, entityID, description, t.UnixNano())
	hash := sha256.Sum256([]byte(payload))
	return hex.EncodeToString(hash[:])
}

const auditLogBaseSelect = `
	SELECT 
		al.id,
		al.workspace_id,
		al.actor_id,
		al.actor_name,
		al.actor_email,
		al.action,
		al.category,
		al.severity,
		al.entity_type,
		al.entity_id,
		al.entity_name,
		al.description,
		al.before_state,
		al.after_state,
		al.metadata,
		al.ip_address,
		al.user_agent,
		al.checksum,
		al.created_at
	FROM audit_logs al
`

func scanAuditLog(scanner interface {
	Scan(dest ...interface{}) error
}) (*model.AuditLog, error) {
	var l model.AuditLog
	err := scanner.Scan(
		&l.ID,
		&l.WorkspaceID,
		&l.ActorID,
		&l.ActorName,
		&l.ActorEmail,
		&l.Action,
		&l.Category,
		&l.Severity,
		&l.EntityType,
		&l.EntityID,
		&l.EntityName,
		&l.Description,
		&l.BeforeState,
		&l.AfterState,
		&l.Metadata,
		&l.IPAddress,
		&l.UserAgent,
		&l.Checksum,
		&l.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &l, nil
}

func (r *auditRepositoryImpl) CreateAuditLog(log *model.AuditLog) (int64, error) {
	now := time.Now()
	if log.CreatedAt.IsZero() {
		log.CreatedAt = now
	}
	if log.Category == "" {
		log.Category = model.AuditCategorySystem
	}
	if log.Severity == "" {
		log.Severity = model.AuditSeverityInfo
	}

	// Auto-enrich actor name/email from users table if blank
	if log.ActorName == "" || log.ActorEmail == "" {
		var uName, uEmail string
		err := r.db.QueryRow(`SELECT username, email FROM users WHERE id = ?`, log.ActorID).Scan(&uName, &uEmail)
		if err == nil {
			if log.ActorName == "" {
				log.ActorName = uName
			}
			if log.ActorEmail == "" {
				log.ActorEmail = uEmail
			}
		}
	}

	if log.Checksum == "" {
		log.Checksum = ComputeChecksum(
			log.WorkspaceID,
			log.ActorID,
			log.Action,
			log.EntityType,
			log.EntityID,
			log.Description,
			log.CreatedAt,
		)
	}

	query := `
	INSERT INTO audit_logs (
		workspace_id, actor_id, actor_name, actor_email, action, category, severity,
		entity_type, entity_id, entity_name, description, before_state, after_state,
		metadata, ip_address, user_agent, checksum, created_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	res, err := r.db.Exec(
		query,
		log.WorkspaceID,
		log.ActorID,
		log.ActorName,
		log.ActorEmail,
		log.Action,
		log.Category,
		log.Severity,
		log.EntityType,
		log.EntityID,
		log.EntityName,
		log.Description,
		log.BeforeState,
		log.AfterState,
		log.Metadata,
		log.IPAddress,
		log.UserAgent,
		log.Checksum,
		log.CreatedAt,
	)
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	log.ID = id
	return id, nil
}

func (r *auditRepositoryImpl) GetAuditLogByID(id int64) (*model.AuditLog, error) {
	query := auditLogBaseSelect + ` WHERE al.id = ?`
	row := r.db.QueryRow(query, id)
	log, err := scanAuditLog(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrAuditLogNotFound
		}
		return nil, err
	}
	return log, nil
}

func (r *auditRepositoryImpl) ListAuditLogs(filter model.AuditLogFilter) ([]*model.AuditLog, int64, error) {
	var conditions []string
	var args []interface{}

	conditions = append(conditions, "al.workspace_id = ?")
	args = append(args, filter.WorkspaceID)

	if filter.ActorID != nil && strings.TrimSpace(*filter.ActorID) != "" {
		conditions = append(conditions, "al.actor_id = ?")
		args = append(args, strings.TrimSpace(*filter.ActorID))
	}
	if filter.Action != nil && strings.TrimSpace(*filter.Action) != "" {
		conditions = append(conditions, "al.action = ?")
		args = append(args, strings.TrimSpace(*filter.Action))
	}
	if filter.Category != nil && strings.TrimSpace(*filter.Category) != "" {
		conditions = append(conditions, "al.category = ?")
		args = append(args, strings.TrimSpace(*filter.Category))
	}
	if filter.Severity != nil && strings.TrimSpace(*filter.Severity) != "" {
		conditions = append(conditions, "al.severity = ?")
		args = append(args, strings.TrimSpace(*filter.Severity))
	}
	if filter.EntityType != nil && strings.TrimSpace(*filter.EntityType) != "" {
		conditions = append(conditions, "al.entity_type = ?")
		args = append(args, strings.TrimSpace(*filter.EntityType))
	}
	if filter.EntityID != nil && strings.TrimSpace(*filter.EntityID) != "" {
		conditions = append(conditions, "al.entity_id = ?")
		args = append(args, strings.TrimSpace(*filter.EntityID))
	}
	if filter.Search != nil && strings.TrimSpace(*filter.Search) != "" {
		term := "%" + strings.TrimSpace(*filter.Search) + "%"
		conditions = append(conditions, "(al.description LIKE ? OR al.actor_name LIKE ? OR al.entity_name LIKE ?)")
		args = append(args, term, term, term)
	}
	if filter.StartDate != nil && strings.TrimSpace(*filter.StartDate) != "" {
		conditions = append(conditions, "al.created_at >= ?")
		args = append(args, strings.TrimSpace(*filter.StartDate))
	}
	if filter.EndDate != nil && strings.TrimSpace(*filter.EndDate) != "" {
		conditions = append(conditions, "al.created_at <= ?")
		args = append(args, strings.TrimSpace(*filter.EndDate))
	}

	whereClause := strings.Join(conditions, " AND ")

	// Count query
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM audit_logs al WHERE %s", whereClause)
	var total int64
	if err := r.db.QueryRow(countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	limit := filter.Limit
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}
	offset := filter.Offset
	if offset < 0 {
		offset = 0
	}

	query := fmt.Sprintf("%s WHERE %s ORDER BY al.created_at DESC, al.id DESC LIMIT ? OFFSET ?", auditLogBaseSelect, whereClause)
	queryArgs := append(args, limit, offset)

	rows, err := r.db.Query(query, queryArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var logs []*model.AuditLog
	for rows.Next() {
		l, err := scanAuditLog(rows)
		if err != nil {
			return nil, 0, err
		}
		logs = append(logs, l)
	}

	if logs == nil {
		logs = []*model.AuditLog{}
	}
	return logs, total, nil
}

func (r *auditRepositoryImpl) PurgeExpiredAuditLogs(workspaceID int64, olderThan time.Time) (int64, error) {
	return 0, nil
}

const retentionPolicyBaseSelect = `
	SELECT 
		id,
		workspace_id,
		retention_days,
		frameworks,
		auto_purge_enabled,
		last_purge_at,
		created_at,
		updated_at
	FROM audit_retention_policies
`

func (r *auditRepositoryImpl) GetRetentionPolicy(workspaceID int64) (*model.AuditRetentionPolicy, error) {
	query := retentionPolicyBaseSelect + ` WHERE workspace_id = ?`
	row := r.db.QueryRow(query, workspaceID)

	var p model.AuditRetentionPolicy
	var lastPurge sql.NullTime
	err := row.Scan(
		&p.ID,
		&p.WorkspaceID,
		&p.RetentionDays,
		&p.Frameworks,
		&p.AutoPurgeEnabled,
		&lastPurge,
		&p.CreatedAt,
		&p.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrRetentionPolicyNotFound
		}
		return nil, err
	}
	if lastPurge.Valid {
		p.LastPurgeAt = &lastPurge.Time
	}
	return &p, nil
}

func (r *auditRepositoryImpl) UpsertRetentionPolicy(policy *model.AuditRetentionPolicy) error {
	now := time.Now()
	if policy.CreatedAt.IsZero() {
		policy.CreatedAt = now
	}
	policy.UpdatedAt = now

	query := `
	INSERT INTO audit_retention_policies (
		workspace_id, retention_days, frameworks, auto_purge_enabled, created_at, updated_at
	) VALUES (?, ?, ?, ?, ?, ?)
	ON CONFLICT(workspace_id) DO UPDATE SET
		retention_days = excluded.retention_days,
		frameworks = excluded.frameworks,
		auto_purge_enabled = excluded.auto_purge_enabled,
		updated_at = excluded.updated_at
	`
	_, err := r.db.Exec(
		query,
		policy.WorkspaceID,
		policy.RetentionDays,
		policy.Frameworks,
		policy.AutoPurgeEnabled,
		policy.CreatedAt,
		policy.UpdatedAt,
	)
	return err
}

const auditAlertBaseSelect = `
	SELECT 
		aa.id,
		aa.workspace_id,
		aa.title,
		aa.description,
		aa.severity,
		aa.status,
		aa.trigger_event_id,
		aa.resolution_notes,
		aa.resolved_by_user_id,
		aa.resolved_at,
		aa.created_at,
		aa.updated_at,
		COALESCE(al.action, '') AS trigger_action,
		COALESCE(al.actor_name, '') AS trigger_actor,
		COALESCE(u.username, '') AS resolved_by_username
	FROM audit_alerts aa
	LEFT JOIN audit_logs al ON aa.trigger_event_id = al.id
	LEFT JOIN users u ON aa.resolved_by_user_id = u.id
`

func scanAuditAlert(scanner interface {
	Scan(dest ...interface{}) error
}) (*model.AuditAlert, error) {
	var a model.AuditAlert
	var triggerID sql.NullInt64
	var resolvedBy sql.NullString
	var resolvedAt sql.NullTime

	err := scanner.Scan(
		&a.ID,
		&a.WorkspaceID,
		&a.Title,
		&a.Description,
		&a.Severity,
		&a.Status,
		&triggerID,
		&a.ResolutionNotes,
		&resolvedBy,
		&resolvedAt,
		&a.CreatedAt,
		&a.UpdatedAt,
		&a.TriggerAction,
		&a.TriggerActor,
		&a.ResolvedByUsername,
	)
	if err != nil {
		return nil, err
	}
	if triggerID.Valid {
		a.TriggerEventID = &triggerID.Int64
	}
	if resolvedBy.Valid {
		a.ResolvedByUserID = resolvedBy.String
	}
	if resolvedAt.Valid {
		a.ResolvedAt = &resolvedAt.Time
	}
	return &a, nil
}

func (r *auditRepositoryImpl) CreateAlert(alert *model.AuditAlert) (int64, error) {
	now := time.Now()
	if alert.CreatedAt.IsZero() {
		alert.CreatedAt = now
	}
	alert.UpdatedAt = now
	if alert.Severity == "" {
		alert.Severity = model.AuditSeverityWarning
	}
	if alert.Status == "" {
		alert.Status = model.AuditAlertStatusOpen
	}

	query := `
	INSERT INTO audit_alerts (
		workspace_id, title, description, severity, status, trigger_event_id,
		resolution_notes, resolved_by_user_id, resolved_at, created_at, updated_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	res, err := r.db.Exec(
		query,
		alert.WorkspaceID,
		alert.Title,
		alert.Description,
		alert.Severity,
		alert.Status,
		alert.TriggerEventID,
		alert.ResolutionNotes,
		alert.ResolvedByUserID,
		alert.ResolvedAt,
		alert.CreatedAt,
		alert.UpdatedAt,
	)
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	alert.ID = id
	return id, nil
}

func (r *auditRepositoryImpl) GetAlertByID(id int64) (*model.AuditAlert, error) {
	query := auditAlertBaseSelect + ` WHERE aa.id = ?`
	row := r.db.QueryRow(query, id)
	alert, err := scanAuditAlert(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrAuditAlertNotFound
		}
		return nil, err
	}
	return alert, nil
}

func (r *auditRepositoryImpl) ListAlerts(filter model.AuditAlertFilter) ([]*model.AuditAlert, error) {
	var conditions []string
	var args []interface{}

	conditions = append(conditions, "aa.workspace_id = ?")
	args = append(args, filter.WorkspaceID)

	if filter.Status != nil && strings.TrimSpace(*filter.Status) != "" {
		conditions = append(conditions, "aa.status = ?")
		args = append(args, strings.TrimSpace(*filter.Status))
	}
	if filter.Severity != nil && strings.TrimSpace(*filter.Severity) != "" {
		conditions = append(conditions, "aa.severity = ?")
		args = append(args, strings.TrimSpace(*filter.Severity))
	}

	whereClause := strings.Join(conditions, " AND ")
	limit := filter.Limit
	if limit <= 0 {
		limit = 50
	}
	offset := filter.Offset
	if offset < 0 {
		offset = 0
	}

	query := fmt.Sprintf("%s WHERE %s ORDER BY aa.created_at DESC LIMIT ? OFFSET ?", auditAlertBaseSelect, whereClause)
	args = append(args, limit, offset)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var alerts []*model.AuditAlert
	for rows.Next() {
		a, err := scanAuditAlert(rows)
		if err != nil {
			return nil, err
		}
		alerts = append(alerts, a)
	}
	if alerts == nil {
		alerts = []*model.AuditAlert{}
	}
	return alerts, nil
}

func (r *auditRepositoryImpl) ResolveAlert(id int64, status, notes, resolvedBy string) error {
	now := time.Now()
	query := `
	UPDATE audit_alerts
	SET status = ?, resolution_notes = ?, resolved_by_user_id = ?, resolved_at = ?, updated_at = ?
	WHERE id = ?
	`
	res, err := r.db.Exec(query, status, notes, resolvedBy, now, now, id)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrAuditAlertNotFound
	}
	return nil
}

func (r *auditRepositoryImpl) CreateExport(export *model.AuditExport) (int64, error) {
	return 0, nil
}

func (r *auditRepositoryImpl) GetExportByID(id int64) (*model.AuditExport, error) {
	return nil, nil
}

func (r *auditRepositoryImpl) ListExports(workspaceID int64) ([]*model.AuditExport, error) {
	return nil, nil
}

func (r *auditRepositoryImpl) GetAuditMetrics(workspaceID int64) (*model.AuditMetrics, error) {
	return nil, nil
}

func (r *auditRepositoryImpl) GetComplianceSummary(workspaceID int64) (*model.ComplianceSummary, error) {
	return nil, nil
}

// Compile-time check
var _ = strings.TrimSpace
