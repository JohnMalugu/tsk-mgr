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

func setupTestCMDBSvc(t *testing.T) (CMDBService, repository.CMDBRepository, *sql.DB) {
	sqliteDB, err := db.InitDB(":memory:")
	if err != nil {
		t.Fatalf("failed to initialize test in-memory db: %v", err)
	}

	_, err = sqliteDB.Exec(`
		INSERT INTO users (id, username, email, password_hash, created_at, updated_at) 
		VALUES ('user_1', 'alice', 'alice@example.com', 'hash', datetime('now'), datetime('now')),
		       ('user_2', 'bob', 'bob@example.com', 'hash', datetime('now'), datetime('now'));

		INSERT INTO workspaces (id, name, slug, owner_id, created_at, updated_at) 
		VALUES (1, 'Engineering WS', 'eng-ws-cmdb', 'user_1', datetime('now'), datetime('now'));

		INSERT INTO projects (id, user_id, name, description, color, created_at, updated_at)
		VALUES (101, 'user_1', 'Payments Platform', 'Payment Infra', '#4f46e5', datetime('now'), datetime('now'));

		INSERT INTO tasks (id, user_id, project_id, title, description, status, completed, due_date, estimate_minutes, priority, created_at, updated_at)
		VALUES (501, 'user_1', 101, 'Postgres Cluster Maintenance', 'Upgrade to Postgres 16', 'todo', 0, datetime('now'), 120, 'high', datetime('now'), datetime('now'));
	`)
	if err != nil {
		t.Fatalf("failed to insert test seed data: %v", err)
	}

	cmdbRepo := repository.NewCMDBRepository(sqliteDB)
	workspaceRepo := repository.NewWorkspaceRepository(sqliteDB)
	projectRepo := repository.NewProjectRepository(sqliteDB)
	taskRepo := repository.NewTaskRepository(sqliteDB)
	userRepo := repository.NewUserRepository(sqliteDB)
	auditRepo := repository.NewAuditRepository(sqliteDB)
	auditSvc := NewAuditService(auditRepo, workspaceRepo, userRepo)

	svc := NewCMDBService(cmdbRepo, workspaceRepo, projectRepo, taskRepo, auditSvc)
	return svc, cmdbRepo, sqliteDB
}

func TestCMDBService_CreateAndGet(t *testing.T) {
	svc, _, _ := setupTestCMDBSvc(t)

	// 1. Missing workspace
	_, err := svc.CreateCI("user_1", model.CreateCIRequest{
		WorkspaceID: 0,
		Name:        "Payment Gateway API",
	})
	if !errors.Is(err, ErrWorkspaceRequired) {
		t.Fatalf("expected ErrWorkspaceRequired, got %v", err)
	}

	// 2. Non-existent workspace
	_, err = svc.CreateCI("user_1", model.CreateCIRequest{
		WorkspaceID: 999,
		Name:        "Payment Gateway API",
	})
	if !errors.Is(err, ErrWorkspaceNotFound) {
		t.Fatalf("expected ErrWorkspaceNotFound, got %v", err)
	}

	// 3. Missing name
	_, err = svc.CreateCI("user_1", model.CreateCIRequest{
		WorkspaceID: 1,
		Name:        "   ",
	})
	if !errors.Is(err, ErrCINameRequired) {
		t.Fatalf("expected ErrCINameRequired, got %v", err)
	}

	// 4. Invalid CI type
	_, err = svc.CreateCI("user_1", model.CreateCIRequest{
		WorkspaceID: 1,
		Name:        "Payment Gateway API",
		CIType:      "virtual_device",
	})
	if !errors.Is(err, ErrInvalidCIType) {
		t.Fatalf("expected ErrInvalidCIType, got %v", err)
	}

	// 5. Invalid Criticality
	_, err = svc.CreateCI("user_1", model.CreateCIRequest{
		WorkspaceID: 1,
		Name:        "Payment Gateway API",
		Criticality: "super_critical",
	})
	if !errors.Is(err, ErrInvalidCICriticality) {
		t.Fatalf("expected ErrInvalidCICriticality, got %v", err)
	}

	// 6. Successful creation
	projID := int64(101)
	ci, err := svc.CreateCI("user_1", model.CreateCIRequest{
		WorkspaceID:  1,
		ProjectID:    &projID,
		Name:         "Payment Gateway API",
		CIType:       model.CITypeService,
		Criticality:  model.CICriticalityTier1,
		Environment:  model.CIEnvProduction,
		IPAddress:    "10.0.1.50",
		Hostname:     "pay-api-prod.internal",
		Version:      "v2.4.1",
		CostPerMonth: 250.00,
		Attributes:   `{"runtime":"go","replicas":4}`,
	})
	if err != nil {
		t.Fatalf("failed to create CI: %v", err)
	}
	if ci.ID <= 0 {
		t.Fatalf("expected positive ID, got %d", ci.ID)
	}
	if ci.CICode != "CI-1" {
		t.Fatalf("expected CI code CI-1, got %s", ci.CICode)
	}
	if ci.Status != model.CIStatusActive {
		t.Fatalf("expected default status active, got %s", ci.Status)
	}
	if ci.OwnerName != "alice" {
		t.Fatalf("expected owner alice, got %s", ci.OwnerName)
	}
	if ci.ProjectName != "Payments Platform" {
		t.Fatalf("expected project Payments Platform, got %s", ci.ProjectName)
	}

	// 7. Get by ID
	fetched, err := svc.GetCI(ci.ID)
	if err != nil {
		t.Fatalf("failed to get CI by ID: %v", err)
	}
	if fetched.Name != ci.Name {
		t.Fatalf("expected name %s, got %s", ci.Name, fetched.Name)
	}

	// 8. Get by Code
	byCode, err := svc.GetCIByCode(1, "CI-1")
	if err != nil {
		t.Fatalf("failed to get CI by code: %v", err)
	}
	if byCode.ID != ci.ID {
		t.Fatalf("expected ID %d, got %d", ci.ID, byCode.ID)
	}
}

