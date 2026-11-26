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

// Service Desk omnichannel contact methods
const (
	SDChannelPortal = "portal"
	SDChannelEmail  = "email"
	SDChannelChat   = "chat"
	SDChannelPhone  = "phone"
	SDChannelSlack  = "slack"
	SDChannelWalkUp = "walk_up"
)

// Service Desk ticket classifications
const (
	SDTicketTypeIncidentReport = "incident_report"
	SDTicketTypeServiceRequest = "service_request"
	SDTicketTypeInquiry        = "inquiry"
	SDTicketTypeComplaint      = "complaint"
)

// Service Desk ticket priority ratings
const (
	SDPriorityP1 = "p1"
	SDPriorityP2 = "p2"
	SDPriorityP3 = "p3"
	SDPriorityP4 = "p4"
)

// Service Desk ticket lifecycle statuses
const (
	SDTicketStatusNew         = "new"
	SDTicketStatusTriaged     = "triaged"
	SDTicketStatusInProgress  = "in_progress"
	SDTicketStatusPendingUser = "pending_user"
	SDTicketStatusResolved    = "resolved"
	SDTicketStatusClosed      = "closed"
)

// IsValidSDChannel validates omnichannel source
func IsValidSDChannel(ch string) bool {
	switch ch {
	case SDChannelPortal, SDChannelEmail, SDChannelChat, SDChannelPhone, SDChannelSlack, SDChannelWalkUp:
		return true
	default:
		return false
	}
}

// IsValidSDTicketType validates ticket classification
func IsValidSDTicketType(tt string) bool {
	switch tt {
	case SDTicketTypeIncidentReport, SDTicketTypeServiceRequest, SDTicketTypeInquiry, SDTicketTypeComplaint:
		return true
	default:
		return false
	}
}

// IsValidSDPriority validates ticket priority
func IsValidSDPriority(p string) bool {
	switch p {
	case SDPriorityP1, SDPriorityP2, SDPriorityP3, SDPriorityP4:
		return true
	default:
		return false
	}
}

// IsValidSDTicketStatus validates ticket lifecycle status
func IsValidSDTicketStatus(st string) bool {
	switch st {
	case SDTicketStatusNew, SDTicketStatusTriaged, SDTicketStatusInProgress,
		SDTicketStatusPendingUser, SDTicketStatusResolved, SDTicketStatusClosed:
		return true
	default:
		return false
	}
}

// ServiceDeskTicket represents a user support interaction captured at the Single Point of Contact (SPOC)
type ServiceDeskTicket struct {
	ID                   string     `json:"id"`
	WorkspaceID          string     `json:"workspace_id"`
	TicketNumber         string     `json:"ticket_number"`
	QueueID              string     `json:"queue_id,omitempty"`
	RequesterID          string     `json:"requester_id"`
	AssigneeID           string     `json:"assignee_id,omitempty"`
	Channel              string     `json:"channel"`
	TicketType           string     `json:"ticket_type"`
	Priority             string     `json:"priority"`
	Status               string     `json:"status"`
	Subject              string     `json:"subject"`
	Description          string     `json:"description,omitempty"`
	FirstContactResolved bool       `json:"first_contact_resolved"`
	FirstResponseAt      *time.Time `json:"first_response_at,omitempty"`
	ResolvedAt           *time.Time `json:"resolved_at,omitempty"`
	LinkedIncidentID     string     `json:"linked_incident_id,omitempty"`
	LinkedKBArticleID    string     `json:"linked_kb_article_id,omitempty"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`
}

// Service Desk interaction direction constants
const (
	SDDirectionInbound      = "inbound"
	SDDirectionOutbound     = "outbound"
	SDDirectionInternalNote = "internal_note"
)

// IsValidSDDirection validates interaction direction
func IsValidSDDirection(dir string) bool {
	switch dir {
	case SDDirectionInbound, SDDirectionOutbound, SDDirectionInternalNote:
		return true
	default:
		return false
	}
}

// TicketInteraction represents an omnichannel message or agent note on a support ticket
type TicketInteraction struct {
	ID          string    `json:"id"`
	WorkspaceID string    `json:"workspace_id"`
	TicketID    string    `json:"ticket_id"`
	AuthorID    string    `json:"author_id"`
	Channel     string    `json:"channel"`
	Direction   string    `json:"direction"`
	Body        string    `json:"body"`
	CreatedAt   time.Time `json:"created_at"`
}

