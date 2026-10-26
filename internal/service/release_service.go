package service

import (
	"errors"
	"strings"

	"github.com/JohnMalugu/tsk-mgr-api/internal/model"
	"github.com/JohnMalugu/tsk-mgr-api/internal/repository"
)

var (
	ErrReleaseNotFound          = errors.New("release not found")
	ErrReleaseNameRequired      = errors.New("release name cannot be empty")
	ErrInvalidReleaseVersion    = errors.New("invalid release version format")
	ErrInvalidReleaseStatus     = errors.New("invalid release status")
	ErrInvalidStatusTransition  = errors.New("invalid release status transition")
	ErrReleaseCodeFrozen        = errors.New("cannot modify tasks in a code-frozen release")
	ErrReleaseAlreadyReleased   = errors.New("cannot modify a released or archived release")
	ErrInvalidReleaseCategory   = errors.New("invalid release item category")
	ErrNoTasksSpecified         = errors.New("no tasks specified")
	ErrReleaseItemNotFound      = errors.New("release item not found")
)

// ReleaseService coordinates release milestones, task scoping, changelog generation, and metrics
type ReleaseService interface {
	CreateRelease(userID string, req model.CreateReleaseRequest) (*model.Release, error)
	GetRelease(id int64) (*model.Release, error)
	ListReleases(filter model.ReleaseFilter) ([]*model.ReleaseSummary, error)
	UpdateRelease(id int64, req model.UpdateReleaseRequest) (*model.Release, error)
	UpdateReleaseStatus(id int64, req model.UpdateReleaseStatusRequest) (*model.Release, error)
	DeleteRelease(id int64) error

	AddTasksToRelease(userID string, releaseID int64, req model.AddReleaseTasksRequest) ([]*model.ReleaseItem, error)
	RemoveTaskFromRelease(releaseID int64, taskID int) error
	UpdateReleaseItem(releaseID int64, taskID int, req model.UpdateReleaseItemRequest) (*model.ReleaseItem, error)
	ListReleaseItems(releaseID int64, category string) ([]*model.ReleaseItem, error)

	GenerateReleaseNotes(releaseID int64, opts model.ReleaseNotesOptions) (*model.GeneratedReleaseNotes, error)
	GetReleaseMetrics(releaseID int64) (*model.ReleaseMetrics, error)
	CloneRelease(userID string, sourceReleaseID int64, req model.CloneReleaseRequest) (*model.Release, error)
}

type releaseServiceImpl struct {
	repo     repository.ReleaseRepository
	taskRepo repository.TaskRepository
}

// NewReleaseService constructs a ReleaseService instance
func NewReleaseService(repo repository.ReleaseRepository, taskRepo repository.TaskRepository) ReleaseService {
	return &releaseServiceImpl{
		repo:     repo,
		taskRepo: taskRepo,
	}
}

func validateReleaseName(name string) (string, error) {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return "", ErrReleaseNameRequired
	}
	if len([]rune(trimmed)) > 100 {
		return "", errors.New("release name cannot exceed 100 characters")
	}
	return trimmed, nil
}

func validateReleaseVersion(version string) (string, error) {
	trimmed := strings.TrimSpace(version)
	if trimmed == "" {
		return "", ErrInvalidReleaseVersion
	}
	if len([]rune(trimmed)) > 50 {
		return "", errors.New("release version cannot exceed 50 characters")
	}
	// Verify version contains reasonable characters: alphanumeric, dots, dashes, plus, 'v' prefix
	for _, r := range trimmed {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '.' || r == '-' || r == '+' || r == '_') {
			return "", ErrInvalidReleaseVersion
		}
	}
	return trimmed, nil
}

func validateReleaseCategory(cat string) (string, error) {
	lower := strings.ToLower(strings.TrimSpace(cat))
	if lower == "" {
		return model.ReleaseCategoryFeature, nil
	}
	if !model.IsValidReleaseCategory(lower) {
		return "", ErrInvalidReleaseCategory
	}
	return lower, nil
}

func validateReleaseStatusTransition(currentStatus, targetStatus string) error {
	if currentStatus == targetStatus {
		return nil
	}
	switch targetStatus {
	case model.ReleaseStatusPlanning, model.ReleaseStatusActive, model.ReleaseStatusFreeze, model.ReleaseStatusReleased, model.ReleaseStatusArchived:
		// Valid target status
	default:
		return ErrInvalidReleaseStatus
	}

	switch currentStatus {
	case model.ReleaseStatusPlanning:
		if targetStatus == model.ReleaseStatusActive || targetStatus == model.ReleaseStatusArchived {
			return nil
		}
	case model.ReleaseStatusActive:
		if targetStatus == model.ReleaseStatusFreeze || targetStatus == model.ReleaseStatusReleased || targetStatus == model.ReleaseStatusPlanning || targetStatus == model.ReleaseStatusArchived {
			return nil
		}
	case model.ReleaseStatusFreeze:
		if targetStatus == model.ReleaseStatusReleased || targetStatus == model.ReleaseStatusActive || targetStatus == model.ReleaseStatusArchived {
			return nil
		}
	case model.ReleaseStatusReleased:
		if targetStatus == model.ReleaseStatusArchived {
			return nil
		}
	case model.ReleaseStatusArchived:
		if targetStatus == model.ReleaseStatusPlanning || targetStatus == model.ReleaseStatusActive {
			return nil
		}
	}

	return ErrInvalidStatusTransition
}


func (s *releaseServiceImpl) CreateRelease(userID string, req model.CreateReleaseRequest) (*model.Release, error) {
	return nil, nil
}

func (s *releaseServiceImpl) GetRelease(id int64) (*model.Release, error) {
	return nil, nil
}

func (s *releaseServiceImpl) ListReleases(filter model.ReleaseFilter) ([]*model.ReleaseSummary, error) {
	return nil, nil
}

func (s *releaseServiceImpl) UpdateRelease(id int64, req model.UpdateReleaseRequest) (*model.Release, error) {
	return nil, nil
}

func (s *releaseServiceImpl) UpdateReleaseStatus(id int64, req model.UpdateReleaseStatusRequest) (*model.Release, error) {
	return nil, nil
}

func (s *releaseServiceImpl) DeleteRelease(id int64) error {
	return nil
}

func (s *releaseServiceImpl) AddTasksToRelease(userID string, releaseID int64, req model.AddReleaseTasksRequest) ([]*model.ReleaseItem, error) {
	return nil, nil
}

func (s *releaseServiceImpl) RemoveTaskFromRelease(releaseID int64, taskID int) error {
	return nil
}

func (s *releaseServiceImpl) UpdateReleaseItem(releaseID int64, taskID int, req model.UpdateReleaseItemRequest) (*model.ReleaseItem, error) {
	return nil, nil
}

func (s *releaseServiceImpl) ListReleaseItems(releaseID int64, category string) ([]*model.ReleaseItem, error) {
	return nil, nil
}

func (s *releaseServiceImpl) GenerateReleaseNotes(releaseID int64, opts model.ReleaseNotesOptions) (*model.GeneratedReleaseNotes, error) {
	return nil, nil
}

func (s *releaseServiceImpl) GetReleaseMetrics(releaseID int64) (*model.ReleaseMetrics, error) {
	return nil, nil
}

func (s *releaseServiceImpl) CloneRelease(userID string, sourceReleaseID int64, req model.CloneReleaseRequest) (*model.Release, error) {
	return nil, nil
}
