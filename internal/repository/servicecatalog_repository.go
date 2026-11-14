package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/JohnMalugu/tsk-mgr-api/internal/model"
)

var (
	ErrCatalogCategoryNotFound = errors.New("service catalog category not found")
	ErrDuplicateCategorySlug   = errors.New("service catalog category slug already in use")
	ErrCatalogItemNotFound     = errors.New("service catalog item not found")
	ErrDuplicateItemSlug       = errors.New("service catalog item slug already in use")
	ErrServiceRequestNotFound  = errors.New("service request not found")
	ErrRequestApprovalNotFound = errors.New("service request approval not found")
)

// ServiceCatalogRepository defines data persistence operations for the IT service catalog & requests
type ServiceCatalogRepository interface {
	// Categories
	CreateCategory(cat *model.ServiceCatalogCategory) (int64, error)
	GetCategoryByID(id int64) (*model.ServiceCatalogCategory, error)
	GetCategoryBySlug(workspaceID int64, slug string) (*model.ServiceCatalogCategory, error)
	ListCategories(filter model.CatalogCategoryFilter) ([]model.ServiceCatalogCategory, error)
	UpdateCategory(cat *model.ServiceCatalogCategory) error
	DeleteCategory(id int64) error

	// Catalog Items
	CreateItem(item *model.ServiceCatalogItem) (int64, error)
	GetItemByID(id int64) (*model.ServiceCatalogItem, error)
	GetItemBySlug(workspaceID int64, slug string) (*model.ServiceCatalogItem, error)
	ListItems(filter model.CatalogItemFilter) ([]model.ServiceCatalogItem, int, error)
	UpdateItem(item *model.ServiceCatalogItem) error
	DeleteItem(id int64) error

	// Service Requests
	CreateRequest(req *model.ServiceRequest) (int64, error)
	GetRequestByID(id int64) (*model.ServiceRequest, error)
	GetRequestByNumber(workspaceID int64, requestNumber string) (*model.ServiceRequest, error)
	ListRequests(filter model.ServiceRequestFilter) ([]model.ServiceRequest, int, error)
	UpdateRequest(req *model.ServiceRequest) error
	DeleteRequest(id int64) error
	GetNextRequestNumber(workspaceID int64) (string, error)

	// Approvals
	CreateApproval(approval *model.ServiceRequestApproval) (int64, error)
	GetApprovalByID(id int64) (*model.ServiceRequestApproval, error)
	ListApprovals(requestID int64) ([]model.ServiceRequestApproval, error)
	UpdateApproval(approval *model.ServiceRequestApproval) error

	// Comments
	CreateComment(comment *model.ServiceRequestComment) (int64, error)
	ListComments(requestID int64) ([]model.ServiceRequestComment, error)

	// Analytics
	GetMetrics(workspaceID int64) (*model.ServiceCatalogMetrics, error)
}

type serviceCatalogRepositoryImpl struct {
	db *sql.DB
}

// NewServiceCatalogRepository constructs a new ServiceCatalogRepository instance
func NewServiceCatalogRepository(db *sql.DB) ServiceCatalogRepository {
	return &serviceCatalogRepositoryImpl{db: db}
}

func scanCategory(scanner interface{ Scan(dest ...interface{}) error }) (*model.ServiceCatalogCategory, error) {
	var c model.ServiceCatalogCategory
	var isActiveInt int
	var itemCount int

	err := scanner.Scan(
		&c.ID, &c.WorkspaceID, &c.Name, &c.Slug, &c.Description,
		&c.Icon, &c.DisplayOrder, &isActiveInt, &c.CreatedAt, &c.UpdatedAt,
		&itemCount,
	)
	if err != nil {
		return nil, err
	}
	c.IsActive = (isActiveInt == 1)
	c.ItemCount = itemCount
	return &c, nil
}

