package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/mattjh1/psi-map/internal/api"
	"github.com/mattjh1/psi-map/internal/logger"
)

var (
	version   = "dev"
	commit    = "none"
	buildTime = "unknown"
)

// @title PSI-Map API
// @version 1.0
// @description API for analyzing website performance using PageSpeed Insights
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /api/v1
// @schemes http https

func main() {
	// Initialize logger
	logger.Init(logger.WithOutput(os.Stderr))
	log := logger.GetLogger()

	// Get port from environment or default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Create API server
	server := api.NewServer(version, commit, buildTime)

	// Create HTTP server
	httpServer := &http.Server{
		Addr:         ":" + port,
		Handler:      server.Router(),
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 300 * time.Second, // Extended for long-running analyses
		IdleTimeout:  120 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		log.Info("Starting PSI-Map API server on port %s", port)
		log.Info("Version: %s, Commit: %s, Build Time: %s", version, commit, buildTime)
		log.Info("Swagger docs available at: http://localhost:%s/swagger/", port)

		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("Failed to start server: %v", err)
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down server...")

	// Give outstanding requests 30 seconds to complete
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		log.Error("Server forced to shutdown: %v", err)
		os.Exit(1)
	}

	log.Info("Server shutdown complete")
}
