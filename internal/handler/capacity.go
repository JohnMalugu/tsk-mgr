package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/JohnMalugu/tsk-mgr-api/internal/model"
	"github.com/JohnMalugu/tsk-mgr-api/internal/service"
)

var _ = strings.TrimPrefix

// CapacityHandler exposes HTTP endpoints for ITIL Capacity and Performance Management
type CapacityHandler struct {
	svc service.CapacityService
}

// GlobalCapacityHandler is the singleton handler instance
var GlobalCapacityHandler *CapacityHandler

// InitCapacityHandler initializes the global capacity handler singleton
func InitCapacityHandler(svc service.CapacityService) {
	GlobalCapacityHandler = &CapacityHandler{svc: svc}
}

// NewCapacityHandler creates a new CapacityHandler instance
func NewCapacityHandler(svc service.CapacityService) *CapacityHandler {
	return &CapacityHandler{svc: svc}
}

func getCapacityUser(r *http.Request) string {
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

func getCapacityWorkspaceID(r *http.Request) string {
	if ws := r.URL.Query().Get("workspace_id"); ws != "" {
		return ws
	}
	if ws := r.Header.Get("X-Workspace-ID"); ws != "" {
		return ws
	}
	return "ws_1"
}

func writeCapacityJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func writeCapacityError(w http.ResponseWriter, status int, msg string) {
	writeCapacityJSON(w, status, map[string]string{"error": msg})
}

func handleCapacityError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrCapacityPoolNotFound):
		writeCapacityError(w, http.StatusNotFound, "capacity pool not found")
	case errors.Is(err, service.ErrCapacityPoolNameRequired):
		writeCapacityError(w, http.StatusBadRequest, "capacity pool name is required")
	case errors.Is(err, service.ErrCapacityPoolNameConflict):
		writeCapacityError(w, http.StatusConflict, "capacity pool with this name already exists in workspace")
	case errors.Is(err, service.ErrInvalidResourceType):
		writeCapacityError(w, http.StatusBadRequest, "invalid resource type")
	case errors.Is(err, service.ErrInvalidPoolStatus):
		writeCapacityError(w, http.StatusBadRequest, "invalid pool status")
	case errors.Is(err, service.ErrInvalidTotalCapacity):
		writeCapacityError(w, http.StatusBadRequest, "total capacity must be greater than zero")
	case errors.Is(err, service.ErrInvalidThresholdPercent):
		writeCapacityError(w, http.StatusBadRequest, "threshold percent must be between 1 and 100")
	case errors.Is(err, service.ErrCapacityMetricNotFound):
		writeCapacityError(w, http.StatusNotFound, "capacity metric not found")
	case errors.Is(err, service.ErrCapacityPlanNotFound):
		writeCapacityError(w, http.StatusNotFound, "capacity plan not found")
	case errors.Is(err, service.ErrInvalidPlanHorizon):
		writeCapacityError(w, http.StatusBadRequest, "invalid capacity plan horizon")
	case errors.Is(err, service.ErrInvalidPlanStatus):
		writeCapacityError(w, http.StatusBadRequest, "invalid capacity plan status")
	case errors.Is(err, service.ErrBottleneckNotFound):
		writeCapacityError(w, http.StatusNotFound, "bottleneck record not found")
	case errors.Is(err, service.ErrInvalidBottleneckSeverity):
		writeCapacityError(w, http.StatusBadRequest, "invalid bottleneck severity")
	case errors.Is(err, service.ErrInvalidBottleneckStatus):
		writeCapacityError(w, http.StatusBadRequest, "invalid bottleneck status")
	default:
		writeCapacityError(w, http.StatusInternalServerError, err.Error())
	}
}

// Pool HTTP Handlers

