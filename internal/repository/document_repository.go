package repository

import (
	"database/sql"
	"errors"

	"github.com/JohnMalugu/tsk-mgr-api/internal/model"
)

var (
	ErrDocumentNotFound         = errors.New("document not found")
	ErrDocumentRevisionNotFound = errors.New("document revision not found")
	ErrDocumentCommentNotFound  = errors.New("document comment not found")
	ErrDocumentTaskLinkNotFound = errors.New("document task link not found")
)

// DocumentRepository defines database operations for documents, revisions, task links, comments, and metrics
type DocumentRepository interface {
	CreateDocument(doc *model.Document) (int64, error)
	GetDocumentByID(id int64) (*model.Document, error)
	GetDocumentBySlug(workspaceID int64, slug string) (*model.Document, error)
	ListDocuments(filter model.DocumentFilter) ([]*model.Document, error)
	UpdateDocument(doc *model.Document) error
	DeleteDocument(id int64) error
	IncrementViewCount(id int64) error
	TogglePinned(id int64, isPinned bool) error

	CreateRevision(rev *model.DocumentRevision) (int64, error)
	GetRevisionByVersion(documentID int64, version int) (*model.DocumentRevision, error)
	ListRevisions(documentID int64) ([]model.DocumentRevision, error)

	LinkTask(link *model.DocumentTaskLink) (int64, error)
	UnlinkTask(documentID int64, taskID int) error
	ListDocumentTasks(documentID int64) ([]model.DocumentTaskLink, error)

	CreateComment(comment *model.DocumentComment) (int64, error)
	GetCommentByID(id int64) (*model.DocumentComment, error)
	ListComments(documentID int64) ([]model.DocumentComment, error)
	ResolveComment(id int64, resolvedBy string) error

	GetDocumentMetrics(workspaceID int64) (*model.DocumentMetrics, error)
}

type documentRepositoryImpl struct {
	db *sql.DB
}

// NewDocumentRepository creates a new DocumentRepository instance
func NewDocumentRepository(db *sql.DB) DocumentRepository {
	return &documentRepositoryImpl{db: db}
}

func (r *documentRepositoryImpl) scanDocument(scanner interface{ Scan(...interface{}) error }) (*model.Document, error) {
	var doc model.Document
	var projectID sql.NullInt64
	var parentDocID sql.NullInt64
	var isPinnedInt int
	var authorUser, authorEmail, projName string

	err := scanner.Scan(
		&doc.ID,
		&doc.WorkspaceID,
		&projectID,
		&parentDocID,
		&doc.AuthorID,
		&doc.Title,
		&doc.Slug,
		&doc.Content,
		&doc.DocType,
		&doc.Status,
		&isPinnedInt,
		&doc.ViewCount,
		&doc.Version,
		&doc.CreatedAt,
		&doc.UpdatedAt,
		&authorUser,
		&authorEmail,
		&projName,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrDocumentNotFound
		}
		return nil, err
	}

	if projectID.Valid {
		pid := int(projectID.Int64)
		doc.ProjectID = &pid
	}
	if parentDocID.Valid {
		pdoc := parentDocID.Int64
		doc.ParentDocumentID = &pdoc
	}
	doc.IsPinned = isPinnedInt == 1
	doc.AuthorUsername = authorUser
	doc.AuthorEmail = authorEmail
	doc.ProjectName = projName
	return &doc, nil
}

const docSelectFields = `
	SELECT 
		d.id, d.workspace_id, d.project_id, d.parent_document_id, d.author_id,
		d.title, d.slug, d.content, d.doc_type, d.status, d.is_pinned,
		d.view_count, d.version, d.created_at, d.updated_at,
		COALESCE(u.username, ''), COALESCE(u.email, ''),
		COALESCE(p.name, '')
	FROM documents d
	LEFT JOIN users u ON d.author_id = u.id
	LEFT JOIN projects p ON d.project_id = p.id
`

