package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/JohnMalugu/tsk-mgr-api/internal/model"
	"github.com/JohnMalugu/tsk-mgr-api/internal/service"
)

type NotificationHandler struct {
	svc service.NotificationService
}

var GlobalNotificationHandler *NotificationHandler

func InitNotificationHandler(svc service.NotificationService) {
	GlobalNotificationHandler = &NotificationHandler{svc: svc}
}

func NewNotificationHandler(svc service.NotificationService) *NotificationHandler {
	return &NotificationHandler{svc: svc}
}

func (h *NotificationHandler) HandleList(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(string)
	if !ok || userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	filter := model.NotificationFilter{
		UserID: userID,
	}

	if t := r.URL.Query().Get("type"); t != "" {
		filter.Type = model.NotificationType(t)
	}
	if unreadStr := r.URL.Query().Get("unreadOnly"); unreadStr == "true" || unreadStr == "1" {
		filter.UnreadOnly = true
	}
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil {
			filter.Limit = limit
		}
	}
	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil {
			filter.Offset = offset
		}
	}

	summary, err := h.svc.ListNotifications(filter)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(summary)
}

func (h *NotificationHandler) HandleGetUnreadCount(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(string)
	if !ok || userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	count, err := h.svc.GetUnreadCount(userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(model.UnreadCountResponse{UnreadCount: count})
}

