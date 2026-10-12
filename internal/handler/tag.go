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
