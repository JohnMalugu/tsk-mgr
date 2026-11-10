package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/JohnMalugu/tsk-mgr-api/internal/model"
)

var (
	ErrCINotFound            = errors.New("configuration item not found")
	ErrRelationshipNotFound  = errors.New("ci relationship not found")
	ErrLinkNotFound          = errors.New("ci entity link not found")
	ErrMaintenanceNotFound   = errors.New("ci maintenance schedule not found")
)

// CMDBRepository manages database persistence for Configuration Items, relationships, entity links, and maintenance schedules
type CMDBRepository interface {
	GetNextCICode(workspaceID int64) (string, error)
	CreateCI(ci *model.ConfigurationItem) (int64, error)
	GetCIByID(id int64) (*model.ConfigurationItem, error)
	GetCIByCode(workspaceID int64, code string) (*model.ConfigurationItem, error)
	ListCIs(filter model.CIFilter) ([]*model.ConfigurationItem, error)
	UpdateCI(ci *model.ConfigurationItem) error
	DeleteCI(id int64) error

	CreateRelationship(rel *model.CIRelationship) (int64, error)
	GetRelationship(id int64) (*model.CIRelationship, error)
	ListRelationships(ciID int64) ([]model.CIRelationship, error)
	DeleteRelationship(id int64) error
	GetUpstreamDependencies(ciID int64, maxDepth int) ([]model.CIDependencyNode, error)
	GetDownstreamDependencies(ciID int64, maxDepth int) ([]model.CIDependencyNode, error)

	CreateEntityLink(link *model.CIEntityLink) (int64, error)
	ListEntityLinksByCI(ciID int64) ([]model.CIEntityLink, error)
	ListCIsByEntity(entityType string, entityID int64) ([]model.CIEntityLink, error)
	DeleteEntityLink(linkID int64) error

	CreateMaintenance(maint *model.CIMaintenanceSchedule) (int64, error)
	GetMaintenanceByID(id int64) (*model.CIMaintenanceSchedule, error)
	ListMaintenancesByCI(ciID int64) ([]model.CIMaintenanceSchedule, error)
	UpdateMaintenanceStatus(id int64, status string) error
	DeleteMaintenance(id int64) error

	GetCMDBMetrics(workspaceID int64) (*model.CMDBMetrics, error)
}

type cmdbRepositoryImpl struct {
	db *sql.DB
}

// NewCMDBRepository creates a new instance of CMDBRepository
func NewCMDBRepository(db *sql.DB) CMDBRepository {
	return &cmdbRepositoryImpl{db: db}
}

func (r *cmdbRepositoryImpl) GetNextCICode(workspaceID int64) (string, error) {
	query := `SELECT COALESCE(MAX(CAST(SUBSTR(ci_code, 4) AS INTEGER)), 0) + 1 FROM configuration_items WHERE workspace_id = ?`
	var nextNum int
	err := r.db.QueryRow(query, workspaceID).Scan(&nextNum)
	if err != nil {
		return "CI-1", err
	}
	return fmt.Sprintf("CI-%d", nextNum), nil
}

const baseCISelect = `
	SELECT 
		c.id, c.workspace_id, c.project_id, c.ci_code, c.name, c.ci_type, c.status,
		c.criticality, c.environment, c.owner_user_id, c.ip_address, c.hostname,
		c.version, c.cost_per_month, c.attributes, c.created_at, c.updated_at,
		COALESCE(u.username, u.email, '') AS owner_name,
		COALESCE(p.name, '') AS project_name,
		(SELECT COUNT(*) FROM ci_relationships WHERE source_ci_id = c.id) AS upstream_count,
		(SELECT COUNT(*) FROM ci_relationships WHERE target_ci_id = c.id) AS downstream_count,
		(SELECT COUNT(*) FROM ci_entity_links WHERE ci_id = c.id) AS linked_entities_count
	FROM configuration_items c
	LEFT JOIN users u ON c.owner_user_id = u.id
	LEFT JOIN projects p ON c.project_id = p.id
`

