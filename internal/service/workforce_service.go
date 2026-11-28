package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/JohnMalugu/tsk-mgr-api/internal/model"
	"github.com/JohnMalugu/tsk-mgr-api/internal/repository"
)

var (
	_ = uuid.New
	_ = fmt.Sprintf
	_ = strings.TrimSpace
	_ = time.Now
	_ = errors.New
)

var (
	ErrWFSkillNotFound          = repository.ErrWFSkillNotFound
	ErrWFSkillCodeConflict      = repository.ErrWFSkillCodeConflict
	ErrWFProfileNotFound        = repository.ErrWFProfileNotFound
	ErrWFCertificationNotFound  = repository.ErrWFCertificationNotFound
	ErrWFRoleCompetencyNotFound = repository.ErrWFRoleCompetencyNotFound
	ErrWFTrainingNotFound       = repository.ErrWFTrainingNotFound
)

// WorkforceService defines business logic and orchestration for ITIL Workforce & Talent Management
type WorkforceService interface {
	// Skills
	CreateSkill(ctx context.Context, workspaceID string, req model.CreateWFSkillRequest) (*model.WorkforceSkill, error)
	GetSkillByID(ctx context.Context, id string) (*model.WorkforceSkill, error)
	GetSkillByCode(ctx context.Context, workspaceID, code string) (*model.WorkforceSkill, error)
	UpdateSkill(ctx context.Context, id string, req model.UpdateWFSkillRequest) (*model.WorkforceSkill, error)
	DeleteSkill(ctx context.Context, id string) error
	ListSkills(ctx context.Context, filter model.WFSkillFilter) ([]model.WorkforceSkill, int, error)

	// Profiles
	UpsertProfile(ctx context.Context, workspaceID string, req model.UpsertWFProfileRequest) (*model.MemberSkillProfile, error)
	GetProfileByID(ctx context.Context, id string) (*model.MemberSkillProfile, error)
	VerifyProfile(ctx context.Context, id string, req model.VerifyWFProfileRequest) (*model.MemberSkillProfile, error)
	DeleteProfile(ctx context.Context, id string) error
	ListProfiles(ctx context.Context, filter model.WFProfileFilter) ([]model.MemberSkillProfile, int, error)

	// Certifications
	CreateCertification(ctx context.Context, workspaceID string, req model.CreateWFCertRequest) (*model.WorkforceCertification, error)
	GetCertificationByID(ctx context.Context, id string) (*model.WorkforceCertification, error)
	UpdateCertification(ctx context.Context, id string, req model.UpdateWFCertRequest) (*model.WorkforceCertification, error)
	DeleteCertification(ctx context.Context, id string) error
	ListCertifications(ctx context.Context, filter model.WFCertFilter) ([]model.WorkforceCertification, int, error)

	// Role Competencies
	CreateRoleCompetency(ctx context.Context, workspaceID string, req model.CreateWFRoleCompetencyRequest) (*model.RoleCompetency, error)
	GetRoleCompetencyByID(ctx context.Context, id string) (*model.RoleCompetency, error)
	DeleteRoleCompetency(ctx context.Context, id string) error
	ListRoleCompetencies(ctx context.Context, filter model.WFRoleCompetencyFilter) ([]model.RoleCompetency, int, error)

	// Trainings
	CreateTraining(ctx context.Context, workspaceID string, req model.CreateWFTrainingRequest) (*model.TrainingEnrollment, error)
	GetTrainingByID(ctx context.Context, id string) (*model.TrainingEnrollment, error)
	UpdateTraining(ctx context.Context, id string, req model.UpdateWFTrainingRequest) (*model.TrainingEnrollment, error)
	DeleteTraining(ctx context.Context, id string) error
	ListTrainings(ctx context.Context, filter model.WFTrainingFilter) ([]model.TrainingEnrollment, int, error)

	// Analytics
	GetWorkforceAnalytics(ctx context.Context, workspaceID string) (*model.WorkforceAnalytics, error)
}

