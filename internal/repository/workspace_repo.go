package repository

import (
	"database/sql"

	"github.com/JohnMalugu/tsk-mgr-api/internal/model"
)

type WorkspaceRepository interface {
	Create(w *model.WorkspaceDB) error
	GetByID(id int) (*model.WorkspaceDB, error)
	GetBySlug(slug string) (*model.WorkspaceDB, error)
	ListUserWorkspaces(userID string) ([]*model.WorkspaceDB, error)
	Update(w *model.WorkspaceDB) error
	Delete(id int) error

	AddMember(m *model.WorkspaceMemberDB) error
	GetMember(workspaceID int, userID string) (*model.WorkspaceMemberDB, error)
	ListMembers(workspaceID int) ([]*model.WorkspaceMemberInfo, error)
	UpdateMemberRole(workspaceID int, userID string, role model.WorkspaceRole) error
	RemoveMember(workspaceID int, userID string) error

	CreateInvite(inv *model.WorkspaceInviteDB) error
	GetInviteByToken(token string) (*model.WorkspaceInviteDB, error)
	UpdateInviteStatus(id int, status string) error
}

type workspaceRepositoryImpl struct {
	db *sql.DB
}

func NewWorkspaceRepository(db *sql.DB) WorkspaceRepository {
	return &workspaceRepositoryImpl{db: db}
}
