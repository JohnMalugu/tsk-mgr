package main

import (
	"fmt"
	"log"
	"net/http"
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
	fmt.Printf("   GET    http://localhost:8080/tasks\n")
	fmt.Printf("   GET    http://localhost:8080/tasks/summary\n")
	fmt.Printf("   POST   http://localhost:8080/tasks/bulk/complete\n")
	fmt.Printf("   POST   http://localhost:8080/tasks\n")
	fmt.Printf("   GET    http://localhost:8080/tasks/{id}\n")
	fmt.Printf("   PUT    http://localhost:8080/tasks/{id}\n")
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
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
