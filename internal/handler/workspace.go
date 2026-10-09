package handler

import (
	"github.com/JohnMalugu/tsk-mgr-api/internal/service"
	"net/http"
)

type WorkspaceHandler struct {
	service service.WorkspaceService
}

func NewWorkspaceHandler(s service.WorkspaceService) *WorkspaceHandler {
	return &WorkspaceHandler{service: s}
}

func (h *WorkspaceHandler) HandleList(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func (h *WorkspaceHandler) HandleCreate(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusCreated)
}
