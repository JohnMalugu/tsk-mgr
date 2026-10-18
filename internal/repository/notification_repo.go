package repository

import (
	"database/sql"
	"fmt"
	"strings"
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
	where := []string{"user_id = ?"}
	args := []interface{}{filter.UserID}

	if filter.Type != "" {
		where = append(where, "type = ?")
		args = append(args, filter.Type)
	}
	if filter.UnreadOnly {
		where = append(where, "is_read = 0")
	}

	whereClause := strings.Join(where, " AND ")

	countQuery := "SELECT COUNT(*) FROM notifications WHERE " + whereClause
	var total int
	if err := r.db.QueryRow(countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	limit := filter.Limit
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	offset := filter.Offset
	if offset < 0 {
		offset = 0
	}

	query := fmt.Sprintf(`
	SELECT id, user_id, COALESCE(actor_id, ''), type, title, message, COALESCE(entity_type, ''), COALESCE(entity_id, ''), is_read, read_at, created_at
	FROM notifications
	WHERE %s
	ORDER BY created_at DESC
	LIMIT ? OFFSET ?`, whereClause)

	listArgs := append(args, limit, offset)
	rows, err := r.db.Query(query, listArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var list []*model.NotificationDB
	for rows.Next() {
		var n model.NotificationDB
		var readAt sql.NullTime
		if err := rows.Scan(
			&n.ID, &n.UserID, &n.ActorID, &n.Type, &n.Title, &n.Message, &n.EntityType, &n.EntityID, &n.IsRead, &readAt, &n.CreatedAt,
		); err != nil {
			return nil, 0, err
		}
		if readAt.Valid {
			n.ReadAt = &readAt.Time
		}
		list = append(list, &n)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

func (r *notificationRepositoryImpl) MarkAsRead(id int, userID string) error {
	now := time.Now().UTC()
	query := `
	UPDATE notifications
	SET is_read = 1, read_at = ?
	WHERE id = ? AND user_id = ? AND is_read = 0`

	res, err := r.db.Exec(query, now, id, userID)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		var exists int
		err := r.db.QueryRow("SELECT COUNT(*) FROM notifications WHERE id = ? AND user_id = ?", id, userID).Scan(&exists)
		if err != nil {
			return err
		}
		if exists == 0 {
			return sql.ErrNoRows
		}
	}
	return nil
}

func (r *notificationRepositoryImpl) MarkAllAsRead(userID string) error {
	now := time.Now().UTC()
	query := `
	UPDATE notifications
	SET is_read = 1, read_at = ?
	WHERE user_id = ? AND is_read = 0`

	_, err := r.db.Exec(query, now, userID)
	return err
}

func (r *notificationRepositoryImpl) Delete(id int, userID string) error {
	query := `DELETE FROM notifications WHERE id = ? AND user_id = ?`
	res, err := r.db.Exec(query, id, userID)
	if err != nil {
		return err
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

func (r *notificationRepositoryImpl) CountUnread(userID string) (int, error) {
	query := `SELECT COUNT(*) FROM notifications WHERE user_id = ? AND is_read = 0`
	var count int
	err := r.db.QueryRow(query, userID).Scan(&count)
	return count, err
}

func (r *notificationRepositoryImpl) GetPreferences(userID string) (*model.NotificationPreferencesDB, error) {
	query := `
	SELECT user_id, email_enabled, in_app_enabled, notify_task_assigned, notify_task_due_soon, notify_comment, notify_invite, updated_at
	FROM notification_preferences
	WHERE user_id = ?`

	var p model.NotificationPreferencesDB
	err := r.db.QueryRow(query, userID).Scan(
		&p.UserID, &p.EmailEnabled, &p.InAppEnabled, &p.NotifyTaskAssigned, &p.NotifyTaskDueSoon, &p.NotifyComment, &p.NotifyInvite, &p.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &p, nil
}

func (r *notificationRepositoryImpl) UpsertPreferences(p *model.NotificationPreferencesDB) error {
	p.UpdatedAt = time.Now().UTC()
	query := `
	INSERT INTO notification_preferences (user_id, email_enabled, in_app_enabled, notify_task_assigned, notify_task_due_soon, notify_comment, notify_invite, updated_at)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	ON CONFLICT(user_id) DO UPDATE SET
		email_enabled = excluded.email_enabled,
		in_app_enabled = excluded.in_app_enabled,
		notify_task_assigned = excluded.notify_task_assigned,
		notify_task_due_soon = excluded.notify_task_due_soon,
		notify_comment = excluded.notify_comment,
		notify_invite = excluded.notify_invite,
		updated_at = excluded.updated_at`

	_, err := r.db.Exec(query, p.UserID, p.EmailEnabled, p.InAppEnabled, p.NotifyTaskAssigned, p.NotifyTaskDueSoon, p.NotifyComment, p.NotifyInvite, p.UpdatedAt)
	return err
}
