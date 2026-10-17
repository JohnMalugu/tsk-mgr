package service

import (
	"errors"
	"strings"

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

func slugify(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	var b strings.Builder
	lastDash := false
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
			lastDash = false
		} else if !lastDash && b.Len() > 0 {
			b.WriteRune('-')
			lastDash = true
		}
	}
	res := strings.Trim(b.String(), "-")
	if res == "" {
		return "workspace"
	}
	return res
}

func (s *workspaceDBServiceImpl) CreateWorkspace(userID string, req *model.CreateWorkspaceRequest) (*model.WorkspaceDB, error) {
	if req == nil {
		return nil, ErrWorkspaceNameRequired
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, ErrWorkspaceNameRequired
	}

	slug := strings.TrimSpace(req.Slug)
	if slug == "" {
		slug = slugify(name)
	} else {
		slug = slugify(slug)
	}

	existing, err := s.repo.GetBySlug(slug)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, ErrWorkspaceSlugTaken
	}

	ws := &model.WorkspaceDB{
		Name:    name,
		Slug:    slug,
		OwnerID: userID,
	}
	if err := s.repo.Create(ws); err != nil {
		return nil, err
	}

	member := &model.WorkspaceMemberDB{
		WorkspaceID: ws.ID,
		UserID:      userID,
		Role:        model.RoleOwner,
	}
	if err := s.repo.AddMember(member); err != nil {
		return nil, err
	}

	return ws, nil
}

func (s *workspaceDBServiceImpl) GetWorkspace(userID string, workspaceID int) (*model.WorkspaceDetailResponse, error) {
	ws, err := s.repo.GetByID(workspaceID)
	if err != nil {
		return nil, err
	}
	if ws == nil {
		return nil, ErrWorkspaceNotFoundDB
	}

	member, err := s.repo.GetMember(workspaceID, userID)
	if err != nil {
		return nil, err
	}
	if member == nil && ws.OwnerID != userID {
		return nil, ErrUnauthorizedWorkspace
	}

	role := model.RoleViewer
	if member != nil {
		role = member.Role
	} else if ws.OwnerID == userID {
		role = model.RoleOwner
	}

	members, err := s.repo.ListMembers(workspaceID)
	if err != nil {
		return nil, err
	}

	return &model.WorkspaceDetailResponse{
		Workspace:   ws,
		MyRole:      role,
		MemberCount: len(members),
		Members:     members,
	}, nil
}

func (s *workspaceDBServiceImpl) ListUserWorkspaces(userID string) ([]*model.WorkspaceDB, error) {
	return s.repo.ListUserWorkspaces(userID)
}

func (s *workspaceDBServiceImpl) UpdateWorkspace(userID string, workspaceID int, req *model.UpdateWorkspaceRequest) (*model.WorkspaceDB, error) {
	ws, err := s.repo.GetByID(workspaceID)
	if err != nil {
		return nil, err
	}
	if ws == nil {
		return nil, ErrWorkspaceNotFoundDB
	}

	member, err := s.repo.GetMember(workspaceID, userID)
	if err != nil {
		return nil, err
	}
	if (member == nil || !member.Role.CanManageWorkspace()) && ws.OwnerID != userID {
		return nil, ErrUnauthorizedWorkspace
	}

	if req == nil || strings.TrimSpace(req.Name) == "" {
		return nil, ErrWorkspaceNameRequired
	}

	ws.Name = strings.TrimSpace(req.Name)
	if err := s.repo.Update(ws); err != nil {
		return nil, err
	}

	return ws, nil
}

func (s *workspaceDBServiceImpl) DeleteWorkspace(userID string, workspaceID int) error {
	ws, err := s.repo.GetByID(workspaceID)
	if err != nil {
		return err
	}
	if ws == nil {
		return ErrWorkspaceNotFoundDB
	}

	if ws.OwnerID != userID {
		member, err := s.repo.GetMember(workspaceID, userID)
		if err != nil {
			return err
		}
		if member == nil || member.Role != model.RoleOwner {
			return ErrUnauthorizedWorkspace
		}
	}

	return s.repo.Delete(workspaceID)
}

