package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

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

// HandleUploadAttachment handles both JSON (base64) and multipart/form-data file uploads
func (h *AttachmentHandler) HandleUploadAttachment(w http.ResponseWriter, r *http.Request, taskID int64) {
	var req model.UploadAttachmentRequest
	req.TaskID = taskID

	contentType := r.Header.Get("Content-Type")
	if strings.HasPrefix(contentType, "multipart/form-data") {
		err := r.ParseMultipartForm(model.MaxAttachmentSizeBytes)
		if err != nil {
			http.Error(w, "Failed to parse multipart form: "+err.Error(), http.StatusBadRequest)
			return
		}

		file, header, err := r.FormFile("file")
		if err != nil {
			http.Error(w, "Missing 'file' field in multipart form", http.StatusBadRequest)
			return
		}
		defer file.Close()

		rawBytes, err := io.ReadAll(file)
		if err != nil {
			http.Error(w, "Failed to read uploaded file: "+err.Error(), http.StatusBadRequest)
			return
		}

		req.RawBytes = rawBytes
		req.Filename = header.Filename
		req.ContentType = header.Header.Get("Content-Type")

		if wsStr := r.FormValue("workspace_id"); wsStr != "" {
			if wsID, err := strconv.ParseInt(wsStr, 10, 64); err == nil {
				req.WorkspaceID = wsID
			}
		}
		if uID := r.FormValue("user_id"); uID != "" {
			req.UserID = uID
		}
	} else {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid input: "+err.Error(), http.StatusBadRequest)
			return
		}
		if req.TaskID == 0 {
			req.TaskID = taskID
		}
	}

	att, err := h.svc.UploadAttachment(&req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidTaskID),
			errors.Is(err, service.ErrInvalidWorkspaceID),
			errors.Is(err, service.ErrInvalidUserID),
			errors.Is(err, service.ErrInvalidFilename),
			errors.Is(err, service.ErrEmptyFile),
			errors.Is(err, service.ErrFileTooLarge),
			errors.Is(err, service.ErrBlockedFileType):
			http.Error(w, err.Error(), http.StatusBadRequest)
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(att)
}

// HandleDownloadAttachment streams the raw binary file with caching headers (ETag, If-None-Match) and Content-Disposition
func (h *AttachmentHandler) HandleDownloadAttachment(w http.ResponseWriter, r *http.Request, id int64) {
	reader, att, etag, err := h.svc.OpenAttachmentFile(id)
	if err != nil {
		if errors.Is(err, service.ErrAttachmentNotFound) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer reader.Close()

	if clientETag := r.Header.Get("If-None-Match"); clientETag != "" && clientETag == etag {
		w.WriteHeader(http.StatusNotModified)
		return
	}

	w.Header().Set("ETag", etag)
	w.Header().Set("Content-Type", att.ContentType)
	w.Header().Set("Content-Length", strconv.FormatInt(att.FileSize, 10))
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", att.Filename))

	w.WriteHeader(http.StatusOK)
	_, _ = io.Copy(w, reader)
}

// HandleListAttachments returns paginated attachments filtered by task, workspace, user, or mime type
func (h *AttachmentHandler) HandleListAttachments(w http.ResponseWriter, r *http.Request, taskID *int64) {
	q := r.URL.Query()
	filter := model.AttachmentFilter{}

	if taskID != nil && *taskID > 0 {
		filter.TaskID = taskID
	} else if tIDStr := q.Get("task_id"); tIDStr != "" {
		if tID, err := strconv.ParseInt(tIDStr, 10, 64); err == nil {
			filter.TaskID = &tID
		}
	}

	if wsStr := q.Get("workspace_id"); wsStr != "" {
		if wsID, err := strconv.ParseInt(wsStr, 10, 64); err == nil {
			filter.WorkspaceID = &wsID
		}
	}

	if uID := q.Get("user_id"); uID != "" {
		filter.UserID = &uID
	}

	if ct := q.Get("content_type"); ct != "" {
		filter.ContentType = &ct
	}

	if archStr := q.Get("is_archived"); archStr != "" {
		if arch, err := strconv.ParseBool(archStr); err == nil {
			filter.IsArchived = &arch
		}
	}

	if limitStr := q.Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			filter.Limit = l
		}
	}

	if offsetStr := q.Get("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil {
			filter.Offset = o
		}
	}

	summary, err := h.svc.ListAttachments(filter)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(summary)
}

// HandleGetStorageMetrics returns storage consumption analytics for a workspace
func (h *AttachmentHandler) HandleGetStorageMetrics(w http.ResponseWriter, r *http.Request, workspaceID int64) {
	if workspaceID <= 0 {
		if wsStr := r.URL.Query().Get("workspace_id"); wsStr != "" {
			if wsID, err := strconv.ParseInt(wsStr, 10, 64); err == nil {
				workspaceID = wsID
			}
		}
	}

	if workspaceID <= 0 {
		http.Error(w, service.ErrInvalidWorkspaceID.Error(), http.StatusBadRequest)
		return
	}

	metrics, err := h.svc.GetStorageMetrics(workspaceID)
	if err != nil {
		if errors.Is(err, service.ErrInvalidWorkspaceID) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(metrics)
}
