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

// ITAssetHandler exposes HTTP endpoints for IT asset management and software licenses
type ITAssetHandler struct {
	svc service.ITAssetService
}

var GlobalITAssetHandler *ITAssetHandler

// InitITAssetHandler initializes the global handler singleton
func InitITAssetHandler(svc service.ITAssetService) {
	GlobalITAssetHandler = &ITAssetHandler{svc: svc}
}

// NewITAssetHandler creates a new ITAssetHandler instance
func NewITAssetHandler(svc service.ITAssetService) *ITAssetHandler {
	return &ITAssetHandler{svc: svc}
}

func getITAssetUser(r *http.Request) string {
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

func writeITAssetJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func writeITAssetError(w http.ResponseWriter, status int, msg string) {
	writeITAssetJSON(w, status, map[string]string{"error": msg})
}

func handleITAssetError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrAssetNotFound):
		writeITAssetError(w, http.StatusNotFound, "asset not found")
	case errors.Is(err, service.ErrAssetNameRequired):
		writeITAssetError(w, http.StatusBadRequest, "asset name is required")
	case errors.Is(err, service.ErrAssetTagRequired):
		writeITAssetError(w, http.StatusBadRequest, "asset tag is required")
	case errors.Is(err, service.ErrInvalidAssetCategory):
		writeITAssetError(w, http.StatusBadRequest, "invalid asset category")
	case errors.Is(err, service.ErrInvalidAssetStatus):
		writeITAssetError(w, http.StatusBadRequest, "invalid asset status")
	case errors.Is(err, service.ErrInvalidDepreciation):
		writeITAssetError(w, http.StatusBadRequest, "invalid depreciation method")
	case errors.Is(err, service.ErrAssetTagConflict):
		writeITAssetError(w, http.StatusConflict, "asset tag already exists in workspace")
	case errors.Is(err, service.ErrAssetAlreadyAssigned):
		writeITAssetError(w, http.StatusConflict, "asset is already assigned")
	case errors.Is(err, service.ErrAssetNotAssigned):
		writeITAssetError(w, http.StatusBadRequest, "asset is not currently assigned")
	case errors.Is(err, service.ErrAssetRetired):
		writeITAssetError(w, http.StatusConflict, "cannot perform operation on retired or disposed asset")
	case errors.Is(err, service.ErrLicenseNotFound):
		writeITAssetError(w, http.StatusNotFound, "software license not found")
	case errors.Is(err, service.ErrLicenseNameRequired):
		writeITAssetError(w, http.StatusBadRequest, "software name is required")
	case errors.Is(err, service.ErrInvalidLicenseType):
		writeITAssetError(w, http.StatusBadRequest, "invalid software license type")
	case errors.Is(err, service.ErrNoSeatsAvailable):
		writeITAssetError(w, http.StatusConflict, "no software license seats available")
	case errors.Is(err, service.ErrLicenseAllocNotFound):
		writeITAssetError(w, http.StatusNotFound, "license allocation not found")
	case errors.Is(err, service.ErrAlreadyAllocated):
		writeITAssetError(w, http.StatusConflict, "license seat already allocated to this target")
	case errors.Is(err, service.ErrAllocationTargetReq):
		writeITAssetError(w, http.StatusBadRequest, "allocation target (user or asset) is required")
	case errors.Is(err, service.ErrMaintenanceRecordNotFound):
		writeITAssetError(w, http.StatusNotFound, "asset maintenance record not found")
	case errors.Is(err, service.ErrInvalidMaintenanceType):
		writeITAssetError(w, http.StatusBadRequest, "invalid maintenance type")
	case errors.Is(err, service.ErrInvalidMaintRecordStatus):
		writeITAssetError(w, http.StatusBadRequest, "invalid maintenance record status")
	case errors.Is(err, service.ErrAuditScanNotFound):
		writeITAssetError(w, http.StatusNotFound, "asset audit record not found")
	case errors.Is(err, service.ErrInvalidAuditStatus):
		writeITAssetError(w, http.StatusBadRequest, "invalid asset audit status")
	case errors.Is(err, service.ErrAuditScanAssetRequired):
		writeITAssetError(w, http.StatusBadRequest, "asset ID or tag is required for audit scan")
	default:
		writeITAssetError(w, http.StatusInternalServerError, err.Error())
	}
}

