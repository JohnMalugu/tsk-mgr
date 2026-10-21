package model

import (
	"strings"
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