type workforceService struct {
	repo repository.WorkforceRepository
}

// NewWorkforceService constructs an ITIL Workforce & Talent Management service
func NewWorkforceService(repo repository.WorkforceRepository) WorkforceService {
	return &workforceService{repo: repo}
}

// CreateSkill validates and adds a skill to the workspace skill catalog
func (s *workforceService) CreateSkill(ctx context.Context, workspaceID string, req model.CreateWFSkillRequest) (*model.WorkforceSkill, error) {
	if req.WorkspaceID == "" {
		req.WorkspaceID = workspaceID
	}
	if err := req.Validate(); err != nil {
		return nil, err
	}

	code := strings.TrimSpace(req.SkillCode)
	if code == "" {
		code = fmt.Sprintf("SKL-%s", strings.ToUpper(uuid.New().String()[:8]))
	}
	cat := req.Category
	if cat == "" {
		cat = model.WFSkillCategorySoftwareEng
	}

	skill := &model.WorkforceSkill{
		WorkspaceID: req.WorkspaceID,
		SkillCode:   code,
		Name:        req.Name,
		Category:    cat,
		Description: req.Description,
		IsCritical:  req.IsCritical,
	}

	if err := s.repo.CreateSkill(ctx, skill); err != nil {
		return nil, err
	}
	return skill, nil
}

// GetSkillByID retrieves a skill by ID
func (s *workforceService) GetSkillByID(ctx context.Context, id string) (*model.WorkforceSkill, error) {
	if strings.TrimSpace(id) == "" {
		return nil, ErrWFSkillNotFound
	}
	return s.repo.GetSkillByID(ctx, id)
}

// GetSkillByCode retrieves a skill by code within a workspace
func (s *workforceService) GetSkillByCode(ctx context.Context, workspaceID, code string) (*model.WorkforceSkill, error) {
	if strings.TrimSpace(code) == "" {
		return nil, ErrWFSkillNotFound
	}
	return s.repo.GetSkillByCode(ctx, workspaceID, code)
}

// UpdateSkill modifies an existing cataloged skill
func (s *workforceService) UpdateSkill(ctx context.Context, id string, req model.UpdateWFSkillRequest) (*model.WorkforceSkill, error) {
	if strings.TrimSpace(id) == "" {
		return nil, ErrWFSkillNotFound
	}
	if err := req.Validate(); err != nil {
		return nil, err
	}

	skill, err := s.repo.GetSkillByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.Name != nil {
		skill.Name = *req.Name
	}
	if req.Category != nil {
		skill.Category = *req.Category
	}
	if req.Description != nil {
		skill.Description = *req.Description
	}
	if req.IsCritical != nil {
		skill.IsCritical = *req.IsCritical
	}

	if err := s.repo.UpdateSkill(ctx, skill); err != nil {
		return nil, err
	}
	return skill, nil
}

// DeleteSkill deletes a skill from the catalog
func (s *workforceService) DeleteSkill(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return ErrWFSkillNotFound
	}
	return s.repo.DeleteSkill(ctx, id)
}

// ListSkills queries cataloged skills matching filter criteria
func (s *workforceService) ListSkills(ctx context.Context, filter model.WFSkillFilter) ([]model.WorkforceSkill, int, error) {
	return s.repo.ListSkills(ctx, filter)
}

