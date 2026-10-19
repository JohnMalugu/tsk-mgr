package model

import (
	"strings"
	"time"
)

// WebhookEventType represents the domain event that triggers a webhook
type WebhookEventType string

const (
	WebhookEventWildcard             WebhookEventType = "*"
	WebhookEventTaskCreated          WebhookEventType = "task.created"
	WebhookEventTaskUpdated          WebhookEventType = "task.updated"
	WebhookEventTaskCompleted        WebhookEventType = "task.completed"
	WebhookEventTaskDeleted          WebhookEventType = "task.deleted"
	WebhookEventCommentCreated       WebhookEventType = "comment.created"
	WebhookEventNotificationCreated  WebhookEventType = "notification.created"
	WebhookEventWorkspaceMemberJoined WebhookEventType = "workspace.member_joined"
	WebhookEventPing                 WebhookEventType = "webhook.ping"
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
