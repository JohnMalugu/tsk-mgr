package service

import (
	"errors"
	"fmt"
	"net/url"
	"strconv"
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
	def, err := s.repo.GetFieldDefinitionByID(id)
	if err != nil {
		return nil, err
	}
	if def == nil {
		return nil, ErrCustomFieldNotFound
	}
	return def, nil
}

func (s *customFieldService) GetFieldDefinitionByKey(workspaceID int64, key string) (*model.CustomFieldDefinition, error) {
	cleanKey := slugifyKey(key)
	if cleanKey == "" {
		return nil, ErrCustomFieldKeyInvalid
	}
	def, err := s.repo.GetFieldDefinitionByKey(workspaceID, cleanKey)
	if err != nil {
		return nil, err
	}
	if def == nil {
		return nil, ErrCustomFieldNotFound
	}
	return def, nil
}

func (s *customFieldService) ListFieldDefinitions(filter model.CustomFieldFilter) (*model.CustomFieldSummary, error) {
	if filter.Limit <= 0 {
		filter.Limit = 20
	}
	if filter.Limit > 100 {
		filter.Limit = 100
	}
	if filter.Offset < 0 {
		filter.Offset = 0
	}

	defs, total, err := s.repo.ListFieldDefinitions(filter)
	if err != nil {
		return nil, err
	}

	return &model.CustomFieldSummary{
		Definitions: defs,
		Total:       total,
		Offset:      filter.Offset,
		Limit:       filter.Limit,
	}, nil
}

func (s *customFieldService) UpdateFieldDefinition(id int64, req *model.UpdateCustomFieldRequest) (*model.CustomFieldDefinition, error) {
	if req == nil {
		return nil, errors.New("request cannot be nil")
	}

	def, err := s.repo.GetFieldDefinitionByID(id)
	if err != nil {
		return nil, err
	}
	if def == nil {
		return nil, ErrCustomFieldNotFound
	}

	if req.Name != nil {
		trimmedName := strings.TrimSpace(*req.Name)
		if trimmedName == "" {
			return nil, ErrCustomFieldNameRequired
		}
		def.Name = trimmedName
	}

	if req.Description != nil {
		def.Description = strings.TrimSpace(*req.Description)
	}

	if req.Required != nil {
		def.Required = *req.Required
	}

	if req.DefaultValue != nil {
		def.DefaultValue = req.DefaultValue
	}

	if req.Options != nil {
		var cleanedOptions []string
		for _, opt := range *req.Options {
			trimmed := strings.TrimSpace(opt)
			if trimmed != "" {
				cleanedOptions = append(cleanedOptions, trimmed)
			}
		}
		if def.Type == model.CustomFieldTypeDropdown && len(cleanedOptions) == 0 {
			return nil, ErrDropdownOptionsRequired
		}
		def.Options = cleanedOptions
	}

	def.UpdatedAt = time.Now()
	if err := s.repo.UpdateFieldDefinition(def); err != nil {
		return nil, err
	}

	return def, nil
}

func (s *customFieldService) DeleteFieldDefinition(id int64) error {
	def, err := s.repo.GetFieldDefinitionByID(id)
	if err != nil {
		return err
	}
	if def == nil {
		return ErrCustomFieldNotFound
	}

	return s.repo.DeleteFieldDefinition(id)
}