// CSATSurvey records post-resolution Customer Satisfaction (1-5) and Net Promoter Score (0-10)
type CSATSurvey struct {
	ID           string    `json:"id"`
	WorkspaceID  string    `json:"workspace_id"`
	TicketID     string    `json:"ticket_id"`
	RespondentID string    `json:"respondent_id"`
	CSATScore    int       `json:"csat_score"`
	NPSScore     int       `json:"nps_score"`
	Comment      string    `json:"comment,omitempty"`
	SubmittedAt  time.Time `json:"submitted_at"`
}

// Request DTOs and validation rules for Service Desk Management API

// CreateSDQueueRequest payload for creating a service desk routing queue
type CreateSDQueueRequest struct {
	WorkspaceID       string `json:"workspace_id"`
	QueueCode         string `json:"queue_code"`
	Name              string `json:"name"`
	Description       string `json:"description,omitempty"`
	Tier              string `json:"tier"`
	DefaultSLAMinutes int    `json:"default_sla_minutes"`
}

func (r *CreateSDQueueRequest) Validate() error {
	if strings.TrimSpace(r.Name) == "" {
		return errors.New("name is required")
	}
	if r.Tier != "" && !IsValidSDTier(r.Tier) {
		return errors.New("invalid tier")
	}
	if r.DefaultSLAMinutes < 0 {
		return errors.New("default_sla_minutes cannot be negative")
	}
	return nil
}

// UpdateSDQueueRequest payload for updating a service desk routing queue
type UpdateSDQueueRequest struct {
	Name              *string `json:"name,omitempty"`
	Description       *string `json:"description,omitempty"`
	Tier              *string `json:"tier,omitempty"`
	DefaultSLAMinutes *int    `json:"default_sla_minutes,omitempty"`
	IsActive          *bool   `json:"is_active,omitempty"`
}

func (r *UpdateSDQueueRequest) Validate() error {
	if r.Name != nil && strings.TrimSpace(*r.Name) == "" {
		return errors.New("name cannot be empty")
	}
	if r.Tier != nil && !IsValidSDTier(*r.Tier) {
		return errors.New("invalid tier")
	}
	if r.DefaultSLAMinutes != nil && *r.DefaultSLAMinutes < 0 {
		return errors.New("default_sla_minutes cannot be negative")
	}
	return nil
}

// CreateSDTicketRequest payload for capturing a new omnichannel support ticket
type CreateSDTicketRequest struct {
	WorkspaceID       string `json:"workspace_id"`
	TicketNumber      string `json:"ticket_number,omitempty"`
	QueueID           string `json:"queue_id,omitempty"`
	RequesterID       string `json:"requester_id"`
	AssigneeID        string `json:"assignee_id,omitempty"`
	Channel           string `json:"channel"`
	TicketType        string `json:"ticket_type"`
	Priority          string `json:"priority"`
	Subject           string `json:"subject"`
	Description       string `json:"description,omitempty"`
	LinkedIncidentID  string `json:"linked_incident_id,omitempty"`
	LinkedKBArticleID string `json:"linked_kb_article_id,omitempty"`
}

func (r *CreateSDTicketRequest) Validate() error {
	if strings.TrimSpace(r.Subject) == "" {
		return errors.New("subject is required")
	}
	if r.Channel != "" && !IsValidSDChannel(r.Channel) {
		return errors.New("invalid channel")
	}
	if r.TicketType != "" && !IsValidSDTicketType(r.TicketType) {
		return errors.New("invalid ticket_type")
	}
	if r.Priority != "" && !IsValidSDPriority(r.Priority) {
		return errors.New("invalid priority")
	}
	return nil
}

// UpdateSDTicketRequest payload for triaging, assigning, or updating a ticket
type UpdateSDTicketRequest struct {
	QueueID           *string `json:"queue_id,omitempty"`
	AssigneeID        *string `json:"assignee_id,omitempty"`
	Priority          *string `json:"priority,omitempty"`
	Status            *string `json:"status,omitempty"`
	Subject           *string `json:"subject,omitempty"`
	Description       *string `json:"description,omitempty"`
	LinkedIncidentID  *string `json:"linked_incident_id,omitempty"`
	LinkedKBArticleID *string `json:"linked_kb_article_id,omitempty"`
}

