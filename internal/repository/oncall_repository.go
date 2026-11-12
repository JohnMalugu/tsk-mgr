package repository

import (
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/JohnMalugu/tsk-mgr-api/internal/model"
)

var (
	ErrScheduleNotFound         = errors.New("on-call schedule not found")
	ErrEscalationPolicyNotFound = errors.New("escalation policy not found")
	ErrEscalationStepNotFound   = errors.New("escalation step not found")
	ErrOnCallOverrideNotFound   = errors.New("on-call override not found")
	ErrOnCallPageNotFound       = errors.New("on-call page not found")
)

// OnCallRepository defines persistent operations for schedules, escalation policies, overrides, and pages
type OnCallRepository interface {
	// Schedules & Participants
	CreateSchedule(s *model.OnCallSchedule, participantIDs []string) (int64, error)
	GetScheduleByID(id int64) (*model.OnCallSchedule, error)
	ListSchedules(filter model.OnCallFilter) ([]model.OnCallSchedule, int, error)
	UpdateSchedule(s *model.OnCallSchedule) error
	DeleteSchedule(id int64) error
	SetParticipants(scheduleID int64, userIDs []string) error
	ListParticipants(scheduleID int64) ([]model.OnCallParticipant, error)

	// Computed Shifts
	SaveShifts(scheduleID int64, shifts []model.OnCallShift) error
	ListShifts(scheduleID int64, startTime, endTime time.Time) ([]model.OnCallShift, error)
	GetCurrentShift(scheduleID int64, now time.Time) (*model.OnCallShift, error)

	// Escalation Policies & Steps
	CreateEscalationPolicy(p *model.EscalationPolicy, steps []model.EscalationStep) (int64, error)
	GetEscalationPolicyByID(id int64) (*model.EscalationPolicy, error)
	GetDefaultEscalationPolicy(workspaceID int64) (*model.EscalationPolicy, error)
	ListEscalationPolicies(workspaceID int64) ([]model.EscalationPolicy, error)
	UpdateEscalationPolicy(p *model.EscalationPolicy) error
	DeleteEscalationPolicy(id int64) error
	SetEscalationSteps(policyID int64, steps []model.EscalationStep) error
	ListEscalationSteps(policyID int64) ([]model.EscalationStep, error)

	// Overrides
	CreateOverride(o *model.OnCallOverride) (int64, error)
	GetOverrideByID(id int64) (*model.OnCallOverride, error)
	ListOverrides(scheduleID int64) ([]model.OnCallOverride, error)
	DeleteOverride(id int64) error
	GetActiveOverride(scheduleID int64, now time.Time) (*model.OnCallOverride, error)

	// Pages & Alerts
	CreatePage(page *model.OnCallPage) (int64, error)
	GetPageByID(id int64) (*model.OnCallPage, error)
	ListPages(workspaceID int64, status string, limit, offset int) ([]model.OnCallPage, int, error)
	UpdatePage(page *model.OnCallPage) error

	// Metrics
	GetOnCallMetrics(workspaceID int64) (*model.OnCallMetrics, error)
}

type oncallRepositoryImpl struct {
	db *sql.DB
}

// NewOnCallRepository creates a new SQLite-backed OnCallRepository
func NewOnCallRepository(db *sql.DB) OnCallRepository {
	return &oncallRepositoryImpl{db: db}
}

func scanSchedule(scanner interface{ Scan(dest ...interface{}) error }) (*model.OnCallSchedule, error) {
	var s model.OnCallSchedule
	var projectID sql.NullInt64
	var endDate sql.NullTime
	var isActive int
	var projectName, creatorName sql.NullString
	var partCount int

	err := scanner.Scan(
		&s.ID, &s.WorkspaceID, &projectID, &s.Name, &s.Description,
		&s.Timezone, &s.RotationType, &s.RotationIntervalDays,
		&s.ShiftStartTime, &s.ShiftDurationHours, &s.HandoverDayOfWeek,
		&s.StartDate, &endDate, &isActive, &s.CreatedByID,
		&s.CreatedAt, &s.UpdatedAt,
		&projectName, &creatorName, &partCount,
	)
	if err != nil {
		return nil, err
	}
	if projectID.Valid {
		pid := projectID.Int64
		s.ProjectID = &pid
	}
	if endDate.Valid {
		ed := endDate.Time
		s.EndDate = &ed
	}
	s.IsActive = (isActive == 1)
	if projectName.Valid {
		s.ProjectName = projectName.String
	}
	if creatorName.Valid {
		s.CreatorName = creatorName.String
	}
	s.ParticipantsCount = partCount
	return &s, nil
}