// UpsertProfile creates or updates a team member's proficiency profile for a skill
func (s *workforceService) UpsertProfile(ctx context.Context, workspaceID string, req model.UpsertWFProfileRequest) (*model.MemberSkillProfile, error) {
	if req.WorkspaceID == "" {
		req.WorkspaceID = workspaceID
	}
	if err := req.Validate(); err != nil {
		return nil, err
	}

	if _, err := s.repo.GetSkillByID(ctx, req.SkillID); err != nil {
		return nil, err
	}

	lvl := req.ProficiencyLevel
	if lvl == "" {
		lvl = model.WFProficiencyNovice
	}

	now := time.Now().UTC()
	existing, err := s.repo.GetProfileByUserAndSkill(ctx, req.WorkspaceID, req.UserID, req.SkillID)
	if err == nil && existing != nil {
		existing.ProficiencyLevel = lvl
		existing.YearsExperience = req.YearsExperience
		existing.LastAssessedAt = now
		if err := s.repo.UpdateProfile(ctx, existing); err != nil {
			return nil, err
		}
		return existing, nil
	}

	profile := &model.MemberSkillProfile{
		WorkspaceID:      req.WorkspaceID,
		UserID:           req.UserID,
		SkillID:          req.SkillID,
		ProficiencyLevel: lvl,
		YearsExperience:  req.YearsExperience,
		LastAssessedAt:   now,
	}

	if err := s.repo.CreateProfile(ctx, profile); err != nil {
		return nil, err
	}
	return profile, nil
}

// GetProfileByID retrieves a member skill profile by ID
func (s *workforceService) GetProfileByID(ctx context.Context, id string) (*model.MemberSkillProfile, error) {
	if strings.TrimSpace(id) == "" {
		return nil, ErrWFProfileNotFound
	}
	return s.repo.GetProfileByID(ctx, id)
}

// VerifyProfile records a manager/lead endorsement of a member's skill proficiency
func (s *workforceService) VerifyProfile(ctx context.Context, id string, req model.VerifyWFProfileRequest) (*model.MemberSkillProfile, error) {
	if strings.TrimSpace(id) == "" {
		return nil, ErrWFProfileNotFound
	}
	if err := req.Validate(); err != nil {
		return nil, err
	}

	profile, err := s.repo.GetProfileByID(ctx, id)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	profile.VerifiedBy = req.VerifiedBy
	profile.VerifiedAt = &now
	profile.LastAssessedAt = now
	if req.ProficiencyLevel != "" {
		profile.ProficiencyLevel = req.ProficiencyLevel
	}

	if err := s.repo.UpdateProfile(ctx, profile); err != nil {
		return nil, err
	}
	return profile, nil
}

// DeleteProfile deletes a member skill profile by ID
func (s *workforceService) DeleteProfile(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return ErrWFProfileNotFound
	}
	return s.repo.DeleteProfile(ctx, id)
}

// ListProfiles queries member skill profiles matching filter criteria
func (s *workforceService) ListProfiles(ctx context.Context, filter model.WFProfileFilter) ([]model.MemberSkillProfile, int, error) {
	return s.repo.ListProfiles(ctx, filter)
}

func evaluateCertStatus(expiresAt *time.Time, now time.Time) string {
	if expiresAt == nil {
		return model.WFCertStatusActive
	}
	if expiresAt.Before(now) {
		return model.WFCertStatusExpired
	}
	if expiresAt.Before(now.Add(30 * 24 * time.Hour)) {
		return model.WFCertStatusExpiringSoon
	}
	return model.WFCertStatusActive
}

// CreateCertification records a professional certification and evaluates its expiration status
func (s *workforceService) CreateCertification(ctx context.Context, workspaceID string, req model.CreateWFCertRequest) (*model.WorkforceCertification, error) {
	if req.WorkspaceID == "" {
		req.WorkspaceID = workspaceID
	}
	if err := req.Validate(); err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	issuedAt := now
	if req.IssuedAt != nil && !req.IssuedAt.IsZero() {
		issuedAt = req.IssuedAt.UTC()
	}
	code := strings.TrimSpace(req.CertCode)
	if code == "" {
		code = fmt.Sprintf("CERT-%s", strings.ToUpper(uuid.New().String()[:8]))
	}

	cert := &model.WorkforceCertification{
		WorkspaceID:  req.WorkspaceID,
		UserID:       req.UserID,
		CertCode:     code,
		Name:         req.Name,
		Issuer:       req.Issuer,
		CredentialID: req.CredentialID,
		IssuedAt:     issuedAt,
		ExpiresAt:    req.ExpiresAt,
		Status:       evaluateCertStatus(req.ExpiresAt, now),
	}

	if err := s.repo.CreateCertification(ctx, cert); err != nil {
		return nil, err
	}
	return cert, nil
}

