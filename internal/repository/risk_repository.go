package repository

import (
	"database/sql"
	"errors"

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

func (r *riskRepositoryImpl) CreateRisk(risk *model.Risk) (int64, error) {
	return 0, errors.New("not implemented")
}

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

func (r *riskRepositoryImpl) GetNextRiskNumber(workspaceID int64) (int, error) {
	return 0, errors.New("not implemented")
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
