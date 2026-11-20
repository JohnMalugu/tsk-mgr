package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/JohnMalugu/tsk-mgr-api/internal/model"
)

var (
	_ = uuid.New
	_ = fmt.Sprintf
	_ = strings.Join
	_ = time.Now
	_ = errors.New
)

var (
	ErrAvailabilityTargetNotFound      = model.ErrAvailabilityTargetNotFound
	ErrAvailabilityTargetNameRequired  = model.ErrAvailabilityTargetNameRequired
	ErrAvailabilityTargetNameConflict  = errors.New("availability target with this name already exists in workspace")
	ErrOutageEventNotFound             = model.ErrOutageEventNotFound
	ErrAvailabilityMeasurementNotFound = model.ErrAvailabilityMeasurementNotFound
	ErrSPOFRecordNotFound              = model.ErrSPOFRecordNotFound
	ErrAvailabilityTestNotFound        = model.ErrAvailabilityTestNotFound
)

// AvailabilityRepository defines data access methods for ITIL Availability Management
type AvailabilityRepository interface {
	// Targets
	CreateTarget(ctx context.Context, target *model.AvailabilityTarget) error
	GetTargetByID(ctx context.Context, id string) (*model.AvailabilityTarget, error)
	GetTargetByName(ctx context.Context, workspaceID string, name string) (*model.AvailabilityTarget, error)
	UpdateTarget(ctx context.Context, target *model.AvailabilityTarget) error
	DeleteTarget(ctx context.Context, id string) error
	ListTargets(ctx context.Context, filter model.AvailabilityFilterCriteria) ([]model.AvailabilityTarget, int, error)

	// Outages
	CreateOutage(ctx context.Context, outage *model.OutageEvent) error
	GetOutageByID(ctx context.Context, id string) (*model.OutageEvent, error)
	UpdateOutage(ctx context.Context, outage *model.OutageEvent) error
	DeleteOutage(ctx context.Context, id string) error
	ListOutages(ctx context.Context, filter model.AvailabilityFilterCriteria) ([]model.OutageEvent, int, error)

	// Measurements
	SaveMeasurement(ctx context.Context, m *model.AvailabilityMeasurement) error
	GetMeasurementByID(ctx context.Context, id string) (*model.AvailabilityMeasurement, error)
	GetMeasurementByPeriod(ctx context.Context, targetID string, period string) (*model.AvailabilityMeasurement, error)
	ListMeasurements(ctx context.Context, filter model.AvailabilityFilterCriteria) ([]model.AvailabilityMeasurement, int, error)

	// SPOFs
	CreateSPOF(ctx context.Context, spof *model.SPOFRecord) error
	GetSPOFByID(ctx context.Context, id string) (*model.SPOFRecord, error)
	UpdateSPOF(ctx context.Context, spof *model.SPOFRecord) error
	DeleteSPOF(ctx context.Context, id string) error
	ListSPOFs(ctx context.Context, filter model.AvailabilityFilterCriteria) ([]model.SPOFRecord, int, error)

	// Availability Tests
	CreateTest(ctx context.Context, test *model.AvailabilityTest) error
	GetTestByID(ctx context.Context, id string) (*model.AvailabilityTest, error)
	UpdateTest(ctx context.Context, test *model.AvailabilityTest) error
	DeleteTest(ctx context.Context, id string) error
	ListTests(ctx context.Context, filter model.AvailabilityFilterCriteria) ([]model.AvailabilityTest, int, error)

	// Posture Analytics
	GetAvailabilityAnalytics(ctx context.Context, workspaceID string) (*model.AvailabilityAnalytics, error)
}

type availabilityRepository struct {
	db *sql.DB
}

// NewAvailabilityRepository creates a new SQLite repository for ITIL Availability Management
func NewAvailabilityRepository(db *sql.DB) AvailabilityRepository {
	return &availabilityRepository{db: db}
}

