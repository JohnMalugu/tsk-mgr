package service

import (
	"errors"
	"testing"
	"time"

	"github.com/JohnMalugu/tsk-mgr-api/internal/db"
	"github.com/JohnMalugu/tsk-mgr-api/internal/model"
	"github.com/JohnMalugu/tsk-mgr-api/internal/repository"
)

func setupTestWorkspaceSvc(t *testing.T) WorkspaceDBService {
	sqliteDB, err := db.InitDB(":memory:")
	if err != nil {
		t.Fatalf("failed to init db: %v", err)
	}

	wsRepo := repository.NewWorkspaceRepository(sqliteDB)
	return NewWorkspaceDBService(wsRepo)
}

func TestWorkspaceService_CreateAndGet(t *testing.T) {
	svc := setupTestWorkspaceSvc(t)
	ownerID := "user-alice"
	otherID := "user-bob"

	// 1. Missing name validation
	_, err := svc.CreateWorkspace(ownerID, &model.CreateWorkspaceRequest{Name: "   "})
	if !errors.Is(err, ErrWorkspaceNameRequired) {
		t.Fatalf("expected ErrWorkspaceNameRequired, got %v", err)
	}

	// 2. Successful creation
	ws, err := svc.CreateWorkspace(ownerID, &model.CreateWorkspaceRequest{
		Name: "Engineering Team",
		Slug: "eng-team",
	})
	if err != nil {
		t.Fatalf("unexpected create error: %v", err)
	}
	if ws.ID == 0 || ws.Slug != "eng-team" || ws.OwnerID != ownerID {
		t.Fatalf("unexpected workspace data: %+v", ws)
	}

	// 3. Duplicate slug fails
	_, err = svc.CreateWorkspace(otherID, &model.CreateWorkspaceRequest{
		Name: "Other Team",
		Slug: "eng-team",
	})
	if !errors.Is(err, ErrWorkspaceSlugTaken) {
		t.Fatalf("expected ErrWorkspaceSlugTaken, got %v", err)
	}

	// 4. Owner can GetWorkspace
	detail, err := svc.GetWorkspace(ownerID, ws.ID)
	if err != nil {
		t.Fatalf("unexpected get error: %v", err)
	}
	if detail.MyRole != model.RoleOwner || detail.MemberCount != 1 {
		t.Fatalf("expected role owner and 1 member, got %v and %d", detail.MyRole, detail.MemberCount)
	}

	// 5. Non-member gets unauthorized
	_, err = svc.GetWorkspace(otherID, ws.ID)
	if !errors.Is(err, ErrUnauthorizedWorkspace) {
		t.Fatalf("expected ErrUnauthorizedWorkspace, got %v", err)
	}
}

func TestWorkspaceService_ListUpdateDelete(t *testing.T) {
	svc := setupTestWorkspaceSvc(t)
	ownerID := "user-alice"
	otherID := "user-bob"

	ws, err := svc.CreateWorkspace(ownerID, &model.CreateWorkspaceRequest{Name: "Product Lab"})
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}

	// List workspaces
	list, err := svc.ListUserWorkspaces(ownerID)
	if err != nil || len(list) != 1 {
		t.Fatalf("expected 1 workspace for owner, got %d, err: %v", len(list), err)
	}
	bobList, err := svc.ListUserWorkspaces(otherID)
	if err != nil || len(bobList) != 0 {
		t.Fatalf("expected 0 workspaces for non-member, got %d", len(bobList))
	}

	// Update workspace
	_, err = svc.UpdateWorkspace(otherID, ws.ID, &model.UpdateWorkspaceRequest{Name: "Hacked Lab"})
	if !errors.Is(err, ErrUnauthorizedWorkspace) {
		t.Fatalf("expected ErrUnauthorizedWorkspace, got %v", err)
	}

	updated, err := svc.UpdateWorkspace(ownerID, ws.ID, &model.UpdateWorkspaceRequest{Name: "Product Innovation Lab"})
	if err != nil {
		t.Fatalf("update failed: %v", err)
	}
	if updated.Name != "Product Innovation Lab" {
		t.Fatalf("expected updated name, got %s", updated.Name)
	}

	// Delete workspace non-owner guard
	err = svc.DeleteWorkspace(otherID, ws.ID)
	if !errors.Is(err, ErrUnauthorizedWorkspace) {
		t.Fatalf("expected ErrUnauthorizedWorkspace on delete, got %v", err)
	}

	err = svc.DeleteWorkspace(ownerID, ws.ID)
	if err != nil {
		t.Fatalf("delete failed: %v", err)
	}

	// Verify not found after delete
	_, err = svc.GetWorkspace(ownerID, ws.ID)
	if !errors.Is(err, ErrWorkspaceNotFoundDB) {
		t.Fatalf("expected ErrWorkspaceNotFoundDB, got %v", err)
	}
}

