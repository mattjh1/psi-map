package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/mattjh1/psi-map/internal/cache"
	"github.com/mattjh1/psi-map/internal/constants"
	"github.com/mattjh1/psi-map/internal/logger"
	"github.com/mattjh1/psi-map/internal/types"
	"github.com/mattjh1/psi-map/internal/utils"
	"github.com/mattjh1/psi-map/runner"
	httpSwagger "github.com/swaggo/http-swagger"
)

// Server represents the API server
type Server struct {
	version    string
	commit     string
	buildTime  string
	router     *mux.Router
	jobManager *JobManager
	startTime  time.Time
}

func WriteJSON(w http.ResponseWriter, status int, v any) {
	log := logger.GetLogger()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Error("failed to encode JSON response: %v", err)
	}
}

// NewServer creates a new API server instance
func NewServer(version, commit, buildTime string) *Server {
	s := &Server{
		version:    version,
		commit:     commit,
		buildTime:  buildTime,
		router:     mux.NewRouter(),
		jobManager: NewJobManager(10, 24*time.Hour), // Max 10 concurrent jobs, 24h TTL
		startTime:  time.Now(),
	}

	s.setupRoutes()
	return s
}

// Router returns the configured router
func (s *Server) Router() http.Handler {
	return s.router
}

// setupRoutes configures all API routes
func (s *Server) setupRoutes() {
	// Add middleware
	s.router.Use(corsMiddleware)
	s.router.Use(loggingMiddleware)

	// Swagger documentation
	s.router.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)

	// API v1 routes
	api := s.router.PathPrefix("/api/v1").Subrouter()

	// Health and info endpoints
	api.HandleFunc("/health", s.healthHandler).Methods("GET")
	api.HandleFunc("/version", s.versionHandler).Methods("GET")

	// Analysis endpoints
	api.HandleFunc("/analyze", s.analyzeHandler).Methods("POST")
	api.HandleFunc("/analyze/status/{jobId}", s.getJobStatusHandler).Methods("GET")
	api.HandleFunc("/analyze/jobs", s.listJobsHandler).Methods("GET")
	api.HandleFunc("/analyze/jobs/{jobId}/cancel", s.cancelJobHandler).Methods("POST")

	// Serve basic info at root
	s.router.HandleFunc("/", s.rootHandler).Methods("GET")
	s.router.PathPrefix("/").Handler(http.NotFoundHandler())
}

// @Summary Health check
// @Description Check if the API is healthy and running
// @Tags health
// @Produce json
// @Success 200 {object} HealthResponse
// @Router /health [get]
func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	uptime := time.Since(s.startTime)

	response := HealthResponse{
		Status:    "healthy",
		Timestamp: time.Now().UTC(),
		Version:   s.version,
		Uptime:    uptime.String(),
	}
	WriteJSON(w, http.StatusOK, response)
}

// @Summary Get version information
// @Description Get version, commit, and build information
// @Tags info
// @Produce json
// @Success 200 {object} VersionResponse
// @Router /version [get]
func (s *Server) versionHandler(w http.ResponseWriter, r *http.Request) {
	response := VersionResponse{
		Version:   s.version,
		Commit:    s.commit,
		BuildTime: s.buildTime,
	}

	WriteJSON(w, http.StatusOK, response)
}

// Basic PSI request (no auth needed):
/*
{
  "sitemap": "https://example.com/sitemap.xml",
  "provider": "psi",
  "async": false
}
*/

// Lighthouse with Bearer token:
/*
{
  "sitemap": "https://example.com/sitemap.xml",
  "provider": "lighthouse",
  "lighthouse_url": "https://lighthouse-api.example.com",
  "async": true,
  "auth": {
    "bearer_token": "your-bearer-token-here"
  }
}
*/

// Lighthouse with Cloudflare Access:
/*
{
  "sitemap": "https://example.com/sitemap.xml",
  "provider": "lighthouse",
  "lighthouse_url": "https://lighthouse-api.mattjh.sh",
  "async": true,
  "auth": {
    "cloudflare_access": {
      "client_id": "$CF_ACCESS_CLIENT_ID",
      "client_secret": "$CF_ACCESS_CLIENT_SECRET"
    }
  }
}
*/

