package service

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/JohnMalugu/tsk-mgr-api/internal/db"
	"github.com/JohnMalugu/tsk-mgr-api/internal/model"
	"github.com/JohnMalugu/tsk-mgr-api/internal/repository"
)

func setupTestVendorSvc(t *testing.T) (VendorService, repository.VendorRepository, *sql.DB) {
	sqliteDB, err := db.InitDB(":memory:")
	if err != nil {
		t.Fatalf("failed to initialize test in-memory db: %v", err)
	}

	_, err = sqliteDB.Exec(`
		INSERT INTO users (id, username, email, password_hash, created_at, updated_at) 
		VALUES ('user_1', 'alice', 'alice@example.com', 'hash', datetime('now'), datetime('now')),
		       ('user_2', 'bob', 'bob@example.com', 'hash', datetime('now'), datetime('now'));

		INSERT INTO workspaces (id, name, slug, owner_id, created_at, updated_at) 
		VALUES (1, 'Engineering WS', 'eng-ws-vendor', 'user_1', datetime('now'), datetime('now'));

		INSERT INTO configuration_items (id, workspace_id, ci_code, name, ci_type, status, criticality, owner_user_id, created_at, updated_at)
		VALUES (10, 1, 'CI-0010', 'Production Kubernetes Cluster', 'service', 'active', 'tier_1_mission_critical', 'user_1', datetime('now'), datetime('now')),
		       (11, 1, 'CI-0011', 'Primary Database Server', 'database', 'active', 'tier_2_business_critical', 'user_1', datetime('now'), datetime('now'));
	`)
	if err != nil {
		t.Fatalf("failed to insert test seed data: %v", err)
	}

	repo := repository.NewVendorRepository(sqliteDB)
	svc := NewVendorService(repo)
	return svc, repo, sqliteDB
}

func TestVendor_CreateAndGet(t *testing.T) {
	svc, _, dbConn := setupTestVendorSvc(t)
	defer dbConn.Close()

	ctx := context.Background()

	// 1. Validation failure: empty name
	_, err := svc.CreateVendor(ctx, &model.CreateVendorRequest{
		WorkspaceID: 1,
		Name:        "",
	})
	if err != ErrVendorNameRequired {
		t.Fatalf("expected ErrVendorNameRequired, got: %v", err)
	}

	// 2. Successful creation with auto slug
	v, err := svc.CreateVendor(ctx, &model.CreateVendorRequest{
		WorkspaceID:            1,
		Name:                   "Amazon Web Services",
		Category:               model.VendorCategoryCloud,
		Tier:                   model.VendorTierStrategic,
		Status:                 model.VendorStatusActive,
		RiskLevel:              model.VendorRiskLow,
		Website:                "https://aws.amazon.com",
		ContactPerson:          "Jeff",
		ContactEmail:           "aws-support@amazon.com",
		SecurityCertifications: "SOC 2 Type II, ISO 27001",
	})
	if err != nil {
		t.Fatalf("failed to create vendor: %v", err)
	}
	if v.ID <= 0 {
		t.Fatalf("expected positive vendor ID, got: %d", v.ID)
	}
	if v.Slug != "amazon-web-services" {
		t.Fatalf("expected slug amazon-web-services, got: %s", v.Slug)
	}

	// 3. Conflict on duplicate slug
	_, err = svc.CreateVendor(ctx, &model.CreateVendorRequest{
		WorkspaceID: 1,
		Name:        "Amazon Web Services",
		Slug:        "amazon-web-services",
	})
	if err != ErrVendorSlugConflict {
		t.Fatalf("expected ErrVendorSlugConflict, got: %v", err)
	}

	// 4. Fetch by ID
	fetched, err := svc.GetVendorByID(ctx, v.ID)
	if err != nil {
		t.Fatalf("failed to get vendor by ID: %v", err)
	}
	if fetched.Name != "Amazon Web Services" {
		t.Fatalf("expected name Amazon Web Services, got: %s", fetched.Name)
	}
	if fetched.Category != model.VendorCategoryCloud {
		t.Fatalf("expected category cloud, got: %s", fetched.Category)
	}

	// 5. Fetch by Slug
	bySlug, err := svc.GetVendorBySlug(ctx, 1, "amazon-web-services")
	if err != nil {
		t.Fatalf("failed to get vendor by slug: %v", err)
	}
	if bySlug.ID != v.ID {
		t.Fatalf("expected ID %d, got: %d", v.ID, bySlug.ID)
	}

	// 6. Not found
	_, err = svc.GetVendorByID(ctx, 99999)
	if err != ErrVendorNotFound {
		t.Fatalf("expected ErrVendorNotFound, got: %v", err)
	}
}