func scanCI(scanner interface{ Scan(dest ...interface{}) error }) (*model.ConfigurationItem, error) {
	var c model.ConfigurationItem
	var projectID sql.NullInt64
	var ownerName, projectName sql.NullString
	var upstreamCount, downstreamCount, linkedCount sql.NullInt64

	err := scanner.Scan(
		&c.ID, &c.WorkspaceID, &projectID, &c.CICode, &c.Name, &c.CIType, &c.Status,
		&c.Criticality, &c.Environment, &c.OwnerUserID, &c.IPAddress, &c.Hostname,
		&c.Version, &c.CostPerMonth, &c.Attributes, &c.CreatedAt, &c.UpdatedAt,
		&ownerName, &projectName,
		&upstreamCount, &downstreamCount, &linkedCount,
	)
	if err != nil {
		return nil, err
	}
	if projectID.Valid {
		c.ProjectID = &projectID.Int64
	}
	if ownerName.Valid {
		c.OwnerName = ownerName.String
	}
	if projectName.Valid {
		c.ProjectName = projectName.String
	}
	if upstreamCount.Valid {
		c.UpstreamCount = int(upstreamCount.Int64)
	}
	if downstreamCount.Valid {
		c.DownstreamCount = int(downstreamCount.Int64)
	}
	if linkedCount.Valid {
		c.LinkedEntitiesCount = int(linkedCount.Int64)
	}
	return &c, nil
}

