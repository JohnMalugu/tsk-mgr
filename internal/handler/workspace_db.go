package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/JohnMalugu/tsk-mgr-api/internal/model"
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

func (h *WorkspaceDBHandler) HandleCreate(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(string)
	if !ok || userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req model.CreateWorkspaceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid input: "+err.Error(), http.StatusBadRequest)
		return
	}

	ws, err := h.svc.CreateWorkspace(userID, &req)
	if err != nil {
		if errors.Is(err, service.ErrWorkspaceNameRequired) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if errors.Is(err, service.ErrWorkspaceSlugTaken) {
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(ws)
}

func (h *WorkspaceDBHandler) HandleList(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(string)
	if !ok || userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	list, err := h.svc.ListUserWorkspaces(userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if list == nil {
		list = []*model.WorkspaceDB{}
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(list)
}

func (h *WorkspaceDBHandler) HandleGet(w http.ResponseWriter, r *http.Request, idStr string) {
	userID, ok := r.Context().Value("userID").(string)
	if !ok || userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid workspace ID", http.StatusBadRequest)
		return
	}

	detail, err := h.svc.GetWorkspace(userID, id)
	if err != nil {
		if errors.Is(err, service.ErrWorkspaceNotFoundDB) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		if errors.Is(err, service.ErrUnauthorizedWorkspace) {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(detail)
}

func (h *WorkspaceDBHandler) HandleUpdate(w http.ResponseWriter, r *http.Request, idStr string) {
	userID, ok := r.Context().Value("userID").(string)
	if !ok || userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid workspace ID", http.StatusBadRequest)
		return
	}

	var req model.UpdateWorkspaceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid input: "+err.Error(), http.StatusBadRequest)
		return
	}

	ws, err := h.svc.UpdateWorkspace(userID, id, &req)
	if err != nil {
		if errors.Is(err, service.ErrWorkspaceNotFoundDB) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		if errors.Is(err, service.ErrUnauthorizedWorkspace) {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}
		if errors.Is(err, service.ErrWorkspaceNameRequired) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(ws)
}

func (h *WorkspaceDBHandler) HandleDelete(w http.ResponseWriter, r *http.Request, idStr string) {
	userID, ok := r.Context().Value("userID").(string)
	if !ok || userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid workspace ID", http.StatusBadRequest)
		return
	}

	if err := h.svc.DeleteWorkspace(userID, id); err != nil {
		if errors.Is(err, service.ErrWorkspaceNotFoundDB) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		if errors.Is(err, service.ErrUnauthorizedWorkspace) {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"message": "workspace deleted successfully"})
}

func (h *WorkspaceDBHandler) HandleListMembers(w http.ResponseWriter, r *http.Request, wsIDStr string) {
	userID, ok := r.Context().Value("userID").(string)
	if !ok || userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	wsID, err := strconv.Atoi(wsIDStr)
	if err != nil {
		http.Error(w, "Invalid workspace ID", http.StatusBadRequest)
		return
	}

	members, err := h.svc.ListMembers(userID, wsID)
	if err != nil {
		if errors.Is(err, service.ErrWorkspaceNotFoundDB) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		if errors.Is(err, service.ErrUnauthorizedWorkspace) {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if members == nil {
		members = []*model.WorkspaceMemberInfo{}
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(members)
}

func (h *WorkspaceDBHandler) HandleAddMember(w http.ResponseWriter, r *http.Request, wsIDStr string) {
	userID, ok := r.Context().Value("userID").(string)
	if !ok || userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	wsID, err := strconv.Atoi(wsIDStr)
	if err != nil {
		http.Error(w, "Invalid workspace ID", http.StatusBadRequest)
		return
	}

	var req model.AddMemberRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid input: "+err.Error(), http.StatusBadRequest)
		return
	}

	member, err := h.svc.AddMember(userID, wsID, &req)
	if err != nil {
		if errors.Is(err, service.ErrWorkspaceNotFoundDB) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		if errors.Is(err, service.ErrUnauthorizedWorkspace) {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}
		if errors.Is(err, service.ErrAlreadyWorkspaceMember) {
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}
		if errors.Is(err, service.ErrInvalidWorkspaceRole) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(member)
}

func (h *WorkspaceDBHandler) HandleUpdateMemberRole(w http.ResponseWriter, r *http.Request, wsIDStr, targetUserID string) {
	userID, ok := r.Context().Value("userID").(string)
	if !ok || userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	wsID, err := strconv.Atoi(wsIDStr)
	if err != nil {
		http.Error(w, "Invalid workspace ID", http.StatusBadRequest)
		return
	}

	var req model.UpdateMemberRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid input: "+err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.svc.UpdateMemberRole(userID, wsID, targetUserID, &req); err != nil {
		if errors.Is(err, service.ErrWorkspaceNotFoundDB) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		if errors.Is(err, service.ErrUnauthorizedWorkspace) {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}
		if errors.Is(err, service.ErrCannotRemoveOwner) || errors.Is(err, service.ErrInvalidWorkspaceRole) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"message": "member role updated successfully"})
}

func (h *WorkspaceDBHandler) HandleRemoveMember(w http.ResponseWriter, r *http.Request, wsIDStr, targetUserID string) {
	userID, ok := r.Context().Value("userID").(string)
	if !ok || userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	wsID, err := strconv.Atoi(wsIDStr)
	if err != nil {
		http.Error(w, "Invalid workspace ID", http.StatusBadRequest)
		return
	}

	if err := h.svc.RemoveMember(userID, wsID, targetUserID); err != nil {
		if errors.Is(err, service.ErrWorkspaceNotFoundDB) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		if errors.Is(err, service.ErrUnauthorizedWorkspace) {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}
		if errors.Is(err, service.ErrCannotRemoveOwner) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"message": "member removed successfully"})
}

func (h *WorkspaceDBHandler) HandleInviteMember(w http.ResponseWriter, r *http.Request, wsIDStr string) {
	userID, ok := r.Context().Value("userID").(string)
	if !ok || userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	wsID, err := strconv.Atoi(wsIDStr)
	if err != nil {
		http.Error(w, "Invalid workspace ID", http.StatusBadRequest)
		return
	}

	var req model.InviteMemberRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid input: "+err.Error(), http.StatusBadRequest)
		return
	}

	inv, err := h.svc.InviteMember(userID, wsID, &req)
	if err != nil {
		if errors.Is(err, service.ErrWorkspaceNotFoundDB) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		if errors.Is(err, service.ErrUnauthorizedWorkspace) {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}
		if errors.Is(err, service.ErrInvalidWorkspaceRole) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(inv)
}

func (h *WorkspaceDBHandler) HandleAcceptInvite(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(string)
	if !ok || userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid input: "+err.Error(), http.StatusBadRequest)
		return
	}

	member, err := h.svc.AcceptInvite(userID, req.Token)
	if err != nil {
		if errors.Is(err, service.ErrInviteNotFoundDB) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		if errors.Is(err, service.ErrInviteExpired) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if errors.Is(err, service.ErrWorkspaceNotFoundDB) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(member)
}




