package service

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/JohnMalugu/tsk-mgr-api/internal/db"
	"github.com/JohnMalugu/tsk-mgr-api/internal/model"
	"github.com/JohnMalugu/tsk-mgr-api/internal/repository"
)

func setupTestWorkforceSvc(t *testing.T) (WorkforceService, repository.WorkforceRepository, *sql.DB) {
	sqliteDB, err := db.InitDB(":memory:")
	if err != nil {
		t.Fatalf("failed to initialize test in-memory db: %v", err)
	}
	sqliteDB.SetMaxOpenConns(1)

	_, err = sqliteDB.Exec(`
		INSERT INTO users (id, username, email, password_hash, created_at, updated_at) 
		VALUES ('user_1', 'alice', 'alice@example.com', 'hash', datetime('now'), datetime('now')),
		       ('user_2', 'bob', 'bob@example.com', 'hash', datetime('now'), datetime('now'));

		INSERT INTO workspaces (id, name, slug, owner_id, created_at, updated_at) 
		VALUES (1, 'Workforce WS', 'wf-ws', 'user_1', datetime('now'), datetime('now'));
	`)
	if err != nil {
		t.Fatalf("failed to insert test seed data: %v", err)
	}

	repo := repository.NewWorkforceRepository(sqliteDB)
	svc := NewWorkforceService(repo)
	return svc, repo, sqliteDB
}

func TestWorkforce_SkillCatalogAndProfiles(t *testing.T) {
	svc, _, dbConn := setupTestWorkforceSvc(t)
	defer dbConn.Close()
	ctx := context.Background()

	// Validation failure: missing name
	if _, err := svc.CreateSkill(ctx, "ws_1", model.CreateWFSkillRequest{Category: model.WFSkillCategoryCloud}); err == nil {
		t.Fatal("expected error for missing skill name")
	}

	// Create critical skill
	skill, err := svc.CreateSkill(ctx, "ws_1", model.CreateWFSkillRequest{
		SkillCode:   "SKL-K8S",
		Name:        "Kubernetes Operations",
		Category:    model.WFSkillCategoryCloud,
		Description: "Production cluster administration and incident triage",
		IsCritical:  true,
	})
	if err != nil || skill.ID == "" {
		t.Fatalf("failed to create skill: %v", err)
	}

	// Duplicate code conflict
	if _, err := svc.CreateSkill(ctx, "ws_1", model.CreateWFSkillRequest{
		SkillCode: "SKL-K8S",
		Name:      "Duplicate Skill",
	}); err != ErrWFSkillCodeConflict {
		t.Fatalf("expected ErrWFSkillCodeConflict, got %v", err)
	}

	// Upsert profile (create)
	p1, err := svc.UpsertProfile(ctx, "ws_1", model.UpsertWFProfileRequest{
		UserID:           "user_1",
		SkillID:          skill.ID,
		ProficiencyLevel: model.WFProficiencyIntermediate,
		YearsExperience:  2.5,
	})
	if err != nil || p1.ID == "" {
		t.Fatalf("failed to create profile: %v", err)
	}

	// Upsert profile again (updates existing record rather than duplicating)
	p2, err := svc.UpsertProfile(ctx, "ws_1", model.UpsertWFProfileRequest{
		UserID:           "user_1",
		SkillID:          skill.ID,
		ProficiencyLevel: model.WFProficiencyAdvanced,
		YearsExperience:  3.0,
	})
	if err != nil || p2.ID != p1.ID || p2.ProficiencyLevel != model.WFProficiencyAdvanced {
		t.Fatalf("expected upsert to update existing profile %s, got %+v, err: %v", p1.ID, p2, err)
	}

	// Manager verification
	verified, err := svc.VerifyProfile(ctx, p1.ID, model.VerifyWFProfileRequest{
		VerifiedBy:       "user_2",
		ProficiencyLevel: model.WFProficiencyExpert,
	})
	if err != nil || verified.VerifiedAt == nil || verified.ProficiencyLevel != model.WFProficiencyExpert {
		t.Fatalf("expected verified expert profile, got %+v, err: %v", verified, err)
	}

	profiles, total, err := svc.ListProfiles(ctx, model.WFProfileFilter{WorkspaceID: "ws_1", VerifiedOnly: true})
	if err != nil || total != 1 || len(profiles) != 1 {
		t.Fatalf("expected 1 verified profile, got total %d, err %v", total, err)
	}
}

