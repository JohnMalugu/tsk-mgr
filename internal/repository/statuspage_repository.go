package repository

import (
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/JohnMalugu/tsk-mgr-api/internal/model"
)

var (
	ErrStatusPageNotFound      = errors.New("status page not found")
	ErrDuplicateStatusPageSlug = errors.New("status page slug already in use")
	ErrComponentNotFound       = errors.New("status component not found")
	ErrComponentGroupNotFound  = errors.New("component group not found")
	ErrStatusIncidentNotFound     = errors.New("status incident not found")
	ErrStatusMaintenanceNotFound = errors.New("maintenance window not found")
	ErrSubscriberNotFound         = errors.New("status subscriber not found")
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

func scanStatusPage(scanner interface{ Scan(dest ...interface{}) error }) (*model.StatusPage, error) {
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

func (r *statusPageRepositoryImpl) CreateComponent(comp *model.StatusComponent) (int64, error) {
	return 0, nil
}

func (r *statusPageRepositoryImpl) GetComponentByID(id int64) (*model.StatusComponent, error) {
	return nil, nil
}

func (r *statusPageRepositoryImpl) ListComponents(pageID int64) ([]model.StatusComponent, error) {
	return nil, nil
}

func (r *statusPageRepositoryImpl) UpdateComponent(comp *model.StatusComponent) error {
	return nil
}

func (r *statusPageRepositoryImpl) DeleteComponent(id int64) error {
	return nil
}

func (r *statusPageRepositoryImpl) UpdateComponentStatus(id int64, status string) error {
	return nil
}

func (r *statusPageRepositoryImpl) CreateIncident(inc *model.StatusIncident) (int64, error) {
	return 0, nil
}

func (r *statusPageRepositoryImpl) GetIncidentByID(id int64) (*model.StatusIncident, error) {
	return nil, nil
}

func (r *statusPageRepositoryImpl) ListIncidents(pageID int64, limit, offset int) ([]model.StatusIncident, int, error) {
	return nil, 0, nil
}

func (r *statusPageRepositoryImpl) UpdateIncident(inc *model.StatusIncident) error {
	return nil
}

func (r *statusPageRepositoryImpl) DeleteIncident(id int64) error {
	return nil
}

func (r *statusPageRepositoryImpl) AddIncidentUpdate(update *model.StatusIncidentUpdate) (int64, error) {
	return 0, nil
}

func (r *statusPageRepositoryImpl) ListIncidentUpdates(incidentID int64) ([]model.StatusIncidentUpdate, error) {
	return nil, nil
}

func (r *statusPageRepositoryImpl) CreateMaintenance(maint *model.StatusMaintenanceWindow) (int64, error) {
	return 0, nil
}

func (r *statusPageRepositoryImpl) GetMaintenanceByID(id int64) (*model.StatusMaintenanceWindow, error) {
	return nil, nil
}

func (r *statusPageRepositoryImpl) ListMaintenances(pageID int64, limit, offset int) ([]model.StatusMaintenanceWindow, int, error) {
	return nil, 0, nil
}

func (r *statusPageRepositoryImpl) UpdateMaintenance(maint *model.StatusMaintenanceWindow) error {
	return nil
}

func (r *statusPageRepositoryImpl) DeleteMaintenance(id int64) error {
	return nil
}

func (r *statusPageRepositoryImpl) CreateSubscriber(sub *model.StatusSubscriber) (int64, error) {
	return 0, nil
}

func (r *statusPageRepositoryImpl) GetSubscriberByID(id int64) (*model.StatusSubscriber, error) {
	return nil, nil
}

func (r *statusPageRepositoryImpl) ListSubscribers(pageID int64) ([]model.StatusSubscriber, error) {
	return nil, nil
}

func (r *statusPageRepositoryImpl) DeleteSubscriber(id int64) error {
	return nil
}

func (r *statusPageRepositoryImpl) GetPageMetrics(pageID int64) (*model.StatusPageMetrics, error) {
	return nil, nil
}

func (r *statusPageRepositoryImpl) ComputeOverallStatus(pageID int64) (string, error) {
	return "", nil
}
