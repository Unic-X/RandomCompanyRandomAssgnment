package main

import (
	"net/http"

	"github.com/Unic-X/slow-server/api"
	"github.com/Unic-X/slow-server/config"
	"github.com/Unic-X/slow-server/logger"
	"github.com/Unic-X/slow-server/middleware"
	"github.com/charmbracelet/log"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)


func main() {
	// Initialize the logger
    go logger.InitLogger()
	
	// Load configuration
	cfg := config.LoadConfig()

	// Set up router and middleware
	router := http.NewServeMux()
	
	// Register API handlers
	router.HandleFunc("/api/data", api.GetDataHandler)
	router.HandleFunc("/api/users", api.GetUsersHandler)
	router.HandleFunc("/api/process", api.ProcessDataHandler)

    router.Handle("/metrics", promhttp.Handler())
	wrappedRouter := middleware.ApplyMetricsMiddleware(router)
	wrappedRouter = middleware.ApplyLoggingMiddleware(wrappedRouter)

	// Start server
	log.Info("Starting server on port", "port", cfg.PortString())
	err := http.ListenAndServe(":"+cfg.PortString(), wrappedRouter)
	if err != nil {
		log.Fatal("Server failed to start", "error", err)
	}
}
