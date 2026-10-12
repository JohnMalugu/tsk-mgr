package handler

import (
	"encoding/json"
	"net/http"

	"github.com/JohnMalugu/tsk-mgr-api/internal/service"
)

type TagHandler struct {
	svc service.TagService
}

var GlobalTagHandler *TagHandler

func InitTagHandler(svc service.TagService) {
	GlobalTagHandler = &TagHandler{svc: svc}
}

func NewTagHandler(svc service.TagService) *TagHandler {
	return &TagHandler{svc: svc}
}

type CreateTagRequest struct {
	Name  string `json:"name"`
	Color string `json:"color"`
}

func (h *TagHandler) HandleCreate(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req CreateTagRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	tag, err := h.svc.CreateTag(userID, req.Name, req.Color)
	if err != nil {
		if err == service.ErrTagNameRequired {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(tag)
}

func (h *TagHandler) HandleList(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	tags, err := h.svc.GetUserTags(userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tags)
}

func (h *TagHandler) HandleGetByID(w http.ResponseWriter, r *http.Request, tagID int) {
	userID, ok := r.Context().Value("userID").(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	tag, err := h.svc.GetTag(tagID, userID)
	if err != nil {
		if err == service.ErrTagNotFound {
			http.Error(w, "Tag not found", http.StatusNotFound)
			return
		}
		if err == service.ErrUnauthorizedTag {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tag)
}

type UpdateTagRequest struct {
	Name  string `json:"name"`
	Color string `json:"color"`
}

func (h *TagHandler) HandleUpdate(w http.ResponseWriter, r *http.Request, tagID int) {
	userID, ok := r.Context().Value("userID").(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req UpdateTagRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	tag, err := h.svc.UpdateTag(tagID, userID, req.Name, req.Color)
	if err != nil {
		if err == service.ErrTagNameRequired {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err == service.ErrTagNotFound {
			http.Error(w, "Tag not found", http.StatusNotFound)
			return
		}
		if err == service.ErrUnauthorizedTag {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tag)
}

func (h *TagHandler) HandleDelete(w http.ResponseWriter, r *http.Request, tagID int) {
	userID, ok := r.Context().Value("userID").(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	err := h.svc.DeleteTag(tagID, userID)
	if err != nil {
		if err == service.ErrTagNotFound {
			http.Error(w, "Tag not found", http.StatusNotFound)
			return
		}
		if err == service.ErrUnauthorizedTag {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

type AttachTagRequest struct {
	TagID int `json:"tag_id"`
}

func (h *TagHandler) HandleAttachTag(w http.ResponseWriter, r *http.Request, taskID int) {
	userID, ok := r.Context().Value("userID").(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req AttachTagRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	err := h.svc.AddTagToTask(taskID, req.TagID, userID)
	if err != nil {
		if err == service.ErrTagNotFound || err == service.ErrTaskNotFoundDB {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		if err == service.ErrUnauthorizedTag || err == service.ErrUnauthorizedTask {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (h *TagHandler) HandleDetachTag(w http.ResponseWriter, r *http.Request, taskID, tagID int) {
	userID, ok := r.Context().Value("userID").(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	err := h.svc.RemoveTagFromTask(taskID, tagID, userID)
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

	w.WriteHeader(http.StatusNoContent)
}

func (h *TagHandler) HandleGetTaskTags(w http.ResponseWriter, r *http.Request, taskID int) {
	userID, ok := r.Context().Value("userID").(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	tags, err := h.svc.GetTaskTags(taskID, userID)
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
	json.NewEncoder(w).Encode(tags)
}