func validateAndExtractValue(def *model.CustomFieldDefinition, req *model.SetCustomFieldValueRequest) (*model.CustomFieldValue, error) {
	val := &model.CustomFieldValue{
		FieldID: def.ID,
	}

	if req.ValueText != nil {
		req.Value = *req.ValueText
	} else if req.ValueNumber != nil {
		req.Value = *req.ValueNumber
	} else if req.ValueBoolean != nil {
		req.Value = *req.ValueBoolean
	} else if req.ValueDate != nil {
		req.Value = *req.ValueDate
	}

	if req.Value == nil {
		if def.Required {
			return nil, ErrRequiredCustomFieldMissing
		}
		return val, nil
	}

	switch def.Type {
	case model.CustomFieldTypeText:
		s := fmt.Sprintf("%v", req.Value)
		s = strings.TrimSpace(s)
		if def.Required && s == "" {
			return nil, ErrRequiredCustomFieldMissing
		}
		val.ValueText = &s

	case model.CustomFieldTypeNumber:
		var num float64
		switch v := req.Value.(type) {
		case float64:
			num = v
		case float32:
			num = float64(v)
		case int:
			num = float64(v)
		case int64:
			num = float64(v)
		case string:
			parsed, err := strconv.ParseFloat(strings.TrimSpace(v), 64)
			if err != nil {
				return nil, ErrInvalidNumberValue
			}
			num = parsed
		default:
			return nil, ErrInvalidNumberValue
		}
		val.ValueNumber = &num

	case model.CustomFieldTypeDropdown:
		s := fmt.Sprintf("%v", req.Value)
		s = strings.TrimSpace(s)
		if def.Required && s == "" {
			return nil, ErrRequiredCustomFieldMissing
		}
		matched := false
		for _, opt := range def.Options {
			if strings.EqualFold(opt, s) {
				s = opt
				matched = true
				break
			}
		}
		if !matched && s != "" {
			return nil, ErrInvalidDropdownValue
		}
		val.ValueText = &s

	case model.CustomFieldTypeBoolean:
		var b bool
		switch v := req.Value.(type) {
		case bool:
			b = v
		case string:
			clean := strings.ToLower(strings.TrimSpace(v))
			if clean == "true" || clean == "1" || clean == "yes" {
				b = true
			} else if clean == "false" || clean == "0" || clean == "no" {
				b = false
			} else {
				return nil, ErrInvalidBooleanValue
			}
		default:
			return nil, ErrInvalidBooleanValue
		}
		val.ValueBoolean = &b

	case model.CustomFieldTypeDate:
		var d time.Time
		switch v := req.Value.(type) {
		case time.Time:
			d = v
		case string:
			clean := strings.TrimSpace(v)
			parsed, err := time.Parse(time.RFC3339, clean)
			if err != nil {
				parsedDateOnly, err2 := time.Parse("2006-01-02", clean)
				if err2 != nil {
					return nil, ErrInvalidDateValue
				}
				parsed = parsedDateOnly
			}
			d = parsed
		default:
			return nil, ErrInvalidDateValue
		}
		val.ValueDate = &d

	case model.CustomFieldTypeURL:
		s := fmt.Sprintf("%v", req.Value)
		s = strings.TrimSpace(s)
		if def.Required && s == "" {
			return nil, ErrRequiredCustomFieldMissing
		}
		if s != "" {
			u, err := url.ParseRequestURI(s)
			if err != nil || u.Scheme == "" || u.Host == "" {
				return nil, ErrInvalidURLValue
			}
		}
		val.ValueText = &s
	}

	return val, nil
}

func (s *customFieldService) SetFieldValue(taskID int64, req *model.SetCustomFieldValueRequest) (*model.CustomFieldValue, error) {
	if req == nil {
		return nil, errors.New("request cannot be nil")
	}

	if req.FieldID <= 0 {
		return nil, ErrCustomFieldNotFound
	}

	def, err := s.repo.GetFieldDefinitionByID(req.FieldID)
	if err != nil {
		return nil, err
	}
	if def == nil {
		return nil, ErrCustomFieldNotFound
	}

	val, err := validateAndExtractValue(def, req)
	if err != nil {
		return nil, err
	}
	val.TaskID = taskID

	if err := s.repo.SetFieldValue(val); err != nil {
		return nil, err
	}

	return val, nil
}

func (s *customFieldService) GetTaskCustomFields(taskID int64) ([]model.TaskCustomFieldValueDTO, error) {
	return s.repo.GetFieldValuesByTask(taskID)
}

func (s *customFieldService) DeleteFieldValue(taskID, fieldID int64) error {
	def, err := s.repo.GetFieldDefinitionByID(fieldID)
	if err != nil {
		return err
	}
	if def == nil {
		return ErrCustomFieldNotFound
	}

	return s.repo.DeleteFieldValue(taskID, fieldID)
}

func (s *customFieldService) BulkSetCustomFields(req *model.BulkSetCustomFieldRequest) error {
	if req == nil {
		return errors.New("request cannot be nil")
	}

	for _, item := range req.Assignments {
		for _, f := range item.Fields {
			fCopy := f
			if _, err := s.SetFieldValue(item.TaskID, &fCopy); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *customFieldService) FindTasksByFieldValue(filter model.CustomFieldValueFilter) ([]int64, int, error) {
	if filter.Limit <= 0 {
		filter.Limit = 20
	}
	if filter.Limit > 100 {
		filter.Limit = 100
	}
	if filter.Offset < 0 {
		filter.Offset = 0
	}

	return s.repo.FindTasksByFieldValue(filter)
}
