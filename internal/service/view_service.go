package service

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

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
	view, err := s.repo.GetViewByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrViewNotFound) {
			return nil, ErrViewNotFound
		}
		return nil, err
	}

	if view.UserID != userID {
		return nil, ErrUnauthorizedViewAccess
	}

	if req.Name != nil {
		name, err := validateViewName(*req.Name)
		if err != nil {
			return nil, err
		}
		view.Name = name
	}

	if req.Description != nil {
		view.Description = strings.TrimSpace(*req.Description)
	}

	if req.Icon != nil {
		icon := strings.TrimSpace(*req.Icon)
		if icon == "" {
			icon = "bookmark"
		}
		view.Icon = icon
	}

	if req.Color != nil {
		color := strings.TrimSpace(*req.Color)
		if color == "" {
			color = "#4F46E5"
		}
		view.Color = color
	}

	if req.FilterCriteria != nil {
		view.FilterCriteria = sanitizeFilterCriteria(req.FilterCriteria)
	}

	if req.SortBy != nil {
		sb, err := normalizeSortBy(*req.SortBy)
		if err != nil {
			return nil, err
		}
		view.SortBy = sb
	}

	if req.SortOrder != nil {
		so, err := normalizeSortOrder(*req.SortOrder)
		if err != nil {
			return nil, err
		}
		view.SortOrder = so
	}

	if req.GroupBy != nil {
		gb, err := normalizeGroupBy(*req.GroupBy)
		if err != nil {
			return nil, err
		}
		view.GroupBy = gb
	}

	if req.IsShared != nil {
		view.IsShared = *req.IsShared
	}

	if req.IsPinned != nil {
		view.IsPinned = *req.IsPinned
	}

	if req.Position != nil {
		view.Position = *req.Position
	}

	if err := s.repo.UpdateView(view); err != nil {
		return nil, err
	}

	return s.repo.GetViewByID(id)
}

func (s *savedViewService) DeleteView(userID string, id int64) error {
	view, err := s.repo.GetViewByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrViewNotFound) {
			return ErrViewNotFound
		}
		return err
	}

	if view.UserID != userID {
		return ErrUnauthorizedViewAccess
	}

	return s.repo.DeleteView(id)
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
	existing, err := s.repo.GetViewByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrViewNotFound) {
			return nil, ErrViewNotFound
		}
		return nil, err
	}

	if !existing.IsShared && existing.UserID != userID {
		return nil, ErrUnauthorizedViewAccess
	}

	targetWs := existing.WorkspaceID
	if req.TargetWorkspaceID != nil && *req.TargetWorkspaceID > 0 {
		targetWs = *req.TargetWorkspaceID
	}

	var newName string
	if strings.TrimSpace(req.NewName) != "" {
		validated, err := validateViewName(req.NewName)
		if err != nil {
			return nil, err
		}
		newName = validated
	} else {
		newName = fmt.Sprintf("%s (Copy)", existing.Name)
		if len(newName) > 100 {
			newName = newName[:100]
		}
	}

	cloned := &model.SavedView{
		WorkspaceID:    targetWs,
		UserID:         userID,
		Name:           newName,
		Description:    existing.Description,
		Icon:           existing.Icon,
		Color:          existing.Color,
		FilterCriteria: existing.FilterCriteria,
		SortBy:         existing.SortBy,
		SortOrder:      existing.SortOrder,
		GroupBy:        existing.GroupBy,
		IsShared:       false,
		IsPinned:       false,
		Position:       existing.Position + 1,
	}

	newID, err := s.repo.CreateView(cloned)
	if err != nil {
		return nil, err
	}

	return s.repo.GetViewByID(newID)
}

var priorityWeights = map[string]int{
	"urgent": 4,
	"high":   3,
	"medium": 2,
	"low":    1,
}

