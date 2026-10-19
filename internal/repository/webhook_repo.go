package repository

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/JohnMalugu/tsk-mgr-api/internal/model"
)

// WebhookRepository defines data persistence operations for webhooks and delivery logs
type WebhookRepository interface {
	CreateSubscription(sub *model.WebhookSubscriptionDB) (int64, error)
	GetSubscriptionByID(id int64) (*model.WebhookSubscriptionDB, error)
	ListSubscriptions(filter model.WebhookFilter) ([]model.WebhookSubscriptionDB, int, error)
	UpdateSubscription(sub *model.WebhookSubscriptionDB) error
	DeleteSubscription(id int64) error
	CreateDelivery(delivery *model.WebhookDeliveryDB) (int64, error)
	GetDeliveryByID(id int64) (*model.WebhookDeliveryDB, error)
	ListDeliveries(filter model.DeliveryFilter) ([]model.WebhookDeliveryDB, int, error)
	GetWebhookStats(subscriptionID int64) (*model.WebhookStats, error)
}

type sqlWebhookRepository struct {
	db *sql.DB
}

// NewWebhookRepository returns an implementation of WebhookRepository using SQLite
func NewWebhookRepository(db *sql.DB) WebhookRepository {
	return &sqlWebhookRepository{db: db}
}

