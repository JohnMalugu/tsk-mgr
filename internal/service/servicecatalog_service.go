package service

import (
	"encoding/json"
	"errors"
	"strings"
	"time"

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
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, ErrCatalogCategoryNameReq
	}

	slug := req.Slug
	if strings.TrimSpace(slug) == "" {
		slug = model.NormalizeSlug(name)
	} else {
		slug = model.NormalizeSlug(slug)
	}
	if slug == "" {
		return nil, ErrCatalogCategorySlugReq
	}

	cat := &model.ServiceCatalogCategory{
		WorkspaceID:  req.WorkspaceID,
		Name:         name,
		Slug:         slug,
		Description:  strings.TrimSpace(req.Description),
		Icon:         strings.TrimSpace(req.Icon),
		DisplayOrder: req.DisplayOrder,
		IsActive:     true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	id, err := s.repo.CreateCategory(cat)
	if err != nil {
		if errors.Is(err, repository.ErrDuplicateCategorySlug) {
			return nil, ErrDuplicateCategorySlug
		}
		return nil, err
	}

	return s.GetCategory(id)
}

func (s *serviceCatalogServiceImpl) GetCategory(id int64) (*model.ServiceCatalogCategory, error) {
	cat, err := s.repo.GetCategoryByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrCatalogCategoryNotFound) {
			return nil, ErrCatalogCategoryNotFound
		}
		return nil, err
	}
	return cat, nil
}

func (s *serviceCatalogServiceImpl) GetCategoryBySlug(workspaceID int64, slug string) (*model.ServiceCatalogCategory, error) {
	cleanSlug := model.NormalizeSlug(slug)
	cat, err := s.repo.GetCategoryBySlug(workspaceID, cleanSlug)
	if err != nil {
		if errors.Is(err, repository.ErrCatalogCategoryNotFound) {
			return nil, ErrCatalogCategoryNotFound
		}
		return nil, err
	}
	return cat, nil
}

func (s *serviceCatalogServiceImpl) ListCategories(filter model.CatalogCategoryFilter) ([]model.ServiceCatalogCategory, error) {
	return s.repo.ListCategories(filter)
}

func (s *serviceCatalogServiceImpl) UpdateCategory(id int64, req model.UpdateCatalogCategoryRequest) (*model.ServiceCatalogCategory, error) {
	existing, err := s.repo.GetCategoryByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrCatalogCategoryNotFound) {
			return nil, ErrCatalogCategoryNotFound
		}
		return nil, err
	}

	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)
		if name == "" {
			return nil, ErrCatalogCategoryNameReq
		}
		existing.Name = name
	}
	if req.Slug != nil {
		slug := model.NormalizeSlug(*req.Slug)
		if slug == "" {
			return nil, ErrCatalogCategorySlugReq
		}
		existing.Slug = slug
	}
	if req.Description != nil {
		existing.Description = strings.TrimSpace(*req.Description)
	}
	if req.Icon != nil {
		existing.Icon = strings.TrimSpace(*req.Icon)
	}
	if req.DisplayOrder != nil {
		existing.DisplayOrder = *req.DisplayOrder
	}
	if req.IsActive != nil {
		existing.IsActive = *req.IsActive
	}

	if err := s.repo.UpdateCategory(existing); err != nil {
		if errors.Is(err, repository.ErrDuplicateCategorySlug) {
			return nil, ErrDuplicateCategorySlug
		}
		if errors.Is(err, repository.ErrCatalogCategoryNotFound) {
			return nil, ErrCatalogCategoryNotFound
		}
		return nil, err
	}

	return s.GetCategory(id)
}

func (s *serviceCatalogServiceImpl) DeleteCategory(id int64) error {
	err := s.repo.DeleteCategory(id)
	if err != nil {
		if errors.Is(err, repository.ErrCatalogCategoryNotFound) {
			return ErrCatalogCategoryNotFound
		}
		return err
	}
	return nil
}

func (s *serviceCatalogServiceImpl) CreateItem(creatorID string, req model.CreateCatalogItemRequest) (*model.ServiceCatalogItem, error) {
	_, err := s.repo.GetCategoryByID(req.CategoryID)
	if err != nil {
		if errors.Is(err, repository.ErrCatalogCategoryNotFound) {
			return nil, ErrCatalogCategoryNotFound
		}
		return nil, err
	}

	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, ErrCatalogItemNameReq
	}

	slug := req.Slug
	if strings.TrimSpace(slug) == "" {
		slug = model.NormalizeSlug(name)
	} else {
		slug = model.NormalizeSlug(slug)
	}
	if slug == "" {
		return nil, ErrCatalogItemSlugReq
	}

	if req.RequiresApproval {
		if err := model.ValidateApprovalRole(req.ApprovalRole); err != nil {
			return nil, err
		}
		if req.ApprovalRole == "" {
			req.ApprovalRole = model.ApprovalRoleManager
		}
	}

	schemaJSON := `{"fields":[]}`
	if req.FormSchema != nil {
		b, err := json.Marshal(req.FormSchema)
		if err == nil {
			schemaJSON = string(b)
		}
	}

	item := &model.ServiceCatalogItem{
		WorkspaceID:               req.WorkspaceID,
		CategoryID:                req.CategoryID,
		Name:                      name,
		Slug:                      slug,
		Description:               strings.TrimSpace(req.Description),
		ShortDescription:          strings.TrimSpace(req.ShortDescription),
		Icon:                      strings.TrimSpace(req.Icon),
		EstimatedFulfillmentHours: req.EstimatedFulfillmentHours,
		CostEstimateCents:         req.CostEstimateCents,
		Currency:                  req.Currency,
		RequiresApproval:          req.RequiresApproval,
		ApprovalRole:              req.ApprovalRole,
		AutoCreateTask:            req.AutoCreateTask,
		TargetProjectID:           req.TargetProjectID,
		CIID:                      req.CIID,
		IsActive:                  true,
		FormSchemaJSON:            schemaJSON,
		CreatedByID:               creatorID,
		CreatedAt:                 time.Now(),
		UpdatedAt:                 time.Now(),
	}

	id, err := s.repo.CreateItem(item)
	if err != nil {
		if errors.Is(err, repository.ErrDuplicateItemSlug) {
			return nil, ErrDuplicateItemSlug
		}
		return nil, err
	}

	return s.GetItem(id)
}

