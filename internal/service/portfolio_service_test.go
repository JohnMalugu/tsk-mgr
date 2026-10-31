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

func setupTestPortfolioSvc(t *testing.T) (PortfolioService, repository.ProjectRepository, *sql.DB) {
	sqliteDB, err := db.InitDB(":memory:")
	if err != nil {
		t.Fatalf("failed to initialize test in-memory db: %v", err)
	}

	_, err = sqliteDB.Exec(`
		INSERT INTO users (id, username, email, password_hash, created_at, updated_at) 
		VALUES ('user_1', 'user1', 'user1@example.com', 'hash', datetime('now'), datetime('now')),
		       ('user_2', 'user2', 'user2@example.com', 'hash', datetime('now'), datetime('now'));

		INSERT INTO workspaces (id, name, slug, owner_id, created_at, updated_at) 
		VALUES (1, 'Engineering WS', 'eng-ws-pf', 'user_1', datetime('now'), datetime('now')),
		       (2, 'Product WS', 'prod-ws-pf', 'user_1', datetime('now'), datetime('now'));

		INSERT INTO projects (id, user_id, name, description, color, created_at, updated_at)
		VALUES (101, 'user_1', 'Core API', 'Core Backend Service', '#4f46e5', datetime('now'), datetime('now')),
		       (102, 'user_1', 'Web Client', 'React Web Frontend', '#10b981', datetime('now'), datetime('now'));
	`)
	if err != nil {
		t.Fatalf("failed to insert test seed data: %v", err)
	}

	portfolioRepo := repository.NewPortfolioRepository(sqliteDB)
	projectRepo := repository.NewProjectRepository(sqliteDB)
	svc := NewPortfolioService(portfolioRepo, projectRepo)
	return svc, projectRepo, sqliteDB
}

