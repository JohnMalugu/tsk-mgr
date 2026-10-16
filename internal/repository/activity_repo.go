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
