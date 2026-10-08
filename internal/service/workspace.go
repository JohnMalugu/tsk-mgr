package service

import "github.com/JohnMalugu/tsk-mgr-api/internal/model"

type WorkspaceService interface {
	Create(w *model.Workspace) error
	Get(id string) (*model.Workspace, error)
	List() ([]*model.Workspace, error)
}

type workspaceServiceImpl struct {}
