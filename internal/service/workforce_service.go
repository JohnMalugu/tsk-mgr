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

// Certifications stubs
func (s *workforceService) CreateCertification(ctx context.Context, workspaceID string, req model.CreateWFCertRequest) (*model.WorkforceCertification, error) {
	return nil, errors.New("not implemented")
}
func (s *workforceService) GetCertificationByID(ctx context.Context, id string) (*model.WorkforceCertification, error) {
	return nil, errors.New("not implemented")
}
func (s *workforceService) UpdateCertification(ctx context.Context, id string, req model.UpdateWFCertRequest) (*model.WorkforceCertification, error) {
	return nil, errors.New("not implemented")
}
func (s *workforceService) DeleteCertification(ctx context.Context, id string) error {
	return errors.New("not implemented")
}
func (s *workforceService) ListCertifications(ctx context.Context, filter model.WFCertFilter) ([]model.WorkforceCertification, int, error) {
	return nil, 0, errors.New("not implemented")
}

// Role Competencies stubs
func (s *workforceService) CreateRoleCompetency(ctx context.Context, workspaceID string, req model.CreateWFRoleCompetencyRequest) (*model.RoleCompetency, error) {
	return nil, errors.New("not implemented")
}
func (s *workforceService) GetRoleCompetencyByID(ctx context.Context, id string) (*model.RoleCompetency, error) {
	return nil, errors.New("not implemented")
}
func (s *workforceService) DeleteRoleCompetency(ctx context.Context, id string) error {
	return errors.New("not implemented")
}
func (s *workforceService) ListRoleCompetencies(ctx context.Context, filter model.WFRoleCompetencyFilter) ([]model.RoleCompetency, int, error) {
	return nil, 0, errors.New("not implemented")
}

// Trainings stubs
func (s *workforceService) CreateTraining(ctx context.Context, workspaceID string, req model.CreateWFTrainingRequest) (*model.TrainingEnrollment, error) {
	return nil, errors.New("not implemented")
}
func (s *workforceService) GetTrainingByID(ctx context.Context, id string) (*model.TrainingEnrollment, error) {
	return nil, errors.New("not implemented")
}
func (s *workforceService) UpdateTraining(ctx context.Context, id string, req model.UpdateWFTrainingRequest) (*model.TrainingEnrollment, error) {
	return nil, errors.New("not implemented")
}
func (s *workforceService) DeleteTraining(ctx context.Context, id string) error {
	return errors.New("not implemented")
}
func (s *workforceService) ListTrainings(ctx context.Context, filter model.WFTrainingFilter) ([]model.TrainingEnrollment, int, error) {
	return nil, 0, errors.New("not implemented")
}

// Analytics stub
func (s *workforceService) GetWorkforceAnalytics(ctx context.Context, workspaceID string) (*model.WorkforceAnalytics, error) {
	return nil, errors.New("not implemented")
}
