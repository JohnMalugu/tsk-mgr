package repository

import (
	"database/sql"
	"time"

	"github.com/JohnMalugu/tsk-mgr-api/internal/model"
)

type ActivityRepository interface {
	Create(activity *model.ActivityDB) error
	GetByID(id int) (*model.ActivityDB, error)
	ListByUserID(userID string, filter *model.ActivityFilter) ([]*model.ActivityDB, error)
	ListByTaskID(taskID int, offset, limit int) ([]*model.ActivityDB, error)
	CountByUserID(userID string, filter *model.ActivityFilter) (int, error)
	GetActivityStats(userID string, from, to *time.Time) ([]model.ActionCount, error)
	PurgeOlderThan(cutoff time.Time) (int64, error)
}

type activityRepositoryImpl struct {
	db *sql.DB
}

func NewActivityRepository(db *sql.DB) ActivityRepository {
	return &activityRepositoryImpl{db: db}
}

func (r *activityRepositoryImpl) Create(activity *model.ActivityDB) error {
	if activity.CreatedAt.IsZero() {
		activity.CreatedAt = time.Now().UTC()
	}
	query := `
	INSERT INTO activities (user_id, task_id, action, summary, details_json, created_at)
	VALUES (?, ?, ?, ?, ?, ?)`

	res, err := r.db.Exec(query,
		activity.UserID,
		activity.TaskID,
		activity.Action,
		activity.Summary,
		activity.DetailsJSON,
		activity.CreatedAt,
	)
	if err != nil {
		return err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return err
	}
	activity.ID = int(id)
	return nil
}

func (r *activityRepositoryImpl) GetByID(id int) (*model.ActivityDB, error) {
	query := `
	SELECT id, user_id, task_id, action, summary, details_json, created_at
	FROM activities
	WHERE id = ?`

	var a model.ActivityDB
	var taskID sql.NullInt64
	var details sql.NullString

	err := r.db.QueryRow(query, id).Scan(
		&a.ID,
		&a.UserID,
		&taskID,
		&a.Action,
		&a.Summary,
		&details,
		&a.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	if taskID.Valid {
		tid := int(taskID.Int64)
		a.TaskID = &tid
	}
	if details.Valid {
		a.DetailsJSON = details.String
	}
	return &a, nil
}

func (r *activityRepositoryImpl) ListByUserID(userID string, filter *model.ActivityFilter) ([]*model.ActivityDB, error) {
	query := `SELECT id, user_id, task_id, action, summary, details_json, created_at FROM activities WHERE user_id = ?`
	args := []interface{}{userID}

	if filter != nil {
		if filter.Action != "" {
			query += ` AND action = ?`
			args = append(args, filter.Action)
		}
		if filter.TaskID != nil {
			query += ` AND task_id = ?`
			args = append(args, *filter.TaskID)
		}
		if filter.From != nil {
			query += ` AND created_at >= ?`
			args = append(args, *filter.From)
		}
		if filter.To != nil {
			query += ` AND created_at <= ?`
			args = append(args, *filter.To)
		}
	}

	query += ` ORDER BY created_at DESC, id DESC`

	limit := 50
	offset := 0
	if filter != nil {
		if filter.Limit > 0 {
			limit = filter.Limit
		}
		if filter.Offset > 0 {
			offset = filter.Offset
		}
	}
	query += ` LIMIT ? OFFSET ?`
	args = append(args, limit, offset)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*model.ActivityDB
	for rows.Next() {
		var a model.ActivityDB
		var taskID sql.NullInt64
		var details sql.NullString
		if err := rows.Scan(&a.ID, &a.UserID, &taskID, &a.Action, &a.Summary, &details, &a.CreatedAt); err != nil {
			return nil, err
		}
		if taskID.Valid {
			tid := int(taskID.Int64)
			a.TaskID = &tid
		}
		if details.Valid {
			a.DetailsJSON = details.String
		}
		list = append(list, &a)
	}
	return list, rows.Err()
}

func (r *activityRepositoryImpl) ListByTaskID(taskID int, offset, limit int) ([]*model.ActivityDB, error) {
	if limit <= 0 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	query := `
	SELECT id, user_id, task_id, action, summary, details_json, created_at
	FROM activities
	WHERE task_id = ?
	ORDER BY created_at DESC, id DESC
	LIMIT ? OFFSET ?`

	rows, err := r.db.Query(query, taskID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*model.ActivityDB
	for rows.Next() {
		var a model.ActivityDB
		var tID sql.NullInt64
		var details sql.NullString
		if err := rows.Scan(&a.ID, &a.UserID, &tID, &a.Action, &a.Summary, &details, &a.CreatedAt); err != nil {
			return nil, err
		}
		if tID.Valid {
			tid := int(tID.Int64)
			a.TaskID = &tid
		}
		if details.Valid {
			a.DetailsJSON = details.String
		}
		list = append(list, &a)
	}
	return list, rows.Err()
}

func (r *activityRepositoryImpl) CountByUserID(userID string, filter *model.ActivityFilter) (int, error) {
	query := `SELECT COUNT(1) FROM activities WHERE user_id = ?`
	args := []interface{}{userID}

	if filter != nil {
		if filter.Action != "" {
			query += ` AND action = ?`
			args = append(args, filter.Action)
		}
		if filter.TaskID != nil {
			query += ` AND task_id = ?`
			args = append(args, *filter.TaskID)
		}
		if filter.From != nil {
			query += ` AND created_at >= ?`
			args = append(args, *filter.From)
		}
		if filter.To != nil {
			query += ` AND created_at <= ?`
			args = append(args, *filter.To)
		}
	}

	var total int
	err := r.db.QueryRow(query, args...).Scan(&total)
	return total, err
}

func (r *activityRepositoryImpl) GetActivityStats(userID string, from, to *time.Time) ([]model.ActionCount, error) {
	query := `
	SELECT action, COUNT(1) as cnt
	FROM activities
	WHERE user_id = ?`
	args := []interface{}{userID}

	if from != nil {
		query += ` AND created_at >= ?`
		args = append(args, *from)
	}
	if to != nil {
		query += ` AND created_at <= ?`
		args = append(args, *to)
	}

	query += ` GROUP BY action ORDER BY cnt DESC, action ASC`

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []model.ActionCount
	for rows.Next() {
		var ac model.ActionCount
		if err := rows.Scan(&ac.Action, &ac.Count); err != nil {
			return nil, err
		}
		stats = append(stats, ac)
	}
	return stats, rows.Err()
}
