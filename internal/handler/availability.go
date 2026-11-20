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

// AvailabilityHandler exposes HTTP endpoints for ITIL Availability Management
type AvailabilityHandler struct {
	svc service.AvailabilityService
}

// GlobalAvailabilityHandler is the singleton handler instance
var GlobalAvailabilityHandler *AvailabilityHandler

// InitAvailabilityHandler initializes the global availability handler singleton
func InitAvailabilityHandler(svc service.AvailabilityService) {
	GlobalAvailabilityHandler = &AvailabilityHandler{svc: svc}
}

// NewAvailabilityHandler creates a new AvailabilityHandler instance
func NewAvailabilityHandler(svc service.AvailabilityService) *AvailabilityHandler {
	return &AvailabilityHandler{svc: svc}
}

func getAvailabilityUser(r *http.Request) string {
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

func getAvailabilityWorkspaceID(r *http.Request) string {
	if ws := r.URL.Query().Get("workspace_id"); ws != "" {
		return ws
	}
	if ws := r.Header.Get("X-Workspace-ID"); ws != "" {
		return ws
	}
	return "ws_1"
}

func writeAvailabilityJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func writeAvailabilityError(w http.ResponseWriter, status int, msg string) {
	writeAvailabilityJSON(w, status, map[string]string{"error": msg})
}

func handleAvailabilityError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrAvailabilityTargetNotFound):
		writeAvailabilityError(w, http.StatusNotFound, "availability target not found")
	case errors.Is(err, service.ErrAvailabilityTargetNameRequired):
		writeAvailabilityError(w, http.StatusBadRequest, "availability target name is required")
	case errors.Is(err, service.ErrAvailabilityTargetNameConflict):
		writeAvailabilityError(w, http.StatusConflict, "availability target with this name already exists in workspace")
	case errors.Is(err, model.ErrInvalidTargetPercentage):
		writeAvailabilityError(w, http.StatusBadRequest, "target percentage must be between 0.0 and 100.0")
	case errors.Is(err, model.ErrInvalidMeasurementWindow):
		writeAvailabilityError(w, http.StatusBadRequest, "invalid measurement window")
	case errors.Is(err, model.ErrInvalidServiceHours):
		writeAvailabilityError(w, http.StatusBadRequest, "invalid service hours")
	case errors.Is(err, model.ErrInvalidTargetStatus):
		writeAvailabilityError(w, http.StatusBadRequest, "invalid target status")
	case errors.Is(err, service.ErrOutageEventNotFound):
		writeAvailabilityError(w, http.StatusNotFound, "outage event not found")
	case errors.Is(err, model.ErrOutageTitleRequired):
		writeAvailabilityError(w, http.StatusBadRequest, "outage title is required")
	case errors.Is(err, model.ErrInvalidOutageType):
		writeAvailabilityError(w, http.StatusBadRequest, "invalid outage type")
	case errors.Is(err, model.ErrInvalidOutageSeverity):
		writeAvailabilityError(w, http.StatusBadRequest, "invalid outage severity")
	case errors.Is(err, model.ErrInvalidOutageTimestamps):
		writeAvailabilityError(w, http.StatusBadRequest, "outage ended_at cannot be before started_at")
	case errors.Is(err, service.ErrAvailabilityMeasurementNotFound):
		writeAvailabilityError(w, http.StatusNotFound, "availability measurement not found")
	case errors.Is(err, model.ErrPeriodRequired):
		writeAvailabilityError(w, http.StatusBadRequest, "measurement period is required")
	case errors.Is(err, model.ErrInvalidOperationalMinutes):
		writeAvailabilityError(w, http.StatusBadRequest, "total operational minutes must be greater than zero")
	case errors.Is(err, service.ErrSPOFRecordNotFound):
		writeAvailabilityError(w, http.StatusNotFound, "single point of failure record not found")
	case errors.Is(err, model.ErrSPOFComponentNameRequired):
		writeAvailabilityError(w, http.StatusBadRequest, "SPOF component name is required")
	case errors.Is(err, model.ErrInvalidRedundancyMechanism):
		writeAvailabilityError(w, http.StatusBadRequest, "invalid redundancy mechanism")
	case errors.Is(err, model.ErrInvalidSPOFRiskLevel):
		writeAvailabilityError(w, http.StatusBadRequest, "invalid SPOF risk level")
	case errors.Is(err, model.ErrInvalidSPOFStatus):
		writeAvailabilityError(w, http.StatusBadRequest, "invalid SPOF status")
	case errors.Is(err, service.ErrAvailabilityTestNotFound):
		writeAvailabilityError(w, http.StatusNotFound, "availability test not found")
	case errors.Is(err, model.ErrAvailabilityTestNameRequired):
		writeAvailabilityError(w, http.StatusBadRequest, "availability test name is required")
	case errors.Is(err, model.ErrInvalidTestType):
		writeAvailabilityError(w, http.StatusBadRequest, "invalid availability test type")
	case errors.Is(err, model.ErrInvalidTestFrequency):
		writeAvailabilityError(w, http.StatusBadRequest, "invalid availability test frequency")
	case errors.Is(err, model.ErrInvalidTestOutcome):
		writeAvailabilityError(w, http.StatusBadRequest, "invalid availability test outcome")
	default:
		writeAvailabilityError(w, http.StatusInternalServerError, err.Error())
	}
}

