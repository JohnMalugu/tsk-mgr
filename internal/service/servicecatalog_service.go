package service

import (
	"github.com/JohnMalugu/tsk-mgr-api/internal/model"
	"github.com/JohnMalugu/tsk-mgr-api/internal/repository"
)

var (
	ErrCatalogCategoryNotFound = model.ErrCatalogCategoryNotFound
	ErrCatalogCategoryNameReq  = model.ErrCatalogCategoryNameReq
	ErrCatalogCategorySlugReq  = model.ErrCatalogCategorySlugReq
	ErrDuplicateCategorySlug    = model.ErrDuplicateCategorySlug
	ErrCatalogItemNotFound     = model.ErrCatalogItemNotFound
	ErrCatalogItemNameReq      = model.ErrCatalogItemNameReq
	ErrCatalogItemSlugReq      = model.ErrCatalogItemSlugReq
	ErrDuplicateItemSlug        = model.ErrDuplicateItemSlug
	ErrInvalidApprovalRole     = model.ErrInvalidApprovalRole
	ErrServiceRequestNotFound  = model.ErrServiceRequestNotFound
	ErrServiceRequestTitleReq  = model.ErrServiceRequestTitleReq
	ErrInvalidRequestStatus    = model.ErrInvalidRequestStatus
	ErrInvalidRequestPriority  = model.ErrInvalidRequestPriority
	ErrRequestApprovalNotFound = model.ErrRequestApprovalNotFound
	ErrInvalidApprovalStatus   = model.ErrInvalidApprovalStatus
	ErrRequestAlreadyDecided   = model.ErrRequestAlreadyDecided
	ErrRequestCommentEmpty     = model.ErrRequestCommentEmpty
)

// ServiceCatalogService orchestrates IT service catalog management, intake questionnaires, approval chains, and fulfillment workflows
type ServiceCatalogService interface {
	// Categories
	CreateCategory(creatorID string, req model.CreateCatalogCategoryRequest) (*model.ServiceCatalogCategory, error)
	GetCategory(id int64) (*model.ServiceCatalogCategory, error)
	GetCategoryBySlug(workspaceID int64, slug string) (*model.ServiceCatalogCategory, error)
	ListCategories(filter model.CatalogCategoryFilter) ([]model.ServiceCatalogCategory, error)
	UpdateCategory(id int64, req model.UpdateCatalogCategoryRequest) (*model.ServiceCatalogCategory, error)
	DeleteCategory(id int64) error

	// Catalog Items
	CreateItem(creatorID string, req model.CreateCatalogItemRequest) (*model.ServiceCatalogItem, error)
	GetItem(id int64) (*model.ServiceCatalogItem, error)
	GetItemBySlug(workspaceID int64, slug string) (*model.ServiceCatalogItem, error)
	ListItems(filter model.CatalogItemFilter) ([]model.ServiceCatalogItem, int, error)
	UpdateItem(id int64, req model.UpdateCatalogItemRequest) (*model.ServiceCatalogItem, error)
	DeleteItem(id int64) error

	// Service Requests & Fulfillment
	SubmitRequest(requesterID string, req model.SubmitServiceRequest) (*model.ServiceRequest, error)
	GetRequest(id int64) (*model.ServiceRequest, error)
	GetRequestByNumber(workspaceID int64, requestNumber string) (*model.ServiceRequest, error)
	ListRequests(filter model.ServiceRequestFilter) ([]model.ServiceRequest, int, error)
	UpdateRequest(id int64, req model.UpdateServiceRequest) (*model.ServiceRequest, error)
	ReviewApproval(approverID string, approvalID int64, req model.ReviewApprovalRequest) (*model.ServiceRequestApproval, error)
	AddComment(userID string, requestID int64, req model.AddRequestCommentRequest) (*model.ServiceRequestComment, error)
	ListComments(requestID int64) ([]model.ServiceRequestComment, error)
	GetMetrics(workspaceID int64) (*model.ServiceCatalogMetrics, error)
}

type serviceCatalogServiceImpl struct {
	repo        repository.ServiceCatalogRepository
	taskRepo    repository.TaskRepository
	projectRepo repository.ProjectRepository
	cmdbRepo    repository.CMDBRepository
}