// CreateAsset handles POST /v2/assets
func (h *ITAssetHandler) CreateAsset(w http.ResponseWriter, r *http.Request) {
	var req model.CreateAssetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeITAssetError(w, http.StatusBadRequest, "invalid JSON payload")
		return
	}
	if req.WorkspaceID <= 0 {
		req.WorkspaceID = 1
	}

	asset, err := h.svc.CreateAsset(r.Context(), &req)
	if err != nil {
		handleITAssetError(w, err)
		return
	}
	writeITAssetJSON(w, http.StatusCreated, asset)
}

// GetAsset handles GET /v2/assets/{id}
func (h *ITAssetHandler) GetAsset(w http.ResponseWriter, r *http.Request, id int64) {
	asset, err := h.svc.GetAssetByID(r.Context(), id)
	if err != nil {
		handleITAssetError(w, err)
		return
	}
	writeITAssetJSON(w, http.StatusOK, asset)
}

// GetAssetByTag handles GET /v2/assets/tags/{tag}
func (h *ITAssetHandler) GetAssetByTag(w http.ResponseWriter, r *http.Request, workspaceID int64, tag string) {
	if workspaceID <= 0 {
		workspaceID = 1
	}
	asset, err := h.svc.GetAssetByTag(r.Context(), workspaceID, tag)
	if err != nil {
		handleITAssetError(w, err)
		return
	}
	writeITAssetJSON(w, http.StatusOK, asset)
}

// UpdateAsset handles PUT /v2/assets/{id}
func (h *ITAssetHandler) UpdateAsset(w http.ResponseWriter, r *http.Request, id int64) {
	var req model.UpdateAssetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeITAssetError(w, http.StatusBadRequest, "invalid JSON payload")
		return
	}

	asset, err := h.svc.UpdateAsset(r.Context(), id, &req)
	if err != nil {
		handleITAssetError(w, err)
		return
	}
	writeITAssetJSON(w, http.StatusOK, asset)
}

// DeleteAsset handles DELETE /v2/assets/{id}
func (h *ITAssetHandler) DeleteAsset(w http.ResponseWriter, r *http.Request, id int64) {
	if err := h.svc.DeleteAsset(r.Context(), id); err != nil {
		handleITAssetError(w, err)
		return
	}
	writeITAssetJSON(w, http.StatusOK, map[string]string{"message": "asset deleted successfully"})
}

// ListAssets handles GET /v2/assets
func (h *ITAssetHandler) ListAssets(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	wsID, _ := strconv.ParseInt(query.Get("workspace_id"), 10, 64)
	if wsID <= 0 {
		wsID = 1
	}
	limit, _ := strconv.Atoi(query.Get("limit"))
	offset, _ := strconv.Atoi(query.Get("offset"))

	filter := model.AssetFilterCriteria{
		WorkspaceID:      wsID,
		Category:         query.Get("category"),
		Status:           query.Get("status"),
		AssignedToUserID: query.Get("assigned_to_user_id"),
		Location:         query.Get("location"),
		Search:           query.Get("search"),
		Limit:            limit,
		Offset:           offset,
	}

	if expDaysStr := query.Get("warranty_expiring_days"); expDaysStr != "" {
		if d, err := strconv.Atoi(expDaysStr); err == nil && d > 0 {
			filter.WarrantyExpiringDays = &d
		}
	}

	assets, total, err := h.svc.ListAssets(r.Context(), filter)
	if err != nil {
		handleITAssetError(w, err)
		return
	}
	writeITAssetJSON(w, http.StatusOK, map[string]interface{}{
		"assets": assets,
		"total":  total,
	})
}

