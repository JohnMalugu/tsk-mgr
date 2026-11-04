package service

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/JohnMalugu/tsk-mgr-api/internal/db"
	"github.com/JohnMalugu/tsk-mgr-api/internal/model"
	"github.com/JohnMalugu/tsk-mgr-api/internal/repository"
)

func setupTestIncidentSvc(t *testing.T) (IncidentService, repository.IncidentRepository, repository.ProjectRepository, repository.TaskRepository, repository.DocumentRepository, *sql.DB) {
	sqliteDB, err := db.InitDB(":memory:")
	if err != nil {
		t.Fatalf("failed to initialize test in-memory db: %v", err)
	}

	_, err = sqliteDB.Exec(`
		INSERT INTO users (id, username, email, password_hash, created_at, updated_at) 
		VALUES ('user_1', 'alice', 'alice@example.com', 'hash', datetime('now'), datetime('now')),
		       ('user_2', 'bob', 'bob@example.com', 'hash', datetime('now'), datetime('now')),
		       ('user_3', 'charlie', 'charlie@example.com', 'hash', datetime('now'), datetime('now'));

		INSERT INTO workspaces (id, name, slug, owner_id, created_at, updated_at) 
		VALUES (1, 'Engineering WS', 'eng-ws-inc', 'user_1', datetime('now'), datetime('now')),
		       (2, 'Product WS', 'prod-ws-inc', 'user_1', datetime('now'), datetime('now'));

		INSERT INTO projects (id, user_id, name, description, color, created_at, updated_at)
		VALUES (101, 'user_1', 'Core API', 'Core Backend Service', '#4f46e5', datetime('now'), datetime('now')),
		       (102, 'user_1', 'Payment Gateway', 'Payment Service', '#10b981', datetime('now'), datetime('now'));

		INSERT INTO tasks (id, user_id, project_id, title, description, status, completed, due_date, estimate_minutes, priority, created_at, updated_at)
		VALUES (501, 'user_1', 101, 'Fix Payment Deadlock', 'Remediation task for payment bug', 'todo', 0, datetime('now'), 60, 'high', datetime('now'), datetime('now')),
		       (502, 'user_1', 101, 'Add Health Check Alerting', 'Preventative monitor', 'in_progress', 0, datetime('now'), 30, 'medium', datetime('now'), datetime('now'));

		INSERT INTO documents (id, workspace_id, project_id, parent_document_id, title, slug, content, doc_type, status, is_pinned, view_count, version, author_id, created_at, updated_at)
		VALUES (601, 1, 101, NULL, 'Payment Outage Postmortem', 'payment-outage-postmortem', '# Root Cause Analysis', 'postmortem', 'published', 0, 1, 1, 'user_1', datetime('now'), datetime('now'));
	`)
	if err != nil {
		t.Fatalf("failed to insert test seed data: %v", err)
	}

	incRepo := repository.NewIncidentRepository(sqliteDB)
	projectRepo := repository.NewProjectRepository(sqliteDB)
	taskRepo := repository.NewTaskRepository(sqliteDB)
	docRepo := repository.NewDocumentRepository(sqliteDB)

	svc := NewIncidentService(incRepo, projectRepo, taskRepo, docRepo)
	return svc, incRepo, projectRepo, taskRepo, docRepo, sqliteDB
}

