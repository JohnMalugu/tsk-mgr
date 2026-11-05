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

func setupTestRiskSvc(t *testing.T) (RiskService, repository.RiskRepository, repository.ProjectRepository, repository.TaskRepository, *sql.DB) {
	sqliteDB, err := db.InitDB(":memory:")
	if err != nil {
		t.Fatalf("failed to initialize test in-memory db: %v", err)
	}

	_, err = sqliteDB.Exec(`
		INSERT INTO users (id, username, email, password_hash, created_at, updated_at) 
		VALUES ('user_1', 'alice', 'alice@example.com', 'hash', datetime('now'), datetime('now')),
		       ('user_2', 'bob', 'bob@example.com', 'hash', datetime('now'), datetime('now'));

		INSERT INTO workspaces (id, name, slug, owner_id, created_at, updated_at) 
		VALUES (1, 'Engineering WS', 'eng-ws-rsk', 'user_1', datetime('now'), datetime('now')),
		       (2, 'Product WS', 'prod-ws-rsk', 'user_1', datetime('now'), datetime('now'));

		INSERT INTO projects (id, user_id, name, description, color, created_at, updated_at)
		VALUES (101, 'user_1', 'Infrastructure', 'Cloud Infrastructure', '#4f46e5', datetime('now'), datetime('now')),
		       (102, 'user_1', 'Billing Engine', 'Billing Pipeline', '#10b981', datetime('now'), datetime('now'));

		INSERT INTO tasks (id, user_id, project_id, title, description, status, completed, due_date, estimate_minutes, priority, created_at, updated_at)
		VALUES (501, 'user_1', 101, 'Add DB Replication', 'Mitigation task for single point of failure', 'todo', 0, datetime('now'), 60, 'high', datetime('now'), datetime('now')),
		       (502, 'user_1', 101, 'Implement Circuit Breakers', 'Contingency measure', 'in_progress', 0, datetime('now'), 45, 'medium', datetime('now'), datetime('now'));
	`)
	if err != nil {
		t.Fatalf("failed to seed test database: %v", err)
	}

	riskRepo := repository.NewRiskRepository(sqliteDB)
	projectRepo := repository.NewProjectRepository(sqliteDB)
	taskRepo := repository.NewTaskRepository(sqliteDB)
	riskSvc := NewRiskService(riskRepo, projectRepo, taskRepo)

	return riskSvc, riskRepo, projectRepo, taskRepo, sqliteDB
}

