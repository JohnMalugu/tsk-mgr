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

// ProblemHandler exposes HTTP endpoints for ITIL Problem Management & Known Error Database (KEDB)
type ProblemHandler struct {
	svc service.ProblemService
}

var GlobalProblemHandler *ProblemHandler

// InitProblemHandler initializes the global problem handler
func InitProblemHandler(svc service.ProblemService) {
	GlobalProblemHandler = &ProblemHandler{svc: svc}
}

// NewProblemHandler creates a new ProblemHandler
func NewProblemHandler(svc service.ProblemService) *ProblemHandler {
	return &ProblemHandler{svc: svc}
}

func getProblemUser(r *http.Request) string {
	if uid, ok := r.Context().Value("userID").(string); ok && uid != "" {
		return uid
	}
	if uid := r.Header.Get("X-User-ID"); uid != "" {
		return uid
	}
	if uid := r.URL.Query().Get("user_id"); uid != "" {
		return uid
	}
	return "user_1"
}

func writeProblemJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func writeProblemError(w http.ResponseWriter, status int, msg string) {
	writeProblemJSON(w, status, map[string]string{"error": msg})
}

func handleProblemServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrProblemNotFound):
		writeProblemError(w, http.StatusNotFound, "problem not found")
	case errors.Is(err, service.ErrWorkspaceNotFound):
		writeProblemError(w, http.StatusNotFound, "workspace not found")
	case errors.Is(err, service.ErrProjectNotFound):
		writeProblemError(w, http.StatusNotFound, "project not found")
	case errors.Is(err, service.ErrIncidentNotFound):
		writeProblemError(w, http.StatusNotFound, "incident not found")
	case errors.Is(err, service.ErrCINotFound):
		writeProblemError(w, http.StatusNotFound, "configuration item not found")
	case errors.Is(err, service.ErrChangeNotFound):
		writeProblemError(w, http.StatusNotFound, "change request not found")
	case errors.Is(err, service.ErrTaskNotFound):
		writeProblemError(w, http.StatusNotFound, "task not found")
	case errors.Is(err, service.ErrProblemLinkNotFound):
		writeProblemError(w, http.StatusNotFound, "problem link not found")
	case errors.Is(err, service.ErrRCAStepNotFound):
		writeProblemError(w, http.StatusNotFound, "rca step not found")
	case errors.Is(err, service.ErrWorkspaceRequired):
		writeProblemError(w, http.StatusBadRequest, "workspace_id is required")
	case errors.Is(err, service.ErrProblemTitleRequired):
		writeProblemError(w, http.StatusBadRequest, "problem title is required")
	case errors.Is(err, service.ErrInvalidProblemStatus):
		writeProblemError(w, http.StatusBadRequest, "invalid problem status")
	case errors.Is(err, service.ErrInvalidProblemCategory):
		writeProblemError(w, http.StatusBadRequest, "invalid problem category")
	case errors.Is(err, service.ErrInvalidProblemPriority):
		writeProblemError(w, http.StatusBadRequest, "invalid problem priority")
	case errors.Is(err, service.ErrInvalidProblemUrgency):
		writeProblemError(w, http.StatusBadRequest, "invalid problem urgency")
	case errors.Is(err, service.ErrInvalidProblemImpact):
		writeProblemError(w, http.StatusBadRequest, "invalid problem impact")
	case errors.Is(err, service.ErrInvalidRCAMethod):
		writeProblemError(w, http.StatusBadRequest, "invalid root cause analysis method")
	case errors.Is(err, service.ErrInvalidProblemTransition):
		writeProblemError(w, http.StatusBadRequest, "illegal problem status transition")
	case errors.Is(err, service.ErrKnownErrorRequiresWorkaround):
		writeProblemError(w, http.StatusBadRequest, "known error requires an approved workaround or root cause description")
	case errors.Is(err, service.ErrProblemLinkExists):
		writeProblemError(w, http.StatusConflict, "problem link already exists")
	default:
		writeProblemError(w, http.StatusBadRequest, err.Error())
	}
}

