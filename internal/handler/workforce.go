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

// ListProfiles handles GET /v2/workforce/profiles
func (h *WorkforceHandler) ListProfiles(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	limit, _ := strconv.Atoi(q.Get("limit"))
	offset, _ := strconv.Atoi(q.Get("offset"))

	filter := model.WFProfileFilter{
		WorkspaceID:      getWFWorkspaceID(r),
		UserID:           q.Get("user_id"),
		SkillID:          q.Get("skill_id"),
		ProficiencyLevel: q.Get("proficiency_level"),
		VerifiedOnly:     q.Get("verified_only") == "true" || q.Get("verified_only") == "1",
		Limit:            limit,
		Offset:           offset,
	}

	profiles, total, err := h.svc.ListProfiles(r.Context(), filter)
	if err != nil {
		handleWFError(w, err)
		return
	}

	writeWFJSON(w, http.StatusOK, map[string]interface{}{
		"profiles": profiles,
		"total":    total,
		"limit":    filter.Limit,
		"offset":   filter.Offset,
	})
}

// UpsertProfile handles POST /v2/workforce/profiles
func (h *WorkforceHandler) UpsertProfile(w http.ResponseWriter, r *http.Request) {
	var req model.UpsertWFProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeWFError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.UserID == "" {
		req.UserID = getWFUser(r)
	}

	profile, err := h.svc.UpsertProfile(r.Context(), getWFWorkspaceID(r), req)
	if err != nil {
		handleWFError(w, err)
		return
	}
	writeWFJSON(w, http.StatusCreated, profile)
}

// GetProfileByID handles GET /v2/workforce/profiles/{id}
func (h *WorkforceHandler) GetProfileByID(w http.ResponseWriter, r *http.Request, id string) {
	profile, err := h.svc.GetProfileByID(r.Context(), id)
	if err != nil {
		handleWFError(w, err)
		return
	}
	writeWFJSON(w, http.StatusOK, profile)
}

// VerifyProfile handles POST /v2/workforce/profiles/{id}/verify
func (h *WorkforceHandler) VerifyProfile(w http.ResponseWriter, r *http.Request, id string) {
	var req model.VerifyWFProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeWFError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.VerifiedBy == "" {
		req.VerifiedBy = getWFUser(r)
	}

	profile, err := h.svc.VerifyProfile(r.Context(), id, req)
	if err != nil {
		handleWFError(w, err)
		return
	}
	writeWFJSON(w, http.StatusOK, profile)
}

// DeleteProfile handles DELETE /v2/workforce/profiles/{id}
func (h *WorkforceHandler) DeleteProfile(w http.ResponseWriter, r *http.Request, id string) {
	if err := h.svc.DeleteProfile(r.Context(), id); err != nil {
		handleWFError(w, err)
		return
	}
	writeWFJSON(w, http.StatusOK, map[string]string{"message": "member skill profile deleted"})
}

// ListCertifications handles GET /v2/workforce/certifications
func (h *WorkforceHandler) ListCertifications(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	limit, _ := strconv.Atoi(q.Get("limit"))
	offset, _ := strconv.Atoi(q.Get("offset"))

	filter := model.WFCertFilter{
		WorkspaceID: getWFWorkspaceID(r),
		UserID:      q.Get("user_id"),
		Issuer:      q.Get("issuer"),
		Status:      q.Get("status"),
		Limit:       limit,
		Offset:      offset,
	}

	certs, total, err := h.svc.ListCertifications(r.Context(), filter)
	if err != nil {
		handleWFError(w, err)
		return
	}

	writeWFJSON(w, http.StatusOK, map[string]interface{}{
		"certifications": certs,
		"total":          total,
		"limit":          filter.Limit,
		"offset":         filter.Offset,
	})
}

// CreateCertification handles POST /v2/workforce/certifications
func (h *WorkforceHandler) CreateCertification(w http.ResponseWriter, r *http.Request) {
	var req model.CreateWFCertRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeWFError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.UserID == "" {
		req.UserID = getWFUser(r)
	}

	cert, err := h.svc.CreateCertification(r.Context(), getWFWorkspaceID(r), req)
	if err != nil {
		handleWFError(w, err)
		return
	}
	writeWFJSON(w, http.StatusCreated, cert)
}

// GetCertificationByID handles GET /v2/workforce/certifications/{id}
func (h *WorkforceHandler) GetCertificationByID(w http.ResponseWriter, r *http.Request, id string) {
	cert, err := h.svc.GetCertificationByID(r.Context(), id)
	if err != nil {
		handleWFError(w, err)
		return
	}
	writeWFJSON(w, http.StatusOK, cert)
}

// UpdateCertification handles PUT /v2/workforce/certifications/{id}
func (h *WorkforceHandler) UpdateCertification(w http.ResponseWriter, r *http.Request, id string) {
	var req model.UpdateWFCertRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeWFError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	cert, err := h.svc.UpdateCertification(r.Context(), id, req)
	if err != nil {
		handleWFError(w, err)
		return
	}
	writeWFJSON(w, http.StatusOK, cert)
}

