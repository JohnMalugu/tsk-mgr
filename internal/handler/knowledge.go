package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/JohnMalugu/tsk-mgr-api/internal/model"
	"github.com/JohnMalugu/tsk-mgr-api/internal/service"
)

var (
	_ = strconv.Atoi
	_ = strings.TrimPrefix
)

// KnowledgeHandler exposes HTTP endpoints for ITIL Knowledge Management (KEDB)
type KnowledgeHandler struct {
	svc service.KnowledgeService
}

// GlobalKnowledgeHandler is the singleton handler instance
var GlobalKnowledgeHandler *KnowledgeHandler

// InitKnowledgeHandler initializes the global Knowledge handler singleton
func InitKnowledgeHandler(svc service.KnowledgeService) {
	GlobalKnowledgeHandler = &KnowledgeHandler{svc: svc}
}

// NewKnowledgeHandler creates a new KnowledgeHandler instance
func NewKnowledgeHandler(svc service.KnowledgeService) *KnowledgeHandler {
	return &KnowledgeHandler{svc: svc}
}

func getKnowledgeUser(r *http.Request) string {
	if uid, ok := r.Context().Value("userID").(string); ok && uid != "" {
		return uid
	}
	if uid := r.Header.Get("X-User-ID"); uid != "" {
		return uid
	}
	if uid := r.URL.Query().Get("user_id"); uid != "" {
		return uid
	}
	return "user_1"
}

func getKnowledgeWorkspaceID(r *http.Request) string {
	if ws := r.URL.Query().Get("workspace_id"); ws != "" {
		return ws
	}
	if ws := r.Header.Get("X-Workspace-ID"); ws != "" {
		return ws
	}
	return "ws_1"
}

func writeKnowledgeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func writeKnowledgeError(w http.ResponseWriter, status int, msg string) {
	writeKnowledgeJSON(w, status, map[string]string{"error": msg})
}

func handleKnowledgeError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrKnowledgeArticleNotFound):
		writeKnowledgeError(w, http.StatusNotFound, "knowledge article not found")
	case errors.Is(err, service.ErrKnowledgeArticleCodeConflict):
		writeKnowledgeError(w, http.StatusConflict, "article with this code already exists in workspace")
	case errors.Is(err, service.ErrKnowledgeRevisionNotFound):
		writeKnowledgeError(w, http.StatusNotFound, "knowledge revision not found")
	case errors.Is(err, service.ErrKnowledgeFeedbackNotFound):
		writeKnowledgeError(w, http.StatusNotFound, "knowledge feedback not found")
	case errors.Is(err, service.ErrKnowledgeReviewNotFound):
		writeKnowledgeError(w, http.StatusNotFound, "knowledge review not found")
	default:
		writeKnowledgeError(w, http.StatusBadRequest, err.Error())
	}
}

// Articles HTTP Endpoints

// ListArticles handles GET /v2/knowledge/articles
func (h *KnowledgeHandler) ListArticles(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	limit, _ := strconv.Atoi(q.Get("limit"))
	offset, _ := strconv.Atoi(q.Get("offset"))

	filter := model.KnowledgeArticleFilter{
		WorkspaceID: getKnowledgeWorkspaceID(r),
		Category:    q.Get("category"),
		Status:      q.Get("status"),
		Visibility:  q.Get("visibility"),
		AuthorID:    q.Get("author_id"),
		CIID:        q.Get("ci_id"),
		Search:      q.Get("search"),
		Limit:       limit,
		Offset:      offset,
	}

	articles, total, err := h.svc.ListArticles(r.Context(), filter)
	if err != nil {
		handleKnowledgeError(w, err)
		return
	}

	writeKnowledgeJSON(w, http.StatusOK, map[string]interface{}{
		"articles": articles,
		"total":    total,
		"limit":    filter.Limit,
		"offset":   filter.Offset,
	})
}

// CreateArticle handles POST /v2/knowledge/articles
func (h *KnowledgeHandler) CreateArticle(w http.ResponseWriter, r *http.Request) {
	var req model.CreateKnowledgeArticleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeKnowledgeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	ws := getKnowledgeWorkspaceID(r)
	if req.WorkspaceID == "" {
		req.WorkspaceID = ws
	}
	if req.AuthorID == "" {
		req.AuthorID = getKnowledgeUser(r)
	}

	article, err := h.svc.CreateArticle(r.Context(), ws, req)
	if err != nil {
		handleKnowledgeError(w, err)
		return
	}

	writeKnowledgeJSON(w, http.StatusCreated, article)
}

