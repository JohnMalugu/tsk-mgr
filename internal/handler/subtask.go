package handler

import (
	"encoding/json"
	"net/http"

	"github.com/JohnMalugu/tsk-mgr-api/internal/service"
)

type SubtaskHandler struct {
	svc service.SubtaskService
}

var GlobalSubtaskHandler *SubtaskHandler

func InitSubtaskHandler(svc service.SubtaskService) {
	GlobalSubtaskHandler = &SubtaskHandler{svc: svc}
}

func NewSubtaskHandler(svc service.SubtaskService) *SubtaskHandler {
	return &SubtaskHandler{svc: svc}
}

type CreateSubtaskRequest struct {
	Title string `json:"title"`
}

func (h *SubtaskHandler) HandleCreate(w http.ResponseWriter, r *http.Request, taskID int) {
	userID, ok := r.Context().Value("userID").(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req CreateSubtaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	subtask, err := h.svc.CreateSubtask(taskID, userID, req.Title)
	if err != nil {
		if err == service.ErrSubtaskTitleRequired {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err == service.ErrTaskNotFoundDB {
			http.Error(w, err.Error(), http.StatusNotFound)
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
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(subtask)
}

func (h *SubtaskHandler) HandleList(w http.ResponseWriter, r *http.Request, taskID int) {
	userID, ok := r.Context().Value("userID").(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	subtasks, err := h.svc.GetSubtasks(taskID, userID)
	if err != nil {
		if err == service.ErrTaskNotFoundDB {
			http.Error(w, err.Error(), http.StatusNotFound)
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
	json.NewEncoder(w).Encode(subtasks)
}

type ToggleSubtaskRequest struct {
	Completed bool `json:"completed"`
}

func (h *SubtaskHandler) HandleToggle(w http.ResponseWriter, r *http.Request, subtaskID int) {
	userID, ok := r.Context().Value("userID").(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req ToggleSubtaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	subtask, err := h.svc.ToggleSubtask(subtaskID, userID, req.Completed)
	if err != nil {
		if err == service.ErrSubtaskNotFound || err == service.ErrTaskNotFoundDB {
			http.Error(w, err.Error(), http.StatusNotFound)
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
	json.NewEncoder(w).Encode(subtask)
}

type UpdateSubtaskRequest struct {
	Title string `json:"title"`
}

func (h *SubtaskHandler) HandleUpdate(w http.ResponseWriter, r *http.Request, subtaskID int) {
	userID, ok := r.Context().Value("userID").(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req UpdateSubtaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	subtask, err := h.svc.UpdateSubtask(subtaskID, userID, req.Title)
	if err != nil {
		if err == service.ErrSubtaskTitleRequired {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err == service.ErrSubtaskNotFound || err == service.ErrTaskNotFoundDB {
			http.Error(w, err.Error(), http.StatusNotFound)
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
	json.NewEncoder(w).Encode(subtask)
}

func (h *SubtaskHandler) HandleDelete(w http.ResponseWriter, r *http.Request, subtaskID int) {
	userID, ok := r.Context().Value("userID").(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	err := h.svc.DeleteSubtask(subtaskID, userID)
	if err != nil {
		if err == service.ErrSubtaskNotFound || err == service.ErrTaskNotFoundDB {
			http.Error(w, err.Error(), http.StatusNotFound)
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

type ReorderSubtasksRequest struct {
	SubtaskIDs []int `json:"subtask_ids"`
}

func (h *SubtaskHandler) HandleReorder(w http.ResponseWriter, r *http.Request, taskID int) {
	userID, ok := r.Context().Value("userID").(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req ReorderSubtasksRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	err := h.svc.ReorderSubtasks(taskID, userID, req.SubtaskIDs)
	if err != nil {
		if err == service.ErrTaskNotFoundDB {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		if err == service.ErrUnauthorizedTask {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