func TestWorkspaceService_MemberLifecycle(t *testing.T) {
	svc := setupTestWorkspaceSvc(t)
	ownerID := "user-alice"
	memberID := "user-bob"
	viewerID := "user-charlie"

	ws, err := svc.CreateWorkspace(ownerID, &model.CreateWorkspaceRequest{Name: "Core Ops"})
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}

	// Add member by non-owner should fail
	_, err = svc.AddMember(memberID, ws.ID, &model.AddMemberRequest{
		UserID: memberID,
		Role:   model.RoleMember,
	})
	if !errors.Is(err, ErrUnauthorizedWorkspace) {
		t.Fatalf("expected ErrUnauthorizedWorkspace, got %v", err)
	}

	// Add member with invalid role
	_, err = svc.AddMember(ownerID, ws.ID, &model.AddMemberRequest{
		UserID: memberID,
		Role:   "superadmin",
	})
	if !errors.Is(err, ErrInvalidWorkspaceRole) {
		t.Fatalf("expected ErrInvalidWorkspaceRole, got %v", err)
	}

	// Add member successfully
	m, err := svc.AddMember(ownerID, ws.ID, &model.AddMemberRequest{
		UserID: memberID,
		Role:   model.RoleMember,
	})
	if err != nil {
		t.Fatalf("add member failed: %v", err)
	}
	if m.Role != model.RoleMember {
		t.Fatalf("expected role member, got %s", m.Role)
	}

	// Duplicate member fails
	_, err = svc.AddMember(ownerID, ws.ID, &model.AddMemberRequest{
		UserID: memberID,
		Role:   model.RoleAdmin,
	})
	if !errors.Is(err, ErrAlreadyWorkspaceMember) {
		t.Fatalf("expected ErrAlreadyWorkspaceMember, got %v", err)
	}

	// Add viewer
	_, err = svc.AddMember(ownerID, ws.ID, &model.AddMemberRequest{
		UserID: viewerID,
		Role:   model.RoleViewer,
	})
	if err != nil {
		t.Fatalf("add viewer failed: %v", err)
	}

	// List members
	members, err := svc.ListMembers(memberID, ws.ID)
	if err != nil || len(members) != 3 {
		t.Fatalf("expected 3 members, got %d, err: %v", len(members), err)
	}

	// Update role
	err = svc.UpdateMemberRole(ownerID, ws.ID, memberID, &model.UpdateMemberRoleRequest{
		Role: model.RoleAdmin,
	})
	if err != nil {
		t.Fatalf("update member role failed: %v", err)
	}

	// Cannot remove or demote owner
	err = svc.UpdateMemberRole(ownerID, ws.ID, ownerID, &model.UpdateMemberRoleRequest{Role: model.RoleViewer})
	if !errors.Is(err, ErrCannotRemoveOwner) {
		t.Fatalf("expected ErrCannotRemoveOwner, got %v", err)
	}

	err = svc.RemoveMember(ownerID, ws.ID, ownerID)
	if !errors.Is(err, ErrCannotRemoveOwner) {
		t.Fatalf("expected ErrCannotRemoveOwner, got %v", err)
	}

	// Self-leave
	err = svc.RemoveMember(viewerID, ws.ID, viewerID)
	if err != nil {
		t.Fatalf("self-leave failed: %v", err)
	}

	membersAfterLeave, err := svc.ListMembers(ownerID, ws.ID)
	if err != nil || len(membersAfterLeave) != 2 {
		t.Fatalf("expected 2 members after leave, got %d", len(membersAfterLeave))
	}
}

func TestWorkspaceService_Invites(t *testing.T) {
	svc := setupTestWorkspaceSvc(t)
	ownerID := "user-alice"
	inviteeID := "user-dave"
	unauthorizedID := "user-eve"

	ws, err := svc.CreateWorkspace(ownerID, &model.CreateWorkspaceRequest{Name: "Growth Team"})
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}

	// Unauthorized invite creation
	_, err = svc.InviteMember(unauthorizedID, ws.ID, &model.InviteMemberRequest{
		Email: "dave@example.com",
		Role:  model.RoleMember,
	})
	if !errors.Is(err, ErrUnauthorizedWorkspace) {
		t.Fatalf("expected ErrUnauthorizedWorkspace, got %v", err)
	}

	// Successful invite
	inv, err := svc.InviteMember(ownerID, ws.ID, &model.InviteMemberRequest{
		Email: "dave@example.com",
		Role:  model.RoleAdmin,
	})
	if err != nil {
		t.Fatalf("invite failed: %v", err)
	}
	if inv.Token == "" || inv.Status != "pending" || inv.ExpiresAt.Before(time.Now()) {
		t.Fatalf("unexpected invite state: %+v", inv)
	}

	// Invalid token fails
	_, err = svc.AcceptInvite(inviteeID, "invalid-token-123")
	if !errors.Is(err, ErrInviteNotFoundDB) {
		t.Fatalf("expected ErrInviteNotFoundDB, got %v", err)
	}

	// Accept invite successfully
	member, err := svc.AcceptInvite(inviteeID, inv.Token)
	if err != nil {
		t.Fatalf("accept invite failed: %v", err)
	}
	if member.Role != model.RoleAdmin || member.UserID != inviteeID {
		t.Fatalf("unexpected member from invite: %+v", member)
	}

	// Reusing token fails with ErrInviteExpired
	_, err = svc.AcceptInvite(inviteeID, inv.Token)
	if !errors.Is(err, ErrInviteExpired) {
		t.Fatalf("expected ErrInviteExpired on reused invite, got %v", err)
	}
}
