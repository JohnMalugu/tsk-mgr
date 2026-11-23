package service

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/JohnMalugu/tsk-mgr-api/internal/db"
	"github.com/JohnMalugu/tsk-mgr-api/internal/model"
	"github.com/JohnMalugu/tsk-mgr-api/internal/repository"
)

var _ = time.Now

func setupTestKnowledgeSvc(t *testing.T) (KnowledgeService, repository.KnowledgeRepository, *sql.DB) {
	sqliteDB, err := db.InitDB(":memory:")
	if err != nil {
		t.Fatalf("failed to initialize test in-memory db: %v", err)
	}
	sqliteDB.SetMaxOpenConns(1)

	_, err = sqliteDB.Exec(`
		INSERT INTO users (id, username, email, password_hash, created_at, updated_at) 
		VALUES ('user_1', 'alice', 'alice@example.com', 'hash', datetime('now'), datetime('now')),
		       ('user_2', 'bob', 'bob@example.com', 'hash', datetime('now'), datetime('now'));

		INSERT INTO workspaces (id, name, slug, owner_id, created_at, updated_at) 
		VALUES (1, 'Knowledge WS', 'kb-ws', 'user_1', datetime('now'), datetime('now'));
	`)
	if err != nil {
		t.Fatalf("failed to insert test seed data: %v", err)
	}

	repo := repository.NewKnowledgeRepository(sqliteDB)
	svc := NewKnowledgeService(repo)
	return svc, repo, sqliteDB
}

func TestKnowledge_ArticleLifecycle(t *testing.T) {
	svc, _, dbConn := setupTestKnowledgeSvc(t)
	defer dbConn.Close()

	ctx := context.Background()

	// 1. Validation failure: missing article_code
	_, err := svc.CreateArticle(ctx, "ws_1", model.CreateKnowledgeArticleRequest{
		Title:    "How to reset VPN credentials",
		Content:  "Follow these steps to reset credentials...",
		Category: model.KnowledgeCategoryHowTo,
	})
	if err == nil {
		t.Fatal("expected error for missing article_code, got nil")
	}

	// 2. Validation failure: invalid category
	_, err = svc.CreateArticle(ctx, "ws_1", model.CreateKnowledgeArticleRequest{
		ArticleCode: "KB-001",
		Title:       "Invalid Category Article",
		Content:     "Content goes here",
		Category:    "unsupported_category",
	})
	if err == nil {
		t.Fatal("expected error for invalid category, got nil")
	}

	// 3. Create draft article successfully
	article, err := svc.CreateArticle(ctx, "ws_1", model.CreateKnowledgeArticleRequest{
		ArticleCode: "KB-VPN-001",
		Title:       "Enterprise GlobalProtect VPN Setup",
		Summary:     "Complete guide for onboarding new team members to corporate VPN",
		Content:     "# Prerequisites\nDownload the official client...",
		Category:    model.KnowledgeCategoryHowTo,
		Visibility:  model.KnowledgeVisibilityInternal,
		CIID:        "ci-net-vpn",
		Tags:        "vpn,network,security",
		AuthorID:    "user_1",
	})
	if err != nil {
		t.Fatalf("failed to create article: %v", err)
	}
	if article.ID == "" || article.Status != model.KnowledgeStatusDraft || article.Version != 1 {
		t.Fatalf("unexpected article state: %+v", article)
	}

	// 4. Duplicate article_code conflict
	_, err = svc.CreateArticle(ctx, "ws_1", model.CreateKnowledgeArticleRequest{
		ArticleCode: "KB-VPN-001",
		Title:       "Duplicate Code Article",
		Content:     "Duplicate content",
		Category:    model.KnowledgeCategoryTroubleshooting,
	})
	if err != ErrKnowledgeArticleCodeConflict {
		t.Fatalf("expected ErrKnowledgeArticleCodeConflict, got %v", err)
	}

	// 5. Get article by ID and by code (increments view count)
	fetched, err := svc.GetArticleByID(ctx, article.ID)
	if err != nil {
		t.Fatalf("failed to get article by ID: %v", err)
	}
	if fetched.Title != "Enterprise GlobalProtect VPN Setup" {
		t.Fatalf("unexpected article title: %s", fetched.Title)
	}

	byCode, err := svc.GetArticleByCode(ctx, "ws_1", "KB-VPN-001")
	if err != nil {
		t.Fatalf("failed to get article by code: %v", err)
	}
	if byCode.ID != article.ID {
		t.Fatalf("ID mismatch: %s vs %s", byCode.ID, article.ID)
	}
	if byCode.ViewCount < 2 {
		t.Fatalf("expected view_count >= 2, got %d", byCode.ViewCount)
	}

	// 6. Update article content (creates v2 revision)
	newContent := "# Prerequisites\nUpdated with MFA verification steps..."
	changeReason := "Added required Okta MFA verification requirements"
	updated, err := svc.UpdateArticle(ctx, article.ID, model.UpdateKnowledgeArticleRequest{
		Content:       &newContent,
		ChangeSummary: changeReason,
	})
	if err != nil {
		t.Fatalf("failed to update article: %v", err)
	}
	if updated.Version != 2 || updated.Content != newContent {
		t.Fatalf("unexpected updated article state: %+v", updated)
	}

	// 7. Publish article
	published, err := svc.PublishArticle(ctx, article.ID)
	if err != nil {
		t.Fatalf("failed to publish article: %v", err)
	}
	if published.Status != model.KnowledgeStatusPublished || published.PublishedAt == nil {
		t.Fatalf("unexpected published state: %+v", published)
	}

	// 8. Archive article
	archived, err := svc.ArchiveArticle(ctx, article.ID)
	if err != nil {
		t.Fatalf("failed to archive article: %v", err)
	}
	if archived.Status != model.KnowledgeStatusArchived {
		t.Fatalf("unexpected archived state: %+v", archived)
	}

	// 9. List articles with search
	list, total, err := svc.ListArticles(ctx, model.KnowledgeArticleFilter{
		WorkspaceID: "ws_1",
		Search:      "GlobalProtect",
	})
	if err != nil {
		t.Fatalf("failed to list articles: %v", err)
	}
	if total != 1 || len(list) != 1 {
		t.Fatalf("expected 1 article, got %d (total %d)", len(list), total)
	}

	// 10. Delete article
	if err := svc.DeleteArticle(ctx, article.ID); err != nil {
		t.Fatalf("failed to delete article: %v", err)
	}
	_, err = svc.GetArticleByID(ctx, article.ID)
	if err != ErrKnowledgeArticleNotFound {
		t.Fatalf("expected ErrKnowledgeArticleNotFound, got %v", err)
	}
}

