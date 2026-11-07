package service

import (
	"database/sql"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/JohnMalugu/tsk-mgr-api/internal/db"
	"github.com/JohnMalugu/tsk-mgr-api/internal/model"
	"github.com/JohnMalugu/tsk-mgr-api/internal/repository"
)

func setupTestAuditSvc(t *testing.T) (AuditService, repository.AuditRepository, *sql.DB) {
	sqliteDB, err := db.InitDB(":memory:")
	if err != nil {
		t.Fatalf("failed to initialize test in-memory db: %v", err)
	}

	_, err = sqliteDB.Exec(`
		INSERT INTO users (id, username, email, password_hash, created_at, updated_at) 
		VALUES ('user_sec_1', 'sec_alice', 'alice@compliance.org', 'hash', datetime('now'), datetime('now')),
		       ('user_sec_admin', 'sec_charlie', 'charlie@compliance.org', 'hash', datetime('now'), datetime('now'));

		INSERT INTO workspaces (id, name, slug, owner_id, created_at, updated_at) 
		VALUES (1, 'Fintech Operations', 'fintech-ops-audit', 'user_sec_admin', datetime('now'), datetime('now')),
		       (2, 'Healthcare Data', 'health-data-audit', 'user_sec_admin', datetime('now'), datetime('now'));

		INSERT INTO workspace_members (workspace_id, user_id, role, joined_at)
		VALUES (1, 'user_sec_1', 'member', datetime('now')),
		       (1, 'user_sec_admin', 'admin', datetime('now')),
		       (2, 'user_sec_admin', 'admin', datetime('now'));
	`)
	if err != nil {
		t.Fatalf("failed to seed test database: %v", err)
	}

	auditRepo := repository.NewAuditRepository(sqliteDB)
	workspaceRepo := repository.NewWorkspaceRepository(sqliteDB)
	userRepo := repository.NewUserRepository(sqliteDB)

	svc := NewAuditService(auditRepo, workspaceRepo, userRepo)
	return svc, auditRepo, sqliteDB
}

func TestAuditService_RecordAndGetEvent(t *testing.T) {
	svc, _, dbConn := setupTestAuditSvc(t)
	defer dbConn.Close()

	// Validation checks
	_, err := svc.RecordEvent("user_sec_1", model.CreateAuditLogRequest{
		WorkspaceID: 0,
		Action:      "auth.login",
	})
	if !errors.Is(err, ErrWorkspaceRequired) {
		t.Fatalf("expected ErrWorkspaceRequired, got %v", err)
	}

	_, err = svc.RecordEvent("user_sec_1", model.CreateAuditLogRequest{
		WorkspaceID: 1,
		Action:      "",
	})
	if !errors.Is(err, model.ErrAuditActionRequired) {
		t.Fatalf("expected ErrAuditActionRequired, got %v", err)
	}

	_, err = svc.RecordEvent("user_sec_1", model.CreateAuditLogRequest{
		WorkspaceID: 1,
		Action:      "auth.login",
		Severity:    "ultra-fatal",
	})
	if !errors.Is(err, model.ErrInvalidAuditSeverity) {
		t.Fatalf("expected ErrInvalidAuditSeverity, got %v", err)
	}

	// Create valid audit event
	log, err := svc.RecordEvent("user_sec_1", model.CreateAuditLogRequest{
		WorkspaceID: 1,
		Action:      model.AuditActionLogin,
		Category:    model.AuditCategoryAuth,
		Severity:    model.AuditSeverityInfo,
		EntityType:  "session",
		EntityID:    "sess_abc123",
		EntityName:  "User Web Session",
		Description: "User successfully authenticated via SSO",
		IPAddress:   "192.168.1.100",
		UserAgent:   "Mozilla/5.0 Chrome/120.0",
		Metadata:    `{"mfa_verified": true, "sso_provider": "okta"}`,
	})
	if err != nil {
		t.Fatalf("RecordEvent failed: %v", err)
	}

	if log.ID <= 0 || log.Checksum == "" || log.ActorName != "sec_alice" || log.ActorEmail != "alice@compliance.org" {
		t.Fatalf("unexpected audit log fields: %+v", log)
	}

	// Fetch event by ID
	fetched, err := svc.GetEvent(log.ID)
	if err != nil {
		t.Fatalf("GetEvent failed: %v", err)
	}
	if fetched.Action != model.AuditActionLogin || fetched.Checksum != log.Checksum {
		t.Fatalf("fetched log does not match created log: %+v", fetched)
	}
}

