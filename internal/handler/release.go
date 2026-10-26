package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/JohnMalugu/tsk-mgr-api/internal/model"
	"github.com/JohnMalugu/tsk-mgr-api/internal/service"
)

// ReleaseHandler provides HTTP controllers for project releases and milestones
type ReleaseHandler struct {
	svc service.ReleaseService
}

var GlobalReleaseHandler *ReleaseHandler

// InitReleaseHandler initializes the global ReleaseHandler singleton
func InitReleaseHandler(svc service.ReleaseService) {
	GlobalReleaseHandler = &ReleaseHandler{svc: svc}
}

// NewReleaseHandler constructs a new ReleaseHandler
func NewReleaseHandler(svc service.ReleaseService) *ReleaseHandler {
	return &ReleaseHandler{svc: svc}
}

func getReleaseUser(r *http.Request) string {
	if uid, ok := r.Context().Value("userID").(string); ok && uid != "" {
		return uid
	}
	if uid := r.Header.Get("X-User-ID"); uid != "" {
		return uid
	}
	if uid := r.URL.Query().Get("user_id"); uid != "" {
		return uid
	}
	return "default_user"
}

func extractReleaseID(path, prefix string) (int64, error) {
	trimmed := strings.TrimPrefix(path, prefix)
	parts := strings.Split(strings.Trim(trimmed, "/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		return 0, errors.New("missing release ID in path")
	}
	id, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return 0, errors.New("invalid release ID format")
	}
	return id, nil
}

func writeReleaseJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(data)
}

func writeReleaseError(w http.ResponseWriter, statusCode int, message string) {
	writeReleaseJSON(w, statusCode, map[string]string{"error": message})
}

// HandleReleasesRoot routes /v2/releases GET (list) and POST (create)
func (h *ReleaseHandler) HandleReleasesRoot(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.HandleListReleases(w, r)
	case http.MethodPost:
		h.HandleCreateRelease(w, r)
	default:
		writeReleaseError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

// HandleCreateRelease provisions a new release milestone
func (h *ReleaseHandler) HandleCreateRelease(w http.ResponseWriter, r *http.Request) {
	userID := getReleaseUser(r)

	var req model.CreateReleaseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeReleaseError(w, http.StatusBadRequest, "Invalid request body: "+err.Error())
		return
	}

	rel, err := h.svc.CreateRelease(userID, req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrWorkspaceRequired),
			errors.Is(err, service.ErrReleaseNameRequired),
			errors.Is(err, service.ErrInvalidReleaseVersion):
			writeReleaseError(w, http.StatusBadRequest, err.Error())
		default:
			writeReleaseError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	writeReleaseJSON(w, http.StatusCreated, rel)
}

