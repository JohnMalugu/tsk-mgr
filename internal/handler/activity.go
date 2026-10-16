package handler

import (
	"github.com/JohnMalugu/tsk-mgr-api/internal/service"
)

type ActivityHandler struct {
	svc service.ActivityService
}

var GlobalActivityHandler *ActivityHandler

func InitActivityHandler(svc service.ActivityService) {
	GlobalActivityHandler = &ActivityHandler{svc: svc}
}

func NewActivityHandler(svc service.ActivityService) *ActivityHandler {
	return &ActivityHandler{svc: svc}
}
