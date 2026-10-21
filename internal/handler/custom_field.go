package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/JohnMalugu/tsk-mgr-api/internal/model"
	"github.com/JohnMalugu/tsk-mgr-api/internal/service"
)

type CustomFieldHandler struct {
	svc service.CustomFieldService
}

var GlobalCustomFieldHandler *CustomFieldHandler

func InitCustomFieldHandler(svc service.CustomFieldService) {
	GlobalCustomFieldHandler = &CustomFieldHandler{svc: svc}
}

func NewCustomFieldHandler(svc service.CustomFieldService) *CustomFieldHandler {
	return &CustomFieldHandler{svc: svc}
}

func (h *CustomFieldHandler) HandleCreateDefinition(w http.ResponseWriter, r *http.Request) {
	var req model.CreateCustomFieldRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid input: "+err.Error(), http.StatusBadRequest)
		return
	}

	def, err := h.svc.CreateFieldDefinition(&req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrCustomFieldNameRequired),
			errors.Is(err, service.ErrInvalidCustomFieldType),
			errors.Is(err, service.ErrDropdownOptionsRequired),
			errors.Is(err, service.ErrCustomFieldKeyInvalid):
			http.Error(w, err.Error(), http.StatusBadRequest)
		case errors.Is(err, service.ErrCustomFieldKeyTaken):
			http.Error(w, err.Error(), http.StatusConflict)
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(def)
}

func (h *CustomFieldHandler) HandleGetDefinition(w http.ResponseWriter, r *http.Request, id int64) {
	def, err := h.svc.GetFieldDefinition(id)
	if err != nil {
		if errors.Is(err, service.ErrCustomFieldNotFound) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(def)
}

func (h *CustomFieldHandler) HandleListDefinitions(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	filter := model.CustomFieldFilter{
		Limit:  20,
		Offset: 0,
	}

	if wsStr := query.Get("workspace_id"); wsStr != "" {
		if wsID, err := strconv.ParseInt(wsStr, 10, 64); err == nil {
			filter.WorkspaceID = wsID
		}
	}
	if pStr := query.Get("project_id"); pStr != "" {
		if pID, err := strconv.ParseInt(pStr, 10, 64); err == nil {
			filter.ProjectID = &pID
		}
	}
	if typeStr := query.Get("type"); typeStr != "" {
		t := model.CustomFieldType(typeStr)
		filter.Type = &t
	}
	if reqStr := query.Get("required"); reqStr != "" {
		reqBool := reqStr == "true" || reqStr == "1"
		filter.Required = &reqBool
	}
	if limitStr := query.Get("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
			filter.Limit = limit
		}
	}
	if offsetStr := query.Get("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil && offset >= 0 {
			filter.Offset = offset
		}
	}

	summary, err := h.svc.ListFieldDefinitions(filter)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(summary)
}

func (h *CustomFieldHandler) HandleUpdateDefinition(w http.ResponseWriter, r *http.Request, id int64) {
	var req model.UpdateCustomFieldRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid input: "+err.Error(), http.StatusBadRequest)
		return
	}

	def, err := h.svc.UpdateFieldDefinition(id, &req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrCustomFieldNotFound):
			http.Error(w, err.Error(), http.StatusNotFound)
		case errors.Is(err, service.ErrCustomFieldNameRequired),
			errors.Is(err, service.ErrDropdownOptionsRequired):
			http.Error(w, err.Error(), http.StatusBadRequest)
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(def)
}

func (h *CustomFieldHandler) HandleDeleteDefinition(w http.ResponseWriter, r *http.Request, id int64) {
	err := h.svc.DeleteFieldDefinition(id)
	if err != nil {
		if errors.Is(err, service.ErrCustomFieldNotFound) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
