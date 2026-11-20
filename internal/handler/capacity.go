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

// Capacity Plan HTTP Handlers

func (h *CapacityHandler) ListPlans(w http.ResponseWriter, r *http.Request) {
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

	filter := model.CapacityPlanFilterCriteria{
		WorkspaceID:     wsID,
		PoolID:          q.Get("pool_id"),
		PlanningHorizon: q.Get("planning_horizon"),
		Status:          q.Get("status"),
		Limit:           limit,
		Offset:          offset,
	}

	list, total, err := h.svc.ListPlans(r.Context(), filter)
	if err != nil {
		handleCapacityError(w, err)
		return
	}

	writeCapacityJSON(w, http.StatusOK, map[string]interface{}{
		"plans":  list,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

func (h *CapacityHandler) CreatePlan(w http.ResponseWriter, r *http.Request) {
	wsID := getCapacityWorkspaceID(r)

	var req model.CreateCapacityPlanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeCapacityError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	plan, err := h.svc.CreatePlan(r.Context(), wsID, req)
	if err != nil {
		handleCapacityError(w, err)
		return
	}

	writeCapacityJSON(w, http.StatusCreated, plan)
}

func (h *CapacityHandler) GetPlanByID(w http.ResponseWriter, r *http.Request, id string) {
	plan, err := h.svc.GetPlanByID(r.Context(), id)
	if err != nil {
		handleCapacityError(w, err)
		return
	}

	writeCapacityJSON(w, http.StatusOK, plan)
}

func (h *CapacityHandler) UpdatePlan(w http.ResponseWriter, r *http.Request, id string) {
	var req model.UpdateCapacityPlanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeCapacityError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	plan, err := h.svc.UpdatePlan(r.Context(), id, req)
	if err != nil {
		handleCapacityError(w, err)
		return
	}

	writeCapacityJSON(w, http.StatusOK, plan)
}

func (h *CapacityHandler) ApprovePlan(w http.ResponseWriter, r *http.Request, id string) {
	approverID := getCapacityUser(r)

	plan, err := h.svc.ApprovePlan(r.Context(), id, approverID)
	if err != nil {
		handleCapacityError(w, err)
		return
	}

	writeCapacityJSON(w, http.StatusOK, plan)
}

func (h *CapacityHandler) DeletePlan(w http.ResponseWriter, r *http.Request, id string) {
	if err := h.svc.DeletePlan(r.Context(), id); err != nil {
		handleCapacityError(w, err)
		return
	}

	writeCapacityJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// Bottleneck HTTP Handlers

func (h *CapacityHandler) ListBottlenecks(w http.ResponseWriter, r *http.Request) {
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

	filter := model.BottleneckFilterCriteria{
		WorkspaceID:  wsID,
		PoolID:       q.Get("pool_id"),
		AffectedCIID: q.Get("affected_ci_id"),
		Severity:     q.Get("severity"),
		Status:       q.Get("status"),
		Limit:        limit,
		Offset:       offset,
	}

	list, total, err := h.svc.ListBottlenecks(r.Context(), filter)
	if err != nil {
		handleCapacityError(w, err)
		return
	}

	writeCapacityJSON(w, http.StatusOK, map[string]interface{}{
		"bottlenecks": list,
		"total":       total,
		"limit":       limit,
		"offset":      offset,
	})
}

func (h *CapacityHandler) CreateBottleneck(w http.ResponseWriter, r *http.Request) {
	wsID := getCapacityWorkspaceID(r)

	var req model.CreateBottleneckRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeCapacityError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	b, err := h.svc.CreateBottleneck(r.Context(), wsID, req)
	if err != nil {
		handleCapacityError(w, err)
		return
	}

	writeCapacityJSON(w, http.StatusCreated, b)
}

func (h *CapacityHandler) GetBottleneckByID(w http.ResponseWriter, r *http.Request, id string) {
	b, err := h.svc.GetBottleneckByID(r.Context(), id)
	if err != nil {
		handleCapacityError(w, err)
		return
	}

	writeCapacityJSON(w, http.StatusOK, b)
}

func (h *CapacityHandler) UpdateBottleneck(w http.ResponseWriter, r *http.Request, id string) {
	var req model.UpdateBottleneckRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeCapacityError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	b, err := h.svc.UpdateBottleneck(r.Context(), id, req)
	if err != nil {
		handleCapacityError(w, err)
		return
	}

	writeCapacityJSON(w, http.StatusOK, b)
}

func (h *CapacityHandler) ResolveBottleneck(w http.ResponseWriter, r *http.Request, id string) {
	var req struct {
		MitigationPlan string `json:"mitigation_plan"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)

	b, err := h.svc.ResolveBottleneck(r.Context(), id, req.MitigationPlan)
	if err != nil {
		handleCapacityError(w, err)
		return
	}

	writeCapacityJSON(w, http.StatusOK, b)
}

func (h *CapacityHandler) DeleteBottleneck(w http.ResponseWriter, r *http.Request, id string) {
	if err := h.svc.DeleteBottleneck(r.Context(), id); err != nil {
		handleCapacityError(w, err)
		return
	}

	writeCapacityJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// Analytics HTTP Handler

func (h *CapacityHandler) GetCapacityAnalytics(w http.ResponseWriter, r *http.Request) {
	wsID := getCapacityWorkspaceID(r)

	report, err := h.svc.GetCapacityAnalytics(r.Context(), wsID)
	if err != nil {
		handleCapacityError(w, err)
		return
	}

	writeCapacityJSON(w, http.StatusOK, report)
}

// ServeHTTP dispatches requests for /v2/capacity
func (h *CapacityHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/v2/capacity")
	path = strings.TrimPrefix(path, "/")

	if path == "analytics" {
		if r.Method == http.MethodGet {
			h.GetCapacityAnalytics(w, r)
		} else {
			writeCapacityError(w, http.StatusMethodNotAllowed, "method not allowed")
		}
		return
	}

	if strings.HasPrefix(path, "forecasts") {
		sub := strings.TrimPrefix(path, "forecasts")
		sub = strings.TrimPrefix(sub, "/")
		if sub == "" {
			if r.Method == http.MethodPost {
				h.GenerateForecast(w, r)
			} else {
				writeCapacityError(w, http.StatusMethodNotAllowed, "method not allowed")
			}
			return
		}
		if r.Method == http.MethodGet {
			h.GetLatestForecast(w, r, sub)
		} else {
			writeCapacityError(w, http.StatusMethodNotAllowed, "method not allowed")
		}
		return
	}

	if strings.HasPrefix(path, "metrics") {
		sub := strings.TrimPrefix(path, "metrics")
		sub = strings.TrimPrefix(sub, "/")
		if sub == "" {
			switch r.Method {
			case http.MethodGet:
				h.ListMetrics(w, r)
			case http.MethodPost:
				h.RecordMetric(w, r)
			default:
				writeCapacityError(w, http.StatusMethodNotAllowed, "method not allowed")
			}
			return
		}
		if r.Method == http.MethodGet {
			h.GetMetricByID(w, r, sub)
		} else {
			writeCapacityError(w, http.StatusMethodNotAllowed, "method not allowed")
		}
		return
	}

	if strings.HasPrefix(path, "plans") {
		sub := strings.TrimPrefix(path, "plans")
		sub = strings.TrimPrefix(sub, "/")
		if sub == "" {
			switch r.Method {
			case http.MethodGet:
				h.ListPlans(w, r)
			case http.MethodPost:
				h.CreatePlan(w, r)
			default:
				writeCapacityError(w, http.StatusMethodNotAllowed, "method not allowed")
			}
			return
		}
		parts := strings.Split(sub, "/")
		planID := parts[0]
		if len(parts) == 1 {
			switch r.Method {
			case http.MethodGet:
				h.GetPlanByID(w, r, planID)
			case http.MethodPut:
				h.UpdatePlan(w, r, planID)
			case http.MethodDelete:
				h.DeletePlan(w, r, planID)
			default:
				writeCapacityError(w, http.StatusMethodNotAllowed, "method not allowed")
			}
			return
		}
		if len(parts) == 2 && parts[1] == "approve" && r.Method == http.MethodPost {
			h.ApprovePlan(w, r, planID)
			return
		}
	}

	if strings.HasPrefix(path, "bottlenecks") {
		sub := strings.TrimPrefix(path, "bottlenecks")
		sub = strings.TrimPrefix(sub, "/")
		if sub == "" {
			switch r.Method {
			case http.MethodGet:
				h.ListBottlenecks(w, r)
			case http.MethodPost:
				h.CreateBottleneck(w, r)
			default:
				writeCapacityError(w, http.StatusMethodNotAllowed, "method not allowed")
			}
			return
		}
		parts := strings.Split(sub, "/")
		bID := parts[0]
		if len(parts) == 1 {
			switch r.Method {
			case http.MethodGet:
				h.GetBottleneckByID(w, r, bID)
			case http.MethodPut:
				h.UpdateBottleneck(w, r, bID)
			case http.MethodDelete:
				h.DeleteBottleneck(w, r, bID)
			default:
				writeCapacityError(w, http.StatusMethodNotAllowed, "method not allowed")
			}
			return
		}
		if len(parts) == 2 && parts[1] == "resolve" && r.Method == http.MethodPost {
			h.ResolveBottleneck(w, r, bID)
			return
		}
	}

	if strings.HasPrefix(path, "pools") {
		sub := strings.TrimPrefix(path, "pools")
		sub = strings.TrimPrefix(sub, "/")
		if sub == "" {
			switch r.Method {
			case http.MethodGet:
				h.ListPools(w, r)
			case http.MethodPost:
				h.CreatePool(w, r)
			default:
				writeCapacityError(w, http.StatusMethodNotAllowed, "method not allowed")
			}
			return
		}
		poolID := sub
		switch r.Method {
		case http.MethodGet:
			h.GetPoolByID(w, r, poolID)
		case http.MethodPut:
			h.UpdatePool(w, r, poolID)
		case http.MethodDelete:
			h.DeletePool(w, r, poolID)
		default:
			writeCapacityError(w, http.StatusMethodNotAllowed, "method not allowed")
		}
		return
	}

	writeCapacityError(w, http.StatusNotFound, "route not found")
}
