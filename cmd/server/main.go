package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/JohnMalugu/tsk-mgr-api/internal/routes"
)

func main() {
	// Register the router for all requests
	http.HandleFunc("/", routes.Router)

	// Start the server
	port := ":8080"
	fmt.Printf("🚀 Server running on http://localhost:8080\n")
	fmt.Printf("📚 Endpoints:\n")
	fmt.Printf("   GET    http://localhost:8080/health\n")
	fmt.Printf("   GET    http://localhost:8080/activity?offset=0&limit=20&action={action}&from={timestamp}&to={timestamp}\n")
	fmt.Printf("   GET    http://localhost:8080/tags?offset=0&limit=20\n")
	fmt.Printf("   GET    http://localhost:8080/timer\n")
	fmt.Printf("   GET    http://localhost:8080/time/report?taskId={id}&from={timestamp}&to={timestamp}\n")
	fmt.Printf("   GET    http://localhost:8080/tasks\n")
	fmt.Printf("   GET    http://localhost:8080/tasks/summary\n")
	fmt.Printf("   GET    http://localhost:8080/tasks/ready\n")
	fmt.Printf("   GET    http://localhost:8080/tasks/blocked\n")
	fmt.Printf("   GET    http://localhost:8080/tasks/upcoming?days=7\n")
	fmt.Printf("   GET    http://localhost:8080/tasks/{id}/occurrences\n")
	fmt.Printf("   GET    http://localhost:8080/tasks/{id}/recurrence/next\n")
	fmt.Printf("   POST   http://localhost:8080/tasks/{id}/recurrence/skip\n")
	fmt.Printf("   GET    http://localhost:8080/tasks?recurring=true|false\n")
	fmt.Printf("   GET    http://localhost:8080/tasks/{id}/checklist\n")
	fmt.Printf("   POST   http://localhost:8080/tasks/{id}/checklist\n")
	fmt.Printf("   PUT    http://localhost:8080/tasks/{id}/checklist/order\n")
	fmt.Printf("   POST   http://localhost:8080/tasks/{id}/checklist/complete\n")
	fmt.Printf("   GET    http://localhost:8080/tasks/{id}/checklist/progress\n")
	fmt.Printf("   PATCH  http://localhost:8080/tasks/{id}/checklist/{itemId}\n")
	fmt.Printf("   DELETE http://localhost:8080/tasks/{id}/checklist/{itemId}\n")
	fmt.Printf("   GET    http://localhost:8080/tasks/{id}/dependencies\n")
	fmt.Printf("   GET    http://localhost:8080/tasks/{id}/dependents\n")
	fmt.Printf("   GET    http://localhost:8080/tasks/{id}/activity?offset=0&limit=20&action={action}\n")
	fmt.Printf("   GET    http://localhost:8080/tasks/{id}/comments\n")
	fmt.Printf("   POST   http://localhost:8080/tasks/{id}/comments\n")
	fmt.Printf("   DELETE http://localhost:8080/tasks/{id}/comments/{commentId}\n")
	fmt.Printf("   PATCH  http://localhost:8080/tasks/{id}/comments/{commentId}\n")
	fmt.Printf("   GET    http://localhost:8080/tasks/{id}/time\n")
	fmt.Printf("   POST   http://localhost:8080/tasks/{id}/time\n")
	fmt.Printf("   POST   http://localhost:8080/tasks/{id}/timer/start\n")
	fmt.Printf("   POST   http://localhost:8080/tasks/{id}/timer/stop\n")
	fmt.Printf("   DELETE http://localhost:8080/tasks/{id}/time/{entryId}\n")
	fmt.Printf("   POST   http://localhost:8080/tasks/{id}/dependencies\n")
	fmt.Printf("   DELETE http://localhost:8080/tasks/{id}/dependencies/{dependencyId}\n")
	fmt.Printf("   POST   http://localhost:8080/tasks/bulk/complete\n")
	fmt.Printf("   POST   http://localhost:8080/tasks/bulk/delete\n")
		fmt.Printf("   POST   http://localhost:8080/tasks/bulk/duplicate\n")
	fmt.Printf("   POST   http://localhost:8080/tasks/bulk/priority\n")
	fmt.Printf("   POST   http://localhost:8080/tasks/bulk/due-date\n")
	fmt.Printf("   POST   http://localhost:8080/tasks/bulk/tags\n")
	fmt.Printf("   POST   http://localhost:8080/tasks/bulk/status\n")
	fmt.Printf("   POST   http://localhost:8080/tasks/bulk/estimate\n")
	fmt.Printf("   POST   http://localhost:8080/tasks\n")
	fmt.Printf("   GET    http://localhost:8080/tasks/{id}\n")
	fmt.Printf("   POST   http://localhost:8080/tasks/{id}/duplicate\n")
	fmt.Printf("   PUT    http://localhost:8080/tasks/{id}\n")
	fmt.Printf("   PATCH  http://localhost:8080/tasks/{id}\n")
	fmt.Printf("   DELETE http://localhost:8080/tasks/{id}\n")
	fmt.Printf("   PATCH  http://localhost:8080/tasks/{id}/complete\n")
	fmt.Printf("   POST   http://localhost:8080/tasks/{id}/reopen\n")
	fmt.Printf("\n")

	server := &http.Server{
		Addr:              port,
		Handler:           nil,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	serverErrors := make(chan error, 1)
	go func() {
		serverErrors <- server.ListenAndServe()
	}()

	shutdownSignals := make(chan os.Signal, 1)
	signal.Notify(shutdownSignals, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(shutdownSignals)

	select {
	case err := <-serverErrors:
		if !errors.Is(err, http.ErrServerClosed) {
			log.Fatal(err)
		}
	case <-shutdownSignals:
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(ctx); err != nil {
			log.Printf("graceful shutdown failed: %v", err)
			if closeErr := server.Close(); closeErr != nil {
				log.Printf("forced shutdown failed: %v", closeErr)
			}
		}
	}
}
