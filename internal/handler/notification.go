package handler

import (
	"encoding/json"
	"errors"
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

func (h *NotificationHandler) HandleMarkAsRead(w http.ResponseWriter, r *http.Request, idStr string) {
	userID, ok := r.Context().Value("userID").(string)
	if !ok || userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid notification ID", http.StatusBadRequest)
		return
	}

	if err := h.svc.MarkAsRead(id, userID); err != nil {
		if errors.Is(err, service.ErrNotificationNotFound) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"message": "notification marked as read"})
}

func (h *NotificationHandler) HandleMarkAllAsRead(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(string)
	if !ok || userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if err := h.svc.MarkAllAsRead(userID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"message": "all notifications marked as read"})
}

func (h *NotificationHandler) HandleDelete(w http.ResponseWriter, r *http.Request, idStr string) {
	userID, ok := r.Context().Value("userID").(string)
	if !ok || userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid notification ID", http.StatusBadRequest)
		return
	}

	if err := h.svc.DeleteNotification(id, userID); err != nil {
		if errors.Is(err, service.ErrNotificationNotFound) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"message": "notification deleted successfully"})
}

func (h *NotificationHandler) HandleSendNotification(w http.ResponseWriter, r *http.Request) {
	actingUserID, ok := r.Context().Value("userID").(string)
	if !ok || actingUserID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req model.CreateNotificationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid input: "+err.Error(), http.StatusBadRequest)
		return
	}

	notif, err := h.svc.Send(&req)
	if err != nil {
		if errors.Is(err, service.ErrNotificationTitleEmpty) ||
			errors.Is(err, service.ErrNotificationUserIDEmpty) ||
			errors.Is(err, service.ErrInvalidNotificationType) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if notif == nil {
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "suppressed_by_preferences"})
		return
	}
	_ = json.NewEncoder(w).Encode(notif)
}

func (h *NotificationHandler) HandleGetPreferences(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(string)
	if !ok || userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	pref, err := h.svc.GetPreferences(userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(pref)
}

func (h *NotificationHandler) HandleUpdatePreferences(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(string)
	if !ok || userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req model.UpdatePreferencesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid input: "+err.Error(), http.StatusBadRequest)
		return
	}

	pref, err := h.svc.UpdatePreferences(userID, &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(pref)
}
