package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/JohnMalugu/tsk-mgr-api/internal/model"
)

var (
	_ = uuid.New
	_ = fmt.Sprintf
	_ = strings.Join
	_ = time.Now
	_ = errors.New
)

var (
	ErrSDQueueNotFound        = errors.New("service desk queue not found")
	ErrSDQueueCodeConflict    = errors.New("service desk queue with this code already exists in workspace")
	ErrSDTicketNotFound       = errors.New("service desk ticket not found")
	ErrSDTicketNumberConflict = errors.New("service desk ticket with this number already exists in workspace")
	ErrSDInteractionNotFound  = errors.New("ticket interaction not found")
	ErrCSATSurveyNotFound     = errors.New("CSAT survey not found")
)

// ServiceDeskRepository defines data access methods for ITIL Service Desk Management
type ServiceDeskRepository interface {
	// Queues
	CreateQueue(ctx context.Context, q *model.ServiceDeskQueue) error
	GetQueueByID(ctx context.Context, id string) (*model.ServiceDeskQueue, error)
	GetQueueByCode(ctx context.Context, workspaceID, code string) (*model.ServiceDeskQueue, error)
	UpdateQueue(ctx context.Context, q *model.ServiceDeskQueue) error
	DeleteQueue(ctx context.Context, id string) error
	ListQueues(ctx context.Context, filter model.SDQueueFilter) ([]model.ServiceDeskQueue, int, error)

	// Tickets
	CreateTicket(ctx context.Context, t *model.ServiceDeskTicket) error
	GetTicketByID(ctx context.Context, id string) (*model.ServiceDeskTicket, error)
	GetTicketByNumber(ctx context.Context, workspaceID, number string) (*model.ServiceDeskTicket, error)
	UpdateTicket(ctx context.Context, t *model.ServiceDeskTicket) error
	DeleteTicket(ctx context.Context, id string) error
	ListTickets(ctx context.Context, filter model.SDTicketFilter) ([]model.ServiceDeskTicket, int, error)

	// Interactions
	CreateInteraction(ctx context.Context, i *model.TicketInteraction) error
	ListInteractions(ctx context.Context, filter model.SDInteractionFilter) ([]model.TicketInteraction, int, error)

	// Surveys
	CreateSurvey(ctx context.Context, s *model.CSATSurvey) error
	ListSurveys(ctx context.Context, filter model.CSATSurveyFilter) ([]model.CSATSurvey, int, error)

	// Posture Analytics
	GetServiceDeskAnalytics(ctx context.Context, workspaceID string) (*model.ServiceDeskAnalytics, error)
}

type sqlServiceDeskRepository struct {
	db *sql.DB
}

// NewServiceDeskRepository initializes a SQLite service desk repository
func NewServiceDeskRepository(db *sql.DB) ServiceDeskRepository {
	return &sqlServiceDeskRepository{db: db}
}

