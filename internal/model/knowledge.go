package model

import (
	"errors"
	"strings"
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

// KnowledgeRevision represents an immutable version snapshot of a knowledge article
type KnowledgeRevision struct {
	ID            string    `json:"id"`
	WorkspaceID   string    `json:"workspace_id"`
	ArticleID     string    `json:"article_id"`
	Version       int       `json:"version"`
	Title         string    `json:"title"`
	Content       string    `json:"content"`
	Summary       string    `json:"summary,omitempty"`
	ChangeSummary string    `json:"change_summary,omitempty"`
	AuthorID      string    `json:"author_id,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
}

// KnowledgeFeedback captures user evaluation, helpfulness votes, and qualitative feedback
type KnowledgeFeedback struct {
	ID          string    `json:"id"`
	WorkspaceID string    `json:"workspace_id"`
	ArticleID   string    `json:"article_id"`
	UserID      string    `json:"user_id"`
	IsHelpful   bool      `json:"is_helpful"`
	Rating      int       `json:"rating"` // 1 - 5 stars
	Comment     string    `json:"comment,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

// Knowledge review status constants
const (
	KnowledgeReviewStatusScheduled = "scheduled"
	KnowledgeReviewStatusCompleted = "completed"
	KnowledgeReviewStatusExpired   = "expired"
)

// IsValidKnowledgeReviewStatus validates review status
func IsValidKnowledgeReviewStatus(status string) bool {
	switch status {
	case KnowledgeReviewStatusScheduled, KnowledgeReviewStatusCompleted, KnowledgeReviewStatusExpired:
		return true
	default:
		return false
	}
}

// KnowledgeReview tracks periodic technical verification and accuracy certification of articles
type KnowledgeReview struct {
	ID            string     `json:"id"`
	WorkspaceID   string     `json:"workspace_id"`
	ArticleID     string     `json:"article_id"`
	ReviewerID    string     `json:"reviewer_id"`
	Status        string     `json:"status"`
	Findings      string     `json:"findings,omitempty"`
	NextReviewDue *time.Time `json:"next_review_due,omitempty"`
	ReviewedAt    *time.Time `json:"reviewed_at,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

// KnowledgeDeflection tracks search self-service interactions that prevent incident ticket escalation
type KnowledgeDeflection struct {
	ID          string    `json:"id"`
	WorkspaceID string    `json:"workspace_id"`
	Query       string    `json:"query"`
	ArticleID   string    `json:"article_id,omitempty"`
	UserID      string    `json:"user_id,omitempty"`
	Deflected   bool      `json:"deflected"`
	Context     string    `json:"context,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

// Request DTOs and validation rules for Knowledge API

// CreateKnowledgeArticleRequest represents the payload for creating an article
type CreateKnowledgeArticleRequest struct {
	WorkspaceID string `json:"workspace_id"`
	ArticleCode string `json:"article_code"`
	Title       string `json:"title"`
	Summary     string `json:"summary,omitempty"`
	Content     string `json:"content"`
	Category    string `json:"category"`
	Visibility  string `json:"visibility"`
	CIID        string `json:"ci_id,omitempty"`
	ProblemID   string `json:"problem_id,omitempty"`
	IncidentID  string `json:"incident_id,omitempty"`
	Tags        string `json:"tags,omitempty"`
	AuthorID    string `json:"author_id,omitempty"`
}

func (r *CreateKnowledgeArticleRequest) Validate() error {
	if strings.TrimSpace(r.WorkspaceID) == "" {
		return errors.New("workspace_id is required")
	}
	if strings.TrimSpace(r.ArticleCode) == "" {
		return errors.New("article_code is required")
	}
	if strings.TrimSpace(r.Title) == "" {
		return errors.New("title is required")
	}
	if strings.TrimSpace(r.Content) == "" {
		return errors.New("content is required")
	}
	if !IsValidKnowledgeCategory(r.Category) {
		return errors.New("invalid category")
	}
	if r.Visibility != "" && !IsValidKnowledgeVisibility(r.Visibility) {
		return errors.New("invalid visibility")
	}
	return nil
}

// UpdateKnowledgeArticleRequest represents the payload for modifying an article
type UpdateKnowledgeArticleRequest struct {
	Title         *string `json:"title,omitempty"`
	Summary       *string `json:"summary,omitempty"`
	Content       *string `json:"content,omitempty"`
	Category      *string `json:"category,omitempty"`
	Status        *string `json:"status,omitempty"`
	Visibility    *string `json:"visibility,omitempty"`
	CIID          *string `json:"ci_id,omitempty"`
	ProblemID     *string `json:"problem_id,omitempty"`
	IncidentID    *string `json:"incident_id,omitempty"`
	Tags          *string `json:"tags,omitempty"`
	ChangeSummary string  `json:"change_summary,omitempty"`
}

func (r *UpdateKnowledgeArticleRequest) Validate() error {
	if r.Category != nil && !IsValidKnowledgeCategory(*r.Category) {
		return errors.New("invalid category")
	}
	if r.Status != nil && !IsValidKnowledgeStatus(*r.Status) {
		return errors.New("invalid status")
	}
	if r.Visibility != nil && !IsValidKnowledgeVisibility(*r.Visibility) {
		return errors.New("invalid visibility")
	}
	return nil
}

// SubmitKnowledgeFeedbackRequest payload for voting and reviewing an article
type SubmitKnowledgeFeedbackRequest struct {
	WorkspaceID string `json:"workspace_id"`
	ArticleID   string `json:"article_id"`
	UserID      string `json:"user_id"`
	IsHelpful   bool   `json:"is_helpful"`
	Rating      int    `json:"rating"`
	Comment     string `json:"comment,omitempty"`
}

func (r *SubmitKnowledgeFeedbackRequest) Validate() error {
	if strings.TrimSpace(r.WorkspaceID) == "" {
		return errors.New("workspace_id is required")
	}
	if strings.TrimSpace(r.ArticleID) == "" {
		return errors.New("article_id is required")
	}
	if r.Rating < 1 || r.Rating > 5 {
		return errors.New("rating must be between 1 and 5")
	}
	return nil
}

// ScheduleKnowledgeReviewRequest payload for scheduling article verification
type ScheduleKnowledgeReviewRequest struct {
	WorkspaceID   string     `json:"workspace_id"`
	ArticleID     string     `json:"article_id"`
	ReviewerID    string     `json:"reviewer_id"`
	NextReviewDue *time.Time `json:"next_review_due,omitempty"`
}

func (r *ScheduleKnowledgeReviewRequest) Validate() error {
	if strings.TrimSpace(r.WorkspaceID) == "" {
		return errors.New("workspace_id is required")
	}
	if strings.TrimSpace(r.ArticleID) == "" {
		return errors.New("article_id is required")
	}
	if strings.TrimSpace(r.ReviewerID) == "" {
		return errors.New("reviewer_id is required")
	}
	return nil
}

// CompleteKnowledgeReviewRequest payload for certifying an article review
type CompleteKnowledgeReviewRequest struct {
	Findings      string     `json:"findings"`
	NextReviewDue *time.Time `json:"next_review_due,omitempty"`
}

func (r *CompleteKnowledgeReviewRequest) Validate() error {
	if strings.TrimSpace(r.Findings) == "" {
		return errors.New("findings is required")
	}
	return nil
}

// LogKnowledgeDeflectionRequest payload for self-service deflection metrics
type LogKnowledgeDeflectionRequest struct {
	WorkspaceID string `json:"workspace_id"`
	Query       string `json:"query"`
	ArticleID   string `json:"article_id,omitempty"`
	UserID      string `json:"user_id,omitempty"`
	Deflected   bool   `json:"deflected"`
	Context     string `json:"context,omitempty"`
}

func (r *LogKnowledgeDeflectionRequest) Validate() error {
	if strings.TrimSpace(r.WorkspaceID) == "" {
		return errors.New("workspace_id is required")
	}
	if strings.TrimSpace(r.Query) == "" {
		return errors.New("query is required")
	}
	return nil
}

// Filters and executive analytics models

// KnowledgeArticleFilter defines query criteria for listing articles
type KnowledgeArticleFilter struct {
	WorkspaceID string
	Category    string
	Status      string
	Visibility  string
	AuthorID    string
	CIID        string
	Search      string
	Limit       int
	Offset      int
}

// KnowledgeRevisionFilter defines query criteria for listing article revisions
type KnowledgeRevisionFilter struct {
	WorkspaceID string
	ArticleID   string
	Limit       int
	Offset      int
}

// KnowledgeFeedbackFilter defines query criteria for listing article feedback
type KnowledgeFeedbackFilter struct {
	WorkspaceID string
	ArticleID   string
	Limit       int
	Offset      int
}

// KnowledgeReviewFilter defines query criteria for listing reviews
type KnowledgeReviewFilter struct {
	WorkspaceID string
	ArticleID   string
	Status      string
	Limit       int
	Offset      int
}

// KnowledgeDeflectionFilter defines query criteria for deflection logs
type KnowledgeDeflectionFilter struct {
	WorkspaceID string
	ArticleID   string
	Limit       int
	Offset      int
}

// KnowledgeAnalytics aggregates executive Knowledge Base health and self-service metrics
type KnowledgeAnalytics struct {
	TotalArticles            int            `json:"total_articles"`
	PublishedArticles        int            `json:"published_articles"`
	DraftArticles            int            `json:"draft_articles"`
	InReviewArticles         int            `json:"in_review_articles"`
	ArchivedArticles         int            `json:"archived_articles"`
	TotalViews               int            `json:"total_views"`
	TotalHelpfulVotes        int            `json:"total_helpful_votes"`
	TotalNotHelpfulVotes     int            `json:"total_not_helpful_votes"`
	AverageHelpfulnessRatio  float64        `json:"average_helpfulness_ratio"`
	TotalDeflectionAttempts  int            `json:"total_deflection_attempts"`
	SuccessfulDeflections    int            `json:"successful_deflections"`
	DeflectionRate           float64        `json:"deflection_rate"`
	PendingReviewsCount      int            `json:"pending_reviews_count"`
	CategoryDistribution     map[string]int `json:"category_distribution"`
	KnowledgeHealthIndex     float64        `json:"knowledge_health_index"`
}
