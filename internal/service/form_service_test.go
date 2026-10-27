package service

import (
	"database/sql"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/JohnMalugu/tsk-mgr-api/internal/db"
	"github.com/JohnMalugu/tsk-mgr-api/internal/model"
	"github.com/JohnMalugu/tsk-mgr-api/internal/repository"
)

func setupTestFormSvc(t *testing.T) (FormService, repository.TaskRepository, *sql.DB) {
	sqliteDB, err := db.InitDB(":memory:")
	if err != nil {
		t.Fatalf("failed to initialize in-memory test db: %v", err)
	}

	_, err = sqliteDB.Exec(`
		INSERT INTO users (id, username, email, password_hash, created_at, updated_at)
		VALUES ('user_1', 'john_doe', 'john@example.com', 'hash', datetime('now'), datetime('now'));

		INSERT INTO workspaces (id, name, slug, owner_id, created_at, updated_at) 
		VALUES (1, 'Engineering WS', 'eng-ws', 'user_1', datetime('now'), datetime('now')),
		       (2, 'Operations WS', 'ops-ws', 'user_1', datetime('now'), datetime('now'));
	`)
	if err != nil {
		t.Fatalf("failed to insert test data: %v", err)
	}

	formRepo := repository.NewFormRepository(sqliteDB)
	taskRepo := repository.NewTaskRepository(sqliteDB)
	svc := NewFormService(formRepo, taskRepo)
	return svc, taskRepo, sqliteDB
}

func TestFormService_Validation(t *testing.T) {
	svc, _, _ := setupTestFormSvc(t)

	// Workspace ID required
	_, err := svc.CreateForm("user_1", model.CreateFormRequest{
		WorkspaceID: 0,
		Title:       "Feature Request",
	})
	if !errors.Is(err, ErrFormWorkspaceRequired) {
		t.Fatalf("expected ErrFormWorkspaceRequired, got: %v", err)
	}

	// Title required
	_, err = svc.CreateForm("user_1", model.CreateFormRequest{
		WorkspaceID: 1,
		Title:       "   ",
	})
	if !errors.Is(err, ErrFormTitleRequired) {
		t.Fatalf("expected ErrFormTitleRequired, got: %v", err)
	}

	// Invalid slug
	_, err = svc.CreateForm("user_1", model.CreateFormRequest{
		WorkspaceID: 1,
		Title:       "Bug Report",
		Slug:        "----",
	})
	if !errors.Is(err, ErrFormInvalidSlug) {
		t.Fatalf("expected ErrFormInvalidSlug, got: %v", err)
	}

	// Auto-derives slug
	form, err := svc.CreateForm("user_1", model.CreateFormRequest{
		WorkspaceID: 1,
		Title:       "Feedback & Suggestions!",
	})
	if err != nil {
		t.Fatalf("unexpected error creating form: %v", err)
	}
	if !strings.Contains(form.Slug, "feedback-suggestions") {
		t.Fatalf("expected slug to contain 'feedback-suggestions', got: %s", form.Slug)
	}
}