// GetArticleByID handles GET /v2/knowledge/articles/{id}
func (h *KnowledgeHandler) GetArticleByID(w http.ResponseWriter, r *http.Request, id string) {
	article, err := h.svc.GetArticleByID(r.Context(), id)
	if err != nil {
		handleKnowledgeError(w, err)
		return
	}
	writeKnowledgeJSON(w, http.StatusOK, article)
}

// UpdateArticle handles PUT /v2/knowledge/articles/{id}
func (h *KnowledgeHandler) UpdateArticle(w http.ResponseWriter, r *http.Request, id string) {
	var req model.UpdateKnowledgeArticleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeKnowledgeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	article, err := h.svc.UpdateArticle(r.Context(), id, req)
	if err != nil {
		handleKnowledgeError(w, err)
		return
	}
	writeKnowledgeJSON(w, http.StatusOK, article)
}

// PublishArticle handles POST /v2/knowledge/articles/{id}/publish
func (h *KnowledgeHandler) PublishArticle(w http.ResponseWriter, r *http.Request, id string) {
	article, err := h.svc.PublishArticle(r.Context(), id)
	if err != nil {
		handleKnowledgeError(w, err)
		return
	}
	writeKnowledgeJSON(w, http.StatusOK, article)
}

// ArchiveArticle handles POST /v2/knowledge/articles/{id}/archive
func (h *KnowledgeHandler) ArchiveArticle(w http.ResponseWriter, r *http.Request, id string) {
	article, err := h.svc.ArchiveArticle(r.Context(), id)
	if err != nil {
		handleKnowledgeError(w, err)
		return
	}
	writeKnowledgeJSON(w, http.StatusOK, article)
}

// DeleteArticle handles DELETE /v2/knowledge/articles/{id}
func (h *KnowledgeHandler) DeleteArticle(w http.ResponseWriter, r *http.Request, id string) {
	if err := h.svc.DeleteArticle(r.Context(), id); err != nil {
		handleKnowledgeError(w, err)
		return
	}
	writeKnowledgeJSON(w, http.StatusOK, map[string]string{"message": "knowledge article deleted successfully"})
}

// Revisions HTTP Endpoints

// ListRevisions handles GET /v2/knowledge/revisions
func (h *KnowledgeHandler) ListRevisions(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	limit, _ := strconv.Atoi(q.Get("limit"))
	offset, _ := strconv.Atoi(q.Get("offset"))

	filter := model.KnowledgeRevisionFilter{
		WorkspaceID: getKnowledgeWorkspaceID(r),
		ArticleID:   q.Get("article_id"),
		Limit:       limit,
		Offset:      offset,
	}

	revisions, total, err := h.svc.ListRevisions(r.Context(), filter)
	if err != nil {
		handleKnowledgeError(w, err)
		return
	}

	writeKnowledgeJSON(w, http.StatusOK, map[string]interface{}{
		"revisions": revisions,
		"total":     total,
		"limit":     filter.Limit,
		"offset":    filter.Offset,
	})
}

// GetRevisionByID handles GET /v2/knowledge/revisions/{id}
func (h *KnowledgeHandler) GetRevisionByID(w http.ResponseWriter, r *http.Request, id string) {
	revision, err := h.svc.GetRevisionByID(r.Context(), id)
	if err != nil {
		handleKnowledgeError(w, err)
		return
	}
	writeKnowledgeJSON(w, http.StatusOK, revision)
}

// Feedback HTTP Endpoints

// ListFeedback handles GET /v2/knowledge/feedback
func (h *KnowledgeHandler) ListFeedback(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	limit, _ := strconv.Atoi(q.Get("limit"))
	offset, _ := strconv.Atoi(q.Get("offset"))

	filter := model.KnowledgeFeedbackFilter{
		WorkspaceID: getKnowledgeWorkspaceID(r),
		ArticleID:   q.Get("article_id"),
		Limit:       limit,
		Offset:      offset,
	}

	feedback, total, err := h.svc.ListFeedback(r.Context(), filter)
	if err != nil {
		handleKnowledgeError(w, err)
		return
	}

	writeKnowledgeJSON(w, http.StatusOK, map[string]interface{}{
		"feedback": feedback,
		"total":    total,
		"limit":    filter.Limit,
		"offset":   filter.Offset,
	})
}

// SubmitFeedback handles POST /v2/knowledge/feedback
func (h *KnowledgeHandler) SubmitFeedback(w http.ResponseWriter, r *http.Request) {
	var req model.SubmitKnowledgeFeedbackRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeKnowledgeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	ws := getKnowledgeWorkspaceID(r)
	if req.WorkspaceID == "" {
		req.WorkspaceID = ws
	}
	if req.UserID == "" {
		req.UserID = getKnowledgeUser(r)
	}

	feedback, err := h.svc.SubmitFeedback(r.Context(), ws, req)
	if err != nil {
		handleKnowledgeError(w, err)
		return
	}

	writeKnowledgeJSON(w, http.StatusCreated, feedback)
}