func TestCMDBService_ListAndFilter(t *testing.T) {
	svc, _, _ := setupTestCMDBSvc(t)

	_, _ = svc.CreateCI("user_1", model.CreateCIRequest{
		WorkspaceID: 1,
		Name:        "Primary Postgres DB",
		CIType:      model.CITypeDatabase,
		Criticality: model.CICriticalityTier1,
	})
	_, _ = svc.CreateCI("user_1", model.CreateCIRequest{
		WorkspaceID: 1,
		Name:        "Redis Cache Cluster",
		CIType:      model.CITypeCluster,
		Criticality: model.CICriticalityTier2,
	})

	// 1. List all
	list, err := svc.ListCIs(model.CIFilter{WorkspaceID: 1})
	if err != nil {
		t.Fatalf("list CIs failed: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("expected 2 CIs, got %d", len(list))
	}

	// 2. Filter by CI type
	dbType := model.CITypeDatabase
	dbList, err := svc.ListCIs(model.CIFilter{WorkspaceID: 1, CIType: &dbType})
	if err != nil {
		t.Fatalf("filter by db failed: %v", err)
	}
	if len(dbList) != 1 || dbList[0].CIType != model.CITypeDatabase {
		t.Fatalf("expected 1 database CI, got %d", len(dbList))
	}

	// 3. Search query
	q := "Redis"
	searchList, err := svc.ListCIs(model.CIFilter{WorkspaceID: 1, Search: &q})
	if err != nil {
		t.Fatalf("search failed: %v", err)
	}
	if len(searchList) != 1 {
		t.Fatalf("expected 1 search result, got %d", len(searchList))
	}
}

func TestCMDBService_UpdateAndDelete(t *testing.T) {
	svc, _, _ := setupTestCMDBSvc(t)

	ci, err := svc.CreateCI("user_1", model.CreateCIRequest{
		WorkspaceID: 1,
		Name:        "Auth Microservice",
		CIType:      model.CITypeService,
	})
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}

	// 1. Update metadata
	newName := "Authentication Core Service"
	newVer := "v3.0.0"
	newCost := 180.50
	updated, err := svc.UpdateCI(ci.ID, "user_1", model.UpdateCIRequest{
		Name:         &newName,
		Version:      &newVer,
		CostPerMonth: &newCost,
	})
	if err != nil {
		t.Fatalf("update failed: %v", err)
	}
	if updated.Name != newName || updated.Version != newVer || updated.CostPerMonth != newCost {
		t.Fatalf("updated values mismatch: %+v", updated)
	}

	// 2. Update status
	stUpdated, err := svc.UpdateCIStatus(ci.ID, "user_1", model.UpdateCIStatusRequest{
		Status: model.CIStatusMaintenance,
		Reason: "Scheduled kernel upgrade",
	})
	if err != nil {
		t.Fatalf("update status failed: %v", err)
	}
	if stUpdated.Status != model.CIStatusMaintenance {
		t.Fatalf("expected status maintenance, got %s", stUpdated.Status)
	}

	// 3. Delete CI
	if err := svc.DeleteCI(ci.ID, "user_1"); err != nil {
		t.Fatalf("delete failed: %v", err)
	}
	_, err = svc.GetCI(ci.ID)
	if !errors.Is(err, ErrCINotFound) {
		t.Fatalf("expected ErrCINotFound after deletion, got %v", err)
	}
}

