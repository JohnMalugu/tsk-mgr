package routes

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/JohnMalugu/tsk-mgr-api/internal/handler"
)

// Router directs HTTP requests to the correct handler
func Router(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("[%s] %s\n", r.Method, r.URL.Path)
	if r.URL.Path == "/health" {
		handler.HandleHealth(w, r)
		return
	}
	if r.URL.Path == "/auth/register" && r.Method == http.MethodPost {
		handler.HandleRegister(w, r)
		return
	}
	if r.URL.Path == "/auth/login" && r.Method == http.MethodPost {
		handler.HandleLogin(w, r)
		return
	}
	if r.URL.Path == "/activity" {
		handler.HandleActivity(w, r)
		return
	}
	if r.URL.Path == "/tags" {
		handler.HandleTags(w, r)
		return
	}
	if r.URL.Path == "/timer" {
		handler.HandleActiveTimer(w, r)
		return
	}
	if r.URL.Path == "/time/report" {
		handler.HandleTimeReport(w, r)
		return
	}
	if r.URL.Path == "/tasks/bulk/duplicate" {
		handler.HandleBulkDuplicate(w, r)
		return
	}
	if strings.HasPrefix(r.URL.Path, "/tasks/") && (strings.HasSuffix(r.URL.Path, "/timer/start") || strings.HasSuffix(r.URL.Path, "/timer/stop")) {
		handler.HandleTaskTimer(w, r)
		return
	}
	if strings.HasPrefix(r.URL.Path, "/tasks/") && strings.HasSuffix(r.URL.Path, "/time") {
		handler.HandleTaskTime(w, r)
		return
	}
	if strings.HasPrefix(r.URL.Path, "/tasks/") && strings.Contains(r.URL.Path, "/time/") {
		handler.HandleTaskTimeEntry(w, r)
		return
	}
	if strings.HasPrefix(r.URL.Path, "/tasks/") && strings.HasSuffix(r.URL.Path, "/activity") {
		handler.HandleTaskActivity(w, r)
		return
	}
	if strings.HasPrefix(r.URL.Path, "/tasks/") && strings.HasSuffix(r.URL.Path, "/comments") {
		handler.HandleTaskComments(w, r)
		return
	}
	if strings.HasPrefix(r.URL.Path, "/tasks/") && strings.Contains(r.URL.Path, "/comments/") {
		handler.HandleTaskCommentItem(w, r)
		return
	}
	if strings.HasPrefix(r.URL.Path, "/tasks/") && strings.HasSuffix(r.URL.Path, "/occurrences") {
		handler.HandleTaskOccurrences(w, r)
		return
	}
	if strings.HasPrefix(r.URL.Path, "/tasks/") && strings.HasSuffix(r.URL.Path, "/recurrence/next") {
		handler.HandleNextRecurrence(w, r)
		return
	}
	if strings.HasPrefix(r.URL.Path, "/tasks/") && strings.HasSuffix(r.URL.Path, "/recurrence/skip") {
		handler.HandleSkipRecurrence(w, r)
		return
	}
	if strings.HasPrefix(r.URL.Path, "/tasks/") && strings.HasSuffix(r.URL.Path, "/duplicate") {
		handler.HandleDuplicateTask(w, r)
		return
	}

	if r.URL.Path == "/tasks" {
		handler.HandleTasks(w, r)
		return
	}

	if r.URL.Path == "/v2/tasks" && r.Method == http.MethodPost {
		handler.GlobalTaskDBHandler.HandleCreate(w, r)
		return
	}

	if r.URL.Path == "/v2/tasks" && r.Method == http.MethodGet {
		handler.GlobalTaskDBHandler.HandleList(w, r)
		return
	}

	if r.URL.Path == "/v2/tasks/blocked" && r.Method == http.MethodGet {
		handler.GlobalTaskDependencyHandler.HandleGetBlockedTasks(w, r)
		return
	}

	if r.URL.Path == "/v2/tasks/ready" && r.Method == http.MethodGet {
		handler.GlobalTaskDependencyHandler.HandleGetReadyTasks(w, r)
		return
	}

	if strings.HasPrefix(r.URL.Path, "/v2/tasks/") {
		trimmed := strings.TrimPrefix(r.URL.Path, "/v2/tasks/")
		if strings.HasSuffix(trimmed, "/complete") && r.Method == http.MethodPatch {
			idStr := strings.TrimSuffix(trimmed, "/complete")
			if id, err := strconv.Atoi(idStr); err == nil {
				handler.GlobalTaskDBHandler.HandleComplete(w, r, id)
				return
			}
		}
		if id, err := strconv.Atoi(trimmed); err == nil {
			switch r.Method {
			case http.MethodGet:
				handler.GlobalTaskDBHandler.HandleGetByID(w, r, id)
				return
			case http.MethodPut:
				handler.GlobalTaskDBHandler.HandleUpdate(w, r, id)
				return
			case http.MethodDelete:
				handler.GlobalTaskDBHandler.HandleDelete(w, r, id)
				return
			}
		}
	}

	if strings.HasPrefix(r.URL.Path, "/v2/tasks/") && strings.Contains(r.URL.Path, "/subtasks") {
		trimmed := strings.TrimPrefix(r.URL.Path, "/v2/tasks/")
		parts := strings.Split(trimmed, "/")
		if len(parts) >= 2 && parts[1] == "subtasks" {
			if taskID, err := strconv.Atoi(parts[0]); err == nil {
				if len(parts) == 2 {
					if r.Method == http.MethodGet {
						handler.GlobalSubtaskHandler.HandleList(w, r, taskID)
						return
					}
					if r.Method == http.MethodPost {
						handler.GlobalSubtaskHandler.HandleCreate(w, r, taskID)
						return
					}
				} else if len(parts) == 3 && parts[2] == "reorder" && r.Method == http.MethodPut {
					handler.GlobalSubtaskHandler.HandleReorder(w, r, taskID)
					return
				}
			}
		}
	}

	if strings.HasPrefix(r.URL.Path, "/v2/subtasks/") {
		trimmed := strings.TrimPrefix(r.URL.Path, "/v2/subtasks/")
		if strings.HasSuffix(trimmed, "/toggle") && r.Method == http.MethodPatch {
			idStr := strings.TrimSuffix(trimmed, "/toggle")
			if id, err := strconv.Atoi(idStr); err == nil {
				handler.GlobalSubtaskHandler.HandleToggle(w, r, id)
				return
			}
		}
		if id, err := strconv.Atoi(trimmed); err == nil {
			switch r.Method {
			case http.MethodPut:
				handler.GlobalSubtaskHandler.HandleUpdate(w, r, id)
				return
			case http.MethodDelete:
				handler.GlobalSubtaskHandler.HandleDelete(w, r, id)
				return
			}
		}
	}

	if strings.HasPrefix(r.URL.Path, "/v2/tasks/") && strings.Contains(r.URL.Path, "/comments") {
		trimmed := strings.TrimPrefix(r.URL.Path, "/v2/tasks/")
		parts := strings.Split(trimmed, "/")
		if len(parts) == 2 && parts[1] == "comments" {
			if taskID, err := strconv.Atoi(parts[0]); err == nil {
				if r.Method == http.MethodGet {
					handler.GlobalCommentHandler.HandleList(w, r, taskID)
					return
				}
				if r.Method == http.MethodPost {
					handler.GlobalCommentHandler.HandleCreate(w, r, taskID)
					return
				}
			}
		}
	}

	if strings.HasPrefix(r.URL.Path, "/v2/comments/") {
		trimmed := strings.TrimPrefix(r.URL.Path, "/v2/comments/")
		if id, err := strconv.Atoi(trimmed); err == nil {
			switch r.Method {
			case http.MethodPut:
				handler.GlobalCommentHandler.HandleUpdate(w, r, id)
				return
			case http.MethodDelete:
				handler.GlobalCommentHandler.HandleDelete(w, r, id)
				return
			}
		}
	}

	if r.URL.Path == "/v2/timer" && r.Method == http.MethodGet {
		handler.GlobalTimeHandler.HandleGetActiveTimer(w, r)
		return
	}

	if r.URL.Path == "/v2/timer/stop" && r.Method == http.MethodPost {
		handler.GlobalTimeHandler.HandleStopTimer(w, r)
		return
	}

	if r.URL.Path == "/v2/time/report" && r.Method == http.MethodGet {
		handler.GlobalTimeHandler.HandleGetTimeReport(w, r)
		return
	}

	if strings.HasPrefix(r.URL.Path, "/v2/time/") {
		trimmed := strings.TrimPrefix(r.URL.Path, "/v2/time/")
		if id, err := strconv.Atoi(trimmed); err == nil && r.Method == http.MethodDelete {
			handler.GlobalTimeHandler.HandleDeleteTimeEntry(w, r, id)
			return
		}
	}

	if strings.HasPrefix(r.URL.Path, "/v2/tasks/") && strings.Contains(r.URL.Path, "/timer/start") && r.Method == http.MethodPost {
		trimmed := strings.TrimPrefix(r.URL.Path, "/v2/tasks/")
		parts := strings.Split(trimmed, "/")
		if len(parts) >= 2 {
			if taskID, err := strconv.Atoi(parts[0]); err == nil {
				handler.GlobalTimeHandler.HandleStartTimer(w, r, taskID)
				return
			}
		}
	}

	if strings.HasPrefix(r.URL.Path, "/v2/tasks/") && strings.Contains(r.URL.Path, "/time/summary") && r.Method == http.MethodGet {
		trimmed := strings.TrimPrefix(r.URL.Path, "/v2/tasks/")
		parts := strings.Split(trimmed, "/")
		if len(parts) >= 2 {
			if taskID, err := strconv.Atoi(parts[0]); err == nil {
				handler.GlobalTimeHandler.HandleGetTaskTimeSummary(w, r, taskID)
				return
			}
		}
	}

	if strings.HasPrefix(r.URL.Path, "/v2/tasks/") && strings.Contains(r.URL.Path, "/time") {
		trimmed := strings.TrimPrefix(r.URL.Path, "/v2/tasks/")
		parts := strings.Split(trimmed, "/")
		if len(parts) == 2 && parts[1] == "time" {
			if taskID, err := strconv.Atoi(parts[0]); err == nil {
				if r.Method == http.MethodGet {
					handler.GlobalTimeHandler.HandleGetTaskTimeEntries(w, r, taskID)
					return
				}
				if r.Method == http.MethodPost {
					handler.GlobalTimeHandler.HandleLogManualTime(w, r, taskID)
					return
				}
			}
		}
	}

	if strings.HasPrefix(r.URL.Path, "/v2/tasks/") && strings.Contains(r.URL.Path, "/dependencies") {
		trimmed := strings.TrimPrefix(r.URL.Path, "/v2/tasks/")
		parts := strings.Split(trimmed, "/")
		if len(parts) >= 2 && parts[1] == "dependencies" {
			if taskID, err := strconv.Atoi(parts[0]); err == nil {
				if len(parts) == 2 {
					if r.Method == http.MethodGet {
						handler.GlobalTaskDependencyHandler.HandleGetDependencies(w, r, taskID)
						return
					}
					if r.Method == http.MethodPost {
						handler.GlobalTaskDependencyHandler.HandleAddDependency(w, r, taskID)
						return
					}
				} else if len(parts) == 3 && r.Method == http.MethodDelete {
					if depID, err := strconv.Atoi(parts[2]); err == nil {
						handler.GlobalTaskDependencyHandler.HandleRemoveDependency(w, r, taskID, depID)
						return
					}
				}
			}
		}
	}

	if strings.HasPrefix(r.URL.Path, "/v2/tasks/") && strings.Contains(r.URL.Path, "/dependents") && r.Method == http.MethodGet {
		trimmed := strings.TrimPrefix(r.URL.Path, "/v2/tasks/")
		parts := strings.Split(trimmed, "/")
		if len(parts) == 2 && parts[1] == "dependents" {
			if taskID, err := strconv.Atoi(parts[0]); err == nil {
				handler.GlobalTaskDependencyHandler.HandleGetDependents(w, r, taskID)
				return
			}
		}
	}

	if strings.HasPrefix(r.URL.Path, "/v2/tasks/") && strings.Contains(r.URL.Path, "/dag") && r.Method == http.MethodGet {
		trimmed := strings.TrimPrefix(r.URL.Path, "/v2/tasks/")
		parts := strings.Split(trimmed, "/")
		if len(parts) == 2 && parts[1] == "dag" {
			if taskID, err := strconv.Atoi(parts[0]); err == nil {
				handler.GlobalTaskDependencyHandler.HandleGetDAGSummary(w, r, taskID)
				return
			}
		}
	}

	if r.URL.Path == "/v2/activity/summary" && r.Method == http.MethodGet {
		handler.GlobalActivityHandler.HandleGetActivitySummary(w, r)
		return
	}

	if r.URL.Path == "/v2/activity/purge" && r.Method == http.MethodPost {
		handler.GlobalActivityHandler.HandlePurgeActivities(w, r)
		return
	}

	if r.URL.Path == "/v2/activity" && r.Method == http.MethodGet {
		handler.GlobalActivityHandler.HandleGetUserActivity(w, r)
		return
	}

	if strings.HasPrefix(r.URL.Path, "/v2/tasks/") && strings.Contains(r.URL.Path, "/activity") && r.Method == http.MethodGet {
		trimmed := strings.TrimPrefix(r.URL.Path, "/v2/tasks/")
		parts := strings.Split(trimmed, "/")
		if len(parts) == 2 && parts[1] == "activity" {
			if taskID, err := strconv.Atoi(parts[0]); err == nil {
				handler.GlobalActivityHandler.HandleGetTaskActivity(w, r, taskID)
				return
			}
		}
	}

	if r.URL.Path == "/projects" && r.Method == http.MethodPost {
		handler.GlobalProjectHandler.HandleCreate(w, r)
		return
	}

	if r.URL.Path == "/projects" && r.Method == http.MethodGet {
		handler.GlobalProjectHandler.HandleList(w, r)
		return
	}

	if strings.HasPrefix(r.URL.Path, "/projects/") {
		trimmed := strings.TrimPrefix(r.URL.Path, "/projects/")
		if id, err := strconv.Atoi(trimmed); err == nil {
			switch r.Method {
			case http.MethodGet:
				handler.GlobalProjectHandler.HandleGetByID(w, r, id)
				return
			case http.MethodPut:
				handler.GlobalProjectHandler.HandleUpdate(w, r, id)
				return
			case http.MethodDelete:
				handler.GlobalProjectHandler.HandleDelete(w, r, id)
				return
			}
		}
	}

	if r.URL.Path == "/v2/tags" && r.Method == http.MethodPost {
		handler.GlobalTagHandler.HandleCreate(w, r)
		return
	}

	if r.URL.Path == "/v2/tags" && r.Method == http.MethodGet {
		handler.GlobalTagHandler.HandleList(w, r)
		return
	}

	if strings.HasPrefix(r.URL.Path, "/v2/tags/") {
		trimmed := strings.TrimPrefix(r.URL.Path, "/v2/tags/")
		if id, err := strconv.Atoi(trimmed); err == nil {
			switch r.Method {
			case http.MethodGet:
				handler.GlobalTagHandler.HandleGetByID(w, r, id)
				return
			case http.MethodPut:
				handler.GlobalTagHandler.HandleUpdate(w, r, id)
				return
			case http.MethodDelete:
				handler.GlobalTagHandler.HandleDelete(w, r, id)
				return
			}
		}
	}

	if strings.HasPrefix(r.URL.Path, "/v2/tasks/") && strings.Contains(r.URL.Path, "/tags") {
		trimmed := strings.TrimPrefix(r.URL.Path, "/v2/tasks/")
		parts := strings.Split(trimmed, "/")
		if len(parts) >= 2 && parts[1] == "tags" {
			if taskID, err := strconv.Atoi(parts[0]); err == nil {
				if len(parts) == 2 {
					if r.Method == http.MethodGet {
						handler.GlobalTagHandler.HandleGetTaskTags(w, r, taskID)
						return
					}
					if r.Method == http.MethodPost {
						handler.GlobalTagHandler.HandleAttachTag(w, r, taskID)
						return
					}
				} else if len(parts) == 3 && r.Method == http.MethodDelete {
					if tagID, err := strconv.Atoi(parts[2]); err == nil {
						handler.GlobalTagHandler.HandleDetachTag(w, r, taskID, tagID)
						return
					}
				}
			}
		}
	}

	if r.URL.Path == "/tasks/summary" {
		handler.HandleTaskSummary(w, r)
		return
	}
	if r.URL.Path == "/tasks/ready" {
		handler.HandleReadyTasks(w, r)
		return
	}
	if r.URL.Path == "/tasks/blocked" {
		handler.HandleBlockedTasks(w, r)
		return
	}
	if r.URL.Path == "/tasks/upcoming" {
		handler.HandleUpcomingTasks(w, r)
		return
	}
	if strings.HasPrefix(r.URL.Path, "/tasks/") && strings.HasSuffix(r.URL.Path, "/checklist/order") {
		handler.HandleChecklistOrder(w, r)
		return
	}
	if strings.HasPrefix(r.URL.Path, "/tasks/") && strings.HasSuffix(r.URL.Path, "/checklist/complete") {
		handler.HandleBulkChecklistCompletion(w, r)
		return
	}
	if strings.HasPrefix(r.URL.Path, "/tasks/") && strings.HasSuffix(r.URL.Path, "/checklist/progress") {
		handler.HandleChecklistProgress(w, r)
		return
	}
	if strings.HasPrefix(r.URL.Path, "/tasks/") && strings.Contains(r.URL.Path, "/checklist") {
		handler.HandleTaskChecklist(w, r)
		return
	}
	if strings.HasPrefix(r.URL.Path, "/tasks/") && strings.HasSuffix(r.URL.Path, "/dependencies") {
		handler.HandleTaskDependencies(w, r)
		return
	}
	if strings.HasPrefix(r.URL.Path, "/tasks/") && strings.HasSuffix(r.URL.Path, "/dependents") {
		handler.HandleTaskDependents(w, r)
		return
	}
	if strings.HasPrefix(r.URL.Path, "/tasks/") && strings.Contains(r.URL.Path, "/dependencies/") {
		handler.HandleTaskDependency(w, r)
		return
	}

	if r.URL.Path == "/tasks/bulk/complete" {
		handler.HandleBulkComplete(w, r)
		return
	}
	if r.URL.Path == "/tasks/bulk/delete" {
		handler.HandleBulkDelete(w, r)
		return
	}
	if r.URL.Path == "/tasks/bulk/priority" {
		handler.HandleBulkPriority(w, r)
		return
	}
	if r.URL.Path == "/tasks/bulk/due-date" {
		handler.HandleBulkDueDate(w, r)
		return
	}
	if r.URL.Path == "/tasks/bulk/tags" {
		handler.HandleBulkTags(w, r)
		return
	}
	if r.URL.Path == "/tasks/bulk/status" {
		handler.HandleBulkStatus(w, r)
		return
	}
	if r.URL.Path == "/tasks/bulk/estimate" {
		handler.HandleBulkEstimate(w, r)
		return
	}

	if strings.HasSuffix(r.URL.Path, "/complete") && strings.HasPrefix(r.URL.Path, "/tasks/") {
		handler.HandleTaskComplete(w, r)
		return
	}
	if strings.HasSuffix(r.URL.Path, "/reopen") && strings.HasPrefix(r.URL.Path, "/tasks/") {
		handler.HandleReopenCanceledTask(w, r)
		return
	}

	if strings.HasPrefix(r.URL.Path, "/tasks/") {
		handler.HandleTaskByID(w, r)
		return
	}

	// Route not found
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotFound)
	json.NewEncoder(w).Encode(map[string]string{"error": "Route not found"})
}
