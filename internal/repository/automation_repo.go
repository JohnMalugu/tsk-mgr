package repository

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/JohnMalugu/tsk-mgr-api/internal/model"
)

// ErrAutomationRuleNotFound is returned when an automation rule cannot be found
var ErrAutomationRuleNotFound = errors.New("automation rule not found")

// AutomationRepository defines persistent operations for automation rules and execution logs
type AutomationRepository interface {
	CreateRule(rule *model.AutomationRule) (int64, error)
	GetRuleByID(id int64) (*model.AutomationRule, error)
	ListRules(workspaceID int64, activeOnly *bool, triggerType string) ([]*model.AutomationRule, error)
	UpdateRule(rule *model.AutomationRule) error
	DeleteRule(id int64) error
	ToggleRuleActive(id int64) (bool, error)
	RecordExecutionLog(log *model.AutomationExecutionLog) (int64, error)
	ListExecutionLogs(filter model.AutomationLogFilter) ([]*model.AutomationExecutionLog, error)
	IncrementRuleExecutionCount(id int64) error
	GetAutomationMetrics(workspaceID int64) (*model.AutomationMetrics, error)
}

type sqlAutomationRepository struct {
	db *sql.DB
}

// NewAutomationRepository creates a new SQLite-backed AutomationRepository
func NewAutomationRepository(db *sql.DB) AutomationRepository {
	return &sqlAutomationRepository{db: db}
}