func TestCMDBService_RelationshipsAndGraph(t *testing.T) {
	svc, _, _ := setupTestCMDBSvc(t)

	// Web App -> API Gateway -> Auth DB
	webApp, err := svc.CreateCI("user_1", model.CreateCIRequest{
		WorkspaceID: 1,
		Name:        "Customer Web Portal",
		CIType:      model.CITypeApplication,
		Criticality: model.CICriticalityTier2,
	})
	if err != nil {
		t.Fatalf("create webApp failed: %v", err)
	}
	apiGw, err := svc.CreateCI("user_1", model.CreateCIRequest{
		WorkspaceID: 1,
		Name:        "Public API Gateway",
		CIType:      model.CITypeService,
		Criticality: model.CICriticalityTier1,
	})
	if err != nil {
		t.Fatalf("create apiGw failed: %v", err)
	}
	authDB, err := svc.CreateCI("user_1", model.CreateCIRequest{
		WorkspaceID: 1,
		Name:        "User Accounts Database",
		CIType:      model.CITypeDatabase,
		Criticality: model.CICriticalityTier1,
	})
	if err != nil {
		t.Fatalf("create authDB failed: %v", err)
	}

	// 1. Self relationship fails
	_, err = svc.CreateRelationship(webApp.ID, model.CreateCIRelationshipRequest{
		TargetCIID: webApp.ID,
	})
	if !errors.Is(err, ErrSelfRelationship) {
		t.Fatalf("expected ErrSelfRelationship, got %v", err)
	}

	// 2. Web App depends_on API Gateway
	rel1, err := svc.CreateRelationship(webApp.ID, model.CreateCIRelationshipRequest{
		TargetCIID:   apiGw.ID,
		RelationType: model.CIRelationDependsOn,
		Description:  "Routes all user requests",
	})
	if err != nil {
		t.Fatalf("create relationship 1 failed: %v", err)
	}
	if rel1.SourceCIID != webApp.ID || rel1.TargetCIID != apiGw.ID {
		t.Fatalf("expected source %d and target %d, got %+v", webApp.ID, apiGw.ID, rel1)
	}

	// 3. Duplicate relationship fails
	_, err = svc.CreateRelationship(webApp.ID, model.CreateCIRelationshipRequest{
		TargetCIID:   apiGw.ID,
		RelationType: model.CIRelationDependsOn,
	})
	if !errors.Is(err, ErrRelationshipExists) {
		t.Fatalf("expected ErrRelationshipExists, got %v", err)
	}

	// 4. API Gateway depends_on Auth DB
	_, err = svc.CreateRelationship(apiGw.ID, model.CreateCIRelationshipRequest{
		TargetCIID:   authDB.ID,
		RelationType: model.CIRelationDependsOn,
	})
	if err != nil {
		t.Fatalf("create relationship 2 failed: %v", err)
	}

	// 5. Dependency Graph from Web App
	graph, err := svc.GetDependencyGraph(webApp.ID, 5)
	if err != nil {
		t.Fatalf("get dependency graph failed: %v", err)
	}
	if graph.RootCI.ID != webApp.ID {
		t.Fatalf("expected root CI %d, got %d", webApp.ID, graph.RootCI.ID)
	}
	if len(graph.Upstream) != 2 {
		t.Fatalf("expected 2 upstream dependencies, got %d", len(graph.Upstream))
	}

	// 6. Blast Radius from Auth DB (downstream impact)
	blast, err := svc.AnalyzeBlastRadius(authDB.ID)
	if err != nil {
		t.Fatalf("analyze blast radius failed: %v", err)
	}
	if blast.DirectImpactCount != 1 || blast.IndirectImpactCount != 1 {
		t.Fatalf("expected direct=1 indirect=1, got direct=%d indirect=%d", blast.DirectImpactCount, blast.IndirectImpactCount)
	}
	if !blast.HasTier1Impact {
		t.Fatalf("expected HasTier1Impact true because API Gateway is tier 1")
	}

	// 7. Delete relationship
	if err := svc.DeleteRelationship(rel1.ID); err != nil {
		t.Fatalf("delete relationship failed: %v", err)
	}
	rels, _ := svc.ListRelationships(webApp.ID)
	if len(rels) != 0 {
		t.Fatalf("expected 0 relationships, got %d", len(rels))
	}
}

