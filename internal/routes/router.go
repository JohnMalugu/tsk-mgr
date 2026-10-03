package routes

import (
	"encoding/json"
	"fmt"
	"net/http"
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

	if r.URL.Path == "/tasks" {
		handler.HandleTasks(w, r)
		return
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

	if strings.HasSuffix(r.URL.Path, "/complete") && strings.HasPrefix(r.URL.Path, "/tasks/") {
		handler.HandleTaskComplete(w, r)
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
