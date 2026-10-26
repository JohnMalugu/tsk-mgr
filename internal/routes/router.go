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

	// V2 Workspaces & Invites
	if r.URL.Path == "/v2/workspaces" {
		if handler.GlobalWorkspaceDBHandler == nil {
			http.Error(w, "Workspace handler not initialized", http.StatusInternalServerError)
			return
		}
		if r.Method == http.MethodPost {
			handler.GlobalWorkspaceDBHandler.HandleCreate(w, r)
			return
		}
		if r.Method == http.MethodGet {
			handler.GlobalWorkspaceDBHandler.HandleList(w, r)
			return
		}
	}

	if r.URL.Path == "/v2/invites/accept" && r.Method == http.MethodPost {
		if handler.GlobalWorkspaceDBHandler != nil {
			handler.GlobalWorkspaceDBHandler.HandleAcceptInvite(w, r)
			return
		}
	}

	if strings.HasPrefix(r.URL.Path, "/v2/workspaces/") {
		if handler.GlobalWorkspaceDBHandler == nil {
			http.Error(w, "Workspace handler not initialized", http.StatusInternalServerError)
			return
		}
		trimmed := strings.TrimPrefix(r.URL.Path, "/v2/workspaces/")
		parts := strings.Split(trimmed, "/")
		wsIDStr := parts[0]

		if len(parts) == 1 {
			switch r.Method {
			case http.MethodGet:
				handler.GlobalWorkspaceDBHandler.HandleGet(w, r, wsIDStr)
				return
			case http.MethodPut, http.MethodPatch:
				handler.GlobalWorkspaceDBHandler.HandleUpdate(w, r, wsIDStr)
				return
			case http.MethodDelete:
				handler.GlobalWorkspaceDBHandler.HandleDelete(w, r, wsIDStr)
				return
			}
		} else if len(parts) == 2 {
			if parts[1] == "members" {
				if r.Method == http.MethodGet {
					handler.GlobalWorkspaceDBHandler.HandleListMembers(w, r, wsIDStr)
					return
				}
				if r.Method == http.MethodPost {
					handler.GlobalWorkspaceDBHandler.HandleAddMember(w, r, wsIDStr)
					return
				}
			} else if parts[1] == "invites" && r.Method == http.MethodPost {
				handler.GlobalWorkspaceDBHandler.HandleInviteMember(w, r, wsIDStr)
				return
			}
		} else if len(parts) == 3 && parts[1] == "members" {
			targetUserID := parts[2]
			if r.Method == http.MethodPut || r.Method == http.MethodPatch {
				handler.GlobalWorkspaceDBHandler.HandleUpdateMemberRole(w, r, wsIDStr, targetUserID)
				return
			}
			if r.Method == http.MethodDelete {
				handler.GlobalWorkspaceDBHandler.HandleRemoveMember(w, r, wsIDStr, targetUserID)
				return
			}
		}
	}

	// V2 Notifications & Preferences
	if r.URL.Path == "/v2/notifications" {
		if handler.GlobalNotificationHandler == nil {
			http.Error(w, "Notification handler not initialized", http.StatusInternalServerError)
			return
		}
		if r.Method == http.MethodGet {
			handler.GlobalNotificationHandler.HandleList(w, r)
			return
		}
		if r.Method == http.MethodPost {
			handler.GlobalNotificationHandler.HandleSendNotification(w, r)
			return
		}
	}

	if r.URL.Path == "/v2/notifications/unread-count" && r.Method == http.MethodGet {
		if handler.GlobalNotificationHandler != nil {
			handler.GlobalNotificationHandler.HandleGetUnreadCount(w, r)
			return
		}
	}

	if r.URL.Path == "/v2/notifications/read-all" && r.Method == http.MethodPost {
		if handler.GlobalNotificationHandler != nil {
			handler.GlobalNotificationHandler.HandleMarkAllAsRead(w, r)
			return
		}
	}

	if r.URL.Path == "/v2/notifications/preferences" {
		if handler.GlobalNotificationHandler == nil {
			http.Error(w, "Notification handler not initialized", http.StatusInternalServerError)
			return
		}
		if r.Method == http.MethodGet {
			handler.GlobalNotificationHandler.HandleGetPreferences(w, r)
			return
		}
		if r.Method == http.MethodPut || r.Method == http.MethodPatch {
			handler.GlobalNotificationHandler.HandleUpdatePreferences(w, r)
			return
		}
	}

	if strings.HasPrefix(r.URL.Path, "/v2/notifications/") {
		if handler.GlobalNotificationHandler == nil {
			http.Error(w, "Notification handler not initialized", http.StatusInternalServerError)
			return
		}
		trimmed := strings.TrimPrefix(r.URL.Path, "/v2/notifications/")
		parts := strings.Split(trimmed, "/")
		idStr := parts[0]

		if len(parts) == 1 {
			if r.Method == http.MethodDelete {
				handler.GlobalNotificationHandler.HandleDelete(w, r, idStr)
				return
			}
		} else if len(parts) == 2 && parts[1] == "read" {
			if r.Method == http.MethodPost || r.Method == http.MethodPatch {
				handler.GlobalNotificationHandler.HandleMarkAsRead(w, r, idStr)
				return
			}
		}
	}

	// V2 Webhooks & Deliveries
	if r.URL.Path == "/v2/webhooks" {
		if handler.GlobalWebhookHandler == nil {
			http.Error(w, "Webhook handler not initialized", http.StatusInternalServerError)
			return
		}
		if r.Method == http.MethodGet {
			handler.GlobalWebhookHandler.HandleListSubscriptions(w, r)
			return
		}
		if r.Method == http.MethodPost {
			handler.GlobalWebhookHandler.HandleCreateSubscription(w, r)
			return
		}
	}

	if strings.HasPrefix(r.URL.Path, "/v2/webhooks/deliveries/") && strings.HasSuffix(r.URL.Path, "/retry") && r.Method == http.MethodPost {
		if handler.GlobalWebhookHandler == nil {
			http.Error(w, "Webhook handler not initialized", http.StatusInternalServerError)
			return
		}
		trimmed := strings.TrimPrefix(r.URL.Path, "/v2/webhooks/deliveries/")
		deliveryIDStr := strings.TrimSuffix(trimmed, "/retry")
		if deliveryID, err := strconv.ParseInt(deliveryIDStr, 10, 64); err == nil {
			handler.GlobalWebhookHandler.HandleRetryDelivery(w, r, deliveryID)
			return
		}
		http.Error(w, "Invalid delivery ID", http.StatusBadRequest)
		return
	}

	if strings.HasPrefix(r.URL.Path, "/v2/webhooks/") {
		if handler.GlobalWebhookHandler == nil {
			http.Error(w, "Webhook handler not initialized", http.StatusInternalServerError)
			return
		}
		trimmed := strings.TrimPrefix(r.URL.Path, "/v2/webhooks/")
		parts := strings.Split(trimmed, "/")
		id, err := strconv.ParseInt(parts[0], 10, 64)
		if err != nil {
			http.Error(w, "Invalid webhook ID", http.StatusBadRequest)
			return
		}

		if len(parts) == 1 {
			if r.Method == http.MethodGet {
				handler.GlobalWebhookHandler.HandleGetSubscription(w, r, id)
				return
			}
			if r.Method == http.MethodPut || r.Method == http.MethodPatch {
				handler.GlobalWebhookHandler.HandleUpdateSubscription(w, r, id)
				return
			}
			if r.Method == http.MethodDelete {
				handler.GlobalWebhookHandler.HandleDeleteSubscription(w, r, id)
				return
			}
		} else if len(parts) == 2 {
			if parts[1] == "ping" && r.Method == http.MethodPost {
				handler.GlobalWebhookHandler.HandlePingSubscription(w, r, id)
				return
			}
			if parts[1] == "deliveries" && r.Method == http.MethodGet {
				handler.GlobalWebhookHandler.HandleListDeliveries(w, r, id)
				return
			}
			if parts[1] == "stats" && r.Method == http.MethodGet {
				handler.GlobalWebhookHandler.HandleGetStats(w, r, id)
				return
			}
		}
	}

	// Sprints API (/v2/sprints)
	if r.URL.Path == "/v2/sprints" {
		if r.Method == http.MethodPost {
			handler.GlobalSprintHandler.HandleCreate(w, r)
			return
		}
		if r.Method == http.MethodGet {
			handler.GlobalSprintHandler.HandleList(w, r)
			return
		}
	}

	if strings.HasPrefix(r.URL.Path, "/v2/sprints/") {
		trimmed := strings.TrimPrefix(r.URL.Path, "/v2/sprints/")
		parts := strings.Split(trimmed, "/")

		if len(parts) == 1 {
			id, err := strconv.ParseInt(parts[0], 10, 64)
			if err == nil {
				switch r.Method {
				case http.MethodGet:
					handler.GlobalSprintHandler.HandleGet(w, r, id)
					return
				case http.MethodPut, http.MethodPatch:
					handler.GlobalSprintHandler.HandleUpdate(w, r, id)
					return
				case http.MethodDelete:
					handler.GlobalSprintHandler.HandleDelete(w, r, id)
					return
				}
			}
		}

		if len(parts) == 2 {
			id, err := strconv.ParseInt(parts[0], 10, 64)
			if err == nil {
				switch parts[1] {
				case "start":
					if r.Method == http.MethodPost {
						handler.GlobalSprintHandler.HandleStart(w, r, id)
						return
					}
				case "complete":
					if r.Method == http.MethodPost {
						handler.GlobalSprintHandler.HandleComplete(w, r, id)
						return
					}
				case "close":
					if r.Method == http.MethodPost {
						handler.GlobalSprintHandler.HandleClose(w, r, id)
						return
					}
				case "tasks":
					if r.Method == http.MethodGet {
						handler.GlobalSprintHandler.HandleGetTasks(w, r, id)
						return
					}
					if r.Method == http.MethodPost {
						handler.GlobalSprintHandler.HandleAddTask(w, r, id)
						return
					}
				case "metrics":
					if r.Method == http.MethodGet {
						handler.GlobalSprintHandler.HandleGetMetrics(w, r, id)
						return
					}
				}
			}
		}

		if len(parts) == 3 && parts[1] == "tasks" {
			sprintID, err1 := strconv.ParseInt(parts[0], 10, 64)
			taskID, err2 := strconv.ParseInt(parts[2], 10, 64)
			if err1 == nil && err2 == nil {
				if r.Method == http.MethodDelete {
					handler.GlobalSprintHandler.HandleRemoveTask(w, r, sprintID, taskID)
					return
				}
				if r.Method == http.MethodPut || r.Method == http.MethodPatch {
					handler.GlobalSprintHandler.HandleUpdateStoryPoints(w, r, sprintID, taskID)
					return
				}
			}
		}
	}

	// Task Custom Fields API (/v2/tasks/{id}/custom-fields)
	if strings.HasPrefix(r.URL.Path, "/v2/tasks/") && strings.Contains(r.URL.Path, "/custom-fields") {
		trimmed := strings.TrimPrefix(r.URL.Path, "/v2/tasks/")
		parts := strings.Split(trimmed, "/")
		if len(parts) >= 2 && parts[1] == "custom-fields" {
			taskID, err := strconv.ParseInt(parts[0], 10, 64)
			if err == nil {
				if len(parts) == 2 {
					if r.Method == http.MethodGet {
						handler.GlobalCustomFieldHandler.HandleGetTaskCustomFields(w, r, taskID)
						return
					}
					if r.Method == http.MethodPost || r.Method == http.MethodPut {
						handler.GlobalCustomFieldHandler.HandleSetTaskCustomField(w, r, taskID)
						return
					}
				}
				if len(parts) == 3 && r.Method == http.MethodDelete {
					fieldID, err := strconv.ParseInt(parts[2], 10, 64)
					if err == nil {
						handler.GlobalCustomFieldHandler.HandleDeleteTaskCustomField(w, r, taskID, fieldID)
						return
					}
				}
			}
		}
	}

	// Custom Field Definitions API (/v2/custom-fields)
	if r.URL.Path == "/v2/custom-fields" {
		if r.Method == http.MethodPost {
			handler.GlobalCustomFieldHandler.HandleCreateDefinition(w, r)
			return
		}
		if r.Method == http.MethodGet {
			handler.GlobalCustomFieldHandler.HandleListDefinitions(w, r)
			return
		}
	}
	if r.URL.Path == "/v2/custom-fields/bulk" && r.Method == http.MethodPost {
		handler.GlobalCustomFieldHandler.HandleBulkSet(w, r)
		return
	}
	if r.URL.Path == "/v2/custom-fields/search" && r.Method == http.MethodGet {
		handler.GlobalCustomFieldHandler.HandleFindTasksByValue(w, r)
		return
	}
	if strings.HasPrefix(r.URL.Path, "/v2/custom-fields/") {
		trimmed := strings.TrimPrefix(r.URL.Path, "/v2/custom-fields/")
		id, err := strconv.ParseInt(trimmed, 10, 64)
		if err == nil {
			switch r.Method {
			case http.MethodGet:
				handler.GlobalCustomFieldHandler.HandleGetDefinition(w, r, id)
				return
			case http.MethodPut, http.MethodPatch:
				handler.GlobalCustomFieldHandler.HandleUpdateDefinition(w, r, id)
				return
			case http.MethodDelete:
				handler.GlobalCustomFieldHandler.HandleDeleteDefinition(w, r, id)
				return
			}
		}
	}

	// Task Attachments API (/v2/tasks/{id}/attachments)
	if strings.HasPrefix(r.URL.Path, "/v2/tasks/") && strings.Contains(r.URL.Path, "/attachments") {
		trimmed := strings.TrimPrefix(r.URL.Path, "/v2/tasks/")
		parts := strings.Split(trimmed, "/")
		if len(parts) >= 2 && parts[1] == "attachments" {
			taskID, err := strconv.ParseInt(parts[0], 10, 64)
			if err == nil {
				if r.Method == http.MethodPost {
					handler.GlobalAttachmentHandler.HandleUploadAttachment(w, r, taskID)
					return
				}
				if r.Method == http.MethodGet {
					handler.GlobalAttachmentHandler.HandleListAttachments(w, r, &taskID)
					return
				}
			}
		}
	}

	// File Assets & Attachments API (/v2/attachments)
	if r.URL.Path == "/v2/attachments" && r.Method == http.MethodGet {
		handler.GlobalAttachmentHandler.HandleListAttachments(w, r, nil)
		return
	}
	if r.URL.Path == "/v2/attachments/metrics" && r.Method == http.MethodGet {
		handler.GlobalAttachmentHandler.HandleGetStorageMetrics(w, r, 0)
		return
	}
	if strings.HasPrefix(r.URL.Path, "/v2/attachments/") {
		trimmed := strings.TrimPrefix(r.URL.Path, "/v2/attachments/")
		parts := strings.Split(trimmed, "/")
		id, err := strconv.ParseInt(parts[0], 10, 64)
		if err == nil {
			if len(parts) == 2 && parts[1] == "download" && r.Method == http.MethodGet {
				handler.GlobalAttachmentHandler.HandleDownloadAttachment(w, r, id)
				return
			}
			if len(parts) == 1 {
				switch r.Method {
				case http.MethodGet:
					handler.GlobalAttachmentHandler.HandleGetAttachment(w, r, id)
					return
				case http.MethodPut, http.MethodPatch:
					handler.GlobalAttachmentHandler.HandleUpdateAttachment(w, r, id)
					return
				case http.MethodDelete:
					handler.GlobalAttachmentHandler.HandleDeleteAttachment(w, r, id)
					return
				}
			}
		}
	}

	// Task Templates & Blueprints API (/v2/templates)
	if r.URL.Path == "/v2/templates" {
		if r.Method == http.MethodPost {
			handler.GlobalTemplateHandler.HandleCreateTemplate(w, r)
			return
		}
		if r.Method == http.MethodGet {
			handler.GlobalTemplateHandler.HandleListTemplates(w, r)
			return
		}
	}
	if r.URL.Path == "/v2/templates/stats" && r.Method == http.MethodGet {
		handler.GlobalTemplateHandler.HandleGetTemplateStats(w, r, 0)
		return
	}
	if r.URL.Path == "/v2/templates/import" && r.Method == http.MethodPost {
		handler.GlobalTemplateHandler.HandleImportTemplate(w, r)
		return
	}
	if strings.HasPrefix(r.URL.Path, "/v2/templates/") {
		trimmed := strings.TrimPrefix(r.URL.Path, "/v2/templates/")
		parts := strings.Split(trimmed, "/")
		id, err := strconv.ParseInt(parts[0], 10, 64)
		if err == nil {
			if len(parts) == 2 {
				switch parts[1] {
				case "instantiate":
					if r.Method == http.MethodPost {
						handler.GlobalTemplateHandler.HandleInstantiateTemplate(w, r, id)
						return
					}
				case "duplicate":
					if r.Method == http.MethodPost {
						handler.GlobalTemplateHandler.HandleDuplicateTemplate(w, r, id)
						return
					}
				case "export":
					if r.Method == http.MethodGet {
						handler.GlobalTemplateHandler.HandleExportTemplate(w, r, id)
						return
					}
				}
			}
			if len(parts) == 1 {
				switch r.Method {
				case http.MethodGet:
					handler.GlobalTemplateHandler.HandleGetTemplate(w, r, id)
					return
				case http.MethodPut, http.MethodPatch:
					handler.GlobalTemplateHandler.HandleUpdateTemplate(w, r, id)
					return
				case http.MethodDelete:
					handler.GlobalTemplateHandler.HandleDeleteTemplate(w, r, id)
					return
				}
			}
		}
	}

	// Saved Views & Custom Filter Presets API (/v2/views)
	if r.URL.Path == "/v2/views" {
		if r.Method == http.MethodPost {
			handler.GlobalSavedViewHandler.HandleCreate(w, r)
			return
		}
		if r.Method == http.MethodGet {
			handler.GlobalSavedViewHandler.HandleList(w, r)
			return
		}
	}
	if strings.HasPrefix(r.URL.Path, "/v2/views/") {
		trimmed := strings.TrimPrefix(r.URL.Path, "/v2/views/")
		parts := strings.Split(trimmed, "/")
		id, err := strconv.ParseInt(parts[0], 10, 64)
		if err == nil {
			if len(parts) == 2 {
				switch parts[1] {
				case "execute":
					if r.Method == http.MethodGet || r.Method == http.MethodPost {
						handler.GlobalSavedViewHandler.HandleExecute(w, r, id)
						return
					}
				case "pin":
					if r.Method == http.MethodPost || r.Method == http.MethodPatch {
						handler.GlobalSavedViewHandler.HandleTogglePin(w, r, id)
						return
					}
				case "duplicate":
					if r.Method == http.MethodPost {
						handler.GlobalSavedViewHandler.HandleDuplicate(w, r, id)
						return
					}
				}
			}
			if len(parts) == 1 {
				switch r.Method {
				case http.MethodGet:
					handler.GlobalSavedViewHandler.HandleGetByID(w, r, id)
					return
				case http.MethodPut, http.MethodPatch:
					handler.GlobalSavedViewHandler.HandleUpdate(w, r, id)
					return
				case http.MethodDelete:
					handler.GlobalSavedViewHandler.HandleDelete(w, r, id)
					return
				}
			}
		}
	}

	// Task Automations & Workflow Engine API (/v2/automations)
	if r.URL.Path == "/v2/automations" {
		if r.Method == http.MethodPost {
			handler.GlobalAutomationHandler.HandleCreateRule(w, r)
			return
		}
		if r.Method == http.MethodGet {
			handler.GlobalAutomationHandler.HandleListRules(w, r)
			return
		}
	}
	if r.URL.Path == "/v2/automations/metrics" && r.Method == http.MethodGet {
		handler.GlobalAutomationHandler.HandleGetMetrics(w, r)
		return
	}
	if r.URL.Path == "/v2/automations/logs" && r.Method == http.MethodGet {
		handler.GlobalAutomationHandler.HandleListLogs(w, r)
		return
	}
	if r.URL.Path == "/v2/automations/dispatch" && r.Method == http.MethodPost {
		handler.GlobalAutomationHandler.HandleDispatchEvent(w, r)
		return
	}
	if strings.HasPrefix(r.URL.Path, "/v2/automations/") {
		trimmed := strings.TrimPrefix(r.URL.Path, "/v2/automations/")
		parts := strings.Split(trimmed, "/")
		id, err := strconv.ParseInt(parts[0], 10, 64)
		if err == nil {
			if len(parts) == 2 {
				switch parts[1] {
				case "toggle":
					if r.Method == http.MethodPost || r.Method == http.MethodPatch {
						handler.GlobalAutomationHandler.HandleToggleRule(w, r, id)
						return
					}
				case "duplicate":
					if r.Method == http.MethodPost {
						handler.GlobalAutomationHandler.HandleDuplicateRule(w, r, id)
						return
					}
				case "dry-run":
					if r.Method == http.MethodPost {
						handler.GlobalAutomationHandler.HandleDryRunRule(w, r, id)
						return
					}
				case "logs":
					if r.Method == http.MethodGet {
						handler.GlobalAutomationHandler.HandleGetRuleLogs(w, r, id)
						return
					}
				}
			}
			if len(parts) == 1 {
				switch r.Method {
				case http.MethodGet:
					handler.GlobalAutomationHandler.HandleGetRule(w, r, id)
					return
				case http.MethodPut, http.MethodPatch:
					handler.GlobalAutomationHandler.HandleUpdateRule(w, r, id)
					return
				case http.MethodDelete:
					handler.GlobalAutomationHandler.HandleDeleteRule(w, r, id)
					return
				}
			}
		}
	}

	// Releases & Milestones Management API (/v2/releases)
	if r.URL.Path == "/v2/releases" {
		handler.GlobalReleaseHandler.HandleReleasesRoot(w, r)
		return
	}
	if strings.HasPrefix(r.URL.Path, "/v2/releases/") {
		trimmed := strings.TrimPrefix(r.URL.Path, "/v2/releases/")
		parts := strings.Split(trimmed, "/")
		id, err := strconv.ParseInt(parts[0], 10, 64)
		if err == nil {
			if len(parts) == 1 {
				switch r.Method {
				case http.MethodGet:
					handler.GlobalReleaseHandler.HandleGetRelease(w, r, id)
					return
				case http.MethodPut, http.MethodPatch:
					handler.GlobalReleaseHandler.HandleUpdateRelease(w, r, id)
					return
				case http.MethodDelete:
					handler.GlobalReleaseHandler.HandleDeleteRelease(w, r, id)
					return
				}
			} else if len(parts) == 2 {
				switch parts[1] {
				case "status":
					handler.GlobalReleaseHandler.HandleUpdateReleaseStatus(w, r)
					return
				case "tasks":
					handler.GlobalReleaseHandler.HandleReleaseTasks(w, r)
					return
				case "changelog", "notes":
					handler.GlobalReleaseHandler.HandleReleaseChangelog(w, r)
					return
				case "metrics":
					handler.GlobalReleaseHandler.HandleReleaseMetrics(w, r)
					return
				case "clone":
					handler.GlobalReleaseHandler.HandleCloneRelease(w, r)
					return
				}
			} else if len(parts) == 3 && parts[1] == "tasks" {
				handler.GlobalReleaseHandler.HandleReleaseTaskDetail(w, r)
				return
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
