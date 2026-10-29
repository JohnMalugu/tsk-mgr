package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/JohnMalugu/tsk-mgr-api/internal/model"
	"github.com/JohnMalugu/tsk-mgr-api/internal/repository"
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
		case "key-results":
			h.HandleGoalKeyResults(w, r, id)
			return
		case "check-ins":
			h.HandleGoalCheckIns(w, r, id)
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

// HandleGoalKeyResults handles GET and POST for /v2/goals/{id}/key-results
func (h *GoalHandler) HandleGoalKeyResults(w http.ResponseWriter, r *http.Request, goalID int64) {
	switch r.Method {
	case http.MethodGet:
		krs, err := h.svc.ListKeyResults(goalID)
		if err != nil {
			if errors.Is(err, service.ErrGoalNotFound) {
				writeGoalError(w, http.StatusNotFound, "Goal not found")
				return
			}
			writeGoalError(w, http.StatusInternalServerError, "Failed to list key results: "+err.Error())
			return
		}
		if krs == nil {
			krs = []model.KeyResult{}
		}
		writeGoalJSON(w, http.StatusOK, krs)

	case http.MethodPost:
		var req model.CreateKeyResultRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeGoalError(w, http.StatusBadRequest, "Invalid request body: "+err.Error())
			return
		}

		kr, err := h.svc.CreateKeyResult(goalID, req)
		if err != nil {
			switch {
			case errors.Is(err, service.ErrGoalNotFound):
				writeGoalError(w, http.StatusNotFound, "Goal not found")
			case errors.Is(err, service.ErrKeyResultTitleRequired),
				errors.Is(err, service.ErrInvalidMetricType),
				errors.Is(err, service.ErrInvalidTargetValue):
				writeGoalError(w, http.StatusBadRequest, err.Error())
			default:
				writeGoalError(w, http.StatusInternalServerError, "Failed to create key result: "+err.Error())
			}
			return
		}
		writeGoalJSON(w, http.StatusCreated, kr)

	default:
		writeGoalError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

// HandleKeyResultsRoot routes /v2/goals/key-results/...
func (h *GoalHandler) HandleKeyResultsRoot(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/v2/goals/key-results/")
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		writeGoalError(w, http.StatusBadRequest, "Missing key result ID in path")
		return
	}

	krID, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		writeGoalError(w, http.StatusBadRequest, "Invalid key result ID format")
		return
	}

	if len(parts) == 1 {
		h.HandleKeyResultByID(w, r, krID)
		return
	}

	if parts[1] == "tasks" {
		if len(parts) == 2 {
			h.HandleKeyResultTasks(w, r, krID)
			return
		}
		if len(parts) == 3 && r.Method == http.MethodDelete {
			taskID, err := strconv.Atoi(parts[2])
			if err != nil {
				writeGoalError(w, http.StatusBadRequest, "Invalid task ID format")
				return
			}
			h.HandleUnlinkTask(w, r, krID, taskID)
			return
		}
	}

	writeGoalError(w, http.StatusNotFound, "Resource not found")
}

