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

func setupTestStatusPageSvc(t *testing.T) (StatusPageService, repository.StatusPageRepository, *sql.DB) {
	sqliteDB, err := db.InitDB(":memory:")
	if err != nil {
		t.Fatalf("failed to initialize test in-memory db: %v", err)
	}

	_, err = sqliteDB.Exec(`
		INSERT INTO users (id, username, email, password_hash, created_at, updated_at) 
		VALUES ('user_1', 'alice', 'alice@example.com', 'hash', datetime('now'), datetime('now')),
		       ('user_2', 'bob', 'bob@example.com', 'hash', datetime('now'), datetime('now'));

		INSERT INTO workspaces (id, name, slug, owner_id, created_at, updated_at) 
		VALUES (1, 'Engineering WS', 'eng-ws-status', 'user_1', datetime('now'), datetime('now'));

		INSERT INTO projects (id, user_id, name, description, color, created_at, updated_at)
		VALUES (101, 'user_1', 'Core Infrastructure', 'Cloud Infrastructure', '#4f46e5', datetime('now'), datetime('now'));
	`)
	if err != nil {
		t.Fatalf("failed to insert test seed data: %v", err)
	}

	statusPageRepo := repository.NewStatusPageRepository(sqliteDB)
	projectRepo := repository.NewProjectRepository(sqliteDB)
	cmdbRepo := repository.NewCMDBRepository(sqliteDB)
	incidentRepo := repository.NewIncidentRepository(sqliteDB)

	svc := NewStatusPageService(statusPageRepo, projectRepo, cmdbRepo, incidentRepo)
	return svc, statusPageRepo, sqliteDB
}

func TestStatusPageService_CreateAndGetPage(t *testing.T) {
	svc, _, _ := setupTestStatusPageSvc(t)

	// 1. Missing name validation
	_, err := svc.CreatePage("user_1", model.CreateStatusPageRequest{
		WorkspaceID: 1,
		Name:        "",
	})
	if !errors.Is(err, ErrStatusPageNameRequired) {
		t.Fatalf("expected ErrStatusPageNameRequired, got %v", err)
	}

	// 2. Successful creation with custom slug
	projID := int64(101)
	page, err := svc.CreatePage("user_1", model.CreateStatusPageRequest{
		WorkspaceID:  1,
		ProjectID:    &projID,
		Name:         "Acme Platform Status",
		Slug:         "acme-status",
		Description:  "Real-time system health and service availability",
		Headline:     "All Systems Normal",
		Visibility:   model.VisibilityPublic,
		CustomDomain: "status.acme.example.com",
		Theme:        model.ThemeDark,
	})
	if err != nil {
		t.Fatalf("unexpected error creating status page: %v", err)
	}
	if page.ID <= 0 {
		t.Fatalf("expected positive page ID, got %d", page.ID)
	}
	if page.Slug != "acme-status" {
		t.Fatalf("expected slug acme-status, got %s", page.Slug)
	}
	if page.CurrentStatus != model.StatusOperational {
		t.Fatalf("expected default operational status, got %s", page.CurrentStatus)
	}

	// 3. Duplicate slug validation
	_, err = svc.CreatePage("user_1", model.CreateStatusPageRequest{
		WorkspaceID: 1,
		Name:        "Duplicate Acme",
		Slug:        "acme-status",
	})
	if !errors.Is(err, ErrDuplicateStatusPageSlug) {
		t.Fatalf("expected ErrDuplicateStatusPageSlug, got %v", err)
	}

	// 4. Get by ID
	fetched, err := svc.GetPage(page.ID)
	if err != nil {
		t.Fatalf("unexpected error fetching page by ID: %v", err)
	}
	if fetched.Name != "Acme Platform Status" {
		t.Fatalf("expected Acme Platform Status, got %s", fetched.Name)
	}

	// 5. Get by Slug
	fetchedBySlug, err := svc.GetPageBySlug("acme-status")
	if err != nil {
		t.Fatalf("unexpected error fetching page by slug: %v", err)
	}
	if fetchedBySlug.ID != page.ID {
		t.Fatalf("expected ID %d, got %d", page.ID, fetchedBySlug.ID)
	}

	// 6. List pages
	pages, total, err := svc.ListPages(model.StatusPageFilter{
		WorkspaceID: 1,
	})
	if err != nil {
		t.Fatalf("unexpected error listing pages: %v", err)
	}
	if total != 1 || len(pages) != 1 {
		t.Fatalf("expected 1 page, got %d total, %d returned", total, len(pages))
	}

	// 7. Update page
	newName := "Acme Cloud Status"
	newTheme := model.ThemeLight
	updated, err := svc.UpdatePage(page.ID, model.UpdateStatusPageRequest{
		Name:  &newName,
		Theme: &newTheme,
	})
	if err != nil {
		t.Fatalf("unexpected error updating page: %v", err)
	}
	if updated.Name != "Acme Cloud Status" || updated.Theme != model.ThemeLight {
		t.Fatalf("expected updated attributes, got name: %s, theme: %s", updated.Name, updated.Theme)
	}

	// 8. Delete page
	if err := svc.DeletePage(page.ID); err != nil {
		t.Fatalf("unexpected error deleting page: %v", err)
	}

	// 9. Verify not found after deletion
	_, err = svc.GetPage(page.ID)
	if !errors.Is(err, ErrStatusPageNotFound) {
		t.Fatalf("expected ErrStatusPageNotFound, got %v", err)
	}
}

