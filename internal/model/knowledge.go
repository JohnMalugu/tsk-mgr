package model

import (
	"time"
)

// Knowledge category constants
const (
	KnowledgeCategoryHowTo           = "how_to"
	KnowledgeCategoryTroubleshooting = "troubleshooting"
	KnowledgeCategoryWorkaround      = "workaround"
	KnowledgeCategorySOP             = "sop"
	KnowledgeCategoryReference       = "reference"
	KnowledgeCategoryFAQ             = "faq"
)

// Knowledge status constants
const (
	KnowledgeStatusDraft      = "draft"
	KnowledgeStatusInReview   = "in_review"
	KnowledgeStatusPublished  = "published"
	KnowledgeStatusArchived   = "archived"
	KnowledgeStatusDeprecated = "deprecated"
)

// Knowledge visibility constants
const (
	KnowledgeVisibilityInternal   = "internal"
	KnowledgeVisibilityPublic     = "public"
	KnowledgeVisibilityRestricted = "restricted"
)

// IsValidKnowledgeCategory validates knowledge category
func IsValidKnowledgeCategory(cat string) bool {
	switch cat {
	case KnowledgeCategoryHowTo, KnowledgeCategoryTroubleshooting, KnowledgeCategoryWorkaround,
		KnowledgeCategorySOP, KnowledgeCategoryReference, KnowledgeCategoryFAQ:
		return true
	default:
		return false
	}
}

// IsValidKnowledgeStatus validates knowledge lifecycle status
func IsValidKnowledgeStatus(status string) bool {
	switch status {
	case KnowledgeStatusDraft, KnowledgeStatusInReview, KnowledgeStatusPublished,
		KnowledgeStatusArchived, KnowledgeStatusDeprecated:
		return true
	default:
		return false
	}
}

// IsValidKnowledgeVisibility validates knowledge visibility scope
func IsValidKnowledgeVisibility(vis string) bool {
	switch vis {
	case KnowledgeVisibilityInternal, KnowledgeVisibilityPublic, KnowledgeVisibilityRestricted:
		return true
	default:
		return false
	}
}

// KnowledgeArticle represents an ITIL Knowledge Base / KEDB record
type KnowledgeArticle struct {
	ID              string     `json:"id"`
	WorkspaceID     string     `json:"workspace_id"`
	ArticleCode     string     `json:"article_code"`
	Title           string     `json:"title"`
	Summary         string     `json:"summary"`
	Content         string     `json:"content"`
	Category        string     `json:"category"`
	Status          string     `json:"status"`
	Visibility      string     `json:"visibility"`
	Version         int        `json:"version"`
	AuthorID        string     `json:"author_id,omitempty"`
	CIID            string     `json:"ci_id,omitempty"`
	ProblemID       string     `json:"problem_id,omitempty"`
	IncidentID      string     `json:"incident_id,omitempty"`
	Tags            string     `json:"tags,omitempty"`
	ViewCount       int        `json:"view_count"`
	HelpfulCount    int        `json:"helpful_count"`
	NotHelpfulCount int        `json:"not_helpful_count"`
	ReviewDueDate   *time.Time `json:"review_due_date,omitempty"`
	PublishedAt     *time.Time `json:"published_at,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}
