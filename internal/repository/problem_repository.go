package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/JohnMalugu/tsk-mgr-api/internal/model"
)

var (
	ErrProblemNotFound     = errors.New("problem not found")
	ErrProblemExists       = errors.New("problem already exists")
	ErrProblemLinkExists   = errors.New("problem link already exists")
	ErrProblemLinkNotFound = errors.New("problem link not found")
	ErrRCAStepNotFound     = errors.New("root cause step not found")
)

// ProblemRepository defines persistent storage operations for ITIL Problem Management & KEDB
type ProblemRepository interface {
	// Sequential code generation
	GetNextProblemCode(workspaceID int64) (string, error)

	// Problem CRUD
	CreateProblem(p *model.Problem) (int64, error)
	GetProblemByID(id int64) (*model.Problem, error)
	GetProblemByCode(workspaceID int64, code string) (*model.Problem, error)
	ListProblems(filter model.ProblemFilter) ([]model.Problem, int, error)
	UpdateProblem(p *model.Problem) error
	DeleteProblem(id int64) error

	// Incident Linkage
	LinkIncident(link *model.ProblemIncidentLink) (int64, error)
	UnlinkIncident(problemID, incidentID int64) error
	ListProblemIncidents(problemID int64) ([]model.ProblemIncidentLink, error)
	GetProblemsForIncident(incidentID int64) ([]model.Problem, error)

	// CI Linkage
	LinkCI(link *model.ProblemCILink) (int64, error)
	UnlinkCI(id int64) error
	ListProblemCIs(problemID int64) ([]model.ProblemCILink, error)
	GetProblemsForCI(ciID int64) ([]model.Problem, error)

	// Change Linkage
	LinkChange(link *model.ProblemChangeLink) (int64, error)
	UnlinkChange(id int64) error
	ListProblemChanges(problemID int64) ([]model.ProblemChangeLink, error)
	GetProblemsForChange(changeID int64) ([]model.Problem, error)

	// Task Linkage
	LinkTask(link *model.ProblemTaskLink) (int64, error)
	UnlinkTask(id int64) error
	ListProblemTasks(problemID int64) ([]model.ProblemTaskLink, error)
	GetProblemsForTask(taskID int) ([]model.Problem, error)

	// 5-Whys RCA Steps
	AddRCAStep(step *model.ProblemRCAStep) (int64, error)
	GetRCAStep(id int64) (*model.ProblemRCAStep, error)
	UpdateRCAStep(step *model.ProblemRCAStep) error
	DeleteRCAStep(id int64) error
	ListRCASteps(problemID int64) ([]model.ProblemRCAStep, error)
	GetNextRCAStepNumber(problemID int64) (int, error)

	// Known Error Database (KEDB)
	ListKnownErrors(filter model.KnownErrorFilter) ([]model.Problem, int, error)

	// Metrics
	GetProblemMetrics(workspaceID int64) (*model.ProblemMetrics, error)
}

type problemRepositoryImpl struct {
	db *sql.DB
}

// NewProblemRepository creates a new SQLite-backed ProblemRepository
func NewProblemRepository(db *sql.DB) ProblemRepository {
	return &problemRepositoryImpl{db: db}
}

func (r *problemRepositoryImpl) GetNextProblemCode(workspaceID int64) (string, error) {
	query := `SELECT problem_code FROM problems WHERE workspace_id = ? ORDER BY id DESC LIMIT 1`
	var lastCode string
	err := r.db.QueryRow(query, workspaceID).Scan(&lastCode)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "PRB-1", nil
		}
		return "PRB-1", err
	}

	parts := strings.Split(lastCode, "-")
	if len(parts) != 2 {
		return "PRB-1", nil
	}
	nextNum, err := strconv.Atoi(parts[1])
	if err != nil {
		return "PRB-1", nil
	}
	return fmt.Sprintf("PRB-%d", nextNum+1), nil
}

