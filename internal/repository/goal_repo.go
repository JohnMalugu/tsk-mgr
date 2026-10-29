package repository

import (
	"database/sql"
	"errors"
	"time"

	"github.com/JohnMalugu/tsk-mgr-api/internal/model"
)

var (
	ErrGoalNotFound        = errors.New("goal not found")
	ErrKeyResultNotFound   = errors.New("key result not found")
	ErrTaskAlreadyLinked   = errors.New("task is already linked to this key result")
	ErrKeyResultTaskLinkNF = errors.New("task link not found for key result")
)

// GoalRepository defines data access operations for Goals, Key Results, task links, check-ins, and metrics
type GoalRepository interface {
	CreateGoal(goal *model.Goal) (int64, error)
	GetGoalByID(id int64) (*model.Goal, error)
	ListGoals(filter model.GoalFilter) ([]*model.GoalSummary, error)
	UpdateGoal(goal *model.Goal) error
	DeleteGoal(id int64) error

	CreateKeyResult(kr *model.KeyResult) (int64, error)
	GetKeyResultByID(id int64) (*model.KeyResult, error)
	ListKeyResults(goalID int64) ([]model.KeyResult, error)
	UpdateKeyResult(kr *model.KeyResult) error
	DeleteKeyResult(id int64) error

	LinkTaskToKeyResult(keyResultID int64, taskID int) error
	UnlinkTask(keyResultID int64, taskID int) error
	ListLinkedTasks(keyResultID int64) ([]model.LinkedTaskInfo, error)

	CreateGoalCheckIn(checkIn *model.GoalCheckIn) (int64, error)
	ListGoalCheckIns(goalID int64) ([]model.GoalCheckIn, error)

	GetGoalTree(goalID int64) (*model.GoalTree, error)
	GetGoalMetrics(workspaceID int64) (*model.GoalMetrics, error)
}

type goalRepositoryImpl struct {
	db *sql.DB
}

// NewGoalRepository constructs a new SQLite-backed GoalRepository instance
func NewGoalRepository(db *sql.DB) GoalRepository {
	return &goalRepositoryImpl{db: db}
}

func scanGoal(scanner interface{ Scan(dest ...interface{}) error }) (*model.Goal, error) {
	var (
		g            model.Goal
		parentGoalID sql.NullInt64
		startDate    sql.NullTime
		endDate      sql.NullTime
	)

	err := scanner.Scan(
		&g.ID, &g.WorkspaceID, &g.OwnerUserID, &parentGoalID,
		&g.Title, &g.Description, &g.Timeframe,
		&startDate, &endDate,
		&g.Status, &g.ProgressPercentage,
		&g.CreatedAt, &g.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrGoalNotFound
		}
		return nil, err
	}

	if parentGoalID.Valid {
		pID := parentGoalID.Int64
		g.ParentGoalID = &pID
	}
	if startDate.Valid {
		g.StartDate = &startDate.Time
	}
	if endDate.Valid {
		g.EndDate = &endDate.Time
	}

	return &g, nil
}

