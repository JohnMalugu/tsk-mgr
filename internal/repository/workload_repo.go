package repository

import (
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/JohnMalugu/tsk-mgr-api/internal/model"
)

var (
	ErrProfileNotFound    = errors.New("resource profile not found")
	ErrProfileExists      = errors.New("resource profile already exists for user in workspace")
	ErrTimeOffNotFound    = errors.New("resource time-off not found")
	ErrAllocationNotFound = errors.New("project resource allocation not found")
)

// WorkloadRepository manages persistence for capacity profiles, time-offs, project allocations, and workload queries
type WorkloadRepository interface {
	CreateProfile(profile *model.ResourceProfile) (int64, error)
	GetProfileByID(id int64) (*model.ResourceProfile, error)
	GetProfileByUserID(workspaceID int64, userID string) (*model.ResourceProfile, error)
	ListProfiles(workspaceID int64, isActive *bool) ([]*model.ResourceProfile, error)
	UpdateProfile(profile *model.ResourceProfile) error
	DeleteProfile(id int64) error

	CreateTimeOff(timeOff *model.ResourceTimeOff) (int64, error)
	GetTimeOffByID(id int64) (*model.ResourceTimeOff, error)
	ListTimeOffs(workspaceID int64, userID string, from, to *time.Time) ([]model.ResourceTimeOff, error)
	UpdateTimeOff(timeOff *model.ResourceTimeOff) error
	DeleteTimeOff(id int64) error

	CreateAllocation(alloc *model.ProjectResourceAllocation) (int64, error)
	GetAllocationByID(id int64) (*model.ProjectResourceAllocation, error)
	ListAllocations(workspaceID int64, projectID int, userID string, status string) ([]model.ProjectResourceAllocation, error)
	UpdateAllocation(alloc *model.ProjectResourceAllocation) error
	DeleteAllocation(id int64) error

	GetUserTasksForWorkload(workspaceID int64, userID string, from, to time.Time) ([]model.Task, error)
	GetWorkspaceWorkloadMetrics(workspaceID int64) (*model.WorkloadMetrics, error)
}

type workloadRepositoryImpl struct {
	db *sql.DB
}

// NewWorkloadRepository creates a new instance of WorkloadRepository
func NewWorkloadRepository(db *sql.DB) WorkloadRepository {
	return &workloadRepositoryImpl{db: db}
}

