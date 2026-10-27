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
	if req.WorkspaceID <= 0 {
		return nil, ErrFormWorkspaceRequired
	}

	title := strings.TrimSpace(req.Title)
	if title == "" {
		return nil, ErrFormTitleRequired
	}

	slug := ""
	if req.Slug != "" {
		slug = normalizeSlug(req.Slug)
		if slug == "" {
			return nil, ErrFormInvalidSlug
		}
	} else {
		slug = deriveSlug(title)
	}

	allowAnon := true
	if req.AllowAnonymous != nil {
		allowAnon = *req.AllowAnonymous
	}

	btnText := strings.TrimSpace(req.SubmitButtonText)
	if btnText == "" {
		btnText = "Submit"
	}

	successMsg := strings.TrimSpace(req.SuccessMessage)
	if successMsg == "" {
		successMsg = "Thank you! Your submission has been received."
	}

	priority := strings.TrimSpace(req.DefaultPriority)
	if priority == "" {
		priority = "medium"
	}

	var fields []model.FormField
	if len(req.Fields) > 0 {
		var err error
		fields, err = validateFieldDefinitions(req.Fields)
		if err != nil {
			return nil, err
		}
	}

	form := &model.IntakeForm{
		WorkspaceID:       req.WorkspaceID,
		ProjectID:         req.ProjectID,
		Title:             title,
		Description:       strings.TrimSpace(req.Description),
		Slug:              slug,
		Status:            model.FormStatusDraft,
		IsPublished:       false,
		AllowAnonymous:    allowAnon,
		SubmitButtonText:  btnText,
		SuccessMessage:    successMsg,
		RedirectURL:       strings.TrimSpace(req.RedirectURL),
		DefaultPriority:   priority,
		DefaultTags:       req.DefaultTags,
		DefaultAssigneeID: strings.TrimSpace(req.DefaultAssigneeID),
		CreatedByUserID:   userID,
	}

	id, err := s.formRepo.CreateForm(form)
	if err != nil {
		if errors.Is(err, repository.ErrDuplicateFormSlug) {
			return nil, ErrFormSlugAlreadyExists
		}
		return nil, err
	}

	if len(fields) > 0 {
		if err := s.formRepo.SetFormFields(id, fields); err != nil {
			return nil, err
		}
	}

	return s.formRepo.GetFormByID(id)
}

func (s *formServiceImpl) GetForm(id int64) (*model.IntakeForm, error) {
	form, err := s.formRepo.GetFormByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrFormNotFound) {
			return nil, ErrFormNotFound
		}
		return nil, err
	}
	return form, nil
}

func (s *formServiceImpl) GetFormBySlug(slug string) (*model.IntakeForm, error) {
	norm := normalizeSlug(slug)
	if norm == "" {
		return nil, ErrFormInvalidSlug
	}
	form, err := s.formRepo.GetFormBySlug(norm)
	if err != nil {
		if errors.Is(err, repository.ErrFormNotFound) {
			return nil, ErrFormNotFound
		}
		return nil, err
	}
	return form, nil
}

func (s *formServiceImpl) GetPublicFormView(slug string) (*model.PublicFormView, error) {
	form, err := s.GetFormBySlug(slug)
	if err != nil {
		return nil, err
	}

	if !form.IsPublished || form.Status != model.FormStatusPublished {
		return nil, ErrFormNotPublished
	}

	fields, err := s.formRepo.GetFormFields(form.ID)
	if err != nil {
		return nil, err
	}

	return &model.PublicFormView{
		Title:            form.Title,
		Description:      form.Description,
		Slug:             form.Slug,
		SubmitButtonText: form.SubmitButtonText,
		AllowAnonymous:   form.AllowAnonymous,
		Fields:           fields,
	}, nil
}

func (s *formServiceImpl) ListForms(filter model.FormFilter) ([]*model.FormSummary, error) {
	return s.formRepo.ListForms(filter)
}