func TestRiskService_CreateAndGet(t *testing.T) {
	svc, _, _, _, dbConn := setupTestRiskSvc(t)
	defer dbConn.Close()

	// 1. Validation: Workspace required
	_, err := svc.CreateRisk("user_1", model.CreateRiskRequest{
		WorkspaceID: 0,
		Title:       "Valid Title",
	})
	if !errors.Is(err, ErrWorkspaceRequired) {
		t.Fatalf("expected ErrWorkspaceRequired, got %v", err)
	}

	// 2. Validation: Title required
	_, err = svc.CreateRisk("user_1", model.CreateRiskRequest{
		WorkspaceID: 1,
		Title:       "   ",
	})
	if !errors.Is(err, ErrRiskTitleRequired) {
		t.Fatalf("expected ErrRiskTitleRequired, got %v", err)
	}

	// 3. Validation: Invalid project
	invalidProjID := 9999
	_, err = svc.CreateRisk("user_1", model.CreateRiskRequest{
		WorkspaceID: 1,
		ProjectID:   &invalidProjID,
		Title:       "Risk with bad project",
	})
	if !errors.Is(err, ErrProjectNotFound) {
		t.Fatalf("expected ErrProjectNotFound, got %v", err)
	}

	// 4. Validation: Invalid probability and impact
	_, err = svc.CreateRisk("user_1", model.CreateRiskRequest{
		WorkspaceID: 1,
		Title:       "Bad prob",
		Probability: 7,
		Impact:      3,
	})
	if !errors.Is(err, ErrInvalidProbability) {
		t.Fatalf("expected ErrInvalidProbability, got %v", err)
	}

	_, err = svc.CreateRisk("user_1", model.CreateRiskRequest{
		WorkspaceID: 1,
		Title:       "Bad impact",
		Probability: 3,
		Impact:      6,
	})
	if !errors.Is(err, ErrInvalidImpact) {
		t.Fatalf("expected ErrInvalidImpact, got %v", err)
	}

	_, err = svc.CreateRisk("user_1", model.CreateRiskRequest{
		WorkspaceID: 1,
		Title:       "Bad impact negative",
		Probability: 3,
		Impact:      -1,
	})
	if !errors.Is(err, ErrInvalidImpact) {
		t.Fatalf("expected ErrInvalidImpact, got %v", err)
	}

	// 5. Validation: Invalid category and strategy
	_, err = svc.CreateRisk("user_1", model.CreateRiskRequest{
		WorkspaceID: 1,
		Title:       "Bad cat",
		Category:    "unknown_category",
	})
	if !errors.Is(err, ErrInvalidRiskCategory) {
		t.Fatalf("expected ErrInvalidRiskCategory, got %v", err)
	}

	_, err = svc.CreateRisk("user_1", model.CreateRiskRequest{
		WorkspaceID: 1,
		Title:       "Bad strat",
		Strategy:    "unknown_strategy",
	})
	if !errors.Is(err, ErrInvalidRiskStrategy) {
		t.Fatalf("expected ErrInvalidRiskStrategy, got %v", err)
	}

	// 6. Valid Creation
	projID := 101
	targetDate := time.Now().Add(14 * 24 * time.Hour).UTC()
	created, err := svc.CreateRisk("user_1", model.CreateRiskRequest{
		WorkspaceID:          1,
		ProjectID:            &projID,
		Title:                "Database Failover Vulnerability",
		Description:          "Primary node lacks automated seamless failover",
		Category:             model.RiskCategoryTechnical,
		Probability:          4,
		Impact:               5,
		Strategy:             model.RiskStrategyMitigate,
		MitigationPlan:       "Deploy Patroni replication cluster",
		ContingencyPlan:      "Manual promotion script with runbook",
		TargetResolutionDate: &targetDate,
	})
	if err != nil {
		t.Fatalf("failed to create valid risk: %v", err)
	}

	if created.ID <= 0 {
		t.Errorf("expected positive risk ID, got %d", created.ID)
	}
	if created.RiskNumber != 1 {
		t.Errorf("expected RiskNumber 1, got %d", created.RiskNumber)
	}
	if created.RiskCode != "RSK-1" {
		t.Errorf("expected RiskCode 'RSK-1', got '%s'", created.RiskCode)
	}
	if created.Score != 20 {
		t.Errorf("expected Score 20 (4*5), got %d", created.Score)
	}
	if created.Severity != model.RiskSeverityCritical {
		t.Errorf("expected Severity critical, got %s", created.Severity)
	}
	if created.Status != model.RiskStatusIdentified {
		t.Errorf("expected Status identified, got %s", created.Status)
	}
	if created.OwnerID == nil || *created.OwnerID != "user_1" {
		t.Errorf("expected OwnerID user_1, got %v", created.OwnerID)
	}

	// 7. Get by ID
	fetched, err := svc.GetRisk(created.ID)
	if err != nil {
		t.Fatalf("failed to get risk by id: %v", err)
	}
	if fetched.Title != created.Title {
		t.Errorf("title mismatch: %s vs %s", fetched.Title, created.Title)
	}

	// 8. Get by Code
	fetchedCode, err := svc.GetRiskByCode(1, "RSK-1")
	if err != nil {
		t.Fatalf("failed to get risk by code: %v", err)
	}
	if fetchedCode.ID != created.ID {
		t.Errorf("id mismatch: %d vs %d", fetchedCode.ID, created.ID)
	}

	// 9. Not Found errors
	_, err = svc.GetRisk(99999)
	if !errors.Is(err, ErrRiskNotFound) {
		t.Fatalf("expected ErrRiskNotFound, got %v", err)
	}
	_, err = svc.GetRiskByCode(1, "RSK-999")
	if !errors.Is(err, ErrRiskNotFound) {
		t.Fatalf("expected ErrRiskNotFound, got %v", err)
	}
}

