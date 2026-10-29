package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/JohnMalugu/tsk-mgr-api/internal/model"
	"github.com/JohnMalugu/tsk-mgr-api/internal/service"
)

// GoalHandler handles HTTP endpoints for Goals, Key Results, task linkages, and OKR progress check-ins
type GoalHandler struct {
	svc service.GoalService
}

var GlobalGoalHandler *GoalHandler

// InitGoalHandler initializes the global goal handler singleton
func InitGoalHandler(svc service.GoalService) {
	GlobalGoalHandler = &GoalHandler{svc: svc}
}

// NewGoalHandler constructs a new GoalHandler instance
func NewGoalHandler(svc service.GoalService) *GoalHandler {
	return &GoalHandler{svc: svc}
}

func getGoalUser(r *http.Request) string {
	if uid, ok := r.Context().Value("userID").(string); ok && uid != "" {
		return uid
	}
	if uid := r.Header.Get("X-User-ID"); uid != "" {
		return uid
	}
	if uid := r.URL.Query().Get("user_id"); uid != "" {
		return uid
	}
	return "default_user"
}

func writeGoalJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func writeGoalError(w http.ResponseWriter, status int, msg string) {
	writeGoalJSON(w, status, map[string]string{"error": msg})
}

// HandleGoalsRoot routes /v2/goals for creating goals (POST) and listing goals (GET)
func (h *GoalHandler) HandleGoalsRoot(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		h.HandleCreateGoal(w, r)
	case http.MethodGet:
		h.HandleListGoals(w, r)
	default:
		writeGoalError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

// HandleCreateGoal creates a new strategic goal
func (h *GoalHandler) HandleCreateGoal(w http.ResponseWriter, r *http.Request) {
	userID := getGoalUser(r)

	var req model.CreateGoalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeGoalError(w, http.StatusBadRequest, "Invalid request body: "+err.Error())
		return
	}

	goal, err := h.svc.CreateGoal(userID, req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrGoalTitleRequired),
			errors.Is(err, service.ErrGoalWorkspaceRequired),
			errors.Is(err, service.ErrGoalTimeframeRequired):
			writeGoalError(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, service.ErrGoalNotFound):
			writeGoalError(w, http.StatusNotFound, "Parent goal not found")
		default:
			writeGoalError(w, http.StatusInternalServerError, "Failed to create goal: "+err.Error())
		}
		return
	}

	writeGoalJSON(w, http.StatusCreated, goal)
}

// HandleListGoals lists goals with optional multi-criteria filters
func (h *GoalHandler) HandleListGoals(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	filter := model.GoalFilter{}

	if wsStr := q.Get("workspace_id"); wsStr != "" {
		if wsID, err := strconv.ParseInt(wsStr, 10, 64); err == nil && wsID > 0 {
			filter.WorkspaceID = &wsID
		}
	}
	if owner := q.Get("owner_user_id"); owner != "" {
		filter.OwnerUserID = owner
	}
	if status := q.Get("status"); status != "" {
		filter.Status = status
	}
	if tf := q.Get("timeframe"); tf != "" {
		filter.Timeframe = tf
	}
	if parentStr := q.Get("parent_goal_id"); parentStr != "" {
		if pID, err := strconv.ParseInt(parentStr, 10, 64); err == nil {
			filter.ParentGoalID = &pID
		}
	}
	if limStr := q.Get("limit"); limStr != "" {
		if lim, err := strconv.Atoi(limStr); err == nil {
			filter.Limit = lim
		}
	}
	if offStr := q.Get("offset"); offStr != "" {
		if off, err := strconv.Atoi(offStr); err == nil {
			filter.Offset = off
		}
	}

	goals, err := h.svc.ListGoals(filter)
	if err != nil {
		writeGoalError(w, http.StatusInternalServerError, "Failed to list goals: "+err.Error())
		return
	}

	if goals == nil {
		goals = []*model.GoalSummary{}
	}

	writeGoalJSON(w, http.StatusOK, goals)
}

// HandleGoalByID routes GET, PUT/PATCH, and DELETE for /v2/goals/{id}
func (h *GoalHandler) HandleGoalByID(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/v2/goals/")
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		writeGoalError(w, http.StatusBadRequest, "Missing goal ID in path")
		return
	}

	id, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		writeGoalError(w, http.StatusBadRequest, "Invalid goal ID format")
		return
	}

	if len(parts) > 1 {
		switch parts[1] {
		case "tree":
			if r.Method == http.MethodGet {
				h.HandleGetGoalTree(w, r, id)
				return
			}
			writeGoalError(w, http.StatusMethodNotAllowed, "Method not allowed")
			return
		}
	}

	switch r.Method {
	case http.MethodGet:
		h.HandleGetGoal(w, r, id)
	case http.MethodPut, http.MethodPatch:
		h.HandleUpdateGoal(w, r, id)
	case http.MethodDelete:
		h.HandleDeleteGoal(w, r, id)
	default:
		writeGoalError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

// HandleGetGoal fetches goal details including key results and latest check-in
func (h *GoalHandler) HandleGetGoal(w http.ResponseWriter, r *http.Request, id int64) {
	detail, err := h.svc.GetGoal(id)
	if err != nil {
		if errors.Is(err, service.ErrGoalNotFound) {
			writeGoalError(w, http.StatusNotFound, "Goal not found")
			return
		}
		writeGoalError(w, http.StatusInternalServerError, "Failed to retrieve goal: "+err.Error())
		return
	}

	writeGoalJSON(w, http.StatusOK, detail)
}

// HandleUpdateGoal updates an existing goal
func (h *GoalHandler) HandleUpdateGoal(w http.ResponseWriter, r *http.Request, id int64) {
	var req model.UpdateGoalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeGoalError(w, http.StatusBadRequest, "Invalid request body: "+err.Error())
		return
	}

	goal, err := h.svc.UpdateGoal(id, req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrGoalNotFound):
			writeGoalError(w, http.StatusNotFound, "Goal not found")
		case errors.Is(err, service.ErrGoalTitleRequired),
			errors.Is(err, service.ErrGoalTimeframeRequired):
			writeGoalError(w, http.StatusBadRequest, err.Error())
		default:
			writeGoalError(w, http.StatusInternalServerError, "Failed to update goal: "+err.Error())
		}
		return
	}

	writeGoalJSON(w, http.StatusOK, goal)
}

// HandleDeleteGoal deletes a goal by ID
func (h *GoalHandler) HandleDeleteGoal(w http.ResponseWriter, r *http.Request, id int64) {
	if err := h.svc.DeleteGoal(id); err != nil {
		if errors.Is(err, service.ErrGoalNotFound) {
			writeGoalError(w, http.StatusNotFound, "Goal not found")
			return
		}
		writeGoalError(w, http.StatusInternalServerError, "Failed to delete goal: "+err.Error())
		return
	}

	writeGoalJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Goal deleted successfully",
	})
}

// HandleGetGoalTree retrieves the hierarchical goal tree with sub-goals
func (h *GoalHandler) HandleGetGoalTree(w http.ResponseWriter, r *http.Request, id int64) {
	tree, err := h.svc.GetGoalTree(id)
	if err != nil {
		if errors.Is(err, service.ErrGoalNotFound) {
			writeGoalError(w, http.StatusNotFound, "Goal not found")
			return
		}
		writeGoalError(w, http.StatusInternalServerError, "Failed to get goal tree: "+err.Error())
		return
	}

	writeGoalJSON(w, http.StatusOK, tree)
}
