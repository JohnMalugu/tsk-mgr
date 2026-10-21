package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/JohnMalugu/tsk-mgr-api/internal/model"
	"github.com/JohnMalugu/tsk-mgr-api/internal/service"
)

// AttachmentHandler handles HTTP requests for file attachments and storage metrics
type AttachmentHandler struct {
	svc service.AttachmentService
}

var GlobalAttachmentHandler *AttachmentHandler

// InitAttachmentHandler initializes the global attachment handler instance
func InitAttachmentHandler(svc service.AttachmentService) {
	GlobalAttachmentHandler = &AttachmentHandler{svc: svc}
}

// NewAttachmentHandler creates a new AttachmentHandler instance
func NewAttachmentHandler(svc service.AttachmentService) *AttachmentHandler {
	return &AttachmentHandler{svc: svc}
}

// HandleGetAttachment returns attachment metadata by ID
func (h *AttachmentHandler) HandleGetAttachment(w http.ResponseWriter, r *http.Request, id int64) {
	att, err := h.svc.GetAttachment(id)
	if err != nil {
		if errors.Is(err, service.ErrAttachmentNotFound) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(att)
}

// HandleUpdateAttachment updates attachment metadata (e.g., filename, archived status)
func (h *AttachmentHandler) HandleUpdateAttachment(w http.ResponseWriter, r *http.Request, id int64) {
	var req model.UpdateAttachmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid input: "+err.Error(), http.StatusBadRequest)
		return
	}

	att, err := h.svc.UpdateAttachment(id, &req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrAttachmentNotFound):
			http.Error(w, err.Error(), http.StatusNotFound)
		case errors.Is(err, service.ErrBlockedFileType):
			http.Error(w, err.Error(), http.StatusBadRequest)
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(att)
}

// HandleDeleteAttachment deletes an attachment record and removes its physical file
func (h *AttachmentHandler) HandleDeleteAttachment(w http.ResponseWriter, r *http.Request, id int64) {
	err := h.svc.DeleteAttachment(id)
	if err != nil {
		if errors.Is(err, service.ErrAttachmentNotFound) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