func TestKnowledge_Revisions(t *testing.T) {
	svc, _, dbConn := setupTestKnowledgeSvc(t)
	defer dbConn.Close()

	ctx := context.Background()

	article, err := svc.CreateArticle(ctx, "ws_1", model.CreateKnowledgeArticleRequest{
		ArticleCode: "KB-REV-001",
		Title:       "Kubernetes Pod Eviction Troubleshooting",
		Content:     "Initial diagnostic commands: kubectl describe pod...",
		Category:    model.KnowledgeCategoryTroubleshooting,
		AuthorID:    "user_1",
	})
	if err != nil {
		t.Fatalf("failed to create article: %v", err)
	}

	// Update 1 (v2)
	v2Content := "Updated with node disk pressure investigation"
	_, _ = svc.UpdateArticle(ctx, article.ID, model.UpdateKnowledgeArticleRequest{
		Content:       &v2Content,
		ChangeSummary: "Added disk pressure commands",
	})

	// Update 2 (v3)
	v3Content := "Updated with OOMKilled cgroup investigation"
	_, _ = svc.UpdateArticle(ctx, article.ID, model.UpdateKnowledgeArticleRequest{
		Content:       &v3Content,
		ChangeSummary: "Added OOM analysis steps",
	})

	// List revisions
	revs, total, err := svc.ListRevisions(ctx, model.KnowledgeRevisionFilter{
		WorkspaceID: "ws_1",
		ArticleID:   article.ID,
	})
	if err != nil {
		t.Fatalf("failed to list revisions: %v", err)
	}
	if total != 3 || len(revs) != 3 {
		t.Fatalf("expected 3 revisions, got %d (total %d)", len(revs), total)
	}
	if revs[0].Version != 3 || revs[1].Version != 2 || revs[2].Version != 1 {
		t.Fatalf("expected descending order versions: %d, %d, %d", revs[0].Version, revs[1].Version, revs[2].Version)
	}

	// Fetch single revision
	r2, err := svc.GetRevisionByID(ctx, revs[1].ID)
	if err != nil {
		t.Fatalf("failed to get revision: %v", err)
	}
	if r2.Version != 2 || r2.ChangeSummary != "Added disk pressure commands" {
		t.Fatalf("unexpected revision data: %+v", r2)
	}
}

