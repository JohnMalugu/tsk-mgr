package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/JohnMalugu/tsk-mgr-api/internal/model"
)

// SprintRepository handles database persistence for sprints and sprint task associations
type SprintRepository interface {
	CreateSprint(sprint *model.SprintDB) (int64, error)
	GetSprintByID(id int64) (*model.SprintDB, error)
	ListSprints(filter model.SprintFilter) ([]model.SprintDB, int, error)
	UpdateSprint(sprint *model.SprintDB) error
	DeleteSprint(id int64) error
	GetActiveSprintByProject(projectID int64) (*model.SprintDB, error)
	AddTaskToSprint(sprintID, taskID int64, storyPoints int) error
	RemoveTaskFromSprint(sprintID, taskID int64) error
	GetSprintTasks(sprintID int64) ([]model.SprintTaskDB, error)
	UpdateTaskStoryPoints(sprintID, taskID int64, storyPoints int) error
	GetSprintMetrics(sprintID int64) (*model.SprintMetrics, error)
}

type sqlSprintRepository struct {
	db *sql.DB
}

// NewSprintRepository returns a new implementation of SprintRepository
func NewSprintRepository(db *sql.DB) SprintRepository {
	return &sqlSprintRepository{db: db}
}

func (r *sqlSprintRepository) CreateSprint(sprint *model.SprintDB) (int64, error) {
	now := time.Now()
	if sprint.CreatedAt.IsZero() {
		sprint.CreatedAt = now
	}
	if sprint.UpdatedAt.IsZero() {
		sprint.UpdatedAt = now
	}
	if sprint.Status == "" {
		sprint.Status = model.SprintStatusPlanned
	}

	query := `
	INSERT INTO sprints (workspace_id, project_id, name, goal, status, start_date, end_date, total_points, completed_points, created_at, updated_at)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	res, err := r.db.Exec(query,
		sprint.WorkspaceID,
		sprint.ProjectID,
		sprint.Name,
		sprint.Goal,
		string(sprint.Status),
		sprint.StartDate,
		sprint.EndDate,
		sprint.TotalPoints,
		sprint.CompletedPoints,
		sprint.CreatedAt,
		sprint.UpdatedAt,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to insert sprint: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	sprint.ID = id
	return id, nil
}

func (r *sqlSprintRepository) GetSprintByID(id int64) (*model.SprintDB, error) {
	query := `
	SELECT id, workspace_id, project_id, name, goal, status, start_date, end_date, total_points, completed_points, created_at, updated_at
	FROM sprints
	WHERE id = ?`

	row := r.db.QueryRow(query, id)
	var s model.SprintDB
	var statusStr string
	err := row.Scan(
		&s.ID,
		&s.WorkspaceID,
		&s.ProjectID,
		&s.Name,
		&s.Goal,
		&statusStr,
		&s.StartDate,
		&s.EndDate,
		&s.TotalPoints,
		&s.CompletedPoints,
		&s.CreatedAt,
		&s.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to query sprint: %w", err)
	}
	s.Status = model.SprintStatus(statusStr)
	return &s, nil
}

func (r *sqlSprintRepository) ListSprints(filter model.SprintFilter) ([]model.SprintDB, int, error) {
	return nil, 0, nil
}

func (r *sqlSprintRepository) UpdateSprint(sprint *model.SprintDB) error {
	return nil
}

func (r *sqlSprintRepository) DeleteSprint(id int64) error {
	return nil
}

func (r *sqlSprintRepository) GetActiveSprintByProject(projectID int64) (*model.SprintDB, error) {
	return nil, nil
}

func (r *sqlSprintRepository) AddTaskToSprint(sprintID, taskID int64, storyPoints int) error {
	return nil
}

func (r *sqlSprintRepository) RemoveTaskFromSprint(sprintID, taskID int64) error {
	return nil
}

func (r *sqlSprintRepository) GetSprintTasks(sprintID int64) ([]model.SprintTaskDB, error) {
	return nil, nil
}

func (r *sqlSprintRepository) UpdateTaskStoryPoints(sprintID, taskID int64, storyPoints int) error {
	return nil
}

func (r *sqlSprintRepository) GetSprintMetrics(sprintID int64) (*model.SprintMetrics, error) {
	return nil, nil
}
