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

var (
	_ = strconv.Atoi
	_ = strings.TrimPrefix
)

// DeploymentHandler exposes HTTP endpoints for Release & Deployment Management
type DeploymentHandler struct {
	svc service.DeploymentService
}

// GlobalDeploymentHandler is the singleton handler instance
var GlobalDeploymentHandler *DeploymentHandler

// InitDeploymentHandler initializes the global Deployment handler singleton
func InitDeploymentHandler(svc service.DeploymentService) {
	GlobalDeploymentHandler = &DeploymentHandler{svc: svc}
}

// NewDeploymentHandler creates a new DeploymentHandler instance
func NewDeploymentHandler(svc service.DeploymentService) *DeploymentHandler {
	return &DeploymentHandler{svc: svc}
}

func getDeploymentUser(r *http.Request) string {
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

func getDeploymentWorkspaceID(r *http.Request) string {
	if ws := r.URL.Query().Get("workspace_id"); ws != "" {
		return ws
	}
	if ws := r.Header.Get("X-Workspace-ID"); ws != "" {
		return ws
	}
	return "ws_1"
}

func writeDeploymentJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func writeDeploymentError(w http.ResponseWriter, status int, msg string) {
	writeDeploymentJSON(w, status, map[string]string{"error": msg})
}

func handleDeploymentError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrDeploymentPipelineNotFound):
		writeDeploymentError(w, http.StatusNotFound, "deployment pipeline not found")
	case errors.Is(err, service.ErrDeploymentPipelineCodeConflict):
		writeDeploymentError(w, http.StatusConflict, "deployment pipeline with this code already exists in workspace")
	case errors.Is(err, service.ErrDeploymentRunNotFound):
		writeDeploymentError(w, http.StatusNotFound, "deployment run not found")
	case errors.Is(err, service.ErrDeploymentGateNotFound):
		writeDeploymentError(w, http.StatusNotFound, "deployment gate not found")
	case errors.Is(err, service.ErrDeploymentStepNotFound):
		writeDeploymentError(w, http.StatusNotFound, "deployment step not found")
	case errors.Is(err, service.ErrRollbackPolicyNotFound):
		writeDeploymentError(w, http.StatusNotFound, "rollback policy not found")
	default:
		writeDeploymentError(w, http.StatusBadRequest, err.Error())
	}
}

// Pipelines HTTP Endpoints

// ListPipelines handles GET /v2/deployments/pipelines
func (h *DeploymentHandler) ListPipelines(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	limit, _ := strconv.Atoi(q.Get("limit"))
	offset, _ := strconv.Atoi(q.Get("offset"))

	filter := model.DeploymentPipelineFilter{
		WorkspaceID:        getDeploymentWorkspaceID(r),
		TargetEnvironment:  q.Get("target_environment"),
		DeploymentStrategy: q.Get("deployment_strategy"),
		Status:             q.Get("status"),
		Limit:              limit,
		Offset:             offset,
	}

	pipelines, total, err := h.svc.ListPipelines(r.Context(), filter)
	if err != nil {
		handleDeploymentError(w, err)
		return
	}

	writeDeploymentJSON(w, http.StatusOK, map[string]interface{}{
		"pipelines": pipelines,
		"total":     total,
		"limit":     filter.Limit,
		"offset":    filter.Offset,
	})
}

// CreatePipeline handles POST /v2/deployments/pipelines
func (h *DeploymentHandler) CreatePipeline(w http.ResponseWriter, r *http.Request) {
	var req model.CreateDeploymentPipelineRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeDeploymentError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	ws := getDeploymentWorkspaceID(r)
	if req.WorkspaceID == "" {
		req.WorkspaceID = ws
	}
	if req.CreatedByUserID == "" {
		req.CreatedByUserID = getDeploymentUser(r)
	}

	pipeline, err := h.svc.CreatePipeline(r.Context(), ws, req)
	if err != nil {
		handleDeploymentError(w, err)
		return
	}

	writeDeploymentJSON(w, http.StatusCreated, pipeline)
}

// GetPipelineByID handles GET /v2/deployments/pipelines/{id}
func (h *DeploymentHandler) GetPipelineByID(w http.ResponseWriter, r *http.Request, id string) {
	pipeline, err := h.svc.GetPipelineByID(r.Context(), id)
	if err != nil {
		handleDeploymentError(w, err)
		return
	}
	writeDeploymentJSON(w, http.StatusOK, pipeline)
}

// UpdatePipeline handles PUT /v2/deployments/pipelines/{id}
func (h *DeploymentHandler) UpdatePipeline(w http.ResponseWriter, r *http.Request, id string) {
	var req model.UpdateDeploymentPipelineRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeDeploymentError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	pipeline, err := h.svc.UpdatePipeline(r.Context(), id, req)
	if err != nil {
		handleDeploymentError(w, err)
		return
	}
	writeDeploymentJSON(w, http.StatusOK, pipeline)
}

// DeletePipeline handles DELETE /v2/deployments/pipelines/{id}
func (h *DeploymentHandler) DeletePipeline(w http.ResponseWriter, r *http.Request, id string) {
	if err := h.svc.DeletePipeline(r.Context(), id); err != nil {
		handleDeploymentError(w, err)
		return
	}
	writeDeploymentJSON(w, http.StatusOK, map[string]string{"message": "deployment pipeline deleted successfully"})
}

