package repository

import (
	"database/sql"
	"time"

	"github.com/JohnMalugu/tsk-mgr-api/internal/model"
)

type TimeEntryRepository interface {
	Create(entry *model.TimeEntryDB) error
	GetByID(id int) (*model.TimeEntryDB, error)
	GetActiveTimerByUser(userID string) (*model.TimeEntryDB, error)
	GetByTaskID(taskID int) ([]*model.TimeEntryDB, error)
	Update(entry *model.TimeEntryDB) error
	Delete(id int) error
	GetUserTimeEntriesInRange(userID string, from, to *time.Time) ([]*model.TimeEntryDB, error)
}

type timeEntryRepositoryImpl struct {
	db *sql.DB
}

func NewTimeEntryRepository(db *sql.DB) TimeEntryRepository {
	return &timeEntryRepositoryImpl{db: db}
}
