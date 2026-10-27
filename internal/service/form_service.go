package service

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/JohnMalugu/tsk-mgr-api/internal/model"
	"github.com/JohnMalugu/tsk-mgr-api/internal/repository"
)

var (
	ErrFormNotFound           = errors.New("form not found")
	ErrFormTitleRequired      = errors.New("form title is required")
	ErrFormWorkspaceRequired  = errors.New("workspace ID is required")
	ErrFormInvalidSlug        = errors.New("form slug is invalid")
	ErrFormSlugAlreadyExists  = errors.New("form slug already exists")
	ErrFormNotPublished       = errors.New("form is not published")
	ErrFormAnonymousForbidden = errors.New("anonymous submissions are not permitted for this form")
	ErrFormRequiredField      = errors.New("required field is missing or empty")
	ErrFormInvalidFieldType   = errors.New("invalid form field type")
	ErrFormInvalidTargetField = errors.New("invalid target task field")
	ErrFormInvalidOption      = errors.New("selected option is invalid")
	ErrFormSubmissionFailed   = errors.New("failed to process form submission")
)

// FormService orchestrates intake forms, field blueprints, submission processing, and intake analytics
type FormService interface {
	CreateForm(userID string, req model.CreateFormRequest) (*model.IntakeForm, error)
	GetForm(id int64) (*model.IntakeForm, error)
	GetFormBySlug(slug string) (*model.IntakeForm, error)
	GetPublicFormView(slug string) (*model.PublicFormView, error)
	ListForms(filter model.FormFilter) ([]*model.FormSummary, error)
	UpdateForm(id int64, req model.UpdateFormRequest) (*model.IntakeForm, error)
	DeleteForm(id int64) error
	PublishForm(id int64, publish bool) (*model.IntakeForm, error)

	SetFormFields(formID int64, req model.SetFormFieldsRequest) ([]model.FormField, error)
	GetFormFields(formID int64) ([]model.FormField, error)

	SubmitForm(slug string, clientIP, userAgent string, req model.SubmitFormRequest) (*model.SubmissionResult, error)
	GetFormSubmissions(formID int64, limit, offset int) ([]*model.FormSubmission, error)
	GetFormAnalytics(formID int64) (*model.FormAnalytics, error)
}

type formServiceImpl struct {
	formRepo repository.FormRepository
	taskRepo repository.TaskRepository
}

// NewFormService instantiates a new FormService
func NewFormService(formRepo repository.FormRepository, taskRepo repository.TaskRepository) FormService {
	return &formServiceImpl{
		formRepo: formRepo,
		taskRepo: taskRepo,
	}
}

func normalizeSlug(raw string) string {
	raw = strings.ToLower(strings.TrimSpace(raw))
	if raw == "" {
		return ""
	}

	var sb strings.Builder
	lastDash := false
	for _, r := range raw {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			sb.WriteRune(r)
			lastDash = false
		} else if r == '-' || r == '_' || r == ' ' || r == '.' {
			if !lastDash && sb.Len() > 0 {
				sb.WriteRune('-')
				lastDash = true
			}
		}
	}
	return strings.Trim(sb.String(), "-")
}

func deriveSlug(title string) string {
	slug := normalizeSlug(title)
	if slug == "" {
		return fmt.Sprintf("intake-%d", time.Now().UnixNano()%100000)
	}
	return slug
}

func validateFieldDefinitions(defs []model.FormFieldDefinition) ([]model.FormField, error) {
	if len(defs) == 0 {
		return []model.FormField{}, nil
	}

	seenKeys := make(map[string]bool)
	fields := make([]model.FormField, 0, len(defs))

	for i, def := range defs {
		key := strings.ToLower(strings.TrimSpace(def.FieldKey))
		if key == "" {
			return nil, fmt.Errorf("%w: field key cannot be empty", ErrFormRequiredField)
		}
		if seenKeys[key] {
			return nil, fmt.Errorf("duplicate field key '%s'", key)
		}
		seenKeys[key] = true

		if !model.IsValidFormFieldType(def.FieldType) {
			return nil, fmt.Errorf("%w: %s", ErrFormInvalidFieldType, def.FieldType)
		}

		target := strings.TrimSpace(def.TargetField)
		if target == "" {
			target = model.FieldTargetCustom
		}
		switch target {
		case model.FieldTargetTitle, model.FieldTargetDescription, model.FieldTargetPriority, model.FieldTargetDueDate, model.FieldTargetCustom:
			// valid
		default:
			return nil, fmt.Errorf("%w: %s", ErrFormInvalidTargetField, target)
		}

		label := strings.TrimSpace(def.Label)
		if label == "" {
			return nil, fmt.Errorf("%w: field label cannot be empty for key %s", ErrFormRequiredField, key)
		}

		if def.FieldType == model.FormFieldTypeDropdown && len(def.Options) == 0 {
			return nil, fmt.Errorf("dropdown field '%s' must have at least one option", key)
		}

		pos := def.Position
		if pos == 0 {
			pos = i + 1
		}

		fields = append(fields, model.FormField{
			FieldKey:     key,
			FieldType:    def.FieldType,
			TargetField:  target,
			Label:        label,
			Placeholder:  def.Placeholder,
			HelpText:     def.HelpText,
			IsRequired:   def.IsRequired,
			Position:     pos,
			Options:      def.Options,
			DefaultValue: def.DefaultValue,
		})
	}

	return fields, nil
}

func (s *formServiceImpl) CreateForm(userID string, req model.CreateFormRequest) (*model.IntakeForm, error) {
	return nil, nil
}

func (s *formServiceImpl) GetForm(id int64) (*model.IntakeForm, error) {
	return nil, nil
}

func (s *formServiceImpl) GetFormBySlug(slug string) (*model.IntakeForm, error) {
	return nil, nil
}

func (s *formServiceImpl) GetPublicFormView(slug string) (*model.PublicFormView, error) {
	return nil, nil
}

func (s *formServiceImpl) ListForms(filter model.FormFilter) ([]*model.FormSummary, error) {
	return nil, nil
}

func (s *formServiceImpl) UpdateForm(id int64, req model.UpdateFormRequest) (*model.IntakeForm, error) {
	return nil, nil
}

func (s *formServiceImpl) DeleteForm(id int64) error {
	return nil
}

func (s *formServiceImpl) PublishForm(id int64, publish bool) (*model.IntakeForm, error) {
	return nil, nil
}

func (s *formServiceImpl) SetFormFields(formID int64, req model.SetFormFieldsRequest) ([]model.FormField, error) {
	return nil, nil
}

func (s *formServiceImpl) GetFormFields(formID int64) ([]model.FormField, error) {
	return nil, nil
}

func (s *formServiceImpl) SubmitForm(slug string, clientIP, userAgent string, req model.SubmitFormRequest) (*model.SubmissionResult, error) {
	return nil, nil
}

func (s *formServiceImpl) GetFormSubmissions(formID int64, limit, offset int) ([]*model.FormSubmission, error) {
	return nil, nil
}

func (s *formServiceImpl) GetFormAnalytics(formID int64) (*model.FormAnalytics, error) {
	return nil, nil
}