// CreateTarget persists a new availability target into the database
func (r *availabilityRepository) CreateTarget(ctx context.Context, target *model.AvailabilityTarget) error {
	if err := target.Validate(); err != nil {
		return err
	}
	if target.ID == "" {
		target.ID = uuid.New().String()
	}
	now := time.Now().UTC()
	if target.CreatedAt.IsZero() {
		target.CreatedAt = now
	}
	target.UpdatedAt = now

	// Check for unique name in workspace
	existing, err := r.GetTargetByName(ctx, target.WorkspaceID, target.Name)
	if err == nil && existing != nil {
		return ErrAvailabilityTargetNameConflict
	}

	query := `
	INSERT INTO availability_targets (
		id, workspace_id, name, target_percentage, measurement_window,
		service_hours, committed_mtbf_hours, target_mttr_minutes,
		ci_id, status, description, created_at, updated_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err = r.db.ExecContext(ctx, query,
		target.ID, target.WorkspaceID, target.Name, target.TargetPercentage, target.MeasurementWindow,
		target.ServiceHours, target.CommittedMTBFHours, target.TargetMTTRMinutes,
		target.CIID, target.Status, target.Description, target.CreatedAt, target.UpdatedAt,
	)
	if err != nil {
		return err
	}
	return nil
}

// GetTargetByID retrieves an availability target by its unique ID
func (r *availabilityRepository) GetTargetByID(ctx context.Context, id string) (*model.AvailabilityTarget, error) {
	query := `
	SELECT id, workspace_id, name, target_percentage, measurement_window,
	       service_hours, committed_mtbf_hours, target_mttr_minutes,
	       ci_id, status, description, created_at, updated_at
	FROM availability_targets WHERE id = ?`

	var t model.AvailabilityTarget
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&t.ID, &t.WorkspaceID, &t.Name, &t.TargetPercentage, &t.MeasurementWindow,
		&t.ServiceHours, &t.CommittedMTBFHours, &t.TargetMTTRMinutes,
		&t.CIID, &t.Status, &t.Description, &t.CreatedAt, &t.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrAvailabilityTargetNotFound
		}
		return nil, err
	}
	return &t, nil
}

// GetTargetByName retrieves an availability target by workspace and name
func (r *availabilityRepository) GetTargetByName(ctx context.Context, workspaceID string, name string) (*model.AvailabilityTarget, error) {
	query := `
	SELECT id, workspace_id, name, target_percentage, measurement_window,
	       service_hours, committed_mtbf_hours, target_mttr_minutes,
	       ci_id, status, description, created_at, updated_at
	FROM availability_targets WHERE workspace_id = ? AND name = ?`

	var t model.AvailabilityTarget
	err := r.db.QueryRowContext(ctx, query, workspaceID, name).Scan(
		&t.ID, &t.WorkspaceID, &t.Name, &t.TargetPercentage, &t.MeasurementWindow,
		&t.ServiceHours, &t.CommittedMTBFHours, &t.TargetMTTRMinutes,
		&t.CIID, &t.Status, &t.Description, &t.CreatedAt, &t.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrAvailabilityTargetNotFound
		}
		return nil, err
	}
	return &t, nil
}

// UpdateTarget updates an existing target in SQLite
func (r *availabilityRepository) UpdateTarget(ctx context.Context, target *model.AvailabilityTarget) error {
	if err := target.Validate(); err != nil {
		return err
	}
	target.UpdatedAt = time.Now().UTC()

	query := `
	UPDATE availability_targets SET
		name = ?, target_percentage = ?, measurement_window = ?,
		service_hours = ?, committed_mtbf_hours = ?, target_mttr_minutes = ?,
		ci_id = ?, status = ?, description = ?, updated_at = ?
	WHERE id = ?`

	res, err := r.db.ExecContext(ctx, query,
		target.Name, target.TargetPercentage, target.MeasurementWindow,
		target.ServiceHours, target.CommittedMTBFHours, target.TargetMTTRMinutes,
		target.CIID, target.Status, target.Description, target.UpdatedAt,
		target.ID,
	)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrAvailabilityTargetNotFound
	}
	return nil
}

// DeleteTarget deletes an availability target by ID
func (r *availabilityRepository) DeleteTarget(ctx context.Context, id string) error {
	query := `DELETE FROM availability_targets WHERE id = ?`
	res, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrAvailabilityTargetNotFound
	}
	return nil
}

// ListTargets returns a filtered list of availability targets
func (r *availabilityRepository) ListTargets(ctx context.Context, filter model.AvailabilityFilterCriteria) ([]model.AvailabilityTarget, int, error) {
	var conditions []string
	var args []interface{}

	if filter.WorkspaceID != "" {
		conditions = append(conditions, "workspace_id = ?")
		args = append(args, filter.WorkspaceID)
	}
	if filter.Status != "" {
		conditions = append(conditions, "status = ?")
		args = append(args, filter.Status)
	}
	if filter.CIID != "" {
		conditions = append(conditions, "ci_id = ?")
		args = append(args, filter.CIID)
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM availability_targets %s", whereClause)
	var total int
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	limit := filter.Limit
	if limit <= 0 {
		limit = 50
	}
	offset := filter.Offset
	if offset < 0 {
		offset = 0
	}

	query := fmt.Sprintf(`
	SELECT id, workspace_id, name, target_percentage, measurement_window,
	       service_hours, committed_mtbf_hours, target_mttr_minutes,
	       ci_id, status, description, created_at, updated_at
	FROM availability_targets %s
	ORDER BY created_at DESC
	LIMIT ? OFFSET ?`, whereClause)

	queryArgs := append(args, limit, offset)
	rows, err := r.db.QueryContext(ctx, query, queryArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var targets []model.AvailabilityTarget
	for rows.Next() {
		var t model.AvailabilityTarget
		err := rows.Scan(
			&t.ID, &t.WorkspaceID, &t.Name, &t.TargetPercentage, &t.MeasurementWindow,
			&t.ServiceHours, &t.CommittedMTBFHours, &t.TargetMTTRMinutes,
			&t.CIID, &t.Status, &t.Description, &t.CreatedAt, &t.UpdatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		targets = append(targets, t)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	return targets, total, nil
}

// CreateOutage inserts a new outage event
func (r *availabilityRepository) CreateOutage(ctx context.Context, outage *model.OutageEvent) error {
	if err := outage.Validate(); err != nil {
		return err
	}
	if outage.ID == "" {
		outage.ID = uuid.New().String()
	}
	now := time.Now().UTC()
	if outage.CreatedAt.IsZero() {
		outage.CreatedAt = now
	}
	outage.UpdatedAt = now

	if outage.EndedAt != nil {
		outage.DurationMinutes = outage.CalculateDuration()
	}

	slaExemptInt := 0
	if outage.IsSLAExempt {
		slaExemptInt = 1
	}

	query := `
	INSERT INTO availability_outages (
		id, workspace_id, target_id, ci_id, title,
		outage_type, severity, started_at, ended_at,
		duration_minutes, root_cause, is_sla_exempt,
		exemption_reason, reported_by, created_at, updated_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := r.db.ExecContext(ctx, query,
		outage.ID, outage.WorkspaceID, outage.TargetID, outage.CIID, outage.Title,
		outage.OutageType, outage.Severity, outage.StartedAt, outage.EndedAt,
		outage.DurationMinutes, outage.RootCause, slaExemptInt,
		outage.ExemptionReason, outage.ReportedBy, outage.CreatedAt, outage.UpdatedAt,
	)
	return err
}

// GetOutageByID retrieves an outage event by ID
func (r *availabilityRepository) GetOutageByID(ctx context.Context, id string) (*model.OutageEvent, error) {
	query := `
	SELECT id, workspace_id, target_id, ci_id, title,
	       outage_type, severity, started_at, ended_at,
	       duration_minutes, root_cause, is_sla_exempt,
	       exemption_reason, reported_by, created_at, updated_at
	FROM availability_outages WHERE id = ?`

	var o model.OutageEvent
	var slaExemptInt int
	var targetID, ciID, rootCause, exemptReason, reportedBy sql.NullString
	var endedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&o.ID, &o.WorkspaceID, &targetID, &ciID, &o.Title,
		&o.OutageType, &o.Severity, &o.StartedAt, &endedAt,
		&o.DurationMinutes, &rootCause, &slaExemptInt,
		&exemptReason, &reportedBy, &o.CreatedAt, &o.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrOutageEventNotFound
		}
		return nil, err
	}
	if targetID.Valid {
		o.TargetID = targetID.String
	}
	if ciID.Valid {
		o.CIID = ciID.String
	}
	if endedAt.Valid {
		t := endedAt.Time
		o.EndedAt = &t
	}
	if rootCause.Valid {
		o.RootCause = rootCause.String
	}
	if exemptReason.Valid {
		o.ExemptionReason = exemptReason.String
	}
	if reportedBy.Valid {
		o.ReportedBy = reportedBy.String
	}
	o.IsSLAExempt = slaExemptInt == 1

	return &o, nil
}

