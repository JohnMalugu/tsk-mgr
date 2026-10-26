package service

import (
	"errors"
	"fmt"
	"strings"
	"time"

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
	if req.WorkspaceID <= 0 {
		return nil, ErrWorkspaceRequired
	}
	validatedName, err := validateReleaseName(req.Name)
	if err != nil {
		return nil, err
	}
	validatedVersion, err := validateReleaseVersion(req.Version)
	if err != nil {
		return nil, err
	}

	rel := &model.Release{
		WorkspaceID:     req.WorkspaceID,
		ProjectID:       req.ProjectID,
		Name:            validatedName,
		Version:         validatedVersion,
		Description:     strings.TrimSpace(req.Description),
		Status:          model.ReleaseStatusPlanning,
		TargetDate:      req.TargetDate,
		CreatedByUserID: userID,
	}

	id, err := s.repo.CreateRelease(rel)
	if err != nil {
		return nil, err
	}

	return s.repo.GetReleaseByID(id)
}

func (s *releaseServiceImpl) GetRelease(id int64) (*model.Release, error) {
	rel, err := s.repo.GetReleaseByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrReleaseNotFound) {
			return nil, ErrReleaseNotFound
		}
		return nil, err
	}
	return rel, nil
}

func (s *releaseServiceImpl) ListReleases(filter model.ReleaseFilter) ([]*model.ReleaseSummary, error) {
	if filter.WorkspaceID <= 0 {
		return nil, ErrWorkspaceRequired
	}
	return s.repo.ListReleases(filter)
}

func (s *releaseServiceImpl) UpdateRelease(id int64, req model.UpdateReleaseRequest) (*model.Release, error) {
	existing, err := s.GetRelease(id)
	if err != nil {
		return nil, err
	}

	if existing.Status == model.ReleaseStatusReleased || existing.Status == model.ReleaseStatusArchived {
		return nil, ErrReleaseAlreadyReleased
	}

	if req.Name != nil {
		val, err := validateReleaseName(*req.Name)
		if err != nil {
			return nil, err
		}
		existing.Name = val
	}
	if req.Version != nil {
		val, err := validateReleaseVersion(*req.Version)
		if err != nil {
			return nil, err
		}
		existing.Version = val
	}
	if req.Description != nil {
		existing.Description = strings.TrimSpace(*req.Description)
	}
	if req.TargetDate != nil {
		existing.TargetDate = req.TargetDate
	}

	if err := s.repo.UpdateRelease(existing); err != nil {
		return nil, err
	}

	return s.repo.GetReleaseByID(id)
}

func (s *releaseServiceImpl) UpdateReleaseStatus(id int64, req model.UpdateReleaseStatusRequest) (*model.Release, error) {
	existing, err := s.GetRelease(id)
	if err != nil {
		return nil, err
	}

	targetStatus := strings.ToLower(strings.TrimSpace(req.Status))
	if err := validateReleaseStatusTransition(existing.Status, targetStatus); err != nil {
		return nil, err
	}

	var releasedAt *time.Time
	if targetStatus == model.ReleaseStatusReleased {
		now := time.Now()
		releasedAt = &now
	} else if existing.ReleasedAt != nil {
		releasedAt = existing.ReleasedAt
	}

	if err := s.repo.UpdateReleaseStatus(id, targetStatus, releasedAt); err != nil {
		return nil, err
	}

	return s.repo.GetReleaseByID(id)
}

func (s *releaseServiceImpl) DeleteRelease(id int64) error {
	err := s.repo.DeleteRelease(id)
	if err != nil {
		if errors.Is(err, repository.ErrReleaseNotFound) {
			return ErrReleaseNotFound
		}
		return err
	}
	return nil
}