// ListProblems handles GET /v2/problems
func (h *ProblemHandler) ListProblems(w http.ResponseWriter, r *http.Request) {
	wsID, _ := strconv.ParseInt(r.URL.Query().Get("workspace_id"), 10, 64)
	if wsID <= 0 {
		writeProblemError(w, http.StatusBadRequest, "workspace_id query param is required")
		return
	}

	filter := model.ProblemFilter{
		WorkspaceID:  wsID,
		Status:       r.URL.Query().Get("status"),
		Category:     r.URL.Query().Get("category"),
		Priority:     r.URL.Query().Get("priority"),
		AssignedTo:   r.URL.Query().Get("assigned_to"),
		AssignedTeam: r.URL.Query().Get("assigned_team"),
		Search:       r.URL.Query().Get("search"),
	}

	if projStr := r.URL.Query().Get("project_id"); projStr != "" {
		if pID, err := strconv.ParseInt(projStr, 10, 64); err == nil {
			filter.ProjectID = &pID
		}
	}

	if kedbStr := r.URL.Query().Get("is_known_error"); kedbStr != "" {
		isKEDB := strings.ToLower(kedbStr) == "true" || kedbStr == "1"
		filter.IsKnownError = &isKEDB
	}

	if limit, err := strconv.Atoi(r.URL.Query().Get("limit")); err == nil {
		filter.Limit = limit
	}
	if offset, err := strconv.Atoi(r.URL.Query().Get("offset")); err == nil {
		filter.Offset = offset
	}

	problems, total, err := h.svc.ListProblems(filter)
	if err != nil {
		handleProblemServiceError(w, err)
		return
	}

	writeProblemJSON(w, http.StatusOK, map[string]interface{}{
		"problems": problems,
		"total":    total,
		"limit":    filter.Limit,
		"offset":   filter.Offset,
	})
}

// CreateProblem handles POST /v2/problems
func (h *ProblemHandler) CreateProblem(w http.ResponseWriter, r *http.Request) {
	var req model.CreateProblemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeProblemError(w, http.StatusBadRequest, "invalid json payload")
		return
	}

	creatorID := getProblemUser(r)
	prob, err := h.svc.CreateProblem(creatorID, req)
	if err != nil {
		handleProblemServiceError(w, err)
		return
	}

	writeProblemJSON(w, http.StatusCreated, prob)
}

// GetProblem handles GET /v2/problems/{id}
func (h *ProblemHandler) GetProblem(w http.ResponseWriter, r *http.Request, id int64) {
	prob, err := h.svc.GetProblem(id)
	if err != nil {
		handleProblemServiceError(w, err)
		return
	}
	writeProblemJSON(w, http.StatusOK, prob)
}

// UpdateProblem handles PUT|PATCH /v2/problems/{id}
func (h *ProblemHandler) UpdateProblem(w http.ResponseWriter, r *http.Request, id int64) {
	var req model.UpdateProblemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeProblemError(w, http.StatusBadRequest, "invalid json payload")
		return
	}

	updaterID := getProblemUser(r)
	prob, err := h.svc.UpdateProblem(id, updaterID, req)
	if err != nil {
		handleProblemServiceError(w, err)
		return
	}

	writeProblemJSON(w, http.StatusOK, prob)
}

// DeleteProblem handles DELETE /v2/problems/{id}
func (h *ProblemHandler) DeleteProblem(w http.ResponseWriter, r *http.Request, id int64) {
	deleterID := getProblemUser(r)
	if err := h.svc.DeleteProblem(id, deleterID); err != nil {
		handleProblemServiceError(w, err)
		return
	}
	writeProblemJSON(w, http.StatusOK, map[string]string{"message": "problem deleted successfully"})
}

// UpdateProblemStatus handles POST|PATCH /v2/problems/{id}/status
func (h *ProblemHandler) UpdateProblemStatus(w http.ResponseWriter, r *http.Request, id int64) {
	var req model.UpdateProblemStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeProblemError(w, http.StatusBadRequest, "invalid json payload")
		return
	}

	updaterID := getProblemUser(r)
	prob, err := h.svc.UpdateProblemStatus(id, updaterID, req)
	if err != nil {
		handleProblemServiceError(w, err)
		return
	}

	writeProblemJSON(w, http.StatusOK, prob)
}
