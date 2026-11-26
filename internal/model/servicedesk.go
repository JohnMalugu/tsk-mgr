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