// UpdateOutage updates an existing outage event
func (r *availabilityRepository) UpdateOutage(ctx context.Context, outage *model.OutageEvent) error {
	if err := outage.Validate(); err != nil {
		return err
	}
	outage.UpdatedAt = time.Now().UTC()
	if outage.EndedAt != nil {
		outage.DurationMinutes = outage.CalculateDuration()
	}

	slaExemptInt := 0
	if outage.IsSLAExempt {
		slaExemptInt = 1
	}

	query := `
	UPDATE availability_outages SET
		title = ?, outage_type = ?, severity = ?, started_at = ?, ended_at = ?,
		duration_minutes = ?, root_cause = ?, is_sla_exempt = ?,
		exemption_reason = ?, reported_by = ?, updated_at = ?
	WHERE id = ?`

	res, err := r.db.ExecContext(ctx, query,
		outage.Title, outage.OutageType, outage.Severity, outage.StartedAt, outage.EndedAt,
		outage.DurationMinutes, outage.RootCause, slaExemptInt,
		outage.ExemptionReason, outage.ReportedBy, outage.UpdatedAt,
		outage.ID,
	)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrOutageEventNotFound
	}
	return nil
}

// DeleteOutage deletes an outage by ID
func (r *availabilityRepository) DeleteOutage(ctx context.Context, id string) error {
	query := `DELETE FROM availability_outages WHERE id = ?`
	res, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrOutageEventNotFound
	}
	return nil
}