// CreateQueue inserts a new service desk support queue
func (r *sqlServiceDeskRepository) CreateQueue(ctx context.Context, q *model.ServiceDeskQueue) error {
	if q.ID == "" {
		q.ID = uuid.New().String()
	}
	now := time.Now().UTC()
	if q.CreatedAt.IsZero() {
		q.CreatedAt = now
	}
	q.UpdatedAt = now
	if q.Tier == "" {
		q.Tier = model.SDTier1
	}
	if q.DefaultSLAMinutes <= 0 {
		q.DefaultSLAMinutes = 60
	}

	var count int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM service_desk_queues WHERE workspace_id = ? AND queue_code = ?", q.WorkspaceID, q.QueueCode).Scan(&count)
	if err == nil && count > 0 {
		return ErrSDQueueCodeConflict
	}

	query := `
		INSERT INTO service_desk_queues (
			id, workspace_id, queue_code, name, description,
			tier, default_sla_minutes, is_active, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err = r.db.ExecContext(ctx, query,
		q.ID, q.WorkspaceID, q.QueueCode, q.Name, q.Description,
		q.Tier, q.DefaultSLAMinutes, q.IsActive, q.CreatedAt, q.UpdatedAt,
	)
	return err
}

// GetQueueByID retrieves a service desk queue by ID
func (r *sqlServiceDeskRepository) GetQueueByID(ctx context.Context, id string) (*model.ServiceDeskQueue, error) {
	query := `
		SELECT id, workspace_id, queue_code, name, description,
		       tier, default_sla_minutes, is_active, created_at, updated_at
		FROM service_desk_queues
		WHERE id = ?`

	var q model.ServiceDeskQueue
	var desc sql.NullString
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&q.ID, &q.WorkspaceID, &q.QueueCode, &q.Name, &desc,
		&q.Tier, &q.DefaultSLAMinutes, &q.IsActive, &q.CreatedAt, &q.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrSDQueueNotFound
		}
		return nil, err
	}
	if desc.Valid {
		q.Description = desc.String
	}
	return &q, nil
}

// GetQueueByCode retrieves a service desk queue by queue_code within a workspace
func (r *sqlServiceDeskRepository) GetQueueByCode(ctx context.Context, workspaceID, code string) (*model.ServiceDeskQueue, error) {
	query := `
		SELECT id, workspace_id, queue_code, name, description,
		       tier, default_sla_minutes, is_active, created_at, updated_at
		FROM service_desk_queues
		WHERE workspace_id = ? AND queue_code = ?`

	var q model.ServiceDeskQueue
	var desc sql.NullString
	err := r.db.QueryRowContext(ctx, query, workspaceID, code).Scan(
		&q.ID, &q.WorkspaceID, &q.QueueCode, &q.Name, &desc,
		&q.Tier, &q.DefaultSLAMinutes, &q.IsActive, &q.CreatedAt, &q.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrSDQueueNotFound
		}
		return nil, err
	}
	if desc.Valid {
		q.Description = desc.String
	}
	return &q, nil
}

// UpdateQueue updates an existing service desk queue
func (r *sqlServiceDeskRepository) UpdateQueue(ctx context.Context, q *model.ServiceDeskQueue) error {
	q.UpdatedAt = time.Now().UTC()

	query := `
		UPDATE service_desk_queues
		SET name = ?, description = ?, tier = ?, default_sla_minutes = ?,
		    is_active = ?, updated_at = ?
		WHERE id = ?`

	res, err := r.db.ExecContext(ctx, query,
		q.Name, q.Description, q.Tier, q.DefaultSLAMinutes,
		q.IsActive, q.UpdatedAt, q.ID,
	)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrSDQueueNotFound
	}
	return nil
}

// DeleteQueue removes a service desk queue
func (r *sqlServiceDeskRepository) DeleteQueue(ctx context.Context, id string) error {
	res, err := r.db.ExecContext(ctx, "DELETE FROM service_desk_queues WHERE id = ?", id)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrSDQueueNotFound
	}
	return nil
}

// ListQueues queries service desk queues based on filter criteria
func (r *sqlServiceDeskRepository) ListQueues(ctx context.Context, filter model.SDQueueFilter) ([]model.ServiceDeskQueue, int, error) {
	var conditions []string
	var args []interface{}

	if filter.WorkspaceID != "" {
		conditions = append(conditions, "workspace_id = ?")
		args = append(args, filter.WorkspaceID)
	}
	if filter.Tier != "" {
		conditions = append(conditions, "tier = ?")
		args = append(args, filter.Tier)
	}
	if filter.IsActive != nil {
		conditions = append(conditions, "is_active = ?")
		args = append(args, *filter.IsActive)
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	var total int
	err := r.db.QueryRowContext(ctx, fmt.Sprintf("SELECT COUNT(*) FROM service_desk_queues %s", whereClause), args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	limit := filter.Limit
	if limit <= 0 {
		limit = 50
	}
	offset := filter.Offset
	if offset < 0 {
		offset = 0
	}

	selectQuery := fmt.Sprintf(`
		SELECT id, workspace_id, queue_code, name, description,
		       tier, default_sla_minutes, is_active, created_at, updated_at
		FROM service_desk_queues
		%s
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?`, whereClause)

	rows, err := r.db.QueryContext(ctx, selectQuery, append(args, limit, offset)...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var queues []model.ServiceDeskQueue
	for rows.Next() {
		var q model.ServiceDeskQueue
		var desc sql.NullString
		if err := rows.Scan(
			&q.ID, &q.WorkspaceID, &q.QueueCode, &q.Name, &desc,
			&q.Tier, &q.DefaultSLAMinutes, &q.IsActive, &q.CreatedAt, &q.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		if desc.Valid {
			q.Description = desc.String
		}
		queues = append(queues, q)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return queues, total, nil
}

// CreateTicket inserts a new omnichannel service desk ticket
func (r *sqlServiceDeskRepository) CreateTicket(ctx context.Context, t *model.ServiceDeskTicket) error {
	if t.ID == "" {
		t.ID = uuid.New().String()
	}
	now := time.Now().UTC()
	if t.CreatedAt.IsZero() {
		t.CreatedAt = now
	}
	t.UpdatedAt = now
	if t.Channel == "" {
		t.Channel = model.SDChannelPortal
	}
	if t.TicketType == "" {
		t.TicketType = model.SDTicketTypeIncidentReport
	}
	if t.Priority == "" {
		t.Priority = model.SDPriorityP3
	}
	if t.Status == "" {
		t.Status = model.SDTicketStatusNew
	}

	var count int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM service_desk_tickets WHERE workspace_id = ? AND ticket_number = ?", t.WorkspaceID, t.TicketNumber).Scan(&count)
	if err == nil && count > 0 {
		return ErrSDTicketNumberConflict
	}

	query := `
		INSERT INTO service_desk_tickets (
			id, workspace_id, ticket_number, queue_id, requester_id,
			assignee_id, channel, ticket_type, priority, status,
			subject, description, first_contact_resolved, first_response_at,
			resolved_at, linked_incident_id, linked_kb_article_id, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err = r.db.ExecContext(ctx, query,
		t.ID, t.WorkspaceID, t.TicketNumber, t.QueueID, t.RequesterID,
		t.AssigneeID, t.Channel, t.TicketType, t.Priority, t.Status,
		t.Subject, t.Description, t.FirstContactResolved, t.FirstResponseAt,
		t.ResolvedAt, t.LinkedIncidentID, t.LinkedKBArticleID, t.CreatedAt, t.UpdatedAt,
	)
	return err
}

func scanServiceDeskTicket(scanner interface {
	Scan(dest ...interface{}) error
}) (*model.ServiceDeskTicket, error) {
	var t model.ServiceDeskTicket
	var queueID, assigneeID, desc, linkedInc, linkedKB sql.NullString
	var firstRespAt, resolvedAt sql.NullTime

	err := scanner.Scan(
		&t.ID, &t.WorkspaceID, &t.TicketNumber, &queueID, &t.RequesterID,
		&assigneeID, &t.Channel, &t.TicketType, &t.Priority, &t.Status,
		&t.Subject, &desc, &t.FirstContactResolved, &firstRespAt,
		&resolvedAt, &linkedInc, &linkedKB, &t.CreatedAt, &t.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	if queueID.Valid {
		t.QueueID = queueID.String
	}
	if assigneeID.Valid {
		t.AssigneeID = assigneeID.String
	}
	if desc.Valid {
		t.Description = desc.String
	}
	if firstRespAt.Valid {
		t.FirstResponseAt = &firstRespAt.Time
	}
	if resolvedAt.Valid {
		t.ResolvedAt = &resolvedAt.Time
	}
	if linkedInc.Valid {
		t.LinkedIncidentID = linkedInc.String
	}
	if linkedKB.Valid {
		t.LinkedKBArticleID = linkedKB.String
	}
	return &t, nil
}

// GetTicketByID retrieves a service desk ticket by ID
func (r *sqlServiceDeskRepository) GetTicketByID(ctx context.Context, id string) (*model.ServiceDeskTicket, error) {
	query := `
		SELECT id, workspace_id, ticket_number, queue_id, requester_id,
		       assignee_id, channel, ticket_type, priority, status,
		       subject, description, first_contact_resolved, first_response_at,
		       resolved_at, linked_incident_id, linked_kb_article_id, created_at, updated_at
		FROM service_desk_tickets
		WHERE id = ?`

	t, err := scanServiceDeskTicket(r.db.QueryRowContext(ctx, query, id))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrSDTicketNotFound
		}
		return nil, err
	}
	return t, nil
}