// DeleteCertification handles DELETE /v2/workforce/certifications/{id}
func (h *WorkforceHandler) DeleteCertification(w http.ResponseWriter, r *http.Request, id string) {
	if err := h.svc.DeleteCertification(r.Context(), id); err != nil {
		handleWFError(w, err)
		return
	}
	writeWFJSON(w, http.StatusOK, map[string]string{"message": "workforce certification deleted"})
}

// ListRoleCompetencies handles GET /v2/workforce/roles
func (h *WorkforceHandler) ListRoleCompetencies(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	limit, _ := strconv.Atoi(q.Get("limit"))
	offset, _ := strconv.Atoi(q.Get("offset"))

	filter := model.WFRoleCompetencyFilter{
		WorkspaceID: getWFWorkspaceID(r),
		RoleName:    q.Get("role_name"),
		SkillID:     q.Get("skill_id"),
		Priority:    q.Get("priority"),
		Limit:       limit,
		Offset:      offset,
	}

	roles, total, err := h.svc.ListRoleCompetencies(r.Context(), filter)
	if err != nil {
		handleWFError(w, err)
		return
	}

	writeWFJSON(w, http.StatusOK, map[string]interface{}{
		"role_competencies": roles,
		"total":             total,
		"limit":             filter.Limit,
		"offset":            filter.Offset,
	})
}

// CreateRoleCompetency handles POST /v2/workforce/roles
func (h *WorkforceHandler) CreateRoleCompetency(w http.ResponseWriter, r *http.Request) {
	var req model.CreateWFRoleCompetencyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeWFError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	rc, err := h.svc.CreateRoleCompetency(r.Context(), getWFWorkspaceID(r), req)
	if err != nil {
		handleWFError(w, err)
		return
	}
	writeWFJSON(w, http.StatusCreated, rc)
}

// DeleteRoleCompetency handles DELETE /v2/workforce/roles/{id}
func (h *WorkforceHandler) DeleteRoleCompetency(w http.ResponseWriter, r *http.Request, id string) {
	if err := h.svc.DeleteRoleCompetency(r.Context(), id); err != nil {
		handleWFError(w, err)
		return
	}
	writeWFJSON(w, http.StatusOK, map[string]string{"message": "role competency deleted"})
}

// ListTrainings handles GET /v2/workforce/trainings
func (h *WorkforceHandler) ListTrainings(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	limit, _ := strconv.Atoi(q.Get("limit"))
	offset, _ := strconv.Atoi(q.Get("offset"))

	filter := model.WFTrainingFilter{
		WorkspaceID: getWFWorkspaceID(r),
		UserID:      q.Get("user_id"),
		SkillID:     q.Get("skill_id"),
		Status:      q.Get("status"),
		Limit:       limit,
		Offset:      offset,
	}

	trainings, total, err := h.svc.ListTrainings(r.Context(), filter)
	if err != nil {
		handleWFError(w, err)
		return
	}

	writeWFJSON(w, http.StatusOK, map[string]interface{}{
		"trainings": trainings,
		"total":     total,
		"limit":     filter.Limit,
		"offset":    filter.Offset,
	})
}

// CreateTraining handles POST /v2/workforce/trainings
func (h *WorkforceHandler) CreateTraining(w http.ResponseWriter, r *http.Request) {
	var req model.CreateWFTrainingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeWFError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.UserID == "" {
		req.UserID = getWFUser(r)
	}

	training, err := h.svc.CreateTraining(r.Context(), getWFWorkspaceID(r), req)
	if err != nil {
		handleWFError(w, err)
		return
	}
	writeWFJSON(w, http.StatusCreated, training)
}

// GetTrainingByID handles GET /v2/workforce/trainings/{id}
func (h *WorkforceHandler) GetTrainingByID(w http.ResponseWriter, r *http.Request, id string) {
	training, err := h.svc.GetTrainingByID(r.Context(), id)
	if err != nil {
		handleWFError(w, err)
		return
	}
	writeWFJSON(w, http.StatusOK, training)
}

// UpdateTraining handles PUT /v2/workforce/trainings/{id}
func (h *WorkforceHandler) UpdateTraining(w http.ResponseWriter, r *http.Request, id string) {
	var req model.UpdateWFTrainingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeWFError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	training, err := h.svc.UpdateTraining(r.Context(), id, req)
	if err != nil {
		handleWFError(w, err)
		return
	}
	writeWFJSON(w, http.StatusOK, training)
}

// DeleteTraining handles DELETE /v2/workforce/trainings/{id}
func (h *WorkforceHandler) DeleteTraining(w http.ResponseWriter, r *http.Request, id string) {
	if err := h.svc.DeleteTraining(r.Context(), id); err != nil {
		handleWFError(w, err)
		return
	}
	writeWFJSON(w, http.StatusOK, map[string]string{"message": "training enrollment deleted"})
}

// Analytics & ServeHTTP stubs
func (h *WorkforceHandler) GetWorkforceAnalytics(w http.ResponseWriter, r *http.Request) {}
func (h *WorkforceHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	writeWFError(w, http.StatusNotFound, "route not found")
}
