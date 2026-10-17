package handler

import (
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
