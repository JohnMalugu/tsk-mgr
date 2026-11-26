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

// Tickets stubs
func (r *sqlServiceDeskRepository) CreateTicket(ctx context.Context, t *model.ServiceDeskTicket) error {
	return errors.New("not implemented")
}
func (r *sqlServiceDeskRepository) GetTicketByID(ctx context.Context, id string) (*model.ServiceDeskTicket, error) {
	return nil, errors.New("not implemented")
}
func (r *sqlServiceDeskRepository) GetTicketByNumber(ctx context.Context, workspaceID, number string) (*model.ServiceDeskTicket, error) {
	return nil, errors.New("not implemented")
}
func (r *sqlServiceDeskRepository) UpdateTicket(ctx context.Context, t *model.ServiceDeskTicket) error {
	return errors.New("not implemented")
}
func (r *sqlServiceDeskRepository) DeleteTicket(ctx context.Context, id string) error {
	return errors.New("not implemented")
}
func (r *sqlServiceDeskRepository) ListTickets(ctx context.Context, filter model.SDTicketFilter) ([]model.ServiceDeskTicket, int, error) {
	return nil, 0, errors.New("not implemented")
}

// Interactions stubs
func (r *sqlServiceDeskRepository) CreateInteraction(ctx context.Context, i *model.TicketInteraction) error {
	return errors.New("not implemented")
}
func (r *sqlServiceDeskRepository) ListInteractions(ctx context.Context, filter model.SDInteractionFilter) ([]model.TicketInteraction, int, error) {
	return nil, 0, errors.New("not implemented")
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
