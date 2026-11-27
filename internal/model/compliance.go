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

// Compliance control domain constants
const (
	GRCDomainAccessControl      = "access_control"
	GRCDomainEncryption         = "encryption"
	GRCDomainChangeManagement   = "change_management"
	GRCDomainIncidentResponse   = "incident_response"
	GRCDomainBusinessContinuity = "business_continuity"
	GRCDomainVendorRisk         = "vendor_risk"
	GRCDomainLoggingMonitoring  = "logging_monitoring"
)

// Compliance control implementation statuses
const (
	GRCControlNotImplemented       = "not_implemented"
	GRCControlPartiallyImplemented = "partially_implemented"
	GRCControlImplemented          = "implemented"
	GRCControlNotApplicable        = "not_applicable"
)

// Compliance control automation levels
const (
	GRCAutomationManual         = "manual"
	GRCAutomationSemiAutomated  = "semi_automated"
	GRCAutomationFullyAutomated = "fully_automated"
)

// IsValidGRCDomain validates control security/operational domain
func IsValidGRCDomain(d string) bool {
	switch d {
	case GRCDomainAccessControl, GRCDomainEncryption, GRCDomainChangeManagement,
		GRCDomainIncidentResponse, GRCDomainBusinessContinuity, GRCDomainVendorRisk, GRCDomainLoggingMonitoring:
		return true
	default:
		return false
	}
}

// IsValidGRCControlStatus validates control implementation status
func IsValidGRCControlStatus(st string) bool {
	switch st {
	case GRCControlNotImplemented, GRCControlPartiallyImplemented, GRCControlImplemented, GRCControlNotApplicable:
		return true
	default:
		return false
	}
}

// IsValidGRCAutomationLevel validates control automation level
func IsValidGRCAutomationLevel(lvl string) bool {
	switch lvl {
	case GRCAutomationManual, GRCAutomationSemiAutomated, GRCAutomationFullyAutomated:
		return true
	default:
		return false
	}
}

// ComplianceControl represents an individual governance or security requirement under a framework
type ComplianceControl struct {
	ID                   string     `json:"id"`
	WorkspaceID          string     `json:"workspace_id"`
	FrameworkID          string     `json:"framework_id"`
	ControlCode          string     `json:"control_code"`
	Title                string     `json:"title"`
	Description          string     `json:"description,omitempty"`
	Domain               string     `json:"domain"`
	ImplementationStatus string     `json:"implementation_status"`
	AutomationLevel      string     `json:"automation_level"`
	OwnerID              string     `json:"owner_id,omitempty"`
	CIID                 string     `json:"ci_id,omitempty"`
	LastTestedAt         *time.Time `json:"last_tested_at,omitempty"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`
}
