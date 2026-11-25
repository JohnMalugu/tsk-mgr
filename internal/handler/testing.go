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

// TestingHandler exposes HTTP endpoints for ITIL Service Validation & Testing
type TestingHandler struct {
	svc service.TestingService
}

// GlobalTestingHandler is the singleton handler instance
var GlobalTestingHandler *TestingHandler

// InitTestingHandler initializes the global TestingHandler singleton
func InitTestingHandler(svc service.TestingService) {
	GlobalTestingHandler = &TestingHandler{svc: svc}
}

// NewTestingHandler creates a new TestingHandler instance
func NewTestingHandler(svc service.TestingService) *TestingHandler {
	return &TestingHandler{svc: svc}
}

func getTestingWorkspaceID(r *http.Request) string {
	if ws := r.URL.Query().Get("workspace_id"); ws != "" {
		return ws
	}
	if ws := r.Header.Get("X-Workspace-ID"); ws != "" {
		return ws
	}
	return "ws_1"
}

func getTestingUser(r *http.Request) string {
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

func writeTestingJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func writeTestingError(w http.ResponseWriter, status int, msg string) {
	writeTestingJSON(w, status, map[string]string{"error": msg})
}

func handleTestingError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrTestPlanNotFound):
		writeTestingError(w, http.StatusNotFound, "test plan not found")
	case errors.Is(err, service.ErrTestPlanCodeConflict):
		writeTestingError(w, http.StatusConflict, "test plan with this code already exists in workspace")
	case errors.Is(err, service.ErrTestCaseNotFound):
		writeTestingError(w, http.StatusNotFound, "test case not found")
	case errors.Is(err, service.ErrTestCaseCodeConflict):
		writeTestingError(w, http.StatusConflict, "test case with this code already exists in workspace")
	case errors.Is(err, service.ErrTestRunNotFound):
		writeTestingError(w, http.StatusNotFound, "test run not found")
	case errors.Is(err, service.ErrQualityGateNotFound):
		writeTestingError(w, http.StatusNotFound, "quality gate not found")
	case errors.Is(err, service.ErrQualityGateNameConflict):
		writeTestingError(w, http.StatusConflict, "quality gate with this name already exists in workspace")
	case errors.Is(err, service.ErrDefectLinkNotFound):
		writeTestingError(w, http.StatusNotFound, "defect link not found")
	default:
		writeTestingError(w, http.StatusBadRequest, err.Error())
	}
}

// ListPlans handles GET /v2/testing/plans
func (h *TestingHandler) ListPlans(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	limit, _ := strconv.Atoi(q.Get("limit"))
	offset, _ := strconv.Atoi(q.Get("offset"))

	filter := model.TestPlanFilter{
		WorkspaceID: getTestingWorkspaceID(r),
		ReleaseID:   q.Get("release_id"),
		ChangeID:    q.Get("change_id"),
		TestType:    q.Get("test_type"),
		Environment: q.Get("environment"),
		Status:      q.Get("status"),
		Search:      q.Get("search"),
		Limit:       limit,
		Offset:      offset,
	}

	plans, total, err := h.svc.ListPlans(r.Context(), filter)
	if err != nil {
		handleTestingError(w, err)
		return
	}

	writeTestingJSON(w, http.StatusOK, map[string]interface{}{
		"plans":  plans,
		"total":  total,
		"limit":  filter.Limit,
		"offset": filter.Offset,
	})
}

// CreatePlan handles POST /v2/testing/plans
func (h *TestingHandler) CreatePlan(w http.ResponseWriter, r *http.Request) {
	var req model.CreateTestPlanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeTestingError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	workspaceID := getTestingWorkspaceID(r)
	if req.CreatedBy == "" {
		req.CreatedBy = getTestingUser(r)
	}

	plan, err := h.svc.CreatePlan(r.Context(), workspaceID, req)
	if err != nil {
		handleTestingError(w, err)
		return
	}

	writeTestingJSON(w, http.StatusCreated, plan)
}

// GetPlanByID handles GET /v2/testing/plans/{id}
func (h *TestingHandler) GetPlanByID(w http.ResponseWriter, r *http.Request, id string) {
	plan, err := h.svc.GetPlanByID(r.Context(), id)
	if err != nil {
		handleTestingError(w, err)
		return
	}
	writeTestingJSON(w, http.StatusOK, plan)
}

// UpdatePlan handles PUT /v2/testing/plans/{id}
func (h *TestingHandler) UpdatePlan(w http.ResponseWriter, r *http.Request, id string) {
	var req model.UpdateTestPlanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeTestingError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	plan, err := h.svc.UpdatePlan(r.Context(), id, req)
	if err != nil {
		handleTestingError(w, err)
		return
	}

	writeTestingJSON(w, http.StatusOK, plan)
}