// CalculateDepreciation handles GET /v2/assets/{id}/depreciation
func (h *ITAssetHandler) CalculateDepreciation(w http.ResponseWriter, r *http.Request, id int64) {
	sched, err := h.svc.CalculateDepreciation(r.Context(), id)
	if err != nil {
		handleITAssetError(w, err)
		return
	}
	writeITAssetJSON(w, http.StatusOK, sched)
}

// AssignAsset handles POST /v2/assets/{id}/assign
func (h *ITAssetHandler) AssignAsset(w http.ResponseWriter, r *http.Request, id int64) {
	var req model.AssignAssetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeITAssetError(w, http.StatusBadRequest, "invalid JSON payload")
		return
	}

	asset, err := h.svc.AssignAsset(r.Context(), id, &req)
	if err != nil {
		handleITAssetError(w, err)
		return
	}
	writeITAssetJSON(w, http.StatusOK, asset)
}

// CheckinAsset handles POST /v2/assets/{id}/checkin
func (h *ITAssetHandler) CheckinAsset(w http.ResponseWriter, r *http.Request, id int64) {
	var body struct {
		Location string `json:"location"`
		Notes    string `json:"notes"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)

	asset, err := h.svc.CheckinAsset(r.Context(), id, body.Location, body.Notes)
	if err != nil {
		handleITAssetError(w, err)
		return
	}
	writeITAssetJSON(w, http.StatusOK, asset)
}

// TransitionStatus handles POST /v2/assets/{id}/status
func (h *ITAssetHandler) TransitionStatus(w http.ResponseWriter, r *http.Request, id int64) {
	var req model.AssetStatusTransitionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeITAssetError(w, http.StatusBadRequest, "invalid JSON payload")
		return
	}

	asset, err := h.svc.TransitionAssetStatus(r.Context(), id, &req)
	if err != nil {
		handleITAssetError(w, err)
		return
	}
	writeITAssetJSON(w, http.StatusOK, asset)
}

// CreateMaintenanceRecord handles POST /v2/assets/{id}/maintenance
func (h *ITAssetHandler) CreateMaintenanceRecord(w http.ResponseWriter, r *http.Request, assetID int64) {
	var req model.CreateMaintenanceRecordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeITAssetError(w, http.StatusBadRequest, "invalid JSON payload")
		return
	}

	record, err := h.svc.CreateMaintenanceRecord(r.Context(), assetID, &req)
	if err != nil {
		handleITAssetError(w, err)
		return
	}
	writeITAssetJSON(w, http.StatusCreated, record)
}

// GetMaintenanceRecord handles GET /v2/assets/maintenance/{id}
func (h *ITAssetHandler) GetMaintenanceRecord(w http.ResponseWriter, r *http.Request, id int64) {
	record, err := h.svc.GetMaintenanceRecordByID(r.Context(), id)
	if err != nil {
		handleITAssetError(w, err)
		return
	}
	writeITAssetJSON(w, http.StatusOK, record)
}

// UpdateMaintenanceRecord handles PUT /v2/assets/maintenance/{id}
func (h *ITAssetHandler) UpdateMaintenanceRecord(w http.ResponseWriter, r *http.Request, id int64) {
	var req model.UpdateMaintenanceRecordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeITAssetError(w, http.StatusBadRequest, "invalid JSON payload")
		return
	}

	record, err := h.svc.UpdateMaintenanceRecord(r.Context(), id, &req)
	if err != nil {
		handleITAssetError(w, err)
		return
	}
	writeITAssetJSON(w, http.StatusOK, record)
}

// DeleteMaintenanceRecord handles DELETE /v2/assets/maintenance/{id}
func (h *ITAssetHandler) DeleteMaintenanceRecord(w http.ResponseWriter, r *http.Request, id int64) {
	if err := h.svc.DeleteMaintenanceRecord(r.Context(), id); err != nil {
		handleITAssetError(w, err)
		return
	}
	writeITAssetJSON(w, http.StatusOK, map[string]string{"message": "maintenance record deleted successfully"})
}

// ListMaintenanceRecords handles GET /v2/assets/{id}/maintenance
func (h *ITAssetHandler) ListMaintenanceRecords(w http.ResponseWriter, r *http.Request, assetID int64) {
	records, err := h.svc.ListMaintenanceRecords(r.Context(), assetID)
	if err != nil {
		handleITAssetError(w, err)
		return
	}
	writeITAssetJSON(w, http.StatusOK, map[string]interface{}{
		"maintenance_records": records,
		"total":               len(records),
	})
}

// RecordAuditScan handles POST /v2/assets/audit
func (h *ITAssetHandler) RecordAuditScan(w http.ResponseWriter, r *http.Request) {
	var req model.AssetAuditScanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeITAssetError(w, http.StatusBadRequest, "invalid JSON payload")
		return
	}

	wsID, _ := strconv.ParseInt(r.URL.Query().Get("workspace_id"), 10, 64)
	if wsID <= 0 {
		wsID = 1
	}

	user := getITAssetUser(r)
	log, err := h.svc.RecordAuditScan(r.Context(), wsID, user, &req)
	if err != nil {
		handleITAssetError(w, err)
		return
	}
	writeITAssetJSON(w, http.StatusCreated, log)
}

// ListAuditLogs handles GET /v2/assets/audit/logs
func (h *ITAssetHandler) ListAuditLogs(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	wsID, _ := strconv.ParseInt(query.Get("workspace_id"), 10, 64)
	if wsID <= 0 {
		wsID = 1
	}
	limit, _ := strconv.Atoi(query.Get("limit"))
	offset, _ := strconv.Atoi(query.Get("offset"))

	var assetID *int64
	if aIDStr := query.Get("asset_id"); aIDStr != "" {
		if id, err := strconv.ParseInt(aIDStr, 10, 64); err == nil && id > 0 {
			assetID = &id
		}
	}

	logs, err := h.svc.ListAuditLogs(r.Context(), wsID, assetID, limit, offset)
	if err != nil {
		handleITAssetError(w, err)
		return
	}
	writeITAssetJSON(w, http.StatusOK, map[string]interface{}{
		"audit_logs": logs,
		"total":      len(logs),
	})
}

// CreateSoftwareLicense handles POST /v2/assets/licenses
func (h *ITAssetHandler) CreateSoftwareLicense(w http.ResponseWriter, r *http.Request) {
	var req model.CreateSoftwareLicenseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeITAssetError(w, http.StatusBadRequest, "invalid JSON payload")
		return
	}
	if req.WorkspaceID <= 0 {
		req.WorkspaceID = 1
	}

	license, err := h.svc.CreateSoftwareLicense(r.Context(), &req)
	if err != nil {
		handleITAssetError(w, err)
		return
	}
	writeITAssetJSON(w, http.StatusCreated, license)
}

// GetSoftwareLicense handles GET /v2/assets/licenses/{id}
func (h *ITAssetHandler) GetSoftwareLicense(w http.ResponseWriter, r *http.Request, id int64) {
	license, err := h.svc.GetSoftwareLicenseByID(r.Context(), id)
	if err != nil {
		handleITAssetError(w, err)
		return
	}
	writeITAssetJSON(w, http.StatusOK, license)
}

// UpdateSoftwareLicense handles PUT /v2/assets/licenses/{id}
func (h *ITAssetHandler) UpdateSoftwareLicense(w http.ResponseWriter, r *http.Request, id int64) {
	var req model.UpdateSoftwareLicenseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeITAssetError(w, http.StatusBadRequest, "invalid JSON payload")
		return
	}

	license, err := h.svc.UpdateSoftwareLicense(r.Context(), id, &req)
	if err != nil {
		handleITAssetError(w, err)
		return
	}
	writeITAssetJSON(w, http.StatusOK, license)
}

// DeleteSoftwareLicense handles DELETE /v2/assets/licenses/{id}
func (h *ITAssetHandler) DeleteSoftwareLicense(w http.ResponseWriter, r *http.Request, id int64) {
	if err := h.svc.DeleteSoftwareLicense(r.Context(), id); err != nil {
		handleITAssetError(w, err)
		return
	}
	writeITAssetJSON(w, http.StatusOK, map[string]string{"message": "software license deleted successfully"})
}

// ListSoftwareLicenses handles GET /v2/assets/licenses
func (h *ITAssetHandler) ListSoftwareLicenses(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	wsID, _ := strconv.ParseInt(query.Get("workspace_id"), 10, 64)
	if wsID <= 0 {
		wsID = 1
	}
	limit, _ := strconv.Atoi(query.Get("limit"))
	offset, _ := strconv.Atoi(query.Get("offset"))

	filter := model.LicenseFilterCriteria{
		WorkspaceID: wsID,
		LicenseType: query.Get("license_type"),
		Publisher:   query.Get("publisher"),
		Search:      query.Get("search"),
		Limit:       limit,
		Offset:      offset,
	}

	if expDaysStr := query.Get("expiring_days"); expDaysStr != "" {
		if d, err := strconv.Atoi(expDaysStr); err == nil && d > 0 {
			filter.ExpiringDays = &d
		}
	}

	licenses, total, err := h.svc.ListSoftwareLicenses(r.Context(), filter)
	if err != nil {
		handleITAssetError(w, err)
		return
	}
	writeITAssetJSON(w, http.StatusOK, map[string]interface{}{
		"licenses": licenses,
		"total":    total,
	})
}

// AllocateLicense handles POST /v2/assets/licenses/{id}/allocations
func (h *ITAssetHandler) AllocateLicense(w http.ResponseWriter, r *http.Request, licenseID int64) {
	var req model.AllocateLicenseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeITAssetError(w, http.StatusBadRequest, "invalid JSON payload")
		return
	}

	user := getITAssetUser(r)
	alloc, err := h.svc.AllocateLicense(r.Context(), licenseID, &req, user)
	if err != nil {
		handleITAssetError(w, err)
		return
	}
	writeITAssetJSON(w, http.StatusCreated, alloc)
}

// GetLicenseAllocation handles GET /v2/assets/licenses/allocations/{id}
func (h *ITAssetHandler) GetLicenseAllocation(w http.ResponseWriter, r *http.Request, id int64) {
	alloc, err := h.svc.GetLicenseAllocationByID(r.Context(), id)
	if err != nil {
		handleITAssetError(w, err)
		return
	}
	writeITAssetJSON(w, http.StatusOK, alloc)
}

// DeallocateLicense handles DELETE /v2/assets/licenses/allocations/{id}
func (h *ITAssetHandler) DeallocateLicense(w http.ResponseWriter, r *http.Request, id int64) {
	if err := h.svc.DeallocateLicense(r.Context(), id); err != nil {
		handleITAssetError(w, err)
		return
	}
	writeITAssetJSON(w, http.StatusOK, map[string]string{"message": "license seat deallocated successfully"})
}

// ListLicenseAllocations handles GET /v2/assets/licenses/{id}/allocations
func (h *ITAssetHandler) ListLicenseAllocations(w http.ResponseWriter, r *http.Request, licenseID int64) {
	allocations, err := h.svc.ListLicenseAllocations(r.Context(), licenseID)
	if err != nil {
		handleITAssetError(w, err)
		return
	}
	writeITAssetJSON(w, http.StatusOK, map[string]interface{}{
		"allocations": allocations,
		"total":       len(allocations),
	})
}

// GetAssetMetrics handles GET /v2/assets/metrics
func (h *ITAssetHandler) GetAssetMetrics(w http.ResponseWriter, r *http.Request) {
	wsID, _ := strconv.ParseInt(r.URL.Query().Get("workspace_id"), 10, 64)
	if wsID <= 0 {
		wsID = 1
	}

	metrics, err := h.svc.GetAssetMetrics(r.Context(), wsID)
	if err != nil {
		handleITAssetError(w, err)
		return
	}
	writeITAssetJSON(w, http.StatusOK, metrics)
}

// ServeHTTP implements http.Handler routing for all /v2/assets endpoints
func (h *ITAssetHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimSuffix(r.URL.Path, "/")

	// GET /v2/assets/metrics
	if path == "/v2/assets/metrics" && r.Method == http.MethodGet {
		h.GetAssetMetrics(w, r)
		return
	}

	// GET /v2/assets/audit/logs
	if path == "/v2/assets/audit/logs" && r.Method == http.MethodGet {
		h.ListAuditLogs(w, r)
		return
	}

	// POST /v2/assets/audit
	if path == "/v2/assets/audit" && r.Method == http.MethodPost {
		h.RecordAuditScan(w, r)
		return
	}

	// /v2/assets/maintenance/{id}
	if strings.HasPrefix(path, "/v2/assets/maintenance/") {
		idStr := strings.TrimPrefix(path, "/v2/assets/maintenance/")
		if id, err := strconv.ParseInt(idStr, 10, 64); err == nil && id > 0 {
			switch r.Method {
			case http.MethodGet:
				h.GetMaintenanceRecord(w, r, id)
				return
			case http.MethodPut:
				h.UpdateMaintenanceRecord(w, r, id)
				return
			case http.MethodDelete:
				h.DeleteMaintenanceRecord(w, r, id)
				return
			default:
				writeITAssetError(w, http.StatusMethodNotAllowed, "method not allowed")
				return
			}
		}
	}

	// /v2/assets/licenses/allocations/{id}
	if strings.HasPrefix(path, "/v2/assets/licenses/allocations/") {
		idStr := strings.TrimPrefix(path, "/v2/assets/licenses/allocations/")
		if id, err := strconv.ParseInt(idStr, 10, 64); err == nil && id > 0 {
			switch r.Method {
			case http.MethodGet:
				h.GetLicenseAllocation(w, r, id)
				return
			case http.MethodDelete:
				h.DeallocateLicense(w, r, id)
				return
			default:
				writeITAssetError(w, http.StatusMethodNotAllowed, "method not allowed")
				return
			}
		}
	}

	// /v2/assets/licenses/{id}/allocations
	if strings.HasPrefix(path, "/v2/assets/licenses/") && strings.HasSuffix(path, "/allocations") {
		trimmed := strings.TrimPrefix(path, "/v2/assets/licenses/")
		trimmed = strings.TrimSuffix(trimmed, "/allocations")
		if licID, err := strconv.ParseInt(trimmed, 10, 64); err == nil && licID > 0 {
			switch r.Method {
			case http.MethodGet:
				h.ListLicenseAllocations(w, r, licID)
				return
			case http.MethodPost:
				h.AllocateLicense(w, r, licID)
				return
			default:
				writeITAssetError(w, http.StatusMethodNotAllowed, "method not allowed")
				return
			}
		}
	}

	// /v2/assets/licenses/{id}
	if strings.HasPrefix(path, "/v2/assets/licenses/") {
		idStr := strings.TrimPrefix(path, "/v2/assets/licenses/")
		if id, err := strconv.ParseInt(idStr, 10, 64); err == nil && id > 0 {
			switch r.Method {
			case http.MethodGet:
				h.GetSoftwareLicense(w, r, id)
				return
			case http.MethodPut:
				h.UpdateSoftwareLicense(w, r, id)
				return
			case http.MethodDelete:
				h.DeleteSoftwareLicense(w, r, id)
				return
			default:
				writeITAssetError(w, http.StatusMethodNotAllowed, "method not allowed")
				return
			}
		}
	}

	// /v2/assets/licenses (root)
	if path == "/v2/assets/licenses" {
		switch r.Method {
		case http.MethodGet:
			h.ListSoftwareLicenses(w, r)
			return
		case http.MethodPost:
			h.CreateSoftwareLicense(w, r)
			return
		default:
			writeITAssetError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
	}

	// /v2/assets/tags/{tag}
	if strings.HasPrefix(path, "/v2/assets/tags/") && r.Method == http.MethodGet {
		tag := strings.TrimPrefix(path, "/v2/assets/tags/")
		wsID, _ := strconv.ParseInt(r.URL.Query().Get("workspace_id"), 10, 64)
		h.GetAssetByTag(w, r, wsID, tag)
		return
	}

	// /v2/assets/{id}/assign
	if strings.HasPrefix(path, "/v2/assets/") && strings.HasSuffix(path, "/assign") && r.Method == http.MethodPost {
		trimmed := strings.TrimPrefix(path, "/v2/assets/")
		trimmed = strings.TrimSuffix(trimmed, "/assign")
		if id, err := strconv.ParseInt(trimmed, 10, 64); err == nil && id > 0 {
			h.AssignAsset(w, r, id)
			return
		}
	}

	// /v2/assets/{id}/checkin
	if strings.HasPrefix(path, "/v2/assets/") && strings.HasSuffix(path, "/checkin") && r.Method == http.MethodPost {
		trimmed := strings.TrimPrefix(path, "/v2/assets/")
		trimmed = strings.TrimSuffix(trimmed, "/checkin")
		if id, err := strconv.ParseInt(trimmed, 10, 64); err == nil && id > 0 {
			h.CheckinAsset(w, r, id)
			return
		}
	}

	// /v2/assets/{id}/status
	if strings.HasPrefix(path, "/v2/assets/") && strings.HasSuffix(path, "/status") && r.Method == http.MethodPost {
		trimmed := strings.TrimPrefix(path, "/v2/assets/")
		trimmed = strings.TrimSuffix(trimmed, "/status")
		if id, err := strconv.ParseInt(trimmed, 10, 64); err == nil && id > 0 {
			h.TransitionStatus(w, r, id)
			return
		}
	}

	// /v2/assets/{id}/depreciation
	if strings.HasPrefix(path, "/v2/assets/") && strings.HasSuffix(path, "/depreciation") && r.Method == http.MethodGet {
		trimmed := strings.TrimPrefix(path, "/v2/assets/")
		trimmed = strings.TrimSuffix(trimmed, "/depreciation")
		if id, err := strconv.ParseInt(trimmed, 10, 64); err == nil && id > 0 {
			h.CalculateDepreciation(w, r, id)
			return
		}
	}

	// /v2/assets/{id}/maintenance
	if strings.HasPrefix(path, "/v2/assets/") && strings.HasSuffix(path, "/maintenance") {
		trimmed := strings.TrimPrefix(path, "/v2/assets/")
		trimmed = strings.TrimSuffix(trimmed, "/maintenance")
		if id, err := strconv.ParseInt(trimmed, 10, 64); err == nil && id > 0 {
			switch r.Method {
			case http.MethodGet:
				h.ListMaintenanceRecords(w, r, id)
				return
			case http.MethodPost:
				h.CreateMaintenanceRecord(w, r, id)
				return
			default:
				writeITAssetError(w, http.StatusMethodNotAllowed, "method not allowed")
				return
			}
		}
	}

	// /v2/assets/{id}
	if strings.HasPrefix(path, "/v2/assets/") {
		idStr := strings.TrimPrefix(path, "/v2/assets/")
		if id, err := strconv.ParseInt(idStr, 10, 64); err == nil && id > 0 {
			switch r.Method {
			case http.MethodGet:
				h.GetAsset(w, r, id)
				return
			case http.MethodPut:
				h.UpdateAsset(w, r, id)
				return
			case http.MethodDelete:
				h.DeleteAsset(w, r, id)
				return
			default:
				writeITAssetError(w, http.StatusMethodNotAllowed, "method not allowed")
				return
			}
		}
	}

	// /v2/assets (root)
	if path == "/v2/assets" {
		switch r.Method {
		case http.MethodGet:
			h.ListAssets(w, r)
			return
		case http.MethodPost:
			h.CreateAsset(w, r)
			return
		default:
			writeITAssetError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
	}

	writeITAssetError(w, http.StatusNotFound, "endpoint not found")
}