func TestStatusPageService_ComponentGroupsAndComponents(t *testing.T) {
	svc, _, _ := setupTestStatusPageSvc(t)

	page, err := svc.CreatePage("user_1", model.CreateStatusPageRequest{
		WorkspaceID: 1,
		Name:        "SaaS Health",
		Slug:        "saas-health",
	})
	if err != nil {
		t.Fatalf("failed to create page: %v", err)
	}

	// 1. Create component groups
	group1, err := svc.CreateComponentGroup(page.ID, model.CreateComponentGroupRequest{
		Name:         "Core APIs",
		Description:  "Critical API Gateways and microservices",
		DisplayOrder: 1,
	})
	if err != nil {
		t.Fatalf("unexpected error creating group: %v", err)
	}

	group2, err := svc.CreateComponentGroup(page.ID, model.CreateComponentGroupRequest{
		Name:         "Databases & Storage",
		Description:  "Database replicas and object storage",
		DisplayOrder: 2,
	})
	if err != nil {
		t.Fatalf("unexpected error creating group 2: %v", err)
	}

	// 2. Create components in group 1
	comp1, err := svc.CreateComponent(page.ID, model.CreateComponentRequest{
		GroupID:      &group1.ID,
		Name:         "Auth & Identity Service",
		Description:  "OIDC and session management",
		Status:       model.StatusOperational,
		DisplayOrder: 1,
		IsMonitored:  true,
	})
	if err != nil {
		t.Fatalf("unexpected error creating component 1: %v", err)
	}

	comp2, err := svc.CreateComponent(page.ID, model.CreateComponentRequest{
		GroupID:      &group1.ID,
		Name:         "Billing API",
		Description:  "Subscription and payments processor",
		Status:       model.StatusOperational,
		DisplayOrder: 2,
		IsMonitored:  true,
	})
	if err != nil {
		t.Fatalf("unexpected error creating component 2: %v", err)
	}

	// 3. Create component in group 2
	comp3, err := svc.CreateComponent(page.ID, model.CreateComponentRequest{
		GroupID:      &group2.ID,
		Name:         "Postgres Primary Cluster",
		Description:  "Transactional relational database",
		Status:       model.StatusOperational,
		DisplayOrder: 1,
		IsMonitored:  true,
	})
	if err != nil {
		t.Fatalf("unexpected error creating component 3: %v", err)
	}

	// 4. List component groups with nested components
	groups, err := svc.ListComponentGroups(page.ID)
	if err != nil {
		t.Fatalf("unexpected error listing groups: %v", err)
	}
	if len(groups) != 2 {
		t.Fatalf("expected 2 groups, got %d", len(groups))
	}
	if groups[0].ComponentsCount != 2 {
		t.Fatalf("expected group 1 to have 2 components, got %d", groups[0].ComponentsCount)
	}
	if groups[1].ComponentsCount != 1 {
		t.Fatalf("expected group 2 to have 1 component, got %d", groups[1].ComponentsCount)
	}

	// 5. Test status roll-up: degrade comp1 -> page becomes degraded
	err = svc.SetComponentStatus(comp1.ID, model.StatusDegradedPerformance)
	if err != nil {
		t.Fatalf("failed to update component status: %v", err)
	}
	p, err := svc.GetPage(page.ID)
	if err != nil {
		t.Fatalf("failed to get page: %v", err)
	}
	if p.CurrentStatus != model.StatusDegradedPerformance {
		t.Fatalf("expected page status degraded_performance, got %s", p.CurrentStatus)
	}

	// 6. Test status roll-up: comp2 has major outage -> page becomes major_outage
	err = svc.SetComponentStatus(comp2.ID, model.StatusMajorOutage)
	if err != nil {
		t.Fatalf("failed to update component status: %v", err)
	}
	p, _ = svc.GetPage(page.ID)
	if p.CurrentStatus != model.StatusMajorOutage {
		t.Fatalf("expected page status major_outage, got %s", p.CurrentStatus)
	}

	// 7. Recover comp2 to operational, comp1 to operational -> page returns to operational
	_ = svc.SetComponentStatus(comp1.ID, model.StatusOperational)
	_ = svc.SetComponentStatus(comp2.ID, model.StatusOperational)
	p, _ = svc.GetPage(page.ID)
	if p.CurrentStatus != model.StatusOperational {
		t.Fatalf("expected page status operational, got %s", p.CurrentStatus)
	}

	// 8. Delete a component
	if err := svc.DeleteComponent(comp3.ID); err != nil {
		t.Fatalf("unexpected error deleting component: %v", err)
	}
	_, err = svc.GetComponent(comp3.ID)
	if !errors.Is(err, ErrComponentNotFound) {
		t.Fatalf("expected ErrComponentNotFound, got %v", err)
	}

	// 9. Delete a component group
	if err := svc.DeleteComponentGroup(group2.ID); err != nil {
		t.Fatalf("unexpected error deleting component group: %v", err)
	}
}

