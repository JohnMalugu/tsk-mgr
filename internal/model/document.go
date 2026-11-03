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

// DocumentRevision preserves historical point-in-time snapshots of document content
type DocumentRevision struct {
	ID            int64     `json:"id"`
	DocumentID    int64     `json:"document_id"`
	Version       int       `json:"version"`
	Title         string    `json:"title"`
	Content       string    `json:"content"`
	AuthorID      string    `json:"author_id"`
	ChangeSummary string    `json:"change_summary,omitempty"`
	CreatedAt     time.Time `json:"created_at"`

	// Enriched fields
	AuthorUsername string `json:"author_username,omitempty"`
}

// Document task link relationship constants
const (
	DocTaskRelRelated            = "related"
	DocTaskRelSpecification      = "specification"
	DocTaskRelMeetingActionItems = "meeting_action_items"
	DocTaskRelPostmortemAction   = "postmortem_action"
)

// DocumentTaskLink establishes bidirectional cross-references between a document and a task
type DocumentTaskLink struct {
	ID           int64     `json:"id"`
	DocumentID   int64     `json:"document_id"`
	TaskID       int       `json:"task_id"`
	Relationship string    `json:"relationship"` // related, specification, meeting_action_items, postmortem_action
	CreatedBy    string    `json:"created_by"`
	CreatedAt    time.Time `json:"created_at"`

	// Enriched fields
	TaskTitle    string `json:"task_title,omitempty"`
	TaskStatus   string `json:"task_status,omitempty"`
	TaskPriority string `json:"task_priority,omitempty"`
}

// DocumentComment represents discussions, feedback, and reviews on a document
type DocumentComment struct {
	ID         int64      `json:"id"`
	DocumentID int64      `json:"document_id"`
	AuthorID   string     `json:"author_id"`
	Content    string     `json:"content"`
	IsResolved bool       `json:"is_resolved"`
	ResolvedBy *string    `json:"resolved_by,omitempty"`
	ResolvedAt *time.Time `json:"resolved_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`

	// Enriched fields
	AuthorUsername   string `json:"author_username,omitempty"`
	ResolvedUsername string `json:"resolved_username,omitempty"`
}



