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

// ListControls handles GET /v2/compliance/controls
func (h *ComplianceHandler) ListControls(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	limit, _ := strconv.Atoi(q.Get("limit"))
	offset, _ := strconv.Atoi(q.Get("offset"))

	filter := model.GRCControlFilter{
		WorkspaceID:          getGRCWorkspaceID(r),
		FrameworkID:          q.Get("framework_id"),
		Domain:               q.Get("domain"),
		ImplementationStatus: q.Get("implementation_status"),
		AutomationLevel:      q.Get("automation_level"),
		OwnerID:              q.Get("owner_id"),
		Search:               q.Get("search"),
		Limit:                limit,
		Offset:               offset,
	}

	controls, total, err := h.svc.ListControls(r.Context(), filter)
	if err != nil {
		handleGRCError(w, err)
		return
	}

	writeGRCJSON(w, http.StatusOK, map[string]interface{}{
		"controls": controls,
		"total":    total,
		"limit":    filter.Limit,
		"offset":   filter.Offset,
	})
}

// CreateControl handles POST /v2/compliance/controls
func (h *ComplianceHandler) CreateControl(w http.ResponseWriter, r *http.Request) {
	var req model.CreateGRCControlRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeGRCError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.OwnerID == "" {
		req.OwnerID = getGRCUser(r)
	}

	ctrl, err := h.svc.CreateControl(r.Context(), getGRCWorkspaceID(r), req)
	if err != nil {
		handleGRCError(w, err)
		return
	}
	writeGRCJSON(w, http.StatusCreated, ctrl)
}

// GetControlByID handles GET /v2/compliance/controls/{id}
func (h *ComplianceHandler) GetControlByID(w http.ResponseWriter, r *http.Request, id string) {
	ctrl, err := h.svc.GetControlByID(r.Context(), id)
	if err != nil {
		handleGRCError(w, err)
		return
	}
	writeGRCJSON(w, http.StatusOK, ctrl)
}

// UpdateControl handles PUT /v2/compliance/controls/{id}
func (h *ComplianceHandler) UpdateControl(w http.ResponseWriter, r *http.Request, id string) {
	var req model.UpdateGRCControlRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeGRCError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	ctrl, err := h.svc.UpdateControl(r.Context(), id, req)
	if err != nil {
		handleGRCError(w, err)
		return
	}
	writeGRCJSON(w, http.StatusOK, ctrl)
}

// DeleteControl handles DELETE /v2/compliance/controls/{id}
func (h *ComplianceHandler) DeleteControl(w http.ResponseWriter, r *http.Request, id string) {
	if err := h.svc.DeleteControl(r.Context(), id); err != nil {
		handleGRCError(w, err)
		return
	}
	writeGRCJSON(w, http.StatusOK, map[string]string{"message": "compliance control deleted"})
}

// ListEvidence handles GET /v2/compliance/evidence
func (h *ComplianceHandler) ListEvidence(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	limit, _ := strconv.Atoi(q.Get("limit"))
	offset, _ := strconv.Atoi(q.Get("offset"))

	filter := model.GRCEvidenceFilter{
		WorkspaceID:  getGRCWorkspaceID(r),
		ControlID:    q.Get("control_id"),
		EvidenceType: q.Get("evidence_type"),
		ReviewStatus: q.Get("review_status"),
		Limit:        limit,
		Offset:       offset,
	}

	evidence, total, err := h.svc.ListEvidence(r.Context(), filter)
	if err != nil {
		handleGRCError(w, err)
		return
	}

	writeGRCJSON(w, http.StatusOK, map[string]interface{}{
		"evidence": evidence,
		"total":    total,
		"limit":    filter.Limit,
		"offset":   filter.Offset,
	})
}

// SubmitEvidence handles POST /v2/compliance/evidence
func (h *ComplianceHandler) SubmitEvidence(w http.ResponseWriter, r *http.Request) {
	var req model.SubmitGRCEvidenceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeGRCError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.CollectedBy == "" {
		req.CollectedBy = getGRCUser(r)
	}

	ev, err := h.svc.SubmitEvidence(r.Context(), getGRCWorkspaceID(r), req)
	if err != nil {
		handleGRCError(w, err)
		return
	}
	writeGRCJSON(w, http.StatusCreated, ev)
}

// GetEvidenceByID handles GET /v2/compliance/evidence/{id}
func (h *ComplianceHandler) GetEvidenceByID(w http.ResponseWriter, r *http.Request, id string) {
	ev, err := h.svc.GetEvidenceByID(r.Context(), id)
	if err != nil {
		handleGRCError(w, err)
		return
	}
	writeGRCJSON(w, http.StatusOK, ev)
}

// ReviewEvidence handles POST /v2/compliance/evidence/{id}/review
func (h *ComplianceHandler) ReviewEvidence(w http.ResponseWriter, r *http.Request, id string) {
	var req model.ReviewGRCEvidenceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeGRCError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.ReviewedBy == "" {
		req.ReviewedBy = getGRCUser(r)
	}

	ev, err := h.svc.ReviewEvidence(r.Context(), id, req)
	if err != nil {
		handleGRCError(w, err)
		return
	}
	writeGRCJSON(w, http.StatusOK, ev)
}

// DeleteEvidence handles DELETE /v2/compliance/evidence/{id}
func (h *ComplianceHandler) DeleteEvidence(w http.ResponseWriter, r *http.Request, id string) {
	if err := h.svc.DeleteEvidence(r.Context(), id); err != nil {
		handleGRCError(w, err)
		return
	}
	writeGRCJSON(w, http.StatusOK, map[string]string{"message": "compliance evidence deleted"})
}

