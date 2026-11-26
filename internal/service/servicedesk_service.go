package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/JohnMalugu/tsk-mgr-api/internal/model"
	"github.com/JohnMalugu/tsk-mgr-api/internal/repository"
)

var (
	_ = fmt.Sprintf
	_ = strings.TrimSpace
	_ = time.Now
	_ = errors.New
)

var (
	ErrSDQueueNotFound        = repository.ErrSDQueueNotFound
	ErrSDQueueCodeConflict    = repository.ErrSDQueueCodeConflict
	ErrSDTicketNotFound       = repository.ErrSDTicketNotFound
	ErrSDTicketNumberConflict = repository.ErrSDTicketNumberConflict
	ErrSDInteractionNotFound  = repository.ErrSDInteractionNotFound
	ErrCSATSurveyNotFound     = repository.ErrCSATSurveyNotFound
)

// ServiceDeskService defines business logic and orchestration for ITIL Service Desk Management
type ServiceDeskService interface {
	// Queues
	CreateQueue(ctx context.Context, workspaceID string, req model.CreateSDQueueRequest) (*model.ServiceDeskQueue, error)
	GetQueueByID(ctx context.Context, id string) (*model.ServiceDeskQueue, error)
	GetQueueByCode(ctx context.Context, workspaceID, code string) (*model.ServiceDeskQueue, error)
	UpdateQueue(ctx context.Context, id string, req model.UpdateSDQueueRequest) (*model.ServiceDeskQueue, error)
	DeleteQueue(ctx context.Context, id string) error
	ListQueues(ctx context.Context, filter model.SDQueueFilter) ([]model.ServiceDeskQueue, int, error)

	// Tickets
	CreateTicket(ctx context.Context, workspaceID string, req model.CreateSDTicketRequest) (*model.ServiceDeskTicket, error)
	GetTicketByID(ctx context.Context, id string) (*model.ServiceDeskTicket, error)
	GetTicketByNumber(ctx context.Context, workspaceID, number string) (*model.ServiceDeskTicket, error)
	UpdateTicket(ctx context.Context, id string, req model.UpdateSDTicketRequest) (*model.ServiceDeskTicket, error)
	ResolveTicket(ctx context.Context, id string, req model.ResolveSDTicketRequest) (*model.ServiceDeskTicket, error)
	CloseTicket(ctx context.Context, id string) (*model.ServiceDeskTicket, error)
	DeleteTicket(ctx context.Context, id string) error
	ListTickets(ctx context.Context, filter model.SDTicketFilter) ([]model.ServiceDeskTicket, int, error)

	// Interactions
	AddInteraction(ctx context.Context, workspaceID string, req model.AddSDInteractionRequest) (*model.TicketInteraction, error)
	ListInteractions(ctx context.Context, filter model.SDInteractionFilter) ([]model.TicketInteraction, int, error)

	// Surveys
	SubmitSurvey(ctx context.Context, workspaceID string, req model.SubmitCSATSurveyRequest) (*model.CSATSurvey, error)
	ListSurveys(ctx context.Context, filter model.CSATSurveyFilter) ([]model.CSATSurvey, int, error)

	// Analytics
	GetServiceDeskAnalytics(ctx context.Context, workspaceID string) (*model.ServiceDeskAnalytics, error)
}

type serviceDeskService struct {
	repo repository.ServiceDeskRepository
}

// NewServiceDeskService constructs an ITIL Service Desk Management service
func NewServiceDeskService(repo repository.ServiceDeskRepository) ServiceDeskService {
	return &serviceDeskService{repo: repo}
}

