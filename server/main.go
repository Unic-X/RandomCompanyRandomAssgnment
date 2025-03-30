package main

import (
    "github.com/charmbracelet/log"
	"net/http"
	"github.com/Unic-X/slow-server/api"
	"github.com/Unic-X/slow-server/config"
	"github.com/Unic-X/slow-server/middleware"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig()

	// Set up router and middleware
	router := http.NewServeMux()
	
	// Register API handlers
	router.HandleFunc("/api/data", api.GetDataHandler)
	router.HandleFunc("/api/users", api.GetUsersHandler)
	router.HandleFunc("/api/process", api.ProcessDataHandler)
	
	// Apply middleware to all requests
	wrappedRouter := middleware.ApplyMetricsMiddleware(router)
	wrappedRouter = middleware.ApplyLoggingMiddleware(wrappedRouter)

	// Start server
	log.Info("Starting server","on port",cfg.Port)
	err := http.ListenAndServe(":"+cfg.PortString(), wrappedRouter)
	if err != nil {
		log.Fatal("Server failed to start: %v", err)
	}
}
