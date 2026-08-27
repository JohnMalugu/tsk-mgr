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