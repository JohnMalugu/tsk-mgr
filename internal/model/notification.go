package model

import "time"

// NotificationType identifies the domain category for an in-app alert.
type NotificationType string

const (
	NotificationTypeTaskAssigned    NotificationType = "task_assigned"
	NotificationTypeTaskDueSoon     NotificationType = "task_due_soon"
	NotificationTypeTaskCompleted   NotificationType = "task_completed"
	NotificationTypeCommentAdded    NotificationType = "comment_added"
	NotificationTypeWorkspaceInvite NotificationType = "workspace_invite"
	NotificationTypeSystemAlert     NotificationType = "system_alert"
)

// IsValid checks if the notification type is recognized.
func (t NotificationType) IsValid() bool {
	switch t {
	case NotificationTypeTaskAssigned,
		NotificationTypeTaskDueSoon,
		NotificationTypeTaskCompleted,
		NotificationTypeCommentAdded,
		NotificationTypeWorkspaceInvite,
		NotificationTypeSystemAlert:
		return true
	default:
		return false
	}
}

// NotificationDB represents a persistent in-app alert in SQLite.
type NotificationDB struct {
	ID         int              `json:"id"`
	UserID     string           `json:"userId"`
	ActorID    string           `json:"actorId,omitempty"`
	Type       NotificationType `json:"type"`
	Title      string           `json:"title"`
	Message    string           `json:"message"`
	EntityType string           `json:"entityType,omitempty"` // task, comment, workspace
	EntityID   string           `json:"entityId,omitempty"`
	IsRead     bool             `json:"isRead"`
	ReadAt     *time.Time       `json:"readAt,omitempty"`
	CreatedAt  time.Time        `json:"createdAt"`
}

// NotificationPreferencesDB manages per-user notification channel and topic toggles.
type NotificationPreferencesDB struct {
	UserID             string    `json:"userId"`
	EmailEnabled       bool      `json:"emailEnabled"`
	InAppEnabled       bool      `json:"inAppEnabled"`
	NotifyTaskAssigned bool      `json:"notifyTaskAssigned"`
	NotifyTaskDueSoon  bool      `json:"notifyTaskDueSoon"`
	NotifyComment      bool      `json:"notifyComment"`
	NotifyInvite       bool      `json:"notifyInvite"`
	UpdatedAt          time.Time `json:"updatedAt"`
}

// DefaultNotificationPreferences returns user defaults with in-app and email alerts enabled.
func DefaultNotificationPreferences(userID string) *NotificationPreferencesDB {
	return &NotificationPreferencesDB{
		UserID:             userID,
		EmailEnabled:       true,
		InAppEnabled:       true,
		NotifyTaskAssigned: true,
		NotifyTaskDueSoon:  true,
		NotifyComment:      true,
		NotifyInvite:       true,
		UpdatedAt:          time.Now().UTC(),
	}
}

