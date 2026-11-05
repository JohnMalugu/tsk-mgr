package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/JohnMalugu/tsk-mgr-api/internal/model"
)

var (
	ErrRiskNotFound = errors.New("risk not found")
)

// RiskRepository manages database operations for risks, mitigation task links, and reviews
type RiskRepository interface {
	CreateRisk(risk *model.Risk) (int64, error)
	GetRiskByID(id int64) (*model.Risk, error)
	GetRiskByCode(workspaceID int64, code string) (*model.Risk, error)
	GetRiskByNumber(workspaceID int64, number int) (*model.Risk, error)
	ListRisks(filter model.RiskFilter) ([]*model.Risk, error)
	UpdateRisk(risk *model.Risk) error
	DeleteRisk(id int64) error
	GetNextRiskNumber(workspaceID int64) (int, error)

	LinkTask(link *model.RiskTaskLink) (int64, error)
	UnlinkTask(riskID int64, taskID int) error
	ListRiskTasks(riskID int64) ([]model.RiskTaskLink, error)

	CreateReview(review *model.RiskReview) (int64, error)
	ListReviews(riskID int64) ([]model.RiskReview, error)

	GetRiskMetrics(workspaceID int64) (*model.RiskMetrics, error)
}

type riskRepositoryImpl struct {
	db *sql.DB
}

// NewRiskRepository creates a new instance of RiskRepository
func NewRiskRepository(db *sql.DB) RiskRepository {
	return &riskRepositoryImpl{db: db}
}

func (r *riskRepositoryImpl) GetNextRiskNumber(workspaceID int64) (int, error) {
	query := `SELECT COALESCE(MAX(risk_number), 0) + 1 FROM risks WHERE workspace_id = ?`
	var nextNum int
	err := r.db.QueryRow(query, workspaceID).Scan(&nextNum)
	if err != nil {
		return 1, err
	}
	return nextNum, nil
}