// GetCertificationByID retrieves a certification by ID
func (s *workforceService) GetCertificationByID(ctx context.Context, id string) (*model.WorkforceCertification, error) {
	if strings.TrimSpace(id) == "" {
		return nil, ErrWFCertificationNotFound
	}
	return s.repo.GetCertificationByID(ctx, id)
}

// UpdateCertification updates credential details, expiry, or status on a certification
func (s *workforceService) UpdateCertification(ctx context.Context, id string, req model.UpdateWFCertRequest) (*model.WorkforceCertification, error) {
	if strings.TrimSpace(id) == "" {
		return nil, ErrWFCertificationNotFound
	}
	if err := req.Validate(); err != nil {
		return nil, err
	}

	cert, err := s.repo.GetCertificationByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.CredentialID != nil {
		cert.CredentialID = *req.CredentialID
	}
	if req.ExpiresAt != nil {
		cert.ExpiresAt = req.ExpiresAt
		cert.Status = evaluateCertStatus(req.ExpiresAt, time.Now().UTC())
	}
	if req.Status != nil {
		cert.Status = *req.Status
	}

	if err := s.repo.UpdateCertification(ctx, cert); err != nil {
		return nil, err
	}
	return cert, nil
}

// DeleteCertification deletes a certification by ID
func (s *workforceService) DeleteCertification(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return ErrWFCertificationNotFound
	}
	return s.repo.DeleteCertification(ctx, id)
}

// ListCertifications queries certifications matching filter criteria
func (s *workforceService) ListCertifications(ctx context.Context, filter model.WFCertFilter) ([]model.WorkforceCertification, int, error) {
	return s.repo.ListCertifications(ctx, filter)
}

// CreateRoleCompetency registers a skill proficiency requirement for an ITIL role
func (s *workforceService) CreateRoleCompetency(ctx context.Context, workspaceID string, req model.CreateWFRoleCompetencyRequest) (*model.RoleCompetency, error) {
	if req.WorkspaceID == "" {
		req.WorkspaceID = workspaceID
	}
	if err := req.Validate(); err != nil {
		return nil, err
	}

	if _, err := s.repo.GetSkillByID(ctx, req.SkillID); err != nil {
		return nil, err
	}

	reqLvl := req.RequiredLevel
	if reqLvl == "" {
		reqLvl = model.WFProficiencyIntermediate
	}
	minHC := req.MinHeadcount
	if minHC <= 0 {
		minHC = 1
	}
	prio := req.Priority
	if prio == "" {
		prio = model.WFPriorityMedium
	}

	rc := &model.RoleCompetency{
		WorkspaceID:   req.WorkspaceID,
		RoleName:      req.RoleName,
		SkillID:       req.SkillID,
		RequiredLevel: reqLvl,
		MinHeadcount:  minHC,
		Priority:      prio,
	}

	if err := s.repo.CreateRoleCompetency(ctx, rc); err != nil {
		return nil, err
	}
	return rc, nil
}

// GetRoleCompetencyByID retrieves a role competency by ID
func (s *workforceService) GetRoleCompetencyByID(ctx context.Context, id string) (*model.RoleCompetency, error) {
	if strings.TrimSpace(id) == "" {
		return nil, ErrWFRoleCompetencyNotFound
	}
	return s.repo.GetRoleCompetencyByID(ctx, id)
}