func (r *sqlWebhookRepository) CreateSubscription(sub *model.WebhookSubscriptionDB) (int64, error) {
	eventsJSON, err := json.Marshal(sub.EventTypes)
	if err != nil {
		return 0, fmt.Errorf("failed to marshal event types: %w", err)
	}

	query := `
	INSERT INTO webhook_subscriptions (workspace_id, user_id, target_url, secret_token, event_types, is_active, description, created_at, updated_at)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`

	now := time.Now()
	if sub.CreatedAt.IsZero() {
		sub.CreatedAt = now
	}
	if sub.UpdatedAt.IsZero() {
		sub.UpdatedAt = now
	}

	res, err := r.db.Exec(query, sub.WorkspaceID, sub.UserID, sub.TargetURL, sub.SecretToken, string(eventsJSON), sub.IsActive, sub.Description, sub.CreatedAt, sub.UpdatedAt)
	if err != nil {
		return 0, fmt.Errorf("failed to insert webhook subscription: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	sub.ID = id
	return id, nil
}

func (r *sqlWebhookRepository) GetSubscriptionByID(id int64) (*model.WebhookSubscriptionDB, error) {
	query := `
	SELECT id, workspace_id, user_id, target_url, secret_token, event_types, is_active, description, created_at, updated_at
	FROM webhook_subscriptions
	WHERE id = ?`

	row := r.db.QueryRow(query, id)
	var sub model.WebhookSubscriptionDB
	var eventsRaw string
	err := row.Scan(&sub.ID, &sub.WorkspaceID, &sub.UserID, &sub.TargetURL, &sub.SecretToken, &eventsRaw, &sub.IsActive, &sub.Description, &sub.CreatedAt, &sub.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to query webhook subscription: %w", err)
	}

	if eventsRaw != "" {
		_ = json.Unmarshal([]byte(eventsRaw), &sub.EventTypes)
	}
	return &sub, nil
}

func (r *sqlWebhookRepository) ListSubscriptions(filter model.WebhookFilter) ([]model.WebhookSubscriptionDB, int, error) {
	var conditions []string
	var args []interface{}

	if filter.WorkspaceID > 0 {
		conditions = append(conditions, "workspace_id = ?")
		args = append(args, filter.WorkspaceID)
	}
	if filter.UserID != "" {
		conditions = append(conditions, "user_id = ?")
		args = append(args, filter.UserID)
	}
	if filter.IsActive != nil {
		conditions = append(conditions, "is_active = ?")
		args = append(args, *filter.IsActive)
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM webhook_subscriptions %s", whereClause)
	var total int
	err := r.db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count subscriptions: %w", err)
	}

	limit := filter.Limit
	if limit <= 0 {
		limit = 20
	}
	offset := filter.Offset
	if offset < 0 {
		offset = 0
	}

	query := fmt.Sprintf(`
	SELECT id, workspace_id, user_id, target_url, secret_token, event_types, is_active, description, created_at, updated_at
	FROM webhook_subscriptions
	%s
	ORDER BY id DESC
	LIMIT ? OFFSET ?`, whereClause)

	queryArgs := append(args, limit, offset)
	rows, err := r.db.Query(query, queryArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list subscriptions: %w", err)
	}
	defer rows.Close()

	var subs []model.WebhookSubscriptionDB
	for rows.Next() {
		var sub model.WebhookSubscriptionDB
		var eventsRaw string
		if err := rows.Scan(&sub.ID, &sub.WorkspaceID, &sub.UserID, &sub.TargetURL, &sub.SecretToken, &eventsRaw, &sub.IsActive, &sub.Description, &sub.CreatedAt, &sub.UpdatedAt); err != nil {
			return nil, 0, err
		}
		if eventsRaw != "" {
			_ = json.Unmarshal([]byte(eventsRaw), &sub.EventTypes)
		}
		if filter.EventType != "" {
			matched := false
			for _, et := range sub.EventTypes {
				if model.MatchesWebhookEventType(et, filter.EventType) {
					matched = true
					break
				}
			}
			if !matched {
				continue
			}
		}
		subs = append(subs, sub)
	}

	return subs, total, nil
}

func (r *sqlWebhookRepository) UpdateSubscription(sub *model.WebhookSubscriptionDB) error {
	eventsJSON, err := json.Marshal(sub.EventTypes)
	if err != nil {
		return fmt.Errorf("failed to marshal event types: %w", err)
	}

	sub.UpdatedAt = time.Now()
	query := `
	UPDATE webhook_subscriptions
	SET target_url = ?, secret_token = ?, event_types = ?, is_active = ?, description = ?, updated_at = ?
	WHERE id = ?`

	res, err := r.db.Exec(query, sub.TargetURL, sub.SecretToken, string(eventsJSON), sub.IsActive, sub.Description, sub.UpdatedAt, sub.ID)
	if err != nil {
		return fmt.Errorf("failed to update webhook subscription: %w", err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *sqlWebhookRepository) DeleteSubscription(id int64) error {
	query := `DELETE FROM webhook_subscriptions WHERE id = ?`
	res, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete webhook subscription: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *sqlWebhookRepository) CreateDelivery(delivery *model.WebhookDeliveryDB) (int64, error) {
	query := `
	INSERT INTO webhook_deliveries (
		subscription_id, event_type, payload_json, status_code, response_headers, response_body, duration_ms, attempt_number, status, error_message, created_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	now := time.Now()
	if delivery.CreatedAt.IsZero() {
		delivery.CreatedAt = now
	}
	if delivery.AttemptNumber <= 0 {
		delivery.AttemptNumber = 1
	}

	res, err := r.db.Exec(
		query,
		delivery.SubscriptionID,
		delivery.EventType,
		delivery.PayloadJSON,
		delivery.StatusCode,
		delivery.ResponseHeaders,
		delivery.ResponseBody,
		delivery.DurationMs,
		delivery.AttemptNumber,
		delivery.Status,
		delivery.ErrorMessage,
		delivery.CreatedAt,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to insert webhook delivery: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	delivery.ID = id
	return id, nil
}

func (r *sqlWebhookRepository) GetDeliveryByID(id int64) (*model.WebhookDeliveryDB, error) {
	query := `
	SELECT id, subscription_id, event_type, payload_json, status_code, response_headers, response_body, duration_ms, attempt_number, status, error_message, created_at
	FROM webhook_deliveries
	WHERE id = ?`

	row := r.db.QueryRow(query, id)
	var d model.WebhookDeliveryDB
	err := row.Scan(
		&d.ID,
		&d.SubscriptionID,
		&d.EventType,
		&d.PayloadJSON,
		&d.StatusCode,
		&d.ResponseHeaders,
		&d.ResponseBody,
		&d.DurationMs,
		&d.AttemptNumber,
		&d.Status,
		&d.ErrorMessage,
		&d.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to query webhook delivery: %w", err)
	}
	return &d, nil
}

func (r *sqlWebhookRepository) ListDeliveries(filter model.DeliveryFilter) ([]model.WebhookDeliveryDB, int, error) {
	var conditions []string
	var args []interface{}

	if filter.SubscriptionID > 0 {
		conditions = append(conditions, "subscription_id = ?")
		args = append(args, filter.SubscriptionID)
	}
	if filter.Status != "" {
		conditions = append(conditions, "status = ?")
		args = append(args, filter.Status)
	}
	if filter.EventType != "" {
		conditions = append(conditions, "event_type = ?")
		args = append(args, filter.EventType)
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM webhook_deliveries %s", whereClause)
	var total int
	err := r.db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count webhook deliveries: %w", err)
	}

	limit := filter.Limit
	if limit <= 0 {
		limit = 20
	}
	offset := filter.Offset
	if offset < 0 {
		offset = 0
	}

	query := fmt.Sprintf(`
	SELECT id, subscription_id, event_type, payload_json, status_code, response_headers, response_body, duration_ms, attempt_number, status, error_message, created_at
	FROM webhook_deliveries
	%s
	ORDER BY id DESC
	LIMIT ? OFFSET ?`, whereClause)

	queryArgs := append(args, limit, offset)
	rows, err := r.db.Query(query, queryArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list webhook deliveries: %w", err)
	}
	defer rows.Close()

	var list []model.WebhookDeliveryDB
	for rows.Next() {
		var d model.WebhookDeliveryDB
		if err := rows.Scan(
			&d.ID,
			&d.SubscriptionID,
			&d.EventType,
			&d.PayloadJSON,
			&d.StatusCode,
			&d.ResponseHeaders,
			&d.ResponseBody,
			&d.DurationMs,
			&d.AttemptNumber,
			&d.Status,
			&d.ErrorMessage,
			&d.CreatedAt,
		); err != nil {
			return nil, 0, err
		}
		list = append(list, d)
	}

	return list, total, nil
}

func (r *sqlWebhookRepository) GetWebhookStats(subscriptionID int64) (*model.WebhookStats, error) {
	query := `
	SELECT 
		COUNT(*) as total,
		COALESCE(SUM(CASE WHEN status = 'success' THEN 1 ELSE 0 END), 0) as successes,
		COALESCE(SUM(CASE WHEN status = 'failed' THEN 1 ELSE 0 END), 0) as failures,
		COALESCE(AVG(duration_ms), 0) as avg_duration,
		MAX(created_at) as last_delivery
	FROM webhook_deliveries
	WHERE subscription_id = ?`

	row := r.db.QueryRow(query, subscriptionID)
	var stats model.WebhookStats
	stats.SubscriptionID = subscriptionID

	var lastDelivery sql.NullTime
	err := row.Scan(
		&stats.TotalDeliveries,
		&stats.SuccessfulDeliveries,
		&stats.FailedDeliveries,
		&stats.AverageDurationMs,
		&lastDelivery,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to aggregate webhook stats: %w", err)
	}

	if lastDelivery.Valid {
		stats.LastDeliveryAt = &lastDelivery.Time
	}

	if stats.TotalDeliveries > 0 {
		stats.SuccessRate = float64(stats.SuccessfulDeliveries) / float64(stats.TotalDeliveries) * 100.0
	}

	return &stats, nil
}