func TestVendor_UpdateAndList(t *testing.T) {
	svc, _, dbConn := setupTestVendorSvc(t)
	defer dbConn.Close()

	ctx := context.Background()

	v, err := svc.CreateVendor(ctx, &model.CreateVendorRequest{
		WorkspaceID: 1,
		Name:        "Datadog Inc",
		Category:    model.VendorCategorySoftware,
		Tier:        model.VendorTierPreferred,
		Status:      model.VendorStatusActive,
		RiskLevel:   model.VendorRiskLow,
	})
	if err != nil {
		t.Fatalf("failed to create vendor: %v", err)
	}

	// Update vendor
	newTier := model.VendorTierStrategic
	newRisk := model.VendorRiskMedium
	newNotes := "Upgraded to strategic monitoring partner"
	updated, err := svc.UpdateVendor(ctx, v.ID, &model.UpdateVendorRequest{
		Tier:      &newTier,
		RiskLevel: &newRisk,
		Notes:     &newNotes,
	})
	if err != nil {
		t.Fatalf("failed to update vendor: %v", err)
	}
	if updated.Tier != model.VendorTierStrategic {
		t.Fatalf("expected tier strategic, got: %s", updated.Tier)
	}
	if updated.RiskLevel != model.VendorRiskMedium {
		t.Fatalf("expected risk medium, got: %s", updated.RiskLevel)
	}
	if updated.Notes != newNotes {
		t.Fatalf("expected updated notes, got: %s", updated.Notes)
	}

	// List vendors with search
	vendors, total, err := svc.ListVendors(ctx, model.VendorFilterCriteria{
		WorkspaceID: 1,
		Search:      "Datadog",
	})
	if err != nil {
		t.Fatalf("failed to list vendors: %v", err)
	}
	if total != 1 || len(vendors) != 1 {
		t.Fatalf("expected 1 vendor, got total=%d, len=%d", total, len(vendors))
	}
}

func TestVendor_Delete(t *testing.T) {
	svc, _, dbConn := setupTestVendorSvc(t)
	defer dbConn.Close()

	ctx := context.Background()

	v, err := svc.CreateVendor(ctx, &model.CreateVendorRequest{
		WorkspaceID: 1,
		Name:        "Temporary Supplier",
	})
	if err != nil {
		t.Fatalf("failed to create vendor: %v", err)
	}

	if err := svc.DeleteVendor(ctx, v.ID); err != nil {
		t.Fatalf("failed to delete vendor: %v", err)
	}

	_, err = svc.GetVendorByID(ctx, v.ID)
	if err != ErrVendorNotFound {
		t.Fatalf("expected ErrVendorNotFound, got: %v", err)
	}
}

func TestVendorContract_LifecycleAndRenewal(t *testing.T) {
	svc, _, dbConn := setupTestVendorSvc(t)
	defer dbConn.Close()

	ctx := context.Background()

	v, err := svc.CreateVendor(ctx, &model.CreateVendorRequest{
		WorkspaceID: 1,
		Name:        "Snowflake Cloud",
		Category:    model.VendorCategoryCloud,
	})
	if err != nil {
		t.Fatalf("failed to create vendor: %v", err)
	}

	startDate := time.Now().UTC().Truncate(time.Second)
	endDate := startDate.AddDate(1, 0, 0)

	// Create contract
	contract, err := svc.CreateContract(ctx, &model.CreateVendorContractRequest{
		WorkspaceID:         1,
		VendorID:            v.ID,
		ContractNumber:      "CTR-SNOW-2026",
		Title:               "Enterprise Data Cloud Agreement",
		ContractType:        model.ContractTypeMSA,
		Status:              model.ContractStatusActive,
		TotalValue:          120000.0,
		AnnualRecurringCost: 120000.0,
		Currency:            "USD",
		StartDate:           startDate,
		EndDate:             endDate,
		AutoRenew:           true,
		RenewalNoticeDays:   60,
		OwnerUserID:         "user_1",
	})
	if err != nil {
		t.Fatalf("failed to create contract: %v", err)
	}
	if contract.ID <= 0 {
		t.Fatalf("expected valid contract ID, got: %d", contract.ID)
	}
	if contract.ContractNumber != "CTR-SNOW-2026" {
		t.Fatalf("expected contract number CTR-SNOW-2026, got: %s", contract.ContractNumber)
	}

	// Fetch by number
	byNum, err := svc.GetContractByNumber(ctx, 1, "CTR-SNOW-2026")
	if err != nil {
		t.Fatalf("failed to get contract by number: %v", err)
	}
	if byNum.ID != contract.ID {
		t.Fatalf("expected contract ID %d, got: %d", contract.ID, byNum.ID)
	}

	// Renew contract
	newEnd := endDate.AddDate(1, 0, 0)
	newTotal := 250000.0
	renewed, err := svc.RenewContract(ctx, contract.ID, &model.RenewContractRequest{
		NewEndDate:    newEnd,
		NewTotalValue: &newTotal,
		Notes:         "Extended for second operating year with expanded compute credits",
	})
	if err != nil {
		t.Fatalf("failed to renew contract: %v", err)
	}
	if !renewed.EndDate.Equal(newEnd) {
		t.Fatalf("expected renewed end date %v, got: %v", newEnd, renewed.EndDate)
	}
	if renewed.TotalValue != 250000.0 {
		t.Fatalf("expected renewed total value 250000.0, got: %f", renewed.TotalValue)
	}

	// List contracts
	contracts, total, err := svc.ListContracts(ctx, model.ContractFilterCriteria{
		WorkspaceID: 1,
		VendorID:    v.ID,
	})
	if err != nil {
		t.Fatalf("failed to list contracts: %v", err)
	}
	if total != 1 || len(contracts) != 1 {
		t.Fatalf("expected 1 contract, got total=%d, len=%d", total, len(contracts))
	}
}

