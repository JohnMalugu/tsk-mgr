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

// DocumentHandler exposes HTTP endpoints for Knowledge Base Documents and Wiki pages
type DocumentHandler struct {
	svc service.DocumentService
}

var GlobalDocumentHandler *DocumentHandler

// InitDocumentHandler initializes the singleton global document handler
func InitDocumentHandler(svc service.DocumentService) {
	GlobalDocumentHandler = &DocumentHandler{svc: svc}
}

// NewDocumentHandler creates a new DocumentHandler instance
func NewDocumentHandler(svc service.DocumentService) *DocumentHandler {
	return &DocumentHandler{svc: svc}
}

func getDocUser(r *http.Request) string {
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

func writeDocJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func writeDocError(w http.ResponseWriter, status int, msg string) {
	writeDocJSON(w, status, map[string]string{"error": msg})
}

// HandleDocuments handles GET and POST for /v2/docs
func (h *DocumentHandler) HandleDocuments(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		wsStr := r.URL.Query().Get("workspace_id")
		if wsStr == "" {
			writeDocError(w, http.StatusBadRequest, "workspace_id query parameter is required")
			return
		}
		wsID, err := strconv.ParseInt(wsStr, 10, 64)
		if err != nil || wsID <= 0 {
			writeDocError(w, http.StatusBadRequest, "invalid workspace_id")
			return
		}

		filter := model.DocumentFilter{
			WorkspaceID: wsID,
			AuthorID:    r.URL.Query().Get("author_id"),
			DocType:     r.URL.Query().Get("doc_type"),
			Status:      r.URL.Query().Get("status"),
			Search:      r.URL.Query().Get("search"),
		}

		if projStr := r.URL.Query().Get("project_id"); projStr != "" {
			if pid, err := strconv.Atoi(projStr); err == nil {
				filter.ProjectID = &pid
			}
		}

		if parentStr := r.URL.Query().Get("parent_id"); parentStr != "" {
			if parentID, err := strconv.ParseInt(parentStr, 10, 64); err == nil {
				filter.ParentDocumentID = &parentID
			}
		}

		if pinnedStr := r.URL.Query().Get("pinned"); pinnedStr != "" {
			pinned := pinnedStr == "true" || pinnedStr == "1"
			filter.IsPinned = &pinned
		}

		if limStr := r.URL.Query().Get("limit"); limStr != "" {
			if lim, err := strconv.Atoi(limStr); err == nil && lim > 0 {
				filter.Limit = lim
			}
		}
		if offStr := r.URL.Query().Get("offset"); offStr != "" {
			if off, err := strconv.Atoi(offStr); err == nil && off >= 0 {
				filter.Offset = off
			}
		}

		docs, err := h.svc.ListDocuments(filter)
		if err != nil {
			if errors.Is(err, service.ErrWorkspaceRequired) {
				writeDocError(w, http.StatusBadRequest, err.Error())
				return
			}
			writeDocError(w, http.StatusInternalServerError, "Failed to list documents: "+err.Error())
			return
		}
		writeDocJSON(w, http.StatusOK, docs)

	case http.MethodPost:
		var req model.CreateDocumentRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeDocError(w, http.StatusBadRequest, "Invalid request payload: "+err.Error())
			return
		}

		authorID := getDocUser(r)
		doc, err := h.svc.CreateDocument(authorID, req)
		if err != nil {
			switch {
			case errors.Is(err, service.ErrWorkspaceRequired),
				errors.Is(err, service.ErrDocumentTitleRequired),
				errors.Is(err, service.ErrInvalidDocType),
				errors.Is(err, service.ErrInvalidPublishStatus):
				writeDocError(w, http.StatusBadRequest, err.Error())
			case errors.Is(err, service.ErrProjectNotFound),
				errors.Is(err, service.ErrParentDocumentNotFound):
				writeDocError(w, http.StatusNotFound, err.Error())
			default:
				writeDocError(w, http.StatusInternalServerError, "Failed to create document: "+err.Error())
			}
			return
		}
		writeDocJSON(w, http.StatusCreated, doc)

	default:
		writeDocError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

// HandleDocumentDetail handles GET, PUT, and DELETE for /v2/docs/{id}
func (h *DocumentHandler) HandleDocumentDetail(w http.ResponseWriter, r *http.Request, id int64) {
	switch r.Method {
	case http.MethodGet:
		recordView := r.URL.Query().Get("record_view") != "false"
		doc, err := h.svc.GetDocument(id, recordView)
		if err != nil {
			if errors.Is(err, service.ErrDocumentNotFound) {
				writeDocError(w, http.StatusNotFound, err.Error())
				return
			}
			writeDocError(w, http.StatusInternalServerError, "Failed to fetch document: "+err.Error())
			return
		}
		writeDocJSON(w, http.StatusOK, doc)

	case http.MethodPut, http.MethodPatch:
		var req model.UpdateDocumentRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeDocError(w, http.StatusBadRequest, "Invalid request payload: "+err.Error())
			return
		}

		authorID := getDocUser(r)
		doc, err := h.svc.UpdateDocument(id, authorID, req)
		if err != nil {
			switch {
			case errors.Is(err, service.ErrDocumentTitleRequired),
				errors.Is(err, service.ErrInvalidDocType),
				errors.Is(err, service.ErrInvalidPublishStatus),
				errors.Is(err, service.ErrCircularDocumentHierarchy):
				writeDocError(w, http.StatusBadRequest, err.Error())
			case errors.Is(err, service.ErrDocumentNotFound),
				errors.Is(err, service.ErrParentDocumentNotFound),
				errors.Is(err, service.ErrProjectNotFound):
				writeDocError(w, http.StatusNotFound, err.Error())
			default:
				writeDocError(w, http.StatusInternalServerError, "Failed to update document: "+err.Error())
			}
			return
		}
		writeDocJSON(w, http.StatusOK, doc)

	case http.MethodDelete:
		err := h.svc.DeleteDocument(id)
		if err != nil {
			if errors.Is(err, service.ErrDocumentNotFound) {
				writeDocError(w, http.StatusNotFound, err.Error())
				return
			}
			writeDocError(w, http.StatusInternalServerError, "Failed to delete document: "+err.Error())
			return
		}
		writeDocJSON(w, http.StatusOK, map[string]string{"message": "Document deleted successfully"})

	default:
		writeDocError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

// HandleDocumentBySlug handles GET /v2/docs/slug/{slug}
func (h *DocumentHandler) HandleDocumentBySlug(w http.ResponseWriter, r *http.Request, slug string) {
	if r.Method != http.MethodGet {
		writeDocError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	wsStr := r.URL.Query().Get("workspace_id")
	if wsStr == "" {
		writeDocError(w, http.StatusBadRequest, "workspace_id query parameter is required")
		return
	}
	wsID, err := strconv.ParseInt(wsStr, 10, 64)
	if err != nil || wsID <= 0 {
		writeDocError(w, http.StatusBadRequest, "invalid workspace_id")
		return
	}

	recordView := r.URL.Query().Get("record_view") != "false"
	doc, err := h.svc.GetDocumentBySlug(wsID, slug, recordView)
	if err != nil {
		if errors.Is(err, service.ErrDocumentNotFound) {
			writeDocError(w, http.StatusNotFound, err.Error())
			return
		}
		writeDocError(w, http.StatusInternalServerError, "Failed to fetch document by slug: "+err.Error())
		return
	}
	writeDocJSON(w, http.StatusOK, doc)
}

// HandleDocumentTree handles GET /v2/docs/tree?workspace_id={id}&project_id={id}
func (h *DocumentHandler) HandleDocumentTree(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeDocError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	wsStr := r.URL.Query().Get("workspace_id")
	if wsStr == "" {
		writeDocError(w, http.StatusBadRequest, "workspace_id query parameter is required")
		return
	}
	wsID, err := strconv.ParseInt(wsStr, 10, 64)
	if err != nil || wsID <= 0 {
		writeDocError(w, http.StatusBadRequest, "invalid workspace_id")
		return
	}

	var projectID *int
	if projStr := r.URL.Query().Get("project_id"); projStr != "" {
		if pid, err := strconv.Atoi(projStr); err == nil {
			projectID = &pid
		}
	}

	tree, err := h.svc.GetDocumentTree(wsID, projectID)
	if err != nil {
		if errors.Is(err, service.ErrWorkspaceRequired) {
			writeDocError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeDocError(w, http.StatusInternalServerError, "Failed to load document tree: "+err.Error())
		return
	}
	writeDocJSON(w, http.StatusOK, tree)
}

// HandlePublish handles POST /v2/docs/{id}/publish
func (h *DocumentHandler) HandlePublish(w http.ResponseWriter, r *http.Request, id int64) {
	if r.Method != http.MethodPost {
		writeDocError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req struct {
		Status string `json:"status"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)
	if req.Status == "" {
		req.Status = model.DocStatusPublished
	}

	doc, err := h.svc.PublishDocument(id, req.Status)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidPublishStatus):
			writeDocError(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, service.ErrDocumentNotFound):
			writeDocError(w, http.StatusNotFound, err.Error())
		default:
			writeDocError(w, http.StatusInternalServerError, "Failed to publish document: "+err.Error())
		}
		return
	}
	writeDocJSON(w, http.StatusOK, doc)
}

// HandlePin handles POST or PATCH /v2/docs/{id}/pin
func (h *DocumentHandler) HandlePin(w http.ResponseWriter, r *http.Request, id int64) {
	if r.Method != http.MethodPost && r.Method != http.MethodPatch {
		writeDocError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	doc, err := h.svc.TogglePin(id)
	if err != nil {
		if errors.Is(err, service.ErrDocumentNotFound) {
			writeDocError(w, http.StatusNotFound, err.Error())
			return
		}
		writeDocError(w, http.StatusInternalServerError, "Failed to toggle pin on document: "+err.Error())
		return
	}
	writeDocJSON(w, http.StatusOK, doc)
}

// HandleRevisions handles GET /v2/docs/{id}/revisions
func (h *DocumentHandler) HandleRevisions(w http.ResponseWriter, r *http.Request, id int64) {
	if r.Method != http.MethodGet {
		writeDocError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	revs, err := h.svc.ListRevisions(id)
	if err != nil {
		if errors.Is(err, service.ErrDocumentNotFound) {
			writeDocError(w, http.StatusNotFound, err.Error())
			return
		}
		writeDocError(w, http.StatusInternalServerError, "Failed to list revisions: "+err.Error())
		return
	}
	writeDocJSON(w, http.StatusOK, revs)
}

// HandleRevisionDetail handles GET /v2/docs/{id}/revisions/{version}
func (h *DocumentHandler) HandleRevisionDetail(w http.ResponseWriter, r *http.Request, id int64, version int) {
	if r.Method != http.MethodGet {
		writeDocError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	rev, err := h.svc.GetRevision(id, version)
	if err != nil {
		if errors.Is(err, service.ErrDocumentNotFound) || errors.Is(err, service.ErrRevisionNotFound) {
			writeDocError(w, http.StatusNotFound, err.Error())
			return
		}
		writeDocError(w, http.StatusInternalServerError, "Failed to get revision: "+err.Error())
		return
	}
	writeDocJSON(w, http.StatusOK, rev)
}

// HandleRollback handles POST /v2/docs/{id}/rollback
func (h *DocumentHandler) HandleRollback(w http.ResponseWriter, r *http.Request, id int64) {
	if r.Method != http.MethodPost {
		writeDocError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req model.RollbackRevisionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeDocError(w, http.StatusBadRequest, "Invalid request payload: "+err.Error())
		return
	}

	authorID := getDocUser(r)
	doc, err := h.svc.RollbackDocument(id, authorID, req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrDocumentNotFound), errors.Is(err, service.ErrRevisionNotFound):
			writeDocError(w, http.StatusNotFound, err.Error())
		default:
			writeDocError(w, http.StatusInternalServerError, "Failed to rollback document: "+err.Error())
		}
		return
	}
	writeDocJSON(w, http.StatusOK, doc)
}

// HandleDocumentTasks handles GET and POST for /v2/docs/{id}/tasks
func (h *DocumentHandler) HandleDocumentTasks(w http.ResponseWriter, r *http.Request, docID int64) {
	switch r.Method {
	case http.MethodGet:
		tasks, err := h.svc.ListDocumentTasks(docID)
		if err != nil {
			if errors.Is(err, service.ErrDocumentNotFound) {
				writeDocError(w, http.StatusNotFound, err.Error())
				return
			}
			writeDocError(w, http.StatusInternalServerError, "Failed to list document tasks: "+err.Error())
			return
		}
		writeDocJSON(w, http.StatusOK, tasks)

	case http.MethodPost:
		var req model.LinkDocumentTaskRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeDocError(w, http.StatusBadRequest, "Invalid request payload: "+err.Error())
			return
		}

		createdBy := getDocUser(r)
		link, err := h.svc.LinkTask(docID, createdBy, req)
		if err != nil {
			switch {
			case errors.Is(err, service.ErrInvalidTaskRelationship):
				writeDocError(w, http.StatusBadRequest, err.Error())
			case errors.Is(err, service.ErrDocumentNotFound), errors.Is(err, service.ErrTaskNotFound):
				writeDocError(w, http.StatusNotFound, err.Error())
			default:
				writeDocError(w, http.StatusInternalServerError, "Failed to link task: "+err.Error())
			}
			return
		}
		writeDocJSON(w, http.StatusCreated, link)

	default:
		writeDocError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

// HandleDocumentTaskDetail handles DELETE /v2/docs/{id}/tasks/{taskId}
func (h *DocumentHandler) HandleDocumentTaskDetail(w http.ResponseWriter, r *http.Request, docID int64, taskID int) {
	if r.Method != http.MethodDelete {
		writeDocError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	err := h.svc.UnlinkTask(docID, taskID)
	if err != nil {
		if errors.Is(err, service.ErrDocumentNotFound) {
			writeDocError(w, http.StatusNotFound, err.Error())
			return
		}
		writeDocError(w, http.StatusInternalServerError, "Failed to unlink task: "+err.Error())
		return
	}
	writeDocJSON(w, http.StatusOK, map[string]string{"message": "Task unlinked successfully"})
}

// HandleDocumentComments handles GET and POST for /v2/docs/{id}/comments
func (h *DocumentHandler) HandleDocumentComments(w http.ResponseWriter, r *http.Request, docID int64) {
	switch r.Method {
	case http.MethodGet:
		comments, err := h.svc.ListComments(docID)
		if err != nil {
			if errors.Is(err, service.ErrDocumentNotFound) {
				writeDocError(w, http.StatusNotFound, err.Error())
				return
			}
			writeDocError(w, http.StatusInternalServerError, "Failed to list document comments: "+err.Error())
			return
		}
		writeDocJSON(w, http.StatusOK, comments)

	case http.MethodPost:
		var req model.CreateDocumentCommentRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeDocError(w, http.StatusBadRequest, "Invalid request payload: "+err.Error())
			return
		}

		authorID := getDocUser(r)
		comment, err := h.svc.CreateComment(docID, authorID, req)
		if err != nil {
			switch {
			case errors.Is(err, service.ErrCommentContentRequired):
				writeDocError(w, http.StatusBadRequest, err.Error())
			case errors.Is(err, service.ErrDocumentNotFound):
				writeDocError(w, http.StatusNotFound, err.Error())
			default:
				writeDocError(w, http.StatusInternalServerError, "Failed to create comment: "+err.Error())
			}
			return
		}
		writeDocJSON(w, http.StatusCreated, comment)

	default:
		writeDocError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

// HandleResolveComment handles POST or PATCH /v2/docs/comments/{commentId}/resolve
func (h *DocumentHandler) HandleResolveComment(w http.ResponseWriter, r *http.Request, commentID int64) {
	if r.Method != http.MethodPost && r.Method != http.MethodPatch {
		writeDocError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	resolvedBy := getDocUser(r)
	comment, err := h.svc.ResolveComment(commentID, resolvedBy)
	if err != nil {
		if errors.Is(err, service.ErrCommentNotFound) {
			writeDocError(w, http.StatusNotFound, err.Error())
			return
		}
		writeDocError(w, http.StatusInternalServerError, "Failed to resolve comment: "+err.Error())
		return
	}
	writeDocJSON(w, http.StatusOK, comment)
}

// HandleMetrics handles GET /v2/docs/metrics?workspace_id={id}
func (h *DocumentHandler) HandleMetrics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeDocError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	wsStr := r.URL.Query().Get("workspace_id")
	if wsStr == "" {
		writeDocError(w, http.StatusBadRequest, "workspace_id query parameter is required")
		return
	}
	wsID, err := strconv.ParseInt(wsStr, 10, 64)
	if err != nil || wsID <= 0 {
		writeDocError(w, http.StatusBadRequest, "invalid workspace_id")
		return
	}

	metrics, err := h.svc.GetDocumentMetrics(wsID)
	if err != nil {
		if errors.Is(err, service.ErrWorkspaceRequired) {
			writeDocError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeDocError(w, http.StatusInternalServerError, "Failed to get document metrics: "+err.Error())
		return
	}
	writeDocJSON(w, http.StatusOK, metrics)
}

// ServeHTTP dispatches requests under /v2/docs
func (h *DocumentHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/v2/docs")
	path = strings.TrimPrefix(path, "/")

	if path == "" {
		h.HandleDocuments(w, r)
		return
	}

	if path == "tree" || path == "tree/" {
		h.HandleDocumentTree(w, r)
		return
	}

	if path == "metrics" || path == "metrics/" {
		h.HandleMetrics(w, r)
		return
	}

	if strings.HasPrefix(path, "slug/") {
		slug := strings.TrimPrefix(path, "slug/")
		h.HandleDocumentBySlug(w, r, slug)
		return
	}

	if strings.HasPrefix(path, "comments/") {
		sub := strings.TrimPrefix(path, "comments/")
		parts := strings.Split(sub, "/")
		if len(parts) == 2 && parts[1] == "resolve" {
			commentID, err := strconv.ParseInt(parts[0], 10, 64)
			if err != nil || commentID <= 0 {
				writeDocError(w, http.StatusBadRequest, "invalid comment id")
				return
			}
			h.HandleResolveComment(w, r, commentID)
			return
		}
		writeDocError(w, http.StatusNotFound, "not found")
		return
	}

	parts := strings.Split(path, "/")
	id, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil || id <= 0 {
		writeDocError(w, http.StatusBadRequest, "invalid document id")
		return
	}

	if len(parts) == 1 || (len(parts) == 2 && parts[1] == "") {
		h.HandleDocumentDetail(w, r, id)
		return
	}

	if len(parts) == 2 {
		switch parts[1] {
		case "publish":
			h.HandlePublish(w, r, id)
		case "pin":
			h.HandlePin(w, r, id)
		case "revisions":
			h.HandleRevisions(w, r, id)
		case "rollback":
			h.HandleRollback(w, r, id)
		case "tasks":
			h.HandleDocumentTasks(w, r, id)
		case "comments":
			h.HandleDocumentComments(w, r, id)
		default:
			writeDocError(w, http.StatusNotFound, "not found")
		}
		return
	}

	if len(parts) == 3 {
		switch parts[1] {
		case "revisions":
			version, err := strconv.Atoi(parts[2])
			if err != nil || version <= 0 {
				writeDocError(w, http.StatusBadRequest, "invalid revision version")
				return
			}
			h.HandleRevisionDetail(w, r, id, version)
		case "tasks":
			taskID, err := strconv.Atoi(parts[2])
			if err != nil || taskID <= 0 {
				writeDocError(w, http.StatusBadRequest, "invalid task id")
				return
			}
			h.HandleDocumentTaskDetail(w, r, id, taskID)
		default:
			writeDocError(w, http.StatusNotFound, "not found")
		}
		return
	}

	writeDocError(w, http.StatusNotFound, "not found")
}


