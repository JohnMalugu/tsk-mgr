package repository

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/JohnMalugu/tsk-mgr-api/internal/model"
)

// CustomFieldRepository defines database access methods for custom field definitions and values
type CustomFieldRepository interface {
	CreateFieldDefinition(def *model.CustomFieldDefinition) (int64, error)
	GetFieldDefinitionByID(id int64) (*model.CustomFieldDefinition, error)
	GetFieldDefinitionByKey(workspaceID int64, key string) (*model.CustomFieldDefinition, error)
	ListFieldDefinitions(filter model.CustomFieldFilter) ([]model.CustomFieldDefinition, int, error)
	UpdateFieldDefinition(def *model.CustomFieldDefinition) error
	DeleteFieldDefinition(id int64) error

	SetFieldValue(val *model.CustomFieldValue) error
	GetFieldValueByTaskAndField(taskID, fieldID int64) (*model.CustomFieldValue, error)
	GetFieldValuesByTask(taskID int64) ([]model.TaskCustomFieldValueDTO, error)
	DeleteFieldValue(taskID, fieldID int64) error
	FindTasksByFieldValue(filter model.CustomFieldValueFilter) ([]int64, int, error)
}

type sqlCustomFieldRepository struct {
	db *sql.DB
}

// NewCustomFieldRepository constructs a new CustomFieldRepository instance
func NewCustomFieldRepository(db *sql.DB) CustomFieldRepository {
	return &sqlCustomFieldRepository{db: db}
}

