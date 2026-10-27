package repository

import (
	"database/sql"
	"errors"

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

func (r *formRepositoryImpl) CreateForm(form *model.IntakeForm) (int64, error) {
	return 0, nil
}

func (r *formRepositoryImpl) GetFormByID(id int64) (*model.IntakeForm, error) {
	return nil, nil
}

func (r *formRepositoryImpl) GetFormBySlug(slug string) (*model.IntakeForm, error) {
	return nil, nil
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
