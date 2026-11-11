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

// Incident Linkages
func (h *ProblemHandler) ListProblemIncidents(w http.ResponseWriter, r *http.Request, problemID int64) {
	links, err := h.svc.ListProblemIncidents(problemID)
	if err != nil {
		handleProblemServiceError(w, err)
		return
	}
	writeProblemJSON(w, http.StatusOK, map[string]interface{}{"incidents": links, "total": len(links)})
}

func (h *ProblemHandler) LinkIncident(w http.ResponseWriter, r *http.Request, problemID int64) {
	var req model.LinkIncidentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeProblemError(w, http.StatusBadRequest, "invalid json payload")
		return
	}
	linkerID := getProblemUser(r)
	link, err := h.svc.LinkIncident(problemID, linkerID, req)
	if err != nil {
		handleProblemServiceError(w, err)
		return
	}
	writeProblemJSON(w, http.StatusCreated, link)
}

func (h *ProblemHandler) UnlinkIncident(w http.ResponseWriter, r *http.Request, problemID, incidentID int64) {
	unlinkerID := getProblemUser(r)
	if err := h.svc.UnlinkIncident(problemID, incidentID, unlinkerID); err != nil {
		handleProblemServiceError(w, err)
		return
	}
	writeProblemJSON(w, http.StatusOK, map[string]string{"message": "incident unlinked successfully"})
}

// CI Linkages
func (h *ProblemHandler) ListProblemCIs(w http.ResponseWriter, r *http.Request, problemID int64) {
	links, err := h.svc.ListProblemCIs(problemID)
	if err != nil {
		handleProblemServiceError(w, err)
		return
	}
	writeProblemJSON(w, http.StatusOK, map[string]interface{}{"cis": links, "total": len(links)})
}

func (h *ProblemHandler) LinkCI(w http.ResponseWriter, r *http.Request, problemID int64) {
	var req model.LinkCIRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeProblemError(w, http.StatusBadRequest, "invalid json payload")
		return
	}
	linkerID := getProblemUser(r)
	link, err := h.svc.LinkCI(problemID, linkerID, req)
	if err != nil {
		handleProblemServiceError(w, err)
		return
	}
	writeProblemJSON(w, http.StatusCreated, link)
}

func (h *ProblemHandler) UnlinkCI(w http.ResponseWriter, r *http.Request, problemID, linkID int64) {
	unlinkerID := getProblemUser(r)
	if err := h.svc.UnlinkCI(problemID, linkID, unlinkerID); err != nil {
		handleProblemServiceError(w, err)
		return
	}
	writeProblemJSON(w, http.StatusOK, map[string]string{"message": "configuration item unlinked successfully"})
}

// Change Linkages
func (h *ProblemHandler) ListProblemChanges(w http.ResponseWriter, r *http.Request, problemID int64) {
	links, err := h.svc.ListProblemChanges(problemID)
	if err != nil {
		handleProblemServiceError(w, err)
		return
	}
	writeProblemJSON(w, http.StatusOK, map[string]interface{}{"changes": links, "total": len(links)})
}

func (h *ProblemHandler) LinkChange(w http.ResponseWriter, r *http.Request, problemID int64) {
	var req model.LinkChangeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeProblemError(w, http.StatusBadRequest, "invalid json payload")
		return
	}
	linkerID := getProblemUser(r)
	link, err := h.svc.LinkChange(problemID, linkerID, req)
	if err != nil {
		handleProblemServiceError(w, err)
		return
	}
	writeProblemJSON(w, http.StatusCreated, link)
}

func (h *ProblemHandler) UnlinkChange(w http.ResponseWriter, r *http.Request, problemID, linkID int64) {
	unlinkerID := getProblemUser(r)
	if err := h.svc.UnlinkChange(problemID, linkID, unlinkerID); err != nil {
		handleProblemServiceError(w, err)
		return
	}
	writeProblemJSON(w, http.StatusOK, map[string]string{"message": "change request unlinked successfully"})
}

