package service

import (
	"context"
	"database/sql"
	"sync"
	"testing"
	"time"

	"github.com/JohnMalugu/tsk-mgr-api/internal/db"
	"github.com/JohnMalugu/tsk-mgr-api/internal/model"
	"github.com/JohnMalugu/tsk-mgr-api/internal/repository"
)

func setupTestCapacitySvc(t *testing.T) (CapacityService, repository.CapacityRepository, *sql.DB) {
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
		VALUES (1, 'Capacity WS', 'capacity-ws', 'user_1', datetime('now'), datetime('now'));
	`)
	if err != nil {
		t.Fatalf("failed to insert test seed data: %v", err)
	}

	repo := repository.NewCapacityRepository(sqliteDB)
	svc := NewCapacityService(repo)
	return svc, repo, sqliteDB
}

func TestCapacity_PoolLifecycle(t *testing.T) {
	svc, _, dbConn := setupTestCapacitySvc(t)
	defer dbConn.Close()

	ctx := context.Background()

	// 1. Validation failure: missing name
	_, err := svc.CreatePool(ctx, "ws_1", model.CreateCapacityPoolRequest{
		ResourceType:  model.ResourceTypeComputeCPU,
		TotalCapacity: 64.0,
	})
	if err != ErrCapacityPoolNameRequired {
		t.Fatalf("expected ErrCapacityPoolNameRequired, got %v", err)
	}

	// 2. Validation failure: invalid resource type
	_, err = svc.CreatePool(ctx, "ws_1", model.CreateCapacityPoolRequest{
		Name:          "Primary Compute",
		ResourceType:  "invalid_type",
		TotalCapacity: 64.0,
	})
	if err != ErrInvalidResourceType {
		t.Fatalf("expected ErrInvalidResourceType, got %v", err)
	}

	// 3. Validation failure: invalid total capacity
	_, err = svc.CreatePool(ctx, "ws_1", model.CreateCapacityPoolRequest{
		Name:          "Primary Compute",
		ResourceType:  model.ResourceTypeComputeCPU,
		TotalCapacity: -5.0,
	})
	if err != ErrInvalidTotalCapacity {
		t.Fatalf("expected ErrInvalidTotalCapacity, got %v", err)
	}

	// 4. Successful pool creation
	pool, err := svc.CreatePool(ctx, "ws_1", model.CreateCapacityPoolRequest{
		Name:                     "K8s Production Compute Cluster",
		ResourceType:             model.ResourceTypeComputeCPU,
		CIID:                     "ci_k8s_prod",
		TotalCapacity:            128.0,
		AllocatedCapacity:        96.0,
		UtilizedCapacity:         32.0,
		WarningThresholdPercent:  75.0,
		CriticalThresholdPercent: 90.0,
		Description:              "Production Kubernetes node CPU capacity pool",
	})
	if err != nil {
		t.Fatalf("failed to create pool: %v", err)
	}
	if pool.ID == "" || pool.Name != "K8s Production Compute Cluster" || pool.Status != model.PoolStatusOptimal {
		t.Fatalf("unexpected pool fields: %+v", pool)
	}

	// 5. Unique name constraint in same workspace
	_, err = svc.CreatePool(ctx, "ws_1", model.CreateCapacityPoolRequest{
		Name:          "K8s Production Compute Cluster",
		ResourceType:  model.ResourceTypeComputeCPU,
		TotalCapacity: 64.0,
	})
	if err != ErrCapacityPoolNameConflict {
		t.Fatalf("expected ErrCapacityPoolNameConflict, got %v", err)
	}

	// 6. GetPoolByID & GetPoolByName
	fetchedByID, err := svc.GetPoolByID(ctx, pool.ID)
	if err != nil {
		t.Fatalf("failed to get pool by id: %v", err)
	}
	if fetchedByID.CIID != "ci_k8s_prod" {
		t.Fatalf("expected CIID ci_k8s_prod, got %s", fetchedByID.CIID)
	}

	fetchedByName, err := svc.GetPoolByName(ctx, "ws_1", "K8s Production Compute Cluster")
	if err != nil {
		t.Fatalf("failed to get pool by name: %v", err)
	}
	if fetchedByName.ID != pool.ID {
		t.Fatalf("mismatched pool id by name")
	}

	// 7. Update pool: increase utilized capacity into warning zone (100 / 128 = 78.1%)
	newUtil := 100.0
	updated, err := svc.UpdatePool(ctx, pool.ID, model.UpdateCapacityPoolRequest{
		UtilizedCapacity: &newUtil,
	})
	if err != nil {
		t.Fatalf("failed to update pool: %v", err)
	}
	if updated.UtilizedCapacity != 100.0 || updated.Status != model.PoolStatusWarning {
		t.Fatalf("expected warning status after update, got %+v", updated)
	}

	// 8. List pools
	pools, total, err := svc.ListPools(ctx, model.CapacityPoolFilterCriteria{
		WorkspaceID:  "ws_1",
		ResourceType: model.ResourceTypeComputeCPU,
	})
	if err != nil {
		t.Fatalf("failed to list pools: %v", err)
	}
	if total != 1 || len(pools) != 1 {
		t.Fatalf("expected 1 pool, got total=%d len=%d", total, len(pools))
	}

	// 9. Delete pool
	if err := svc.DeletePool(ctx, pool.ID); err != nil {
		t.Fatalf("failed to delete pool: %v", err)
	}
	_, err = svc.GetPoolByID(ctx, pool.ID)
	if err != ErrCapacityPoolNotFound {
		t.Fatalf("expected ErrCapacityPoolNotFound, got %v", err)
	}
}

func TestCapacity_MetricsAndTelemetry(t *testing.T) {
	svc, _, dbConn := setupTestCapacitySvc(t)
	defer dbConn.Close()

	ctx := context.Background()

	// Pool setup
	pool, err := svc.CreatePool(ctx, "ws_1", model.CreateCapacityPoolRequest{
		Name:          "Database RAM Cache Pool",
		ResourceType:  model.ResourceTypeMemoryRAM,
		TotalCapacity: 256.0,
	})
	if err != nil {
		t.Fatalf("failed to create pool: %v", err)
	}

	// 1. Validation failure: missing metric name
	_, err = svc.RecordMetric(ctx, "ws_1", model.RecordCapacityMetricRequest{
		PoolID:       pool.ID,
		AverageValue: 128.0,
	})
	if err == nil {
		t.Fatalf("expected validation error for missing metric_name")
	}

	// 2. Successful metric ingestion
	now := time.Now().UTC()
	metric1, err := svc.RecordMetric(ctx, "ws_1", model.RecordCapacityMetricRequest{
		PoolID:       pool.ID,
		MetricName:   "memory_ram_utilization",
		SampledAt:    &now,
		AverageValue: 180.0,
		PeakValue:    210.0,
		P95Value:     200.0,
		P99Value:     208.0,
	})
	if err != nil {
		t.Fatalf("failed to record metric: %v", err)
	}
	if metric1.ID == "" || metric1.AverageValue != 180.0 {
		t.Fatalf("unexpected metric fields: %+v", metric1)
	}

	// Check that pool utilized capacity automatically updated to 180 GB
	updatedPool, err := svc.GetPoolByID(ctx, pool.ID)
	if err != nil {
		t.Fatalf("failed to get pool: %v", err)
	}
	if updatedPool.UtilizedCapacity != 180.0 {
		t.Fatalf("expected utilized capacity 180.0, got %f", updatedPool.UtilizedCapacity)
	}

	// 3. Record anomaly metric
	metric2, err := svc.RecordMetric(ctx, "ws_1", model.RecordCapacityMetricRequest{
		PoolID:       pool.ID,
		MetricName:   "memory_spike_anomaly",
		SampledAt:    &now,
		AverageValue: 250.0,
		PeakValue:    255.0,
		IsAnomaly:    true,
		AnomalyScore: 0.98,
	})
	if err != nil {
		t.Fatalf("failed to record anomaly metric: %v", err)
	}
	if !metric2.IsAnomaly || metric2.AnomalyScore != 0.98 {
		t.Fatalf("unexpected anomaly metric: %+v", metric2)
	}

	// 4. GetMetricByID & ListMetrics
	fetched, err := svc.GetMetricByID(ctx, metric1.ID)
	if err != nil {
		t.Fatalf("failed to get metric: %v", err)
	}
	if fetched.MetricName != "memory_ram_utilization" {
		t.Fatalf("metric name mismatch: %s", fetched.MetricName)
	}

	isAnomaly := true
	anomalyList, totalAnomalies, err := svc.ListMetrics(ctx, model.CapacityMetricFilterCriteria{
		WorkspaceID: "ws_1",
		PoolID:      pool.ID,
		IsAnomaly:   &isAnomaly,
	})
	if err != nil {
		t.Fatalf("failed to list anomaly metrics: %v", err)
	}
	if totalAnomalies != 1 || len(anomalyList) != 1 {
		t.Fatalf("expected 1 anomaly metric, got total=%d len=%d", totalAnomalies, len(anomalyList))
	}
}

func TestCapacity_Forecasting(t *testing.T) {
	svc, _, dbConn := setupTestCapacitySvc(t)
	defer dbConn.Close()

	ctx := context.Background()

	// High utilization pool (850 / 1000 = 85%)
	pool, err := svc.CreatePool(ctx, "ws_1", model.CreateCapacityPoolRequest{
		Name:                     "NVMe Storage Cluster",
		ResourceType:             model.ResourceTypeStorageDisk,
		TotalCapacity:            1000.0,
		UtilizedCapacity:         850.0,
		WarningThresholdPercent:  75.0,
		CriticalThresholdPercent: 90.0,
	})
	if err != nil {
		t.Fatalf("failed to create pool: %v", err)
	}

	// Generate forecast for 90 days
	forecast, err := svc.GenerateForecast(ctx, "ws_1", pool.ID, 90)
	if err != nil {
		t.Fatalf("failed to generate forecast: %v", err)
	}
	if forecast.CurrentUtilizationPercent != 85.0 {
		t.Fatalf("expected current utilization 85.0, got %f", forecast.CurrentUtilizationPercent)
	}
	if forecast.DaysToExhaustion <= 0 {
		t.Fatalf("expected positive days to exhaustion under growth, got %d", forecast.DaysToExhaustion)
	}
	if forecast.RecommendedAction != model.ActionScaleUp && forecast.RecommendedAction != model.ActionScaleOut {
		t.Fatalf("expected scale action, got %s", forecast.RecommendedAction)
	}

	// Get latest forecast
	latest, err := svc.GetLatestForecast(ctx, pool.ID)
	if err != nil {
		t.Fatalf("failed to get latest forecast: %v", err)
	}
	if latest.ID != forecast.ID {
		t.Fatalf("forecast id mismatch")
	}
}

func TestCapacity_PlansAndBottlenecks(t *testing.T) {
	svc, _, dbConn := setupTestCapacitySvc(t)
	defer dbConn.Close()

	ctx := context.Background()

	pool, err := svc.CreatePool(ctx, "ws_1", model.CreateCapacityPoolRequest{
		Name:          "Database Cluster IOPS",
		ResourceType:  model.ResourceTypeDatabaseIOPS,
		TotalCapacity: 10000.0,
	})
	if err != nil {
		t.Fatalf("failed to create pool: %v", err)
	}

	// 1. Capacity Plan Creation
	plan, err := svc.CreatePlan(ctx, "ws_1", model.CreateCapacityPlanRequest{
		PoolID:           pool.ID,
		Title:            "Q3 Aurora IOPS Scaling",
		PlanningHorizon:  model.PlanHorizonMediumTerm,
		RequiredCapacity: 25000.0,
		EstimatedCost:    7500.0,
		Currency:         "USD",
		Justification:    "Anticipated peak holiday traffic workload",
	})
	if err != nil {
		t.Fatalf("failed to create plan: %v", err)
	}
	if plan.ID == "" || plan.Status != model.PlanStatusDraft || plan.CurrentCapacity != 10000.0 {
		t.Fatalf("unexpected plan fields: %+v", plan)
	}

	// 2. Approve Plan
	approved, err := svc.ApprovePlan(ctx, plan.ID, "user_1")
	if err != nil {
		t.Fatalf("failed to approve plan: %v", err)
	}
	if approved.Status != model.PlanStatusApproved || approved.ApprovedBy != "user_1" {
		t.Fatalf("plan approval mismatch: %+v", approved)
	}

	// 3. List Plans
	plans, totalPlans, err := svc.ListPlans(ctx, model.CapacityPlanFilterCriteria{
		WorkspaceID: "ws_1",
		Status:      model.PlanStatusApproved,
	})
	if err != nil {
		t.Fatalf("failed to list plans: %v", err)
	}
	if totalPlans != 1 || len(plans) != 1 {
		t.Fatalf("expected 1 approved plan, got total=%d len=%d", totalPlans, len(plans))
	}

	// 4. Bottleneck Creation
	bottleneck, err := svc.CreateBottleneck(ctx, "ws_1", model.CreateBottleneckRequest{
		PoolID:            pool.ID,
		AffectedCIID:      "ci_aurora_primary",
		Title:             "Disk IOPS Queue Saturation",
		Description:       "Write queues elevated past 95% during batch ETL",
		Severity:          model.BottleneckSeverityCritical,
		SaturationPercent: 96.5,
		MitigationPlan:    "Scale cluster storage class to Provisioned IOPS",
	})
	if err != nil {
		t.Fatalf("failed to create bottleneck: %v", err)
	}
	if bottleneck.ID == "" || bottleneck.Status != model.BottleneckStatusOpen {
		t.Fatalf("unexpected bottleneck fields: %+v", bottleneck)
	}

	// 5. Resolve Bottleneck
	resolved, err := svc.ResolveBottleneck(ctx, bottleneck.ID, "ETL batch window optimized and storage class upgraded")
	if err != nil {
		t.Fatalf("failed to resolve bottleneck: %v", err)
	}
	if resolved.Status != model.BottleneckStatusResolved || resolved.ResolvedAt == nil {
		t.Fatalf("bottleneck resolution mismatch: %+v", resolved)
	}

	// 6. List Bottlenecks
	bottlenecks, totalB, err := svc.ListBottlenecks(ctx, model.BottleneckFilterCriteria{
		WorkspaceID: "ws_1",
		PoolID:      pool.ID,
	})
	if err != nil {
		t.Fatalf("failed to list bottlenecks: %v", err)
	}
	if totalB != 1 || len(bottlenecks) != 1 {
		t.Fatalf("expected 1 bottleneck, got total=%d len=%d", totalB, len(bottlenecks))
	}
}

func TestCapacity_AnalyticsReport(t *testing.T) {
	svc, _, dbConn := setupTestCapacitySvc(t)
	defer dbConn.Close()

	ctx := context.Background()

	// 1. Optimal pool (20 / 100 = 20%)
	_, err := svc.CreatePool(ctx, "ws_1", model.CreateCapacityPoolRequest{
		Name:             "Optimal Pool",
		ResourceType:     model.ResourceTypeComputeCPU,
		TotalCapacity:    100.0,
		UtilizedCapacity: 20.0,
	})
	if err != nil {
		t.Fatalf("failed to create optimal pool: %v", err)
	}

	// 2. Warning pool (80 / 100 = 80%)
	_, err = svc.CreatePool(ctx, "ws_1", model.CreateCapacityPoolRequest{
		Name:             "Warning Pool",
		ResourceType:     model.ResourceTypeMemoryRAM,
		TotalCapacity:    100.0,
		UtilizedCapacity: 80.0,
	})
	if err != nil {
		t.Fatalf("failed to create warning pool: %v", err)
	}

	// 3. Analytics
	report, err := svc.GetCapacityAnalytics(ctx, "ws_1")
	if err != nil {
		t.Fatalf("failed to get analytics report: %v", err)
	}

	if report.TotalPools != 2 {
		t.Fatalf("expected 2 pools, got %d", report.TotalPools)
	}
	if report.OptimalPoolsCount != 1 || report.WarningPoolsCount != 1 {
		t.Fatalf("expected 1 optimal and 1 warning pool, got opt=%d warn=%d", report.OptimalPoolsCount, report.WarningPoolsCount)
	}
	if report.AverageUtilizationPercent != 50.0 {
		t.Fatalf("expected 50.0%% average utilization, got %f", report.AverageUtilizationPercent)
	}
	if report.CapacityHealthIndex <= 0 || report.CapacityHealthIndex > 100 {
		t.Fatalf("expected capacity health index between 0 and 100, got %f", report.CapacityHealthIndex)
	}
}

func TestCapacity_ConcurrencyAndRace(t *testing.T) {
	svc, _, dbConn := setupTestCapacitySvc(t)
	defer dbConn.Close()

	ctx := context.Background()

	pool, err := svc.CreatePool(ctx, "ws_1", model.CreateCapacityPoolRequest{
		Name:          "Concurrent Network Bandwidth Pool",
		ResourceType:  model.ResourceTypeNetworkBandwidth,
		TotalCapacity: 10000.0,
	})
	if err != nil {
		t.Fatalf("failed to create pool: %v", err)
	}

	var wg sync.WaitGroup
	numWorkers := 10

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			_, _ = svc.RecordMetric(ctx, "ws_1", model.RecordCapacityMetricRequest{
				PoolID:       pool.ID,
				MetricName:   "network_egress_mbps",
				AverageValue: float64(100 * (workerID + 1)),
				PeakValue:    float64(150 * (workerID + 1)),
			})
			_, _ = svc.GetCapacityAnalytics(ctx, "ws_1")
		}(i)
	}

	wg.Wait()

	metrics, total, err := svc.ListMetrics(ctx, model.CapacityMetricFilterCriteria{
		WorkspaceID: "ws_1",
		PoolID:      pool.ID,
	})
	if err != nil {
		t.Fatalf("failed to list metrics: %v", err)
	}
	if total != numWorkers || len(metrics) != numWorkers {
		t.Fatalf("expected %d metrics, got total=%d len=%d", numWorkers, total, len(metrics))
	}
}
