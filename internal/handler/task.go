package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	appError "github.com/JohnMalugu/tsk-mgr-api/internal/error"
	"github.com/JohnMalugu/tsk-mgr-api/internal/model"
	"github.com/JohnMalugu/tsk-mgr-api/internal/service"
	"github.com/JohnMalugu/tsk-mgr-api/internal/validation"
)

// HandleTasks handles collection requests for /tasks.
func HandleTasks(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		respondJSON(w, http.StatusOK, service.GetAllTasks())
	case http.MethodPost:
		createTask(w, r)
	default:
		w.Header().Set("Allow", http.MethodGet+", "+http.MethodPost)
		respondError(w, r, http.StatusMethodNotAllowed, "method not allowed")
	}
}

// HandleTaskByID handles requests for /tasks/{id}.
func HandleTaskByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		respondError(w, r, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	idText := strings.TrimPrefix(r.URL.Path, "/tasks/")
	id, err := strconv.Atoi(idText)
	if err != nil || id < 1 {
		respondError(w, r, http.StatusBadRequest, "task id must be a positive integer")
		return
	}

	task := service.GetTaskByID(id)
	if task == nil {
		respondError(w, r, http.StatusNotFound, "task not found")
		return
	}
	respondJSON(w, http.StatusOK, task)
}

func createTask(w http.ResponseWriter, r *http.Request) {
	var task model.Task
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&task); err != nil {
		respondError(w, r, http.StatusBadRequest, "request body must be valid JSON")
		return
	}

	if errors := validation.ValidateTask(&task); len(errors) > 0 {
		respondJSON(w, http.StatusBadRequest, map[string]interface{}{"errors": errors})
		return
	}

	respondJSON(w, http.StatusCreated, service.CreateTask(task))
}

func respondJSON(w http.ResponseWriter, status int, value interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func respondError(w http.ResponseWriter, r *http.Request, status int, message string) {
	appError.RespondWithError(w, r, appError.NewAppError(status, message, nil))
}

func respondError(w http.ResponseWriter, r *http.Request, status int, message string) {
	appError.RespondWithError(w, r, appError.NewAppError(status, message, nil))
}