// Task Linkages
func (h *ProblemHandler) ListProblemTasks(w http.ResponseWriter, r *http.Request, problemID int64) {
	links, err := h.svc.ListProblemTasks(problemID)
	if err != nil {
		handleProblemServiceError(w, err)
		return
	}
	writeProblemJSON(w, http.StatusOK, map[string]interface{}{"tasks": links, "total": len(links)})
}

func (h *ProblemHandler) LinkTask(w http.ResponseWriter, r *http.Request, problemID int64) {
	var req model.LinkTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeProblemError(w, http.StatusBadRequest, "invalid json payload")
		return
	}
	linkerID := getProblemUser(r)
	link, err := h.svc.LinkTask(problemID, linkerID, req)
	if err != nil {
		handleProblemServiceError(w, err)
		return
	}
	writeProblemJSON(w, http.StatusCreated, link)
}

func (h *ProblemHandler) UnlinkTask(w http.ResponseWriter, r *http.Request, problemID, linkID int64) {
	unlinkerID := getProblemUser(r)
	if err := h.svc.UnlinkTask(problemID, linkID, unlinkerID); err != nil {
		handleProblemServiceError(w, err)
		return
	}
	writeProblemJSON(w, http.StatusOK, map[string]string{"message": "task unlinked successfully"})
}

// 5-Whys RCA Steps
func (h *ProblemHandler) ListRCASteps(w http.ResponseWriter, r *http.Request, problemID int64) {
	steps, err := h.svc.ListRCASteps(problemID)
	if err != nil {
		handleProblemServiceError(w, err)
		return
	}
	writeProblemJSON(w, http.StatusOK, map[string]interface{}{"rca_steps": steps, "total": len(steps)})
}

func (h *ProblemHandler) AddRCAStep(w http.ResponseWriter, r *http.Request, problemID int64) {
	var req model.CreateRCAStepRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeProblemError(w, http.StatusBadRequest, "invalid json payload")
		return
	}
	authorID := getProblemUser(r)
	step, err := h.svc.AddRCAStep(problemID, authorID, req)
	if err != nil {
		handleProblemServiceError(w, err)
		return
	}
	writeProblemJSON(w, http.StatusCreated, step)
}

func (h *ProblemHandler) UpdateRCAStep(w http.ResponseWriter, r *http.Request, stepID int64) {
	var req model.UpdateRCAStepRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeProblemError(w, http.StatusBadRequest, "invalid json payload")
		return
	}
	updaterID := getProblemUser(r)
	step, err := h.svc.UpdateRCAStep(stepID, updaterID, req)
	if err != nil {
		handleProblemServiceError(w, err)
		return
	}
	writeProblemJSON(w, http.StatusOK, step)
}

func (h *ProblemHandler) DeleteRCAStep(w http.ResponseWriter, r *http.Request, stepID int64) {
	deleterID := getProblemUser(r)
	if err := h.svc.DeleteRCAStep(stepID, deleterID); err != nil {
		handleProblemServiceError(w, err)
		return
	}
	writeProblemJSON(w, http.StatusOK, map[string]string{"message": "rca step deleted successfully"})
}

// Known Error Database (KEDB)
func (h *ProblemHandler) PublishKnownError(w http.ResponseWriter, r *http.Request, problemID int64) {
	var req model.PublishKnownErrorRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeProblemError(w, http.StatusBadRequest, "invalid json payload")
		return
	}
	publisherID := getProblemUser(r)
	prob, err := h.svc.PublishKnownError(problemID, publisherID, req)
	if err != nil {
		handleProblemServiceError(w, err)
		return
	}
	writeProblemJSON(w, http.StatusOK, prob)
}

func (h *ProblemHandler) RetireKnownError(w http.ResponseWriter, r *http.Request, problemID int64) {
	actorID := getProblemUser(r)
	prob, err := h.svc.RetireKnownError(problemID, actorID)
	if err != nil {
		handleProblemServiceError(w, err)
		return
	}
	writeProblemJSON(w, http.StatusOK, prob)
}

