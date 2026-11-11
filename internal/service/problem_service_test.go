package service

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/JohnMalugu/tsk-mgr-api/internal/db"
	"github.com/JohnMalugu/tsk-mgr-api/internal/model"
	"github.com/JohnMalugu/tsk-mgr-api/internal/repository"
)

func setupTestProblemSvc(t *testing.T) (ProblemService, repository.ProblemRepository, *sql.DB) {
	sqliteDB, err := db.InitDB(":memory:")
	if err != nil {
		t.Fatalf("failed to initialize test in-memory db: %v", err)
	}

	_, err = sqliteDB.Exec(`
		INSERT INTO users (id, username, email, password_hash, created_at, updated_at) 
		VALUES ('user_1', 'alice', 'alice@example.com', 'hash', datetime('now'), datetime('now')),
		       ('user_2', 'bob', 'bob@example.com', 'hash', datetime('now'), datetime('now'));

		INSERT INTO workspaces (id, name, slug, owner_id, created_at, updated_at) 
		VALUES (1, 'Engineering WS', 'eng-ws-prb', 'user_1', datetime('now'), datetime('now'));

		INSERT INTO projects (id, user_id, name, description, color, created_at, updated_at)
		VALUES (101, 'user_1', 'Core Payments', 'Payment Infrastructure', '#4f46e5', datetime('now'), datetime('now'));

		INSERT INTO tasks (id, user_id, project_id, title, description, status, completed, due_date, estimate_minutes, priority, created_at, updated_at)
		VALUES (501, 'user_1', 101, 'Tune Connection Pool', 'Fix pool timeout', 'todo', 0, datetime('now'), 60, 'high', datetime('now'), datetime('now'));

		INSERT INTO incidents (id, workspace_id, project_id, incident_number, title, summary, severity, status, reported_by_id, started_at, created_at, updated_at)
		VALUES (301, 1, 101, 1, 'Payment Gateway DB Connection Spike', 'DB pool exhausted', 'sev1', 'resolved', 'user_1', datetime('now'), datetime('now'), datetime('now'));

		INSERT INTO configuration_items (id, workspace_id, project_id, ci_code, name, ci_type, status, criticality, environment, owner_user_id, created_at, updated_at)
		VALUES (201, 1, 101, 'CI-1', 'Payments Postgres DB', 'database', 'active', 'tier_1_mission_critical', 'production', 'user_1', datetime('now'), datetime('now'));

		INSERT INTO change_requests (id, workspace_id, project_id, change_number, title, description, change_type, risk_level, impact_level, status, reason_for_change, implementation_plan, rollback_plan, test_plan, requested_by_user_id, created_at, updated_at)
		VALUES (401, 1, 101, 'CHG-1', 'Upgrade Database Pool Parameters', 'Increase max_connections to 200', 'standard', 'medium', 'medium', 'scheduled', 'Prevent connection exhaustion', 'Apply config and restart', 'Revert config', 'Verify connections', 'user_1', datetime('now'), datetime('now'));
	`)
	if err != nil {
		t.Fatalf("failed to insert test seed data: %v", err)
	}

	problemRepo := repository.NewProblemRepository(sqliteDB)
	workspaceRepo := repository.NewWorkspaceRepository(sqliteDB)
	projectRepo := repository.NewProjectRepository(sqliteDB)
	taskRepo := repository.NewTaskRepository(sqliteDB)
	incidentRepo := repository.NewIncidentRepository(sqliteDB)
	cmdbRepo := repository.NewCMDBRepository(sqliteDB)
	changeRepo := repository.NewChangeRepository(sqliteDB)
	userRepo := repository.NewUserRepository(sqliteDB)
	auditRepo := repository.NewAuditRepository(sqliteDB)
	auditSvc := NewAuditService(auditRepo, workspaceRepo, userRepo)

	svc := NewProblemService(problemRepo, workspaceRepo, projectRepo, taskRepo, incidentRepo, cmdbRepo, changeRepo, auditSvc)
	return svc, problemRepo, sqliteDB
}