// Reviews HTTP Endpoints

// ListReviews handles GET /v2/knowledge/reviews
func (h *KnowledgeHandler) ListReviews(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	limit, _ := strconv.Atoi(q.Get("limit"))
	offset, _ := strconv.Atoi(q.Get("offset"))

	filter := model.KnowledgeReviewFilter{
		WorkspaceID: getKnowledgeWorkspaceID(r),
		ArticleID:   q.Get("article_id"),
		Status:      q.Get("status"),
		Limit:       limit,
		Offset:      offset,
	}

	reviews, total, err := h.svc.ListReviews(r.Context(), filter)
	if err != nil {
		handleKnowledgeError(w, err)
		return
	}

	writeKnowledgeJSON(w, http.StatusOK, map[string]interface{}{
		"reviews": reviews,
		"total":   total,
		"limit":   filter.Limit,
		"offset":  filter.Offset,
	})
}

// ScheduleReview handles POST /v2/knowledge/reviews
func (h *KnowledgeHandler) ScheduleReview(w http.ResponseWriter, r *http.Request) {
	var req model.ScheduleKnowledgeReviewRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeKnowledgeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	ws := getKnowledgeWorkspaceID(r)
	if req.WorkspaceID == "" {
		req.WorkspaceID = ws
	}
	if req.ReviewerID == "" {
		req.ReviewerID = getKnowledgeUser(r)
	}

	review, err := h.svc.ScheduleReview(r.Context(), ws, req)
	if err != nil {
		handleKnowledgeError(w, err)
		return
	}

	writeKnowledgeJSON(w, http.StatusCreated, review)
}

// GetReviewByID handles GET /v2/knowledge/reviews/{id}
func (h *KnowledgeHandler) GetReviewByID(w http.ResponseWriter, r *http.Request, id string) {
	review, err := h.svc.GetReviewByID(r.Context(), id)
	if err != nil {
		handleKnowledgeError(w, err)
		return
	}
	writeKnowledgeJSON(w, http.StatusOK, review)
}

// CompleteReview handles POST /v2/knowledge/reviews/{id}/complete
func (h *KnowledgeHandler) CompleteReview(w http.ResponseWriter, r *http.Request, id string) {
	var req model.CompleteKnowledgeReviewRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeKnowledgeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	review, err := h.svc.CompleteReview(r.Context(), id, req)
	if err != nil {
		handleKnowledgeError(w, err)
		return
	}
	writeKnowledgeJSON(w, http.StatusOK, review)
}

// Deflections and Lookup HTTP Endpoints

// ListDeflections handles GET /v2/knowledge/deflections
func (h *KnowledgeHandler) ListDeflections(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	limit, _ := strconv.Atoi(q.Get("limit"))
	offset, _ := strconv.Atoi(q.Get("offset"))

	filter := model.KnowledgeDeflectionFilter{
		WorkspaceID: getKnowledgeWorkspaceID(r),
		ArticleID:   q.Get("article_id"),
		Limit:       limit,
		Offset:      offset,
	}

	deflections, total, err := h.svc.ListDeflections(r.Context(), filter)
	if err != nil {
		handleKnowledgeError(w, err)
		return
	}

	writeKnowledgeJSON(w, http.StatusOK, map[string]interface{}{
		"deflections": deflections,
		"total":       total,
		"limit":       filter.Limit,
		"offset":      filter.Offset,
	})
}

// LogDeflection handles POST /v2/knowledge/deflections
func (h *KnowledgeHandler) LogDeflection(w http.ResponseWriter, r *http.Request) {
	var req model.LogKnowledgeDeflectionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeKnowledgeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	ws := getKnowledgeWorkspaceID(r)
	if req.WorkspaceID == "" {
		req.WorkspaceID = ws
	}
	if req.UserID == "" {
		req.UserID = getKnowledgeUser(r)
	}

	deflection, err := h.svc.LogDeflection(r.Context(), ws, req)
	if err != nil {
		handleKnowledgeError(w, err)
		return
	}

	writeKnowledgeJSON(w, http.StatusCreated, deflection)
}

// GetArticleByCode handles GET /v2/knowledge/code/{code}
func (h *KnowledgeHandler) GetArticleByCode(w http.ResponseWriter, r *http.Request, code string) {
	ws := getKnowledgeWorkspaceID(r)
	article, err := h.svc.GetArticleByCode(r.Context(), ws, code)
	if err != nil {
		handleKnowledgeError(w, err)
		return
	}
	writeKnowledgeJSON(w, http.StatusOK, article)
}