// @Summary Analyze website performance
// @Description Start performance analysis of a website sitemap with optional authentication for external services
// @Tags analysis
// @Accept json
// @Produce json
// @Param request body AnalysisRequest true "Analysis configuration with optional authentication"
// @Success 200 {object} AnalysisResponse "Synchronous analysis completed"
// @Success 202 {object} AsyncAnalysisResponse "Asynchronous analysis started"
// @Failure 400 {object} ErrorResponse "Bad request"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /analyze [post]
func (s *Server) analyzeHandler(w http.ResponseWriter, r *http.Request) {
	log := logger.GetLogger()

	var req AnalysisRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, "Invalid JSON", err)
		return
	}

	// Validate request
	if err := req.Validate(); err != nil {
		s.writeError(w, http.StatusBadRequest, "Validation error", err)
		return
	}

	// Set defaults
	req.SetDefaults()

	log.Info("Analysis request for: %s (provider: %s, async: %v)", req.Sitemap, req.Provider, req.Async)

	// Convert API authentication config to internal types
	var authConfig *types.AuthenticationConfig
	if req.Auth != nil {
		authConfig = &types.AuthenticationConfig{
			BearerToken: req.Auth.BearerToken,
		}

		if req.Auth.CloudflareAccess != nil {
			authConfig.CloudflareAccess = &types.CloudflareAccessConfig{
				ClientID:     req.Auth.CloudflareAccess.ClientID,
				ClientSecret: req.Auth.CloudflareAccess.ClientSecret,
			}
		}
	}

	// Create analysis configuration
	config := &types.AnalysisConfig{
		Sitemap:       req.Sitemap,
		OutputFormat:  constants.JSON,
		UseStdout:     false,
		StartServer:   false,
		MaxWorkers:    req.MaxWorkers,
		CacheTTL:      req.CacheTTL,
		Provider:      req.Provider,
		LighthouseURL: req.LighthouseURL,
		Auth:          authConfig,
	}

	if req.Async {
		// Handle async request
		s.handleAsyncAnalysis(w, config)
	} else {
		// Handle synchronous request
		s.handleSyncAnalysis(w, config)
	}
}

// @Summary Get analysis job status
// @Description Get the status and results of an analysis job
// @Tags analysis
// @Produce json
// @Param jobId path string true "Job ID"
// @Success 200 {object} JobStatusResponse
// @Failure 404 {object} ErrorResponse "Job not found"
// @Router /analyze/status/{jobId} [get]
func (s *Server) getJobStatusHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	jobID := vars["jobId"]

	job, exists := s.jobManager.GetJob(jobID)
	if !exists {
		s.writeError(w, http.StatusNotFound, "Job not found", fmt.Errorf("job %s not found", jobID))
		return
	}

	response := JobStatusResponse{
		JobID:     job.ID,
		Status:    string(job.Status),
		Progress:  job.Progress,
		Timestamp: time.Now().UTC(),
	}

	if job.Started != nil {
		elapsed := time.Since(*job.Started)
		response.Duration = elapsed.String()
	}

	if job.Status == JobStatusCompleted {
		response.Results = ConvertPageResults(job.Results)
		summary := s.createSummary(job.Results)
		response.Summary = &summary
	}

	if job.Error != nil {
		response.Error = job.Error.Error()
	}

	WriteJSON(w, http.StatusOK, response)
}

// @Summary List analysis jobs
// @Description Get a list of analysis jobs with optional filtering
// @Tags analysis
// @Produce json
// @Param status query string false "Filter by status" Enums(pending,running,completed,failed,canceled)
// @Param limit query int false "Limit number of results" default(50)
// @Param offset query int false "Offset for pagination" default(0)
// @Success 200 {object} JobListResponse
// @Router /analyze/jobs [get]
func (s *Server) listJobsHandler(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	statusFilter := JobStatus(r.URL.Query().Get("status"))

	limit := 50
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}

	offset := 0
	if o := r.URL.Query().Get("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	// Get jobs from manager
	jobs := s.jobManager.ListJobs(statusFilter, limit, offset)

	// Convert to summaries
	summaries := make([]JobSummary, len(jobs))
	for i, job := range jobs {
		summaries[i] = JobSummary{
			JobID:     job.ID,
			Status:    string(job.Status),
			Created:   job.Created,
			Completed: job.Completed,
			Sitemap:   job.Config.Sitemap,
			Progress:  job.Progress,
		}
	}

	response := JobListResponse{
		Jobs:  summaries,
		Total: len(summaries),
		Page:  (offset / limit) + 1,
		Limit: limit,
	}

	WriteJSON(w, http.StatusOK, response)
}

