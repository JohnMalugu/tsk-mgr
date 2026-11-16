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

// -------------------------------------------------------------
// Vendor Endpoints
// -------------------------------------------------------------

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

// -------------------------------------------------------------
// Contract Endpoints
// -------------------------------------------------------------

// CreateContract handles POST /v2/contracts
func (h *VendorHandler) CreateContract(w http.ResponseWriter, r *http.Request) {
	var req model.CreateVendorContractRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeVendorError(w, http.StatusBadRequest, "invalid json request body")
		return
	}
	if req.WorkspaceID <= 0 {
		req.WorkspaceID = getVendorWorkspaceID(r)
	}

	contract, err := h.svc.CreateContract(r.Context(), &req)
	if err != nil {
		handleVendorError(w, err)
		return
	}
	writeVendorJSON(w, http.StatusCreated, contract)
}

// GetContract handles GET /v2/contracts/{id}
func (h *VendorHandler) GetContract(w http.ResponseWriter, r *http.Request, id int64) {
	contract, err := h.svc.GetContractByID(r.Context(), id)
	if err != nil {
		handleVendorError(w, err)
		return
	}
	writeVendorJSON(w, http.StatusOK, contract)
}

// GetContractByNumber handles GET /v2/contracts/by-number/{number}
func (h *VendorHandler) GetContractByNumber(w http.ResponseWriter, r *http.Request, number string) {
	wsID := getVendorWorkspaceID(r)
	contract, err := h.svc.GetContractByNumber(r.Context(), wsID, number)
	if err != nil {
		handleVendorError(w, err)
		return
	}
	writeVendorJSON(w, http.StatusOK, contract)
}

// UpdateContract handles PUT/PATCH /v2/contracts/{id}
func (h *VendorHandler) UpdateContract(w http.ResponseWriter, r *http.Request, id int64) {
	var req model.UpdateVendorContractRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeVendorError(w, http.StatusBadRequest, "invalid json request body")
		return
	}

	contract, err := h.svc.UpdateContract(r.Context(), id, &req)
	if err != nil {
		handleVendorError(w, err)
		return
	}
	writeVendorJSON(w, http.StatusOK, contract)
}

// DeleteContract handles DELETE /v2/contracts/{id}
func (h *VendorHandler) DeleteContract(w http.ResponseWriter, r *http.Request, id int64) {
	if err := h.svc.DeleteContract(r.Context(), id); err != nil {
		handleVendorError(w, err)
		return
	}
	writeVendorJSON(w, http.StatusOK, map[string]string{"message": "contract deleted successfully"})
}

// ListContracts handles GET /v2/contracts
func (h *VendorHandler) ListContracts(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	filter := model.ContractFilterCriteria{
		WorkspaceID:  getVendorWorkspaceID(r),
		ContractType: q.Get("contract_type"),
		Status:       q.Get("status"),
		Search:       q.Get("search"),
	}

	if vIDStr := q.Get("vendor_id"); vIDStr != "" {
		if vid, err := strconv.ParseInt(vIDStr, 10, 64); err == nil && vid > 0 {
			filter.VendorID = vid
		}
	}
	if expStr := q.Get("expiring_within_days"); expStr != "" {
		if days, err := strconv.Atoi(expStr); err == nil && days > 0 {
			filter.ExpiringDays = &days
		}
	} else if expStr := q.Get("expiring_days"); expStr != "" {
		if days, err := strconv.Atoi(expStr); err == nil && days > 0 {
			filter.ExpiringDays = &days
		}
	}
	if owner := q.Get("owner_user_id"); owner != "" {
		filter.OwnerUserID = owner
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

	contracts, total, err := h.svc.ListContracts(r.Context(), filter)
	if err != nil {
		handleVendorError(w, err)
		return
	}

	if contracts == nil {
		contracts = []model.VendorContract{}
	}

	writeVendorJSON(w, http.StatusOK, map[string]interface{}{
		"contracts": contracts,
		"total":     total,
		"limit":     filter.Limit,
		"offset":    filter.Offset,
	})
}

// RenewContract handles POST /v2/contracts/{id}/renew
func (h *VendorHandler) RenewContract(w http.ResponseWriter, r *http.Request, id int64) {
	var req model.RenewContractRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeVendorError(w, http.StatusBadRequest, "invalid json request body")
		return
	}

	contract, err := h.svc.RenewContract(r.Context(), id, &req)
	if err != nil {
		handleVendorError(w, err)
		return
	}
	writeVendorJSON(w, http.StatusOK, contract)
}