func TestWorkforce_CertificationsAndExpiration(t *testing.T) {
	svc, _, dbConn := setupTestWorkforceSvc(t)
	defer dbConn.Close()
	ctx := context.Background()

	now := time.Now().UTC()
	activeExp := now.Add(180 * 24 * time.Hour)
	soonExp := now.Add(10 * 24 * time.Hour)
	pastExp := now.Add(-5 * 24 * time.Hour)

	cActive, err := svc.CreateCertification(ctx, "ws_1", model.CreateWFCertRequest{
		UserID:    "user_1",
		CertCode:  "AWS-SAP",
		Name:      "AWS Solutions Architect Professional",
		Issuer:    "Amazon Web Services",
		ExpiresAt: &activeExp,
	})
	if err != nil || cActive.Status != model.WFCertStatusActive {
		t.Fatalf("expected active cert, got %+v, err: %v", cActive, err)
	}

	cSoon, err := svc.CreateCertification(ctx, "ws_1", model.CreateWFCertRequest{
		UserID:    "user_1",
		CertCode:  "CKA",
		Name:      "Certified Kubernetes Administrator",
		Issuer:    "CNCF",
		ExpiresAt: &soonExp,
	})
	if err != nil || cSoon.Status != model.WFCertStatusExpiringSoon {
		t.Fatalf("expected expiring_soon cert, got %+v, err: %v", cSoon, err)
	}

	cExpired, err := svc.CreateCertification(ctx, "ws_1", model.CreateWFCertRequest{
		UserID:    "user_2",
		CertCode:  "ITIL4-MP",
		Name:      "ITIL 4 Managing Professional",
		Issuer:    "PeopleCert",
		ExpiresAt: &pastExp,
	})
	if err != nil || cExpired.Status != model.WFCertStatusExpired {
		t.Fatalf("expected expired cert, got %+v, err: %v", cExpired, err)
	}

	// Renew expired cert -> status automatically transitions back to active
	renewedExp := now.Add(365 * 24 * time.Hour)
	renewed, err := svc.UpdateCertification(ctx, cExpired.ID, model.UpdateWFCertRequest{ExpiresAt: &renewedExp})
	if err != nil || renewed.Status != model.WFCertStatusActive {
		t.Fatalf("expected renewed cert to become active, got %+v, err: %v", renewed, err)
	}
}