func TestKnowledge_FeedbackAndRating(t *testing.T) {
	svc, _, dbConn := setupTestKnowledgeSvc(t)
	defer dbConn.Close()

	ctx := context.Background()

	article, err := svc.CreateArticle(ctx, "ws_1", model.CreateKnowledgeArticleRequest{
		ArticleCode: "KB-RATING-01",
		Title:       "OAuth2 Token Refresh Protocol",
		Content:     "Technical specification for refreshing user tokens",
		Category:    model.KnowledgeCategoryReference,
	})
	if err != nil {
		t.Fatalf("failed to create article: %v", err)
	}

	// 1. Validation failure: invalid rating
	_, err = svc.SubmitFeedback(ctx, "ws_1", model.SubmitKnowledgeFeedbackRequest{
		ArticleID: article.ID,
		UserID:    "user_2",
		IsHelpful: true,
		Rating:    10, // out of bounds
	})
	if err == nil {
		t.Fatal("expected error for rating > 5, got nil")
	}

	// 2. Submit helpful feedback
	f1, err := svc.SubmitFeedback(ctx, "ws_1", model.SubmitKnowledgeFeedbackRequest{
		ArticleID: article.ID,
		UserID:    "user_1",
		IsHelpful: true,
		Rating:    5,
		Comment:   "Resolved my token expiry issue immediately!",
	})
	if err != nil {
		t.Fatalf("failed to submit helpful feedback: %v", err)
	}
	if !f1.IsHelpful || f1.Rating != 5 {
		t.Fatalf("unexpected feedback: %+v", f1)
	}

	// 3. Submit unhelpful feedback
	_, err = svc.SubmitFeedback(ctx, "ws_1", model.SubmitKnowledgeFeedbackRequest{
		ArticleID: article.ID,
		UserID:    "user_2",
		IsHelpful: false,
		Rating:    2,
		Comment:   "Missing code snippet for Go client",
	})
	if err != nil {
		t.Fatalf("failed to submit unhelpful feedback: %v", err)
	}

	// Verify article counters
	art, err := svc.GetArticleByID(ctx, article.ID)
	if err != nil {
		t.Fatalf("failed to get article: %v", err)
	}
	if art.HelpfulCount != 1 || art.NotHelpfulCount != 1 {
		t.Fatalf("expected 1 helpful and 1 unhelpful vote, got %d and %d", art.HelpfulCount, art.NotHelpfulCount)
	}

	// List feedback
	list, total, err := svc.ListFeedback(ctx, model.KnowledgeFeedbackFilter{
		WorkspaceID: "ws_1",
		ArticleID:   article.ID,
	})
	if err != nil {
		t.Fatalf("failed to list feedback: %v", err)
	}
	if total != 2 || len(list) != 2 {
		t.Fatalf("expected 2 feedback items, got %d (total %d)", len(list), total)
	}
}

func TestKnowledge_Reviews(t *testing.T) {
	svc, _, dbConn := setupTestKnowledgeSvc(t)
	defer dbConn.Close()

	ctx := context.Background()

	article, err := svc.CreateArticle(ctx, "ws_1", model.CreateKnowledgeArticleRequest{
		ArticleCode: "KB-REV-SCHEDULE-01",
		Title:       "Postgres Major Version Upgrade Runbook",
		Content:     "SOP for zero-downtime logical replication upgrades",
		Category:    model.KnowledgeCategorySOP,
	})
	if err != nil {
		t.Fatalf("failed to create article: %v", err)
	}

	// 1. Validation failure: missing reviewer_id
	_, err = svc.ScheduleReview(ctx, "ws_1", model.ScheduleKnowledgeReviewRequest{
		ArticleID: article.ID,
	})
	if err == nil {
		t.Fatal("expected error for missing reviewer_id, got nil")
	}

	// 2. Schedule review
	dueDate := time.Now().UTC().AddDate(0, 3, 0) // 3 months
	review, err := svc.ScheduleReview(ctx, "ws_1", model.ScheduleKnowledgeReviewRequest{
		ArticleID:     article.ID,
		ReviewerID:    "user_1",
		NextReviewDue: &dueDate,
	})
	if err != nil {
		t.Fatalf("failed to schedule review: %v", err)
	}
	if review.Status != model.KnowledgeReviewStatusScheduled {
		t.Fatalf("expected scheduled status, got %s", review.Status)
	}

	// 3. Complete review
	completed, err := svc.CompleteReview(ctx, review.ID, model.CompleteKnowledgeReviewRequest{
		Findings: "Verified against Postgres 16; commands and flags remain accurate",
	})
	if err != nil {
		t.Fatalf("failed to complete review: %v", err)
	}
	if completed.Status != model.KnowledgeReviewStatusCompleted || completed.ReviewedAt == nil {
		t.Fatalf("unexpected completed review: %+v", completed)
	}

	// List reviews
	list, total, err := svc.ListReviews(ctx, model.KnowledgeReviewFilter{
		WorkspaceID: "ws_1",
		ArticleID:   article.ID,
	})
	if err != nil {
		t.Fatalf("failed to list reviews: %v", err)
	}
	if total != 1 || len(list) != 1 {
		t.Fatalf("expected 1 review, got %d (total %d)", len(list), total)
	}
}

