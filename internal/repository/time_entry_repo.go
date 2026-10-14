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

func (r *timeEntryRepositoryImpl) Create(entry *model.TimeEntryDB) error {
	query := `
	INSERT INTO time_entries (task_id, user_id, started_at, ended_at, duration_seconds, note, created_at, updated_at)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?)`

	result, err := r.db.Exec(query,
		entry.TaskID, entry.UserID, entry.StartedAt, entry.EndedAt,
		entry.DurationSeconds, entry.Note, entry.CreatedAt, entry.UpdatedAt)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err == nil {
		entry.ID = int(id)
	}

	return err
}

func (r *timeEntryRepositoryImpl) GetByID(id int) (*model.TimeEntryDB, error) {
	query := `
	SELECT id, task_id, user_id, started_at, ended_at, duration_seconds, note, created_at, updated_at
	FROM time_entries WHERE id = ?`

	row := r.db.QueryRow(query, id)
	var e model.TimeEntryDB

	err := row.Scan(&e.ID, &e.TaskID, &e.UserID, &e.StartedAt, &e.EndedAt,
		&e.DurationSeconds, &e.Note, &e.CreatedAt, &e.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &e, nil
}

func (r *timeEntryRepositoryImpl) GetActiveTimerByUser(userID string) (*model.TimeEntryDB, error) {
	query := `
	SELECT id, task_id, user_id, started_at, ended_at, duration_seconds, note, created_at, updated_at
	FROM time_entries WHERE user_id = ? AND ended_at IS NULL ORDER BY started_at DESC LIMIT 1`

	row := r.db.QueryRow(query, userID)
	var e model.TimeEntryDB

	err := row.Scan(&e.ID, &e.TaskID, &e.UserID, &e.StartedAt, &e.EndedAt,
		&e.DurationSeconds, &e.Note, &e.CreatedAt, &e.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &e, nil
}

func (r *timeEntryRepositoryImpl) GetByTaskID(taskID int) ([]*model.TimeEntryDB, error) {
	query := `
	SELECT id, task_id, user_id, started_at, ended_at, duration_seconds, note, created_at, updated_at
	FROM time_entries WHERE task_id = ? ORDER BY started_at DESC`

	rows, err := r.db.Query(query, taskID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*model.TimeEntryDB
	for rows.Next() {
		var e model.TimeEntryDB
		err := rows.Scan(&e.ID, &e.TaskID, &e.UserID, &e.StartedAt, &e.EndedAt,
			&e.DurationSeconds, &e.Note, &e.CreatedAt, &e.UpdatedAt)
		if err != nil {
			return nil, err
		}
		list = append(list, &e)
	}

	return list, rows.Err()
}

func (r *timeEntryRepositoryImpl) Update(entry *model.TimeEntryDB) error {
	query := `
	UPDATE time_entries
	SET ended_at = ?, duration_seconds = ?, note = ?, updated_at = ?
	WHERE id = ?`

	_, err := r.db.Exec(query,
		entry.EndedAt, entry.DurationSeconds, entry.Note, entry.UpdatedAt, entry.ID)
	return err
}

func (r *timeEntryRepositoryImpl) Delete(id int) error {
	query := `DELETE FROM time_entries WHERE id = ?`
	_, err := r.db.Exec(query, id)
	return err
}

func (r *timeEntryRepositoryImpl) GetUserTimeEntriesInRange(userID string, from, to *time.Time) ([]*model.TimeEntryDB, error) {
	query := `
	SELECT id, task_id, user_id, started_at, ended_at, duration_seconds, note, created_at, updated_at
	FROM time_entries WHERE user_id = ?`

	args := []interface{}{userID}
	if from != nil {
		query += " AND started_at >= ?"
		args = append(args, *from)
	}
	if to != nil {
		query += " AND started_at <= ?"
		args = append(args, *to)
	}
	query += " ORDER BY started_at ASC"

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*model.TimeEntryDB
	for rows.Next() {
		var e model.TimeEntryDB
		err := rows.Scan(&e.ID, &e.TaskID, &e.UserID, &e.StartedAt, &e.EndedAt,
			&e.DurationSeconds, &e.Note, &e.CreatedAt, &e.UpdatedAt)
		if err != nil {
			return nil, err
		}
		list = append(list, &e)
	}

	return list, rows.Err()
}







