package service

import (
	"testing"

	"github.com/JohnMalugu/tsk-mgr-api/internal/db"
	"github.com/JohnMalugu/tsk-mgr-api/internal/repository"
)

func setupTestProjectSvc(t *testing.T) ProjectService {
	sqliteDB, err := db.InitDB(":memory:")
	if err != nil {
		t.Fatalf("failed to init db: %v", err)
	}

	repo := repository.NewProjectRepository(sqliteDB)
	return NewProjectService(repo)
}

func TestProjectService_CRUD(t *testing.T) {
	svc := setupTestProjectSvc(t)
	userID := "user-123"

	// 1. Validation error on empty name
	_, err := svc.CreateProject(userID, "   ", "desc", "#ffffff")
	if err != ErrProjectNameRequired {
		t.Fatalf("expected ErrProjectNameRequired, got %v", err)
	}

	// 2. Create project
	proj, err := svc.CreateProject(userID, "DevOps", "Infrastructure tasks", "#ff5722")
	if err != nil {
		t.Fatalf("failed to create project: %v", err)
	}
	if proj.ID == 0 {
		t.Errorf("expected positive project ID")
	}
	if proj.Name != "DevOps" {
		t.Errorf("expected name DevOps, got %s", proj.Name)
	}

	// 3. Get project
	got, err := svc.GetProject(proj.ID, userID)
	if err != nil {
		t.Fatalf("failed to get project: %v", err)
	}
	if got.ID != proj.ID {
		t.Errorf("expected ID %d, got %d", proj.ID, got.ID)
	}

	// 4. Unauthorized access
	_, err = svc.GetProject(proj.ID, "user-999")
	if err != ErrUnauthorizedProject {
		t.Fatalf("expected ErrUnauthorizedProject, got %v", err)
	}

	// 5. Update project
	updated, err := svc.UpdateProject(proj.ID, userID, "Cloud & DevOps", "Updated description", "#00bcd4")
	if err != nil {
		t.Fatalf("failed to update project: %v", err)
	}
	if updated.Name != "Cloud & DevOps" {
		t.Errorf("expected Cloud & DevOps, got %s", updated.Name)
	}

	// 6. List projects
	list, err := svc.GetUserProjects(userID)
	if err != nil {
		t.Fatalf("failed to list projects: %v", err)
	}
	if len(list) != 1 {
		t.Errorf("expected 1 project, got %d", len(list))
	}

	// 7. Delete project
	err = svc.DeleteProject(proj.ID, userID)
	if err != nil {
		t.Fatalf("failed to delete project: %v", err)
	}

	// 8. Confirm deleted
	_, err = svc.GetProject(proj.ID, userID)
	if err != ErrProjectNotFound {
		t.Fatalf("expected ErrProjectNotFound, got %v", err)
	}
}
