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

// CSIHandler exposes HTTP endpoints for Continual Service Improvement
type CSIHandler struct {
	svc service.CSIService
}

// GlobalCSIHandler is the singleton handler instance
var GlobalCSIHandler *CSIHandler

// InitCSIHandler initializes the global CSI handler singleton
func InitCSIHandler(svc service.CSIService) {
	GlobalCSIHandler = &CSIHandler{svc: svc}
}

// NewCSIHandler creates a new CSIHandler instance
func NewCSIHandler(svc service.CSIService) *CSIHandler {
	return &CSIHandler{svc: svc}
}

func getCSIUser(r *http.Request) string {
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

func getCSIWorkspaceID(r *http.Request) string {
	if ws := r.URL.Query().Get("workspace_id"); ws != "" {
		return ws
	}
	if ws := r.Header.Get("X-Workspace-ID"); ws != "" {
		return ws
	}
	return "ws_1"
}

func writeCSIJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func writeCSIError(w http.ResponseWriter, status int, msg string) {
	writeCSIJSON(w, status, map[string]string{"error": msg})
}

func handleCSIError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrCSIInitiativeNotFound):
		writeCSIError(w, http.StatusNotFound, "continual improvement initiative not found")
	case errors.Is(err, service.ErrCSIInitiativeTitleRequired):
		writeCSIError(w, http.StatusBadRequest, "initiative title is required")
	case errors.Is(err, service.ErrCSIInitiativeCodeConflict):
		writeCSIError(w, http.StatusConflict, "initiative with this code already exists in workspace")
	case errors.Is(err, model.ErrInvalidCSIAlignment):
		writeCSIError(w, http.StatusBadRequest, "invalid strategic alignment category")
	case errors.Is(err, model.ErrInvalidCSIStatus):
		writeCSIError(w, http.StatusBadRequest, "invalid CSI initiative status")
	case errors.Is(err, model.ErrInvalidCSIPriority):
		writeCSIError(w, http.StatusBadRequest, "invalid CSI initiative priority")
	case errors.Is(err, model.ErrInvalidCSISource):
		writeCSIError(w, http.StatusBadRequest, "invalid CSI initiative source")
	case errors.Is(err, service.ErrCSIBenchmarkNotFound):
		writeCSIError(w, http.StatusNotFound, "CSI benchmark metric not found")
	case errors.Is(err, service.ErrCSIBenchmarkMetricNameRequired):
		writeCSIError(w, http.StatusBadRequest, "benchmark metric name is required")
	case errors.Is(err, model.ErrInvalidCSIMetricUnit):
		writeCSIError(w, http.StatusBadRequest, "invalid CSI benchmark metric unit")
	case errors.Is(err, service.ErrCSIBenefitNotFound):
		writeCSIError(w, http.StatusNotFound, "CSI benefit realization record not found")
	case errors.Is(err, model.ErrInvalidCSIBenefitStatus):
		writeCSIError(w, http.StatusBadRequest, "invalid CSI benefit realization status")
	case errors.Is(err, model.ErrInvalidEffortHours):
		writeCSIError(w, http.StatusBadRequest, "effort hours must be non-negative")
	case errors.Is(err, service.ErrCSIActionNotFound):
		writeCSIError(w, http.StatusNotFound, "CSI action item not found")
	case errors.Is(err, model.ErrCSIActionDescriptionRequired):
		writeCSIError(w, http.StatusBadRequest, "action description is required")
	case errors.Is(err, model.ErrInvalidCSIActionStatus):
		writeCSIError(w, http.StatusBadRequest, "invalid CSI action status")
	case errors.Is(err, service.ErrCSIReviewNotFound):
		writeCSIError(w, http.StatusNotFound, "CSI review record not found")
	case errors.Is(err, model.ErrCSIReviewNotesRequired):
		writeCSIError(w, http.StatusBadRequest, "review notes or findings are required")
	case errors.Is(err, model.ErrInvalidCSIReviewOutcome):
		writeCSIError(w, http.StatusBadRequest, "invalid CSI review outcome")
	default:
		writeCSIError(w, http.StatusInternalServerError, err.Error())
	}
}

// Initiative Endpoints