func (r *sqlCustomFieldRepository) CreateFieldDefinition(def *model.CustomFieldDefinition) (int64, error) {
	now := time.Now()
	if def.CreatedAt.IsZero() {
		def.CreatedAt = now
	}
	if def.UpdatedAt.IsZero() {
		def.UpdatedAt = now
	}

	optionsJSON, err := json.Marshal(def.Options)
	if err != nil {
		optionsJSON = []byte("[]")
	}

	query := `
	INSERT INTO custom_field_definitions (
		workspace_id, project_id, name, field_key, description, type, required, options, default_value, created_at, updated_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	res, err := r.db.Exec(
		query,
		def.WorkspaceID,
		def.ProjectID,
		def.Name,
		def.FieldKey,
		def.Description,
		string(def.Type),
		def.Required,
		string(optionsJSON),
		def.DefaultValue,
		def.CreatedAt,
		def.UpdatedAt,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to create custom field definition: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get last insert ID: %w", err)
	}
	def.ID = id
	return id, nil
}

func (r *sqlCustomFieldRepository) GetFieldDefinitionByID(id int64) (*model.CustomFieldDefinition, error) {
	query := `
	SELECT id, workspace_id, project_id, name, field_key, description, type, required, options, default_value, created_at, updated_at
	FROM custom_field_definitions
	WHERE id = ?`

	row := r.db.QueryRow(query, id)
	var def model.CustomFieldDefinition
	var typeStr, optionsStr string
	err := row.Scan(
		&def.ID,
		&def.WorkspaceID,
		&def.ProjectID,
		&def.Name,
		&def.FieldKey,
		&def.Description,
		&typeStr,
		&def.Required,
		&optionsStr,
		&def.DefaultValue,
		&def.CreatedAt,
		&def.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get custom field definition by id: %w", err)
	}

	def.Type = model.CustomFieldType(typeStr)
	if optionsStr != "" {
		_ = json.Unmarshal([]byte(optionsStr), &def.Options)
	}
	if def.Options == nil {
		def.Options = []string{}
	}

	return &def, nil
}

func (r *sqlCustomFieldRepository) GetFieldDefinitionByKey(workspaceID int64, key string) (*model.CustomFieldDefinition, error) {
	query := `
	SELECT id, workspace_id, project_id, name, field_key, description, type, required, options, default_value, created_at, updated_at
	FROM custom_field_definitions
	WHERE workspace_id = ? AND field_key = ?`

	row := r.db.QueryRow(query, workspaceID, key)
	var def model.CustomFieldDefinition
	var typeStr, optionsStr string
	err := row.Scan(
		&def.ID,
		&def.WorkspaceID,
		&def.ProjectID,
		&def.Name,
		&def.FieldKey,
		&def.Description,
		&typeStr,
		&def.Required,
		&optionsStr,
		&def.DefaultValue,
		&def.CreatedAt,
		&def.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get custom field definition by key: %w", err)
	}

	def.Type = model.CustomFieldType(typeStr)
	if optionsStr != "" {
		_ = json.Unmarshal([]byte(optionsStr), &def.Options)
	}
	if def.Options == nil {
		def.Options = []string{}
	}

	return &def, nil
}

func (r *sqlCustomFieldRepository) ListFieldDefinitions(filter model.CustomFieldFilter) ([]model.CustomFieldDefinition, int, error) {
	var conditions []string
	var args []interface{}

	if filter.WorkspaceID > 0 {
		conditions = append(conditions, "workspace_id = ?")
		args = append(args, filter.WorkspaceID)
	}
	if filter.ProjectID != nil {
		conditions = append(conditions, "(project_id = ? OR project_id IS NULL)")
		args = append(args, *filter.ProjectID)
	}
	if filter.Type != nil {
		conditions = append(conditions, "type = ?")
		args = append(args, string(*filter.Type))
	}
	if filter.Required != nil {
		conditions = append(conditions, "required = ?")
		args = append(args, *filter.Required)
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM custom_field_definitions %s", whereClause)
	var total int
	err := r.db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count custom field definitions: %w", err)
	}

	limit := filter.Limit
	if limit <= 0 {
		limit = 20
	}
	offset := filter.Offset
	if offset < 0 {
		offset = 0
	}

	query := fmt.Sprintf(`
	SELECT id, workspace_id, project_id, name, field_key, description, type, required, options, default_value, created_at, updated_at
	FROM custom_field_definitions
	%s
	ORDER BY id ASC
	LIMIT ? OFFSET ?`, whereClause)

	queryArgs := append(args, limit, offset)
	rows, err := r.db.Query(query, queryArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list custom field definitions: %w", err)
	}
	defer rows.Close()

	var defs []model.CustomFieldDefinition
	for rows.Next() {
		var def model.CustomFieldDefinition
		var typeStr, optionsStr string
		err := rows.Scan(
			&def.ID,
			&def.WorkspaceID,
			&def.ProjectID,
			&def.Name,
			&def.FieldKey,
			&def.Description,
			&typeStr,
			&def.Required,
			&optionsStr,
			&def.DefaultValue,
			&def.CreatedAt,
			&def.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan custom field definition: %w", err)
		}
		def.Type = model.CustomFieldType(typeStr)
		if optionsStr != "" {
			_ = json.Unmarshal([]byte(optionsStr), &def.Options)
		}
		if def.Options == nil {
			def.Options = []string{}
		}
		defs = append(defs, def)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("row error listing custom field definitions: %w", err)
	}

	return defs, total, nil
}

func (r *sqlCustomFieldRepository) UpdateFieldDefinition(def *model.CustomFieldDefinition) error {
	def.UpdatedAt = time.Now()
	optionsJSON, err := json.Marshal(def.Options)
	if err != nil {
		optionsJSON = []byte("[]")
	}

	query := `
	UPDATE custom_field_definitions
	SET name = ?, description = ?, required = ?, options = ?, default_value = ?, updated_at = ?
	WHERE id = ?`

	res, err := r.db.Exec(
		query,
		def.Name,
		def.Description,
		def.Required,
		string(optionsJSON),
		def.DefaultValue,
		def.UpdatedAt,
		def.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update custom field definition: %w", err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *sqlCustomFieldRepository) DeleteFieldDefinition(id int64) error {
	query := `DELETE FROM custom_field_definitions WHERE id = ?`
	res, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete custom field definition: %w", err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *sqlCustomFieldRepository) SetFieldValue(val *model.CustomFieldValue) error {
	now := time.Now()
	if val.CreatedAt.IsZero() {
		val.CreatedAt = now
	}
	val.UpdatedAt = now

	query := `
	INSERT INTO custom_field_values (
		field_id, task_id, value_text, value_number, value_boolean, value_date, created_at, updated_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	ON CONFLICT(field_id, task_id) DO UPDATE SET
		value_text = excluded.value_text,
		value_number = excluded.value_number,
		value_boolean = excluded.value_boolean,
		value_date = excluded.value_date,
		updated_at = excluded.updated_at`

	res, err := r.db.Exec(
		query,
		val.FieldID,
		val.TaskID,
		val.ValueText,
		val.ValueNumber,
		val.ValueBoolean,
		val.ValueDate,
		val.CreatedAt,
		val.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to set custom field value: %w", err)
	}

	id, err := res.LastInsertId()
	if err == nil && id > 0 {
		val.ID = id
	}
	return nil
}

func (r *sqlCustomFieldRepository) GetFieldValueByTaskAndField(taskID, fieldID int64) (*model.CustomFieldValue, error) {
	query := `
	SELECT id, field_id, task_id, value_text, value_number, value_boolean, value_date, created_at, updated_at
	FROM custom_field_values
	WHERE task_id = ? AND field_id = ?`

	row := r.db.QueryRow(query, taskID, fieldID)
	var val model.CustomFieldValue
	err := row.Scan(
		&val.ID,
		&val.FieldID,
		&val.TaskID,
		&val.ValueText,
		&val.ValueNumber,
		&val.ValueBoolean,
		&val.ValueDate,
		&val.CreatedAt,
		&val.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get custom field value: %w", err)
	}
	return &val, nil
}

func (r *sqlCustomFieldRepository) GetFieldValuesByTask(taskID int64) ([]model.TaskCustomFieldValueDTO, error) {
	query := `
	SELECT 
		d.id, d.name, d.field_key, d.type, d.required, d.options,
		v.value_text, v.value_number, v.value_boolean, v.value_date, v.updated_at
	FROM custom_field_values v
	JOIN custom_field_definitions d ON v.field_id = d.id
	WHERE v.task_id = ?
	ORDER BY d.id ASC`

	rows, err := r.db.Query(query, taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to get custom field values by task: %w", err)
	}
	defer rows.Close()

	var result []model.TaskCustomFieldValueDTO
	for rows.Next() {
		var dto model.TaskCustomFieldValueDTO
		var typeStr, optionsStr string
		err := rows.Scan(
			&dto.FieldID,
			&dto.FieldName,
			&dto.FieldKey,
			&typeStr,
			&dto.Required,
			&optionsStr,
			&dto.ValueText,
			&dto.ValueNumber,
			&dto.ValueBoolean,
			&dto.ValueDate,
			&dto.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan task custom field value: %w", err)
		}

		dto.Type = model.CustomFieldType(typeStr)
		if optionsStr != "" {
			_ = json.Unmarshal([]byte(optionsStr), &dto.Options)
		}
		if dto.Options == nil {
			dto.Options = []string{}
		}

		switch dto.Type {
		case model.CustomFieldTypeNumber:
			if dto.ValueNumber != nil {
				dto.Value = *dto.ValueNumber
			}
		case model.CustomFieldTypeBoolean:
			if dto.ValueBoolean != nil {
				dto.Value = *dto.ValueBoolean
			}
		case model.CustomFieldTypeDate:
			if dto.ValueDate != nil {
				dto.Value = *dto.ValueDate
			}
		default:
			if dto.ValueText != nil {
				dto.Value = *dto.ValueText
			}
		}

		result = append(result, dto)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row error fetching task custom field values: %w", err)
	}

	return result, nil
}

func (r *sqlCustomFieldRepository) DeleteFieldValue(taskID, fieldID int64) error {
	query := `DELETE FROM custom_field_values WHERE task_id = ? AND field_id = ?`
	res, err := r.db.Exec(query, taskID, fieldID)
	if err != nil {
		return fmt.Errorf("failed to delete custom field value: %w", err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *sqlCustomFieldRepository) FindTasksByFieldValue(filter model.CustomFieldValueFilter) ([]int64, int, error) {
	return nil, 0, nil
}