func (r *workloadRepositoryImpl) CreateProfile(profile *model.ResourceProfile) (int64, error) {
	// Check if already exists for this workspace and user
	var count int
	checkQuery := `SELECT COUNT(*) FROM resource_profiles WHERE workspace_id = ? AND user_id = ?`
	err := r.db.QueryRow(checkQuery, profile.WorkspaceID, profile.UserID).Scan(&count)
	if err != nil {
		return 0, err
	}
	if count > 0 {
		return 0, ErrProfileExists
	}

	query := `
	INSERT INTO resource_profiles (
		workspace_id, user_id, job_title, weekly_capacity_hours, daily_capacity_hours,
		skills, hourly_rate, currency, is_active, created_at, updated_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	now := time.Now()
	if profile.CreatedAt.IsZero() {
		profile.CreatedAt = now
	}
	profile.UpdatedAt = now

	skillsJSON, err := json.Marshal(profile.Skills)
	if err != nil {
		skillsJSON = []byte("[]")
	}

	isActiveInt := 1
	if !profile.IsActive {
		isActiveInt = 0
	}

	res, err := r.db.Exec(query,
		profile.WorkspaceID,
		profile.UserID,
		profile.JobTitle,
		profile.WeeklyCapacityHours,
		profile.DailyCapacityHours,
		string(skillsJSON),
		profile.HourlyRate,
		profile.Currency,
		isActiveInt,
		profile.CreatedAt,
		profile.UpdatedAt,
	)
	if err != nil {
		return 0, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	profile.ID = id
	return id, nil
}

func (r *workloadRepositoryImpl) GetProfileByID(id int64) (*model.ResourceProfile, error) {
	query := `
	SELECT 
		rp.id, rp.workspace_id, rp.user_id, rp.job_title, rp.weekly_capacity_hours,
		rp.daily_capacity_hours, rp.skills, rp.hourly_rate, rp.currency, rp.is_active,
		rp.created_at, rp.updated_at,
		u.username, u.email
	FROM resource_profiles rp
	JOIN users u ON rp.user_id = u.id
	WHERE rp.id = ?`

	row := r.db.QueryRow(query, id)
	return r.scanProfile(row)
}

func (r *workloadRepositoryImpl) GetProfileByUserID(workspaceID int64, userID string) (*model.ResourceProfile, error) {
	query := `
	SELECT 
		rp.id, rp.workspace_id, rp.user_id, rp.job_title, rp.weekly_capacity_hours,
		rp.daily_capacity_hours, rp.skills, rp.hourly_rate, rp.currency, rp.is_active,
		rp.created_at, rp.updated_at,
		u.username, u.email
	FROM resource_profiles rp
	JOIN users u ON rp.user_id = u.id
	WHERE rp.workspace_id = ? AND rp.user_id = ?`

	row := r.db.QueryRow(query, workspaceID, userID)
	return r.scanProfile(row)
}

type rowScanner interface {
	Scan(dest ...interface{}) error
}

func (r *workloadRepositoryImpl) scanProfile(row rowScanner) (*model.ResourceProfile, error) {
	var (
		p           model.ResourceProfile
		skillsJSON  string
		isActiveInt int
	)
	err := row.Scan(
		&p.ID,
		&p.WorkspaceID,
		&p.UserID,
		&p.JobTitle,
		&p.WeeklyCapacityHours,
		&p.DailyCapacityHours,
		&skillsJSON,
		&p.HourlyRate,
		&p.Currency,
		&isActiveInt,
		&p.CreatedAt,
		&p.UpdatedAt,
		&p.Username,
		&p.Email,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrProfileNotFound
		}
		return nil, err
	}

	p.IsActive = isActiveInt == 1
	var skills []string
	if err := json.Unmarshal([]byte(skillsJSON), &skills); err == nil {
		p.Skills = skills
	}
	if p.Skills == nil {
		p.Skills = []string{}
	}

	return &p, nil
}

func (r *workloadRepositoryImpl) ListProfiles(workspaceID int64, isActive *bool) ([]*model.ResourceProfile, error) {
	baseQuery := `
	SELECT 
		rp.id, rp.workspace_id, rp.user_id, rp.job_title, rp.weekly_capacity_hours,
		rp.daily_capacity_hours, rp.skills, rp.hourly_rate, rp.currency, rp.is_active,
		rp.created_at, rp.updated_at,
		u.username, u.email
	FROM resource_profiles rp
	JOIN users u ON rp.user_id = u.id
	WHERE rp.workspace_id = ?`

	var args []interface{}
	args = append(args, workspaceID)

	if isActive != nil {
		baseQuery += ` AND rp.is_active = ?`
		activeInt := 0
		if *isActive {
			activeInt = 1
		}
		args = append(args, activeInt)
	}

	baseQuery += ` ORDER BY u.username ASC`

	rows, err := r.db.Query(baseQuery, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var profiles []*model.ResourceProfile
	for rows.Next() {
		p, err := r.scanProfile(rows)
		if err != nil {
			return nil, err
		}
		profiles = append(profiles, p)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	if profiles == nil {
		profiles = []*model.ResourceProfile{}
	}
	return profiles, nil
}

func (r *workloadRepositoryImpl) UpdateProfile(profile *model.ResourceProfile) error {
	query := `
	UPDATE resource_profiles SET
		job_title = ?,
		weekly_capacity_hours = ?,
		daily_capacity_hours = ?,
		skills = ?,
		hourly_rate = ?,
		currency = ?,
		is_active = ?,
		updated_at = ?
	WHERE id = ?`

	profile.UpdatedAt = time.Now()
	skillsJSON, err := json.Marshal(profile.Skills)
	if err != nil {
		skillsJSON = []byte("[]")
	}

	isActiveInt := 1
	if !profile.IsActive {
		isActiveInt = 0
	}

	res, err := r.db.Exec(query,
		profile.JobTitle,
		profile.WeeklyCapacityHours,
		profile.DailyCapacityHours,
		string(skillsJSON),
		profile.HourlyRate,
		profile.Currency,
		isActiveInt,
		profile.UpdatedAt,
		profile.ID,
	)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrProfileNotFound
	}
	return nil
}
func (r *workloadRepositoryImpl) DeleteProfile(id int64) error {
	query := `DELETE FROM resource_profiles WHERE id = ?`
	res, err := r.db.Exec(query, id)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrProfileNotFound
	}
	return nil
}

func (r *workloadRepositoryImpl) CreateTimeOff(timeOff *model.ResourceTimeOff) (int64, error) {
	query := `
	INSERT INTO resource_time_offs (
		workspace_id, user_id, start_date, end_date, hours_per_day,
		time_off_type, description, created_at, updated_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`

	now := time.Now()
	if timeOff.CreatedAt.IsZero() {
		timeOff.CreatedAt = now
	}
	timeOff.UpdatedAt = now
	if timeOff.HoursPerDay <= 0 {
		timeOff.HoursPerDay = model.DefaultDailyCapacityHours
	}
	if timeOff.TimeOffType == "" {
		timeOff.TimeOffType = model.TimeOffTypeVacation
	}

	res, err := r.db.Exec(query,
		timeOff.WorkspaceID,
		timeOff.UserID,
		timeOff.StartDate,
		timeOff.EndDate,
		timeOff.HoursPerDay,
		timeOff.TimeOffType,
		timeOff.Description,
		timeOff.CreatedAt,
		timeOff.UpdatedAt,
	)
	if err != nil {
		return 0, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	timeOff.ID = id
	return id, nil
}

func (r *workloadRepositoryImpl) GetTimeOffByID(id int64) (*model.ResourceTimeOff, error) {
	query := `
	SELECT 
		rto.id, rto.workspace_id, rto.user_id, rto.start_date, rto.end_date,
		rto.hours_per_day, rto.time_off_type, rto.description, rto.created_at, rto.updated_at,
		u.username
	FROM resource_time_offs rto
	JOIN users u ON rto.user_id = u.id
	WHERE rto.id = ?`

	row := r.db.QueryRow(query, id)
	return r.scanTimeOff(row)
}

func (r *workloadRepositoryImpl) ListTimeOffs(workspaceID int64, userID string, from, to *time.Time) ([]model.ResourceTimeOff, error) {
	baseQuery := `
	SELECT 
		rto.id, rto.workspace_id, rto.user_id, rto.start_date, rto.end_date,
		rto.hours_per_day, rto.time_off_type, rto.description, rto.created_at, rto.updated_at,
		u.username
	FROM resource_time_offs rto
	JOIN users u ON rto.user_id = u.id
	WHERE rto.workspace_id = ?`

	var args []interface{}
	args = append(args, workspaceID)

	if userID != "" {
		baseQuery += ` AND rto.user_id = ?`
		args = append(args, userID)
	}
	if from != nil {
		baseQuery += ` AND rto.end_date >= ?`
		args = append(args, *from)
	}
	if to != nil {
		baseQuery += ` AND rto.start_date <= ?`
		args = append(args, *to)
	}

	baseQuery += ` ORDER BY rto.start_date ASC`

	rows, err := r.db.Query(baseQuery, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []model.ResourceTimeOff
	for rows.Next() {
		to, err := r.scanTimeOff(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, *to)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	if list == nil {
		list = []model.ResourceTimeOff{}
	}
	return list, nil
}

func (r *workloadRepositoryImpl) scanTimeOff(row rowScanner) (*model.ResourceTimeOff, error) {
	var to model.ResourceTimeOff
	err := row.Scan(
		&to.ID,
		&to.WorkspaceID,
		&to.UserID,
		&to.StartDate,
		&to.EndDate,
		&to.HoursPerDay,
		&to.TimeOffType,
		&to.Description,
		&to.CreatedAt,
		&to.UpdatedAt,
		&to.Username,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrTimeOffNotFound
		}
		return nil, err
	}
	return &to, nil
}
func (r *workloadRepositoryImpl) UpdateTimeOff(timeOff *model.ResourceTimeOff) error {
	query := `
	UPDATE resource_time_offs SET
		start_date = ?,
		end_date = ?,
		hours_per_day = ?,
		time_off_type = ?,
		description = ?,
		updated_at = ?
	WHERE id = ?`

	timeOff.UpdatedAt = time.Now()
	res, err := r.db.Exec(query,
		timeOff.StartDate,
		timeOff.EndDate,
		timeOff.HoursPerDay,
		timeOff.TimeOffType,
		timeOff.Description,
		timeOff.UpdatedAt,
		timeOff.ID,
	)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrTimeOffNotFound
	}
	return nil
}

func (r *workloadRepositoryImpl) DeleteTimeOff(id int64) error {
	query := `DELETE FROM resource_time_offs WHERE id = ?`
	res, err := r.db.Exec(query, id)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrTimeOffNotFound
	}
	return nil
}

func (r *workloadRepositoryImpl) CreateAllocation(alloc *model.ProjectResourceAllocation) (int64, error) {
	query := `
	INSERT INTO project_resource_allocations (
		workspace_id, project_id, user_id, allocated_hours_per_week,
		start_date, end_date, status, notes, created_at, updated_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	now := time.Now()
	if alloc.CreatedAt.IsZero() {
		alloc.CreatedAt = now
	}
	alloc.UpdatedAt = now
	if alloc.Status == "" {
		alloc.Status = model.AllocationStatusPlanned
	}

	res, err := r.db.Exec(query,
		alloc.WorkspaceID,
		alloc.ProjectID,
		alloc.UserID,
		alloc.AllocatedHoursPerWeek,
		alloc.StartDate,
		alloc.EndDate,
		alloc.Status,
		alloc.Notes,
		alloc.CreatedAt,
		alloc.UpdatedAt,
	)
	if err != nil {
		return 0, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	alloc.ID = id
	return id, nil
}

func (r *workloadRepositoryImpl) GetAllocationByID(id int64) (*model.ProjectResourceAllocation, error) {
	query := `
	SELECT 
		pra.id, pra.workspace_id, pra.project_id, pra.user_id, pra.allocated_hours_per_week,
		pra.start_date, pra.end_date, pra.status, pra.notes, pra.created_at, pra.updated_at,
		pr.name, pr.color, u.username
	FROM project_resource_allocations pra
	JOIN projects pr ON pra.project_id = pr.id
	JOIN users u ON pra.user_id = u.id
	WHERE pra.id = ?`

	row := r.db.QueryRow(query, id)
	return r.scanAllocation(row)
}

func (r *workloadRepositoryImpl) ListAllocations(workspaceID int64, projectID int, userID string, status string) ([]model.ProjectResourceAllocation, error) {
	baseQuery := `
	SELECT 
		pra.id, pra.workspace_id, pra.project_id, pra.user_id, pra.allocated_hours_per_week,
		pra.start_date, pra.end_date, pra.status, pra.notes, pra.created_at, pra.updated_at,
		pr.name, pr.color, u.username
	FROM project_resource_allocations pra
	JOIN projects pr ON pra.project_id = pr.id
	JOIN users u ON pra.user_id = u.id
	WHERE pra.workspace_id = ?`

	var args []interface{}
	args = append(args, workspaceID)

	if projectID > 0 {
		baseQuery += ` AND pra.project_id = ?`
		args = append(args, projectID)
	}
	if userID != "" {
		baseQuery += ` AND pra.user_id = ?`
		args = append(args, userID)
	}
	if status != "" {
		baseQuery += ` AND pra.status = ?`
		args = append(args, status)
	}

	baseQuery += ` ORDER BY pra.start_date ASC`

	rows, err := r.db.Query(baseQuery, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []model.ProjectResourceAllocation
	for rows.Next() {
		alloc, err := r.scanAllocation(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, *alloc)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	if list == nil {
		list = []model.ProjectResourceAllocation{}
	}
	return list, nil
}

func (r *workloadRepositoryImpl) UpdateAllocation(alloc *model.ProjectResourceAllocation) error {
	query := `
	UPDATE project_resource_allocations SET
		allocated_hours_per_week = ?,
		start_date = ?,
		end_date = ?,
		status = ?,
		notes = ?,
		updated_at = ?
	WHERE id = ?`

	alloc.UpdatedAt = time.Now()
	res, err := r.db.Exec(query,
		alloc.AllocatedHoursPerWeek,
		alloc.StartDate,
		alloc.EndDate,
		alloc.Status,
		alloc.Notes,
		alloc.UpdatedAt,
		alloc.ID,
	)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrAllocationNotFound
	}
	return nil
}

func (r *workloadRepositoryImpl) DeleteAllocation(id int64) error {
	query := `DELETE FROM project_resource_allocations WHERE id = ?`
	res, err := r.db.Exec(query, id)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrAllocationNotFound
	}
	return nil
}

func (r *workloadRepositoryImpl) scanAllocation(row rowScanner) (*model.ProjectResourceAllocation, error) {
	var a model.ProjectResourceAllocation
	err := row.Scan(
		&a.ID,
		&a.WorkspaceID,
		&a.ProjectID,
		&a.UserID,
		&a.AllocatedHoursPerWeek,
		&a.StartDate,
		&a.EndDate,
		&a.Status,
		&a.Notes,
		&a.CreatedAt,
		&a.UpdatedAt,
		&a.ProjectName,
		&a.ProjectColor,
		&a.Username,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrAllocationNotFound
		}
		return nil, err
	}
	return &a, nil
}

func (r *workloadRepositoryImpl) GetUserTasksForWorkload(workspaceID int64, userID string, from, to time.Time) ([]model.Task, error) {
	query := `
	SELECT id, user_id, project_id, title, description, created_at, updated_at, due_date, estimate_minutes, completed, status, priority
	FROM tasks
	WHERE user_id = ? AND due_date >= ? AND due_date <= ?
	ORDER BY due_date ASC`

	rows, err := r.db.Query(query, userID, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []model.Task
	for rows.Next() {
		var task model.Task
		err := rows.Scan(
			&task.ID, &task.UserID, &task.ProjectID, &task.Title, &task.Description,
			&task.CreatedAt, &task.UpdatedAt, &task.DueDate,
			&task.EstimateMinutes, &task.Completed, &task.Status, &task.Priority,
		)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	if tasks == nil {
		tasks = []model.Task{}
	}
	return tasks, nil
}
func (r *workloadRepositoryImpl) GetWorkspaceWorkloadMetrics(workspaceID int64) (*model.WorkloadMetrics, error) {
	metrics := &model.WorkloadMetrics{
		WorkspaceID: workspaceID,
	}

	// 1. Profile capacity totals
	profQuery := `
	SELECT 
		COUNT(id),
		COALESCE(SUM(CASE WHEN is_active = 1 THEN 1 ELSE 0 END), 0),
		COALESCE(SUM(CASE WHEN is_active = 1 THEN weekly_capacity_hours ELSE 0 END), 0.0)
	FROM resource_profiles
	WHERE workspace_id = ?`

	err := r.db.QueryRow(profQuery, workspaceID).Scan(
		&metrics.TotalProfiles,
		&metrics.ActiveProfiles,
		&metrics.TotalWeeklyCapacityHours,
	)
	if err != nil {
		return nil, err
	}

	// 2. Active project allocation hours
	allocQuery := `
	SELECT COALESCE(SUM(allocated_hours_per_week), 0.0)
	FROM project_resource_allocations
	WHERE workspace_id = ? AND status IN ('planned', 'active')`

	err = r.db.QueryRow(allocQuery, workspaceID).Scan(&metrics.TotalAllocatedProjectHours)
	if err != nil {
		return nil, err
	}

	// 3. Open task estimated hours assigned to active workspace resources
	taskHoursQuery := `
	SELECT COALESCE(SUM(t.estimate_minutes) / 60.0, 0.0)
	FROM tasks t
	JOIN resource_profiles rp ON t.user_id = rp.user_id
	WHERE rp.workspace_id = ? AND rp.is_active = 1 AND (t.status != 'completed' AND t.completed = 0)`

	err = r.db.QueryRow(taskHoursQuery, workspaceID).Scan(&metrics.TotalAssignedTaskHours)
	if err != nil {
		return nil, err
	}

	// 4. Scheduled future/active time-off entries
	timeOffQuery := `
	SELECT COUNT(id)
	FROM resource_time_offs
	WHERE workspace_id = ? AND end_date >= datetime('now')`

	err = r.db.QueryRow(timeOffQuery, workspaceID).Scan(&metrics.TotalScheduledTimeOffDays)
	if err != nil {
		return nil, err
	}

	// 5. Utilization rate calculation
	if metrics.TotalWeeklyCapacityHours > 0 {
		totalDemand := metrics.TotalAssignedTaskHours + metrics.TotalAllocatedProjectHours
		rate := (totalDemand / metrics.TotalWeeklyCapacityHours) * 100.0
		metrics.AverageUtilizationRate = float64(int(rate*10)) / 10.0
	}

	return metrics, nil
}
