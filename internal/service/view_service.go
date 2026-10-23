package service

import (
	"errors"

	"github.com/JohnMalugu/tsk-mgr-api/internal/model"
	"github.com/JohnMalugu/tsk-mgr-api/internal/repository"
)

// Domain errors for saved views
var (
	ErrViewNotFound           = errors.New("saved view not found")
	ErrInvalidViewName        = errors.New("view name is required and must not exceed 100 characters")
	ErrInvalidSortField       = errors.New("invalid sort field")
	ErrInvalidSortOrder       = errors.New("invalid sort order (must be 'asc' or 'desc')")
	ErrInvalidGroupBy         = errors.New("invalid group_by field")
	ErrWorkspaceRequired      = errors.New("workspace_id is required")
	ErrUnauthorizedViewAccess = errors.New("unauthorized view access")
)

// SavedViewService orchestrates saved views, filter validation, pinning, cloning, and execution
type SavedViewService interface {
	CreateView(userID string, req model.CreateViewRequest) (*model.SavedView, error)
	GetView(userID string, id int64) (*model.SavedView, error)
	ListViews(userID string, workspaceID int64, pinnedOnly bool) ([]*model.SavedView, error)
	UpdateView(userID string, id int64, req model.UpdateViewRequest) (*model.SavedView, error)
	DeleteView(userID string, id int64) error
	TogglePinView(userID string, id int64) (bool, error)
	DuplicateView(userID string, id int64, req model.DuplicateViewRequest) (*model.SavedView, error)
	ExecuteView(userID string, id int64) (*model.ViewExecutionResult, error)
}

type savedViewService struct {
	repo     repository.SavedViewRepository
	taskRepo repository.TaskRepository
}

// NewSavedViewService constructs a new SavedViewService instance
func NewSavedViewService(repo repository.SavedViewRepository, taskRepo repository.TaskRepository) SavedViewService {
	return &savedViewService{
		repo:     repo,
		taskRepo: taskRepo,
	}
}

func (s *savedViewService) CreateView(userID string, req model.CreateViewRequest) (*model.SavedView, error) {
	return nil, nil
}

func (s *savedViewService) GetView(userID string, id int64) (*model.SavedView, error) {
	return nil, nil
}

func (s *savedViewService) ListViews(userID string, workspaceID int64, pinnedOnly bool) ([]*model.SavedView, error) {
	return nil, nil
}

func (s *savedViewService) UpdateView(userID string, id int64, req model.UpdateViewRequest) (*model.SavedView, error) {
	return nil, nil
}

func (s *savedViewService) DeleteView(userID string, id int64) error {
	return nil
}

func (s *savedViewService) TogglePinView(userID string, id int64) (bool, error) {
	return false, nil
}

func (s *savedViewService) DuplicateView(userID string, id int64, req model.DuplicateViewRequest) (*model.SavedView, error) {
	return nil, nil
}

func (s *savedViewService) ExecuteView(userID string, id int64) (*model.ViewExecutionResult, error) {
	return nil, nil
}