func (s *serviceCatalogServiceImpl) GetItem(id int64) (*model.ServiceCatalogItem, error) {
	item, err := s.repo.GetItemByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrCatalogItemNotFound) {
			return nil, ErrCatalogItemNotFound
		}
		return nil, err
	}

	if item.FormSchemaJSON != "" {
		var schema model.CatalogFormSchema
		if err := json.Unmarshal([]byte(item.FormSchemaJSON), &schema); err == nil {
			item.FormSchema = &schema
		}
	}

	return item, nil
}

func (s *serviceCatalogServiceImpl) GetItemBySlug(workspaceID int64, slug string) (*model.ServiceCatalogItem, error) {
	cleanSlug := model.NormalizeSlug(slug)
	item, err := s.repo.GetItemBySlug(workspaceID, cleanSlug)
	if err != nil {
		if errors.Is(err, repository.ErrCatalogItemNotFound) {
			return nil, ErrCatalogItemNotFound
		}
		return nil, err
	}

	if item.FormSchemaJSON != "" {
		var schema model.CatalogFormSchema
		if err := json.Unmarshal([]byte(item.FormSchemaJSON), &schema); err == nil {
			item.FormSchema = &schema
		}
	}

	return item, nil
}

func (s *serviceCatalogServiceImpl) ListItems(filter model.CatalogItemFilter) ([]model.ServiceCatalogItem, int, error) {
	items, total, err := s.repo.ListItems(filter)
	if err != nil {
		return nil, 0, err
	}

	for i := range items {
		if items[i].FormSchemaJSON != "" {
			var schema model.CatalogFormSchema
			if err := json.Unmarshal([]byte(items[i].FormSchemaJSON), &schema); err == nil {
				items[i].FormSchema = &schema
			}
		}
	}

	return items, total, nil
}

func (s *serviceCatalogServiceImpl) UpdateItem(id int64, req model.UpdateCatalogItemRequest) (*model.ServiceCatalogItem, error) {
	existing, err := s.repo.GetItemByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrCatalogItemNotFound) {
			return nil, ErrCatalogItemNotFound
		}
		return nil, err
	}

	if req.CategoryID != nil {
		if _, err := s.repo.GetCategoryByID(*req.CategoryID); err != nil {
			return nil, ErrCatalogCategoryNotFound
		}
		existing.CategoryID = *req.CategoryID
	}
	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)
		if name == "" {
			return nil, ErrCatalogItemNameReq
		}
		existing.Name = name
	}
	if req.Slug != nil {
		slug := model.NormalizeSlug(*req.Slug)
		if slug == "" {
			return nil, ErrCatalogItemSlugReq
		}
		existing.Slug = slug
	}
	if req.Description != nil {
		existing.Description = strings.TrimSpace(*req.Description)
	}
	if req.ShortDescription != nil {
		existing.ShortDescription = strings.TrimSpace(*req.ShortDescription)
	}
	if req.Icon != nil {
		existing.Icon = strings.TrimSpace(*req.Icon)
	}
	if req.EstimatedFulfillmentHours != nil {
		existing.EstimatedFulfillmentHours = *req.EstimatedFulfillmentHours
	}
	if req.CostEstimateCents != nil {
		existing.CostEstimateCents = *req.CostEstimateCents
	}
	if req.Currency != nil {
		existing.Currency = *req.Currency
	}
	if req.RequiresApproval != nil {
		existing.RequiresApproval = *req.RequiresApproval
	}
	if req.ApprovalRole != nil {
		if err := model.ValidateApprovalRole(*req.ApprovalRole); err != nil {
			return nil, err
		}
		existing.ApprovalRole = *req.ApprovalRole
	}
	if req.AutoCreateTask != nil {
		existing.AutoCreateTask = *req.AutoCreateTask
	}
	if req.TargetProjectID != nil {
		existing.TargetProjectID = req.TargetProjectID
	}
	if req.CIID != nil {
		existing.CIID = req.CIID
	}
	if req.IsActive != nil {
		existing.IsActive = *req.IsActive
	}
	if req.FormSchema != nil {
		b, err := json.Marshal(req.FormSchema)
		if err == nil {
			existing.FormSchemaJSON = string(b)
		}
	}

	if err := s.repo.UpdateItem(existing); err != nil {
		if errors.Is(err, repository.ErrDuplicateItemSlug) {
			return nil, ErrDuplicateItemSlug
		}
		if errors.Is(err, repository.ErrCatalogItemNotFound) {
			return nil, ErrCatalogItemNotFound
		}
		return nil, err
	}

	return s.GetItem(id)
}

func (s *serviceCatalogServiceImpl) DeleteItem(id int64) error {
	err := s.repo.DeleteItem(id)
	if err != nil {
		if errors.Is(err, repository.ErrCatalogItemNotFound) {
			return ErrCatalogItemNotFound
		}
		return err
	}
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