// Runs HTTP Endpoints

// ListRuns handles GET /v2/deployments/runs
func (h *DeploymentHandler) ListRuns(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	limit, _ := strconv.Atoi(q.Get("limit"))
	offset, _ := strconv.Atoi(q.Get("offset"))

	filter := model.DeploymentRunFilter{
		WorkspaceID: getDeploymentWorkspaceID(r),
		PipelineID:  q.Get("pipeline_id"),
		Environment: q.Get("environment"),
		Status:      q.Get("status"),
		TriggerType: q.Get("trigger_type"),
		Limit:       limit,
		Offset:      offset,
	}

	runs, total, err := h.svc.ListRuns(r.Context(), filter)
	if err != nil {
		handleDeploymentError(w, err)
		return
	}

	writeDeploymentJSON(w, http.StatusOK, map[string]interface{}{
		"runs":   runs,
		"total":  total,
		"limit":  filter.Limit,
		"offset": filter.Offset,
	})
}

// CreateRun handles POST /v2/deployments/runs
func (h *DeploymentHandler) CreateRun(w http.ResponseWriter, r *http.Request) {
	var req model.CreateDeploymentRunRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeDeploymentError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	ws := getDeploymentWorkspaceID(r)
	if req.WorkspaceID == "" {
		req.WorkspaceID = ws
	}
	if req.DeployedByUserID == "" {
		req.DeployedByUserID = getDeploymentUser(r)
	}

	run, err := h.svc.CreateRun(r.Context(), ws, req)
	if err != nil {
		handleDeploymentError(w, err)
		return
	}

	writeDeploymentJSON(w, http.StatusCreated, run)
}

// GetRunByID handles GET /v2/deployments/runs/{id}
func (h *DeploymentHandler) GetRunByID(w http.ResponseWriter, r *http.Request, id string) {
	run, err := h.svc.GetRunByID(r.Context(), id)
	if err != nil {
		handleDeploymentError(w, err)
		return
	}
	writeDeploymentJSON(w, http.StatusOK, run)
}

// UpdateRunStatus handles PUT/PATCH /v2/deployments/runs/{id}/status
func (h *DeploymentHandler) UpdateRunStatus(w http.ResponseWriter, r *http.Request, id string) {
	var req model.UpdateDeploymentRunStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeDeploymentError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	run, err := h.svc.UpdateRunStatus(r.Context(), id, req)
	if err != nil {
		handleDeploymentError(w, err)
		return
	}
	writeDeploymentJSON(w, http.StatusOK, run)
}

// Gates HTTP Endpoints

// ListGates handles GET /v2/deployments/gates
func (h *DeploymentHandler) ListGates(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	limit, _ := strconv.Atoi(q.Get("limit"))
	offset, _ := strconv.Atoi(q.Get("offset"))

	filter := model.DeploymentGateFilter{
		WorkspaceID: getDeploymentWorkspaceID(r),
		RunID:       q.Get("run_id"),
		GateType:    q.Get("gate_type"),
		Status:      q.Get("status"),
		Limit:       limit,
		Offset:      offset,
	}

	gates, total, err := h.svc.ListGates(r.Context(), filter)
	if err != nil {
		handleDeploymentError(w, err)
		return
	}

	writeDeploymentJSON(w, http.StatusOK, map[string]interface{}{
		"gates":  gates,
		"total":  total,
		"limit":  filter.Limit,
		"offset": filter.Offset,
	})
}

// CreateGate handles POST /v2/deployments/gates
func (h *DeploymentHandler) CreateGate(w http.ResponseWriter, r *http.Request) {
	var req model.CreateDeploymentGateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeDeploymentError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	ws := getDeploymentWorkspaceID(r)
	if req.WorkspaceID == "" {
		req.WorkspaceID = ws
	}

	gate, err := h.svc.CreateGate(r.Context(), ws, req)
	if err != nil {
		handleDeploymentError(w, err)
		return
	}

	writeDeploymentJSON(w, http.StatusCreated, gate)
}

// GetGateByID handles GET /v2/deployments/gates/{id}
func (h *DeploymentHandler) GetGateByID(w http.ResponseWriter, r *http.Request, id string) {
	gate, err := h.svc.GetGateByID(r.Context(), id)
	if err != nil {
		handleDeploymentError(w, err)
		return
	}
	writeDeploymentJSON(w, http.StatusOK, gate)
}

// EvaluateGate handles POST /v2/deployments/gates/{id}/evaluate
func (h *DeploymentHandler) EvaluateGate(w http.ResponseWriter, r *http.Request, id string) {
	var req model.EvaluateDeploymentGateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeDeploymentError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.EvaluatorID == "" {
		req.EvaluatorID = getDeploymentUser(r)
	}

	gate, err := h.svc.EvaluateGate(r.Context(), id, req)
	if err != nil {
		handleDeploymentError(w, err)
		return
	}
	writeDeploymentJSON(w, http.StatusOK, gate)
}
