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

// Steps HTTP Endpoints

// ListSteps handles GET /v2/deployments/steps
func (h *DeploymentHandler) ListSteps(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	limit, _ := strconv.Atoi(q.Get("limit"))
	offset, _ := strconv.Atoi(q.Get("offset"))

	filter := model.DeploymentStepFilter{
		WorkspaceID: getDeploymentWorkspaceID(r),
		RunID:       q.Get("run_id"),
		Status:      q.Get("status"),
		Limit:       limit,
		Offset:      offset,
	}

	steps, total, err := h.svc.ListSteps(r.Context(), filter)
	if err != nil {
		handleDeploymentError(w, err)
		return
	}

	writeDeploymentJSON(w, http.StatusOK, map[string]interface{}{
		"steps":  steps,
		"total":  total,
		"limit":  filter.Limit,
		"offset": filter.Offset,
	})
}

// CreateStep handles POST /v2/deployments/steps
func (h *DeploymentHandler) CreateStep(w http.ResponseWriter, r *http.Request) {
	var req model.CreateDeploymentStepRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeDeploymentError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	ws := getDeploymentWorkspaceID(r)
	if req.WorkspaceID == "" {
		req.WorkspaceID = ws
	}

	step, err := h.svc.CreateStep(r.Context(), ws, req)
	if err != nil {
		handleDeploymentError(w, err)
		return
	}

	writeDeploymentJSON(w, http.StatusCreated, step)
}

// GetStepByID handles GET /v2/deployments/steps/{id}
func (h *DeploymentHandler) GetStepByID(w http.ResponseWriter, r *http.Request, id string) {
	step, err := h.svc.GetStepByID(r.Context(), id)
	if err != nil {
		handleDeploymentError(w, err)
		return
	}
	writeDeploymentJSON(w, http.StatusOK, step)
}

// PromoteStep handles POST /v2/deployments/steps/{id}/promote
func (h *DeploymentHandler) PromoteStep(w http.ResponseWriter, r *http.Request, id string) {
	var req model.PromoteDeploymentStepRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeDeploymentError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	step, err := h.svc.PromoteStep(r.Context(), id, req)
	if err != nil {
		handleDeploymentError(w, err)
		return
	}
	writeDeploymentJSON(w, http.StatusOK, step)
}

// Rollback Policies HTTP Endpoints

// ListRollbackPolicies handles GET /v2/deployments/rollbacks
func (h *DeploymentHandler) ListRollbackPolicies(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	limit, _ := strconv.Atoi(q.Get("limit"))
	offset, _ := strconv.Atoi(q.Get("offset"))

	filter := model.RollbackPolicyFilter{
		WorkspaceID:     getDeploymentWorkspaceID(r),
		PipelineID:      q.Get("pipeline_id"),
		ExecutionStatus: q.Get("execution_status"),
		Limit:           limit,
		Offset:          offset,
	}

	policies, total, err := h.svc.ListRollbackPolicies(r.Context(), filter)
	if err != nil {
		handleDeploymentError(w, err)
		return
	}

	writeDeploymentJSON(w, http.StatusOK, map[string]interface{}{
		"rollbacks": policies,
		"total":     total,
		"limit":     filter.Limit,
		"offset":    filter.Offset,
	})
}

// CreateRollbackPolicy handles POST /v2/deployments/rollbacks
func (h *DeploymentHandler) CreateRollbackPolicy(w http.ResponseWriter, r *http.Request) {
	var req model.CreateRollbackPolicyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeDeploymentError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	ws := getDeploymentWorkspaceID(r)
	if req.WorkspaceID == "" {
		req.WorkspaceID = ws
	}

	policy, err := h.svc.CreateRollbackPolicy(r.Context(), ws, req)
	if err != nil {
		handleDeploymentError(w, err)
		return
	}

	writeDeploymentJSON(w, http.StatusCreated, policy)
}

// GetRollbackPolicyByID handles GET /v2/deployments/rollbacks/{id}
func (h *DeploymentHandler) GetRollbackPolicyByID(w http.ResponseWriter, r *http.Request, id string) {
	policy, err := h.svc.GetRollbackPolicyByID(r.Context(), id)
	if err != nil {
		handleDeploymentError(w, err)
		return
	}
	writeDeploymentJSON(w, http.StatusOK, policy)
}

// ExecuteRollback handles POST /v2/deployments/rollbacks/{id}/execute
func (h *DeploymentHandler) ExecuteRollback(w http.ResponseWriter, r *http.Request, id string) {
	var req model.ExecuteRollbackRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeDeploymentError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	policy, err := h.svc.ExecuteRollback(r.Context(), id, req)
	if err != nil {
		handleDeploymentError(w, err)
		return
	}
	writeDeploymentJSON(w, http.StatusOK, policy)
}

// Analytics and Router HTTP Endpoints

// GetDeploymentAnalytics handles GET /v2/deployments/analytics
func (h *DeploymentHandler) GetDeploymentAnalytics(w http.ResponseWriter, r *http.Request) {
	ws := getDeploymentWorkspaceID(r)
	analytics, err := h.svc.GetDeploymentAnalytics(r.Context(), ws)
	if err != nil {
		handleDeploymentError(w, err)
		return
	}
	writeDeploymentJSON(w, http.StatusOK, analytics)
}

