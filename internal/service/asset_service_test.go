package service

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/JohnMalugu/tsk-mgr-api/internal/db"
	"github.com/JohnMalugu/tsk-mgr-api/internal/model"
	"github.com/JohnMalugu/tsk-mgr-api/internal/repository"
)

func setupTestAssetSvc(t *testing.T) (ITAssetService, repository.ITAssetRepository, *sql.DB) {
	sqliteDB, err := db.InitDB(":memory:")
	if err != nil {
		t.Fatalf("failed to initialize test in-memory db: %v", err)
	}

	_, err = sqliteDB.Exec(`
		INSERT INTO users (id, username, email, password_hash, created_at, updated_at) 
		VALUES ('user_1', 'alice', 'alice@example.com', 'hash', datetime('now'), datetime('now')),
		       ('user_2', 'bob', 'bob@example.com', 'hash', datetime('now'), datetime('now'));

		INSERT INTO workspaces (id, name, slug, owner_id, created_at, updated_at) 
		VALUES (1, 'Engineering WS', 'eng-ws-asset', 'user_1', datetime('now'), datetime('now'));

		INSERT INTO configuration_items (id, workspace_id, ci_type, name, ci_code, criticality, status, environment, owner_user_id, created_at, updated_at)
		VALUES (101, 1, 'server', 'Prod-App-Server-01', 'SRV-01', 'tier_1_mission_critical', 'active', 'production', 'user_1', datetime('now'), datetime('now'));
	`)
	if err != nil {
		t.Fatalf("failed to insert test seed data: %v", err)
	}

	repo := repository.NewITAssetRepository(sqliteDB)
	svc := NewITAssetService(repo)
	return svc, repo, sqliteDB
}

func TestAsset_CreateAndGet(t *testing.T) {
	svc, _, _ := setupTestAssetSvc(t)
	ctx := context.Background()

	pDate := time.Now().UTC().AddDate(-1, 0, 0) // 12 months ago
	wDate := time.Now().UTC().AddDate(2, 0, 0)  // 24 months future

	asset, err := svc.CreateAsset(ctx, &model.CreateAssetRequest{
		WorkspaceID:        1,
		Name:               "MacBook Pro 16",
		Category:           model.AssetCategoryLaptop,
		Make:               "Apple",
		Model:              "M3 Max",
		SerialNumber:       "C02G12345678",
		PurchaseDate:       &pDate,
		PurchaseCost:       3600.0,
		ResidualValue:      600.0,
		UsefulLifeMonths:   36,
		WarrantyExpiry:     &wDate,
		DepreciationMethod: model.DepreciationStraightLine,
		Location:           "HQ - 4th Floor",
	})
	if err != nil {
		t.Fatalf("failed to create asset: %v", err)
	}

	if asset.ID <= 0 {
		t.Errorf("expected positive asset ID, got %d", asset.ID)
	}
	if asset.AssetTag == "" {
		t.Errorf("expected auto-generated asset tag, got empty")
	}
	if asset.Category != model.AssetCategoryLaptop {
		t.Errorf("expected category laptop, got %s", asset.Category)
	}
	if asset.Status != model.AssetStatusInStock {
		t.Errorf("expected status in_stock, got %s", asset.Status)
	}
	if asset.AgeMonths < 11 || asset.AgeMonths > 13 {
		t.Errorf("expected age ~12 months, got %d", asset.AgeMonths)
	}
	if asset.CurrentBookValue >= 3600.0 || asset.CurrentBookValue <= 600.0 {
		t.Errorf("expected book value between 600 and 3600, got %f", asset.CurrentBookValue)
	}

	// Lookup by ID
	fetched, err := svc.GetAssetByID(ctx, asset.ID)
	if err != nil {
		t.Fatalf("failed to get asset by ID: %v", err)
	}
	if fetched.Name != asset.Name || fetched.AssetTag != asset.AssetTag {
		t.Errorf("fetched asset mismatch: expected %s, got %s", asset.Name, fetched.Name)
	}

	// Lookup by Tag
	byTag, err := svc.GetAssetByTag(ctx, 1, asset.AssetTag)
	if err != nil {
		t.Fatalf("failed to get asset by tag: %v", err)
	}
	if byTag.ID != asset.ID {
		t.Errorf("asset ID mismatch for tag query: expected %d, got %d", asset.ID, byTag.ID)
	}

	// Tag conflict check
	_, err = svc.CreateAsset(ctx, &model.CreateAssetRequest{
		WorkspaceID: 1,
		AssetTag:    asset.AssetTag,
		Name:        "Duplicate Tag Laptop",
		Category:    model.AssetCategoryLaptop,
	})
	if !errors.Is(err, ErrAssetTagConflict) {
		t.Errorf("expected ErrAssetTagConflict, got %v", err)
	}
}

