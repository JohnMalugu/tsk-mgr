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
