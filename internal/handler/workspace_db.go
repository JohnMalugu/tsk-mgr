package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/JohnMalugu/tsk-mgr-api/internal/model"
	"github.com/JohnMalugu/tsk-mgr-api/internal/service"
)

type WorkspaceDBHandler struct {
	svc service.WorkspaceDBService
}

var GlobalWorkspaceDBHandler *WorkspaceDBHandler

func InitWorkspaceDBHandler(svc service.WorkspaceDBService) {
	GlobalWorkspaceDBHandler = &WorkspaceDBHandler{svc: svc}
}

func NewWorkspaceDBHandler(svc service.WorkspaceDBService) *WorkspaceDBHandler {
	return &WorkspaceDBHandler{svc: svc}
}

func (h *WorkspaceDBHandler) HandleCreate(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(string)
	if !ok || userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req model.CreateWorkspaceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid input: "+err.Error(), http.StatusBadRequest)
		return
	}

	ws, err := h.svc.CreateWorkspace(userID, &req)
	if err != nil {
		if errors.Is(err, service.ErrWorkspaceNameRequired) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if errors.Is(err, service.ErrWorkspaceSlugTaken) {
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(ws)
}

func (h *WorkspaceDBHandler) HandleList(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(string)
	if !ok || userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	list, err := h.svc.ListUserWorkspaces(userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if list == nil {
		list = []*model.WorkspaceDB{}
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(list)
}

func (h *WorkspaceDBHandler) HandleGet(w http.ResponseWriter, r *http.Request, idStr string) {
	userID, ok := r.Context().Value("userID").(string)
	if !ok || userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid workspace ID", http.StatusBadRequest)
		return
	}

	detail, err := h.svc.GetWorkspace(userID, id)
	if err != nil {
		if errors.Is(err, service.ErrWorkspaceNotFoundDB) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		if errors.Is(err, service.ErrUnauthorizedWorkspace) {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(detail)
}

