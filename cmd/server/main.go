cat > cmd/server/main.go << 'EOF'
package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/yourusername/task-manager-api/internal/routes"
)

func main() {
	// Register the router for all requests
	http.HandleFunc("/", routes.Router)

	// Start the server
	port := ":8080"
	fmt.Printf("🚀 Server running on http://localhost:8080\n")
	fmt.Printf("📚 Endpoints:\n")
	fmt.Printf("   GET    http://localhost:8080/tasks\n")
	fmt.Printf("   POST   http://localhost:8080/tasks\n")
	fmt.Printf("   GET    http://localhost:8080/tasks/{id}\n")
	fmt.Printf("\n")

	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatal(err)
	}
}
EOF