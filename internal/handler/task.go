package handler

// Handler holds dependencies for all handlers
type Handler struct {
	taskService *service.TaskService
}