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
