package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

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
