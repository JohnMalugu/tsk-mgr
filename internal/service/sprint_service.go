package service

import (
	"errors"
	"strings"
	"time"

	"github.com/JohnMalugu/tsk-mgr-api/internal/model"
	"github.com/JohnMalugu/tsk-mgr-api/internal/repository"
)

var (
	ErrSprintNotFound       = errors.New("sprint not found")
	ErrSprintNameRequired   = errors.New("sprint name is required")
	ErrInvalidSprintStatus  = errors.New("invalid sprint status")
	ErrActiveSprintExists   = errors.New("another sprint is already active in this project")
	ErrSprintCannotStart    = errors.New("only planned sprints can be started")
	ErrSprintCannotComplete      = errors.New("only active sprints can be completed")
	ErrSprintCannotClose         = errors.New("only completed sprints can be closed")
	ErrSprintClosed              = errors.New("cannot modify a closed sprint")
	ErrCannotDeleteActiveSprint  = errors.New("cannot delete an active sprint")
	ErrTaskAlreadyInSprint       = errors.New("task is already assigned to this sprint")
	ErrInvalidStoryPoints        = errors.New("story points must be non-negative")
)

// SprintService handles agile sprint lifecycle, task scoping, and velocity metrics
type SprintService interface {
	CreateSprint(req *model.CreateSprintRequest) (*model.SprintDB, error)
	GetSprint(id int64) (*model.SprintDB, error)
	ListSprints(filter model.SprintFilter) (*model.SprintSummary, error)
	UpdateSprint(id int64, req *model.UpdateSprintRequest) (*model.SprintDB, error)
	DeleteSprint(id int64) error
	StartSprint(id int64) (*model.SprintDB, error)
	CompleteSprint(id int64) (*model.SprintDB, error)
	CloseSprint(id int64) (*model.SprintDB, error)
	AddTaskToSprint(sprintID int64, req *model.AddTaskToSprintRequest) error
	RemoveTaskFromSprint(sprintID, taskID int64) error
	GetSprintTasks(sprintID int64) ([]model.SprintTaskDB, error)
	UpdateTaskStoryPoints(sprintID, taskID int64, storyPoints int) error
	GetSprintMetrics(sprintID int64) (*model.SprintMetrics, error)
}

type sprintService struct {
	repo repository.SprintRepository
}

// NewSprintService creates a new SprintService instance
func NewSprintService(repo repository.SprintRepository) SprintService {
	return &sprintService{repo: repo}
}

func (s *sprintService) CreateSprint(req *model.CreateSprintRequest) (*model.SprintDB, error) {
	if req == nil {
		return nil, errors.New("request cannot be nil")
	}

	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, ErrSprintNameRequired
	}

	sprint := &model.SprintDB{
		WorkspaceID: req.WorkspaceID,
		ProjectID:   req.ProjectID,
		Name:        name,
		Goal:        strings.TrimSpace(req.Goal),
		Status:      model.SprintStatusPlanned,
		StartDate:   req.StartDate,
		EndDate:     req.EndDate,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	id, err := s.repo.CreateSprint(sprint)
	if err != nil {
		return nil, err
	}
	sprint.ID = id
	return sprint, nil
}

func (s *sprintService) GetSprint(id int64) (*model.SprintDB, error) {
	sprint, err := s.repo.GetSprintByID(id)
	if err != nil {
		return nil, err
	}
	if sprint == nil {
		return nil, ErrSprintNotFound
	}
	return sprint, nil
}

func (s *sprintService) ListSprints(filter model.SprintFilter) (*model.SprintSummary, error) {
	if filter.Limit <= 0 {
		filter.Limit = 20
	}
	if filter.Limit > 100 {
		filter.Limit = 100
	}
	if filter.Offset < 0 {
		filter.Offset = 0
	}

	sprints, total, err := s.repo.ListSprints(filter)
	if err != nil {
		return nil, err
	}

	return &model.SprintSummary{
		Sprints: sprints,
		Total:   total,
		Offset:  filter.Offset,
		Limit:   filter.Limit,
	}, nil
}

