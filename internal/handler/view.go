package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/JohnMalugu/tsk-mgr-api/internal/model"
	"github.com/JohnMalugu/tsk-mgr-api/internal/service"
)

// SavedViewHandler handles HTTP endpoints for saved views and custom filter presets
type SavedViewHandler struct {
	svc service.SavedViewService
}

var GlobalSavedViewHandler *SavedViewHandler

// InitSavedViewHandler initializes the global saved view handler
func InitSavedViewHandler(svc service.SavedViewService) {
	GlobalSavedViewHandler = &SavedViewHandler{svc: svc}
}

// NewSavedViewHandler constructs a new SavedViewHandler instance
func NewSavedViewHandler(svc service.SavedViewService) *SavedViewHandler {
	return &SavedViewHandler{svc: svc}
}

func getViewUser(r *http.Request) string {
	if uid, ok := r.Context().Value("userID").(string); ok && uid != "" {
		return uid
	}
	if uid := r.Header.Get("X-User-ID"); uid != "" {
		return uid
	}
	if uid := r.URL.Query().Get("user_id"); uid != "" {
		return uid
	}
	return "default_user"
}

// HandleCreate processes creation of a new saved view
func (h *SavedViewHandler) HandleCreate(w http.ResponseWriter, r *http.Request) {
	userID := getViewUser(r)

	var req model.CreateViewRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request body: " + err.Error()})
		return
	}

	view, err := h.svc.CreateView(userID, req)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case errors.Is(err, service.ErrWorkspaceRequired),
			errors.Is(err, service.ErrInvalidViewName),
			errors.Is(err, service.ErrInvalidSortField),
			errors.Is(err, service.ErrInvalidSortOrder),
			errors.Is(err, service.ErrInvalidGroupBy):
			w.WriteHeader(http.StatusBadRequest)
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
		_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(view)
}

// HandleGetByID retrieves a saved view by its ID
func (h *SavedViewHandler) HandleGetByID(w http.ResponseWriter, r *http.Request, id int64) {
	userID := getViewUser(r)

	view, err := h.svc.GetView(userID, id)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case errors.Is(err, service.ErrViewNotFound):
			w.WriteHeader(http.StatusNotFound)
		case errors.Is(err, service.ErrUnauthorizedViewAccess):
			w.WriteHeader(http.StatusForbidden)
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
		_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(view)
}

// HandleList lists saved views for a workspace
func (h *SavedViewHandler) HandleList(w http.ResponseWriter, r *http.Request) {
	userID := getViewUser(r)

	wsIDStr := r.URL.Query().Get("workspace_id")
	if wsIDStr == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "workspace_id query parameter is required"})
		return
	}

	wsID, err := strconv.ParseInt(wsIDStr, 10, 64)
	if err != nil || wsID <= 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "invalid workspace_id"})
		return
	}

	pinned := r.URL.Query().Get("pinned") == "true" || r.URL.Query().Get("pinned") == "1"

	views, err := h.svc.ListViews(userID, wsID, pinned)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	if views == nil {
		views = []*model.SavedView{}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(views)
}

// HandleUpdate updates an existing saved view
func (h *SavedViewHandler) HandleUpdate(w http.ResponseWriter, r *http.Request, id int64) {
	userID := getViewUser(r)

	var req model.UpdateViewRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request body: " + err.Error()})
		return
	}

	view, err := h.svc.UpdateView(userID, id, req)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case errors.Is(err, service.ErrViewNotFound):
			w.WriteHeader(http.StatusNotFound)
		case errors.Is(err, service.ErrUnauthorizedViewAccess):
			w.WriteHeader(http.StatusForbidden)
		case errors.Is(err, service.ErrInvalidViewName),
			errors.Is(err, service.ErrInvalidSortField),
			errors.Is(err, service.ErrInvalidSortOrder),
			errors.Is(err, service.ErrInvalidGroupBy):
			w.WriteHeader(http.StatusBadRequest)
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
		_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(view)
}

// HandleDelete removes a saved view
func (h *SavedViewHandler) HandleDelete(w http.ResponseWriter, r *http.Request, id int64) {
	userID := getViewUser(r)

	if err := h.svc.DeleteView(userID, id); err != nil {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case errors.Is(err, service.ErrViewNotFound):
			w.WriteHeader(http.StatusNotFound)
		case errors.Is(err, service.ErrUnauthorizedViewAccess):
			w.WriteHeader(http.StatusForbidden)
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
		_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"message": "View deleted successfully"})
}

// HandleDuplicate duplicates a saved view with an optional new name
func (h *SavedViewHandler) HandleDuplicate(w http.ResponseWriter, r *http.Request, id int64) {
	userID := getViewUser(r)

	var req model.DuplicateViewRequest
	if r.Body != nil {
		_ = json.NewDecoder(r.Body).Decode(&req)
	}

	cloned, err := h.svc.DuplicateView(userID, id, req)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case errors.Is(err, service.ErrViewNotFound):
			w.WriteHeader(http.StatusNotFound)
		case errors.Is(err, service.ErrUnauthorizedViewAccess):
			w.WriteHeader(http.StatusForbidden)
		case errors.Is(err, service.ErrInvalidViewName):
			w.WriteHeader(http.StatusBadRequest)
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
		_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(cloned)
}

// HandleTogglePin toggles the pinned status of a saved view
func (h *SavedViewHandler) HandleTogglePin(w http.ResponseWriter, r *http.Request, id int64) {
	userID := getViewUser(r)

	isPinned, err := h.svc.TogglePinView(userID, id)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case errors.Is(err, service.ErrViewNotFound):
			w.WriteHeader(http.StatusNotFound)
		case errors.Is(err, service.ErrUnauthorizedViewAccess):
			w.WriteHeader(http.StatusForbidden)
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
		_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"id":        id,
		"is_pinned": isPinned,
	})
}

// HandleExecute executes a saved view and returns filtered tasks with grouping metrics
func (h *SavedViewHandler) HandleExecute(w http.ResponseWriter, r *http.Request, id int64) {
	userID := getViewUser(r)

	result, err := h.svc.ExecuteView(userID, id)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case errors.Is(err, service.ErrViewNotFound):
			w.WriteHeader(http.StatusNotFound)
		case errors.Is(err, service.ErrUnauthorizedViewAccess):
			w.WriteHeader(http.StatusForbidden)
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
		_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(result)
}

