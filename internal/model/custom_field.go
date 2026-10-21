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
