package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/JohnMalugu/tsk-mgr-api/internal/model"
	"github.com/JohnMalugu/tsk-mgr-api/internal/service"
)

type WebhookHandler struct {
	svc service.WebhookService
}

var GlobalWebhookHandler *WebhookHandler

// InitWebhookHandler initializes the global webhook handler instance
func InitWebhookHandler(svc service.WebhookService) {
	GlobalWebhookHandler = &WebhookHandler{svc: svc}
}

// NewWebhookHandler returns a new WebhookHandler instance
func NewWebhookHandler(svc service.WebhookService) *WebhookHandler {
	return &WebhookHandler{svc: svc}
}

func getWebhookUser(r *http.Request) string {
	if uid, ok := r.Context().Value("userID").(string); ok && uid != "" {
		return uid
	}
	if uid := r.Header.Get("X-User-ID"); uid != "" {
		return uid
	}
	return ""
}

// HandleCreateSubscription registers a new webhook subscription
func (h *WebhookHandler) HandleCreateSubscription(w http.ResponseWriter, r *http.Request) {
	userID := getWebhookUser(r)
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req model.CreateWebhookRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON payload: "+err.Error(), http.StatusBadRequest)
		return
	}

	sub, err := h.svc.CreateSubscription(userID, &req)
	if err != nil {
		if errors.Is(err, service.ErrInvalidTargetURL) ||
			errors.Is(err, service.ErrNoEventsSpecified) ||
			errors.Is(err, service.ErrInvalidEventType) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(sub)
}

// HandleGetSubscription retrieves a single subscription by its ID
func (h *WebhookHandler) HandleGetSubscription(w http.ResponseWriter, r *http.Request, id int64) {
	sub, err := h.svc.GetSubscription(id)
	if err != nil {
		if errors.Is(err, service.ErrWebhookNotFound) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(sub)
}

// HandleListSubscriptions lists webhook subscriptions matching optional query parameters
func (h *WebhookHandler) HandleListSubscriptions(w http.ResponseWriter, r *http.Request) {
	filter := model.WebhookFilter{}

	if wsStr := r.URL.Query().Get("workspace_id"); wsStr != "" {
		if wsID, err := strconv.ParseInt(wsStr, 10, 64); err == nil {
			filter.WorkspaceID = wsID
		}
	}
	if uid := r.URL.Query().Get("user_id"); uid != "" {
		filter.UserID = uid
	}
	if actStr := r.URL.Query().Get("is_active"); actStr != "" {
		act := actStr == "true" || actStr == "1"
		filter.IsActive = &act
	}
	if et := r.URL.Query().Get("event_type"); et != "" {
		filter.EventType = et
	}
	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if off, err := strconv.Atoi(offsetStr); err == nil {
			filter.Offset = off
		}
	}
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if lim, err := strconv.Atoi(limitStr); err == nil {
			filter.Limit = lim
		}
	}

	summary, err := h.svc.ListSubscriptions(filter)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(summary)
}

// HandleUpdateSubscription modifies an existing subscription
func (h *WebhookHandler) HandleUpdateSubscription(w http.ResponseWriter, r *http.Request, id int64) {
	userID := getWebhookUser(r)

	var req model.UpdateWebhookRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON payload: "+err.Error(), http.StatusBadRequest)
		return
	}

	sub, err := h.svc.UpdateSubscription(id, userID, &req)
	if err != nil {
		if errors.Is(err, service.ErrWebhookNotFound) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		if errors.Is(err, service.ErrUnauthorizedWebhookAccess) {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}
		if errors.Is(err, service.ErrInvalidTargetURL) ||
			errors.Is(err, service.ErrNoEventsSpecified) ||
			errors.Is(err, service.ErrInvalidEventType) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(sub)
}

// HandleDeleteSubscription removes an existing subscription
func (h *WebhookHandler) HandleDeleteSubscription(w http.ResponseWriter, r *http.Request, id int64) {
	userID := getWebhookUser(r)

	if err := h.svc.DeleteSubscription(id, userID); err != nil {
		if errors.Is(err, service.ErrWebhookNotFound) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		if errors.Is(err, service.ErrUnauthorizedWebhookAccess) {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"message": "Webhook subscription deleted successfully"})
}

// HandlePingSubscription triggers a live test ping delivery to the subscription's target URL
func (h *WebhookHandler) HandlePingSubscription(w http.ResponseWriter, r *http.Request, id int64) {
	userID := getWebhookUser(r)

	delivery, err := h.svc.PingWebhook(id, userID)
	if err != nil {
		if errors.Is(err, service.ErrWebhookNotFound) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		if errors.Is(err, service.ErrUnauthorizedWebhookAccess) {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(delivery)
}

// HandleListDeliveries returns delivery records for a webhook subscription
func (h *WebhookHandler) HandleListDeliveries(w http.ResponseWriter, r *http.Request, subscriptionID int64) {
	filter := model.DeliveryFilter{
		SubscriptionID: subscriptionID,
	}

	if status := r.URL.Query().Get("status"); status != "" {
		filter.Status = status
	}
	if et := r.URL.Query().Get("event_type"); et != "" {
		filter.EventType = et
	}
	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if off, err := strconv.Atoi(offsetStr); err == nil {
			filter.Offset = off
		}
	}
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if lim, err := strconv.Atoi(limitStr); err == nil {
			filter.Limit = lim
		}
	}

	summary, err := h.svc.ListDeliveries(filter)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(summary)
}

// HandleRetryDelivery re-attempts a previously recorded delivery
func (h *WebhookHandler) HandleRetryDelivery(w http.ResponseWriter, r *http.Request, deliveryID int64) {
	delivery, err := h.svc.RetryDelivery(deliveryID)
	if err != nil {
		if errors.Is(err, service.ErrDeliveryNotFound) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		if errors.Is(err, service.ErrWebhookNotFound) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(delivery)
}

// HandleGetStats returns delivery analytics and metrics for a subscription
func (h *WebhookHandler) HandleGetStats(w http.ResponseWriter, r *http.Request, subscriptionID int64) {
	stats, err := h.svc.GetWebhookStats(subscriptionID)
	if err != nil {
		if errors.Is(err, service.ErrWebhookNotFound) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(stats)
}