func TestAsset_UpdateAndList(t *testing.T) {
	svc, _, _ := setupTestAssetSvc(t)
	ctx := context.Background()

	asset, err := svc.CreateAsset(ctx, &model.CreateAssetRequest{
		WorkspaceID: 1,
		Name:        "Dell PowerEdge R750",
		Category:    model.AssetCategoryServer,
		Make:        "Dell",
		Model:       "PowerEdge R750",
		Location:    "DC-East-Rack-12",
	})
	if err != nil {
		t.Fatalf("failed to create asset: %v", err)
	}

	newName := "Dell PowerEdge R750 Rack 12"
	newLocation := "DC-West-Rack-04"
	newCost := 8500.0

	updated, err := svc.UpdateAsset(ctx, asset.ID, &model.UpdateAssetRequest{
		Name:         &newName,
		Location:     &newLocation,
		PurchaseCost: &newCost,
	})
	if err != nil {
		t.Fatalf("failed to update asset: %v", err)
	}
	if updated.Name != newName || updated.Location != newLocation || updated.PurchaseCost != newCost {
		t.Errorf("update failed: name=%s, loc=%s, cost=%f", updated.Name, updated.Location, updated.PurchaseCost)
	}

	// List assets with filtering
	list, total, err := svc.ListAssets(ctx, model.AssetFilterCriteria{
		WorkspaceID: 1,
		Category:    model.AssetCategoryServer,
		Search:      "PowerEdge",
	})
	if err != nil {
		t.Fatalf("failed to list assets: %v", err)
	}
	if total != 1 || len(list) != 1 {
		t.Errorf("expected 1 server asset in list, got total=%d len=%d", total, len(list))
	}
}