func TestProblemService_CreateAndGet(t *testing.T) {
	svc, _, _ := setupTestProblemSvc(t)

	// 1. Missing workspace
	_, err := svc.CreateProblem("user_1", model.CreateProblemRequest{
		WorkspaceID: 0,
		Title:       "Recurring Postgres Connection Exhaustion",
	})
	if !errors.Is(err, ErrWorkspaceRequired) {
		t.Fatalf("expected ErrWorkspaceRequired, got %v", err)
	}

	// 2. Non-existent workspace
	_, err = svc.CreateProblem("user_1", model.CreateProblemRequest{
		WorkspaceID: 999,
		Title:       "Recurring Postgres Connection Exhaustion",
	})
	if !errors.Is(err, ErrWorkspaceNotFound) {
		t.Fatalf("expected ErrWorkspaceNotFound, got %v", err)
	}

	// 3. Missing title
	_, err = svc.CreateProblem("user_1", model.CreateProblemRequest{
		WorkspaceID: 1,
		Title:       "   ",
	})
	if !errors.Is(err, ErrProblemTitleRequired) {
		t.Fatalf("expected ErrProblemTitleRequired, got %v", err)
	}

	// 4. Invalid category
	_, err = svc.CreateProblem("user_1", model.CreateProblemRequest{
		WorkspaceID: 1,
		Title:       "Recurring Postgres Connection Exhaustion",
		Category:    "unknown_category",
	})
	if !errors.Is(err, model.ErrInvalidProblemCategory) {
		t.Fatalf("expected ErrInvalidProblemCategory, got %v", err)
	}

	// 5. Successful problem creation
	projID := int64(101)
	prob1, err := svc.CreateProblem("user_1", model.CreateProblemRequest{
		WorkspaceID: 1,
		ProjectID:   &projID,
		Title:       "Recurring Postgres Connection Exhaustion",
		Description: "Multiple services fail with connection pool timeouts during load spikes",
		Category:    model.ProblemCategoryDatabase,
		Priority:    model.ProblemPriorityCritical,
		Urgency:     model.ProblemUrgencyCritical,
		Impact:      model.ProblemImpactCritical,
		RCAMethod:   model.RCAMethod5Whys,
		IncidentIDs: []int64{301},
		CIIDs:       []int64{201},
	})
	if err != nil {
		t.Fatalf("failed to create problem: %v", err)
	}

	if prob1.ProblemCode != "PRB-1" {
		t.Fatalf("expected problem code PRB-1, got %s", prob1.ProblemCode)
	}
	if prob1.Status != model.ProblemStatusLogged {
		t.Fatalf("expected status logged, got %s", prob1.Status)
	}
	if prob1.LinkedIncidentsCount != 1 {
		t.Fatalf("expected 1 linked incident, got %d", prob1.LinkedIncidentsCount)
	}
	if prob1.LinkedCIsCount != 1 {
		t.Fatalf("expected 1 linked CI, got %d", prob1.LinkedCIsCount)
	}

	// 6. Create second problem to verify sequential code generation
	prob2, err := svc.CreateProblem("user_1", model.CreateProblemRequest{
		WorkspaceID: 1,
		Title:       "Redis Cache Memory Leak Under High Concurrency",
		Category:    model.ProblemCategoryInfrastructure,
		Priority:    model.ProblemPriorityHigh,
	})
	if err != nil {
		t.Fatalf("failed to create second problem: %v", err)
	}
	if prob2.ProblemCode != "PRB-2" {
		t.Fatalf("expected problem code PRB-2, got %s", prob2.ProblemCode)
	}

	// 7. Get by ID
	fetched, err := svc.GetProblem(prob1.ID)
	if err != nil {
		t.Fatalf("get problem by id failed: %v", err)
	}
	if fetched.Title != prob1.Title {
		t.Fatalf("expected title %s, got %s", prob1.Title, fetched.Title)
	}

	// 8. Get by Code
	fetchedByCode, err := svc.GetProblemByCode(1, "PRB-2")
	if err != nil {
		t.Fatalf("get problem by code failed: %v", err)
	}
	if fetchedByCode.ID != prob2.ID {
		t.Fatalf("expected id %d, got %d", prob2.ID, fetchedByCode.ID)
	}

	// 9. Non-existent problem
	_, err = svc.GetProblem(9999)
	if !errors.Is(err, ErrProblemNotFound) {
		t.Fatalf("expected ErrProblemNotFound, got %v", err)
	}
}