func TestStatusPageService_IncidentsAndTimeline(t *testing.T) {
	svc, _, _ := setupTestStatusPageSvc(t)

	page, err := svc.CreatePage("user_1", model.CreateStatusPageRequest{
		WorkspaceID: 1,
		Name:        "Incident Test Page",
		Slug:        "incident-test-page",
	})
	if err != nil {
		t.Fatalf("failed to create page: %v", err)
	}

	comp, err := svc.CreateComponent(page.ID, model.CreateComponentRequest{
		Name:        "Edge CDN",
		Status:      model.StatusOperational,
		IsMonitored: true,
	})
	if err != nil {
		t.Fatalf("failed to create component: %v", err)
	}

	// 1. Missing title validation
	_, err = svc.CreateIncident("user_1", page.ID, model.CreateStatusIncidentRequest{
		Title: "",
	})
	if !errors.Is(err, ErrStatusIncidentTitleReq) {
		t.Fatalf("expected ErrStatusIncidentTitleReq, got %v", err)
	}

	// 2. Create incident with affected component and initial timeline update
	inc, err := svc.CreateIncident("user_1", page.ID, model.CreateStatusIncidentRequest{
		Title:                "Elevated 504 Gateway Timeouts in EU Region",
		Status:               model.StatusIncidentInvestigating,
		Impact:               model.ImpactMajor,
		Message:              "We are investigating elevated error rates affecting edge proxy nodes.",
		AffectedComponentIDs: []int64{comp.ID},
		ComponentStatusMap: map[int64]string{
			comp.ID: model.StatusPartialOutage,
		},
	})
	if err != nil {
		t.Fatalf("unexpected error creating incident: %v", err)
	}
	if inc.ID <= 0 {
		t.Fatalf("expected valid incident ID, got %d", inc.ID)
	}
	if len(inc.AffectedComponentIDs) != 1 || inc.AffectedComponentIDs[0] != comp.ID {
		t.Fatalf("expected affected component ID %d, got %v", comp.ID, inc.AffectedComponentIDs)
	}
	if len(inc.Updates) != 1 {
		t.Fatalf("expected 1 initial update, got %d", len(inc.Updates))
	}
	if inc.Updates[0].Status != model.StatusIncidentInvestigating {
		t.Fatalf("expected initial status investigating, got %s", inc.Updates[0].Status)
	}

	// Verify component status was updated to partial_outage
	updatedComp, _ := svc.GetComponent(comp.ID)
	if updatedComp.Status != model.StatusPartialOutage {
		t.Fatalf("expected component status partial_outage, got %s", updatedComp.Status)
	}

	// 3. Add timeline updates: identified
	updateIdentified, err := svc.AddIncidentUpdate("user_1", inc.ID, model.AddStatusIncidentUpdateRequest{
		Status:  model.StatusIncidentIdentified,
		Message: "Root cause isolated to upstream BGP route flap with transit provider.",
	})
	if err != nil {
		t.Fatalf("unexpected error adding update: %v", err)
	}
	if updateIdentified.Status != model.StatusIncidentIdentified {
		t.Fatalf("expected status identified, got %s", updateIdentified.Status)
	}

	// 4. Add timeline update: monitoring
	_, err = svc.AddIncidentUpdate("user_1", inc.ID, model.AddStatusIncidentUpdateRequest{
		Status:  model.StatusIncidentMonitoring,
		Message: "Traffic re-routed around degraded link. Metrics returning to baseline.",
	})
	if err != nil {
		t.Fatalf("unexpected error adding monitoring update: %v", err)
	}

	// 5. Add timeline update: resolved + restore component status
	updateResolved, err := svc.AddIncidentUpdate("user_1", inc.ID, model.AddStatusIncidentUpdateRequest{
		Status:  model.StatusIncidentResolved,
		Message: "Transit provider resolved routing flap. All services fully operational.",
		ComponentStatusMap: map[int64]string{
			comp.ID: model.StatusOperational,
		},
	})
	if err != nil {
		t.Fatalf("unexpected error adding resolved update: %v", err)
	}
	if updateResolved.Status != model.StatusIncidentResolved {
		t.Fatalf("expected status resolved, got %s", updateResolved.Status)
	}

	// 6. Verify incident is marked resolved with timestamp
	fetchedInc, err := svc.GetIncident(inc.ID)
	if err != nil {
		t.Fatalf("unexpected error fetching incident: %v", err)
	}
	if fetchedInc.Status != model.StatusIncidentResolved {
		t.Fatalf("expected resolved status, got %s", fetchedInc.Status)
	}
	if fetchedInc.ResolvedAt == nil {
		t.Fatalf("expected resolved_at to be populated")
	}
	if len(fetchedInc.Updates) != 4 {
		t.Fatalf("expected 4 updates in timeline, got %d", len(fetchedInc.Updates))
	}

	// 7. Verify component is restored to operational
	updatedComp, _ = svc.GetComponent(comp.ID)
	if updatedComp.Status != model.StatusOperational {
		t.Fatalf("expected component status operational, got %s", updatedComp.Status)
	}

	// 8. List incidents
	list, total, err := svc.ListIncidents(page.ID, 10, 0)
	if err != nil {
		t.Fatalf("unexpected error listing incidents: %v", err)
	}
	if total != 1 || len(list) != 1 {
		t.Fatalf("expected 1 incident, got %d total", total)
	}

	// 9. Delete incident
	if err := svc.DeleteIncident(inc.ID); err != nil {
		t.Fatalf("unexpected error deleting incident: %v", err)
	}
	_, err = svc.GetIncident(inc.ID)
	if !errors.Is(err, ErrStatusIncidentNotFound) {
		t.Fatalf("expected ErrStatusIncidentNotFound, got %v", err)
	}
}

