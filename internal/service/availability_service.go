package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/JohnMalugu/tsk-mgr-api/internal/model"
	"github.com/JohnMalugu/tsk-mgr-api/internal/repository"
)

var (
	_ = strings.TrimSpace
	_ = time.Now
	_ = errors.New
)

var (
	ErrAvailabilityTargetNotFound      = repository.ErrAvailabilityTargetNotFound
	ErrAvailabilityTargetNameRequired  = repository.ErrAvailabilityTargetNameRequired
	ErrAvailabilityTargetNameConflict  = repository.ErrAvailabilityTargetNameConflict
	ErrOutageEventNotFound             = repository.ErrOutageEventNotFound
	ErrAvailabilityMeasurementNotFound = repository.ErrAvailabilityMeasurementNotFound
	ErrSPOFRecordNotFound              = repository.ErrSPOFRecordNotFound
	ErrAvailabilityTestNotFound        = repository.ErrAvailabilityTestNotFound
)

// AvailabilityService defines business logic and orchestration for ITIL Availability Management
type AvailabilityService interface {
	// Targets
	CreateTarget(ctx context.Context, workspaceID string, req model.CreateAvailabilityTargetRequest) (*model.AvailabilityTarget, error)
	GetTargetByID(ctx context.Context, id string) (*model.AvailabilityTarget, error)
	GetTargetByName(ctx context.Context, workspaceID string, name string) (*model.AvailabilityTarget, error)
	UpdateTarget(ctx context.Context, id string, req model.UpdateAvailabilityTargetRequest) (*model.AvailabilityTarget, error)
	DeleteTarget(ctx context.Context, id string) error
	ListTargets(ctx context.Context, filter model.AvailabilityFilterCriteria) ([]model.AvailabilityTarget, int, error)

	// Outages
	RecordOutage(ctx context.Context, workspaceID string, req model.CreateOutageEventRequest) (*model.OutageEvent, error)
	GetOutageByID(ctx context.Context, id string) (*model.OutageEvent, error)
	UpdateOutage(ctx context.Context, id string, req model.UpdateOutageEventRequest) (*model.OutageEvent, error)
	ResolveOutage(ctx context.Context, id string, req model.ResolveOutageRequest) (*model.OutageEvent, error)
	DeleteOutage(ctx context.Context, id string) error
	ListOutages(ctx context.Context, filter model.AvailabilityFilterCriteria) ([]model.OutageEvent, int, error)

	// Measurements
	RecordMeasurement(ctx context.Context, workspaceID string, req model.CreateMeasurementRequest) (*model.AvailabilityMeasurement, error)
	CalculatePeriodMeasurement(ctx context.Context, workspaceID string, targetID string, period string, totalOperationalMinutes float64) (*model.AvailabilityMeasurement, error)
	GetMeasurementByID(ctx context.Context, id string) (*model.AvailabilityMeasurement, error)
	ListMeasurements(ctx context.Context, filter model.AvailabilityFilterCriteria) ([]model.AvailabilityMeasurement, int, error)

	// SPOFs
	RegisterSPOF(ctx context.Context, workspaceID string, req model.CreateSPOFRequest) (*model.SPOFRecord, error)
	GetSPOFByID(ctx context.Context, id string) (*model.SPOFRecord, error)
	UpdateSPOF(ctx context.Context, id string, req model.UpdateSPOFRequest) (*model.SPOFRecord, error)
	MitigateSPOF(ctx context.Context, id string, mitigationPlan string) (*model.SPOFRecord, error)
	DeleteSPOF(ctx context.Context, id string) error
	ListSPOFs(ctx context.Context, filter model.AvailabilityFilterCriteria) ([]model.SPOFRecord, int, error)

	// Availability Tests
	ScheduleTest(ctx context.Context, workspaceID string, req model.CreateAvailabilityTestRequest) (*model.AvailabilityTest, error)
	GetTestByID(ctx context.Context, id string) (*model.AvailabilityTest, error)
	UpdateTest(ctx context.Context, id string, req model.UpdateAvailabilityTestRequest) (*model.AvailabilityTest, error)
	RecordTestResult(ctx context.Context, id string, req model.RecordTestResultRequest) (*model.AvailabilityTest, error)
	DeleteTest(ctx context.Context, id string) error
	ListTests(ctx context.Context, filter model.AvailabilityFilterCriteria) ([]model.AvailabilityTest, int, error)

	// Posture Analytics
	GetAvailabilityAnalytics(ctx context.Context, workspaceID string) (*model.AvailabilityAnalytics, error)
}

