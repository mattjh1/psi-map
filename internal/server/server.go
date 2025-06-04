package server

import (
	"bytes"
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
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

	"github.com/mattjh1/psi-map/internal/types"
)

//go:embed templates/*.html templates/partials/*.html
var templateFS embed.FS

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
		Addr:    ":" + s.port,
		Handler: mux,
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
	json.NewEncoder(w).Encode(s.results)
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
	json.NewEncoder(w).Encode(s.results[index])
}

// handleStatic serves static CSS/JS (embedded in template for simplicity)
func (s *Server) handleStatic(w http.ResponseWriter, r *http.Request) {
	// For now, we'll embed CSS/JS in the HTML template
	// In a more complex setup, you'd serve actual static files here
	http.NotFound(w, r)
}

// generateSummary creates aggregate statistics from results
func (s *Server) generateSummary() types.ReportSummary {
	summary := types.ReportSummary{
		TotalPages:        len(s.results),
		AverageScores:     make(map[string]float64),
		ScoreDistribution: make(map[string][]int),
	}

	var (
		totalScores = make(map[string]float64)
		scoreCounts = make(map[string]int)
		fastestTime = time.Hour
		slowestTime time.Duration
	)

	categories := []string{"performance", "accessibility", "best_practices", "seo"}
	for _, cat := range categories {
		summary.ScoreDistribution[cat] = []int{0, 0, 0}
	}

	for i := range s.results {
		pageResult := &s.results[i]

		// Count this as successful if either mobile or desktop succeeded
		if pageResult.Mobile.Error == nil || pageResult.Desktop.Error == nil {
			summary.SuccessfulPages++
		} else {
			summary.FailedPages++
			continue
		}

		// Track timing
		if pageResult.Duration < fastestTime {
			fastestTime = pageResult.Duration
			summary.FastestPage = &pageResult.Mobile // or create a new PageResult reference
		}
		if pageResult.Duration > slowestTime {
			slowestTime = pageResult.Duration
			summary.SlowestPage = &pageResult.Mobile
		}

		// Process mobile scores
		if pageResult.Mobile.Error == nil && pageResult.Mobile.Scores != nil {
			s.processScores(pageResult.Mobile.Scores, totalScores, scoreCounts, summary.ScoreDistribution)
		}

		// Process desktop scores
		if pageResult.Desktop.Error == nil && pageResult.Desktop.Scores != nil {
			s.processScores(pageResult.Desktop.Scores, totalScores, scoreCounts, summary.ScoreDistribution)
		}
	}

	// Calculate averages
	for category, total := range totalScores {
		if count := scoreCounts[category]; count > 0 {
			summary.AverageScores[category] = total / float64(count)
		}
	}

	return summary
}

// Helper function to process scores
func (s *Server) processScores(scores *types.CategoryScores, totalScores map[string]float64, scoreCounts map[string]int, distribution map[string][]int) {
	scoreMap := map[string]float64{
		"performance":    scores.Performance,
		"accessibility":  scores.Accessibility,
		"best_practices": scores.BestPractices,
		"seo":            scores.SEO,
	}

	for category, score := range scoreMap {
		if score > 0 {
			totalScores[category] += score
			scoreCounts[category]++

			var bucket int
			switch {
			case score >= 90:
				bucket = 0 // good
			case score >= 50:
				bucket = 1 // needs improvement
			default:
				bucket = 2 // poor
			}
			distribution[category][bucket]++
		}
	}
}

// GenerateHTMLFile generates an HTML file from results without starting a server
func GenerateHTMLFile(results []types.PageResult, filename string) error {
	tmpl, err := loadReportTemplateFromFS()
	if err != nil {
		return fmt.Errorf("template parsing error: %v", err)
	}

	// Create a temporary server instance to generate the summary
	s := &Server{results: results}

	data := struct {
		Results   []types.PageResult
		Summary   types.ReportSummary
		Generated time.Time
	}{
		Results:   results,
		Summary:   s.generateSummary(),
		Generated: time.Now(),
	}

	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %v", err)
	}
	defer file.Close()

	return tmpl.Execute(file, data)
}

// GenerateSummary creates a summary from results without needing a server instance
func GenerateSummary(results []types.PageResult) types.ReportSummary {
	s := &Server{results: results}
	return s.generateSummary()
}

func loadReportTemplateFromFS() (*template.Template, error) {
	return template.New("report").Funcs(template.FuncMap{
		"formatDuration": formatDuration,
		"formatScore":    formatScore,
		"getGradeClass":  getGradeClass,
		"getScoreClass":  getScoreClass,
		"hasMetrics":     hasMetrics,
		"formatBytes":    formatBytes,
		"toJSON":         toJSON,
		"add":            func(a, b int) int { return a + b },
		"mul":            func(a, b int) int { return a * b },
		"dict":           dict,
		"getResult":      getResult,
	}).ParseFS(templateFS, "templates/*.html", "templates/partials/*.html")
}

// Utility functions for templates
func formatDuration(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%.0fms", float64(d.Nanoseconds())/1000000)
	}
	return fmt.Sprintf("%.1fs", d.Seconds())
}

func formatScore(score float64) string {
	if score == 0 {
		return "N/A"
	}
	return fmt.Sprintf("%.0f", score)
}

func getGradeClass(grade string) string {
	switch grade {
	case "good":
		return "badge-success"
	case "needs-improvement":
		return "badge-warning"
	case "poor":
		return "badge-danger"
	default:
		return "badge-secondary"
	}
}

func getScoreClass(score float64) string {
	switch {
	case score >= 90:
		return "text-success"
	case score >= 50:
		return "text-warning"
	default:
		return "text-danger"
	}
}

func hasMetrics(result types.Result) bool {
	return result.Metrics != nil
}

func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

func toJSON(v any) template.JS {
	b, err := json.Marshal(v)
	if err != nil {
		return template.JS("null")
	}
	return template.JS(b)
}

func dict(values ...any) map[string]any {
	if len(values)%2 != 0 {
		panic("dict requires an even number of arguments")
	}
	dict := make(map[string]any, len(values)/2)
	for i := 0; i < len(values); i += 2 {
		key, ok := values[i].(string)
		if !ok {
			panic("dict keys must be strings")
		}
		dict[key] = values[i+1]
	}
	return dict
}

func getResult(page types.PageResult, strategy string) types.Result {
	if strategy == "mobile" {
		return page.Mobile
	}
	return page.Desktop
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

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return s.server.Shutdown(ctx)
}