func (r *serviceCatalogRepositoryImpl) CreateCategory(cat *model.ServiceCatalogCategory) (int64, error) {
	now := time.Now()
	if cat.CreatedAt.IsZero() {
		cat.CreatedAt = now
	}
	if cat.UpdatedAt.IsZero() {
		cat.UpdatedAt = now
	}

	isActiveInt := 0
	if cat.IsActive {
		isActiveInt = 1
	}

	var existingCount int
	_ = r.db.QueryRow(`SELECT COUNT(*) FROM service_catalog_categories WHERE workspace_id = ? AND slug = ?`,
		cat.WorkspaceID, cat.Slug).Scan(&existingCount)
	if existingCount > 0 {
		return 0, ErrDuplicateCategorySlug
	}

	res, err := r.db.Exec(`
		INSERT INTO service_catalog_categories (
			workspace_id, name, slug, description, icon, display_order, is_active, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, cat.WorkspaceID, cat.Name, cat.Slug, cat.Description, cat.Icon, cat.DisplayOrder, isActiveInt, cat.CreatedAt, cat.UpdatedAt)
	if err != nil {
		return 0, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	cat.ID = id
	return id, nil
}

func (r *serviceCatalogRepositoryImpl) GetCategoryByID(id int64) (*model.ServiceCatalogCategory, error) {
	query := `
		SELECT c.id, c.workspace_id, c.name, c.slug, c.description,
		       c.icon, c.display_order, c.is_active, c.created_at, c.updated_at,
		       (SELECT COUNT(*) FROM service_catalog_items WHERE category_id = c.id AND is_active = 1)
		FROM service_catalog_categories c
		WHERE c.id = ?
	`
	cat, err := scanCategory(r.db.QueryRow(query, id))
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrCatalogCategoryNotFound
		}
		return nil, err
	}
	return cat, nil
}

func (r *serviceCatalogRepositoryImpl) GetCategoryBySlug(workspaceID int64, slug string) (*model.ServiceCatalogCategory, error) {
	query := `
		SELECT c.id, c.workspace_id, c.name, c.slug, c.description,
		       c.icon, c.display_order, c.is_active, c.created_at, c.updated_at,
		       (SELECT COUNT(*) FROM service_catalog_items WHERE category_id = c.id AND is_active = 1)
		FROM service_catalog_categories c
		WHERE c.workspace_id = ? AND c.slug = ?
	`
	cat, err := scanCategory(r.db.QueryRow(query, workspaceID, slug))
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrCatalogCategoryNotFound
		}
		return nil, err
	}
	return cat, nil
}

func (r *serviceCatalogRepositoryImpl) ListCategories(filter model.CatalogCategoryFilter) ([]model.ServiceCatalogCategory, error) {
	query := `
		SELECT c.id, c.workspace_id, c.name, c.slug, c.description,
		       c.icon, c.display_order, c.is_active, c.created_at, c.updated_at,
		       (SELECT COUNT(*) FROM service_catalog_items WHERE category_id = c.id AND is_active = 1)
		FROM service_catalog_categories c
		WHERE c.workspace_id = ?
	`
	args := []interface{}{filter.WorkspaceID}

	if filter.IsActive != nil {
		activeInt := 0
		if *filter.IsActive {
			activeInt = 1
		}
		query += " AND c.is_active = ?"
		args = append(args, activeInt)
	}

	if filter.Search != "" {
		keyword := "%" + strings.ToLower(filter.Search) + "%"
		query += " AND (LOWER(c.name) LIKE ? OR LOWER(c.description) LIKE ?)"
		args = append(args, keyword, keyword)
	}

	query += " ORDER BY c.display_order ASC, c.name ASC"

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []model.ServiceCatalogCategory
	for rows.Next() {
		cat, err := scanCategory(rows)
		if err != nil {
			return nil, err
		}
		categories = append(categories, *cat)
	}
	if categories == nil {
		categories = []model.ServiceCatalogCategory{}
	}
	return categories, nil
}

func (r *serviceCatalogRepositoryImpl) UpdateCategory(cat *model.ServiceCatalogCategory) error {
	cat.UpdatedAt = time.Now()
	isActiveInt := 0
	if cat.IsActive {
		isActiveInt = 1
	}

	res, err := r.db.Exec(`
		UPDATE service_catalog_categories
		SET name = ?, slug = ?, description = ?, icon = ?, display_order = ?, is_active = ?, updated_at = ?
		WHERE id = ?
	`, cat.Name, cat.Slug, cat.Description, cat.Icon, cat.DisplayOrder, isActiveInt, cat.UpdatedAt, cat.ID)
	if err != nil {
		return err
	}

	rowsAff, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAff == 0 {
		return ErrCatalogCategoryNotFound
	}
	return nil
}

func (r *serviceCatalogRepositoryImpl) DeleteCategory(id int64) error {
	res, err := r.db.Exec(`DELETE FROM service_catalog_categories WHERE id = ?`, id)
	if err != nil {
		return err
	}
	rowsAff, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAff == 0 {
		return ErrCatalogCategoryNotFound
	}
	return nil
}
func scanCatalogItem(scanner interface{ Scan(dest ...interface{}) error }) (*model.ServiceCatalogItem, error) {
	var item model.ServiceCatalogItem
	var targetProjectID, ciID sql.NullInt64
	var reqApprInt, autoTaskInt, isActiveInt int
	var categoryName, projectName, ciName sql.NullString
	var reqCount int

	err := scanner.Scan(
		&item.ID, &item.WorkspaceID, &item.CategoryID, &item.Name, &item.Slug,
		&item.Description, &item.ShortDescription, &item.Icon,
		&item.EstimatedFulfillmentHours, &item.CostEstimateCents, &item.Currency,
		&reqApprInt, &item.ApprovalRole, &autoTaskInt, &targetProjectID, &ciID,
		&isActiveInt, &item.FormSchemaJSON, &item.CreatedByID, &item.CreatedAt, &item.UpdatedAt,
		&categoryName, &projectName, &ciName, &reqCount,
	)
	if err != nil {
		return nil, err
	}

	item.RequiresApproval = (reqApprInt == 1)
	item.AutoCreateTask = (autoTaskInt == 1)
	item.IsActive = (isActiveInt == 1)

	if targetProjectID.Valid {
		pid := targetProjectID.Int64
		item.TargetProjectID = &pid
	}
	if ciID.Valid {
		cid := ciID.Int64
		item.CIID = &cid
	}
	if categoryName.Valid {
		item.CategoryName = categoryName.String
	}
	if projectName.Valid {
		item.ProjectName = projectName.String
	}
	if ciName.Valid {
		item.CIName = ciName.String
	}
	item.RequestCount = reqCount

	return &item, nil
}

func (r *serviceCatalogRepositoryImpl) CreateItem(item *model.ServiceCatalogItem) (int64, error) {
	now := time.Now()
	if item.CreatedAt.IsZero() {
		item.CreatedAt = now
	}
	if item.UpdatedAt.IsZero() {
		item.UpdatedAt = now
	}
	if item.Currency == "" {
		item.Currency = "USD"
	}
	if item.FormSchemaJSON == "" {
		item.FormSchemaJSON = `{"fields":[]}`
	}

	reqApprInt := 0
	if item.RequiresApproval {
		reqApprInt = 1
	}
	autoTaskInt := 0
	if item.AutoCreateTask {
		autoTaskInt = 1
	}
	isActiveInt := 0
	if item.IsActive {
		isActiveInt = 1
	}

	var existingCount int
	_ = r.db.QueryRow(`SELECT COUNT(*) FROM service_catalog_items WHERE workspace_id = ? AND slug = ?`,
		item.WorkspaceID, item.Slug).Scan(&existingCount)
	if existingCount > 0 {
		return 0, ErrDuplicateItemSlug
	}

	res, err := r.db.Exec(`
		INSERT INTO service_catalog_items (
			workspace_id, category_id, name, slug, description, short_description, icon,
			estimated_fulfillment_hours, cost_estimate_cents, currency,
			requires_approval, approval_role, auto_create_task, target_project_id, ci_id,
			is_active, form_schema_json, created_by_id, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, item.WorkspaceID, item.CategoryID, item.Name, item.Slug, item.Description, item.ShortDescription, item.Icon,
		item.EstimatedFulfillmentHours, item.CostEstimateCents, item.Currency,
		reqApprInt, item.ApprovalRole, autoTaskInt, item.TargetProjectID, item.CIID,
		isActiveInt, item.FormSchemaJSON, item.CreatedByID, item.CreatedAt, item.UpdatedAt)
	if err != nil {
		return 0, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	item.ID = id
	return id, nil
}

func (r *serviceCatalogRepositoryImpl) GetItemByID(id int64) (*model.ServiceCatalogItem, error) {
	query := `
		SELECT i.id, i.workspace_id, i.category_id, i.name, i.slug,
		       i.description, i.short_description, i.icon,
		       i.estimated_fulfillment_hours, i.cost_estimate_cents, i.currency,
		       i.requires_approval, i.approval_role, i.auto_create_task, i.target_project_id, i.ci_id,
		       i.is_active, i.form_schema_json, i.created_by_id, i.created_at, i.updated_at,
		       c.name, p.name, ci.name,
		       (SELECT COUNT(*) FROM service_requests WHERE catalog_item_id = i.id)
		FROM service_catalog_items i
		LEFT JOIN service_catalog_categories c ON i.category_id = c.id
		LEFT JOIN projects p ON i.target_project_id = p.id
		LEFT JOIN configuration_items ci ON i.ci_id = ci.id
		WHERE i.id = ?
	`
	item, err := scanCatalogItem(r.db.QueryRow(query, id))
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrCatalogItemNotFound
		}
		return nil, err
	}
	return item, nil
}

