package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/JohnMalugu/tsk-mgr-api/internal/model"
	"github.com/JohnMalugu/tsk-mgr-api/internal/service"
)

// FormHandler handles HTTP requests for public forms and intake workflows
type FormHandler struct {
	svc service.FormService
}

var GlobalFormHandler *FormHandler

// InitFormHandler initializes the global FormHandler singleton
func InitFormHandler(svc service.FormService) {
	GlobalFormHandler = &FormHandler{svc: svc}
}

// NewFormHandler creates a new FormHandler
func NewFormHandler(svc service.FormService) *FormHandler {
	return &FormHandler{svc: svc}
}

func getFormUser(r *http.Request) string {
	if uid, ok := r.Context().Value("userID").(string); ok && uid != "" {
		return uid
	}
	if uid := r.Header.Get("X-User-ID"); uid != "" {
		return uid
	}
	if uid := r.URL.Query().Get("user_id"); uid != "" {
		return uid
	}
	return "default_user"
}

func extractFormID(path, prefix string) (int64, error) {
	trimmed := strings.TrimPrefix(path, prefix)
	parts := strings.Split(strings.Trim(trimmed, "/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		return 0, errors.New("missing form ID in path")
	}
	id, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return 0, errors.New("invalid form ID format")
	}
	return id, nil
}

func writeFormJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(data)
}

func writeFormError(w http.ResponseWriter, statusCode int, message string) {
	writeFormJSON(w, statusCode, map[string]string{"error": message})
}