// ServeHTTP routes HTTP requests under /v2/deployments
func (h *DeploymentHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/v2/deployments")
	path = strings.TrimPrefix(path, "/")

	if path == "" {
		writeDeploymentJSON(w, http.StatusOK, map[string]string{
			"service": "ITIL Release & Deployment Management API",
			"version": "v2",
		})
		return
	}

	if path == "analytics" {
		if r.Method == http.MethodGet {
			h.GetDeploymentAnalytics(w, r)
		} else {
			writeDeploymentError(w, http.StatusMethodNotAllowed, "method not allowed")
		}
		return
	}

	parts := strings.Split(path, "/")
	resource := parts[0]

	switch resource {
	case "pipelines":
		if len(parts) == 1 {
			switch r.Method {
			case http.MethodGet:
				h.ListPipelines(w, r)
			case http.MethodPost:
				h.CreatePipeline(w, r)
			default:
				writeDeploymentError(w, http.StatusMethodNotAllowed, "method not allowed")
			}
			return
		}
		pipelineID := parts[1]
		if len(parts) == 2 {
			switch r.Method {
			case http.MethodGet:
				h.GetPipelineByID(w, r, pipelineID)
			case http.MethodPut:
				h.UpdatePipeline(w, r, pipelineID)
			case http.MethodDelete:
				h.DeletePipeline(w, r, pipelineID)
			default:
				writeDeploymentError(w, http.StatusMethodNotAllowed, "method not allowed")
			}
			return
		}

	case "runs":
		if len(parts) == 1 {
			switch r.Method {
			case http.MethodGet:
				h.ListRuns(w, r)
			case http.MethodPost:
				h.CreateRun(w, r)
			default:
				writeDeploymentError(w, http.StatusMethodNotAllowed, "method not allowed")
			}
			return
		}
		runID := parts[1]
		if len(parts) == 2 {
			if r.Method == http.MethodGet {
				h.GetRunByID(w, r, runID)
			} else {
				writeDeploymentError(w, http.StatusMethodNotAllowed, "method not allowed")
			}
			return
		}
		if len(parts) == 3 && parts[2] == "status" {
			if r.Method == http.MethodPut || r.Method == http.MethodPatch {
				h.UpdateRunStatus(w, r, runID)
			} else {
				writeDeploymentError(w, http.StatusMethodNotAllowed, "method not allowed")
			}
			return
		}

	case "gates":
		if len(parts) == 1 {
			switch r.Method {
			case http.MethodGet:
				h.ListGates(w, r)
			case http.MethodPost:
				h.CreateGate(w, r)
			default:
				writeDeploymentError(w, http.StatusMethodNotAllowed, "method not allowed")
			}
			return
		}
		gateID := parts[1]
		if len(parts) == 2 {
			if r.Method == http.MethodGet {
				h.GetGateByID(w, r, gateID)
			} else {
				writeDeploymentError(w, http.StatusMethodNotAllowed, "method not allowed")
			}
			return
		}
		if len(parts) == 3 && parts[2] == "evaluate" {
			if r.Method == http.MethodPost {
				h.EvaluateGate(w, r, gateID)
			} else {
				writeDeploymentError(w, http.StatusMethodNotAllowed, "method not allowed")
			}
			return
		}

	case "steps":
		if len(parts) == 1 {
			switch r.Method {
			case http.MethodGet:
				h.ListSteps(w, r)
			case http.MethodPost:
				h.CreateStep(w, r)
			default:
				writeDeploymentError(w, http.StatusMethodNotAllowed, "method not allowed")
			}
			return
		}
		stepID := parts[1]
		if len(parts) == 2 {
			if r.Method == http.MethodGet {
				h.GetStepByID(w, r, stepID)
			} else {
				writeDeploymentError(w, http.StatusMethodNotAllowed, "method not allowed")
			}
			return
		}
		if len(parts) == 3 && parts[2] == "promote" {
			if r.Method == http.MethodPost {
				h.PromoteStep(w, r, stepID)
			} else {
				writeDeploymentError(w, http.StatusMethodNotAllowed, "method not allowed")
			}
			return
		}

	case "rollbacks":
		if len(parts) == 1 {
			switch r.Method {
			case http.MethodGet:
				h.ListRollbackPolicies(w, r)
			case http.MethodPost:
				h.CreateRollbackPolicy(w, r)
			default:
				writeDeploymentError(w, http.StatusMethodNotAllowed, "method not allowed")
			}
			return
		}
		rollbackID := parts[1]
		if len(parts) == 2 {
			if r.Method == http.MethodGet {
				h.GetRollbackPolicyByID(w, r, rollbackID)
			} else {
				writeDeploymentError(w, http.StatusMethodNotAllowed, "method not allowed")
			}
			return
		}
		if len(parts) == 3 && parts[2] == "execute" {
			if r.Method == http.MethodPost {
				h.ExecuteRollback(w, r, rollbackID)
			} else {
				writeDeploymentError(w, http.StatusMethodNotAllowed, "method not allowed")
			}
			return
		}
	}

	writeDeploymentError(w, http.StatusNotFound, "route not found")
}
