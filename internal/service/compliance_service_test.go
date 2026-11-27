package service

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"testing"

	"github.com/JohnMalugu/tsk-mgr-api/internal/db"
	"github.com/JohnMalugu/tsk-mgr-api/internal/model"
	"github.com/JohnMalugu/tsk-mgr-api/internal/repository"
)

func setupTestComplianceSvc(t *testing.T) (ComplianceService, repository.ComplianceRepository, *sql.DB) {
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
		VALUES (1, 'Compliance WS', 'grc-ws', 'user_1', datetime('now'), datetime('now'));
	`)
	if err != nil {
		t.Fatalf("failed to insert test seed data: %v", err)
	}

	repo := repository.NewComplianceRepository(sqliteDB)
	svc := NewComplianceService(repo)
	return svc, repo, sqliteDB
}

func TestCompliance_FrameworkAndControlLifecycle(t *testing.T) {
	svc, _, dbConn := setupTestComplianceSvc(t)
	defer dbConn.Close()
	ctx := context.Background()

	// Validation failure: missing name
	if _, err := svc.CreateFramework(ctx, "ws_1", model.CreateGRCFrameworkRequest{StandardType: model.GRCStandardSOC2}); err == nil {
		t.Fatal("expected error for missing framework name")
	}

	// Create SOC 2 framework
	fw, err := svc.CreateFramework(ctx, "ws_1", model.CreateGRCFrameworkRequest{
		FrameworkCode: "SOC2-2026",
		Name:          "SOC 2 Type II",
		StandardType:  model.GRCStandardSOC2,
		Version:       "2026.1",
		Description:   "Trust Services Criteria for Security and Availability",
		OwnerID:       "user_1",
	})
	if err != nil || fw.ID == "" {
		t.Fatalf("failed to create framework: %v", err)
	}

	// Duplicate framework code conflict
	if _, err := svc.CreateFramework(ctx, "ws_1", model.CreateGRCFrameworkRequest{
		FrameworkCode: "SOC2-2026",
		Name:          "Duplicate Framework",
	}); err != ErrGRCFrameworkCodeConflict {
		t.Fatalf("expected ErrGRCFrameworkCodeConflict, got %v", err)
	}

	// Lookup by code & update
	byCode, err := svc.GetFrameworkByCode(ctx, "ws_1", "SOC2-2026")
	if err != nil || byCode.ID != fw.ID {
		t.Fatalf("unexpected GetFrameworkByCode: %+v, err: %v", byCode, err)
	}

	// Create control under framework
	ctrl, err := svc.CreateControl(ctx, "ws_1", model.CreateGRCControlRequest{
		FrameworkID:          fw.ID,
		ControlCode:          "CC6.1",
		Title:                "Logical Access Controls",
		Domain:               model.GRCDomainAccessControl,
		ImplementationStatus: model.GRCControlPartiallyImplemented,
		AutomationLevel:      model.GRCAutomationSemiAutomated,
		OwnerID:              "user_1",
	})
	if err != nil || ctrl.ID == "" {
		t.Fatalf("failed to create control: %v", err)
	}

	// Duplicate control code conflict
	if _, err := svc.CreateControl(ctx, "ws_1", model.CreateGRCControlRequest{
		FrameworkID: fw.ID,
		ControlCode: "CC6.1",
		Title:       "Duplicate Control",
	}); err != ErrGRCControlCodeConflict {
		t.Fatalf("expected ErrGRCControlCodeConflict, got %v", err)
	}

	// Update control status to implemented
	impl := model.GRCControlImplemented
	updatedCtrl, err := svc.UpdateControl(ctx, ctrl.ID, model.UpdateGRCControlRequest{ImplementationStatus: &impl})
	if err != nil || updatedCtrl.ImplementationStatus != model.GRCControlImplemented {
		t.Fatalf("failed to update control: %+v, err: %v", updatedCtrl, err)
	}

	// List controls
	controls, total, err := svc.ListControls(ctx, model.GRCControlFilter{WorkspaceID: "ws_1", FrameworkID: fw.ID})
	if err != nil || total != 1 || len(controls) != 1 {
		t.Fatalf("expected 1 control, got total %d, err %v", total, err)
	}

	// Delete control and framework
	if err := svc.DeleteControl(ctx, ctrl.ID); err != nil {
		t.Fatalf("failed to delete control: %v", err)
	}
	if err := svc.DeleteFramework(ctx, fw.ID); err != nil {
		t.Fatalf("failed to delete framework: %v", err)
	}
}

func TestCompliance_EvidenceSubmissionAndReview(t *testing.T) {
	svc, _, dbConn := setupTestComplianceSvc(t)
	defer dbConn.Close()
	ctx := context.Background()

	fw, _ := svc.CreateFramework(ctx, "ws_1", model.CreateGRCFrameworkRequest{
		FrameworkCode: "ISO-27001",
		Name:          "ISO/IEC 27001:2022",
		StandardType:  model.GRCStandardISO27001,
	})
	ctrl, _ := svc.CreateControl(ctx, "ws_1", model.CreateGRCControlRequest{
		FrameworkID:          fw.ID,
		ControlCode:          "A.8.24",
		Title:                "Use of Cryptography",
		Domain:               model.GRCDomainEncryption,
		ImplementationStatus: model.GRCControlImplemented,
	})

	// Submit evidence
	ev, err := svc.SubmitEvidence(ctx, "ws_1", model.SubmitGRCEvidenceRequest{
		ControlID:    ctrl.ID,
		Title:        "TLS 1.3 Cipher Suite Config Snapshot",
		EvidenceType: model.GRCEvidenceConfigSnapshot,
		ArtifactURL:  "s3://audit-vault/iso27001/tls-config-2026.json",
		CollectedBy:  "user_1",
	})
	if err != nil || ev.ReviewStatus != model.GRCEvidenceStatusPending {
		t.Fatalf("expected pending evidence, got %+v, err: %v", ev, err)
	}

	// Approve evidence -> should stamp LastTestedAt on parent control
	reviewed, err := svc.ReviewEvidence(ctx, ev.ID, model.ReviewGRCEvidenceRequest{
		ReviewStatus: model.GRCEvidenceStatusApproved,
		ReviewedBy:   "user_2",
	})
	if err != nil || reviewed.ReviewStatus != model.GRCEvidenceStatusApproved || reviewed.ReviewedAt == nil {
		t.Fatalf("expected approved evidence with timestamp, got %+v, err: %v", reviewed, err)
	}

	refreshedCtrl, err := svc.GetControlByID(ctx, ctrl.ID)
	if err != nil || refreshedCtrl.LastTestedAt == nil {
		t.Fatalf("expected control LastTestedAt to be stamped upon evidence approval, got %+v", refreshedCtrl)
	}

	evList, total, err := svc.ListEvidence(ctx, model.GRCEvidenceFilter{WorkspaceID: "ws_1", ControlID: ctrl.ID})
	if err != nil || total != 1 || len(evList) != 1 {
		t.Fatalf("expected 1 evidence item, got total %d", total)
	}
}

func TestCompliance_AssessmentsFindingsAndAnalytics(t *testing.T) {
	svc, _, dbConn := setupTestComplianceSvc(t)
	defer dbConn.Close()
	ctx := context.Background()

	fw, _ := svc.CreateFramework(ctx, "ws_1", model.CreateGRCFrameworkRequest{
		FrameworkCode: "PCI-4.0",
		Name:          "PCI-DSS v4.0",
		StandardType:  model.GRCStandardPCIDSS,
	})

	c1, _ := svc.CreateControl(ctx, "ws_1", model.CreateGRCControlRequest{
		FrameworkID:          fw.ID,
		ControlCode:          "REQ-3.5",
		Title:                "PAN Encryption at Rest",
		Domain:               model.GRCDomainEncryption,
		ImplementationStatus: model.GRCControlImplemented,
	})
	c2, _ := svc.CreateControl(ctx, "ws_1", model.CreateGRCControlRequest{
		FrameworkID:          fw.ID,
		ControlCode:          "REQ-10.2",
		Title:                "Automated Audit Trails",
		Domain:               model.GRCDomainLoggingMonitoring,
		ImplementationStatus: model.GRCControlNotImplemented,
	})

	ev, _ := svc.SubmitEvidence(ctx, "ws_1", model.SubmitGRCEvidenceRequest{
		ControlID:   c1.ID,
		Title:       "KMS Key Rotation Policy",
		ArtifactURL: "https://vault.internal/kms-policy.pdf",
	})
	_, _ = svc.ReviewEvidence(ctx, ev.ID, model.ReviewGRCEvidenceRequest{
		ReviewStatus: model.GRCEvidenceStatusApproved,
		ReviewedBy:   "auditor_1",
	})

	// Schedule assessment
	assess, err := svc.CreateAssessment(ctx, "ws_1", model.CreateGRCAssessmentRequest{
		FrameworkID:    fw.ID,
		AssessmentCode: "AUD-PCI-Q4",
		Title:          "Q4 PCI-DSS Readiness Audit",
		AuditorName:    "Deloitte Cyber",
		AuditType:      model.GRCAuditTypeExternal,
	})
	if err != nil || assess.Status != model.GRCAssessmentInProgress {
		t.Fatalf("failed to create assessment: %+v, err: %v", assess, err)
	}

	// Log 2 findings: 1 critical open, 1 medium remediated
	_, err = svc.CreateFinding(ctx, "ws_1", model.CreateGRCFindingRequest{
		AssessmentID:    assess.ID,
		ControlID:       c2.ID,
		Title:           "Missing immutable WORM storage for audit logs",
		Severity:        model.GRCFindingSeverityCritical,
		RemediationPlan: "Enable S3 Object Lock on audit bucket",
	})
	if err != nil {
		t.Fatalf("failed to create finding 1: %v", err)
	}

	f2, err := svc.CreateFinding(ctx, "ws_1", model.CreateGRCFindingRequest{
		AssessmentID: assess.ID,
		ControlID:    c1.ID,
		Title:        "Stale key alias in staging KMS",
		Severity:     model.GRCFindingSeverityMedium,
	})
	if err != nil {
		t.Fatalf("failed to create finding 2: %v", err)
	}

	remStatus := model.GRCFindingStatusRemediated
	updatedF2, err := svc.UpdateFinding(ctx, f2.ID, model.UpdateGRCFindingRequest{Status: &remStatus})
	if err != nil || updatedF2.RemediatedAt == nil {
		t.Fatalf("expected finding 2 remediated with timestamp, got %+v, err: %v", updatedF2, err)
	}

	// Complete assessment
	completed, err := svc.CompleteAssessment(ctx, assess.ID, model.CompleteGRCAssessmentRequest{
		OverallScore: 84.5,
		Summary:      "Strong encryption posture; logging WORM remediation pending",
	})
	if err != nil || completed.Status != model.GRCAssessmentCompleted || completed.EndDate == nil {
		t.Fatalf("failed to complete assessment: %+v, err: %v", completed, err)
	}

	// Verify analytics
	analytics, err := svc.GetComplianceAnalytics(ctx, "ws_1")
	if err != nil {
		t.Fatalf("failed to compute compliance analytics: %v", err)
	}
	if analytics.TotalFrameworks != 1 || analytics.ActiveFrameworks != 1 {
		t.Fatalf("unexpected framework counts: %+v", analytics)
	}
	if analytics.TotalControls != 2 || analytics.ImplementedControls != 1 || analytics.ControlCoverageRate != 50.0 {
		t.Fatalf("unexpected control coverage: %+v", analytics)
	}
	if analytics.ApprovedEvidence != 1 || analytics.CompletedAssessments != 1 {
		t.Fatalf("unexpected evidence/assessment counts: %+v", analytics)
	}
	if analytics.OpenFindings != 1 || analytics.CriticalOpenFindings != 1 || analytics.RemediatedFindings != 1 {
		t.Fatalf("unexpected finding counts: %+v", analytics)
	}
	if analytics.ComplianceHealthIndex != 40.0 {
		t.Fatalf("expected ComplianceHealthIndex 40.0 (50 - 10), got %.2f", analytics.ComplianceHealthIndex)
	}
}

func TestCompliance_Concurrency(t *testing.T) {
	svc, _, dbConn := setupTestComplianceSvc(t)
	defer dbConn.Close()
	ctx := context.Background()

	fw, err := svc.CreateFramework(ctx, "ws_1", model.CreateGRCFrameworkRequest{
		FrameworkCode: "NIST-CONC",
		Name:          "NIST CSF Concurrency Test",
		StandardType:  model.GRCStandardNISTCSF,
	})
	if err != nil {
		t.Fatalf("failed to create framework: %v", err)
	}

	concurrency := 15
	var wg sync.WaitGroup
	errCh := make(chan error, concurrency)

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			ctrl, err := svc.CreateControl(ctx, "ws_1", model.CreateGRCControlRequest{
				FrameworkID:          fw.ID,
				Title:                fmt.Sprintf("Automated Control %d", idx),
				Domain:               model.GRCDomainAccessControl,
				ImplementationStatus: model.GRCControlImplemented,
			})
			if err != nil {
				errCh <- err
				return
			}
			_, err = svc.SubmitEvidence(ctx, "ws_1", model.SubmitGRCEvidenceRequest{
				ControlID:   ctrl.ID,
				Title:       fmt.Sprintf("Automated Evidence %d", idx),
				ArtifactURL: fmt.Sprintf("https://audit.local/ev-%d.json", idx),
			})
			if err != nil {
				errCh <- err
			}
		}(i)
	}

	wg.Wait()
	close(errCh)

	for err := range errCh {
		if err != nil {
			t.Fatalf("concurrent compliance error: %v", err)
		}
	}

	controls, total, err := svc.ListControls(ctx, model.GRCControlFilter{WorkspaceID: "ws_1"})
	if err != nil || total != concurrency || len(controls) != concurrency {
		t.Fatalf("expected %d controls, got total %d, err %v", concurrency, total, err)
	}
}