func TestAsset_AssignmentAndLifecycle(t *testing.T) {
	svc, _, _ := setupTestAssetSvc(t)
	ctx := context.Background()

	pDate := time.Now().UTC().AddDate(0, -6, 0) // 6 months ago
	asset, err := svc.CreateAsset(ctx, &model.CreateAssetRequest{
		WorkspaceID:        1,
		Name:               "ThinkPad X1 Carbon",
		Category:           model.AssetCategoryLaptop,
		PurchaseCost:       2400.0,
		ResidualValue:      400.0,
		UsefulLifeMonths:   24,
		PurchaseDate:       &pDate,
		DepreciationMethod: model.DepreciationStraightLine,
	})
	if err != nil {
		t.Fatalf("failed to create asset: %v", err)
	}

	// Assign asset to user_1
	targetUser := "user_1"
	assigned, err := svc.AssignAsset(ctx, asset.ID, &model.AssignAssetRequest{
		UserID:     &targetUser,
		Department: "Platform Engineering",
		Location:   "Building B - Room 201",
		Notes:      "Assigned for remote onboarding",
	})
	if err != nil {
		t.Fatalf("failed to assign asset: %v", err)
	}
	if assigned.Status != model.AssetStatusInUse {
		t.Errorf("expected status in_use, got %s", assigned.Status)
	}
	if assigned.AssignedToUserID == nil || *assigned.AssignedToUserID != targetUser {
		t.Errorf("expected assigned user_1, got %v", assigned.AssignedToUserID)
	}
	if assigned.AssignedUserName != "alice" {
		t.Errorf("expected enriched username alice, got %s", assigned.AssignedUserName)
	}

	// Checkin asset
	checkedIn, err := svc.CheckinAsset(ctx, asset.ID, "Storage Locker 3", "Returned after hardware upgrade")
	if err != nil {
		t.Fatalf("failed to checkin asset: %v", err)
	}
	if checkedIn.Status != model.AssetStatusInStock {
		t.Errorf("expected status in_stock after checkin, got %s", checkedIn.Status)
	}
	if checkedIn.AssignedToUserID != nil {
		t.Errorf("expected assigned user to be cleared, got %v", checkedIn.AssignedToUserID)
	}

	// Transition status
	repaired, err := svc.TransitionAssetStatus(ctx, asset.ID, &model.AssetStatusTransitionRequest{
		Status: model.AssetStatusInRepair,
		Notes:  "Sent to vendor for battery replacement",
	})
	if err != nil {
		t.Fatalf("failed to transition status: %v", err)
	}
	if repaired.Status != model.AssetStatusInRepair {
		t.Errorf("expected status in_repair, got %s", repaired.Status)
	}

	// Calculate depreciation
	sched, err := svc.CalculateDepreciation(ctx, asset.ID)
	if err != nil {
		t.Fatalf("failed to calculate depreciation: %v", err)
	}
	if sched.AgeMonths < 5 || sched.AgeMonths > 7 {
		t.Errorf("expected age ~6 months, got %d", sched.AgeMonths)
	}
	expectedMonthly := (2400.0 - 400.0) / 24.0 // ~83.33
	if sched.MonthlyDepreciation < 80.0 || sched.MonthlyDepreciation > 85.0 {
		t.Errorf("unexpected monthly depreciation: %f (expected ~%f)", sched.MonthlyDepreciation, expectedMonthly)
	}
	if sched.CurrentBookValue >= 2400.0 || sched.CurrentBookValue <= 400.0 {
		t.Errorf("unexpected book value: %f", sched.CurrentBookValue)
	}
}

func TestAsset_MaintenanceRecords(t *testing.T) {
	svc, _, _ := setupTestAssetSvc(t)
	ctx := context.Background()

	asset, err := svc.CreateAsset(ctx, &model.CreateAssetRequest{
		WorkspaceID: 1,
		Name:        "Cisco Catalyst Switch 9300",
		Category:    model.AssetCategoryNetwork,
	})
	if err != nil {
		t.Fatalf("failed to create asset: %v", err)
	}

	schedDate := time.Now().UTC().AddDate(0, 0, 7)
	rec, err := svc.CreateMaintenanceRecord(ctx, asset.ID, &model.CreateMaintenanceRecordRequest{
		MaintenanceType: model.MaintTypeUpgrade,
		Cost:            450.0,
		PerformedBy:     "NetOps Team",
		Description:     "Firmware upgrade to IOS-XE 17.9",
		ScheduledDate:   &schedDate,
	})
	if err != nil {
		t.Fatalf("failed to create maintenance record: %v", err)
	}
	if rec.ID <= 0 || rec.Status != model.MaintRecordScheduled {
		t.Errorf("unexpected maintenance record created: id=%d status=%s", rec.ID, rec.Status)
	}

	// Update maintenance to completed
	compStatus := model.MaintRecordCompleted
	compDate := time.Now().UTC()
	updatedRec, err := svc.UpdateMaintenanceRecord(ctx, rec.ID, &model.UpdateMaintenanceRecordRequest{
		Status:        &compStatus,
		CompletedDate: &compDate,
	})
	if err != nil {
		t.Fatalf("failed to update maintenance record: %v", err)
	}
	if updatedRec.Status != model.MaintRecordCompleted {
		t.Errorf("expected status completed, got %s", updatedRec.Status)
	}

	// List maintenance records
	records, err := svc.ListMaintenanceRecords(ctx, asset.ID)
	if err != nil {
		t.Fatalf("failed to list maintenance records: %v", err)
	}
	if len(records) != 1 {
		t.Errorf("expected 1 maintenance record, got %d", len(records))
	}

	// Delete maintenance record
	if err := svc.DeleteMaintenanceRecord(ctx, rec.ID); err != nil {
		t.Fatalf("failed to delete maintenance record: %v", err)
	}
	recordsAfter, err := svc.ListMaintenanceRecords(ctx, asset.ID)
	if err != nil {
		t.Fatalf("failed to list records after delete: %v", err)
	}
	if len(recordsAfter) != 0 {
		t.Errorf("expected 0 records after delete, got %d", len(recordsAfter))
	}
}