func TestVendor_Evaluations(t *testing.T) {
	svc, _, dbConn := setupTestVendorSvc(t)
	defer dbConn.Close()

	ctx := context.Background()

	v, err := svc.CreateVendor(ctx, &model.CreateVendorRequest{
		WorkspaceID: 1,
		Name:        "PagerDuty",
		Category:    model.VendorCategorySoftware,
	})
	if err != nil {
		t.Fatalf("failed to create vendor: %v", err)
	}

	// Create evaluation
	eval, err := svc.CreateEvaluation(ctx, v.ID, "user_1", &model.CreateVendorEvaluationRequest{
		EvaluationPeriod:    "2026-Q1",
		QualityScore:        4.5,
		ResponsivenessScore: 4.0,
		SLAScore:            5.0,
		PricingScore:        4.5,
		Feedback:            "Excellent alerting reliability and incident escalation response times",
	})
	if err != nil {
		t.Fatalf("failed to create evaluation: %v", err)
	}
	expectedOverall := (4.5 + 4.0 + 5.0 + 4.5) / 4.0
	if eval.OverallScore != expectedOverall {
		t.Fatalf("expected overall score %f, got: %f", expectedOverall, eval.OverallScore)
	}

	// List evaluations
	evals, err := svc.ListEvaluations(ctx, v.ID)
	if err != nil {
		t.Fatalf("failed to list evaluations: %v", err)
	}
	if len(evals) != 1 {
		t.Fatalf("expected 1 evaluation, got: %d", len(evals))
	}

	// Delete evaluation
	if err := svc.DeleteEvaluation(ctx, eval.ID); err != nil {
		t.Fatalf("failed to delete evaluation: %v", err)
	}
	evalsAfter, err := svc.ListEvaluations(ctx, v.ID)
	if err != nil {
		t.Fatalf("failed to list evaluations after delete: %v", err)
	}
	if len(evalsAfter) != 0 {
		t.Fatalf("expected 0 evaluations, got: %d", len(evalsAfter))
	}
}

func TestContract_Milestones(t *testing.T) {
	svc, _, dbConn := setupTestVendorSvc(t)
	defer dbConn.Close()

	ctx := context.Background()

	v, err := svc.CreateVendor(ctx, &model.CreateVendorRequest{
		WorkspaceID: 1,
		Name:        "Accenture Consulting",
		Category:    model.VendorCategoryConsulting,
	})
	if err != nil {
		t.Fatalf("failed to create vendor: %v", err)
	}

	startDate := time.Now().UTC()
	contract, err := svc.CreateContract(ctx, &model.CreateVendorContractRequest{
		WorkspaceID:  1,
		VendorID:     v.ID,
		Title:        "Cloud Migration Services",
		ContractType: model.ContractTypeSOW,
		StartDate:    startDate,
		EndDate:      startDate.AddDate(0, 6, 0),
		TotalValue:   75000.0,
	})
	if err != nil {
		t.Fatalf("failed to create contract: %v", err)
	}

	// Create milestone
	dueDate := startDate.AddDate(0, 2, 0)
	milestone, err := svc.CreateMilestone(ctx, contract.ID, &model.CreateContractMilestoneRequest{
		Title:       "Phase 1 Architecture Sign-off",
		Description: "Complete VPC architecture and security posture review",
		DueDate:     dueDate,
		Amount:      25000.0,
	})
	if err != nil {
		t.Fatalf("failed to create milestone: %v", err)
	}
	if milestone.Status != model.ContractMilestonePending {
		t.Fatalf("expected status pending, got: %s", milestone.Status)
	}

	// Update milestone to completed
	completedStatus := model.ContractMilestoneCompleted
	updated, err := svc.UpdateMilestone(ctx, milestone.ID, &model.UpdateContractMilestoneRequest{
		Status: &completedStatus,
	})
	if err != nil {
		t.Fatalf("failed to update milestone: %v", err)
	}
	if updated.Status != model.ContractMilestoneCompleted {
		t.Fatalf("expected status completed, got: %s", updated.Status)
	}
	if updated.CompletedAt == nil {
		t.Fatalf("expected CompletedAt to be automatically populated")
	}

	// List milestones
	milestones, err := svc.ListMilestones(ctx, contract.ID)
	if err != nil {
		t.Fatalf("failed to list milestones: %v", err)
	}
	if len(milestones) != 1 {
		t.Fatalf("expected 1 milestone, got: %d", len(milestones))
	}

	// Delete milestone
	if err := svc.DeleteMilestone(ctx, milestone.ID); err != nil {
		t.Fatalf("failed to delete milestone: %v", err)
	}
}

