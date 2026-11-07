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
	return nil, 0, nil
}

func (r *auditRepositoryImpl) PurgeExpiredAuditLogs(workspaceID int64, olderThan time.Time) (int64, error) {
	return 0, nil
}

func (r *auditRepositoryImpl) GetRetentionPolicy(workspaceID int64) (*model.AuditRetentionPolicy, error) {
	return nil, nil
}

func (r *auditRepositoryImpl) UpsertRetentionPolicy(policy *model.AuditRetentionPolicy) error {
	return nil
}

func (r *auditRepositoryImpl) CreateAlert(alert *model.AuditAlert) (int64, error) {
	return 0, nil
}

func (r *auditRepositoryImpl) GetAlertByID(id int64) (*model.AuditAlert, error) {
	return nil, nil
}

func (r *auditRepositoryImpl) ListAlerts(filter model.AuditAlertFilter) ([]*model.AuditAlert, error) {
	return nil, nil
}

func (r *auditRepositoryImpl) ResolveAlert(id int64, status, notes, resolvedBy string) error {
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
