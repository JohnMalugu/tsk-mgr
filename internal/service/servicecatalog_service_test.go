package service

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/JohnMalugu/tsk-mgr-api/internal/db"
	"github.com/JohnMalugu/tsk-mgr-api/internal/model"
	"github.com/JohnMalugu/tsk-mgr-api/internal/repository"
)

func setupTestServiceCatalogSvc(t *testing.T) (ServiceCatalogService, repository.ServiceCatalogRepository, *sql.DB) {
	sqliteDB, err := db.InitDB(":memory:")
	if err != nil {
		t.Fatalf("failed to initialize test in-memory db: %v", err)
	}

	_, err = sqliteDB.Exec(`
		INSERT INTO users (id, username, email, password_hash, created_at, updated_at) 
		VALUES ('user_1', 'alice', 'alice@example.com', 'hash', datetime('now'), datetime('now')),
		       ('user_2', 'bob', 'bob@example.com', 'hash', datetime('now'), datetime('now'));

		INSERT INTO workspaces (id, name, slug, owner_id, created_at, updated_at) 
		VALUES (1, 'Engineering WS', 'eng-ws-catalog', 'user_1', datetime('now'), datetime('now'));

		INSERT INTO projects (id, user_id, name, description, color, created_at, updated_at)
		VALUES (101, 'user_1', 'IT Operations Backlog', 'IT Support and Fulfillment', '#4f46e5', datetime('now'), datetime('now'));
	`)
	if err != nil {
		t.Fatalf("failed to insert test seed data: %v", err)
	}

	catalogRepo := repository.NewServiceCatalogRepository(sqliteDB)
	taskRepo := repository.NewTaskRepository(sqliteDB)
	projectRepo := repository.NewProjectRepository(sqliteDB)
	cmdbRepo := repository.NewCMDBRepository(sqliteDB)

	svc := NewServiceCatalogService(catalogRepo, taskRepo, projectRepo, cmdbRepo)
	return svc, catalogRepo, sqliteDB
}

func TestServiceCatalogService_Categories(t *testing.T) {
	svc, _, _ := setupTestServiceCatalogSvc(t)

	// 1. Missing name validation
	_, err := svc.CreateCategory("user_1", model.CreateCatalogCategoryRequest{
		WorkspaceID: 1,
		Name:        "",
	})
	if !errors.Is(err, ErrCatalogCategoryNameReq) {
		t.Fatalf("expected ErrCatalogCategoryNameReq, got %v", err)
	}

	// 2. Successful creation
	cat, err := svc.CreateCategory("user_1", model.CreateCatalogCategoryRequest{
		WorkspaceID:  1,
		Name:         "Cloud Infrastructure",
		Slug:         "cloud-infra",
		Description:  "AWS and GCP cloud resource provisioning",
		Icon:         "cloud",
		DisplayOrder: 1,
	})
	if err != nil {
		t.Fatalf("unexpected error creating category: %v", err)
	}
	if cat.ID <= 0 {
		t.Fatalf("expected positive category ID, got %d", cat.ID)
	}
	if cat.Slug != "cloud-infra" {
		t.Fatalf("expected slug cloud-infra, got %s", cat.Slug)
	}

	// 3. Duplicate slug validation
	_, err = svc.CreateCategory("user_1", model.CreateCatalogCategoryRequest{
		WorkspaceID: 1,
		Name:        "Duplicate Cloud",
		Slug:        "cloud-infra",
	})
	if !errors.Is(err, ErrDuplicateCategorySlug) {
		t.Fatalf("expected ErrDuplicateCategorySlug, got %v", err)
	}

	// 4. Get category by ID
	fetched, err := svc.GetCategory(cat.ID)
	if err != nil {
		t.Fatalf("unexpected error fetching category by ID: %v", err)
	}
	if fetched.Name != "Cloud Infrastructure" {
		t.Fatalf("expected Cloud Infrastructure, got %s", fetched.Name)
	}

	// 5. Get category by slug
	fetchedBySlug, err := svc.GetCategoryBySlug(1, "cloud-infra")
	if err != nil {
		t.Fatalf("unexpected error fetching category by slug: %v", err)
	}
	if fetchedBySlug.ID != cat.ID {
		t.Fatalf("expected ID %d, got %d", cat.ID, fetchedBySlug.ID)
	}

	// 6. List categories
	cats, err := svc.ListCategories(model.CatalogCategoryFilter{
		WorkspaceID: 1,
	})
	if err != nil {
		t.Fatalf("unexpected error listing categories: %v", err)
	}
	if len(cats) != 1 {
		t.Fatalf("expected 1 category, got %d", len(cats))
	}

	// 7. Update category
	newName := "Cloud & Dedicated Infrastructure"
	updated, err := svc.UpdateCategory(cat.ID, model.UpdateCatalogCategoryRequest{
		Name: &newName,
	})
	if err != nil {
		t.Fatalf("unexpected error updating category: %v", err)
	}
	if updated.Name != "Cloud & Dedicated Infrastructure" {
		t.Fatalf("expected updated name, got %s", updated.Name)
	}

	// 8. Delete category
	if err := svc.DeleteCategory(cat.ID); err != nil {
		t.Fatalf("unexpected error deleting category: %v", err)
	}
	_, err = svc.GetCategory(cat.ID)
	if !errors.Is(err, ErrCatalogCategoryNotFound) {
		t.Fatalf("expected ErrCatalogCategoryNotFound, got %v", err)
	}
}

