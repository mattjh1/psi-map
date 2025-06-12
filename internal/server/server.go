package server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"strconv"
	"syscall"
	"time"

	"github.com/mattjh1/psi-map/internal/constants"
	"github.com/mattjh1/psi-map/internal/types"
)

type Server struct {
	results []types.PageResult
	port    string
	server  *http.Server
}

// Start initializes and starts the web server
func Start(results []types.PageResult, port string) error {
	// Find an available port if the default is taken
	availablePort, err := findAvailablePort(port)
	if err != nil {
		return fmt.Errorf("failed to find available port: %v", err)
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
		fmt.Printf("[INFO] Starting web server on http://localhost:%s\n", s.port)
		fmt.Println("[INFO] Press Ctrl+C to stop the server")

		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("Server error: %v", err)
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
	tmpl, err := loadReportTemplateFromFS()
	if err != nil {
		http.Error(w, fmt.Sprintf("Template error: %v", err), http.StatusInternalServerError)
		return
	}

	data := struct {
		Results   []types.PageResult
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
		http.Error(w, fmt.Sprintf("Template execution error: %v", err), http.StatusInternalServerError)
		return
	}

	// Only write to ResponseWriter after template succeeded
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(buf.Bytes())
}

// handleAPIResults serves JSON data for all results
func (s *Server) handleAPIResults(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(s.results); err != nil {
		http.Error(w, fmt.Sprintf("failed to encode results: %v", err), http.StatusInternalServerError)
	}
}

// handleAPIResult serves JSON data for a specific result
func (s *Server) handleAPIResult(w http.ResponseWriter, r *http.Request) {
	indexStr := r.URL.Path[len("/api/results/"):]
	index, err := strconv.Atoi(indexStr)
	if err != nil || index < 0 || index >= len(s.results) {
		http.Error(w, "Invalid result index", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(s.results[index]); err != nil {
		http.Error(w, fmt.Sprintf("failed to encode results: %v", err), http.StatusInternalServerError)
	}
}

// handleStatic serves static CSS/JS (embedded in template for simplicity)
func (s *Server) handleStatic(w http.ResponseWriter, r *http.Request) {
	// For now, we'll embed CSS/JS in the HTML template
	// In a more complex setup, you'd serve actual static files here
	http.NotFound(w, r)
}

// findAvailablePort finds an available port starting from the given port
func findAvailablePort(preferredPort string) (string, error) {
	port, err := strconv.Atoi(preferredPort)
	if err != nil {
		port = 8080
	}

	for i := range 10 {
		testPort := port + i
		listener, err := net.Listen("tcp", fmt.Sprintf(":%d", testPort))
		if err == nil {
			listener.Close()
			return strconv.Itoa(testPort), nil
		}
	}

	return "", fmt.Errorf("no available ports found")
}

// openBrowser opens the URL in the default browser
func openBrowser(url string) {
	var err error
	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		fmt.Printf("Please open %s in your browser\n", url)
		return
	}

	if err != nil {
		fmt.Printf("Could not auto-open browser. Please visit: %s\n", url)
	}
}

// waitForShutdown waits for interrupt signal and gracefully shuts down
func (s *Server) waitForShutdown() error {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	fmt.Println("\n[INFO] Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), constants.ShutdownTimeout)
	defer cancel()

	return s.server.Shutdown(ctx)
}