// ListInitiatives handles GET /v2/csi/initiatives
func (h *CSIHandler) ListInitiatives(w http.ResponseWriter, r *http.Request) {
	wsID := getCSIWorkspaceID(r)
	q := r.URL.Query()

	limit := 50
	offset := 0
	if l := q.Get("limit"); l != "" {
		if val, err := strconv.Atoi(l); err == nil && val > 0 {
			limit = val
		}
	}
	if o := q.Get("offset"); o != "" {
		if val, err := strconv.Atoi(o); err == nil && val >= 0 {
			offset = val
		}
	}

	filter := model.CSIFilterCriteria{
		WorkspaceID:        wsID,
		StrategicAlignment: q.Get("strategic_alignment"),
		Priority:           q.Get("priority"),
		Status:             q.Get("status"),
		Source:             q.Get("source"),
		CIID:               q.Get("ci_id"),
		Limit:              limit,
		Offset:             offset,
	}

	initiatives, total, err := h.svc.ListInitiatives(r.Context(), filter)
	if err != nil {
		handleCSIError(w, err)
		return
	}

	writeCSIJSON(w, http.StatusOK, map[string]interface{}{
		"initiatives": initiatives,
		"total":       total,
		"limit":       limit,
		"offset":      offset,
	})
}

// CreateInitiative handles POST /v2/csi/initiatives
func (h *CSIHandler) CreateInitiative(w http.ResponseWriter, r *http.Request) {
	wsID := getCSIWorkspaceID(r)

	var req model.CreateCSIInitiativeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeCSIError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	initiative, err := h.svc.CreateInitiative(r.Context(), wsID, req)
	if err != nil {
		handleCSIError(w, err)
		return
	}

	writeCSIJSON(w, http.StatusCreated, initiative)
}

// GetInitiativeByID handles GET /v2/csi/initiatives/:id
func (h *CSIHandler) GetInitiativeByID(w http.ResponseWriter, r *http.Request, id string) {
	initiative, err := h.svc.GetInitiativeByID(r.Context(), id)
	if err != nil {
		handleCSIError(w, err)
		return
	}
	writeCSIJSON(w, http.StatusOK, initiative)
}

// UpdateInitiative handles PUT /v2/csi/initiatives/:id
func (h *CSIHandler) UpdateInitiative(w http.ResponseWriter, r *http.Request, id string) {
	var req model.UpdateCSIInitiativeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeCSIError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	initiative, err := h.svc.UpdateInitiative(r.Context(), id, req)
	if err != nil {
		handleCSIError(w, err)
		return
	}
	writeCSIJSON(w, http.StatusOK, initiative)
}

// CompleteInitiative handles POST /v2/csi/initiatives/:id/complete
func (h *CSIHandler) CompleteInitiative(w http.ResponseWriter, r *http.Request, id string) {
	initiative, err := h.svc.CompleteInitiative(r.Context(), id)
	if err != nil {
		handleCSIError(w, err)
		return
	}
	writeCSIJSON(w, http.StatusOK, initiative)
}

// DeleteInitiative handles DELETE /v2/csi/initiatives/:id
func (h *CSIHandler) DeleteInitiative(w http.ResponseWriter, r *http.Request, id string) {
	if err := h.svc.DeleteInitiative(r.Context(), id); err != nil {
		handleCSIError(w, err)
		return
	}
	writeCSIJSON(w, http.StatusOK, map[string]string{"message": "continual improvement initiative deleted successfully"})
}

// Benchmark Endpoints

// ListBenchmarks handles GET /v2/csi/initiatives/:id/benchmarks
func (h *CSIHandler) ListBenchmarks(w http.ResponseWriter, r *http.Request, initiativeID string) {
	benchmarks, err := h.svc.ListBenchmarks(r.Context(), initiativeID)
	if err != nil {
		handleCSIError(w, err)
		return
	}
	writeCSIJSON(w, http.StatusOK, map[string]interface{}{
		"benchmarks": benchmarks,
		"total":      len(benchmarks),
	})
}

// AddBenchmark handles POST /v2/csi/initiatives/:id/benchmarks
func (h *CSIHandler) AddBenchmark(w http.ResponseWriter, r *http.Request, initiativeID string) {
	wsID := getCSIWorkspaceID(r)

	var req model.CreateCSIBenchmarkRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeCSIError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	benchmark, err := h.svc.AddBenchmark(r.Context(), wsID, initiativeID, req)
	if err != nil {
		handleCSIError(w, err)
		return
	}
	writeCSIJSON(w, http.StatusCreated, benchmark)
}