func TestAsset_SoftwareLicensesAndAllocations(t *testing.T) {
	svc, _, _ := setupTestAssetSvc(t)
	ctx := context.Background()

	// 1. Create software license with 2 seats
	expDate := time.Now().UTC().AddDate(1, 0, 0)
	lic, err := svc.CreateSoftwareLicense(ctx, &model.CreateSoftwareLicenseRequest{
		WorkspaceID:    1,
		SoftwareName:   "IntelliJ IDEA Ultimate",
		Publisher:      "JetBrains",
		LicenseKey:     "JB-ULT-98765-XYZ",
		LicenseType:    model.LicenseTypeSeatBased,
		TotalSeats:     2,
		PurchaseCost:   1200.0,
		ExpirationDate: &expDate,
	})
	if err != nil {
		t.Fatalf("failed to create software license: %v", err)
	}
	if lic.TotalSeats != 2 || lic.AvailableSeats != 2 || lic.AllocatedSeats != 0 {
		t.Errorf("unexpected license seat counts: total=%d avail=%d alloc=%d", lic.TotalSeats, lic.AvailableSeats, lic.AllocatedSeats)
	}

	// Create test asset for allocation
	asset, err := svc.CreateAsset(ctx, &model.CreateAssetRequest{
		WorkspaceID: 1,
		Name:        "Dev Workstation 01",
		Category:    model.AssetCategoryDesktop,
	})
	if err != nil {
		t.Fatalf("failed to create asset: %v", err)
	}

	// 2. Allocate seat 1 to asset
	alloc1, err := svc.AllocateLicense(ctx, lic.ID, &model.AllocateLicenseRequest{
		AssetID: &asset.ID,
		Notes:   "Assigned to developer workstation",
	}, "admin")
	if err != nil {
		t.Fatalf("failed to allocate seat 1: %v", err)
	}
	if alloc1.ID <= 0 {
		t.Errorf("expected positive allocation ID, got %d", alloc1.ID)
	}

	// 3. Allocate seat 2 to user_1
	targetUser := "user_1"
	alloc2, err := svc.AllocateLicense(ctx, lic.ID, &model.AllocateLicenseRequest{
		UserID: &targetUser,
		Notes:  "Assigned to Alice",
	}, "admin")
	if err != nil {
		t.Fatalf("failed to allocate seat 2: %v", err)
	}
	if alloc2.ID <= 0 {
		t.Errorf("expected positive allocation ID, got %d", alloc2.ID)
	}

	// 4. Third allocation should fail with ErrNoSeatsAvailable
	targetUser2 := "user_2"
	_, err = svc.AllocateLicense(ctx, lic.ID, &model.AllocateLicenseRequest{
		UserID: &targetUser2,
	}, "admin")
	if !errors.Is(err, ErrNoSeatsAvailable) {
		t.Errorf("expected ErrNoSeatsAvailable, got %v", err)
	}

	// Check updated license seat counts
	licUpdated, err := svc.GetSoftwareLicenseByID(ctx, lic.ID)
	if err != nil {
		t.Fatalf("failed to get license: %v", err)
	}
	if licUpdated.AllocatedSeats != 2 || licUpdated.AvailableSeats != 0 {
		t.Errorf("expected 2 allocated and 0 available, got alloc=%d avail=%d", licUpdated.AllocatedSeats, licUpdated.AvailableSeats)
	}

	// 5. Deallocate seat 1
	if err := svc.DeallocateLicense(ctx, alloc1.ID); err != nil {
		t.Fatalf("failed to deallocate seat: %v", err)
	}

	licAfterDealloc, err := svc.GetSoftwareLicenseByID(ctx, lic.ID)
	if err != nil {
		t.Fatalf("failed to get license after deallocation: %v", err)
	}
	if licAfterDealloc.AllocatedSeats != 1 || licAfterDealloc.AvailableSeats != 1 {
		t.Errorf("expected 1 allocated and 1 available, got alloc=%d avail=%d", licAfterDealloc.AllocatedSeats, licAfterDealloc.AvailableSeats)
	}
}