func (s *formServiceImpl) UpdateForm(id int64, req model.UpdateFormRequest) (*model.IntakeForm, error) {
	form, err := s.GetForm(id)
	if err != nil {
		return nil, err
	}

	if req.Title != nil {
		t := strings.TrimSpace(*req.Title)
		if t == "" {
			return nil, ErrFormTitleRequired
		}
		form.Title = t
	}
	if req.Description != nil {
		form.Description = strings.TrimSpace(*req.Description)
	}
	if req.Slug != nil {
		sl := normalizeSlug(*req.Slug)
		if sl == "" {
			return nil, ErrFormInvalidSlug
		}
		form.Slug = sl
	}
	if req.AllowAnonymous != nil {
		form.AllowAnonymous = *req.AllowAnonymous
	}
	if req.SubmitButtonText != nil {
		btn := strings.TrimSpace(*req.SubmitButtonText)
		if btn != "" {
			form.SubmitButtonText = btn
		}
	}
	if req.SuccessMessage != nil {
		msg := strings.TrimSpace(*req.SuccessMessage)
		if msg != "" {
			form.SuccessMessage = msg
		}
	}
	if req.RedirectURL != nil {
		form.RedirectURL = strings.TrimSpace(*req.RedirectURL)
	}
	if req.DefaultPriority != nil {
		p := strings.TrimSpace(*req.DefaultPriority)
		if p != "" {
			form.DefaultPriority = p
		}
	}
	if req.DefaultTags != nil {
		form.DefaultTags = *req.DefaultTags
	}
	if req.DefaultAssigneeID != nil {
		form.DefaultAssigneeID = strings.TrimSpace(*req.DefaultAssigneeID)
	}
	if req.ProjectID != nil {
		form.ProjectID = req.ProjectID
	}

	err = s.formRepo.UpdateForm(form)
	if err != nil {
		if errors.Is(err, repository.ErrDuplicateFormSlug) {
			return nil, ErrFormSlugAlreadyExists
		}
		if errors.Is(err, repository.ErrFormNotFound) {
			return nil, ErrFormNotFound
		}
		return nil, err
	}

	return s.formRepo.GetFormByID(id)
}

func (s *formServiceImpl) DeleteForm(id int64) error {
	err := s.formRepo.DeleteForm(id)
	if err != nil {
		if errors.Is(err, repository.ErrFormNotFound) {
			return ErrFormNotFound
		}
		return err
	}
	return nil
}

func (s *formServiceImpl) PublishForm(id int64, publish bool) (*model.IntakeForm, error) {
	_, err := s.GetForm(id)
	if err != nil {
		return nil, err
	}

	status := model.FormStatusPublished
	if !publish {
		status = model.FormStatusDraft
	}

	if err := s.formRepo.ToggleFormPublished(id, publish, status); err != nil {
		if errors.Is(err, repository.ErrFormNotFound) {
			return nil, ErrFormNotFound
		}
		return nil, err
	}

	return s.formRepo.GetFormByID(id)
}

func (s *formServiceImpl) SetFormFields(formID int64, req model.SetFormFieldsRequest) ([]model.FormField, error) {
	_, err := s.GetForm(formID)
	if err != nil {
		return nil, err
	}

	fields, err := validateFieldDefinitions(req.Fields)
	if err != nil {
		return nil, err
	}

	if err := s.formRepo.SetFormFields(formID, fields); err != nil {
		if errors.Is(err, repository.ErrFormNotFound) {
			return nil, ErrFormNotFound
		}
		return nil, err
	}

	return s.formRepo.GetFormFields(formID)
}

func (s *formServiceImpl) GetFormFields(formID int64) ([]model.FormField, error) {
	_, err := s.GetForm(formID)
	if err != nil {
		return nil, err
	}

	return s.formRepo.GetFormFields(formID)
}

