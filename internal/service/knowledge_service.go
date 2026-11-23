package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/JohnMalugu/tsk-mgr-api/internal/model"
	"github.com/JohnMalugu/tsk-mgr-api/internal/repository"
)

var (
	_ = strings.TrimSpace
	_ = time.Now
	_ = errors.New
)

var (
	ErrKnowledgeArticleNotFound     = repository.ErrKnowledgeArticleNotFound
	ErrKnowledgeArticleCodeConflict = repository.ErrKnowledgeArticleCodeConflict
	ErrKnowledgeRevisionNotFound    = repository.ErrKnowledgeRevisionNotFound
	ErrKnowledgeFeedbackNotFound    = repository.ErrKnowledgeFeedbackNotFound
	ErrKnowledgeReviewNotFound      = repository.ErrKnowledgeReviewNotFound
	ErrInvalidArticleStatus         = errors.New("invalid article status")
)

// KnowledgeService defines business logic and orchestration for ITIL Knowledge Management (KEDB)
type KnowledgeService interface {
	// Articles
	CreateArticle(ctx context.Context, workspaceID string, req model.CreateKnowledgeArticleRequest) (*model.KnowledgeArticle, error)
	GetArticleByID(ctx context.Context, id string) (*model.KnowledgeArticle, error)
	GetArticleByCode(ctx context.Context, workspaceID, code string) (*model.KnowledgeArticle, error)
	UpdateArticle(ctx context.Context, id string, req model.UpdateKnowledgeArticleRequest) (*model.KnowledgeArticle, error)
	PublishArticle(ctx context.Context, id string) (*model.KnowledgeArticle, error)
	ArchiveArticle(ctx context.Context, id string) (*model.KnowledgeArticle, error)
	DeleteArticle(ctx context.Context, id string) error
	ListArticles(ctx context.Context, filter model.KnowledgeArticleFilter) ([]model.KnowledgeArticle, int, error)

	// Revisions
	GetRevisionByID(ctx context.Context, id string) (*model.KnowledgeRevision, error)
	ListRevisions(ctx context.Context, filter model.KnowledgeRevisionFilter) ([]model.KnowledgeRevision, int, error)

	// Feedback
	SubmitFeedback(ctx context.Context, workspaceID string, req model.SubmitKnowledgeFeedbackRequest) (*model.KnowledgeFeedback, error)
	ListFeedback(ctx context.Context, filter model.KnowledgeFeedbackFilter) ([]model.KnowledgeFeedback, int, error)

	// Reviews
	ScheduleReview(ctx context.Context, workspaceID string, req model.ScheduleKnowledgeReviewRequest) (*model.KnowledgeReview, error)
	GetReviewByID(ctx context.Context, id string) (*model.KnowledgeReview, error)
	CompleteReview(ctx context.Context, id string, req model.CompleteKnowledgeReviewRequest) (*model.KnowledgeReview, error)
	ListReviews(ctx context.Context, filter model.KnowledgeReviewFilter) ([]model.KnowledgeReview, int, error)

	// Deflections
	LogDeflection(ctx context.Context, workspaceID string, req model.LogKnowledgeDeflectionRequest) (*model.KnowledgeDeflection, error)
	ListDeflections(ctx context.Context, filter model.KnowledgeDeflectionFilter) ([]model.KnowledgeDeflection, int, error)

	// Posture Analytics
	GetKnowledgeAnalytics(ctx context.Context, workspaceID string) (*model.KnowledgeAnalytics, error)
}

type knowledgeService struct {
	repo repository.KnowledgeRepository
}

// NewKnowledgeService constructs an ITIL Knowledge Management service
func NewKnowledgeService(repo repository.KnowledgeRepository) KnowledgeService {
	return &knowledgeService{repo: repo}
}