func TestAsset_AuditScanAndMetrics(t *testing.T) {
	svc, _, _ := setupTestAssetSvc(t)
	ctx := context.Background()

	pDate := time.Now().UTC().AddDate(-1, 0, 0)
	wDate := time.Now().UTC().AddDate(0, 0, 15) // expires in 15 days

	asset, err := svc.CreateAsset(ctx, &model.CreateAssetRequest{
		WorkspaceID:    1,
		Name:           "Server Alpha",
		Category:       model.AssetCategoryServer,
		PurchaseCost:   5000.0,
		ResidualValue:  1000.0,
		PurchaseDate:   &pDate,
		WarrantyExpiry: &wDate,
	})
	if err != nil {
		t.Fatalf("failed to create asset: %v", err)
	}

	// Record audit scan
	scan, err := svc.RecordAuditScan(ctx, 1, "user_1", &model.AssetAuditScanRequest{
		AssetTag:    asset.AssetTag,
		AuditStatus: model.AuditStatusVerified,
		Location:    "DC-East-Row-4",
		Notes:       "Physical barcode scanned during Q4 audit",
	})
	if err != nil {
		t.Fatalf("failed to record audit scan: %v", err)
	}
	if scan.ID <= 0 || scan.AuditStatus != model.AuditStatusVerified {
		t.Errorf("unexpected audit scan: id=%d status=%s", scan.ID, scan.AuditStatus)
	}

	// List audit logs
	logs, err := svc.ListAuditLogs(ctx, 1, &asset.ID, 10, 0)
	if err != nil {
		t.Fatalf("failed to list audit logs: %v", err)
	}
	if len(logs) != 1 {
		t.Errorf("expected 1 audit log, got %d", len(logs))
	}

	// Metrics
	metrics, err := svc.GetAssetMetrics(ctx, 1)
	if err != nil {
		t.Fatalf("failed to get asset metrics: %v", err)
	}
	if metrics.TotalAssets < 1 {
		t.Errorf("expected >= 1 asset in metrics, got %d", metrics.TotalAssets)
	}
	if metrics.TotalPurchaseValue < 5000.0 {
		t.Errorf("expected >= 5000 purchase value, got %f", metrics.TotalPurchaseValue)
	}
	if metrics.WarrantyExpiringIn30Days < 1 {
		t.Errorf("expected >= 1 warranty expiring in 30 days, got %d", metrics.WarrantyExpiringIn30Days)
	}
	if metrics.CategoryBreakdown[model.AssetCategoryServer] < 1 {
		t.Errorf("expected server category count >= 1, got %d", metrics.CategoryBreakdown[model.AssetCategoryServer])
	}
}
