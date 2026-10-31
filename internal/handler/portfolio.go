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

// HandlePortfolioSubroutes routes sub-paths under /v2/portfolios/{id}/...
func (h *PortfolioHandler) HandlePortfolioSubroutes(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/v2/portfolios/")
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		writePortfolioError(w, http.StatusBadRequest, "Missing portfolio ID in path")
		return
	}

	id, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		writePortfolioError(w, http.StatusBadRequest, "Invalid portfolio ID format")
		return
	}

	if len(parts) == 1 {
		h.HandlePortfolioByID(w, r, id)
		return
	}

	switch parts[1] {
	case "projects":
		if len(parts) == 2 {
			h.HandlePortfolioProjects(w, r, id)
			return
		}
		if len(parts) == 3 {
			projID, err := strconv.Atoi(parts[2])
			if err != nil {
				writePortfolioError(w, http.StatusBadRequest, "Invalid project ID format")
				return
			}
			h.HandlePortfolioProjectItem(w, r, id, projID)
			return
		}

	case "milestones":
		if len(parts) == 2 {
			h.HandlePortfolioMilestones(w, r, id)
			return
		}
		if len(parts) == 3 {
			msID, err := strconv.ParseInt(parts[2], 10, 64)
			if err != nil {
				writePortfolioError(w, http.StatusBadRequest, "Invalid milestone ID format")
				return
			}
			h.HandlePortfolioMilestoneItem(w, r, id, msID)
			return
		}

	case "updates":
		if len(parts) == 2 {
			h.HandlePortfolioUpdates(w, r, id)
			return
		}

	case "metrics":
		if len(parts) == 2 {
			h.HandlePortfolioMetrics(w, r, id)
			return
		}
	}

	writePortfolioError(w, http.StatusNotFound, "Endpoint not found")
}