// Articles implementation
func (s *knowledgeService) CreateArticle(ctx context.Context, workspaceID string, req model.CreateKnowledgeArticleRequest) (*model.KnowledgeArticle, error) {
	if req.WorkspaceID == "" {
		req.WorkspaceID = workspaceID
	}
	if err := req.Validate(); err != nil {
		return nil, err
	}

	visibility := req.Visibility
	if visibility == "" {
		visibility = model.KnowledgeVisibilityInternal
	}

	now := time.Now().UTC()
	article := &model.KnowledgeArticle{
		WorkspaceID: req.WorkspaceID,
		ArticleCode: strings.TrimSpace(req.ArticleCode),
		Title:       strings.TrimSpace(req.Title),
		Summary:     strings.TrimSpace(req.Summary),
		Content:     strings.TrimSpace(req.Content),
		Category:    req.Category,
		Status:      model.KnowledgeStatusDraft,
		Visibility:  visibility,
		Version:     1,
		AuthorID:    strings.TrimSpace(req.AuthorID),
		CIID:        strings.TrimSpace(req.CIID),
		ProblemID:   strings.TrimSpace(req.ProblemID),
		IncidentID:  strings.TrimSpace(req.IncidentID),
		Tags:        strings.TrimSpace(req.Tags),
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := s.repo.CreateArticle(ctx, article); err != nil {
		return nil, err
	}

	// Create initial version revision
	rev := &model.KnowledgeRevision{
		WorkspaceID:   article.WorkspaceID,
		ArticleID:     article.ID,
		Version:       1,
		Title:         article.Title,
		Content:       article.Content,
		Summary:       article.Summary,
		ChangeSummary: "Initial article draft created",
		AuthorID:      article.AuthorID,
		CreatedAt:     now,
	}
	_ = s.repo.CreateRevision(ctx, rev)

	return article, nil
}

func (s *knowledgeService) GetArticleByID(ctx context.Context, id string) (*model.KnowledgeArticle, error) {
	article, err := s.repo.GetArticleByID(ctx, id)
	if err != nil {
		return nil, err
	}
	_ = s.repo.IncrementViewCount(ctx, id)
	return article, nil
}

func (s *knowledgeService) GetArticleByCode(ctx context.Context, workspaceID, code string) (*model.KnowledgeArticle, error) {
	article, err := s.repo.GetArticleByCode(ctx, workspaceID, code)
	if err != nil {
		return nil, err
	}
	_ = s.repo.IncrementViewCount(ctx, article.ID)
	return article, nil
}

func (s *knowledgeService) UpdateArticle(ctx context.Context, id string, req model.UpdateKnowledgeArticleRequest) (*model.KnowledgeArticle, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	existing, err := s.repo.GetArticleByID(ctx, id)
	if err != nil {
		return nil, err
	}

	contentChanged := false
	if req.Title != nil && *req.Title != existing.Title {
		existing.Title = strings.TrimSpace(*req.Title)
		contentChanged = true
	}
	if req.Summary != nil {
		existing.Summary = strings.TrimSpace(*req.Summary)
	}
	if req.Content != nil && *req.Content != existing.Content {
		existing.Content = strings.TrimSpace(*req.Content)
		contentChanged = true
	}
	if req.Category != nil {
		existing.Category = *req.Category
	}
	if req.Status != nil {
		existing.Status = *req.Status
	}
	if req.Visibility != nil {
		existing.Visibility = *req.Visibility
	}
	if req.CIID != nil {
		existing.CIID = strings.TrimSpace(*req.CIID)
	}
	if req.ProblemID != nil {
		existing.ProblemID = strings.TrimSpace(*req.ProblemID)
	}
	if req.IncidentID != nil {
		existing.IncidentID = strings.TrimSpace(*req.IncidentID)
	}
	if req.Tags != nil {
		existing.Tags = strings.TrimSpace(*req.Tags)
	}

	if contentChanged {
		existing.Version++
		now := time.Now().UTC()
		changeSum := req.ChangeSummary
		if changeSum == "" {
			changeSum = "Content revised"
		}
		rev := &model.KnowledgeRevision{
			WorkspaceID:   existing.WorkspaceID,
			ArticleID:     existing.ID,
			Version:       existing.Version,
			Title:         existing.Title,
			Content:       existing.Content,
			Summary:       existing.Summary,
			ChangeSummary: changeSum,
			AuthorID:      existing.AuthorID,
			CreatedAt:     now,
		}
		_ = s.repo.CreateRevision(ctx, rev)
	}

	if err := s.repo.UpdateArticle(ctx, existing); err != nil {
		return nil, err
	}
	return existing, nil
}

func (s *knowledgeService) PublishArticle(ctx context.Context, id string) (*model.KnowledgeArticle, error) {
	existing, err := s.repo.GetArticleByID(ctx, id)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	existing.Status = model.KnowledgeStatusPublished
	existing.PublishedAt = &now

	if err := s.repo.UpdateArticle(ctx, existing); err != nil {
		return nil, err
	}
	return existing, nil
}

func (s *knowledgeService) ArchiveArticle(ctx context.Context, id string) (*model.KnowledgeArticle, error) {
	existing, err := s.repo.GetArticleByID(ctx, id)
	if err != nil {
		return nil, err
	}

	existing.Status = model.KnowledgeStatusArchived
	if err := s.repo.UpdateArticle(ctx, existing); err != nil {
		return nil, err
	}
	return existing, nil
}

func (s *knowledgeService) DeleteArticle(ctx context.Context, id string) error {
	return s.repo.DeleteArticle(ctx, id)
}

func (s *knowledgeService) ListArticles(ctx context.Context, filter model.KnowledgeArticleFilter) ([]model.KnowledgeArticle, int, error) {
	return s.repo.ListArticles(ctx, filter)
}

func (s *knowledgeService) GetRevisionByID(ctx context.Context, id string) (*model.KnowledgeRevision, error) {
	return s.repo.GetRevisionByID(ctx, id)
}

func (s *knowledgeService) ListRevisions(ctx context.Context, filter model.KnowledgeRevisionFilter) ([]model.KnowledgeRevision, int, error) {
	return s.repo.ListRevisions(ctx, filter)
}
// Feedback implementation
func (s *knowledgeService) SubmitFeedback(ctx context.Context, workspaceID string, req model.SubmitKnowledgeFeedbackRequest) (*model.KnowledgeFeedback, error) {
	if req.WorkspaceID == "" {
		req.WorkspaceID = workspaceID
	}
	if err := req.Validate(); err != nil {
		return nil, err
	}

	article, err := s.repo.GetArticleByID(ctx, req.ArticleID)
	if err != nil {
		return nil, err
	}

	feedback := &model.KnowledgeFeedback{
		WorkspaceID: req.WorkspaceID,
		ArticleID:   article.ID,
		UserID:      strings.TrimSpace(req.UserID),
		IsHelpful:   req.IsHelpful,
		Rating:      req.Rating,
		Comment:     strings.TrimSpace(req.Comment),
		CreatedAt:   time.Now().UTC(),
	}

	if err := s.repo.CreateFeedback(ctx, feedback); err != nil {
		return nil, err
	}
	return feedback, nil
}

func (s *knowledgeService) ListFeedback(ctx context.Context, filter model.KnowledgeFeedbackFilter) ([]model.KnowledgeFeedback, int, error) {
	return s.repo.ListFeedback(ctx, filter)
}

// Reviews implementation
func (s *knowledgeService) ScheduleReview(ctx context.Context, workspaceID string, req model.ScheduleKnowledgeReviewRequest) (*model.KnowledgeReview, error) {
	if req.WorkspaceID == "" {
		req.WorkspaceID = workspaceID
	}
	if err := req.Validate(); err != nil {
		return nil, err
	}

	article, err := s.repo.GetArticleByID(ctx, req.ArticleID)
	if err != nil {
		return nil, err
	}

	review := &model.KnowledgeReview{
		WorkspaceID:   req.WorkspaceID,
		ArticleID:     article.ID,
		ReviewerID:    strings.TrimSpace(req.ReviewerID),
		Status:        model.KnowledgeReviewStatusScheduled,
		NextReviewDue: req.NextReviewDue,
	}

	if err := s.repo.CreateReview(ctx, review); err != nil {
		return nil, err
	}
	return review, nil
}

func (s *knowledgeService) GetReviewByID(ctx context.Context, id string) (*model.KnowledgeReview, error) {
	return s.repo.GetReviewByID(ctx, id)
}

func (s *knowledgeService) CompleteReview(ctx context.Context, id string, req model.CompleteKnowledgeReviewRequest) (*model.KnowledgeReview, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	review, err := s.repo.GetReviewByID(ctx, id)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	review.Status = model.KnowledgeReviewStatusCompleted
	review.Findings = strings.TrimSpace(req.Findings)
	review.ReviewedAt = &now
	if req.NextReviewDue != nil {
		review.NextReviewDue = req.NextReviewDue
	}

	if err := s.repo.UpdateReview(ctx, review); err != nil {
		return nil, err
	}
	return review, nil
}

func (s *knowledgeService) ListReviews(ctx context.Context, filter model.KnowledgeReviewFilter) ([]model.KnowledgeReview, int, error) {
	return s.repo.ListReviews(ctx, filter)
}
// Deflections implementation
func (s *knowledgeService) LogDeflection(ctx context.Context, workspaceID string, req model.LogKnowledgeDeflectionRequest) (*model.KnowledgeDeflection, error) {
	if req.WorkspaceID == "" {
		req.WorkspaceID = workspaceID
	}
	if err := req.Validate(); err != nil {
		return nil, err
	}

	deflection := &model.KnowledgeDeflection{
		WorkspaceID: req.WorkspaceID,
		Query:       strings.TrimSpace(req.Query),
		ArticleID:   strings.TrimSpace(req.ArticleID),
		UserID:      strings.TrimSpace(req.UserID),
		Deflected:   req.Deflected,
		Context:     strings.TrimSpace(req.Context),
		CreatedAt:   time.Now().UTC(),
	}

	if err := s.repo.LogDeflection(ctx, deflection); err != nil {
		return nil, err
	}
	return deflection, nil
}

func (s *knowledgeService) ListDeflections(ctx context.Context, filter model.KnowledgeDeflectionFilter) ([]model.KnowledgeDeflection, int, error) {
	return s.repo.ListDeflections(ctx, filter)
}
// Posture Analytics implementation
func (s *knowledgeService) GetKnowledgeAnalytics(ctx context.Context, workspaceID string) (*model.KnowledgeAnalytics, error) {
	if strings.TrimSpace(workspaceID) == "" {
		return nil, errors.New("workspace_id is required")
	}
	return s.repo.GetKnowledgeAnalytics(ctx, workspaceID)
}
