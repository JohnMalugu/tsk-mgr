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

// Cases stubs
func (h *TestingHandler) ListCases(w http.ResponseWriter, r *http.Request) {}
func (h *TestingHandler) CreateCase(w http.ResponseWriter, r *http.Request) {}
func (h *TestingHandler) GetCaseByID(w http.ResponseWriter, r *http.Request, id string) {}
func (h *TestingHandler) UpdateCase(w http.ResponseWriter, r *http.Request, id string) {}
func (h *TestingHandler) DeleteCase(w http.ResponseWriter, r *http.Request, id string) {}

// Runs stubs
func (h *TestingHandler) ListRuns(w http.ResponseWriter, r *http.Request) {}
func (h *TestingHandler) ExecuteRun(w http.ResponseWriter, r *http.Request) {}
func (h *TestingHandler) GetRunByID(w http.ResponseWriter, r *http.Request, id string) {}

// Quality Gates stubs
func (h *TestingHandler) ListQualityGates(w http.ResponseWriter, r *http.Request) {}
func (h *TestingHandler) CreateQualityGate(w http.ResponseWriter, r *http.Request) {}
func (h *TestingHandler) GetQualityGateByID(w http.ResponseWriter, r *http.Request, id string) {}
func (h *TestingHandler) EvaluateQualityGate(w http.ResponseWriter, r *http.Request, id string) {}
func (h *TestingHandler) WaiveQualityGate(w http.ResponseWriter, r *http.Request, id string) {}
func (h *TestingHandler) DeleteQualityGate(w http.ResponseWriter, r *http.Request, id string) {}

// Defects stubs
func (h *TestingHandler) ListDefects(w http.ResponseWriter, r *http.Request) {}
func (h *TestingHandler) LinkDefect(w http.ResponseWriter, r *http.Request) {}

// Analytics stub
func (h *TestingHandler) GetTestingAnalytics(w http.ResponseWriter, r *http.Request) {}

// ServeHTTP stub
func (h *TestingHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	writeTestingError(w, http.StatusNotFound, "route not found")
}