// ListAssessments handles GET /v2/compliance/assessments
func (h *ComplianceHandler) ListAssessments(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	limit, _ := strconv.Atoi(q.Get("limit"))
	offset, _ := strconv.Atoi(q.Get("offset"))

	filter := model.GRCAssessmentFilter{
		WorkspaceID: getGRCWorkspaceID(r),
		FrameworkID: q.Get("framework_id"),
		AuditType:   q.Get("audit_type"),
		Status:      q.Get("status"),
		Limit:       limit,
		Offset:      offset,
	}

	assessments, total, err := h.svc.ListAssessments(r.Context(), filter)
	if err != nil {
		handleGRCError(w, err)
		return
	}

	writeGRCJSON(w, http.StatusOK, map[string]interface{}{
		"assessments": assessments,
		"total":       total,
		"limit":       filter.Limit,
		"offset":      filter.Offset,
	})
}

// CreateAssessment handles POST /v2/compliance/assessments
func (h *ComplianceHandler) CreateAssessment(w http.ResponseWriter, r *http.Request) {
	var req model.CreateGRCAssessmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeGRCError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	assessment, err := h.svc.CreateAssessment(r.Context(), getGRCWorkspaceID(r), req)
	if err != nil {
		handleGRCError(w, err)
		return
	}
	writeGRCJSON(w, http.StatusCreated, assessment)
}

// GetAssessmentByID handles GET /v2/compliance/assessments/{id}
func (h *ComplianceHandler) GetAssessmentByID(w http.ResponseWriter, r *http.Request, id string) {
	assessment, err := h.svc.GetAssessmentByID(r.Context(), id)
	if err != nil {
		handleGRCError(w, err)
		return
	}
	writeGRCJSON(w, http.StatusOK, assessment)
}

// CompleteAssessment handles POST /v2/compliance/assessments/{id}/complete
func (h *ComplianceHandler) CompleteAssessment(w http.ResponseWriter, r *http.Request, id string) {
	var req model.CompleteGRCAssessmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeGRCError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	assessment, err := h.svc.CompleteAssessment(r.Context(), id, req)
	if err != nil {
		handleGRCError(w, err)
		return
	}
	writeGRCJSON(w, http.StatusOK, assessment)
}

// DeleteAssessment handles DELETE /v2/compliance/assessments/{id}
func (h *ComplianceHandler) DeleteAssessment(w http.ResponseWriter, r *http.Request, id string) {
	if err := h.svc.DeleteAssessment(r.Context(), id); err != nil {
		handleGRCError(w, err)
		return
	}
	writeGRCJSON(w, http.StatusOK, map[string]string{"message": "compliance assessment deleted"})
}

// ListFindings handles GET /v2/compliance/findings
func (h *ComplianceHandler) ListFindings(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	limit, _ := strconv.Atoi(q.Get("limit"))
	offset, _ := strconv.Atoi(q.Get("offset"))

	filter := model.GRCFindingFilter{
		WorkspaceID:  getGRCWorkspaceID(r),
		AssessmentID: q.Get("assessment_id"),
		ControlID:    q.Get("control_id"),
		Severity:     q.Get("severity"),
		Status:       q.Get("status"),
		OwnerID:      q.Get("owner_id"),
		Limit:        limit,
		Offset:       offset,
	}

	findings, total, err := h.svc.ListFindings(r.Context(), filter)
	if err != nil {
		handleGRCError(w, err)
		return
	}

	writeGRCJSON(w, http.StatusOK, map[string]interface{}{
		"findings": findings,
		"total":    total,
		"limit":    filter.Limit,
		"offset":   filter.Offset,
	})
}

// CreateFinding handles POST /v2/compliance/findings
func (h *ComplianceHandler) CreateFinding(w http.ResponseWriter, r *http.Request) {
	var req model.CreateGRCFindingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeGRCError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.OwnerID == "" {
		req.OwnerID = getGRCUser(r)
	}

	finding, err := h.svc.CreateFinding(r.Context(), getGRCWorkspaceID(r), req)
	if err != nil {
		handleGRCError(w, err)
		return
	}
	writeGRCJSON(w, http.StatusCreated, finding)
}

// GetFindingByID handles GET /v2/compliance/findings/{id}
func (h *ComplianceHandler) GetFindingByID(w http.ResponseWriter, r *http.Request, id string) {
	finding, err := h.svc.GetFindingByID(r.Context(), id)
	if err != nil {
		handleGRCError(w, err)
		return
	}
	writeGRCJSON(w, http.StatusOK, finding)
}

// UpdateFinding handles PUT /v2/compliance/findings/{id}
func (h *ComplianceHandler) UpdateFinding(w http.ResponseWriter, r *http.Request, id string) {
	var req model.UpdateGRCFindingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeGRCError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	finding, err := h.svc.UpdateFinding(r.Context(), id, req)
	if err != nil {
		handleGRCError(w, err)
		return
	}
	writeGRCJSON(w, http.StatusOK, finding)
}

// DeleteFinding handles DELETE /v2/compliance/findings/{id}
func (h *ComplianceHandler) DeleteFinding(w http.ResponseWriter, r *http.Request, id string) {
	if err := h.svc.DeleteFinding(r.Context(), id); err != nil {
		handleGRCError(w, err)
		return
	}
	writeGRCJSON(w, http.StatusOK, map[string]string{"message": "compliance finding deleted"})
}

// Analytics & ServeHTTP stubs
func (h *ComplianceHandler) GetComplianceAnalytics(w http.ResponseWriter, r *http.Request) {}
func (h *ComplianceHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	writeGRCError(w, http.StatusNotFound, "route not found")
}