func (r *serviceCatalogRepositoryImpl) GetItemBySlug(workspaceID int64, slug string) (*model.ServiceCatalogItem, error) {
	query := `
		SELECT i.id, i.workspace_id, i.category_id, i.name, i.slug,
		       i.description, i.short_description, i.icon,
		       i.estimated_fulfillment_hours, i.cost_estimate_cents, i.currency,
		       i.requires_approval, i.approval_role, i.auto_create_task, i.target_project_id, i.ci_id,
		       i.is_active, i.form_schema_json, i.created_by_id, i.created_at, i.updated_at,
		       c.name, p.name, ci.name,
		       (SELECT COUNT(*) FROM service_requests WHERE catalog_item_id = i.id)
		FROM service_catalog_items i
		LEFT JOIN service_catalog_categories c ON i.category_id = c.id
		LEFT JOIN projects p ON i.target_project_id = p.id
		LEFT JOIN configuration_items ci ON i.ci_id = ci.id
		WHERE i.workspace_id = ? AND i.slug = ?
	`
	item, err := scanCatalogItem(r.db.QueryRow(query, workspaceID, slug))
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrCatalogItemNotFound
		}
		return nil, err
	}
	return item, nil
}

func (r *serviceCatalogRepositoryImpl) ListItems(filter model.CatalogItemFilter) ([]model.ServiceCatalogItem, int, error) {
	countQuery := `
		SELECT COUNT(*)
		FROM service_catalog_items i
		WHERE i.workspace_id = ?
	`
	args := []interface{}{filter.WorkspaceID}

	if filter.CategoryID != nil {
		countQuery += " AND i.category_id = ?"
		args = append(args, *filter.CategoryID)
	}
	if filter.RequiresApproval != nil {
		reqAppr := 0
		if *filter.RequiresApproval {
			reqAppr = 1
		}
		countQuery += " AND i.requires_approval = ?"
		args = append(args, reqAppr)
	}
	if filter.IsActive != nil {
		act := 0
		if *filter.IsActive {
			act = 1
		}
		countQuery += " AND i.is_active = ?"
		args = append(args, act)
	}
	if filter.Search != "" {
		keyword := "%" + strings.ToLower(filter.Search) + "%"
		countQuery += " AND (LOWER(i.name) LIKE ? OR LOWER(i.description) LIKE ? OR LOWER(i.short_description) LIKE ?)"
		args = append(args, keyword, keyword, keyword)
	}

	var totalCount int
	if err := r.db.QueryRow(countQuery, args...).Scan(&totalCount); err != nil {
		return nil, 0, err
	}

	dataQuery := `
		SELECT i.id, i.workspace_id, i.category_id, i.name, i.slug,
		       i.description, i.short_description, i.icon,
		       i.estimated_fulfillment_hours, i.cost_estimate_cents, i.currency,
		       i.requires_approval, i.approval_role, i.auto_create_task, i.target_project_id, i.ci_id,
		       i.is_active, i.form_schema_json, i.created_by_id, i.created_at, i.updated_at,
		       c.name, p.name, ci.name,
		       (SELECT COUNT(*) FROM service_requests WHERE catalog_item_id = i.id)
		FROM service_catalog_items i
		LEFT JOIN service_catalog_categories c ON i.category_id = c.id
		LEFT JOIN projects p ON i.target_project_id = p.id
		LEFT JOIN configuration_items ci ON i.ci_id = ci.id
		WHERE i.workspace_id = ?
	`
	dataArgs := []interface{}{filter.WorkspaceID}

	if filter.CategoryID != nil {
		dataQuery += " AND i.category_id = ?"
		dataArgs = append(dataArgs, *filter.CategoryID)
	}
	if filter.RequiresApproval != nil {
		reqAppr := 0
		if *filter.RequiresApproval {
			reqAppr = 1
		}
		dataQuery += " AND i.requires_approval = ?"
		dataArgs = append(dataArgs, reqAppr)
	}
	if filter.IsActive != nil {
		act := 0
		if *filter.IsActive {
			act = 1
		}
		dataQuery += " AND i.is_active = ?"
		dataArgs = append(dataArgs, act)
	}
	if filter.Search != "" {
		keyword := "%" + strings.ToLower(filter.Search) + "%"
		dataQuery += " AND (LOWER(i.name) LIKE ? OR LOWER(i.description) LIKE ? OR LOWER(i.short_description) LIKE ?)"
		dataArgs = append(dataArgs, keyword, keyword, keyword)
	}

	dataQuery += " ORDER BY i.name ASC, i.id ASC"

	if filter.Limit > 0 {
		dataQuery += " LIMIT ?"
		dataArgs = append(dataArgs, filter.Limit)
		if filter.Offset > 0 {
			dataQuery += " OFFSET ?"
			dataArgs = append(dataArgs, filter.Offset)
		}
	}

	rows, err := r.db.Query(dataQuery, dataArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var items []model.ServiceCatalogItem
	for rows.Next() {
		item, err := scanCatalogItem(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, *item)
	}
	if items == nil {
		items = []model.ServiceCatalogItem{}
	}
	return items, totalCount, nil
}

func (r *serviceCatalogRepositoryImpl) UpdateItem(item *model.ServiceCatalogItem) error {
	item.UpdatedAt = time.Now()
	reqApprInt := 0
	if item.RequiresApproval {
		reqApprInt = 1
	}
	autoTaskInt := 0
	if item.AutoCreateTask {
		autoTaskInt = 1
	}
	isActiveInt := 0
	if item.IsActive {
		isActiveInt = 1
	}

	res, err := r.db.Exec(`
		UPDATE service_catalog_items
		SET category_id = ?, name = ?, slug = ?, description = ?, short_description = ?, icon = ?,
		    estimated_fulfillment_hours = ?, cost_estimate_cents = ?, currency = ?,
		    requires_approval = ?, approval_role = ?, auto_create_task = ?, target_project_id = ?, ci_id = ?,
		    is_active = ?, form_schema_json = ?, updated_at = ?
		WHERE id = ?
	`, item.CategoryID, item.Name, item.Slug, item.Description, item.ShortDescription, item.Icon,
		item.EstimatedFulfillmentHours, item.CostEstimateCents, item.Currency,
		reqApprInt, item.ApprovalRole, autoTaskInt, item.TargetProjectID, item.CIID,
		isActiveInt, item.FormSchemaJSON, item.UpdatedAt, item.ID)
	if err != nil {
		return err
	}

	rowsAff, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAff == 0 {
		return ErrCatalogItemNotFound
	}
	return nil
}

func (r *serviceCatalogRepositoryImpl) DeleteItem(id int64) error {
	res, err := r.db.Exec(`DELETE FROM service_catalog_items WHERE id = ?`, id)
	if err != nil {
		return err
	}
	rowsAff, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAff == 0 {
		return ErrCatalogItemNotFound
	}
	return nil
}
func scanServiceRequest(scanner interface{ Scan(dest ...interface{}) error }) (*model.ServiceRequest, error) {
	var req model.ServiceRequest
	var assignedToID sql.NullString
	var dueAt, approvedAt, fulfilledAt, cancelledAt sql.NullTime
	var taskID sql.NullInt64
	var itemName, requesterName, assignedToName, taskTitle sql.NullString

	err := scanner.Scan(
		&req.ID, &req.WorkspaceID, &req.CatalogItemID, &req.RequestNumber,
		&req.Title, &req.Description, &req.Status, &req.Priority,
		&req.RequesterID, &assignedToID, &req.FormDataJSON, &dueAt,
		&req.SubmittedAt, &approvedAt, &fulfilledAt, &cancelledAt,
		&req.CancellationReason, &req.RejectionReason, &taskID,
		&req.CreatedAt, &req.UpdatedAt,
		&itemName, &requesterName, &assignedToName, &taskTitle,
	)
	if err != nil {
		return nil, err
	}

	if assignedToID.Valid {
		at := assignedToID.String
		req.AssignedToID = &at
	}
	if dueAt.Valid {
		da := dueAt.Time
		req.DueAt = &da
	}
	if approvedAt.Valid {
		aa := approvedAt.Time
		req.ApprovedAt = &aa
	}
	if fulfilledAt.Valid {
		fa := fulfilledAt.Time
		req.FulfilledAt = &fa
	}
	if cancelledAt.Valid {
		ca := cancelledAt.Time
		req.CancelledAt = &ca
	}
	if taskID.Valid {
		tid := int(taskID.Int64)
		req.TaskID = &tid
	}
	if itemName.Valid {
		req.CatalogItemName = itemName.String
	}
	if requesterName.Valid {
		req.RequesterName = requesterName.String
	}
	if assignedToName.Valid {
		req.AssignedToName = assignedToName.String
	}
	if taskTitle.Valid {
		req.TaskTitle = taskTitle.String
	}

	return &req, nil
}

func (r *serviceCatalogRepositoryImpl) GetNextRequestNumber(workspaceID int64) (string, error) {
	var count int
	err := r.db.QueryRow(`SELECT COUNT(*) FROM service_requests WHERE workspace_id = ?`, workspaceID).Scan(&count)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("REQ-%04d", count+1), nil
}

func (r *serviceCatalogRepositoryImpl) CreateRequest(req *model.ServiceRequest) (int64, error) {
	now := time.Now()
	if req.CreatedAt.IsZero() {
		req.CreatedAt = now
	}
	if req.UpdatedAt.IsZero() {
		req.UpdatedAt = now
	}
	if req.SubmittedAt.IsZero() {
		req.SubmittedAt = now
	}
	if req.Status == "" {
		req.Status = model.RequestStatusSubmitted
	}
	if req.Priority == "" {
		req.Priority = model.RequestPriorityMedium
	}
	if req.FormDataJSON == "" {
		req.FormDataJSON = "{}"
	}

	if req.RequestNumber == "" {
		num, err := r.GetNextRequestNumber(req.WorkspaceID)
		if err != nil {
			return 0, err
		}
		req.RequestNumber = num
	}

	res, err := r.db.Exec(`
		INSERT INTO service_requests (
			workspace_id, catalog_item_id, request_number, title, description,
			status, priority, requester_id, assigned_to_id, form_data_json,
			due_at, submitted_at, approved_at, fulfilled_at, cancelled_at,
			cancellation_reason, rejection_reason, task_id, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, req.WorkspaceID, req.CatalogItemID, req.RequestNumber, req.Title, req.Description,
		req.Status, req.Priority, req.RequesterID, req.AssignedToID, req.FormDataJSON,
		req.DueAt, req.SubmittedAt, req.ApprovedAt, req.FulfilledAt, req.CancelledAt,
		req.CancellationReason, req.RejectionReason, req.TaskID, req.CreatedAt, req.UpdatedAt)
	if err != nil {
		return 0, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	req.ID = id
	return id, nil
}

func (r *serviceCatalogRepositoryImpl) GetRequestByID(id int64) (*model.ServiceRequest, error) {
	query := `
		SELECT sr.id, sr.workspace_id, sr.catalog_item_id, sr.request_number,
		       sr.title, sr.description, sr.status, sr.priority,
		       sr.requester_id, sr.assigned_to_id, sr.form_data_json, sr.due_at,
		       sr.submitted_at, sr.approved_at, sr.fulfilled_at, sr.cancelled_at,
		       sr.cancellation_reason, sr.rejection_reason, sr.task_id,
		       sr.created_at, sr.updated_at,
		       i.name, u1.username, u2.username, t.title
		FROM service_requests sr
		LEFT JOIN service_catalog_items i ON sr.catalog_item_id = i.id
		LEFT JOIN users u1 ON sr.requester_id = u1.id
		LEFT JOIN users u2 ON sr.assigned_to_id = u2.id
		LEFT JOIN tasks t ON sr.task_id = t.id
		WHERE sr.id = ?
	`
	req, err := scanServiceRequest(r.db.QueryRow(query, id))
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrServiceRequestNotFound
		}
		return nil, err
	}

	approvals, err := r.ListApprovals(id)
	if err == nil {
		req.Approvals = approvals
	}
	comments, err := r.ListComments(id)
	if err == nil {
		req.Comments = comments
	}

	return req, nil
}

func (r *serviceCatalogRepositoryImpl) GetRequestByNumber(workspaceID int64, requestNumber string) (*model.ServiceRequest, error) {
	query := `
		SELECT sr.id, sr.workspace_id, sr.catalog_item_id, sr.request_number,
		       sr.title, sr.description, sr.status, sr.priority,
		       sr.requester_id, sr.assigned_to_id, sr.form_data_json, sr.due_at,
		       sr.submitted_at, sr.approved_at, sr.fulfilled_at, sr.cancelled_at,
		       sr.cancellation_reason, sr.rejection_reason, sr.task_id,
		       sr.created_at, sr.updated_at,
		       i.name, u1.username, u2.username, t.title
		FROM service_requests sr
		LEFT JOIN service_catalog_items i ON sr.catalog_item_id = i.id
		LEFT JOIN users u1 ON sr.requester_id = u1.id
		LEFT JOIN users u2 ON sr.assigned_to_id = u2.id
		LEFT JOIN tasks t ON sr.task_id = t.id
		WHERE sr.workspace_id = ? AND sr.request_number = ?
	`
	req, err := scanServiceRequest(r.db.QueryRow(query, workspaceID, requestNumber))
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrServiceRequestNotFound
		}
		return nil, err
	}

	approvals, err := r.ListApprovals(req.ID)
	if err == nil {
		req.Approvals = approvals
	}
	comments, err := r.ListComments(req.ID)
	if err == nil {
		req.Comments = comments
	}

	return req, nil
}

func (r *serviceCatalogRepositoryImpl) ListRequests(filter model.ServiceRequestFilter) ([]model.ServiceRequest, int, error) {
	countQuery := `
		SELECT COUNT(*)
		FROM service_requests sr
		WHERE sr.workspace_id = ?
	`
	args := []interface{}{filter.WorkspaceID}

	if filter.CatalogItemID != nil {
		countQuery += " AND sr.catalog_item_id = ?"
		args = append(args, *filter.CatalogItemID)
	}
	if filter.RequesterID != "" {
		countQuery += " AND sr.requester_id = ?"
		args = append(args, filter.RequesterID)
	}
	if filter.AssignedToID != "" {
		countQuery += " AND sr.assigned_to_id = ?"
		args = append(args, filter.AssignedToID)
	}
	if filter.Status != "" {
		countQuery += " AND sr.status = ?"
		args = append(args, filter.Status)
	}
	if filter.Priority != "" {
		countQuery += " AND sr.priority = ?"
		args = append(args, filter.Priority)
	}
	if filter.Search != "" {
		keyword := "%" + strings.ToLower(filter.Search) + "%"
		countQuery += " AND (LOWER(sr.title) LIKE ? OR LOWER(sr.description) LIKE ? OR LOWER(sr.request_number) LIKE ?)"
		args = append(args, keyword, keyword, keyword)
	}

	var totalCount int
	if err := r.db.QueryRow(countQuery, args...).Scan(&totalCount); err != nil {
		return nil, 0, err
	}

	dataQuery := `
		SELECT sr.id, sr.workspace_id, sr.catalog_item_id, sr.request_number,
		       sr.title, sr.description, sr.status, sr.priority,
		       sr.requester_id, sr.assigned_to_id, sr.form_data_json, sr.due_at,
		       sr.submitted_at, sr.approved_at, sr.fulfilled_at, sr.cancelled_at,
		       sr.cancellation_reason, sr.rejection_reason, sr.task_id,
		       sr.created_at, sr.updated_at,
		       i.name, u1.username, u2.username, t.title
		FROM service_requests sr
		LEFT JOIN service_catalog_items i ON sr.catalog_item_id = i.id
		LEFT JOIN users u1 ON sr.requester_id = u1.id
		LEFT JOIN users u2 ON sr.assigned_to_id = u2.id
		LEFT JOIN tasks t ON sr.task_id = t.id
		WHERE sr.workspace_id = ?
	`
	dataArgs := []interface{}{filter.WorkspaceID}

	if filter.CatalogItemID != nil {
		dataQuery += " AND sr.catalog_item_id = ?"
		dataArgs = append(dataArgs, *filter.CatalogItemID)
	}
	if filter.RequesterID != "" {
		dataQuery += " AND sr.requester_id = ?"
		dataArgs = append(dataArgs, filter.RequesterID)
	}
	if filter.AssignedToID != "" {
		dataQuery += " AND sr.assigned_to_id = ?"
		dataArgs = append(dataArgs, filter.AssignedToID)
	}
	if filter.Status != "" {
		dataQuery += " AND sr.status = ?"
		dataArgs = append(dataArgs, filter.Status)
	}
	if filter.Priority != "" {
		dataQuery += " AND sr.priority = ?"
		dataArgs = append(dataArgs, filter.Priority)
	}
	if filter.Search != "" {
		keyword := "%" + strings.ToLower(filter.Search) + "%"
		dataQuery += " AND (LOWER(sr.title) LIKE ? OR LOWER(sr.description) LIKE ? OR LOWER(sr.request_number) LIKE ?)"
		dataArgs = append(dataArgs, keyword, keyword, keyword)
	}

	dataQuery += " ORDER BY sr.submitted_at DESC, sr.id DESC"

	if filter.Limit > 0 {
		dataQuery += " LIMIT ?"
		dataArgs = append(dataArgs, filter.Limit)
		if filter.Offset > 0 {
			dataQuery += " OFFSET ?"
			dataArgs = append(dataArgs, filter.Offset)
		}
	}

	rows, err := r.db.Query(dataQuery, dataArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var requests []model.ServiceRequest
	for rows.Next() {
		req, err := scanServiceRequest(rows)
		if err != nil {
			return nil, 0, err
		}
		requests = append(requests, *req)
	}
	if requests == nil {
		requests = []model.ServiceRequest{}
	}
	return requests, totalCount, nil
}

func (r *serviceCatalogRepositoryImpl) UpdateRequest(req *model.ServiceRequest) error {
	req.UpdatedAt = time.Now()
	res, err := r.db.Exec(`
		UPDATE service_requests
		SET title = ?, description = ?, status = ?, priority = ?, assigned_to_id = ?,
		    form_data_json = ?, due_at = ?, approved_at = ?, fulfilled_at = ?,
		    cancelled_at = ?, cancellation_reason = ?, rejection_reason = ?,
		    task_id = ?, updated_at = ?
		WHERE id = ?
	`, req.Title, req.Description, req.Status, req.Priority, req.AssignedToID,
		req.FormDataJSON, req.DueAt, req.ApprovedAt, req.FulfilledAt,
		req.CancelledAt, req.CancellationReason, req.RejectionReason,
		req.TaskID, req.UpdatedAt, req.ID)
	if err != nil {
		return err
	}

	rowsAff, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAff == 0 {
		return ErrServiceRequestNotFound
	}
	return nil
}

func (r *serviceCatalogRepositoryImpl) DeleteRequest(id int64) error {
	res, err := r.db.Exec(`DELETE FROM service_requests WHERE id = ?`, id)
	if err != nil {
		return err
	}
	rowsAff, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAff == 0 {
		return ErrServiceRequestNotFound
	}
	return nil
}
func (r *serviceCatalogRepositoryImpl) CreateApproval(approval *model.ServiceRequestApproval) (int64, error) {
	return 0, nil
}
func (r *serviceCatalogRepositoryImpl) GetApprovalByID(id int64) (*model.ServiceRequestApproval, error) {
	return nil, nil
}
func (r *serviceCatalogRepositoryImpl) ListApprovals(requestID int64) ([]model.ServiceRequestApproval, error) {
	return nil, nil
}
func (r *serviceCatalogRepositoryImpl) UpdateApproval(approval *model.ServiceRequestApproval) error {
	return nil
}
func (r *serviceCatalogRepositoryImpl) CreateComment(comment *model.ServiceRequestComment) (int64, error) {
	return 0, nil
}
func (r *serviceCatalogRepositoryImpl) ListComments(requestID int64) ([]model.ServiceRequestComment, error) {
	return nil, nil
}
func (r *serviceCatalogRepositoryImpl) GetMetrics(workspaceID int64) (*model.ServiceCatalogMetrics, error) {
	return nil, nil
}