func TestAuditService_ListEventsAndFilters(t *testing.T) {
	svc, _, dbConn := setupTestAuditSvc(t)
	defer dbConn.Close()

	// Seed multiple audit events
	events := []model.CreateAuditLogRequest{
		{
			WorkspaceID: 1,
			Action:      model.AuditActionLogin,
			Category:    model.AuditCategoryAuth,
			Severity:    model.AuditSeverityInfo,
			Description: "Alice logged in successfully",
		},
		{
			WorkspaceID: 1,
			Action:      model.AuditActionPermissionGranted,
			Category:    model.AuditCategoryAuthz,
			Severity:    model.AuditSeverityWarning,
			Description: "Granted admin permission to worker node",
		},
		{
			WorkspaceID: 1,
			Action:      model.AuditActionDataExported,
			Category:    model.AuditCategoryDataAccess,
			Severity:    model.AuditSeverityWarning,
			Description: "Customer PII export initiated",
		},
		{
			WorkspaceID: 1,
			Action:      "secret.rotate",
			Category:    model.AuditCategorySecurity,
			Severity:    model.AuditSeverityCritical,
			Description: "Production database credentials rotated",
		},
	}

	for _, e := range events {
		if _, err := svc.RecordEvent("user_sec_1", e); err != nil {
			t.Fatalf("failed to record seed event: %v", err)
		}
	}

	// Filter by workspace (all 4)
	logs, total, err := svc.ListEvents(model.AuditLogFilter{WorkspaceID: 1})
	if err != nil || total != 4 || len(logs) != 4 {
		t.Fatalf("expected 4 logs, got total=%d len=%d err=%v", total, len(logs), err)
	}

	// Filter by category: security
	secCat := model.AuditCategorySecurity
	secLogs, secTotal, err := svc.ListEvents(model.AuditLogFilter{
		WorkspaceID: 1,
		Category:    &secCat,
	})
	if err != nil || secTotal != 1 || len(secLogs) != 1 {
		t.Fatalf("expected 1 security log, got %d", secTotal)
	}
	if secLogs[0].Action != "secret.rotate" {
		t.Fatalf("expected secret.rotate action, got %s", secLogs[0].Action)
	}

	// Filter by severity: critical
	critSev := model.AuditSeverityCritical
	critLogs, critTotal, err := svc.ListEvents(model.AuditLogFilter{
		WorkspaceID: 1,
		Severity:    &critSev,
	})
	if err != nil || critTotal != 1 || len(critLogs) != 1 {
		t.Fatalf("expected 1 critical log, got %d", critTotal)
	}

	// Search filter: "PII"
	search := "PII"
	searchLogs, searchTotal, err := svc.ListEvents(model.AuditLogFilter{
		WorkspaceID: 1,
		Search:      &search,
	})
	if err != nil || searchTotal != 1 || len(searchLogs) != 1 {
		t.Fatalf("expected 1 search match for PII, got %d", searchTotal)
	}
}