func TestPortfolioService_PortfolioCRUD(t *testing.T) {
	svc, _, _ := setupTestPortfolioSvc(t)

	// 1. Validation: Name required
	_, err := svc.CreatePortfolio("user_1", model.CreatePortfolioRequest{
		WorkspaceID: 1,
		Name:        "   ",
	})
	if !errors.Is(err, ErrPortfolioNameRequired) {
		t.Fatalf("expected ErrPortfolioNameRequired, got: %v", err)
	}

	// 2. Validation: Workspace required
	_, err = svc.CreatePortfolio("user_1", model.CreatePortfolioRequest{
		WorkspaceID: 0,
		Name:        "Strategic Cloud",
	})
	if !errors.Is(err, ErrWorkspaceRequired) {
		t.Fatalf("expected ErrWorkspaceRequired, got: %v", err)
	}

	// 3. Create portfolio successfully
	budget := 150000.0
	startDate := time.Now().Add(-24 * time.Hour)
	targetDate := time.Now().Add(180 * 24 * time.Hour)

	pf, err := svc.CreatePortfolio("user_1", model.CreatePortfolioRequest{
		WorkspaceID:    1,
		Name:           "Strategic Cloud Modernization",
		Description:    "Migrate services to microservices architecture",
		Color:          "#6366f1",
		StartDate:      &startDate,
		TargetDate:     &targetDate,
		BudgetCurrency: "USD",
		BudgetAmount:   &budget,
	})
	if err != nil {
		t.Fatalf("failed to create portfolio: %v", err)
	}
	if pf.ID <= 0 {
		t.Fatalf("expected positive portfolio ID, got %d", pf.ID)
	}
	if pf.Status != model.PortfolioStatusOnTrack {
		t.Fatalf("expected initial status %q, got %q", model.PortfolioStatusOnTrack, pf.Status)
	}
	if pf.BudgetAmount != 150000.0 {
		t.Fatalf("expected budget 150000.0, got %f", pf.BudgetAmount)
	}

	// 4. Get portfolio detail
	detail, err := svc.GetPortfolio(pf.ID)
	if err != nil {
		t.Fatalf("failed to get portfolio: %v", err)
	}
	if detail.Portfolio.Name != "Strategic Cloud Modernization" {
		t.Fatalf("unexpected portfolio name: %s", detail.Portfolio.Name)
	}
	if len(detail.Projects) != 0 || len(detail.Milestones) != 0 {
		t.Fatalf("expected empty initial projects and milestones, got %d and %d", len(detail.Projects), len(detail.Milestones))
	}
	if detail.OverallProgress != 0 {
		t.Fatalf("expected initial overall progress 0, got %f", detail.OverallProgress)
	}

	// 5. Get non-existent portfolio
	_, err = svc.GetPortfolio(999999)
	if !errors.Is(err, ErrPortfolioNotFound) {
		t.Fatalf("expected ErrPortfolioNotFound, got: %v", err)
	}

	// 6. Update portfolio validations
	emptyName := "  "
	_, err = svc.UpdatePortfolio(pf.ID, model.UpdatePortfolioRequest{
		Name: &emptyName,
	})
	if !errors.Is(err, ErrPortfolioNameRequired) {
		t.Fatalf("expected ErrPortfolioNameRequired, got: %v", err)
	}

	invalidStatus := "in_limbo"
	_, err = svc.UpdatePortfolio(pf.ID, model.UpdatePortfolioRequest{
		Status: &invalidStatus,
	})
	if !errors.Is(err, ErrInvalidPortfolioStatus) {
		t.Fatalf("expected ErrInvalidPortfolioStatus, got: %v", err)
	}

	// 7. Successful update
	newName := "Cloud Modernization 2.0"
	newStatus := model.PortfolioStatusAtRisk
	newActualCost := 35000.0
	updated, err := svc.UpdatePortfolio(pf.ID, model.UpdatePortfolioRequest{
		Name:       &newName,
		Status:     &newStatus,
		ActualCost: &newActualCost,
	})
	if err != nil {
		t.Fatalf("failed to update portfolio: %v", err)
	}
	if updated.Name != newName || updated.Status != newStatus || updated.ActualCost != 35000.0 {
		t.Fatalf("portfolio update not applied properly: %+v", updated)
	}

	// 8. List portfolios with filter
	wsID := int64(1)
	list, err := svc.ListPortfolios(model.PortfolioFilter{
		WorkspaceID: &wsID,
		Status:      model.PortfolioStatusAtRisk,
	})
	if err != nil {
		t.Fatalf("failed to list portfolios: %v", err)
	}
	if len(list) != 1 || list[0].ID != pf.ID {
		t.Fatalf("expected 1 portfolio matching filter, got %d", len(list))
	}

	// 9. Delete portfolio
	err = svc.DeletePortfolio(pf.ID)
	if err != nil {
		t.Fatalf("failed to delete portfolio: %v", err)
	}

	// Check it no longer exists
	_, err = svc.GetPortfolio(pf.ID)
	if !errors.Is(err, ErrPortfolioNotFound) {
		t.Fatalf("expected ErrPortfolioNotFound after delete, got: %v", err)
	}

	// Delete non-existent
	err = svc.DeletePortfolio(pf.ID)
	if !errors.Is(err, ErrPortfolioNotFound) {
		t.Fatalf("expected ErrPortfolioNotFound on repeat delete, got: %v", err)
	}
}