func TestIncidentService_CreateAndGet(t *testing.T) {
	svc, _, _, _, _, _ := setupTestIncidentSvc(t)

	// 1. Validation: Workspace required
	_, err := svc.CreateIncident("user_1", model.CreateIncidentRequest{
		WorkspaceID: 0,
		Title:       "Payment API Outage",
	})
	if !errors.Is(err, ErrWorkspaceRequired) {
		t.Fatalf("expected ErrWorkspaceRequired, got %v", err)
	}

	// 2. Validation: Title required
	_, err = svc.CreateIncident("user_1", model.CreateIncidentRequest{
		WorkspaceID: 1,
		Title:       "   ",
	})
	if !errors.Is(err, ErrIncidentTitleRequired) {
		t.Fatalf("expected ErrIncidentTitleRequired, got %v", err)
	}

	// 3. Validation: Invalid severity
	_, err = svc.CreateIncident("user_1", model.CreateIncidentRequest{
		WorkspaceID: 1,
		Title:       "Payment Outage",
		Severity:    "sev99",
	})
	if !errors.Is(err, ErrInvalidIncidentSeverity) {
		t.Fatalf("expected ErrInvalidIncidentSeverity, got %v", err)
	}

	// 4. Validation: Project not found
	invalidProj := 9999
	_, err = svc.CreateIncident("user_1", model.CreateIncidentRequest{
		WorkspaceID: 1,
		ProjectID:   &invalidProj,
		Title:       "Payment Outage",
	})
	if !errors.Is(err, ErrProjectNotFound) {
		t.Fatalf("expected ErrProjectNotFound, got %v", err)
	}

	// 5. Successful creation of sev0 incident (Critical outage)
	projID := 101
	inc1, err := svc.CreateIncident("user_1", model.CreateIncidentRequest{
		WorkspaceID:      1,
		ProjectID:        &projID,
		Title:            "Database Connection Pool Exhaustion",
		Summary:          "Core API unable to acquire SQLite connections",
		Severity:         model.IncidentSev0,
		ImpactedServices: []string{"core-api", "billing"},
	})
	if err != nil {
		t.Fatalf("failed to create incident: %v", err)
	}
	if inc1.ID <= 0 {
		t.Fatalf("expected positive incident ID, got %d", inc1.ID)
	}
	if inc1.IncidentNumber != 1 {
		t.Fatalf("expected incident number 1, got %d", inc1.IncidentNumber)
	}
	if inc1.Status != model.IncidentStatusReported {
		t.Fatalf("expected status reported, got %s", inc1.Status)
	}
	// Postmortem must be automatically pending for sev0
	if inc1.PostmortemStatus != model.PostmortemPending {
		t.Fatalf("expected postmortem pending for sev0, got %s", inc1.PostmortemStatus)
	}
	if len(inc1.ImpactedServices) != 2 {
		t.Fatalf("expected 2 impacted services, got %d", len(inc1.ImpactedServices))
	}

	// 6. Successful creation of second incident (sev2, sequential numbering)
	inc2, err := svc.CreateIncident("user_2", model.CreateIncidentRequest{
		WorkspaceID: 1,
		Title:       "Slow Webhook Ingestion",
		Severity:    model.IncidentSev2,
	})
	if err != nil {
		t.Fatalf("failed to create second incident: %v", err)
	}
	if inc2.IncidentNumber != 2 {
		t.Fatalf("expected sequential incident number 2, got %d", inc2.IncidentNumber)
	}
	if inc2.PostmortemStatus != model.PostmortemNotRequired {
		t.Fatalf("expected postmortem not_required for sev2, got %s", inc2.PostmortemStatus)
	}

	// 7. Get by ID
	fetched, err := svc.GetIncident(inc1.ID)
	if err != nil {
		t.Fatalf("failed to get incident by ID: %v", err)
	}
	if fetched.Title != inc1.Title {
		t.Fatalf("expected title '%s', got '%s'", inc1.Title, fetched.Title)
	}

	// 8. Get by Number
	byNum, err := svc.GetIncidentByNumber(1, 1)
	if err != nil {
		t.Fatalf("failed to get incident by number: %v", err)
	}
	if byNum.ID != inc1.ID {
		t.Fatalf("expected incident ID %d, got %d", inc1.ID, byNum.ID)
	}

	// 9. Non-existent returns ErrIncidentNotFound
	_, err = svc.GetIncident(99999)
	if !errors.Is(err, ErrIncidentNotFound) {
		t.Fatalf("expected ErrIncidentNotFound, got %v", err)
	}
}

func TestIncidentService_List(t *testing.T) {
	svc, _, _, _, _, _ := setupTestIncidentSvc(t)

	proj101 := 101
	_, _ = svc.CreateIncident("user_1", model.CreateIncidentRequest{
		WorkspaceID: 1,
		ProjectID:   &proj101,
		Title:       "Auth Token Expiry Failure",
		Severity:    model.IncidentSev1,
	})
	_, _ = svc.CreateIncident("user_1", model.CreateIncidentRequest{
		WorkspaceID: 1,
		Title:       "Search Index Desync",
		Severity:    model.IncidentSev3,
	})

	// List all in workspace 1
	all, err := svc.ListIncidents(model.IncidentFilter{WorkspaceID: 1})
	if err != nil {
		t.Fatalf("failed to list incidents: %v", err)
	}
	if len(all) != 2 {
		t.Fatalf("expected 2 incidents, got %d", len(all))
	}

	// Filter by severity
	sev1s, err := svc.ListIncidents(model.IncidentFilter{WorkspaceID: 1, Severity: model.IncidentSev1})
	if err != nil {
		t.Fatalf("failed to filter by severity: %v", err)
	}
	if len(sev1s) != 1 {
		t.Fatalf("expected 1 sev1 incident, got %d", len(sev1s))
	}

	// Filter by search query
	searched, err := svc.ListIncidents(model.IncidentFilter{WorkspaceID: 1, Search: "Token"})
	if err != nil {
		t.Fatalf("failed to search incidents: %v", err)
	}
	if len(searched) != 1 {
		t.Fatalf("expected 1 matching incident, got %d", len(searched))
	}
}

