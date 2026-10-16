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