func TestServiceCatalogService_CatalogItems(t *testing.T) {
	svc, _, _ := setupTestServiceCatalogSvc(t)

	cat, err := svc.CreateCategory("user_1", model.CreateCatalogCategoryRequest{
		WorkspaceID: 1,
		Name:        "Access & Permissions",
		Slug:        "access-permissions",
	})
	if err != nil {
		t.Fatalf("failed to create category: %v", err)
	}

	// 1. Missing item name validation
	_, err = svc.CreateItem("user_1", model.CreateCatalogItemRequest{
		WorkspaceID: 1,
		CategoryID:  cat.ID,
		Name:        "",
	})
	if !errors.Is(err, ErrCatalogItemNameReq) {
		t.Fatalf("expected ErrCatalogItemNameReq, got %v", err)
	}

	// 2. Non-existent category
	_, err = svc.CreateItem("user_1", model.CreateCatalogItemRequest{
		WorkspaceID: 1,
		CategoryID:  9999,
		Name:        "Production Access",
	})
	if !errors.Is(err, ErrCatalogCategoryNotFound) {
		t.Fatalf("expected ErrCatalogCategoryNotFound, got %v", err)
	}

	// 3. Create item with form schema
	projID := int64(101)
	schema := &model.CatalogFormSchema{
		Fields: []model.CatalogFormField{
			{
				ID:       "env",
				Name:     "environment",
				Label:    "Target Environment",
				Type:     model.FieldTypeSelect,
				Required: true,
				Options:  []string{"staging", "production"},
			},
			{
				ID:       "justification",
				Name:     "business_justification",
				Label:    "Business Justification",
				Type:     model.FieldTypeTextArea,
				Required: true,
			},
		},
	}

	item, err := svc.CreateItem("user_1", model.CreateCatalogItemRequest{
		WorkspaceID:               1,
		CategoryID:                cat.ID,
		Name:                      "Production Database Access Grant",
		Slug:                      "prod-db-access",
		Description:               "Temporary elevated access to production database read-replicas",
		ShortDescription:          "Read-only production DB access",
		Icon:                      "database-lock",
		EstimatedFulfillmentHours: 2.0,
		CostEstimateCents:         0,
		Currency:                  "USD",
		RequiresApproval:          true,
		ApprovalRole:              model.ApprovalRoleSecurity,
		AutoCreateTask:            true,
		TargetProjectID:           &projID,
		FormSchema:                schema,
	})
	if err != nil {
		t.Fatalf("unexpected error creating item: %v", err)
	}
	if item.ID <= 0 {
		t.Fatalf("expected positive item ID, got %d", item.ID)
	}
	if !item.RequiresApproval || item.ApprovalRole != model.ApprovalRoleSecurity {
		t.Fatalf("expected security approval required, got %v, role %s", item.RequiresApproval, item.ApprovalRole)
	}

	// 4. Duplicate slug validation
	_, err = svc.CreateItem("user_1", model.CreateCatalogItemRequest{
		WorkspaceID: 1,
		CategoryID:  cat.ID,
		Name:        "Duplicate Prod DB Access",
		Slug:        "prod-db-access",
	})
	if !errors.Is(err, ErrDuplicateItemSlug) {
		t.Fatalf("expected ErrDuplicateItemSlug, got %v", err)
	}

	// 5. Get item by ID and verify form schema deserialization
	fetched, err := svc.GetItem(item.ID)
	if err != nil {
		t.Fatalf("unexpected error getting item: %v", err)
	}
	if fetched.FormSchema == nil || len(fetched.FormSchema.Fields) != 2 {
		t.Fatalf("expected 2 schema fields deserialized, got %v", fetched.FormSchema)
	}
	if fetched.CategoryName != "Access & Permissions" {
		t.Fatalf("expected category name Access & Permissions, got %s", fetched.CategoryName)
	}

	// 6. List items
	items, total, err := svc.ListItems(model.CatalogItemFilter{
		WorkspaceID: 1,
	})
	if err != nil {
		t.Fatalf("unexpected error listing items: %v", err)
	}
	if total != 1 || len(items) != 1 {
		t.Fatalf("expected 1 item, got total %d", total)
	}

	// 7. Update item
	newShort := "Elevated DB Read Access"
	updated, err := svc.UpdateItem(item.ID, model.UpdateCatalogItemRequest{
		ShortDescription: &newShort,
	})
	if err != nil {
		t.Fatalf("unexpected error updating item: %v", err)
	}
	if updated.ShortDescription != newShort {
		t.Fatalf("expected updated short description, got %s", updated.ShortDescription)
	}

	// 8. Delete item
	if err := svc.DeleteItem(item.ID); err != nil {
		t.Fatalf("unexpected error deleting item: %v", err)
	}
	_, err = svc.GetItem(item.ID)
	if !errors.Is(err, ErrCatalogItemNotFound) {
		t.Fatalf("expected ErrCatalogItemNotFound, got %v", err)
	}
}

