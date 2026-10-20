package repository

import (
	"database/sql"

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
	return 0, nil
}

func (r *sqlSprintRepository) GetSprintByID(id int64) (*model.SprintDB, error) {
	return nil, nil
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
