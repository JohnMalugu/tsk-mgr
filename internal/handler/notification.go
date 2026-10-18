package handler

import (
	"github.com/JohnMalugu/tsk-mgr-api/internal/service"
)

type NotificationHandler struct {
	svc service.NotificationService
}

var GlobalNotificationHandler *NotificationHandler

func InitNotificationHandler(svc service.NotificationService) {
	GlobalNotificationHandler = &NotificationHandler{svc: svc}
}

func NewNotificationHandler(svc service.NotificationService) *NotificationHandler {
	return &NotificationHandler{svc: svc}
}
