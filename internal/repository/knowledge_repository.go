package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/JohnMalugu/tsk-mgr-api/internal/model"
)

var (
	_ = uuid.New
	_ = fmt.Sprintf
	_ = strings.Join
	_ = time.Now
	_ = errors.New
)

var (
	ErrKnowledgeArticleNotFound     = errors.New("knowledge article not found")
	ErrKnowledgeArticleCodeConflict = errors.New("article with this code already exists in workspace")
	ErrKnowledgeRevisionNotFound    = errors.New("knowledge revision not found")
	ErrKnowledgeFeedbackNotFound    = errors.New("knowledge feedback not found")
	ErrKnowledgeReviewNotFound      = errors.New("knowledge review not found")
)

// KnowledgeRepository defines data access methods for ITIL Knowledge Management (KEDB)
type KnowledgeRepository interface {
	// Articles
	CreateArticle(ctx context.Context, a *model.KnowledgeArticle) error
	GetArticleByID(ctx context.Context, id string) (*model.KnowledgeArticle, error)
	GetArticleByCode(ctx context.Context, workspaceID, code string) (*model.KnowledgeArticle, error)
	UpdateArticle(ctx context.Context, a *model.KnowledgeArticle) error
	DeleteArticle(ctx context.Context, id string) error
	IncrementViewCount(ctx context.Context, id string) error
	ListArticles(ctx context.Context, filter model.KnowledgeArticleFilter) ([]model.KnowledgeArticle, int, error)

	// Revisions
	CreateRevision(ctx context.Context, r *model.KnowledgeRevision) error
	GetRevisionByID(ctx context.Context, id string) (*model.KnowledgeRevision, error)
	ListRevisions(ctx context.Context, filter model.KnowledgeRevisionFilter) ([]model.KnowledgeRevision, int, error)

	// Feedback
	CreateFeedback(ctx context.Context, f *model.KnowledgeFeedback) error
	ListFeedback(ctx context.Context, filter model.KnowledgeFeedbackFilter) ([]model.KnowledgeFeedback, int, error)

	// Reviews
	CreateReview(ctx context.Context, rev *model.KnowledgeReview) error
	GetReviewByID(ctx context.Context, id string) (*model.KnowledgeReview, error)
	UpdateReview(ctx context.Context, rev *model.KnowledgeReview) error
	ListReviews(ctx context.Context, filter model.KnowledgeReviewFilter) ([]model.KnowledgeReview, int, error)

	// Deflections
	LogDeflection(ctx context.Context, d *model.KnowledgeDeflection) error
	ListDeflections(ctx context.Context, filter model.KnowledgeDeflectionFilter) ([]model.KnowledgeDeflection, int, error)

	// Posture Analytics
	GetKnowledgeAnalytics(ctx context.Context, workspaceID string) (*model.KnowledgeAnalytics, error)
}

type sqlKnowledgeRepository struct {
	db *sql.DB
}

// NewKnowledgeRepository initializes a SQLite knowledge repository
func NewKnowledgeRepository(db *sql.DB) KnowledgeRepository {
	return &sqlKnowledgeRepository{db: db}
}

