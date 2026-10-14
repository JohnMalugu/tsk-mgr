package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

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

func (h *TimeHandler) HandleGetActiveTimer(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	entry, err := h.svc.GetActiveTimer(userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if entry == nil {
		json.NewEncoder(w).Encode(map[string]interface{}{"active": false})
		return
	}

	json.NewEncoder(w).Encode(entry)
}

type LogManualTimeRequest struct {
	StartedAt string `json:"started_at"`
	EndedAt   string `json:"ended_at"`
	Note      string `json:"note"`
}

func (h *TimeHandler) HandleLogManualTime(w http.ResponseWriter, r *http.Request, taskID int) {
	userID, ok := r.Context().Value("userID").(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req LogManualTimeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	startedAt, err := parseTimeFlexible(req.StartedAt)
	if err != nil {
		http.Error(w, "Invalid started_at format", http.StatusBadRequest)
		return
	}

	endedAt, err := parseTimeFlexible(req.EndedAt)
	if err != nil {
		http.Error(w, "Invalid ended_at format", http.StatusBadRequest)
		return
	}

	entry, err := h.svc.LogManualTime(taskID, userID, startedAt, endedAt, req.Note)
	if err != nil {
		if err == service.ErrInvalidTimeRange {
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
	json.NewEncoder(w).Encode(entry)
}

func parseTimeFlexible(s string) (time.Time, error) {
	formats := []string{
		time.RFC3339,
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05",
	}
	for _, f := range formats {
		if t, err := time.Parse(f, s); err == nil {
			return t, nil
		}
	}
	return time.Time{}, errors.New("cannot parse time")
}

func (h *TimeHandler) HandleGetTaskTimeEntries(w http.ResponseWriter, r *http.Request, taskID int) {
	userID, ok := r.Context().Value("userID").(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	entries, err := h.svc.GetTaskTimeEntries(taskID, userID)
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
	json.NewEncoder(w).Encode(entries)
}

func (h *TimeHandler) HandleDeleteTimeEntry(w http.ResponseWriter, r *http.Request, entryID int) {
	userID, ok := r.Context().Value("userID").(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	err := h.svc.DeleteTimeEntry(entryID, userID)
	if err != nil {
		if err == service.ErrTimeEntryNotFoundDB {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		if err == service.ErrUnauthorizedTimeEntry {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *TimeHandler) HandleGetTaskTimeSummary(w http.ResponseWriter, r *http.Request, taskID int) {
	userID, ok := r.Context().Value("userID").(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	summary, err := h.svc.GetTaskTimeSummary(taskID, userID)
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
	json.NewEncoder(w).Encode(summary)
}

func (h *TimeHandler) HandleGetTimeReport(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var fromPtr *time.Time
	if fromStr := r.URL.Query().Get("from"); fromStr != "" {
		if t, err := parseTimeFlexible(fromStr); err == nil {
			fromPtr = &t
		}
	}

	var toPtr *time.Time
	if toStr := r.URL.Query().Get("to"); toStr != "" {
		if t, err := parseTimeFlexible(toStr); err == nil {
			toPtr = &t
		}
	}

	report, err := h.svc.GetTimeReport(userID, fromPtr, toPtr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(report)
}



