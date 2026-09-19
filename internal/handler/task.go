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
		completed, err := completionFilter(r)
		if err != nil {
			respondError(w, r, http.StatusBadRequest, "completed must be true or false")
			return
		}
		respondJSON(w, http.StatusOK, service.GetTasks(completed, r.URL.Query().Get("q")))
	case http.MethodPost:
		createTask(w, r)
	default:
		w.Header().Set("Allow", http.MethodGet+", "+http.MethodPost)
		respondError(w, r, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func completionFilter(r *http.Request) (*bool, error) {
	value := r.URL.Query().Get("completed")
	if value == "" {
		return nil, nil
	}

	completed, err := strconv.ParseBool(value)
	if err != nil {
		return nil, err
	}
	return &completed, nil
}

// HandleTaskByID handles requests for /tasks/{id}.
func HandleTaskByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodPut && r.Method != http.MethodDelete {
		w.Header().Set("Allow", http.MethodGet+", "+http.MethodPut+", "+http.MethodDelete)
		respondError(w, r, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	idText := strings.TrimPrefix(r.URL.Path, "/tasks/")
	id, err := strconv.Atoi(idText)
	if err != nil || id < 1 {
		respondError(w, r, http.StatusBadRequest, "task id must be a positive integer")
		return
	}

	switch r.Method {
	case http.MethodGet:
		task := service.GetTaskByID(id)
		if task == nil {
			respondError(w, r, http.StatusNotFound, "task not found")
			return
		}
		respondJSON(w, http.StatusOK, task)
	case http.MethodPut:
		updateTask(w, r, id)
	case http.MethodDelete:
		if !service.DeleteTask(id) {
			respondError(w, r, http.StatusNotFound, "task not found")
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func createTask(w http.ResponseWriter, r *http.Request) {
	task, ok := decodeTask(w, r)
	if !ok {
		return
	}

	respondJSON(w, http.StatusCreated, service.CreateTask(task))
}

func updateTask(w http.ResponseWriter, r *http.Request, id int) {
	task, ok := decodeTask(w, r)
	if !ok {
		return
	}

	updated := service.UpdateTask(id, task)
	if updated == nil {
		respondError(w, r, http.StatusNotFound, "task not found")
		return
	}
	respondJSON(w, http.StatusOK, updated)
}

func decodeTask(w http.ResponseWriter, r *http.Request) (model.Task, bool) {
	var task model.Task
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&task); err != nil {
		respondError(w, r, http.StatusBadRequest, "request body must be valid JSON")
		return model.Task{}, false
	}

	if errors := validation.ValidateTask(&task); len(errors) > 0 {
		respondJSON(w, http.StatusBadRequest, map[string]interface{}{"errors": errors})
		return model.Task{}, false
	}
	return task, true
}

func respondJSON(w http.ResponseWriter, status int, value interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func respondError(w http.ResponseWriter, r *http.Request, status int, message string) {
	appError.RespondWithError(w, r, appError.NewAppError(status, message, nil))
}
