package handler

import "github.com/JohnMalugu/tsk-mgr-api/internal/service"

type WorkspaceHandler struct {
	service service.WorkspaceService
}

func NewWorkspaceHandler(s service.WorkspaceService) *WorkspaceHandler {
	return &WorkspaceHandler{service: s}
}