func TestServiceCatalogService_ServiceRequests_NoApproval(t *testing.T) {
	svc, _, _ := setupTestServiceCatalogSvc(t)

	cat, err := svc.CreateCategory("user_1", model.CreateCatalogCategoryRequest{
		WorkspaceID: 1,
		Name:        "Developer Tools",
		Slug:        "dev-tools",
	})
	if err != nil {
		t.Fatalf("failed to create category: %v", err)
	}

	projID := int64(101)
	item, err := svc.CreateItem("user_1", model.CreateCatalogItemRequest{
		WorkspaceID:               1,
		CategoryID:                cat.ID,
		Name:                      "Developer Cloud Sandbox",
		Slug:                      "dev-cloud-sandbox",
		EstimatedFulfillmentHours: 1.0,
		CostEstimateCents:         2500, // $25.00
		RequiresApproval:          false,
		AutoCreateTask:            true,
		TargetProjectID:           &projID,
	})
	if err != nil {
		t.Fatalf("failed to create item: %v", err)
	}

	// 1. Submit service request
	req, err := svc.SubmitRequest("user_1", model.SubmitServiceRequest{
		CatalogItemID: item.ID,
		Title:         "Sandbox environment for feature prototype",
		Description:   "Need 2-node K8s sandbox cluster for load testing",
		Priority:      model.RequestPriorityHigh,
		FormData: map[string]interface{}{
			"cluster_size": 2,
			"region":       "us-east-1",
		},
	})
	if err != nil {
		t.Fatalf("unexpected error submitting request: %v", err)
	}
	if req.ID <= 0 {
		t.Fatalf("expected positive request ID, got %d", req.ID)
	}
	if req.RequestNumber != "REQ-0001" {
		t.Fatalf("expected request number REQ-0001, got %s", req.RequestNumber)
	}
	// Because AutoCreateTask was true and RequiresApproval was false, task should be generated and status in_progress
	if req.Status != model.RequestStatusInProgress {
		t.Fatalf("expected status in_progress, got %s", req.Status)
	}
	if req.TaskID == nil {
		t.Fatalf("expected generated TaskID, got nil")
	}

	// 2. Add comments
	comm1, err := svc.AddComment("user_1", req.ID, model.AddRequestCommentRequest{
		Comment:    "Please deploy with Helm 3 pre-installed.",
		IsInternal: false,
	})
	if err != nil {
		t.Fatalf("unexpected error adding comment: %v", err)
	}
	if comm1.Comment != "Please deploy with Helm 3 pre-installed." {
		t.Fatalf("unexpected comment content: %s", comm1.Comment)
	}

	// Add internal team note
	_, err = svc.AddComment("user_2", req.ID, model.AddRequestCommentRequest{
		Comment:    "Sandbox provisioning script executed successfully.",
		IsInternal: true,
	})
	if err != nil {
		t.Fatalf("unexpected error adding internal comment: %v", err)
	}

	comments, err := svc.ListComments(req.ID)
	if err != nil {
		t.Fatalf("unexpected error listing comments: %v", err)
	}
	if len(comments) != 2 {
		t.Fatalf("expected 2 comments, got %d", len(comments))
	}

	// 3. Complete fulfillment
	fulfilledStatus := model.RequestStatusFulfilled
	updatedReq, err := svc.UpdateRequest(req.ID, model.UpdateServiceRequest{
		Status: &fulfilledStatus,
	})
	if err != nil {
		t.Fatalf("unexpected error updating request: %v", err)
	}
	if updatedReq.Status != model.RequestStatusFulfilled {
		t.Fatalf("expected fulfilled status, got %s", updatedReq.Status)
	}
	if updatedReq.FulfilledAt == nil {
		t.Fatalf("expected fulfilled_at timestamp to be populated")
	}

	// 4. Get by request number
	byNum, err := svc.GetRequestByNumber(1, "REQ-0001")
	if err != nil {
		t.Fatalf("unexpected error getting request by number: %v", err)
	}
	if byNum.ID != req.ID {
		t.Fatalf("expected ID %d, got %d", req.ID, byNum.ID)
	}
}

