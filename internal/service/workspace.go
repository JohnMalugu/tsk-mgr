package service

import "github.com/JohnMalugu/tsk-mgr-api/internal/model"

type WorkspaceService interface {
	Create(w *model.Workspace) error
	Get(id string) (*model.Workspace, error)
	List() ([]*model.Workspace, error)
}

type workspaceServiceImpl struct {}

func NewWorkspaceService() WorkspaceService {
	return &workspaceServiceImpl{}
}

func (s *workspaceServiceImpl) Create(w *model.Workspace) error { return nil }
func (s *workspaceServiceImpl) Get(id string) (*model.Workspace, error) { return nil, nil }
