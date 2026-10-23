package service

import (
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/JohnMalugu/tsk-mgr-api/internal/db"
	"github.com/JohnMalugu/tsk-mgr-api/internal/model"
	"github.com/JohnMalugu/tsk-mgr-api/internal/repository"
)

func setupTestTemplateSvc(t *testing.T) (TemplateService, *sql.DB) {
	sqliteDB, err := db.InitDB(":memory:")
	if err != nil {
		t.Fatalf("failed to init in-memory db: %v", err)
	}

	templateRepo := repository.NewTemplateRepository(sqliteDB)
	taskRepo := repository.NewTaskRepository(sqliteDB)
	subtaskRepo := repository.NewSubtaskRepository(sqliteDB)
	svc := NewTemplateService(templateRepo, taskRepo, subtaskRepo)
	return svc, sqliteDB
}

func TestTemplateService_Validation(t *testing.T) {
	svc, _ := setupTestTemplateSvc(t)

	// 1. Title required
	_, err := svc.CreateTemplate(&model.CreateTemplateRequest{
		WorkspaceID:      1,
		CreatedByUserID:  "user-1",
		Title:            "   ",
		Category:         model.TemplateCategoryGeneral,
		DefaultPriority:  "medium",
		EstimatedMinutes: 30,
	})
	if !errors.Is(err, ErrTemplateTitleRequired) {
		t.Fatalf("expected ErrTemplateTitleRequired, got: %v", err)
	}

	// 2. Workspace ID required
	_, err = svc.CreateTemplate(&model.CreateTemplateRequest{
		WorkspaceID:      0,
		CreatedByUserID:  "user-1",
		Title:            "Bug Template",
		Category:         model.TemplateCategoryGeneral,
		DefaultPriority:  "medium",
		EstimatedMinutes: 30,
	})
	if !errors.Is(err, ErrInvalidWorkspaceID) {
		t.Fatalf("expected ErrInvalidWorkspaceID, got: %v", err)
	}

	// 3. User ID required
	_, err = svc.CreateTemplate(&model.CreateTemplateRequest{
		WorkspaceID:      1,
		CreatedByUserID:  "   ",
		Title:            "Bug Template",
		Category:         model.TemplateCategoryGeneral,
		DefaultPriority:  "medium",
		EstimatedMinutes: 30,
	})
	if !errors.Is(err, ErrInvalidUserID) {
		t.Fatalf("expected ErrInvalidUserID, got: %v", err)
	}

	// 4. Invalid Category
	_, err = svc.CreateTemplate(&model.CreateTemplateRequest{
		WorkspaceID:      1,
		CreatedByUserID:  "user-1",
		Title:            "Bug Template",
		Category:         "unknown_category_xyz",
		DefaultPriority:  "medium",
		EstimatedMinutes: 30,
	})
	if !errors.Is(err, ErrInvalidCategory) {
		t.Fatalf("expected ErrInvalidCategory, got: %v", err)
	}

	// 5. Invalid Priority
	_, err = svc.CreateTemplate(&model.CreateTemplateRequest{
		WorkspaceID:      1,
		CreatedByUserID:  "user-1",
		Title:            "Bug Template",
		Category:         model.TemplateCategoryBugReport,
		DefaultPriority:  "super_critical",
		EstimatedMinutes: 30,
	})
	if !errors.Is(err, ErrInvalidPriority) {
		t.Fatalf("expected ErrInvalidPriority, got: %v", err)
	}

	// 6. Negative Estimate
	_, err = svc.CreateTemplate(&model.CreateTemplateRequest{
		WorkspaceID:      1,
		CreatedByUserID:  "user-1",
		Title:            "Bug Template",
		Category:         model.TemplateCategoryBugReport,
		DefaultPriority:  "high",
		EstimatedMinutes: -15,
	})
	if !errors.Is(err, ErrNegativeEstimate) {
		t.Fatalf("expected ErrNegativeEstimate, got: %v", err)
	}
}

