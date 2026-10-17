package model

import "time"

// WorkspaceRole represents the access tier of a workspace member.
type WorkspaceRole string

const (
	RoleOwner  WorkspaceRole = "owner"
	RoleAdmin  WorkspaceRole = "admin"
	RoleMember WorkspaceRole = "member"
	RoleViewer WorkspaceRole = "viewer"
)

func (r WorkspaceRole) IsValid() bool {
	switch r {
	case RoleOwner, RoleAdmin, RoleMember, RoleViewer:
		return true
	default:
		return false
	}
}

// CanManageMembers returns true if role can add/remove members.
func (r WorkspaceRole) CanManageMembers() bool {
	return r == RoleOwner || r == RoleAdmin
}

// CanManageWorkspace returns true if role can update workspace settings.
func (r WorkspaceRole) CanManageWorkspace() bool {
	return r == RoleOwner || r == RoleAdmin
}

// WorkspaceDB represents a persistent team workspace in SQLite.
type WorkspaceDB struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug"`
	OwnerID   string    `json:"ownerId"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}
