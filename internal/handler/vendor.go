package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/JohnMalugu/tsk-mgr-api/internal/model"
	"github.com/JohnMalugu/tsk-mgr-api/internal/service"
)

// VendorHandler exposes HTTP endpoints for the ITIL Supplier & Contract Management System
type VendorHandler struct {
	svc service.VendorService
}

// GlobalVendorHandler is the singleton handler instance
var GlobalVendorHandler *VendorHandler

// InitVendorHandler initializes the global vendor handler
func InitVendorHandler(svc service.VendorService) {
	GlobalVendorHandler = &VendorHandler{svc: svc}
}

// NewVendorHandler creates a new VendorHandler instance
func NewVendorHandler(svc service.VendorService) *VendorHandler {
	return &VendorHandler{svc: svc}
}

func getVendorUser(r *http.Request) string {
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

func getVendorWorkspaceID(r *http.Request) int64 {
	if wsIDStr := r.URL.Query().Get("workspace_id"); wsIDStr != "" {
		if id, err := strconv.ParseInt(wsIDStr, 10, 64); err == nil && id > 0 {
			return id
		}
	}
	if wsIDStr := r.Header.Get("X-Workspace-ID"); wsIDStr != "" {
		if id, err := strconv.ParseInt(wsIDStr, 10, 64); err == nil && id > 0 {
			return id
		}
	}
	return 1
}

func writeVendorJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func writeVendorError(w http.ResponseWriter, status int, msg string) {
	writeVendorJSON(w, status, map[string]string{"error": msg})
}

func handleVendorError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrVendorNotFound):
		writeVendorError(w, http.StatusNotFound, "vendor not found")
	case errors.Is(err, service.ErrVendorNameRequired):
		writeVendorError(w, http.StatusBadRequest, "vendor name is required")
	case errors.Is(err, service.ErrVendorSlugRequired):
		writeVendorError(w, http.StatusBadRequest, "vendor slug is required")
	case errors.Is(err, service.ErrVendorSlugConflict):
		writeVendorError(w, http.StatusConflict, "vendor slug already exists")
	case errors.Is(err, service.ErrInvalidVendorCategory):
		writeVendorError(w, http.StatusBadRequest, "invalid vendor category")
	case errors.Is(err, service.ErrInvalidVendorTier):
		writeVendorError(w, http.StatusBadRequest, "invalid vendor tier")
	case errors.Is(err, service.ErrInvalidVendorStatus):
		writeVendorError(w, http.StatusBadRequest, "invalid vendor status")
	case errors.Is(err, service.ErrInvalidVendorRisk):
		writeVendorError(w, http.StatusBadRequest, "invalid vendor risk level")
	case errors.Is(err, service.ErrContractNotFound):
		writeVendorError(w, http.StatusNotFound, "contract not found")
	case errors.Is(err, service.ErrContractTitleRequired):
		writeVendorError(w, http.StatusBadRequest, "contract title is required")
	case errors.Is(err, service.ErrContractNumberRequired):
		writeVendorError(w, http.StatusBadRequest, "contract number is required")
	case errors.Is(err, service.ErrContractNumberConflict):
		writeVendorError(w, http.StatusConflict, "contract number already exists")
	case errors.Is(err, service.ErrInvalidContractType):
		writeVendorError(w, http.StatusBadRequest, "invalid contract type")
	case errors.Is(err, service.ErrInvalidContractStatus):
		writeVendorError(w, http.StatusBadRequest, "invalid contract status")
	case errors.Is(err, service.ErrInvalidContractDates):
		writeVendorError(w, http.StatusBadRequest, "end date must be after start date")
	case errors.Is(err, service.ErrEvaluationNotFound):
		writeVendorError(w, http.StatusNotFound, "vendor evaluation not found")
	case errors.Is(err, service.ErrInvalidScore):
		writeVendorError(w, http.StatusBadRequest, "score must be between 1.0 and 5.0")
	case errors.Is(err, service.ErrContractMilestoneNotFound):
		writeVendorError(w, http.StatusNotFound, "contract milestone not found")
	case errors.Is(err, service.ErrInvalidMilestoneStatus):
		writeVendorError(w, http.StatusBadRequest, "invalid contract milestone status")
	case errors.Is(err, service.ErrVendorCILinkNotFound):
		writeVendorError(w, http.StatusNotFound, "vendor CI link not found")
	case errors.Is(err, service.ErrVendorCILinkConflict):
		writeVendorError(w, http.StatusConflict, "vendor CI link already exists")
	default:
		writeVendorError(w, http.StatusInternalServerError, err.Error())
	}
}