// @Summary Cancel analysis job
// @Description Cancel a running or pending analysis job
// @Tags analysis
// @Produce json
// @Param jobId path string true "Job ID"
// @Success 200 {object} JobStatusResponse "Job canceled successfully"
// @Failure 400 {object} ErrorResponse "Job cannot be canceled"
// @Failure 404 {object} ErrorResponse "Job not found"
// @Router /analyze/jobs/{jobId}/cancel [post]
func (s *Server) cancelJobHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	jobID := vars["jobId"]

	if err := s.jobManager.CancelJob(jobID); err != nil {
		if err.Error() == fmt.Sprintf("job not found: %s", jobID) {
			s.writeError(w, http.StatusNotFound, "Job not found", err)
		} else {
			s.writeError(w, http.StatusBadRequest, "Cannot cancel job", err)
		}
		return
	}

	// Return updated job status
	job, _ := s.jobManager.GetJob(jobID)
	response := JobStatusResponse{
		JobID:     job.ID,
		Status:    string(job.Status),
		Progress:  job.Progress,
		Timestamp: time.Now().UTC(),
	}

	WriteJSON(w, http.StatusOK, response)
}

// Root handler provides basic API information
func (s *Server) rootHandler(w http.ResponseWriter, r *http.Request) {
	info := map[string]any{
		"name":        "PSI-Map API",
		"description": "API for analyzing website performance using PageSpeed Insights",
		"version":     s.version,
		"docs":        "/swagger/",
		"endpoints": map[string]string{
			"health":  "/api/v1/health",
			"version": "/api/v1/version",
			"analyze": "/api/v1/analyze",
			"jobs":    "/api/v1/analyze/jobs",
		},
	}

	WriteJSON(w, http.StatusOK, info)
}

// handleAsyncAnalysis handles asynchronous analysis requests
func (s *Server) handleAsyncAnalysis(w http.ResponseWriter, config *types.AnalysisConfig) {
	log := logger.GetLogger()

	// Create job
	job, err := s.jobManager.CreateJob(config)
	if err != nil {
		s.writeError(w, http.StatusServiceUnavailable, "Cannot create job", err)
		return
	}

	// Start job
	if err := s.jobManager.StartJob(job.ID); err != nil {
		s.writeError(w, http.StatusInternalServerError, "Cannot start job", err)
		return
	}

	log.Info("Started async analysis job: %s", job.ID)

	response := AsyncAnalysisResponse{
		JobID:     job.ID,
		Status:    "started",
		Timestamp: time.Now().UTC(),
		StatusURL: fmt.Sprintf("/api/v1/analyze/status/%s", job.ID),
	}

	WriteJSON(w, http.StatusAccepted, response)
}

// handleSyncAnalysis handles synchronous analysis requests
func (s *Server) handleSyncAnalysis(w http.ResponseWriter, config *types.AnalysisConfig) {
	log := logger.GetLogger()
	start := time.Now()

	// Execute analysis (reusing CLI logic)
	results, err := s.executeAnalysis(config)
	if err != nil {
		log.Error("Synchronous analysis failed: %v", err)
		s.writeError(w, http.StatusInternalServerError, "Analysis failed", err)
		return
	}

	elapsed := time.Since(start)

	// Create response
	response := AnalysisResponse{
		Status:    "completed",
		Timestamp: time.Now().UTC(),
		Results:   ConvertPageResults(results),
		Summary:   s.createSummary(results),
		Duration:  elapsed.String(),
	}

	WriteJSON(w, http.StatusOK, response)
}