// Target Endpoints

// ListTargets handles GET /v2/availability/targets
func (h *AvailabilityHandler) ListTargets(w http.ResponseWriter, r *http.Request) {
	wsID := getAvailabilityWorkspaceID(r)
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

	filter := model.AvailabilityFilterCriteria{
		WorkspaceID: wsID,
		Status:      q.Get("status"),
		CIID:        q.Get("ci_id"),
		Limit:       limit,
		Offset:      offset,
	}

	targets, total, err := h.svc.ListTargets(r.Context(), filter)
	if err != nil {
		handleAvailabilityError(w, err)
		return
	}

	writeAvailabilityJSON(w, http.StatusOK, map[string]interface{}{
		"targets": targets,
		"total":   total,
		"limit":   limit,
		"offset":  offset,
	})
}

// CreateTarget handles POST /v2/availability/targets
func (h *AvailabilityHandler) CreateTarget(w http.ResponseWriter, r *http.Request) {
	wsID := getAvailabilityWorkspaceID(r)

	var req model.CreateAvailabilityTargetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeAvailabilityError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	target, err := h.svc.CreateTarget(r.Context(), wsID, req)
	if err != nil {
		handleAvailabilityError(w, err)
		return
	}

	writeAvailabilityJSON(w, http.StatusCreated, target)
}

// GetTargetByID handles GET /v2/availability/targets/:id
func (h *AvailabilityHandler) GetTargetByID(w http.ResponseWriter, r *http.Request, id string) {
	target, err := h.svc.GetTargetByID(r.Context(), id)
	if err != nil {
		handleAvailabilityError(w, err)
		return
	}
	writeAvailabilityJSON(w, http.StatusOK, target)
}

// UpdateTarget handles PUT /v2/availability/targets/:id
func (h *AvailabilityHandler) UpdateTarget(w http.ResponseWriter, r *http.Request, id string) {
	var req model.UpdateAvailabilityTargetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeAvailabilityError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	target, err := h.svc.UpdateTarget(r.Context(), id, req)
	if err != nil {
		handleAvailabilityError(w, err)
		return
	}
	writeAvailabilityJSON(w, http.StatusOK, target)
}

// DeleteTarget handles DELETE /v2/availability/targets/:id
func (h *AvailabilityHandler) DeleteTarget(w http.ResponseWriter, r *http.Request, id string) {
	if err := h.svc.DeleteTarget(r.Context(), id); err != nil {
		handleAvailabilityError(w, err)
		return
	}
	writeAvailabilityJSON(w, http.StatusOK, map[string]string{"message": "availability target deleted successfully"})
}

// Outage Endpoints

// ListOutages handles GET /v2/availability/outages
func (h *AvailabilityHandler) ListOutages(w http.ResponseWriter, r *http.Request) {
	wsID := getAvailabilityWorkspaceID(r)
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

	filter := model.AvailabilityFilterCriteria{
		WorkspaceID: wsID,
		TargetID:    q.Get("target_id"),
		CIID:        q.Get("ci_id"),
		OutageType:  q.Get("outage_type"),
		Severity:    q.Get("severity"),
		Limit:       limit,
		Offset:      offset,
	}

	outages, total, err := h.svc.ListOutages(r.Context(), filter)
	if err != nil {
		handleAvailabilityError(w, err)
		return
	}

	writeAvailabilityJSON(w, http.StatusOK, map[string]interface{}{
		"outages": outages,
		"total":   total,
		"limit":   limit,
		"offset":  offset,
	})
}