func (r *goalRepositoryImpl) CreateGoal(goal *model.Goal) (int64, error) {
	now := time.Now()
	if goal.CreatedAt.IsZero() {
		goal.CreatedAt = now
	}
	if goal.UpdatedAt.IsZero() {
		goal.UpdatedAt = now
	}
	if goal.Status == "" {
		goal.Status = model.GoalStatusDraft
	}

	query := `
	INSERT INTO goals (
		workspace_id, owner_user_id, parent_goal_id,
		title, description, timeframe,
		start_date, end_date,
		status, progress_percentage,
		created_at, updated_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	res, err := r.db.Exec(query,
		goal.WorkspaceID, goal.OwnerUserID, goal.ParentGoalID,
		goal.Title, goal.Description, goal.Timeframe,
		goal.StartDate, goal.EndDate,
		goal.Status, goal.ProgressPercentage,
		goal.CreatedAt, goal.UpdatedAt,
	)
	if err != nil {
		return 0, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	goal.ID = id
	return id, nil
}

func (r *goalRepositoryImpl) GetGoalByID(id int64) (*model.Goal, error) {
	query := `
	SELECT id, workspace_id, owner_user_id, parent_goal_id,
		title, description, timeframe,
		start_date, end_date,
		status, progress_percentage,
		created_at, updated_at
	FROM goals
	WHERE id = ?`

	return scanGoal(r.db.QueryRow(query, id))
}

func (r *goalRepositoryImpl) ListGoals(filter model.GoalFilter) ([]*model.GoalSummary, error) {
	query := `
	SELECT g.id, g.title, g.timeframe, g.status, g.progress_percentage,
		COUNT(kr.id) as key_result_count, g.owner_user_id
	FROM goals g
	LEFT JOIN key_results kr ON g.id = kr.goal_id
	WHERE 1=1`

	var args []interface{}

	if filter.WorkspaceID != nil && *filter.WorkspaceID > 0 {
		query += " AND g.workspace_id = ?"
		args = append(args, *filter.WorkspaceID)
	}
	if filter.OwnerUserID != "" {
		query += " AND g.owner_user_id = ?"
		args = append(args, filter.OwnerUserID)
	}
	if filter.Status != "" {
		query += " AND g.status = ?"
		args = append(args, filter.Status)
	}
	if filter.Timeframe != "" {
		query += " AND g.timeframe = ?"
		args = append(args, filter.Timeframe)
	}
	if filter.ParentGoalID != nil {
		query += " AND g.parent_goal_id = ?"
		args = append(args, *filter.ParentGoalID)
	}

	query += " GROUP BY g.id ORDER BY g.created_at DESC"

	if filter.Limit > 0 {
		query += " LIMIT ?"
		args = append(args, filter.Limit)
		if filter.Offset > 0 {
			query += " OFFSET ?"
			args = append(args, filter.Offset)
		}
	}

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var summaries []*model.GoalSummary
	for rows.Next() {
		var s model.GoalSummary
		if err := rows.Scan(
			&s.ID, &s.Title, &s.Timeframe, &s.Status,
			&s.ProgressPercentage, &s.KeyResultCount, &s.OwnerUserID,
		); err != nil {
			return nil, err
		}
		summaries = append(summaries, &s)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	if summaries == nil {
		summaries = []*model.GoalSummary{}
	}

	return summaries, nil
}

func (r *goalRepositoryImpl) UpdateGoal(goal *model.Goal) error {
	goal.UpdatedAt = time.Now()

	query := `
	UPDATE goals
	SET owner_user_id = ?, parent_goal_id = ?,
		title = ?, description = ?, timeframe = ?,
		start_date = ?, end_date = ?,
		status = ?, progress_percentage = ?,
		updated_at = ?
	WHERE id = ?`

	res, err := r.db.Exec(query,
		goal.OwnerUserID, goal.ParentGoalID,
		goal.Title, goal.Description, goal.Timeframe,
		goal.StartDate, goal.EndDate,
		goal.Status, goal.ProgressPercentage,
		goal.UpdatedAt, goal.ID,
	)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrGoalNotFound
	}

	return nil
}

func (r *goalRepositoryImpl) DeleteGoal(id int64) error {
	query := `DELETE FROM goals WHERE id = ?`
	res, err := r.db.Exec(query, id)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrGoalNotFound
	}

	return nil
}

func (r *goalRepositoryImpl) CreateKeyResult(kr *model.KeyResult) (int64, error) {
	return 0, nil
}

func (r *goalRepositoryImpl) GetKeyResultByID(id int64) (*model.KeyResult, error) {
	return nil, nil
}

func (r *goalRepositoryImpl) ListKeyResults(goalID int64) ([]model.KeyResult, error) {
	return nil, nil
}

func (r *goalRepositoryImpl) UpdateKeyResult(kr *model.KeyResult) error {
	return nil
}

func (r *goalRepositoryImpl) DeleteKeyResult(id int64) error {
	return nil
}

func (r *goalRepositoryImpl) LinkTaskToKeyResult(keyResultID int64, taskID int) error {
	return nil
}

func (r *goalRepositoryImpl) UnlinkTask(keyResultID int64, taskID int) error {
	return nil
}

func (r *goalRepositoryImpl) ListLinkedTasks(keyResultID int64) ([]model.LinkedTaskInfo, error) {
	return nil, nil
}

func (r *goalRepositoryImpl) CreateGoalCheckIn(checkIn *model.GoalCheckIn) (int64, error) {
	return 0, nil
}

func (r *goalRepositoryImpl) ListGoalCheckIns(goalID int64) ([]model.GoalCheckIn, error) {
	return nil, nil
}

func (r *goalRepositoryImpl) GetGoalTree(goalID int64) (*model.GoalTree, error) {
	return nil, nil
}

func (r *goalRepositoryImpl) GetGoalMetrics(workspaceID int64) (*model.GoalMetrics, error) {
	return nil, nil
}