// DeletePlan handles DELETE /v2/testing/plans/{id}
func (h *TestingHandler) DeletePlan(w http.ResponseWriter, r *http.Request, id string) {
	if err := h.svc.DeletePlan(r.Context(), id); err != nil {
		handleTestingError(w, err)
		return
	}
	writeTestingJSON(w, http.StatusOK, map[string]string{"message": "test plan deleted"})
}

// ListCases handles GET /v2/testing/cases
func (h *TestingHandler) ListCases(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	limit, _ := strconv.Atoi(q.Get("limit"))
	offset, _ := strconv.Atoi(q.Get("offset"))

	filter := model.TestCaseFilter{
		WorkspaceID: getTestingWorkspaceID(r),
		PlanID:      q.Get("plan_id"),
		Category:    q.Get("category"),
		Severity:    q.Get("severity"),
		Status:      q.Get("status"),
		Search:      q.Get("search"),
		Limit:       limit,
		Offset:      offset,
	}

	cases, total, err := h.svc.ListCases(r.Context(), filter)
	if err != nil {
		handleTestingError(w, err)
		return
	}

	writeTestingJSON(w, http.StatusOK, map[string]interface{}{
		"cases":  cases,
		"total":  total,
		"limit":  filter.Limit,
		"offset": filter.Offset,
	})
}

// CreateCase handles POST /v2/testing/cases
func (h *TestingHandler) CreateCase(w http.ResponseWriter, r *http.Request) {
	var req model.CreateTestCaseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeTestingError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	workspaceID := getTestingWorkspaceID(r)
	testCase, err := h.svc.CreateCase(r.Context(), workspaceID, req)
	if err != nil {
		handleTestingError(w, err)
		return
	}

	writeTestingJSON(w, http.StatusCreated, testCase)
}

// GetCaseByID handles GET /v2/testing/cases/{id}
func (h *TestingHandler) GetCaseByID(w http.ResponseWriter, r *http.Request, id string) {
	testCase, err := h.svc.GetCaseByID(r.Context(), id)
	if err != nil {
		handleTestingError(w, err)
		return
	}
	writeTestingJSON(w, http.StatusOK, testCase)
}

// UpdateCase handles PUT /v2/testing/cases/{id}
func (h *TestingHandler) UpdateCase(w http.ResponseWriter, r *http.Request, id string) {
	var req model.UpdateTestCaseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeTestingError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	testCase, err := h.svc.UpdateCase(r.Context(), id, req)
	if err != nil {
		handleTestingError(w, err)
		return
	}

	writeTestingJSON(w, http.StatusOK, testCase)
}

// DeleteCase handles DELETE /v2/testing/cases/{id}
func (h *TestingHandler) DeleteCase(w http.ResponseWriter, r *http.Request, id string) {
	if err := h.svc.DeleteCase(r.Context(), id); err != nil {
		handleTestingError(w, err)
		return
	}
	writeTestingJSON(w, http.StatusOK, map[string]string{"message": "test case deleted"})
}

// ListRuns handles GET /v2/testing/runs
func (h *TestingHandler) ListRuns(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	limit, _ := strconv.Atoi(q.Get("limit"))
	offset, _ := strconv.Atoi(q.Get("offset"))

	filter := model.TestRunFilter{
		WorkspaceID: getTestingWorkspaceID(r),
		PlanID:      q.Get("plan_id"),
		CaseID:      q.Get("case_id"),
		Status:      q.Get("status"),
		ExecutedBy:  q.Get("executed_by"),
		Limit:       limit,
		Offset:      offset,
	}

	runs, total, err := h.svc.ListRuns(r.Context(), filter)
	if err != nil {
		handleTestingError(w, err)
		return
	}

	writeTestingJSON(w, http.StatusOK, map[string]interface{}{
		"runs":   runs,
		"total":  total,
		"limit":  filter.Limit,
		"offset": filter.Offset,
	})
}

// ExecuteRun handles POST /v2/testing/runs
func (h *TestingHandler) ExecuteRun(w http.ResponseWriter, r *http.Request) {
	var req model.ExecuteTestRunRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeTestingError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	workspaceID := getTestingWorkspaceID(r)
	if req.ExecutedBy == "" {
		req.ExecutedBy = getTestingUser(r)
	}

	run, err := h.svc.ExecuteRun(r.Context(), workspaceID, req)
	if err != nil {
		handleTestingError(w, err)
		return
	}

	writeTestingJSON(w, http.StatusCreated, run)
}

// GetRunByID handles GET /v2/testing/runs/{id}
func (h *TestingHandler) GetRunByID(w http.ResponseWriter, r *http.Request, id string) {
	run, err := h.svc.GetRunByID(r.Context(), id)
	if err != nil {
		handleTestingError(w, err)
		return
	}
	writeTestingJSON(w, http.StatusOK, run)
}

