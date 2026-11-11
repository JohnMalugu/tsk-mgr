package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/JohnMalugu/tsk-mgr-api/internal/model"
)

var (
	ErrProblemNotFound     = errors.New("problem not found")
	ErrProblemExists       = errors.New("problem already exists")
	ErrProblemLinkExists   = errors.New("problem link already exists")
	ErrProblemLinkNotFound = errors.New("problem link not found")
	ErrRCAStepNotFound     = errors.New("root cause step not found")
)

// ProblemRepository defines persistent storage operations for ITIL Problem Management & KEDB
type ProblemRepository interface {
	// Sequential code generation
	GetNextProblemCode(workspaceID int64) (string, error)

	// Problem CRUD
	CreateProblem(p *model.Problem) (int64, error)
	GetProblemByID(id int64) (*model.Problem, error)
	GetProblemByCode(workspaceID int64, code string) (*model.Problem, error)
	ListProblems(filter model.ProblemFilter) ([]model.Problem, int, error)
	UpdateProblem(p *model.Problem) error
	DeleteProblem(id int64) error

	// Incident Linkage
	LinkIncident(link *model.ProblemIncidentLink) (int64, error)
	UnlinkIncident(problemID, incidentID int64) error
	ListProblemIncidents(problemID int64) ([]model.ProblemIncidentLink, error)
	GetProblemsForIncident(incidentID int64) ([]model.Problem, error)

	// CI Linkage
	LinkCI(link *model.ProblemCILink) (int64, error)
	UnlinkCI(id int64) error
	ListProblemCIs(problemID int64) ([]model.ProblemCILink, error)
	GetProblemsForCI(ciID int64) ([]model.Problem, error)

	// Change Linkage
	LinkChange(link *model.ProblemChangeLink) (int64, error)
	UnlinkChange(id int64) error
	ListProblemChanges(problemID int64) ([]model.ProblemChangeLink, error)
	GetProblemsForChange(changeID int64) ([]model.Problem, error)

	// Task Linkage
	LinkTask(link *model.ProblemTaskLink) (int64, error)
	UnlinkTask(id int64) error
	ListProblemTasks(problemID int64) ([]model.ProblemTaskLink, error)
	GetProblemsForTask(taskID int) ([]model.Problem, error)

	// 5-Whys RCA Steps
	AddRCAStep(step *model.ProblemRCAStep) (int64, error)
	GetRCAStep(id int64) (*model.ProblemRCAStep, error)
	UpdateRCAStep(step *model.ProblemRCAStep) error
	DeleteRCAStep(id int64) error
	ListRCASteps(problemID int64) ([]model.ProblemRCAStep, error)
	GetNextRCAStepNumber(problemID int64) (int, error)

	// Known Error Database (KEDB)
	ListKnownErrors(filter model.KnownErrorFilter) ([]model.Problem, int, error)

	// Metrics
	GetProblemMetrics(workspaceID int64) (*model.ProblemMetrics, error)
}

type problemRepositoryImpl struct {
	db *sql.DB
}

// NewProblemRepository creates a new SQLite-backed ProblemRepository
func NewProblemRepository(db *sql.DB) ProblemRepository {
	return &problemRepositoryImpl{db: db}
}

func (r *problemRepositoryImpl) GetNextProblemCode(workspaceID int64) (string, error) {
	query := `SELECT problem_code FROM problems WHERE workspace_id = ? ORDER BY id DESC LIMIT 1`
	var lastCode string
	err := r.db.QueryRow(query, workspaceID).Scan(&lastCode)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "PRB-1", nil
		}
		return "PRB-1", err
	}

	parts := strings.Split(lastCode, "-")
	if len(parts) != 2 {
		return "PRB-1", nil
	}
	nextNum, err := strconv.Atoi(parts[1])
	if err != nil {
		return "PRB-1", nil
	}
	return fmt.Sprintf("PRB-%d", nextNum+1), nil
}