// HandleFormsRoot routes /v2/forms GET (list) and POST (create)
func (h *FormHandler) HandleFormsRoot(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.HandleListForms(w, r)
	case http.MethodPost:
		h.HandleCreateForm(w, r)
	default:
		writeFormError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

// HandleCreateForm creates a new intake form
func (h *FormHandler) HandleCreateForm(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeFormError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	userID := getFormUser(r)

	var req model.CreateFormRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeFormError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	form, err := h.svc.CreateForm(userID, req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrFormWorkspaceRequired), errors.Is(err, service.ErrFormTitleRequired), errors.Is(err, service.ErrFormInvalidSlug):
			writeFormError(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, service.ErrFormSlugAlreadyExists):
			writeFormError(w, http.StatusConflict, err.Error())
		default:
			writeFormError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	writeFormJSON(w, http.StatusCreated, form)
}

// HandleListForms lists intake forms matching query filters
func (h *FormHandler) HandleListForms(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeFormError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var filter model.FormFilter
	q := r.URL.Query()

	if wsStr := q.Get("workspace_id"); wsStr != "" {
		if wsID, err := strconv.ParseInt(wsStr, 10, 64); err == nil {
			filter.WorkspaceID = &wsID
		}
	}
	if prjStr := q.Get("project_id"); prjStr != "" {
		if prjID, err := strconv.ParseInt(prjStr, 10, 64); err == nil {
			filter.ProjectID = &prjID
		}
	}
	if status := q.Get("status"); status != "" {
		filter.Status = &status
	}
	if pubStr := q.Get("is_published"); pubStr != "" {
		pub := pubStr == "true" || pubStr == "1"
		filter.IsPublished = &pub
	}
	filter.Search = q.Get("search")

	forms, err := h.svc.ListForms(filter)
	if err != nil {
		writeFormError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeFormJSON(w, http.StatusOK, forms)
}

// HandleGetForm fetches a form by its ID
func (h *FormHandler) HandleGetForm(w http.ResponseWriter, r *http.Request, id int64) {
	if r.Method != http.MethodGet {
		writeFormError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	form, err := h.svc.GetForm(id)
	if err != nil {
		if errors.Is(err, service.ErrFormNotFound) {
			writeFormError(w, http.StatusNotFound, "Form not found")
			return
		}
		writeFormError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeFormJSON(w, http.StatusOK, form)
}

// HandleUpdateForm updates intake form settings
func (h *FormHandler) HandleUpdateForm(w http.ResponseWriter, r *http.Request, id int64) {
	var req model.UpdateFormRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeFormError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	form, err := h.svc.UpdateForm(id, req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrFormNotFound):
			writeFormError(w, http.StatusNotFound, "Form not found")
		case errors.Is(err, service.ErrFormTitleRequired), errors.Is(err, service.ErrFormInvalidSlug):
			writeFormError(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, service.ErrFormSlugAlreadyExists):
			writeFormError(w, http.StatusConflict, err.Error())
		default:
			writeFormError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	writeFormJSON(w, http.StatusOK, form)
}

// HandleDeleteForm removes an intake form and all associated fields
func (h *FormHandler) HandleDeleteForm(w http.ResponseWriter, r *http.Request, id int64) {
	if r.Method != http.MethodDelete {
		writeFormError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	if err := h.svc.DeleteForm(id); err != nil {
		if errors.Is(err, service.ErrFormNotFound) {
			writeFormError(w, http.StatusNotFound, "Form not found")
			return
		}
		writeFormError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeFormJSON(w, http.StatusOK, map[string]string{"message": "Form deleted successfully"})
}

// HandlePublishForm publishes or unpublishes an intake form
func (h *FormHandler) HandlePublishForm(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeFormError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	id, err := extractFormID(r.URL.Path, "/v2/forms/")
	if err != nil {
		writeFormError(w, http.StatusBadRequest, err.Error())
		return
	}

	var body struct {
		Publish *bool `json:"publish"`
	}
	publish := true
	if err := json.NewDecoder(r.Body).Decode(&body); err == nil && body.Publish != nil {
		publish = *body.Publish
	}

	form, err := h.svc.PublishForm(id, publish)
	if err != nil {
		if errors.Is(err, service.ErrFormNotFound) {
			writeFormError(w, http.StatusNotFound, "Form not found")
			return
		}
		writeFormError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeFormJSON(w, http.StatusOK, form)
}

// HandleFormFields manages form field configuration
func (h *FormHandler) HandleFormFields(w http.ResponseWriter, r *http.Request) {
	id, err := extractFormID(r.URL.Path, "/v2/forms/")
	if err != nil {
		writeFormError(w, http.StatusBadRequest, err.Error())
		return
	}

	switch r.Method {
	case http.MethodGet:
		fields, err := h.svc.GetFormFields(id)
		if err != nil {
			if errors.Is(err, service.ErrFormNotFound) {
				writeFormError(w, http.StatusNotFound, "Form not found")
				return
			}
			writeFormError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeFormJSON(w, http.StatusOK, fields)

	case http.MethodPut, http.MethodPost:
		var req model.SetFormFieldsRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeFormError(w, http.StatusBadRequest, "Invalid request body")
			return
		}

		fields, err := h.svc.SetFormFields(id, req)
		if err != nil {
			switch {
			case errors.Is(err, service.ErrFormNotFound):
				writeFormError(w, http.StatusNotFound, "Form not found")
			case errors.Is(err, service.ErrFormRequiredField), errors.Is(err, service.ErrFormInvalidFieldType), errors.Is(err, service.ErrFormInvalidTargetField):
				writeFormError(w, http.StatusBadRequest, err.Error())
			default:
				writeFormError(w, http.StatusInternalServerError, err.Error())
			}
			return
		}
		writeFormJSON(w, http.StatusOK, fields)

	default:
		writeFormError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

// HandlePublicFormView serves the unauthenticated public intake form definition
func (h *FormHandler) HandlePublicFormView(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeFormError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	trimmed := strings.TrimPrefix(r.URL.Path, "/v2/forms/public/")
	slug := strings.Trim(trimmed, "/")
	if slug == "" {
		writeFormError(w, http.StatusBadRequest, "Missing form slug")
		return
	}

	view, err := h.svc.GetPublicFormView(slug)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrFormNotFound):
			writeFormError(w, http.StatusNotFound, "Form not found")
		case errors.Is(err, service.ErrFormNotPublished):
			writeFormError(w, http.StatusForbidden, "Form is currently not accepting submissions")
		default:
			writeFormError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	writeFormJSON(w, http.StatusOK, view)
}

// HandlePublicFormSubmit processes public submission and provisions tasks
func (h *FormHandler) HandlePublicFormSubmit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeFormError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	trimmed := strings.TrimPrefix(r.URL.Path, "/v2/forms/public/")
	slug := strings.TrimSuffix(trimmed, "/submit")
	slug = strings.Trim(slug, "/")
	if slug == "" {
		writeFormError(w, http.StatusBadRequest, "Missing form slug")
		return
	}

	var req model.SubmitFormRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeFormError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	clientIP := r.Header.Get("X-Forwarded-For")
	if clientIP == "" {
		clientIP = r.RemoteAddr
	}
	userAgent := r.UserAgent()

	result, err := h.svc.SubmitForm(slug, clientIP, userAgent, req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrFormNotFound):
			writeFormError(w, http.StatusNotFound, "Form not found")
		case errors.Is(err, service.ErrFormNotPublished):
			writeFormError(w, http.StatusForbidden, "Form is not published")
		case errors.Is(err, service.ErrFormAnonymousForbidden):
			writeFormError(w, http.StatusUnauthorized, "Submitter email is required for this form")
		case errors.Is(err, service.ErrFormRequiredField), errors.Is(err, service.ErrFormInvalidOption):
			writeFormError(w, http.StatusBadRequest, err.Error())
		default:
			writeFormError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	writeFormJSON(w, http.StatusCreated, result)
}

// HandleFormSubmissions retrieves submission logs for a form
func (h *FormHandler) HandleFormSubmissions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeFormError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	id, err := extractFormID(r.URL.Path, "/v2/forms/")
	if err != nil {
		writeFormError(w, http.StatusBadRequest, err.Error())
		return
	}

	limit := 50
	offset := 0
	q := r.URL.Query()
	if lStr := q.Get("limit"); lStr != "" {
		if l, err := strconv.Atoi(lStr); err == nil {
			limit = l
		}
	}
	if oStr := q.Get("offset"); oStr != "" {
		if o, err := strconv.Atoi(oStr); err == nil {
			offset = o
		}
	}

	subs, err := h.svc.GetFormSubmissions(id, limit, offset)
	if err != nil {
		if errors.Is(err, service.ErrFormNotFound) {
			writeFormError(w, http.StatusNotFound, "Form not found")
			return
		}
		writeFormError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeFormJSON(w, http.StatusOK, subs)
}

// HandleFormAnalytics computes intake metrics and submission counts
func (h *FormHandler) HandleFormAnalytics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeFormError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	id, err := extractFormID(r.URL.Path, "/v2/forms/")
	if err != nil {
		writeFormError(w, http.StatusBadRequest, err.Error())
		return
	}

	analytics, err := h.svc.GetFormAnalytics(id)
	if err != nil {
		if errors.Is(err, service.ErrFormNotFound) {
			writeFormError(w, http.StatusNotFound, "Form not found")
			return
		}
		writeFormError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeFormJSON(w, http.StatusOK, analytics)
}

