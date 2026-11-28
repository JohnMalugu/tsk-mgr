package model

import (
	"errors"
	"strings"
	"time"
)

var (
	_ = errors.New
	_ = strings.TrimSpace
	_ = time.Now
)

// Workforce skill category constants
const (
	WFSkillCategoryCloud       = "cloud"
	WFSkillCategorySecurity    = "security"
	WFSkillCategoryNetworking  = "networking"
	WFSkillCategoryDatabase    = "database"
	WFSkillCategoryDevOps      = "devops"
	WFSkillCategoryITSM        = "itsm"
	WFSkillCategorySoftwareEng = "software_engineering"
	WFSkillCategoryLeadership  = "leadership"
)

// IsValidWFSkillCategory validates workforce skill category
func IsValidWFSkillCategory(cat string) bool {
	switch cat {
	case WFSkillCategoryCloud, WFSkillCategorySecurity, WFSkillCategoryNetworking,
		WFSkillCategoryDatabase, WFSkillCategoryDevOps, WFSkillCategoryITSM,
		WFSkillCategorySoftwareEng, WFSkillCategoryLeadership:
		return true
	default:
		return false
	}
}

// WorkforceSkill represents a cataloged technical or operational capability
type WorkforceSkill struct {
	ID          string    `json:"id"`
	WorkspaceID string    `json:"workspace_id"`
	SkillCode   string    `json:"skill_code"`
	Name        string    `json:"name"`
	Category    string    `json:"category"`
	Description string    `json:"description,omitempty"`
	IsCritical  bool      `json:"is_critical"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Workforce proficiency level constants
const (
	WFProficiencyNovice       = "novice"
	WFProficiencyIntermediate = "intermediate"
	WFProficiencyAdvanced     = "advanced"
	WFProficiencyExpert       = "expert"
)

// IsValidWFProficiency validates skill proficiency level
func IsValidWFProficiency(lvl string) bool {
	switch lvl {
	case WFProficiencyNovice, WFProficiencyIntermediate, WFProficiencyAdvanced, WFProficiencyExpert:
		return true
	default:
		return false
	}
}

// MemberSkillProfile records a team member's assessed proficiency and verification in a skill
type MemberSkillProfile struct {
	ID               string     `json:"id"`
	WorkspaceID      string     `json:"workspace_id"`
	UserID           string     `json:"user_id"`
	SkillID          string     `json:"skill_id"`
	ProficiencyLevel string     `json:"proficiency_level"`
	YearsExperience  float64    `json:"years_experience"`
	VerifiedBy       string     `json:"verified_by,omitempty"`
	VerifiedAt       *time.Time `json:"verified_at,omitempty"`
	LastAssessedAt   time.Time  `json:"last_assessed_at"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

// Workforce certification status constants
const (
	WFCertStatusActive       = "active"
	WFCertStatusExpiringSoon = "expiring_soon"
	WFCertStatusExpired      = "expired"
	WFCertStatusRevoked      = "revoked"
)

// IsValidWFCertStatus validates certification status
func IsValidWFCertStatus(st string) bool {
	switch st {
	case WFCertStatusActive, WFCertStatusExpiringSoon, WFCertStatusExpired, WFCertStatusRevoked:
		return true
	default:
		return false
	}
}

// WorkforceCertification tracks professional credentials and expiration dates for staff
type WorkforceCertification struct {
	ID           string     `json:"id"`
	WorkspaceID  string     `json:"workspace_id"`
	UserID       string     `json:"user_id"`
	CertCode     string     `json:"cert_code"`
	Name         string     `json:"name"`
	Issuer       string     `json:"issuer"`
	CredentialID string     `json:"credential_id,omitempty"`
	IssuedAt     time.Time  `json:"issued_at"`
	ExpiresAt    *time.Time `json:"expires_at,omitempty"`
	Status       string     `json:"status"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

// Role competency priority constants
const (
	WFPriorityCritical = "critical"
	WFPriorityHigh     = "high"
	WFPriorityMedium   = "medium"
	WFPriorityLow      = "low"
)

// IsValidWFPriority validates role competency priority
func IsValidWFPriority(p string) bool {
	switch p {
	case WFPriorityCritical, WFPriorityHigh, WFPriorityMedium, WFPriorityLow:
		return true
	default:
		return false
	}
}

// RoleCompetency defines required skill proficiency and headcount benchmarks for an ITIL role
type RoleCompetency struct {
	ID            string    `json:"id"`
	WorkspaceID   string    `json:"workspace_id"`
	RoleName      string    `json:"role_name"`
	SkillID       string    `json:"skill_id"`
	RequiredLevel string    `json:"required_level"`
	MinHeadcount  int       `json:"min_headcount"`
	Priority      string    `json:"priority"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// Training enrollment lifecycle statuses
const (
	WFTrainingAssigned   = "assigned"
	WFTrainingInProgress = "in_progress"
	WFTrainingCompleted  = "completed"
	WFTrainingDropped    = "dropped"
)

// IsValidWFTrainingStatus validates training enrollment status
func IsValidWFTrainingStatus(st string) bool {
	switch st {
	case WFTrainingAssigned, WFTrainingInProgress, WFTrainingCompleted, WFTrainingDropped:
		return true
	default:
		return false
	}
}

// TrainingEnrollment tracks upskilling course assignments and completion progress
type TrainingEnrollment struct {
	ID          string     `json:"id"`
	WorkspaceID string     `json:"workspace_id"`
	UserID      string     `json:"user_id"`
	SkillID     string     `json:"skill_id"`
	CourseTitle string     `json:"course_title"`
	Provider    string     `json:"provider,omitempty"`
	Status      string     `json:"status"`
	ProgressPct int        `json:"progress_pct"`
	DueDate     *time.Time `json:"due_date,omitempty"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// Request DTOs and validation rules for Workforce & Talent Management API

// CreateWFSkillRequest payload for adding a skill to the catalog
type CreateWFSkillRequest struct {
	WorkspaceID string `json:"workspace_id"`
	SkillCode   string `json:"skill_code"`
	Name        string `json:"name"`
	Category    string `json:"category"`
	Description string `json:"description,omitempty"`
	IsCritical  bool   `json:"is_critical"`
}

func (r *CreateWFSkillRequest) Validate() error {
	if strings.TrimSpace(r.Name) == "" {
		return errors.New("name is required")
	}
	if r.Category != "" && !IsValidWFSkillCategory(r.Category) {
		return errors.New("invalid category")
	}
	return nil
}

// UpdateWFSkillRequest payload for updating a cataloged skill
type UpdateWFSkillRequest struct {
	Name        *string `json:"name,omitempty"`
	Category    *string `json:"category,omitempty"`
	Description *string `json:"description,omitempty"`
	IsCritical  *bool   `json:"is_critical,omitempty"`
}

func (r *UpdateWFSkillRequest) Validate() error {
	if r.Name != nil && strings.TrimSpace(*r.Name) == "" {
		return errors.New("name cannot be empty")
	}
	if r.Category != nil && !IsValidWFSkillCategory(*r.Category) {
		return errors.New("invalid category")
	}
	return nil
}

// UpsertWFProfileRequest payload for recording a team member's proficiency in a skill
type UpsertWFProfileRequest struct {
	WorkspaceID      string  `json:"workspace_id"`
	UserID           string  `json:"user_id"`
	SkillID          string  `json:"skill_id"`
	ProficiencyLevel string  `json:"proficiency_level"`
	YearsExperience  float64 `json:"years_experience"`
}

func (r *UpsertWFProfileRequest) Validate() error {
	if strings.TrimSpace(r.UserID) == "" {
		return errors.New("user_id is required")
	}
	if strings.TrimSpace(r.SkillID) == "" {
		return errors.New("skill_id is required")
	}
	if r.ProficiencyLevel != "" && !IsValidWFProficiency(r.ProficiencyLevel) {
		return errors.New("invalid proficiency_level")
	}
	if r.YearsExperience < 0 {
		return errors.New("years_experience cannot be negative")
	}
	return nil
}

// VerifyWFProfileRequest payload for manager/lead endorsement of a member's skill profile
type VerifyWFProfileRequest struct {
	VerifiedBy       string `json:"verified_by"`
	ProficiencyLevel string `json:"proficiency_level,omitempty"`
}

func (r *VerifyWFProfileRequest) Validate() error {
	if strings.TrimSpace(r.VerifiedBy) == "" {
		return errors.New("verified_by is required")
	}
	if r.ProficiencyLevel != "" && !IsValidWFProficiency(r.ProficiencyLevel) {
		return errors.New("invalid proficiency_level")
	}
	return nil
}

// CreateWFCertRequest payload for registering a professional certification
type CreateWFCertRequest struct {
	WorkspaceID  string     `json:"workspace_id"`
	UserID       string     `json:"user_id"`
	CertCode     string     `json:"cert_code"`
	Name         string     `json:"name"`
	Issuer       string     `json:"issuer"`
	CredentialID string     `json:"credential_id,omitempty"`
	IssuedAt     *time.Time `json:"issued_at,omitempty"`
	ExpiresAt    *time.Time `json:"expires_at,omitempty"`
}

func (r *CreateWFCertRequest) Validate() error {
	if strings.TrimSpace(r.UserID) == "" {
		return errors.New("user_id is required")
	}
	if strings.TrimSpace(r.Name) == "" {
		return errors.New("name is required")
	}
	if strings.TrimSpace(r.Issuer) == "" {
		return errors.New("issuer is required")
	}
	return nil
}

// UpdateWFCertRequest payload for updating certification status or renewal dates
type UpdateWFCertRequest struct {
	CredentialID *string    `json:"credential_id,omitempty"`
	ExpiresAt    *time.Time `json:"expires_at,omitempty"`
	Status       *string    `json:"status,omitempty"`
}

func (r *UpdateWFCertRequest) Validate() error {
	if r.Status != nil && !IsValidWFCertStatus(*r.Status) {
		return errors.New("invalid status")
	}
	return nil
}

// CreateWFRoleCompetencyRequest payload for defining a role competency requirement
type CreateWFRoleCompetencyRequest struct {
	WorkspaceID   string `json:"workspace_id"`
	RoleName      string `json:"role_name"`
	SkillID       string `json:"skill_id"`
	RequiredLevel string `json:"required_level"`
	MinHeadcount  int    `json:"min_headcount"`
	Priority      string `json:"priority"`
}

func (r *CreateWFRoleCompetencyRequest) Validate() error {
	if strings.TrimSpace(r.RoleName) == "" {
		return errors.New("role_name is required")
	}
	if strings.TrimSpace(r.SkillID) == "" {
		return errors.New("skill_id is required")
	}
	if r.RequiredLevel != "" && !IsValidWFProficiency(r.RequiredLevel) {
		return errors.New("invalid required_level")
	}
	if r.Priority != "" && !IsValidWFPriority(r.Priority) {
		return errors.New("invalid priority")
	}
	if r.MinHeadcount < 0 {
		return errors.New("min_headcount cannot be negative")
	}
	return nil
}

// CreateWFTrainingRequest payload for assigning a training enrollment
type CreateWFTrainingRequest struct {
	WorkspaceID string     `json:"workspace_id"`
	UserID      string     `json:"user_id"`
	SkillID     string     `json:"skill_id"`
	CourseTitle string     `json:"course_title"`
	Provider    string     `json:"provider,omitempty"`
	DueDate     *time.Time `json:"due_date,omitempty"`
}

func (r *CreateWFTrainingRequest) Validate() error {
	if strings.TrimSpace(r.UserID) == "" {
		return errors.New("user_id is required")
	}
	if strings.TrimSpace(r.SkillID) == "" {
		return errors.New("skill_id is required")
	}
	if strings.TrimSpace(r.CourseTitle) == "" {
		return errors.New("course_title is required")
	}
	return nil
}

// UpdateWFTrainingRequest payload for advancing training progress or completion
type UpdateWFTrainingRequest struct {
	Status      *string `json:"status,omitempty"`
	ProgressPct *int    `json:"progress_pct,omitempty"`
}

func (r *UpdateWFTrainingRequest) Validate() error {
	if r.Status != nil && !IsValidWFTrainingStatus(*r.Status) {
		return errors.New("invalid status")
	}
	if r.ProgressPct != nil && (*r.ProgressPct < 0 || *r.ProgressPct > 100) {
		return errors.New("progress_pct must be between 0 and 100")
	}
	return nil
}

// Query filters and executive readiness analytics models for Workforce & Talent Management

// WFSkillFilter defines query criteria for the skill catalog
type WFSkillFilter struct {
	WorkspaceID string
	Category    string
	IsCritical  *bool
	Search      string
	Limit       int
	Offset      int
}

// WFProfileFilter defines query criteria for member skill profiles
type WFProfileFilter struct {
	WorkspaceID      string
	UserID           string
	SkillID          string
	ProficiencyLevel string
	VerifiedOnly     bool
	Limit            int
	Offset           int
}

// WFCertFilter defines query criteria for professional certifications
type WFCertFilter struct {
	WorkspaceID string
	UserID      string
	Issuer      string
	Status      string
	Limit       int
	Offset      int
}

// WFRoleCompetencyFilter defines query criteria for role competency requirements
type WFRoleCompetencyFilter struct {
	WorkspaceID string
	RoleName    string
	SkillID     string
	Priority    string
	Limit       int
	Offset      int
}

// WFTrainingFilter defines query criteria for training enrollments
type WFTrainingFilter struct {
	WorkspaceID string
	UserID      string
	SkillID     string
	Status      string
	Limit       int
	Offset      int
}

// WorkforceAnalytics represents executive Workforce Readiness, certification compliance, and skill coverage KPIs
type WorkforceAnalytics struct {
	TotalSkills             int            `json:"total_skills"`
	CriticalSkills          int            `json:"critical_skills"`
	TotalSkillProfiles      int            `json:"total_skill_profiles"`
	VerifiedProfiles        int            `json:"verified_profiles"`
	ActiveCertifications    int            `json:"active_certifications"`
	ExpiringCertifications  int            `json:"expiring_certifications"`
	ExpiredCertifications   int            `json:"expired_certifications"`
	TotalRoleCompetencies   int            `json:"total_role_competencies"`
	ActiveTrainings         int            `json:"active_trainings"`
	CompletedTrainings      int            `json:"completed_trainings"`
	TrainingCompletionRate  float64        `json:"training_completion_rate"`
	CategoryDistribution    map[string]int `json:"category_distribution"`
	ProficiencyDistribution map[string]int `json:"proficiency_distribution"`
	WorkforceReadinessIndex float64        `json:"workforce_readiness_index"`
}