func (s *savedViewService) ExecuteView(userID string, id int64) (*model.ViewExecutionResult, error) {
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

	_ = s.repo.IncrementViewUsage(id)

	// Fetch tasks for the current user and view owner if shared
	tasksMap := make(map[int]*model.Task)
	userTasks, err := s.taskRepo.GetByUserID(userID)
	if err == nil {
		for _, t := range userTasks {
			tasksMap[t.ID] = t
		}
	}
	if view.IsShared && view.UserID != userID {
		ownerTasks, err := s.taskRepo.GetByUserID(view.UserID)
		if err == nil {
			for _, t := range ownerTasks {
				tasksMap[t.ID] = t
			}
		}
	}

	var allTasks []*model.Task
	for _, t := range tasksMap {
		allTasks = append(allTasks, t)
	}

	fc := view.FilterCriteria
	var filtered []*model.Task

	statusSet := make(map[string]bool)
	for _, st := range fc.Status {
		statusSet[strings.ToLower(st)] = true
	}

	prioritySet := make(map[string]bool)
	for _, pr := range fc.Priority {
		prioritySet[strings.ToLower(pr)] = true
	}

	tagSet := make(map[string]bool)
	for _, tg := range fc.Tags {
		tagSet[strings.ToLower(tg)] = true
	}

	for _, task := range allTasks {
		// Status filter
		if len(statusSet) > 0 {
			if !statusSet[strings.ToLower(task.Status)] {
				continue
			}
		}

		// Priority filter
		if len(prioritySet) > 0 {
			if !prioritySet[strings.ToLower(task.Priority)] {
				continue
			}
		}

		// Project ID filter
		if fc.ProjectID != nil {
			if task.ProjectID == nil || *task.ProjectID != *fc.ProjectID {
				continue
			}
		}

		// Assignee filter
		if fc.AssigneeUserID != nil && *fc.AssigneeUserID != "" {
			if task.UserID != *fc.AssigneeUserID {
				continue
			}
		}

		// Tags filter (at least one tag match)
		if len(tagSet) > 0 {
			matchedTag := false
			for _, tg := range task.Tags {
				if tagSet[strings.ToLower(tg)] {
					matchedTag = true
					break
				}
			}
			if !matchedTag {
				continue
			}
		}

		// Has Due Date filter
		if fc.HasDueDate != nil {
			hasDue := !task.DueDate.IsZero()
			if hasDue != *fc.HasDueDate {
				continue
			}
		}

		// Due Before filter
		if fc.DueBefore != nil {
			if task.DueDate.IsZero() || !task.DueDate.Before(*fc.DueBefore) {
				continue
			}
		}

		// Due After filter
		if fc.DueAfter != nil {
			if task.DueDate.IsZero() || !task.DueDate.After(*fc.DueAfter) {
				continue
			}
		}

		// Search term filter (title or description)
		if fc.Search != "" {
			searchLower := strings.ToLower(fc.Search)
			titleMatch := strings.Contains(strings.ToLower(task.Title), searchLower)
			descMatch := strings.Contains(strings.ToLower(task.Description), searchLower)
			if !titleMatch && !descMatch {
				continue
			}
		}

		filtered = append(filtered, task)
	}

	// Sort tasks
	isAsc := view.SortOrder == model.SortOrderAsc
	sort.SliceStable(filtered, func(i, j int) bool {
		ti, tj := filtered[i], filtered[j]
		var cmp bool
		switch view.SortBy {
		case model.SortByDueDate:
			if ti.DueDate.IsZero() && !tj.DueDate.IsZero() {
				cmp = false
			} else if !ti.DueDate.IsZero() && tj.DueDate.IsZero() {
				cmp = true
			} else {
				cmp = ti.DueDate.Before(tj.DueDate)
			}
		case model.SortByPriority:
			wi := priorityWeights[strings.ToLower(ti.Priority)]
			wj := priorityWeights[strings.ToLower(tj.Priority)]
			cmp = wi < wj
		case model.SortByTitle:
			cmp = strings.ToLower(ti.Title) < strings.ToLower(tj.Title)
		case model.SortByStatus:
			cmp = strings.ToLower(ti.Status) < strings.ToLower(tj.Status)
		case model.SortByUpdatedAt:
			cmp = ti.UpdatedAt.Before(tj.UpdatedAt)
		case model.SortByCreatedAt:
			fallthrough
		default:
			cmp = ti.CreatedAt.Before(tj.CreatedAt)
		}

		if isAsc {
			return cmp
		}
		return !cmp
	})

	// Group Aggregations
	var groupCounts []model.ViewGroupCount
	if view.GroupBy != "" && view.GroupBy != model.GroupByNone {
		counts := make(map[string]int)
		for _, task := range filtered {
			var key string
			switch view.GroupBy {
			case model.GroupByStatus:
				key = task.Status
				if key == "" {
					key = "unspecified"
				}
			case model.GroupByPriority:
				key = task.Priority
				if key == "" {
					key = "unspecified"
				}
			case model.GroupByProjectID:
				if task.ProjectID != nil {
					key = fmt.Sprintf("%d", *task.ProjectID)
				} else {
					key = "no_project"
				}
			case model.GroupByAssignee:
				key = task.UserID
				if key == "" {
					key = "unassigned"
				}
			default:
				key = "all"
			}
			counts[key]++
		}

		var keys []string
		for k := range counts {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			groupCounts = append(groupCounts, model.ViewGroupCount{
				GroupKey: k,
				Count:    counts[k],
			})
		}
	}

	return &model.ViewExecutionResult{
		View:        view,
		TotalTasks:  len(filtered),
		Tasks:       filtered,
		GroupCounts: groupCounts,
		ExecutedAt:  time.Now(),
	}, nil
}
