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

func setupTestOnCallSvc(t *testing.T) (OnCallService, repository.OnCallRepository, *sql.DB) {
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
		VALUES (1, 'Engineering WS', 'eng-ws-oncall', 'user_1', datetime('now'), datetime('now'));

		INSERT INTO projects (id, user_id, name, description, color, created_at, updated_at)
		VALUES (101, 'user_1', 'Infrastructure Reliability', 'Reliability and on-call services', '#4f46e5', datetime('now'), datetime('now'));

		INSERT INTO incidents (id, workspace_id, project_id, incident_number, title, summary, severity, status, reported_by_id, started_at, created_at, updated_at)
		VALUES (301, 1, 101, 1, 'Production Outage in US-East', 'High latency and error rates in ingress', 'sev1', 'investigating', 'user_1', datetime('now'), datetime('now'), datetime('now'));
	`)
	if err != nil {
		t.Fatalf("failed to insert test seed data: %v", err)
	}

	oncallRepo := repository.NewOnCallRepository(sqliteDB)
	projectRepo := repository.NewProjectRepository(sqliteDB)
	userRepo := repository.NewUserRepository(sqliteDB)

	svc := NewOnCallService(oncallRepo, projectRepo, userRepo)
	return svc, oncallRepo, sqliteDB
}

func TestOnCallSchedule_CreateAndGet(t *testing.T) {
	svc, _, _ := setupTestOnCallSvc(t)

	req := model.CreateScheduleRequest{
		WorkspaceID:          1,
		Name:                 "Primary SRE Rotation",
		Description:          "24/7 Primary rotation for backend services",
		Timezone:             "UTC",
		RotationType:         model.RotationWeekly,
		RotationIntervalDays: 7,
		ShiftStartTime:       "09:00",
		ShiftDurationHours:   24,
		ParticipantUserIDs:   []string{"user_1", "user_2"},
	}

	sched, err := svc.CreateSchedule("user_1", req)
	if err != nil {
		t.Fatalf("unexpected error creating schedule: %v", err)
	}
	if sched.ID <= 0 {
		t.Fatalf("expected positive schedule ID, got %d", sched.ID)
	}
	if sched.Name != "Primary SRE Rotation" {
		t.Errorf("expected schedule name 'Primary SRE Rotation', got %s", sched.Name)
	}
	if sched.ParticipantsCount != 2 {
		t.Errorf("expected 2 participants, got %d", sched.ParticipantsCount)
	}

	// Verify GetSchedule
	fetched, err := svc.GetSchedule(sched.ID)
	if err != nil {
		t.Fatalf("unexpected error getting schedule: %v", err)
	}
	if fetched.ID != sched.ID {
		t.Errorf("expected ID %d, got %d", sched.ID, fetched.ID)
	}
}

func TestOnCallSchedule_ValidationErrors(t *testing.T) {
	svc, _, _ := setupTestOnCallSvc(t)

	// Missing name
	_, err := svc.CreateSchedule("user_1", model.CreateScheduleRequest{
		WorkspaceID:  1,
		RotationType: model.RotationWeekly,
	})
	if !errors.Is(err, ErrScheduleNameRequired) {
		t.Errorf("expected ErrScheduleNameRequired, got %v", err)
	}

	// Invalid rotation type
	_, err = svc.CreateSchedule("user_1", model.CreateScheduleRequest{
		WorkspaceID:  1,
		Name:         "Invalid Rotation Sched",
		RotationType: "monthly_invalid",
	})
	if !errors.Is(err, ErrInvalidRotationType) {
		t.Errorf("expected ErrInvalidRotationType, got %v", err)
	}

	// Invalid timezone
	_, err = svc.CreateSchedule("user_1", model.CreateScheduleRequest{
		WorkspaceID:  1,
		Name:         "Invalid TZ Sched",
		RotationType: model.RotationWeekly,
		Timezone:     "Mars/Olympus_Mons",
	})
	if !errors.Is(err, ErrInvalidTimezone) {
		t.Errorf("expected ErrInvalidTimezone, got %v", err)
	}
}

func TestOnCallSchedule_ListUpdateAndDelete(t *testing.T) {
	svc, _, _ := setupTestOnCallSvc(t)

	// Create 2 schedules
	s1, err := svc.CreateSchedule("user_1", model.CreateScheduleRequest{
		WorkspaceID: 1,
		Name:        "Schedule Alpha",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	s2, err := svc.CreateSchedule("user_1", model.CreateScheduleRequest{
		WorkspaceID: 1,
		Name:        "Schedule Beta",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// List schedules
	schedules, total, err := svc.ListSchedules(model.OnCallFilter{WorkspaceID: 1})
	if err != nil {
		t.Fatalf("unexpected error listing schedules: %v", err)
	}
	if total != 2 || len(schedules) != 2 {
		t.Fatalf("expected 2 schedules, got total=%d len=%d", total, len(schedules))
	}

	// Update schedule Alpha
	newName := "Schedule Alpha Refactored"
	newInterval := 14
	updated, err := svc.UpdateSchedule(s1.ID, model.UpdateScheduleRequest{
		Name:                 &newName,
		RotationIntervalDays: &newInterval,
	})
	if err != nil {
		t.Fatalf("unexpected error updating schedule: %v", err)
	}
	if updated.Name != newName {
		t.Errorf("expected updated name %s, got %s", newName, updated.Name)
	}

	// Delete schedule Beta
	if err := svc.DeleteSchedule(s2.ID); err != nil {
		t.Fatalf("unexpected error deleting schedule: %v", err)
	}

	// Verify s2 is gone
	_, err = svc.GetSchedule(s2.ID)
	if !errors.Is(err, ErrScheduleNotFound) {
		t.Errorf("expected ErrScheduleNotFound, got %v", err)
	}
}

func TestOnCallSchedule_ParticipantsAndShifts(t *testing.T) {
	svc, _, _ := setupTestOnCallSvc(t)

	sched, err := svc.CreateSchedule("user_1", model.CreateScheduleRequest{
		WorkspaceID:          1,
		Name:                 "DevOps Rotation",
		RotationType:         model.RotationWeekly,
		RotationIntervalDays: 7,
		ShiftStartTime:       "09:00",
		ShiftDurationHours:   24,
	})
	if err != nil {
		t.Fatalf("failed to create schedule: %v", err)
	}

	// Set participants
	if err := svc.SetParticipants(sched.ID, []string{"user_1", "user_2", "user_3"}); err != nil {
		t.Fatalf("failed to set participants: %v", err)
	}

	parts, err := svc.ListParticipants(sched.ID)
	if err != nil {
		t.Fatalf("failed to list participants: %v", err)
	}
	if len(parts) != 3 {
		t.Fatalf("expected 3 participants, got %d", len(parts))
	}
	if parts[0].UserID != "user_1" || parts[1].UserID != "user_2" || parts[2].UserID != "user_3" {
		t.Errorf("unexpected participant order: %+v", parts)
	}

	// Generate shifts for 28 days
	shifts, err := svc.GenerateShifts(sched.ID, 28)
	if err != nil {
		t.Fatalf("failed to generate shifts: %v", err)
	}
	if len(shifts) == 0 {
		t.Fatalf("expected generated shifts, got 0")
	}

	// List shifts
	now := time.Now()
	listedShifts, err := svc.ListShifts(sched.ID, now.AddDate(0, 0, -1), now.AddDate(0, 0, 30))
	if err != nil {
		t.Fatalf("failed to list shifts: %v", err)
	}
	if len(listedShifts) == 0 {
		t.Errorf("expected listed shifts, got none")
	}
}

func TestOnCall_ActiveResponderAndOverride(t *testing.T) {
	svc, _, _ := setupTestOnCallSvc(t)

	now := time.Now()
	sched, err := svc.CreateSchedule("user_1", model.CreateScheduleRequest{
		WorkspaceID:          1,
		Name:                 "Platform Rotation",
		RotationType:         model.RotationDaily,
		RotationIntervalDays: 1,
		ShiftStartTime:       "00:00",
		ShiftDurationHours:   24,
		StartDate:            now.Truncate(24 * time.Hour),
		ParticipantUserIDs:   []string{"user_1", "user_2"},
	})
	if err != nil {
		t.Fatalf("failed to create schedule: %v", err)
	}

	// Resolve active responder under normal shift
	resp, err := svc.GetActiveResponder(sched.ID)
	if err != nil {
		t.Fatalf("failed to get active responder: %v", err)
	}
	if resp.UserID != "user_1" {
		t.Errorf("expected user_1 as initial responder, got %s", resp.UserID)
	}
	if resp.IsOverride {
		t.Errorf("expected IsOverride=false, got true")
	}

	// Add an active override for user_3 right now
	override, err := svc.CreateOverride("user_1", sched.ID, model.CreateOverrideRequest{
		OverrideUserID: "user_3",
		StartTime:      now.Add(-1 * time.Hour),
		EndTime:        now.Add(5 * time.Hour),
		Reason:         "Coverage swap for team training",
	})
	if err != nil {
		t.Fatalf("failed to create override: %v", err)
	}
	if override.ID <= 0 {
		t.Fatalf("expected positive override ID, got %d", override.ID)
	}

	// Active responder should now resolve to user_3
	activeResp, err := svc.GetActiveResponder(sched.ID)
	if err != nil {
		t.Fatalf("failed to get active responder during override: %v", err)
	}
	if activeResp.UserID != "user_3" {
		t.Errorf("expected user_3 during override, got %s", activeResp.UserID)
	}
	if !activeResp.IsOverride {
		t.Errorf("expected IsOverride=true, got false")
	}

	// Delete override
	if err := svc.DeleteOverride(override.ID); err != nil {
		t.Fatalf("failed to delete override: %v", err)
	}

	// Active responder reverts back to user_1
	revertedResp, err := svc.GetActiveResponder(sched.ID)
	if err != nil {
		t.Fatalf("failed to get active responder after override deletion: %v", err)
	}
	if revertedResp.UserID != "user_1" {
		t.Errorf("expected user_1 after override removal, got %s", revertedResp.UserID)
	}
}

func TestEscalationPolicies_Lifecycle(t *testing.T) {
	svc, _, _ := setupTestOnCallSvc(t)

	// Create escalation policy with 3 steps
	req := model.CreateEscalationPolicyRequest{
		WorkspaceID: 1,
		Name:        "Critical Outage Escalation Policy",
		Description: "Multi-tiered paging rules for severity 1 incidents",
		RepeatCount: 2,
		IsDefault:   true,
		Steps: []model.CreateEscalationStepRequest{
			{TargetType: model.TargetUser, TargetID: "user_1", DelayMinutes: 5},
			{TargetType: model.TargetUser, TargetID: "user_2", DelayMinutes: 10},
			{TargetType: model.TargetWebhook, TargetID: "https://ops.example.com/webhook", DelayMinutes: 15},
		},
	}

	policy, err := svc.CreateEscalationPolicy(req)
	if err != nil {
		t.Fatalf("failed to create escalation policy: %v", err)
	}
	if policy.ID <= 0 {
		t.Fatalf("expected positive policy ID, got %d", policy.ID)
	}
	if policy.StepsCount != 3 {
		t.Errorf("expected 3 steps, got %d", policy.StepsCount)
	}
	if !policy.IsDefault {
		t.Errorf("expected policy to be default, got false")
	}

	// Default policy lookup
	defaultPol, err := svc.GetDefaultEscalationPolicy(1)
	if err != nil {
		t.Fatalf("failed to get default policy: %v", err)
	}
	if defaultPol.ID != policy.ID {
		t.Errorf("expected default policy ID %d, got %d", policy.ID, defaultPol.ID)
	}

	// List policies
	policies, err := svc.ListEscalationPolicies(1)
	if err != nil {
		t.Fatalf("failed to list policies: %v", err)
	}
	if len(policies) != 1 {
		t.Errorf("expected 1 policy, got %d", len(policies))
	}

	// Update policy
	updated, err := svc.UpdateEscalationPolicy(policy.ID, model.CreateEscalationPolicyRequest{
		Name:        "Updated Critical Policy",
		RepeatCount: 1,
	})
	if err != nil {
		t.Fatalf("failed to update policy: %v", err)
	}
	if updated.Name != "Updated Critical Policy" {
		t.Errorf("expected updated name, got %s", updated.Name)
	}

	// Delete policy
	if err := svc.DeleteEscalationPolicy(policy.ID); err != nil {
		t.Fatalf("failed to delete policy: %v", err)
	}
	_, err = svc.GetEscalationPolicy(policy.ID)
	if !errors.Is(err, ErrEscalationPolicyNotFound) {
		t.Errorf("expected ErrEscalationPolicyNotFound, got %v", err)
	}
}

func TestOnCallPage_LifecycleWorkflows(t *testing.T) {
	svc, _, _ := setupTestOnCallSvc(t)

	// Create policy
	policy, err := svc.CreateEscalationPolicy(model.CreateEscalationPolicyRequest{
		WorkspaceID: 1,
		Name:        "Paging Policy",
		IsDefault:   true,
		Steps: []model.CreateEscalationStepRequest{
			{TargetType: model.TargetUser, TargetID: "user_1", DelayMinutes: 5},
			{TargetType: model.TargetUser, TargetID: "user_2", DelayMinutes: 10},
		},
	})
	if err != nil {
		t.Fatalf("failed to create policy: %v", err)
	}

	incID := int64(301)
	// 1. Trigger Page
	page, err := svc.TriggerPage("user_1", model.TriggerPageRequest{
		WorkspaceID:   1,
		PolicyID:      policy.ID,
		IncidentID:    &incID,
		Title:         "High Error Rate on Gateway",
		Summary:       "500 Internal Server Errors surged above 10%",
		TriggerSource: model.PageTriggerIncident,
	})
	if err != nil {
		t.Fatalf("failed to trigger page: %v", err)
	}
	if page.Status != model.PageStatusTriggered {
		t.Errorf("expected status 'triggered', got %s", page.Status)
	}
	if page.CurrentStepOrder != 1 {
		t.Errorf("expected step order 1, got %d", page.CurrentStepOrder)
	}
	if page.PagedUserID == nil || *page.PagedUserID != "user_1" {
		t.Errorf("expected paged user 'user_1', got %v", page.PagedUserID)
	}

	// 2. Acknowledge Page
	ackPage, err := svc.AcknowledgePage(page.ID, "user_1", model.AcknowledgePageRequest{
		Notes: "Investigating traffic spike in US-East",
	})
	if err != nil {
		t.Fatalf("failed to acknowledge page: %v", err)
	}
	if ackPage.Status != model.PageStatusAcknowledged {
		t.Errorf("expected status 'acknowledged', got %s", ackPage.Status)
	}
	if ackPage.AcknowledgedBy == nil || *ackPage.AcknowledgedBy != "user_1" {
		t.Errorf("expected acknowledged_by 'user_1', got %v", ackPage.AcknowledgedBy)
	}
	if ackPage.AcknowledgedAt == nil {
		t.Errorf("expected acknowledged_at timestamp to be set")
	}

	// 3. Escalate Page to Step 2
	escPage, err := svc.EscalatePage(page.ID)
	if err != nil {
		t.Fatalf("failed to escalate page: %v", err)
	}
	if escPage.Status != model.PageStatusEscalated {
		t.Errorf("expected status 'escalated', got %s", escPage.Status)
	}
	if escPage.CurrentStepOrder != 2 {
		t.Errorf("expected current_step_order 2, got %d", escPage.CurrentStepOrder)
	}
	if escPage.PagedUserID == nil || *escPage.PagedUserID != "user_2" {
		t.Errorf("expected paged user 'user_2' at step 2, got %v", escPage.PagedUserID)
	}

	// 4. Resolve Page
	resPage, err := svc.ResolvePage(page.ID, model.ResolvePageRequest{
		ResolutionNotes: "Traffic rate-limiting applied. Error rates dropped to normal.",
	})
	if err != nil {
		t.Fatalf("failed to resolve page: %v", err)
	}
	if resPage.Status != model.PageStatusResolved {
		t.Errorf("expected status 'resolved', got %s", resPage.Status)
	}
	if resPage.ResolvedAt == nil {
		t.Errorf("expected resolved_at timestamp to be set")
	}

	// 5. List Pages
	pages, total, err := svc.ListPages(1, "", 10, 0)
	if err != nil {
		t.Fatalf("failed to list pages: %v", err)
	}
	if total != 1 || len(pages) != 1 {
		t.Errorf("expected 1 total page, got total=%d len=%d", total, len(pages))
	}
}

func TestOnCallMetrics(t *testing.T) {
	svc, _, _ := setupTestOnCallSvc(t)

	// Create schedule
	sched, err := svc.CreateSchedule("user_1", model.CreateScheduleRequest{
		WorkspaceID: 1,
		Name:        "Metrics Schedule",
	})
	if err != nil {
		t.Fatalf("failed to create schedule: %v", err)
	}

	// Create policy
	policy, err := svc.CreateEscalationPolicy(model.CreateEscalationPolicyRequest{
		WorkspaceID: 1,
		Name:        "Metrics Policy",
		IsDefault:   true,
		Steps: []model.CreateEscalationStepRequest{
			{TargetType: model.TargetUser, TargetID: "user_1"},
		},
	})
	if err != nil {
		t.Fatalf("failed to create policy: %v", err)
	}

	// Trigger and resolve page
	p, err := svc.TriggerPage("user_1", model.TriggerPageRequest{
		WorkspaceID: 1,
		PolicyID:    policy.ID,
		Title:       "Sample Incident Page",
	})
	if err != nil {
		t.Fatalf("failed to trigger page: %v", err)
	}
	_, _ = svc.AcknowledgePage(p.ID, "user_1", model.AcknowledgePageRequest{})
	_, _ = svc.ResolvePage(p.ID, model.ResolvePageRequest{})

	// Fetch metrics
	metrics, err := svc.GetOnCallMetrics(1)
	if err != nil {
		t.Fatalf("failed to fetch metrics: %v", err)
	}
	if metrics.TotalSchedules < 1 {
		t.Errorf("expected TotalSchedules >= 1, got %d", metrics.TotalSchedules)
	}
	if metrics.TotalPolicies < 1 {
		t.Errorf("expected TotalPolicies >= 1, got %d", metrics.TotalPolicies)
	}
	if metrics.TotalPagesTriggered < 1 {
		t.Errorf("expected TotalPagesTriggered >= 1, got %d", metrics.TotalPagesTriggered)
	}
	if metrics.ResolvedPages < 1 {
		t.Errorf("expected ResolvedPages >= 1, got %d", metrics.ResolvedPages)
	}

	_ = sched
}
