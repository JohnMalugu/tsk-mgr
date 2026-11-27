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

// Compliance regulatory standard types
const (
	GRCStandardSOC2     = "soc2"
	GRCStandardISO27001 = "iso27001"
	GRCStandardHIPAA    = "hipaa"
	GRCStandardPCIDSS   = "pci_dss"
	GRCStandardGDPR     = "gdpr"
	GRCStandardNISTCSF  = "nist_csf"
	GRCStandardCustom   = "custom"
)

// Compliance framework statuses
const (
	GRCFrameworkStatusDraft    = "draft"
	GRCFrameworkStatusActive   = "active"
	GRCFrameworkStatusArchived = "archived"
)

// IsValidGRCStandard validates regulatory standard type
func IsValidGRCStandard(std string) bool {
	switch std {
	case GRCStandardSOC2, GRCStandardISO27001, GRCStandardHIPAA,
		GRCStandardPCIDSS, GRCStandardGDPR, GRCStandardNISTCSF, GRCStandardCustom:
		return true
	default:
		return false
	}
}

// IsValidGRCFrameworkStatus validates framework lifecycle status
func IsValidGRCFrameworkStatus(st string) bool {
	switch st {
	case GRCFrameworkStatusDraft, GRCFrameworkStatusActive, GRCFrameworkStatusArchived:
		return true
	default:
		return false
	}
}

// ComplianceFramework represents a regulatory or internal governance standard
type ComplianceFramework struct {
	ID            string    `json:"id"`
	WorkspaceID   string    `json:"workspace_id"`
	FrameworkCode string    `json:"framework_code"`
	Name          string    `json:"name"`
	StandardType  string    `json:"standard_type"`
	Version       string    `json:"version"`
	Description   string    `json:"description,omitempty"`
	OwnerID       string    `json:"owner_id,omitempty"`
	Status        string    `json:"status"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}
