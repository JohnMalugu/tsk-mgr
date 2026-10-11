package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/JohnMalugu/tsk-mgr-api/internal/service"
)

type TaskDBHandler struct {
	svc service.TaskDBService
}

var GlobalTaskDBHandler *TaskDBHandler

func InitTaskDBHandler(svc service.TaskDBService) {
	GlobalTaskDBHandler = &TaskDBHandler{svc: svc}
}

func NewTaskDBHandler(svc service.TaskDBService) *TaskDBHandler {
	return &TaskDBHandler{svc: svc}
}

type CreateTaskRequest struct {
	Title           string    `json:"title"`
	Description     string    `json:"description"`
	DueDate         time.Time `json:"due_date"`
	EstimateMinutes int       `json:"estimate_minutes"`
	Priority        string    `json:"priority"`
}

func (h *TaskDBHandler) HandleCreate(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req CreateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	task, err := h.svc.CreateTask(userID, req.Title, req.Description, req.DueDate, req.EstimateMinutes, req.Priority)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(task)
}

func (h *TaskDBHandler) HandleList(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	tasks, err := h.svc.GetUserTasks(userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(tasks)
}

func (h *TaskDBHandler) HandleGetByID(w http.ResponseWriter, r *http.Request, taskID int) {
	userID, ok := r.Context().Value("userID").(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	task, err := h.svc.GetTask(taskID, userID)
	if err != nil {
		if err == service.ErrTaskNotFoundDB {
			http.Error(w, "Task not found", http.StatusNotFound)
			return
		}
		if err == service.ErrUnauthorizedTask {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(task)
}

type UpdateTaskRequest struct {
	Title           string    `json:"title"`
	Description     string    `json:"description"`
	DueDate         time.Time `json:"due_date"`
	EstimateMinutes int       `json:"estimate_minutes"`
	Completed       bool      `json:"completed"`
	Status          string    `json:"status"`
	Priority        string    `json:"priority"`
}

func (h *TaskDBHandler) HandleUpdate(w http.ResponseWriter, r *http.Request, taskID int) {
	userID, ok := r.Context().Value("userID").(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req UpdateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	task, err := h.svc.UpdateTask(taskID, userID, req.Title, req.Description, req.DueDate, req.EstimateMinutes, req.Completed, req.Status, req.Priority)
	if err != nil {
		if err == service.ErrTaskNotFoundDB {
			http.Error(w, "Task not found", http.StatusNotFound)
			return
		}
		if err == service.ErrUnauthorizedTask {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(task)
}

func (h *TaskDBHandler) HandleDelete(w http.ResponseWriter, r *http.Request, taskID int) {
	userID, ok := r.Context().Value("userID").(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	err := h.svc.DeleteTask(taskID, userID)
	if err != nil {
		if err == service.ErrTaskNotFoundDB {
			http.Error(w, "Task not found", http.StatusNotFound)
			return
		}
		if err == service.ErrUnauthorizedTask {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *TaskDBHandler) HandleComplete(w http.ResponseWriter, r *http.Request, taskID int) {
	userID, ok := r.Context().Value("userID").(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	task, err := h.svc.CompleteTask(taskID, userID, true)
	if err != nil {
		if err == service.ErrTaskNotFoundDB {
			http.Error(w, "Task not found", http.StatusNotFound)
			return
		}
		if err == service.ErrUnauthorizedTask {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(task)
}