// CreateOutage handles POST /v2/availability/outages
func (h *AvailabilityHandler) CreateOutage(w http.ResponseWriter, r *http.Request) {
	wsID := getAvailabilityWorkspaceID(r)

	var req model.CreateOutageEventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeAvailabilityError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	outage, err := h.svc.RecordOutage(r.Context(), wsID, req)
	if err != nil {
		handleAvailabilityError(w, err)
		return
	}

	writeAvailabilityJSON(w, http.StatusCreated, outage)
}

// GetOutageByID handles GET /v2/availability/outages/:id
func (h *AvailabilityHandler) GetOutageByID(w http.ResponseWriter, r *http.Request, id string) {
	outage, err := h.svc.GetOutageByID(r.Context(), id)
	if err != nil {
		handleAvailabilityError(w, err)
		return
	}
	writeAvailabilityJSON(w, http.StatusOK, outage)
}

// UpdateOutage handles PUT /v2/availability/outages/:id
func (h *AvailabilityHandler) UpdateOutage(w http.ResponseWriter, r *http.Request, id string) {
	var req model.UpdateOutageEventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeAvailabilityError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	outage, err := h.svc.UpdateOutage(r.Context(), id, req)
	if err != nil {
		handleAvailabilityError(w, err)
		return
	}
	writeAvailabilityJSON(w, http.StatusOK, outage)
}

// ResolveOutage handles POST /v2/availability/outages/:id/resolve
func (h *AvailabilityHandler) ResolveOutage(w http.ResponseWriter, r *http.Request, id string) {
	var req model.ResolveOutageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeAvailabilityError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	outage, err := h.svc.ResolveOutage(r.Context(), id, req)
	if err != nil {
		handleAvailabilityError(w, err)
		return
	}
	writeAvailabilityJSON(w, http.StatusOK, outage)
}

// DeleteOutage handles DELETE /v2/availability/outages/:id
func (h *AvailabilityHandler) DeleteOutage(w http.ResponseWriter, r *http.Request, id string) {
	if err := h.svc.DeleteOutage(r.Context(), id); err != nil {
		handleAvailabilityError(w, err)
		return
	}
	writeAvailabilityJSON(w, http.StatusOK, map[string]string{"message": "outage event deleted successfully"})
}

// Measurement Endpoints

// ListMeasurements handles GET /v2/availability/measurements
func (h *AvailabilityHandler) ListMeasurements(w http.ResponseWriter, r *http.Request) {
	wsID := getAvailabilityWorkspaceID(r)
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

	filter := model.AvailabilityFilterCriteria{
		WorkspaceID: wsID,
		TargetID:    q.Get("target_id"),
		Period:      q.Get("period"),
		Limit:       limit,
		Offset:      offset,
	}

	measurements, total, err := h.svc.ListMeasurements(r.Context(), filter)
	if err != nil {
		handleAvailabilityError(w, err)
		return
	}

	writeAvailabilityJSON(w, http.StatusOK, map[string]interface{}{
		"measurements": measurements,
		"total":        total,
		"limit":        limit,
		"offset":       offset,
	})
}

// RecordMeasurement handles POST /v2/availability/measurements
func (h *AvailabilityHandler) RecordMeasurement(w http.ResponseWriter, r *http.Request) {
	wsID := getAvailabilityWorkspaceID(r)

	var req model.CreateMeasurementRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeAvailabilityError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	m, err := h.svc.RecordMeasurement(r.Context(), wsID, req)
	if err != nil {
		handleAvailabilityError(w, err)
		return
	}

	writeAvailabilityJSON(w, http.StatusCreated, m)
}

// CalculateMeasurement handles POST /v2/availability/measurements/calculate
func (h *AvailabilityHandler) CalculateMeasurement(w http.ResponseWriter, r *http.Request) {
	wsID := getAvailabilityWorkspaceID(r)

	var req struct {
		TargetID                string  `json:"target_id"`
		Period                  string  `json:"period"`
		TotalOperationalMinutes float64 `json:"total_operational_minutes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeAvailabilityError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	m, err := h.svc.CalculatePeriodMeasurement(r.Context(), wsID, req.TargetID, req.Period, req.TotalOperationalMinutes)
	if err != nil {
		handleAvailabilityError(w, err)
		return
	}

	writeAvailabilityJSON(w, http.StatusOK, m)
}

// GetMeasurementByID handles GET /v2/availability/measurements/:id
func (h *AvailabilityHandler) GetMeasurementByID(w http.ResponseWriter, r *http.Request, id string) {
	m, err := h.svc.GetMeasurementByID(r.Context(), id)
	if err != nil {
		handleAvailabilityError(w, err)
		return
	}
	writeAvailabilityJSON(w, http.StatusOK, m)
}