func (r *documentRepositoryImpl) CreateDocument(doc *model.Document) (int64, error) {
	query := `
		INSERT INTO documents (
			workspace_id, project_id, parent_document_id, author_id,
			title, slug, content, doc_type, status, is_pinned,
			view_count, version, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, datetime('now'), datetime('now'))
	`
	pinned := 0
	if doc.IsPinned {
		pinned = 1
	}
	res, err := r.db.Exec(query,
		doc.WorkspaceID,
		doc.ProjectID,
		doc.ParentDocumentID,
		doc.AuthorID,
		doc.Title,
		doc.Slug,
		doc.Content,
		doc.DocType,
		doc.Status,
		pinned,
		doc.ViewCount,
		doc.Version,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (r *documentRepositoryImpl) GetDocumentByID(id int64) (*model.Document, error) {
	query := docSelectFields + ` WHERE d.id = ?`
	row := r.db.QueryRow(query, id)
	return r.scanDocument(row)
}

func (r *documentRepositoryImpl) GetDocumentBySlug(workspaceID int64, slug string) (*model.Document, error) {
	query := docSelectFields + ` WHERE d.workspace_id = ? AND d.slug = ?`
	row := r.db.QueryRow(query, workspaceID, slug)
	return r.scanDocument(row)
}
func (r *documentRepositoryImpl) ListDocuments(filter model.DocumentFilter) ([]*model.Document, error) {
	query := docSelectFields + ` WHERE 1=1`
	var args []interface{}

	if filter.WorkspaceID > 0 {
		query += ` AND d.workspace_id = ?`
		args = append(args, filter.WorkspaceID)
	}
	if filter.ProjectID != nil {
		query += ` AND d.project_id = ?`
		args = append(args, *filter.ProjectID)
	}
	if filter.AuthorID != "" {
		query += ` AND d.author_id = ?`
		args = append(args, filter.AuthorID)
	}
	if filter.ParentDocumentID != nil {
		query += ` AND d.parent_document_id = ?`
		args = append(args, *filter.ParentDocumentID)
	}
	if filter.DocType != "" {
		query += ` AND d.doc_type = ?`
		args = append(args, filter.DocType)
	}
	if filter.Status != "" {
		query += ` AND d.status = ?`
		args = append(args, filter.Status)
	}
	if filter.IsPinned != nil {
		pinnedVal := 0
		if *filter.IsPinned {
			pinnedVal = 1
		}
		query += ` AND d.is_pinned = ?`
		args = append(args, pinnedVal)
	}
	if filter.Search != "" {
		term := "%" + filter.Search + "%"
		query += ` AND (d.title LIKE ? OR d.content LIKE ?)`
		args = append(args, term, term)
	}

	query += ` ORDER BY d.is_pinned DESC, d.updated_at DESC`

	if filter.Limit > 0 {
		query += ` LIMIT ?`
		args = append(args, filter.Limit)
		if filter.Offset > 0 {
			query += ` OFFSET ?`
			args = append(args, filter.Offset)
		}
	}

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var docs []*model.Document
	for rows.Next() {
		doc, err := r.scanDocument(rows)
		if err != nil {
			return nil, err
		}
		docs = append(docs, doc)
	}
	if docs == nil {
		docs = []*model.Document{}
	}
	return docs, rows.Err()
}
func (r *documentRepositoryImpl) UpdateDocument(doc *model.Document) error {
	return nil
}
func (r *documentRepositoryImpl) DeleteDocument(id int64) error {
	return nil
}
func (r *documentRepositoryImpl) IncrementViewCount(id int64) error {
	return nil
}
func (r *documentRepositoryImpl) TogglePinned(id int64, isPinned bool) error {
	return nil
}
func (r *documentRepositoryImpl) CreateRevision(rev *model.DocumentRevision) (int64, error) {
	return 0, nil
}
func (r *documentRepositoryImpl) GetRevisionByVersion(documentID int64, version int) (*model.DocumentRevision, error) {
	return nil, nil
}
func (r *documentRepositoryImpl) ListRevisions(documentID int64) ([]model.DocumentRevision, error) {
	return nil, nil
}
func (r *documentRepositoryImpl) LinkTask(link *model.DocumentTaskLink) (int64, error) {
	return 0, nil
}
func (r *documentRepositoryImpl) UnlinkTask(documentID int64, taskID int) error {
	return nil
}
func (r *documentRepositoryImpl) ListDocumentTasks(documentID int64) ([]model.DocumentTaskLink, error) {
	return nil, nil
}
func (r *documentRepositoryImpl) CreateComment(comment *model.DocumentComment) (int64, error) {
	return 0, nil
}
func (r *documentRepositoryImpl) GetCommentByID(id int64) (*model.DocumentComment, error) {
	return nil, nil
}
func (r *documentRepositoryImpl) ListComments(documentID int64) ([]model.DocumentComment, error) {
	return nil, nil
}
func (r *documentRepositoryImpl) ResolveComment(id int64, resolvedBy string) error {
	return nil
}
func (r *documentRepositoryImpl) GetDocumentMetrics(workspaceID int64) (*model.DocumentMetrics, error) {
	return nil, nil
}