func TestCMDBService_EntityLinkage(t *testing.T) {
	svc, _, _ := setupTestCMDBSvc(t)

	ci, _ := svc.CreateCI("user_1", model.CreateCIRequest{
		WorkspaceID: 1,
		Name:        "Database Replica",
		CIType:      model.CITypeDatabase,
	})

	// 1. Link to task 501
	link, err := svc.LinkEntity(ci.ID, model.CreateCIEntityLinkRequest{
		EntityType:  model.CILinkEntityTask,
		EntityID:    501,
		ImpactLevel: model.CIImpactDirect,
		Notes:       "Task upgrades this replica",
	})
	if err != nil {
		t.Fatalf("link entity failed: %v", err)
	}
	if link.CIID != ci.ID || link.EntityID != 501 {
		t.Fatalf("expected link to task 501, got %+v", link)
	}

	// 2. Duplicate link fails
	_, err = svc.LinkEntity(ci.ID, model.CreateCIEntityLinkRequest{
		EntityType: model.CILinkEntityTask,
		EntityID:   501,
	})
	if !errors.Is(err, ErrLinkAlreadyExists) {
		t.Fatalf("expected ErrLinkAlreadyExists, got %v", err)
	}

	// 3. List links by CI
	links, err := svc.ListEntityLinks(ci.ID)
	if err != nil || len(links) != 1 {
		t.Fatalf("expected 1 link, got %d (err: %v)", len(links), err)
	}

	// 4. List CIs for entity
	ciLinks, err := svc.ListCIsForEntity(model.CILinkEntityTask, 501)
	if err != nil || len(ciLinks) != 1 {
		t.Fatalf("expected 1 CI link for task 501, got %d (err: %v)", len(ciLinks), err)
	}

	// 5. Unlink
	if err := svc.UnlinkEntity(link.ID); err != nil {
		t.Fatalf("unlink failed: %v", err)
	}
	afterUnlink, _ := svc.ListEntityLinks(ci.ID)
	if len(afterUnlink) != 0 {
		t.Fatalf("expected 0 links after unlink, got %d", len(afterUnlink))
	}
}

func TestCMDBService_MaintenanceSchedules(t *testing.T) {
	svc, _, _ := setupTestCMDBSvc(t)

	ci, _ := svc.CreateCI("user_1", model.CreateCIRequest{
		WorkspaceID: 1,
		Name:        "Kubernetes Worker Pool",
		CIType:      model.CITypeCluster,
	})

	// 1. Invalid date range (end before start)
	_, err := svc.CreateMaintenance(ci.ID, "user_1", model.CreateCIMaintenanceRequest{
		Title:         "OS Security Patching",
		WindowStartAt: "2026-11-20T04:00:00Z",
		WindowEndAt:   "2026-11-20T02:00:00Z",
	})
	if !errors.Is(err, ErrInvalidMaintDates) {
		t.Fatalf("expected ErrInvalidMaintDates, got %v", err)
	}

	// 2. Valid schedule
	maint, err := svc.CreateMaintenance(ci.ID, "user_1", model.CreateCIMaintenanceRequest{
		Title:                   "OS Security Patching",
		Description:             "Rolling node reboot",
		WindowStartAt:           "2026-11-20T02:00:00Z",
		WindowEndAt:             "2026-11-20T04:00:00Z",
		ExpectedDowntimeMinutes: 15,
	})
	if err != nil {
		t.Fatalf("create maintenance failed: %v", err)
	}
	if maint.Status != model.CIMaintStatusScheduled {
		t.Fatalf("expected status scheduled, got %s", maint.Status)
	}

	// 3. Update maintenance status
	updatedMaint, err := svc.UpdateMaintenanceStatus(maint.ID, model.UpdateCIMaintenanceStatusRequest{
		Status: model.CIMaintStatusCompleted,
	})
	if err != nil {
		t.Fatalf("update maintenance status failed: %v", err)
	}
	if updatedMaint.Status != model.CIMaintStatusCompleted {
		t.Fatalf("expected status completed, got %s", updatedMaint.Status)
	}

	// 4. List maintenances
	list, err := svc.ListMaintenances(ci.ID)
	if err != nil || len(list) != 1 {
		t.Fatalf("expected 1 maintenance, got %d (err: %v)", len(list), err)
	}

	// 5. Delete maintenance
	if err := svc.DeleteMaintenance(maint.ID); err != nil {
		t.Fatalf("delete maintenance failed: %v", err)
	}
	listAfter, _ := svc.ListMaintenances(ci.ID)
	if len(listAfter) != 0 {
		t.Fatalf("expected 0 maintenances after delete, got %d", len(listAfter))
	}
}