func (r *oncallRepositoryImpl) CreateSchedule(s *model.OnCallSchedule, participantIDs []string) (int64, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	isActiveInt := 0
	if s.IsActive {
		isActiveInt = 1
	}

	now := time.Now()
	if s.CreatedAt.IsZero() {
		s.CreatedAt = now
	}
	if s.UpdatedAt.IsZero() {
		s.UpdatedAt = now
	}
	if s.Timezone == "" {
		s.Timezone = "UTC"
	}
	if s.RotationType == "" {
		s.RotationType = model.RotationWeekly
	}
	if s.RotationIntervalDays <= 0 {
		s.RotationIntervalDays = 7
	}
	if s.ShiftStartTime == "" {
		s.ShiftStartTime = "09:00"
	}
	if s.ShiftDurationHours <= 0 {
		s.ShiftDurationHours = 24
	}

	res, err := tx.Exec(`
		INSERT INTO oncall_schedules (
			workspace_id, project_id, name, description, timezone, rotation_type,
			rotation_interval_days, shift_start_time, shift_duration_hours, handover_day_of_week,
			start_date, end_date, is_active, created_by_id, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, s.WorkspaceID, s.ProjectID, s.Name, s.Description, s.Timezone, s.RotationType,
		s.RotationIntervalDays, s.ShiftStartTime, s.ShiftDurationHours, s.HandoverDayOfWeek,
		s.StartDate, s.EndDate, isActiveInt, s.CreatedByID, s.CreatedAt, s.UpdatedAt)
	if err != nil {
		return 0, err
	}

	scheduleID, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	s.ID = scheduleID

	for i, uid := range participantIDs {
		_, err = tx.Exec(`
			INSERT INTO oncall_participants (schedule_id, user_id, position_order, created_at)
			VALUES (?, ?, ?, ?)
		`, scheduleID, uid, i+1, now)
		if err != nil {
			return 0, err
		}
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return scheduleID, nil
}

func (r *oncallRepositoryImpl) GetScheduleByID(id int64) (*model.OnCallSchedule, error) {
	query := `
		SELECT s.id, s.workspace_id, s.project_id, s.name, s.description,
		       s.timezone, s.rotation_type, s.rotation_interval_days,
		       s.shift_start_time, s.shift_duration_hours, s.handover_day_of_week,
		       s.start_date, s.end_date, s.is_active, s.created_by_id,
		       s.created_at, s.updated_at,
		       COALESCE(p.name, ''), COALESCE(u.name, ''),
		       (SELECT COUNT(*) FROM oncall_participants WHERE schedule_id = s.id)
		FROM oncall_schedules s
		LEFT JOIN projects p ON s.project_id = p.id
		LEFT JOIN users u ON s.created_by_id = u.id
		WHERE s.id = ?
	`
	sc, err := scanSchedule(r.db.QueryRow(query, id))
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrScheduleNotFound
		}
		return nil, err
	}
	return sc, nil
}

func (r *oncallRepositoryImpl) ListSchedules(filter model.OnCallFilter) ([]model.OnCallSchedule, int, error) {
	var conditions []string
	var args []interface{}

	conditions = append(conditions, "s.workspace_id = ?")
	args = append(args, filter.WorkspaceID)

	if filter.ProjectID != nil {
		conditions = append(conditions, "s.project_id = ?")
		args = append(args, *filter.ProjectID)
	}
	if filter.IsActive != nil {
		act := 0
		if *filter.IsActive {
			act = 1
		}
		conditions = append(conditions, "s.is_active = ?")
		args = append(args, act)
	}
	if filter.Search != "" {
		conditions = append(conditions, "(s.name LIKE ? OR s.description LIKE ?)")
		pattern := "%" + filter.Search + "%"
		args = append(args, pattern, pattern)
	}

	whereClause := " WHERE " + strings.Join(conditions, " AND ")

	countQuery := `SELECT COUNT(*) FROM oncall_schedules s` + whereClause
	var totalCount int
	if err := r.db.QueryRow(countQuery, args...).Scan(&totalCount); err != nil {
		return nil, 0, err
	}

	dataQuery := `
		SELECT s.id, s.workspace_id, s.project_id, s.name, s.description,
		       s.timezone, s.rotation_type, s.rotation_interval_days,
		       s.shift_start_time, s.shift_duration_hours, s.handover_day_of_week,
		       s.start_date, s.end_date, s.is_active, s.created_by_id,
		       s.created_at, s.updated_at,
		       COALESCE(p.name, ''), COALESCE(u.name, ''),
		       (SELECT COUNT(*) FROM oncall_participants WHERE schedule_id = s.id)
		FROM oncall_schedules s
		LEFT JOIN projects p ON s.project_id = p.id
		LEFT JOIN users u ON s.created_by_id = u.id
	` + whereClause + " ORDER BY s.id DESC"

	if filter.Limit > 0 {
		dataQuery += " LIMIT ?"
		args = append(args, filter.Limit)
		if filter.Offset > 0 {
			dataQuery += " OFFSET ?"
			args = append(args, filter.Offset)
		}
	}

	rows, err := r.db.Query(dataQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var schedules []model.OnCallSchedule
	for rows.Next() {
		sc, err := scanSchedule(rows)
		if err != nil {
			return nil, 0, err
		}
		schedules = append(schedules, *sc)
	}
	return schedules, totalCount, nil
}

func (r *oncallRepositoryImpl) UpdateSchedule(s *model.OnCallSchedule) error {
	isActiveInt := 0
	if s.IsActive {
		isActiveInt = 1
	}
	s.UpdatedAt = time.Now()
	res, err := r.db.Exec(`
		UPDATE oncall_schedules
		SET name = ?, description = ?, timezone = ?, rotation_type = ?,
		    rotation_interval_days = ?, shift_start_time = ?, shift_duration_hours = ?,
		    handover_day_of_week = ?, start_date = ?, end_date = ?, is_active = ?,
		    updated_at = ?
		WHERE id = ?
	`, s.Name, s.Description, s.Timezone, s.RotationType, s.RotationIntervalDays,
		s.ShiftStartTime, s.ShiftDurationHours, s.HandoverDayOfWeek, s.StartDate,
		s.EndDate, isActiveInt, s.UpdatedAt, s.ID)
	if err != nil {
		return err
	}
	rowsAff, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAff == 0 {
		return ErrScheduleNotFound
	}
	return nil
}

func (r *oncallRepositoryImpl) DeleteSchedule(id int64) error {
	res, err := r.db.Exec(`DELETE FROM oncall_schedules WHERE id = ?`, id)
	if err != nil {
		return err
	}
	rowsAff, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAff == 0 {
		return ErrScheduleNotFound
	}
	return nil
}

func (r *oncallRepositoryImpl) SetParticipants(scheduleID int64, userIDs []string) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec(`DELETE FROM oncall_participants WHERE schedule_id = ?`, scheduleID)
	if err != nil {
		return err
	}

	now := time.Now()
	for i, uid := range userIDs {
		_, err = tx.Exec(`
			INSERT INTO oncall_participants (schedule_id, user_id, position_order, created_at)
			VALUES (?, ?, ?, ?)
		`, scheduleID, uid, i+1, now)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *oncallRepositoryImpl) ListParticipants(scheduleID int64) ([]model.OnCallParticipant, error) {
	rows, err := r.db.Query(`
		SELECT p.id, p.schedule_id, p.user_id, p.position_order, p.created_at,
		       COALESCE(u.name, ''), COALESCE(u.email, '')
		FROM oncall_participants p
		LEFT JOIN users u ON p.user_id = u.id
		WHERE p.schedule_id = ?
		ORDER BY p.position_order ASC
	`, scheduleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var participants []model.OnCallParticipant
	for rows.Next() {
		var p model.OnCallParticipant
		var name, email sql.NullString
		if err := rows.Scan(&p.ID, &p.ScheduleID, &p.UserID, &p.PositionOrder, &p.CreatedAt, &name, &email); err != nil {
			return nil, err
		}
		p.UserName = name.String
		p.UserEmail = email.String
		participants = append(participants, p)
	}
	return participants, nil
}

func (r *oncallRepositoryImpl) SaveShifts(scheduleID int64, shifts []model.OnCallShift) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec(`DELETE FROM oncall_shifts WHERE schedule_id = ?`, scheduleID)
	if err != nil {
		return err
	}

	now := time.Now()
	for _, shift := range shifts {
		_, err = tx.Exec(`
			INSERT INTO oncall_shifts (schedule_id, user_id, start_time, end_time, created_at)
			VALUES (?, ?, ?, ?, ?)
		`, scheduleID, shift.UserID, shift.StartTime, shift.EndTime, now)
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (r *oncallRepositoryImpl) ListShifts(scheduleID int64, startTime, endTime time.Time) ([]model.OnCallShift, error) {
	now := time.Now()
	rows, err := r.db.Query(`
		SELECT s.schedule_id, s.user_id, s.start_time, s.end_time,
		       COALESCE(u.name, ''), COALESCE(u.email, '')
		FROM oncall_shifts s
		LEFT JOIN users u ON s.user_id = u.id
		WHERE s.schedule_id = ? AND s.end_time >= ? AND s.start_time <= ?
		ORDER BY s.start_time ASC
	`, scheduleID, startTime, endTime)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var shifts []model.OnCallShift
	for rows.Next() {
		var s model.OnCallShift
		var name, email sql.NullString
		if err := rows.Scan(&s.ScheduleID, &s.UserID, &s.StartTime, &s.EndTime, &name, &email); err != nil {
			return nil, err
		}
		s.UserName = name.String
		s.UserEmail = email.String
		s.IsActiveNow = now.After(s.StartTime) && now.Before(s.EndTime)
		shifts = append(shifts, s)
	}
	return shifts, nil
}

func (r *oncallRepositoryImpl) GetCurrentShift(scheduleID int64, now time.Time) (*model.OnCallShift, error) {
	var s model.OnCallShift
	var name, email sql.NullString
	err := r.db.QueryRow(`
		SELECT s.schedule_id, s.user_id, s.start_time, s.end_time,
		       COALESCE(u.name, ''), COALESCE(u.email, '')
		FROM oncall_shifts s
		LEFT JOIN users u ON s.user_id = u.id
		WHERE s.schedule_id = ? AND s.start_time <= ? AND s.end_time > ?
		ORDER BY s.start_time ASC
		LIMIT 1
	`, scheduleID, now, now).Scan(&s.ScheduleID, &s.UserID, &s.StartTime, &s.EndTime, &name, &email)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	s.UserName = name.String
	s.UserEmail = email.String
	s.IsActiveNow = true
	return &s, nil
}

func scanEscalationPolicy(scanner interface{ Scan(dest ...interface{}) error }) (*model.EscalationPolicy, error) {
	var p model.EscalationPolicy
	var projectID sql.NullInt64
	var projectName sql.NullString
	var isDefaultInt int
	var stepsCount int

	err := scanner.Scan(
		&p.ID, &p.WorkspaceID, &projectID, &p.Name, &p.Description,
		&p.RepeatCount, &isDefaultInt, &p.CreatedAt, &p.UpdatedAt,
		&projectName, &stepsCount,
	)
	if err != nil {
		return nil, err
	}
	if projectID.Valid {
		pid := projectID.Int64
		p.ProjectID = &pid
	}
	p.IsDefault = (isDefaultInt == 1)
	if projectName.Valid {
		p.ProjectName = projectName.String
	}
	p.StepsCount = stepsCount
	return &p, nil
}

func (r *oncallRepositoryImpl) CreateEscalationPolicy(p *model.EscalationPolicy, steps []model.EscalationStep) (int64, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	isDefaultInt := 0
	if p.IsDefault {
		isDefaultInt = 1
		_, err = tx.Exec(`UPDATE escalation_policies SET is_default = 0 WHERE workspace_id = ?`, p.WorkspaceID)
		if err != nil {
			return 0, err
		}
	}

	now := time.Now()
	if p.CreatedAt.IsZero() {
		p.CreatedAt = now
	}
	if p.UpdatedAt.IsZero() {
		p.UpdatedAt = now
	}

	res, err := tx.Exec(`
		INSERT INTO escalation_policies (
			workspace_id, project_id, name, description, repeat_count, is_default, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, p.WorkspaceID, p.ProjectID, p.Name, p.Description, p.RepeatCount, isDefaultInt, p.CreatedAt, p.UpdatedAt)
	if err != nil {
		return 0, err
	}

	policyID, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	p.ID = policyID

	for i, step := range steps {
		stepOrder := step.StepOrder
		if stepOrder <= 0 {
			stepOrder = i + 1
		}
		delay := step.DelayMinutes
		if delay <= 0 {
			delay = 5
		}
		_, err = tx.Exec(`
			INSERT INTO escalation_steps (policy_id, step_order, target_type, target_id, delay_minutes, created_at)
			VALUES (?, ?, ?, ?, ?, ?)
		`, policyID, stepOrder, step.TargetType, step.TargetID, delay, now)
		if err != nil {
			return 0, err
		}
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return policyID, nil
}

func (r *oncallRepositoryImpl) GetEscalationPolicyByID(id int64) (*model.EscalationPolicy, error) {
	query := `
		SELECT p.id, p.workspace_id, p.project_id, p.name, p.description,
		       p.repeat_count, p.is_default, p.created_at, p.updated_at,
		       COALESCE(proj.name, ''),
		       (SELECT COUNT(*) FROM escalation_steps WHERE policy_id = p.id)
		FROM escalation_policies p
		LEFT JOIN projects proj ON p.project_id = proj.id
		WHERE p.id = ?
	`
	policy, err := scanEscalationPolicy(r.db.QueryRow(query, id))
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrEscalationPolicyNotFound
		}
		return nil, err
	}

	steps, err := r.ListEscalationSteps(id)
	if err != nil {
		return nil, err
	}
	policy.Steps = steps
	policy.StepsCount = len(steps)
	return policy, nil
}