func TestTemplateService_CRUD(t *testing.T) {
	svc, _ := setupTestTemplateSvc(t)

	// Create
	created, err := svc.CreateTemplate(&model.CreateTemplateRequest{
		WorkspaceID:      10,
		CreatedByUserID:  "lead-architect",
		Title:            "Standard Feature Blueprint",
		Description:      "End-to-end feature delivery lifecycle",
		Category:         model.TemplateCategoryFeature,
		DefaultPriority:  "high",
		EstimatedMinutes: 120,
		Tags:             []string{"feature", "backend", "v2"},
		ChecklistItems: []model.ChecklistBlueprint{
			{Text: "Write unit tests with race detector"},
			{Text: "Update documentation in README"},
		},
		Subtasks: []model.SubtaskBlueprint{
			{Title: "Schema migration and entity design", Position: 1},
			{Title: "Service business logic and validation", Position: 2},
			{Title: "HTTP route wiring and handler tests", Position: 3},
		},
	})
	if err != nil {
		t.Fatalf("failed to create template: %v", err)
	}
	if created.ID == 0 {
		t.Fatalf("expected assigned ID")
	}

	// Get
	fetched, err := svc.GetTemplate(created.ID)
	if err != nil {
		t.Fatalf("failed to get template: %v", err)
	}
	if fetched.Title != "Standard Feature Blueprint" {
		t.Fatalf("title mismatch: %s", fetched.Title)
	}
	if len(fetched.Subtasks) != 3 {
		t.Fatalf("expected 3 subtasks, got: %d", len(fetched.Subtasks))
	}
	if len(fetched.ChecklistItems) != 2 {
		t.Fatalf("expected 2 checklist items, got: %d", len(fetched.ChecklistItems))
	}
	if len(fetched.Tags) != 3 {
		t.Fatalf("expected 3 tags, got: %d", len(fetched.Tags))
	}

	// List
	wsID := int64(10)
	cat := model.TemplateCategoryFeature
	summary, err := svc.ListTemplates(model.TemplateFilter{
		WorkspaceID: &wsID,
		Category:    &cat,
	})
	if err != nil {
		t.Fatalf("failed to list templates: %v", err)
	}
	if summary.Total != 1 {
		t.Fatalf("expected 1 template in summary, got: %d", summary.Total)
	}

	// Update
	newTitle := "Enhanced Feature Blueprint"
	archived := true
	updated, err := svc.UpdateTemplate(created.ID, &model.UpdateTemplateRequest{
		Title:      &newTitle,
		IsArchived: &archived,
	})
	if err != nil {
		t.Fatalf("failed to update template: %v", err)
	}
	if updated.Title != "Enhanced Feature Blueprint" {
		t.Fatalf("updated title mismatch: %s", updated.Title)
	}
	if !updated.IsArchived {
		t.Fatalf("expected is_archived true")
	}

	// Delete
	if err := svc.DeleteTemplate(created.ID); err != nil {
		t.Fatalf("failed to delete template: %v", err)
	}
	_, err = svc.GetTemplate(created.ID)
	if !errors.Is(err, ErrTemplateNotFound) {
		t.Fatalf("expected ErrTemplateNotFound after delete, got: %v", err)
	}
}

