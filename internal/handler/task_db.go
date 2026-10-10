package handler

import (
	"encoding/json"
	"net/http"
	"time"
	"strconv"

	"github.com/JohnMalugu/tsk-mgr-api/internal/service"
)

type TaskDBHandler struct {
	svc service.TaskDBService
}

func NewTaskDBHandler(svc service.TaskDBService) *TaskDBHandler {
	return &TaskDBHandler{svc: svc}
}