func TestProblemService_ListAndFilter(t *testing.T) {
	svc, _, _ := setupTestProblemSvc(t)

	// Create 3 problems
	_, _ = svc.CreateProblem("user_1", model.CreateProblemRequest{
		WorkspaceID: 1,
		Title:       "Postgres Connection Saturation",
		Category:    model.ProblemCategoryDatabase,
		Priority:    model.ProblemPriorityCritical,
	})
	_, _ = svc.CreateProblem("user_1", model.CreateProblemRequest{
		WorkspaceID: 1,
		Title:       "Kafka Consumer Lag in Notification Service",
		Category:    model.ProblemCategorySoftware,
		Priority:    model.ProblemPriorityMedium,
	})
	_, _ = svc.CreateProblem("user_1", model.CreateProblemRequest{
		WorkspaceID: 1,
		Title:       "BGP Route Flapping in US-East Region",
		Category:    model.ProblemCategoryNetwork,
		Priority:    model.ProblemPriorityHigh,
	})

	// 1. List all
	list, total, err := svc.ListProblems(model.ProblemFilter{WorkspaceID: 1})
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	if total != 3 || len(list) != 3 {
		t.Fatalf("expected 3 problems, got total=%d len=%d", total, len(list))
	}

	// 2. Filter by Category
	dbList, total, err := svc.ListProblems(model.ProblemFilter{
		WorkspaceID: 1,
		Category:    model.ProblemCategoryDatabase,
	})
	if err != nil {
		t.Fatalf("filter category failed: %v", err)
	}
	if total != 1 || len(dbList) != 1 || dbList[0].Category != model.ProblemCategoryDatabase {
		t.Fatalf("expected 1 database problem, got %d", total)
	}

	// 3. Search
	searchList, total, err := svc.ListProblems(model.ProblemFilter{
		WorkspaceID: 1,
		Search:      "Kafka",
	})
	if err != nil {
		t.Fatalf("search failed: %v", err)
	}
	if total != 1 || len(searchList) != 1 || searchList[0].Title != "Kafka Consumer Lag in Notification Service" {
		t.Fatalf("expected 1 match for Kafka, got %d", total)
	}
}

func TestProblemService_UpdateAndDelete(t *testing.T) {
	svc, _, _ := setupTestProblemSvc(t)

	prob, err := svc.CreateProblem("user_1", model.CreateProblemRequest{
		WorkspaceID: 1,
		Title:       "Initial Problem Title",
		Category:    model.ProblemCategorySoftware,
		Priority:    model.ProblemPriorityLow,
	})
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}

	// Update fields
	newTitle := "Updated Critical Software Problem"
	newPri := model.ProblemPriorityCritical
	rootCause := "Null pointer dereference in auth handler"
	updated, err := svc.UpdateProblem(prob.ID, "user_1", model.UpdateProblemRequest{
		Title:            &newTitle,
		Priority:         &newPri,
		RootCauseSummary: &rootCause,
	})
	if err != nil {
		t.Fatalf("update failed: %v", err)
	}
	if updated.Title != newTitle || updated.Priority != newPri || updated.RootCauseSummary != rootCause {
		t.Fatalf("updated attributes mismatch: %+v", updated)
	}

	// Delete
	if err := svc.DeleteProblem(prob.ID, "user_1"); err != nil {
		t.Fatalf("delete failed: %v", err)
	}
	_, err = svc.GetProblem(prob.ID)
	if !errors.Is(err, ErrProblemNotFound) {
		t.Fatalf("expected ErrProblemNotFound after deletion, got %v", err)
	}
}

