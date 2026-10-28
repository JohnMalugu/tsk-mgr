package repository

import (
	"database/sql"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/JohnMalugu/tsk-mgr-api/internal/model"
)

var (
	ErrSlaPolicyNotFound      = errors.New("sla policy not found")
	ErrTaskSlaNotFound        = errors.New("task sla record not found")
	ErrEscalationRuleNotFound = errors.New("escalation rule not found")
	ErrDuplicatePolicyName    = errors.New("sla policy name already exists in workspace")
)

// SlaRepository provides data access for SLA policies, escalation rules, task SLAs, and breach logs
type SlaRepository interface {
	CreatePolicy(policy *model.SlaPolicy) (int64, error)
	GetPolicyByID(id int64) (*model.SlaPolicy, error)
	GetDefaultPolicy(workspaceID int64, priority string) (*model.SlaPolicy, error)
	ListPolicies(filter model.SlaFilter) ([]*model.SlaSummary, error)
	UpdatePolicy(policy *model.SlaPolicy) error
	TogglePolicyActive(id int64, isActive bool) error
	DeletePolicy(id int64) error

	SetEscalationRules(policyID int64, rules []model.EscalationRule) error
	GetEscalationRules(policyID int64) ([]model.EscalationRule, error)
	DeleteEscalationRule(ruleID int64) error

	CreateTaskSlaRecord(record *model.TaskSlaRecord) (int64, error)
	GetTaskSlaRecord(id int64) (*model.TaskSlaRecord, error)
	GetTaskSlaRecordByTaskID(taskID int) (*model.TaskSlaRecord, error)
	UpdateTaskSlaRecord(record *model.TaskSlaRecord) error

	LogBreach(breach *model.SlaBreachLog) (int64, error)
	ListBreaches(filter model.BreachFilter) ([]*model.SlaBreachLog, error)
	GetSlaMetrics(workspaceID int64) (*model.SlaMetrics, error)
}

type slaRepositoryImpl struct {
	db *sql.DB
}

// NewSlaRepository constructs a new SQLite-backed SlaRepository
func NewSlaRepository(db *sql.DB) SlaRepository {
	return &slaRepositoryImpl{db: db}
}

func scanPolicy(scanner interface{ Scan(dest ...interface{}) error }) (*model.SlaPolicy, error) {
	var (
		p         model.SlaPolicy
		tiersJSON string
		bizHours  int
		isDef     int
		isAct     int
	)

	err := scanner.Scan(
		&p.ID,
		&p.WorkspaceID,
		&p.Name,
		&p.Description,
		&tiersJSON,
		&p.TargetResponseMinutes,
		&p.TargetResolutionMinutes,
		&bizHours,
		&isDef,
		&isAct,
		&p.CreatedByUserID,
		&p.CreatedAt,
		&p.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrSlaPolicyNotFound
		}
		return nil, err
	}

	p.BusinessHoursOnly = bizHours == 1
	p.IsDefault = isDef == 1
	p.IsActive = isAct == 1

	if tiersJSON != "" {
		_ = json.Unmarshal([]byte(tiersJSON), &p.PriorityTiers)
	}
	if p.PriorityTiers == nil {
		p.PriorityTiers = []string{}
	}

	return &p, nil
}