// NewServiceCatalogService constructs a new ServiceCatalogService
func NewServiceCatalogService(
	repo repository.ServiceCatalogRepository,
	taskRepo repository.TaskRepository,
	projectRepo repository.ProjectRepository,
	cmdbRepo repository.CMDBRepository,
) ServiceCatalogService {
	return &serviceCatalogServiceImpl{
		repo:        repo,
		taskRepo:    taskRepo,
		projectRepo: projectRepo,
		cmdbRepo:    cmdbRepo,
	}
}

// Stubs for subsequent commits
func (s *serviceCatalogServiceImpl) CreateCategory(creatorID string, req model.CreateCatalogCategoryRequest) (*model.ServiceCatalogCategory, error) {
	return nil, nil
}
func (s *serviceCatalogServiceImpl) GetCategory(id int64) (*model.ServiceCatalogCategory, error) {
	return nil, nil
}
func (s *serviceCatalogServiceImpl) GetCategoryBySlug(workspaceID int64, slug string) (*model.ServiceCatalogCategory, error) {
	return nil, nil
}
func (s *serviceCatalogServiceImpl) ListCategories(filter model.CatalogCategoryFilter) ([]model.ServiceCatalogCategory, error) {
	return nil, nil
}
func (s *serviceCatalogServiceImpl) UpdateCategory(id int64, req model.UpdateCatalogCategoryRequest) (*model.ServiceCatalogCategory, error) {
	return nil, nil
}
func (s *serviceCatalogServiceImpl) DeleteCategory(id int64) error {
	return nil
}
func (s *serviceCatalogServiceImpl) CreateItem(creatorID string, req model.CreateCatalogItemRequest) (*model.ServiceCatalogItem, error) {
	return nil, nil
}
func (s *serviceCatalogServiceImpl) GetItem(id int64) (*model.ServiceCatalogItem, error) {
	return nil, nil
}
func (s *serviceCatalogServiceImpl) GetItemBySlug(workspaceID int64, slug string) (*model.ServiceCatalogItem, error) {
	return nil, nil
}
func (s *serviceCatalogServiceImpl) ListItems(filter model.CatalogItemFilter) ([]model.ServiceCatalogItem, int, error) {
	return nil, 0, nil
}
func (s *serviceCatalogServiceImpl) UpdateItem(id int64, req model.UpdateCatalogItemRequest) (*model.ServiceCatalogItem, error) {
	return nil, nil
}
func (s *serviceCatalogServiceImpl) DeleteItem(id int64) error {
	return nil
}
func (s *serviceCatalogServiceImpl) SubmitRequest(requesterID string, req model.SubmitServiceRequest) (*model.ServiceRequest, error) {
	return nil, nil
}
func (s *serviceCatalogServiceImpl) GetRequest(id int64) (*model.ServiceRequest, error) {
	return nil, nil
}
func (s *serviceCatalogServiceImpl) GetRequestByNumber(workspaceID int64, requestNumber string) (*model.ServiceRequest, error) {
	return nil, nil
}
func (s *serviceCatalogServiceImpl) ListRequests(filter model.ServiceRequestFilter) ([]model.ServiceRequest, int, error) {
	return nil, 0, nil
}
func (s *serviceCatalogServiceImpl) UpdateRequest(id int64, req model.UpdateServiceRequest) (*model.ServiceRequest, error) {
	return nil, nil
}
func (s *serviceCatalogServiceImpl) ReviewApproval(approverID string, approvalID int64, req model.ReviewApprovalRequest) (*model.ServiceRequestApproval, error) {
	return nil, nil
}
func (s *serviceCatalogServiceImpl) AddComment(userID string, requestID int64, req model.AddRequestCommentRequest) (*model.ServiceRequestComment, error) {
	return nil, nil
}
func (s *serviceCatalogServiceImpl) ListComments(requestID int64) ([]model.ServiceRequestComment, error) {
	return nil, nil
}
func (s *serviceCatalogServiceImpl) GetMetrics(workspaceID int64) (*model.ServiceCatalogMetrics, error) {
	return nil, nil
}
