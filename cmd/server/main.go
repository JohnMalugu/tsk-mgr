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
	fmt.Printf("   GET    http://localhost:8080/tasks\n")
	fmt.Printf("   GET    http://localhost:8080/tasks/summary\n")
	fmt.Printf("   POST   http://localhost:8080/tasks/bulk/complete\n")
	fmt.Printf("   POST   http://localhost:8080/tasks/bulk/delete\n")
	fmt.Printf("   POST   http://localhost:8080/tasks/bulk/priority\n")
	fmt.Printf("   POST   http://localhost:8080/tasks\n")
	fmt.Printf("   GET    http://localhost:8080/tasks/{id}\n")
	fmt.Printf("   PUT    http://localhost:8080/tasks/{id}\n")
	fmt.Printf("   PATCH  http://localhost:8080/tasks/{id}\n")
	fmt.Printf("   DELETE http://localhost:8080/tasks/{id}\n")
	fmt.Printf("   PATCH  http://localhost:8080/tasks/{id}/complete\n")
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