func TestIncidentService_UpdateAndDelete(t *testing.T) {
	svc, _, _, _, _, _ := setupTestIncidentSvc(t)

	inc, err := svc.CreateIncident("user_1", model.CreateIncidentRequest{
		WorkspaceID: 1,
		Title:       "Original Incident Title",
		Summary:     "Original summary",
	})
	if err != nil {
		t.Fatalf("failed to create incident: %v", err)
	}

	// Update fields
	newTitle := "Updated Critical Incident Title"
	newSummary := "Updated incident impact summary"
	rootCause := "Configuration drift in production env"
	updated, err := svc.UpdateIncident(inc.ID, model.UpdateIncidentRequest{
		Title:     &newTitle,
		Summary:   &newSummary,
		RootCause: &rootCause,
	})
	if err != nil {
		t.Fatalf("failed to update incident: %v", err)
	}
	if updated.Title != newTitle {
		t.Fatalf("expected updated title '%s', got '%s'", newTitle, updated.Title)
	}
	if updated.RootCause != rootCause {
		t.Fatalf("expected updated root cause '%s', got '%s'", rootCause, updated.RootCause)
	}

	// Empty title rejected
	emptyTitle := "   "
	_, err = svc.UpdateIncident(inc.ID, model.UpdateIncidentRequest{Title: &emptyTitle})
	if !errors.Is(err, ErrIncidentTitleRequired) {
		t.Fatalf("expected ErrIncidentTitleRequired, got %v", err)
	}

	// Delete incident
	err = svc.DeleteIncident(inc.ID)
	if err != nil {
		t.Fatalf("failed to delete incident: %v", err)
	}
	_, err = svc.GetIncident(inc.ID)
	if !errors.Is(err, ErrIncidentNotFound) {
		t.Fatalf("expected ErrIncidentNotFound after delete, got %v", err)
	}
}