// -------------------------------------------------------------
// Vendor Evaluation Endpoints
// -------------------------------------------------------------

// CreateEvaluation handles POST /v2/vendors/{vendorId}/evaluations
func (h *VendorHandler) CreateEvaluation(w http.ResponseWriter, r *http.Request, vendorID int64) {
	var req model.CreateVendorEvaluationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeVendorError(w, http.StatusBadRequest, "invalid json request body")
		return
	}

	evaluatorID := getVendorUser(r)
	eval, err := h.svc.CreateEvaluation(r.Context(), vendorID, evaluatorID, &req)
	if err != nil {
		handleVendorError(w, err)
		return
	}
	writeVendorJSON(w, http.StatusCreated, eval)
}

// ListEvaluations handles GET /v2/vendors/{vendorId}/evaluations
func (h *VendorHandler) ListEvaluations(w http.ResponseWriter, r *http.Request, vendorID int64) {
	evals, err := h.svc.ListEvaluations(r.Context(), vendorID)
	if err != nil {
		handleVendorError(w, err)
		return
	}
	if evals == nil {
		evals = []model.VendorEvaluation{}
	}
	writeVendorJSON(w, http.StatusOK, map[string]interface{}{
		"evaluations": evals,
		"total":       len(evals),
	})
}

// DeleteEvaluation handles DELETE /v2/vendors/evaluations/{id}
func (h *VendorHandler) DeleteEvaluation(w http.ResponseWriter, r *http.Request, id int64) {
	if err := h.svc.DeleteEvaluation(r.Context(), id); err != nil {
		handleVendorError(w, err)
		return
	}
	writeVendorJSON(w, http.StatusOK, map[string]string{"message": "evaluation deleted successfully"})
}

// -------------------------------------------------------------
// Contract Milestone Endpoints
// -------------------------------------------------------------

// CreateMilestone handles POST /v2/contracts/{contractId}/milestones
func (h *VendorHandler) CreateMilestone(w http.ResponseWriter, r *http.Request, contractID int64) {
	var req model.CreateContractMilestoneRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeVendorError(w, http.StatusBadRequest, "invalid json request body")
		return
	}

	m, err := h.svc.CreateMilestone(r.Context(), contractID, &req)
	if err != nil {
		handleVendorError(w, err)
		return
	}
	writeVendorJSON(w, http.StatusCreated, m)
}

// GetMilestone handles GET /v2/contracts/milestones/{id}
func (h *VendorHandler) GetMilestone(w http.ResponseWriter, r *http.Request, id int64) {
	m, err := h.svc.GetMilestoneByID(r.Context(), id)
	if err != nil {
		handleVendorError(w, err)
		return
	}
	writeVendorJSON(w, http.StatusOK, m)
}

// UpdateMilestone handles PUT/PATCH /v2/contracts/milestones/{id}
func (h *VendorHandler) UpdateMilestone(w http.ResponseWriter, r *http.Request, id int64) {
	var req model.UpdateContractMilestoneRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeVendorError(w, http.StatusBadRequest, "invalid json request body")
		return
	}

	m, err := h.svc.UpdateMilestone(r.Context(), id, &req)
	if err != nil {
		handleVendorError(w, err)
		return
	}
	writeVendorJSON(w, http.StatusOK, m)
}

// DeleteMilestone handles DELETE /v2/contracts/milestones/{id}
func (h *VendorHandler) DeleteMilestone(w http.ResponseWriter, r *http.Request, id int64) {
	if err := h.svc.DeleteMilestone(r.Context(), id); err != nil {
		handleVendorError(w, err)
		return
	}
	writeVendorJSON(w, http.StatusOK, map[string]string{"message": "milestone deleted successfully"})
}

// ListMilestones handles GET /v2/contracts/{contractId}/milestones
func (h *VendorHandler) ListMilestones(w http.ResponseWriter, r *http.Request, contractID int64) {
	milestones, err := h.svc.ListMilestones(r.Context(), contractID)
	if err != nil {
		handleVendorError(w, err)
		return
	}
	if milestones == nil {
		milestones = []model.ContractMilestone{}
	}
	writeVendorJSON(w, http.StatusOK, map[string]interface{}{
		"milestones": milestones,
		"total":      len(milestones),
	})
}

// -------------------------------------------------------------
// Vendor CI Link & Metrics Endpoints
// -------------------------------------------------------------

