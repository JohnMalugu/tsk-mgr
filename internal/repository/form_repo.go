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
	query := `
	SELECT 
		f.id, f.workspace_id, f.project_id, f.title, f.slug, f.status, f.is_published,
		COALESCE((SELECT COUNT(*) FROM intake_form_fields iff WHERE iff.form_id = f.id), 0) AS field_count,
		f.submission_count, f.created_at, f.updated_at
	FROM intake_forms f
	WHERE 1=1`

	var args []interface{}

	if filter.WorkspaceID != nil {
		query += " AND f.workspace_id = ?"
		args = append(args, *filter.WorkspaceID)
	}
	if filter.ProjectID != nil {
		query += " AND f.project_id = ?"
		args = append(args, *filter.ProjectID)
	}
	if filter.Status != nil && *filter.Status != "" {
		query += " AND f.status = ?"
		args = append(args, *filter.Status)
	}
	if filter.IsPublished != nil {
		pubVal := 0
		if *filter.IsPublished {
			pubVal = 1
		}
		query += " AND f.is_published = ?"
		args = append(args, pubVal)
	}
	if filter.Search != "" {
		query += " AND (f.title LIKE ? OR f.slug LIKE ? OR f.description LIKE ?)"
		pattern := "%" + filter.Search + "%"
		args = append(args, pattern, pattern, pattern)
	}

	query += " ORDER BY f.created_at DESC"

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	summaries := make([]*model.FormSummary, 0)
	for rows.Next() {
		var (
			s           model.FormSummary
			projectID   sql.NullInt64
			isPublished int
		)

		err := rows.Scan(
			&s.ID,
			&s.WorkspaceID,
			&projectID,
			&s.Title,
			&s.Slug,
			&s.Status,
			&isPublished,
			&s.FieldCount,
			&s.SubmissionCount,
			&s.CreatedAt,
			&s.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		if projectID.Valid {
			s.ProjectID = &projectID.Int64
		}
		s.IsPublished = isPublished == 1
		summaries = append(summaries, &s)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return summaries, nil
}

func (r *formRepositoryImpl) UpdateForm(form *model.IntakeForm) error {
	form.UpdatedAt = time.Now()

	tagsJSON, err := json.Marshal(form.DefaultTags)
	if err != nil {
		return err
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
	UPDATE intake_forms
	SET title = ?, description = ?, slug = ?, status = ?,
		is_published = ?, allow_anonymous = ?, submit_button_text = ?,
		success_message = ?, redirect_url = ?, default_priority = ?,
		default_tags = ?, default_assignee_id = ?, project_id = ?, updated_at = ?
	WHERE id = ?`

	res, err := r.db.Exec(query,
		form.Title, form.Description, form.Slug, form.Status,
		isPublished, allowAnon, form.SubmitButtonText,
		form.SuccessMessage, form.RedirectURL, form.DefaultPriority,
		string(tagsJSON), assigneeID, projectID, form.UpdatedAt,
		form.ID,
	)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed: intake_forms.slug") {
			return ErrDuplicateFormSlug
		}
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrFormNotFound
	}

	return nil
}

func (r *formRepositoryImpl) ToggleFormPublished(id int64, isPublished bool, status string) error {
	pubVal := 0
	if isPublished {
		pubVal = 1
	}

	query := `
	UPDATE intake_forms
	SET is_published = ?, status = ?, updated_at = ?
	WHERE id = ?`

	res, err := r.db.Exec(query, pubVal, status, time.Now(), id)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrFormNotFound
	}

	return nil
}

func (r *formRepositoryImpl) DeleteForm(id int64) error {
	query := `DELETE FROM intake_forms WHERE id = ?`
	res, err := r.db.Exec(query, id)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrFormNotFound
	}

	return nil
}


func (r *formRepositoryImpl) SetFormFields(formID int64, fields []model.FormField) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var exists int
	err = tx.QueryRow("SELECT 1 FROM intake_forms WHERE id = ?", formID).Scan(&exists)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrFormNotFound
		}
		return err
	}

	_, err = tx.Exec("DELETE FROM intake_form_fields WHERE form_id = ?", formID)
	if err != nil {
		return err
	}

	insertQuery := `
	INSERT INTO intake_form_fields (
		form_id, field_key, field_type, target_field, label,
		placeholder, help_text, is_required, position, options, default_value
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	for i, f := range fields {
		pos := f.Position
		if pos == 0 {
			pos = i + 1
		}
		if f.Options == nil {
			f.Options = []string{}
		}
		optsJSON, err := json.Marshal(f.Options)
		if err != nil {
			return err
		}

		reqVal := 0
		if f.IsRequired {
			reqVal = 1
		}

		_, err = tx.Exec(insertQuery,
			formID, f.FieldKey, f.FieldType, f.TargetField, f.Label,
			f.Placeholder, f.HelpText, reqVal, pos, string(optsJSON), f.DefaultValue,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *formRepositoryImpl) GetFormFields(formID int64) ([]model.FormField, error) {
	query := `
	SELECT id, form_id, field_key, field_type, target_field, label,
		placeholder, help_text, is_required, position, options, default_value
	FROM intake_form_fields
	WHERE form_id = ?
	ORDER BY position ASC, id ASC`

	rows, err := r.db.Query(query, formID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	fields := make([]model.FormField, 0)
	for rows.Next() {
		var (
			f        model.FormField
			reqVal   int
			optsJSON string
		)

		err := rows.Scan(
			&f.ID,
			&f.FormID,
			&f.FieldKey,
			&f.FieldType,
			&f.TargetField,
			&f.Label,
			&f.Placeholder,
			&f.HelpText,
			&reqVal,
			&f.Position,
			&optsJSON,
			&f.DefaultValue,
		)
		if err != nil {
			return nil, err
		}

		f.IsRequired = reqVal == 1
		if optsJSON != "" {
			_ = json.Unmarshal([]byte(optsJSON), &f.Options)
		}
		if f.Options == nil {
			f.Options = []string{}
		}

		fields = append(fields, f)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return fields, nil
}

func (r *formRepositoryImpl) DeleteFormField(formID int64, fieldKey string) error {
	query := `DELETE FROM intake_form_fields WHERE form_id = ? AND field_key = ?`
	res, err := r.db.Exec(query, formID, fieldKey)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrFormFieldNotFound
	}

	return nil
}


func (r *formRepositoryImpl) RecordSubmission(sub *model.FormSubmission) (int64, error) {
	if sub.SubmittedAt.IsZero() {
		sub.SubmittedAt = time.Now()
	}

	if sub.Responses == nil {
		sub.Responses = make(map[string]interface{})
	}
	respJSON, err := json.Marshal(sub.Responses)
	if err != nil {
		return 0, err
	}

	tx, err := r.db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	var taskIDVal sql.NullInt64
	if sub.TaskID != nil {
		taskIDVal = sql.NullInt64{Int64: int64(*sub.TaskID), Valid: true}
	}

	insertQuery := `
	INSERT INTO intake_form_submissions (
		form_id, task_id, submitter_email, submitter_ip, user_agent, responses, submitted_at
	) VALUES (?, ?, ?, ?, ?, ?, ?)`

	res, err := tx.Exec(insertQuery,
		sub.FormID, taskIDVal, sub.SubmitterEmail, sub.SubmitterIP, sub.UserAgent,
		string(respJSON), sub.SubmittedAt,
	)
	if err != nil {
		return 0, err
	}

	subID, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	sub.ID = subID

	_, err = tx.Exec("UPDATE intake_forms SET submission_count = submission_count + 1 WHERE id = ?", sub.FormID)
	if err != nil {
		return 0, err
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}

	return subID, nil
}

func (r *formRepositoryImpl) ListSubmissions(formID int64, limit, offset int) ([]*model.FormSubmission, error) {
	if limit <= 0 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	query := `
	SELECT id, form_id, task_id, submitter_email, submitter_ip, user_agent, responses, submitted_at
	FROM intake_form_submissions
	WHERE form_id = ?
	ORDER BY submitted_at DESC
	LIMIT ? OFFSET ?`

	rows, err := r.db.Query(query, formID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	submissions := make([]*model.FormSubmission, 0)
	for rows.Next() {
		var (
			sub      model.FormSubmission
			taskID   sql.NullInt64
			respJSON string
		)

		err := rows.Scan(
			&sub.ID,
			&sub.FormID,
			&taskID,
			&sub.SubmitterEmail,
			&sub.SubmitterIP,
			&sub.UserAgent,
			&respJSON,
			&sub.SubmittedAt,
		)
		if err != nil {
			return nil, err
		}

		if taskID.Valid {
			tid := int(taskID.Int64)
			sub.TaskID = &tid
		}
		if respJSON != "" {
			_ = json.Unmarshal([]byte(respJSON), &sub.Responses)
		}
		if sub.Responses == nil {
			sub.Responses = make(map[string]interface{})
		}

		submissions = append(submissions, &sub)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return submissions, nil
}

func (r *formRepositoryImpl) GetFormAnalytics(formID int64) (*model.FormAnalytics, error) {
	var (
		title     string
		totalSubs int
	)
	err := r.db.QueryRow("SELECT title, submission_count FROM intake_forms WHERE id = ?", formID).Scan(&title, &totalSubs)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrFormNotFound
		}
		return nil, err
	}

	analytics := &model.FormAnalytics{
		FormID:           formID,
		FormTitle:        title,
		TotalSubmissions: totalSubs,
		DailySubmissions: make(map[string]int),
	}

	// Submissions in last 7 days
	_ = r.db.QueryRow(`
		SELECT COUNT(*) 
		FROM intake_form_submissions 
		WHERE form_id = ? AND submitted_at >= datetime('now', '-7 days')`, formID).Scan(&analytics.SubmissionsLast7d)

	// Submissions in last 30 days
	_ = r.db.QueryRow(`
		SELECT COUNT(*) 
		FROM intake_form_submissions 
		WHERE form_id = ? AND submitted_at >= datetime('now', '-30 days')`, formID).Scan(&analytics.SubmissionsLast30d)

	// First and last submitted at
	var (
		firstSub sql.NullTime
		lastSub  sql.NullTime
	)
	_ = r.db.QueryRow(`
		SELECT MIN(submitted_at), MAX(submitted_at) 
		FROM intake_form_submissions 
		WHERE form_id = ?`, formID).Scan(&firstSub, &lastSub)

	if firstSub.Valid {
		analytics.FirstSubmittedAt = &firstSub.Time
	}
	if lastSub.Valid {
		analytics.LastSubmittedAt = &lastSub.Time
	}

	// Daily counts for last 30 days
	dailyRows, err := r.db.Query(`
		SELECT strftime('%Y-%m-%d', submitted_at) AS day, COUNT(*) 
		FROM intake_form_submissions 
		WHERE form_id = ? AND submitted_at >= datetime('now', '-30 days')
		GROUP BY day 
		ORDER BY day ASC`, formID)
	if err == nil {
		defer dailyRows.Close()
		for dailyRows.Next() {
			var day string
			var count int
			if err := dailyRows.Scan(&day, &count); err == nil {
				analytics.DailySubmissions[day] = count
			}
		}
	}

	return analytics, nil
}