func TestKnowledge_Deflections(t *testing.T) {
	svc, _, dbConn := setupTestKnowledgeSvc(t)
	defer dbConn.Close()

	ctx := context.Background()

	article, _ := svc.CreateArticle(ctx, "ws_1", model.CreateKnowledgeArticleRequest{
		ArticleCode: "KB-DEFLECT-01",
		Title:       "Printer IP Configuration",
		Content:     "Follow these instructions to connect to office floor printer",
		Category:    model.KnowledgeCategoryHowTo,
	})

	// 1. Validation failure: missing query
	_, err := svc.LogDeflection(ctx, "ws_1", model.LogKnowledgeDeflectionRequest{
		ArticleID: article.ID,
	})
	if err == nil {
		t.Fatal("expected error for missing query, got nil")
	}

	// 2. Log successful deflection
	d1, err := svc.LogDeflection(ctx, "ws_1", model.LogKnowledgeDeflectionRequest{
		Query:     "printer setup 4th floor",
		ArticleID: article.ID,
		UserID:    "user_2",
		Deflected: true,
		Context:   "service_portal_search",
	})
	if err != nil {
		t.Fatalf("failed to log deflection: %v", err)
	}
	if !d1.Deflected {
		t.Fatal("expected deflected = true")
	}

	// 3. Log unsuccessful deflection
	_, _ = svc.LogDeflection(ctx, "ws_1", model.LogKnowledgeDeflectionRequest{
		Query:     "printer out of paper error",
		ArticleID: article.ID,
		UserID:    "user_2",
		Deflected: false,
		Context:   "incident_form_prompt",
	})

	// List deflections
	list, total, err := svc.ListDeflections(ctx, model.KnowledgeDeflectionFilter{
		WorkspaceID: "ws_1",
	})
	if err != nil {
		t.Fatalf("failed to list deflections: %v", err)
	}
	if total != 2 || len(list) != 2 {
		t.Fatalf("expected 2 deflections, got %d (total %d)", len(list), total)
	}
}

