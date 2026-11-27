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

// Audit evidence artifact types
const (
	GRCEvidenceScreenshot     = "screenshot"
	GRCEvidenceLogExport      = "log_export"
	GRCEvidencePolicyDoc      = "policy_doc"
	GRCEvidenceConfigSnapshot = "config_snapshot"
	GRCEvidencePenTestReport  = "pen_test_report"
	GRCEvidenceAttestation    = "attestation"
)

// Audit evidence review statuses
const (
	GRCEvidenceStatusPending  = "pending"
	GRCEvidenceStatusApproved = "approved"
	GRCEvidenceStatusRejected = "rejected"
	GRCEvidenceStatusExpired  = "expired"
)

// IsValidGRCEvidenceType validates evidence artifact type
func IsValidGRCEvidenceType(et string) bool {
	switch et {
	case GRCEvidenceScreenshot, GRCEvidenceLogExport, GRCEvidencePolicyDoc,
		GRCEvidenceConfigSnapshot, GRCEvidencePenTestReport, GRCEvidenceAttestation:
		return true
	default:
		return false
	}
}

// IsValidGRCEvidenceStatus validates evidence review status
func IsValidGRCEvidenceStatus(st string) bool {
	switch st {
	case GRCEvidenceStatusPending, GRCEvidenceStatusApproved, GRCEvidenceStatusRejected, GRCEvidenceStatusExpired:
		return true
	default:
		return false
	}
}

