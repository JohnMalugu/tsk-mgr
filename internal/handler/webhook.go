package handler

import (
	"encoding/json"
	"errors"
	"net/http"

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