// GetTicketByNumber retrieves a service desk ticket by ticket_number within a workspace
func (r *sqlServiceDeskRepository) GetTicketByNumber(ctx context.Context, workspaceID, number string) (*model.ServiceDeskTicket, error) {
	query := `
		SELECT id, workspace_id, ticket_number, queue_id, requester_id,
		       assignee_id, channel, ticket_type, priority, status,
		       subject, description, first_contact_resolved, first_response_at,
		       resolved_at, linked_incident_id, linked_kb_article_id, created_at, updated_at
		FROM service_desk_tickets
		WHERE workspace_id = ? AND ticket_number = ?`

	t, err := scanServiceDeskTicket(r.db.QueryRowContext(ctx, query, workspaceID, number))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrSDTicketNotFound
		}
		return nil, err
	}
	return t, nil
}

// UpdateTicket updates an existing service desk ticket
func (r *sqlServiceDeskRepository) UpdateTicket(ctx context.Context, t *model.ServiceDeskTicket) error {
	t.UpdatedAt = time.Now().UTC()

	query := `
		UPDATE service_desk_tickets
		SET queue_id = ?, assignee_id = ?, priority = ?, status = ?,
		    subject = ?, description = ?, first_contact_resolved = ?,
		    first_response_at = ?, resolved_at = ?, linked_incident_id = ?,
		    linked_kb_article_id = ?, updated_at = ?
		WHERE id = ?`

	res, err := r.db.ExecContext(ctx, query,
		t.QueueID, t.AssigneeID, t.Priority, t.Status,
		t.Subject, t.Description, t.FirstContactResolved,
		t.FirstResponseAt, t.ResolvedAt, t.LinkedIncidentID,
		t.LinkedKBArticleID, t.UpdatedAt, t.ID,
	)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrSDTicketNotFound
	}
	return nil
}

