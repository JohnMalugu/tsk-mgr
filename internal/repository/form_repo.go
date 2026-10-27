package repository

import (
	"database/sql"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/JohnMalugu/tsk-mgr-api/internal/model"
)

var (
	ErrFormNotFound      = errors.New("form not found")
	ErrDuplicateFormSlug = errors.New("form slug already exists")
	ErrFormFieldNotFound = errors.New("form field not found")
)

// FormRepository defines data access methods for intake forms, fields, and submissions
type FormRepository interface {
	CreateForm(form *model.IntakeForm) (int64, error)
	GetFormByID(id int64) (*model.IntakeForm, error)
	GetFormBySlug(slug string) (*model.IntakeForm, error)
	ListForms(filter model.FormFilter) ([]*model.FormSummary, error)
	UpdateForm(form *model.IntakeForm) error
	ToggleFormPublished(id int64, isPublished bool, status string) error
	DeleteForm(id int64) error

	SetFormFields(formID int64, fields []model.FormField) error
	GetFormFields(formID int64) ([]model.FormField, error)
	DeleteFormField(formID int64, fieldKey string) error

	RecordSubmission(sub *model.FormSubmission) (int64, error)
	ListSubmissions(formID int64, limit, offset int) ([]*model.FormSubmission, error)
	GetFormAnalytics(formID int64) (*model.FormAnalytics, error)
}

type formRepositoryImpl struct {
	db *sql.DB
}

// NewFormRepository constructs a SQLite-backed FormRepository
func NewFormRepository(db *sql.DB) FormRepository {
	return &formRepositoryImpl{db: db}
}

func scanForm(scanner interface{ Scan(dest ...interface{}) error }) (*model.IntakeForm, error) {
	var (
		form            model.IntakeForm
		projectID       sql.NullInt64
		defaultAssignee sql.NullString
		defaultTagsJSON string
		isPublished     int
		allowAnon       int
	)

	err := scanner.Scan(
		&form.ID,
		&form.WorkspaceID,
		&projectID,
		&form.Title,
		&form.Description,
		&form.Slug,
		&form.Status,
		&isPublished,
		&allowAnon,
		&form.SubmitButtonText,
		&form.SuccessMessage,
		&form.RedirectURL,
		&form.DefaultPriority,
		&defaultTagsJSON,
		&defaultAssignee,
		&form.SubmissionCount,
		&form.CreatedByUserID,
		&form.CreatedAt,
		&form.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrFormNotFound
		}
		return nil, err
	}

	if projectID.Valid {
		form.ProjectID = &projectID.Int64
	}
	if defaultAssignee.Valid {
		form.DefaultAssigneeID = defaultAssignee.String
	}
	form.IsPublished = isPublished == 1
	form.AllowAnonymous = allowAnon == 1

	if defaultTagsJSON != "" {
		_ = json.Unmarshal([]byte(defaultTagsJSON), &form.DefaultTags)
	}
	if form.DefaultTags == nil {
		form.DefaultTags = []string{}
	}

	return &form, nil
}

func (r *formRepositoryImpl) CreateForm(form *model.IntakeForm) (int64, error) {
	now := time.Now()
	form.CreatedAt = now
	form.UpdatedAt = now
	if form.Status == "" {
		form.Status = model.FormStatusDraft
	}
	if form.SubmitButtonText == "" {
		form.SubmitButtonText = "Submit"
	}
	if form.SuccessMessage == "" {
		form.SuccessMessage = "Thank you! Your submission has been received."
	}
	if form.DefaultPriority == "" {
		form.DefaultPriority = "medium"
	}
	if form.DefaultTags == nil {
		form.DefaultTags = []string{}
	}

	tagsJSON, err := json.Marshal(form.DefaultTags)
	if err != nil {
		return 0, err
	}

	isPublished := 0
	if form.IsPublished {
		isPublished = 1
	}
	allowAnon := 1
	if !form.AllowAnonymous {
		allowAnon = 0
	}

	var projectID sql.NullInt64
	if form.ProjectID != nil {
		projectID = sql.NullInt64{Int64: *form.ProjectID, Valid: true}
	}
	var assigneeID sql.NullString
	if form.DefaultAssigneeID != "" {
		assigneeID = sql.NullString{String: form.DefaultAssigneeID, Valid: true}
	}

	query := `
	INSERT INTO intake_forms (
		workspace_id, project_id, title, description, slug, status,
		is_published, allow_anonymous, submit_button_text, success_message,
		redirect_url, default_priority, default_tags, default_assignee_id,
		submission_count, created_by_user_id, created_at, updated_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	res, err := r.db.Exec(query,
		form.WorkspaceID, projectID, form.Title, form.Description, form.Slug, form.Status,
		isPublished, allowAnon, form.SubmitButtonText, form.SuccessMessage,
		form.RedirectURL, form.DefaultPriority, string(tagsJSON), assigneeID,
		form.SubmissionCount, form.CreatedByUserID, form.CreatedAt, form.UpdatedAt,
	)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed: intake_forms.slug") {
			return 0, ErrDuplicateFormSlug
		}
		return 0, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	form.ID = id
	return id, nil
}

func (r *formRepositoryImpl) GetFormByID(id int64) (*model.IntakeForm, error) {
	query := `
	SELECT id, workspace_id, project_id, title, description, slug, status,
		is_published, allow_anonymous, submit_button_text, success_message,
		redirect_url, default_priority, default_tags, default_assignee_id,
		submission_count, created_by_user_id, created_at, updated_at
	FROM intake_forms
	WHERE id = ?`

	return scanForm(r.db.QueryRow(query, id))
}

func (r *formRepositoryImpl) GetFormBySlug(slug string) (*model.IntakeForm, error) {
	query := `
	SELECT id, workspace_id, project_id, title, description, slug, status,
		is_published, allow_anonymous, submit_button_text, success_message,
		redirect_url, default_priority, default_tags, default_assignee_id,
		submission_count, created_by_user_id, created_at, updated_at
	FROM intake_forms
	WHERE slug = ?`

	return scanForm(r.db.QueryRow(query, slug))
}


func (r *formRepositoryImpl) ListForms(filter model.FormFilter) ([]*model.FormSummary, error) {
	return nil, nil
}

func (r *formRepositoryImpl) UpdateForm(form *model.IntakeForm) error {
	return nil
}

func (r *formRepositoryImpl) ToggleFormPublished(id int64, isPublished bool, status string) error {
	return nil
}

func (r *formRepositoryImpl) DeleteForm(id int64) error {
	return nil
}

func (r *formRepositoryImpl) SetFormFields(formID int64, fields []model.FormField) error {
	return nil
}

func (r *formRepositoryImpl) GetFormFields(formID int64) ([]model.FormField, error) {
	return nil, nil
}

func (r *formRepositoryImpl) DeleteFormField(formID int64, fieldKey string) error {
	return nil
}

func (r *formRepositoryImpl) RecordSubmission(sub *model.FormSubmission) (int64, error) {
	return 0, nil
}

func (r *formRepositoryImpl) ListSubmissions(formID int64, limit, offset int) ([]*model.FormSubmission, error) {
	return nil, nil
}

func (r *formRepositoryImpl) GetFormAnalytics(formID int64) (*model.FormAnalytics, error) {
	return nil, nil
}
