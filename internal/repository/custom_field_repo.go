package repository

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
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
	return nil, nil
}

func (r *sqlCustomFieldRepository) ListFieldDefinitions(filter model.CustomFieldFilter) ([]model.CustomFieldDefinition, int, error) {
	return nil, 0, nil
}

func (r *sqlCustomFieldRepository) UpdateFieldDefinition(def *model.CustomFieldDefinition) error {
	return nil
}

func (r *sqlCustomFieldRepository) DeleteFieldDefinition(id int64) error {
	return nil
}

func (r *sqlCustomFieldRepository) SetFieldValue(val *model.CustomFieldValue) error {
	return nil
}

func (r *sqlCustomFieldRepository) GetFieldValueByTaskAndField(taskID, fieldID int64) (*model.CustomFieldValue, error) {
	return nil, nil
}

func (r *sqlCustomFieldRepository) GetFieldValuesByTask(taskID int64) ([]model.TaskCustomFieldValueDTO, error) {
	return nil, nil
}

func (r *sqlCustomFieldRepository) DeleteFieldValue(taskID, fieldID int64) error {
	return nil
}

func (r *sqlCustomFieldRepository) FindTasksByFieldValue(filter model.CustomFieldValueFilter) ([]int64, int, error) {
	return nil, 0, nil
}