// CreateQueue validates and persists a new service desk support queue
func (s *serviceDeskService) CreateQueue(ctx context.Context, workspaceID string, req model.CreateSDQueueRequest) (*model.ServiceDeskQueue, error) {
	if req.WorkspaceID == "" {
		req.WorkspaceID = workspaceID
	}
	if err := req.Validate(); err != nil {
		return nil, err
	}

	queueCode := strings.TrimSpace(req.QueueCode)
	if queueCode == "" {
		queueCode = fmt.Sprintf("Q-%04d", time.Now().UnixNano()%10000)
	}
	tier := req.Tier
	if tier == "" {
		tier = model.SDTier1
	}
	slaMinutes := req.DefaultSLAMinutes
	if slaMinutes <= 0 {
		slaMinutes = 60
	}

	queue := &model.ServiceDeskQueue{
		WorkspaceID:       req.WorkspaceID,
		QueueCode:         queueCode,
		Name:              req.Name,
		Description:       req.Description,
		Tier:              tier,
		DefaultSLAMinutes: slaMinutes,
		IsActive:          true,
	}

	if err := s.repo.CreateQueue(ctx, queue); err != nil {
		return nil, err
	}
	return queue, nil
}

// GetQueueByID retrieves a support queue by ID
func (s *serviceDeskService) GetQueueByID(ctx context.Context, id string) (*model.ServiceDeskQueue, error) {
	if strings.TrimSpace(id) == "" {
		return nil, ErrSDQueueNotFound
	}
	return s.repo.GetQueueByID(ctx, id)
}

// GetQueueByCode retrieves a support queue by code within a workspace
func (s *serviceDeskService) GetQueueByCode(ctx context.Context, workspaceID, code string) (*model.ServiceDeskQueue, error) {
	if strings.TrimSpace(code) == "" {
		return nil, ErrSDQueueNotFound
	}
	return s.repo.GetQueueByCode(ctx, workspaceID, code)
}

// UpdateQueue updates an existing support queue
func (s *serviceDeskService) UpdateQueue(ctx context.Context, id string, req model.UpdateSDQueueRequest) (*model.ServiceDeskQueue, error) {
	if strings.TrimSpace(id) == "" {
		return nil, ErrSDQueueNotFound
	}
	if err := req.Validate(); err != nil {
		return nil, err
	}

	queue, err := s.repo.GetQueueByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.Name != nil {
		queue.Name = *req.Name
	}
	if req.Description != nil {
		queue.Description = *req.Description
	}
	if req.Tier != nil {
		queue.Tier = *req.Tier
	}
	if req.DefaultSLAMinutes != nil {
		queue.DefaultSLAMinutes = *req.DefaultSLAMinutes
	}
	if req.IsActive != nil {
		queue.IsActive = *req.IsActive
	}

	if err := s.repo.UpdateQueue(ctx, queue); err != nil {
		return nil, err
	}
	return queue, nil
}

// DeleteQueue deletes a support queue by ID
func (s *serviceDeskService) DeleteQueue(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return ErrSDQueueNotFound
	}
	return s.repo.DeleteQueue(ctx, id)
}

// ListQueues lists support queues matching filter criteria
func (s *serviceDeskService) ListQueues(ctx context.Context, filter model.SDQueueFilter) ([]model.ServiceDeskQueue, int, error) {
	return s.repo.ListQueues(ctx, filter)
}

// CreateTicket captures an incoming omnichannel support ticket
func (s *serviceDeskService) CreateTicket(ctx context.Context, workspaceID string, req model.CreateSDTicketRequest) (*model.ServiceDeskTicket, error) {
	if req.WorkspaceID == "" {
		req.WorkspaceID = workspaceID
	}
	if err := req.Validate(); err != nil {
		return nil, err
	}

	if req.QueueID != "" {
		if _, err := s.repo.GetQueueByID(ctx, req.QueueID); err != nil {
			return nil, err
		}
	}

	ticketNum := strings.TrimSpace(req.TicketNumber)
	if ticketNum == "" {
		ticketNum = fmt.Sprintf("SD-%05d", time.Now().UnixNano()%100000)
	}
	requesterID := strings.TrimSpace(req.RequesterID)
	if requesterID == "" {
		requesterID = "user_1"
	}
	channel := req.Channel
	if channel == "" {
		channel = model.SDChannelPortal
	}
	ticketType := req.TicketType
	if ticketType == "" {
		ticketType = model.SDTicketTypeIncidentReport
	}
	priority := req.Priority
	if priority == "" {
		priority = model.SDPriorityP3
	}
	status := model.SDTicketStatusNew
	if req.QueueID != "" || req.AssigneeID != "" {
		status = model.SDTicketStatusTriaged
	}

	ticket := &model.ServiceDeskTicket{
		WorkspaceID:       req.WorkspaceID,
		TicketNumber:      ticketNum,
		QueueID:           req.QueueID,
		RequesterID:       requesterID,
		AssigneeID:        req.AssigneeID,
		Channel:           channel,
		TicketType:        ticketType,
		Priority:          priority,
		Status:            status,
		Subject:           req.Subject,
		Description:       req.Description,
		LinkedIncidentID:  req.LinkedIncidentID,
		LinkedKBArticleID: req.LinkedKBArticleID,
	}

	if err := s.repo.CreateTicket(ctx, ticket); err != nil {
		return nil, err
	}
	return ticket, nil
}