// executeAnalysis reuses the CLI analysis logic
func (s *Server) executeAnalysis(config *types.AnalysisConfig) ([]*types.PageResult, error) {
	log := logger.GetLogger()

	// Parse input to get URLs
	urls, err := utils.ParseSitemap(config.Sitemap)
	if err != nil {
		return nil, fmt.Errorf("failed to parse sitemap: %w", err)
	}
	log.Info("Found %d URLs to analyze", len(urls))

	// Initialize cache store
	cacheDir, err := utils.GetCacheDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get cache directory: %w", err)
	}
	store, err := cache.NewFilesystemCacheStore(cacheDir)
	if err != nil {
		return nil, fmt.Errorf("failed to create cache store: %w", err)
	}
	defer store.Close()

	// Load or create cache index
	hash, err := utils.CalculateSitemapHash(config.Sitemap, urls)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate sitemap hash: %w", err)
	}
	index, idxFilename, err := cache.LoadOrCreateIndex(store, hash, urls, config.Sitemap)
	if err != nil {
		return nil, fmt.Errorf("failed to load or create cache index: %w", err)
	}

	// Check URL-level cache
	ttl := time.Duration(config.CacheTTL) * time.Hour
	cachedResults, missingURLs, err := cache.CheckURLCache(store, index, ttl)
	if err != nil {
		log.Warn("Cache check failed: %v", err)
		missingURLs = urls
		cachedResults = nil
	}

	// Report cache status
	cachedCount := len(cachedResults)
	missingCount := len(missingURLs)
	if cachedCount > 0 {
		log.Tagged("CACHE", "Found %d cached result(s), %d URL(s) need analysis", "🎯", cachedCount, missingCount)
	}

	var newResults []*types.PageResult
	// Only analyze missing URLs
	if missingCount > 0 {
		log.Tagged("ANALYZE", "Starting analysis of %d URL(s)...", "🔍", missingCount)
		newResults = runner.RunBatch(missingURLs, config)

		// Save new results to cache
		if _, err := cache.SaveResults(store, index, idxFilename, newResults); err != nil {
			log.Error("Failed to save cache: %v", err)
		} else {
			log.Tagged("CACHE", "%d new result(s) cached successfully", "💾", len(newResults))
		}
	}

	// Combine cached and new results
	return cache.CombineResultsInOrder(urls, cachedResults, newResults), nil
}

// createSummary generates a summary of results
func (s *Server) createSummary(results []*types.PageResult) AnalysisSummary {
	if len(results) == 0 {
		return AnalysisSummary{}
	}

	var totalPerformance, totalAccessibility, totalBestPractices, totalSEO float64
	var fastestURL, slowestURL string
	var fastestTime, slowestTime time.Duration
	validResults := 0
	failedResults := 0

	for i, result := range results {
		// Check for failures
		if result.Mobile == nil && result.Desktop == nil {
			failedResults++
			continue
		}

		// Get scores (prefer mobile, fallback to desktop)
		var scores *types.CategoryScores
		if result.Mobile != nil && result.Mobile.Scores != nil {
			scores = result.Mobile.Scores
		} else if result.Desktop != nil && result.Desktop.Scores != nil {
			scores = result.Desktop.Scores
		}

		if scores != nil {
			totalPerformance += scores.Performance
			totalAccessibility += scores.Accessibility
			totalBestPractices += scores.BestPractices
			totalSEO += scores.SEO
			validResults++
		}

		// Track fastest/slowest by duration
		if i == 0 || result.Duration < fastestTime {
			fastestTime = result.Duration
			fastestURL = result.URL
		}
		if i == 0 || result.Duration > slowestTime {
			slowestTime = result.Duration
			slowestURL = result.URL
		}
	}

	summary := AnalysisSummary{
		TotalURLs:    len(results),
		AnalyzedURLs: validResults,
		FailedURLs:   failedResults,
		FastestURL:   fastestURL,
		SlowestURL:   slowestURL,
	}

	if validResults > 0 {
		summary.AveragePerformance = totalPerformance / float64(validResults)
		summary.AverageAccessibility = totalAccessibility / float64(validResults)
		summary.AverageBestPractices = totalBestPractices / float64(validResults)
		summary.AverageSEO = totalSEO / float64(validResults)
	}

	return summary
}

// writeError writes an error response
func (s *Server) writeError(w http.ResponseWriter, code int, message string, err error) {
	log := logger.GetLogger()
	log.Error("%s: %v", message, err)

	response := ErrorResponse{
		Error:   message,
		Code:    code,
		Message: err.Error(),
		Time:    time.Now().UTC(),
	}

	WriteJSON(w, code, response)
}

// Middleware functions
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		log := logger.GetLogger()

		// Log request
		log.Tagged("HTTP", "%s %s", "🌐", r.Method, r.URL.Path)

		next.ServeHTTP(w, r)

		// Log response time
		elapsed := time.Since(start)
		log.Tagged("HTTP", "Completed in %v", "⚡", elapsed)
	})
}