type availabilityService struct {
	repo repository.AvailabilityRepository
}

// NewAvailabilityService instantiates the availability management service
func NewAvailabilityService(repo repository.AvailabilityRepository) AvailabilityService {
	return &availabilityService{repo: repo}
}

// CreateTarget creates and registers an availability target with default or configured SLA parameters
func (s *availabilityService) CreateTarget(ctx context.Context, workspaceID string, req model.CreateAvailabilityTargetRequest) (*model.AvailabilityTarget, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	target := &model.AvailabilityTarget{
		WorkspaceID:        workspaceID,
		Name:               strings.TrimSpace(req.Name),
		TargetPercentage:   req.TargetPercentage,
		MeasurementWindow:  req.MeasurementWindow,
		ServiceHours:       req.ServiceHours,
		CommittedMTBFHours: req.CommittedMTBFHours,
		TargetMTTRMinutes:  req.TargetMTTRMinutes,
		CIID:               strings.TrimSpace(req.CIID),
		Status:             req.Status,
		Description:        strings.TrimSpace(req.Description),
	}

	if err := s.repo.CreateTarget(ctx, target); err != nil {
		return nil, err
	}
	return target, nil
}

// GetTargetByID retrieves an availability target by ID
func (s *availabilityService) GetTargetByID(ctx context.Context, id string) (*model.AvailabilityTarget, error) {
	return s.repo.GetTargetByID(ctx, id)
}

// GetTargetByName retrieves a target by name within a workspace
func (s *availabilityService) GetTargetByName(ctx context.Context, workspaceID string, name string) (*model.AvailabilityTarget, error) {
	return s.repo.GetTargetByName(ctx, workspaceID, name)
}

// UpdateTarget applies updates to an existing availability target
func (s *availabilityService) UpdateTarget(ctx context.Context, id string, req model.UpdateAvailabilityTargetRequest) (*model.AvailabilityTarget, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	target, err := s.repo.GetTargetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.Name != nil {
		target.Name = strings.TrimSpace(*req.Name)
	}
	if req.TargetPercentage != nil {
		target.TargetPercentage = *req.TargetPercentage
	}
	if req.MeasurementWindow != nil {
		target.MeasurementWindow = *req.MeasurementWindow
	}
	if req.ServiceHours != nil {
		target.ServiceHours = *req.ServiceHours
	}
	if req.CommittedMTBFHours != nil {
		target.CommittedMTBFHours = *req.CommittedMTBFHours
	}
	if req.TargetMTTRMinutes != nil {
		target.TargetMTTRMinutes = *req.TargetMTTRMinutes
	}
	if req.CIID != nil {
		target.CIID = strings.TrimSpace(*req.CIID)
	}
	if req.Status != nil {
		target.Status = *req.Status
	}
	if req.Description != nil {
		target.Description = strings.TrimSpace(*req.Description)
	}

	if err := s.repo.UpdateTarget(ctx, target); err != nil {
		return nil, err
	}
	return target, nil
}

// DeleteTarget deletes an availability target
func (s *availabilityService) DeleteTarget(ctx context.Context, id string) error {
	return s.repo.DeleteTarget(ctx, id)
}

// ListTargets returns targets matching filter criteria
func (s *availabilityService) ListTargets(ctx context.Context, filter model.AvailabilityFilterCriteria) ([]model.AvailabilityTarget, int, error) {
	return s.repo.ListTargets(ctx, filter)
}