func (r *riskRepositoryImpl) CreateRisk(risk *model.Risk) (int64, error) {
	if risk.RiskNumber <= 0 {
		nextNum, err := r.GetNextRiskNumber(risk.WorkspaceID)
		if err != nil {
			return 0, err
		}
		risk.RiskNumber = nextNum
	}
	if risk.RiskCode == "" {
		risk.RiskCode = fmt.Sprintf("RSK-%d", risk.RiskNumber)
	}

	risk.Score, risk.Severity = model.CalculateRiskSeverity(risk.Probability, risk.Impact)

	now := time.Now().UTC()
	if risk.CreatedAt.IsZero() {
		risk.CreatedAt = now
	}
	if risk.UpdatedAt.IsZero() {
		risk.UpdatedAt = now
	}

	query := `
	INSERT INTO risks (
		workspace_id, project_id, risk_number, risk_code, title, description,
		category, probability, impact, score, severity, status, strategy,
		mitigation_plan, contingency_plan, owner_id, target_resolution_date, closed_at,
		created_at, updated_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	res, err := r.db.Exec(query,
		risk.WorkspaceID, risk.ProjectID, risk.RiskNumber, risk.RiskCode, risk.Title, risk.Description,
		risk.Category, risk.Probability, risk.Impact, risk.Score, risk.Severity, risk.Status, risk.Strategy,
		risk.MitigationPlan, risk.ContingencyPlan, risk.OwnerID, risk.TargetResolutionDate, risk.ClosedAt,
		risk.CreatedAt, risk.UpdatedAt,
	)
	if err != nil {
		return 0, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	risk.ID = id
	return id, nil
}

func scanRisk(scanner interface {
	Scan(dest ...interface{}) error
}) (*model.Risk, error) {
	var rk model.Risk
	var projectID sql.NullInt32
	var ownerID sql.NullString
	var targetResolutionDate, closedAt sql.NullTime
	var ownerUsername, projectName sql.NullString
	var taskLinksCount, openTaskLinksCount sql.NullInt32

	err := scanner.Scan(
		&rk.ID, &rk.WorkspaceID, &projectID, &rk.RiskNumber, &rk.RiskCode, &rk.Title, &rk.Description,
		&rk.Category, &rk.Probability, &rk.Impact, &rk.Score, &rk.Severity, &rk.Status, &rk.Strategy,
		&rk.MitigationPlan, &rk.ContingencyPlan, &ownerID, &targetResolutionDate, &closedAt,
		&rk.CreatedAt, &rk.UpdatedAt,
		&ownerUsername, &projectName,
		&taskLinksCount, &openTaskLinksCount,
	)
	if err != nil {
		return nil, err
	}

	if projectID.Valid {
		pid := int(projectID.Int32)
		rk.ProjectID = &pid
	}
	if ownerID.Valid {
		oid := ownerID.String
		rk.OwnerID = &oid
	}
	if targetResolutionDate.Valid {
		t := targetResolutionDate.Time
		rk.TargetResolutionDate = &t
	}
	if closedAt.Valid {
		t := closedAt.Time
		rk.ClosedAt = &t
	}
	if ownerUsername.Valid {
		rk.OwnerUsername = ownerUsername.String
	}
	if projectName.Valid {
		rk.ProjectName = projectName.String
	}
	if taskLinksCount.Valid {
		rk.TaskLinksCount = int(taskLinksCount.Int32)
	}
	if openTaskLinksCount.Valid {
		rk.OpenTaskLinksCount = int(openTaskLinksCount.Int32)
	}

	return &rk, nil
}

const riskSelectQuery = `
	SELECT 
		r.id, r.workspace_id, r.project_id, r.risk_number, r.risk_code, r.title, r.description,
		r.category, r.probability, r.impact, r.score, r.severity, r.status, r.strategy,
		r.mitigation_plan, r.contingency_plan, r.owner_id, r.target_resolution_date, r.closed_at,
		r.created_at, r.updated_at,
		u.username AS owner_username,
		p.name AS project_name,
		(SELECT COUNT(*) FROM risk_task_links rtl WHERE rtl.risk_id = r.id) AS task_links_count,
		(SELECT COUNT(*) FROM risk_task_links rtl JOIN tasks t ON rtl.task_id = t.id WHERE rtl.risk_id = r.id AND t.completed = 0) AS open_task_links_count
	FROM risks r
	LEFT JOIN users u ON r.owner_id = u.id
	LEFT JOIN projects p ON r.project_id = p.id
`


func (r *riskRepositoryImpl) GetRiskByID(id int64) (*model.Risk, error) {
	query := riskSelectQuery + " WHERE r.id = ?"
	row := r.db.QueryRow(query, id)
	risk, err := scanRisk(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrRiskNotFound
		}
		return nil, err
	}
	return risk, nil
}

func (r *riskRepositoryImpl) GetRiskByCode(workspaceID int64, code string) (*model.Risk, error) {
	query := riskSelectQuery + " WHERE r.workspace_id = ? AND r.risk_code = ?"
	row := r.db.QueryRow(query, workspaceID, code)
	risk, err := scanRisk(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrRiskNotFound
		}
		return nil, err
	}
	return risk, nil
}

func (r *riskRepositoryImpl) GetRiskByNumber(workspaceID int64, number int) (*model.Risk, error) {
	query := riskSelectQuery + " WHERE r.workspace_id = ? AND r.risk_number = ?"
	row := r.db.QueryRow(query, workspaceID, number)
	risk, err := scanRisk(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrRiskNotFound
		}
		return nil, err
	}
	return risk, nil
}

func (r *riskRepositoryImpl) ListRisks(filter model.RiskFilter) ([]*model.Risk, error) {
	var conditions []string
	var args []interface{}

	conditions = append(conditions, "r.workspace_id = ?")
	args = append(args, filter.WorkspaceID)

	if filter.ProjectID != nil {
		conditions = append(conditions, "r.project_id = ?")
		args = append(args, *filter.ProjectID)
	}
	if filter.Category != "" {
		conditions = append(conditions, "r.category = ?")
		args = append(args, filter.Category)
	}
	if filter.Severity != "" {
		conditions = append(conditions, "r.severity = ?")
		args = append(args, filter.Severity)
	}
	if filter.Status != "" {
		conditions = append(conditions, "r.status = ?")
		args = append(args, filter.Status)
	}
	if filter.Strategy != "" {
		conditions = append(conditions, "r.strategy = ?")
		args = append(args, filter.Strategy)
	}
	if filter.OwnerID != "" {
		conditions = append(conditions, "r.owner_id = ?")
		args = append(args, filter.OwnerID)
	}
	if filter.MinScore > 0 {
		conditions = append(conditions, "r.score >= ?")
		args = append(args, filter.MinScore)
	}
	if filter.MaxScore > 0 {
		conditions = append(conditions, "r.score <= ?")
		args = append(args, filter.MaxScore)
	}
	if filter.Search != "" {
		conditions = append(conditions, "(r.title LIKE ? OR r.description LIKE ? OR r.risk_code LIKE ?)")
		pattern := "%" + filter.Search + "%"
		args = append(args, pattern, pattern, pattern)
	}

	whereClause := " WHERE " + strings.Join(conditions, " AND ")
	orderClause := " ORDER BY r.score DESC, r.created_at DESC"

	limit := filter.Limit
	if limit <= 0 {
		limit = 50
	} else if limit > 100 {
		limit = 100
	}
	offset := filter.Offset
	if offset < 0 {
		offset = 0
	}

	paginationClause := " LIMIT ? OFFSET ?"
	args = append(args, limit, offset)

	query := riskSelectQuery + whereClause + orderClause + paginationClause
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var risks []*model.Risk
	for rows.Next() {
		rk, err := scanRisk(rows)
		if err != nil {
			return nil, err
		}
		risks = append(risks, rk)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return risks, nil
}

func (r *riskRepositoryImpl) UpdateRisk(risk *model.Risk) error {
	risk.Score, risk.Severity = model.CalculateRiskSeverity(risk.Probability, risk.Impact)
	risk.UpdatedAt = time.Now().UTC()

	query := `
	UPDATE risks SET
		project_id = ?,
		title = ?,
		description = ?,
		category = ?,
		probability = ?,
		impact = ?,
		score = ?,
		severity = ?,
		status = ?,
		strategy = ?,
		mitigation_plan = ?,
		contingency_plan = ?,
		owner_id = ?,
		target_resolution_date = ?,
		closed_at = ?,
		updated_at = ?
	WHERE id = ?`

	res, err := r.db.Exec(query,
		risk.ProjectID,
		risk.Title,
		risk.Description,
		risk.Category,
		risk.Probability,
		risk.Impact,
		risk.Score,
		risk.Severity,
		risk.Status,
		risk.Strategy,
		risk.MitigationPlan,
		risk.ContingencyPlan,
		risk.OwnerID,
		risk.TargetResolutionDate,
		risk.ClosedAt,
		risk.UpdatedAt,
		risk.ID,
	)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrRiskNotFound
	}
	return nil
}

func (r *riskRepositoryImpl) DeleteRisk(id int64) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec("DELETE FROM risk_reviews WHERE risk_id = ?", id); err != nil {
		return err
	}
	if _, err := tx.Exec("DELETE FROM risk_task_links WHERE risk_id = ?", id); err != nil {
		return err
	}

	res, err := tx.Exec("DELETE FROM risks WHERE id = ?", id)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrRiskNotFound
	}

	return tx.Commit()
}

func (r *riskRepositoryImpl) LinkTask(link *model.RiskTaskLink) (int64, error) {
	if link.ActionType == "" {
		link.ActionType = model.RiskActionMitigation
	}
	now := time.Now().UTC()
	if link.CreatedAt.IsZero() {
		link.CreatedAt = now
	}

	query := `
	INSERT INTO risk_task_links (risk_id, task_id, action_type, notes, created_at)
	VALUES (?, ?, ?, ?, ?)
	ON CONFLICT(risk_id, task_id) DO UPDATE SET
		action_type = excluded.action_type,
		notes = excluded.notes`

	res, err := r.db.Exec(query, link.RiskID, link.TaskID, link.ActionType, link.Notes, link.CreatedAt)
	if err != nil {
		return 0, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	link.ID = id
	return id, nil
}

func (r *riskRepositoryImpl) UnlinkTask(riskID int64, taskID int) error {
	query := `DELETE FROM risk_task_links WHERE risk_id = ? AND task_id = ?`
	res, err := r.db.Exec(query, riskID, taskID)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return errors.New("risk task link not found")
	}
	return nil
}

func (r *riskRepositoryImpl) ListRiskTasks(riskID int64) ([]model.RiskTaskLink, error) {
	query := `
	SELECT 
		rtl.id, rtl.risk_id, rtl.task_id, rtl.action_type, rtl.notes, rtl.created_at,
		t.title, t.status, t.priority, t.completed
	FROM risk_task_links rtl
	JOIN tasks t ON rtl.task_id = t.id
	WHERE rtl.risk_id = ?
	ORDER BY rtl.created_at ASC`

	rows, err := r.db.Query(query, riskID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var links []model.RiskTaskLink
	for rows.Next() {
		var l model.RiskTaskLink
		var completed bool
		err := rows.Scan(
			&l.ID, &l.RiskID, &l.TaskID, &l.ActionType, &l.Notes, &l.CreatedAt,
			&l.TaskTitle, &l.TaskStatus, &l.TaskPriority, &completed,
		)
		if err != nil {
			return nil, err
		}
		l.TaskCompleted = completed
		links = append(links, l)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return links, nil
}

func (r *riskRepositoryImpl) CreateReview(review *model.RiskReview) (int64, error) {
	now := time.Now().UTC()
	if review.CreatedAt.IsZero() {
		review.CreatedAt = now
	}

	query := `
	INSERT INTO risk_reviews (
		risk_id, reviewer_id, previous_probability, previous_impact, previous_score,
		new_probability, new_impact, new_score, previous_status, new_status, comments, created_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	res, err := r.db.Exec(query,
		review.RiskID, review.ReviewerID, review.PreviousProbability, review.PreviousImpact, review.PreviousScore,
		review.NewProbability, review.NewImpact, review.NewScore, review.PreviousStatus, review.NewStatus, review.Comments, review.CreatedAt,
	)
	if err != nil {
		return 0, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	review.ID = id
	return id, nil
}

func (r *riskRepositoryImpl) ListReviews(riskID int64) ([]model.RiskReview, error) {
	query := `
	SELECT 
		rr.id, rr.risk_id, rr.reviewer_id, rr.previous_probability, rr.previous_impact, rr.previous_score,
		rr.new_probability, rr.new_impact, rr.new_score, rr.previous_status, rr.new_status, rr.comments, rr.created_at,
		COALESCE(u.username, '') AS reviewer_username
	FROM risk_reviews rr
	LEFT JOIN users u ON rr.reviewer_id = u.id
	WHERE rr.risk_id = ?
	ORDER BY rr.created_at DESC`

	rows, err := r.db.Query(query, riskID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reviews []model.RiskReview
	for rows.Next() {
		var rev model.RiskReview
		err := rows.Scan(
			&rev.ID, &rev.RiskID, &rev.ReviewerID, &rev.PreviousProbability, &rev.PreviousImpact, &rev.PreviousScore,
			&rev.NewProbability, &rev.NewImpact, &rev.NewScore, &rev.PreviousStatus, &rev.NewStatus, &rev.Comments, &rev.CreatedAt,
			&rev.ReviewerUsername,
		)
		if err != nil {
			return nil, err
		}
		reviews = append(reviews, rev)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return reviews, nil
}

func (r *riskRepositoryImpl) GetRiskMetrics(workspaceID int64) (*model.RiskMetrics, error) {
	metrics := &model.RiskMetrics{
		WorkspaceID:       workspaceID,
		SeverityBreakdown: make(map[string]int),
		StatusBreakdown:   make(map[string]int),
		CategoryBreakdown: make(map[string]int),
		StrategyBreakdown: make(map[string]int),
		CriticalRisks:     []model.Risk{},
	}

	// 1. Overall counts and exposure
	summaryQuery := `
	SELECT 
		COUNT(*),
		COALESCE(SUM(CASE WHEN status != 'closed' THEN 1 ELSE 0 END), 0),
		COALESCE(SUM(CASE WHEN status = 'closed' THEN 1 ELSE 0 END), 0),
		COALESCE(SUM(CASE WHEN status != 'closed' THEN score ELSE 0 END), 0),
		COALESCE(AVG(CASE WHEN status != 'closed' THEN score ELSE NULL END), 0.0)
	FROM risks
	WHERE workspace_id = ?`

	err := r.db.QueryRow(summaryQuery, workspaceID).Scan(
		&metrics.TotalRisks,
		&metrics.ActiveRisks,
		&metrics.ClosedRisks,
		&metrics.TotalExposureScore,
		&metrics.AverageRiskScore,
	)
	if err != nil {
		return nil, err
	}

	// 2. Severity breakdown (active)
	sevRows, err := r.db.Query(`SELECT severity, COUNT(*) FROM risks WHERE workspace_id = ? AND status != 'closed' GROUP BY severity`, workspaceID)
	if err != nil {
		return nil, err
	}
	defer sevRows.Close()
	for sevRows.Next() {
		var sev string
		var count int
		if err := sevRows.Scan(&sev, &count); err == nil {
			metrics.SeverityBreakdown[sev] = count
		}
	}

	// 3. Status breakdown (all)
	statusRows, err := r.db.Query(`SELECT status, COUNT(*) FROM risks WHERE workspace_id = ? GROUP BY status`, workspaceID)
	if err != nil {
		return nil, err
	}
	defer statusRows.Close()
	for statusRows.Next() {
		var st string
		var count int
		if err := statusRows.Scan(&st, &count); err == nil {
			metrics.StatusBreakdown[st] = count
		}
	}

	// 4. Category breakdown (active)
	catRows, err := r.db.Query(`SELECT category, COUNT(*) FROM risks WHERE workspace_id = ? AND status != 'closed' GROUP BY category`, workspaceID)
	if err != nil {
		return nil, err
	}
	defer catRows.Close()
	for catRows.Next() {
		var cat string
		var count int
		if err := catRows.Scan(&cat, &count); err == nil {
			metrics.CategoryBreakdown[cat] = count
		}
	}

	// 5. Strategy breakdown (active)
	stratRows, err := r.db.Query(`SELECT strategy, COUNT(*) FROM risks WHERE workspace_id = ? AND status != 'closed' GROUP BY strategy`, workspaceID)
	if err != nil {
		return nil, err
	}
	defer stratRows.Close()
	for stratRows.Next() {
		var strat string
		var count int
		if err := stratRows.Scan(&strat, &count); err == nil {
			metrics.StrategyBreakdown[strat] = count
		}
	}

	// 6. 5x5 heatmap matrix (active risks)
	matrixRows, err := r.db.Query(`SELECT probability, impact, COUNT(*) FROM risks WHERE workspace_id = ? AND status != 'closed' GROUP BY probability, impact`, workspaceID)
	if err != nil {
		return nil, err
	}
	defer matrixRows.Close()
	for matrixRows.Next() {
		var p, i, count int
		if err := matrixRows.Scan(&p, &i, &count); err == nil {
			if p >= 1 && p <= 5 && i >= 1 && i <= 5 {
				metrics.HeatmapMatrix[p-1][i-1] = count
			}
		}
	}

	// 7. Top critical / high risks
	critQuery := riskSelectQuery + ` WHERE r.workspace_id = ? AND r.status != 'closed' AND r.severity IN ('critical', 'high') ORDER BY r.score DESC, r.created_at DESC LIMIT 5`
	critRows, err := r.db.Query(critQuery, workspaceID)
	if err != nil {
		return nil, err
	}
	defer critRows.Close()
	for critRows.Next() {
		rk, err := scanRisk(critRows)
		if err != nil {
			return nil, err
		}
		metrics.CriticalRisks = append(metrics.CriticalRisks, *rk)
	}

	return metrics, nil
}