func TestCMDBService_Metrics(t *testing.T) {
	svc, _, _ := setupTestCMDBSvc(t)

	// Create 2 active CIs and 1 in maintenance
	c1, _ := svc.CreateCI("user_1", model.CreateCIRequest{
		WorkspaceID:  1,
		Name:         "Auth Service",
		CIType:       model.CITypeService,
		Criticality:  model.CICriticalityTier1,
		CostPerMonth: 100.0,
	})
	c2, _ := svc.CreateCI("user_1", model.CreateCIRequest{
		WorkspaceID:  1,
		Name:         "Postgres Primary",
		CIType:       model.CITypeDatabase,
		Criticality:  model.CICriticalityTier1,
		CostPerMonth: 300.0,
	})
	_, _ = svc.CreateRelationship(c1.ID, model.CreateCIRelationshipRequest{
		TargetCIID:   c2.ID,
		RelationType: model.CIRelationDependsOn,
	})

	c3, _ := svc.CreateCI("user_1", model.CreateCIRequest{
		WorkspaceID:  1,
		Name:         "Batch Worker",
		CIType:       model.CITypeServer,
		Criticality:  model.CICriticalityTier3,
		CostPerMonth: 50.0,
	})
	_, _ = svc.UpdateCIStatus(c3.ID, "user_1", model.UpdateCIStatusRequest{
		Status: model.CIStatusMaintenance,
	})

	now := time.Now().UTC()
	startStr := now.Add(1 * time.Hour).Format(time.RFC3339)
	endStr := now.Add(2 * time.Hour).Format(time.RFC3339)
	_, _ = svc.CreateMaintenance(c1.ID, "user_1", model.CreateCIMaintenanceRequest{
		Title:         "Minor Patch",
		WindowStartAt: startStr,
		WindowEndAt:   endStr,
	})

	metrics, err := svc.GetCMDBMetrics(1)
	if err != nil {
		t.Fatalf("get CMDB metrics failed: %v", err)
	}
	if metrics.TotalCIs != 3 {
		t.Fatalf("expected 3 total CIs, got %d", metrics.TotalCIs)
	}
	if metrics.ActiveCount != 2 {
		t.Fatalf("expected 2 active CIs, got %d", metrics.ActiveCount)
	}
	if metrics.MaintenanceCount != 1 {
		t.Fatalf("expected 1 maintenance CI, got %d", metrics.MaintenanceCount)
	}
	if metrics.ServiceCount != 1 || metrics.DatabaseCount != 1 || metrics.ServerCount != 1 {
		t.Fatalf("type counts mismatch: service=%d db=%d server=%d", metrics.ServiceCount, metrics.DatabaseCount, metrics.ServerCount)
	}
	if metrics.Tier1Count != 2 {
		t.Fatalf("expected 2 tier 1 CIs, got %d", metrics.Tier1Count)
	}
	if metrics.TotalRelationships != 1 {
		t.Fatalf("expected 1 relationship, got %d", metrics.TotalRelationships)
	}
	if metrics.ScheduledMaintenances != 1 {
		t.Fatalf("expected 1 scheduled maintenance, got %d", metrics.ScheduledMaintenances)
	}
	if metrics.TotalMonthlyCost != 450.0 {
		t.Fatalf("expected monthly cost 450.0, got %f", metrics.TotalMonthlyCost)
	}
}
