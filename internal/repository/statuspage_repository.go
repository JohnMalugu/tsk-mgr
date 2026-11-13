package repository

import (
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/JohnMalugu/tsk-mgr-api/internal/model"
)

var (
	ErrStatusPageNotFound        = errors.New("status page not found")
	ErrDuplicateStatusPageSlug   = errors.New("status page slug already in use")
	ErrComponentNotFound         = errors.New("status component not found")
	ErrComponentGroupNotFound    = errors.New("component group not found")
	ErrStatusIncidentNotFound    = errors.New("status incident not found")
	ErrStatusMaintenanceNotFound = errors.New("maintenance window not found")
	ErrSubscriberNotFound        = errors.New("status subscriber not found")
)

// StatusPageRepository defines persistence operations for status pages, components, incidents, maintenances, and subscribers
type StatusPageRepository interface {
	// Status Pages
	CreatePage(page *model.StatusPage) (int64, error)
	GetPageByID(id int64) (*model.StatusPage, error)
	GetPageBySlug(slug string) (*model.StatusPage, error)
	ListPages(filter model.StatusPageFilter) ([]model.StatusPage, int, error)
	UpdatePage(page *model.StatusPage) error
	DeletePage(id int64) error

	// Component Groups
	CreateComponentGroup(group *model.ComponentGroup) (int64, error)
	GetComponentGroupByID(id int64) (*model.ComponentGroup, error)
	ListComponentGroups(pageID int64) ([]model.ComponentGroup, error)
	UpdateComponentGroup(group *model.ComponentGroup) error
	DeleteComponentGroup(id int64) error

	// Status Components
	CreateComponent(comp *model.StatusComponent) (int64, error)
	GetComponentByID(id int64) (*model.StatusComponent, error)
	ListComponents(pageID int64) ([]model.StatusComponent, error)
	UpdateComponent(comp *model.StatusComponent) error
	DeleteComponent(id int64) error
	UpdateComponentStatus(id int64, status string) error

	// Status Incidents & Timeline Updates
	CreateIncident(inc *model.StatusIncident) (int64, error)
	GetIncidentByID(id int64) (*model.StatusIncident, error)
	ListIncidents(pageID int64, limit, offset int) ([]model.StatusIncident, int, error)
	UpdateIncident(inc *model.StatusIncident) error
	DeleteIncident(id int64) error
	AddIncidentUpdate(update *model.StatusIncidentUpdate) (int64, error)
	ListIncidentUpdates(incidentID int64) ([]model.StatusIncidentUpdate, error)

	// Maintenance Windows
	CreateMaintenance(maint *model.StatusMaintenanceWindow) (int64, error)
	GetMaintenanceByID(id int64) (*model.StatusMaintenanceWindow, error)
	ListMaintenances(pageID int64, limit, offset int) ([]model.StatusMaintenanceWindow, int, error)
	UpdateMaintenance(maint *model.StatusMaintenanceWindow) error
	DeleteMaintenance(id int64) error

	// Subscribers
	CreateSubscriber(sub *model.StatusSubscriber) (int64, error)
	GetSubscriberByID(id int64) (*model.StatusSubscriber, error)
	ListSubscribers(pageID int64) ([]model.StatusSubscriber, error)
	DeleteSubscriber(id int64) error

	// Metrics & Aggregated Status
	GetPageMetrics(pageID int64) (*model.StatusPageMetrics, error)
	ComputeOverallStatus(pageID int64) (string, error)
}

type statusPageRepositoryImpl struct {
	db *sql.DB
}

// NewStatusPageRepository creates a new SQLite-backed StatusPageRepository instance
func NewStatusPageRepository(db *sql.DB) StatusPageRepository {
	return &statusPageRepositoryImpl{db: db}
}

func scanStatusPage(scanner interface {
	Scan(dest ...interface{}) error
}) (*model.StatusPage, error) {
	var sp model.StatusPage
	var projectID sql.NullInt64
	var projectName, creatorName sql.NullString
	var isActiveInt int
	var compCount, incCount int

	err := scanner.Scan(
		&sp.ID, &sp.WorkspaceID, &projectID, &sp.Name, &sp.Slug,
		&sp.Description, &sp.Headline, &sp.Visibility, &sp.CustomDomain,
		&sp.Theme, &sp.CurrentStatus, &isActiveInt, &sp.CreatedByID,
		&sp.CreatedAt, &sp.UpdatedAt,
		&projectName, &creatorName, &compCount, &incCount,
	)
	if err != nil {
		return nil, err
	}
	if projectID.Valid {
		pid := projectID.Int64
		sp.ProjectID = &pid
	}
	sp.IsActive = (isActiveInt == 1)
	if projectName.Valid {
		sp.ProjectName = projectName.String
	}
	if creatorName.Valid {
		sp.CreatorName = creatorName.String
	}
	sp.ComponentsCount = compCount
	sp.ActiveIncidentsCount = incCount
	return &sp, nil
}