func (r *cmdbRepositoryImpl) CreateCI(ci *model.ConfigurationItem) (int64, error) {
	if ci.CICode == "" {
		code, err := r.GetNextCICode(ci.WorkspaceID)
		if err != nil {
			return 0, err
		}
		ci.CICode = code
	}
	now := time.Now().UTC()
	if ci.CreatedAt.IsZero() {
		ci.CreatedAt = now
	}
	if ci.UpdatedAt.IsZero() {
		ci.UpdatedAt = now
	}
	if ci.CIType == "" {
		ci.CIType = model.CITypeService
	}
	if ci.Status == "" {
		ci.Status = model.CIStatusActive
	}
	if ci.Criticality == "" {
		ci.Criticality = model.CICriticalityTier2
	}
	if ci.Environment == "" {
		ci.Environment = model.CIEnvProduction
	}
	if ci.Attributes == "" {
		ci.Attributes = "{}"
	}

	query := `
	INSERT INTO configuration_items (
		workspace_id, project_id, ci_code, name, ci_type, status,
		criticality, environment, owner_user_id, ip_address, hostname,
		version, cost_per_month, attributes, created_at, updated_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	res, err := r.db.Exec(query,
		ci.WorkspaceID, ci.ProjectID, ci.CICode, ci.Name, ci.CIType, ci.Status,
		ci.Criticality, ci.Environment, ci.OwnerUserID, ci.IPAddress, ci.Hostname,
		ci.Version, ci.CostPerMonth, ci.Attributes, ci.CreatedAt, ci.UpdatedAt,
	)
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	ci.ID = id
	return id, nil
}

func (r *cmdbRepositoryImpl) GetCIByID(id int64) (*model.ConfigurationItem, error) {
	query := baseCISelect + ` WHERE c.id = ?`
	row := r.db.QueryRow(query, id)
	ci, err := scanCI(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrCINotFound
		}
		return nil, err
	}
	return ci, nil
}

func (r *cmdbRepositoryImpl) GetCIByCode(workspaceID int64, code string) (*model.ConfigurationItem, error) {
	query := baseCISelect + ` WHERE c.workspace_id = ? AND c.ci_code = ?`
	row := r.db.QueryRow(query, workspaceID, code)
	ci, err := scanCI(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrCINotFound
		}
		return nil, err
	}
	return ci, nil
}

func (r *cmdbRepositoryImpl) ListCIs(filter model.CIFilter) ([]*model.ConfigurationItem, error) {
	var conditions []string
	var args []interface{}

	if filter.WorkspaceID > 0 {
		conditions = append(conditions, "c.workspace_id = ?")
		args = append(args, filter.WorkspaceID)
	}
	if filter.ProjectID != nil {
		conditions = append(conditions, "c.project_id = ?")
		args = append(args, *filter.ProjectID)
	}
	if filter.CIType != nil && *filter.CIType != "" {
		conditions = append(conditions, "c.ci_type = ?")
		args = append(args, *filter.CIType)
	}
	if filter.Status != nil && *filter.Status != "" {
		conditions = append(conditions, "c.status = ?")
		args = append(args, *filter.Status)
	}
	if filter.Criticality != nil && *filter.Criticality != "" {
		conditions = append(conditions, "c.criticality = ?")
		args = append(args, *filter.Criticality)
	}
	if filter.Environment != nil && *filter.Environment != "" {
		conditions = append(conditions, "c.environment = ?")
		args = append(args, *filter.Environment)
	}
	if filter.OwnerUserID != nil && *filter.OwnerUserID != "" {
		conditions = append(conditions, "c.owner_user_id = ?")
		args = append(args, *filter.OwnerUserID)
	}
	if filter.Search != nil && *filter.Search != "" {
		pattern := "%" + *filter.Search + "%"
		conditions = append(conditions, "(c.name LIKE ? OR c.hostname LIKE ? OR c.ip_address LIKE ? OR c.ci_code LIKE ?)")
		args = append(args, pattern, pattern, pattern, pattern)
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = " WHERE " + strings.Join(conditions, " AND ")
	}

	query := baseCISelect + whereClause + " ORDER BY c.created_at DESC"
	if filter.Limit > 0 {
		query += " LIMIT ?"
		args = append(args, filter.Limit)
		if filter.Offset > 0 {
			query += " OFFSET ?"
			args = append(args, filter.Offset)
		}
	}

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*model.ConfigurationItem
	for rows.Next() {
		ci, err := scanCI(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, ci)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return list, nil
}

func (r *cmdbRepositoryImpl) UpdateCI(ci *model.ConfigurationItem) error {
	ci.UpdatedAt = time.Now().UTC()
	query := `
	UPDATE configuration_items SET
		project_id = ?,
		name = ?,
		ci_type = ?,
		status = ?,
		criticality = ?,
		environment = ?,
		owner_user_id = ?,
		ip_address = ?,
		hostname = ?,
		version = ?,
		cost_per_month = ?,
		attributes = ?,
		updated_at = ?
	WHERE id = ?`

	res, err := r.db.Exec(query,
		ci.ProjectID,
		ci.Name,
		ci.CIType,
		ci.Status,
		ci.Criticality,
		ci.Environment,
		ci.OwnerUserID,
		ci.IPAddress,
		ci.Hostname,
		ci.Version,
		ci.CostPerMonth,
		ci.Attributes,
		ci.UpdatedAt,
		ci.ID,
	)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrCINotFound
	}
	return nil
}

func (r *cmdbRepositoryImpl) DeleteCI(id int64) error {
	res, err := r.db.Exec("DELETE FROM configuration_items WHERE id = ?", id)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrCINotFound
	}
	return nil
}

const baseRelSelect = `
	SELECT 
		r.id, r.source_ci_id, r.target_ci_id, r.relation_type, r.description, r.created_at,
		COALESCE(src.name, '') AS source_ci_name,
		COALESCE(src.ci_code, '') AS source_ci_code,
		COALESCE(src.ci_type, '') AS source_ci_type,
		COALESCE(tgt.name, '') AS target_ci_name,
		COALESCE(tgt.ci_code, '') AS target_ci_code,
		COALESCE(tgt.ci_type, '') AS target_ci_type
	FROM ci_relationships r
	LEFT JOIN configuration_items src ON r.source_ci_id = src.id
	LEFT JOIN configuration_items tgt ON r.target_ci_id = tgt.id
`

func scanRel(scanner interface{ Scan(dest ...interface{}) error }) (*model.CIRelationship, error) {
	var r model.CIRelationship
	var srcName, srcCode, srcType, tgtName, tgtCode, tgtType sql.NullString
	err := scanner.Scan(
		&r.ID, &r.SourceCIID, &r.TargetCIID, &r.RelationType, &r.Description, &r.CreatedAt,
		&srcName, &srcCode, &srcType, &tgtName, &tgtCode, &tgtType,
	)
	if err != nil {
		return nil, err
	}
	if srcName.Valid {
		r.SourceCIName = srcName.String
	}
	if srcCode.Valid {
		r.SourceCICode = srcCode.String
	}
	if srcType.Valid {
		r.SourceCIType = srcType.String
	}
	if tgtName.Valid {
		r.TargetCIName = tgtName.String
	}
	if tgtCode.Valid {
		r.TargetCICode = tgtCode.String
	}
	if tgtType.Valid {
		r.TargetCIType = tgtType.String
	}
	return &r, nil
}

func (r *cmdbRepositoryImpl) CreateRelationship(rel *model.CIRelationship) (int64, error) {
	if rel.CreatedAt.IsZero() {
		rel.CreatedAt = time.Now().UTC()
	}
	if rel.RelationType == "" {
		rel.RelationType = model.CIRelationDependsOn
	}

	query := `
	INSERT INTO ci_relationships (source_ci_id, target_ci_id, relation_type, description, created_at)
	VALUES (?, ?, ?, ?, ?)`

	res, err := r.db.Exec(query, rel.SourceCIID, rel.TargetCIID, rel.RelationType, rel.Description, rel.CreatedAt)
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	rel.ID = id
	return id, nil
}

func (r *cmdbRepositoryImpl) GetRelationship(id int64) (*model.CIRelationship, error) {
	query := baseRelSelect + ` WHERE r.id = ?`
	row := r.db.QueryRow(query, id)
	rel, err := scanRel(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrRelationshipNotFound
		}
		return nil, err
	}
	return rel, nil
}

func (r *cmdbRepositoryImpl) ListRelationships(ciID int64) ([]model.CIRelationship, error) {
	query := baseRelSelect + ` WHERE r.source_ci_id = ? OR r.target_ci_id = ? ORDER BY r.created_at ASC`
	rows, err := r.db.Query(query, ciID, ciID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rels []model.CIRelationship
	for rows.Next() {
		rel, err := scanRel(rows)
		if err != nil {
			return nil, err
		}
		rels = append(rels, *rel)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return rels, nil
}

func (r *cmdbRepositoryImpl) DeleteRelationship(id int64) error {
	res, err := r.db.Exec("DELETE FROM ci_relationships WHERE id = ?", id)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrRelationshipNotFound
	}
	return nil
}

func (r *cmdbRepositoryImpl) GetUpstreamDependencies(ciID int64, maxDepth int) ([]model.CIDependencyNode, error) {
	if maxDepth <= 0 || maxDepth > 10 {
		maxDepth = 5
	}
	visited := make(map[int64]bool)
	visited[ciID] = true

	var nodes []model.CIDependencyNode
	currentLevel := []int64{ciID}

	for depth := 1; depth <= maxDepth && len(currentLevel) > 0; depth++ {
		var nextLevel []int64
		for _, currID := range currentLevel {
			rows, err := r.db.Query(`SELECT target_ci_id, relation_type FROM ci_relationships WHERE source_ci_id = ?`, currID)
			if err != nil {
				return nil, err
			}
			for rows.Next() {
				var targetID int64
				var relType string
				if err := rows.Scan(&targetID, &relType); err != nil {
					rows.Close()
					return nil, err
				}
				if !visited[targetID] {
					visited[targetID] = true
					targetCI, err := r.GetCIByID(targetID)
					if err == nil && targetCI != nil {
						nodes = append(nodes, model.CIDependencyNode{
							CI:           *targetCI,
							RelationType: relType,
							Depth:        depth,
							Direction:    "upstream",
						})
						nextLevel = append(nextLevel, targetID)
					}
				}
			}
			rows.Close()
		}
		currentLevel = nextLevel
	}
	return nodes, nil
}

func (r *cmdbRepositoryImpl) GetDownstreamDependencies(ciID int64, maxDepth int) ([]model.CIDependencyNode, error) {
	if maxDepth <= 0 || maxDepth > 10 {
		maxDepth = 5
	}
	visited := make(map[int64]bool)
	visited[ciID] = true

	var nodes []model.CIDependencyNode
	currentLevel := []int64{ciID}

	for depth := 1; depth <= maxDepth && len(currentLevel) > 0; depth++ {
		var nextLevel []int64
		for _, currID := range currentLevel {
			rows, err := r.db.Query(`SELECT source_ci_id, relation_type FROM ci_relationships WHERE target_ci_id = ?`, currID)
			if err != nil {
				return nil, err
			}
			for rows.Next() {
				var sourceID int64
				var relType string
				if err := rows.Scan(&sourceID, &relType); err != nil {
					rows.Close()
					return nil, err
				}
				if !visited[sourceID] {
					visited[sourceID] = true
					sourceCI, err := r.GetCIByID(sourceID)
					if err == nil && sourceCI != nil {
						nodes = append(nodes, model.CIDependencyNode{
							CI:           *sourceCI,
							RelationType: relType,
							Depth:        depth,
							Direction:    "downstream",
						})
						nextLevel = append(nextLevel, sourceID)
					}
				}
			}
			rows.Close()
		}
		currentLevel = nextLevel
	}
	return nodes, nil
}

const baseLinkSelect = `
	SELECT 
		l.id, l.ci_id, l.entity_type, l.entity_id, l.impact_level, l.notes, l.created_at,
		COALESCE(c.name, '') AS ci_name,
		COALESCE(c.ci_code, '') AS ci_code,
		CASE 
			WHEN l.entity_type = 'task' THEN (SELECT COALESCE(title, '') FROM tasks WHERE id = l.entity_id)
			WHEN l.entity_type = 'change' THEN (SELECT COALESCE(title, '') FROM change_requests WHERE id = l.entity_id)
			WHEN l.entity_type = 'incident' THEN (SELECT COALESCE(title, '') FROM incidents WHERE id = l.entity_id)
			ELSE ''
		END AS entity_title,
		CASE 
			WHEN l.entity_type = 'task' THEN (SELECT COALESCE(status, '') FROM tasks WHERE id = l.entity_id)
			WHEN l.entity_type = 'change' THEN (SELECT COALESCE(status, '') FROM change_requests WHERE id = l.entity_id)
			WHEN l.entity_type = 'incident' THEN (SELECT COALESCE(status, '') FROM incidents WHERE id = l.entity_id)
			ELSE ''
		END AS entity_status
	FROM ci_entity_links l
	LEFT JOIN configuration_items c ON l.ci_id = c.id
`

func scanLink(scanner interface{ Scan(dest ...interface{}) error }) (*model.CIEntityLink, error) {
	var l model.CIEntityLink
	var ciName, ciCode, entityTitle, entityStatus sql.NullString
	err := scanner.Scan(
		&l.ID, &l.CIID, &l.EntityType, &l.EntityID, &l.ImpactLevel, &l.Notes, &l.CreatedAt,
		&ciName, &ciCode, &entityTitle, &entityStatus,
	)
	if err != nil {
		return nil, err
	}
	if ciName.Valid {
		l.CIName = ciName.String
	}
	if ciCode.Valid {
		l.CICode = ciCode.String
	}
	if entityTitle.Valid {
		l.EntityTitle = entityTitle.String
	}
	if entityStatus.Valid {
		l.EntityStatus = entityStatus.String
	}
	return &l, nil
}

func (r *cmdbRepositoryImpl) CreateEntityLink(link *model.CIEntityLink) (int64, error) {
	if link.CreatedAt.IsZero() {
		link.CreatedAt = time.Now().UTC()
	}
	if link.ImpactLevel == "" {
		link.ImpactLevel = model.CIImpactDirect
	}

	query := `
	INSERT INTO ci_entity_links (ci_id, entity_type, entity_id, impact_level, notes, created_at)
	VALUES (?, ?, ?, ?, ?, ?)`

	res, err := r.db.Exec(query, link.CIID, link.EntityType, link.EntityID, link.ImpactLevel, link.Notes, link.CreatedAt)
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	link.ID = id
	return id, nil
}

func (r *cmdbRepositoryImpl) ListEntityLinksByCI(ciID int64) ([]model.CIEntityLink, error) {
	query := baseLinkSelect + ` WHERE l.ci_id = ? ORDER BY l.created_at ASC`
	rows, err := r.db.Query(query, ciID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var links []model.CIEntityLink
	for rows.Next() {
		link, err := scanLink(rows)
		if err != nil {
			return nil, err
		}
		links = append(links, *link)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return links, nil
}

func (r *cmdbRepositoryImpl) ListCIsByEntity(entityType string, entityID int64) ([]model.CIEntityLink, error) {
	query := baseLinkSelect + ` WHERE l.entity_type = ? AND l.entity_id = ? ORDER BY l.created_at ASC`
	rows, err := r.db.Query(query, entityType, entityID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var links []model.CIEntityLink
	for rows.Next() {
		link, err := scanLink(rows)
		if err != nil {
			return nil, err
		}
		links = append(links, *link)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return links, nil
}

func (r *cmdbRepositoryImpl) DeleteEntityLink(linkID int64) error {
	res, err := r.db.Exec("DELETE FROM ci_entity_links WHERE id = ?", linkID)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrLinkNotFound
	}
	return nil
}

func (r *cmdbRepositoryImpl) CreateMaintenance(maint *model.CIMaintenanceSchedule) (int64, error) {
	return 0, nil
}

func (r *cmdbRepositoryImpl) GetMaintenanceByID(id int64) (*model.CIMaintenanceSchedule, error) {
	return nil, nil
}

func (r *cmdbRepositoryImpl) ListMaintenancesByCI(ciID int64) ([]model.CIMaintenanceSchedule, error) {
	return nil, nil
}

func (r *cmdbRepositoryImpl) UpdateMaintenanceStatus(id int64, status string) error {
	return nil
}

func (r *cmdbRepositoryImpl) DeleteMaintenance(id int64) error {
	return nil
}

func (r *cmdbRepositoryImpl) GetCMDBMetrics(workspaceID int64) (*model.CMDBMetrics, error) {
	return nil, nil
}