func TestTemplateService_Duplicate(t *testing.T) {
	svc, _ := setupTestTemplateSvc(t)

	original, err := svc.CreateTemplate(&model.CreateTemplateRequest{
		WorkspaceID:      5,
		CreatedByUserID:  "qa-lead",
		Title:            "Regression Smoke Suite",
		Description:      "Automated test pass before deployment",
		Category:         model.TemplateCategoryReview,
		DefaultPriority:  "urgent",
		EstimatedMinutes: 45,
		Tags:             []string{"qa", "ci"},
		Subtasks: []model.SubtaskBlueprint{
			{Title: "Run unit test suite", Position: 1},
			{Title: "Verify smoke API endpoints", Position: 2},
		},
	})
	if err != nil {
		t.Fatalf("failed to create original template: %v", err)
	}

	cloned, err := svc.DuplicateTemplate(original.ID)
	if err != nil {
		t.Fatalf("failed to duplicate template: %v", err)
	}
	if cloned.ID == original.ID {
		t.Fatalf("cloned template must have a unique ID")
	}
	if cloned.Title != "Copy of Regression Smoke Suite" {
		t.Fatalf("expected title 'Copy of Regression Smoke Suite', got %q", cloned.Title)
	}
	if len(cloned.Subtasks) != 2 {
		t.Fatalf("expected 2 cloned subtasks, got: %d", len(cloned.Subtasks))
	}
	if cloned.UsageCount != 0 {
		t.Fatalf("cloned template usage count must be 0, got: %d", cloned.UsageCount)
	}
}

func TestTemplateService_Instantiate(t *testing.T) {
	svc, _ := setupTestTemplateSvc(t)

	tmpl, err := svc.CreateTemplate(&model.CreateTemplateRequest{
		WorkspaceID:      1,
		CreatedByUserID:  "lead-eng",
		Title:            "Hotfix Incident Blueprint",
		Description:      "Standard rapid hotfix mitigation workflow",
		Category:         model.TemplateCategoryBugReport,
		DefaultPriority:  "urgent",
		EstimatedMinutes: 60,
		Tags:             []string{"hotfix", "prod"},
		ChecklistItems: []model.ChecklistBlueprint{
			{Text: "Notify on-call commander"},
			{Text: "Deploy patch to staging"},
		},
		Subtasks: []model.SubtaskBlueprint{
			{Title: "Isolate root cause and reproduce", Position: 1},
			{Title: "Implement fix and run test suite", Position: 2},
			{Title: "Verify in production and resolve incident", Position: 3},
		},
	})
	if err != nil {
		t.Fatalf("failed to create template: %v", err)
	}

	due := time.Now().Add(4 * time.Hour)
	result, err := svc.InstantiateTask(tmpl.ID, &model.TemplateInstantiationRequest{
		Title:          "PROD-9982 Database Latency Spike Mitigation",
		AssigneeUserID: "eng-alex",
		DueDate:        &due,
	})
	if err != nil {
		t.Fatalf("failed to instantiate template: %v", err)
	}

	if result.Task == nil || result.Task.ID == 0 {
		t.Fatalf("instantiated task was not persisted properly")
	}
	if result.Task.Title != "PROD-9982 Database Latency Spike Mitigation" {
		t.Fatalf("task title mismatch: %s", result.Task.Title)
	}
	if result.SubtasksCreated != 3 {
		t.Fatalf("expected 3 subtasks created, got: %d", result.SubtasksCreated)
	}
	if result.ChecklistItemsCreated != 2 {
		t.Fatalf("expected 2 checklist items, got: %d", result.ChecklistItemsCreated)
	}

	// Verify template usage count incremented
	refreshed, err := svc.GetTemplate(tmpl.ID)
	if err != nil {
		t.Fatalf("failed to get template: %v", err)
	}
	if refreshed.UsageCount != 1 {
		t.Fatalf("expected usage count 1, got: %d", refreshed.UsageCount)
	}
	if refreshed.LastUsedAt == nil {
		t.Fatalf("expected last_used_at to be populated")
	}

	// Test archived template instantiation prevention
	archived := true
	_, err = svc.UpdateTemplate(tmpl.ID, &model.UpdateTemplateRequest{
		IsArchived: &archived,
	})
	if err != nil {
		t.Fatalf("failed to archive template: %v", err)
	}

	_, err = svc.InstantiateTask(tmpl.ID, &model.TemplateInstantiationRequest{
		AssigneeUserID: "eng-alex",
	})
	if !errors.Is(err, ErrTemplateArchived) {
		t.Fatalf("expected ErrTemplateArchived, got: %v", err)
	}
}