// DeleteTicket removes a service desk ticket
func (r *sqlServiceDeskRepository) DeleteTicket(ctx context.Context, id string) error {
	res, err := r.db.ExecContext(ctx, "DELETE FROM service_desk_tickets WHERE id = ?", id)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrSDTicketNotFound
	}
	return nil
}

// ListTickets queries service desk tickets based on filter criteria
func (r *sqlServiceDeskRepository) ListTickets(ctx context.Context, filter model.SDTicketFilter) ([]model.ServiceDeskTicket, int, error) {
	var conditions []string
	var args []interface{}

	if filter.WorkspaceID != "" {
		conditions = append(conditions, "workspace_id = ?")
		args = append(args, filter.WorkspaceID)
	}
	if filter.QueueID != "" {
		conditions = append(conditions, "queue_id = ?")
		args = append(args, filter.QueueID)
	}
	if filter.RequesterID != "" {
		conditions = append(conditions, "requester_id = ?")
		args = append(args, filter.RequesterID)
	}
	if filter.AssigneeID != "" {
		conditions = append(conditions, "assignee_id = ?")
		args = append(args, filter.AssigneeID)
	}
	if filter.Channel != "" {
		conditions = append(conditions, "channel = ?")
		args = append(args, filter.Channel)
	}
	if filter.TicketType != "" {
		conditions = append(conditions, "ticket_type = ?")
		args = append(args, filter.TicketType)
	}
	if filter.Priority != "" {
		conditions = append(conditions, "priority = ?")
		args = append(args, filter.Priority)
	}
	if filter.Status != "" {
		conditions = append(conditions, "status = ?")
		args = append(args, filter.Status)
	}
	if filter.Search != "" {
		conditions = append(conditions, "(subject LIKE ? OR ticket_number LIKE ? OR description LIKE ?)")
		s := "%" + filter.Search + "%"
		args = append(args, s, s, s)
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	var total int
	err := r.db.QueryRowContext(ctx, fmt.Sprintf("SELECT COUNT(*) FROM service_desk_tickets %s", whereClause), args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	limit := filter.Limit
	if limit <= 0 {
		limit = 50
	}
	offset := filter.Offset
	if offset < 0 {
		offset = 0
	}

	selectQuery := fmt.Sprintf(`
		SELECT id, workspace_id, ticket_number, queue_id, requester_id,
		       assignee_id, channel, ticket_type, priority, status,
		       subject, description, first_contact_resolved, first_response_at,
		       resolved_at, linked_incident_id, linked_kb_article_id, created_at, updated_at
		FROM service_desk_tickets
		%s
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?`, whereClause)

	rows, err := r.db.QueryContext(ctx, selectQuery, append(args, limit, offset)...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var tickets []model.ServiceDeskTicket
	for rows.Next() {
		t, err := scanServiceDeskTicket(rows)
		if err != nil {
			return nil, 0, err
		}
		tickets = append(tickets, *t)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return tickets, total, nil
}

// CreateInteraction inserts a new omnichannel interaction message or note on a ticket
func (r *sqlServiceDeskRepository) CreateInteraction(ctx context.Context, i *model.TicketInteraction) error {
	if i.ID == "" {
		i.ID = uuid.New().String()
	}
	if i.CreatedAt.IsZero() {
		i.CreatedAt = time.Now().UTC()
	}
	if i.Channel == "" {
		i.Channel = model.SDChannelPortal
	}
	if i.Direction == "" {
		i.Direction = model.SDDirectionInbound
	}

	query := `
		INSERT INTO service_desk_interactions (
			id, workspace_id, ticket_id, author_id, channel, direction, body, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := r.db.ExecContext(ctx, query,
		i.ID, i.WorkspaceID, i.TicketID, i.AuthorID, i.Channel, i.Direction, i.Body, i.CreatedAt,
	)
	return err
}

// ListInteractions queries interaction messages for a ticket or workspace
func (r *sqlServiceDeskRepository) ListInteractions(ctx context.Context, filter model.SDInteractionFilter) ([]model.TicketInteraction, int, error) {
	var conditions []string
	var args []interface{}

	if filter.WorkspaceID != "" {
		conditions = append(conditions, "workspace_id = ?")
		args = append(args, filter.WorkspaceID)
	}
	if filter.TicketID != "" {
		conditions = append(conditions, "ticket_id = ?")
		args = append(args, filter.TicketID)
	}
	if filter.Direction != "" {
		conditions = append(conditions, "direction = ?")
		args = append(args, filter.Direction)
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	var total int
	err := r.db.QueryRowContext(ctx, fmt.Sprintf("SELECT COUNT(*) FROM service_desk_interactions %s", whereClause), args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	limit := filter.Limit
	if limit <= 0 {
		limit = 50
	}
	offset := filter.Offset
	if offset < 0 {
		offset = 0
	}

	selectQuery := fmt.Sprintf(`
		SELECT id, workspace_id, ticket_id, author_id, channel, direction, body, created_at
		FROM service_desk_interactions
		%s
		ORDER BY created_at ASC
		LIMIT ? OFFSET ?`, whereClause)

	rows, err := r.db.QueryContext(ctx, selectQuery, append(args, limit, offset)...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var interactions []model.TicketInteraction
	for rows.Next() {
		var i model.TicketInteraction
		if err := rows.Scan(
			&i.ID, &i.WorkspaceID, &i.TicketID, &i.AuthorID,
			&i.Channel, &i.Direction, &i.Body, &i.CreatedAt,
		); err != nil {
			return nil, 0, err
		}
		interactions = append(interactions, i)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return interactions, total, nil
}

// Surveys stubs
func (r *sqlServiceDeskRepository) CreateSurvey(ctx context.Context, s *model.CSATSurvey) error {
	return errors.New("not implemented")
}
func (r *sqlServiceDeskRepository) ListSurveys(ctx context.Context, filter model.CSATSurveyFilter) ([]model.CSATSurvey, int, error) {
	return nil, 0, errors.New("not implemented")
}

// Analytics stub
func (r *sqlServiceDeskRepository) GetServiceDeskAnalytics(ctx context.Context, workspaceID string) (*model.ServiceDeskAnalytics, error) {
	return nil, errors.New("not implemented")
}