func (r *statusPageRepositoryImpl) CreatePage(page *model.StatusPage) (int64, error) {
	now := time.Now()
	if page.CreatedAt.IsZero() {
		page.CreatedAt = now
	}
	if page.UpdatedAt.IsZero() {
		page.UpdatedAt = now
	}
	if page.Visibility == "" {
		page.Visibility = model.VisibilityPublic
	}
	if page.Theme == "" {
		page.Theme = model.ThemeSystem
	}
	if page.CurrentStatus == "" {
		page.CurrentStatus = model.StatusOperational
	}

	isActiveInt := 0
	if page.IsActive {
		isActiveInt = 1
	}

	// Check duplicate slug
	var existingCount int
	_ = r.db.QueryRow(`SELECT COUNT(*) FROM status_pages WHERE slug = ?`, page.Slug).Scan(&existingCount)
	if existingCount > 0 {
		return 0, ErrDuplicateStatusPageSlug
	}

	res, err := r.db.Exec(`
		INSERT INTO status_pages (
			workspace_id, project_id, name, slug, description, headline,
			visibility, custom_domain, theme, current_status, is_active,
			created_by_id, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, page.WorkspaceID, page.ProjectID, page.Name, page.Slug, page.Description,
		page.Headline, page.Visibility, page.CustomDomain, page.Theme,
		page.CurrentStatus, isActiveInt, page.CreatedByID, page.CreatedAt, page.UpdatedAt)
	if err != nil {
		return 0, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	page.ID = id
	return id, nil
}

func (r *statusPageRepositoryImpl) GetPageByID(id int64) (*model.StatusPage, error) {
	query := `
		SELECT sp.id, sp.workspace_id, sp.project_id, sp.name, sp.slug,
		       sp.description, sp.headline, sp.visibility, sp.custom_domain,
		       sp.theme, sp.current_status, sp.is_active, sp.created_by_id,
		       sp.created_at, sp.updated_at,
		       COALESCE(p.name, ''), COALESCE(u.username, ''),
		       (SELECT COUNT(*) FROM status_components WHERE page_id = sp.id),
		       (SELECT COUNT(*) FROM status_incidents WHERE page_id = sp.id AND status != 'resolved')
		FROM status_pages sp
		LEFT JOIN projects p ON sp.project_id = p.id
		LEFT JOIN users u ON sp.created_by_id = u.id
		WHERE sp.id = ?
	`
	sp, err := scanStatusPage(r.db.QueryRow(query, id))
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrStatusPageNotFound
		}
		return nil, err
	}
	return sp, nil
}

func (r *statusPageRepositoryImpl) GetPageBySlug(slug string) (*model.StatusPage, error) {
	query := `
		SELECT sp.id, sp.workspace_id, sp.project_id, sp.name, sp.slug,
		       sp.description, sp.headline, sp.visibility, sp.custom_domain,
		       sp.theme, sp.current_status, sp.is_active, sp.created_by_id,
		       sp.created_at, sp.updated_at,
		       COALESCE(p.name, ''), COALESCE(u.username, ''),
		       (SELECT COUNT(*) FROM status_components WHERE page_id = sp.id),
		       (SELECT COUNT(*) FROM status_incidents WHERE page_id = sp.id AND status != 'resolved')
		FROM status_pages sp
		LEFT JOIN projects p ON sp.project_id = p.id
		LEFT JOIN users u ON sp.created_by_id = u.id
		WHERE sp.slug = ?
	`
	sp, err := scanStatusPage(r.db.QueryRow(query, slug))
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrStatusPageNotFound
		}
		return nil, err
	}
	return sp, nil
}

func (r *statusPageRepositoryImpl) ListPages(filter model.StatusPageFilter) ([]model.StatusPage, int, error) {
	var conditions []string
	var args []interface{}

	conditions = append(conditions, "sp.workspace_id = ?")
	args = append(args, filter.WorkspaceID)

	if filter.ProjectID != nil {
		conditions = append(conditions, "sp.project_id = ?")
		args = append(args, *filter.ProjectID)
	}
	if strings.TrimSpace(filter.Visibility) != "" {
		conditions = append(conditions, "sp.visibility = ?")
		args = append(args, strings.TrimSpace(filter.Visibility))
	}
	if strings.TrimSpace(filter.Search) != "" {
		conditions = append(conditions, "(sp.name LIKE ? OR sp.description LIKE ? OR sp.slug LIKE ?)")
		pattern := "%" + strings.TrimSpace(filter.Search) + "%"
		args = append(args, pattern, pattern, pattern)
	}

	whereClause := " WHERE " + strings.Join(conditions, " AND ")

	countQuery := `SELECT COUNT(*) FROM status_pages sp` + whereClause
	var totalCount int
	if err := r.db.QueryRow(countQuery, args...).Scan(&totalCount); err != nil {
		return nil, 0, err
	}

	dataQuery := `
		SELECT sp.id, sp.workspace_id, sp.project_id, sp.name, sp.slug,
		       sp.description, sp.headline, sp.visibility, sp.custom_domain,
		       sp.theme, sp.current_status, sp.is_active, sp.created_by_id,
		       sp.created_at, sp.updated_at,
		       COALESCE(p.name, ''), COALESCE(u.username, ''),
		       (SELECT COUNT(*) FROM status_components WHERE page_id = sp.id),
		       (SELECT COUNT(*) FROM status_incidents WHERE page_id = sp.id AND status != 'resolved')
		FROM status_pages sp
		LEFT JOIN projects p ON sp.project_id = p.id
		LEFT JOIN users u ON sp.created_by_id = u.id
	` + whereClause + " ORDER BY sp.id DESC"

	if filter.Limit > 0 {
		dataQuery += " LIMIT ?"
		args = append(args, filter.Limit)
		if filter.Offset > 0 {
			dataQuery += " OFFSET ?"
			args = append(args, filter.Offset)
		}
	}

	rows, err := r.db.Query(dataQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var pages []model.StatusPage
	for rows.Next() {
		sp, err := scanStatusPage(rows)
		if err != nil {
			return nil, 0, err
		}
		pages = append(pages, *sp)
	}
	return pages, totalCount, nil
}

func (r *statusPageRepositoryImpl) UpdatePage(page *model.StatusPage) error {
	isActiveInt := 0
	if page.IsActive {
		isActiveInt = 1
	}
	page.UpdatedAt = time.Now()

	res, err := r.db.Exec(`
		UPDATE status_pages
		SET name = ?, slug = ?, description = ?, headline = ?,
		    visibility = ?, custom_domain = ?, theme = ?, current_status = ?,
		    is_active = ?, updated_at = ?
		WHERE id = ?
	`, page.Name, page.Slug, page.Description, page.Headline,
		page.Visibility, page.CustomDomain, page.Theme, page.CurrentStatus,
		isActiveInt, page.UpdatedAt, page.ID)
	if err != nil {
		return err
	}

	rowsAff, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAff == 0 {
		return ErrStatusPageNotFound
	}
	return nil
}

func (r *statusPageRepositoryImpl) DeletePage(id int64) error {
	res, err := r.db.Exec(`DELETE FROM status_pages WHERE id = ?`, id)
	if err != nil {
		return err
	}
	rowsAff, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAff == 0 {
		return ErrStatusPageNotFound
	}
	return nil
}

func (r *statusPageRepositoryImpl) CreateComponentGroup(group *model.ComponentGroup) (int64, error) {
	now := time.Now()
	if group.CreatedAt.IsZero() {
		group.CreatedAt = now
	}
	if group.UpdatedAt.IsZero() {
		group.UpdatedAt = now
	}

	collapsedInt := 0
	if group.Collapsed {
		collapsedInt = 1
	}

	res, err := r.db.Exec(`
		INSERT INTO status_component_groups (
			page_id, name, description, display_order, collapsed, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?)
	`, group.PageID, group.Name, group.Description, group.DisplayOrder, collapsedInt, group.CreatedAt, group.UpdatedAt)
	if err != nil {
		return 0, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	group.ID = id
	return id, nil
}

func (r *statusPageRepositoryImpl) GetComponentGroupByID(id int64) (*model.ComponentGroup, error) {
	var g model.ComponentGroup
	var collapsedInt int
	var compCount int

	err := r.db.QueryRow(`
		SELECT id, page_id, name, description, display_order, collapsed, created_at, updated_at,
		       (SELECT COUNT(*) FROM status_components WHERE group_id = status_component_groups.id)
		FROM status_component_groups
		WHERE id = ?
	`, id).Scan(&g.ID, &g.PageID, &g.Name, &g.Description, &g.DisplayOrder, &collapsedInt, &g.CreatedAt, &g.UpdatedAt, &compCount)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrComponentGroupNotFound
		}
		return nil, err
	}
	g.Collapsed = (collapsedInt == 1)
	g.ComponentsCount = compCount
	return &g, nil
}

func (r *statusPageRepositoryImpl) ListComponentGroups(pageID int64) ([]model.ComponentGroup, error) {
	rows, err := r.db.Query(`
		SELECT id, page_id, name, description, display_order, collapsed, created_at, updated_at,
		       (SELECT COUNT(*) FROM status_components WHERE group_id = status_component_groups.id)
		FROM status_component_groups
		WHERE page_id = ?
		ORDER BY display_order ASC, id ASC
	`, pageID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var groups []model.ComponentGroup
	for rows.Next() {
		var g model.ComponentGroup
		var collapsedInt int
		var compCount int
		if err := rows.Scan(&g.ID, &g.PageID, &g.Name, &g.Description, &g.DisplayOrder, &collapsedInt, &g.CreatedAt, &g.UpdatedAt, &compCount); err != nil {
			return nil, err
		}
		g.Collapsed = (collapsedInt == 1)
		g.ComponentsCount = compCount
		groups = append(groups, g)
	}
	return groups, nil
}

func (r *statusPageRepositoryImpl) UpdateComponentGroup(group *model.ComponentGroup) error {
	group.UpdatedAt = time.Now()
	collapsedInt := 0
	if group.Collapsed {
		collapsedInt = 1
	}

	res, err := r.db.Exec(`
		UPDATE status_component_groups
		SET name = ?, description = ?, display_order = ?, collapsed = ?, updated_at = ?
		WHERE id = ?
	`, group.Name, group.Description, group.DisplayOrder, collapsedInt, group.UpdatedAt, group.ID)
	if err != nil {
		return err
	}

	rowsAff, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAff == 0 {
		return ErrComponentGroupNotFound
	}
	return nil
}

func (r *statusPageRepositoryImpl) DeleteComponentGroup(id int64) error {
	res, err := r.db.Exec(`DELETE FROM status_component_groups WHERE id = ?`, id)
	if err != nil {
		return err
	}
	rowsAff, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAff == 0 {
		return ErrComponentGroupNotFound
	}
	return nil
}

func scanStatusComponent(scanner interface {
	Scan(dest ...interface{}) error
}) (*model.StatusComponent, error) {
	var c model.StatusComponent
	var groupID, ciID sql.NullInt64
	var groupName, ciName, ciCode sql.NullString
	var isMonitoredInt int

	err := scanner.Scan(
		&c.ID, &c.PageID, &groupID, &ciID, &c.Name, &c.Description,
		&c.Status, &c.DisplayOrder, &isMonitoredInt, &c.CreatedAt, &c.UpdatedAt,
		&groupName, &ciName, &ciCode,
	)
	if err != nil {
		return nil, err
	}
	if groupID.Valid {
		gid := groupID.Int64
		c.GroupID = &gid
	}
	if ciID.Valid {
		cid := ciID.Int64
		c.CIID = &cid
	}
	c.IsMonitored = (isMonitoredInt == 1)
	if groupName.Valid {
		c.GroupName = groupName.String
	}
	if ciName.Valid {
		c.CIName = ciName.String
	}
	if ciCode.Valid {
		c.CICode = ciCode.String
	}
	return &c, nil
}

func (r *statusPageRepositoryImpl) CreateComponent(comp *model.StatusComponent) (int64, error) {
	now := time.Now()
	if comp.CreatedAt.IsZero() {
		comp.CreatedAt = now
	}
	if comp.UpdatedAt.IsZero() {
		comp.UpdatedAt = now
	}
	if comp.Status == "" {
		comp.Status = model.StatusOperational
	}

	isMonitoredInt := 0
	if comp.IsMonitored {
		isMonitoredInt = 1
	}

	res, err := r.db.Exec(`
		INSERT INTO status_components (
			page_id, group_id, ci_id, name, description, status,
			display_order, is_monitored, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, comp.PageID, comp.GroupID, comp.CIID, comp.Name, comp.Description,
		comp.Status, comp.DisplayOrder, isMonitoredInt, comp.CreatedAt, comp.UpdatedAt)
	if err != nil {
		return 0, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	comp.ID = id

	_, _ = r.ComputeOverallStatus(comp.PageID)
	return id, nil
}

func (r *statusPageRepositoryImpl) GetComponentByID(id int64) (*model.StatusComponent, error) {
	query := `
		SELECT c.id, c.page_id, c.group_id, c.ci_id, c.name, c.description,
		       c.status, c.display_order, c.is_monitored, c.created_at, c.updated_at,
		       COALESCE(g.name, ''), COALESCE(ci.name, ''), COALESCE(ci.ci_code, '')
		FROM status_components c
		LEFT JOIN status_component_groups g ON c.group_id = g.id
		LEFT JOIN configuration_items ci ON c.ci_id = ci.id
		WHERE c.id = ?
	`
	comp, err := scanStatusComponent(r.db.QueryRow(query, id))
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrComponentNotFound
		}
		return nil, err
	}
	return comp, nil
}