func TestFormService_CRUD(t *testing.T) {
	svc, _, _ := setupTestFormSvc(t)

	anon := false
	form, err := svc.CreateForm("user_1", model.CreateFormRequest{
		WorkspaceID:     1,
		Title:           "Internal IT Requests",
		Description:     "Submit IT hardware and access requests",
		Slug:            "it-requests",
		AllowAnonymous:  &anon,
		DefaultPriority: "high",
		DefaultTags:     []string{"it", "helpdesk"},
	})
	if err != nil {
		t.Fatalf("failed to create form: %v", err)
	}
	if form.ID == 0 {
		t.Fatalf("expected nonzero form ID")
	}
	if form.AllowAnonymous {
		t.Fatalf("expected AllowAnonymous=false")
	}
	if form.DefaultPriority != "high" {
		t.Fatalf("expected high priority, got %s", form.DefaultPriority)
	}

	// Get by ID
	fetched, err := svc.GetForm(form.ID)
	if err != nil {
		t.Fatalf("failed to get form: %v", err)
	}
	if fetched.Title != "Internal IT Requests" {
		t.Fatalf("expected title 'Internal IT Requests', got: %s", fetched.Title)
	}

	// Get by Slug
	bySlug, err := svc.GetFormBySlug("it-requests")
	if err != nil {
		t.Fatalf("failed to get form by slug: %v", err)
	}
	if bySlug.ID != form.ID {
		t.Fatalf("form ID mismatch: expected %d, got %d", form.ID, bySlug.ID)
	}

	// Duplicate slug forbidden
	_, err = svc.CreateForm("user_1", model.CreateFormRequest{
		WorkspaceID: 1,
		Title:       "Another Form",
		Slug:        "it-requests",
	})
	if !errors.Is(err, ErrFormSlugAlreadyExists) {
		t.Fatalf("expected ErrFormSlugAlreadyExists, got: %v", err)
	}

	// Update form
	newTitle := "IT & Hardware Helpdesk"
	newPriority := "urgent"
	updated, err := svc.UpdateForm(form.ID, model.UpdateFormRequest{
		Title:           &newTitle,
		DefaultPriority: &newPriority,
	})
	if err != nil {
		t.Fatalf("failed to update form: %v", err)
	}
	if updated.Title != newTitle || updated.DefaultPriority != newPriority {
		t.Fatalf("update mismatch: title=%s, priority=%s", updated.Title, updated.DefaultPriority)
	}

	// List forms
	wsID := int64(1)
	summaries, err := svc.ListForms(model.FormFilter{
		WorkspaceID: &wsID,
	})
	if err != nil {
		t.Fatalf("failed to list forms: %v", err)
	}
	if len(summaries) != 1 {
		t.Fatalf("expected 1 form summary, got %d", len(summaries))
	}
	if summaries[0].ID != form.ID {
		t.Fatalf("summary ID mismatch: expected %d, got %d", form.ID, summaries[0].ID)
	}

	// Delete form
	if err := svc.DeleteForm(form.ID); err != nil {
		t.Fatalf("failed to delete form: %v", err)
	}
	_, err = svc.GetForm(form.ID)
	if !errors.Is(err, ErrFormNotFound) {
		t.Fatalf("expected ErrFormNotFound after deletion, got: %v", err)
	}
}

func TestFormService_PublishAndPublicView(t *testing.T) {
	svc, _, _ := setupTestFormSvc(t)

	form, err := svc.CreateForm("user_1", model.CreateFormRequest{
		WorkspaceID: 1,
		Title:       "Customer Feedback Portal",
		Slug:        "customer-feedback",
	})
	if err != nil {
		t.Fatalf("failed to create form: %v", err)
	}

	// When draft, public view should return ErrFormNotPublished
	_, err = svc.GetPublicFormView(form.Slug)
	if !errors.Is(err, ErrFormNotPublished) {
		t.Fatalf("expected ErrFormNotPublished for draft form, got: %v", err)
	}

	// Publish form
	published, err := svc.PublishForm(form.ID, true)
	if err != nil {
		t.Fatalf("failed to publish form: %v", err)
	}
	if !published.IsPublished || published.Status != model.FormStatusPublished {
		t.Fatalf("expected published form, got status=%s, is_published=%v", published.Status, published.IsPublished)
	}

	// Public view succeeds
	view, err := svc.GetPublicFormView(form.Slug)
	if err != nil {
		t.Fatalf("failed to get public form view: %v", err)
	}
	if view.Title != form.Title || view.Slug != form.Slug {
		t.Fatalf("public view metadata mismatch: title=%s, slug=%s", view.Title, view.Slug)
	}

	// Unpublish form
	unpublished, err := svc.PublishForm(form.ID, false)
	if err != nil {
		t.Fatalf("failed to unpublish form: %v", err)
	}
	if unpublished.IsPublished || unpublished.Status != model.FormStatusDraft {
		t.Fatalf("expected draft form after unpublishing, got: %s", unpublished.Status)
	}
}