// GetBenchmarkByID handles GET /v2/csi/benchmarks/:id
func (h *CSIHandler) GetBenchmarkByID(w http.ResponseWriter, r *http.Request, id string) {
	benchmark, err := h.svc.GetBenchmarkByID(r.Context(), id)
	if err != nil {
		handleCSIError(w, err)
		return
	}
	writeCSIJSON(w, http.StatusOK, benchmark)
}

// UpdateBenchmark handles PUT /v2/csi/benchmarks/:id
func (h *CSIHandler) UpdateBenchmark(w http.ResponseWriter, r *http.Request, id string) {
	var req model.UpdateCSIBenchmarkRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeCSIError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	benchmark, err := h.svc.UpdateBenchmark(r.Context(), id, req)
	if err != nil {
		handleCSIError(w, err)
		return
	}
	writeCSIJSON(w, http.StatusOK, benchmark)
}

// VerifyBenchmark handles POST /v2/csi/benchmarks/:id/verify
func (h *CSIHandler) VerifyBenchmark(w http.ResponseWriter, r *http.Request, id string) {
	var req struct {
		CurrentValue float64 `json:"current_value"`
		Notes        string  `json:"notes,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeCSIError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	benchmark, err := h.svc.VerifyBenchmark(r.Context(), id, req.CurrentValue, req.Notes)
	if err != nil {
		handleCSIError(w, err)
		return
	}
	writeCSIJSON(w, http.StatusOK, benchmark)
}

// DeleteBenchmark handles DELETE /v2/csi/benchmarks/:id
func (h *CSIHandler) DeleteBenchmark(w http.ResponseWriter, r *http.Request, id string) {
	if err := h.svc.DeleteBenchmark(r.Context(), id); err != nil {
		handleCSIError(w, err)
		return
	}
	writeCSIJSON(w, http.StatusOK, map[string]string{"message": "benchmark deleted successfully"})
}

// Benefit Endpoints

// GetBenefit handles GET /v2/csi/initiatives/:id/benefit
func (h *CSIHandler) GetBenefit(w http.ResponseWriter, r *http.Request, initiativeID string) {
	benefit, err := h.svc.GetBenefit(r.Context(), initiativeID)
	if err != nil {
		handleCSIError(w, err)
		return
	}
	writeCSIJSON(w, http.StatusOK, benefit)
}

// RecordBenefit handles POST /v2/csi/initiatives/:id/benefit
func (h *CSIHandler) RecordBenefit(w http.ResponseWriter, r *http.Request, initiativeID string) {
	wsID := getCSIWorkspaceID(r)

	var req model.RecordCSIBenefitRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeCSIError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	benefit, err := h.svc.RecordBenefit(r.Context(), wsID, initiativeID, req)
	if err != nil {
		handleCSIError(w, err)
		return
	}
	writeCSIJSON(w, http.StatusCreated, benefit)
}

// UpdateBenefit handles PUT /v2/csi/initiatives/:id/benefit
func (h *CSIHandler) UpdateBenefit(w http.ResponseWriter, r *http.Request, initiativeID string) {
	var req model.UpdateCSIBenefitRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeCSIError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	benefit, err := h.svc.UpdateBenefit(r.Context(), initiativeID, req)
	if err != nil {
		handleCSIError(w, err)
		return
	}
	writeCSIJSON(w, http.StatusOK, benefit)
}

// Action Endpoints

// ListActions handles GET /v2/csi/initiatives/:id/actions
func (h *CSIHandler) ListActions(w http.ResponseWriter, r *http.Request, initiativeID string) {
	actions, err := h.svc.ListActions(r.Context(), initiativeID)
	if err != nil {
		handleCSIError(w, err)
		return
	}
	writeCSIJSON(w, http.StatusOK, map[string]interface{}{
		"actions": actions,
		"total":   len(actions),
	})
}

// CreateAction handles POST /v2/csi/initiatives/:id/actions
func (h *CSIHandler) CreateAction(w http.ResponseWriter, r *http.Request, initiativeID string) {
	wsID := getCSIWorkspaceID(r)

	var req model.CreateCSIActionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeCSIError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	action, err := h.svc.CreateAction(r.Context(), wsID, initiativeID, req)
	if err != nil {
		handleCSIError(w, err)
		return
	}
	writeCSIJSON(w, http.StatusCreated, action)
}

// GetActionByID handles GET /v2/csi/actions/:id
func (h *CSIHandler) GetActionByID(w http.ResponseWriter, r *http.Request, id string) {
	action, err := h.svc.GetActionByID(r.Context(), id)
	if err != nil {
		handleCSIError(w, err)
		return
	}
	writeCSIJSON(w, http.StatusOK, action)
}

// UpdateAction handles PUT /v2/csi/actions/:id
func (h *CSIHandler) UpdateAction(w http.ResponseWriter, r *http.Request, id string) {
	var req model.UpdateCSIActionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeCSIError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	action, err := h.svc.UpdateAction(r.Context(), id, req)
	if err != nil {
		handleCSIError(w, err)
		return
	}
	writeCSIJSON(w, http.StatusOK, action)
}

// CompleteAction handles POST /v2/csi/actions/:id/complete
func (h *CSIHandler) CompleteAction(w http.ResponseWriter, r *http.Request, id string) {
	var req struct {
		VerificationEvidence string `json:"verification_evidence"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeCSIError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	action, err := h.svc.CompleteAction(r.Context(), id, req.VerificationEvidence)
	if err != nil {
		handleCSIError(w, err)
		return
	}
	writeCSIJSON(w, http.StatusOK, action)
}

// DeleteAction handles DELETE /v2/csi/actions/:id
func (h *CSIHandler) DeleteAction(w http.ResponseWriter, r *http.Request, id string) {
	if err := h.svc.DeleteAction(r.Context(), id); err != nil {
		handleCSIError(w, err)
		return
	}
	writeCSIJSON(w, http.StatusOK, map[string]string{"message": "action item deleted successfully"})
}

// Review Endpoints

// ListReviews handles GET /v2/csi/initiatives/:id/reviews
func (h *CSIHandler) ListReviews(w http.ResponseWriter, r *http.Request, initiativeID string) {
	reviews, err := h.svc.ListReviews(r.Context(), initiativeID)
	if err != nil {
		handleCSIError(w, err)
		return
	}
	writeCSIJSON(w, http.StatusOK, map[string]interface{}{
		"reviews": reviews,
		"total":   len(reviews),
	})
}

// RecordReview handles POST /v2/csi/initiatives/:id/reviews
func (h *CSIHandler) RecordReview(w http.ResponseWriter, r *http.Request, initiativeID string) {
	wsID := getCSIWorkspaceID(r)

	var req model.RecordCSIReviewRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeCSIError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	review, err := h.svc.RecordReview(r.Context(), wsID, initiativeID, req)
	if err != nil {
		handleCSIError(w, err)
		return
	}
	writeCSIJSON(w, http.StatusCreated, review)
}

// GetReviewByID handles GET /v2/csi/reviews/:id
func (h *CSIHandler) GetReviewByID(w http.ResponseWriter, r *http.Request, id string) {
	review, err := h.svc.GetReviewByID(r.Context(), id)
	if err != nil {
		handleCSIError(w, err)
		return
	}
	writeCSIJSON(w, http.StatusOK, review)
}

// Analytics Endpoints

// GetAnalytics handles GET /v2/csi/analytics
func (h *CSIHandler) GetAnalytics(w http.ResponseWriter, r *http.Request) {
	wsID := getCSIWorkspaceID(r)

	analytics, err := h.svc.GetCSIAnalytics(r.Context(), wsID)
	if err != nil {
		handleCSIError(w, err)
		return
	}
	writeCSIJSON(w, http.StatusOK, analytics)
}

// ServeHTTP multiplexes sub-routes under /v2/csi
func (h *CSIHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/v2/csi")
	path = strings.TrimPrefix(path, "/")

	// Analytics
	if path == "analytics" {
		if r.Method == http.MethodGet {
			h.GetAnalytics(w, r)
		} else {
			writeCSIError(w, http.StatusMethodNotAllowed, "method not allowed")
		}
		return
	}

	// Direct Benchmarks sub-path (/v2/csi/benchmarks/:id)
	if strings.HasPrefix(path, "benchmarks") {
		sub := strings.TrimPrefix(path, "benchmarks")
		sub = strings.TrimPrefix(sub, "/")
		parts := strings.Split(sub, "/")
		benchmarkID := parts[0]
		if len(parts) == 1 {
			switch r.Method {
			case http.MethodGet:
				h.GetBenchmarkByID(w, r, benchmarkID)
			case http.MethodPut:
				h.UpdateBenchmark(w, r, benchmarkID)
			case http.MethodDelete:
				h.DeleteBenchmark(w, r, benchmarkID)
			default:
				writeCSIError(w, http.StatusMethodNotAllowed, "method not allowed")
			}
			return
		}
		if len(parts) == 2 && parts[1] == "verify" && r.Method == http.MethodPost {
			h.VerifyBenchmark(w, r, benchmarkID)
			return
		}
	}

	// Direct Actions sub-path (/v2/csi/actions/:id)
	if strings.HasPrefix(path, "actions") {
		sub := strings.TrimPrefix(path, "actions")
		sub = strings.TrimPrefix(sub, "/")
		parts := strings.Split(sub, "/")
		actionID := parts[0]
		if len(parts) == 1 {
			switch r.Method {
			case http.MethodGet:
				h.GetActionByID(w, r, actionID)
			case http.MethodPut:
				h.UpdateAction(w, r, actionID)
			case http.MethodDelete:
				h.DeleteAction(w, r, actionID)
			default:
				writeCSIError(w, http.StatusMethodNotAllowed, "method not allowed")
			}
			return
		}
		if len(parts) == 2 && parts[1] == "complete" && r.Method == http.MethodPost {
			h.CompleteAction(w, r, actionID)
			return
		}
	}

	// Direct Reviews sub-path (/v2/csi/reviews/:id)
	if strings.HasPrefix(path, "reviews") {
		sub := strings.TrimPrefix(path, "reviews")
		sub = strings.TrimPrefix(sub, "/")
		if r.Method == http.MethodGet {
			h.GetReviewByID(w, r, sub)
		} else {
			writeCSIError(w, http.StatusMethodNotAllowed, "method not allowed")
		}
		return
	}

	// Initiatives (/v2/csi/initiatives)
	if strings.HasPrefix(path, "initiatives") {
		sub := strings.TrimPrefix(path, "initiatives")
		sub = strings.TrimPrefix(sub, "/")
		if sub == "" {
			switch r.Method {
			case http.MethodGet:
				h.ListInitiatives(w, r)
			case http.MethodPost:
				h.CreateInitiative(w, r)
			default:
				writeCSIError(w, http.StatusMethodNotAllowed, "method not allowed")
			}
			return
		}
		parts := strings.Split(sub, "/")
		initiativeID := parts[0]
		if len(parts) == 1 {
			switch r.Method {
			case http.MethodGet:
				h.GetInitiativeByID(w, r, initiativeID)
			case http.MethodPut:
				h.UpdateInitiative(w, r, initiativeID)
			case http.MethodDelete:
				h.DeleteInitiative(w, r, initiativeID)
			default:
				writeCSIError(w, http.StatusMethodNotAllowed, "method not allowed")
			}
			return
		}

		// Nested initiative routes
		switch parts[1] {
		case "complete":
			if r.Method == http.MethodPost {
				h.CompleteInitiative(w, r, initiativeID)
			} else {
				writeCSIError(w, http.StatusMethodNotAllowed, "method not allowed")
			}
			return
		case "benchmarks":
			switch r.Method {
			case http.MethodGet:
				h.ListBenchmarks(w, r, initiativeID)
			case http.MethodPost:
				h.AddBenchmark(w, r, initiativeID)
			default:
				writeCSIError(w, http.StatusMethodNotAllowed, "method not allowed")
			}
			return
		case "benefit":
			switch r.Method {
			case http.MethodGet:
				h.GetBenefit(w, r, initiativeID)
			case http.MethodPost:
				h.RecordBenefit(w, r, initiativeID)
			case http.MethodPut:
				h.UpdateBenefit(w, r, initiativeID)
			default:
				writeCSIError(w, http.StatusMethodNotAllowed, "method not allowed")
			}
			return
		case "actions":
			switch r.Method {
			case http.MethodGet:
				h.ListActions(w, r, initiativeID)
			case http.MethodPost:
				h.CreateAction(w, r, initiativeID)
			default:
				writeCSIError(w, http.StatusMethodNotAllowed, "method not allowed")
			}
			return
		case "reviews":
			switch r.Method {
			case http.MethodGet:
				h.ListReviews(w, r, initiativeID)
			case http.MethodPost:
				h.RecordReview(w, r, initiativeID)
			default:
				writeCSIError(w, http.StatusMethodNotAllowed, "method not allowed")
			}
			return
		}
	}

	writeCSIError(w, http.StatusNotFound, "route not found")
}