// Articles implementation
func (r *sqlKnowledgeRepository) CreateArticle(ctx context.Context, a *model.KnowledgeArticle) error {
	if a.ID == "" {
		a.ID = uuid.New().String()
	}
	now := time.Now().UTC()
	if a.CreatedAt.IsZero() {
		a.CreatedAt = now
	}
	a.UpdatedAt = now
	if a.Version <= 0 {
		a.Version = 1
	}

	// Check for unique code in workspace
	var count int
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM knowledge_articles WHERE workspace_id = ? AND article_code = ?`, a.WorkspaceID, a.ArticleCode).Scan(&count)
	if err == nil && count > 0 {
		return ErrKnowledgeArticleCodeConflict
	}

	query := `
		INSERT INTO knowledge_articles (
			id, workspace_id, article_code, title, summary, content,
			category, status, visibility, version, author_id, ci_id,
			problem_id, incident_id, tags, view_count, helpful_count,
			not_helpful_count, review_due_date, published_at, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err = r.db.ExecContext(ctx, query,
		a.ID, a.WorkspaceID, a.ArticleCode, a.Title, a.Summary, a.Content,
		a.Category, a.Status, a.Visibility, a.Version, a.AuthorID, a.CIID,
		a.ProblemID, a.IncidentID, a.Tags, a.ViewCount, a.HelpfulCount,
		a.NotHelpfulCount, a.ReviewDueDate, a.PublishedAt, a.CreatedAt, a.UpdatedAt,
	)
	return err
}

func (r *sqlKnowledgeRepository) GetArticleByID(ctx context.Context, id string) (*model.KnowledgeArticle, error) {
	query := `
		SELECT id, workspace_id, article_code, title, summary, content,
		       category, status, visibility, version, author_id, ci_id,
		       problem_id, incident_id, tags, view_count, helpful_count,
		       not_helpful_count, review_due_date, published_at, created_at, updated_at
		FROM knowledge_articles
		WHERE id = ?
	`
	var a model.KnowledgeArticle
	var summary, author, ci, prob, inc, tags sql.NullString
	var reviewDue, pubAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&a.ID, &a.WorkspaceID, &a.ArticleCode, &a.Title, &summary, &a.Content,
		&a.Category, &a.Status, &a.Visibility, &a.Version, &author, &ci,
		&prob, &inc, &tags, &a.ViewCount, &a.HelpfulCount,
		&a.NotHelpfulCount, &reviewDue, &pubAt, &a.CreatedAt, &a.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrKnowledgeArticleNotFound
		}
		return nil, err
	}
	a.Summary = summary.String
	a.AuthorID = author.String
	a.CIID = ci.String
	a.ProblemID = prob.String
	a.IncidentID = inc.String
	a.Tags = tags.String
	if reviewDue.Valid {
		a.ReviewDueDate = &reviewDue.Time
	}
	if pubAt.Valid {
		a.PublishedAt = &pubAt.Time
	}
	return &a, nil
}

func (r *sqlKnowledgeRepository) GetArticleByCode(ctx context.Context, workspaceID, code string) (*model.KnowledgeArticle, error) {
	query := `
		SELECT id, workspace_id, article_code, title, summary, content,
		       category, status, visibility, version, author_id, ci_id,
		       problem_id, incident_id, tags, view_count, helpful_count,
		       not_helpful_count, review_due_date, published_at, created_at, updated_at
		FROM knowledge_articles
		WHERE workspace_id = ? AND article_code = ?
	`
	var a model.KnowledgeArticle
	var summary, author, ci, prob, inc, tags sql.NullString
	var reviewDue, pubAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, workspaceID, code).Scan(
		&a.ID, &a.WorkspaceID, &a.ArticleCode, &a.Title, &summary, &a.Content,
		&a.Category, &a.Status, &a.Visibility, &a.Version, &author, &ci,
		&prob, &inc, &tags, &a.ViewCount, &a.HelpfulCount,
		&a.NotHelpfulCount, &reviewDue, &pubAt, &a.CreatedAt, &a.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrKnowledgeArticleNotFound
		}
		return nil, err
	}
	a.Summary = summary.String
	a.AuthorID = author.String
	a.CIID = ci.String
	a.ProblemID = prob.String
	a.IncidentID = inc.String
	a.Tags = tags.String
	if reviewDue.Valid {
		a.ReviewDueDate = &reviewDue.Time
	}
	if pubAt.Valid {
		a.PublishedAt = &pubAt.Time
	}
	return &a, nil
}

func (r *sqlKnowledgeRepository) UpdateArticle(ctx context.Context, a *model.KnowledgeArticle) error {
	a.UpdatedAt = time.Now().UTC()
	query := `
		UPDATE knowledge_articles
		SET title = ?, summary = ?, content = ?, category = ?, status = ?,
		    visibility = ?, version = ?, ci_id = ?, problem_id = ?, incident_id = ?,
		    tags = ?, helpful_count = ?, not_helpful_count = ?, review_due_date = ?,
		    published_at = ?, updated_at = ?
		WHERE id = ?
	`
	res, err := r.db.ExecContext(ctx, query,
		a.Title, a.Summary, a.Content, a.Category, a.Status,
		a.Visibility, a.Version, a.CIID, a.ProblemID, a.IncidentID,
		a.Tags, a.HelpfulCount, a.NotHelpfulCount, a.ReviewDueDate,
		a.PublishedAt, a.UpdatedAt, a.ID,
	)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrKnowledgeArticleNotFound
	}
	return nil
}