func (r *statusPageRepositoryImpl) ListComponents(pageID int64) ([]model.StatusComponent, error) {
	query := `
		SELECT c.id, c.page_id, c.group_id, c.ci_id, c.name, c.description,
		       c.status, c.display_order, c.is_monitored, c.created_at, c.updated_at,
		       COALESCE(g.name, ''), COALESCE(ci.name, ''), COALESCE(ci.ci_code, '')
		FROM status_components c
		LEFT JOIN status_component_groups g ON c.group_id = g.id
		LEFT JOIN configuration_items ci ON c.ci_id = ci.id
		WHERE c.page_id = ?
		ORDER BY c.display_order ASC, c.id ASC
	`
	rows, err := r.db.Query(query, pageID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var components []model.StatusComponent
	for rows.Next() {
		comp, err := scanStatusComponent(rows)
		if err != nil {
			return nil, err
		}
		components = append(components, *comp)
	}
	return components, nil
}

func (r *statusPageRepositoryImpl) UpdateComponent(comp *model.StatusComponent) error {
	comp.UpdatedAt = time.Now()
	isMonitoredInt := 0
	if comp.IsMonitored {
		isMonitoredInt = 1
	}

	res, err := r.db.Exec(`
		UPDATE status_components
		SET group_id = ?, ci_id = ?, name = ?, description = ?, status = ?,
		    display_order = ?, is_monitored = ?, updated_at = ?
		WHERE id = ?
	`, comp.GroupID, comp.CIID, comp.Name, comp.Description, comp.Status,
		comp.DisplayOrder, isMonitoredInt, comp.UpdatedAt, comp.ID)
	if err != nil {
		return err
	}

	rowsAff, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAff == 0 {
		return ErrComponentNotFound
	}

	_, _ = r.ComputeOverallStatus(comp.PageID)
	return nil
}

func (r *statusPageRepositoryImpl) DeleteComponent(id int64) error {
	var pageID int64
	_ = r.db.QueryRow(`SELECT page_id FROM status_components WHERE id = ?`, id).Scan(&pageID)

	res, err := r.db.Exec(`DELETE FROM status_components WHERE id = ?`, id)
	if err != nil {
		return err
	}
	rowsAff, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAff == 0 {
		return ErrComponentNotFound
	}

	if pageID > 0 {
		_, _ = r.ComputeOverallStatus(pageID)
	}
	return nil
}

func (r *statusPageRepositoryImpl) UpdateComponentStatus(id int64, status string) error {
	var pageID int64
	err := r.db.QueryRow(`SELECT page_id FROM status_components WHERE id = ?`, id).Scan(&pageID)
	if err != nil {
		if err == sql.ErrNoRows {
			return ErrComponentNotFound
		}
		return err
	}

	now := time.Now()
	res, err := r.db.Exec(`UPDATE status_components SET status = ?, updated_at = ? WHERE id = ?`, status, now, id)
	if err != nil {
		return err
	}
	rowsAff, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAff == 0 {
		return ErrComponentNotFound
	}

	_, _ = r.ComputeOverallStatus(pageID)
	return nil
}

func (r *statusPageRepositoryImpl) ComputeOverallStatus(pageID int64) (string, error) {
	rows, err := r.db.Query(`SELECT status FROM status_components WHERE page_id = ? AND is_monitored = 1`, pageID)
	if err != nil {
		return model.StatusOperational, err
	}
	defer rows.Close()

	hasMajorOutage := false
	hasPartialOutage := false
	hasDegraded := false
	hasMaintenance := false

	for rows.Next() {
		var st string
		if err := rows.Scan(&st); err == nil {
			switch st {
			case model.StatusMajorOutage:
				hasMajorOutage = true
			case model.StatusPartialOutage:
				hasPartialOutage = true
			case model.StatusDegradedPerformance:
				hasDegraded = true
			case model.StatusUnderMaintenance:
				hasMaintenance = true
			}
		}
	}

	overall := model.StatusOperational
	if hasMajorOutage {
		overall = model.StatusMajorOutage
	} else if hasPartialOutage {
		overall = model.StatusPartialOutage
	} else if hasDegraded {
		overall = model.StatusDegradedPerformance
	} else if hasMaintenance {
		overall = model.StatusUnderMaintenance
	}

	now := time.Now()
	_, _ = r.db.Exec(`UPDATE status_pages SET current_status = ?, updated_at = ? WHERE id = ?`, overall, now, pageID)
	return overall, nil
}

func scanStatusIncident(scanner interface {
	Scan(dest ...interface{}) error
}) (*model.StatusIncident, error) {
	var inc model.StatusIncident
	var incidentID sql.NullInt64
	var resolvedAt sql.NullTime
	var incidentNumber sql.NullInt64

	err := scanner.Scan(
		&inc.ID, &inc.PageID, &incidentID, &inc.Title, &inc.Status,
		&inc.Impact, &inc.StartedAt, &resolvedAt, &inc.AffectedComponentsJSON,
		&inc.CreatedByID, &inc.CreatedAt, &inc.UpdatedAt,
		&incidentNumber,
	)
	if err != nil {
		return nil, err
	}
	if incidentID.Valid {
		iid := incidentID.Int64
		inc.IncidentID = &iid
	}
	if resolvedAt.Valid {
		rt := resolvedAt.Time
		inc.ResolvedAt = &rt
	}
	if incidentNumber.Valid {
		in := int(incidentNumber.Int64)
		inc.IncidentNumber = &in
	}
	return &inc, nil
}

func (r *statusPageRepositoryImpl) CreateIncident(inc *model.StatusIncident) (int64, error) {
	now := time.Now()
	if inc.CreatedAt.IsZero() {
		inc.CreatedAt = now
	}
	if inc.UpdatedAt.IsZero() {
		inc.UpdatedAt = now
	}
	if inc.StartedAt.IsZero() {
		inc.StartedAt = now
	}
	if inc.Status == "" {
		inc.Status = model.StatusIncidentInvestigating
	}
	if inc.Impact == "" {
		inc.Impact = model.ImpactMinor
	}
	if inc.AffectedComponentsJSON == "" {
		inc.AffectedComponentsJSON = "[]"
	}

	res, err := r.db.Exec(`
		INSERT INTO status_incidents (
			page_id, incident_id, title, status, impact, started_at,
			resolved_at, affected_components_json, created_by_id, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, inc.PageID, inc.IncidentID, inc.Title, inc.Status, inc.Impact,
		inc.StartedAt, inc.ResolvedAt, inc.AffectedComponentsJSON, inc.CreatedByID, inc.CreatedAt, inc.UpdatedAt)
	if err != nil {
		return 0, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	inc.ID = id
	return id, nil
}

func (r *statusPageRepositoryImpl) GetIncidentByID(id int64) (*model.StatusIncident, error) {
	query := `
		SELECT si.id, si.page_id, si.incident_id, si.title, si.status,
		       si.impact, si.started_at, si.resolved_at, si.affected_components_json,
		       si.created_by_id, si.created_at, si.updated_at,
		       i.incident_number
		FROM status_incidents si
		LEFT JOIN incidents i ON si.incident_id = i.id
		WHERE si.id = ?
	`
	inc, err := scanStatusIncident(r.db.QueryRow(query, id))
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrStatusIncidentNotFound
		}
		return nil, err
	}

	updates, err := r.ListIncidentUpdates(id)
	if err == nil {
		inc.Updates = updates
	}
	return inc, nil
}

func (r *statusPageRepositoryImpl) ListIncidents(pageID int64, limit, offset int) ([]model.StatusIncident, int, error) {
	countQuery := `SELECT COUNT(*) FROM status_incidents WHERE page_id = ?`
	var totalCount int
	if err := r.db.QueryRow(countQuery, pageID).Scan(&totalCount); err != nil {
		return nil, 0, err
	}

	dataQuery := `
		SELECT si.id, si.page_id, si.incident_id, si.title, si.status,
		       si.impact, si.started_at, si.resolved_at, si.affected_components_json,
		       si.created_by_id, si.created_at, si.updated_at,
		       i.incident_number
		FROM status_incidents si
		LEFT JOIN incidents i ON si.incident_id = i.id
		WHERE si.page_id = ?
		ORDER BY si.started_at DESC, si.id DESC
	`
	var args []interface{}
	args = append(args, pageID)
	if limit > 0 {
		dataQuery += " LIMIT ?"
		args = append(args, limit)
		if offset > 0 {
			dataQuery += " OFFSET ?"
			args = append(args, offset)
		}
	}

	rows, err := r.db.Query(dataQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var incidents []model.StatusIncident
	for rows.Next() {
		inc, err := scanStatusIncident(rows)
		if err != nil {
			return nil, 0, err
		}
		updates, _ := r.ListIncidentUpdates(inc.ID)
		inc.Updates = updates
		incidents = append(incidents, *inc)
	}
	return incidents, totalCount, nil
}

func (r *statusPageRepositoryImpl) UpdateIncident(inc *model.StatusIncident) error {
	inc.UpdatedAt = time.Now()
	res, err := r.db.Exec(`
		UPDATE status_incidents
		SET title = ?, status = ?, impact = ?, resolved_at = ?,
		    affected_components_json = ?, updated_at = ?
		WHERE id = ?
	`, inc.Title, inc.Status, inc.Impact, inc.ResolvedAt,
		inc.AffectedComponentsJSON, inc.UpdatedAt, inc.ID)
	if err != nil {
		return err
	}
	rowsAff, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAff == 0 {
		return ErrStatusIncidentNotFound
	}
	return nil
}

func (r *statusPageRepositoryImpl) DeleteIncident(id int64) error {
	res, err := r.db.Exec(`DELETE FROM status_incidents WHERE id = ?`, id)
	if err != nil {
		return err
	}
	rowsAff, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAff == 0 {
		return ErrStatusIncidentNotFound
	}
	return nil
}

func (r *statusPageRepositoryImpl) AddIncidentUpdate(update *model.StatusIncidentUpdate) (int64, error) {
	now := time.Now()
	if update.CreatedAt.IsZero() {
		update.CreatedAt = now
	}

	res, err := r.db.Exec(`
		INSERT INTO status_incident_updates (
			status_incident_id, status, message, posted_by_id, created_at
		) VALUES (?, ?, ?, ?, ?)
	`, update.StatusIncidentID, update.Status, update.Message, update.PostedByID, update.CreatedAt)
	if err != nil {
		return 0, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	update.ID = id

	// Sync incident status & timestamp if resolved
	if update.Status == model.StatusIncidentResolved {
		_, _ = r.db.Exec(`UPDATE status_incidents SET status = ?, resolved_at = ?, updated_at = ? WHERE id = ?`,
			update.Status, now, now, update.StatusIncidentID)
	} else {
		_, _ = r.db.Exec(`UPDATE status_incidents SET status = ?, updated_at = ? WHERE id = ?`,
			update.Status, now, update.StatusIncidentID)
	}

	return id, nil
}

func (r *statusPageRepositoryImpl) ListIncidentUpdates(incidentID int64) ([]model.StatusIncidentUpdate, error) {
	query := `
		SELECT u.id, u.status_incident_id, u.status, u.message, u.posted_by_id, u.created_at,
		       COALESCE(usr.username, '')
		FROM status_incident_updates u
		LEFT JOIN users usr ON u.posted_by_id = usr.id
		WHERE u.status_incident_id = ?
		ORDER BY u.created_at DESC, u.id DESC
	`
	rows, err := r.db.Query(query, incidentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var updates []model.StatusIncidentUpdate
	for rows.Next() {
		var u model.StatusIncidentUpdate
		var posterName sql.NullString
		if err := rows.Scan(&u.ID, &u.StatusIncidentID, &u.Status, &u.Message, &u.PostedByID, &u.CreatedAt, &posterName); err != nil {
			return nil, err
		}
		if posterName.Valid {
			u.PosterName = posterName.String
		}
		updates = append(updates, u)
	}
	return updates, nil
}

func scanMaintenanceWindow(scanner interface {
	Scan(dest ...interface{}) error
}) (*model.StatusMaintenanceWindow, error) {
	var m model.StatusMaintenanceWindow
	var actualStart, actualEnd sql.NullTime

	err := scanner.Scan(
		&m.ID, &m.PageID, &m.Title, &m.Description, &m.Status,
		&m.ScheduledStartAt, &m.ScheduledEndAt, &actualStart, &actualEnd,
		&m.AffectedComponentsJSON, &m.CreatedByID, &m.CreatedAt, &m.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	if actualStart.Valid {
		as := actualStart.Time
		m.ActualStartAt = &as
	}
	if actualEnd.Valid {
		ae := actualEnd.Time
		m.ActualEndAt = &ae
	}
	return &m, nil
}

func (r *statusPageRepositoryImpl) CreateMaintenance(maint *model.StatusMaintenanceWindow) (int64, error) {
	now := time.Now()
	if maint.CreatedAt.IsZero() {
		maint.CreatedAt = now
	}
	if maint.UpdatedAt.IsZero() {
		maint.UpdatedAt = now
	}
	if maint.Status == "" {
		maint.Status = model.MaintenanceStatusScheduled
	}
	if maint.AffectedComponentsJSON == "" {
		maint.AffectedComponentsJSON = "[]"
	}

	res, err := r.db.Exec(`
		INSERT INTO status_maintenance_windows (
			page_id, title, description, status, scheduled_start_at, scheduled_end_at,
			actual_start_at, actual_end_at, affected_components_json, created_by_id,
			created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, maint.PageID, maint.Title, maint.Description, maint.Status, maint.ScheduledStartAt,
		maint.ScheduledEndAt, maint.ActualStartAt, maint.ActualEndAt, maint.AffectedComponentsJSON,
		maint.CreatedByID, maint.CreatedAt, maint.UpdatedAt)
	if err != nil {
		return 0, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	maint.ID = id
	return id, nil
}

func (r *statusPageRepositoryImpl) GetMaintenanceByID(id int64) (*model.StatusMaintenanceWindow, error) {
	query := `
		SELECT id, page_id, title, description, status, scheduled_start_at, scheduled_end_at,
		       actual_start_at, actual_end_at, affected_components_json, created_by_id,
		       created_at, updated_at
		FROM status_maintenance_windows
		WHERE id = ?
	`
	m, err := scanMaintenanceWindow(r.db.QueryRow(query, id))
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrStatusMaintenanceNotFound
		}
		return nil, err
	}
	return m, nil
}

func (r *statusPageRepositoryImpl) ListMaintenances(pageID int64, limit, offset int) ([]model.StatusMaintenanceWindow, int, error) {
	countQuery := `SELECT COUNT(*) FROM status_maintenance_windows WHERE page_id = ?`
	var totalCount int
	if err := r.db.QueryRow(countQuery, pageID).Scan(&totalCount); err != nil {
		return nil, 0, err
	}

	dataQuery := `
		SELECT id, page_id, title, description, status, scheduled_start_at, scheduled_end_at,
		       actual_start_at, actual_end_at, affected_components_json, created_by_id,
		       created_at, updated_at
		FROM status_maintenance_windows
		WHERE page_id = ?
		ORDER BY scheduled_start_at DESC, id DESC
	`
	var args []interface{}
	args = append(args, pageID)
	if limit > 0 {
		dataQuery += " LIMIT ?"
		args = append(args, limit)
		if offset > 0 {
			dataQuery += " OFFSET ?"
			args = append(args, offset)
		}
	}

	rows, err := r.db.Query(dataQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var list []model.StatusMaintenanceWindow
	for rows.Next() {
		m, err := scanMaintenanceWindow(rows)
		if err != nil {
			return nil, 0, err
		}
		list = append(list, *m)
	}
	return list, totalCount, nil
}

func (r *statusPageRepositoryImpl) UpdateMaintenance(maint *model.StatusMaintenanceWindow) error {
	maint.UpdatedAt = time.Now()
	res, err := r.db.Exec(`
		UPDATE status_maintenance_windows
		SET title = ?, description = ?, status = ?, scheduled_start_at = ?,
		    scheduled_end_at = ?, actual_start_at = ?, actual_end_at = ?,
		    affected_components_json = ?, updated_at = ?
		WHERE id = ?
	`, maint.Title, maint.Description, maint.Status, maint.ScheduledStartAt,
		maint.ScheduledEndAt, maint.ActualStartAt, maint.ActualEndAt,
		maint.AffectedComponentsJSON, maint.UpdatedAt, maint.ID)
	if err != nil {
		return err
	}
	rowsAff, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAff == 0 {
		return ErrStatusMaintenanceNotFound
	}
	return nil
}

func (r *statusPageRepositoryImpl) DeleteMaintenance(id int64) error {
	res, err := r.db.Exec(`DELETE FROM status_maintenance_windows WHERE id = ?`, id)
	if err != nil {
		return err
	}
	rowsAff, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAff == 0 {
		return ErrStatusMaintenanceNotFound
	}
	return nil
}

func scanSubscriber(scanner interface {
	Scan(dest ...interface{}) error
}) (*model.StatusSubscriber, error) {
	var s model.StatusSubscriber
	var isVerifiedInt int

	err := scanner.Scan(
		&s.ID, &s.PageID, &s.SubscriberType, &s.TargetEndpoint,
		&s.VerificationToken, &isVerifiedInt, &s.SubscribedComponentsJSON,
		&s.CreatedAt, &s.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	s.IsVerified = (isVerifiedInt == 1)
	return &s, nil
}

func (r *statusPageRepositoryImpl) CreateSubscriber(sub *model.StatusSubscriber) (int64, error) {
	now := time.Now()
	if sub.CreatedAt.IsZero() {
		sub.CreatedAt = now
	}
	if sub.UpdatedAt.IsZero() {
		sub.UpdatedAt = now
	}
	if sub.SubscriberType == "" {
		sub.SubscriberType = model.SubscriberTypeEmail
	}
	if sub.SubscribedComponentsJSON == "" {
		sub.SubscribedComponentsJSON = "[]"
	}
	isVerifiedInt := 1

	res, err := r.db.Exec(`
		INSERT INTO status_subscribers (
			page_id, subscriber_type, target_endpoint, verification_token,
			is_verified, subscribed_components_json, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, sub.PageID, sub.SubscriberType, sub.TargetEndpoint, sub.VerificationToken,
		isVerifiedInt, sub.SubscribedComponentsJSON, sub.CreatedAt, sub.UpdatedAt)
	if err != nil {
		return 0, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	sub.ID = id
	return id, nil
}

func (r *statusPageRepositoryImpl) GetSubscriberByID(id int64) (*model.StatusSubscriber, error) {
	query := `
		SELECT id, page_id, subscriber_type, target_endpoint,
		       verification_token, is_verified, subscribed_components_json,
		       created_at, updated_at
		FROM status_subscribers
		WHERE id = ?
	`
	sub, err := scanSubscriber(r.db.QueryRow(query, id))
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrSubscriberNotFound
		}
		return nil, err
	}
	return sub, nil
}

func (r *statusPageRepositoryImpl) ListSubscribers(pageID int64) ([]model.StatusSubscriber, error) {
	query := `
		SELECT id, page_id, subscriber_type, target_endpoint,
		       verification_token, is_verified, subscribed_components_json,
		       created_at, updated_at
		FROM status_subscribers
		WHERE page_id = ?
		ORDER BY id DESC
	`
	rows, err := r.db.Query(query, pageID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var subs []model.StatusSubscriber
	for rows.Next() {
		sub, err := scanSubscriber(rows)
		if err != nil {
			return nil, err
		}
		subs = append(subs, *sub)
	}
	return subs, nil
}

func (r *statusPageRepositoryImpl) DeleteSubscriber(id int64) error {
	res, err := r.db.Exec(`DELETE FROM status_subscribers WHERE id = ?`, id)
	if err != nil {
		return err
	}
	rowsAff, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAff == 0 {
		return ErrSubscriberNotFound
	}
	return nil
}

func (r *statusPageRepositoryImpl) GetPageMetrics(pageID int64) (*model.StatusPageMetrics, error) {
	metrics := &model.StatusPageMetrics{
		UptimePercentage30d: 100.0,
		UptimePercentage60d: 100.0,
		UptimePercentage90d: 100.0,
	}

	// 1. Component breakdown
	rows, err := r.db.Query(`SELECT status, COUNT(*) FROM status_components WHERE page_id = ? GROUP BY status`, pageID)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var st string
			var count int
			if err := rows.Scan(&st, &count); err == nil {
				metrics.TotalComponents += count
				switch st {
				case model.StatusOperational:
					metrics.OperationalComponents = count
				case model.StatusDegradedPerformance:
					metrics.DegradedComponents = count
				case model.StatusPartialOutage, model.StatusMajorOutage:
					metrics.OutageComponents += count
				case model.StatusUnderMaintenance:
					metrics.UnderMaintenanceComponents = count
				}
			}
		}
	}

	// 2. Active incidents count
	_ = r.db.QueryRow(`
		SELECT COUNT(*) FROM status_incidents WHERE page_id = ? AND status != 'resolved'
	`, pageID).Scan(&metrics.ActiveIncidents)

	// 3. Resolved incidents in last 30 days
	_ = r.db.QueryRow(`
		SELECT COUNT(*) FROM status_incidents
		WHERE page_id = ? AND status = 'resolved' AND resolved_at >= datetime('now', '-30 days')
	`, pageID).Scan(&metrics.ResolvedIncidents30d)

	// 4. Scheduled maintenance count
	_ = r.db.QueryRow(`
		SELECT COUNT(*) FROM status_maintenance_windows
		WHERE page_id = ? AND status IN ('scheduled', 'in_progress')
	`, pageID).Scan(&metrics.ScheduledMaintenances)

	// 5. Total subscribers
	_ = r.db.QueryRow(`
		SELECT COUNT(*) FROM status_subscribers WHERE page_id = ?
	`, pageID).Scan(&metrics.TotalSubscribers)

	// 6. Uptime percentages based on major/critical outage incident duration
	calculateUptime := func(days int) float64 {
		totalMinutes := float64(days * 1440)
		var downtimeMinutes float64
		_ = r.db.QueryRow(`
			SELECT COALESCE(SUM((julianday(COALESCE(resolved_at, datetime('now'))) - julianday(started_at)) * 1440.0), 0.0)
			FROM status_incidents
			WHERE page_id = ? AND impact IN ('major', 'critical') AND started_at >= datetime('now', printf('-%d days', ?))
		`, pageID, days).Scan(&downtimeMinutes)

		if downtimeMinutes <= 0.0 {
			return 100.0
		}
		pct := 100.0 - (downtimeMinutes / totalMinutes * 100.0)
		if pct < 0.0 {
			return 0.0
		}
		if pct > 100.0 {
			return 100.0
		}
		return pct
	}

	metrics.UptimePercentage30d = calculateUptime(30)
	metrics.UptimePercentage60d = calculateUptime(60)
	metrics.UptimePercentage90d = calculateUptime(90)

	return metrics, nil
}
