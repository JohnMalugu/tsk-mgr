package service

import (
	"errors"

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
	return nil, nil
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
