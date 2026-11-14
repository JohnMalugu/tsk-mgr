package repository

import (
	"database/sql"
	"errors"
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
func (r *serviceCatalogRepositoryImpl) CreateItem(item *model.ServiceCatalogItem) (int64, error) {
	return 0, nil
}
func (r *serviceCatalogRepositoryImpl) GetItemByID(id int64) (*model.ServiceCatalogItem, error) {
	return nil, nil
}
func (r *serviceCatalogRepositoryImpl) GetItemBySlug(workspaceID int64, slug string) (*model.ServiceCatalogItem, error) {
	return nil, nil
}
func (r *serviceCatalogRepositoryImpl) ListItems(filter model.CatalogItemFilter) ([]model.ServiceCatalogItem, int, error) {
	return nil, 0, nil
}
func (r *serviceCatalogRepositoryImpl) UpdateItem(item *model.ServiceCatalogItem) error {
	return nil
}
func (r *serviceCatalogRepositoryImpl) DeleteItem(id int64) error {
	return nil
}
func (r *serviceCatalogRepositoryImpl) CreateRequest(req *model.ServiceRequest) (int64, error) {
	return 0, nil
}
func (r *serviceCatalogRepositoryImpl) GetRequestByID(id int64) (*model.ServiceRequest, error) {
	return nil, nil
}
func (r *serviceCatalogRepositoryImpl) GetRequestByNumber(workspaceID int64, requestNumber string) (*model.ServiceRequest, error) {
	return nil, nil
}
func (r *serviceCatalogRepositoryImpl) ListRequests(filter model.ServiceRequestFilter) ([]model.ServiceRequest, int, error) {
	return nil, 0, nil
}
func (r *serviceCatalogRepositoryImpl) UpdateRequest(req *model.ServiceRequest) error {
	return nil
}
func (r *serviceCatalogRepositoryImpl) DeleteRequest(id int64) error {
	return nil
}
func (r *serviceCatalogRepositoryImpl) GetNextRequestNumber(workspaceID int64) (string, error) {
	return "", nil
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
