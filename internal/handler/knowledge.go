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
