package model

import (
	"strings"
	"time"
)

// WebhookEventType represents the domain event that triggers a webhook
type WebhookEventType string

const (
	WebhookEventWildcard              WebhookEventType = "*"
	WebhookEventTaskCreated           WebhookEventType = "task.created"
	WebhookEventTaskUpdated           WebhookEventType = "task.updated"
	WebhookEventTaskCompleted         WebhookEventType = "task.completed"
	WebhookEventTaskDeleted           WebhookEventType = "task.deleted"
	WebhookEventCommentCreated        WebhookEventType = "comment.created"
	WebhookEventNotificationCreated   WebhookEventType = "notification.created"
	WebhookEventWorkspaceMemberJoined WebhookEventType = "workspace.member_joined"
	WebhookEventPing                  WebhookEventType = "webhook.ping"
)

// SupportedWebhookEventTypes list all valid event types
var SupportedWebhookEventTypes = []WebhookEventType{
	WebhookEventWildcard,
	WebhookEventTaskCreated,
	WebhookEventTaskUpdated,
	WebhookEventTaskCompleted,
	WebhookEventTaskDeleted,
	WebhookEventCommentCreated,
	WebhookEventNotificationCreated,
	WebhookEventWorkspaceMemberJoined,
	WebhookEventPing,
}

// IsValidWebhookEventType validates if an event type string is supported
func IsValidWebhookEventType(eventType string) bool {
	trimmed := strings.TrimSpace(eventType)
	for _, supported := range SupportedWebhookEventTypes {
		if string(supported) == trimmed {
			return true
		}
	}
	return false
}

// MatchesWebhookEventType checks if a subscription event pattern matches an actual event
func MatchesWebhookEventType(subscriptionPattern, actualEvent string) bool {
	if subscriptionPattern == string(WebhookEventWildcard) {
		return true
	}
	return strings.EqualFold(strings.TrimSpace(subscriptionPattern), strings.TrimSpace(actualEvent))
}

// WebhookSubscriptionDB represents a webhook subscriber registered in a workspace
type WebhookSubscriptionDB struct {
	ID          int64     `json:"id"`
	WorkspaceID int64     `json:"workspace_id"`
	UserID      string    `json:"user_id"`
	TargetURL   string    `json:"target_url"`
	SecretToken string    `json:"secret_token,omitempty"`
	EventTypes  []string  `json:"event_types"`
	IsActive    bool      `json:"is_active"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Delivery status constants
const (
	WebhookDeliveryStatusSuccess = "success"
	WebhookDeliveryStatusFailed  = "failed"
	WebhookDeliveryStatusPending = "pending"
)

// WebhookDeliveryDB represents an HTTP delivery attempt and outcome for a webhook
type WebhookDeliveryDB struct {
	ID              int64     `json:"id"`
	SubscriptionID  int64     `json:"subscription_id"`
	EventType       string    `json:"event_type"`
	PayloadJSON     string    `json:"payload_json"`
	StatusCode      int       `json:"status_code"`
	ResponseHeaders string    `json:"response_headers"`
	ResponseBody    string    `json:"response_body"`
	DurationMs      int64     `json:"duration_ms"`
	AttemptNumber   int       `json:"attempt_number"`
	Status          string    `json:"status"`
	ErrorMessage    string    `json:"error_message,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
}

// WebhookEvent represents the standard JSON envelope delivered to external endpoints
type WebhookEvent struct {
	EventID      string      `json:"event_id"`
	EventType    string      `json:"event_type"`
	Timestamp    time.Time   `json:"timestamp"`
	WorkspaceID  int64       `json:"workspace_id"`
	ActorUserID  string      `json:"actor_user_id,omitempty"`
	ResourceType string      `json:"resource_type"`
	ResourceID   string      `json:"resource_id"`
	Data         interface{} `json:"data"`
}

// WebhookFilter defines criteria for querying registered webhooks
type WebhookFilter struct {
	WorkspaceID int64  `json:"workspace_id"`
	UserID      string `json:"user_id"`
	IsActive    *bool  `json:"is_active"`
	EventType   string `json:"event_type"`
	Offset      int    `json:"offset"`
	Limit       int    `json:"limit"`
}

// DeliveryFilter defines criteria for querying delivery execution logs
type DeliveryFilter struct {
	SubscriptionID int64  `json:"subscription_id"`
	Status         string `json:"status"`
	EventType      string `json:"event_type"`
	Offset         int    `json:"offset"`
	Limit          int    `json:"limit"`
}

// WebhookSummary represents a paginated list of webhook subscriptions
type WebhookSummary struct {
	Subscriptions []WebhookSubscriptionDB `json:"subscriptions"`
	Total         int                     `json:"total"`
	Offset        int                     `json:"offset"`
	Limit         int                     `json:"limit"`
}

// DeliverySummary represents a paginated list of webhook delivery records
type DeliverySummary struct {
	Deliveries []WebhookDeliveryDB `json:"deliveries"`
	Total      int                 `json:"total"`
	Offset     int                 `json:"offset"`
	Limit      int                 `json:"limit"`
}

// WebhookStats represents aggregated health and delivery metrics for a webhook subscription
type WebhookStats struct {
	SubscriptionID       int64      `json:"subscription_id"`
	TotalDeliveries      int        `json:"total_deliveries"`
	SuccessfulDeliveries int        `json:"successful_deliveries"`
	FailedDeliveries     int        `json:"failed_deliveries"`
	SuccessRate          float64    `json:"success_rate"`
	AverageDurationMs    float64    `json:"average_duration_ms"`
	LastDeliveryAt       *time.Time `json:"last_delivery_at,omitempty"`
}

// CreateWebhookRequest represents the payload to register a new webhook
type CreateWebhookRequest struct {
	WorkspaceID int64    `json:"workspace_id"`
	TargetURL   string   `json:"target_url"`
	SecretToken string   `json:"secret_token,omitempty"`
	EventTypes  []string `json:"event_types"`
	Description string   `json:"description,omitempty"`
	IsActive    *bool    `json:"is_active,omitempty"`
}

// UpdateWebhookRequest represents updates to an existing webhook subscription
type UpdateWebhookRequest struct {
	TargetURL   *string  `json:"target_url,omitempty"`
	SecretToken *string  `json:"secret_token,omitempty"`
	EventTypes  []string `json:"event_types,omitempty"`
	Description *string  `json:"description,omitempty"`
	IsActive    *bool    `json:"is_active,omitempty"`
}