func (s *releaseServiceImpl) AddTasksToRelease(userID string, releaseID int64, req model.AddReleaseTasksRequest) ([]*model.ReleaseItem, error) {
	rel, err := s.GetRelease(releaseID)
	if err != nil {
		return nil, err
	}

	if rel.Status == model.ReleaseStatusFreeze {
		return nil, ErrReleaseCodeFrozen
	}
	if rel.Status == model.ReleaseStatusReleased || rel.Status == model.ReleaseStatusArchived {
		return nil, ErrReleaseAlreadyReleased
	}
	if len(req.Tasks) == 0 {
		return nil, ErrNoTasksSpecified
	}

	for _, t := range req.Tasks {
		cat, err := validateReleaseCategory(t.Category)
		if err != nil {
			return nil, err
		}
		incChangelog := true
		if t.IncludeInChangelog != nil {
			incChangelog = *t.IncludeInChangelog
		}

		item := &model.ReleaseItem{
			ReleaseID:          releaseID,
			TaskID:             t.TaskID,
			Category:           cat,
			IncludeInChangelog: incChangelog,
			CustomNotes:        strings.TrimSpace(t.CustomNotes),
			AddedByUserID:      userID,
		}
		if err := s.repo.AddReleaseItem(item); err != nil {
			return nil, err
		}
	}

	return s.repo.ListReleaseItems(releaseID, "")
}

func (s *releaseServiceImpl) RemoveTaskFromRelease(releaseID int64, taskID int) error {
	rel, err := s.GetRelease(releaseID)
	if err != nil {
		return err
	}

	if rel.Status == model.ReleaseStatusFreeze {
		return ErrReleaseCodeFrozen
	}
	if rel.Status == model.ReleaseStatusReleased || rel.Status == model.ReleaseStatusArchived {
		return ErrReleaseAlreadyReleased
	}

	err = s.repo.RemoveReleaseItem(releaseID, taskID)
	if err != nil {
		if errors.Is(err, repository.ErrReleaseItemNotFound) {
			return ErrReleaseItemNotFound
		}
		return err
	}
	return nil
}

func (s *releaseServiceImpl) UpdateReleaseItem(releaseID int64, taskID int, req model.UpdateReleaseItemRequest) (*model.ReleaseItem, error) {
	rel, err := s.GetRelease(releaseID)
	if err != nil {
		return nil, err
	}

	if rel.Status == model.ReleaseStatusReleased || rel.Status == model.ReleaseStatusArchived {
		return nil, ErrReleaseAlreadyReleased
	}

	existing, err := s.repo.GetReleaseItem(releaseID, taskID)
	if err != nil {
		if errors.Is(err, repository.ErrReleaseItemNotFound) {
			return nil, ErrReleaseItemNotFound
		}
		return nil, err
	}

	if req.Category != nil {
		cat, err := validateReleaseCategory(*req.Category)
		if err != nil {
			return nil, err
		}
		existing.Category = cat
	}
	if req.IncludeInChangelog != nil {
		existing.IncludeInChangelog = *req.IncludeInChangelog
	}
	if req.CustomNotes != nil {
		existing.CustomNotes = strings.TrimSpace(*req.CustomNotes)
	}

	if err := s.repo.UpdateReleaseItem(existing); err != nil {
		return nil, err
	}

	return s.repo.GetReleaseItem(releaseID, taskID)
}

func (s *releaseServiceImpl) ListReleaseItems(releaseID int64, category string) ([]*model.ReleaseItem, error) {
	if _, err := s.GetRelease(releaseID); err != nil {
		return nil, err
	}
	return s.repo.ListReleaseItems(releaseID, category)
}

