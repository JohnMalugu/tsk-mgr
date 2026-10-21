package model

import (
	"strings"
	"time"
)

// CustomFieldType represents the data type for a custom field definition
type CustomFieldType string

const (
	CustomFieldTypeText     CustomFieldType = "text"
	CustomFieldTypeNumber   CustomFieldType = "number"
	CustomFieldTypeDropdown CustomFieldType = "dropdown"
	CustomFieldTypeBoolean  CustomFieldType = "boolean"
	CustomFieldTypeDate     CustomFieldType = "date"
	CustomFieldTypeURL      CustomFieldType = "url"
)

// ValidCustomFieldTypes defines all supported custom field types
var ValidCustomFieldTypes = []CustomFieldType{
	CustomFieldTypeText,
	CustomFieldTypeNumber,
	CustomFieldTypeDropdown,
	CustomFieldTypeBoolean,
	CustomFieldTypeDate,
	CustomFieldTypeURL,
}

// IsValidCustomFieldType checks if a given string corresponds to a valid CustomFieldType
func IsValidCustomFieldType(t string) bool {
	clean := strings.ToLower(strings.TrimSpace(t))
	for _, valid := range ValidCustomFieldTypes {
		if string(valid) == clean {
			return true
		}
	}
	return false
}

// CustomFieldDefinition represents a custom attribute schema scoped to a workspace and optionally a project
type CustomFieldDefinition struct {
	ID           int64           `json:"id"`
	WorkspaceID  int64           `json:"workspace_id"`
	ProjectID    *int64          `json:"project_id,omitempty"`
	Name         string          `json:"name"`
	FieldKey     string          `json:"field_key"`
	Description  string          `json:"description,omitempty"`
	Type         CustomFieldType `json:"type"`
	Required     bool            `json:"required"`
	Options      []string        `json:"options,omitempty"`
	DefaultValue *string         `json:"default_value,omitempty"`
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`
}

// CustomFieldValue represents the stored typed value for a custom field attached to a task
type CustomFieldValue struct {
	ID           int64      `json:"id"`
	FieldID      int64      `json:"field_id"`
	TaskID       int64      `json:"task_id"`
	ValueText    *string    `json:"value_text,omitempty"`
	ValueNumber  *float64   `json:"value_number,omitempty"`
	ValueBoolean *bool      `json:"value_boolean,omitempty"`
	ValueDate    *time.Time `json:"value_date,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

// TaskCustomFieldValueDTO represents an enriched custom field value with full field metadata for API responses
type TaskCustomFieldValueDTO struct {
	FieldID      int64           `json:"field_id"`
	FieldName    string          `json:"field_name"`
	FieldKey     string          `json:"field_key"`
	Type         CustomFieldType `json:"type"`
	Required     bool            `json:"required"`
	Options      []string        `json:"options,omitempty"`
	Value        interface{}     `json:"value"`
	ValueText    *string         `json:"value_text,omitempty"`
	ValueNumber  *float64        `json:"value_number,omitempty"`
	ValueBoolean *bool           `json:"value_boolean,omitempty"`
	ValueDate    *time.Time      `json:"value_date,omitempty"`
	UpdatedAt    time.Time       `json:"updated_at"`
}

// CreateCustomFieldRequest represents the payload to create a new custom field definition
type CreateCustomFieldRequest struct {
	WorkspaceID  int64           `json:"workspace_id"`
	ProjectID    *int64          `json:"project_id,omitempty"`
	Name         string          `json:"name"`
	FieldKey     string          `json:"field_key,omitempty"`
	Description  string          `json:"description,omitempty"`
	Type         CustomFieldType `json:"type"`
	Required     bool            `json:"required"`
	Options      []string        `json:"options,omitempty"`
	DefaultValue *string         `json:"default_value,omitempty"`
}

// UpdateCustomFieldRequest represents the payload to update an existing custom field definition
type UpdateCustomFieldRequest struct {
	Name         *string   `json:"name,omitempty"`
	Description  *string   `json:"description,omitempty"`
	Required     *bool     `json:"required,omitempty"`
	Options      *[]string `json:"options,omitempty"`
	DefaultValue *string   `json:"default_value,omitempty"`
}

// SetCustomFieldValueRequest represents payload to assign or update a custom field value on a task
type SetCustomFieldValueRequest struct {
	FieldID      int64       `json:"field_id,omitempty"`
	FieldKey     string      `json:"field_key,omitempty"`
	Value        interface{} `json:"value"`
	ValueText    *string     `json:"value_text,omitempty"`
	ValueNumber  *float64    `json:"value_number,omitempty"`
	ValueBoolean *bool       `json:"value_boolean,omitempty"`
	ValueDate    *time.Time  `json:"value_date,omitempty"`
}

// BulkSetCustomFieldItem represents a single field assignment within a bulk update
type BulkSetCustomFieldItem struct {
	TaskID int64                      `json:"task_id"`
	Fields []SetCustomFieldValueRequest `json:"fields"`
}

// BulkSetCustomFieldRequest represents the payload to set custom fields across multiple tasks
type BulkSetCustomFieldRequest struct {
	Assignments []BulkSetCustomFieldItem `json:"assignments"`
}

// CustomFieldFilter provides search and pagination criteria for field definitions
type CustomFieldFilter struct {
	WorkspaceID int64            `json:"workspace_id"`
	ProjectID   *int64           `json:"project_id,omitempty"`
	Type        *CustomFieldType `json:"type,omitempty"`
	Required    *bool            `json:"required,omitempty"`
	Offset      int              `json:"offset"`
	Limit       int              `json:"limit"`
}

// CustomFieldValueFilter provides criteria to search for tasks by specific custom field value
type CustomFieldValueFilter struct {
	WorkspaceID int64  `json:"workspace_id"`
	FieldID     int64  `json:"field_id,omitempty"`
	FieldKey    string `json:"field_key,omitempty"`
	QueryValue  string `json:"query_value"`
	Offset      int    `json:"offset"`
	Limit       int    `json:"limit"`
}

// CustomFieldSummary represents a paginated collection of field definitions
type CustomFieldSummary struct {
	Definitions []CustomFieldDefinition `json:"definitions"`
	Total       int                     `json:"total"`
	Offset      int                     `json:"offset"`
	Limit       int                     `json:"limit"`
}
