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