// LinkCI handles POST /v2/vendors/{vendorId}/ci-links
func (h *VendorHandler) LinkCI(w http.ResponseWriter, r *http.Request, vendorID int64) {
	var req model.LinkVendorCIRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeVendorError(w, http.StatusBadRequest, "invalid json request body")
		return
	}

	link, err := h.svc.LinkCI(r.Context(), vendorID, &req)
	if err != nil {
		handleVendorError(w, err)
		return
	}
	writeVendorJSON(w, http.StatusCreated, link)
}

// UnlinkCI handles DELETE /v2/vendors/{vendorId}/ci-links/{ciId}
func (h *VendorHandler) UnlinkCI(w http.ResponseWriter, r *http.Request, vendorID, ciID int64) {
	if err := h.svc.UnlinkCI(r.Context(), vendorID, ciID); err != nil {
		handleVendorError(w, err)
		return
	}
	writeVendorJSON(w, http.StatusOK, map[string]string{"message": "vendor CI link deleted successfully"})
}

// ListCILinks handles GET /v2/vendors/{vendorId}/ci-links
func (h *VendorHandler) ListCILinks(w http.ResponseWriter, r *http.Request, vendorID int64) {
	links, err := h.svc.ListCILinks(r.Context(), vendorID)
	if err != nil {
		handleVendorError(w, err)
		return
	}
	if links == nil {
		links = []model.VendorCILink{}
	}
	writeVendorJSON(w, http.StatusOK, map[string]interface{}{
		"ci_links": links,
		"total":    len(links),
	})
}

// GetVendorMetrics handles GET /v2/vendors/metrics
func (h *VendorHandler) GetVendorMetrics(w http.ResponseWriter, r *http.Request) {
	wsID := getVendorWorkspaceID(r)
	metrics, err := h.svc.GetVendorMetrics(r.Context(), wsID)
	if err != nil {
		handleVendorError(w, err)
		return
	}
	writeVendorJSON(w, http.StatusOK, metrics)
}

// -------------------------------------------------------------
// ServeHTTP Router Dispatcher
// -------------------------------------------------------------

// ServeHTTP routes HTTP requests matching /v2/vendors and /v2/contracts
func (h *VendorHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if strings.HasPrefix(r.URL.Path, "/v2/vendors") {
		h.serveVendors(w, r)
		return
	}
	if strings.HasPrefix(r.URL.Path, "/v2/contracts") {
		h.serveContracts(w, r)
		return
	}
	writeVendorError(w, http.StatusNotFound, "not found")
}