func (r *problemRepositoryImpl) CreateProblem(p *model.Problem) (int64, error) {
	return 0, nil
}

func (r *problemRepositoryImpl) GetProblemByID(id int64) (*model.Problem, error) {
	return nil, nil
}

func (r *problemRepositoryImpl) GetProblemByCode(workspaceID int64, code string) (*model.Problem, error) {
	return nil, nil
}

func (r *problemRepositoryImpl) ListProblems(filter model.ProblemFilter) ([]model.Problem, int, error) {
	return nil, 0, nil
}

func (r *problemRepositoryImpl) UpdateProblem(p *model.Problem) error {
	return nil
}

func (r *problemRepositoryImpl) DeleteProblem(id int64) error {
	return nil
}

func (r *problemRepositoryImpl) LinkIncident(link *model.ProblemIncidentLink) (int64, error) {
	return 0, nil
}

func (r *problemRepositoryImpl) UnlinkIncident(problemID, incidentID int64) error {
	return nil
}

func (r *problemRepositoryImpl) ListProblemIncidents(problemID int64) ([]model.ProblemIncidentLink, error) {
	return nil, nil
}

func (r *problemRepositoryImpl) GetProblemsForIncident(incidentID int64) ([]model.Problem, error) {
	return nil, nil
}

func (r *problemRepositoryImpl) LinkCI(link *model.ProblemCILink) (int64, error) {
	return 0, nil
}

func (r *problemRepositoryImpl) UnlinkCI(id int64) error {
	return nil
}

func (r *problemRepositoryImpl) ListProblemCIs(problemID int64) ([]model.ProblemCILink, error) {
	return nil, nil
}

func (r *problemRepositoryImpl) GetProblemsForCI(ciID int64) ([]model.Problem, error) {
	return nil, nil
}

func (r *problemRepositoryImpl) LinkChange(link *model.ProblemChangeLink) (int64, error) {
	return 0, nil
}

func (r *problemRepositoryImpl) UnlinkChange(id int64) error {
	return nil
}

func (r *problemRepositoryImpl) ListProblemChanges(problemID int64) ([]model.ProblemChangeLink, error) {
	return nil, nil
}

func (r *problemRepositoryImpl) GetProblemsForChange(changeID int64) ([]model.Problem, error) {
	return nil, nil
}

func (r *problemRepositoryImpl) LinkTask(link *model.ProblemTaskLink) (int64, error) {
	return 0, nil
}

func (r *problemRepositoryImpl) UnlinkTask(id int64) error {
	return nil
}

func (r *problemRepositoryImpl) ListProblemTasks(problemID int64) ([]model.ProblemTaskLink, error) {
	return nil, nil
}

func (r *problemRepositoryImpl) GetProblemsForTask(taskID int) ([]model.Problem, error) {
	return nil, nil
}

func (r *problemRepositoryImpl) AddRCAStep(step *model.ProblemRCAStep) (int64, error) {
	return 0, nil
}

func (r *problemRepositoryImpl) GetRCAStep(id int64) (*model.ProblemRCAStep, error) {
	return nil, nil
}

func (r *problemRepositoryImpl) UpdateRCAStep(step *model.ProblemRCAStep) error {
	return nil
}

func (r *problemRepositoryImpl) DeleteRCAStep(id int64) error {
	return nil
}

func (r *problemRepositoryImpl) ListRCASteps(problemID int64) ([]model.ProblemRCAStep, error) {
	return nil, nil
}

func (r *problemRepositoryImpl) GetNextRCAStepNumber(problemID int64) (int, error) {
	return 0, nil
}

func (r *problemRepositoryImpl) ListKnownErrors(filter model.KnownErrorFilter) ([]model.Problem, int, error) {
	return nil, 0, nil
}

func (r *problemRepositoryImpl) GetProblemMetrics(workspaceID int64) (*model.ProblemMetrics, error) {
	return nil, nil
}