// GetTicketByID retrieves a support ticket by ID
func (s *serviceDeskService) GetTicketByID(ctx context.Context, id string) (*model.ServiceDeskTicket, error) {
	if strings.TrimSpace(id) == "" {
		return nil, ErrSDTicketNotFound
	}
	return s.repo.GetTicketByID(ctx, id)
}

// GetTicketByNumber retrieves a support ticket by ticket number within a workspace
func (s *serviceDeskService) GetTicketByNumber(ctx context.Context, workspaceID, number string) (*model.ServiceDeskTicket, error) {
	if strings.TrimSpace(number) == "" {
		return nil, ErrSDTicketNotFound
	}
	return s.repo.GetTicketByNumber(ctx, workspaceID, number)
}

// UpdateTicket updates ticket assignment, queue routing, priority, or status
func (s *serviceDeskService) UpdateTicket(ctx context.Context, id string, req model.UpdateSDTicketRequest) (*model.ServiceDeskTicket, error) {
	if strings.TrimSpace(id) == "" {
		return nil, ErrSDTicketNotFound
	}
	if err := req.Validate(); err != nil {
		return nil, err
	}

	ticket, err := s.repo.GetTicketByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.QueueID != nil {
		if *req.QueueID != "" {
			if _, err := s.repo.GetQueueByID(ctx, *req.QueueID); err != nil {
				return nil, err
			}
		}
		ticket.QueueID = *req.QueueID
	}
	if req.AssigneeID != nil {
		ticket.AssigneeID = *req.AssigneeID
		if ticket.Status == model.SDTicketStatusNew && *req.AssigneeID != "" {
			ticket.Status = model.SDTicketStatusTriaged
		}
	}
	if req.Priority != nil {
		ticket.Priority = *req.Priority
	}
	if req.Status != nil {
		ticket.Status = *req.Status
	}
	if req.Subject != nil {
		ticket.Subject = *req.Subject
	}
	if req.Description != nil {
		ticket.Description = *req.Description
	}
	if req.LinkedIncidentID != nil {
		ticket.LinkedIncidentID = *req.LinkedIncidentID
	}
	if req.LinkedKBArticleID != nil {
		ticket.LinkedKBArticleID = *req.LinkedKBArticleID
	}

	if err := s.repo.UpdateTicket(ctx, ticket); err != nil {
		return nil, err
	}
	return ticket, nil
}

// ResolveTicket marks a support ticket as resolved and captures First Contact Resolution (FCR)
func (s *serviceDeskService) ResolveTicket(ctx context.Context, id string, req model.ResolveSDTicketRequest) (*model.ServiceDeskTicket, error) {
	if strings.TrimSpace(id) == "" {
		return nil, ErrSDTicketNotFound
	}

	ticket, err := s.repo.GetTicketByID(ctx, id)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	ticket.Status = model.SDTicketStatusResolved
	ticket.FirstContactResolved = req.FirstContactResolved
	ticket.ResolvedAt = &now
	if ticket.FirstResponseAt == nil {
		ticket.FirstResponseAt = &now
	}
	if req.LinkedKBArticleID != "" {
		ticket.LinkedKBArticleID = req.LinkedKBArticleID
	}

	if err := s.repo.UpdateTicket(ctx, ticket); err != nil {
		return nil, err
	}

	if strings.TrimSpace(req.ResolutionNote) != "" {
		author := ticket.AssigneeID
		if author == "" {
			author = "system"
		}
		_ = s.repo.CreateInteraction(ctx, &model.TicketInteraction{
			WorkspaceID: ticket.WorkspaceID,
			TicketID:    ticket.ID,
			AuthorID:    author,
			Channel:     ticket.Channel,
			Direction:   model.SDDirectionOutbound,
			Body:        req.ResolutionNote,
			CreatedAt:   now,
		})
	}

	return ticket, nil
}