func (r *UpdateSDTicketRequest) Validate() error {
	if r.Subject != nil && strings.TrimSpace(*r.Subject) == "" {
		return errors.New("subject cannot be empty")
	}
	if r.Priority != nil && !IsValidSDPriority(*r.Priority) {
		return errors.New("invalid priority")
	}
	if r.Status != nil && !IsValidSDTicketStatus(*r.Status) {
		return errors.New("invalid status")
	}
	return nil
}

// ResolveSDTicketRequest payload for resolving a support ticket and flagging First Contact Resolution (FCR)
type ResolveSDTicketRequest struct {
	FirstContactResolved bool   `json:"first_contact_resolved"`
	LinkedKBArticleID    string `json:"linked_kb_article_id,omitempty"`
	ResolutionNote       string `json:"resolution_note,omitempty"`
}

// AddSDInteractionRequest payload for logging an omnichannel message or agent note on a ticket
type AddSDInteractionRequest struct {
	WorkspaceID string `json:"workspace_id"`
	TicketID    string `json:"ticket_id"`
	AuthorID    string `json:"author_id"`
	Channel     string `json:"channel"`
	Direction   string `json:"direction"`
	Body        string `json:"body"`
}

func (r *AddSDInteractionRequest) Validate() error {
	if strings.TrimSpace(r.TicketID) == "" {
		return errors.New("ticket_id is required")
	}
	if strings.TrimSpace(r.Body) == "" {
		return errors.New("body is required")
	}
	if r.Channel != "" && !IsValidSDChannel(r.Channel) {
		return errors.New("invalid channel")
	}
	if r.Direction != "" && !IsValidSDDirection(r.Direction) {
		return errors.New("invalid direction")
	}
	return nil
}

// SubmitCSATSurveyRequest payload for submitting post-resolution CSAT/NPS feedback
type SubmitCSATSurveyRequest struct {
	WorkspaceID  string `json:"workspace_id"`
	TicketID     string `json:"ticket_id"`
	RespondentID string `json:"respondent_id"`
	CSATScore    int    `json:"csat_score"`
	NPSScore     int    `json:"nps_score"`
	Comment      string `json:"comment,omitempty"`
}

func (r *SubmitCSATSurveyRequest) Validate() error {
	if strings.TrimSpace(r.TicketID) == "" {
		return errors.New("ticket_id is required")
	}
	if r.CSATScore < 1 || r.CSATScore > 5 {
		return errors.New("csat_score must be between 1 and 5")
	}
	if r.NPSScore < 0 || r.NPSScore > 10 {
		return errors.New("nps_score must be between 0 and 10")
	}
	return nil
}

// Query filters and posture analytics models for Service Desk Management

// SDQueueFilter defines query criteria for browsing support queues
type SDQueueFilter struct {
	WorkspaceID string
	Tier        string
	IsActive    *bool
	Limit       int
	Offset      int
}

// SDTicketFilter defines query criteria for browsing omnichannel tickets
type SDTicketFilter struct {
	WorkspaceID string
	QueueID     string
	RequesterID string
	AssigneeID  string
	Channel     string
	TicketType  string
	Priority    string
	Status      string
	Search      string
	Limit       int
	Offset      int
}

// SDInteractionFilter defines query criteria for ticket interaction threads
type SDInteractionFilter struct {
	WorkspaceID string
	TicketID    string
	Direction   string
	Limit       int
	Offset      int
}

// CSATSurveyFilter defines query criteria for customer satisfaction surveys
type CSATSurveyFilter struct {
	WorkspaceID string
	TicketID    string
	MinScore    int
	Limit       int
	Offset      int
}

// ServiceDeskAnalytics represents executive Service Desk FCR, CSAT, and omnichannel posture
type ServiceDeskAnalytics struct {
	TotalTickets           int            `json:"total_tickets"`
	OpenTickets            int            `json:"open_tickets"`
	ResolvedTickets        int            `json:"resolved_tickets"`
	ClosedTickets          int            `json:"closed_tickets"`
	FCRRate                float64        `json:"fcr_rate"`
	AvgCSATScore           float64        `json:"avg_csat_score"`
	AvgNPSScore            float64        `json:"avg_nps_score"`
	TotalSurveys           int            `json:"total_surveys"`
	ActiveQueues           int            `json:"active_queues"`
	ChannelDistribution    map[string]int `json:"channel_distribution"`
	PriorityDistribution   map[string]int `json:"priority_distribution"`
	ServiceDeskHealthIndex float64        `json:"service_desk_health_index"`
}