func TestIncidentService_LifecycleTransitions(t *testing.T) {
	svc, _, _, _, _, _ := setupTestIncidentSvc(t)

	inc, err := svc.CreateIncident("user_1", model.CreateIncidentRequest{
		WorkspaceID: 1,
		Title:       "Payment Webhook Delay",
		Severity:    model.IncidentSev2,
	})
	if err != nil {
		t.Fatalf("failed to create incident: %v", err)
	}

	// 1. Transition: reported -> investigating (should set AcknowledgedAt)
	s1, err := svc.ChangeStatus(inc.ID, "user_1", model.ChangeIncidentStatusRequest{
		Status:  model.IncidentStatusInvestigating,
		Message: "SRE team paged and actively investigating logs",
	})
	if err != nil {
		t.Fatalf("failed to change status to investigating: %v", err)
	}
	if s1.Status != model.IncidentStatusInvestigating {
		t.Fatalf("expected status investigating, got %s", s1.Status)
	}
	if s1.AcknowledgedAt == nil {
		t.Fatalf("expected AcknowledgedAt timestamp to be populated")
	}

	// 2. Transition: investigating -> monitoring (should set MitigatedAt)
	s2, err := svc.ChangeStatus(inc.ID, "user_1", model.ChangeIncidentStatusRequest{
		Status:  model.IncidentStatusMonitoring,
		Message: "Traffic rerouted via backup payment processor",
	})
	if err != nil {
		t.Fatalf("failed to change status to monitoring: %v", err)
	}
	if s2.MitigatedAt == nil {
		t.Fatalf("expected MitigatedAt timestamp to be populated")
	}

	// 3. Transition: monitoring -> resolved (should set ResolvedAt)
	s3, err := svc.ChangeStatus(inc.ID, "user_1", model.ChangeIncidentStatusRequest{
		Status:  model.IncidentStatusResolved,
		Message: "Primary processor connection re-established and healthy",
	})
	if err != nil {
		t.Fatalf("failed to change status to resolved: %v", err)
	}
	if s3.ResolvedAt == nil {
		t.Fatalf("expected ResolvedAt timestamp to be populated")
	}

	// 4. Transition: resolved -> closed (should set ClosedAt)
	s4, err := svc.ChangeStatus(inc.ID, "user_1", model.ChangeIncidentStatusRequest{
		Status: model.IncidentStatusClosed,
	})
	if err != nil {
		t.Fatalf("failed to change status to closed: %v", err)
	}
	if s4.ClosedAt == nil {
		t.Fatalf("expected ClosedAt timestamp to be populated")
	}

	// 5. Invalid status rejected
	_, err = svc.ChangeStatus(inc.ID, "user_1", model.ChangeIncidentStatusRequest{Status: "not_a_valid_status"})
	if !errors.Is(err, ErrInvalidIncidentStatus) {
		t.Fatalf("expected ErrInvalidIncidentStatus, got %v", err)
	}

	// 6. Severity Escalation: sev2 -> sev0 (should automatically escalate postmortem to pending)
	sevChanged, err := svc.ChangeSeverity(inc.ID, "user_1", model.ChangeIncidentSeverityRequest{
		Severity: model.IncidentSev0,
		Reason:   "Outage impact discovered to affect all live transactions",
	})
	if err != nil {
		t.Fatalf("failed to escalate severity: %v", err)
	}
	if sevChanged.Severity != model.IncidentSev0 {
		t.Fatalf("expected sev0, got %s", sevChanged.Severity)
	}
	if sevChanged.PostmortemStatus != model.PostmortemPending {
		t.Fatalf("expected postmortem pending after escalation, got %s", sevChanged.PostmortemStatus)
	}

	// 7. Invalid severity rejected
	_, err = svc.ChangeSeverity(inc.ID, "user_1", model.ChangeIncidentSeverityRequest{Severity: "invalid_sev"})
	if !errors.Is(err, ErrInvalidIncidentSeverity) {
		t.Fatalf("expected ErrInvalidIncidentSeverity, got %v", err)
	}
}

func TestIncidentService_RolesAndTimeline(t *testing.T) {
	svc, _, _, _, _, _ := setupTestIncidentSvc(t)

	inc, err := svc.CreateIncident("user_1", model.CreateIncidentRequest{
		WorkspaceID: 1,
		Title:       "Production Latency Spike",
	})
	if err != nil {
		t.Fatalf("failed to create incident: %v", err)
	}

	// 1. Assign incident commander and scribe
	cmd := "user_2"
	scb := "user_3"
	assigned, err := svc.AssignRoles(inc.ID, "user_1", model.AssignIncidentRolesRequest{
		CommanderID: &cmd,
		ScribeID:    &scb,
	})
	if err != nil {
		t.Fatalf("failed to assign roles: %v", err)
	}
	if assigned.CommanderID == nil || *assigned.CommanderID != "user_2" {
		t.Fatalf("expected commander user_2, got %v", assigned.CommanderID)
	}
	if assigned.ScribeID == nil || *assigned.ScribeID != "user_3" {
		t.Fatalf("expected scribe user_3, got %v", assigned.ScribeID)
	}

	// 2. Add custom timeline event
	ev, err := svc.AddTimelineEvent(inc.ID, "user_2", model.CreateIncidentTimelineRequest{
		EventType: model.TimelineEventMitigation,
		Message:   "Restarted web server workers to clear thread leak",
		Metadata:  `{"nodes_restarted": 4}`,
	})
	if err != nil {
		t.Fatalf("failed to add timeline event: %v", err)
	}
	if ev.ID <= 0 {
		t.Fatalf("expected positive event ID, got %d", ev.ID)
	}

	// 3. Validation: Empty timeline message rejected
	_, err = svc.AddTimelineEvent(inc.ID, "user_2", model.CreateIncidentTimelineRequest{
		Message: "   ",
	})
	if !errors.Is(err, ErrTimelineMessageRequired) {
		t.Fatalf("expected ErrTimelineMessageRequired, got %v", err)
	}

	// 4. Validation: Invalid event type rejected
	_, err = svc.AddTimelineEvent(inc.ID, "user_2", model.CreateIncidentTimelineRequest{
		EventType: "bogus_event",
		Message:   "Some note",
	})
	if !errors.Is(err, ErrInvalidTimelineEventType) {
		t.Fatalf("expected ErrInvalidTimelineEventType, got %v", err)
	}

	// 5. List timeline
	timeline, err := svc.ListTimeline(inc.ID)
	if err != nil {
		t.Fatalf("failed to list timeline: %v", err)
	}
	// Initial declaration + role assignment + mitigation event = at least 3 events
	if len(timeline) < 3 {
		t.Fatalf("expected at least 3 timeline events, got %d", len(timeline))
	}
}

