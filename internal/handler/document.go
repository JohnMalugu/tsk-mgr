package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

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

