package service

import (
	"errors"
	"fmt"
	"strings"
	"time"

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

func docSlugify(title string) string {
	s := slugify(title)
	if s == "" || s == "workspace" {
		s = fmt.Sprintf("doc-%d", time.Now().UnixNano()%100000)
	}
	return s
}

func isValidDocType(dt string) bool {
	switch dt {
	case model.DocTypeGeneral, model.DocTypeRFC, model.DocTypeDesignDoc,
		model.DocTypeMeetingNotes, model.DocTypeRunbook, model.DocTypePostmortem:
		return true
	default:
		return false
	}
}

func isValidPublishStatus(st string) bool {
	switch st {
	case model.DocStatusDraft, model.DocStatusInReview, model.DocStatusPublished, model.DocStatusArchived:
		return true
	default:
		return false
	}
}

func (s *documentServiceImpl) CreateDocument(authorID string, req model.CreateDocumentRequest) (*model.Document, error) {
	if req.WorkspaceID <= 0 {
		return nil, ErrWorkspaceRequired
	}
	title := strings.TrimSpace(req.Title)
	if title == "" {
		return nil, ErrDocumentTitleRequired
	}

	if authorID == "" {
		authorID = "user_1"
	}

	docType := req.DocType
	if docType == "" {
		docType = model.DocTypeGeneral
	} else if !isValidDocType(docType) {
		return nil, ErrInvalidDocType
	}

	status := req.Status
	if status == "" {
		status = model.DocStatusDraft
	} else if !isValidPublishStatus(status) {
		return nil, ErrInvalidPublishStatus
	}

	if req.ProjectID != nil && s.projectRepo != nil {
		p, err := s.projectRepo.GetByID(*req.ProjectID)
		if err != nil || p == nil {
			return nil, ErrProjectNotFound
		}
	}

	if req.ParentDocumentID != nil {
		parentDoc, err := s.repo.GetDocumentByID(*req.ParentDocumentID)
		if err != nil || parentDoc == nil || parentDoc.WorkspaceID != req.WorkspaceID {
			return nil, ErrParentDocumentNotFound
		}
	}

	baseSlug := req.Slug
	if baseSlug == "" {
		baseSlug = docSlugify(title)
	} else {
		baseSlug = docSlugify(baseSlug)
	}
	slug := baseSlug
	counter := 2
	for {
		existing, _ := s.repo.GetDocumentBySlug(req.WorkspaceID, slug)
		if existing == nil {
			break
		}
		slug = fmt.Sprintf("%s-%d", baseSlug, counter)
		counter++
	}

	doc := &model.Document{
		WorkspaceID:      req.WorkspaceID,
		ProjectID:        req.ProjectID,
		ParentDocumentID: req.ParentDocumentID,
		AuthorID:         authorID,
		Title:            title,
		Slug:             slug,
		Content:          req.Content,
		DocType:          docType,
		Status:           status,
		IsPinned:         req.IsPinned,
		ViewCount:        0,
		Version:          1,
	}

	id, err := s.repo.CreateDocument(doc)
	if err != nil {
		return nil, err
	}
	doc.ID = id

	// Initial revision
	_, _ = s.repo.CreateRevision(&model.DocumentRevision{
		DocumentID:    id,
		Version:       1,
		Title:         doc.Title,
		Content:       doc.Content,
		AuthorID:      authorID,
		ChangeSummary: "Initial creation",
	})

	return s.repo.GetDocumentByID(id)
}

func (s *documentServiceImpl) GetDocument(id int64, recordView bool) (*model.Document, error) {
	doc, err := s.repo.GetDocumentByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrDocumentNotFound) {
			return nil, ErrDocumentNotFound
		}
		return nil, err
	}
	if recordView {
		_ = s.repo.IncrementViewCount(id)
		doc.ViewCount++
	}
	return doc, nil
}

func (s *documentServiceImpl) GetDocumentBySlug(workspaceID int64, slug string, recordView bool) (*model.Document, error) {
	if workspaceID <= 0 {
		return nil, ErrWorkspaceRequired
	}
	doc, err := s.repo.GetDocumentBySlug(workspaceID, slug)
	if err != nil {
		if errors.Is(err, repository.ErrDocumentNotFound) {
			return nil, ErrDocumentNotFound
		}
		return nil, err
	}
	if recordView {
		_ = s.repo.IncrementViewCount(doc.ID)
		doc.ViewCount++
	}
	return doc, nil
}

