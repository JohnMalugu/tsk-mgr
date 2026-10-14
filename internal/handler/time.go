package handler

import (
	"encoding/json"
	"net/http"

	"github.com/JohnMalugu/tsk-mgr-api/internal/service"
)

type TimeHandler struct {
	svc service.TimeTrackingService
}

var GlobalTimeHandler *TimeHandler

func InitTimeHandler(svc service.TimeTrackingService) {
	GlobalTimeHandler = &TimeHandler{svc: svc}
}

func NewTimeHandler(svc service.TimeTrackingService) *TimeHandler {
	return &TimeHandler{svc: svc}
}

type StartTimerRequest struct {
	Note string `json:"note"`
}

func (h *TimeHandler) HandleStartTimer(w http.ResponseWriter, r *http.Request, taskID int) {
	userID, ok := r.Context().Value("userID").(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req StartTimerRequest
	_ = json.NewDecoder(r.Body).Decode(&req)

	entry, err := h.svc.StartTimer(taskID, userID, req.Note)
	if err != nil {
		if err == service.ErrTaskNotFoundDB {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		if err == service.ErrUnauthorizedTask {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		if err == service.ErrActiveTimerExists {
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(entry)
}

func (h *TimeHandler) HandleStopTimer(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	entry, err := h.svc.StopTimer(userID)
	if err != nil {
		if err == service.ErrNoActiveTimerFound {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(entry)
}
