package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/JohnMalugu/tsk-mgr-api/internal/model"
	"github.com/JohnMalugu/tsk-mgr-api/internal/service"
)

// PortfolioHandler manages HTTP endpoints for Portfolios, linked projects, milestones, updates, and metrics
type PortfolioHandler struct {
	svc service.PortfolioService
}

var GlobalPortfolioHandler *PortfolioHandler

// InitPortfolioHandler initializes the global portfolio handler singleton
func InitPortfolioHandler(svc service.PortfolioService) {
	GlobalPortfolioHandler = &PortfolioHandler{svc: svc}
}

// NewPortfolioHandler creates a new PortfolioHandler instance
func NewPortfolioHandler(svc service.PortfolioService) *PortfolioHandler {
	return &PortfolioHandler{svc: svc}
}

func getPortfolioUser(r *http.Request) string {
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

func writePortfolioJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func writePortfolioError(w http.ResponseWriter, status int, msg string) {
	writePortfolioJSON(w, status, map[string]string{"error": msg})
}

// HandlePortfoliosRoot handles GET and POST for /v2/portfolios
func (h *PortfolioHandler) HandlePortfoliosRoot(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.HandleListPortfolios(w, r)
	case http.MethodPost:
		h.HandleCreatePortfolio(w, r)
	default:
		writePortfolioError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

// HandleCreatePortfolio handles POST /v2/portfolios
func (h *PortfolioHandler) HandleCreatePortfolio(w http.ResponseWriter, r *http.Request) {
	userID := getPortfolioUser(r)

	var req model.CreatePortfolioRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writePortfolioError(w, http.StatusBadRequest, "Invalid request body: "+err.Error())
		return
	}

	portfolio, err := h.svc.CreatePortfolio(userID, req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrPortfolioNameRequired), errors.Is(err, service.ErrWorkspaceRequired):
			writePortfolioError(w, http.StatusBadRequest, err.Error())
		default:
			writePortfolioError(w, http.StatusInternalServerError, "Failed to create portfolio: "+err.Error())
		}
		return
	}

	writePortfolioJSON(w, http.StatusCreated, portfolio)
}

// HandleListPortfolios handles GET /v2/portfolios with filters
func (h *PortfolioHandler) HandleListPortfolios(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	filter := model.PortfolioFilter{
		OwnerUserID: q.Get("owner_user_id"),
		Status:      q.Get("status"),
		Search:      q.Get("search"),
	}

	if wsStr := q.Get("workspace_id"); wsStr != "" {
		if wsID, err := strconv.ParseInt(wsStr, 10, 64); err == nil && wsID > 0 {
			filter.WorkspaceID = &wsID
		}
	}
	if limStr := q.Get("limit"); limStr != "" {
		if lim, err := strconv.Atoi(limStr); err == nil {
			filter.Limit = lim
		}
	}
	if offStr := q.Get("offset"); offStr != "" {
		if off, err := strconv.Atoi(offStr); err == nil {
			filter.Offset = off
		}
	}

	portfolios, err := h.svc.ListPortfolios(filter)
	if err != nil {
		writePortfolioError(w, http.StatusInternalServerError, "Failed to list portfolios: "+err.Error())
		return
	}

	if portfolios == nil {
		portfolios = []*model.PortfolioSummary{}
	}

	writePortfolioJSON(w, http.StatusOK, portfolios)
}

// HandlePortfolioByID routes operations for /v2/portfolios/{id}
func (h *PortfolioHandler) HandlePortfolioByID(w http.ResponseWriter, r *http.Request, id int64) {
	switch r.Method {
	case http.MethodGet:
		h.HandleGetPortfolio(w, r, id)
	case http.MethodPut, http.MethodPatch:
		h.HandleUpdatePortfolio(w, r, id)
	case http.MethodDelete:
		h.HandleDeletePortfolio(w, r, id)
	default:
		writePortfolioError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

// HandleGetPortfolio handles GET /v2/portfolios/{id}
func (h *PortfolioHandler) HandleGetPortfolio(w http.ResponseWriter, r *http.Request, id int64) {
	detail, err := h.svc.GetPortfolio(id)
	if err != nil {
		if errors.Is(err, service.ErrPortfolioNotFound) {
			writePortfolioError(w, http.StatusNotFound, "Portfolio not found")
			return
		}
		writePortfolioError(w, http.StatusInternalServerError, "Failed to get portfolio: "+err.Error())
		return
	}

	writePortfolioJSON(w, http.StatusOK, detail)
}

// HandleUpdatePortfolio handles PUT/PATCH /v2/portfolios/{id}
func (h *PortfolioHandler) HandleUpdatePortfolio(w http.ResponseWriter, r *http.Request, id int64) {
	var req model.UpdatePortfolioRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writePortfolioError(w, http.StatusBadRequest, "Invalid request body: "+err.Error())
		return
	}

	portfolio, err := h.svc.UpdatePortfolio(id, req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrPortfolioNotFound):
			writePortfolioError(w, http.StatusNotFound, "Portfolio not found")
		case errors.Is(err, service.ErrPortfolioNameRequired), errors.Is(err, service.ErrInvalidPortfolioStatus):
			writePortfolioError(w, http.StatusBadRequest, err.Error())
		default:
			writePortfolioError(w, http.StatusInternalServerError, "Failed to update portfolio: "+err.Error())
		}
		return
	}

	writePortfolioJSON(w, http.StatusOK, portfolio)
}

// HandleDeletePortfolio handles DELETE /v2/portfolios/{id}
func (h *PortfolioHandler) HandleDeletePortfolio(w http.ResponseWriter, r *http.Request, id int64) {
	if err := h.svc.DeletePortfolio(id); err != nil {
		if errors.Is(err, service.ErrPortfolioNotFound) {
			writePortfolioError(w, http.StatusNotFound, "Portfolio not found")
			return
		}
		writePortfolioError(w, http.StatusInternalServerError, "Failed to delete portfolio: "+err.Error())
		return
	}

	writePortfolioJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Portfolio deleted successfully",
	})
}
