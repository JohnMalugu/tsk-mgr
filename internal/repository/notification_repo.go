package repository

import (
	"database/sql"

	"github.com/JohnMalugu/tsk-mgr-api/internal/model"
)

type NotificationRepository interface {
	Create(n *model.NotificationDB) error
	GetByID(id int) (*model.NotificationDB, error)
	List(filter model.NotificationFilter) ([]*model.NotificationDB, int, error)
	MarkAsRead(id int, userID string) error
	MarkAllAsRead(userID string) error
	Delete(id int, userID string) error
	CountUnread(userID string) (int, error)

	GetPreferences(userID string) (*model.NotificationPreferencesDB, error)
	UpsertPreferences(p *model.NotificationPreferencesDB) error
}

type notificationRepositoryImpl struct {
	db *sql.DB
}

func NewNotificationRepository(db *sql.DB) NotificationRepository {
	return &notificationRepositoryImpl{db: db}
}

func (r *notificationRepositoryImpl) Create(n *model.NotificationDB) error {
	return nil
}

func (r *notificationRepositoryImpl) GetByID(id int) (*model.NotificationDB, error) {
	return nil, nil
}

func (r *notificationRepositoryImpl) List(filter model.NotificationFilter) ([]*model.NotificationDB, int, error) {
	return nil, 0, nil
}

func (r *notificationRepositoryImpl) MarkAsRead(id int, userID string) error {
	return nil
}

func (r *notificationRepositoryImpl) MarkAllAsRead(userID string) error {
	return nil
}

func (r *notificationRepositoryImpl) Delete(id int, userID string) error {
	return nil
}

func (r *notificationRepositoryImpl) CountUnread(userID string) (int, error) {
	return 0, nil
}

func (r *notificationRepositoryImpl) GetPreferences(userID string) (*model.NotificationPreferencesDB, error) {
	return nil, nil
}

func (r *notificationRepositoryImpl) UpsertPreferences(p *model.NotificationPreferencesDB) error {
	return nil
}