// RecordOutage records a new outage event, calculating duration if already concluded
func (s *availabilityService) RecordOutage(ctx context.Context, workspaceID string, req model.CreateOutageEventRequest) (*model.OutageEvent, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	outage := &model.OutageEvent{
		WorkspaceID:     workspaceID,
		TargetID:        strings.TrimSpace(req.TargetID),
		CIID:            strings.TrimSpace(req.CIID),
		Title:           strings.TrimSpace(req.Title),
		OutageType:      req.OutageType,
		Severity:        req.Severity,
		StartedAt:       req.StartedAt,
		EndedAt:         req.EndedAt,
		RootCause:       strings.TrimSpace(req.RootCause),
		IsSLAExempt:     req.IsSLAExempt,
		ExemptionReason: strings.TrimSpace(req.ExemptionReason),
		ReportedBy:      strings.TrimSpace(req.ReportedBy),
	}

	if outage.EndedAt != nil {
		outage.DurationMinutes = outage.CalculateDuration()
	}

	if err := s.repo.CreateOutage(ctx, outage); err != nil {
		return nil, err
	}
	return outage, nil
}

// GetOutageByID retrieves an outage event by ID
func (s *availabilityService) GetOutageByID(ctx context.Context, id string) (*model.OutageEvent, error) {
	return s.repo.GetOutageByID(ctx, id)
}

// UpdateOutage modifies details of an ongoing or concluded outage event
func (s *availabilityService) UpdateOutage(ctx context.Context, id string, req model.UpdateOutageEventRequest) (*model.OutageEvent, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	outage, err := s.repo.GetOutageByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.Title != nil {
		outage.Title = strings.TrimSpace(*req.Title)
	}
	if req.OutageType != nil {
		outage.OutageType = *req.OutageType
	}
	if req.Severity != nil {
		outage.Severity = *req.Severity
	}
	if req.StartedAt != nil {
		outage.StartedAt = *req.StartedAt
	}
	if req.EndedAt != nil {
		outage.EndedAt = req.EndedAt
	}
	if req.RootCause != nil {
		outage.RootCause = strings.TrimSpace(*req.RootCause)
	}
	if req.IsSLAExempt != nil {
		outage.IsSLAExempt = *req.IsSLAExempt
	}
	if req.ExemptionReason != nil {
		outage.ExemptionReason = strings.TrimSpace(*req.ExemptionReason)
	}

	if outage.EndedAt != nil {
		outage.DurationMinutes = outage.CalculateDuration()
	}

	if err := s.repo.UpdateOutage(ctx, outage); err != nil {
		return nil, err
	}
	return outage, nil
}

// ResolveOutage marks an active outage as resolved and updates duration
func (s *availabilityService) ResolveOutage(ctx context.Context, id string, req model.ResolveOutageRequest) (*model.OutageEvent, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	outage, err := s.repo.GetOutageByID(ctx, id)
	if err != nil {
		return nil, err
	}

	endTime := req.EndedAt
	if endTime.IsZero() {
		endTime = time.Now().UTC()
	}
	outage.EndedAt = &endTime
	if req.RootCause != "" {
		outage.RootCause = strings.TrimSpace(req.RootCause)
	}
	if req.IsSLAExempt != nil {
		outage.IsSLAExempt = *req.IsSLAExempt
	}
	if req.ExemptionReason != "" {
		outage.ExemptionReason = strings.TrimSpace(req.ExemptionReason)
	}

	outage.DurationMinutes = outage.CalculateDuration()

	if err := s.repo.UpdateOutage(ctx, outage); err != nil {
		return nil, err
	}
	return outage, nil
}

// DeleteOutage deletes an outage record
func (s *availabilityService) DeleteOutage(ctx context.Context, id string) error {
	return s.repo.DeleteOutage(ctx, id)
}