func (s *releaseServiceImpl) GenerateReleaseNotes(releaseID int64, opts model.ReleaseNotesOptions) (*model.GeneratedReleaseNotes, error) {
	rel, err := s.GetRelease(releaseID)
	if err != nil {
		return nil, err
	}

	items, err := s.repo.ListReleaseItems(releaseID, "")
	if err != nil {
		return nil, err
	}

	format := strings.ToLower(strings.TrimSpace(opts.Format))
	if format != model.ReleaseNotesFormatJSON {
		format = model.ReleaseNotesFormatMarkdown
	}

	categoryTitles := map[string]string{
		model.ReleaseCategoryFeature:     "Features",
		model.ReleaseCategoryBugfix:      "Bug Fixes",
		model.ReleaseCategoryImprovement: "Improvements",
		model.ReleaseCategorySecurity:    "Security Patches",
		model.ReleaseCategoryChore:       "Chores & Maintenance",
	}

	categoryOrder := []string{
		model.ReleaseCategoryFeature,
		model.ReleaseCategoryBugfix,
		model.ReleaseCategoryImprovement,
		model.ReleaseCategorySecurity,
		model.ReleaseCategoryChore,
	}

	sections := make(map[string][]model.ChangelogEntry)
	totalItems := 0

	for _, item := range items {
		if !item.IncludeInChangelog {
			continue
		}
		if !opts.IncludeUncompleted && !item.TaskCompleted {
			continue
		}

		entry := model.ChangelogEntry{
			TaskID:      item.TaskID,
			Title:       item.TaskTitle,
			Category:    item.Category,
			Completed:   item.TaskCompleted,
			CustomNotes: item.CustomNotes,
		}
		catKey := item.Category
		if catKey == "" {
			catKey = model.ReleaseCategoryFeature
		}
		sections[catKey] = append(sections[catKey], entry)
		totalItems++
	}

	var contentBuilder strings.Builder
	if format == model.ReleaseNotesFormatMarkdown {
		header := opts.HeaderTitle
		if header == "" {
			header = fmt.Sprintf("%s (%s)", rel.Name, rel.Version)
		}
		contentBuilder.WriteString(fmt.Sprintf("# %s\n\n", header))

		if rel.Description != "" {
			contentBuilder.WriteString(fmt.Sprintf("%s\n\n", rel.Description))
		}

		if rel.ReleasedAt != nil {
			contentBuilder.WriteString(fmt.Sprintf("**Released On:** %s\n\n", rel.ReleasedAt.Format("2006-01-02")))
		} else if rel.TargetDate != nil {
			contentBuilder.WriteString(fmt.Sprintf("**Target Date:** %s (Status: %s)\n\n", rel.TargetDate.Format("2006-01-02"), rel.Status))
		} else {
			contentBuilder.WriteString(fmt.Sprintf("**Status:** %s\n\n", rel.Status))
		}

		for _, cat := range categoryOrder {
			entries := sections[cat]
			if len(entries) == 0 {
				continue
			}
			title := categoryTitles[cat]
			contentBuilder.WriteString(fmt.Sprintf("### %s\n\n", title))
			for _, e := range entries {
				desc := e.Title
				if e.CustomNotes != "" {
					desc = fmt.Sprintf("%s - %s", e.Title, e.CustomNotes)
				}
				statusTag := ""
				if !e.Completed {
					statusTag = " *(in progress)*"
				}
				contentBuilder.WriteString(fmt.Sprintf("- %s (#%d)%s\n", desc, e.TaskID, statusTag))
			}
			contentBuilder.WriteString("\n")
		}

		if totalItems == 0 {
			contentBuilder.WriteString("*No changelog items recorded for this release.*\n")
		}
	}

	return &model.GeneratedReleaseNotes{
		ReleaseID:   rel.ID,
		Name:        rel.Name,
		Version:     rel.Version,
		Status:      rel.Status,
		ReleasedAt:  rel.ReleasedAt,
		Format:      format,
		Content:     contentBuilder.String(),
		Sections:    sections,
		TotalItems:  totalItems,
		GeneratedAt: time.Now(),
	}, nil
}

func (s *releaseServiceImpl) GetReleaseMetrics(releaseID int64) (*model.ReleaseMetrics, error) {
	return nil, nil
}

func (s *releaseServiceImpl) CloneRelease(userID string, sourceReleaseID int64, req model.CloneReleaseRequest) (*model.Release, error) {
	return nil, nil
}