func TestTemplateService_ExportAndImport(t *testing.T) {
	svc, _ := setupTestTemplateSvc(t)

	orig, err := svc.CreateTemplate(&model.CreateTemplateRequest{
		WorkspaceID:      1,
		CreatedByUserID:  "product-manager",
		Title:            "Quarterly Roadmap Item",
		Description:      "Standard epic decomposition blueprint",
		Category:         model.TemplateCategoryGeneral,
		DefaultPriority:  "medium",
		EstimatedMinutes: 180,
		Tags:             []string{"roadmap", "q4"},
	})
	if err != nil {
		t.Fatalf("failed to create template: %v", err)
	}

	bundle, err := svc.ExportTemplate(orig.ID)
	if err != nil {
		t.Fatalf("failed to export template: %v", err)
	}
	if bundle.Version != "1.0" {
		t.Fatalf("expected version 1.0, got: %s", bundle.Version)
	}
	if bundle.Template.Title != "Quarterly Roadmap Item" {
		t.Fatalf("bundle title mismatch: %s", bundle.Template.Title)
	}

	// Import into Workspace 2
	imported, err := svc.ImportTemplate(&model.ImportTemplateRequest{
		WorkspaceID:     2,
		CreatedByUserID: "new-admin",
		Bundle:          *bundle,
	})
	if err != nil {
		t.Fatalf("failed to import template: %v", err)
	}
	if imported.WorkspaceID != 2 {
		t.Fatalf("expected workspace ID 2, got: %d", imported.WorkspaceID)
	}
	if imported.Title != "Quarterly Roadmap Item" {
		t.Fatalf("imported title mismatch: %s", imported.Title)
	}
}

func TestTemplateService_Stats(t *testing.T) {
	svc, _ := setupTestTemplateSvc(t)
	wsID := int64(77)

	t1, err := svc.CreateTemplate(&model.CreateTemplateRequest{
		WorkspaceID:     wsID,
		CreatedByUserID: "user-1",
		Title:           "Bug Template A",
		Category:        model.TemplateCategoryBugReport,
	})
	if err != nil {
		t.Fatalf("create t1 failed: %v", err)
	}

	t2, err := svc.CreateTemplate(&model.CreateTemplateRequest{
		WorkspaceID:     wsID,
		CreatedByUserID: "user-1",
		Title:           "Feature Template B",
		Category:        model.TemplateCategoryFeature,
	})
	if err != nil {
		t.Fatalf("create t2 failed: %v", err)
	}

	// Instantiate t1 twice
	_, _ = svc.InstantiateTask(t1.ID, &model.TemplateInstantiationRequest{AssigneeUserID: "user-1"})
	_, _ = svc.InstantiateTask(t1.ID, &model.TemplateInstantiationRequest{AssigneeUserID: "user-1"})

	// Instantiate t2 once
	_, _ = svc.InstantiateTask(t2.ID, &model.TemplateInstantiationRequest{AssigneeUserID: "user-1"})

	stats, err := svc.GetTemplateStats(wsID)
	if err != nil {
		t.Fatalf("get template stats failed: %v", err)
	}

	if stats.TotalTemplates != 2 {
		t.Fatalf("expected 2 total templates, got: %d", stats.TotalTemplates)
	}
	if stats.ActiveTemplates != 2 {
		t.Fatalf("expected 2 active templates, got: %d", stats.ActiveTemplates)
	}
	if stats.TotalInstantiations != 3 {
		t.Fatalf("expected 3 total instantiations, got: %d", stats.TotalInstantiations)
	}
	if len(stats.CategoryBreakdown) != 2 {
		t.Fatalf("expected 2 categories in breakdown, got: %d", len(stats.CategoryBreakdown))
	}
	if len(stats.MostPopular) != 2 {
		t.Fatalf("expected 2 most popular templates, got: %d", len(stats.MostPopular))
	}
	if stats.MostPopular[0].ID != t1.ID {
		t.Fatalf("expected t1 to be most popular with 2 instantiations")
	}
}
