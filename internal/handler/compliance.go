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

// ComplianceHandler exposes HTTP endpoints for ITIL GRC & Compliance Management
type ComplianceHandler struct {
	svc service.ComplianceService
}

// GlobalComplianceHandler is the singleton handler instance
var GlobalComplianceHandler *ComplianceHandler

// InitComplianceHandler initializes the global ComplianceHandler singleton
func InitComplianceHandler(svc service.ComplianceService) {
	GlobalComplianceHandler = &ComplianceHandler{svc: svc}
}

// NewComplianceHandler creates a new ComplianceHandler instance
func NewComplianceHandler(svc service.ComplianceService) *ComplianceHandler {
	return &ComplianceHandler{svc: svc}
}

func getGRCWorkspaceID(r *http.Request) string {
	if ws := r.URL.Query().Get("workspace_id"); ws != "" {
		return ws
	}
	if ws := r.Header.Get("X-Workspace-ID"); ws != "" {
		return ws
	}
	return "ws_1"
}

func getGRCUser(r *http.Request) string {
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

func writeGRCJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func writeGRCError(w http.ResponseWriter, status int, msg string) {
	writeGRCJSON(w, status, map[string]string{"error": msg})
}

func handleGRCError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrGRCFrameworkNotFound):
		writeGRCError(w, http.StatusNotFound, "compliance framework not found")
	case errors.Is(err, service.ErrGRCFrameworkCodeConflict):
		writeGRCError(w, http.StatusConflict, "compliance framework with this code already exists in workspace")
	case errors.Is(err, service.ErrGRCControlNotFound):
		writeGRCError(w, http.StatusNotFound, "compliance control not found")
	case errors.Is(err, service.ErrGRCControlCodeConflict):
		writeGRCError(w, http.StatusConflict, "compliance control with this code already exists in workspace")
	case errors.Is(err, service.ErrGRCEvidenceNotFound):
		writeGRCError(w, http.StatusNotFound, "compliance evidence not found")
	case errors.Is(err, service.ErrGRCAssessmentNotFound):
		writeGRCError(w, http.StatusNotFound, "compliance assessment not found")
	case errors.Is(err, service.ErrGRCFindingNotFound):
		writeGRCError(w, http.StatusNotFound, "compliance finding not found")
	default:
		writeGRCError(w, http.StatusBadRequest, err.Error())
	}
}

// ListFrameworks handles GET /v2/compliance/frameworks
func (h *ComplianceHandler) ListFrameworks(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	limit, _ := strconv.Atoi(q.Get("limit"))
	offset, _ := strconv.Atoi(q.Get("offset"))

	filter := model.GRCFrameworkFilter{
		WorkspaceID:  getGRCWorkspaceID(r),
		StandardType: q.Get("standard_type"),
		Status:       q.Get("status"),
		Search:       q.Get("search"),
		Limit:        limit,
		Offset:       offset,
	}

	frameworks, total, err := h.svc.ListFrameworks(r.Context(), filter)
	if err != nil {
		handleGRCError(w, err)
		return
	}

	writeGRCJSON(w, http.StatusOK, map[string]interface{}{
		"frameworks": frameworks,
		"total":      total,
		"limit":      filter.Limit,
		"offset":     filter.Offset,
	})
}

// CreateFramework handles POST /v2/compliance/frameworks
func (h *ComplianceHandler) CreateFramework(w http.ResponseWriter, r *http.Request) {
	var req model.CreateGRCFrameworkRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeGRCError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.OwnerID == "" {
		req.OwnerID = getGRCUser(r)
	}

	fw, err := h.svc.CreateFramework(r.Context(), getGRCWorkspaceID(r), req)
	if err != nil {
		handleGRCError(w, err)
		return
	}
	writeGRCJSON(w, http.StatusCreated, fw)
}

// GetFrameworkByID handles GET /v2/compliance/frameworks/{id}
func (h *ComplianceHandler) GetFrameworkByID(w http.ResponseWriter, r *http.Request, id string) {
	fw, err := h.svc.GetFrameworkByID(r.Context(), id)
	if err != nil {
		handleGRCError(w, err)
		return
	}
	writeGRCJSON(w, http.StatusOK, fw)
}

// UpdateFramework handles PUT /v2/compliance/frameworks/{id}
func (h *ComplianceHandler) UpdateFramework(w http.ResponseWriter, r *http.Request, id string) {
	var req model.UpdateGRCFrameworkRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeGRCError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	fw, err := h.svc.UpdateFramework(r.Context(), id, req)
	if err != nil {
		handleGRCError(w, err)
		return
	}
	writeGRCJSON(w, http.StatusOK, fw)
}

// DeleteFramework handles DELETE /v2/compliance/frameworks/{id}
func (h *ComplianceHandler) DeleteFramework(w http.ResponseWriter, r *http.Request, id string) {
	if err := h.svc.DeleteFramework(r.Context(), id); err != nil {
		handleGRCError(w, err)
		return
	}
	writeGRCJSON(w, http.StatusOK, map[string]string{"message": "compliance framework deleted"})
}

// Controls stubs
func (h *ComplianceHandler) ListControls(w http.ResponseWriter, r *http.Request) {}
func (h *ComplianceHandler) CreateControl(w http.ResponseWriter, r *http.Request) {}
func (h *ComplianceHandler) GetControlByID(w http.ResponseWriter, r *http.Request, id string) {}
func (h *ComplianceHandler) UpdateControl(w http.ResponseWriter, r *http.Request, id string) {}
func (h *ComplianceHandler) DeleteControl(w http.ResponseWriter, r *http.Request, id string) {}

// Evidence stubs
func (h *ComplianceHandler) ListEvidence(w http.ResponseWriter, r *http.Request) {}
func (h *ComplianceHandler) SubmitEvidence(w http.ResponseWriter, r *http.Request) {}
func (h *ComplianceHandler) GetEvidenceByID(w http.ResponseWriter, r *http.Request, id string) {}
func (h *ComplianceHandler) ReviewEvidence(w http.ResponseWriter, r *http.Request, id string) {}
func (h *ComplianceHandler) DeleteEvidence(w http.ResponseWriter, r *http.Request, id string) {}

// Assessments & Findings stubs
func (h *ComplianceHandler) ListAssessments(w http.ResponseWriter, r *http.Request) {}
func (h *ComplianceHandler) CreateAssessment(w http.ResponseWriter, r *http.Request) {}
func (h *ComplianceHandler) GetAssessmentByID(w http.ResponseWriter, r *http.Request, id string) {}
func (h *ComplianceHandler) CompleteAssessment(w http.ResponseWriter, r *http.Request, id string) {}
func (h *ComplianceHandler) DeleteAssessment(w http.ResponseWriter, r *http.Request, id string) {}
func (h *ComplianceHandler) ListFindings(w http.ResponseWriter, r *http.Request) {}
func (h *ComplianceHandler) CreateFinding(w http.ResponseWriter, r *http.Request) {}
func (h *ComplianceHandler) GetFindingByID(w http.ResponseWriter, r *http.Request, id string) {}
func (h *ComplianceHandler) UpdateFinding(w http.ResponseWriter, r *http.Request, id string) {}
func (h *ComplianceHandler) DeleteFinding(w http.ResponseWriter, r *http.Request, id string) {}

// Analytics & ServeHTTP stubs
func (h *ComplianceHandler) GetComplianceAnalytics(w http.ResponseWriter, r *http.Request) {}
func (h *ComplianceHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	writeGRCError(w, http.StatusNotFound, "route not found")
}