func (r *sqlKnowledgeRepository) DeleteArticle(ctx context.Context, id string) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM knowledge_articles WHERE id = ?`, id)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrKnowledgeArticleNotFound
	}
	return nil
}

func (r *sqlKnowledgeRepository) IncrementViewCount(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE knowledge_articles SET view_count = view_count + 1 WHERE id = ?`, id)
	return err
}

func (r *sqlKnowledgeRepository) ListArticles(ctx context.Context, filter model.KnowledgeArticleFilter) ([]model.KnowledgeArticle, int, error) {
	var conditions []string
	var args []interface{}

	if filter.WorkspaceID != "" {
		conditions = append(conditions, "workspace_id = ?")
		args = append(args, filter.WorkspaceID)
	}
	if filter.Category != "" {
		conditions = append(conditions, "category = ?")
		args = append(args, filter.Category)
	}
	if filter.Status != "" {
		conditions = append(conditions, "status = ?")
		args = append(args, filter.Status)
	}
	if filter.Visibility != "" {
		conditions = append(conditions, "visibility = ?")
		args = append(args, filter.Visibility)
	}
	if filter.AuthorID != "" {
		conditions = append(conditions, "author_id = ?")
		args = append(args, filter.AuthorID)
	}
	if filter.CIID != "" {
		conditions = append(conditions, "ci_id = ?")
		args = append(args, filter.CIID)
	}
	if filter.Search != "" {
		searchTerm := "%" + filter.Search + "%"
		conditions = append(conditions, "(title LIKE ? OR summary LIKE ? OR tags LIKE ?)")
		args = append(args, searchTerm, searchTerm, searchTerm)
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM knowledge_articles %s", whereClause)
	var total int
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	limit := filter.Limit
	if limit <= 0 {
		limit = 50
	}
	offset := filter.Offset
	if offset < 0 {
		offset = 0
	}

	query := fmt.Sprintf(`
		SELECT id, workspace_id, article_code, title, summary, content,
		       category, status, visibility, version, author_id, ci_id,
		       problem_id, incident_id, tags, view_count, helpful_count,
		       not_helpful_count, review_due_date, published_at, created_at, updated_at
		FROM knowledge_articles
		%s
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`, whereClause)

	argsWithPaging := append(args, limit, offset)
	rows, err := r.db.QueryContext(ctx, query, argsWithPaging...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var articles []model.KnowledgeArticle
	for rows.Next() {
		var a model.KnowledgeArticle
		var summary, author, ci, prob, inc, tags sql.NullString
		var reviewDue, pubAt sql.NullTime

		err := rows.Scan(
			&a.ID, &a.WorkspaceID, &a.ArticleCode, &a.Title, &summary, &a.Content,
			&a.Category, &a.Status, &a.Visibility, &a.Version, &author, &ci,
			&prob, &inc, &tags, &a.ViewCount, &a.HelpfulCount,
			&a.NotHelpfulCount, &reviewDue, &pubAt, &a.CreatedAt, &a.UpdatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		a.Summary = summary.String
		a.AuthorID = author.String
		a.CIID = ci.String
		a.ProblemID = prob.String
		a.IncidentID = inc.String
		a.Tags = tags.String
		if reviewDue.Valid {
			a.ReviewDueDate = &reviewDue.Time
		}
		if pubAt.Valid {
			a.PublishedAt = &pubAt.Time
		}
		articles = append(articles, a)
	}
	return articles, total, rows.Err()
}
// Revisions implementation
func (r *sqlKnowledgeRepository) CreateRevision(ctx context.Context, rev *model.KnowledgeRevision) error {
	if rev.ID == "" {
		rev.ID = uuid.New().String()
	}
	if rev.CreatedAt.IsZero() {
		rev.CreatedAt = time.Now().UTC()
	}

	query := `
		INSERT INTO knowledge_revisions (
			id, workspace_id, article_id, version, title, content,
			summary, change_summary, author_id, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err := r.db.ExecContext(ctx, query,
		rev.ID, rev.WorkspaceID, rev.ArticleID, rev.Version, rev.Title, rev.Content,
		rev.Summary, rev.ChangeSummary, rev.AuthorID, rev.CreatedAt,
	)
	return err
}

func (r *sqlKnowledgeRepository) GetRevisionByID(ctx context.Context, id string) (*model.KnowledgeRevision, error) {
	query := `
		SELECT id, workspace_id, article_id, version, title, content,
		       summary, change_summary, author_id, created_at
		FROM knowledge_revisions
		WHERE id = ?
	`
	var rev model.KnowledgeRevision
	var summary, changeSum, author sql.NullString

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&rev.ID, &rev.WorkspaceID, &rev.ArticleID, &rev.Version, &rev.Title, &rev.Content,
		&summary, &changeSum, &author, &rev.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrKnowledgeRevisionNotFound
		}
		return nil, err
	}
	rev.Summary = summary.String
	rev.ChangeSummary = changeSum.String
	rev.AuthorID = author.String
	return &rev, nil
}

func (r *sqlKnowledgeRepository) ListRevisions(ctx context.Context, filter model.KnowledgeRevisionFilter) ([]model.KnowledgeRevision, int, error) {
	var conditions []string
	var args []interface{}

	if filter.WorkspaceID != "" {
		conditions = append(conditions, "workspace_id = ?")
		args = append(args, filter.WorkspaceID)
	}
	if filter.ArticleID != "" {
		conditions = append(conditions, "article_id = ?")
		args = append(args, filter.ArticleID)
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM knowledge_revisions %s", whereClause)
	var total int
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	limit := filter.Limit
	if limit <= 0 {
		limit = 50
	}
	offset := filter.Offset
	if offset < 0 {
		offset = 0
	}

	query := fmt.Sprintf(`
		SELECT id, workspace_id, article_id, version, title, content,
		       summary, change_summary, author_id, created_at
		FROM knowledge_revisions
		%s
		ORDER BY version DESC
		LIMIT ? OFFSET ?
	`, whereClause)

	argsWithPaging := append(args, limit, offset)
	rows, err := r.db.QueryContext(ctx, query, argsWithPaging...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var revisions []model.KnowledgeRevision
	for rows.Next() {
		var rev model.KnowledgeRevision
		var summary, changeSum, author sql.NullString

		err := rows.Scan(
			&rev.ID, &rev.WorkspaceID, &rev.ArticleID, &rev.Version, &rev.Title, &rev.Content,
			&summary, &changeSum, &author, &rev.CreatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		rev.Summary = summary.String
		rev.ChangeSummary = changeSum.String
		rev.AuthorID = author.String
		revisions = append(revisions, rev)
	}
	return revisions, total, rows.Err()
}
// Feedback implementation
func (r *sqlKnowledgeRepository) CreateFeedback(ctx context.Context, f *model.KnowledgeFeedback) error {
	if f.ID == "" {
		f.ID = uuid.New().String()
	}
	if f.CreatedAt.IsZero() {
		f.CreatedAt = time.Now().UTC()
	}

	query := `
		INSERT INTO knowledge_feedback (
			id, workspace_id, article_id, user_id, is_helpful, rating, comment, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err := r.db.ExecContext(ctx, query,
		f.ID, f.WorkspaceID, f.ArticleID, f.UserID, f.IsHelpful, f.Rating, f.Comment, f.CreatedAt,
	)
	if err != nil {
		return err
	}

	// Increment helpful or not helpful counter on article
	if f.IsHelpful {
		_, _ = r.db.ExecContext(ctx, `UPDATE knowledge_articles SET helpful_count = helpful_count + 1 WHERE id = ?`, f.ArticleID)
	} else {
		_, _ = r.db.ExecContext(ctx, `UPDATE knowledge_articles SET not_helpful_count = not_helpful_count + 1 WHERE id = ?`, f.ArticleID)
	}
	return nil
}

func (r *sqlKnowledgeRepository) ListFeedback(ctx context.Context, filter model.KnowledgeFeedbackFilter) ([]model.KnowledgeFeedback, int, error) {
	var conditions []string
	var args []interface{}

	if filter.WorkspaceID != "" {
		conditions = append(conditions, "workspace_id = ?")
		args = append(args, filter.WorkspaceID)
	}
	if filter.ArticleID != "" {
		conditions = append(conditions, "article_id = ?")
		args = append(args, filter.ArticleID)
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM knowledge_feedback %s", whereClause)
	var total int
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	limit := filter.Limit
	if limit <= 0 {
		limit = 50
	}
	offset := filter.Offset
	if offset < 0 {
		offset = 0
	}

	query := fmt.Sprintf(`
		SELECT id, workspace_id, article_id, user_id, is_helpful, rating, comment, created_at
		FROM knowledge_feedback
		%s
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`, whereClause)

	argsWithPaging := append(args, limit, offset)
	rows, err := r.db.QueryContext(ctx, query, argsWithPaging...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var list []model.KnowledgeFeedback
	for rows.Next() {
		var f model.KnowledgeFeedback
		var comment sql.NullString
		err := rows.Scan(
			&f.ID, &f.WorkspaceID, &f.ArticleID, &f.UserID, &f.IsHelpful,
			&f.Rating, &comment, &f.CreatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		f.Comment = comment.String
		list = append(list, f)
	}
	return list, total, rows.Err()
}
// Reviews implementation
func (r *sqlKnowledgeRepository) CreateReview(ctx context.Context, rev *model.KnowledgeReview) error {
	if rev.ID == "" {
		rev.ID = uuid.New().String()
	}
	now := time.Now().UTC()
	if rev.CreatedAt.IsZero() {
		rev.CreatedAt = now
	}
	rev.UpdatedAt = now
	if rev.Status == "" {
		rev.Status = model.KnowledgeReviewStatusScheduled
	}

	query := `
		INSERT INTO knowledge_reviews (
			id, workspace_id, article_id, reviewer_id, status,
			findings, next_review_due, reviewed_at, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err := r.db.ExecContext(ctx, query,
		rev.ID, rev.WorkspaceID, rev.ArticleID, rev.ReviewerID, rev.Status,
		rev.Findings, rev.NextReviewDue, rev.ReviewedAt, rev.CreatedAt, rev.UpdatedAt,
	)
	if err != nil {
		return err
	}

	if rev.NextReviewDue != nil {
		_, _ = r.db.ExecContext(ctx, `UPDATE knowledge_articles SET review_due_date = ?, updated_at = ? WHERE id = ?`, rev.NextReviewDue, now, rev.ArticleID)
	}
	return nil
}

func (r *sqlKnowledgeRepository) GetReviewByID(ctx context.Context, id string) (*model.KnowledgeReview, error) {
	query := `
		SELECT id, workspace_id, article_id, reviewer_id, status,
		       findings, next_review_due, reviewed_at, created_at, updated_at
		FROM knowledge_reviews
		WHERE id = ?
	`
	var rev model.KnowledgeReview
	var findings sql.NullString
	var nextDue, reviewedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&rev.ID, &rev.WorkspaceID, &rev.ArticleID, &rev.ReviewerID, &rev.Status,
		&findings, &nextDue, &reviewedAt, &rev.CreatedAt, &rev.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrKnowledgeReviewNotFound
		}
		return nil, err
	}
	rev.Findings = findings.String
	if nextDue.Valid {
		rev.NextReviewDue = &nextDue.Time
	}
	if reviewedAt.Valid {
		rev.ReviewedAt = &reviewedAt.Time
	}
	return &rev, nil
}

func (r *sqlKnowledgeRepository) UpdateReview(ctx context.Context, rev *model.KnowledgeReview) error {
	rev.UpdatedAt = time.Now().UTC()
	query := `
		UPDATE knowledge_reviews
		SET status = ?, findings = ?, next_review_due = ?, reviewed_at = ?, updated_at = ?
		WHERE id = ?
	`
	res, err := r.db.ExecContext(ctx, query,
		rev.Status, rev.Findings, rev.NextReviewDue, rev.ReviewedAt, rev.UpdatedAt, rev.ID,
	)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrKnowledgeReviewNotFound
	}

	if rev.NextReviewDue != nil {
		_, _ = r.db.ExecContext(ctx, `UPDATE knowledge_articles SET review_due_date = ?, updated_at = ? WHERE id = ?`, rev.NextReviewDue, rev.UpdatedAt, rev.ArticleID)
	}
	return nil
}

func (r *sqlKnowledgeRepository) ListReviews(ctx context.Context, filter model.KnowledgeReviewFilter) ([]model.KnowledgeReview, int, error) {
	var conditions []string
	var args []interface{}

	if filter.WorkspaceID != "" {
		conditions = append(conditions, "workspace_id = ?")
		args = append(args, filter.WorkspaceID)
	}
	if filter.ArticleID != "" {
		conditions = append(conditions, "article_id = ?")
		args = append(args, filter.ArticleID)
	}
	if filter.Status != "" {
		conditions = append(conditions, "status = ?")
		args = append(args, filter.Status)
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM knowledge_reviews %s", whereClause)
	var total int
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	limit := filter.Limit
	if limit <= 0 {
		limit = 50
	}
	offset := filter.Offset
	if offset < 0 {
		offset = 0
	}

	query := fmt.Sprintf(`
		SELECT id, workspace_id, article_id, reviewer_id, status,
		       findings, next_review_due, reviewed_at, created_at, updated_at
		FROM knowledge_reviews
		%s
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`, whereClause)

	argsWithPaging := append(args, limit, offset)
	rows, err := r.db.QueryContext(ctx, query, argsWithPaging...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var reviews []model.KnowledgeReview
	for rows.Next() {
		var rev model.KnowledgeReview
		var findings sql.NullString
		var nextDue, reviewedAt sql.NullTime

		err := rows.Scan(
			&rev.ID, &rev.WorkspaceID, &rev.ArticleID, &rev.ReviewerID, &rev.Status,
			&findings, &nextDue, &reviewedAt, &rev.CreatedAt, &rev.UpdatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		rev.Findings = findings.String
		if nextDue.Valid {
			rev.NextReviewDue = &nextDue.Time
		}
		if reviewedAt.Valid {
			rev.ReviewedAt = &reviewedAt.Time
		}
		reviews = append(reviews, rev)
	}
	return reviews, total, rows.Err()
}

// Deflections implementation
func (r *sqlKnowledgeRepository) LogDeflection(ctx context.Context, d *model.KnowledgeDeflection) error {
	if d.ID == "" {
		d.ID = uuid.New().String()
	}
	if d.CreatedAt.IsZero() {
		d.CreatedAt = time.Now().UTC()
	}

	query := `
		INSERT INTO knowledge_deflections (
			id, workspace_id, query, article_id, user_id, deflected, context, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err := r.db.ExecContext(ctx, query,
		d.ID, d.WorkspaceID, d.Query, d.ArticleID, d.UserID, d.Deflected, d.Context, d.CreatedAt,
	)
	return err
}

func (r *sqlKnowledgeRepository) ListDeflections(ctx context.Context, filter model.KnowledgeDeflectionFilter) ([]model.KnowledgeDeflection, int, error) {
	var conditions []string
	var args []interface{}

	if filter.WorkspaceID != "" {
		conditions = append(conditions, "workspace_id = ?")
		args = append(args, filter.WorkspaceID)
	}
	if filter.ArticleID != "" {
		conditions = append(conditions, "article_id = ?")
		args = append(args, filter.ArticleID)
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM knowledge_deflections %s", whereClause)
	var total int
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	limit := filter.Limit
	if limit <= 0 {
		limit = 50
	}
	offset := filter.Offset
	if offset < 0 {
		offset = 0
	}

	query := fmt.Sprintf(`
		SELECT id, workspace_id, query, article_id, user_id, deflected, context, created_at
		FROM knowledge_deflections
		%s
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`, whereClause)

	argsWithPaging := append(args, limit, offset)
	rows, err := r.db.QueryContext(ctx, query, argsWithPaging...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var deflections []model.KnowledgeDeflection
	for rows.Next() {
		var d model.KnowledgeDeflection
		var articleID, userID, contextStr sql.NullString

		err := rows.Scan(
			&d.ID, &d.WorkspaceID, &d.Query, &articleID, &userID,
			&d.Deflected, &contextStr, &d.CreatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		d.ArticleID = articleID.String
		d.UserID = userID.String
		d.Context = contextStr.String
		deflections = append(deflections, d)
	}
	return deflections, total, rows.Err()
}
// Posture Analytics implementation
func (r *sqlKnowledgeRepository) GetKnowledgeAnalytics(ctx context.Context, workspaceID string) (*model.KnowledgeAnalytics, error) {
	analytics := &model.KnowledgeAnalytics{
		CategoryDistribution: make(map[string]int),
	}

	// 1. Article counts by status
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM knowledge_articles WHERE workspace_id = ?`, workspaceID).Scan(&analytics.TotalArticles)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	_ = r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM knowledge_articles WHERE workspace_id = ? AND status = ?`, workspaceID, model.KnowledgeStatusPublished).Scan(&analytics.PublishedArticles)
	_ = r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM knowledge_articles WHERE workspace_id = ? AND status = ?`, workspaceID, model.KnowledgeStatusDraft).Scan(&analytics.DraftArticles)
	_ = r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM knowledge_articles WHERE workspace_id = ? AND status = ?`, workspaceID, model.KnowledgeStatusInReview).Scan(&analytics.InReviewArticles)
	_ = r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM knowledge_articles WHERE workspace_id = ? AND status = ?`, workspaceID, model.KnowledgeStatusArchived).Scan(&analytics.ArchivedArticles)

	// 2. Engagement & helpfulness metrics
	_ = r.db.QueryRowContext(ctx, `
		SELECT COALESCE(SUM(view_count), 0), COALESCE(SUM(helpful_count), 0), COALESCE(SUM(not_helpful_count), 0)
		FROM knowledge_articles
		WHERE workspace_id = ?
	`, workspaceID).Scan(&analytics.TotalViews, &analytics.TotalHelpfulVotes, &analytics.TotalNotHelpfulVotes)

	totalVotes := analytics.TotalHelpfulVotes + analytics.TotalNotHelpfulVotes
	if totalVotes > 0 {
		analytics.AverageHelpfulnessRatio = (float64(analytics.TotalHelpfulVotes) / float64(totalVotes)) * 100.0
	}

	// 3. Deflection metrics
	_ = r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM knowledge_deflections WHERE workspace_id = ?`, workspaceID).Scan(&analytics.TotalDeflectionAttempts)
	_ = r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM knowledge_deflections WHERE workspace_id = ? AND deflected = 1`, workspaceID).Scan(&analytics.SuccessfulDeflections)

	if analytics.TotalDeflectionAttempts > 0 {
		analytics.DeflectionRate = (float64(analytics.SuccessfulDeflections) / float64(analytics.TotalDeflectionAttempts)) * 100.0
	}

	// 4. Verification review backlog
	_ = r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM knowledge_reviews WHERE workspace_id = ? AND status = ?`, workspaceID, model.KnowledgeReviewStatusScheduled).Scan(&analytics.PendingReviewsCount)

	// 5. Category distribution
	catRows, err := r.db.QueryContext(ctx, `SELECT category, COUNT(*) FROM knowledge_articles WHERE workspace_id = ? GROUP BY category`, workspaceID)
	if err == nil {
		defer catRows.Close()
		for catRows.Next() {
			var cat string
			var count int
			if err := catRows.Scan(&cat, &count); err == nil {
				analytics.CategoryDistribution[cat] = count
			}
		}
	}

	// 6. Knowledge Base Health Index (0 - 100)
	score := 50.0 // baseline starting score
	if analytics.TotalArticles > 0 {
		pubRatio := float64(analytics.PublishedArticles) / float64(analytics.TotalArticles)
		score += (pubRatio * 25.0)
	}
	if totalVotes > 0 {
		score += ((analytics.AverageHelpfulnessRatio / 100.0) * 20.0)
	} else {
		score += 10.0 // neutral if no votes yet
	}
	if analytics.TotalDeflectionAttempts > 0 {
		score += ((analytics.DeflectionRate / 100.0) * 15.0)
	} else {
		score += 5.0
	}

	// Penalty for review backlog
	if analytics.PendingReviewsCount > 0 {
		penalty := float64(analytics.PendingReviewsCount) * 1.5
		if penalty > 15.0 {
			penalty = 15.0
		}
		score -= penalty
	}

	if score < 0.0 {
		score = 0.0
	}
	if score > 100.0 {
		score = 100.0
	}
	analytics.KnowledgeHealthIndex = score

	return analytics, nil
}