// CreateVendor handles POST /v2/vendors
func (h *VendorHandler) CreateVendor(w http.ResponseWriter, r *http.Request) {
	var req model.CreateVendorRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeVendorError(w, http.StatusBadRequest, "invalid json request body")
		return
	}
	if req.WorkspaceID <= 0 {
		req.WorkspaceID = getVendorWorkspaceID(r)
	}

	vendor, err := h.svc.CreateVendor(r.Context(), &req)
	if err != nil {
		handleVendorError(w, err)
		return
	}
	writeVendorJSON(w, http.StatusCreated, vendor)
}

// GetVendor handles GET /v2/vendors/{id}
func (h *VendorHandler) GetVendor(w http.ResponseWriter, r *http.Request, id int64) {
	vendor, err := h.svc.GetVendorByID(r.Context(), id)
	if err != nil {
		handleVendorError(w, err)
		return
	}
	writeVendorJSON(w, http.StatusOK, vendor)
}

// GetVendorBySlug handles GET /v2/vendors/by-slug/{slug}
func (h *VendorHandler) GetVendorBySlug(w http.ResponseWriter, r *http.Request, slug string) {
	wsID := getVendorWorkspaceID(r)
	vendor, err := h.svc.GetVendorBySlug(r.Context(), wsID, slug)
	if err != nil {
		handleVendorError(w, err)
		return
	}
	writeVendorJSON(w, http.StatusOK, vendor)
}

// UpdateVendor handles PUT/PATCH /v2/vendors/{id}
func (h *VendorHandler) UpdateVendor(w http.ResponseWriter, r *http.Request, id int64) {
	var req model.UpdateVendorRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeVendorError(w, http.StatusBadRequest, "invalid json request body")
		return
	}

	vendor, err := h.svc.UpdateVendor(r.Context(), id, &req)
	if err != nil {
		handleVendorError(w, err)
		return
	}
	writeVendorJSON(w, http.StatusOK, vendor)
}

// DeleteVendor handles DELETE /v2/vendors/{id}
func (h *VendorHandler) DeleteVendor(w http.ResponseWriter, r *http.Request, id int64) {
	if err := h.svc.DeleteVendor(r.Context(), id); err != nil {
		handleVendorError(w, err)
		return
	}
	writeVendorJSON(w, http.StatusOK, map[string]string{"message": "vendor deleted successfully"})
}

// ListVendors handles GET /v2/vendors
func (h *VendorHandler) ListVendors(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	filter := model.VendorFilterCriteria{
		WorkspaceID: getVendorWorkspaceID(r),
		Category:    q.Get("category"),
		Tier:        q.Get("tier"),
		Status:      q.Get("status"),
		RiskLevel:   q.Get("risk_level"),
		Search:      q.Get("search"),
	}

	if limitStr := q.Get("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
			filter.Limit = limit
		}
	}
	if offsetStr := q.Get("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil && offset >= 0 {
			filter.Offset = offset
		}
	}

	vendors, total, err := h.svc.ListVendors(r.Context(), filter)
	if err != nil {
		handleVendorError(w, err)
		return
	}

	if vendors == nil {
		vendors = []model.Vendor{}
	}

	writeVendorJSON(w, http.StatusOK, map[string]interface{}{
		"vendors": vendors,
		"total":   total,
		"limit":   filter.Limit,
		"offset":  filter.Offset,
	})
}
