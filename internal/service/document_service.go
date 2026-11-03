package service

import (
	"errors"

	"github.com/JohnMalugu/tsk-mgr-api/internal/model"
	"github.com/JohnMalugu/tsk-mgr-api/internal/repository"
)

var (
	ErrDocumentNotFound          = errors.New("document not found")
	ErrDocumentTitleRequired     = errors.New("document title is required")
	ErrInvalidDocType            = errors.New("invalid document type")
	ErrInvalidPublishStatus      = errors.New("invalid document publishing status")
	ErrParentDocumentNotFound    = errors.New("parent document not found")
	ErrCircularDocumentHierarchy = errors.New("circular hierarchy detected: document cannot be its own ancestor")
	ErrRevisionNotFound          = errors.New("document revision not found")
	ErrInvalidTaskRelationship   = errors.New("invalid document task relationship")
)

// DocumentService manages wiki pages, versions, revisions, task cross-references, comments, and metrics
type DocumentService interface {
	CreateDocument(authorID string, req model.CreateDocumentRequest) (*model.Document, error)
	GetDocument(id int64, recordView bool) (*model.Document, error)
	GetDocumentBySlug(workspaceID int64, slug string, recordView bool) (*model.Document, error)
	ListDocuments(filter model.DocumentFilter) ([]*model.Document, error)
	UpdateDocument(id int64, authorID string, req model.UpdateDocumentRequest) (*model.Document, error)
	DeleteDocument(id int64) error
	TogglePin(id int64) (*model.Document, error)

	GetDocumentTree(workspaceID int64, projectID *int) ([]model.DocumentTreeNode, error)
	PublishDocument(id int64, status string) (*model.Document, error)

	ListRevisions(documentID int64) ([]model.DocumentRevision, error)
	GetRevision(documentID int64, version int) (*model.DocumentRevision, error)
	RollbackDocument(documentID int64, authorID string, req model.RollbackRevisionRequest) (*model.Document, error)

	LinkTask(documentID int64, createdBy string, req model.LinkDocumentTaskRequest) (*model.DocumentTaskLink, error)
	UnlinkTask(documentID int64, taskID int) error
	ListDocumentTasks(documentID int64) ([]model.DocumentTaskLink, error)

	CreateComment(documentID int64, authorID string, req model.CreateDocumentCommentRequest) (*model.DocumentComment, error)
	ListComments(documentID int64) ([]model.DocumentComment, error)
	ResolveComment(commentID int64, resolvedBy string) (*model.DocumentComment, error)

	GetDocumentMetrics(workspaceID int64) (*model.DocumentMetrics, error)
}

type documentServiceImpl struct {
	repo        repository.DocumentRepository
	projectRepo repository.ProjectRepository
	taskRepo    repository.TaskRepository
}

// NewDocumentService constructs a new DocumentService instance
func NewDocumentService(
	repo repository.DocumentRepository,
	projectRepo repository.ProjectRepository,
	taskRepo repository.TaskRepository,
) DocumentService {
	return &documentServiceImpl{
		repo:        repo,
		projectRepo: projectRepo,
		taskRepo:    taskRepo,
	}
}

func (s *documentServiceImpl) CreateDocument(authorID string, req model.CreateDocumentRequest) (*model.Document, error) {
	return nil, nil
}
func (s *documentServiceImpl) GetDocument(id int64, recordView bool) (*model.Document, error) {
	return nil, nil
}
func (s *documentServiceImpl) GetDocumentBySlug(workspaceID int64, slug string, recordView bool) (*model.Document, error) {
	return nil, nil
}
func (s *documentServiceImpl) ListDocuments(filter model.DocumentFilter) ([]*model.Document, error) {
	return nil, nil
}
func (s *documentServiceImpl) UpdateDocument(id int64, authorID string, req model.UpdateDocumentRequest) (*model.Document, error) {
	return nil, nil
}
func (s *documentServiceImpl) DeleteDocument(id int64) error {
	return nil
}
func (s *documentServiceImpl) TogglePin(id int64) (*model.Document, error) {
	return nil, nil
}
func (s *documentServiceImpl) GetDocumentTree(workspaceID int64, projectID *int) ([]model.DocumentTreeNode, error) {
	return nil, nil
}
func (s *documentServiceImpl) PublishDocument(id int64, status string) (*model.Document, error) {
	return nil, nil
}
func (s *documentServiceImpl) ListRevisions(documentID int64) ([]model.DocumentRevision, error) {
	return nil, nil
}
func (s *documentServiceImpl) GetRevision(documentID int64, version int) (*model.DocumentRevision, error) {
	return nil, nil
}
func (s *documentServiceImpl) RollbackDocument(documentID int64, authorID string, req model.RollbackRevisionRequest) (*model.Document, error) {
	return nil, nil
}
func (s *documentServiceImpl) LinkTask(documentID int64, createdBy string, req model.LinkDocumentTaskRequest) (*model.DocumentTaskLink, error) {
	return nil, nil
}
func (s *documentServiceImpl) UnlinkTask(documentID int64, taskID int) error {
	return nil
}
func (s *documentServiceImpl) ListDocumentTasks(documentID int64) ([]model.DocumentTaskLink, error) {
	return nil, nil
}
func (s *documentServiceImpl) CreateComment(documentID int64, authorID string, req model.CreateDocumentCommentRequest) (*model.DocumentComment, error) {
	return nil, nil
}
func (s *documentServiceImpl) ListComments(documentID int64) ([]model.DocumentComment, error) {
	return nil, nil
}
func (s *documentServiceImpl) ResolveComment(commentID int64, resolvedBy string) (*model.DocumentComment, error) {
	return nil, nil
}
func (s *documentServiceImpl) GetDocumentMetrics(workspaceID int64) (*model.DocumentMetrics, error) {
	return nil, nil
}