func (h *VendorHandler) serveVendors(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/v2/vendors")
	path = strings.TrimPrefix(path, "/")

	if path == "" {
		switch r.Method {
		case http.MethodGet:
			h.ListVendors(w, r)
		case http.MethodPost:
			h.CreateVendor(w, r)
		default:
			writeVendorError(w, http.StatusMethodNotAllowed, "method not allowed")
		}
		return
	}

	if path == "metrics" {
		if r.Method != http.MethodGet {
			writeVendorError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		h.GetVendorMetrics(w, r)
		return
	}

	if strings.HasPrefix(path, "by-slug/") {
		slug := strings.TrimPrefix(path, "by-slug/")
		if r.Method != http.MethodGet {
			writeVendorError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		h.GetVendorBySlug(w, r, slug)
		return
	}

	if strings.HasPrefix(path, "evaluations/") {
		idStr := strings.TrimPrefix(path, "evaluations/")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil || id <= 0 {
			writeVendorError(w, http.StatusBadRequest, "invalid evaluation id")
			return
		}
		if r.Method == http.MethodDelete {
			h.DeleteEvaluation(w, r, id)
			return
		}
		writeVendorError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	parts := strings.Split(path, "/")
	id, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil || id <= 0 {
		writeVendorError(w, http.StatusBadRequest, "invalid vendor id")
		return
	}

	if len(parts) == 1 {
		switch r.Method {
		case http.MethodGet:
			h.GetVendor(w, r, id)
		case http.MethodPut, http.MethodPatch:
			h.UpdateVendor(w, r, id)
		case http.MethodDelete:
			h.DeleteVendor(w, r, id)
		default:
			writeVendorError(w, http.StatusMethodNotAllowed, "method not allowed")
		}
		return
	}

	if parts[1] == "evaluations" {
		if len(parts) == 2 {
			switch r.Method {
			case http.MethodGet:
				h.ListEvaluations(w, r, id)
			case http.MethodPost:
				h.CreateEvaluation(w, r, id)
			default:
				writeVendorError(w, http.StatusMethodNotAllowed, "method not allowed")
			}
			return
		}
		if len(parts) == 3 {
			evalID, err := strconv.ParseInt(parts[2], 10, 64)
			if err != nil || evalID <= 0 {
				writeVendorError(w, http.StatusBadRequest, "invalid evaluation id")
				return
			}
			if r.Method == http.MethodDelete {
				h.DeleteEvaluation(w, r, evalID)
				return
			}
			writeVendorError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
	}

	if parts[1] == "ci-links" {
		if len(parts) == 2 {
			switch r.Method {
			case http.MethodGet:
				h.ListCILinks(w, r, id)
			case http.MethodPost:
				h.LinkCI(w, r, id)
			default:
				writeVendorError(w, http.StatusMethodNotAllowed, "method not allowed")
			}
			return
		}
		if len(parts) == 3 {
			ciID, err := strconv.ParseInt(parts[2], 10, 64)
			if err != nil || ciID <= 0 {
				writeVendorError(w, http.StatusBadRequest, "invalid configuration item id")
				return
			}
			if r.Method == http.MethodDelete {
				h.UnlinkCI(w, r, id, ciID)
				return
			}
			writeVendorError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
	}

	writeVendorError(w, http.StatusNotFound, "endpoint not found")
}

func (h *VendorHandler) serveContracts(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/v2/contracts")
	path = strings.TrimPrefix(path, "/")

	if path == "" {
		switch r.Method {
		case http.MethodGet:
			h.ListContracts(w, r)
		case http.MethodPost:
			h.CreateContract(w, r)
		default:
			writeVendorError(w, http.StatusMethodNotAllowed, "method not allowed")
		}
		return
	}

	if strings.HasPrefix(path, "by-number/") {
		number := strings.TrimPrefix(path, "by-number/")
		if r.Method != http.MethodGet {
			writeVendorError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		h.GetContractByNumber(w, r, number)
		return
	}

	if strings.HasPrefix(path, "milestones/") {
		mIDStr := strings.TrimPrefix(path, "milestones/")
		mID, err := strconv.ParseInt(mIDStr, 10, 64)
		if err != nil || mID <= 0 {
			writeVendorError(w, http.StatusBadRequest, "invalid milestone id")
			return
		}
		switch r.Method {
		case http.MethodGet:
			h.GetMilestone(w, r, mID)
		case http.MethodPut, http.MethodPatch:
			h.UpdateMilestone(w, r, mID)
		case http.MethodDelete:
			h.DeleteMilestone(w, r, mID)
		default:
			writeVendorError(w, http.StatusMethodNotAllowed, "method not allowed")
		}
		return
	}

	parts := strings.Split(path, "/")
	id, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil || id <= 0 {
		writeVendorError(w, http.StatusBadRequest, "invalid contract id")
		return
	}

	if len(parts) == 1 {
		switch r.Method {
		case http.MethodGet:
			h.GetContract(w, r, id)
		case http.MethodPut, http.MethodPatch:
			h.UpdateContract(w, r, id)
		case http.MethodDelete:
			h.DeleteContract(w, r, id)
		default:
			writeVendorError(w, http.StatusMethodNotAllowed, "method not allowed")
		}
		return
	}

	if parts[1] == "renew" {
		if r.Method == http.MethodPost {
			h.RenewContract(w, r, id)
			return
		}
		writeVendorError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	if parts[1] == "milestones" {
		if len(parts) == 2 {
			switch r.Method {
			case http.MethodGet:
				h.ListMilestones(w, r, id)
			case http.MethodPost:
				h.CreateMilestone(w, r, id)
			default:
				writeVendorError(w, http.StatusMethodNotAllowed, "method not allowed")
			}
			return
		}
		if len(parts) == 3 {
			mID, err := strconv.ParseInt(parts[2], 10, 64)
			if err != nil || mID <= 0 {
				writeVendorError(w, http.StatusBadRequest, "invalid milestone id")
				return
			}
			switch r.Method {
			case http.MethodGet:
				h.GetMilestone(w, r, mID)
			case http.MethodPut, http.MethodPatch:
				h.UpdateMilestone(w, r, mID)
			case http.MethodDelete:
				h.DeleteMilestone(w, r, mID)
			default:
				writeVendorError(w, http.StatusMethodNotAllowed, "method not allowed")
			}
			return
		}
	}

	writeVendorError(w, http.StatusNotFound, "endpoint not found")
}