func (s *sprintService) UpdateSprint(id int64, req *model.UpdateSprintRequest) (*model.SprintDB, error) {
	if req == nil {
		return nil, errors.New("request cannot be nil")
	}

	sprint, err := s.repo.GetSprintByID(id)
	if err != nil {
		return nil, err
	}
	if sprint == nil {
		return nil, ErrSprintNotFound
	}

	if sprint.Status == model.SprintStatusArchived {
		return nil, ErrSprintClosed
	}

	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)
		if name == "" {
			return nil, ErrSprintNameRequired
		}
		sprint.Name = name
	}
	if req.Goal != nil {
		sprint.Goal = strings.TrimSpace(*req.Goal)
	}
	if req.StartDate != nil {
		sprint.StartDate = req.StartDate
	}
	if req.EndDate != nil {
		sprint.EndDate = req.EndDate
	}
	if req.Status != nil {
		if !model.IsValidSprintStatus(string(*req.Status)) {
			return nil, ErrInvalidSprintStatus
		}
		sprint.Status = *req.Status
	}
	sprint.UpdatedAt = time.Now()

	if err := s.repo.UpdateSprint(sprint); err != nil {
		return nil, err
	}

	return sprint, nil
}

func (s *sprintService) DeleteSprint(id int64) error {
	sprint, err := s.repo.GetSprintByID(id)
	if err != nil {
		return err
	}
	if sprint == nil {
		return ErrSprintNotFound
	}
	if sprint.Status == model.SprintStatusActive {
		return ErrCannotDeleteActiveSprint
	}

	return s.repo.DeleteSprint(id)
}

func (s *sprintService) StartSprint(id int64) (*model.SprintDB, error) {
	sprint, err := s.repo.GetSprintByID(id)
	if err != nil {
		return nil, err
	}
	if sprint == nil {
		return nil, ErrSprintNotFound
	}

	if sprint.Status != model.SprintStatusPlanned {
		return nil, ErrSprintCannotStart
	}

	active, err := s.repo.GetActiveSprintByProject(sprint.ProjectID)
	if err != nil {
		return nil, err
	}
	if active != nil && active.ID != sprint.ID {
		return nil, ErrActiveSprintExists
	}

	now := time.Now()
	sprint.Status = model.SprintStatusActive
	if sprint.StartDate == nil {
		sprint.StartDate = &now
	}
	sprint.UpdatedAt = now

	if err := s.repo.UpdateSprint(sprint); err != nil {
		return nil, err
	}

	return sprint, nil
}

func (s *sprintService) CompleteSprint(id int64) (*model.SprintDB, error) {
	sprint, err := s.repo.GetSprintByID(id)
	if err != nil {
		return nil, err
	}
	if sprint == nil {
		return nil, ErrSprintNotFound
	}

	if sprint.Status != model.SprintStatusActive {
		return nil, ErrSprintCannotComplete
	}

	now := time.Now()
	sprint.Status = model.SprintStatusCompleted
	if sprint.EndDate == nil {
		sprint.EndDate = &now
	}
	sprint.UpdatedAt = now

	if err := s.repo.UpdateSprint(sprint); err != nil {
		return nil, err
	}

	return sprint, nil
}

func (s *sprintService) CloseSprint(id int64) (*model.SprintDB, error) {
	sprint, err := s.repo.GetSprintByID(id)
	if err != nil {
		return nil, err
	}
	if sprint == nil {
		return nil, ErrSprintNotFound
	}

	if sprint.Status != model.SprintStatusCompleted {
		return nil, ErrSprintCannotClose
	}

	sprint.Status = model.SprintStatusArchived
	sprint.UpdatedAt = time.Now()

	if err := s.repo.UpdateSprint(sprint); err != nil {
		return nil, err
	}

	return sprint, nil
}

func (s *sprintService) AddTaskToSprint(sprintID int64, req *model.AddTaskToSprintRequest) error {
	return nil
}

func (s *sprintService) RemoveTaskFromSprint(sprintID, taskID int64) error {
	return nil
}

func (s *sprintService) GetSprintTasks(sprintID int64) ([]model.SprintTaskDB, error) {
	return nil, nil
}

func (s *sprintService) UpdateTaskStoryPoints(sprintID, taskID int64, storyPoints int) error {
	return nil
}

func (s *sprintService) GetSprintMetrics(sprintID int64) (*model.SprintMetrics, error) {
	return nil, nil
}