// ListOutages lists outage events matching criteria
func (r *availabilityRepository) ListOutages(ctx context.Context, filter model.AvailabilityFilterCriteria) ([]model.OutageEvent, int, error) {
	var conditions []string
	var args []interface{}

	if filter.WorkspaceID != "" {
		conditions = append(conditions, "workspace_id = ?")
		args = append(args, filter.WorkspaceID)
	}
	if filter.TargetID != "" {
		conditions = append(conditions, "target_id = ?")
		args = append(args, filter.TargetID)
	}
	if filter.CIID != "" {
		conditions = append(conditions, "ci_id = ?")
		args = append(args, filter.CIID)
	}
	if filter.OutageType != "" {
		conditions = append(conditions, "outage_type = ?")
		args = append(args, filter.OutageType)
	}
	if filter.Severity != "" {
		conditions = append(conditions, "severity = ?")
		args = append(args, filter.Severity)
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM availability_outages %s", whereClause)
	var total int
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	limit := filter.Limit
	if limit <= 0 {
		limit = 50
	}
	offset := filter.Offset
	if offset < 0 {
		offset = 0
	}

	query := fmt.Sprintf(`
	SELECT id, workspace_id, target_id, ci_id, title,
	       outage_type, severity, started_at, ended_at,
	       duration_minutes, root_cause, is_sla_exempt,
	       exemption_reason, reported_by, created_at, updated_at
	FROM availability_outages %s
	ORDER BY started_at DESC
	LIMIT ? OFFSET ?`, whereClause)

	queryArgs := append(args, limit, offset)
	rows, err := r.db.QueryContext(ctx, query, queryArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var outages []model.OutageEvent
	for rows.Next() {
		var o model.OutageEvent
		var slaExemptInt int
		var targetID, ciID, rootCause, exemptReason, reportedBy sql.NullString
		var endedAt sql.NullTime

		err := rows.Scan(
			&o.ID, &o.WorkspaceID, &targetID, &ciID, &o.Title,
			&o.OutageType, &o.Severity, &o.StartedAt, &endedAt,
			&o.DurationMinutes, &rootCause, &slaExemptInt,
			&exemptReason, &reportedBy, &o.CreatedAt, &o.UpdatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		if targetID.Valid {
			o.TargetID = targetID.String
		}
		if ciID.Valid {
			o.CIID = ciID.String
		}
		if endedAt.Valid {
			t := endedAt.Time
			o.EndedAt = &t
		}
		if rootCause.Valid {
			o.RootCause = rootCause.String
		}
		if exemptReason.Valid {
			o.ExemptionReason = exemptReason.String
		}
		if reportedBy.Valid {
			o.ReportedBy = reportedBy.String
		}
		o.IsSLAExempt = slaExemptInt == 1

		outages = append(outages, o)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	return outages, total, nil
}

// SaveMeasurement creates or updates an availability measurement record for a target and period
func (r *availabilityRepository) SaveMeasurement(ctx context.Context, m *model.AvailabilityMeasurement) error {
	if err := m.Validate(); err != nil {
		return err
	}
	if m.ID == "" {
		m.ID = uuid.New().String()
	}
	now := time.Now().UTC()
	if m.CreatedAt.IsZero() {
		m.CreatedAt = now
	}
	m.UpdatedAt = now
	if m.CalculatedAt.IsZero() {
		m.CalculatedAt = now
	}

	breachedInt := 0
	if m.IsSLABreached {
		breachedInt = 1
	}

	// Check if already exists for target + period
	existing, err := r.GetMeasurementByPeriod(ctx, m.TargetID, m.Period)
	if err == nil && existing != nil {
		m.ID = existing.ID
		m.CreatedAt = existing.CreatedAt
		updateQuery := `
		UPDATE availability_measurements SET
			total_operational_minutes = ?, total_downtime_minutes = ?,
			exempt_downtime_minutes = ?, actual_uptime_percentage = ?,
			mtbf_hours = ?, mttr_minutes = ?, outage_count = ?,
			is_sla_breached = ?, calculated_at = ?, updated_at = ?
		WHERE id = ?`

		_, err = r.db.ExecContext(ctx, updateQuery,
			m.TotalOperationalMinutes, m.TotalDowntimeMinutes,
			m.ExemptDowntimeMinutes, m.ActualUptimePercentage,
			m.MTBFHours, m.MTTRMinutes, m.OutageCount,
			breachedInt, m.CalculatedAt, m.UpdatedAt,
			m.ID,
		)
		return err
	}

	insertQuery := `
	INSERT INTO availability_measurements (
		id, workspace_id, target_id, period,
		total_operational_minutes, total_downtime_minutes,
		exempt_downtime_minutes, actual_uptime_percentage,
		mtbf_hours, mttr_minutes, outage_count,
		is_sla_breached, calculated_at, created_at, updated_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err = r.db.ExecContext(ctx, insertQuery,
		m.ID, m.WorkspaceID, m.TargetID, m.Period,
		m.TotalOperationalMinutes, m.TotalDowntimeMinutes,
		m.ExemptDowntimeMinutes, m.ActualUptimePercentage,
		m.MTBFHours, m.MTTRMinutes, m.OutageCount,
		breachedInt, m.CalculatedAt, m.CreatedAt, m.UpdatedAt,
	)
	return err
}

// GetMeasurementByID retrieves an availability measurement by its ID
func (r *availabilityRepository) GetMeasurementByID(ctx context.Context, id string) (*model.AvailabilityMeasurement, error) {
	query := `
	SELECT id, workspace_id, target_id, period,
	       total_operational_minutes, total_downtime_minutes,
	       exempt_downtime_minutes, actual_uptime_percentage,
	       mtbf_hours, mttr_minutes, outage_count,
	       is_sla_breached, calculated_at, created_at, updated_at
	FROM availability_measurements WHERE id = ?`

	var m model.AvailabilityMeasurement
	var breachedInt int

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&m.ID, &m.WorkspaceID, &m.TargetID, &m.Period,
		&m.TotalOperationalMinutes, &m.TotalDowntimeMinutes,
		&m.ExemptDowntimeMinutes, &m.ActualUptimePercentage,
		&m.MTBFHours, &m.MTTRMinutes, &m.OutageCount,
		&breachedInt, &m.CalculatedAt, &m.CreatedAt, &m.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrAvailabilityMeasurementNotFound
		}
		return nil, err
	}
	m.IsSLABreached = breachedInt == 1
	return &m, nil
}

// GetMeasurementByPeriod retrieves an availability measurement for a specific target and period string
func (r *availabilityRepository) GetMeasurementByPeriod(ctx context.Context, targetID string, period string) (*model.AvailabilityMeasurement, error) {
	query := `
	SELECT id, workspace_id, target_id, period,
	       total_operational_minutes, total_downtime_minutes,
	       exempt_downtime_minutes, actual_uptime_percentage,
	       mtbf_hours, mttr_minutes, outage_count,
	       is_sla_breached, calculated_at, created_at, updated_at
	FROM availability_measurements WHERE target_id = ? AND period = ?`

	var m model.AvailabilityMeasurement
	var breachedInt int

	err := r.db.QueryRowContext(ctx, query, targetID, period).Scan(
		&m.ID, &m.WorkspaceID, &m.TargetID, &m.Period,
		&m.TotalOperationalMinutes, &m.TotalDowntimeMinutes,
		&m.ExemptDowntimeMinutes, &m.ActualUptimePercentage,
		&m.MTBFHours, &m.MTTRMinutes, &m.OutageCount,
		&breachedInt, &m.CalculatedAt, &m.CreatedAt, &m.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrAvailabilityMeasurementNotFound
		}
		return nil, err
	}
	m.IsSLABreached = breachedInt == 1
	return &m, nil
}

// ListMeasurements returns a filtered list of availability measurements
func (r *availabilityRepository) ListMeasurements(ctx context.Context, filter model.AvailabilityFilterCriteria) ([]model.AvailabilityMeasurement, int, error) {
	var conditions []string
	var args []interface{}

	if filter.WorkspaceID != "" {
		conditions = append(conditions, "workspace_id = ?")
		args = append(args, filter.WorkspaceID)
	}
	if filter.TargetID != "" {
		conditions = append(conditions, "target_id = ?")
		args = append(args, filter.TargetID)
	}
	if filter.Period != "" {
		conditions = append(conditions, "period = ?")
		args = append(args, filter.Period)
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM availability_measurements %s", whereClause)
	var total int
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	limit := filter.Limit
	if limit <= 0 {
		limit = 50
	}
	offset := filter.Offset
	if offset < 0 {
		offset = 0
	}

	query := fmt.Sprintf(`
	SELECT id, workspace_id, target_id, period,
	       total_operational_minutes, total_downtime_minutes,
	       exempt_downtime_minutes, actual_uptime_percentage,
	       mtbf_hours, mttr_minutes, outage_count,
	       is_sla_breached, calculated_at, created_at, updated_at
	FROM availability_measurements %s
	ORDER BY calculated_at DESC
	LIMIT ? OFFSET ?`, whereClause)

	queryArgs := append(args, limit, offset)
	rows, err := r.db.QueryContext(ctx, query, queryArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var measurements []model.AvailabilityMeasurement
	for rows.Next() {
		var m model.AvailabilityMeasurement
		var breachedInt int
		err := rows.Scan(
			&m.ID, &m.WorkspaceID, &m.TargetID, &m.Period,
			&m.TotalOperationalMinutes, &m.TotalDowntimeMinutes,
			&m.ExemptDowntimeMinutes, &m.ActualUptimePercentage,
			&m.MTBFHours, &m.MTTRMinutes, &m.OutageCount,
			&breachedInt, &m.CalculatedAt, &m.CreatedAt, &m.UpdatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		m.IsSLABreached = breachedInt == 1
		measurements = append(measurements, m)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	return measurements, total, nil
}

func (r *availabilityRepository) CreateSPOF(ctx context.Context, spof *model.SPOFRecord) error {
	return errors.New("not implemented")
}

func (r *availabilityRepository) GetSPOFByID(ctx context.Context, id string) (*model.SPOFRecord, error) {
	return nil, errors.New("not implemented")
}

func (r *availabilityRepository) UpdateSPOF(ctx context.Context, spof *model.SPOFRecord) error {
	return errors.New("not implemented")
}

func (r *availabilityRepository) DeleteSPOF(ctx context.Context, id string) error {
	return errors.New("not implemented")
}

func (r *availabilityRepository) ListSPOFs(ctx context.Context, filter model.AvailabilityFilterCriteria) ([]model.SPOFRecord, int, error) {
	return nil, 0, errors.New("not implemented")
}

func (r *availabilityRepository) CreateTest(ctx context.Context, test *model.AvailabilityTest) error {
	return errors.New("not implemented")
}

func (r *availabilityRepository) GetTestByID(ctx context.Context, id string) (*model.AvailabilityTest, error) {
	return nil, errors.New("not implemented")
}

func (r *availabilityRepository) UpdateTest(ctx context.Context, test *model.AvailabilityTest) error {
	return errors.New("not implemented")
}

func (r *availabilityRepository) DeleteTest(ctx context.Context, id string) error {
	return errors.New("not implemented")
}

func (r *availabilityRepository) ListTests(ctx context.Context, filter model.AvailabilityFilterCriteria) ([]model.AvailabilityTest, int, error) {
	return nil, 0, errors.New("not implemented")
}

func (r *availabilityRepository) GetAvailabilityAnalytics(ctx context.Context, workspaceID string) (*model.AvailabilityAnalytics, error) {
	return nil, errors.New("not implemented")
}