func (h *ProblemHandler) ListKnownErrors(w http.ResponseWriter, r *http.Request) {
	wsID, _ := strconv.ParseInt(r.URL.Query().Get("workspace_id"), 10, 64)
	if wsID <= 0 {
		writeProblemError(w, http.StatusBadRequest, "workspace_id query param is required")
		return
	}

	filter := model.KnownErrorFilter{
		WorkspaceID: wsID,
		Category:    r.URL.Query().Get("category"),
		Search:      r.URL.Query().Get("search"),
	}

	if limit, err := strconv.Atoi(r.URL.Query().Get("limit")); err == nil {
		filter.Limit = limit
	}
	if offset, err := strconv.Atoi(r.URL.Query().Get("offset")); err == nil {
		filter.Offset = offset
	}

	errorsList, total, err := h.svc.ListKnownErrors(filter)
	if err != nil {
		handleProblemServiceError(w, err)
		return
	}

	writeProblemJSON(w, http.StatusOK, map[string]interface{}{
		"known_errors": errorsList,
		"total":        total,
		"limit":        filter.Limit,
		"offset":       filter.Offset,
	})
}

// Metrics
func (h *ProblemHandler) GetProblemMetrics(w http.ResponseWriter, r *http.Request) {
	wsID, _ := strconv.ParseInt(r.URL.Query().Get("workspace_id"), 10, 64)
	if wsID <= 0 {
		writeProblemError(w, http.StatusBadRequest, "workspace_id query param is required")
		return
	}

	metrics, err := h.svc.GetProblemMetrics(wsID)
	if err != nil {
		handleProblemServiceError(w, err)
		return
	}

	writeProblemJSON(w, http.StatusOK, metrics)
}