func TestStatusPageService_MaintenanceWindows(t *testing.T) {
	svc, _, _ := setupTestStatusPageSvc(t)

	page, err := svc.CreatePage("user_1", model.CreateStatusPageRequest{
		WorkspaceID: 1,
		Name:        "Maintenance Test Page",
		Slug:        "maint-test-page",
	})
	if err != nil {
		t.Fatalf("failed to create page: %v", err)
	}

	comp, err := svc.CreateComponent(page.ID, model.CreateComponentRequest{
		Name:        "Database Cluster",
		Status:      model.StatusOperational,
		IsMonitored: true,
	})
	if err != nil {
		t.Fatalf("failed to create component: %v", err)
	}

	now := time.Now()
	startTime := now.Add(2 * time.Hour)
	endTime := now.Add(4 * time.Hour)

	// 1. Invalid time validation: end before start
	_, err = svc.CreateMaintenance("user_1", page.ID, model.CreateMaintenanceRequest{
		Title:            "Invalid Time Window",
		ScheduledStartAt: endTime,
		ScheduledEndAt:   startTime,
	})
	if !errors.Is(err, ErrInvalidMaintenanceTimes) {
		t.Fatalf("expected ErrInvalidMaintenanceTimes, got %v", err)
	}

	// 2. Create valid maintenance window
	maint, err := svc.CreateMaintenance("user_1", page.ID, model.CreateMaintenanceRequest{
		Title:                "Scheduled PostgreSQL Minor Upgrade",
		Description:          "Upgrading database engine to version 16.3 with failover test",
		ScheduledStartAt:     startTime,
		ScheduledEndAt:       endTime,
		AffectedComponentIDs: []int64{comp.ID},
	})
	if err != nil {
		t.Fatalf("unexpected error creating maintenance: %v", err)
	}
	if maint.Status != model.MaintenanceStatusScheduled {
		t.Fatalf("expected status scheduled, got %s", maint.Status)
	}
	if len(maint.AffectedComponentIDs) != 1 || maint.AffectedComponentIDs[0] != comp.ID {
		t.Fatalf("expected component ID %d in affected list, got %v", comp.ID, maint.AffectedComponentIDs)
	}

	// 3. Transition to in_progress -> component automatically marked under_maintenance
	inProgStatus := model.MaintenanceStatusInProgress
	updated, err := svc.UpdateMaintenance(maint.ID, model.UpdateMaintenanceRequest{
		Status: &inProgStatus,
	})
	if err != nil {
		t.Fatalf("unexpected error updating maintenance: %v", err)
	}
	if updated.Status != model.MaintenanceStatusInProgress {
		t.Fatalf("expected status in_progress, got %s", updated.Status)
	}
	if updated.ActualStartAt == nil {
		t.Fatalf("expected actual_start_at to be set")
	}

	updatedComp, _ := svc.GetComponent(comp.ID)
	if updatedComp.Status != model.StatusUnderMaintenance {
		t.Fatalf("expected component status under_maintenance, got %s", updatedComp.Status)
	}

	// 4. Transition to completed -> component restored to operational
	completedStatus := model.MaintenanceStatusCompleted
	completed, err := svc.UpdateMaintenance(maint.ID, model.UpdateMaintenanceRequest{
		Status: &completedStatus,
	})
	if err != nil {
		t.Fatalf("unexpected error completing maintenance: %v", err)
	}
	if completed.Status != model.MaintenanceStatusCompleted {
		t.Fatalf("expected status completed, got %s", completed.Status)
	}
	if completed.ActualEndAt == nil {
		t.Fatalf("expected actual_end_at to be set")
	}

	updatedComp, _ = svc.GetComponent(comp.ID)
	if updatedComp.Status != model.StatusOperational {
		t.Fatalf("expected component status operational, got %s", updatedComp.Status)
	}

	// 5. List maintenances
	maints, total, err := svc.ListMaintenances(page.ID, 10, 0)
	if err != nil {
		t.Fatalf("unexpected error listing maintenances: %v", err)
	}
	if total != 1 || len(maints) != 1 {
		t.Fatalf("expected 1 maintenance, got %d", total)
	}

	// 6. Delete maintenance
	if err := svc.DeleteMaintenance(maint.ID); err != nil {
		t.Fatalf("unexpected error deleting maintenance: %v", err)
	}
	_, err = svc.GetMaintenance(maint.ID)
	if !errors.Is(err, ErrStatusMaintenanceNotFound) {
		t.Fatalf("expected ErrStatusMaintenanceNotFound, got %v", err)
	}
}