func TestPortfolioService_Projects(t *testing.T) {
	svc, _, sqliteDB := setupTestPortfolioSvc(t)

	// Create a portfolio
	pf, err := svc.CreatePortfolio("user_1", model.CreatePortfolioRequest{
		WorkspaceID: 1,
		Name:        "Mobile & Backend Initiatives",
	})
	if err != nil {
		t.Fatalf("failed to create portfolio: %v", err)
	}

	// 1. Add non-existent project
	_, err = svc.AddProject(pf.ID, model.AddPortfolioProjectRequest{
		ProjectID: 9999,
	})
	if !errors.Is(err, ErrProjectNotFound) {
		t.Fatalf("expected ErrProjectNotFound, got: %v", err)
	}

	// 2. Add valid project
	allocBudget := 50000.0
	item, err := svc.AddProject(pf.ID, model.AddPortfolioProjectRequest{
		ProjectID:       101,
		Priority:        model.PortfolioProjectPriorityHigh,
		AllocatedBudget: &allocBudget,
	})
	if err != nil {
		t.Fatalf("failed to add project to portfolio: %v", err)
	}
	if item.ProjectID != 101 || item.ProjectName != "Core API" {
		t.Fatalf("unexpected project details: %+v", item)
	}
	if item.Priority != model.PortfolioProjectPriorityHigh {
		t.Fatalf("expected priority high, got: %s", item.Priority)
	}

	// 3. Add duplicate project
	_, err = svc.AddProject(pf.ID, model.AddPortfolioProjectRequest{
		ProjectID: 101,
	})
	if !errors.Is(err, ErrPortfolioProjectExists) {
		t.Fatalf("expected ErrPortfolioProjectExists, got: %v", err)
	}

	// 4. List projects
	projects, err := svc.ListProjects(pf.ID)
	if err != nil {
		t.Fatalf("failed to list projects: %v", err)
	}
	if len(projects) != 1 || projects[0].ProjectID != 101 {
		t.Fatalf("expected 1 project in list, got %d", len(projects))
	}

	// 5. Update portfolio project
	newPriority := model.PortfolioProjectPriorityCritical
	spent := 12500.0
	updatedItem, err := svc.UpdateProject(pf.ID, 101, model.UpdatePortfolioProjectRequest{
		Priority:    &newPriority,
		SpentAmount: &spent,
	})
	if err != nil {
		t.Fatalf("failed to update project item: %v", err)
	}
	if updatedItem.Priority != model.PortfolioProjectPriorityCritical || updatedItem.SpentAmount != 12500.0 {
		t.Fatalf("update not applied properly: %+v", updatedItem)
	}

	// 6. Insert tasks to test task progress rollup
	_, err = sqliteDB.Exec(`
		INSERT INTO tasks (user_id, project_id, title, status, completed, due_date, created_at, updated_at, priority)
		VALUES ('user_1', 101, 'Task 1', 'completed', 1, datetime('now'), datetime('now'), datetime('now'), 'high'),
		       ('user_1', 101, 'Task 2', 'in_progress', 0, datetime('now'), datetime('now'), datetime('now'), 'medium');
	`)
	if err != nil {
		t.Fatalf("failed to insert tasks: %v", err)
	}

	// Check portfolio detail rollup
	detail, err := svc.GetPortfolio(pf.ID)
	if err != nil {
		t.Fatalf("failed to get portfolio: %v", err)
	}
	if len(detail.Projects) != 1 {
		t.Fatalf("expected 1 project, got %d", len(detail.Projects))
	}
	if detail.Projects[0].TaskCount != 2 || detail.Projects[0].CompletedTaskCount != 1 {
		t.Fatalf("expected 2 tasks and 1 completed, got %d and %d", detail.Projects[0].TaskCount, detail.Projects[0].CompletedTaskCount)
	}
	if detail.OverallProgress != 50.0 {
		t.Fatalf("expected 50%% overall progress, got: %f", detail.OverallProgress)
	}

	// 7. Remove project
	err = svc.RemoveProject(pf.ID, 101)
	if err != nil {
		t.Fatalf("failed to remove project: %v", err)
	}

	// Removing again returns ErrPortfolioProjectNF
	err = svc.RemoveProject(pf.ID, 101)
	if !errors.Is(err, ErrPortfolioProjectNF) {
		t.Fatalf("expected ErrPortfolioProjectNF, got: %v", err)
	}
}

func TestPortfolioService_Milestones(t *testing.T) {
	svc, _, _ := setupTestPortfolioSvc(t)

	pf, err := svc.CreatePortfolio("user_1", model.CreatePortfolioRequest{
		WorkspaceID: 1,
		Name:        "Milestone Tracking Initiative",
	})
	if err != nil {
		t.Fatalf("failed to create portfolio: %v", err)
	}

	// 1. Validation: Title required
	targetDate := time.Now().Add(30 * 24 * time.Hour)
	_, err = svc.CreateMilestone(pf.ID, model.CreatePortfolioMilestoneRequest{
		Title:      "  ",
		TargetDate: targetDate,
	})
	if !errors.Is(err, ErrMilestoneTitleRequired) {
		t.Fatalf("expected ErrMilestoneTitleRequired, got: %v", err)
	}

	// 2. Validation: Target date required
	_, err = svc.CreateMilestone(pf.ID, model.CreatePortfolioMilestoneRequest{
		Title: "Alpha Release",
	})
	if !errors.Is(err, ErrMilestoneDateRequired) {
		t.Fatalf("expected ErrMilestoneDateRequired, got: %v", err)
	}

	// 3. Valid milestone
	ms, err := svc.CreateMilestone(pf.ID, model.CreatePortfolioMilestoneRequest{
		Title:       "Alpha Release",
		Description: "Complete core architecture test",
		TargetDate:  targetDate,
	})
	if err != nil {
		t.Fatalf("failed to create milestone: %v", err)
	}
	if ms.Status != model.MilestoneStatusPending {
		t.Fatalf("expected pending status, got: %s", ms.Status)
	}

	// 4. List milestones
	milestones, err := svc.ListMilestones(pf.ID)
	if err != nil {
		t.Fatalf("failed to list milestones: %v", err)
	}
	if len(milestones) != 1 || milestones[0].ID != ms.ID {
		t.Fatalf("expected 1 milestone, got %d", len(milestones))
	}

	// 5. Update milestone to completed
	completedStatus := model.MilestoneStatusCompleted
	updatedMs, err := svc.UpdateMilestone(ms.ID, model.UpdatePortfolioMilestoneRequest{
		Status: &completedStatus,
	})
	if err != nil {
		t.Fatalf("failed to update milestone: %v", err)
	}
	if updatedMs.Status != model.MilestoneStatusCompleted {
		t.Fatalf("expected status completed, got: %s", updatedMs.Status)
	}
	if updatedMs.CompletedAt == nil {
		t.Fatalf("expected CompletedAt to be set when milestone is completed")
	}

	// 6. Delete milestone
	err = svc.DeleteMilestone(ms.ID)
	if err != nil {
		t.Fatalf("failed to delete milestone: %v", err)
	}

	// Delete non-existent
	err = svc.DeleteMilestone(ms.ID)
	if !errors.Is(err, ErrMilestoneNotFound) {
		t.Fatalf("expected ErrMilestoneNotFound on repeat delete, got: %v", err)
	}
}