// CloseTicket transitions a ticket to final closed state
func (s *serviceDeskService) CloseTicket(ctx context.Context, id string) (*model.ServiceDeskTicket, error) {
	if strings.TrimSpace(id) == "" {
		return nil, ErrSDTicketNotFound
	}

	ticket, err := s.repo.GetTicketByID(ctx, id)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	ticket.Status = model.SDTicketStatusClosed
	if ticket.ResolvedAt == nil {
		ticket.ResolvedAt = &now
	}

	if err := s.repo.UpdateTicket(ctx, ticket); err != nil {
		return nil, err
	}
	return ticket, nil
}

// DeleteTicket removes a support ticket by ID
func (s *serviceDeskService) DeleteTicket(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return ErrSDTicketNotFound
	}
	return s.repo.DeleteTicket(ctx, id)
}

// ListTickets queries support tickets matching filter criteria
func (s *serviceDeskService) ListTickets(ctx context.Context, filter model.SDTicketFilter) ([]model.ServiceDeskTicket, int, error) {
	return s.repo.ListTickets(ctx, filter)
}

// AddInteraction appends an omnichannel message or agent note to a ticket and tracks first response SLA
func (s *serviceDeskService) AddInteraction(ctx context.Context, workspaceID string, req model.AddSDInteractionRequest) (*model.TicketInteraction, error) {
	if req.WorkspaceID == "" {
		req.WorkspaceID = workspaceID
	}
	if err := req.Validate(); err != nil {
		return nil, err
	}

	ticket, err := s.repo.GetTicketByID(ctx, req.TicketID)
	if err != nil {
		return nil, err
	}

	author := strings.TrimSpace(req.AuthorID)
	if author == "" {
		author = "user_1"
	}
	channel := req.Channel
	if channel == "" {
		channel = ticket.Channel
	}
	direction := req.Direction
	if direction == "" {
		direction = model.SDDirectionInbound
	}

	now := time.Now().UTC()
	interaction := &model.TicketInteraction{
		WorkspaceID: req.WorkspaceID,
		TicketID:    req.TicketID,
		AuthorID:    author,
		Channel:     channel,
		Direction:   direction,
		Body:        req.Body,
		CreatedAt:   now,
	}

	if err := s.repo.CreateInteraction(ctx, interaction); err != nil {
		return nil, err
	}

	// Stamp first response timestamp and advance status on first outbound agent reply
	if direction == model.SDDirectionOutbound {
		updated := false
		if ticket.FirstResponseAt == nil {
			ticket.FirstResponseAt = &now
			updated = true
		}
		if ticket.Status == model.SDTicketStatusNew || ticket.Status == model.SDTicketStatusTriaged {
			ticket.Status = model.SDTicketStatusInProgress
			updated = true
		}
		if updated {
			_ = s.repo.UpdateTicket(ctx, ticket)
		}
	}

	return interaction, nil
}

// ListInteractions queries interaction threads based on filter criteria
func (s *serviceDeskService) ListInteractions(ctx context.Context, filter model.SDInteractionFilter) ([]model.TicketInteraction, int, error) {
	return s.repo.ListInteractions(ctx, filter)
}

// Surveys stubs
func (s *serviceDeskService) SubmitSurvey(ctx context.Context, workspaceID string, req model.SubmitCSATSurveyRequest) (*model.CSATSurvey, error) {
	return nil, errors.New("not implemented")
}
func (s *serviceDeskService) ListSurveys(ctx context.Context, filter model.CSATSurveyFilter) ([]model.CSATSurvey, int, error) {
	return nil, 0, errors.New("not implemented")
}

// Analytics stub
func (s *serviceDeskService) GetServiceDeskAnalytics(ctx context.Context, workspaceID string) (*model.ServiceDeskAnalytics, error) {
	return nil, errors.New("not implemented")
}