// DeleteRoleCompetency deletes a role competency by ID
func (s *workforceService) DeleteRoleCompetency(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return ErrWFRoleCompetencyNotFound
	}
	return s.repo.DeleteRoleCompetency(ctx, id)
}

// ListRoleCompetencies queries role competencies matching filter criteria
func (s *workforceService) ListRoleCompetencies(ctx context.Context, filter model.WFRoleCompetencyFilter) ([]model.RoleCompetency, int, error) {
	return s.repo.ListRoleCompetencies(ctx, filter)
}

// CreateTraining assigns an upskilling training course to a team member
func (s *workforceService) CreateTraining(ctx context.Context, workspaceID string, req model.CreateWFTrainingRequest) (*model.TrainingEnrollment, error) {
	if req.WorkspaceID == "" {
		req.WorkspaceID = workspaceID
	}
	if err := req.Validate(); err != nil {
		return nil, err
	}

	if _, err := s.repo.GetSkillByID(ctx, req.SkillID); err != nil {
		return nil, err
	}

	training := &model.TrainingEnrollment{
		WorkspaceID: req.WorkspaceID,
		UserID:      req.UserID,
		SkillID:     req.SkillID,
		CourseTitle: req.CourseTitle,
		Provider:    req.Provider,
		Status:      model.WFTrainingAssigned,
		ProgressPct: 0,
		DueDate:     req.DueDate,
	}

	if err := s.repo.CreateTraining(ctx, training); err != nil {
		return nil, err
	}
	return training, nil
}

// GetTrainingByID retrieves a training enrollment by ID
func (s *workforceService) GetTrainingByID(ctx context.Context, id string) (*model.TrainingEnrollment, error) {
	if strings.TrimSpace(id) == "" {
		return nil, ErrWFTrainingNotFound
	}
	return s.repo.GetTrainingByID(ctx, id)
}

// UpdateTraining advances progress or completes a training enrollment
func (s *workforceService) UpdateTraining(ctx context.Context, id string, req model.UpdateWFTrainingRequest) (*model.TrainingEnrollment, error) {
	if strings.TrimSpace(id) == "" {
		return nil, ErrWFTrainingNotFound
	}
	if err := req.Validate(); err != nil {
		return nil, err
	}

	training, err := s.repo.GetTrainingByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.ProgressPct != nil {
		training.ProgressPct = *req.ProgressPct
		if *req.ProgressPct > 0 && *req.ProgressPct < 100 && training.Status == model.WFTrainingAssigned {
			training.Status = model.WFTrainingInProgress
		}
		if *req.ProgressPct == 100 {
			training.Status = model.WFTrainingCompleted
			now := time.Now().UTC()
			training.CompletedAt = &now
		}
	}
	if req.Status != nil {
		training.Status = *req.Status
		if *req.Status == model.WFTrainingCompleted {
			training.ProgressPct = 100
			if training.CompletedAt == nil {
				now := time.Now().UTC()
				training.CompletedAt = &now
			}
		}
	}

	if err := s.repo.UpdateTraining(ctx, training); err != nil {
		return nil, err
	}
	return training, nil
}

// DeleteTraining removes a training enrollment by ID
func (s *workforceService) DeleteTraining(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return ErrWFTrainingNotFound
	}
	return s.repo.DeleteTraining(ctx, id)
}

// ListTrainings queries training enrollments matching filter criteria
func (s *workforceService) ListTrainings(ctx context.Context, filter model.WFTrainingFilter) ([]model.TrainingEnrollment, int, error) {
	return s.repo.ListTrainings(ctx, filter)
}

// GetWorkforceAnalytics computes executive workforce readiness and skill gap KPIs
func (s *workforceService) GetWorkforceAnalytics(ctx context.Context, workspaceID string) (*model.WorkforceAnalytics, error) {
	if strings.TrimSpace(workspaceID) == "" {
		return nil, errors.New("workspace_id is required")
	}
	return s.repo.GetWorkforceAnalytics(ctx, workspaceID)
}