// HandleKeyResultByID handles GET, PUT/PATCH, and DELETE for /v2/goals/key-results/{id}
func (h *GoalHandler) HandleKeyResultByID(w http.ResponseWriter, r *http.Request, krID int64) {
	switch r.Method {
	case http.MethodGet:
		kr, err := h.svc.GetKeyResult(krID)
		if err != nil {
			if errors.Is(err, service.ErrKeyResultNotFound) {
				writeGoalError(w, http.StatusNotFound, "Key result not found")
				return
			}
			writeGoalError(w, http.StatusInternalServerError, "Failed to retrieve key result: "+err.Error())
			return
		}
		writeGoalJSON(w, http.StatusOK, kr)

	case http.MethodPut, http.MethodPatch:
		var req model.UpdateKeyResultRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeGoalError(w, http.StatusBadRequest, "Invalid request body: "+err.Error())
			return
		}

		kr, err := h.svc.UpdateKeyResult(krID, req)
		if err != nil {
			switch {
			case errors.Is(err, service.ErrKeyResultNotFound):
				writeGoalError(w, http.StatusNotFound, "Key result not found")
			case errors.Is(err, service.ErrKeyResultTitleRequired),
				errors.Is(err, service.ErrInvalidMetricType),
				errors.Is(err, service.ErrInvalidTargetValue):
				writeGoalError(w, http.StatusBadRequest, err.Error())
			default:
				writeGoalError(w, http.StatusInternalServerError, "Failed to update key result: "+err.Error())
			}
			return
		}
		writeGoalJSON(w, http.StatusOK, kr)

	case http.MethodDelete:
		if err := h.svc.DeleteKeyResult(krID); err != nil {
			if errors.Is(err, service.ErrKeyResultNotFound) {
				writeGoalError(w, http.StatusNotFound, "Key result not found")
				return
			}
			writeGoalError(w, http.StatusInternalServerError, "Failed to delete key result: "+err.Error())
			return
		}
		writeGoalJSON(w, http.StatusOK, map[string]interface{}{
			"success": true,
			"message": "Key result deleted successfully",
		})

	default:
		writeGoalError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

// HandleKeyResultTasks handles GET and POST for /v2/goals/key-results/{id}/tasks
func (h *GoalHandler) HandleKeyResultTasks(w http.ResponseWriter, r *http.Request, krID int64) {
	switch r.Method {
	case http.MethodGet:
		tasks, err := h.svc.ListLinkedTasks(krID)
		if err != nil {
			if errors.Is(err, service.ErrKeyResultNotFound) {
				writeGoalError(w, http.StatusNotFound, "Key result not found")
				return
			}
			writeGoalError(w, http.StatusInternalServerError, "Failed to list linked tasks: "+err.Error())
			return
		}
		if tasks == nil {
			tasks = []model.LinkedTaskInfo{}
		}
		writeGoalJSON(w, http.StatusOK, tasks)

	case http.MethodPost:
		var req model.LinkTaskToKeyResultRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeGoalError(w, http.StatusBadRequest, "Invalid request body: "+err.Error())
			return
		}

		if err := h.svc.LinkTask(krID, req.TaskID); err != nil {
			switch {
			case errors.Is(err, service.ErrKeyResultNotFound):
				writeGoalError(w, http.StatusNotFound, "Key result not found")
			case errors.Is(err, service.ErrTaskNotFound):
				writeGoalError(w, http.StatusNotFound, "Task not found")
			case errors.Is(err, repository.ErrTaskAlreadyLinked):
				writeGoalError(w, http.StatusBadRequest, err.Error())
			default:
				writeGoalError(w, http.StatusInternalServerError, "Failed to link task: "+err.Error())
			}
			return
		}

		writeGoalJSON(w, http.StatusCreated, map[string]interface{}{
			"success": true,
			"message": "Task linked successfully",
		})

	default:
		writeGoalError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

// HandleUnlinkTask handles DELETE /v2/goals/key-results/{id}/tasks/{taskId}
func (h *GoalHandler) HandleUnlinkTask(w http.ResponseWriter, r *http.Request, krID int64, taskID int) {
	if err := h.svc.UnlinkTask(krID, taskID); err != nil {
		switch {
		case errors.Is(err, service.ErrKeyResultNotFound):
			writeGoalError(w, http.StatusNotFound, "Key result not found")
		case errors.Is(err, repository.ErrKeyResultTaskLinkNF):
			writeGoalError(w, http.StatusNotFound, "Task link not found")
		default:
			writeGoalError(w, http.StatusInternalServerError, "Failed to unlink task: "+err.Error())
		}
		return
	}

	writeGoalJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Task unlinked successfully",
	})
}

// HandleGoalCheckIns handles GET and POST for /v2/goals/{id}/check-ins
func (h *GoalHandler) HandleGoalCheckIns(w http.ResponseWriter, r *http.Request, goalID int64) {
	switch r.Method {
	case http.MethodGet:
		checkIns, err := h.svc.ListCheckIns(goalID)
		if err != nil {
			if errors.Is(err, service.ErrGoalNotFound) {
				writeGoalError(w, http.StatusNotFound, "Goal not found")
				return
			}
			writeGoalError(w, http.StatusInternalServerError, "Failed to list check-ins: "+err.Error())
			return
		}
		if checkIns == nil {
			checkIns = []model.GoalCheckIn{}
		}
		writeGoalJSON(w, http.StatusOK, checkIns)

	case http.MethodPost:
		userID := getGoalUser(r)

		var req model.GoalCheckInRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeGoalError(w, http.StatusBadRequest, "Invalid request body: "+err.Error())
			return
		}

		checkIn, err := h.svc.CreateCheckIn(userID, goalID, req)
		if err != nil {
			switch {
			case errors.Is(err, service.ErrGoalNotFound):
				writeGoalError(w, http.StatusNotFound, "Goal not found")
			case errors.Is(err, service.ErrInvalidConfidenceRating):
				writeGoalError(w, http.StatusBadRequest, err.Error())
			default:
				writeGoalError(w, http.StatusInternalServerError, "Failed to record check-in: "+err.Error())
			}
			return
		}
		writeGoalJSON(w, http.StatusCreated, checkIn)

	default:
		writeGoalError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

// HandleGoalMetrics handles GET /v2/goals/metrics
func (h *GoalHandler) HandleGoalMetrics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeGoalError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var workspaceID int64 = 1
	if wsStr := r.URL.Query().Get("workspace_id"); wsStr != "" {
		if wsID, err := strconv.ParseInt(wsStr, 10, 64); err == nil && wsID > 0 {
			workspaceID = wsID
		}
	}

	metrics, err := h.svc.GetGoalMetrics(workspaceID)
	if err != nil {
		writeGoalError(w, http.StatusInternalServerError, "Failed to retrieve goal metrics: "+err.Error())
		return
	}

	writeGoalJSON(w, http.StatusOK, metrics)
}