const baseProblemSelect = `
	SELECT 
		p.id, p.workspace_id, p.project_id, p.problem_code, p.title, p.description,
		p.status, p.category, p.priority, p.urgency, p.impact,
		p.assigned_to, p.assigned_team, p.root_cause_category, p.root_cause_summary,
		p.workaround, p.permanent_solution, p.is_known_error, p.known_error_published_at,
		p.rca_method, p.created_by_id, p.started_at, p.resolved_at, p.closed_at,
		p.created_at, p.updated_at,
		COALESCE(pr.name, '') AS project_name,
		COALESCE(u_assignee.username, u_assignee.email, '') AS assignee_name,
		COALESCE(u_creator.username, u_creator.email, '') AS creator_name,
		(SELECT COUNT(*) FROM problem_incident_links WHERE problem_id = p.id) AS linked_incidents_count,
		(SELECT COUNT(*) FROM problem_ci_links WHERE problem_id = p.id) AS linked_cis_count,
		(SELECT COUNT(*) FROM problem_change_links WHERE problem_id = p.id) AS linked_changes_count,
		(SELECT COUNT(*) FROM problem_task_links WHERE problem_id = p.id) AS linked_tasks_count,
		(SELECT COUNT(*) FROM problem_rca_steps WHERE problem_id = p.id) AS rca_steps_count
	FROM problems p
	LEFT JOIN projects pr ON p.project_id = pr.id
	LEFT JOIN users u_assignee ON p.assigned_to = u_assignee.id
	LEFT JOIN users u_creator ON p.created_by_id = u_creator.id
`

func scanProblem(scanner interface {
	Scan(dest ...interface{}) error
}) (*model.Problem, error) {
	var p model.Problem
	var projectID sql.NullInt64
	var assignedTo, projectName, assigneeName, creatorName sql.NullString
	var isKnownError int
	var knownErrorPublishedAt, startedAt, resolvedAt, closedAt *time.Time
	var incCount, ciCount, chgCount, tskCount, rcaCount sql.NullInt64

	err := scanner.Scan(
		&p.ID, &p.WorkspaceID, &projectID, &p.ProblemCode, &p.Title, &p.Description,
		&p.Status, &p.Category, &p.Priority, &p.Urgency, &p.Impact,
		&assignedTo, &p.AssignedTeam, &p.RootCauseCategory, &p.RootCauseSummary,
		&p.Workaround, &p.PermanentSolution, &isKnownError, &knownErrorPublishedAt,
		&p.RCAMethod, &p.CreatedByID, &startedAt, &resolvedAt, &closedAt,
		&p.CreatedAt, &p.UpdatedAt,
		&projectName, &assigneeName, &creatorName,
		&incCount, &ciCount, &chgCount, &tskCount, &rcaCount,
	)
	if err != nil {
		return nil, err
	}

	if projectID.Valid {
		p.ProjectID = &projectID.Int64
	}
	if assignedTo.Valid && assignedTo.String != "" {
		str := assignedTo.String
		p.AssignedTo = &str
	}
	p.IsKnownError = isKnownError == 1
	p.KnownErrorPublishedAt = knownErrorPublishedAt
	p.StartedAt = startedAt
	p.ResolvedAt = resolvedAt
	p.ClosedAt = closedAt
	if projectName.Valid {
		p.ProjectName = projectName.String
	}
	if assigneeName.Valid {
		p.AssigneeName = assigneeName.String
	}
	if creatorName.Valid {
		p.CreatorName = creatorName.String
	}
	if incCount.Valid {
		p.LinkedIncidentsCount = int(incCount.Int64)
	}
	if ciCount.Valid {
		p.LinkedCIsCount = int(ciCount.Int64)
	}
	if chgCount.Valid {
		p.LinkedChangesCount = int(chgCount.Int64)
	}
	if tskCount.Valid {
		p.LinkedTasksCount = int(tskCount.Int64)
	}
	if rcaCount.Valid {
		p.RCAStepsCount = int(rcaCount.Int64)
	}

	return &p, nil
}