// ListOutages returns outages matching query filter
func (s *availabilityService) ListOutages(ctx context.Context, filter model.AvailabilityFilterCriteria) ([]model.OutageEvent, int, error) {
	return s.repo.ListOutages(ctx, filter)
}

// RecordMeasurement manually records or overwrites a periodic availability calculation
func (s *availabilityService) RecordMeasurement(ctx context.Context, workspaceID string, req model.CreateMeasurementRequest) (*model.AvailabilityMeasurement, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	target, err := s.repo.GetTargetByID(ctx, req.TargetID)
	if err != nil {
		return nil, err
	}

	m := &model.AvailabilityMeasurement{
		WorkspaceID:             workspaceID,
		TargetID:                req.TargetID,
		Period:                  strings.TrimSpace(req.Period),
		TotalOperationalMinutes: req.TotalOperationalMinutes,
		TotalDowntimeMinutes:   req.TotalDowntimeMinutes,
		ExemptDowntimeMinutes:  req.ExemptDowntimeMinutes,
		MTBFHours:              req.MTBFHours,
		MTTRMinutes:            req.MTTRMinutes,
		OutageCount:            req.OutageCount,
		CalculatedAt:           time.Now().UTC(),
	}

	m.ActualUptimePercentage = m.ComputeUptimePercentage()
	m.IsSLABreached = m.CheckSLABreach(target.TargetPercentage)

	if err := s.repo.SaveMeasurement(ctx, m); err != nil {
		return nil, err
	}
	return m, nil
}

// CalculatePeriodMeasurement automates measurement rollup from outages for a specific target and period
func (s *availabilityService) CalculatePeriodMeasurement(ctx context.Context, workspaceID string, targetID string, period string, totalOperationalMinutes float64) (*model.AvailabilityMeasurement, error) {
	target, err := s.repo.GetTargetByID(ctx, targetID)
	if err != nil {
		return nil, err
	}

	if totalOperationalMinutes <= 0 {
		switch target.MeasurementWindow {
		case model.MeasurementWindowDaily:
			totalOperationalMinutes = 24.0 * 60.0
		case model.MeasurementWindowWeekly:
			totalOperationalMinutes = 7.0 * 24.0 * 60.0
		case model.MeasurementWindowQuarterly:
			totalOperationalMinutes = 90.0 * 24.0 * 60.0
		case model.MeasurementWindowAnnual:
			totalOperationalMinutes = 365.0 * 24.0 * 60.0
		default: // Monthly
			totalOperationalMinutes = 30.0 * 24.0 * 60.0
		}
	}

	// Fetch outages for this target
	outages, _, err := s.repo.ListOutages(ctx, model.AvailabilityFilterCriteria{
		WorkspaceID: workspaceID,
		TargetID:    targetID,
		Limit:       500,
	})
	if err != nil {
		return nil, err
	}

	totalDowntime := 0.0
	exemptDowntime := 0.0
	outageCount := len(outages)

	for _, o := range outages {
		dur := o.DurationMinutes
		if dur <= 0 && o.EndedAt != nil {
			dur = o.CalculateDuration()
		}
		totalDowntime += dur
		if o.IsSLAExempt {
			exemptDowntime += dur
		}
	}

	var mttr float64
	var mtbf float64
	if outageCount > 0 {
		mttr = totalDowntime / float64(outageCount)
		uptimeMins := totalOperationalMinutes - totalDowntime
		if uptimeMins < 0 {
			uptimeMins = 0
		}
		mtbf = (uptimeMins / float64(outageCount)) / 60.0
	} else {
		mtbf = totalOperationalMinutes / 60.0
	}

	m := &model.AvailabilityMeasurement{
		WorkspaceID:             workspaceID,
		TargetID:                targetID,
		Period:                  strings.TrimSpace(period),
		TotalOperationalMinutes: totalOperationalMinutes,
		TotalDowntimeMinutes:   totalDowntime,
		ExemptDowntimeMinutes:  exemptDowntime,
		MTBFHours:              mtbf,
		MTTRMinutes:            mttr,
		OutageCount:            outageCount,
		CalculatedAt:           time.Now().UTC(),
	}

	m.ActualUptimePercentage = m.ComputeUptimePercentage()
	m.IsSLABreached = m.CheckSLABreach(target.TargetPercentage)

	if err := s.repo.SaveMeasurement(ctx, m); err != nil {
		return nil, err
	}
	return m, nil
}