func TestIncidentService_TaskLinks(t *testing.T) {
	svc, _, _, _, _, _ := setupTestIncidentSvc(t)

	inc, err := svc.CreateIncident("user_1", model.CreateIncidentRequest{
		WorkspaceID: 1,
		Title:       "Data Serialization Bug",
	})
	if err != nil {
		t.Fatalf("failed to create incident: %v", err)
	}

	// 1. Invalid task type rejected
	_, err = svc.LinkTask(inc.ID, "user_1", model.LinkIncidentTaskRequest{
		TaskID:   501,
		TaskType: "invalid_type",
	})
	if !errors.Is(err, ErrInvalidIncidentTaskType) {
		t.Fatalf("expected ErrInvalidIncidentTaskType, got %v", err)
	}

	// 2. Non-existent task rejected
	_, err = svc.LinkTask(inc.ID, "user_1", model.LinkIncidentTaskRequest{
		TaskID:   9999,
		TaskType: model.IncidentTaskTypeMitigation,
	})
	if !errors.Is(err, ErrTaskNotFound) {
		t.Fatalf("expected ErrTaskNotFound, got %v", err)
	}

	// 3. Valid task link
	link, err := svc.LinkTask(inc.ID, "user_1", model.LinkIncidentTaskRequest{
		TaskID:   501,
		TaskType: model.IncidentTaskTypeMitigation,
	})
	if err != nil {
		t.Fatalf("failed to link task: %v", err)
	}
	if link.TaskID != 501 || link.TaskType != model.IncidentTaskTypeMitigation {
		t.Fatalf("unexpected task link result: %+v", link)
	}

	// 4. Duplicate link updates task type
	updatedLink, err := svc.LinkTask(inc.ID, "user_1", model.LinkIncidentTaskRequest{
		TaskID:   501,
		TaskType: model.IncidentTaskTypePreventative,
	})
	if err != nil {
		t.Fatalf("failed to update task link: %v", err)
	}
	if updatedLink.TaskType != model.IncidentTaskTypePreventative {
		t.Fatalf("expected updated task type preventative, got %s", updatedLink.TaskType)
	}

	// 5. List linked tasks with enriched title, status, priority
	tasks, err := svc.ListIncidentTasks(inc.ID)
	if err != nil {
		t.Fatalf("failed to list incident tasks: %v", err)
	}
	if len(tasks) != 1 {
		t.Fatalf("expected 1 linked task, got %d", len(tasks))
	}
	if tasks[0].TaskTitle != "Fix Payment Deadlock" {
		t.Fatalf("expected task title 'Fix Payment Deadlock', got '%s'", tasks[0].TaskTitle)
	}

	// 6. Unlink task
	err = svc.UnlinkTask(inc.ID, 501)
	if err != nil {
		t.Fatalf("failed to unlink task: %v", err)
	}
	tasksAfter, err := svc.ListIncidentTasks(inc.ID)
	if err != nil {
		t.Fatalf("failed to list incident tasks after unlink: %v", err)
	}
	if len(tasksAfter) != 0 {
		t.Fatalf("expected 0 linked tasks, got %d", len(tasksAfter))
	}
}

