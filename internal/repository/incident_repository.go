package repository

import (
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/JohnMalugu/tsk-mgr-api/internal/model"
)

var (
	ErrIncidentNotFound = errors.New("incident not found")
)

// IncidentRepository manages database operations for incidents, timelines, and remediation tasks
type IncidentRepository interface {
	CreateIncident(incident *model.Incident) (int64, error)
	GetIncidentByID(id int64) (*model.Incident, error)
	GetIncidentByNumber(workspaceID int64, incidentNumber int) (*model.Incident, error)
	ListIncidents(filter model.IncidentFilter) ([]*model.Incident, error)
	UpdateIncident(incident *model.Incident) error
	DeleteIncident(id int64) error
	GetNextIncidentNumber(workspaceID int64) (int, error)

	CreateTimelineEvent(event *model.IncidentTimeline) (int64, error)
	ListTimelineEvents(incidentID int64) ([]model.IncidentTimeline, error)

	LinkTask(link *model.IncidentTaskLink) (int64, error)
	UnlinkTask(incidentID int64, taskID int) error
	ListIncidentTasks(incidentID int64) ([]model.IncidentTaskLink, error)

	GetIncidentMetrics(workspaceID int64) (*model.IncidentMetrics, error)
}

type incidentRepositoryImpl struct {
	db *sql.DB
}

// NewIncidentRepository creates a new instance of IncidentRepository
func NewIncidentRepository(db *sql.DB) IncidentRepository {
	return &incidentRepositoryImpl{db: db}
}

func (r *incidentRepositoryImpl) GetNextIncidentNumber(workspaceID int64) (int, error) {
	query := `SELECT COALESCE(MAX(incident_number), 0) + 1 FROM incidents WHERE workspace_id = ?`
	var nextNum int
	err := r.db.QueryRow(query, workspaceID).Scan(&nextNum)
	if err != nil {
		return 1, err
	}
	return nextNum, nil
}