// HandleListReleases retrieves summarized releases with filtering
func (h *ReleaseHandler) HandleListReleases(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	var wsID int64
	if wsStr := q.Get("workspace_id"); wsStr != "" {
		wsID, _ = strconv.ParseInt(wsStr, 10, 64)
	}

	var projID *int64
	if projStr := q.Get("project_id"); projStr != "" {
		if pid, err := strconv.ParseInt(projStr, 10, 64); err == nil {
			projID = &pid
		}
	}

	limit := 50
	if lStr := q.Get("limit"); lStr != "" {
		if l, err := strconv.Atoi(lStr); err == nil && l > 0 {
			limit = l
		}
	}

	offset := 0
	if oStr := q.Get("offset"); oStr != "" {
		if o, err := strconv.Atoi(oStr); err == nil && o >= 0 {
			offset = o
		}
	}

	filter := model.ReleaseFilter{
		WorkspaceID: wsID,
		ProjectID:   projID,
		Status:      q.Get("status"),
		Query:       q.Get("q"),
		Limit:       limit,
		Offset:      offset,
	}

	summaries, err := h.svc.ListReleases(filter)
	if err != nil {
		if errors.Is(err, service.ErrWorkspaceRequired) {
			writeReleaseError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeReleaseError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if summaries == nil {
		summaries = []*model.ReleaseSummary{}
	}

	writeReleaseJSON(w, http.StatusOK, map[string]interface{}{
		"releases": summaries,
		"count":    len(summaries),
	})
}

// HandleReleaseDetail routes /v2/releases/{id} GET, PUT, and DELETE
func (h *ReleaseHandler) HandleReleaseDetail(w http.ResponseWriter, r *http.Request) {
	id, err := extractReleaseID(r.URL.Path, "/v2/releases/")
	if err != nil {
		writeReleaseError(w, http.StatusBadRequest, err.Error())
		return
	}

	switch r.Method {
	case http.MethodGet:
		h.HandleGetRelease(w, r, id)
	case http.MethodPut:
		h.HandleUpdateRelease(w, r, id)
	case http.MethodDelete:
		h.HandleDeleteRelease(w, r, id)
	default:
		writeReleaseError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

// HandleGetRelease returns single release details
func (h *ReleaseHandler) HandleGetRelease(w http.ResponseWriter, r *http.Request, id int64) {
	rel, err := h.svc.GetRelease(id)
	if err != nil {
		if errors.Is(err, service.ErrReleaseNotFound) {
			writeReleaseError(w, http.StatusNotFound, "Release not found")
			return
		}
		writeReleaseError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeReleaseJSON(w, http.StatusOK, rel)
}

// HandleUpdateRelease modifies release metadata
func (h *ReleaseHandler) HandleUpdateRelease(w http.ResponseWriter, r *http.Request, id int64) {
	var req model.UpdateReleaseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeReleaseError(w, http.StatusBadRequest, "Invalid request body: "+err.Error())
		return
	}

	updated, err := h.svc.UpdateRelease(id, req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrReleaseNotFound):
			writeReleaseError(w, http.StatusNotFound, "Release not found")
		case errors.Is(err, service.ErrReleaseAlreadyReleased),
			errors.Is(err, service.ErrReleaseNameRequired),
			errors.Is(err, service.ErrInvalidReleaseVersion):
			writeReleaseError(w, http.StatusBadRequest, err.Error())
		default:
			writeReleaseError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	writeReleaseJSON(w, http.StatusOK, updated)
}

// HandleDeleteRelease deletes a release milestone
func (h *ReleaseHandler) HandleDeleteRelease(w http.ResponseWriter, r *http.Request, id int64) {
	err := h.svc.DeleteRelease(id)
	if err != nil {
		if errors.Is(err, service.ErrReleaseNotFound) {
			writeReleaseError(w, http.StatusNotFound, "Release not found")
			return
		}
		writeReleaseError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeReleaseJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Release deleted successfully",
		"id":      id,
	})
}

// HandleUpdateReleaseStatus transitions release lifecycle status (e.g. code_freeze, released)
func (h *ReleaseHandler) HandleUpdateReleaseStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodPatch {
		writeReleaseError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	id, err := extractReleaseID(r.URL.Path, "/v2/releases/")
	if err != nil {
		writeReleaseError(w, http.StatusBadRequest, err.Error())
		return
	}

	var req model.UpdateReleaseStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeReleaseError(w, http.StatusBadRequest, "Invalid request body: "+err.Error())
		return
	}

	updated, err := h.svc.UpdateReleaseStatus(id, req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrReleaseNotFound):
			writeReleaseError(w, http.StatusNotFound, "Release not found")
		case errors.Is(err, service.ErrInvalidReleaseStatus),
			errors.Is(err, service.ErrInvalidStatusTransition):
			writeReleaseError(w, http.StatusBadRequest, err.Error())
		default:
			writeReleaseError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	writeReleaseJSON(w, http.StatusOK, updated)
}

func extractReleaseAndTaskIDs(path string) (int64, int, error) {
	trimmed := strings.TrimPrefix(path, "/v2/releases/")
	parts := strings.Split(strings.Trim(trimmed, "/"), "/")
	if len(parts) < 3 || parts[1] != "tasks" {
		return 0, 0, errors.New("invalid path format, expected /v2/releases/{id}/tasks/{taskId}")
	}
	relID, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return 0, 0, errors.New("invalid release ID")
	}
	taskID, err := strconv.Atoi(parts[2])
	if err != nil {
		return 0, 0, errors.New("invalid task ID")
	}
	return relID, taskID, nil
}

// HandleReleaseTasks routes GET (list tasks) and POST (attach tasks) for /v2/releases/{id}/tasks
func (h *ReleaseHandler) HandleReleaseTasks(w http.ResponseWriter, r *http.Request) {
	id, err := extractReleaseID(r.URL.Path, "/v2/releases/")
	if err != nil {
		writeReleaseError(w, http.StatusBadRequest, err.Error())
		return
	}

	switch r.Method {
	case http.MethodGet:
		cat := r.URL.Query().Get("category")
		items, err := h.svc.ListReleaseItems(id, cat)
		if err != nil {
			if errors.Is(err, service.ErrReleaseNotFound) {
				writeReleaseError(w, http.StatusNotFound, "Release not found")
				return
			}
			writeReleaseError(w, http.StatusInternalServerError, err.Error())
			return
		}
		if items == nil {
			items = []*model.ReleaseItem{}
		}
		writeReleaseJSON(w, http.StatusOK, map[string]interface{}{
			"release_id": id,
			"tasks":      items,
			"count":      len(items),
		})

	case http.MethodPost:
		userID := getReleaseUser(r)
		var req model.AddReleaseTasksRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeReleaseError(w, http.StatusBadRequest, "Invalid request body: "+err.Error())
			return
		}

		items, err := h.svc.AddTasksToRelease(userID, id, req)
		if err != nil {
			switch {
			case errors.Is(err, service.ErrReleaseNotFound):
				writeReleaseError(w, http.StatusNotFound, "Release not found")
			case errors.Is(err, service.ErrReleaseCodeFrozen),
				errors.Is(err, service.ErrReleaseAlreadyReleased),
				errors.Is(err, service.ErrInvalidReleaseCategory),
				errors.Is(err, service.ErrNoTasksSpecified):
				writeReleaseError(w, http.StatusBadRequest, err.Error())
			default:
				writeReleaseError(w, http.StatusInternalServerError, err.Error())
			}
			return
		}

		writeReleaseJSON(w, http.StatusOK, map[string]interface{}{
			"message": "Tasks attached to release successfully",
			"tasks":   items,
			"count":   len(items),
		})

	default:
		writeReleaseError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

// HandleReleaseTaskDetail routes PUT/PATCH (update) and DELETE (detach) for /v2/releases/{id}/tasks/{taskId}
func (h *ReleaseHandler) HandleReleaseTaskDetail(w http.ResponseWriter, r *http.Request) {
	relID, taskID, err := extractReleaseAndTaskIDs(r.URL.Path)
	if err != nil {
		writeReleaseError(w, http.StatusBadRequest, err.Error())
		return
	}

	switch r.Method {
	case http.MethodPut, http.MethodPatch:
		var req model.UpdateReleaseItemRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeReleaseError(w, http.StatusBadRequest, "Invalid request body: "+err.Error())
			return
		}

		updated, err := h.svc.UpdateReleaseItem(relID, taskID, req)
		if err != nil {
			switch {
			case errors.Is(err, service.ErrReleaseNotFound), errors.Is(err, service.ErrReleaseItemNotFound):
				writeReleaseError(w, http.StatusNotFound, err.Error())
			case errors.Is(err, service.ErrReleaseAlreadyReleased), errors.Is(err, service.ErrInvalidReleaseCategory):
				writeReleaseError(w, http.StatusBadRequest, err.Error())
			default:
				writeReleaseError(w, http.StatusInternalServerError, err.Error())
			}
			return
		}

		writeReleaseJSON(w, http.StatusOK, updated)

	case http.MethodDelete:
		err := h.svc.RemoveTaskFromRelease(relID, taskID)
		if err != nil {
			switch {
			case errors.Is(err, service.ErrReleaseNotFound), errors.Is(err, service.ErrReleaseItemNotFound):
				writeReleaseError(w, http.StatusNotFound, err.Error())
			case errors.Is(err, service.ErrReleaseCodeFrozen), errors.Is(err, service.ErrReleaseAlreadyReleased):
				writeReleaseError(w, http.StatusBadRequest, err.Error())
			default:
				writeReleaseError(w, http.StatusInternalServerError, err.Error())
			}
			return
		}

		writeReleaseJSON(w, http.StatusOK, map[string]interface{}{
			"message":    "Task detached from release successfully",
			"release_id": relID,
			"task_id":    taskID,
		})

	default:
		writeReleaseError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}


