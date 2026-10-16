package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/JohnMalugu/tsk-mgr-api/internal/model"
	"github.com/JohnMalugu/tsk-mgr-api/internal/service"
)

type ActivityHandler struct {
	svc service.ActivityService
}

var GlobalActivityHandler *ActivityHandler

func InitActivityHandler(svc service.ActivityService) {
	GlobalActivityHandler = &ActivityHandler{svc: svc}
}

func NewActivityHandler(svc service.ActivityService) *ActivityHandler {
	return &ActivityHandler{svc: svc}
}

func parseTimeParam(str string) *time.Time {
	if str == "" {
		return nil
	}
	formats := []string{
		time.RFC3339,
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05",
		"2006-01-02",
	}
	for _, f := range formats {
		if t, err := time.Parse(f, str); err == nil {
			utc := t.UTC()
			return &utc
		}
	}
	return nil
}

func (h *ActivityHandler) HandleGetUserActivity(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	q := r.URL.Query()
	filter := &model.ActivityFilter{
		Action: q.Get("action"),
		From:   parseTimeParam(q.Get("from")),
		To:     parseTimeParam(q.Get("to")),
	}

	if tIDStr := q.Get("taskId"); tIDStr != "" {
		if tID, err := strconv.Atoi(tIDStr); err == nil {
			filter.TaskID = &tID
		}
	}

	if offStr := q.Get("offset"); offStr != "" {
		if off, err := strconv.Atoi(offStr); err == nil {
			filter.Offset = off
		}
	}

	if limStr := q.Get("limit"); limStr != "" {
		if lim, err := strconv.Atoi(limStr); err == nil {
			filter.Limit = lim
		}
	}

	res, err := h.svc.GetUserActivity(userID, filter)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

func (h *ActivityHandler) HandleGetTaskActivity(w http.ResponseWriter, r *http.Request, taskID int) {
	userID, ok := r.Context().Value("userID").(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	q := r.URL.Query()
	offset := 0
	limit := 20

	if offStr := q.Get("offset"); offStr != "" {
		if off, err := strconv.Atoi(offStr); err == nil {
			offset = off
		}
	}

	if limStr := q.Get("limit"); limStr != "" {
		if lim, err := strconv.Atoi(limStr); err == nil {
			limit = lim
		}
	}

	res, err := h.svc.GetTaskActivity(taskID, userID, offset, limit)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrTaskNotFoundDB):
			http.Error(w, err.Error(), http.StatusNotFound)
		case errors.Is(err, service.ErrUnauthorizedTask):
			http.Error(w, err.Error(), http.StatusForbidden)
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

func (h *ActivityHandler) HandleGetActivitySummary(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	q := r.URL.Query()
	from := parseTimeParam(q.Get("from"))
	to := parseTimeParam(q.Get("to"))

	summary, err := h.svc.GetActivitySummary(userID, from, to)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(summary)
}

type PurgeActivitiesRequest struct {
	RetentionDays int `json:"retentionDays"`
}

func (h *ActivityHandler) HandlePurgeActivities(w http.ResponseWriter, r *http.Request) {
	_, ok := r.Context().Value("userID").(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	days := 90
	if r.Body != nil {
		var req PurgeActivitiesRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err == nil && req.RetentionDays > 0 {
			days = req.RetentionDays
		}
	}
	if qDays := r.URL.Query().Get("days"); qDays != "" {
		if d, err := strconv.Atoi(qDays); err == nil && d > 0 {
			days = d
		}
	}

	purged, err := h.svc.PurgeOldActivities(days)
	if err != nil {
		if errors.Is(err, service.ErrInvalidRetentionCutoff) {
			http.Error(w, err.Error(), http.StatusBadRequest)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"purged":        purged,
		"retentionDays": days,
	})
}