func (r *problemRepositoryImpl) CreateProblem(p *model.Problem) (int64, error) {
	if p.ProblemCode == "" {
		code, err := r.GetNextProblemCode(p.WorkspaceID)
		if err != nil {
			return 0, err
		}
		p.ProblemCode = code
	}
	now := time.Now().UTC()
	if p.CreatedAt.IsZero() {
		p.CreatedAt = now
	}
	if p.UpdatedAt.IsZero() {
		p.UpdatedAt = now
	}
	if p.Status == "" {
		p.Status = model.ProblemStatusLogged
	}
	if p.Category == "" {
		p.Category = model.ProblemCategorySoftware
	}
	if p.Priority == "" {
		p.Priority = model.ProblemPriorityMedium
	}
	if p.Urgency == "" {
		p.Urgency = model.ProblemUrgencyMedium
	}
	if p.Impact == "" {
		p.Impact = model.ProblemImpactMedium
	}

	isKnownInt := 0
	if p.IsKnownError {
		isKnownInt = 1
	}

	query := `
	INSERT INTO problems (
		workspace_id, project_id, problem_code, title, description,
		status, category, priority, urgency, impact,
		assigned_to, assigned_team, root_cause_category, root_cause_summary,
		workaround, permanent_solution, is_known_error, known_error_published_at,
		rca_method, created_by_id, started_at, resolved_at, closed_at,
		created_at, updated_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	res, err := r.db.Exec(query,
		p.WorkspaceID, p.ProjectID, p.ProblemCode, p.Title, p.Description,
		p.Status, p.Category, p.Priority, p.Urgency, p.Impact,
		p.AssignedTo, p.AssignedTeam, p.RootCauseCategory, p.RootCauseSummary,
		p.Workaround, p.PermanentSolution, isKnownInt, p.KnownErrorPublishedAt,
		p.RCAMethod, p.CreatedByID, p.StartedAt, p.ResolvedAt, p.ClosedAt,
		p.CreatedAt, p.UpdatedAt,
	)
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	p.ID = id
	return id, nil
}

func (r *problemRepositoryImpl) GetProblemByID(id int64) (*model.Problem, error) {
	query := baseProblemSelect + ` WHERE p.id = ?`
	row := r.db.QueryRow(query, id)
	p, err := scanProblem(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrProblemNotFound
		}
		return nil, err
	}
	return p, nil
}

func (r *problemRepositoryImpl) GetProblemByCode(workspaceID int64, code string) (*model.Problem, error) {
	query := baseProblemSelect + ` WHERE p.workspace_id = ? AND p.problem_code = ?`
	row := r.db.QueryRow(query, workspaceID, code)
	p, err := scanProblem(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrProblemNotFound
		}
		return nil, err
	}
	return p, nil
}

func (r *problemRepositoryImpl) ListProblems(filter model.ProblemFilter) ([]model.Problem, int, error) {
	var conditions []string
	var args []interface{}

	if filter.WorkspaceID > 0 {
		conditions = append(conditions, "p.workspace_id = ?")
		args = append(args, filter.WorkspaceID)
	}

	if filter.ProjectID != nil {
		conditions = append(conditions, "p.project_id = ?")
		args = append(args, *filter.ProjectID)
	}

	if filter.Status != "" {
		conditions = append(conditions, "p.status = ?")
		args = append(args, strings.ToLower(strings.TrimSpace(filter.Status)))
	}

	if filter.Category != "" {
		conditions = append(conditions, "p.category = ?")
		args = append(args, strings.ToLower(strings.TrimSpace(filter.Category)))
	}

	if filter.Priority != "" {
		conditions = append(conditions, "p.priority = ?")
		args = append(args, strings.ToLower(strings.TrimSpace(filter.Priority)))
	}

	if filter.AssignedTo != "" {
		conditions = append(conditions, "p.assigned_to = ?")
		args = append(args, strings.TrimSpace(filter.AssignedTo))
	}

	if filter.AssignedTeam != "" {
		conditions = append(conditions, "p.assigned_team = ?")
		args = append(args, strings.TrimSpace(filter.AssignedTeam))
	}

	if filter.IsKnownError != nil {
		if *filter.IsKnownError {
			conditions = append(conditions, "p.is_known_error = 1")
		} else {
			conditions = append(conditions, "p.is_known_error = 0")
		}
	}

	if search := strings.TrimSpace(filter.Search); search != "" {
		pattern := "%" + search + "%"
		conditions = append(conditions, "(p.title LIKE ? OR p.problem_code LIKE ? OR p.description LIKE ? OR p.workaround LIKE ? OR p.root_cause_summary LIKE ?)")
		args = append(args, pattern, pattern, pattern, pattern, pattern)
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = " WHERE " + strings.Join(conditions, " AND ")
	}

	countQuery := `SELECT COUNT(*) FROM problems p` + whereClause
	var total int
	if err := r.db.QueryRow(countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

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

	dataQuery := baseProblemSelect + whereClause + ` ORDER BY p.id DESC LIMIT ? OFFSET ?`
	queryArgs := append(args, limit, offset)

	rows, err := r.db.Query(dataQuery, queryArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var problems []model.Problem
	for rows.Next() {
		p, err := scanProblem(rows)
		if err != nil {
			return nil, 0, err
		}
		problems = append(problems, *p)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	if problems == nil {
		problems = []model.Problem{}
	}
	return problems, total, nil
}

func (r *problemRepositoryImpl) UpdateProblem(p *model.Problem) error {
	p.UpdatedAt = time.Now().UTC()
	isKnownInt := 0
	if p.IsKnownError {
		isKnownInt = 1
	}

	query := `
	UPDATE problems SET
		project_id = ?,
		title = ?,
		description = ?,
		status = ?,
		category = ?,
		priority = ?,
		urgency = ?,
		impact = ?,
		assigned_to = ?,
		assigned_team = ?,
		root_cause_category = ?,
		root_cause_summary = ?,
		workaround = ?,
		permanent_solution = ?,
		is_known_error = ?,
		known_error_published_at = ?,
		rca_method = ?,
		started_at = ?,
		resolved_at = ?,
		closed_at = ?,
		updated_at = ?
	WHERE id = ?`

	res, err := r.db.Exec(query,
		p.ProjectID,
		p.Title,
		p.Description,
		p.Status,
		p.Category,
		p.Priority,
		p.Urgency,
		p.Impact,
		p.AssignedTo,
		p.AssignedTeam,
		p.RootCauseCategory,
		p.RootCauseSummary,
		p.Workaround,
		p.PermanentSolution,
		isKnownInt,
		p.KnownErrorPublishedAt,
		p.RCAMethod,
		p.StartedAt,
		p.ResolvedAt,
		p.ClosedAt,
		p.UpdatedAt,
		p.ID,
	)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrProblemNotFound
	}
	return nil
}

func (r *problemRepositoryImpl) DeleteProblem(id int64) error {
	res, err := r.db.Exec("DELETE FROM problems WHERE id = ?", id)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrProblemNotFound
	}
	return nil
}

func (r *problemRepositoryImpl) LinkIncident(link *model.ProblemIncidentLink) (int64, error) {
	if link.LinkedAt.IsZero() {
		link.LinkedAt = time.Now().UTC()
	}

	query := `
	INSERT INTO problem_incident_links (problem_id, incident_id, notes, linked_at)
	VALUES (?, ?, ?, ?)`

	res, err := r.db.Exec(query, link.ProblemID, link.IncidentID, link.Notes, link.LinkedAt)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return 0, ErrProblemLinkExists
		}
		return 0, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	link.ID = id
	return id, nil
}

func (r *problemRepositoryImpl) UnlinkIncident(problemID, incidentID int64) error {
	res, err := r.db.Exec("DELETE FROM problem_incident_links WHERE problem_id = ? AND incident_id = ?", problemID, incidentID)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrProblemLinkNotFound
	}
	return nil
}

func (r *problemRepositoryImpl) ListProblemIncidents(problemID int64) ([]model.ProblemIncidentLink, error) {
	query := `
	SELECT 
		l.id, l.problem_id, l.incident_id, l.notes, l.linked_at,
		COALESCE(i.incident_number, 0) AS incident_number,
		COALESCE(i.title, '') AS incident_title,
		COALESCE(i.severity, '') AS incident_severity,
		COALESCE(i.status, '') AS incident_status
	FROM problem_incident_links l
	LEFT JOIN incidents i ON l.incident_id = i.id
	WHERE l.problem_id = ?
	ORDER BY l.linked_at DESC`

	rows, err := r.db.Query(query, problemID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var links []model.ProblemIncidentLink
	for rows.Next() {
		var l model.ProblemIncidentLink
		if err := rows.Scan(
			&l.ID, &l.ProblemID, &l.IncidentID, &l.Notes, &l.LinkedAt,
			&l.IncidentNumber, &l.IncidentTitle, &l.IncidentSeverity, &l.IncidentStatus,
		); err != nil {
			return nil, err
		}
		links = append(links, l)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if links == nil {
		links = []model.ProblemIncidentLink{}
	}
	return links, nil
}

func (r *problemRepositoryImpl) GetProblemsForIncident(incidentID int64) ([]model.Problem, error) {
	query := baseProblemSelect + ` INNER JOIN problem_incident_links pil ON p.id = pil.problem_id WHERE pil.incident_id = ? ORDER BY p.id DESC`
	rows, err := r.db.Query(query, incidentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var problems []model.Problem
	for rows.Next() {
		p, err := scanProblem(rows)
		if err != nil {
			return nil, err
		}
		problems = append(problems, *p)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if problems == nil {
		problems = []model.Problem{}
	}
	return problems, nil
}

func (r *problemRepositoryImpl) LinkCI(link *model.ProblemCILink) (int64, error) {
	if link.LinkedAt.IsZero() {
		link.LinkedAt = time.Now().UTC()
	}
	if link.Relationship == "" {
		link.Relationship = model.ProblemCIRelRootCause
	}

	query := `
	INSERT INTO problem_ci_links (problem_id, ci_id, relationship, notes, linked_at)
	VALUES (?, ?, ?, ?, ?)`

	res, err := r.db.Exec(query, link.ProblemID, link.CIID, link.Relationship, link.Notes, link.LinkedAt)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return 0, ErrProblemLinkExists
		}
		return 0, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	link.ID = id
	return id, nil
}

func (r *problemRepositoryImpl) UnlinkCI(id int64) error {
	res, err := r.db.Exec("DELETE FROM problem_ci_links WHERE id = ?", id)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrProblemLinkNotFound
	}
	return nil
}

func (r *problemRepositoryImpl) ListProblemCIs(problemID int64) ([]model.ProblemCILink, error) {
	query := `
	SELECT 
		l.id, l.problem_id, l.ci_id, l.relationship, l.notes, l.linked_at,
		COALESCE(c.ci_code, '') AS ci_code,
		COALESCE(c.name, '') AS ci_name,
		COALESCE(c.ci_type, '') AS ci_type,
		COALESCE(c.criticality, '') AS criticality,
		COALESCE(c.status, '') AS ci_status
	FROM problem_ci_links l
	LEFT JOIN configuration_items c ON l.ci_id = c.id
	WHERE l.problem_id = ?
	ORDER BY l.linked_at DESC`

	rows, err := r.db.Query(query, problemID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var links []model.ProblemCILink
	for rows.Next() {
		var l model.ProblemCILink
		if err := rows.Scan(
			&l.ID, &l.ProblemID, &l.CIID, &l.Relationship, &l.Notes, &l.LinkedAt,
			&l.CICode, &l.CIName, &l.CIType, &l.Criticality, &l.CIStatus,
		); err != nil {
			return nil, err
		}
		links = append(links, l)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if links == nil {
		links = []model.ProblemCILink{}
	}
	return links, nil
}

func (r *problemRepositoryImpl) GetProblemsForCI(ciID int64) ([]model.Problem, error) {
	query := baseProblemSelect + ` INNER JOIN problem_ci_links pcl ON p.id = pcl.problem_id WHERE pcl.ci_id = ? ORDER BY p.id DESC`
	rows, err := r.db.Query(query, ciID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var problems []model.Problem
	for rows.Next() {
		p, err := scanProblem(rows)
		if err != nil {
			return nil, err
		}
		problems = append(problems, *p)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if problems == nil {
		problems = []model.Problem{}
	}
	return problems, nil
}

func (r *problemRepositoryImpl) LinkChange(link *model.ProblemChangeLink) (int64, error) {
	return 0, nil
}

func (r *problemRepositoryImpl) UnlinkChange(id int64) error {
	return nil
}

func (r *problemRepositoryImpl) ListProblemChanges(problemID int64) ([]model.ProblemChangeLink, error) {
	return nil, nil
}

func (r *problemRepositoryImpl) GetProblemsForChange(changeID int64) ([]model.Problem, error) {
	return nil, nil
}

func (r *problemRepositoryImpl) LinkTask(link *model.ProblemTaskLink) (int64, error) {
	return 0, nil
}

func (r *problemRepositoryImpl) UnlinkTask(id int64) error {
	return nil
}

func (r *problemRepositoryImpl) ListProblemTasks(problemID int64) ([]model.ProblemTaskLink, error) {
	return nil, nil
}

func (r *problemRepositoryImpl) GetProblemsForTask(taskID int) ([]model.Problem, error) {
	return nil, nil
}

func (r *problemRepositoryImpl) AddRCAStep(step *model.ProblemRCAStep) (int64, error) {
	return 0, nil
}

func (r *problemRepositoryImpl) GetRCAStep(id int64) (*model.ProblemRCAStep, error) {
	return nil, nil
}

func (r *problemRepositoryImpl) UpdateRCAStep(step *model.ProblemRCAStep) error {
	return nil
}

func (r *problemRepositoryImpl) DeleteRCAStep(id int64) error {
	return nil
}

func (r *problemRepositoryImpl) ListRCASteps(problemID int64) ([]model.ProblemRCAStep, error) {
	return nil, nil
}

func (r *problemRepositoryImpl) GetNextRCAStepNumber(problemID int64) (int, error) {
	return 0, nil
}

func (r *problemRepositoryImpl) ListKnownErrors(filter model.KnownErrorFilter) ([]model.Problem, int, error) {
	return nil, 0, nil
}

func (r *problemRepositoryImpl) GetProblemMetrics(workspaceID int64) (*model.ProblemMetrics, error) {
	return nil, nil
}
