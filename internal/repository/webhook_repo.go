package repository

import (
	"database/sql"
	"encoding/json"
	"fmt"
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
	return nil, 0, nil
}

func (r *sqlWebhookRepository) UpdateSubscription(sub *model.WebhookSubscriptionDB) error {
	return nil
}

func (r *sqlWebhookRepository) DeleteSubscription(id int64) error {
	return nil
}

func (r *sqlWebhookRepository) CreateDelivery(delivery *model.WebhookDeliveryDB) (int64, error) {
	return 0, nil
}

func (r *sqlWebhookRepository) GetDeliveryByID(id int64) (*model.WebhookDeliveryDB, error) {
	return nil, nil
}

func (r *sqlWebhookRepository) ListDeliveries(filter model.DeliveryFilter) ([]model.WebhookDeliveryDB, int, error) {
	return nil, 0, nil
}

func (r *sqlWebhookRepository) GetWebhookStats(subscriptionID int64) (*model.WebhookStats, error) {
	return nil, nil
}

