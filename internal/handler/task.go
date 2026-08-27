package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/yourusername/task-manager-api/internal/error"
	"github.com/yourusername/task-manager-api/internal/model"
	"github.com/yourusername/task-manager-api/internal/service"
	"github.com/yourusername/task-manager-api/internal/validation"
)

// Handler holds dependencies for all handlers
type Handler struct {
	taskService *service.TaskService
}

// NewHandler creates a new Handler
func NewHandler(taskService *service.TaskService) *Handler {
	return &Handler{
		taskService: taskService,
	}
}

// HandleTasks handles GET and POST requests for /tasks
func (h *Handler) HandleTasks(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		h.HandleGetAllTasks(w, r)
		return
	}

	if r.Method == http.MethodPost {
		h.HandleCreateTask(w, r)
		return
	}

	appErr := &apierror.AppError{
		Code:    http.StatusMethodNotAllowed,
		Message: "Method not allowed",
	}
	apierror.RespondWithError(w, r, appErr)
}

// HandleTasks handles GET and POST requests for /tasks