func (r *incidentRepositoryImpl) CreateIncident(incident *model.Incident) (int64, error) {
	if incident.IncidentNumber <= 0 {
		nextNum, err := r.GetNextIncidentNumber(incident.WorkspaceID)
		if err != nil {
			return 0, err
		}
		incident.IncidentNumber = nextNum
	}

	servicesJSON, err := json.Marshal(incident.ImpactedServices)
	if err != nil {
		servicesJSON = []byte("[]")
	}

	query := `
	INSERT INTO incidents (
		workspace_id, project_id, incident_number, title, summary,
		severity, status, commander_id, scribe_id, reported_by_id,
		postmortem_status, postmortem_doc_id, root_cause, impacted_services,
		started_at, acknowledged_at, mitigated_at, resolved_at, closed_at,
		created_at, updated_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	now := time.Now().UTC()
	if incident.CreatedAt.IsZero() {
		incident.CreatedAt = now
	}
	if incident.UpdatedAt.IsZero() {
		incident.UpdatedAt = now
	}
	if incident.StartedAt.IsZero() {
		incident.StartedAt = now
	}

	res, err := r.db.Exec(query,
		incident.WorkspaceID, incident.ProjectID, incident.IncidentNumber, incident.Title, incident.Summary,
		incident.Severity, incident.Status, incident.CommanderID, incident.ScribeID, incident.ReportedByID,
		incident.PostmortemStatus, incident.PostmortemDocID, incident.RootCause, string(servicesJSON),
		incident.StartedAt, incident.AcknowledgedAt, incident.MitigatedAt, incident.ResolvedAt, incident.ClosedAt,
		incident.CreatedAt, incident.UpdatedAt,
	)
	if err != nil {
		return 0, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	incident.ID = id
	return id, nil
}

func scanIncident(scanner interface{ Scan(dest ...interface{}) error }) (*model.Incident, error) {
	var inc model.Incident
	var servicesJSON string
	var projectID sql.NullInt32
	var commanderID, scribeID sql.NullString
	var postmortemDocID sql.NullInt64
	var acknowledgedAt, mitigatedAt, resolvedAt, closedAt sql.NullTime
	var commanderUsername, scribeUsername, reporterUsername, projectName sql.NullString

	err := scanner.Scan(
		&inc.ID, &inc.WorkspaceID, &projectID, &inc.IncidentNumber, &inc.Title, &inc.Summary,
		&inc.Severity, &inc.Status, &commanderID, &scribeID, &inc.ReportedByID,
		&inc.PostmortemStatus, &postmortemDocID, &inc.RootCause, &servicesJSON,
		&inc.StartedAt, &acknowledgedAt, &mitigatedAt, &resolvedAt, &closedAt,
		&inc.CreatedAt, &inc.UpdatedAt,
		&commanderUsername, &scribeUsername, &reporterUsername, &projectName,
	)
	if err != nil {
		return nil, err
	}

	if projectID.Valid {
		pid := int(projectID.Int32)
		inc.ProjectID = &pid
	}
	if commanderID.Valid {
		cid := commanderID.String
		inc.CommanderID = &cid
	}
	if scribeID.Valid {
		sid := scribeID.String
		inc.ScribeID = &sid
	}
	if postmortemDocID.Valid {
		inc.PostmortemDocID = &postmortemDocID.Int64
	}
	if acknowledgedAt.Valid {
		t := acknowledgedAt.Time
		inc.AcknowledgedAt = &t
	}
	if mitigatedAt.Valid {
		t := mitigatedAt.Time
		inc.MitigatedAt = &t
	}
	if resolvedAt.Valid {
		t := resolvedAt.Time
		inc.ResolvedAt = &t
	}
	if closedAt.Valid {
		t := closedAt.Time
		inc.ClosedAt = &t
	}

	if commanderUsername.Valid {
		inc.CommanderUsername = commanderUsername.String
	}
	if scribeUsername.Valid {
		inc.ScribeUsername = scribeUsername.String
	}
	if reporterUsername.Valid {
		inc.ReporterUsername = reporterUsername.String
	}
	if projectName.Valid {
		inc.ProjectName = projectName.String
	}

	if servicesJSON != "" {
		_ = json.Unmarshal([]byte(servicesJSON), &inc.ImpactedServices)
	}
	if inc.ImpactedServices == nil {
		inc.ImpactedServices = []string{}
	}

	// Calculate operational metrics
	if inc.AcknowledgedAt != nil && !inc.AcknowledgedAt.IsZero() {
		mins := int(inc.AcknowledgedAt.Sub(inc.StartedAt).Minutes())
		if mins >= 0 {
			inc.TimeToAcknowledgeMinutes = &mins
		}
	}
	if inc.ResolvedAt != nil && !inc.ResolvedAt.IsZero() {
		mins := int(inc.ResolvedAt.Sub(inc.StartedAt).Minutes())
		if mins >= 0 {
			inc.TimeToResolveMinutes = &mins
			inc.DurationMinutes = &mins
		}
	} else {
		mins := int(time.Since(inc.StartedAt).Minutes())
		if mins >= 0 {
			inc.DurationMinutes = &mins
		}
	}

	return &inc, nil
}


const incidentSelectQuery = `
	SELECT 
		i.id, i.workspace_id, i.project_id, i.incident_number, i.title, i.summary,
		i.severity, i.status, i.commander_id, i.scribe_id, i.reported_by_id,
		i.postmortem_status, i.postmortem_doc_id, i.root_cause, i.impacted_services,
		i.started_at, i.acknowledged_at, i.mitigated_at, i.resolved_at, i.closed_at,
		i.created_at, i.updated_at,
		u_cmd.username AS commander_username,
		u_scb.username AS scribe_username,
		u_rep.username AS reporter_username,
		p.name AS project_name
	FROM incidents i
	LEFT JOIN users u_cmd ON i.commander_id = u_cmd.id
	LEFT JOIN users u_scb ON i.scribe_id = u_scb.id
	LEFT JOIN users u_rep ON i.reported_by_id = u_rep.id
	LEFT JOIN projects p ON i.project_id = p.id
`

func (r *incidentRepositoryImpl) GetIncidentByID(id int64) (*model.Incident, error) {
	query := incidentSelectQuery + " WHERE i.id = ?"
	row := r.db.QueryRow(query, id)
	inc, err := scanIncident(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrIncidentNotFound
		}
		return nil, err
	}
	return inc, nil
}

func (r *incidentRepositoryImpl) GetIncidentByNumber(workspaceID int64, incidentNumber int) (*model.Incident, error) {
	query := incidentSelectQuery + " WHERE i.workspace_id = ? AND i.incident_number = ?"
	row := r.db.QueryRow(query, workspaceID, incidentNumber)
	inc, err := scanIncident(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrIncidentNotFound
		}
		return nil, err
	}
	return inc, nil
}

func (r *incidentRepositoryImpl) ListIncidents(filter model.IncidentFilter) ([]*model.Incident, error) {
	query := incidentSelectQuery + " WHERE i.workspace_id = ?"
	args := []interface{}{filter.WorkspaceID}

	if filter.ProjectID != nil {
		query += " AND i.project_id = ?"
		args = append(args, *filter.ProjectID)
	}
	if filter.Severity != "" {
		query += " AND i.severity = ?"
		args = append(args, filter.Severity)
	}
	if filter.Status != "" {
		query += " AND i.status = ?"
		args = append(args, filter.Status)
	}
	if filter.CommanderID != "" {
		query += " AND i.commander_id = ?"
		args = append(args, filter.CommanderID)
	}
	if filter.ReportedByID != "" {
		query += " AND i.reported_by_id = ?"
		args = append(args, filter.ReportedByID)
	}
	if filter.Search != "" {
		query += " AND (i.title LIKE ? OR i.summary LIKE ? OR i.root_cause LIKE ?)"
		pat := "%" + filter.Search + "%"
		args = append(args, pat, pat, pat)
	}
	if filter.From != nil {
		query += " AND i.started_at >= ?"
		args = append(args, *filter.From)
	}
	if filter.To != nil {
		query += " AND i.started_at <= ?"
		args = append(args, *filter.To)
	}

	query += " ORDER BY i.started_at DESC, i.id DESC"

	limit := filter.Limit
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	query += " LIMIT ?"
	args = append(args, limit)

	if filter.Offset > 0 {
		query += " OFFSET ?"
		args = append(args, filter.Offset)
	}

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var incidents []*model.Incident
	for rows.Next() {
		inc, err := scanIncident(rows)
		if err != nil {
			return nil, err
		}
		incidents = append(incidents, inc)
	}
	if incidents == nil {
		incidents = []*model.Incident{}
	}
	return incidents, rows.Err()
}

func (r *incidentRepositoryImpl) UpdateIncident(incident *model.Incident) error {
	servicesJSON, err := json.Marshal(incident.ImpactedServices)
	if err != nil {
		servicesJSON = []byte("[]")
	}

	incident.UpdatedAt = time.Now().UTC()

	query := `
	UPDATE incidents SET
		project_id = ?,
		title = ?,
		summary = ?,
		severity = ?,
		status = ?,
		commander_id = ?,
		scribe_id = ?,
		postmortem_status = ?,
		postmortem_doc_id = ?,
		root_cause = ?,
		impacted_services = ?,
		started_at = ?,
		acknowledged_at = ?,
		mitigated_at = ?,
		resolved_at = ?,
		closed_at = ?,
		updated_at = ?
	WHERE id = ?`

	res, err := r.db.Exec(query,
		incident.ProjectID,
		incident.Title,
		incident.Summary,
		incident.Severity,
		incident.Status,
		incident.CommanderID,
		incident.ScribeID,
		incident.PostmortemStatus,
		incident.PostmortemDocID,
		incident.RootCause,
		string(servicesJSON),
		incident.StartedAt,
		incident.AcknowledgedAt,
		incident.MitigatedAt,
		incident.ResolvedAt,
		incident.ClosedAt,
		incident.UpdatedAt,
		incident.ID,
	)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrIncidentNotFound
	}
	return nil
}

func (r *incidentRepositoryImpl) DeleteIncident(id int64) error {
	// Cascade cleanup of child tables
	_, _ = r.db.Exec(`DELETE FROM incident_timelines WHERE incident_id = ?`, id)
	_, _ = r.db.Exec(`DELETE FROM incident_task_links WHERE incident_id = ?`, id)

	res, err := r.db.Exec(`DELETE FROM incidents WHERE id = ?`, id)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrIncidentNotFound
	}
	return nil
}

func (r *incidentRepositoryImpl) CreateTimelineEvent(event *model.IncidentTimeline) (int64, error) {
	if event.CreatedAt.IsZero() {
		event.CreatedAt = time.Now().UTC()
	}

	query := `
	INSERT INTO incident_timelines (incident_id, author_id, event_type, message, metadata, created_at)
	VALUES (?, ?, ?, ?, ?, ?)`

	res, err := r.db.Exec(query,
		event.IncidentID,
		event.AuthorID,
		event.EventType,
		event.Message,
		event.Metadata,
		event.CreatedAt,
	)
	if err != nil {
		return 0, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	event.ID = id
	return id, nil
}

func (r *incidentRepositoryImpl) ListTimelineEvents(incidentID int64) ([]model.IncidentTimeline, error) {
	query := `
	SELECT 
		t.id, t.incident_id, t.author_id, t.event_type, t.message, t.metadata, t.created_at,
		u.username AS author_username
	FROM incident_timelines t
	LEFT JOIN users u ON t.author_id = u.id
	WHERE t.incident_id = ?
	ORDER BY t.created_at ASC, t.id ASC`

	rows, err := r.db.Query(query, incidentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []model.IncidentTimeline
	for rows.Next() {
		var ev model.IncidentTimeline
		var authorUsername sql.NullString

		err := rows.Scan(
			&ev.ID, &ev.IncidentID, &ev.AuthorID, &ev.EventType, &ev.Message, &ev.Metadata, &ev.CreatedAt,
			&authorUsername,
		)
		if err != nil {
			return nil, err
		}
		if authorUsername.Valid {
			ev.AuthorUsername = authorUsername.String
		}
		events = append(events, ev)
	}
	if events == nil {
		events = []model.IncidentTimeline{}
	}
	return events, rows.Err()
}

func (r *incidentRepositoryImpl) LinkTask(link *model.IncidentTaskLink) (int64, error) {
	return 0, nil
}

func (r *incidentRepositoryImpl) UnlinkTask(incidentID int64, taskID int) error {
	return nil
}

func (r *incidentRepositoryImpl) ListIncidentTasks(incidentID int64) ([]model.IncidentTaskLink, error) {
	return nil, nil
}

func (r *incidentRepositoryImpl) GetIncidentMetrics(workspaceID int64) (*model.IncidentMetrics, error) {
	return nil, nil
}