// ServeHTTP routes HTTP requests for /v2/problems and /v2/kedb
func (h *ProblemHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	// 1. Direct KEDB route: /v2/kedb
	if strings.HasPrefix(path, "/v2/kedb") {
		if r.Method == http.MethodGet {
			h.ListKnownErrors(w, r)
			return
		}
		writeProblemError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	// 2. /v2/problems routes
	path = strings.TrimPrefix(path, "/v2/problems")
	path = strings.TrimPrefix(path, "/")

	// GET /v2/problems/metrics?workspace_id=...
	if path == "metrics" {
		if r.Method == http.MethodGet {
			h.GetProblemMetrics(w, r)
			return
		}
		writeProblemError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	// Standalone RCA step operations: /v2/problems/rca-steps/{id}
	if strings.HasPrefix(path, "rca-steps/") {
		stepIDStr := strings.TrimPrefix(path, "rca-steps/")
		stepID, err := strconv.ParseInt(stepIDStr, 10, 64)
		if err != nil || stepID <= 0 {
			writeProblemError(w, http.StatusBadRequest, "invalid rca step id")
			return
		}
		switch r.Method {
		case http.MethodPut, http.MethodPatch:
			h.UpdateRCAStep(w, r, stepID)
		case http.MethodDelete:
			h.DeleteRCAStep(w, r, stepID)
		default:
			writeProblemError(w, http.StatusMethodNotAllowed, "method not allowed")
		}
		return
	}

	// Root: /v2/problems
	if path == "" {
		switch r.Method {
		case http.MethodGet:
			h.ListProblems(w, r)
		case http.MethodPost:
			h.CreateProblem(w, r)
		default:
			writeProblemError(w, http.StatusMethodNotAllowed, "method not allowed")
		}
		return
	}

	parts := strings.Split(path, "/")
	problemID, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil || problemID <= 0 {
		writeProblemError(w, http.StatusBadRequest, "invalid problem id")
		return
	}

	// /v2/problems/{id}
	if len(parts) == 1 {
		switch r.Method {
		case http.MethodGet:
			h.GetProblem(w, r, problemID)
		case http.MethodPut, http.MethodPatch:
			h.UpdateProblem(w, r, problemID)
		case http.MethodDelete:
			h.DeleteProblem(w, r, problemID)
		default:
			writeProblemError(w, http.StatusMethodNotAllowed, "method not allowed")
		}
		return
	}

	// Sub-resource endpoints
	switch parts[1] {
	case "status":
		if r.Method == http.MethodPost || r.Method == http.MethodPatch {
			h.UpdateProblemStatus(w, r, problemID)
			return
		}
		writeProblemError(w, http.StatusMethodNotAllowed, "method not allowed")

	case "incidents":
		if len(parts) == 2 {
			switch r.Method {
			case http.MethodGet:
				h.ListProblemIncidents(w, r, problemID)
			case http.MethodPost:
				h.LinkIncident(w, r, problemID)
			default:
				writeProblemError(w, http.StatusMethodNotAllowed, "method not allowed")
			}
			return
		}
		if len(parts) == 3 && r.Method == http.MethodDelete {
			incidentID, err := strconv.ParseInt(parts[2], 10, 64)
			if err != nil || incidentID <= 0 {
				writeProblemError(w, http.StatusBadRequest, "invalid incident id")
				return
			}
			h.UnlinkIncident(w, r, problemID, incidentID)
			return
		}
		writeProblemError(w, http.StatusMethodNotAllowed, "method not allowed")

	case "cis":
		if len(parts) == 2 {
			switch r.Method {
			case http.MethodGet:
				h.ListProblemCIs(w, r, problemID)
			case http.MethodPost:
				h.LinkCI(w, r, problemID)
			default:
				writeProblemError(w, http.StatusMethodNotAllowed, "method not allowed")
			}
			return
		}
		if len(parts) == 3 && r.Method == http.MethodDelete {
			linkID, err := strconv.ParseInt(parts[2], 10, 64)
			if err != nil || linkID <= 0 {
				writeProblemError(w, http.StatusBadRequest, "invalid ci link id")
				return
			}
			h.UnlinkCI(w, r, problemID, linkID)
			return
		}
		writeProblemError(w, http.StatusMethodNotAllowed, "method not allowed")

	case "changes":
		if len(parts) == 2 {
			switch r.Method {
			case http.MethodGet:
				h.ListProblemChanges(w, r, problemID)
			case http.MethodPost:
				h.LinkChange(w, r, problemID)
			default:
				writeProblemError(w, http.StatusMethodNotAllowed, "method not allowed")
			}
			return
		}
		if len(parts) == 3 && r.Method == http.MethodDelete {
			linkID, err := strconv.ParseInt(parts[2], 10, 64)
			if err != nil || linkID <= 0 {
				writeProblemError(w, http.StatusBadRequest, "invalid change link id")
				return
			}
			h.UnlinkChange(w, r, problemID, linkID)
			return
		}
		writeProblemError(w, http.StatusMethodNotAllowed, "method not allowed")

	case "tasks":
		if len(parts) == 2 {
			switch r.Method {
			case http.MethodGet:
				h.ListProblemTasks(w, r, problemID)
			case http.MethodPost:
				h.LinkTask(w, r, problemID)
			default:
				writeProblemError(w, http.StatusMethodNotAllowed, "method not allowed")
			}
			return
		}
		if len(parts) == 3 && r.Method == http.MethodDelete {
			linkID, err := strconv.ParseInt(parts[2], 10, 64)
			if err != nil || linkID <= 0 {
				writeProblemError(w, http.StatusBadRequest, "invalid task link id")
				return
			}
			h.UnlinkTask(w, r, problemID, linkID)
			return
		}
		writeProblemError(w, http.StatusMethodNotAllowed, "method not allowed")

	case "rca":
		if len(parts) == 2 {
			switch r.Method {
			case http.MethodGet:
				h.ListRCASteps(w, r, problemID)
			case http.MethodPost:
				h.AddRCAStep(w, r, problemID)
			default:
				writeProblemError(w, http.StatusMethodNotAllowed, "method not allowed")
			}
			return
		}
		writeProblemError(w, http.StatusMethodNotAllowed, "method not allowed")

	case "known-error":
		switch r.Method {
		case http.MethodPost, http.MethodPatch:
			h.PublishKnownError(w, r, problemID)
		case http.MethodDelete:
			h.RetireKnownError(w, r, problemID)
		default:
			writeProblemError(w, http.StatusMethodNotAllowed, "method not allowed")
		}
		return

	default:
		writeProblemError(w, http.StatusNotFound, "endpoint not found")
	}
}