func (h *CapacityHandler) ListPools(w http.ResponseWriter, r *http.Request) {
	wsID := getCapacityWorkspaceID(r)
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

	filter := model.CapacityPoolFilterCriteria{
		WorkspaceID:  wsID,
		ResourceType: q.Get("resource_type"),
		Status:       q.Get("status"),
		CIID:         q.Get("ci_id"),
		Search:       q.Get("search"),
		Limit:        limit,
		Offset:       offset,
	}

	list, total, err := h.svc.ListPools(r.Context(), filter)
	if err != nil {
		handleCapacityError(w, err)
		return
	}

	writeCapacityJSON(w, http.StatusOK, map[string]interface{}{
		"pools":  list,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

func (h *CapacityHandler) CreatePool(w http.ResponseWriter, r *http.Request) {
	wsID := getCapacityWorkspaceID(r)

	var req model.CreateCapacityPoolRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeCapacityError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	pool, err := h.svc.CreatePool(r.Context(), wsID, req)
	if err != nil {
		handleCapacityError(w, err)
		return
	}

	writeCapacityJSON(w, http.StatusCreated, pool)
}

func (h *CapacityHandler) GetPoolByID(w http.ResponseWriter, r *http.Request, id string) {
	pool, err := h.svc.GetPoolByID(r.Context(), id)
	if err != nil {
		handleCapacityError(w, err)
		return
	}

	writeCapacityJSON(w, http.StatusOK, pool)
}

func (h *CapacityHandler) UpdatePool(w http.ResponseWriter, r *http.Request, id string) {
	var req model.UpdateCapacityPoolRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeCapacityError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	pool, err := h.svc.UpdatePool(r.Context(), id, req)
	if err != nil {
		handleCapacityError(w, err)
		return
	}

	writeCapacityJSON(w, http.StatusOK, pool)
}

func (h *CapacityHandler) DeletePool(w http.ResponseWriter, r *http.Request, id string) {
	if err := h.svc.DeletePool(r.Context(), id); err != nil {
		handleCapacityError(w, err)
		return
	}

	writeCapacityJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// Telemetry & Metrics HTTP Handlers

func (h *CapacityHandler) RecordMetric(w http.ResponseWriter, r *http.Request) {
	wsID := getCapacityWorkspaceID(r)

	var req model.RecordCapacityMetricRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeCapacityError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	metric, err := h.svc.RecordMetric(r.Context(), wsID, req)
	if err != nil {
		handleCapacityError(w, err)
		return
	}

	writeCapacityJSON(w, http.StatusCreated, metric)
}

func (h *CapacityHandler) GetMetricByID(w http.ResponseWriter, r *http.Request, id string) {
	metric, err := h.svc.GetMetricByID(r.Context(), id)
	if err != nil {
		handleCapacityError(w, err)
		return
	}

	writeCapacityJSON(w, http.StatusOK, metric)
}

func (h *CapacityHandler) ListMetrics(w http.ResponseWriter, r *http.Request) {
	wsID := getCapacityWorkspaceID(r)
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

	filter := model.CapacityMetricFilterCriteria{
		WorkspaceID: wsID,
		PoolID:      q.Get("pool_id"),
		MetricName:  q.Get("metric_name"),
		Limit:       limit,
		Offset:      offset,
	}

	if a := q.Get("is_anomaly"); a != "" {
		val := (a == "true" || a == "1")
		filter.IsAnomaly = &val
	}
	if fd := q.Get("from_date"); fd != "" {
		if t, err := time.Parse(time.RFC3339, fd); err == nil {
			filter.FromDate = &t
		} else if t, err := time.Parse("2006-01-02", fd); err == nil {
			filter.FromDate = &t
		}
	}
	if td := q.Get("to_date"); td != "" {
		if t, err := time.Parse(time.RFC3339, td); err == nil {
			filter.ToDate = &t
		} else if t, err := time.Parse("2006-01-02", td); err == nil {
			filter.ToDate = &t
		}
	}

	list, total, err := h.svc.ListMetrics(r.Context(), filter)
	if err != nil {
		handleCapacityError(w, err)
		return
	}

	writeCapacityJSON(w, http.StatusOK, map[string]interface{}{
		"metrics": list,
		"total":   total,
		"limit":   limit,
		"offset":  offset,
	})
}

// Forecasting HTTP Handlers

func (h *CapacityHandler) GenerateForecast(w http.ResponseWriter, r *http.Request) {
	wsID := getCapacityWorkspaceID(r)

	var req struct {
		PoolID      string `json:"pool_id"`
		HorizonDays int    `json:"horizon_days"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeCapacityError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	forecast, err := h.svc.GenerateForecast(r.Context(), wsID, req.PoolID, req.HorizonDays)
	if err != nil {
		handleCapacityError(w, err)
		return
	}

	writeCapacityJSON(w, http.StatusCreated, forecast)
}

func (h *CapacityHandler) GetLatestForecast(w http.ResponseWriter, r *http.Request, poolID string) {
	forecast, err := h.svc.GetLatestForecast(r.Context(), poolID)
	if err != nil {
		handleCapacityError(w, err)
		return
	}
	if forecast == nil {
		writeCapacityError(w, http.StatusNotFound, "no forecast found for capacity pool")
		return
	}

	writeCapacityJSON(w, http.StatusOK, forecast)
}
