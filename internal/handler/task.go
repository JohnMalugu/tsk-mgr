package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	appError "github.com/JohnMalugu/tsk-mgr-api/internal/error"
	"github.com/JohnMalugu/tsk-mgr-api/internal/model"
	"github.com/JohnMalugu/tsk-mgr-api/internal/service"
	"github.com/JohnMalugu/tsk-mgr-api/internal/validation"
)

// HandleTasks handles collection requests for /tasks.
func HandleTasks(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		completed, err := completionFilter(r)
		if err != nil {
			respondError(w, r, http.StatusBadRequest, "completed must be true or false")
			return
		}
		overdue, err := overdueFilter(r)
		if err != nil {
			respondError(w, r, http.StatusBadRequest, "overdue must be true or false")
			return
		}
		recurring, err := recurringFilter(r)
		if err != nil {
			respondError(w, r, http.StatusBadRequest, "recurring must be true or false")
			return
		}
		priority, err := priorityFilter(r)
		if err != nil {
			respondError(w, r, http.StatusBadRequest, "priority must be low, medium, or high")
			return
		}
		tag, err := tagFilter(r)
		if err != nil {
			respondError(w, r, http.StatusBadRequest, "tag filter is invalid")
			return
		}
		dueAfter, err := dateFilter(r, "dueAfter")
		if err != nil {
			respondError(w, r, http.StatusBadRequest, "dueAfter must be an RFC3339 timestamp")
			return
		}
		dueBefore, err := dateFilter(r, "dueBefore")
		if err != nil {
			respondError(w, r, http.StatusBadRequest, "dueBefore must be an RFC3339 timestamp")
			return
		}
		if dueAfter != nil && dueBefore != nil && dueAfter.After(*dueBefore) {
			respondError(w, r, http.StatusBadRequest, "dueAfter must not be later than dueBefore")
			return
		}
		offset, limit, err := pagination(r)
		if err != nil {
			respondError(w, r, http.StatusBadRequest, err.Error())
			return
		}
		sortBy, descending, err := sorting(r)
		if err != nil {
			respondError(w, r, http.StatusBadRequest, err.Error())
			return
		}
		respondJSON(w, http.StatusOK, service.GetTasks(completed, overdue, recurring, r.URL.Query().Get("q"), priority, tag, dueAfter, dueBefore, offset, limit, sortBy, descending))
	case http.MethodPost:
		createTask(w, r)
	default:
		w.Header().Set("Allow", http.MethodGet+", "+http.MethodPost)
		respondError(w, r, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func dateFilter(r *http.Request, name string) (*time.Time, error) {
	value := r.URL.Query().Get(name)
	if value == "" {
		return nil, nil
	}
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return nil, err
	}
	return &parsed, nil
}

func pagination(r *http.Request) (int, int, error) {
	offset, err := queryInt(r, "offset", 0)
	if err != nil || offset < 0 {
		return 0, 0, fmt.Errorf("offset must be a non-negative integer")
	}
	limit, err := queryInt(r, "limit", 20)
	if err != nil || limit < 1 || limit > 100 {
		return 0, 0, fmt.Errorf("limit must be between 1 and 100")
	}
	return offset, limit, nil
}

func queryInt(r *http.Request, name string, fallback int) (int, error) {
	value := r.URL.Query().Get(name)
	if value == "" {
		return fallback, nil
	}
	return strconv.Atoi(value)
}

func sorting(r *http.Request) (string, bool, error) {
	sortBy := r.URL.Query().Get("sort")
	if sortBy == "" {
		sortBy = "id"
	}
	if sortBy != "id" && sortBy != "title" && sortBy != "dueDate" && sortBy != "priority" && sortBy != "completed" {
		return "", false, fmt.Errorf("sort must be id, title, dueDate, priority, or completed")
	}

	order := r.URL.Query().Get("order")
	if order == "" || order == "asc" {
		return sortBy, false, nil
	}
	if order == "desc" {
		return sortBy, true, nil
	}
	return "", false, fmt.Errorf("order must be asc or desc")
}

func completionFilter(r *http.Request) (*bool, error) {
	value := r.URL.Query().Get("completed")
	if value == "" {
		return nil, nil
	}

	completed, err := strconv.ParseBool(value)
	if err != nil {
		return nil, err
	}
	return &completed, nil
}

func overdueFilter(r *http.Request) (*bool, error) {
	value := r.URL.Query().Get("overdue")
	if value == "" {
		return nil, nil
	}
	overdue, err := strconv.ParseBool(value)
	if err != nil {
		return nil, err
	}
	return &overdue, nil
}

func recurringFilter(r *http.Request) (*bool, error) {
	value := r.URL.Query().Get("recurring")
	if value == "" {
		return nil, nil
	}
	recurring, err := strconv.ParseBool(value)
	if err != nil {
		return nil, err
	}
	return &recurring, nil
}

func priorityFilter(r *http.Request) (*string, error) {
	value := r.URL.Query().Get("priority")
	if value == "" {
		return nil, nil
	}

	switch strings.ToLower(value) {
	case "low", "medium", "high":
		return &value, nil
	default:
		return nil, fmt.Errorf("invalid priority")
	}
}

func tagFilter(r *http.Request) (*string, error) {
	value := strings.TrimSpace(r.URL.Query().Get("tag"))
	if value == "" {
		return nil, nil
	}
	return &value, nil
}

// HandleTaskByID handles requests for /tasks/{id}.
func HandleTaskByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodPut && r.Method != http.MethodPatch && r.Method != http.MethodDelete {
		w.Header().Set("Allow", http.MethodGet+", "+http.MethodPut+", "+http.MethodPatch+", "+http.MethodDelete)
		respondError(w, r, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	idText := strings.TrimPrefix(r.URL.Path, "/tasks/")
	id, err := strconv.Atoi(idText)
	if err != nil || id < 1 {
		respondError(w, r, http.StatusBadRequest, "task id must be a positive integer")
		return
	}

	switch r.Method {
	case http.MethodGet:
		task := service.GetTaskByID(id)
		if task == nil {
			respondError(w, r, http.StatusNotFound, "task not found")
			return
		}
		respondJSON(w, http.StatusOK, task)
	case http.MethodPut:
		updateTask(w, r, id)
	case http.MethodPatch:
		patchTask(w, r, id)
	case http.MethodDelete:
		if !service.DeleteTask(id) {
			respondError(w, r, http.StatusNotFound, "task not found")
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func createTask(w http.ResponseWriter, r *http.Request) {
	task, ok := decodeTask(w, r)
	if !ok {
		return
	}
	created, err := service.CreateTaskWithDependencies(task, task.DependsOn)
	if err != nil {
		respondDependencyError(w, r, err)
		return
	}
	respondJSON(w, http.StatusCreated, created)
}

func HandleTaskComplete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
		w.Header().Set("Allow", http.MethodPatch)
		respondError(w, r, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	idText := strings.TrimPrefix(r.URL.Path, "/tasks/")
	idText = strings.TrimSuffix(idText, "/complete")
	id, err := strconv.Atoi(idText)
	if err != nil || id < 1 {
		respondError(w, r, http.StatusBadRequest, "task id must be a positive integer")
		return
	}

	completed := true
	if r.Body != nil {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			respondError(w, r, http.StatusBadRequest, "request body must be valid JSON")
			return
		}
		if len(bytes.TrimSpace(body)) > 0 {
			var payload struct {
				Completed *bool `json:"completed"`
			}
			if err := json.Unmarshal(body, &payload); err != nil {
				respondError(w, r, http.StatusBadRequest, "request body must be valid JSON")
				return
			}
			if payload.Completed != nil {
				completed = *payload.Completed
			}
		}
	}

	updated, err := service.SetTaskCompletionChecked(id, completed)
	if err != nil {
		if errors.Is(err, service.ErrTaskBlocked) {
			respondError(w, r, http.StatusConflict, err.Error())
		} else {
			respondError(w, r, http.StatusNotFound, err.Error())
		}
		return
	}
	respondJSON(w, http.StatusOK, updated)
}

func HandleTaskSummary(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		respondError(w, r, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	respondJSON(w, http.StatusOK, service.GetTaskSummary())
}

func HandleReadyTasks(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		respondError(w, r, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	respondJSON(w, http.StatusOK, service.GetReadyTasks())
}

func HandleTaskDependencies(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodGet+", "+http.MethodPost)
		respondError(w, r, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	path := strings.TrimSuffix(r.URL.Path, "/dependencies")
	id, err := strconv.Atoi(strings.TrimPrefix(path, "/tasks/"))
	if err != nil || id < 1 {
		respondError(w, r, http.StatusBadRequest, "task id must be a positive integer")
		return
	}
	if r.Method == http.MethodPost {
		var payload struct {
			DependsOn int `json:"dependsOn"`
		}
		decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&payload); err != nil || payload.DependsOn < 1 {
			respondError(w, r, http.StatusBadRequest, "dependsOn must be a positive task id")
			return
		}
		task, err := service.AddTaskDependency(id, payload.DependsOn)
		if err != nil {
			switch {
			case errors.Is(err, service.ErrTaskNotFound):
				respondError(w, r, http.StatusNotFound, err.Error())
			case errors.Is(err, service.ErrDependencySelfReference), errors.Is(err, service.ErrDependencyAlreadyExists), errors.Is(err, service.ErrDependencyCycle):
				respondError(w, r, http.StatusConflict, err.Error())
			default:
				respondError(w, r, http.StatusInternalServerError, "could not add dependency")
			}
			return
		}
		respondJSON(w, http.StatusCreated, task)
		return
	}
	dependencies, found := service.GetTaskDependencies(id)
	if !found {
		respondError(w, r, http.StatusNotFound, "task not found")
		return
	}
	respondJSON(w, http.StatusOK, dependencies)
}

func HandleTaskDependency(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		w.Header().Set("Allow", http.MethodDelete)
		respondError(w, r, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/tasks/"), "/dependencies/")
	if len(parts) != 2 {
		respondError(w, r, http.StatusBadRequest, "dependency route is invalid")
		return
	}
	taskID, taskErr := strconv.Atoi(parts[0])
	dependencyID, dependencyErr := strconv.Atoi(parts[1])
	if taskErr != nil || dependencyErr != nil || taskID < 1 || dependencyID < 1 {
		respondError(w, r, http.StatusBadRequest, "task ids must be positive integers")
		return
	}
	_, err := service.RemoveTaskDependency(taskID, dependencyID)
	if err != nil {
		if errors.Is(err, service.ErrDependencyNotFound) {
			respondError(w, r, http.StatusNotFound, err.Error())
		} else {
			respondError(w, r, http.StatusNotFound, err.Error())
		}
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func HandleTaskChecklist(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/tasks/")
	parts := strings.Split(path, "/checklist")
	if len(parts) != 2 {
		respondError(w, r, http.StatusBadRequest, "checklist route is invalid")
		return
	}
	taskID, err := strconv.Atoi(parts[0])
	if err != nil || taskID < 1 {
		respondError(w, r, http.StatusBadRequest, "task id must be a positive integer")
		return
	}
	if parts[1] != "" {
		handleChecklistItem(w, r, taskID, strings.TrimPrefix(parts[1], "/"))
		return
	}
	switch r.Method {
	case http.MethodGet:
		items, found := service.GetTaskChecklist(taskID)
		if !found {
			respondError(w, r, http.StatusNotFound, "task not found")
			return
		}
		respondJSON(w, http.StatusOK, items)
	case http.MethodPost:
		var payload struct {
			Text string `json:"text"`
		}
		decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&payload); err != nil {
			respondError(w, r, http.StatusBadRequest, "request body must be valid JSON")
			return
		}
		item, err := service.AddChecklistItem(taskID, payload.Text)
		if err != nil {
			respondChecklistError(w, r, err)
			return
		}
		respondJSON(w, http.StatusCreated, item)
	default:
		w.Header().Set("Allow", http.MethodGet+", "+http.MethodPost)
		respondError(w, r, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func HandleChecklistOrder(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		w.Header().Set("Allow", http.MethodPut)
		respondError(w, r, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	idText := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/tasks/"), "/checklist/order")
	taskID, err := strconv.Atoi(idText)
	if err != nil || taskID < 1 {
		respondError(w, r, http.StatusBadRequest, "task id must be a positive integer")
		return
	}
	var payload struct {
		IDs []int `json:"ids"`
	}
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&payload); err != nil {
		respondError(w, r, http.StatusBadRequest, "request body must be valid JSON")
		return
	}
	items, err := service.ReorderChecklist(taskID, payload.IDs)
	if err != nil {
		if errors.Is(err, service.ErrChecklistOrderInvalid) {
			respondError(w, r, http.StatusConflict, err.Error())
		} else {
			respondChecklistError(w, r, err)
		}
		return
	}
	respondJSON(w, http.StatusOK, items)
}

func HandleChecklistProgress(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		respondError(w, r, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	idText := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/tasks/"), "/checklist/progress")
	taskID, err := strconv.Atoi(idText)
	if err != nil || taskID < 1 {
		respondError(w, r, http.StatusBadRequest, "task id must be a positive integer")
		return
	}
	progress, found := service.GetChecklistProgress(taskID)
	if !found {
		respondError(w, r, http.StatusNotFound, "task not found")
		return
	}
	respondJSON(w, http.StatusOK, progress)
}

func respondChecklistError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, service.ErrTaskNotFound), errors.Is(err, service.ErrChecklistItemNotFound):
		respondError(w, r, http.StatusNotFound, err.Error())
	case errors.Is(err, service.ErrChecklistLimitReached):
		respondError(w, r, http.StatusConflict, err.Error())
	default:
		respondError(w, r, http.StatusBadRequest, err.Error())
	}
}

func handleChecklistItem(w http.ResponseWriter, r *http.Request, taskID int, itemIDText string) {
	itemID, err := strconv.Atoi(itemIDText)
	if err != nil || itemID < 1 {
		respondError(w, r, http.StatusBadRequest, "checklist item id must be a positive integer")
		return
	}
	if r.Method != http.MethodPatch && r.Method != http.MethodDelete {
		w.Header().Set("Allow", http.MethodPatch+", "+http.MethodDelete)
		respondError(w, r, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	if r.Method == http.MethodPatch {
		var patch struct {
			Text      *string `json:"text"`
			Completed *bool   `json:"completed"`
		}
		decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&patch); err != nil {
			respondError(w, r, http.StatusBadRequest, "request body must be valid JSON")
			return
		}
		if patch.Text == nil && patch.Completed == nil {
			respondError(w, r, http.StatusBadRequest, "text or completed must be provided")
			return
		}
		item, err := service.UpdateChecklistItem(taskID, itemID, patch.Text, patch.Completed)
		if err != nil {
			respondChecklistError(w, r, err)
			return
		}
		respondJSON(w, http.StatusOK, item)
		return
	}
	if err := service.DeleteChecklistItem(taskID, itemID); err != nil {
		respondChecklistError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func HandleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		respondError(w, r, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func HandleActivity(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		respondError(w, r, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	offset, limit, err := pagination(r)
	if err != nil {
		respondError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	from, to, err := activityTimeRange(r)
	if err != nil {
		respondError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	action, err := activityActionFilter(r)
	if err != nil {
		respondError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, service.GetActivities(nil, action, from, to, offset, limit))
}

func HandleTaskActivity(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		respondError(w, r, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	idText := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/tasks/"), "/activity")
	taskID, err := strconv.Atoi(idText)
	if err != nil || taskID < 1 {
		respondError(w, r, http.StatusBadRequest, "task id must be a positive integer")
		return
	}
	offset, limit, err := pagination(r)
	if err != nil {
		respondError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	from, to, err := activityTimeRange(r)
	if err != nil {
		respondError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	action, err := activityActionFilter(r)
	if err != nil {
		respondError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, service.GetActivities(&taskID, action, from, to, offset, limit))
}

func HandleTaskOccurrences(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		respondError(w, r, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	taskID, err := taskIDFromSuffix(r.URL.Path, "/occurrences")
	if err != nil {
		respondError(w, r, http.StatusBadRequest, "task id must be a positive integer")
		return
	}
	occurrences, err := service.GetTaskOccurrences(taskID)
	if err != nil {
		if errors.Is(err, service.ErrTaskNotFound) || errors.Is(err, service.ErrRecurrenceNotFound) {
			respondError(w, r, http.StatusNotFound, err.Error())
		} else {
			respondError(w, r, http.StatusInternalServerError, "could not load occurrences")
		}
		return
	}
	respondJSON(w, http.StatusOK, occurrences)
}

func HandleNextRecurrence(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		respondError(w, r, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	taskID, err := taskIDFromSuffix(r.URL.Path, "/recurrence/next")
	if err != nil {
		respondError(w, r, http.StatusBadRequest, "task id must be a positive integer")
		return
	}
	dueDate, err := service.GetNextOccurrenceDueDate(taskID)
	if err != nil {
		if errors.Is(err, service.ErrTaskNotFound) || errors.Is(err, service.ErrRecurrenceNotFound) {
			respondError(w, r, http.StatusNotFound, err.Error())
		} else {
			respondError(w, r, http.StatusInternalServerError, "could not preview recurrence")
		}
		return
	}
	respondJSON(w, http.StatusOK, map[string]time.Time{"nextDueDate": dueDate})
}

func HandleActiveTimer(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		respondError(w, r, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	respondJSON(w, http.StatusOK, service.GetActiveTimer())
}

func HandleTimeReport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		respondError(w, r, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var taskID *int
	if value := r.URL.Query().Get("taskId"); value != "" {
		parsed, err := strconv.Atoi(value)
		if err != nil || parsed < 1 {
			respondError(w, r, http.StatusBadRequest, "taskId must be a positive integer")
			return
		}
		taskID = &parsed
	}
	from, to, err := activityTimeRange(r)
	if err != nil {
		respondError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, service.GetTimeReport(taskID, from, to))
}

func HandleTaskTime(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodGet+", "+http.MethodPost)
		respondError(w, r, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	if r.Method == http.MethodPost {
		var payload struct {
			StartedAt time.Time `json:"startedAt"`
			EndedAt   time.Time `json:"endedAt"`
			Note      string    `json:"note"`
		}
		decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&payload); err != nil {
			respondError(w, r, http.StatusBadRequest, "request body must be valid JSON")
			return
		}
		taskID, err := taskIDFromSuffix(r.URL.Path, "/time")
		if err != nil {
			respondError(w, r, http.StatusBadRequest, "task id must be a positive integer")
			return
		}
		entry, err := service.AddManualTimeEntry(taskID, payload.StartedAt, payload.EndedAt, payload.Note)
		if err != nil {
			respondTimeEntryError(w, r, err)
			return
		}
		respondJSON(w, http.StatusCreated, entry)
		return
	}
	taskID, err := taskIDFromSuffix(r.URL.Path, "/time")
	if err != nil {
		respondError(w, r, http.StatusBadRequest, "task id must be a positive integer")
		return
	}
	entries, found := service.GetTaskTimeEntries(taskID)
	if !found {
		respondError(w, r, http.StatusNotFound, "task not found")
		return
	}
	summary, _ := service.GetTaskTimeSummary(taskID)
	respondJSON(w, http.StatusOK, map[string]interface{}{"entries": entries, "summary": summary})
}

func HandleTaskTimeEntry(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		w.Header().Set("Allow", http.MethodDelete)
		respondError(w, r, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/tasks/"), "/time/")
	if len(parts) != 2 {
		respondError(w, r, http.StatusBadRequest, "time-entry route is invalid")
		return
	}
	taskID, taskErr := strconv.Atoi(parts[0])
	entryID, entryErr := strconv.Atoi(parts[1])
	if taskErr != nil || entryErr != nil || taskID < 1 || entryID < 1 {
		respondError(w, r, http.StatusBadRequest, "task and entry ids must be positive integers")
		return
	}
	if err := service.DeleteTimeEntry(taskID, entryID); err != nil {
		respondTimeEntryError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func respondTimeEntryError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, service.ErrTaskNotFound), errors.Is(err, service.ErrTimeEntryNotFound):
		respondError(w, r, http.StatusNotFound, err.Error())
	case errors.Is(err, service.ErrTimeEntryOverlap), errors.Is(err, service.ErrTimerAlreadyRunning):
		respondError(w, r, http.StatusConflict, err.Error())
	default:
		respondError(w, r, http.StatusBadRequest, err.Error())
	}
}

func HandleTaskTimer(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		respondError(w, r, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	if strings.HasSuffix(r.URL.Path, "/timer/start") {
		taskID, err := taskIDFromSuffix(r.URL.Path, "/timer/start")
		if err != nil {
			respondError(w, r, http.StatusBadRequest, "task id must be a positive integer")
			return
		}
		var payload struct {
			Note string `json:"note"`
		}
		if r.Body != nil {
			decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
			decoder.DisallowUnknownFields()
			if err := decoder.Decode(&payload); err != nil && err != io.EOF {
				respondError(w, r, http.StatusBadRequest, "request body must be valid JSON")
				return
			}
		}
		entry, err := service.StartTaskTimer(taskID, payload.Note)
		if err != nil {
			switch {
			case errors.Is(err, service.ErrTaskNotFound):
				respondError(w, r, http.StatusNotFound, err.Error())
			case errors.Is(err, service.ErrTimerAlreadyRunning):
				respondError(w, r, http.StatusConflict, err.Error())
			default:
				respondError(w, r, http.StatusBadRequest, err.Error())
			}
			return
		}
		respondJSON(w, http.StatusCreated, entry)
		return
	}
	taskID, err := taskIDFromSuffix(r.URL.Path, "/timer/stop")
	if err != nil {
		respondError(w, r, http.StatusBadRequest, "task id must be a positive integer")
		return
	}
	entry, err := service.StopTaskTimer(taskID)
	if err != nil {
		if errors.Is(err, service.ErrTaskNotFound) {
			respondError(w, r, http.StatusNotFound, err.Error())
		} else {
			respondError(w, r, http.StatusConflict, err.Error())
		}
		return
	}
	respondJSON(w, http.StatusOK, entry)
}

func taskIDFromSuffix(path, suffix string) (int, error) {
	id, err := strconv.Atoi(strings.TrimSuffix(strings.TrimPrefix(path, "/tasks/"), suffix))
	if err != nil || id < 1 {
		return 0, fmt.Errorf("invalid task id")
	}
	return id, nil
}

func activityTimeRange(r *http.Request) (*time.Time, *time.Time, error) {
	from, err := dateFilter(r, "from")
	if err != nil {
		return nil, nil, fmt.Errorf("from must be an RFC3339 timestamp")
	}
	to, err := dateFilter(r, "to")
	if err != nil {
		return nil, nil, fmt.Errorf("to must be an RFC3339 timestamp")
	}
	if from != nil && to != nil && from.After(*to) {
		return nil, nil, fmt.Errorf("from must not be later than to")
	}
	return from, to, nil
}

func activityActionFilter(r *http.Request) (string, error) {
	action := r.URL.Query().Get("action")
	if action != "" && !service.IsActivityAction(action) {
		return "", fmt.Errorf("action is not a supported activity type")
	}
	return action, nil
}

func HandleBulkComplete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		respondError(w, r, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var payload struct {
		IDs       []int `json:"ids"`
		Completed *bool `json:"completed"`
	}
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&payload); err != nil {
		respondError(w, r, http.StatusBadRequest, "request body must be valid JSON")
		return
	}
	if len(payload.IDs) == 0 {
		respondError(w, r, http.StatusBadRequest, "ids must contain at least one task id")
		return
	}
	for _, id := range payload.IDs {
		if id < 1 {
			respondError(w, r, http.StatusBadRequest, "task ids must be positive integers")
			return
		}
	}

	completed := true
	if payload.Completed != nil {
		completed = *payload.Completed
	}
	result, err := service.BulkSetTaskCompletionChecked(payload.IDs, completed)
	if err != nil {
		if errors.Is(err, service.ErrTaskBlocked) {
			respondError(w, r, http.StatusConflict, err.Error())
		} else {
			respondError(w, r, http.StatusNotFound, "one or more tasks not found")
		}
		return
	}
	respondJSON(w, http.StatusOK, result)
}

func HandleBulkDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		respondError(w, r, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var payload struct {
		IDs []int `json:"ids"`
	}
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&payload); err != nil {
		respondError(w, r, http.StatusBadRequest, "request body must be valid JSON")
		return
	}
	if len(payload.IDs) == 0 {
		respondError(w, r, http.StatusBadRequest, "ids must contain at least one task id")
		return
	}
	for _, id := range payload.IDs {
		if id < 1 {
			respondError(w, r, http.StatusBadRequest, "task ids must be positive integers")
			return
		}
	}
	result, ok := service.BulkDeleteTasks(payload.IDs)
	if !ok {
		respondError(w, r, http.StatusNotFound, "one or more tasks not found")
		return
	}
	respondJSON(w, http.StatusOK, result)
}

func HandleBulkPriority(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		respondError(w, r, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var payload struct {
		IDs      []int  `json:"ids"`
		Priority string `json:"priority"`
	}
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&payload); err != nil {
		respondError(w, r, http.StatusBadRequest, "request body must be valid JSON")
		return
	}
	if len(payload.IDs) == 0 {
		respondError(w, r, http.StatusBadRequest, "ids must contain at least one task id")
		return
	}
	priority := strings.ToLower(strings.TrimSpace(payload.Priority))
	if priority != "low" && priority != "medium" && priority != "high" {
		respondError(w, r, http.StatusBadRequest, "priority must be low, medium, or high")
		return
	}
	for _, id := range payload.IDs {
		if id < 1 {
			respondError(w, r, http.StatusBadRequest, "task ids must be positive integers")
			return
		}
	}
	result, ok := service.BulkSetTaskPriority(payload.IDs, priority)
	if !ok {
		respondError(w, r, http.StatusNotFound, "one or more tasks not found")
		return
	}
	respondJSON(w, http.StatusOK, result)
}

func updateTask(w http.ResponseWriter, r *http.Request, id int) {
	task, ok := decodeTask(w, r)
	if !ok {
		return
	}

	updated, err := service.UpdateTaskWithDependencies(id, task, task.DependsOn)
	if err != nil {
		respondDependencyError(w, r, err)
		return
	}
	respondJSON(w, http.StatusOK, updated)
}

func patchTask(w http.ResponseWriter, r *http.Request, id int) {
	var patch struct {
		Title           *string    `json:"title"`
		Description     *string    `json:"description"`
		DueDate         *time.Time `json:"dueDate"`
		Completed       *bool      `json:"completed"`
		Priority        *string    `json:"priority"`
		Tags            *[]string  `json:"tags"`
		DependsOn       *[]int     `json:"dependsOn"`
		EstimateMinutes *int       `json:"estimateMinutes"`
	}
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&patch); err != nil {
		respondError(w, r, http.StatusBadRequest, "request body must be valid JSON")
		return
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		respondError(w, r, http.StatusBadRequest, "request body must contain a single JSON value")
		return
	}
	if patch.Title == nil && patch.Description == nil && patch.DueDate == nil && patch.Completed == nil && patch.Priority == nil && patch.Tags == nil && patch.DependsOn == nil && patch.EstimateMinutes == nil {
		respondError(w, r, http.StatusBadRequest, "at least one task field must be provided")
		return
	}
	task := service.GetTaskByID(id)
	if task == nil {
		respondError(w, r, http.StatusNotFound, "task not found")
		return
	}
	if patch.Title != nil {
		task.Title = *patch.Title
	}
	if patch.Description != nil {
		task.Description = *patch.Description
	}
	if patch.DueDate != nil {
		task.DueDate = *patch.DueDate
	}
	if patch.Completed != nil {
		task.Completed = *patch.Completed
	}
	if patch.Priority != nil {
		task.Priority = *patch.Priority
	}
	if patch.Tags != nil {
		task.Tags = *patch.Tags
	}
	if patch.DependsOn != nil {
		task.DependsOn = *patch.DependsOn
	}
	if patch.EstimateMinutes != nil {
		task.EstimateMinutes = *patch.EstimateMinutes
	}
	if validationErrors := validation.ValidateTask(task); len(validationErrors) > 0 {
		respondJSON(w, http.StatusBadRequest, map[string]interface{}{"errors": validationErrors})
		return
	}
	updated, err := service.UpdateTaskWithDependencies(id, *task, task.DependsOn)
	if err != nil {
		respondDependencyError(w, r, err)
		return
	}
	respondJSON(w, http.StatusOK, updated)
}

func decodeTask(w http.ResponseWriter, r *http.Request) (model.Task, bool) {
	var task model.Task
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&task); err != nil {
		respondError(w, r, http.StatusBadRequest, "request body must be valid JSON")
		return model.Task{}, false
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		respondError(w, r, http.StatusBadRequest, "request body must contain a single JSON value")
		return model.Task{}, false
	}

	if errors := validation.ValidateTask(&task); len(errors) > 0 {
		respondJSON(w, http.StatusBadRequest, map[string]interface{}{"errors": errors})
		return model.Task{}, false
	}
	return task, true
}

func respondDependencyError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, service.ErrTaskNotFound):
		respondError(w, r, http.StatusNotFound, err.Error())
	case errors.Is(err, service.ErrDependencySelfReference), errors.Is(err, service.ErrDependencyAlreadyExists), errors.Is(err, service.ErrDependencyCycle):
		respondError(w, r, http.StatusConflict, err.Error())
	default:
		respondError(w, r, http.StatusInternalServerError, "could not update dependencies")
	}
}

func respondJSON(w http.ResponseWriter, status int, value interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func respondError(w http.ResponseWriter, r *http.Request, status int, message string) {
	appError.RespondWithError(w, r, appError.NewAppError(status, message, nil))
}
