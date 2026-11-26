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

// Service Desk support tier constants
const (
	SDTier1          = "tier_1"
	SDTier2          = "tier_2"
	SDTier3          = "tier_3"
	SDTierSpecialist = "specialist"
)

// IsValidSDTier validates service desk support tier
func IsValidSDTier(tier string) bool {
	switch tier {
	case SDTier1, SDTier2, SDTier3, SDTierSpecialist:
		return true
	default:
		return false
	}
}

// ServiceDeskQueue represents a support routing queue or desk team
type ServiceDeskQueue struct {
	ID                string    `json:"id"`
	WorkspaceID       string    `json:"workspace_id"`
	QueueCode         string    `json:"queue_code"`
	Name              string    `json:"name"`
	Description       string    `json:"description,omitempty"`
	Tier              string    `json:"tier"`
	DefaultSLAMinutes int       `json:"default_sla_minutes"`
	IsActive          bool      `json:"is_active"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}
