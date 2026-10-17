package service

import (
	"errors"

	"github.com/JohnMalugu/tsk-mgr-api/internal/model"
	"github.com/JohnMalugu/tsk-mgr-api/internal/repository"
)

var (
	ErrWorkspaceNotFoundDB    = errors.New("workspace not found")
	ErrWorkspaceSlugTaken     = errors.New("workspace slug already taken")
	ErrWorkspaceNameRequired  = errors.New("workspace name is required")
	ErrUnauthorizedWorkspace  = errors.New("unauthorized: insufficient workspace permissions")
	ErrCannotRemoveOwner      = errors.New("cannot remove workspace owner")
	ErrAlreadyWorkspaceMember = errors.New("user is already a member of this workspace")
	ErrInvalidWorkspaceRole   = errors.New("invalid workspace role")
	ErrInviteNotFoundDB       = errors.New("invite not found")
	ErrInviteExpired          = errors.New("invite has expired or already been accepted")
)

type WorkspaceDBService interface {
	CreateWorkspace(userID string, req *model.CreateWorkspaceRequest) (*model.WorkspaceDB, error)
	GetWorkspace(userID string, workspaceID int) (*model.WorkspaceDetailResponse, error)
	ListUserWorkspaces(userID string) ([]*model.WorkspaceDB, error)
	UpdateWorkspace(userID string, workspaceID int, req *model.UpdateWorkspaceRequest) (*model.WorkspaceDB, error)
	DeleteWorkspace(userID string, workspaceID int) error

	AddMember(actingUserID string, workspaceID int, req *model.AddMemberRequest) (*model.WorkspaceMemberDB, error)
	ListMembers(actingUserID string, workspaceID int) ([]*model.WorkspaceMemberInfo, error)
	UpdateMemberRole(actingUserID string, workspaceID int, targetUserID string, req *model.UpdateMemberRoleRequest) error
	RemoveMember(actingUserID string, workspaceID int, targetUserID string) error

	InviteMember(actingUserID string, workspaceID int, req *model.InviteMemberRequest) (*model.WorkspaceInviteDB, error)
	AcceptInvite(userID string, token string) (*model.WorkspaceMemberDB, error)
}

type workspaceDBServiceImpl struct {
	repo repository.WorkspaceRepository
}

func NewWorkspaceDBService(repo repository.WorkspaceRepository) WorkspaceDBService {
	return &workspaceDBServiceImpl{repo: repo}
}

func (s *workspaceDBServiceImpl) CreateWorkspace(userID string, req *model.CreateWorkspaceRequest) (*model.WorkspaceDB, error) {
	return nil, nil
}

func (s *workspaceDBServiceImpl) GetWorkspace(userID string, workspaceID int) (*model.WorkspaceDetailResponse, error) {
	return nil, nil
}

func (s *workspaceDBServiceImpl) ListUserWorkspaces(userID string) ([]*model.WorkspaceDB, error) {
	return nil, nil
}

func (s *workspaceDBServiceImpl) UpdateWorkspace(userID string, workspaceID int, req *model.UpdateWorkspaceRequest) (*model.WorkspaceDB, error) {
	return nil, nil
}

func (s *workspaceDBServiceImpl) DeleteWorkspace(userID string, workspaceID int) error {
	return nil
}

func (s *workspaceDBServiceImpl) AddMember(actingUserID string, workspaceID int, req *model.AddMemberRequest) (*model.WorkspaceMemberDB, error) {
	return nil, nil
}

func (s *workspaceDBServiceImpl) ListMembers(actingUserID string, workspaceID int) ([]*model.WorkspaceMemberInfo, error) {
	return nil, nil
}

func (s *workspaceDBServiceImpl) UpdateMemberRole(actingUserID string, workspaceID int, targetUserID string, req *model.UpdateMemberRoleRequest) error {
	return nil
}

func (s *workspaceDBServiceImpl) RemoveMember(actingUserID string, workspaceID int, targetUserID string) error {
	return nil
}

func (s *workspaceDBServiceImpl) InviteMember(actingUserID string, workspaceID int, req *model.InviteMemberRequest) (*model.WorkspaceInviteDB, error) {
	return nil, nil
}

func (s *workspaceDBServiceImpl) AcceptInvite(userID string, token string) (*model.WorkspaceMemberDB, error) {
	return nil, nil
}