// GetMeasurementByID retrieves an availability measurement by ID
func (s *availabilityService) GetMeasurementByID(ctx context.Context, id string) (*model.AvailabilityMeasurement, error) {
	return s.repo.GetMeasurementByID(ctx, id)
}

// ListMeasurements returns measurements matching filter criteria
func (s *availabilityService) ListMeasurements(ctx context.Context, filter model.AvailabilityFilterCriteria) ([]model.AvailabilityMeasurement, int, error) {
	return s.repo.ListMeasurements(ctx, filter)
}

// GetAvailabilityAnalytics computes executive posture, resilience health index, and SLA compliance
func (s *availabilityService) GetAvailabilityAnalytics(ctx context.Context, workspaceID string) (*model.AvailabilityAnalytics, error) {
	return s.repo.GetAvailabilityAnalytics(ctx, workspaceID)
}

// RegisterSPOF logs a new architectural single point of failure
func (s *availabilityService) RegisterSPOF(ctx context.Context, workspaceID string, req model.CreateSPOFRequest) (*model.SPOFRecord, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	spof := &model.SPOFRecord{
		WorkspaceID:         workspaceID,
		ComponentName:       strings.TrimSpace(req.ComponentName),
		CIID:                strings.TrimSpace(req.CIID),
		RiskLevel:           req.RiskLevel,
		RedundancyMechanism: req.RedundancyMechanism,
		MitigationPlan:      strings.TrimSpace(req.MitigationPlan),
		Status:              req.Status,
		ReviewDate:          req.ReviewDate,
		OwnerID:             strings.TrimSpace(req.OwnerID),
	}

	if err := s.repo.CreateSPOF(ctx, spof); err != nil {
		return nil, err
	}
	return spof, nil
}

// GetSPOFByID retrieves a SPOF record by ID
func (s *availabilityService) GetSPOFByID(ctx context.Context, id string) (*model.SPOFRecord, error) {
	return s.repo.GetSPOFByID(ctx, id)
}

// UpdateSPOF updates fields of an existing SPOF record
func (s *availabilityService) UpdateSPOF(ctx context.Context, id string, req model.UpdateSPOFRequest) (*model.SPOFRecord, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	spof, err := s.repo.GetSPOFByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.ComponentName != nil {
		spof.ComponentName = strings.TrimSpace(*req.ComponentName)
	}
	if req.RiskLevel != nil {
		spof.RiskLevel = *req.RiskLevel
	}
	if req.RedundancyMechanism != nil {
		spof.RedundancyMechanism = *req.RedundancyMechanism
	}
	if req.MitigationPlan != nil {
		spof.MitigationPlan = strings.TrimSpace(*req.MitigationPlan)
	}
	if req.Status != nil {
		spof.Status = *req.Status
	}
	if req.ReviewDate != nil {
		spof.ReviewDate = *req.ReviewDate
	}
	if req.OwnerID != nil {
		spof.OwnerID = strings.TrimSpace(*req.OwnerID)
	}

	if err := s.repo.UpdateSPOF(ctx, spof); err != nil {
		return nil, err
	}
	return spof, nil
}

// MitigateSPOF updates a SPOF's status to mitigated and stores the plan
func (s *availabilityService) MitigateSPOF(ctx context.Context, id string, mitigationPlan string) (*model.SPOFRecord, error) {
	spof, err := s.repo.GetSPOFByID(ctx, id)
	if err != nil {
		return nil, err
	}
	spof.Status = model.SPOFStatusMitigated
	if mitigationPlan != "" {
		spof.MitigationPlan = strings.TrimSpace(mitigationPlan)
	}
	if err := s.repo.UpdateSPOF(ctx, spof); err != nil {
		return nil, err
	}
	return spof, nil
}