// Analytics and Router HTTP Endpoints

// GetKnowledgeAnalytics handles GET /v2/knowledge/analytics
func (h *KnowledgeHandler) GetKnowledgeAnalytics(w http.ResponseWriter, r *http.Request) {
	ws := getKnowledgeWorkspaceID(r)
	analytics, err := h.svc.GetKnowledgeAnalytics(r.Context(), ws)
	if err != nil {
		handleKnowledgeError(w, err)
		return
	}
	writeKnowledgeJSON(w, http.StatusOK, analytics)
}

// ServeHTTP routes HTTP requests under /v2/knowledge
func (h *KnowledgeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/v2/knowledge")
	path = strings.TrimPrefix(path, "/")

	if path == "" {
		writeKnowledgeJSON(w, http.StatusOK, map[string]string{
			"service": "ITIL Knowledge Management API",
			"version": "v2",
		})
		return
	}

	if path == "analytics" {
		if r.Method == http.MethodGet {
			h.GetKnowledgeAnalytics(w, r)
		} else {
			writeKnowledgeError(w, http.StatusMethodNotAllowed, "method not allowed")
		}
		return
	}

	parts := strings.Split(path, "/")
	resource := parts[0]

	switch resource {
	case "code":
		if len(parts) == 2 && r.Method == http.MethodGet {
			h.GetArticleByCode(w, r, parts[1])
			return
		}

	case "articles":
		if len(parts) == 1 {
			switch r.Method {
			case http.MethodGet:
				h.ListArticles(w, r)
			case http.MethodPost:
				h.CreateArticle(w, r)
			default:
				writeKnowledgeError(w, http.StatusMethodNotAllowed, "method not allowed")
			}
			return
		}
		articleID := parts[1]
		if len(parts) == 2 {
			switch r.Method {
			case http.MethodGet:
				h.GetArticleByID(w, r, articleID)
			case http.MethodPut:
				h.UpdateArticle(w, r, articleID)
			case http.MethodDelete:
				h.DeleteArticle(w, r, articleID)
			default:
				writeKnowledgeError(w, http.StatusMethodNotAllowed, "method not allowed")
			}
			return
		}
		if len(parts) == 3 {
			switch parts[2] {
			case "publish":
				if r.Method == http.MethodPost {
					h.PublishArticle(w, r, articleID)
				} else {
					writeKnowledgeError(w, http.StatusMethodNotAllowed, "method not allowed")
				}
				return
			case "archive":
				if r.Method == http.MethodPost {
					h.ArchiveArticle(w, r, articleID)
				} else {
					writeKnowledgeError(w, http.StatusMethodNotAllowed, "method not allowed")
				}
				return
			}
		}

	case "revisions":
		if len(parts) == 1 {
			if r.Method == http.MethodGet {
				h.ListRevisions(w, r)
			} else {
				writeKnowledgeError(w, http.StatusMethodNotAllowed, "method not allowed")
			}
			return
		}
		if len(parts) == 2 && r.Method == http.MethodGet {
			h.GetRevisionByID(w, r, parts[1])
			return
		}

	case "feedback":
		if len(parts) == 1 {
			switch r.Method {
			case http.MethodGet:
				h.ListFeedback(w, r)
			case http.MethodPost:
				h.SubmitFeedback(w, r)
			default:
				writeKnowledgeError(w, http.StatusMethodNotAllowed, "method not allowed")
			}
			return
		}

	case "reviews":
		if len(parts) == 1 {
			switch r.Method {
			case http.MethodGet:
				h.ListReviews(w, r)
			case http.MethodPost:
				h.ScheduleReview(w, r)
			default:
				writeKnowledgeError(w, http.StatusMethodNotAllowed, "method not allowed")
			}
			return
		}
		reviewID := parts[1]
		if len(parts) == 2 {
			if r.Method == http.MethodGet {
				h.GetReviewByID(w, r, reviewID)
			} else {
				writeKnowledgeError(w, http.StatusMethodNotAllowed, "method not allowed")
			}
			return
		}
		if len(parts) == 3 && parts[2] == "complete" {
			if r.Method == http.MethodPost {
				h.CompleteReview(w, r, reviewID)
			} else {
				writeKnowledgeError(w, http.StatusMethodNotAllowed, "method not allowed")
			}
			return
		}

	case "deflections":
		if len(parts) == 1 {
			switch r.Method {
			case http.MethodGet:
				h.ListDeflections(w, r)
			case http.MethodPost:
				h.LogDeflection(w, r)
			default:
				writeKnowledgeError(w, http.StatusMethodNotAllowed, "method not allowed")
			}
			return
		}
	}

	writeKnowledgeError(w, http.StatusNotFound, "route not found")
}
