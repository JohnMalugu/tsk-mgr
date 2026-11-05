package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/JohnMalugu/tsk-mgr-api/internal/model"
	"github.com/JohnMalugu/tsk-mgr-api/internal/service"
)

// RiskHandler exposes HTTP endpoints for enterprise risk management
type RiskHandler struct {
	svc service.RiskService
}

var GlobalRiskHandler *RiskHandler

// InitRiskHandler initializes the singleton risk handler
func InitRiskHandler(svc service.RiskService) {
	GlobalRiskHandler = &RiskHandler{svc: svc}
}

// NewRiskHandler creates a new RiskHandler instance
func NewRiskHandler(svc service.RiskService) *RiskHandler {
	return &RiskHandler{svc: svc}
}

func getRiskUser(r *http.Request) string {
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

func writeRiskJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func writeRiskError(w http.ResponseWriter, status int, msg string) {
	writeRiskJSON(w, status, map[string]string{"error": msg})
}

func handleRiskServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrRiskNotFound):
		writeRiskError(w, http.StatusNotFound, "risk not found")
	case errors.Is(err, service.ErrWorkspaceRequired):
		writeRiskError(w, http.StatusBadRequest, "workspace_id is required")
	case errors.Is(err, service.ErrProjectNotFound):
		writeRiskError(w, http.StatusNotFound, "project not found")
	case errors.Is(err, service.ErrTaskNotFound):
		writeRiskError(w, http.StatusNotFound, "task not found")
	case errors.Is(err, service.ErrRiskTitleRequired):
		writeRiskError(w, http.StatusBadRequest, "risk title is required")
	case errors.Is(err, service.ErrInvalidRiskCategory):
		writeRiskError(w, http.StatusBadRequest, "invalid risk category")
	case errors.Is(err, service.ErrInvalidRiskStatus):
		writeRiskError(w, http.StatusBadRequest, "invalid risk status")
	case errors.Is(err, service.ErrInvalidRiskStrategy):
		writeRiskError(w, http.StatusBadRequest, "invalid risk strategy")
	case errors.Is(err, service.ErrInvalidProbability):
		writeRiskError(w, http.StatusBadRequest, "probability must be between 1 and 5")
	case errors.Is(err, service.ErrInvalidImpact):
		writeRiskError(w, http.StatusBadRequest, "impact must be between 1 and 5")
	case errors.Is(err, service.ErrInvalidRiskAction):
		writeRiskError(w, http.StatusBadRequest, "invalid risk action type (must be mitigation, contingency, or investigation)")
	case errors.Is(err, service.ErrReviewCommentReq):
		writeRiskError(w, http.StatusBadRequest, "review comment is required")
	default:
		writeRiskError(w, http.StatusInternalServerError, err.Error())
	}
}

// ListRisks handles GET /v2/risks with query filtering
func (h *RiskHandler) ListRisks(w http.ResponseWriter, r *http.Request) {
	wsStr := r.URL.Query().Get("workspace_id")
	if wsStr == "" {
		writeRiskError(w, http.StatusBadRequest, "workspace_id is required")
		return
	}
	wsID, err := strconv.ParseInt(wsStr, 10, 64)
	if err != nil || wsID <= 0 {
		writeRiskError(w, http.StatusBadRequest, "invalid workspace_id")
		return
	}

	filter := model.RiskFilter{
		WorkspaceID: wsID,
		Category:    r.URL.Query().Get("category"),
		Severity:    r.URL.Query().Get("severity"),
		Status:      r.URL.Query().Get("status"),
		Strategy:    r.URL.Query().Get("strategy"),
		OwnerID:     r.URL.Query().Get("owner_id"),
		Search:      r.URL.Query().Get("search"),
	}

	if projStr := r.URL.Query().Get("project_id"); projStr != "" {
		if pid, err := strconv.Atoi(projStr); err == nil {
			filter.ProjectID = &pid
		}
	}
	if minStr := r.URL.Query().Get("min_score"); minStr != "" {
		filter.MinScore, _ = strconv.Atoi(minStr)
	}
	if maxStr := r.URL.Query().Get("max_score"); maxStr != "" {
		filter.MaxScore, _ = strconv.Atoi(maxStr)
	}
	if limStr := r.URL.Query().Get("limit"); limStr != "" {
		filter.Limit, _ = strconv.Atoi(limStr)
	}
	if offStr := r.URL.Query().Get("offset"); offStr != "" {
		filter.Offset, _ = strconv.Atoi(offStr)
	}

	risks, err := h.svc.ListRisks(filter)
	if err != nil {
		handleRiskServiceError(w, err)
		return
	}
	if risks == nil {
		risks = []*model.Risk{}
	}

	writeRiskJSON(w, http.StatusOK, map[string]interface{}{
		"risks": risks,
		"count": len(risks),
	})
}

// CreateRisk handles POST /v2/risks
func (h *RiskHandler) CreateRisk(w http.ResponseWriter, r *http.Request) {
	var req model.CreateRiskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeRiskError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	creatorID := getRiskUser(r)
	risk, err := h.svc.CreateRisk(creatorID, req)
	if err != nil {
		handleRiskServiceError(w, err)
		return
	}

	writeRiskJSON(w, http.StatusCreated, risk)
}

// GetRisk handles GET /v2/risks/{id}
func (h *RiskHandler) GetRisk(w http.ResponseWriter, r *http.Request, id int64) {
	risk, err := h.svc.GetRisk(id)
	if err != nil {
		handleRiskServiceError(w, err)
		return
	}
	writeRiskJSON(w, http.StatusOK, risk)
}

// UpdateRisk handles PUT/PATCH /v2/risks/{id}
func (h *RiskHandler) UpdateRisk(w http.ResponseWriter, r *http.Request, id int64) {
	var req model.UpdateRiskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeRiskError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	risk, err := h.svc.UpdateRisk(id, req)
	if err != nil {
		handleRiskServiceError(w, err)
		return
	}
	writeRiskJSON(w, http.StatusOK, risk)
}

// DeleteRisk handles DELETE /v2/risks/{id}
func (h *RiskHandler) DeleteRisk(w http.ResponseWriter, r *http.Request, id int64) {
	if err := h.svc.DeleteRisk(id); err != nil {
		handleRiskServiceError(w, err)
		return
	}
	writeRiskJSON(w, http.StatusOK, map[string]string{"message": "risk deleted successfully"})
}
