package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/JohnMalugu/tsk-mgr-api/internal/model"
	"github.com/JohnMalugu/tsk-mgr-api/internal/service"
)

// TemplateHandler handles HTTP requests for task templates and blueprints
type TemplateHandler struct {
	svc service.TemplateService
}

var GlobalTemplateHandler *TemplateHandler

// InitTemplateHandler initializes the global template handler singleton
func InitTemplateHandler(svc service.TemplateService) {
	GlobalTemplateHandler = &TemplateHandler{svc: svc}
}

// NewTemplateHandler creates a new TemplateHandler instance
func NewTemplateHandler(svc service.TemplateService) *TemplateHandler {
	return &TemplateHandler{svc: svc}
}

// HandleCreateTemplate handles template definition creation
func (h *TemplateHandler) HandleCreateTemplate(w http.ResponseWriter, r *http.Request) {
	var req model.CreateTemplateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid input: "+err.Error(), http.StatusBadRequest)
		return
	}

	tpl, err := h.svc.CreateTemplate(&req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrTemplateTitleRequired),
			errors.Is(err, service.ErrInvalidWorkspaceID),
			errors.Is(err, service.ErrInvalidUserID),
			errors.Is(err, service.ErrInvalidCategory),
			errors.Is(err, service.ErrInvalidPriority),
			errors.Is(err, service.ErrNegativeEstimate):
			http.Error(w, err.Error(), http.StatusBadRequest)
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(tpl)
}

// HandleGetTemplate returns template details by ID
func (h *TemplateHandler) HandleGetTemplate(w http.ResponseWriter, r *http.Request, id int64) {
	tpl, err := h.svc.GetTemplate(id)
	if err != nil {
		if errors.Is(err, service.ErrTemplateNotFound) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(tpl)
}

// HandleListTemplates returns paginated task templates filtered by workspace, category, or search
func (h *TemplateHandler) HandleListTemplates(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	filter := model.TemplateFilter{}

	if wsStr := q.Get("workspace_id"); wsStr != "" {
		if wsID, err := strconv.ParseInt(wsStr, 10, 64); err == nil {
			filter.WorkspaceID = &wsID
		}
	}
	if cat := q.Get("category"); cat != "" {
		filter.Category = &cat
	}
	if s := q.Get("search"); s != "" {
		filter.SearchQuery = &s
	}
	if archStr := q.Get("is_archived"); archStr != "" {
		if arch, err := strconv.ParseBool(archStr); err == nil {
			filter.IsArchived = &arch
		}
	}
	if limitStr := q.Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			filter.Limit = l
		}
	}
	if offsetStr := q.Get("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil {
			filter.Offset = o
		}
	}

	summary, err := h.svc.ListTemplates(filter)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(summary)
}

// HandleUpdateTemplate modifies template properties
func (h *TemplateHandler) HandleUpdateTemplate(w http.ResponseWriter, r *http.Request, id int64) {
	var req model.UpdateTemplateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid input: "+err.Error(), http.StatusBadRequest)
		return
	}

	tpl, err := h.svc.UpdateTemplate(id, &req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrTemplateNotFound):
			http.Error(w, err.Error(), http.StatusNotFound)
		case errors.Is(err, service.ErrTemplateTitleRequired),
			errors.Is(err, service.ErrInvalidCategory),
			errors.Is(err, service.ErrInvalidPriority),
			errors.Is(err, service.ErrNegativeEstimate):
			http.Error(w, err.Error(), http.StatusBadRequest)
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(tpl)
}

// HandleDeleteTemplate removes a template
func (h *TemplateHandler) HandleDeleteTemplate(w http.ResponseWriter, r *http.Request, id int64) {
	err := h.svc.DeleteTemplate(id)
	if err != nil {
		if errors.Is(err, service.ErrTemplateNotFound) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// HandleDuplicateTemplate clones an existing template
func (h *TemplateHandler) HandleDuplicateTemplate(w http.ResponseWriter, r *http.Request, id int64) {
	cloned, err := h.svc.DuplicateTemplate(id)
	if err != nil {
		if errors.Is(err, service.ErrTemplateNotFound) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(cloned)
}

// HandleExportTemplate exports a template definition as a portable bundle
func (h *TemplateHandler) HandleExportTemplate(w http.ResponseWriter, r *http.Request, id int64) {
	bundle, err := h.svc.ExportTemplate(id)
	if err != nil {
		if errors.Is(err, service.ErrTemplateNotFound) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(bundle)
}

// HandleImportTemplate imports a template bundle into a target workspace
func (h *TemplateHandler) HandleImportTemplate(w http.ResponseWriter, r *http.Request) {
	var req model.ImportTemplateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid input: "+err.Error(), http.StatusBadRequest)
		return
	}

	tpl, err := h.svc.ImportTemplate(&req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidWorkspaceID),
			errors.Is(err, service.ErrInvalidUserID),
			errors.Is(err, service.ErrInvalidBundle),
			errors.Is(err, service.ErrTemplateTitleRequired):
			http.Error(w, err.Error(), http.StatusBadRequest)
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(tpl)
}