func TestPortfolioService_Updates(t *testing.T) {
	svc, _, _ := setupTestPortfolioSvc(t)

	pf, err := svc.CreatePortfolio("user_1", model.CreatePortfolioRequest{
		WorkspaceID: 1,
		Name:        "Executive Status Reporting",
	})
	if err != nil {
		t.Fatalf("failed to create portfolio: %v", err)
	}

	// 1. Validation: Summary required
	_, err = svc.CreateUpdate("user_1", pf.ID, model.CreatePortfolioUpdateRequest{
		Summary:      "   ",
		HealthStatus: model.PortfolioStatusOnTrack,
	})
	if !errors.Is(err, ErrUpdateSummaryRequired) {
		t.Fatalf("expected ErrUpdateSummaryRequired, got: %v", err)
	}

	// 2. Validation: Invalid health status
	_, err = svc.CreateUpdate("user_1", pf.ID, model.CreatePortfolioUpdateRequest{
		Summary:      "Weekly briefing",
		HealthStatus: "critical_alert",
	})
	if !errors.Is(err, ErrInvalidHealthStatus) {
		t.Fatalf("expected ErrInvalidHealthStatus, got: %v", err)
	}

	// 3. Create executive update (health status = off_track)
	upd, err := svc.CreateUpdate("user_1", pf.ID, model.CreatePortfolioUpdateRequest{
		HealthStatus:    model.PortfolioStatusOffTrack,
		Summary:         "Delays due to upstream API migration dependencies",
		Accomplishments: "Shipped auth service",
		Blockers:        "Third-party vendor downtime",
		NextSteps:       "Follow up with infrastructure provider",
	})
	if err != nil {
		t.Fatalf("failed to create update: %v", err)
	}
	if upd.HealthStatus != model.PortfolioStatusOffTrack {
		t.Fatalf("expected off_track health status, got %s", upd.HealthStatus)
	}

	// Verify portfolio status automatically synchronized to off_track
	detail, err := svc.GetPortfolio(pf.ID)
	if err != nil {
		t.Fatalf("failed to get portfolio: %v", err)
	}
	if detail.Portfolio.Status != model.PortfolioStatusOffTrack {
		t.Fatalf("expected portfolio status to sync to off_track, got: %s", detail.Portfolio.Status)
	}
	if detail.LatestUpdate == nil || detail.LatestUpdate.ID != upd.ID {
		t.Fatalf("expected detail.LatestUpdate to be populated with newly created update")
	}

	// 4. List updates
	updates, err := svc.ListUpdates(pf.ID)
	if err != nil {
		t.Fatalf("failed to list updates: %v", err)
	}
	if len(updates) != 1 || updates[0].ID != upd.ID {
		t.Fatalf("expected 1 update, got %d", len(updates))
	}
}

