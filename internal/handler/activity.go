package handler

import (
	"encoding/json"
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