func TestRiskService_List(t *testing.T) {
	svc, _, _, _, dbConn := setupTestRiskSvc(t)
	defer dbConn.Close()

	// Seed multiple risks
	categories := []string{model.RiskCategoryTechnical, model.RiskCategoryFinancial, model.RiskCategoryOperational}
	for i, cat := range categories {
		prob := i + 2
		imp := i + 2
		_, err := svc.CreateRisk("user_1", model.CreateRiskRequest{
			WorkspaceID: 1,
			Title:       "Test Risk " + cat,
			Category:    cat,
			Probability: prob,
			Impact:      imp,
		})
		if err != nil {
			t.Fatalf("failed to seed risk: %v", err)
		}
	}

	// 1. List all for workspace 1
	list, err := svc.ListRisks(model.RiskFilter{WorkspaceID: 1})
	if err != nil {
		t.Fatalf("failed to list risks: %v", err)
	}
	if len(list) != 3 {
		t.Fatalf("expected 3 risks, got %d", len(list))
	}

	// 2. Filter by Category
	techList, err := svc.ListRisks(model.RiskFilter{
		WorkspaceID: 1,
		Category:    model.RiskCategoryTechnical,
	})
	if err != nil {
		t.Fatalf("failed to list technical risks: %v", err)
	}
	if len(techList) != 1 || techList[0].Category != model.RiskCategoryTechnical {
		t.Fatalf("expected 1 technical risk, got %d", len(techList))
	}

	// 3. Filter by Search
	searchList, err := svc.ListRisks(model.RiskFilter{
		WorkspaceID: 1,
		Search:      "financial",
	})
	if err != nil {
		t.Fatalf("failed to search risks: %v", err)
	}
	if len(searchList) != 1 {
		t.Fatalf("expected 1 search result, got %d", len(searchList))
	}

	// 4. Workspace 0 error
	_, err = svc.ListRisks(model.RiskFilter{WorkspaceID: 0})
	if !errors.Is(err, ErrWorkspaceRequired) {
		t.Fatalf("expected ErrWorkspaceRequired, got %v", err)
	}
}

func TestRiskService_UpdateAndDelete(t *testing.T) {
	svc, _, _, _, dbConn := setupTestRiskSvc(t)
	defer dbConn.Close()

	created, err := svc.CreateRisk("user_1", model.CreateRiskRequest{
		WorkspaceID: 1,
		Title:       "Original Title",
		Description: "Original Desc",
		Category:    model.RiskCategoryOperational,
	})
	if err != nil {
		t.Fatalf("failed to create risk: %v", err)
	}

	// Update fields
	newTitle := "Updated Title"
	newCat := model.RiskCategoryCompliance
	newStrat := model.RiskStrategyTransfer
	updated, err := svc.UpdateRisk(created.ID, model.UpdateRiskRequest{
		Title:    &newTitle,
		Category: &newCat,
		Strategy: &newStrat,
	})
	if err != nil {
		t.Fatalf("failed to update risk: %v", err)
	}
	if updated.Title != newTitle {
		t.Errorf("expected title %s, got %s", newTitle, updated.Title)
	}
	if updated.Category != newCat {
		t.Errorf("expected category %s, got %s", newCat, updated.Category)
	}
	if updated.Strategy != newStrat {
		t.Errorf("expected strategy %s, got %s", newStrat, updated.Strategy)
	}

	// Delete
	err = svc.DeleteRisk(created.ID)
	if err != nil {
		t.Fatalf("failed to delete risk: %v", err)
	}

	_, err = svc.GetRisk(created.ID)
	if !errors.Is(err, ErrRiskNotFound) {
		t.Fatalf("expected ErrRiskNotFound after deletion, got %v", err)
	}
}

func TestRiskService_AssessRisk(t *testing.T) {
	svc, _, _, _, dbConn := setupTestRiskSvc(t)
	defer dbConn.Close()

	created, err := svc.CreateRisk("user_1", model.CreateRiskRequest{
		WorkspaceID: 1,
		Title:       "Vendor Dependency Outage",
		Probability: 2,
		Impact:      2, // Score = 4 (low)
	})
	if err != nil {
		t.Fatalf("failed to create risk: %v", err)
	}
	if created.Score != 4 || created.Severity != model.RiskSeverityLow {
		t.Fatalf("expected initial score 4/low, got %d/%s", created.Score, created.Severity)
	}

	// Re-assess to Prob 5, Impact 5 -> Score 25 (critical)
	assessed, err := svc.AssessRisk(created.ID, "user_2", model.AssessRiskRequest{
		Probability: 5,
		Impact:      5,
		Comments:    "Third-party provider had 3 major incidents this month",
	})
	if err != nil {
		t.Fatalf("failed to assess risk: %v", err)
	}
	if assessed.Score != 25 || assessed.Severity != model.RiskSeverityCritical {
		t.Fatalf("expected assessed score 25/critical, got %d/%s", assessed.Score, assessed.Severity)
	}

	// Verify review audit trail
	reviews, err := svc.ListReviews(created.ID)
	if err != nil {
		t.Fatalf("failed to list reviews: %v", err)
	}
	if len(reviews) != 1 {
		t.Fatalf("expected 1 review, got %d", len(reviews))
	}
	rev := reviews[0]
	if rev.PreviousScore != 4 || rev.NewScore != 25 {
		t.Errorf("expected score change 4 -> 25, got %d -> %d", rev.PreviousScore, rev.NewScore)
	}
	if rev.ReviewerID != "user_2" {
		t.Errorf("expected reviewer user_2, got %s", rev.ReviewerID)
	}
}

