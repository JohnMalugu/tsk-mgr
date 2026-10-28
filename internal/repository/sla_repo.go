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
	return nil
}

func (r *slaRepositoryImpl) GetEscalationRules(policyID int64) ([]model.EscalationRule, error) {
	return nil, nil
}

func (r *slaRepositoryImpl) DeleteEscalationRule(ruleID int64) error {
	return nil
}

func (r *slaRepositoryImpl) CreateTaskSlaRecord(record *model.TaskSlaRecord) (int64, error) {
	return 0, nil
}

func (r *slaRepositoryImpl) GetTaskSlaRecord(id int64) (*model.TaskSlaRecord, error) {
	return nil, nil
}

func (r *slaRepositoryImpl) GetTaskSlaRecordByTaskID(taskID int) (*model.TaskSlaRecord, error) {
	return nil, nil
}

func (r *slaRepositoryImpl) UpdateTaskSlaRecord(record *model.TaskSlaRecord) error {
	return nil
}

func (r *slaRepositoryImpl) LogBreach(breach *model.SlaBreachLog) (int64, error) {
	return 0, nil
}

func (r *slaRepositoryImpl) ListBreaches(filter model.BreachFilter) ([]*model.SlaBreachLog, error) {
	return nil, nil
}

func (r *slaRepositoryImpl) GetSlaMetrics(workspaceID int64) (*model.SlaMetrics, error) {
	return nil, nil
}