func (r *slaRepositoryImpl) CreatePolicy(policy *model.SlaPolicy) (int64, error) {
	now := time.Now()
	policy.CreatedAt = now
	policy.UpdatedAt = now
	if policy.PriorityTiers == nil {
		policy.PriorityTiers = []string{}
	}

	tiersJSON, err := json.Marshal(policy.PriorityTiers)
	if err != nil {
		return 0, err
	}

	bizVal := 0
	if policy.BusinessHoursOnly {
		bizVal = 1
	}
	defVal := 0
	if policy.IsDefault {
		defVal = 1
	}
	actVal := 1
	if !policy.IsActive {
		actVal = 0
	}

	tx, err := r.db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	if policy.IsDefault {
		_, _ = tx.Exec("UPDATE sla_policies SET is_default = 0 WHERE workspace_id = ?", policy.WorkspaceID)
	}

	query := `
	INSERT INTO sla_policies (
		workspace_id, name, description, priority_tiers,
		target_response_minutes, target_resolution_minutes,
		business_hours_only, is_default, is_active,
		created_by_user_id, created_at, updated_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	res, err := tx.Exec(query,
		policy.WorkspaceID, policy.Name, policy.Description, string(tiersJSON),
		policy.TargetResponseMinutes, policy.TargetResolutionMinutes,
		bizVal, defVal, actVal,
		policy.CreatedByUserID, policy.CreatedAt, policy.UpdatedAt,
	)
	if err != nil {
		return 0, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	policy.ID = id

	if err := tx.Commit(); err != nil {
		return 0, err
	}

	return id, nil
}

func (r *slaRepositoryImpl) GetPolicyByID(id int64) (*model.SlaPolicy, error) {
	query := `
	SELECT id, workspace_id, name, description, priority_tiers,
		target_response_minutes, target_resolution_minutes,
		business_hours_only, is_default, is_active,
		created_by_user_id, created_at, updated_at
	FROM sla_policies
	WHERE id = ?`

	return scanPolicy(r.db.QueryRow(query, id))
}

func (r *slaRepositoryImpl) GetDefaultPolicy(workspaceID int64, priority string) (*model.SlaPolicy, error) {
	if priority != "" {
		rows, err := r.db.Query(`
			SELECT id, workspace_id, name, description, priority_tiers,
				target_response_minutes, target_resolution_minutes,
				business_hours_only, is_default, is_active,
				created_by_user_id, created_at, updated_at
			FROM sla_policies
			WHERE workspace_id = ? AND is_active = 1
			ORDER BY is_default DESC, id ASC`, workspaceID)
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				p, err := scanPolicy(rows)
				if err == nil && p != nil {
					for _, tier := range p.PriorityTiers {
						if strings.EqualFold(tier, priority) {
							return p, nil
						}
					}
				}
			}
		}
	}

	query := `
	SELECT id, workspace_id, name, description, priority_tiers,
		target_response_minutes, target_resolution_minutes,
		business_hours_only, is_default, is_active,
		created_by_user_id, created_at, updated_at
	FROM sla_policies
	WHERE workspace_id = ? AND is_active = 1 AND is_default = 1
	LIMIT 1`

	p, err := scanPolicy(r.db.QueryRow(query, workspaceID))
	if err == nil {
		return p, nil
	}

	queryFallback := `
	SELECT id, workspace_id, name, description, priority_tiers,
		target_response_minutes, target_resolution_minutes,
		business_hours_only, is_default, is_active,
		created_by_user_id, created_at, updated_at
	FROM sla_policies
	WHERE workspace_id = ? AND is_active = 1
	ORDER BY id ASC
	LIMIT 1`

	return scanPolicy(r.db.QueryRow(queryFallback, workspaceID))
}


func (r *slaRepositoryImpl) ListPolicies(filter model.SlaFilter) ([]*model.SlaSummary, error) {
	query := `
	SELECT 
		p.id, p.workspace_id, p.name, p.priority_tiers,
		p.target_response_minutes, p.target_resolution_minutes,
		p.business_hours_only, p.is_default, p.is_active,
		COALESCE((SELECT COUNT(*) FROM sla_escalation_rules er WHERE er.policy_id = p.id), 0) AS rule_count,
		COALESCE((SELECT COUNT(*) FROM task_sla_records tsr WHERE tsr.policy_id = p.id), 0) AS tracked_task_count,
		p.created_at, p.updated_at
	FROM sla_policies p
	WHERE 1=1`

	var args []interface{}

	if filter.WorkspaceID != nil {
		query += " AND p.workspace_id = ?"
		args = append(args, *filter.WorkspaceID)
	}
	if filter.IsActive != nil {
		actVal := 0
		if *filter.IsActive {
			actVal = 1
		}
		query += " AND p.is_active = ?"
		args = append(args, actVal)
	}
	if filter.Search != "" {
		pattern := "%" + filter.Search + "%"
		query += " AND (p.name LIKE ? OR p.description LIKE ?)"
		args = append(args, pattern, pattern)
	}

	query += " ORDER BY p.is_default DESC, p.created_at ASC"

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	summaries := make([]*model.SlaSummary, 0)
	for rows.Next() {
		var (
			s         model.SlaSummary
			tiersJSON string
			bizVal    int
			defVal    int
			actVal    int
		)

		err := rows.Scan(
			&s.ID,
			&s.WorkspaceID,
			&s.Name,
			&tiersJSON,
			&s.TargetResponseMinutes,
			&s.TargetResolutionMinutes,
			&bizVal,
			&defVal,
			&actVal,
			&s.RuleCount,
			&s.TrackedTaskCount,
			&s.CreatedAt,
			&s.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		s.BusinessHoursOnly = bizVal == 1
		s.IsDefault = defVal == 1
		s.IsActive = actVal == 1

		if tiersJSON != "" {
			_ = json.Unmarshal([]byte(tiersJSON), &s.PriorityTiers)
		}
		if s.PriorityTiers == nil {
			s.PriorityTiers = []string{}
		}

		if filter.Priority != "" {
			match := false
			for _, t := range s.PriorityTiers {
				if strings.EqualFold(t, filter.Priority) {
					match = true
					break
				}
			}
			if !match {
				continue
			}
		}

		summaries = append(summaries, &s)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return summaries, nil
}


func (r *slaRepositoryImpl) UpdatePolicy(policy *model.SlaPolicy) error {
	policy.UpdatedAt = time.Now()
	if policy.PriorityTiers == nil {
		policy.PriorityTiers = []string{}
	}

	tiersJSON, err := json.Marshal(policy.PriorityTiers)
	if err != nil {
		return err
	}

	bizVal := 0
	if policy.BusinessHoursOnly {
		bizVal = 1
	}
	defVal := 0
	if policy.IsDefault {
		defVal = 1
	}
	actVal := 0
	if policy.IsActive {
		actVal = 1
	}

	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if policy.IsDefault {
		_, _ = tx.Exec("UPDATE sla_policies SET is_default = 0 WHERE workspace_id = ? AND id != ?", policy.WorkspaceID, policy.ID)
	}

	query := `
	UPDATE sla_policies
	SET name = ?, description = ?, priority_tiers = ?,
		target_response_minutes = ?, target_resolution_minutes = ?,
		business_hours_only = ?, is_default = ?, is_active = ?,
		updated_at = ?
	WHERE id = ?`

	res, err := tx.Exec(query,
		policy.Name, policy.Description, string(tiersJSON),
		policy.TargetResponseMinutes, policy.TargetResolutionMinutes,
		bizVal, defVal, actVal, policy.UpdatedAt,
		policy.ID,
	)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrSlaPolicyNotFound
	}

	return tx.Commit()
}

func (r *slaRepositoryImpl) TogglePolicyActive(id int64, isActive bool) error {
	actVal := 0
	if isActive {
		actVal = 1
	}

	query := `UPDATE sla_policies SET is_active = ?, updated_at = ? WHERE id = ?`
	res, err := r.db.Exec(query, actVal, time.Now(), id)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrSlaPolicyNotFound
	}

	return nil
}

func (r *slaRepositoryImpl) DeletePolicy(id int64) error {
	query := `DELETE FROM sla_policies WHERE id = ?`
	res, err := r.db.Exec(query, id)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrSlaPolicyNotFound
	}

	return nil
}


func (r *slaRepositoryImpl) SetEscalationRules(policyID int64, rules []model.EscalationRule) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var exists int
	err = tx.QueryRow("SELECT 1 FROM sla_policies WHERE id = ?", policyID).Scan(&exists)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrSlaPolicyNotFound
		}
		return err
	}

	_, err = tx.Exec("DELETE FROM sla_escalation_rules WHERE policy_id = ?", policyID)
	if err != nil {
		return err
	}

	insertQuery := `
	INSERT INTO sla_escalation_rules (
		policy_id, trigger_event, threshold_percentage, action_type,
		target_user_id, target_priority, created_at
	) VALUES (?, ?, ?, ?, ?, ?, ?)`

	now := time.Now()
	for _, rule := range rules {
		_, err = tx.Exec(insertQuery,
			policyID, rule.TriggerEvent, rule.ThresholdPercentage, rule.ActionType,
			rule.TargetUserID, rule.TargetPriority, now,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *slaRepositoryImpl) GetEscalationRules(policyID int64) ([]model.EscalationRule, error) {
	var exists int
	err := r.db.QueryRow("SELECT 1 FROM sla_policies WHERE id = ?", policyID).Scan(&exists)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrSlaPolicyNotFound
		}
		return nil, err
	}

	query := `
	SELECT id, policy_id, trigger_event, threshold_percentage, action_type,
		target_user_id, target_priority, created_at
	FROM sla_escalation_rules
	WHERE policy_id = ?
	ORDER BY threshold_percentage ASC, id ASC`

	rows, err := r.db.Query(query, policyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	rules := make([]model.EscalationRule, 0)
	for rows.Next() {
		var rule model.EscalationRule
		err := rows.Scan(
			&rule.ID,
			&rule.PolicyID,
			&rule.TriggerEvent,
			&rule.ThresholdPercentage,
			&rule.ActionType,
			&rule.TargetUserID,
			&rule.TargetPriority,
			&rule.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		rules = append(rules, rule)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return rules, nil
}

func (r *slaRepositoryImpl) DeleteEscalationRule(ruleID int64) error {
	query := `DELETE FROM sla_escalation_rules WHERE id = ?`
	res, err := r.db.Exec(query, ruleID)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrEscalationRuleNotFound
	}

	return nil
}


func scanTaskSla(scanner interface{ Scan(dest ...interface{}) error }) (*model.TaskSlaRecord, error) {
	var (
		rec         model.TaskSlaRecord
		respondedAt sql.NullTime
		resolvedAt  sql.NullTime
		pausedAt    sql.NullTime
		isPausedVal int
	)

	err := scanner.Scan(
		&rec.ID,
		&rec.TaskID,
		&rec.PolicyID,
		&rec.WorkspaceID,
		&rec.ResponseDeadline,
		&rec.ResolutionDeadline,
		&respondedAt,
		&resolvedAt,
		&rec.ResponseStatus,
		&rec.ResolutionStatus,
		&isPausedVal,
		&pausedAt,
		&rec.TotalPausedMinutes,
		&rec.CreatedAt,
		&rec.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrTaskSlaNotFound
		}
		return nil, err
	}

	if respondedAt.Valid {
		rec.RespondedAt = &respondedAt.Time
	}
	if resolvedAt.Valid {
		rec.ResolvedAt = &resolvedAt.Time
	}
	if pausedAt.Valid {
		rec.PausedAt = &pausedAt.Time
	}
	rec.IsPaused = isPausedVal == 1

	return &rec, nil
}

func (r *slaRepositoryImpl) CreateTaskSlaRecord(record *model.TaskSlaRecord) (int64, error) {
	now := time.Now()
	if record.CreatedAt.IsZero() {
		record.CreatedAt = now
	}
	if record.UpdatedAt.IsZero() {
		record.UpdatedAt = now
	}
	if record.ResponseStatus == "" {
		record.ResponseStatus = model.SlaStagePending
	}
	if record.ResolutionStatus == "" {
		record.ResolutionStatus = model.SlaStagePending
	}

	pausedVal := 0
	if record.IsPaused {
		pausedVal = 1
	}

	query := `
	INSERT INTO task_sla_records (
		task_id, policy_id, workspace_id,
		response_deadline, resolution_deadline,
		responded_at, resolved_at,
		response_status, resolution_status,
		is_paused, paused_at, total_paused_minutes,
		created_at, updated_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	ON CONFLICT(task_id) DO UPDATE SET
		policy_id = excluded.policy_id,
		workspace_id = excluded.workspace_id,
		response_deadline = excluded.response_deadline,
		resolution_deadline = excluded.resolution_deadline,
		responded_at = excluded.responded_at,
		resolved_at = excluded.resolved_at,
		response_status = excluded.response_status,
		resolution_status = excluded.resolution_status,
		is_paused = excluded.is_paused,
		paused_at = excluded.paused_at,
		total_paused_minutes = excluded.total_paused_minutes,
		updated_at = excluded.updated_at`

	res, err := r.db.Exec(query,
		record.TaskID, record.PolicyID, record.WorkspaceID,
		record.ResponseDeadline, record.ResolutionDeadline,
		record.RespondedAt, record.ResolvedAt,
		record.ResponseStatus, record.ResolutionStatus,
		pausedVal, record.PausedAt, record.TotalPausedMinutes,
		record.CreatedAt, record.UpdatedAt,
	)
	if err != nil {
		return 0, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	record.ID = id
	return id, nil
}

func (r *slaRepositoryImpl) GetTaskSlaRecord(id int64) (*model.TaskSlaRecord, error) {
	query := `
	SELECT id, task_id, policy_id, workspace_id,
		response_deadline, resolution_deadline,
		responded_at, resolved_at,
		response_status, resolution_status,
		is_paused, paused_at, total_paused_minutes,
		created_at, updated_at
	FROM task_sla_records
	WHERE id = ?`

	return scanTaskSla(r.db.QueryRow(query, id))
}

func (r *slaRepositoryImpl) GetTaskSlaRecordByTaskID(taskID int) (*model.TaskSlaRecord, error) {
	query := `
	SELECT id, task_id, policy_id, workspace_id,
		response_deadline, resolution_deadline,
		responded_at, resolved_at,
		response_status, resolution_status,
		is_paused, paused_at, total_paused_minutes,
		created_at, updated_at
	FROM task_sla_records
	WHERE task_id = ?`

	return scanTaskSla(r.db.QueryRow(query, taskID))
}

func (r *slaRepositoryImpl) UpdateTaskSlaRecord(record *model.TaskSlaRecord) error {
	record.UpdatedAt = time.Now()
	pausedVal := 0
	if record.IsPaused {
		pausedVal = 1
	}

	query := `
	UPDATE task_sla_records
	SET response_deadline = ?, resolution_deadline = ?,
		responded_at = ?, resolved_at = ?,
		response_status = ?, resolution_status = ?,
		is_paused = ?, paused_at = ?, total_paused_minutes = ?,
		updated_at = ?
	WHERE id = ?`

	res, err := r.db.Exec(query,
		record.ResponseDeadline, record.ResolutionDeadline,
		record.RespondedAt, record.ResolvedAt,
		record.ResponseStatus, record.ResolutionStatus,
		pausedVal, record.PausedAt, record.TotalPausedMinutes,
		record.UpdatedAt, record.ID,
	)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrTaskSlaNotFound
	}

	return nil
}


func (r *slaRepositoryImpl) LogBreach(breach *model.SlaBreachLog) (int64, error) {
	if breach.BreachedAt.IsZero() {
		breach.BreachedAt = time.Now()
	}

	escVal := 0
	if breach.EscalationTriggered {
		escVal = 1
	}

	query := `
	INSERT INTO sla_breach_logs (
		task_id, policy_id, workspace_id,
		breach_type, breached_at, minutes_overdue,
		escalation_triggered, escalation_action, details
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`

	res, err := r.db.Exec(query,
		breach.TaskID, breach.PolicyID, breach.WorkspaceID,
		breach.BreachType, breach.BreachedAt, breach.MinutesOverdue,
		escVal, breach.EscalationAction, breach.Details,
	)
	if err != nil {
		return 0, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	breach.ID = id
	return id, nil
}

func (r *slaRepositoryImpl) ListBreaches(filter model.BreachFilter) ([]*model.SlaBreachLog, error) {
	if filter.Limit <= 0 {
		filter.Limit = 50
	}
	if filter.Offset < 0 {
		filter.Offset = 0
	}

	query := `
	SELECT id, task_id, policy_id, workspace_id,
		breach_type, breached_at, minutes_overdue,
		escalation_triggered, escalation_action, details
	FROM sla_breach_logs
	WHERE 1=1`

	var args []interface{}

	if filter.WorkspaceID != nil {
		query += " AND workspace_id = ?"
		args = append(args, *filter.WorkspaceID)
	}
	if filter.PolicyID != nil {
		query += " AND policy_id = ?"
		args = append(args, *filter.PolicyID)
	}
	if filter.TaskID != nil {
		query += " AND task_id = ?"
		args = append(args, *filter.TaskID)
	}
	if filter.BreachType != "" {
		query += " AND breach_type = ?"
		args = append(args, filter.BreachType)
	}

	query += " ORDER BY breached_at DESC LIMIT ? OFFSET ?"
	args = append(args, filter.Limit, filter.Offset)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	breaches := make([]*model.SlaBreachLog, 0)
	for rows.Next() {
		var (
			b      model.SlaBreachLog
			escVal int
		)

		err := rows.Scan(
			&b.ID,
			&b.TaskID,
			&b.PolicyID,
			&b.WorkspaceID,
			&b.BreachType,
			&b.BreachedAt,
			&b.MinutesOverdue,
			&escVal,
			&b.EscalationAction,
			&b.Details,
		)
		if err != nil {
			return nil, err
		}

		b.EscalationTriggered = escVal == 1
		breaches = append(breaches, &b)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return breaches, nil
}

func (r *slaRepositoryImpl) GetSlaMetrics(workspaceID int64) (*model.SlaMetrics, error) {
	metrics := &model.SlaMetrics{
		WorkspaceID:        workspaceID,
		BreachesByPriority: make(map[string]int),
	}

	// Total tracked tasks
	_ = r.db.QueryRow("SELECT COUNT(*) FROM task_sla_records WHERE workspace_id = ?", workspaceID).Scan(&metrics.TotalTrackedTasks)

	// Active SLA count
	_ = r.db.QueryRow("SELECT COUNT(*) FROM task_sla_records WHERE workspace_id = ? AND resolution_status = 'pending'", workspaceID).Scan(&metrics.ActiveSlaCount)

	// Response counts
	_ = r.db.QueryRow("SELECT COUNT(*) FROM task_sla_records WHERE workspace_id = ? AND response_status = 'met'", workspaceID).Scan(&metrics.MetResponseCount)
	_ = r.db.QueryRow("SELECT COUNT(*) FROM task_sla_records WHERE workspace_id = ? AND response_status = 'breached'", workspaceID).Scan(&metrics.BreachedResponseCount)

	// Resolution counts
	_ = r.db.QueryRow("SELECT COUNT(*) FROM task_sla_records WHERE workspace_id = ? AND resolution_status = 'met'", workspaceID).Scan(&metrics.MetResolutionCount)
	_ = r.db.QueryRow("SELECT COUNT(*) FROM task_sla_records WHERE workspace_id = ? AND resolution_status = 'breached'", workspaceID).Scan(&metrics.BreachedResolutionCount)

	// Compliance rate %
	totalDecided := metrics.MetResolutionCount + metrics.BreachedResolutionCount
	if totalDecided > 0 {
		metrics.ComplianceRate = (float64(metrics.MetResolutionCount) / float64(totalDecided)) * 100.0
	} else if metrics.TotalTrackedTasks > 0 {
		if metrics.BreachedResponseCount == 0 {
			metrics.ComplianceRate = 100.0
		} else {
			totalResp := metrics.MetResponseCount + metrics.BreachedResponseCount
			if totalResp > 0 {
				metrics.ComplianceRate = (float64(metrics.MetResponseCount) / float64(totalResp)) * 100.0
			}
		}
	} else {
		metrics.ComplianceRate = 100.0
	}

	// Average response minutes
	var avgResp sql.NullFloat64
	_ = r.db.QueryRow(`
		SELECT AVG((julianday(responded_at) - julianday(created_at)) * 1440.0)
		FROM task_sla_records
		WHERE workspace_id = ? AND responded_at IS NOT NULL`, workspaceID).Scan(&avgResp)
	if avgResp.Valid {
		metrics.AverageResponseMin = avgResp.Float64
	}

	// Average resolution minutes
	var avgRes sql.NullFloat64
	_ = r.db.QueryRow(`
		SELECT AVG((julianday(resolved_at) - julianday(created_at)) * 1440.0)
		FROM task_sla_records
		WHERE workspace_id = ? AND resolved_at IS NOT NULL`, workspaceID).Scan(&avgRes)
	if avgRes.Valid {
		metrics.AverageResolutionMin = avgRes.Float64
	}

	// Breaches grouped by task priority
	rows, err := r.db.Query(`
		SELECT t.priority, COUNT(*)
		FROM sla_breach_logs b
		JOIN tasks t ON t.id = b.task_id
		WHERE b.workspace_id = ?
		GROUP BY t.priority`, workspaceID)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var priority string
			var count int
			if err := rows.Scan(&priority, &count); err == nil {
				metrics.BreachesByPriority[priority] = count
			}
		}
	}

	return metrics, nil
}