func TestRiskService_ChangeStatus(t *testing.T) {
	svc, _, _, _, dbConn := setupTestRiskSvc(t)
	defer dbConn.Close()

	created, err := svc.CreateRisk("user_1", model.CreateRiskRequest{
		WorkspaceID: 1,
		Title:       "API Rate Limit Exhaustion",
	})
	if err != nil {
		t.Fatalf("failed to create risk: %v", err)
	}

	// Change status to mitigating
	updated, err := svc.ChangeStatus(created.ID, "user_1", model.ChangeRiskStatusRequest{
		Status:   model.RiskStatusMitigating,
		Comments: "Engineering team implementing Redis rate limiters",
	})
	if err != nil {
		t.Fatalf("failed to change status to mitigating: %v", err)
	}
	if updated.Status != model.RiskStatusMitigating {
		t.Fatalf("expected status mitigating, got %s", updated.Status)
	}

	// Change status to closed -> should set ClosedAt
	closed, err := svc.ChangeStatus(created.ID, "user_1", model.ChangeRiskStatusRequest{
		Status:   model.RiskStatusClosed,
		Comments: "Rate limiting deployed to production and verified",
	})
	if err != nil {
		t.Fatalf("failed to close risk: %v", err)
	}
	if closed.Status != model.RiskStatusClosed {
		t.Fatalf("expected status closed, got %s", closed.Status)
	}
	if closed.ClosedAt == nil {
		t.Fatalf("expected ClosedAt timestamp, got nil")
	}

	// Reopen to monitored -> should clear ClosedAt
	reopened, err := svc.ChangeStatus(created.ID, "user_1", model.ChangeRiskStatusRequest{
		Status:   model.RiskStatusMonitored,
		Comments: "Reopening for extended monitoring during peak traffic",
	})
	if err != nil {
		t.Fatalf("failed to reopen risk: %v", err)
	}
	if reopened.ClosedAt != nil {
		t.Fatalf("expected ClosedAt to be nil on reopen, got %v", reopened.ClosedAt)
	}
}

func TestRiskService_TaskLinks(t *testing.T) {
	svc, _, _, _, dbConn := setupTestRiskSvc(t)
	defer dbConn.Close()

	risk, err := svc.CreateRisk("user_1", model.CreateRiskRequest{
		WorkspaceID: 1,
		Title:       "Data Loss on Hardware Failure",
	})
	if err != nil {
		t.Fatalf("failed to create risk: %v", err)
	}

	// 1. Link task 501
	link, err := svc.LinkTask(risk.ID, model.LinkRiskTaskRequest{
		TaskID:     501,
		ActionType: model.RiskActionMitigation,
		Notes:      "Setup replica nodes in separate availability zones",
	})
	if err != nil {
		t.Fatalf("failed to link task: %v", err)
	}
	if link.TaskID != 501 || link.ActionType != model.RiskActionMitigation {
		t.Fatalf("invalid link payload: %+v", link)
	}
	if link.TaskTitle != "Add DB Replication" {
		t.Errorf("expected TaskTitle 'Add DB Replication', got '%s'", link.TaskTitle)
	}

	// 2. Link non-existent task -> ErrTaskNotFound
	_, err = svc.LinkTask(risk.ID, model.LinkRiskTaskRequest{
		TaskID: 9999,
	})
	if !errors.Is(err, ErrTaskNotFound) {
		t.Fatalf("expected ErrTaskNotFound, got %v", err)
	}

	// 3. List linked tasks
	links, err := svc.ListRiskTasks(risk.ID)
	if err != nil {
		t.Fatalf("failed to list risk tasks: %v", err)
	}
	if len(links) != 1 {
		t.Fatalf("expected 1 task link, got %d", len(links))
	}

	// 4. Unlink task
	err = svc.UnlinkTask(risk.ID, 501)
	if err != nil {
		t.Fatalf("failed to unlink task: %v", err)
	}

	linksAfter, err := svc.ListRiskTasks(risk.ID)
	if err != nil {
		t.Fatalf("failed to list risk tasks after unlink: %v", err)
	}
	if len(linksAfter) != 0 {
		t.Fatalf("expected 0 task links, got %d", len(linksAfter))
	}
}