func TestServiceCatalogService_ServiceRequests_WithApproval(t *testing.T) {
	svc, _, _ := setupTestServiceCatalogSvc(t)

	cat, err := svc.CreateCategory("user_1", model.CreateCatalogCategoryRequest{
		WorkspaceID: 1,
		Name:        "Hardware & Devices",
		Slug:        "hardware-devices",
	})
	if err != nil {
		t.Fatalf("failed to create category: %v", err)
	}

	projID := int64(101)
	item, err := svc.CreateItem("user_1", model.CreateCatalogItemRequest{
		WorkspaceID:       1,
		CategoryID:        cat.ID,
		Name:              "MacBook Pro 16-inch M3 Max",
		Slug:              "mbp-16-m3",
		CostEstimateCents: 349900,
		RequiresApproval:  true,
		ApprovalRole:      model.ApprovalRoleManager,
		AutoCreateTask:    true,
		TargetProjectID:   &projID,
	})
	if err != nil {
		t.Fatalf("failed to create item: %v", err)
	}

	// 1. Submit request requiring approval
	req, err := svc.SubmitRequest("user_2", model.SubmitServiceRequest{
		CatalogItemID: item.ID,
		Title:         "Workstation refresh for ML modeling",
		Description:   "Upgrading laptop for local LLM fine-tuning",
		Priority:      model.RequestPriorityMedium,
	})
	if err != nil {
		t.Fatalf("unexpected error submitting request: %v", err)
	}
	if req.Status != model.RequestStatusPendingApproval {
		t.Fatalf("expected pending_approval status, got %s", req.Status)
	}
	if len(req.Approvals) != 1 {
		t.Fatalf("expected 1 initial approval step, got %d", len(req.Approvals))
	}

	approvalID := req.Approvals[0].ID

	// 2. Reject request scenario
	approval, err := svc.ReviewApproval("user_1", approvalID, model.ReviewApprovalRequest{
		Status:   model.ApprovalStatusRejected,
		Comments: "Budget cycle closed until next quarter.",
	})
	if err != nil {
		t.Fatalf("unexpected error reviewing approval: %v", err)
	}
	if approval.Status != model.ApprovalStatusRejected {
		t.Fatalf("expected rejected approval status, got %s", approval.Status)
	}

	parentReq, err := svc.GetRequest(req.ID)
	if err != nil {
		t.Fatalf("failed to get parent request: %v", err)
	}
	if parentReq.Status != model.RequestStatusRejected {
		t.Fatalf("expected request status rejected, got %s", parentReq.Status)
	}
	if parentReq.RejectionReason != "Budget cycle closed until next quarter." {
		t.Fatalf("unexpected rejection reason: %s", parentReq.RejectionReason)
	}

	// 3. Double-decision guard: trying to review already decided approval must fail
	_, err = svc.ReviewApproval("user_1", approvalID, model.ReviewApprovalRequest{
		Status:   model.ApprovalStatusApproved,
		Comments: "Changed my mind",
	})
	if !errors.Is(err, ErrRequestAlreadyDecided) {
		t.Fatalf("expected ErrRequestAlreadyDecided, got %v", err)
	}

	// 4. Submit second request and approve it
	req2, err := svc.SubmitRequest("user_2", model.SubmitServiceRequest{
		CatalogItemID: item.ID,
		Title:         "Replacement laptop due to hardware fault",
		Priority:      model.RequestPriorityUrgent,
	})
	if err != nil {
		t.Fatalf("failed to submit request 2: %v", err)
	}

	appr2, err := svc.ReviewApproval("user_1", req2.Approvals[0].ID, model.ReviewApprovalRequest{
		Status:   model.ApprovalStatusApproved,
		Comments: "Approved for urgent replacement.",
	})
	if err != nil {
		t.Fatalf("unexpected error approving request 2: %v", err)
	}
	if appr2.Status != model.ApprovalStatusApproved {
		t.Fatalf("expected approved status, got %s", appr2.Status)
	}

	parentReq2, _ := svc.GetRequest(req2.ID)
	// Because AutoCreateTask was true, approving auto-created the task and placed request in_progress!
	if parentReq2.Status != model.RequestStatusInProgress {
		t.Fatalf("expected request 2 status in_progress, got %s", parentReq2.Status)
	}
	if parentReq2.ApprovedAt == nil {
		t.Fatalf("expected approved_at timestamp to be populated")
	}
	if parentReq2.TaskID == nil {
		t.Fatalf("expected generated TaskID, got nil")
	}
}

