package repository

import (
	"database/sql"
	"errors"
	"fmt"
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
	return nil, errors.New("not implemented")
}

func (r *riskRepositoryImpl) GetRiskByCode(workspaceID int64, code string) (*model.Risk, error) {
	return nil, errors.New("not implemented")
}

func (r *riskRepositoryImpl) GetRiskByNumber(workspaceID int64, number int) (*model.Risk, error) {
	return nil, errors.New("not implemented")
}

func (r *riskRepositoryImpl) ListRisks(filter model.RiskFilter) ([]*model.Risk, error) {
	return nil, errors.New("not implemented")
}

func (r *riskRepositoryImpl) UpdateRisk(risk *model.Risk) error {
	return errors.New("not implemented")
}

func (r *riskRepositoryImpl) DeleteRisk(id int64) error {
	return errors.New("not implemented")
}

func (r *riskRepositoryImpl) LinkTask(link *model.RiskTaskLink) (int64, error) {
	return 0, errors.New("not implemented")
}

func (r *riskRepositoryImpl) UnlinkTask(riskID int64, taskID int) error {
	return errors.New("not implemented")
}

func (r *riskRepositoryImpl) ListRiskTasks(riskID int64) ([]model.RiskTaskLink, error) {
	return nil, errors.New("not implemented")
}

func (r *riskRepositoryImpl) CreateReview(review *model.RiskReview) (int64, error) {
	return 0, errors.New("not implemented")
}

func (r *riskRepositoryImpl) ListReviews(riskID int64) ([]model.RiskReview, error) {
	return nil, errors.New("not implemented")
}

func (r *riskRepositoryImpl) GetRiskMetrics(workspaceID int64) (*model.RiskMetrics, error) {
	return nil, errors.New("not implemented")
}