func (s *formServiceImpl) SubmitForm(slug string, clientIP, userAgent string, req model.SubmitFormRequest) (*model.SubmissionResult, error) {
	form, err := s.GetFormBySlug(slug)
	if err != nil {
		return nil, err
	}

	if !form.IsPublished || form.Status != model.FormStatusPublished {
		return nil, ErrFormNotPublished
	}

	submitterEmail := strings.TrimSpace(req.SubmitterEmail)
	if !form.AllowAnonymous && submitterEmail == "" {
		return nil, ErrFormAnonymousForbidden
	}

	fields, err := s.formRepo.GetFormFields(form.ID)
	if err != nil {
		return nil, err
	}

	if req.Responses == nil {
		req.Responses = make(map[string]interface{})
	}

	taskTitle := ""
	taskDescription := ""
	taskPriority := form.DefaultPriority
	if taskPriority == "" {
		taskPriority = "medium"
	}
	taskDueDate := time.Now().AddDate(0, 0, 7)

	for _, field := range fields {
		val, exists := req.Responses[field.FieldKey]
		if field.IsRequired {
			if !exists || val == nil || strings.TrimSpace(fmt.Sprintf("%v", val)) == "" {
				return nil, fmt.Errorf("%w: %s (%s)", ErrFormRequiredField, field.Label, field.FieldKey)
			}
		}

		if exists && val != nil {
			strVal := strings.TrimSpace(fmt.Sprintf("%v", val))

			if field.FieldType == model.FormFieldTypeDropdown && strVal != "" {
				validOpt := false
				for _, opt := range field.Options {
					if strings.EqualFold(opt, strVal) {
						validOpt = true
						strVal = opt
						break
					}
				}
				if !validOpt {
					return nil, fmt.Errorf("%w: %s for field %s", ErrFormInvalidOption, strVal, field.FieldKey)
				}
			}

			switch field.TargetField {
			case model.FieldTargetTitle:
				if strVal != "" {
					taskTitle = strVal
				}
			case model.FieldTargetDescription:
				if strVal != "" {
					taskDescription = strVal
				}
			case model.FieldTargetPriority:
				if strVal != "" {
					taskPriority = strings.ToLower(strVal)
				}
			case model.FieldTargetDueDate:
				if strVal != "" {
					if parsed, err := time.Parse(time.RFC3339, strVal); err == nil {
						taskDueDate = parsed
					} else if parsed, err := time.Parse("2006-01-02", strVal); err == nil {
						taskDueDate = parsed
					}
				}
			}
		}
	}

	if taskTitle == "" {
		taskTitle = fmt.Sprintf("Form Submission: %s (#%d)", form.Title, form.SubmissionCount+1)
	}

	var descBuilder strings.Builder
	if taskDescription != "" {
		descBuilder.WriteString(taskDescription)
		descBuilder.WriteString("\n\n---\n")
	}
	descBuilder.WriteString("### Form Intake Details\n")
	descBuilder.WriteString(fmt.Sprintf("- **Form**: %s\n", form.Title))
	if submitterEmail != "" {
		descBuilder.WriteString(fmt.Sprintf("- **Submitted by**: %s\n", submitterEmail))
	}
	for _, field := range fields {
		if field.TargetField == model.FieldTargetCustom {
			if v, ok := req.Responses[field.FieldKey]; ok && v != nil {
				descBuilder.WriteString(fmt.Sprintf("- **%s**: %v\n", field.Label, v))
			}
		}
	}
	taskDescription = descBuilder.String()

	taskUserID := form.CreatedByUserID
	if form.DefaultAssigneeID != "" {
		taskUserID = form.DefaultAssigneeID
	}

	now := time.Now()
	task := &model.Task{
		UserID:      taskUserID,
		Title:       taskTitle,
		Description: taskDescription,
		CreatedAt:   now,
		UpdatedAt:   now,
		DueDate:     taskDueDate,
		Completed:   false,
		Status:      model.TaskStatusTodo,
		Priority:    taskPriority,
		Tags:        form.DefaultTags,
	}
	if form.ProjectID != nil {
		pid := int(*form.ProjectID)
		task.ProjectID = &pid
	}

	if err := s.taskRepo.Create(task); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrFormSubmissionFailed, err)
	}

	sub := &model.FormSubmission{
		FormID:         form.ID,
		TaskID:         &task.ID,
		SubmitterEmail: submitterEmail,
		SubmitterIP:    clientIP,
		UserAgent:      userAgent,
		Responses:      req.Responses,
		SubmittedAt:    now,
	}

	subID, err := s.formRepo.RecordSubmission(sub)
	if err != nil {
		return nil, err
	}

	return &model.SubmissionResult{
		SubmissionID:   subID,
		FormID:         form.ID,
		TaskID:         task.ID,
		TaskTitle:      task.Title,
		SuccessMessage: form.SuccessMessage,
		RedirectURL:    form.RedirectURL,
	}, nil
}

func (s *formServiceImpl) GetFormSubmissions(formID int64, limit, offset int) ([]*model.FormSubmission, error) {
	_, err := s.GetForm(formID)
	if err != nil {
		return nil, err
	}

	return s.formRepo.ListSubmissions(formID, limit, offset)
}

func (s *formServiceImpl) GetFormAnalytics(formID int64) (*model.FormAnalytics, error) {
	_, err := s.GetForm(formID)
	if err != nil {
		return nil, err
	}

	return s.formRepo.GetFormAnalytics(formID)
}