func (s *documentServiceImpl) ListDocuments(filter model.DocumentFilter) ([]*model.Document, error) {
	if filter.WorkspaceID <= 0 {
		return nil, ErrWorkspaceRequired
	}
	return s.repo.ListDocuments(filter)
}

func (s *documentServiceImpl) UpdateDocument(id int64, authorID string, req model.UpdateDocumentRequest) (*model.Document, error) {
	doc, err := s.repo.GetDocumentByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrDocumentNotFound) {
			return nil, ErrDocumentNotFound
		}
		return nil, err
	}

	if authorID == "" {
		authorID = doc.AuthorID
	}

	var contentChanged bool
	if req.Title != nil {
		newTitle := strings.TrimSpace(*req.Title)
		if newTitle == "" {
			return nil, ErrDocumentTitleRequired
		}
		if newTitle != doc.Title {
			doc.Title = newTitle
			contentChanged = true
		}
	}

	if req.Content != nil && *req.Content != doc.Content {
		doc.Content = *req.Content
		contentChanged = true
	}

	if req.Slug != nil {
		cleanSlug := docSlugify(*req.Slug)
		if cleanSlug != "" && cleanSlug != doc.Slug {
			existing, _ := s.repo.GetDocumentBySlug(doc.WorkspaceID, cleanSlug)
			if existing != nil && existing.ID != doc.ID {
				cleanSlug = fmt.Sprintf("%s-%d", cleanSlug, id)
			}
			doc.Slug = cleanSlug
		}
	}

	if req.DocType != nil {
		if !isValidDocType(*req.DocType) {
			return nil, ErrInvalidDocType
		}
		doc.DocType = *req.DocType
	}

	if req.Status != nil {
		if !isValidPublishStatus(*req.Status) {
			return nil, ErrInvalidPublishStatus
		}
		doc.Status = *req.Status
	}

	if req.IsPinned != nil {
		doc.IsPinned = *req.IsPinned
	}

	if req.ProjectID != nil {
		if *req.ProjectID > 0 && s.projectRepo != nil {
			p, err := s.projectRepo.GetByID(*req.ProjectID)
			if err != nil || p == nil {
				return nil, ErrProjectNotFound
			}
			doc.ProjectID = req.ProjectID
		} else if *req.ProjectID == 0 {
			doc.ProjectID = nil
		}
	}

	if req.ParentDocumentID != nil {
		parentID := *req.ParentDocumentID
		if parentID > 0 {
			if parentID == doc.ID {
				return nil, ErrCircularDocumentHierarchy
			}
			parentDoc, err := s.repo.GetDocumentByID(parentID)
			if err != nil || parentDoc == nil || parentDoc.WorkspaceID != doc.WorkspaceID {
				return nil, ErrParentDocumentNotFound
			}
			// Check if new parent is a descendant of doc.ID
			curr := parentDoc
			for curr != nil && curr.ParentDocumentID != nil {
				if *curr.ParentDocumentID == doc.ID {
					return nil, ErrCircularDocumentHierarchy
				}
				p, err := s.repo.GetDocumentByID(*curr.ParentDocumentID)
				if err != nil {
					break
				}
				curr = p
			}
			doc.ParentDocumentID = &parentID
		} else {
			doc.ParentDocumentID = nil
		}
	}

	if contentChanged {
		doc.Version++
		summary := req.ChangeSummary
		if summary == "" {
			summary = fmt.Sprintf("Updated to version %d", doc.Version)
		}
		_, _ = s.repo.CreateRevision(&model.DocumentRevision{
			DocumentID:    doc.ID,
			Version:       doc.Version,
			Title:         doc.Title,
			Content:       doc.Content,
			AuthorID:      authorID,
			ChangeSummary: summary,
		})
	}

	err = s.repo.UpdateDocument(doc)
	if err != nil {
		return nil, err
	}
	return s.repo.GetDocumentByID(id)
}

func (s *documentServiceImpl) DeleteDocument(id int64) error {
	err := s.repo.DeleteDocument(id)
	if err != nil {
		if errors.Is(err, repository.ErrDocumentNotFound) {
			return ErrDocumentNotFound
		}
		return err
	}
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