func TestProblemService_LifecycleTransitions(t *testing.T) {
	svc, _, _ := setupTestProblemSvc(t)

	prob, err := svc.CreateProblem("user_1", model.CreateProblemRequest{
		WorkspaceID: 1,
		Title:       "Memory Thrashing Under Peak Load",
		Category:    model.ProblemCategoryInfrastructure,
	})
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}

	// 1. Illegal transition: logged -> resolved directly
	_, err = svc.UpdateProblemStatus(prob.ID, "user_1", model.UpdateProblemStatusRequest{
		Status: model.ProblemStatusResolved,
	})
	if !errors.Is(err, ErrInvalidProblemTransition) {
		t.Fatalf("expected ErrInvalidProblemTransition from logged to resolved, got %v", err)
	}

	// 2. Transition logged -> investigating
	prob, err = svc.UpdateProblemStatus(prob.ID, "user_1", model.UpdateProblemStatusRequest{
		Status: model.ProblemStatusInvestigating,
		Reason: "Team assigned to reproduce leak",
	})
	if err != nil {
		t.Fatalf("transition to investigating failed: %v", err)
	}
	if prob.Status != model.ProblemStatusInvestigating {
		t.Fatalf("expected investigating, got %s", prob.Status)
	}

	// 3. Transition investigating -> workaround_found
	prob, err = svc.UpdateProblemStatus(prob.ID, "user_1", model.UpdateProblemStatusRequest{
		Status:     model.ProblemStatusWorkaroundFound,
		Workaround: "Restart worker process every 4 hours via cron",
	})
	if err != nil {
		t.Fatalf("transition to workaround_found failed: %v", err)
	}
	if prob.Workaround != "Restart worker process every 4 hours via cron" {
		t.Fatalf("workaround not persisted: %s", prob.Workaround)
	}

	// 4. Transition workaround_found -> known_error
	prob, err = svc.UpdateProblemStatus(prob.ID, "user_1", model.UpdateProblemStatusRequest{
		Status: model.ProblemStatusKnownError,
		Reason: "Workaround published for operational awareness",
	})
	if err != nil {
		t.Fatalf("transition to known_error failed: %v", err)
	}
	if !prob.IsKnownError || prob.KnownErrorPublishedAt == nil {
		t.Fatalf("expected IsKnownError true and KnownErrorPublishedAt set")
	}

	// 5. Transition known_error -> resolved
	prob, err = svc.UpdateProblemStatus(prob.ID, "user_1", model.UpdateProblemStatusRequest{
		Status:            model.ProblemStatusResolved,
		PermanentSolution: "Patched memory leak in v2.4.1 release",
	})
	if err != nil {
		t.Fatalf("transition to resolved failed: %v", err)
	}
	if prob.ResolvedAt == nil {
		t.Fatalf("expected ResolvedAt timestamp set")
	}

	// 6. Transition resolved -> closed
	prob, err = svc.UpdateProblemStatus(prob.ID, "user_1", model.UpdateProblemStatusRequest{
		Status: model.ProblemStatusClosed,
		Reason: "Production verified stable for 7 days post-fix",
	})
	if err != nil {
		t.Fatalf("transition to closed failed: %v", err)
	}
	if prob.ClosedAt == nil {
		t.Fatalf("expected ClosedAt timestamp set")
	}

	// 7. Reopen from closed to investigating
	prob, err = svc.UpdateProblemStatus(prob.ID, "user_1", model.UpdateProblemStatusRequest{
		Status: model.ProblemStatusInvestigating,
		Reason: "Leak re-emerged on secondary cluster",
	})
	if err != nil {
		t.Fatalf("reopening failed: %v", err)
	}
	if prob.ResolvedAt != nil || prob.ClosedAt != nil {
		t.Fatalf("expected ResolvedAt and ClosedAt cleared upon reopening")
	}
}