func TestVendor_CILinksAndMetrics(t *testing.T) {
	svc, _, dbConn := setupTestVendorSvc(t)
	defer dbConn.Close()

	ctx := context.Background()

	v, err := svc.CreateVendor(ctx, &model.CreateVendorRequest{
		WorkspaceID: 1,
		Name:        "Red Hat",
		Category:    model.VendorCategorySoftware,
		Tier:        model.VendorTierStrategic,
		Status:      model.VendorStatusActive,
		RiskLevel:   model.VendorRiskLow,
	})
	if err != nil {
		t.Fatalf("failed to create vendor: %v", err)
	}

	// Link CI 10
	link, err := svc.LinkCI(ctx, v.ID, &model.LinkVendorCIRequest{
		CIID:         10,
		Relationship: model.VendorCIRelSupports,
	})
	if err != nil {
		t.Fatalf("failed to link CI: %v", err)
	}
	if link.CIID != 10 {
		t.Fatalf("expected CIID 10, got: %d", link.CIID)
	}

	// List CI links
	links, err := svc.ListCILinks(ctx, v.ID)
	if err != nil {
		t.Fatalf("failed to list CI links: %v", err)
	}
	if len(links) != 1 {
		t.Fatalf("expected 1 link, got: %d", len(links))
	}
	if links[0].CIName != "Production Kubernetes Cluster" {
		t.Fatalf("expected CI name Production Kubernetes Cluster, got: %s", links[0].CIName)
	}

	// Unlink CI
	if err := svc.UnlinkCI(ctx, v.ID, 10); err != nil {
		t.Fatalf("failed to unlink CI: %v", err)
	}
	linksAfter, err := svc.ListCILinks(ctx, v.ID)
	if err != nil {
		t.Fatalf("failed to list CI links after unlink: %v", err)
	}
	if len(linksAfter) != 0 {
		t.Fatalf("expected 0 links after unlink, got: %d", len(linksAfter))
	}

	// Create contract for metrics calculation
	startDate := time.Now().UTC()
	_, err = svc.CreateContract(ctx, &model.CreateVendorContractRequest{
		WorkspaceID:         1,
		VendorID:            v.ID,
		Title:               "OpenShift Enterprise Support",
		ContractType:        model.ContractTypeMaintenance,
		Status:              model.ContractStatusActive,
		StartDate:           startDate,
		EndDate:             startDate.AddDate(1, 0, 0),
		TotalValue:          45000.0,
		AnnualRecurringCost: 45000.0,
	})
	if err != nil {
		t.Fatalf("failed to create contract: %v", err)
	}

	// Create evaluation for metrics calculation
	_, err = svc.CreateEvaluation(ctx, v.ID, "user_1", &model.CreateVendorEvaluationRequest{
		EvaluationPeriod:    "2026-Q1",
		QualityScore:        4.8,
		ResponsivenessScore: 4.8,
		SLAScore:            5.0,
		PricingScore:        4.6,
	})
	if err != nil {
		t.Fatalf("failed to create evaluation: %v", err)
	}

	// Verify metrics
	metrics, err := svc.GetVendorMetrics(ctx, 1)
	if err != nil {
		t.Fatalf("failed to get vendor metrics: %v", err)
	}
	if metrics.TotalVendors < 1 {
		t.Fatalf("expected at least 1 vendor in metrics, got: %d", metrics.TotalVendors)
	}
	if metrics.ActiveContracts < 1 {
		t.Fatalf("expected at least 1 active contract in metrics, got: %d", metrics.ActiveContracts)
	}
	if metrics.TotalAnnualSpend <= 0 {
		t.Fatalf("expected positive annual spend, got: %f", metrics.TotalAnnualSpend)
	}
	if metrics.AverageVendorScore <= 0 {
		t.Fatalf("expected positive average evaluation score, got: %f", metrics.AverageVendorScore)
	}
}
