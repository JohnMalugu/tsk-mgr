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

// CreateSPOF logs a new Single Point of Failure
func (r *availabilityRepository) CreateSPOF(ctx context.Context, spof *model.SPOFRecord) error {
	if err := spof.Validate(); err != nil {
		return err
	}
	if spof.ID == "" {
		spof.ID = uuid.New().String()
	}
	now := time.Now().UTC()
	if spof.CreatedAt.IsZero() {
		spof.CreatedAt = now
	}
	spof.UpdatedAt = now

	query := `
	INSERT INTO availability_spofs (
		id, workspace_id, component_name, ci_id,
		risk_level, redundancy_mechanism, mitigation_plan,
		status, review_date, owner_id, created_at, updated_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := r.db.ExecContext(ctx, query,
		spof.ID, spof.WorkspaceID, spof.ComponentName, spof.CIID,
		spof.RiskLevel, spof.RedundancyMechanism, spof.MitigationPlan,
		spof.Status, spof.ReviewDate, spof.OwnerID, spof.CreatedAt, spof.UpdatedAt,
	)
	return err
}

// GetSPOFByID retrieves a SPOF record by ID
func (r *availabilityRepository) GetSPOFByID(ctx context.Context, id string) (*model.SPOFRecord, error) {
	query := `
	SELECT id, workspace_id, component_name, ci_id,
	       risk_level, redundancy_mechanism, mitigation_plan,
	       status, review_date, owner_id, created_at, updated_at
	FROM availability_spofs WHERE id = ?`

	var s model.SPOFRecord
	var ciID, mitigation, ownerID sql.NullString

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&s.ID, &s.WorkspaceID, &s.ComponentName, &ciID,
		&s.RiskLevel, &s.RedundancyMechanism, &mitigation,
		&s.Status, &s.ReviewDate, &ownerID, &s.CreatedAt, &s.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrSPOFRecordNotFound
		}
		return nil, err
	}
	if ciID.Valid {
		s.CIID = ciID.String
	}
	if mitigation.Valid {
		s.MitigationPlan = mitigation.String
	}
	if ownerID.Valid {
		s.OwnerID = ownerID.String
	}
	return &s, nil
}

// UpdateSPOF modifies an existing SPOF record
func (r *availabilityRepository) UpdateSPOF(ctx context.Context, spof *model.SPOFRecord) error {
	if err := spof.Validate(); err != nil {
		return err
	}
	spof.UpdatedAt = time.Now().UTC()

	query := `
	UPDATE availability_spofs SET
		component_name = ?, ci_id = ?, risk_level = ?,
		redundancy_mechanism = ?, mitigation_plan = ?, status = ?,
		review_date = ?, owner_id = ?, updated_at = ?
	WHERE id = ?`

	res, err := r.db.ExecContext(ctx, query,
		spof.ComponentName, spof.CIID, spof.RiskLevel,
		spof.RedundancyMechanism, spof.MitigationPlan, spof.Status,
		spof.ReviewDate, spof.OwnerID, spof.UpdatedAt,
		spof.ID,
	)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrSPOFRecordNotFound
	}
	return nil
}

// DeleteSPOF removes a SPOF record
func (r *availabilityRepository) DeleteSPOF(ctx context.Context, id string) error {
	query := `DELETE FROM availability_spofs WHERE id = ?`
	res, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrSPOFRecordNotFound
	}
	return nil
}

// ListSPOFs lists SPOF records matching filter criteria
func (r *availabilityRepository) ListSPOFs(ctx context.Context, filter model.AvailabilityFilterCriteria) ([]model.SPOFRecord, int, error) {
	var conditions []string
	var args []interface{}

	if filter.WorkspaceID != "" {
		conditions = append(conditions, "workspace_id = ?")
		args = append(args, filter.WorkspaceID)
	}
	if filter.RiskLevel != "" {
		conditions = append(conditions, "risk_level = ?")
		args = append(args, filter.RiskLevel)
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

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM availability_spofs %s", whereClause)
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
	SELECT id, workspace_id, component_name, ci_id,
	       risk_level, redundancy_mechanism, mitigation_plan,
	       status, review_date, owner_id, created_at, updated_at
	FROM availability_spofs %s
	ORDER BY created_at DESC
	LIMIT ? OFFSET ?`, whereClause)

	queryArgs := append(args, limit, offset)
	rows, err := r.db.QueryContext(ctx, query, queryArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var spofs []model.SPOFRecord
	for rows.Next() {
		var s model.SPOFRecord
		var ciID, mitigation, ownerID sql.NullString
		err := rows.Scan(
			&s.ID, &s.WorkspaceID, &s.ComponentName, &ciID,
			&s.RiskLevel, &s.RedundancyMechanism, &mitigation,
			&s.Status, &s.ReviewDate, &ownerID, &s.CreatedAt, &s.UpdatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		if ciID.Valid {
			s.CIID = ciID.String
		}
		if mitigation.Valid {
			s.MitigationPlan = mitigation.String
		}
		if ownerID.Valid {
			s.OwnerID = ownerID.String
		}
		spofs = append(spofs, s)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	return spofs, total, nil
}

// CreateTest logs a new availability failover test schedule
func (r *availabilityRepository) CreateTest(ctx context.Context, test *model.AvailabilityTest) error {
	if err := test.Validate(); err != nil {
		return err
	}
	if test.ID == "" {
		test.ID = uuid.New().String()
	}
	now := time.Now().UTC()
	if test.CreatedAt.IsZero() {
		test.CreatedAt = now
	}
	test.UpdatedAt = now

	query := `
	INSERT INTO availability_tests (
		id, workspace_id, target_id, name, test_type,
		frequency, last_tested_at, next_test_due,
		outcome, findings, conducted_by, created_at, updated_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := r.db.ExecContext(ctx, query,
		test.ID, test.WorkspaceID, test.TargetID, test.Name, test.TestType,
		test.Frequency, test.LastTestedAt, test.NextTestDue,
		test.Outcome, test.Findings, test.ConductedBy, test.CreatedAt, test.UpdatedAt,
	)
	return err
}

// GetTestByID retrieves an availability test by ID
func (r *availabilityRepository) GetTestByID(ctx context.Context, id string) (*model.AvailabilityTest, error) {
	query := `
	SELECT id, workspace_id, target_id, name, test_type,
	       frequency, last_tested_at, next_test_due,
	       outcome, findings, conducted_by, created_at, updated_at
	FROM availability_tests WHERE id = ?`

	var t model.AvailabilityTest
	var targetID, findings, conductedBy sql.NullString
	var lastTested sql.NullTime

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&t.ID, &t.WorkspaceID, &targetID, &t.Name, &t.TestType,
		&t.Frequency, &lastTested, &t.NextTestDue,
		&t.Outcome, &findings, &conductedBy, &t.CreatedAt, &t.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrAvailabilityTestNotFound
		}
		return nil, err
	}
	if targetID.Valid {
		t.TargetID = targetID.String
	}
	if lastTested.Valid {
		val := lastTested.Time
		t.LastTestedAt = &val
	}
	if findings.Valid {
		t.Findings = findings.String
	}
	if conductedBy.Valid {
		t.ConductedBy = conductedBy.String
	}
	return &t, nil
}

// UpdateTest updates an availability test
func (r *availabilityRepository) UpdateTest(ctx context.Context, test *model.AvailabilityTest) error {
	if err := test.Validate(); err != nil {
		return err
	}
	test.UpdatedAt = time.Now().UTC()

	query := `
	UPDATE availability_tests SET
		name = ?, test_type = ?, frequency = ?, last_tested_at = ?,
		next_test_due = ?, outcome = ?, findings = ?, conducted_by = ?, updated_at = ?
	WHERE id = ?`

	res, err := r.db.ExecContext(ctx, query,
		test.Name, test.TestType, test.Frequency, test.LastTestedAt,
		test.NextTestDue, test.Outcome, test.Findings, test.ConductedBy, test.UpdatedAt,
		test.ID,
	)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrAvailabilityTestNotFound
	}
	return nil
}

// DeleteTest removes an availability test schedule
func (r *availabilityRepository) DeleteTest(ctx context.Context, id string) error {
	query := `DELETE FROM availability_tests WHERE id = ?`
	res, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrAvailabilityTestNotFound
	}
	return nil
}

// ListTests lists availability tests matching filter criteria
func (r *availabilityRepository) ListTests(ctx context.Context, filter model.AvailabilityFilterCriteria) ([]model.AvailabilityTest, int, error) {
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
	if filter.Outcome != "" {
		conditions = append(conditions, "outcome = ?")
		args = append(args, filter.Outcome)
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM availability_tests %s", whereClause)
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
	SELECT id, workspace_id, target_id, name, test_type,
	       frequency, last_tested_at, next_test_due,
	       outcome, findings, conducted_by, created_at, updated_at
	FROM availability_tests %s
	ORDER BY next_test_due ASC
	LIMIT ? OFFSET ?`, whereClause)

	queryArgs := append(args, limit, offset)
	rows, err := r.db.QueryContext(ctx, query, queryArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var tests []model.AvailabilityTest
	for rows.Next() {
		var t model.AvailabilityTest
		var targetID, findings, conductedBy sql.NullString
		var lastTested sql.NullTime
		err := rows.Scan(
			&t.ID, &t.WorkspaceID, &targetID, &t.Name, &t.TestType,
			&t.Frequency, &lastTested, &t.NextTestDue,
			&t.Outcome, &findings, &conductedBy, &t.CreatedAt, &t.UpdatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		if targetID.Valid {
			t.TargetID = targetID.String
		}
		if lastTested.Valid {
			val := lastTested.Time
			t.LastTestedAt = &val
		}
		if findings.Valid {
			t.Findings = findings.String
		}
		if conductedBy.Valid {
			t.ConductedBy = conductedBy.String
		}
		tests = append(tests, t)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	return tests, total, nil
}

// GetAvailabilityAnalytics aggregates resilience, SLA compliance, MTBF/MTTR, and SPOFs into an executive summary
func (r *availabilityRepository) GetAvailabilityAnalytics(ctx context.Context, workspaceID string) (*model.AvailabilityAnalytics, error) {
	analytics := &model.AvailabilityAnalytics{
		OverallAvailabilityPercentage: 100.0,
		SLAComplianceRate:             100.0,
		TopOutagesByCI:                []model.TopOutageSummary{},
	}

	// 1. Targets count
	_ = r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM availability_targets WHERE workspace_id = ?", workspaceID).Scan(&analytics.TotalTargets)
	_ = r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM availability_targets WHERE workspace_id = ? AND status = 'active'", workspaceID).Scan(&analytics.ActiveTargets)

	// 2. Measurements aggregation
	var totalMeasurements int
	_ = r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM availability_measurements WHERE workspace_id = ?", workspaceID).Scan(&totalMeasurements)

	if totalMeasurements > 0 {
		var avgUptime, avgMTBF, avgMTTR sql.NullFloat64
		_ = r.db.QueryRowContext(ctx, `
			SELECT AVG(actual_uptime_percentage), AVG(NULLIF(mtbf_hours, 0)), AVG(NULLIF(mttr_minutes, 0))
			FROM availability_measurements WHERE workspace_id = ?`, workspaceID).Scan(&avgUptime, &avgMTBF, &avgMTTR)

		if avgUptime.Valid {
			analytics.OverallAvailabilityPercentage = avgUptime.Float64
		}
		if avgMTBF.Valid {
			analytics.AverageMTBFHours = avgMTBF.Float64
		}
		if avgMTTR.Valid {
			analytics.AverageMTTRMinutes = avgMTTR.Float64
		}

		_ = r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM availability_measurements WHERE workspace_id = ? AND is_sla_breached = 1", workspaceID).Scan(&analytics.SLABreachCount)
		compliant := totalMeasurements - analytics.SLABreachCount
		analytics.SLAComplianceRate = (float64(compliant) / float64(totalMeasurements)) * 100.0
	}

	// 3. Outages
	_ = r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM availability_outages WHERE workspace_id = ?", workspaceID).Scan(&analytics.TotalOutagesRecorded)
	_ = r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM availability_outages WHERE workspace_id = ? AND ended_at IS NULL", workspaceID).Scan(&analytics.ActiveOutages)

	// 4. SPOFs
	_ = r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM availability_spofs WHERE workspace_id = ?", workspaceID).Scan(&analytics.TotalSPOFCount)
	_ = r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM availability_spofs WHERE workspace_id = ? AND risk_level = 'critical' AND status = 'open'", workspaceID).Scan(&analytics.CriticalSPOFCount)

	// 5. Tests
	now := time.Now().UTC()
	future30Days := now.AddDate(0, 1, 0)
	_ = r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM availability_tests WHERE workspace_id = ? AND next_test_due <= ?", workspaceID, future30Days).Scan(&analytics.UpcomingTestsCount)

	// 6. Top outages by CI
	rows, err := r.db.QueryContext(ctx, `
		SELECT ci_id, COUNT(*), COALESCE(SUM(duration_minutes), 0)
		FROM availability_outages
		WHERE workspace_id = ? AND ci_id IS NOT NULL AND ci_id != ''
		GROUP BY ci_id
		ORDER BY SUM(duration_minutes) DESC
		LIMIT 5`, workspaceID)
	if err == nil && rows != nil {
		defer rows.Close()
		for rows.Next() {
			var top model.TopOutageSummary
			if scanErr := rows.Scan(&top.CIID, &top.OutageCount, &top.TotalDowntimeMinutes); scanErr == nil {
				analytics.TopOutagesByCI = append(analytics.TopOutagesByCI, top)
			}
		}
	}

	analytics.CalculateHealthIndex()
	return analytics, nil
}