func TestPortfolioService_Metrics(t *testing.T) {
	svc, _, sqliteDB := setupTestPortfolioSvc(t)

	budget := 100000.0
	pf, err := svc.CreatePortfolio("user_1", model.CreatePortfolioRequest{
		WorkspaceID:  1,
		Name:         "Metrics Analytics Portfolio",
		BudgetAmount: &budget,
	})
	if err != nil {
		t.Fatalf("failed to create portfolio: %v", err)
	}

	// Link project 101 and 102
	b1 := 40000.0
	b2 := 60000.0
	_, err = svc.AddProject(pf.ID, model.AddPortfolioProjectRequest{
		ProjectID:       101,
		AllocatedBudget: &b1,
	})
	if err != nil {
		t.Fatalf("failed to add project 101: %v", err)
	}
	_, err = svc.AddProject(pf.ID, model.AddPortfolioProjectRequest{
		ProjectID:       102,
		AllocatedBudget: &b2,
	})
	if err != nil {
		t.Fatalf("failed to add project 102: %v", err)
	}

	// Insert tasks: project 101 (2 completed, 1 in_progress, 1 overdue in_progress)
	yesterday := time.Now().Add(-24 * time.Hour).Format(time.RFC3339)
	tomorrow := time.Now().Add(24 * time.Hour).Format(time.RFC3339)

	_, err = sqliteDB.Exec(`
		INSERT INTO tasks (user_id, project_id, title, status, completed, due_date, created_at, updated_at, priority)
		VALUES ('user_1', 101, 'Task A', 'completed', 1, ?, datetime('now'), datetime('now'), 'high'),
		       ('user_1', 101, 'Task B', 'completed', 1, ?, datetime('now'), datetime('now'), 'medium'),
		       ('user_1', 101, 'Task C', 'in_progress', 0, ?, datetime('now'), datetime('now'), 'low'),
		       ('user_1', 101, 'Task D', 'in_progress', 0, ?, datetime('now'), datetime('now'), 'critical');
	`, tomorrow, tomorrow, tomorrow, yesterday)
	if err != nil {
		t.Fatalf("failed to insert tasks: %v", err)
	}

	// Add milestone
	msDate := time.Now().Add(30 * 24 * time.Hour)
	ms, err := svc.CreateMilestone(pf.ID, model.CreatePortfolioMilestoneRequest{
		Title:      "Midterm Review",
		TargetDate: msDate,
	})
	if err != nil {
		t.Fatalf("failed to create milestone: %v", err)
	}

	// Complete milestone
	comp := model.MilestoneStatusCompleted
	_, err = svc.UpdateMilestone(ms.ID, model.UpdatePortfolioMilestoneRequest{
		Status: &comp,
	})
	if err != nil {
		t.Fatalf("failed to complete milestone: %v", err)
	}

	// 1. Fetch portfolio metrics
	m, err := svc.GetPortfolioMetrics(pf.ID)
	if err != nil {
		t.Fatalf("failed to get portfolio metrics: %v", err)
	}
	if m.TotalProjects != 2 {
		t.Fatalf("expected 2 projects, got %d", m.TotalProjects)
	}
	if m.TotalTasks != 4 || m.CompletedTasks != 2 {
		t.Fatalf("expected 4 total tasks and 2 completed tasks, got %d and %d", m.TotalTasks, m.CompletedTasks)
	}
	if m.OverdueTasks != 1 {
		t.Fatalf("expected 1 overdue task, got %d", m.OverdueTasks)
	}
	if m.ProgressPercentage != 50.0 {
		t.Fatalf("expected 50%% progress, got: %f", m.ProgressPercentage)
	}
	if m.TotalMilestones != 1 || m.CompletedMilestones != 1 || m.MilestoneCompletionRate != 100.0 {
		t.Fatalf("unexpected milestone metrics: %+v", m)
	}

	// 2. Fetch workspace metrics
	wm, err := svc.GetWorkspaceMetrics(1)
	if err != nil {
		t.Fatalf("failed to get workspace metrics: %v", err)
	}
	if wm.TotalPortfolios != 1 {
		t.Fatalf("expected 1 portfolio in workspace 1, got %d", wm.TotalPortfolios)
	}
	if wm.TotalBudgetAmount != 100000.0 {
		t.Fatalf("expected 100000.0 total budget, got %f", wm.TotalBudgetAmount)
	}

	// 3. Workspace metrics validation
	_, err = svc.GetWorkspaceMetrics(0)
	if !errors.Is(err, ErrWorkspaceRequired) {
		t.Fatalf("expected ErrWorkspaceRequired for workspace 0, got: %v", err)
	}
}