// ComplianceEvidence represents verifiable audit proof collected for a compliance control
type ComplianceEvidence struct {
	ID           string     `json:"id"`
	WorkspaceID  string     `json:"workspace_id"`
	ControlID    string     `json:"control_id"`
	Title        string     `json:"title"`
	EvidenceType string     `json:"evidence_type"`
	ArtifactURL  string     `json:"artifact_url"`
	CollectedBy  string     `json:"collected_by"`
	ValidFrom    time.Time  `json:"valid_from"`
	ValidUntil   *time.Time `json:"valid_until,omitempty"`
	ReviewStatus string     `json:"review_status"`
	ReviewedBy   string     `json:"reviewed_by,omitempty"`
	ReviewedAt   *time.Time `json:"reviewed_at,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
}

// Compliance audit type constants
const (
	GRCAuditTypeInternal      = "internal"
	GRCAuditTypeExternal      = "external"
	GRCAuditTypeCertification = "certification"
	GRCAuditTypeReadiness     = "readiness"
)

// Compliance assessment lifecycle statuses
const (
	GRCAssessmentPlanned    = "planned"
	GRCAssessmentInProgress = "in_progress"
	GRCAssessmentCompleted  = "completed"
	GRCAssessmentCancelled  = "cancelled"
)

// IsValidGRCAuditType validates audit type
func IsValidGRCAuditType(at string) bool {
	switch at {
	case GRCAuditTypeInternal, GRCAuditTypeExternal, GRCAuditTypeCertification, GRCAuditTypeReadiness:
		return true
	default:
		return false
	}
}

// IsValidGRCAssessmentStatus validates assessment status
func IsValidGRCAssessmentStatus(st string) bool {
	switch st {
	case GRCAssessmentPlanned, GRCAssessmentInProgress, GRCAssessmentCompleted, GRCAssessmentCancelled:
		return true
	default:
		return false
	}
}

// ComplianceAssessment represents a formal compliance audit or readiness review against a framework
type ComplianceAssessment struct {
	ID             string     `json:"id"`
	WorkspaceID    string     `json:"workspace_id"`
	FrameworkID    string     `json:"framework_id"`
	AssessmentCode string     `json:"assessment_code"`
	Title          string     `json:"title"`
	AuditorName    string     `json:"auditor_name"`
	AuditType      string     `json:"audit_type"`
	Status         string     `json:"status"`
	StartDate      time.Time  `json:"start_date"`
	EndDate        *time.Time `json:"end_date,omitempty"`
	OverallScore   float64    `json:"overall_score"`
	Summary        string     `json:"summary,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// Compliance finding severity ratings
const (
	GRCFindingSeverityCritical = "critical"
	GRCFindingSeverityHigh     = "high"
	GRCFindingSeverityMedium   = "medium"
	GRCFindingSeverityLow      = "low"
)

// Compliance finding remediation statuses
const (
	GRCFindingStatusOpen          = "open"
	GRCFindingStatusInRemediation = "in_remediation"
	GRCFindingStatusRemediated    = "remediated"
	GRCFindingStatusAcceptedRisk  = "accepted_risk"
)

// IsValidGRCFindingSeverity validates finding severity
func IsValidGRCFindingSeverity(sev string) bool {
	switch sev {
	case GRCFindingSeverityCritical, GRCFindingSeverityHigh, GRCFindingSeverityMedium, GRCFindingSeverityLow:
		return true
	default:
		return false
	}
}

// IsValidGRCFindingStatus validates finding remediation status
func IsValidGRCFindingStatus(st string) bool {
	switch st {
	case GRCFindingStatusOpen, GRCFindingStatusInRemediation, GRCFindingStatusRemediated, GRCFindingStatusAcceptedRisk:
		return true
	default:
		return false
	}
}

// ComplianceFinding represents an audit non-conformity or control gap requiring remediation
type ComplianceFinding struct {
	ID              string     `json:"id"`
	WorkspaceID     string     `json:"workspace_id"`
	AssessmentID    string     `json:"assessment_id"`
	ControlID       string     `json:"control_id"`
	FindingCode     string     `json:"finding_code"`
	Title           string     `json:"title"`
	Description     string     `json:"description,omitempty"`
	Severity        string     `json:"severity"`
	Status          string     `json:"status"`
	RemediationPlan string     `json:"remediation_plan,omitempty"`
	OwnerID         string     `json:"owner_id,omitempty"`
	DueDate         *time.Time `json:"due_date,omitempty"`
	RemediatedAt    *time.Time `json:"remediated_at,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

// Request DTOs and validation rules for Compliance & GRC Management API

// CreateGRCFrameworkRequest payload for registering a regulatory compliance framework
type CreateGRCFrameworkRequest struct {
	WorkspaceID   string `json:"workspace_id"`
	FrameworkCode string `json:"framework_code"`
	Name          string `json:"name"`
	StandardType  string `json:"standard_type"`
	Version       string `json:"version"`
	Description   string `json:"description,omitempty"`
	OwnerID       string `json:"owner_id,omitempty"`
}

func (r *CreateGRCFrameworkRequest) Validate() error {
	if strings.TrimSpace(r.Name) == "" {
		return errors.New("name is required")
	}
	if r.StandardType != "" && !IsValidGRCStandard(r.StandardType) {
		return errors.New("invalid standard_type")
	}
	return nil
}

// UpdateGRCFrameworkRequest payload for modifying a compliance framework
type UpdateGRCFrameworkRequest struct {
	Name        *string `json:"name,omitempty"`
	Version     *string `json:"version,omitempty"`
	Description *string `json:"description,omitempty"`
	OwnerID     *string `json:"owner_id,omitempty"`
	Status      *string `json:"status,omitempty"`
}

func (r *UpdateGRCFrameworkRequest) Validate() error {
	if r.Name != nil && strings.TrimSpace(*r.Name) == "" {
		return errors.New("name cannot be empty")
	}
	if r.Status != nil && !IsValidGRCFrameworkStatus(*r.Status) {
		return errors.New("invalid status")
	}
	return nil
}

// CreateGRCControlRequest payload for adding a compliance control under a framework
type CreateGRCControlRequest struct {
	WorkspaceID          string `json:"workspace_id"`
	FrameworkID          string `json:"framework_id"`
	ControlCode          string `json:"control_code"`
	Title                string `json:"title"`
	Description          string `json:"description,omitempty"`
	Domain               string `json:"domain"`
	ImplementationStatus string `json:"implementation_status"`
	AutomationLevel      string `json:"automation_level"`
	OwnerID              string `json:"owner_id,omitempty"`
	CIID                 string `json:"ci_id,omitempty"`
}

func (r *CreateGRCControlRequest) Validate() error {
	if strings.TrimSpace(r.FrameworkID) == "" {
		return errors.New("framework_id is required")
	}
	if strings.TrimSpace(r.Title) == "" {
		return errors.New("title is required")
	}
	if r.Domain != "" && !IsValidGRCDomain(r.Domain) {
		return errors.New("invalid domain")
	}
	if r.ImplementationStatus != "" && !IsValidGRCControlStatus(r.ImplementationStatus) {
		return errors.New("invalid implementation_status")
	}
	if r.AutomationLevel != "" && !IsValidGRCAutomationLevel(r.AutomationLevel) {
		return errors.New("invalid automation_level")
	}
	return nil
}

// UpdateGRCControlRequest payload for updating a compliance control
type UpdateGRCControlRequest struct {
	Title                *string `json:"title,omitempty"`
	Description          *string `json:"description,omitempty"`
	Domain               *string `json:"domain,omitempty"`
	ImplementationStatus *string `json:"implementation_status,omitempty"`
	AutomationLevel      *string `json:"automation_level,omitempty"`
	OwnerID              *string `json:"owner_id,omitempty"`
	CIID                 *string `json:"ci_id,omitempty"`
}

func (r *UpdateGRCControlRequest) Validate() error {
	if r.Title != nil && strings.TrimSpace(*r.Title) == "" {
		return errors.New("title cannot be empty")
	}
	if r.Domain != nil && !IsValidGRCDomain(*r.Domain) {
		return errors.New("invalid domain")
	}
	if r.ImplementationStatus != nil && !IsValidGRCControlStatus(*r.ImplementationStatus) {
		return errors.New("invalid implementation_status")
	}
	if r.AutomationLevel != nil && !IsValidGRCAutomationLevel(*r.AutomationLevel) {
		return errors.New("invalid automation_level")
	}
	return nil
}

// SubmitGRCEvidenceRequest payload for attaching audit evidence to a control
type SubmitGRCEvidenceRequest struct {
	WorkspaceID  string     `json:"workspace_id"`
	ControlID    string     `json:"control_id"`
	Title        string     `json:"title"`
	EvidenceType string     `json:"evidence_type"`
	ArtifactURL  string     `json:"artifact_url"`
	CollectedBy  string     `json:"collected_by"`
	ValidFrom    *time.Time `json:"valid_from,omitempty"`
	ValidUntil   *time.Time `json:"valid_until,omitempty"`
}

func (r *SubmitGRCEvidenceRequest) Validate() error {
	if strings.TrimSpace(r.ControlID) == "" {
		return errors.New("control_id is required")
	}
	if strings.TrimSpace(r.Title) == "" {
		return errors.New("title is required")
	}
	if strings.TrimSpace(r.ArtifactURL) == "" {
		return errors.New("artifact_url is required")
	}
	if r.EvidenceType != "" && !IsValidGRCEvidenceType(r.EvidenceType) {
		return errors.New("invalid evidence_type")
	}
	return nil
}

// ReviewGRCEvidenceRequest payload for approving or rejecting audit evidence
type ReviewGRCEvidenceRequest struct {
	ReviewStatus string `json:"review_status"`
	ReviewedBy   string `json:"reviewed_by"`
}

func (r *ReviewGRCEvidenceRequest) Validate() error {
	if !IsValidGRCEvidenceStatus(r.ReviewStatus) {
		return errors.New("invalid review_status")
	}
	if strings.TrimSpace(r.ReviewedBy) == "" {
		return errors.New("reviewed_by is required")
	}
	return nil
}

// CreateGRCAssessmentRequest payload for scheduling a compliance audit assessment
type CreateGRCAssessmentRequest struct {
	WorkspaceID    string     `json:"workspace_id"`
	FrameworkID    string     `json:"framework_id"`
	AssessmentCode string     `json:"assessment_code,omitempty"`
	Title          string     `json:"title"`
	AuditorName    string     `json:"auditor_name"`
	AuditType      string     `json:"audit_type"`
	StartDate      *time.Time `json:"start_date,omitempty"`
	Summary        string     `json:"summary,omitempty"`
}

func (r *CreateGRCAssessmentRequest) Validate() error {
	if strings.TrimSpace(r.FrameworkID) == "" {
		return errors.New("framework_id is required")
	}
	if strings.TrimSpace(r.Title) == "" {
		return errors.New("title is required")
	}
	if r.AuditType != "" && !IsValidGRCAuditType(r.AuditType) {
		return errors.New("invalid audit_type")
	}
	return nil
}

// CompleteGRCAssessmentRequest payload for closing an assessment with a final score
type CompleteGRCAssessmentRequest struct {
	OverallScore float64 `json:"overall_score"`
	Summary      string  `json:"summary,omitempty"`
}

func (r *CompleteGRCAssessmentRequest) Validate() error {
	if r.OverallScore < 0 || r.OverallScore > 100 {
		return errors.New("overall_score must be between 0 and 100")
	}
	return nil
}

// CreateGRCFindingRequest payload for logging an audit non-conformity finding
type CreateGRCFindingRequest struct {
	WorkspaceID     string     `json:"workspace_id"`
	AssessmentID    string     `json:"assessment_id"`
	ControlID       string     `json:"control_id"`
	FindingCode     string     `json:"finding_code,omitempty"`
	Title           string     `json:"title"`
	Description     string     `json:"description,omitempty"`
	Severity        string     `json:"severity"`
	RemediationPlan string     `json:"remediation_plan,omitempty"`
	OwnerID         string     `json:"owner_id,omitempty"`
	DueDate         *time.Time `json:"due_date,omitempty"`
}

func (r *CreateGRCFindingRequest) Validate() error {
	if strings.TrimSpace(r.AssessmentID) == "" {
		return errors.New("assessment_id is required")
	}
	if strings.TrimSpace(r.ControlID) == "" {
		return errors.New("control_id is required")
	}
	if strings.TrimSpace(r.Title) == "" {
		return errors.New("title is required")
	}
	if r.Severity != "" && !IsValidGRCFindingSeverity(r.Severity) {
		return errors.New("invalid severity")
	}
	return nil
}

// UpdateGRCFindingRequest payload for advancing remediation on an audit finding
type UpdateGRCFindingRequest struct {
	Status          *string    `json:"status,omitempty"`
	Severity        *string    `json:"severity,omitempty"`
	RemediationPlan *string    `json:"remediation_plan,omitempty"`
	OwnerID         *string    `json:"owner_id,omitempty"`
	DueDate         *time.Time `json:"due_date,omitempty"`
}

func (r *UpdateGRCFindingRequest) Validate() error {
	if r.Status != nil && !IsValidGRCFindingStatus(*r.Status) {
		return errors.New("invalid status")
	}
	if r.Severity != nil && !IsValidGRCFindingSeverity(*r.Severity) {
		return errors.New("invalid severity")
	}
	return nil
}
