package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/JohnMalugu/tsk-mgr-api/internal/service"
)

type TaskDependencyHandler struct {
	svc service.TaskDependencyService
}

var GlobalTaskDependencyHandler *TaskDependencyHandler

func InitTaskDependencyHandler(svc service.TaskDependencyService) {
	GlobalTaskDependencyHandler = &TaskDependencyHandler{svc: svc}
}

func NewTaskDependencyHandler(svc service.TaskDependencyService) *TaskDependencyHandler {
	return &TaskDependencyHandler{svc: svc}
}

type AddDependencyRequest struct {
	DependsOnTaskID int `json:"dependsOnTaskId"`
}

func (h *TaskDependencyHandler) HandleAddDependency(w http.ResponseWriter, r *http.Request, taskID int) {
	userID, ok := r.Context().Value("userID").(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req AddDependencyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	dep, err := h.svc.AddDependency(taskID, req.DependsOnTaskID, userID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrTaskNotFoundDB):
			http.Error(w, err.Error(), http.StatusNotFound)
		case errors.Is(err, service.ErrUnauthorizedTask):
			http.Error(w, err.Error(), http.StatusForbidden)
		case errors.Is(err, service.ErrDependencySelfRefDB):
			http.Error(w, err.Error(), http.StatusBadRequest)
		case errors.Is(err, service.ErrDependencyCycleDB):
			http.Error(w, err.Error(), http.StatusBadRequest)
		case errors.Is(err, service.ErrDependencyExistsDB):
			http.Error(w, err.Error(), http.StatusConflict)
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(dep)
}