func (s *workspaceDBServiceImpl) AddMember(actingUserID string, workspaceID int, req *model.AddMemberRequest) (*model.WorkspaceMemberDB, error) {
	if req == nil || strings.TrimSpace(req.UserID) == "" {
		return nil, errors.New("user_id is required")
	}

	ws, err := s.repo.GetByID(workspaceID)
	if err != nil {
		return nil, err
	}
	if ws == nil {
		return nil, ErrWorkspaceNotFoundDB
	}

	actor, err := s.repo.GetMember(workspaceID, actingUserID)
	if err != nil {
		return nil, err
	}
	if (actor == nil || !actor.Role.CanManageMembers()) && ws.OwnerID != actingUserID {
		return nil, ErrUnauthorizedWorkspace
	}

	role := req.Role
	if role == "" {
		role = model.RoleMember
	}
	if !role.IsValid() {
		return nil, ErrInvalidWorkspaceRole
	}

	existing, err := s.repo.GetMember(workspaceID, req.UserID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, ErrAlreadyWorkspaceMember
	}

	m := &model.WorkspaceMemberDB{
		WorkspaceID: workspaceID,
		UserID:      strings.TrimSpace(req.UserID),
		Role:        role,
	}
	if err := s.repo.AddMember(m); err != nil {
		return nil, err
	}

	return m, nil
}

func (s *workspaceDBServiceImpl) ListMembers(actingUserID string, workspaceID int) ([]*model.WorkspaceMemberInfo, error) {
	ws, err := s.repo.GetByID(workspaceID)
	if err != nil {
		return nil, err
	}
	if ws == nil {
		return nil, ErrWorkspaceNotFoundDB
	}

	actor, err := s.repo.GetMember(workspaceID, actingUserID)
	if err != nil {
		return nil, err
	}
	if actor == nil && ws.OwnerID != actingUserID {
		return nil, ErrUnauthorizedWorkspace
	}

	return s.repo.ListMembers(workspaceID)
}

func (s *workspaceDBServiceImpl) UpdateMemberRole(actingUserID string, workspaceID int, targetUserID string, req *model.UpdateMemberRoleRequest) error {
	if req == nil || !req.Role.IsValid() {
		return ErrInvalidWorkspaceRole
	}

	ws, err := s.repo.GetByID(workspaceID)
	if err != nil {
		return err
	}
	if ws == nil {
		return ErrWorkspaceNotFoundDB
	}

	actor, err := s.repo.GetMember(workspaceID, actingUserID)
	if err != nil {
		return err
	}
	if (actor == nil || !actor.Role.CanManageMembers()) && ws.OwnerID != actingUserID {
		return ErrUnauthorizedWorkspace
	}

	if ws.OwnerID == targetUserID {
		return ErrCannotRemoveOwner
	}

	target, err := s.repo.GetMember(workspaceID, targetUserID)
	if err != nil {
		return err
	}
	if target == nil {
		return errors.New("member not found")
	}

	return s.repo.UpdateMemberRole(workspaceID, targetUserID, req.Role)
}

func (s *workspaceDBServiceImpl) RemoveMember(actingUserID string, workspaceID int, targetUserID string) error {
	ws, err := s.repo.GetByID(workspaceID)
	if err != nil {
		return err
	}
	if ws == nil {
		return ErrWorkspaceNotFoundDB
	}

	if ws.OwnerID == targetUserID {
		return ErrCannotRemoveOwner
	}

	if actingUserID != targetUserID {
		actor, err := s.repo.GetMember(workspaceID, actingUserID)
		if err != nil {
			return err
		}
		if (actor == nil || !actor.Role.CanManageMembers()) && ws.OwnerID != actingUserID {
			return ErrUnauthorizedWorkspace
		}
	}

	target, err := s.repo.GetMember(workspaceID, targetUserID)
	if err != nil {
		return err
	}
	if target == nil {
		return errors.New("member not found")
	}

	return s.repo.RemoveMember(workspaceID, targetUserID)
}

func (s *workspaceDBServiceImpl) InviteMember(actingUserID string, workspaceID int, req *model.InviteMemberRequest) (*model.WorkspaceInviteDB, error) {
	return nil, nil
}

func (s *workspaceDBServiceImpl) AcceptInvite(userID string, token string) (*model.WorkspaceMemberDB, error) {
	return nil, nil
}