func TestProblemService_Associations(t *testing.T) {
	svc, _, _ := setupTestProblemSvc(t)

	prob, err := svc.CreateProblem("user_1", model.CreateProblemRequest{
		WorkspaceID: 1,
		Title:       "Postgres Disk Full Crisis",
		Category:    model.ProblemCategoryDatabase,
	})
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}

	// 1. Link Incident
	incLink, err := svc.LinkIncident(prob.ID, "user_1", model.LinkIncidentRequest{
		IncidentID: 301,
		Notes:      "Triggering outage incident",
	})
	if err != nil {
		t.Fatalf("link incident failed: %v", err)
	}
	if incLink.IncidentNumber != 1 {
		t.Fatalf("expected incident number 1, got %d", incLink.IncidentNumber)
	}

	// Duplicate link fails
	_, err = svc.LinkIncident(prob.ID, "user_1", model.LinkIncidentRequest{
		IncidentID: 301,
	})
	if !errors.Is(err, ErrProblemLinkExists) {
		t.Fatalf("expected ErrProblemLinkExists, got %v", err)
	}

	// List Incidents
	incList, err := svc.ListProblemIncidents(prob.ID)
	if err != nil || len(incList) != 1 {
		t.Fatalf("expected 1 linked incident, got %d", len(incList))
	}

	// Unlink Incident
	if err := svc.UnlinkIncident(prob.ID, 301, "user_1"); err != nil {
		t.Fatalf("unlink incident failed: %v", err)
	}
	incList, _ = svc.ListProblemIncidents(prob.ID)
	if len(incList) != 0 {
		t.Fatalf("expected 0 linked incidents after unlink, got %d", len(incList))
	}

	// 2. Link CI
	ciLink, err := svc.LinkCI(prob.ID, "user_1", model.LinkCIRequest{
		CIID:         201,
		Relationship: model.ProblemCIRelRootCause,
		Notes:        "Root cause postgres cluster",
	})
	if err != nil {
		t.Fatalf("link CI failed: %v", err)
	}
	if ciLink.CICode != "CI-1" {
		t.Fatalf("expected CI-1, got %s", ciLink.CICode)
	}

	// 3. Link Change
	chgLink, err := svc.LinkChange(prob.ID, "user_1", model.LinkChangeRequest{
		ChangeID:     401,
		Relationship: model.ProblemChangeRelPermanentFix,
		Notes:        "Permanent fix deployment change",
	})
	if err != nil {
		t.Fatalf("link Change failed: %v", err)
	}
	if chgLink.ChangeCode != "CHG-1" {
		t.Fatalf("expected CHG-1, got %s", chgLink.ChangeCode)
	}

	// 4. Link Task
	taskLink, err := svc.LinkTask(prob.ID, "user_1", model.LinkTaskRequest{
		TaskID:       501,
		Relationship: model.ProblemTaskRelInvestigation,
		Notes:        "Deep dive into connection parameters",
	})
	if err != nil {
		t.Fatalf("link Task failed: %v", err)
	}
	if taskLink.TaskTitle != "Tune Connection Pool" {
		t.Fatalf("expected task title 'Tune Connection Pool', got %s", taskLink.TaskTitle)
	}

	// Enriched counts on problem
	p, _ := svc.GetProblem(prob.ID)
	if p.LinkedCIsCount != 1 || p.LinkedChangesCount != 1 || p.LinkedTasksCount != 1 {
		t.Fatalf("expected linked counts 1, 1, 1, got CIs=%d Changes=%d Tasks=%d", p.LinkedCIsCount, p.LinkedChangesCount, p.LinkedTasksCount)
	}
}

