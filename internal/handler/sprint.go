package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/JohnMalugu/tsk-mgr-api/internal/model"
	"github.com/JohnMalugu/tsk-mgr-api/internal/service"
)

type SprintHandler struct {
	svc service.SprintService
}

var GlobalSprintHandler *SprintHandler

func InitSprintHandler(svc service.SprintService) {
	GlobalSprintHandler = &SprintHandler{svc: svc}
}

func NewSprintHandler(svc service.SprintService) *SprintHandler {
	return &SprintHandler{svc: svc}
}

func (h *SprintHandler) HandleCreate(w http.ResponseWriter, r *http.Request) {
	var req model.CreateSprintRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid input: "+err.Error(), http.StatusBadRequest)
		return
	}

	sprint, err := h.svc.CreateSprint(&req)
	if err != nil {
		if errors.Is(err, service.ErrSprintNameRequired) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(sprint)
}

func (h *SprintHandler) HandleGet(w http.ResponseWriter, r *http.Request, id int64) {
	sprint, err := h.svc.GetSprint(id)
	if err != nil {
		if errors.Is(err, service.ErrSprintNotFound) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(sprint)
}

func (h *SprintHandler) HandleList(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	filter := model.SprintFilter{
		Limit:  20,
		Offset: 0,
	}

	if wsStr := query.Get("workspace_id"); wsStr != "" {
		if wsID, err := strconv.ParseInt(wsStr, 10, 64); err == nil {
			filter.WorkspaceID = wsID
		}
	}
	if pStr := query.Get("project_id"); pStr != "" {
		if pID, err := strconv.ParseInt(pStr, 10, 64); err == nil {
			filter.ProjectID = pID
		}
	}
	if status := query.Get("status"); status != "" {
		filter.Status = model.SprintStatus(status)
	}
	if limitStr := query.Get("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
			filter.Limit = limit
		}
	}
	if offsetStr := query.Get("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil && offset >= 0 {
			filter.Offset = offset
		}
	}

	summary, err := h.svc.ListSprints(filter)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(summary)
}

func (h *SprintHandler) HandleUpdate(w http.ResponseWriter, r *http.Request, id int64) {
	var req model.UpdateSprintRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid input: "+err.Error(), http.StatusBadRequest)
		return
	}

	sprint, err := h.svc.UpdateSprint(id, &req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrSprintNotFound):
			http.Error(w, err.Error(), http.StatusNotFound)
		case errors.Is(err, service.ErrSprintNameRequired),
			errors.Is(err, service.ErrInvalidSprintStatus),
			errors.Is(err, service.ErrSprintClosed):
			http.Error(w, err.Error(), http.StatusBadRequest)
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(sprint)
}

func (h *SprintHandler) HandleDelete(w http.ResponseWriter, r *http.Request, id int64) {
	err := h.svc.DeleteSprint(id)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrSprintNotFound):
			http.Error(w, err.Error(), http.StatusNotFound)
		case errors.Is(err, service.ErrCannotDeleteActiveSprint):
			http.Error(w, err.Error(), http.StatusBadRequest)
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *SprintHandler) HandleStart(w http.ResponseWriter, r *http.Request, id int64) {
	sprint, err := h.svc.StartSprint(id)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrSprintNotFound):
			http.Error(w, err.Error(), http.StatusNotFound)
		case errors.Is(err, service.ErrSprintCannotStart):
			http.Error(w, err.Error(), http.StatusBadRequest)
		case errors.Is(err, service.ErrActiveSprintExists):
			http.Error(w, err.Error(), http.StatusConflict)
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(sprint)
}

func (h *SprintHandler) HandleComplete(w http.ResponseWriter, r *http.Request, id int64) {
	sprint, err := h.svc.CompleteSprint(id)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrSprintNotFound):
			http.Error(w, err.Error(), http.StatusNotFound)
		case errors.Is(err, service.ErrSprintCannotComplete):
			http.Error(w, err.Error(), http.StatusBadRequest)
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(sprint)
}

func (h *SprintHandler) HandleClose(w http.ResponseWriter, r *http.Request, id int64) {
	sprint, err := h.svc.CloseSprint(id)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrSprintNotFound):
			http.Error(w, err.Error(), http.StatusNotFound)
		case errors.Is(err, service.ErrSprintCannotClose):
			http.Error(w, err.Error(), http.StatusBadRequest)
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(sprint)
}

func (h *SprintHandler) HandleAddTask(w http.ResponseWriter, r *http.Request, sprintID int64) {
	var req model.AddTaskToSprintRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid input: "+err.Error(), http.StatusBadRequest)
		return
	}

	err := h.svc.AddTaskToSprint(sprintID, &req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrSprintNotFound):
			http.Error(w, err.Error(), http.StatusNotFound)
		case errors.Is(err, service.ErrSprintClosed), errors.Is(err, service.ErrInvalidStoryPoints):
			http.Error(w, err.Error(), http.StatusBadRequest)
		case errors.Is(err, service.ErrTaskAlreadyInSprint):
			http.Error(w, err.Error(), http.StatusConflict)
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (h *SprintHandler) HandleRemoveTask(w http.ResponseWriter, r *http.Request, sprintID, taskID int64) {
	err := h.svc.RemoveTaskFromSprint(sprintID, taskID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrSprintNotFound):
			http.Error(w, err.Error(), http.StatusNotFound)
		case errors.Is(err, service.ErrSprintClosed):
			http.Error(w, err.Error(), http.StatusBadRequest)
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *SprintHandler) HandleGetTasks(w http.ResponseWriter, r *http.Request, sprintID int64) {
	tasks, err := h.svc.GetSprintTasks(sprintID)
	if err != nil {
		if errors.Is(err, service.ErrSprintNotFound) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(tasks)
}

func (h *SprintHandler) HandleUpdateStoryPoints(w http.ResponseWriter, r *http.Request, sprintID, taskID int64) {
	var req model.UpdateStoryPointsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid input: "+err.Error(), http.StatusBadRequest)
		return
	}

	err := h.svc.UpdateTaskStoryPoints(sprintID, taskID, req.StoryPoints)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrSprintNotFound):
			http.Error(w, err.Error(), http.StatusNotFound)
		case errors.Is(err, service.ErrSprintClosed), errors.Is(err, service.ErrInvalidStoryPoints):
			http.Error(w, err.Error(), http.StatusBadRequest)
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *SprintHandler) HandleGetMetrics(w http.ResponseWriter, r *http.Request, sprintID int64) {
	metrics, err := h.svc.GetSprintMetrics(sprintID)
	if err != nil {
		if errors.Is(err, service.ErrSprintNotFound) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(metrics)
}
