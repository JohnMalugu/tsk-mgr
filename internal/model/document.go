package model

import "time"

// Document type constants
const (
	DocTypeGeneral      = "general"
	DocTypeRFC          = "rfc"
	DocTypeDesignDoc    = "design_doc"
	DocTypeMeetingNotes = "meeting_notes"
	DocTypeRunbook      = "runbook"
	DocTypePostmortem   = "postmortem"
)

// Document publishing lifecycle status constants
const (
	DocStatusDraft     = "draft"
	DocStatusInReview  = "in_review"
	DocStatusPublished = "published"
	DocStatusArchived  = "archived"
)

// Document represents a knowledge base wiki page or architectural document
type Document struct {
	ID               int64     `json:"id"`
	WorkspaceID      int64     `json:"workspace_id"`
	ProjectID        *int      `json:"project_id,omitempty"`
	ParentDocumentID *int64    `json:"parent_document_id,omitempty"`
	AuthorID         string    `json:"author_id"`
	Title            string    `json:"title"`
	Slug             string    `json:"slug"`
	Content          string    `json:"content"`
	DocType          string    `json:"doc_type"`
	Status           string    `json:"status"`
	IsPinned         bool      `json:"is_pinned"`
	ViewCount        int       `json:"view_count"`
	Version          int       `json:"version"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`

	// Enriched fields
	AuthorUsername string `json:"author_username,omitempty"`
	AuthorEmail    string `json:"author_email,omitempty"`
	ProjectName    string `json:"project_name,omitempty"`
}
