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
	return nil, nil
}
func (r *workloadRepositoryImpl) UpdateProfile(profile *model.ResourceProfile) error {
	return nil
}
func (r *workloadRepositoryImpl) DeleteProfile(id int64) error {
	return nil
}

func (r *workloadRepositoryImpl) CreateTimeOff(timeOff *model.ResourceTimeOff) (int64, error) {
	return 0, nil
}
func (r *workloadRepositoryImpl) GetTimeOffByID(id int64) (*model.ResourceTimeOff, error) {
	return nil, nil
}
func (r *workloadRepositoryImpl) ListTimeOffs(workspaceID int64, userID string, from, to *time.Time) ([]model.ResourceTimeOff, error) {
	return nil, nil
}
func (r *workloadRepositoryImpl) UpdateTimeOff(timeOff *model.ResourceTimeOff) error {
	return nil
}
func (r *workloadRepositoryImpl) DeleteTimeOff(id int64) error {
	return nil
}

func (r *workloadRepositoryImpl) CreateAllocation(alloc *model.ProjectResourceAllocation) (int64, error) {
	return 0, nil
}
func (r *workloadRepositoryImpl) GetAllocationByID(id int64) (*model.ProjectResourceAllocation, error) {
	return nil, nil
}
func (r *workloadRepositoryImpl) ListAllocations(workspaceID int64, projectID int, userID string, status string) ([]model.ProjectResourceAllocation, error) {
	return nil, nil
}
func (r *workloadRepositoryImpl) UpdateAllocation(alloc *model.ProjectResourceAllocation) error {
	return nil
}
func (r *workloadRepositoryImpl) DeleteAllocation(id int64) error {
	return nil
}

func (r *workloadRepositoryImpl) GetUserTasksForWorkload(workspaceID int64, userID string, from, to time.Time) ([]model.Task, error) {
	return nil, nil
}
func (r *workloadRepositoryImpl) GetWorkspaceWorkloadMetrics(workspaceID int64) (*model.WorkloadMetrics, error) {
	return nil, nil
}