func TestStatusPageService_SubscribersAndMetrics(t *testing.T) {
	svc, _, _ := setupTestStatusPageSvc(t)

	page, err := svc.CreatePage("user_1", model.CreateStatusPageRequest{
		WorkspaceID: 1,
		Name:        "Metrics & Subscribers Page",
		Slug:        "metrics-subs-page",
	})
	if err != nil {
		t.Fatalf("failed to create page: %v", err)
	}

	comp, err := svc.CreateComponent(page.ID, model.CreateComponentRequest{
		Name:        "Web Frontend",
		Status:      model.StatusOperational,
		IsMonitored: true,
	})
	if err != nil {
		t.Fatalf("failed to create component: %v", err)
	}

	// 1. Missing target validation
	_, err = svc.CreateSubscriber(page.ID, model.CreateSubscriberRequest{
		TargetEndpoint: "",
	})
	if !errors.Is(err, ErrSubscriberTargetRequired) {
		t.Fatalf("expected ErrSubscriberTargetRequired, got %v", err)
	}

	// 2. Invalid subscriber type
	_, err = svc.CreateSubscriber(page.ID, model.CreateSubscriberRequest{
		SubscriberType: "carrier_pigeon",
		TargetEndpoint: "user@example.com",
	})
	if !errors.Is(err, ErrInvalidSubscriberType) {
		t.Fatalf("expected ErrInvalidSubscriberType, got %v", err)
	}

	// 3. Create email subscriber
	subEmail, err := svc.CreateSubscriber(page.ID, model.CreateSubscriberRequest{
		SubscriberType:         model.SubscriberTypeEmail,
		TargetEndpoint:         "devops@example.com",
		SubscribedComponentIDs: []int64{comp.ID},
	})
	if err != nil {
		t.Fatalf("unexpected error creating email subscriber: %v", err)
	}
	if subEmail.ID <= 0 {
		t.Fatalf("expected valid subscriber ID, got %d", subEmail.ID)
	}
	if len(subEmail.SubscribedComponentIDs) != 1 || subEmail.SubscribedComponentIDs[0] != comp.ID {
		t.Fatalf("expected component subscription, got %v", subEmail.SubscribedComponentIDs)
	}

	// 4. Create webhook subscriber
	subWebhook, err := svc.CreateSubscriber(page.ID, model.CreateSubscriberRequest{
		SubscriberType: model.SubscriberTypeWebhook,
		TargetEndpoint: "https://ops.example.com/status-webhook",
	})
	if err != nil {
		t.Fatalf("unexpected error creating webhook subscriber: %v", err)
	}
	if subWebhook.SubscriberType != model.SubscriberTypeWebhook {
		t.Fatalf("expected webhook subscriber, got %s", subWebhook.SubscriberType)
	}

	// 5. List subscribers
	subs, err := svc.ListSubscribers(page.ID)
	if err != nil {
		t.Fatalf("unexpected error listing subscribers: %v", err)
	}
	if len(subs) != 2 {
		t.Fatalf("expected 2 subscribers, got %d", len(subs))
	}

	// 6. Delete subscriber
	if err := svc.DeleteSubscriber(subWebhook.ID); err != nil {
		t.Fatalf("unexpected error deleting subscriber: %v", err)
	}
	subsAfter, _ := svc.ListSubscribers(page.ID)
	if len(subsAfter) != 1 {
		t.Fatalf("expected 1 subscriber after deletion, got %d", len(subsAfter))
	}

	// 7. Get status page metrics
	metrics, err := svc.GetPageMetrics(page.ID)
	if err != nil {
		t.Fatalf("unexpected error getting page metrics: %v", err)
	}
	if metrics.TotalComponents != 1 {
		t.Fatalf("expected 1 total component, got %d", metrics.TotalComponents)
	}
	if metrics.OperationalComponents != 1 {
		t.Fatalf("expected 1 operational component, got %d", metrics.OperationalComponents)
	}
	if metrics.TotalSubscribers != 1 {
		t.Fatalf("expected 1 subscriber in metrics, got %d", metrics.TotalSubscribers)
	}
	if metrics.UptimePercentage30d < 99.0 {
		t.Fatalf("expected ~100%% uptime, got %f", metrics.UptimePercentage30d)
	}

	// 8. Get public status overview
	pubPage, groups, incidents, err := svc.GetPublicStatusOverview("metrics-subs-page")
	if err != nil {
		t.Fatalf("unexpected error getting public overview: %v", err)
	}
	if pubPage.ID != page.ID {
		t.Fatalf("expected page ID %d, got %d", page.ID, pubPage.ID)
	}
	if groups == nil {
		t.Fatalf("expected groups slice to be non-nil")
	}
	if incidents == nil {
		t.Fatalf("expected incidents slice to be non-nil")
	}
}