// DeleteSPOF deletes a SPOF record
func (s *availabilityService) DeleteSPOF(ctx context.Context, id string) error {
	return s.repo.DeleteSPOF(ctx, id)
}

// ListSPOFs returns SPOF records matching filter criteria
func (s *availabilityService) ListSPOFs(ctx context.Context, filter model.AvailabilityFilterCriteria) ([]model.SPOFRecord, int, error) {
	return s.repo.ListSPOFs(ctx, filter)
}

// ScheduleTest registers a new failover test schedule
func (s *availabilityService) ScheduleTest(ctx context.Context, workspaceID string, req model.CreateAvailabilityTestRequest) (*model.AvailabilityTest, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	test := &model.AvailabilityTest{
		WorkspaceID: workspaceID,
		TargetID:    strings.TrimSpace(req.TargetID),
		Name:        strings.TrimSpace(req.Name),
		TestType:    req.TestType,
		Frequency:   req.Frequency,
		NextTestDue: req.NextTestDue,
		Outcome:     model.TestOutcomePending,
		Findings:    strings.TrimSpace(req.Findings),
		ConductedBy: strings.TrimSpace(req.ConductedBy),
	}

	if err := s.repo.CreateTest(ctx, test); err != nil {
		return nil, err
	}
	return test, nil
}

// GetTestByID retrieves an availability test by ID
func (s *availabilityService) GetTestByID(ctx context.Context, id string) (*model.AvailabilityTest, error) {
	return s.repo.GetTestByID(ctx, id)
}

// UpdateTest modifies test schedule details
func (s *availabilityService) UpdateTest(ctx context.Context, id string, req model.UpdateAvailabilityTestRequest) (*model.AvailabilityTest, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	test, err := s.repo.GetTestByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.Name != nil {
		test.Name = strings.TrimSpace(*req.Name)
	}
	if req.TestType != nil {
		test.TestType = *req.TestType
	}
	if req.Frequency != nil {
		test.Frequency = *req.Frequency
	}
	if req.NextTestDue != nil {
		test.NextTestDue = *req.NextTestDue
	}
	if req.Findings != nil {
		test.Findings = strings.TrimSpace(*req.Findings)
	}
	if req.ConductedBy != nil {
		test.ConductedBy = strings.TrimSpace(*req.ConductedBy)
	}

	if err := s.repo.UpdateTest(ctx, test); err != nil {
		return nil, err
	}
	return test, nil
}

// RecordTestResult records the outcome, conducted timestamp, and next test date
func (s *availabilityService) RecordTestResult(ctx context.Context, id string, req model.RecordTestResultRequest) (*model.AvailabilityTest, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	test, err := s.repo.GetTestByID(ctx, id)
	if err != nil {
		return nil, err
	}

	testedAt := req.TestedAt
	if testedAt.IsZero() {
		testedAt = time.Now().UTC()
	}
	test.LastTestedAt = &testedAt
	test.Outcome = req.Outcome
	if req.Findings != "" {
		test.Findings = strings.TrimSpace(req.Findings)
	}
	test.NextTestDue = req.NextTestDue
	if req.ConductedBy != "" {
		test.ConductedBy = strings.TrimSpace(req.ConductedBy)
	}

	if err := s.repo.UpdateTest(ctx, test); err != nil {
		return nil, err
	}
	return test, nil
}

// DeleteTest deletes an availability test record
func (s *availabilityService) DeleteTest(ctx context.Context, id string) error {
	return s.repo.DeleteTest(ctx, id)
}

// ListTests returns availability tests matching filter criteria
func (s *availabilityService) ListTests(ctx context.Context, filter model.AvailabilityFilterCriteria) ([]model.AvailabilityTest, int, error) {
	return s.repo.ListTests(ctx, filter)
}


