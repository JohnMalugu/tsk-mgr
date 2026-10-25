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
	return nil
}

func (r *sqlAutomationRepository) DeleteRule(id int64) error {
	return nil
}

func (r *sqlAutomationRepository) ToggleRuleActive(id int64) (bool, error) {
	return false, nil
}

func (r *sqlAutomationRepository) RecordExecutionLog(log *model.AutomationExecutionLog) (int64, error) {
	return 0, nil
}

func (r *sqlAutomationRepository) ListExecutionLogs(filter model.AutomationLogFilter) ([]*model.AutomationExecutionLog, error) {
	return nil, nil
}

func (r *sqlAutomationRepository) IncrementRuleExecutionCount(id int64) error {
	return nil
}

func (r *sqlAutomationRepository) GetAutomationMetrics(workspaceID int64) (*model.AutomationMetrics, error) {
	return nil, nil
}