// ListQualityGates handles GET /v2/testing/gates
func (h *TestingHandler) ListQualityGates(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	limit, _ := strconv.Atoi(q.Get("limit"))
	offset, _ := strconv.Atoi(q.Get("offset"))

	filter := model.QualityGateFilter{
		WorkspaceID: getTestingWorkspaceID(r),
		ReleaseID:   q.Get("release_id"),
		ChangeID:    q.Get("change_id"),
		Status:      q.Get("status"),
		Limit:       limit,
		Offset:      offset,
	}

	gates, total, err := h.svc.ListQualityGates(r.Context(), filter)
	if err != nil {
		handleTestingError(w, err)
		return
	}

	writeTestingJSON(w, http.StatusOK, map[string]interface{}{
		"quality_gates": gates,
		"total":         total,
		"limit":         filter.Limit,
		"offset":        filter.Offset,
	})
}

// CreateQualityGate handles POST /v2/testing/gates
func (h *TestingHandler) CreateQualityGate(w http.ResponseWriter, r *http.Request) {
	var req model.CreateQualityGateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeTestingError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	workspaceID := getTestingWorkspaceID(r)
	gate, err := h.svc.CreateQualityGate(r.Context(), workspaceID, req)
	if err != nil {
		handleTestingError(w, err)
		return
	}

	writeTestingJSON(w, http.StatusCreated, gate)
}

// GetQualityGateByID handles GET /v2/testing/gates/{id}
func (h *TestingHandler) GetQualityGateByID(w http.ResponseWriter, r *http.Request, id string) {
	gate, err := h.svc.GetQualityGateByID(r.Context(), id)
	if err != nil {
		handleTestingError(w, err)
		return
	}
	writeTestingJSON(w, http.StatusOK, gate)
}

// EvaluateQualityGate handles POST /v2/testing/gates/{id}/evaluate
func (h *TestingHandler) EvaluateQualityGate(w http.ResponseWriter, r *http.Request, id string) {
	gate, err := h.svc.EvaluateQualityGate(r.Context(), id)
	if err != nil {
		handleTestingError(w, err)
		return
	}
	writeTestingJSON(w, http.StatusOK, gate)
}

// WaiveQualityGate handles POST /v2/testing/gates/{id}/waive
func (h *TestingHandler) WaiveQualityGate(w http.ResponseWriter, r *http.Request, id string) {
	var req model.WaiveQualityGateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeTestingError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.WaivedBy == "" {
		req.WaivedBy = getTestingUser(r)
	}

	gate, err := h.svc.WaiveQualityGate(r.Context(), id, req)
	if err != nil {
		handleTestingError(w, err)
		return
	}

	writeTestingJSON(w, http.StatusOK, gate)
}

// DeleteQualityGate handles DELETE /v2/testing/gates/{id}
func (h *TestingHandler) DeleteQualityGate(w http.ResponseWriter, r *http.Request, id string) {
	if err := h.svc.DeleteQualityGate(r.Context(), id); err != nil {
		handleTestingError(w, err)
		return
	}
	writeTestingJSON(w, http.StatusOK, map[string]string{"message": "quality gate deleted"})
}

// ListDefects handles GET /v2/testing/defects
func (h *TestingHandler) ListDefects(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	limit, _ := strconv.Atoi(q.Get("limit"))
	offset, _ := strconv.Atoi(q.Get("offset"))

	filter := model.DefectLinkFilter{
		WorkspaceID: getTestingWorkspaceID(r),
		RunID:       q.Get("run_id"),
		IncidentID:  q.Get("incident_id"),
		ProblemID:   q.Get("problem_id"),
		Status:      q.Get("status"),
		Limit:       limit,
		Offset:      offset,
	}

	defects, total, err := h.svc.ListDefects(r.Context(), filter)
	if err != nil {
		handleTestingError(w, err)
		return
	}

	writeTestingJSON(w, http.StatusOK, map[string]interface{}{
		"defects": defects,
		"total":   total,
		"limit":   filter.Limit,
		"offset":  filter.Offset,
	})
}

// LinkDefect handles POST /v2/testing/defects
func (h *TestingHandler) LinkDefect(w http.ResponseWriter, r *http.Request) {
	var req model.LinkDefectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeTestingError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	workspaceID := getTestingWorkspaceID(r)
	defect, err := h.svc.LinkDefect(r.Context(), workspaceID, req)
	if err != nil {
		handleTestingError(w, err)
		return
	}

	writeTestingJSON(w, http.StatusCreated, defect)
}

// GetTestingAnalytics handles GET /v2/testing/analytics
func (h *TestingHandler) GetTestingAnalytics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeTestingError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	workspaceID := getTestingWorkspaceID(r)
	analytics, err := h.svc.GetTestingAnalytics(r.Context(), workspaceID)
	if err != nil {
		handleTestingError(w, err)
		return
	}

	writeTestingJSON(w, http.StatusOK, analytics)
}

// ServeHTTP stub
func (h *TestingHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	writeTestingError(w, http.StatusNotFound, "route not found")
}
