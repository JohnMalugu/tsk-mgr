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

// WorkforceHandler exposes HTTP endpoints for ITIL Workforce & Talent Management
type WorkforceHandler struct {
	svc service.WorkforceService
}

// GlobalWorkforceHandler is the singleton handler instance
var GlobalWorkforceHandler *WorkforceHandler

// InitWorkforceHandler initializes the global WorkforceHandler singleton
func InitWorkforceHandler(svc service.WorkforceService) {
	GlobalWorkforceHandler = &WorkforceHandler{svc: svc}
}

// NewWorkforceHandler creates a new WorkforceHandler instance
func NewWorkforceHandler(svc service.WorkforceService) *WorkforceHandler {
	return &WorkforceHandler{svc: svc}
}

func getWFWorkspaceID(r *http.Request) string {
	if ws := r.URL.Query().Get("workspace_id"); ws != "" {
		return ws
	}
	if ws := r.Header.Get("X-Workspace-ID"); ws != "" {
		return ws
	}
	return "ws_1"
}

func getWFUser(r *http.Request) string {
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

func writeWFJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func writeWFError(w http.ResponseWriter, status int, msg string) {
	writeWFJSON(w, status, map[string]string{"error": msg})
}

func handleWFError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrWFSkillNotFound):
		writeWFError(w, http.StatusNotFound, "workforce skill not found")
	case errors.Is(err, service.ErrWFSkillCodeConflict):
		writeWFError(w, http.StatusConflict, "workforce skill with this code already exists in workspace")
	case errors.Is(err, service.ErrWFProfileNotFound):
		writeWFError(w, http.StatusNotFound, "member skill profile not found")
	case errors.Is(err, service.ErrWFCertificationNotFound):
		writeWFError(w, http.StatusNotFound, "workforce certification not found")
	case errors.Is(err, service.ErrWFRoleCompetencyNotFound):
		writeWFError(w, http.StatusNotFound, "role competency not found")
	case errors.Is(err, service.ErrWFTrainingNotFound):
		writeWFError(w, http.StatusNotFound, "training enrollment not found")
	default:
		writeWFError(w, http.StatusBadRequest, err.Error())
	}
}

// ListSkills handles GET /v2/workforce/skills
func (h *WorkforceHandler) ListSkills(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	limit, _ := strconv.Atoi(q.Get("limit"))
	offset, _ := strconv.Atoi(q.Get("offset"))

	filter := model.WFSkillFilter{
		WorkspaceID: getWFWorkspaceID(r),
		Category:    q.Get("category"),
		Search:      q.Get("search"),
		Limit:       limit,
		Offset:      offset,
	}
	if critStr := q.Get("is_critical"); critStr != "" {
		crit := critStr == "true" || critStr == "1"
		filter.IsCritical = &crit
	}

	skills, total, err := h.svc.ListSkills(r.Context(), filter)
	if err != nil {
		handleWFError(w, err)
		return
	}

	writeWFJSON(w, http.StatusOK, map[string]interface{}{
		"skills": skills,
		"total":  total,
		"limit":  filter.Limit,
		"offset": filter.Offset,
	})
}

// CreateSkill handles POST /v2/workforce/skills
func (h *WorkforceHandler) CreateSkill(w http.ResponseWriter, r *http.Request) {
	var req model.CreateWFSkillRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeWFError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	skill, err := h.svc.CreateSkill(r.Context(), getWFWorkspaceID(r), req)
	if err != nil {
		handleWFError(w, err)
		return
	}
	writeWFJSON(w, http.StatusCreated, skill)
}

// GetSkillByID handles GET /v2/workforce/skills/{id}
func (h *WorkforceHandler) GetSkillByID(w http.ResponseWriter, r *http.Request, id string) {
	skill, err := h.svc.GetSkillByID(r.Context(), id)
	if err != nil {
		handleWFError(w, err)
		return
	}
	writeWFJSON(w, http.StatusOK, skill)
}

// UpdateSkill handles PUT /v2/workforce/skills/{id}
func (h *WorkforceHandler) UpdateSkill(w http.ResponseWriter, r *http.Request, id string) {
	var req model.UpdateWFSkillRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeWFError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	skill, err := h.svc.UpdateSkill(r.Context(), id, req)
	if err != nil {
		handleWFError(w, err)
		return
	}
	writeWFJSON(w, http.StatusOK, skill)
}

// DeleteSkill handles DELETE /v2/workforce/skills/{id}
func (h *WorkforceHandler) DeleteSkill(w http.ResponseWriter, r *http.Request, id string) {
	if err := h.svc.DeleteSkill(r.Context(), id); err != nil {
		handleWFError(w, err)
		return
	}
	writeWFJSON(w, http.StatusOK, map[string]string{"message": "workforce skill deleted"})
}

// Profiles stubs
func (h *WorkforceHandler) ListProfiles(w http.ResponseWriter, r *http.Request) {}
func (h *WorkforceHandler) UpsertProfile(w http.ResponseWriter, r *http.Request) {}
func (h *WorkforceHandler) GetProfileByID(w http.ResponseWriter, r *http.Request, id string) {}
func (h *WorkforceHandler) VerifyProfile(w http.ResponseWriter, r *http.Request, id string) {}
func (h *WorkforceHandler) DeleteProfile(w http.ResponseWriter, r *http.Request, id string) {}

// Certifications stubs
func (h *WorkforceHandler) ListCertifications(w http.ResponseWriter, r *http.Request) {}
func (h *WorkforceHandler) CreateCertification(w http.ResponseWriter, r *http.Request) {}
func (h *WorkforceHandler) GetCertificationByID(w http.ResponseWriter, r *http.Request, id string) {}
func (h *WorkforceHandler) UpdateCertification(w http.ResponseWriter, r *http.Request, id string) {}
func (h *WorkforceHandler) DeleteCertification(w http.ResponseWriter, r *http.Request, id string) {}

// Roles & Trainings stubs
func (h *WorkforceHandler) ListRoleCompetencies(w http.ResponseWriter, r *http.Request) {}
func (h *WorkforceHandler) CreateRoleCompetency(w http.ResponseWriter, r *http.Request) {}
func (h *WorkforceHandler) DeleteRoleCompetency(w http.ResponseWriter, r *http.Request, id string) {}
func (h *WorkforceHandler) ListTrainings(w http.ResponseWriter, r *http.Request) {}
func (h *WorkforceHandler) CreateTraining(w http.ResponseWriter, r *http.Request) {}
func (h *WorkforceHandler) GetTrainingByID(w http.ResponseWriter, r *http.Request, id string) {}
func (h *WorkforceHandler) UpdateTraining(w http.ResponseWriter, r *http.Request, id string) {}
func (h *WorkforceHandler) DeleteTraining(w http.ResponseWriter, r *http.Request, id string) {}

// Analytics & ServeHTTP stubs
func (h *WorkforceHandler) GetWorkforceAnalytics(w http.ResponseWriter, r *http.Request) {}
func (h *WorkforceHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	writeWFError(w, http.StatusNotFound, "route not found")
}