func TestAuditService_RetentionPolicyAndPurge(t *testing.T) {
	svc, _, dbConn := setupTestAuditSvc(t)
	defer dbConn.Close()

	// Default policy
	policy, err := svc.GetRetentionPolicy(1)
	if err != nil {
		t.Fatalf("GetRetentionPolicy failed: %v", err)
	}
	if policy.RetentionDays != 365 {
		t.Fatalf("expected default retention 365 days, got %d", policy.RetentionDays)
	}

	// Validation: cannot set retention < 30 days
	_, err = svc.ConfigureRetentionPolicy(1, model.RetentionPolicyRequest{
		WorkspaceID:   1,
		RetentionDays: 10,
	})
	if !errors.Is(err, ErrInvalidRetentionPeriod) {
		t.Fatalf("expected ErrInvalidRetentionPeriod, got %v", err)
	}

	// Configure valid 90-day retention policy
	updatedPolicy, err := svc.ConfigureRetentionPolicy(1, model.RetentionPolicyRequest{
		WorkspaceID:      1,
		RetentionDays:    90,
		Frameworks:       []string{model.ComplianceSOC2, model.ComplianceISO27001},
		AutoPurgeEnabled: true,
	})
	if err != nil {
		t.Fatalf("ConfigureRetentionPolicy failed: %v", err)
	}
	if updatedPolicy.RetentionDays != 90 || !strings.Contains(updatedPolicy.Frameworks, "iso27001") {
		t.Fatalf("unexpected updated policy: %+v", updatedPolicy)
	}

	// Seed old audit log created 120 days ago directly into DB
	oldDate := time.Now().AddDate(0, 0, -120)
	_, err = dbConn.Exec(`
		INSERT INTO audit_logs (
			workspace_id, actor_id, actor_name, action, category, severity, description, checksum, created_at
		) VALUES (1, 'user_sec_1', 'sec_alice', 'old.event', 'system', 'info', 'Old event 120 days ago', 'hash123', ?)
	`, oldDate)
	if err != nil {
		t.Fatalf("failed to insert old test log: %v", err)
	}

	// Record a current event
	_, _ = svc.RecordEvent("user_sec_1", model.CreateAuditLogRequest{
		WorkspaceID: 1,
		Action:      "current.event",
		Description: "Current active event",
	})

	// Purge expired events
	purgedCount, err := svc.PurgeExpiredEvents(1)
	if err != nil {
		t.Fatalf("PurgeExpiredEvents failed: %v", err)
	}
	if purgedCount != 1 {
		t.Fatalf("expected 1 purged log, got %d", purgedCount)
	}

	// Verify only 1 log remains
	logs, total, _ := svc.ListEvents(model.AuditLogFilter{WorkspaceID: 1})
	if total != 1 || len(logs) != 1 || logs[0].Action != "current.event" {
		t.Fatalf("expected only current.event to remain after purge, got total=%d", total)
	}
}

func TestAuditService_AnomalyAlerts(t *testing.T) {
	svc, _, dbConn := setupTestAuditSvc(t)
	defer dbConn.Close()

	// Recording a failed login should auto-trigger an anomaly alert
	_, err := svc.RecordEvent("user_sec_1", model.CreateAuditLogRequest{
		WorkspaceID: 1,
		Action:      model.AuditActionFailedLogin,
		Category:    model.AuditCategoryAuth,
		Severity:    model.AuditSeverityWarning,
		Description: "Invalid password attempt for admin",
		IPAddress:   "203.0.113.42",
	})
	if err != nil {
		t.Fatalf("failed to record failed login event: %v", err)
	}

	// Verify alert was created
	alerts, err := svc.ListAlerts(model.AuditAlertFilter{WorkspaceID: 1})
	if err != nil {
		t.Fatalf("ListAlerts failed: %v", err)
	}
	if len(alerts) != 1 {
		t.Fatalf("expected 1 alert, got %d", len(alerts))
	}
	alert := alerts[0]
	if alert.Status != model.AuditAlertStatusOpen {
		t.Fatalf("expected alert status open, got %s", alert.Status)
	}

	// Fetch alert
	fetchedAlert, err := svc.GetAlert(alert.ID)
	if err != nil {
		t.Fatalf("GetAlert failed: %v", err)
	}
	if fetchedAlert.ID != alert.ID {
		t.Fatalf("alert IDs do not match: %d vs %d", fetchedAlert.ID, alert.ID)
	}

	// Invalid status
	_, err = svc.ResolveAlert(alert.ID, "user_sec_admin", model.ResolveAlertRequest{
		Status: "invalid_status",
	})
	if !errors.Is(err, ErrInvalidAlertStatus) {
		t.Fatalf("expected ErrInvalidAlertStatus, got %v", err)
	}

	// Resolve alert
	resolved, err := svc.ResolveAlert(alert.ID, "user_sec_admin", model.ResolveAlertRequest{
		Status:          model.AuditAlertStatusResolved,
		ResolutionNotes: "Investigated IP; false alarm during user credential reset.",
	})
	if err != nil {
		t.Fatalf("ResolveAlert failed: %v", err)
	}
	if resolved.Status != model.AuditAlertStatusResolved || resolved.ResolutionNotes == "" {
		t.Fatalf("unexpected resolved alert state: %+v", resolved)
	}
}

