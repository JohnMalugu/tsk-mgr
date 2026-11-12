package repository

import (
	"database/sql"
	"errors"
	"time"

	"github.com/JohnMalugu/tsk-mgr-api/internal/model"
)

var (
	ErrScheduleNotFound         = errors.New("on-call schedule not found")
	ErrEscalationPolicyNotFound = errors.New("escalation policy not found")
	ErrEscalationStepNotFound   = errors.New("escalation step not found")
	ErrOnCallOverrideNotFound   = errors.New("on-call override not found")
	ErrOnCallPageNotFound       = errors.New("on-call page not found")
)

// OnCallRepository defines persistent operations for schedules, escalation policies, overrides, and pages
type OnCallRepository interface {
	// Schedules & Participants
	CreateSchedule(s *model.OnCallSchedule, participantIDs []string) (int64, error)
	GetScheduleByID(id int64) (*model.OnCallSchedule, error)
	ListSchedules(filter model.OnCallFilter) ([]model.OnCallSchedule, int, error)
	UpdateSchedule(s *model.OnCallSchedule) error
	DeleteSchedule(id int64) error
	SetParticipants(scheduleID int64, userIDs []string) error
	ListParticipants(scheduleID int64) ([]model.OnCallParticipant, error)

	// Computed Shifts
	SaveShifts(scheduleID int64, shifts []model.OnCallShift) error
	ListShifts(scheduleID int64, startTime, endTime time.Time) ([]model.OnCallShift, error)
	GetCurrentShift(scheduleID int64, now time.Time) (*model.OnCallShift, error)

	// Escalation Policies & Steps
	CreateEscalationPolicy(p *model.EscalationPolicy, steps []model.EscalationStep) (int64, error)
	GetEscalationPolicyByID(id int64) (*model.EscalationPolicy, error)
	GetDefaultEscalationPolicy(workspaceID int64) (*model.EscalationPolicy, error)
	ListEscalationPolicies(workspaceID int64) ([]model.EscalationPolicy, error)
	UpdateEscalationPolicy(p *model.EscalationPolicy) error
	DeleteEscalationPolicy(id int64) error
	SetEscalationSteps(policyID int64, steps []model.EscalationStep) error
	ListEscalationSteps(policyID int64) ([]model.EscalationStep, error)

	// Overrides
	CreateOverride(o *model.OnCallOverride) (int64, error)
	GetOverrideByID(id int64) (*model.OnCallOverride, error)
	ListOverrides(scheduleID int64) ([]model.OnCallOverride, error)
	DeleteOverride(id int64) error
	GetActiveOverride(scheduleID int64, now time.Time) (*model.OnCallOverride, error)

	// Pages & Alerts
	CreatePage(page *model.OnCallPage) (int64, error)
	GetPageByID(id int64) (*model.OnCallPage, error)
	ListPages(workspaceID int64, status string, limit, offset int) ([]model.OnCallPage, int, error)
	UpdatePage(page *model.OnCallPage) error

	// Metrics
	GetOnCallMetrics(workspaceID int64) (*model.OnCallMetrics, error)
}

type oncallRepositoryImpl struct {
	db *sql.DB
}

// NewOnCallRepository creates a new SQLite-backed OnCallRepository
func NewOnCallRepository(db *sql.DB) OnCallRepository {
	return &oncallRepositoryImpl{db: db}
}

func (r *oncallRepositoryImpl) CreateSchedule(s *model.OnCallSchedule, participantIDs []string) (int64, error) {
	return 0, nil
}

func (r *oncallRepositoryImpl) GetScheduleByID(id int64) (*model.OnCallSchedule, error) {
	return nil, nil
}

func (r *oncallRepositoryImpl) ListSchedules(filter model.OnCallFilter) ([]model.OnCallSchedule, int, error) {
	return nil, 0, nil
}

func (r *oncallRepositoryImpl) UpdateSchedule(s *model.OnCallSchedule) error {
	return nil
}

func (r *oncallRepositoryImpl) DeleteSchedule(id int64) error {
	return nil
}

func (r *oncallRepositoryImpl) SetParticipants(scheduleID int64, userIDs []string) error {
	return nil
}

func (r *oncallRepositoryImpl) ListParticipants(scheduleID int64) ([]model.OnCallParticipant, error) {
	return nil, nil
}

func (r *oncallRepositoryImpl) SaveShifts(scheduleID int64, shifts []model.OnCallShift) error {
	return nil
}

func (r *oncallRepositoryImpl) ListShifts(scheduleID int64, startTime, endTime time.Time) ([]model.OnCallShift, error) {
	return nil, nil
}

func (r *oncallRepositoryImpl) GetCurrentShift(scheduleID int64, now time.Time) (*model.OnCallShift, error) {
	return nil, nil
}

func (r *oncallRepositoryImpl) CreateEscalationPolicy(p *model.EscalationPolicy, steps []model.EscalationStep) (int64, error) {
	return 0, nil
}

func (r *oncallRepositoryImpl) GetEscalationPolicyByID(id int64) (*model.EscalationPolicy, error) {
	return nil, nil
}

func (r *oncallRepositoryImpl) GetDefaultEscalationPolicy(workspaceID int64) (*model.EscalationPolicy, error) {
	return nil, nil
}

func (r *oncallRepositoryImpl) ListEscalationPolicies(workspaceID int64) ([]model.EscalationPolicy, error) {
	return nil, nil
}

func (r *oncallRepositoryImpl) UpdateEscalationPolicy(p *model.EscalationPolicy) error {
	return nil
}

func (r *oncallRepositoryImpl) DeleteEscalationPolicy(id int64) error {
	return nil
}

func (r *oncallRepositoryImpl) SetEscalationSteps(policyID int64, steps []model.EscalationStep) error {
	return nil
}

func (r *oncallRepositoryImpl) ListEscalationSteps(policyID int64) ([]model.EscalationStep, error) {
	return nil, nil
}

func (r *oncallRepositoryImpl) CreateOverride(o *model.OnCallOverride) (int64, error) {
	return 0, nil
}

func (r *oncallRepositoryImpl) GetOverrideByID(id int64) (*model.OnCallOverride, error) {
	return nil, nil
}

func (r *oncallRepositoryImpl) ListOverrides(scheduleID int64) ([]model.OnCallOverride, error) {
	return nil, nil
}

func (r *oncallRepositoryImpl) DeleteOverride(id int64) error {
	return nil
}

func (r *oncallRepositoryImpl) GetActiveOverride(scheduleID int64, now time.Time) (*model.OnCallOverride, error) {
	return nil, nil
}

func (r *oncallRepositoryImpl) CreatePage(page *model.OnCallPage) (int64, error) {
	return 0, nil
}

func (r *oncallRepositoryImpl) GetPageByID(id int64) (*model.OnCallPage, error) {
	return nil, nil
}

func (r *oncallRepositoryImpl) ListPages(workspaceID int64, status string, limit, offset int) ([]model.OnCallPage, int, error) {
	return nil, 0, nil
}

func (r *oncallRepositoryImpl) UpdatePage(page *model.OnCallPage) error {
	return nil
}

func (r *oncallRepositoryImpl) GetOnCallMetrics(workspaceID int64) (*model.OnCallMetrics, error) {
	return nil, nil
}
