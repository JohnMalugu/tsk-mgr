package service

import (
	"errors"
	"strings"

	"github.com/JohnMalugu/tsk-mgr-api/internal/model"
	"github.com/JohnMalugu/tsk-mgr-api/internal/repository"
)

var (
	ErrNotificationNotFound     = errors.New("notification not found")
	ErrUnauthorizedNotification = errors.New("unauthorized notification access")
	ErrInvalidNotificationType  = errors.New("invalid notification type")
	ErrNotificationTitleEmpty   = errors.New("notification title is required")
	ErrNotificationUserIDEmpty  = errors.New("notification target user_id is required")
)

type NotificationService interface {
	Send(req *model.CreateNotificationRequest) (*model.NotificationDB, error)
	ListNotifications(filter model.NotificationFilter) (*model.NotificationSummaryResponse, error)
	MarkAsRead(id int, userID string) error
	MarkAllAsRead(userID string) error
	DeleteNotification(id int, userID string) error
	GetUnreadCount(userID string) (int, error)

	GetPreferences(userID string) (*model.NotificationPreferencesDB, error)
	UpdatePreferences(userID string, req *model.UpdatePreferencesRequest) (*model.NotificationPreferencesDB, error)
}

type notificationServiceImpl struct {
	repo repository.NotificationRepository
}

func NewNotificationService(repo repository.NotificationRepository) NotificationService {
	return &notificationServiceImpl{repo: repo}
}

func (s *notificationServiceImpl) Send(req *model.CreateNotificationRequest) (*model.NotificationDB, error) {
	if req == nil {
		return nil, ErrNotificationTitleEmpty
	}
	userID := strings.TrimSpace(req.UserID)
	if userID == "" {
		return nil, ErrNotificationUserIDEmpty
	}
	title := strings.TrimSpace(req.Title)
	if title == "" {
		return nil, ErrNotificationTitleEmpty
	}
	if !req.Type.IsValid() {
		return nil, ErrInvalidNotificationType
	}

	// Check user preferences
	pref, err := s.repo.GetPreferences(userID)
	if err != nil {
		return nil, err
	}
	if pref != nil {
		if !pref.InAppEnabled {
			return nil, nil // Silently suppressed by user preference
		}
		switch req.Type {
		case model.NotificationTypeTaskAssigned:
			if !pref.NotifyTaskAssigned {
				return nil, nil
			}
		case model.NotificationTypeTaskDueSoon:
			if !pref.NotifyTaskDueSoon {
				return nil, nil
			}
		case model.NotificationTypeCommentAdded:
			if !pref.NotifyComment {
				return nil, nil
			}
		case model.NotificationTypeWorkspaceInvite:
			if !pref.NotifyInvite {
				return nil, nil
			}
		}
	}

	notif := &model.NotificationDB{
		UserID:     userID,
		Type:       req.Type,
		Title:      title,
		Message:    strings.TrimSpace(req.Message),
		EntityType: req.EntityType,
		EntityID:   req.EntityID,
		IsRead:     false,
	}

	if err := s.repo.Create(notif); err != nil {
		return nil, err
	}

	return notif, nil
}

func (s *notificationServiceImpl) ListNotifications(filter model.NotificationFilter) (*model.NotificationSummaryResponse, error) {
	return nil, nil
}

func (s *notificationServiceImpl) MarkAsRead(id int, userID string) error {
	return nil
}

func (s *notificationServiceImpl) MarkAllAsRead(userID string) error {
	return nil
}

func (s *notificationServiceImpl) DeleteNotification(id int, userID string) error {
	return nil
}

func (s *notificationServiceImpl) GetUnreadCount(userID string) (int, error) {
	return 0, nil
}

func (s *notificationServiceImpl) GetPreferences(userID string) (*model.NotificationPreferencesDB, error) {
	return nil, nil
}

func (s *notificationServiceImpl) UpdatePreferences(userID string, req *model.UpdatePreferencesRequest) (*model.NotificationPreferencesDB, error) {
	return nil, nil
}
