package repository

import (
	"database/sql"
	"time"

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
	now := time.Now().UTC()
	if n.CreatedAt.IsZero() {
		n.CreatedAt = now
	}

	query := `
	INSERT INTO notifications (user_id, actor_id, type, title, message, entity_type, entity_id, is_read, read_at, created_at)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	res, err := r.db.Exec(query, n.UserID, n.ActorID, n.Type, n.Title, n.Message, n.EntityType, n.EntityID, n.IsRead, n.ReadAt, n.CreatedAt)
	if err != nil {
		return err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return err
	}
	n.ID = int(id)
	return nil
}

func (r *notificationRepositoryImpl) GetByID(id int) (*model.NotificationDB, error) {
	query := `
	SELECT id, user_id, COALESCE(actor_id, ''), type, title, message, COALESCE(entity_type, ''), COALESCE(entity_id, ''), is_read, read_at, created_at
	FROM notifications
	WHERE id = ?`

	var n model.NotificationDB
	var readAt sql.NullTime
	err := r.db.QueryRow(query, id).Scan(
		&n.ID, &n.UserID, &n.ActorID, &n.Type, &n.Title, &n.Message, &n.EntityType, &n.EntityID, &n.IsRead, &readAt, &n.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	if readAt.Valid {
		n.ReadAt = &readAt.Time
	}
	return &n, nil
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