func TestProblemService_RCA5Whys(t *testing.T) {
	svc, _, _ := setupTestProblemSvc(t)

	prob, err := svc.CreateProblem("user_1", model.CreateProblemRequest{
		WorkspaceID: 1,
		Title:       "Production Payment Failures",
		Category:    model.ProblemCategorySoftware,
		RCAMethod:   model.RCAMethod5Whys,
	})
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}

	// Add 5 Whys steps
	steps := []struct {
		q, a string
	}{
		{"Why did checkout fail?", "Payment service returned HTTP 504 Gateway Timeout"},
		{"Why did Payment service time out?", "Database queries hung waiting for free connection pool slots"},
		{"Why were pool slots exhausted?", "Connection pool idle timeout was set to 0, holding stale connections"},
		{"Why was idle timeout 0?", "A recent configuration refactor replaced custom settings with default 0"},
		{"Why did the refactor lack validation?", "Automated integration tests did not assert connection pool configuration parameters"},
	}

	for _, s := range steps {
		step, err := svc.AddRCAStep(prob.ID, "user_1", model.CreateRCAStepRequest{
			Question: s.q,
			Answer:   s.a,
		})
		if err != nil {
			t.Fatalf("failed to add RCA step: %v", err)
		}
		if step.StepNumber <= 0 {
			t.Fatalf("expected positive step number, got %d", step.StepNumber)
		}
	}

	// List steps
	rcaList, err := svc.ListRCASteps(prob.ID)
	if err != nil || len(rcaList) != 5 {
		t.Fatalf("expected 5 RCA steps, got %d", len(rcaList))
	}
	if rcaList[4].StepNumber != 5 {
		t.Fatalf("expected 5th step number 5, got %d", rcaList[4].StepNumber)
	}

	// Update step 5 evidence
	evidence := "https://metrics.internal/dashboards/db-pools?spike=2026-11-11"
	updated, err := svc.UpdateRCAStep(rcaList[4].ID, "user_1", model.UpdateRCAStepRequest{
		Evidence: &evidence,
	})
	if err != nil {
		t.Fatalf("update RCA step failed: %v", err)
	}
	if updated.Evidence != evidence {
		t.Fatalf("expected evidence %s, got %s", evidence, updated.Evidence)
	}

	// Delete step 1
	if err := svc.DeleteRCAStep(rcaList[0].ID, "user_1"); err != nil {
		t.Fatalf("delete RCA step failed: %v", err)
	}
	rcaListAfter, _ := svc.ListRCASteps(prob.ID)
	if len(rcaListAfter) != 4 {
		t.Fatalf("expected 4 RCA steps after deletion, got %d", len(rcaListAfter))
	}
}

func TestProblemService_KEDBAndMetrics(t *testing.T) {
	svc, _, _ := setupTestProblemSvc(t)

	prob, err := svc.CreateProblem("user_1", model.CreateProblemRequest{
		WorkspaceID: 1,
		Title:       "Memory Saturation in Go HTTP Microservices",
		Category:    model.ProblemCategorySoftware,
		Priority:    model.ProblemPriorityCritical,
	})
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}

	// 1. Publish to KEDB
	published, err := svc.PublishKnownError(prob.ID, "user_1", model.PublishKnownErrorRequest{
		Workaround:        "Set GOMEMLIMIT=1800MiB and GOGC=80 in deployment manifest",
		PermanentSolution: "Upgrade to Go 1.23 with optimized GC pacing",
		RootCauseSummary:  "Unbounded slice reallocation in JSON serializer during multipart file uploads",
	})
	if err != nil {
		t.Fatalf("publish known error failed: %v", err)
	}
	if !published.IsKnownError {
		t.Fatalf("expected IsKnownError true")
	}
	if published.Status != model.ProblemStatusKnownError {
		t.Fatalf("expected status known_error, got %s", published.Status)
	}

	// 2. Search KEDB
	kedbList, total, err := svc.ListKnownErrors(model.KnownErrorFilter{
		WorkspaceID: 1,
		Search:      "GOMEMLIMIT",
	})
	if err != nil {
		t.Fatalf("search KEDB failed: %v", err)
	}
	if total != 1 || len(kedbList) != 1 {
		t.Fatalf("expected 1 KEDB match, got %d", total)
	}

	// 3. Problem Metrics
	metrics, err := svc.GetProblemMetrics(1)
	if err != nil {
		t.Fatalf("get problem metrics failed: %v", err)
	}
	if metrics.TotalProblems != 1 {
		t.Fatalf("expected 1 total problem, got %d", metrics.TotalProblems)
	}
	if metrics.KnownErrorsCount != 1 {
		t.Fatalf("expected 1 known error, got %d", metrics.KnownErrorsCount)
	}
	if metrics.CategoryCounts[model.ProblemCategorySoftware] != 1 {
		t.Fatalf("expected 1 software problem, got %d", metrics.CategoryCounts[model.ProblemCategorySoftware])
	}
	if metrics.PriorityCounts[model.ProblemPriorityCritical] != 1 {
		t.Fatalf("expected 1 critical problem, got %d", metrics.PriorityCounts[model.ProblemPriorityCritical])
	}

	// 4. Retire from KEDB
	retired, err := svc.RetireKnownError(prob.ID, "user_1")
	if err != nil {
		t.Fatalf("retire known error failed: %v", err)
	}
	if retired.IsKnownError {
		t.Fatalf("expected IsKnownError false after retirement")
	}
}