func TestIncidentService_Postmortem(t *testing.T) {
	svc, _, _, _, _, _ := setupTestIncidentSvc(t)

	inc, err := svc.CreateIncident("user_1", model.CreateIncidentRequest{
		WorkspaceID: 1,
		Title:       "Redis Memory OOM Crash",
		Severity:    model.IncidentSev1,
	})
	if err != nil {
		t.Fatalf("failed to create incident: %v", err)
	}

	// 1. Invalid postmortem status rejected
	_, err = svc.UpdatePostmortem(inc.ID, model.UpdatePostmortemRequest{
		PostmortemStatus: "invalid_pm_status",
	})
	if !errors.Is(err, ErrInvalidPostmortemStatus) {
		t.Fatalf("expected ErrInvalidPostmortemStatus, got %v", err)
	}

	// 2. Link non-existent document rejected
	invalidDoc := int64(99999)
	_, err = svc.UpdatePostmortem(inc.ID, model.UpdatePostmortemRequest{
		PostmortemDocID: &invalidDoc,
	})
	if !errors.Is(err, ErrDocumentNotFound) {
		t.Fatalf("expected ErrDocumentNotFound, got %v", err)
	}

	// 3. Valid link to doc 601 and update status to in_review
	validDoc := int64(601)
	cause := "Key eviction policy was disabled leading to OOM"
	updated, err := svc.UpdatePostmortem(inc.ID, model.UpdatePostmortemRequest{
		PostmortemStatus: model.PostmortemInReview,
		PostmortemDocID:  &validDoc,
		RootCause:        &cause,
	})
	if err != nil {
		t.Fatalf("failed to update postmortem: %v", err)
	}
	if updated.PostmortemStatus != model.PostmortemInReview {
		t.Fatalf("expected postmortem in_review, got %s", updated.PostmortemStatus)
	}
	if updated.PostmortemDocID == nil || *updated.PostmortemDocID != 601 {
		t.Fatalf("expected postmortem_doc_id 601, got %v", updated.PostmortemDocID)
	}
	if updated.RootCause != cause {
		t.Fatalf("expected root cause '%s', got '%s'", cause, updated.RootCause)
	}
}

func TestIncidentService_Metrics(t *testing.T) {
	svc, _, _, _, _, _ := setupTestIncidentSvc(t)

	// Create sev0 resolved
	inc1, err := svc.CreateIncident("user_1", model.CreateIncidentRequest{
		WorkspaceID: 1,
		Title:       "Major Cluster Outage",
		Severity:    model.IncidentSev0,
	})
	if err != nil {
		t.Fatalf("failed to create inc1: %v", err)
	}
	_, _ = svc.ChangeStatus(inc1.ID, "user_1", model.ChangeIncidentStatusRequest{Status: model.IncidentStatusInvestigating})
	_, _ = svc.ChangeStatus(inc1.ID, "user_1", model.ChangeIncidentStatusRequest{Status: model.IncidentStatusResolved})
	_, _ = svc.UpdatePostmortem(inc1.ID, model.UpdatePostmortemRequest{PostmortemStatus: model.PostmortemCompleted})

	// Create sev2 active
	_, _ = svc.CreateIncident("user_1", model.CreateIncidentRequest{
		WorkspaceID: 1,
		Title:       "Minor CSS Glitch",
		Severity:    model.IncidentSev3,
	})

	metrics, err := svc.GetIncidentMetrics(1)
	if err != nil {
		t.Fatalf("failed to get incident metrics: %v", err)
	}

	if metrics.TotalIncidents != 2 {
		t.Fatalf("expected 2 total incidents, got %d", metrics.TotalIncidents)
	}
	if metrics.ActiveIncidents != 1 {
		t.Fatalf("expected 1 active incident, got %d", metrics.ActiveIncidents)
	}
	if metrics.ResolvedIncidents != 1 {
		t.Fatalf("expected 1 resolved incident, got %d", metrics.ResolvedIncidents)
	}
	if metrics.SeverityBreakdown[model.IncidentSev0] != 1 {
		t.Fatalf("expected 1 sev0 incident in breakdown, got %d", metrics.SeverityBreakdown[model.IncidentSev0])
	}
	if metrics.SeverityBreakdown[model.IncidentSev3] != 1 {
		t.Fatalf("expected 1 sev3 incident in breakdown, got %d", metrics.SeverityBreakdown[model.IncidentSev3])
	}
	if metrics.PostmortemComplianceRate != 100.0 {
		t.Fatalf("expected 100%% postmortem compliance rate, got %f", metrics.PostmortemComplianceRate)
	}
	if len(metrics.RecentIncidents) != 2 {
		t.Fatalf("expected 2 recent incidents, got %d", len(metrics.RecentIncidents))
	}
}
