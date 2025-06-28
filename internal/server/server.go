package server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"mime"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"strconv"
	"syscall"
	"time"

	"github.com/mattjh1/psi-map/internal/constants"
	"github.com/mattjh1/psi-map/internal/logger"
	"github.com/mattjh1/psi-map/internal/types"
)

type Server struct {
	results []*types.PageResult
	port    string
	server  *http.Server
}

// Start initializes and starts the web server
func Start(results []*types.PageResult, port string) error {
	log := logger.GetLogger()

	// Find an available port if the default is taken
	availablePort, err := findAvailablePort(port)
	if err != nil {
		return fmt.Errorf("failed to find available port: %w", err)
	}

	s := &Server{
		results: results,
		port:    availablePort,
	}

	// Setup routes
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handleReport)
	mux.HandleFunc("/api/results", s.handleAPIResults)
	mux.HandleFunc("/api/results/", s.handleAPIResult)
	mux.HandleFunc("/api/report-data", s.handleReportData)
	mux.HandleFunc("/static/", s.handleStatic)

	s.server = &http.Server{
		Addr:              ":" + s.port,
		Handler:           mux,
		ReadHeaderTimeout: constants.ReadHeaderTimeout, // Prevents Slowloris attacks
		ReadTimeout:       constants.ReadTimeout,
		WriteTimeout:      constants.WriteTimeout,
		IdleTimeout:       constants.IdleTimeout,
	}

	// Start server in goroutine
	go func() {
		log.Tagged("SERVER", "Starting web server on http://localhost:%s", "🌐", s.port)
		log.Info("Press Ctrl+C to stop the server")
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("Server error: %v", err)
		}
	}()

	// Auto-open browser
	go func() {
		time.Sleep(1 * time.Second) // Give server time to start
		openBrowser(fmt.Sprintf("http://localhost:%s", s.port))
	}()

	// Wait for interrupt signal
	return s.waitForShutdown()
}

// handleReport serves the main report page
func (s *Server) handleReport(w http.ResponseWriter, r *http.Request) {
	log := logger.GetLogger()

	tmpl, err := loadReportTemplateFromFS()
	if err != nil {
		log.Error("Template loading error: %v", err)
		http.Error(w, fmt.Sprintf("Template error: %v", err), http.StatusInternalServerError)
		return
	}

	data := struct {
		Results   []*types.PageResult
		Summary   types.ReportSummary
		Generated time.Time
	}{
		Results:   s.results,
		Summary:   s.generateSummary(),
		Generated: time.Now(),
	}

	// Use a bytes.Buffer to catch template output first
	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, "report.html", data); err != nil {
		log.Error("Template execution error: %v", err)
		http.Error(w, fmt.Sprintf("Template execution error: %v", err), http.StatusInternalServerError)
		return
	}

	// Only write to ResponseWriter after template succeeded
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(buf.Bytes())
}

// handleAPIResults serves JSON data for all results
func (s *Server) handleAPIResults(w http.ResponseWriter, r *http.Request) {
	log := logger.GetLogger()

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(s.results); err != nil {
		log.Error("Failed to encode API results: %v", err)
		http.Error(w, fmt.Sprintf("failed to encode results: %v", err), http.StatusInternalServerError)
	}
}

// handleAPIResult serves JSON data for a specific result
func (s *Server) handleAPIResult(w http.ResponseWriter, r *http.Request) {
	log := logger.GetLogger()

	indexStr := r.URL.Path[len("/api/results/"):]
	index, err := strconv.Atoi(indexStr)
	if err != nil || index < 0 || index >= len(s.results) {
		log.Warn("Invalid result index requested: %s", indexStr)
		http.Error(w, "Invalid result index", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(s.results[index]); err != nil {
		log.Error("Failed to encode single API result: %v", err)
		http.Error(w, fmt.Sprintf("failed to encode results: %v", err), http.StatusInternalServerError)
	}
}

// handleStatic serves static CSS and JS files with correct MIME types
func (s *Server) handleStatic(w http.ResponseWriter, r *http.Request) {
	// Create a file server for the static directory
	fs := http.FileServer(http.Dir("internal/server/static"))

	// Strip the /static/ prefix
	path := r.URL.Path
	if len(path) >= len("/static/") {
		path = path[len("/static/"):]
	}

	// Set the correct MIME type based on file extension
	ext := filepath.Ext(path)
	mimeType := mime.TypeByExtension(ext)
	if mimeType == "" {
		// Fallback MIME types
		switch ext {
		case ".css":
			mimeType = "text/css"
		case ".js":
			mimeType = "application/javascript"
		default:
			mimeType = "application/octet-stream"
		}
	}

	// Set headers
	w.Header().Set("Content-Type", mimeType)
	w.Header().Set("Cache-Control", "max-age=31536000") // Cache for 1 year

	// Serve the file with stripped prefix
	http.StripPrefix("/static/", fs).ServeHTTP(w, r)
}

// handleReportData serves Results and Summary as JSON
func (s *Server) handleReportData(w http.ResponseWriter, r *http.Request) {
	log := logger.GetLogger()

	data := struct {
		Results []*types.PageResult `json:"results"`
		Summary types.ReportSummary `json:"summary"`
	}{
		Results: s.results,
		Summary: s.generateSummary(),
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Error("Failed to encode report data: %v", err)
		http.Error(w, fmt.Sprintf("Error encoding JSON: %v", err), http.StatusInternalServerError)
		return
	}
}

// findAvailablePort finds an available port starting from the given port
func findAvailablePort(preferredPort string) (string, error) {
	log := logger.GetLogger()

	port, err := strconv.Atoi(preferredPort)
	if err != nil {
		log.Warn("Invalid port '%s', defaulting to 8080", preferredPort)
		port = 8080
	}

	for i := range 10 {
		testPort := port + i
		listener, err := net.Listen("tcp", fmt.Sprintf(":%d", testPort))
		if err == nil {
			listener.Close()
			if testPort != port {
				log.Tagged("SERVER", "Port %d was busy, using port %d instead", "🔄", port, testPort)
			}
			return strconv.Itoa(testPort), nil
		}
	}

	return "", fmt.Errorf("no available ports found")
}

// openBrowser opens the URL in the default browser
func openBrowser(url string) {
	log := logger.GetLogger()

	var err error
	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		log.Tagged("SERVER", "Please open %s in your browser", "🌐", url)
		return
	}

	if err != nil {
		log.Warn("Could not auto-open browser: %v", err)
		log.Tagged("SERVER", "Please visit: %s", "🌐", url)
	}
}

// waitForShutdown waits for interrupt signal and gracefully shuts down
func (s *Server) waitForShutdown() error {
	log := logger.GetLogger()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Tagged("SERVER", "Shutting down server gracefully...", "🛑")

	ctx, cancel := context.WithTimeout(context.Background(), constants.ShutdownTimeout)
	defer cancel()

	if err := s.server.Shutdown(ctx); err != nil {
		log.Error("Error during server shutdown: %v", err)
		return err
	}

	log.Success("Server shutdown completed")
	return nil
}
