package handler

import (
	"net/http"

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