func TestAuditService_ExportAndComplianceMetrics(t *testing.T) {
	svc, _, dbConn := setupTestAuditSvc(t)
	defer dbConn.Close()

	// Seed logs
	_, _ = svc.RecordEvent("user_sec_1", model.CreateAuditLogRequest{
		WorkspaceID: 1,
		Action:      model.AuditActionLogin,
		Category:    model.AuditCategoryAuth,
		Severity:    model.AuditSeverityInfo,
		Description: "Normal login",
	})
	_, _ = svc.RecordEvent("user_sec_1", model.CreateAuditLogRequest{
		WorkspaceID: 1,
		Action:      "database.snapshot",
		Category:    model.AuditCategorySecurity,
		Severity:    model.AuditSeverityInfo,
		Description: "Nightly backup",
	})

	// Generate JSON export
	exportJSON, rawBytes, err := svc.GenerateExport("user_sec_admin", model.ExportAuditLogsRequest{
		WorkspaceID: 1,
		Format:      "json",
	})
	if err != nil {
		t.Fatalf("GenerateExport JSON failed: %v", err)
	}
	if exportJSON.RecordCount != 2 || exportJSON.Checksum == "" {
		t.Fatalf("unexpected JSON export: %+v", exportJSON)
	}

	var parsedJSON []model.AuditLog
	if err := json.Unmarshal(rawBytes, &parsedJSON); err != nil || len(parsedJSON) != 2 {
		t.Fatalf("failed to parse export JSON bytes: %v", err)
	}

	// Generate CSV export
	exportCSV, csvBytes, err := svc.GenerateExport("user_sec_admin", model.ExportAuditLogsRequest{
		WorkspaceID: 1,
		Format:      "csv",
	})
	if err != nil {
		t.Fatalf("GenerateExport CSV failed: %v", err)
	}
	if exportCSV.RecordCount != 2 || exportCSV.Checksum == "" {
		t.Fatalf("unexpected CSV export metadata: %+v", exportCSV)
	}
	if !strings.Contains(string(csvBytes), "database.snapshot") {
		t.Fatalf("CSV export does not contain expected actions: %s", string(csvBytes))
	}

	// List exports
	exports, err := svc.ListExports(1)
	if err != nil || len(exports) != 2 {
		t.Fatalf("expected 2 exports in list, got %d", len(exports))
	}

	// Metrics
	metrics, err := svc.GetAuditMetrics(1)
	if err != nil {
		t.Fatalf("GetAuditMetrics failed: %v", err)
	}
	if metrics.TotalLogs != 2 || metrics.InfoCount != 2 {
		t.Fatalf("unexpected audit metrics: %+v", metrics)
	}

	// Compliance Summary
	summary, err := svc.GetComplianceSummary(1)
	if err != nil {
		t.Fatalf("GetComplianceSummary failed: %v", err)
	}
	if summary.TotalEventsLogged != 2 || !summary.AuditTrailIntegrityVerified || summary.ComplianceScore != 100.0 {
		t.Fatalf("unexpected compliance summary: %+v", summary)
	}
}
