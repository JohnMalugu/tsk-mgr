package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/JohnMalugu/tsk-mgr-api/internal/model"
	"github.com/JohnMalugu/tsk-mgr-api/internal/repository"
	"github.com/JohnMalugu/tsk-mgr-api/internal/service"
)

// AuditHandler exposes HTTP endpoints for compliance audit trails, security alerts, and policies
type AuditHandler struct {
	svc service.AuditService
}

var GlobalAuditHandler *AuditHandler

// InitAuditHandler initializes the singleton audit handler
func InitAuditHandler(svc service.AuditService) {
	GlobalAuditHandler = &AuditHandler{svc: svc}
}

// NewAuditHandler creates a new AuditHandler instance
func NewAuditHandler(svc service.AuditService) *AuditHandler {
	return &AuditHandler{svc: svc}
}

func getAuditUser(r *http.Request) string {
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

func writeAuditJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func writeAuditError(w http.ResponseWriter, status int, msg string) {
	writeAuditJSON(w, status, map[string]string{"error": msg})
}

func handleAuditServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrAuditLogNotFoundDB), errors.Is(err, repository.ErrAuditLogNotFound):
		writeAuditError(w, http.StatusNotFound, "audit log not found")
	case errors.Is(err, service.ErrAuditAlertNotFoundDB), errors.Is(err, repository.ErrAuditAlertNotFound):
		writeAuditError(w, http.StatusNotFound, "audit alert not found")
	case errors.Is(err, service.ErrRetentionPolicyNotFound), errors.Is(err, repository.ErrRetentionPolicyNotFound):
		writeAuditError(w, http.StatusNotFound, "retention policy not found")
	case errors.Is(err, service.ErrAuditExportNotFound), errors.Is(err, repository.ErrAuditExportNotFound):
		writeAuditError(w, http.StatusNotFound, "audit export not found")
	case errors.Is(err, service.ErrWorkspaceNotFound):
		writeAuditError(w, http.StatusNotFound, "workspace not found")
	case errors.Is(err, service.ErrWorkspaceRequired):
		writeAuditError(w, http.StatusBadRequest, "workspace_id is required")
	case errors.Is(err, service.ErrInvalidRetentionPeriod):
		writeAuditError(w, http.StatusBadRequest, "retention period must be at least 30 days")
	case errors.Is(err, service.ErrInvalidAlertStatus):
		writeAuditError(w, http.StatusBadRequest, "invalid alert status: must be acknowledged, resolved, or dismissed")
	case errors.Is(err, service.ErrExportFormatUnsupported):
		writeAuditError(w, http.StatusBadRequest, "unsupported export format: only 'json' and 'csv' are supported")
	case errors.Is(err, model.ErrAuditActionRequired):
		writeAuditError(w, http.StatusBadRequest, "audit action is required")
	case errors.Is(err, model.ErrInvalidAuditSeverity):
		writeAuditError(w, http.StatusBadRequest, "invalid audit severity level (info, warning, critical)")
	default:
		writeAuditError(w, http.StatusInternalServerError, err.Error())
	}
}

// RecordEvent handles POST /v2/audit
func (h *AuditHandler) RecordEvent(w http.ResponseWriter, r *http.Request) {
	var req model.CreateAuditLogRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeAuditError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	actorID := getAuditUser(r)

	log, err := h.svc.RecordEvent(actorID, req)
	if err != nil {
		handleAuditServiceError(w, err)
		return
	}
	writeAuditJSON(w, http.StatusCreated, log)
}

// GetEvent handles GET /v2/audit/{id}
func (h *AuditHandler) GetEvent(w http.ResponseWriter, r *http.Request, id int64) {
	log, err := h.svc.GetEvent(id)
	if err != nil {
		handleAuditServiceError(w, err)
		return
	}
	writeAuditJSON(w, http.StatusOK, log)
}

// ListEvents handles GET /v2/audit
func (h *AuditHandler) ListEvents(w http.ResponseWriter, r *http.Request) {
	wsIDStr := r.URL.Query().Get("workspace_id")
	if wsIDStr == "" {
		writeAuditError(w, http.StatusBadRequest, "workspace_id is required")
		return
	}
	wsID, err := strconv.ParseInt(wsIDStr, 10, 64)
	if err != nil || wsID <= 0 {
		writeAuditError(w, http.StatusBadRequest, "invalid workspace_id")
		return
	}

	filter := model.AuditLogFilter{
		WorkspaceID: wsID,
	}

	if actorID := r.URL.Query().Get("actor_id"); actorID != "" {
		filter.ActorID = &actorID
	}
	if action := r.URL.Query().Get("action"); action != "" {
		filter.Action = &action
	}
	if cat := r.URL.Query().Get("category"); cat != "" {
		filter.Category = &cat
	}
	if sev := r.URL.Query().Get("severity"); sev != "" {
		filter.Severity = &sev
	}
	if entityType := r.URL.Query().Get("entity_type"); entityType != "" {
		filter.EntityType = &entityType
	}
	if entityID := r.URL.Query().Get("entity_id"); entityID != "" {
		filter.EntityID = &entityID
	}
	if search := r.URL.Query().Get("search"); search != "" {
		filter.Search = &search
	}
	if startDate := r.URL.Query().Get("start_date"); startDate != "" {
		filter.StartDate = &startDate
	}
	if endDate := r.URL.Query().Get("end_date"); endDate != "" {
		filter.EndDate = &endDate
	}
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
			filter.Limit = limit
		}
	}
	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil && offset >= 0 {
			filter.Offset = offset
		}
	}

	logs, total, err := h.svc.ListEvents(filter)
	if err != nil {
		handleAuditServiceError(w, err)
		return
	}

	resp := map[string]interface{}{
		"data":   logs,
		"total":  total,
		"limit":  filter.Limit,
		"offset": filter.Offset,
	}
	writeAuditJSON(w, http.StatusOK, resp)
}