func TestFormService_FieldsManagement(t *testing.T) {
	svc, _, _ := setupTestFormSvc(t)

	form, err := svc.CreateForm("user_1", model.CreateFormRequest{
		WorkspaceID: 1,
		Title:       "Bug Tracker Intake",
		Slug:        "bug-tracker",
	})
	if err != nil {
		t.Fatalf("failed to create form: %v", err)
	}

	// Invalid field type
	_, err = svc.SetFormFields(form.ID, model.SetFormFieldsRequest{
		Fields: []model.FormFieldDefinition{
			{FieldKey: "f1", FieldType: "unsupported", Label: "Label"},
		},
	})
	if !errors.Is(err, ErrFormInvalidFieldType) {
		t.Fatalf("expected ErrFormInvalidFieldType, got: %v", err)
	}

	// Dropdown with no options
	_, err = svc.SetFormFields(form.ID, model.SetFormFieldsRequest{
		Fields: []model.FormFieldDefinition{
			{FieldKey: "severity", FieldType: model.FormFieldTypeDropdown, Label: "Severity", Options: []string{}},
		},
	})
	if err == nil {
		t.Fatalf("expected error for empty dropdown options")
	}

	// Valid fields definition
	fields, err := svc.SetFormFields(form.ID, model.SetFormFieldsRequest{
		Fields: []model.FormFieldDefinition{
			{
				FieldKey:    "issue_summary",
				FieldType:   model.FormFieldTypeText,
				TargetField: model.FieldTargetTitle,
				Label:       "Issue Summary",
				IsRequired:  true,
				Position:    1,
			},
			{
				FieldKey:    "steps_to_reproduce",
				FieldType:   model.FormFieldTypeParagraph,
				TargetField: model.FieldTargetDescription,
				Label:       "Steps to Reproduce",
				IsRequired:  true,
				Position:    2,
			},
			{
				FieldKey:    "severity_level",
				FieldType:   model.FormFieldTypeDropdown,
				TargetField: model.FieldTargetPriority,
				Label:       "Severity",
				Options:     []string{"low", "medium", "high", "urgent"},
				Position:    3,
			},
			{
				FieldKey:    "app_version",
				FieldType:   model.FormFieldTypeText,
				TargetField: model.FieldTargetCustom,
				Label:       "App Version",
				Position:    4,
			},
		},
	})
	if err != nil {
		t.Fatalf("failed to set form fields: %v", err)
	}
	if len(fields) != 4 {
		t.Fatalf("expected 4 fields, got %d", len(fields))
	}
	if fields[0].FieldKey != "issue_summary" || fields[0].TargetField != model.FieldTargetTitle {
		t.Fatalf("first field mismatch: key=%s, target=%s", fields[0].FieldKey, fields[0].TargetField)
	}

	// Get form fields
	fetched, err := svc.GetFormFields(form.ID)
	if err != nil {
		t.Fatalf("failed to get form fields: %v", err)
	}
	if len(fetched) != 4 {
		t.Fatalf("expected 4 fetched fields, got %d", len(fetched))
	}
}