func TestWorkforce_RolesTrainingsAndAnalytics(t *testing.T) {
	svc, _, dbConn := setupTestWorkforceSvc(t)
	defer dbConn.Close()
	ctx := context.Background()

	s1, _ := svc.CreateSkill(ctx, "ws_1", model.CreateWFSkillRequest{
		SkillCode:  "SKL-SEC",
		Name:       "Zero Trust Architecture",
		Category:   model.WFSkillCategorySecurity,
		IsCritical: true,
	})
	s2, _ := svc.CreateSkill(ctx, "ws_1", model.CreateWFSkillRequest{
		SkillCode: "SKL-PG",
		Name:      "PostgreSQL Performance Tuning",
		Category:  model.WFSkillCategoryDatabase,
	})

	// Profiles: 1 verified, 1 unverified
	p1, _ := svc.UpsertProfile(ctx, "ws_1", model.UpsertWFProfileRequest{
		UserID: "user_1", SkillID: s1.ID, ProficiencyLevel: model.WFProficiencyExpert, YearsExperience: 5,
	})
	_, _ = svc.VerifyProfile(ctx, p1.ID, model.VerifyWFProfileRequest{VerifiedBy: "user_2"})
	_, _ = svc.UpsertProfile(ctx, "ws_1", model.UpsertWFProfileRequest{
		UserID: "user_2", SkillID: s2.ID, ProficiencyLevel: model.WFProficiencyIntermediate, YearsExperience: 2,
	})

	// Role competency
	rc, err := svc.CreateRoleCompetency(ctx, "ws_1", model.CreateWFRoleCompetencyRequest{
		RoleName:      "Staff SRE",
		SkillID:       s1.ID,
		RequiredLevel: model.WFProficiencyAdvanced,
		MinHeadcount:  2,
		Priority:      model.WFPriorityCritical,
	})
	if err != nil || rc.ID == "" {
		t.Fatalf("failed to create role competency: %v", err)
	}

	// Trainings: 1 completed (via 100% progress), 1 in progress
	t1, err := svc.CreateTraining(ctx, "ws_1", model.CreateWFTrainingRequest{
		UserID: "user_2", SkillID: s1.ID, CourseTitle: "Zero Trust Deep Dive", Provider: "Coursera",
	})
	if err != nil {
		t.Fatalf("failed to create training 1: %v", err)
	}
	pct100 := 100
	t1Done, err := svc.UpdateTraining(ctx, t1.ID, model.UpdateWFTrainingRequest{ProgressPct: &pct100})
	if err != nil || t1Done.Status != model.WFTrainingCompleted || t1Done.CompletedAt == nil {
		t.Fatalf("expected training 1 completed with timestamp, got %+v, err: %v", t1Done, err)
	}

	t2, _ := svc.CreateTraining(ctx, "ws_1", model.CreateWFTrainingRequest{
		UserID: "user_1", SkillID: s2.ID, CourseTitle: "PG Internals", Provider: "PostgresPro",
	})
	pct50 := 50
	t2Prog, _ := svc.UpdateTraining(ctx, t2.ID, model.UpdateWFTrainingRequest{ProgressPct: &pct50})
	if t2Prog.Status != model.WFTrainingInProgress {
		t.Fatalf("expected training 2 in_progress, got %s", t2Prog.Status)
	}

	// Verify analytics
	analytics, err := svc.GetWorkforceAnalytics(ctx, "ws_1")
	if err != nil {
		t.Fatalf("failed to get workforce analytics: %v", err)
	}
	if analytics.TotalSkills != 2 || analytics.CriticalSkills != 1 {
		t.Fatalf("unexpected skill counts: %+v", analytics)
	}
	if analytics.TotalSkillProfiles != 2 || analytics.VerifiedProfiles != 1 {
		t.Fatalf("unexpected profile counts: %+v", analytics)
	}
	if analytics.TotalRoleCompetencies != 1 {
		t.Fatalf("expected 1 role competency, got %d", analytics.TotalRoleCompetencies)
	}
	if analytics.CompletedTrainings != 1 || analytics.ActiveTrainings != 1 || analytics.TrainingCompletionRate != 50.0 {
		t.Fatalf("unexpected training KPIs: %+v", analytics)
	}
	if analytics.WorkforceReadinessIndex != 50.0 {
		t.Fatalf("expected WorkforceReadinessIndex 50.0, got %.2f", analytics.WorkforceReadinessIndex)
	}
}

func TestWorkforce_Concurrency(t *testing.T) {
	svc, _, dbConn := setupTestWorkforceSvc(t)
	defer dbConn.Close()
	ctx := context.Background()

	skill, err := svc.CreateSkill(ctx, "ws_1", model.CreateWFSkillRequest{
		SkillCode: "SKL-GO",
		Name:      "Go Concurrency",
		Category:  model.WFSkillCategorySoftwareEng,
	})
	if err != nil {
		t.Fatalf("failed to create skill: %v", err)
	}

	concurrency := 15
	var wg sync.WaitGroup
	errCh := make(chan error, concurrency)

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			userID := fmt.Sprintf("member_%d", idx)
			if _, err := svc.UpsertProfile(ctx, "ws_1", model.UpsertWFProfileRequest{
				UserID:           userID,
				SkillID:          skill.ID,
				ProficiencyLevel: model.WFProficiencyAdvanced,
				YearsExperience:  float64(idx%5 + 1),
			}); err != nil {
				errCh <- err
				return
			}
			if _, err := svc.CreateTraining(ctx, "ws_1", model.CreateWFTrainingRequest{
				UserID:      userID,
				SkillID:     skill.ID,
				CourseTitle: fmt.Sprintf("Advanced Go Patterns %d", idx),
			}); err != nil {
				errCh <- err
			}
		}(i)
	}

	wg.Wait()
	close(errCh)

	for err := range errCh {
		if err != nil {
			t.Fatalf("concurrent workforce error: %v", err)
		}
	}

	profiles, total, err := svc.ListProfiles(ctx, model.WFProfileFilter{WorkspaceID: "ws_1"})
	if err != nil || total != concurrency || len(profiles) != concurrency {
		t.Fatalf("expected %d profiles, got total %d, err %v", concurrency, total, err)
	}
}
