package service

import (
	"errors"
	"strings"

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

var allowedSortFields = map[string]bool{
	model.SortByCreatedAt: true,
	model.SortByDueDate:   true,
	model.SortByPriority:  true,
	model.SortByTitle:     true,
	model.SortByStatus:    true,
	model.SortByUpdatedAt: true,
}

var allowedSortOrders = map[string]bool{
	model.SortOrderAsc:  true,
	model.SortOrderDesc: true,
}

var allowedGroupBys = map[string]bool{
	model.GroupByNone:      true,
	model.GroupByStatus:    true,
	model.GroupByPriority:  true,
	model.GroupByProjectID: true,
	model.GroupByAssignee:  true,
}

func validateViewName(name string) (string, error) {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" || len(trimmed) > 100 {
		return "", ErrInvalidViewName
	}
	return trimmed, nil
}

func normalizeSortBy(field string) (string, error) {
	f := strings.ToLower(strings.TrimSpace(field))
	if f == "" {
		return model.SortByCreatedAt, nil
	}
	if !allowedSortFields[f] {
		return "", ErrInvalidSortField
	}
	return f, nil
}

func normalizeSortOrder(order string) (string, error) {
	o := strings.ToLower(strings.TrimSpace(order))
	if o == "" {
		return model.SortOrderDesc, nil
	}
	if !allowedSortOrders[o] {
		return "", ErrInvalidSortOrder
	}
	return o, nil
}

func normalizeGroupBy(group string) (string, error) {
	g := strings.ToLower(strings.TrimSpace(group))
	if g == "" {
		return model.GroupByNone, nil
	}
	if !allowedGroupBys[g] {
		return "", ErrInvalidGroupBy
	}
	return g, nil
}

func sanitizeFilterCriteria(criteria *model.ViewFilterCriteria) model.ViewFilterCriteria {
	if criteria == nil {
		return model.ViewFilterCriteria{}
	}
	res := *criteria

	var cleanedStatuses []string
	for _, s := range res.Status {
		t := strings.ToLower(strings.TrimSpace(s))
		if t != "" {
			cleanedStatuses = append(cleanedStatuses, t)
		}
	}
	res.Status = cleanedStatuses

	var cleanedPriorities []string
	for _, p := range res.Priority {
		t := strings.ToLower(strings.TrimSpace(p))
		if t != "" {
			cleanedPriorities = append(cleanedPriorities, t)
		}
	}
	res.Priority = cleanedPriorities

	var cleanedTags []string
	for _, tag := range res.Tags {
		t := strings.TrimSpace(tag)
		if t != "" {
			cleanedTags = append(cleanedTags, t)
		}
	}
	res.Tags = cleanedTags
	res.Search = strings.TrimSpace(res.Search)

	return res
}


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
	if req.WorkspaceID <= 0 {
		return nil, ErrWorkspaceRequired
	}

	name, err := validateViewName(req.Name)
	if err != nil {
		return nil, err
	}

	sortBy, err := normalizeSortBy(req.SortBy)
	if err != nil {
		return nil, err
	}

	sortOrder, err := normalizeSortOrder(req.SortOrder)
	if err != nil {
		return nil, err
	}

	groupBy, err := normalizeGroupBy(req.GroupBy)
	if err != nil {
		return nil, err
	}

	fc := sanitizeFilterCriteria(req.FilterCriteria)

	icon := strings.TrimSpace(req.Icon)
	if icon == "" {
		icon = "bookmark"
	}

	color := strings.TrimSpace(req.Color)
	if color == "" {
		color = "#4F46E5"
	}

	isShared := false
	if req.IsShared != nil {
		isShared = *req.IsShared
	}

	isPinned := false
	if req.IsPinned != nil {
		isPinned = *req.IsPinned
	}

	position := 0
	if req.Position != nil {
		position = *req.Position
	}

	view := &model.SavedView{
		WorkspaceID:    req.WorkspaceID,
		UserID:         userID,
		Name:           name,
		Description:    strings.TrimSpace(req.Description),
		Icon:           icon,
		Color:          color,
		FilterCriteria: fc,
		SortBy:         sortBy,
		SortOrder:      sortOrder,
		GroupBy:        groupBy,
		IsShared:       isShared,
		IsPinned:       isPinned,
		Position:       position,
	}

	id, err := s.repo.CreateView(view)
	if err != nil {
		return nil, err
	}

	return s.repo.GetViewByID(id)
}

func (s *savedViewService) GetView(userID string, id int64) (*model.SavedView, error) {
	view, err := s.repo.GetViewByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrViewNotFound) {
			return nil, ErrViewNotFound
		}
		return nil, err
	}

	if !view.IsShared && view.UserID != userID {
		return nil, ErrUnauthorizedViewAccess
	}

	return view, nil
}


func (s *savedViewService) ListViews(userID string, workspaceID int64, pinnedOnly bool) ([]*model.SavedView, error) {
	if workspaceID <= 0 {
		return nil, ErrWorkspaceRequired
	}

	return s.repo.ListViews(workspaceID, userID, pinnedOnly)
}

func (s *savedViewService) UpdateView(userID string, id int64, req model.UpdateViewRequest) (*model.SavedView, error) {
	return nil, nil
}

func (s *savedViewService) DeleteView(userID string, id int64) error {
	return nil
}

func (s *savedViewService) TogglePinView(userID string, id int64) (bool, error) {
	view, err := s.repo.GetViewByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrViewNotFound) {
			return false, ErrViewNotFound
		}
		return false, err
	}

	if !view.IsShared && view.UserID != userID {
		return false, ErrUnauthorizedViewAccess
	}

	return s.repo.TogglePinView(id)
}


func (s *savedViewService) DuplicateView(userID string, id int64, req model.DuplicateViewRequest) (*model.SavedView, error) {
	return nil, nil
}

func (s *savedViewService) ExecuteView(userID string, id int64) (*model.ViewExecutionResult, error) {
	return nil, nil
}