func TestFormService_PublicSubmissionAndTaskCreation(t *testing.T) {
	svc, taskRepo, _ := setupTestFormSvc(t)

	form, err := svc.CreateForm("user_1", model.CreateFormRequest{
		WorkspaceID: 1,
		Title:       "Contact Us Intake",
		Slug:        "contact-us",
		Fields: []model.FormFieldDefinition{
			{
				FieldKey:    "subject",
				FieldType:   model.FormFieldTypeText,
				TargetField: model.FieldTargetTitle,
				Label:       "Subject",
				IsRequired:  true,
			},
			{
				FieldKey:    "message",
				FieldType:   model.FormFieldTypeParagraph,
				TargetField: model.FieldTargetDescription,
				Label:       "Message",
				IsRequired:  true,
			},
			{
				FieldKey:    "department",
				FieldType:   model.FormFieldTypeDropdown,
				TargetField: model.FieldTargetCustom,
				Label:       "Department",
				Options:     []string{"Sales", "Support", "Billing"},
				IsRequired:  false,
			},
		},
	})
	if err != nil {
		t.Fatalf("failed to create form: %v", err)
	}

	// Submitting when unpublished fails
	_, err = svc.SubmitForm(form.Slug, "127.0.0.1", "curl/8.0", model.SubmitFormRequest{
		Responses: map[string]interface{}{"subject": "Hello"},
	})
	if !errors.Is(err, ErrFormNotPublished) {
		t.Fatalf("expected ErrFormNotPublished, got: %v", err)
	}

	// Publish form
	_, err = svc.PublishForm(form.ID, true)
	if err != nil {
		t.Fatalf("failed to publish form: %v", err)
	}

	// Missing required field 'subject'
	_, err = svc.SubmitForm(form.Slug, "127.0.0.1", "curl/8.0", model.SubmitFormRequest{
		Responses: map[string]interface{}{
			"message": "Only message provided",
		},
	})
	if !errors.Is(err, ErrFormRequiredField) {
		t.Fatalf("expected ErrFormRequiredField, got: %v", err)
	}

	// Invalid dropdown option
	_, err = svc.SubmitForm(form.Slug, "127.0.0.1", "curl/8.0", model.SubmitFormRequest{
		Responses: map[string]interface{}{
			"subject":    "Enterprise Quote",
			"message":    "We need 50 seats",
			"department": "NonexistentDepartment",
		},
	})
	if !errors.Is(err, ErrFormInvalidOption) {
		t.Fatalf("expected ErrFormInvalidOption, got: %v", err)
	}

	// Valid submission
	res, err := svc.SubmitForm(form.Slug, "192.168.1.100", "Mozilla/5.0", model.SubmitFormRequest{
		SubmitterEmail: "alice@example.com",
		Responses: map[string]interface{}{
			"subject":    "Enterprise Contract Question",
			"message":    "Interested in annual billing options.",
			"department": "Sales",
		},
	})
	if err != nil {
		t.Fatalf("failed to submit form: %v", err)
	}
	if res.SubmissionID == 0 {
		t.Fatalf("expected valid submission ID")
	}
	if res.TaskID == 0 {
		t.Fatalf("expected created task ID")
	}
	if res.TaskTitle != "Enterprise Contract Question" {
		t.Fatalf("task title mismatch: expected 'Enterprise Contract Question', got '%s'", res.TaskTitle)
	}

	// Verify task exists in repository
	createdTask, err := taskRepo.GetByID(res.TaskID)
	if err != nil {
		t.Fatalf("failed to get created task: %v", err)
	}
	if createdTask.Title != "Enterprise Contract Question" {
		t.Fatalf("created task title mismatch: %s", createdTask.Title)
	}
	if !strings.Contains(createdTask.Description, "Interested in annual billing options.") {
		t.Fatalf("created task description missing body: %s", createdTask.Description)
	}
	if !strings.Contains(createdTask.Description, "Sales") {
		t.Fatalf("created task description missing custom department: %s", createdTask.Description)
	}

	// Verify submission count incremented
	formAfter, err := svc.GetForm(form.ID)
	if err != nil {
		t.Fatalf("failed to get form: %v", err)
	}
	if formAfter.SubmissionCount != 1 {
		t.Fatalf("expected submission count 1, got %d", formAfter.SubmissionCount)
	}

	// Anonymous submission enforcement
	anonFalse := false
	_, err = svc.UpdateForm(form.ID, model.UpdateFormRequest{
		AllowAnonymous: &anonFalse,
	})
	if err != nil {
		t.Fatalf("failed to update allow_anonymous: %v", err)
	}

	_, err = svc.SubmitForm(form.Slug, "127.0.0.1", "curl/8.0", model.SubmitFormRequest{
		SubmitterEmail: "", // empty email
		Responses: map[string]interface{}{
			"subject": "Anonymous Request",
			"message": "Anonymous message",
		},
	})
	if !errors.Is(err, ErrFormAnonymousForbidden) {
		t.Fatalf("expected ErrFormAnonymousForbidden, got: %v", err)
	}
}

func TestFormService_SubmissionsAndAnalytics(t *testing.T) {
	svc, _, _ := setupTestFormSvc(t)

	form, err := svc.CreateForm("user_1", model.CreateFormRequest{
		WorkspaceID: 1,
		Title:       "Survey 2026",
		Slug:        "survey-2026",
	})
	if err != nil {
		t.Fatalf("failed to create form: %v", err)
	}
	_, _ = svc.PublishForm(form.ID, true)

	// Submit multiple entries
	for i := 1; i <= 3; i++ {
		_, err := svc.SubmitForm(form.Slug, "127.0.0.1", "tester", model.SubmitFormRequest{
			Responses: map[string]interface{}{
				"feedback": "Great work!",
			},
		})
		if err != nil {
			t.Fatalf("failed submission %d: %v", i, err)
		}
	}

	// List submissions
	subs, err := svc.GetFormSubmissions(form.ID, 10, 0)
	if err != nil {
		t.Fatalf("failed to get submissions: %v", err)
	}
	if len(subs) != 3 {
		t.Fatalf("expected 3 submissions, got %d", len(subs))
	}

	// Analytics
	analytics, err := svc.GetFormAnalytics(form.ID)
	if err != nil {
		t.Fatalf("failed to get analytics: %v", err)
	}
	if analytics.TotalSubmissions != 3 {
		t.Fatalf("expected total 3, got %d", analytics.TotalSubmissions)
	}
	if analytics.SubmissionsLast7d != 3 {
		t.Fatalf("expected last 7d to be 3, got %d", analytics.SubmissionsLast7d)
	}
	if analytics.FirstSubmittedAt == nil || analytics.LastSubmittedAt == nil {
		t.Fatalf("expected non-nil timestamps in analytics")
	}
	today := time.Now().Format("2006-01-02")
	if count, ok := analytics.DailySubmissions[today]; !ok || count != 3 {
		t.Fatalf("expected daily count of 3 for %s, got: %d", today, count)
	}
}