func TestServiceCatalogService_AnalyticsAndMetrics(t *testing.T) {
	svc, _, _ := setupTestServiceCatalogSvc(t)

	cat, err := svc.CreateCategory("user_1", model.CreateCatalogCategoryRequest{
		WorkspaceID: 1,
		Name:        "Cloud Services",
		Slug:        "cloud-services",
	})
	if err != nil {
		t.Fatalf("failed to create category: %v", err)
	}

	item, err := svc.CreateItem("user_1", model.CreateCatalogItemRequest{
		WorkspaceID:       1,
		CategoryID:        cat.ID,
		Name:              "Redis Managed Cluster",
		Slug:              "redis-cluster",
		CostEstimateCents: 5000,
		RequiresApproval:  false,
	})
	if err != nil {
		t.Fatalf("failed to create item: %v", err)
	}

	// Submit and fulfill request
	r1, err := svc.SubmitRequest("user_1", model.SubmitServiceRequest{
		CatalogItemID: item.ID,
		Title:         "Production Redis Cache",
	})
	if err != nil {
		t.Fatalf("failed to submit request: %v", err)
	}

	fulfilledStatus := model.RequestStatusFulfilled
	_, _ = svc.UpdateRequest(r1.ID, model.UpdateServiceRequest{
		Status: &fulfilledStatus,
	})

	// Get metrics
	metrics, err := svc.GetMetrics(1)
	if err != nil {
		t.Fatalf("unexpected error getting metrics: %v", err)
	}
	if metrics.TotalCategories != 1 {
		t.Fatalf("expected 1 category in metrics, got %d", metrics.TotalCategories)
	}
	if metrics.TotalItems != 1 {
		t.Fatalf("expected 1 item in metrics, got %d", metrics.TotalItems)
	}
	if metrics.TotalRequests != 1 {
		t.Fatalf("expected 1 total request, got %d", metrics.TotalRequests)
	}
	if metrics.FulfilledRequests != 1 {
		t.Fatalf("expected 1 fulfilled request, got %d", metrics.FulfilledRequests)
	}
	if metrics.TotalCostEstimateCents != 5000 {
		t.Fatalf("expected 5000 cost estimate, got %d", metrics.TotalCostEstimateCents)
	}
	if len(metrics.MostRequestedItems) != 1 || metrics.MostRequestedItems[0].ItemID != item.ID {
		t.Fatalf("expected Redis Managed Cluster in top items, got %v", metrics.MostRequestedItems)
	}
}
