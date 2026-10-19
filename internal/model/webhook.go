package model

import (
	"strings"
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
