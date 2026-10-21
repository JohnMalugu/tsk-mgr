package service

import (
	"errors"
	"strings"
	"time"

	"github.com/JohnMalugu/tsk-mgr-api/internal/model"
	"github.com/JohnMalugu/tsk-mgr-api/internal/repository"
)

var (
	ErrCustomFieldNotFound        = errors.New("custom field definition not found")
	ErrCustomFieldNameRequired    = errors.New("custom field name is required")
	ErrCustomFieldKeyInvalid      = errors.New("invalid custom field key")
	ErrCustomFieldKeyTaken        = errors.New("custom field key already exists in workspace")
	ErrInvalidCustomFieldType     = errors.New("invalid custom field type")
	ErrDropdownOptionsRequired    = errors.New("dropdown fields must define at least one option")
	ErrInvalidDropdownValue       = errors.New("value must be one of the configured dropdown options")
	ErrInvalidNumberValue         = errors.New("value must be a valid number")
	ErrInvalidBooleanValue        = errors.New("value must be a valid boolean")
	ErrInvalidDateValue           = errors.New("value must be a valid RFC3339 date")
	ErrInvalidURLValue            = errors.New("value must be a valid URL")
	ErrRequiredCustomFieldMissing = errors.New("custom field is required")
)

type CustomFieldService interface {
	CreateFieldDefinition(req *model.CreateCustomFieldRequest) (*model.CustomFieldDefinition, error)
	GetFieldDefinition(id int64) (*model.CustomFieldDefinition, error)
	GetFieldDefinitionByKey(workspaceID int64, key string) (*model.CustomFieldDefinition, error)
	ListFieldDefinitions(filter model.CustomFieldFilter) (*model.CustomFieldSummary, error)
	UpdateFieldDefinition(id int64, req *model.UpdateCustomFieldRequest) (*model.CustomFieldDefinition, error)
	DeleteFieldDefinition(id int64) error

	SetFieldValue(taskID int64, req *model.SetCustomFieldValueRequest) (*model.CustomFieldValue, error)
	GetTaskCustomFields(taskID int64) ([]model.TaskCustomFieldValueDTO, error)
	DeleteFieldValue(taskID, fieldID int64) error
	BulkSetCustomFields(req *model.BulkSetCustomFieldRequest) error
	FindTasksByFieldValue(filter model.CustomFieldValueFilter) ([]int64, int, error)
}

type customFieldService struct {
	repo repository.CustomFieldRepository
}

func NewCustomFieldService(repo repository.CustomFieldRepository) CustomFieldService {
	return &customFieldService{repo: repo}
}

func slugifyKey(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	var sb strings.Builder
	lastUnderscore := false
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			sb.WriteRune(r)
			lastUnderscore = false
		} else if r == ' ' || r == '-' || r == '_' {
			if !lastUnderscore && sb.Len() > 0 {
				sb.WriteRune('_')
				lastUnderscore = true
			}
		}
	}
	return strings.Trim(sb.String(), "_")
}

func (s *customFieldService) CreateFieldDefinition(req *model.CreateCustomFieldRequest) (*model.CustomFieldDefinition, error) {
	if req == nil {
		return nil, errors.New("request cannot be nil")
	}

	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, ErrCustomFieldNameRequired
	}

	if !model.IsValidCustomFieldType(string(req.Type)) {
		return nil, ErrInvalidCustomFieldType
	}

	var cleanedOptions []string
	if req.Type == model.CustomFieldTypeDropdown {
		for _, opt := range req.Options {
			trimmed := strings.TrimSpace(opt)
			if trimmed != "" {
				cleanedOptions = append(cleanedOptions, trimmed)
			}
		}
		if len(cleanedOptions) == 0 {
			return nil, ErrDropdownOptionsRequired
		}
	}

	fieldKey := slugifyKey(req.FieldKey)
	if fieldKey == "" {
		fieldKey = slugifyKey(name)
	}
	if fieldKey == "" {
		return nil, ErrCustomFieldKeyInvalid
	}

	existing, err := s.repo.GetFieldDefinitionByKey(req.WorkspaceID, fieldKey)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, ErrCustomFieldKeyTaken
	}

	def := &model.CustomFieldDefinition{
		WorkspaceID:  req.WorkspaceID,
		ProjectID:    req.ProjectID,
		Name:         name,
		FieldKey:     fieldKey,
		Description:  strings.TrimSpace(req.Description),
		Type:         req.Type,
		Required:     req.Required,
		Options:      cleanedOptions,
		DefaultValue: req.DefaultValue,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	id, err := s.repo.CreateFieldDefinition(def)
	if err != nil {
		return nil, err
	}
	def.ID = id
	return def, nil
}

func (s *customFieldService) GetFieldDefinition(id int64) (*model.CustomFieldDefinition, error) {
	return nil, nil
}

func (s *customFieldService) GetFieldDefinitionByKey(workspaceID int64, key string) (*model.CustomFieldDefinition, error) {
	return nil, nil
}

func (s *customFieldService) ListFieldDefinitions(filter model.CustomFieldFilter) (*model.CustomFieldSummary, error) {
	return nil, nil
}

func (s *customFieldService) UpdateFieldDefinition(id int64, req *model.UpdateCustomFieldRequest) (*model.CustomFieldDefinition, error) {
	return nil, nil
}

func (s *customFieldService) DeleteFieldDefinition(id int64) error {
	return nil
}

func (s *customFieldService) SetFieldValue(taskID int64, req *model.SetCustomFieldValueRequest) (*model.CustomFieldValue, error) {
	return nil, nil
}

func (s *customFieldService) GetTaskCustomFields(taskID int64) ([]model.TaskCustomFieldValueDTO, error) {
	return nil, nil
}

func (s *customFieldService) DeleteFieldValue(taskID, fieldID int64) error {
	return nil
}

func (s *customFieldService) BulkSetCustomFields(req *model.BulkSetCustomFieldRequest) error {
	return nil
}

func (s *customFieldService) FindTasksByFieldValue(filter model.CustomFieldValueFilter) ([]int64, int, error) {
	return nil, 0, nil
}
