package service

import (
	"database/sql"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/JohnMalugu/tsk-mgr-api/internal/db"
	"github.com/JohnMalugu/tsk-mgr-api/internal/model"
	"github.com/JohnMalugu/tsk-mgr-api/internal/repository"
)

func setupTestReleaseSvc(t *testing.T) (ReleaseService, repository.TaskRepository, *sql.DB) {
	sqliteDB, err := db.InitDB(":memory:")
	if err != nil {
		t.Fatalf("failed to initialize test in-memory db: %v", err)
	}

	_, err = sqliteDB.Exec(`
		INSERT INTO workspaces (id, name, slug, owner_id, created_at, updated_at) 
		VALUES (1, 'Engineering WS', 'eng-ws-rel', 'user_1', datetime('now'), datetime('now')),
		       (2, 'Platform WS', 'plat-ws-rel', 'user_1', datetime('now'), datetime('now'))
	`)
	if err != nil {
		t.Fatalf("failed to insert test workspaces: %v", err)
	}

	relRepo := repository.NewReleaseRepository(sqliteDB)
	taskRepo := repository.NewTaskRepository(sqliteDB)
	svc := NewReleaseService(relRepo, taskRepo)
	return svc, taskRepo, sqliteDB
}

func TestReleaseService_Validation(t *testing.T) {
	svc, _, _ := setupTestReleaseSvc(t)

	// 1. Missing workspace ID
	_, err := svc.CreateRelease("user_1", model.CreateReleaseRequest{
		WorkspaceID: 0,
		Name:        "Release v1.0",
		Version:     "v1.0.0",
	})
	if !errors.Is(err, ErrWorkspaceRequired) {
		t.Fatalf("expected ErrWorkspaceRequired, got: %v", err)
	}

	// 2. Empty release name
	_, err = svc.CreateRelease("user_1", model.CreateReleaseRequest{
		WorkspaceID: 1,
		Name:        "   ",
		Version:     "v1.0.0",
	})
	if !errors.Is(err, ErrReleaseNameRequired) {
		t.Fatalf("expected ErrReleaseNameRequired, got: %v", err)
	}

	// 3. Name exceeds 100 runes
	_, err = svc.CreateRelease("user_1", model.CreateReleaseRequest{
		WorkspaceID: 1,
		Name:        strings.Repeat("r", 101),
		Version:     "v1.0.0",
	})
	if err == nil || !strings.Contains(err.Error(), "cannot exceed 100 characters") {
		t.Fatalf("expected name length error, got: %v", err)
	}

	// 4. Empty release version
	_, err = svc.CreateRelease("user_1", model.CreateReleaseRequest{
		WorkspaceID: 1,
		Name:        "Q1 Release",
		Version:     "   ",
	})
	if !errors.Is(err, ErrInvalidReleaseVersion) {
		t.Fatalf("expected ErrInvalidReleaseVersion, got: %v", err)
	}

	// 5. Invalid release version characters
	_, err = svc.CreateRelease("user_1", model.CreateReleaseRequest{
		WorkspaceID: 1,
		Name:        "Q1 Release",
		Version:     "v1.0.0@#$",
	})
	if !errors.Is(err, ErrInvalidReleaseVersion) {
		t.Fatalf("expected ErrInvalidReleaseVersion for special chars, got: %v", err)
	}

	// 6. Invalid status transitions
	if err := validateReleaseStatusTransition(model.ReleaseStatusPlanning, "unknown_status"); !errors.Is(err, ErrInvalidReleaseStatus) {
		t.Fatalf("expected ErrInvalidReleaseStatus, got: %v", err)
	}
	if err := validateReleaseStatusTransition(model.ReleaseStatusPlanning, model.ReleaseStatusReleased); !errors.Is(err, ErrInvalidStatusTransition) {
		t.Fatalf("expected ErrInvalidStatusTransition from planning to released, got: %v", err)
	}
	if err := validateReleaseStatusTransition(model.ReleaseStatusReleased, model.ReleaseStatusPlanning); !errors.Is(err, ErrInvalidStatusTransition) {
		t.Fatalf("expected ErrInvalidStatusTransition from released to planning, got: %v", err)
	}

	// 7. Invalid category
	if _, err := validateReleaseCategory("nonexistent"); !errors.Is(err, ErrInvalidReleaseCategory) {
		t.Fatalf("expected ErrInvalidReleaseCategory, got: %v", err)
	}
}

func TestReleaseService_CRUD(t *testing.T) {
	svc, _, _ := setupTestReleaseSvc(t)

	targetDate := time.Now().Add(14 * 24 * time.Hour)
	created, err := svc.CreateRelease("user_1", model.CreateReleaseRequest{
		WorkspaceID: 1,
		Name:        "Spring Platform Release",
		Version:     "v2.5.0",
		Description: "Major architecture upgrades and UI improvements",
		TargetDate:  &targetDate,
	})
	if err != nil {
		t.Fatalf("failed to create release: %v", err)
	}
	if created.ID == 0 {
		t.Fatalf("expected non-zero ID")
	}
	if created.Status != model.ReleaseStatusPlanning {
		t.Fatalf("expected initial status to be planning, got '%s'", created.Status)
	}

	// GetRelease
	fetched, err := svc.GetRelease(created.ID)
	if err != nil {
		t.Fatalf("failed to get release: %v", err)
	}
	if fetched.Name != "Spring Platform Release" || fetched.Version != "v2.5.0" {
		t.Fatalf("mismatched release data: %+v", fetched)
	}

	// UpdateRelease
	newName := "Spring Platform Enterprise Release"
	newVer := "v2.5.1"
	updated, err := svc.UpdateRelease(created.ID, model.UpdateReleaseRequest{
		Name:    &newName,
		Version: &newVer,
	})
	if err != nil {
		t.Fatalf("failed to update release: %v", err)
	}
	if updated.Name != newName || updated.Version != newVer {
		t.Fatalf("updated release mismatch: %+v", updated)
	}

	// ListReleases
	list, err := svc.ListReleases(model.ReleaseFilter{
		WorkspaceID: 1,
	})
	if err != nil {
		t.Fatalf("failed to list releases: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("expected 1 release in list, got %d", len(list))
	}
	if list[0].ID != created.ID {
		t.Fatalf("expected summary to match created release ID")
	}

	// DeleteRelease
	err = svc.DeleteRelease(created.ID)
	if err != nil {
		t.Fatalf("failed to delete release: %v", err)
	}

	_, err = svc.GetRelease(created.ID)
	if !errors.Is(err, ErrReleaseNotFound) {
		t.Fatalf("expected ErrReleaseNotFound for deleted release, got: %v", err)
	}
}
