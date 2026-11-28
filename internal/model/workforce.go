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