func (r *oncallRepositoryImpl) GetDefaultEscalationPolicy(workspaceID int64) (*model.EscalationPolicy, error) {
	query := `
		SELECT p.id, p.workspace_id, p.project_id, p.name, p.description,
		       p.repeat_count, p.is_default, p.created_at, p.updated_at,
		       COALESCE(proj.name, ''),
		       (SELECT COUNT(*) FROM escalation_steps WHERE policy_id = p.id)
		FROM escalation_policies p
		LEFT JOIN projects proj ON p.project_id = proj.id
		WHERE p.workspace_id = ? AND p.is_default = 1
		LIMIT 1
	`
	policy, err := scanEscalationPolicy(r.db.QueryRow(query, workspaceID))
	if err != nil {
		if err == sql.ErrNoRows {
			fallbackQuery := `
				SELECT p.id, p.workspace_id, p.project_id, p.name, p.description,
				       p.repeat_count, p.is_default, p.created_at, p.updated_at,
				       COALESCE(proj.name, ''),
				       (SELECT COUNT(*) FROM escalation_steps WHERE policy_id = p.id)
				FROM escalation_policies p
				LEFT JOIN projects proj ON p.project_id = proj.id
				WHERE p.workspace_id = ?
				ORDER BY p.id ASC
				LIMIT 1
			`
			policy, err = scanEscalationPolicy(r.db.QueryRow(fallbackQuery, workspaceID))
			if err != nil {
				if err == sql.ErrNoRows {
					return nil, ErrEscalationPolicyNotFound
				}
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	steps, err := r.ListEscalationSteps(policy.ID)
	if err != nil {
		return nil, err
	}
	policy.Steps = steps
	policy.StepsCount = len(steps)
	return policy, nil
}

func (r *oncallRepositoryImpl) ListEscalationPolicies(workspaceID int64) ([]model.EscalationPolicy, error) {
	query := `
		SELECT p.id, p.workspace_id, p.project_id, p.name, p.description,
		       p.repeat_count, p.is_default, p.created_at, p.updated_at,
		       COALESCE(proj.name, ''),
		       (SELECT COUNT(*) FROM escalation_steps WHERE policy_id = p.id)
		FROM escalation_policies p
		LEFT JOIN projects proj ON p.project_id = proj.id
		WHERE p.workspace_id = ?
		ORDER BY p.is_default DESC, p.id ASC
	`
	rows, err := r.db.Query(query, workspaceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var policies []model.EscalationPolicy
	for rows.Next() {
		p, err := scanEscalationPolicy(rows)
		if err != nil {
			return nil, err
		}
		policies = append(policies, *p)
	}
	return policies, nil
}

func (r *oncallRepositoryImpl) UpdateEscalationPolicy(p *model.EscalationPolicy) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	isDefaultInt := 0
	if p.IsDefault {
		isDefaultInt = 1
		_, err = tx.Exec(`UPDATE escalation_policies SET is_default = 0 WHERE workspace_id = ? AND id != ?`, p.WorkspaceID, p.ID)
		if err != nil {
			return err
		}
	}

	p.UpdatedAt = time.Now()
	res, err := tx.Exec(`
		UPDATE escalation_policies
		SET name = ?, description = ?, repeat_count = ?, is_default = ?, updated_at = ?
		WHERE id = ?
	`, p.Name, p.Description, p.RepeatCount, isDefaultInt, p.UpdatedAt, p.ID)
	if err != nil {
		return err
	}

	rowsAff, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAff == 0 {
		return ErrEscalationPolicyNotFound
	}

	return tx.Commit()
}

func (r *oncallRepositoryImpl) DeleteEscalationPolicy(id int64) error {
	res, err := r.db.Exec(`DELETE FROM escalation_policies WHERE id = ?`, id)
	if err != nil {
		return err
	}
	rowsAff, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAff == 0 {
		return ErrEscalationPolicyNotFound
	}
	return nil
}

func (r *oncallRepositoryImpl) SetEscalationSteps(policyID int64, steps []model.EscalationStep) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec(`DELETE FROM escalation_steps WHERE policy_id = ?`, policyID)
	if err != nil {
		return err
	}

	now := time.Now()
	for i, step := range steps {
		stepOrder := step.StepOrder
		if stepOrder <= 0 {
			stepOrder = i + 1
		}
		delay := step.DelayMinutes
		if delay <= 0 {
			delay = 5
		}
		_, err = tx.Exec(`
			INSERT INTO escalation_steps (policy_id, step_order, target_type, target_id, delay_minutes, created_at)
			VALUES (?, ?, ?, ?, ?, ?)
		`, policyID, stepOrder, step.TargetType, step.TargetID, delay, now)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *oncallRepositoryImpl) ListEscalationSteps(policyID int64) ([]model.EscalationStep, error) {
	rows, err := r.db.Query(`
		SELECT id, policy_id, step_order, target_type, target_id, delay_minutes, created_at
		FROM escalation_steps
		WHERE policy_id = ?
		ORDER BY step_order ASC
	`, policyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var steps []model.EscalationStep
	for rows.Next() {
		var s model.EscalationStep
		if err := rows.Scan(&s.ID, &s.PolicyID, &s.StepOrder, &s.TargetType, &s.TargetID, &s.DelayMinutes, &s.CreatedAt); err != nil {
			return nil, err
		}
		steps = append(steps, s)
	}

	for i := range steps {
		switch steps[i].TargetType {
		case model.TargetUser:
			var uname sql.NullString
			_ = r.db.QueryRow(`SELECT name FROM users WHERE id = ?`, steps[i].TargetID).Scan(&uname)
			if uname.Valid {
				steps[i].TargetName = uname.String
			}
		case model.TargetSchedule:
			var sname sql.NullString
			_ = r.db.QueryRow(`SELECT name FROM oncall_schedules WHERE id = ?`, steps[i].TargetID).Scan(&sname)
			if sname.Valid {
				steps[i].TargetName = sname.String
			}
		case model.TargetWebhook:
			steps[i].TargetName = "Webhook Endpoint"
		}
	}
	return steps, nil
}

func scanOverride(scanner interface{ Scan(dest ...interface{}) error }, now time.Time) (*model.OnCallOverride, error) {
	var o model.OnCallOverride
	var origUserID sql.NullString
	var origUserName, overUserName, schedName sql.NullString

	err := scanner.Scan(
		&o.ID, &o.ScheduleID, &origUserID, &o.OverrideUserID,
		&o.StartTime, &o.EndTime, &o.Reason, &o.CreatedByID, &o.CreatedAt,
		&origUserName, &overUserName, &schedName,
	)
	if err != nil {
		return nil, err
	}
	if origUserID.Valid {
		uid := origUserID.String
		o.OriginalUserID = &uid
	}
	if origUserName.Valid {
		o.OriginalUserName = origUserName.String
	}
	if overUserName.Valid {
		o.OverrideUserName = overUserName.String
	}
	if schedName.Valid {
		o.ScheduleName = schedName.String
	}
	o.IsActiveNow = now.After(o.StartTime) && now.Before(o.EndTime)
	return &o, nil
}

func (r *oncallRepositoryImpl) CreateOverride(o *model.OnCallOverride) (int64, error) {
	now := time.Now()
	if o.CreatedAt.IsZero() {
		o.CreatedAt = now
	}
	res, err := r.db.Exec(`
		INSERT INTO oncall_overrides (
			schedule_id, original_user_id, override_user_id, start_time, end_time, reason, created_by_id, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, o.ScheduleID, o.OriginalUserID, o.OverrideUserID, o.StartTime, o.EndTime, o.Reason, o.CreatedByID, o.CreatedAt)
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	o.ID = id
	return id, nil
}

func (r *oncallRepositoryImpl) GetOverrideByID(id int64) (*model.OnCallOverride, error) {
	query := `
		SELECT o.id, o.schedule_id, o.original_user_id, o.override_user_id,
		       o.start_time, o.end_time, o.reason, o.created_by_id, o.created_at,
		       COALESCE(u_orig.name, ''), COALESCE(u_over.name, ''), COALESCE(s.name, '')
		FROM oncall_overrides o
		LEFT JOIN users u_orig ON o.original_user_id = u_orig.id
		LEFT JOIN users u_over ON o.override_user_id = u_over.id
		LEFT JOIN oncall_schedules s ON o.schedule_id = s.id
		WHERE o.id = ?
	`
	o, err := scanOverride(r.db.QueryRow(query, id), time.Now())
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrOnCallOverrideNotFound
		}
		return nil, err
	}
	return o, nil
}

func (r *oncallRepositoryImpl) ListOverrides(scheduleID int64) ([]model.OnCallOverride, error) {
	now := time.Now()
	query := `
		SELECT o.id, o.schedule_id, o.original_user_id, o.override_user_id,
		       o.start_time, o.end_time, o.reason, o.created_by_id, o.created_at,
		       COALESCE(u_orig.name, ''), COALESCE(u_over.name, ''), COALESCE(s.name, '')
		FROM oncall_overrides o
		LEFT JOIN users u_orig ON o.original_user_id = u_orig.id
		LEFT JOIN users u_over ON o.override_user_id = u_over.id
		LEFT JOIN oncall_schedules s ON o.schedule_id = s.id
		WHERE o.schedule_id = ?
		ORDER BY o.start_time DESC
	`
	rows, err := r.db.Query(query, scheduleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var overrides []model.OnCallOverride
	for rows.Next() {
		o, err := scanOverride(rows, now)
		if err != nil {
			return nil, err
		}
		overrides = append(overrides, *o)
	}
	return overrides, nil
}

func (r *oncallRepositoryImpl) DeleteOverride(id int64) error {
	res, err := r.db.Exec(`DELETE FROM oncall_overrides WHERE id = ?`, id)
	if err != nil {
		return err
	}
	rowsAff, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAff == 0 {
		return ErrOnCallOverrideNotFound
	}
	return nil
}

func (r *oncallRepositoryImpl) GetActiveOverride(scheduleID int64, now time.Time) (*model.OnCallOverride, error) {
	query := `
		SELECT o.id, o.schedule_id, o.original_user_id, o.override_user_id,
		       o.start_time, o.end_time, o.reason, o.created_by_id, o.created_at,
		       COALESCE(u_orig.name, ''), COALESCE(u_over.name, ''), COALESCE(s.name, '')
		FROM oncall_overrides o
		LEFT JOIN users u_orig ON o.original_user_id = u_orig.id
		LEFT JOIN users u_over ON o.override_user_id = u_over.id
		LEFT JOIN oncall_schedules s ON o.schedule_id = s.id
		WHERE o.schedule_id = ? AND o.start_time <= ? AND o.end_time > ?
		ORDER BY o.id DESC
		LIMIT 1
	`
	o, err := scanOverride(r.db.QueryRow(query, scheduleID, now, now), now)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return o, nil
}

func scanPage(scanner interface{ Scan(dest ...interface{}) error }) (*model.OnCallPage, error) {
	var page model.OnCallPage
	var incidentID sql.NullInt64
	var pagedUserID, ackBy sql.NullString
	var ackAt, resAt sql.NullTime
	var policyName, incTitle, pagedUserName, ackUserName sql.NullString
	var incNumber sql.NullInt64

	err := scanner.Scan(
		&page.ID, &page.WorkspaceID, &page.PolicyID, &incidentID,
		&page.Title, &page.Summary, &page.Status, &page.CurrentStepOrder,
		&pagedUserID, &ackBy, &ackAt, &resAt, &page.TriggerSource,
		&page.CreatedByID, &page.CreatedAt, &page.UpdatedAt,
		&policyName, &incNumber, &incTitle, &pagedUserName, &ackUserName,
	)
	if err != nil {
		return nil, err
	}

	if incidentID.Valid {
		iid := incidentID.Int64
		page.IncidentID = &iid
	}
	if pagedUserID.Valid {
		puid := pagedUserID.String
		page.PagedUserID = &puid
	}
	if ackBy.Valid {
		ab := ackBy.String
		page.AcknowledgedBy = &ab
	}
	if ackAt.Valid {
		at := ackAt.Time
		page.AcknowledgedAt = &at
	}
	if resAt.Valid {
		rt := resAt.Time
		page.ResolvedAt = &rt
	}

	if policyName.Valid {
		page.PolicyName = policyName.String
	}
	if incNumber.Valid {
		page.IncidentNumber = int(incNumber.Int64)
	}
	if incTitle.Valid {
		page.IncidentTitle = incTitle.String
	}
	if pagedUserName.Valid {
		page.PagedUserName = pagedUserName.String
	}
	if ackUserName.Valid {
		page.AcknowledgedUserName = ackUserName.String
	}
	return &page, nil
}

func (r *oncallRepositoryImpl) CreatePage(page *model.OnCallPage) (int64, error) {
	now := time.Now()
	if page.CreatedAt.IsZero() {
		page.CreatedAt = now
	}
	if page.UpdatedAt.IsZero() {
		page.UpdatedAt = now
	}
	if page.Status == "" {
		page.Status = model.PageStatusTriggered
	}
	if page.CurrentStepOrder <= 0 {
		page.CurrentStepOrder = 1
	}
	if page.TriggerSource == "" {
		page.TriggerSource = model.PageTriggerManual
	}

	res, err := r.db.Exec(`
		INSERT INTO oncall_pages (
			workspace_id, policy_id, incident_id, title, summary, status,
			current_step_order, paged_user_id, acknowledged_by, acknowledged_at, resolved_at,
			trigger_source, created_by_id, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, page.WorkspaceID, page.PolicyID, page.IncidentID, page.Title, page.Summary, page.Status,
		page.CurrentStepOrder, page.PagedUserID, page.AcknowledgedBy, page.AcknowledgedAt, page.ResolvedAt,
		page.TriggerSource, page.CreatedByID, page.CreatedAt, page.UpdatedAt)
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	page.ID = id
	return id, nil
}

func (r *oncallRepositoryImpl) GetPageByID(id int64) (*model.OnCallPage, error) {
	query := `
		SELECT p.id, p.workspace_id, p.policy_id, p.incident_id,
		       p.title, p.summary, p.status, p.current_step_order,
		       p.paged_user_id, p.acknowledged_by, p.acknowledged_at, p.resolved_at,
		       p.trigger_source, p.created_by_id, p.created_at, p.updated_at,
		       COALESCE(ep.name, ''), inc.incident_number, COALESCE(inc.title, ''),
		       COALESCE(u_paged.name, ''), COALESCE(u_ack.name, '')
		FROM oncall_pages p
		LEFT JOIN escalation_policies ep ON p.policy_id = ep.id
		LEFT JOIN incidents inc ON p.incident_id = inc.id
		LEFT JOIN users u_paged ON p.paged_user_id = u_paged.id
		LEFT JOIN users u_ack ON p.acknowledged_by = u_ack.id
		WHERE p.id = ?
	`
	page, err := scanPage(r.db.QueryRow(query, id))
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrOnCallPageNotFound
		}
		return nil, err
	}
	return page, nil
}

func (r *oncallRepositoryImpl) ListPages(workspaceID int64, status string, limit, offset int) ([]model.OnCallPage, int, error) {
	var conditions []string
	var args []interface{}

	conditions = append(conditions, "p.workspace_id = ?")
	args = append(args, workspaceID)

	if strings.TrimSpace(status) != "" {
		conditions = append(conditions, "p.status = ?")
		args = append(args, strings.TrimSpace(status))
	}

	whereClause := " WHERE " + strings.Join(conditions, " AND ")

	countQuery := `SELECT COUNT(*) FROM oncall_pages p` + whereClause
	var totalCount int
	if err := r.db.QueryRow(countQuery, args...).Scan(&totalCount); err != nil {
		return nil, 0, err
	}

	dataQuery := `
		SELECT p.id, p.workspace_id, p.policy_id, p.incident_id,
		       p.title, p.summary, p.status, p.current_step_order,
		       p.paged_user_id, p.acknowledged_by, p.acknowledged_at, p.resolved_at,
		       p.trigger_source, p.created_by_id, p.created_at, p.updated_at,
		       COALESCE(ep.name, ''), inc.incident_number, COALESCE(inc.title, ''),
		       COALESCE(u_paged.name, ''), COALESCE(u_ack.name, '')
		FROM oncall_pages p
		LEFT JOIN escalation_policies ep ON p.policy_id = ep.id
		LEFT JOIN incidents inc ON p.incident_id = inc.id
		LEFT JOIN users u_paged ON p.paged_user_id = u_paged.id
		LEFT JOIN users u_ack ON p.acknowledged_by = u_ack.id
	` + whereClause + " ORDER BY p.id DESC"

	if limit > 0 {
		dataQuery += " LIMIT ?"
		args = append(args, limit)
		if offset > 0 {
			dataQuery += " OFFSET ?"
			args = append(args, offset)
		}
	}

	rows, err := r.db.Query(dataQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var pages []model.OnCallPage
	for rows.Next() {
		pg, err := scanPage(rows)
		if err != nil {
			return nil, 0, err
		}
		pages = append(pages, *pg)
	}
	return pages, totalCount, nil
}

func (r *oncallRepositoryImpl) UpdatePage(page *model.OnCallPage) error {
	page.UpdatedAt = time.Now()
	res, err := r.db.Exec(`
		UPDATE oncall_pages
		SET status = ?, current_step_order = ?, paged_user_id = ?,
		    acknowledged_by = ?, acknowledged_at = ?, resolved_at = ?, updated_at = ?
		WHERE id = ?
	`, page.Status, page.CurrentStepOrder, page.PagedUserID,
		page.AcknowledgedBy, page.AcknowledgedAt, page.ResolvedAt, page.UpdatedAt, page.ID)
	if err != nil {
		return err
	}
	rowsAff, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAff == 0 {
		return ErrOnCallPageNotFound
	}
	return nil
}

func (r *oncallRepositoryImpl) GetOnCallMetrics(workspaceID int64) (*model.OnCallMetrics, error) {
	metrics := &model.OnCallMetrics{
		PagesByStatus: make(map[string]int),
	}

	now := time.Now()

	// Total and active schedules
	_ = r.db.QueryRow(`SELECT COUNT(*) FROM oncall_schedules WHERE workspace_id = ?`, workspaceID).Scan(&metrics.TotalSchedules)
	_ = r.db.QueryRow(`SELECT COUNT(*) FROM oncall_schedules WHERE workspace_id = ? AND is_active = 1`, workspaceID).Scan(&metrics.ActiveSchedules)

	// Total policies
	_ = r.db.QueryRow(`SELECT COUNT(*) FROM escalation_policies WHERE workspace_id = ?`, workspaceID).Scan(&metrics.TotalPolicies)

	// Active overrides
	_ = r.db.QueryRow(`
		SELECT COUNT(*)
		FROM oncall_overrides o
		JOIN oncall_schedules s ON o.schedule_id = s.id
		WHERE s.workspace_id = ? AND o.start_time <= ? AND o.end_time > ?
	`, workspaceID, now, now).Scan(&metrics.ActiveOverrides)

	// Pages stats
	_ = r.db.QueryRow(`SELECT COUNT(*) FROM oncall_pages WHERE workspace_id = ?`, workspaceID).Scan(&metrics.TotalPagesTriggered)
	_ = r.db.QueryRow(`SELECT COUNT(*) FROM oncall_pages WHERE workspace_id = ? AND acknowledged_at IS NOT NULL`, workspaceID).Scan(&metrics.AcknowledgedPages)
	_ = r.db.QueryRow(`SELECT COUNT(*) FROM oncall_pages WHERE workspace_id = ? AND resolved_at IS NOT NULL`, workspaceID).Scan(&metrics.ResolvedPages)

	// Mean times
	_ = r.db.QueryRow(`
		SELECT COALESCE(AVG((julianday(acknowledged_at) - julianday(created_at)) * 1440.0), 0.0)
		FROM oncall_pages
		WHERE workspace_id = ? AND acknowledged_at IS NOT NULL
	`, workspaceID).Scan(&metrics.MeanTimeToAcknowledgeMinutes)

	_ = r.db.QueryRow(`
		SELECT COALESCE(AVG((julianday(resolved_at) - julianday(created_at)) * 1440.0), 0.0)
		FROM oncall_pages
		WHERE workspace_id = ? AND resolved_at IS NOT NULL
	`, workspaceID).Scan(&metrics.MeanTimeToResolveMinutes)

	// Breakdown by status
	rows, err := r.db.Query(`SELECT status, COUNT(*) FROM oncall_pages WHERE workspace_id = ? GROUP BY status`, workspaceID)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var st string
			var count int
			if err := rows.Scan(&st, &count); err == nil {
				metrics.PagesByStatus[st] = count
			}
		}
	}

	return metrics, nil
}