func TestRiskService_Reviews(t *testing.T) {
	svc, _, _, _, dbConn := setupTestRiskSvc(t)
	defer dbConn.Close()

	risk, err := svc.CreateRisk("user_1", model.CreateRiskRequest{
		WorkspaceID: 1,
		Title:       "Single Datacenter Dependency",
		Probability: 3,
		Impact:      4,
	})
	if err != nil {
		t.Fatalf("failed to create risk: %v", err)
	}

	// Create review with comments and prob reduction
	newP := 2
	rev, err := svc.CreateReview(risk.ID, "user_2", model.CreateRiskReviewRequest{
		NewProbability: &newP,
		Comments:       "Multi-region failover tested successfully in staging",
	})
	if err != nil {
		t.Fatalf("failed to create review: %v", err)
	}
	if rev.NewProbability != 2 || rev.NewScore != 8 {
		t.Errorf("expected new prob 2, score 8 (2*4), got prob %d, score %d", rev.NewProbability, rev.NewScore)
	}

	// Verify risk updated
	updatedRisk, err := svc.GetRisk(risk.ID)
	if err != nil {
		t.Fatalf("failed to get risk: %v", err)
	}
	if updatedRisk.Score != 8 || updatedRisk.Severity != model.RiskSeverityMedium {
		t.Fatalf("expected risk score 8/medium, got %d/%s", updatedRisk.Score, updatedRisk.Severity)
	}

	// Empty comments error
	_, err = svc.CreateReview(risk.ID, "user_2", model.CreateRiskReviewRequest{
		Comments: "   ",
	})
	if !errors.Is(err, ErrReviewCommentReq) {
		t.Fatalf("expected ErrReviewCommentReq, got %v", err)
	}
}

func TestRiskService_Metrics(t *testing.T) {
	svc, _, _, _, dbConn := setupTestRiskSvc(t)
	defer dbConn.Close()

	// Seed 3 risks with different severities and statuses
	// 1: Critical (5*5=25) active
	_, err := svc.CreateRisk("user_1", model.CreateRiskRequest{
		WorkspaceID: 1,
		Title:       "R1 Critical",
		Category:    model.RiskCategorySecurity,
		Probability: 5,
		Impact:      5,
	})
	if err != nil {
		t.Fatalf("failed to seed R1: %v", err)
	}

	// 2: Medium (3*3=9) active
	_, err = svc.CreateRisk("user_1", model.CreateRiskRequest{
		WorkspaceID: 1,
		Title:       "R2 Medium",
		Category:    model.RiskCategoryOperational,
		Probability: 3,
		Impact:      3,
	})
	if err != nil {
		t.Fatalf("failed to seed R2: %v", err)
	}

	// 3: Low (1*2=2) closed
	r3, err := svc.CreateRisk("user_1", model.CreateRiskRequest{
		WorkspaceID: 1,
		Title:       "R3 Closed",
		Category:    model.RiskCategoryFinancial,
		Probability: 1,
		Impact:      2,
	})
	if err != nil {
		t.Fatalf("failed to seed R3: %v", err)
	}
	_, err = svc.ChangeStatus(r3.ID, "user_1", model.ChangeRiskStatusRequest{
		Status: model.RiskStatusClosed,
	})
	if err != nil {
		t.Fatalf("failed to close R3: %v", err)
	}

	// Fetch metrics
	metrics, err := svc.GetRiskMetrics(1)
	if err != nil {
		t.Fatalf("failed to get risk metrics: %v", err)
	}

	if metrics.TotalRisks != 3 {
		t.Errorf("expected 3 total risks, got %d", metrics.TotalRisks)
	}
	if metrics.ActiveRisks != 2 {
		t.Errorf("expected 2 active risks, got %d", metrics.ActiveRisks)
	}
	if metrics.ClosedRisks != 1 {
		t.Errorf("expected 1 closed risk, got %d", metrics.ClosedRisks)
	}
	if metrics.TotalExposureScore != 34 { // 25 + 9
		t.Errorf("expected total exposure score 34 (25+9), got %d", metrics.TotalExposureScore)
	}
	if metrics.SeverityBreakdown[model.RiskSeverityCritical] != 1 {
		t.Errorf("expected 1 critical risk, got %d", metrics.SeverityBreakdown[model.RiskSeverityCritical])
	}
	if metrics.HeatmapMatrix[4][4] != 1 { // Probability 5, Impact 5 -> index [4][4]
		t.Errorf("expected 1 risk at heatmap [4][4], got %d", metrics.HeatmapMatrix[4][4])
	}
	if len(metrics.CriticalRisks) != 1 {
		t.Errorf("expected 1 critical risk in top list, got %d", len(metrics.CriticalRisks))
	}
}