// HandlePortfolioProjects handles GET and POST for /v2/portfolios/{id}/projects
func (h *PortfolioHandler) HandlePortfolioProjects(w http.ResponseWriter, r *http.Request, portfolioID int64) {
	switch r.Method {
	case http.MethodGet:
		projects, err := h.svc.ListProjects(portfolioID)
		if err != nil {
			if errors.Is(err, service.ErrPortfolioNotFound) {
				writePortfolioError(w, http.StatusNotFound, "Portfolio not found")
				return
			}
			writePortfolioError(w, http.StatusInternalServerError, "Failed to list projects: "+err.Error())
			return
		}
		if projects == nil {
			projects = []model.PortfolioProject{}
		}
		writePortfolioJSON(w, http.StatusOK, projects)

	case http.MethodPost:
		var req model.AddPortfolioProjectRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writePortfolioError(w, http.StatusBadRequest, "Invalid request body: "+err.Error())
			return
		}

		item, err := h.svc.AddProject(portfolioID, req)
		if err != nil {
			switch {
			case errors.Is(err, service.ErrPortfolioNotFound):
				writePortfolioError(w, http.StatusNotFound, "Portfolio not found")
			case errors.Is(err, service.ErrProjectNotFound):
				writePortfolioError(w, http.StatusNotFound, "Project not found")
			case errors.Is(err, service.ErrPortfolioProjectExists):
				writePortfolioError(w, http.StatusBadRequest, err.Error())
			default:
				writePortfolioError(w, http.StatusInternalServerError, "Failed to add project: "+err.Error())
			}
			return
		}

		writePortfolioJSON(w, http.StatusCreated, item)

	default:
		writePortfolioError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

// HandlePortfolioProjectItem handles PUT/PATCH and DELETE for /v2/portfolios/{id}/projects/{projectId}
func (h *PortfolioHandler) HandlePortfolioProjectItem(w http.ResponseWriter, r *http.Request, portfolioID int64, projectID int) {
	switch r.Method {
	case http.MethodPut, http.MethodPatch:
		var req model.UpdatePortfolioProjectRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writePortfolioError(w, http.StatusBadRequest, "Invalid request body: "+err.Error())
			return
		}

		item, err := h.svc.UpdateProject(portfolioID, projectID, req)
		if err != nil {
			if errors.Is(err, service.ErrPortfolioProjectNF) {
				writePortfolioError(w, http.StatusNotFound, "Project not mapped to portfolio")
				return
			}
			writePortfolioError(w, http.StatusInternalServerError, "Failed to update project mapping: "+err.Error())
			return
		}
		writePortfolioJSON(w, http.StatusOK, item)

	case http.MethodDelete:
		if err := h.svc.RemoveProject(portfolioID, projectID); err != nil {
			if errors.Is(err, service.ErrPortfolioProjectNF) {
				writePortfolioError(w, http.StatusNotFound, "Project not mapped to portfolio")
				return
			}
			writePortfolioError(w, http.StatusInternalServerError, "Failed to remove project: "+err.Error())
			return
		}
		writePortfolioJSON(w, http.StatusOK, map[string]interface{}{
			"success": true,
			"message": "Project removed from portfolio",
		})

	default:
		writePortfolioError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

// HandlePortfolioMilestones handles GET and POST for /v2/portfolios/{id}/milestones
func (h *PortfolioHandler) HandlePortfolioMilestones(w http.ResponseWriter, r *http.Request, portfolioID int64) {
	switch r.Method {
	case http.MethodGet:
		milestones, err := h.svc.ListMilestones(portfolioID)
		if err != nil {
			if errors.Is(err, service.ErrPortfolioNotFound) {
				writePortfolioError(w, http.StatusNotFound, "Portfolio not found")
				return
			}
			writePortfolioError(w, http.StatusInternalServerError, "Failed to list milestones: "+err.Error())
			return
		}
		if milestones == nil {
			milestones = []model.PortfolioMilestone{}
		}
		writePortfolioJSON(w, http.StatusOK, milestones)

	case http.MethodPost:
		var req model.CreatePortfolioMilestoneRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writePortfolioError(w, http.StatusBadRequest, "Invalid request body: "+err.Error())
			return
		}

		milestone, err := h.svc.CreateMilestone(portfolioID, req)
		if err != nil {
			switch {
			case errors.Is(err, service.ErrPortfolioNotFound):
				writePortfolioError(w, http.StatusNotFound, "Portfolio not found")
			case errors.Is(err, service.ErrMilestoneTitleRequired), errors.Is(err, service.ErrMilestoneDateRequired):
				writePortfolioError(w, http.StatusBadRequest, err.Error())
			default:
				writePortfolioError(w, http.StatusInternalServerError, "Failed to create milestone: "+err.Error())
			}
			return
		}

		writePortfolioJSON(w, http.StatusCreated, milestone)

	default:
		writePortfolioError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

// HandlePortfolioMilestoneItem handles PUT/PATCH and DELETE for /v2/portfolios/{id}/milestones/{milestoneId}
func (h *PortfolioHandler) HandlePortfolioMilestoneItem(w http.ResponseWriter, r *http.Request, portfolioID int64, milestoneID int64) {
	switch r.Method {
	case http.MethodPut, http.MethodPatch:
		var req model.UpdatePortfolioMilestoneRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writePortfolioError(w, http.StatusBadRequest, "Invalid request body: "+err.Error())
			return
		}

		milestone, err := h.svc.UpdateMilestone(milestoneID, req)
		if err != nil {
			switch {
			case errors.Is(err, service.ErrMilestoneNotFound):
				writePortfolioError(w, http.StatusNotFound, "Milestone not found")
			case errors.Is(err, service.ErrMilestoneTitleRequired):
				writePortfolioError(w, http.StatusBadRequest, err.Error())
			default:
				writePortfolioError(w, http.StatusInternalServerError, "Failed to update milestone: "+err.Error())
			}
			return
		}
		writePortfolioJSON(w, http.StatusOK, milestone)

	case http.MethodDelete:
		if err := h.svc.DeleteMilestone(milestoneID); err != nil {
			if errors.Is(err, service.ErrMilestoneNotFound) {
				writePortfolioError(w, http.StatusNotFound, "Milestone not found")
				return
			}
			writePortfolioError(w, http.StatusInternalServerError, "Failed to delete milestone: "+err.Error())
			return
		}
		writePortfolioJSON(w, http.StatusOK, map[string]interface{}{
			"success": true,
			"message": "Milestone deleted successfully",
		})

	default:
		writePortfolioError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

// HandlePortfolioUpdates handles GET and POST for /v2/portfolios/{id}/updates
func (h *PortfolioHandler) HandlePortfolioUpdates(w http.ResponseWriter, r *http.Request, portfolioID int64) {
	switch r.Method {
	case http.MethodGet:
		updates, err := h.svc.ListUpdates(portfolioID)
		if err != nil {
			if errors.Is(err, service.ErrPortfolioNotFound) {
				writePortfolioError(w, http.StatusNotFound, "Portfolio not found")
				return
			}
			writePortfolioError(w, http.StatusInternalServerError, "Failed to list updates: "+err.Error())
			return
		}
		if updates == nil {
			updates = []model.PortfolioUpdate{}
		}
		writePortfolioJSON(w, http.StatusOK, updates)

	case http.MethodPost:
		authorUserID := getPortfolioUser(r)

		var req model.CreatePortfolioUpdateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writePortfolioError(w, http.StatusBadRequest, "Invalid request body: "+err.Error())
			return
		}

		update, err := h.svc.CreateUpdate(authorUserID, portfolioID, req)
		if err != nil {
			switch {
			case errors.Is(err, service.ErrPortfolioNotFound):
				writePortfolioError(w, http.StatusNotFound, "Portfolio not found")
			case errors.Is(err, service.ErrUpdateSummaryRequired), errors.Is(err, service.ErrInvalidHealthStatus):
				writePortfolioError(w, http.StatusBadRequest, err.Error())
			default:
				writePortfolioError(w, http.StatusInternalServerError, "Failed to submit executive update: "+err.Error())
			}
			return
		}

		writePortfolioJSON(w, http.StatusCreated, update)

	default:
		writePortfolioError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

// HandlePortfolioMetrics handles GET /v2/portfolios/{id}/metrics
func (h *PortfolioHandler) HandlePortfolioMetrics(w http.ResponseWriter, r *http.Request, portfolioID int64) {
	if r.Method != http.MethodGet {
		writePortfolioError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	metrics, err := h.svc.GetPortfolioMetrics(portfolioID)
	if err != nil {
		if errors.Is(err, service.ErrPortfolioNotFound) {
			writePortfolioError(w, http.StatusNotFound, "Portfolio not found")
			return
		}
		writePortfolioError(w, http.StatusInternalServerError, "Failed to get portfolio metrics: "+err.Error())
		return
	}

	writePortfolioJSON(w, http.StatusOK, metrics)
}

// HandleWorkspacePortfolioMetrics handles GET /v2/portfolios/metrics
func (h *PortfolioHandler) HandleWorkspacePortfolioMetrics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writePortfolioError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var workspaceID int64 = 1
	if wsStr := r.URL.Query().Get("workspace_id"); wsStr != "" {
		if wsID, err := strconv.ParseInt(wsStr, 10, 64); err == nil && wsID > 0 {
			workspaceID = wsID
		}
	}

	metrics, err := h.svc.GetWorkspaceMetrics(workspaceID)
	if err != nil {
		writePortfolioError(w, http.StatusInternalServerError, "Failed to get workspace portfolio metrics: "+err.Error())
		return
	}

	writePortfolioJSON(w, http.StatusOK, metrics)
}