func (r *sqlAutomationRepository) CreateRule(rule *model.AutomationRule) (int64, error) {
	now := time.Now()
	if rule.CreatedAt.IsZero() {
		rule.CreatedAt = now
	}
	if rule.UpdatedAt.IsZero() {
		rule.UpdatedAt = now
	}

	condJSON, err := json.Marshal(rule.Conditions)
	if err != nil {
		condJSON = []byte("[]")
	}

	actJSON, err := json.Marshal(rule.Actions)
	if err != nil {
		actJSON = []byte("[]")
	}

	activeInt := 0
	if rule.IsActive {
		activeInt = 1
	}

	query := `
	INSERT INTO automation_rules (
		workspace_id, created_by_user_id, name, description, trigger_type,
		conditions_json, actions_json, is_active, execution_count,
		last_triggered_at, created_at, updated_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	res, err := r.db.Exec(query,
		rule.WorkspaceID, rule.CreatedByUserID, rule.Name, rule.Description, rule.TriggerType,
		string(condJSON), string(actJSON), activeInt, rule.ExecutionCount,
		rule.LastTriggeredAt, rule.CreatedAt, rule.UpdatedAt,
	)
	if err != nil {
		return 0, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	rule.ID = id
	return id, nil
}

func (r *sqlAutomationRepository) GetRuleByID(id int64) (*model.AutomationRule, error) {
	query := `
	SELECT id, workspace_id, created_by_user_id, name, description, trigger_type,
	       conditions_json, actions_json, is_active, execution_count,
	       last_triggered_at, created_at, updated_at
	FROM automation_rules
	WHERE id = ?`

	row := r.db.QueryRow(query, id)

	var rule model.AutomationRule
	var condJSON, actJSON string
	var activeInt int

	err := row.Scan(
		&rule.ID, &rule.WorkspaceID, &rule.CreatedByUserID, &rule.Name, &rule.Description, &rule.TriggerType,
		&condJSON, &actJSON, &activeInt, &rule.ExecutionCount,
		&rule.LastTriggeredAt, &rule.CreatedAt, &rule.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrAutomationRuleNotFound
		}
		return nil, err
	}

	rule.IsActive = activeInt == 1
	if condJSON != "" {
		_ = json.Unmarshal([]byte(condJSON), &rule.Conditions)
	}
	if actJSON != "" {
		_ = json.Unmarshal([]byte(actJSON), &rule.Actions)
	}

	return &rule, nil
}


func (r *sqlAutomationRepository) ListRules(workspaceID int64, activeOnly *bool, triggerType string) ([]*model.AutomationRule, error) {
	var conditions []string
	var args []interface{}

	conditions = append(conditions, "workspace_id = ?")
	args = append(args, workspaceID)

	if activeOnly != nil {
		if *activeOnly {
			conditions = append(conditions, "is_active = 1")
		} else {
			conditions = append(conditions, "is_active = 0")
		}
	}

	if strings.TrimSpace(triggerType) != "" {
		conditions = append(conditions, "trigger_type = ?")
		args = append(args, strings.TrimSpace(triggerType))
	}

	query := fmt.Sprintf(`
	SELECT id, workspace_id, created_by_user_id, name, description, trigger_type,
	       conditions_json, actions_json, is_active, execution_count,
	       last_triggered_at, created_at, updated_at
	FROM automation_rules
	WHERE %s
	ORDER BY created_at DESC`, strings.Join(conditions, " AND "))

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rules []*model.AutomationRule
	for rows.Next() {
		var rule model.AutomationRule
		var condJSON, actJSON string
		var activeInt int

		err := rows.Scan(
			&rule.ID, &rule.WorkspaceID, &rule.CreatedByUserID, &rule.Name, &rule.Description, &rule.TriggerType,
			&condJSON, &actJSON, &activeInt, &rule.ExecutionCount,
			&rule.LastTriggeredAt, &rule.CreatedAt, &rule.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		rule.IsActive = activeInt == 1
		if condJSON != "" {
			_ = json.Unmarshal([]byte(condJSON), &rule.Conditions)
		}
		if actJSON != "" {
			_ = json.Unmarshal([]byte(actJSON), &rule.Actions)
		}

		rules = append(rules, &rule)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return rules, nil
}


func (r *sqlAutomationRepository) UpdateRule(rule *model.AutomationRule) error {
	rule.UpdatedAt = time.Now()

	condJSON, err := json.Marshal(rule.Conditions)
	if err != nil {
		condJSON = []byte("[]")
	}

	actJSON, err := json.Marshal(rule.Actions)
	if err != nil {
		actJSON = []byte("[]")
	}

	activeInt := 0
	if rule.IsActive {
		activeInt = 1
	}

	query := `
	UPDATE automation_rules SET
		name = ?,
		description = ?,
		trigger_type = ?,
		conditions_json = ?,
		actions_json = ?,
		is_active = ?,
		updated_at = ?
	WHERE id = ?`

	res, err := r.db.Exec(query,
		rule.Name, rule.Description, rule.TriggerType,
		string(condJSON), string(actJSON), activeInt, rule.UpdatedAt,
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
		return ErrAutomationRuleNotFound
	}

	return nil
}

func (r *sqlAutomationRepository) DeleteRule(id int64) error {
	query := `DELETE FROM automation_rules WHERE id = ?`
	res, err := r.db.Exec(query, id)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrAutomationRuleNotFound
	}

	return nil
}

func (r *sqlAutomationRepository) ToggleRuleActive(id int64) (bool, error) {
	now := time.Now()
	query := `
	UPDATE automation_rules SET
		is_active = CASE WHEN is_active = 1 THEN 0 ELSE 1 END,
		updated_at = ?
	WHERE id = ?`

	res, err := r.db.Exec(query, now, id)
	if err != nil {
		return false, err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return false, err
	}
	if rows == 0 {
		return false, ErrAutomationRuleNotFound
	}

	var activeInt int
	err = r.db.QueryRow("SELECT is_active FROM automation_rules WHERE id = ?", id).Scan(&activeInt)
	if err != nil {
		return false, err
	}

	return activeInt == 1, nil
}


func (r *sqlAutomationRepository) RecordExecutionLog(log *model.AutomationExecutionLog) (int64, error) {
	if log.ExecutedAt.IsZero() {
		log.ExecutedAt = time.Now()
	}

	actionsJSON, err := json.Marshal(log.ActionsExecuted)
	if err != nil {
		actionsJSON = []byte("[]")
	}

	query := `
	INSERT INTO automation_execution_logs (
		rule_id, workspace_id, trigger_event, status,
		actions_executed, error_message, duration_ms, executed_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`

	res, err := r.db.Exec(query,
		log.RuleID, log.WorkspaceID, log.TriggerEvent, log.Status,
		string(actionsJSON), log.ErrorMessage, log.DurationMs, log.ExecutedAt,
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

func (r *sqlAutomationRepository) ListExecutionLogs(filter model.AutomationLogFilter) ([]*model.AutomationExecutionLog, error) {
	var conditions []string
	var args []interface{}

	conditions = append(conditions, "workspace_id = ?")
	args = append(args, filter.WorkspaceID)

	if filter.RuleID != nil {
		conditions = append(conditions, "rule_id = ?")
		args = append(args, *filter.RuleID)
	}

	if strings.TrimSpace(filter.Status) != "" {
		conditions = append(conditions, "status = ?")
		args = append(args, strings.TrimSpace(filter.Status))
	}

	limit := filter.Limit
	if limit <= 0 {
		limit = 50
	} else if limit > 100 {
		limit = 100
	}

	offset := filter.Offset
	if offset < 0 {
		offset = 0
	}

	query := fmt.Sprintf(`
	SELECT id, rule_id, workspace_id, trigger_event, status,
	       actions_executed, error_message, duration_ms, executed_at
	FROM automation_execution_logs
	WHERE %s
	ORDER BY executed_at DESC
	LIMIT ? OFFSET ?`, strings.Join(conditions, " AND "))

	args = append(args, limit, offset)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []*model.AutomationExecutionLog
	for rows.Next() {
		var l model.AutomationExecutionLog
		var actionsJSON string

		err := rows.Scan(
			&l.ID, &l.RuleID, &l.WorkspaceID, &l.TriggerEvent, &l.Status,
			&actionsJSON, &l.ErrorMessage, &l.DurationMs, &l.ExecutedAt,
		)
		if err != nil {
			return nil, err
		}

		if actionsJSON != "" {
			_ = json.Unmarshal([]byte(actionsJSON), &l.ActionsExecuted)
		}

		logs = append(logs, &l)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return logs, nil
}


func (r *sqlAutomationRepository) IncrementRuleExecutionCount(id int64) error {
	now := time.Now()
	query := `
	UPDATE automation_rules SET
		execution_count = execution_count + 1,
		last_triggered_at = ?,
		updated_at = ?
	WHERE id = ?`

	res, err := r.db.Exec(query, now, now, id)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrAutomationRuleNotFound
	}

	return nil
}

func (r *sqlAutomationRepository) GetAutomationMetrics(workspaceID int64) (*model.AutomationMetrics, error) {
	metrics := &model.AutomationMetrics{
		WorkspaceID: workspaceID,
	}

	ruleQuery := `
	SELECT COUNT(*), COALESCE(SUM(CASE WHEN is_active = 1 THEN 1 ELSE 0 END), 0)
	FROM automation_rules
	WHERE workspace_id = ?`

	err := r.db.QueryRow(ruleQuery, workspaceID).Scan(&metrics.TotalRules, &metrics.ActiveRules)
	if err != nil {
		return nil, err
	}

	logQuery := `
	SELECT COUNT(*),
	       COALESCE(SUM(CASE WHEN status = 'success' THEN 1 ELSE 0 END), 0),
	       COALESCE(SUM(CASE WHEN status = 'failed' THEN 1 ELSE 0 END), 0),
	       COALESCE(SUM(CASE WHEN status = 'skipped' THEN 1 ELSE 0 END), 0)
	FROM automation_execution_logs
	WHERE workspace_id = ?`

	err = r.db.QueryRow(logQuery, workspaceID).Scan(
		&metrics.TotalExecutions,
		&metrics.SuccessfulExecutions,
		&metrics.FailedExecutions,
		&metrics.SkippedExecutions,
	)
	if err != nil {
		return nil, err
	}

	topQuery := `
	SELECT id, name, execution_count, last_triggered_at
	FROM automation_rules
	WHERE workspace_id = ? AND execution_count > 0
	ORDER BY execution_count DESC
	LIMIT 5`

	rows, err := r.db.Query(topQuery, workspaceID)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var s model.AutomationRuleStat
			if err := rows.Scan(&s.RuleID, &s.RuleName, &s.ExecutionCount, &s.LastTriggeredAt); err == nil {
				metrics.TopRules = append(metrics.TopRules, s)
			}
		}
	}

	return metrics, nil
}