func TestKnowledge_Analytics(t *testing.T) {
	svc, _, dbConn := setupTestKnowledgeSvc(t)
	defer dbConn.Close()

	ctx := context.Background()

	// Create 2 articles: 1 published, 1 draft
	a1, _ := svc.CreateArticle(ctx, "ws_1", model.CreateKnowledgeArticleRequest{
		ArticleCode: "KB-ANALYTICS-01",
		Title:       "Database Tuning Guide",
		Content:     "Detailed tuning guide",
		Category:    model.KnowledgeCategoryReference,
	})
	_, _ = svc.PublishArticle(ctx, a1.ID)

	// Simulate views and feedback on a1
	_, _ = svc.GetArticleByID(ctx, a1.ID) // view 1
	_, _ = svc.GetArticleByID(ctx, a1.ID) // view 2
	_, _ = svc.SubmitFeedback(ctx, "ws_1", model.SubmitKnowledgeFeedbackRequest{
		ArticleID: a1.ID,
		UserID:    "user_1",
		IsHelpful: true,
		Rating:    5,
	})

	_, _ = svc.CreateArticle(ctx, "ws_1", model.CreateKnowledgeArticleRequest{
		ArticleCode: "KB-ANALYTICS-02",
		Title:       "Kafka Consumer Lag Resolution",
		Content:     "Troubleshooting consumer lag",
		Category:    model.KnowledgeCategoryTroubleshooting,
	})

	// Log 2 deflections: 1 successful
	_, _ = svc.LogDeflection(ctx, "ws_1", model.LogKnowledgeDeflectionRequest{
		Query:     "kafka lag",
		ArticleID: a1.ID,
		Deflected: true,
	})
	_, _ = svc.LogDeflection(ctx, "ws_1", model.LogKnowledgeDeflectionRequest{
		Query:     "postgres tuning",
		ArticleID: a1.ID,
		Deflected: false,
	})

	analytics, err := svc.GetKnowledgeAnalytics(ctx, "ws_1")
	if err != nil {
		t.Fatalf("failed to get analytics: %v", err)
	}

	if analytics.TotalArticles != 2 || analytics.PublishedArticles != 1 || analytics.DraftArticles != 1 {
		t.Fatalf("unexpected article counts: %+v", analytics)
	}
	if analytics.TotalViews < 2 {
		t.Fatalf("expected total views >= 2, got %d", analytics.TotalViews)
	}
	if analytics.TotalHelpfulVotes != 1 || analytics.AverageHelpfulnessRatio != 100.0 {
		t.Fatalf("unexpected helpfulness votes: %+v", analytics)
	}
	if analytics.TotalDeflectionAttempts != 2 || analytics.SuccessfulDeflections != 1 || analytics.DeflectionRate != 50.0 {
		t.Fatalf("unexpected deflection metrics: %+v", analytics)
	}
	if analytics.KnowledgeHealthIndex <= 0.0 || analytics.KnowledgeHealthIndex > 100.0 {
		t.Fatalf("unexpected health index: %f", analytics.KnowledgeHealthIndex)
	}
	if analytics.CategoryDistribution[model.KnowledgeCategoryReference] != 1 || analytics.CategoryDistribution[model.KnowledgeCategoryTroubleshooting] != 1 {
		t.Fatalf("unexpected category distribution: %+v", analytics.CategoryDistribution)
	}
}

func TestKnowledge_Concurrency(t *testing.T) {
	svc, _, dbConn := setupTestKnowledgeSvc(t)
	defer dbConn.Close()

	ctx := context.Background()

	const concurrency = 10
	var wg sync.WaitGroup
	errCh := make(chan error, concurrency*3)

	baseArticle, err := svc.CreateArticle(ctx, "ws_1", model.CreateKnowledgeArticleRequest{
		ArticleCode: "KB-CONC-BASE",
		Title:       "Concurrent Knowledge Base Benchmark",
		Content:     "Base article content for concurrent access",
		Category:    model.KnowledgeCategoryFAQ,
	})
	if err != nil {
		t.Fatalf("failed to create base article: %v", err)
	}

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()

			// 1. Concurrent create article
			_, aErr := svc.CreateArticle(ctx, "ws_1", model.CreateKnowledgeArticleRequest{
				ArticleCode: fmt.Sprintf("KB-CONC-%02d", idx),
				Title:       fmt.Sprintf("Concurrent Guide %d", idx),
				Content:     "Sample guide content",
				Category:    model.KnowledgeCategoryHowTo,
			})
			if aErr != nil {
				errCh <- aErr
				return
			}

			// 2. Concurrent read / view increment
			_, rErr := svc.GetArticleByID(ctx, baseArticle.ID)
			if rErr != nil {
				errCh <- rErr
				return
			}

			// 3. Concurrent feedback
			_, fErr := svc.SubmitFeedback(ctx, "ws_1", model.SubmitKnowledgeFeedbackRequest{
				ArticleID: baseArticle.ID,
				UserID:    fmt.Sprintf("user_%d", idx),
				IsHelpful: idx%2 == 0,
				Rating:    4,
			})
			if fErr != nil {
				errCh <- fErr
				return
			}
		}(i)
	}

	wg.Wait()
	close(errCh)

	for cErr := range errCh {
		if cErr != nil {
			t.Fatalf("concurrency error encountered: %v", cErr)
		}
	}

	analytics, err := svc.GetKnowledgeAnalytics(ctx, "ws_1")
	if err != nil {
		t.Fatalf("failed to get analytics after concurrency: %v", err)
	}
	if analytics.TotalArticles < concurrency+1 {
		t.Fatalf("expected at least %d articles, got %d", concurrency+1, analytics.TotalArticles)
	}
}
