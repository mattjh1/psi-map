package api

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/mattjh1/psi-map/internal/cache"
	"github.com/mattjh1/psi-map/internal/logger"
	"github.com/mattjh1/psi-map/internal/types"
	"github.com/mattjh1/psi-map/internal/utils"
	"github.com/mattjh1/psi-map/runner"
)

// JobStatus represents the status of an analysis job
type JobStatus string

const (
	JobStatusPending   JobStatus = "pending"
	JobStatusRunning   JobStatus = "running"
	JobStatusCompleted JobStatus = "completed"
	JobStatusFailed    JobStatus = "failed"
	JobStatusCanceled  JobStatus = "canceled"
)

// Job represents an analysis job
type Job struct {
	ID        string                `json:"job_id"`
	Status    JobStatus             `json:"status"`
	Config    *types.AnalysisConfig `json:"config"`
	Results   []*types.PageResult   `json:"results"`
	Progress  JobProgress           `json:"progress"`
	Error     error                 `json:"error,omitempty"`
	Created   time.Time             `json:"created"`
	Started   *time.Time            `json:"started,omitempty"`
	Completed *time.Time            `json:"completed,omitempty"`
	Context   context.Context       `json:"-"`
	Cancel    context.CancelFunc    `json:"-"`
}

// JobManager manages analysis jobs
type JobManager struct {
	jobs       map[string]*Job
	mutex      sync.RWMutex
	maxJobs    int
	cleanupTTL time.Duration
}

// NewJobManager creates a new job manager
func NewJobManager(maxJobs int, cleanupTTL time.Duration) *JobManager {
	jm := &JobManager{
		jobs:       make(map[string]*Job),
		maxJobs:    maxJobs,
		cleanupTTL: cleanupTTL,
	}

	// Start cleanup routine
	go jm.cleanupRoutine()

	return jm
}

// CreateJob creates a new analysis job
func (jm *JobManager) CreateJob(config *types.AnalysisConfig) (*Job, error) {
	jm.mutex.Lock()
	defer jm.mutex.Unlock()

	// Check if we've reached max jobs
	if len(jm.jobs) >= jm.maxJobs {
		return nil, fmt.Errorf("maximum number of concurrent jobs reached (%d)", jm.maxJobs)
	}

	// Create new job
	jobID := uuid.New().String()
	ctx, cancel := context.WithCancel(context.Background())

	job := &Job{
		ID:      jobID,
		Status:  JobStatusPending,
		Config:  config,
		Created: time.Now(),
		Context: ctx,
		Cancel:  cancel,
	}

	jm.jobs[jobID] = job
	return job, nil
}

// GetJob retrieves a job by ID
func (jm *JobManager) GetJob(jobID string) (*Job, bool) {
	jm.mutex.RLock()
	defer jm.mutex.RUnlock()

	job, exists := jm.jobs[jobID]
	return job, exists
}

// ListJobs returns all jobs (with optional filtering)
func (jm *JobManager) ListJobs(status JobStatus, limit, offset int) []*Job {
	jm.mutex.RLock()
	defer jm.mutex.RUnlock()

	var filteredJobs []*Job
	for _, job := range jm.jobs {
		if status == "" || job.Status == status {
			filteredJobs = append(filteredJobs, job)
		}
	}

	// Apply pagination
	start := offset
	if start >= len(filteredJobs) {
		return []*Job{}
	}

	end := start + limit
	if end > len(filteredJobs) {
		end = len(filteredJobs)
	}

	return filteredJobs[start:end]
}

// StartJob starts executing an analysis job
func (jm *JobManager) StartJob(jobID string) error {
	jm.mutex.Lock()
	job, exists := jm.jobs[jobID]
	if !exists {
		jm.mutex.Unlock()
		return fmt.Errorf("job not found: %s", jobID)
	}

	if job.Status != JobStatusPending {
		jm.mutex.Unlock()
		return fmt.Errorf("job %s is not in pending status", jobID)
	}

	job.Status = JobStatusRunning
	now := time.Now()
	job.Started = &now
	jm.mutex.Unlock()

	// Start job in goroutine
	go jm.executeJob(job)

	return nil
}

// CancelJob cancels a running job
func (jm *JobManager) CancelJob(jobID string) error {
	jm.mutex.Lock()
	defer jm.mutex.Unlock()

	job, exists := jm.jobs[jobID]
	if !exists {
		return fmt.Errorf("job not found: %s", jobID)
	}

	if job.Status != JobStatusRunning && job.Status != JobStatusPending {
		return fmt.Errorf("job %s cannot be canceled (status: %s)", jobID, job.Status)
	}

	job.Status = JobStatusCanceled
	job.Cancel()

	return nil
}

// executeJob executes the analysis job
func (jm *JobManager) executeJob(job *Job) {
	log := logger.GetLogger()
	defer func() {
		if r := recover(); r != nil {
			log.Error("Job %s panicked: %v", job.ID, r)
			jm.markJobFailed(job, fmt.Errorf("job panicked: %v", r))
		}
	}()

	log.Info("Starting job %s", job.ID)

	// Parse sitemap to get URLs
	urls, err := utils.ParseSitemap(job.Config.Sitemap)
	if err != nil {
		jm.markJobFailed(job, fmt.Errorf("failed to parse sitemap: %w", err))
		return
	}

	// Update progress
	jm.updateJobProgress(job, 0, len(urls))

	// Initialize cache store
	cacheDir, err := utils.GetCacheDir()
	if err != nil {
		jm.markJobFailed(job, fmt.Errorf("failed to get cache directory: %w", err))
		return
	}
	store, err := cache.NewFilesystemCacheStore(cacheDir)
	if err != nil {
		jm.markJobFailed(job, fmt.Errorf("failed to create cache store: %w", err))
		return
	}
	defer store.Close()

	// Load or create cache index
	hash, err := utils.CalculateSitemapHash(job.Config.Sitemap, urls)
	if err != nil {
		jm.markJobFailed(job, fmt.Errorf("failed to calculate sitemap hash: %w", err))
		return
	}
	index, idxFilename, err := cache.LoadOrCreateIndex(store, hash, urls, job.Config.Sitemap)
	if err != nil {
		jm.markJobFailed(job, fmt.Errorf("failed to load or create cache index: %w", err))
		return
	}

	// Check cache
	ttl := time.Duration(job.Config.CacheTTL) * time.Hour
	cachedResults, missingURLs, err := cache.CheckURLCache(store, index, ttl)
	if err != nil {
		log.Warn("Cache check failed for job %s: %v", job.ID, err)
		missingURLs = urls
		cachedResults = nil
	}

	var newResults []*types.PageResult

	// Analyze missing URLs if any
	if len(missingURLs) > 0 {
		log.Info("Job %s analyzing %d URLs", job.ID, len(missingURLs))

		// Create a progress tracking version of the runner
		newResults = jm.runBatchWithProgress(job, missingURLs)

		// Check if job was canceled
		if job.Status == JobStatusCanceled {
			log.Info("Job %s was canceled", job.ID)
			return
		}

		// Save to cache
		if _, err := cache.SaveResults(store, index, idxFilename, newResults); err != nil {
			log.Error("Failed to save cache for job %s: %v", job.ID, err)
		}
	}

	// Combine results
	allResults := cache.CombineResultsInOrder(urls, cachedResults, newResults)

	// Mark job as completed
	jm.markJobCompleted(job, allResults)
	log.Info("Job %s completed successfully", job.ID)
}

// runBatchWithProgress runs the batch analysis with progress tracking
func (jm *JobManager) runBatchWithProgress(job *Job, urls []string) []*types.PageResult {
	// We'll reuse the existing runner but need to track progress
	// For now, we'll use the existing runner and update progress periodically
	// In a production system, you might want to modify the runner to accept progress callbacks

	results := runner.RunBatch(urls, job.Config)

	// Update final progress
	jm.updateJobProgress(job, len(urls), len(urls))

	return results
}

// updateJobProgress updates the progress of a job
func (jm *JobManager) updateJobProgress(job *Job, current, total int) {
	jm.mutex.Lock()
	defer jm.mutex.Unlock()

	job.Progress.Current = current
	job.Progress.Total = total
	if total > 0 {
		job.Progress.Percent = int((float64(current) / float64(total)) * 100)
	}
}

// markJobCompleted marks a job as completed
func (jm *JobManager) markJobCompleted(job *Job, results []*types.PageResult) {
	jm.mutex.Lock()
	defer jm.mutex.Unlock()

	job.Status = JobStatusCompleted
	job.Results = results
	now := time.Now()
	job.Completed = &now
}

// markJobFailed marks a job as failed
func (jm *JobManager) markJobFailed(job *Job, err error) {
	jm.mutex.Lock()
	defer jm.mutex.Unlock()

	job.Status = JobStatusFailed
	job.Error = err
	now := time.Now()
	job.Completed = &now
}

// cleanupRoutine periodically cleans up old completed/failed jobs
func (jm *JobManager) cleanupRoutine() {
	ticker := time.NewTicker(time.Hour) // Run every hour
	defer ticker.Stop()

	for range ticker.C {
		jm.cleanup()
	}
}

// cleanup removes old jobs
func (jm *JobManager) cleanup() {
	jm.mutex.Lock()
	defer jm.mutex.Unlock()

	cutoff := time.Now().Add(-jm.cleanupTTL)

	for jobID, job := range jm.jobs {
		// Clean up completed/failed jobs older than TTL
		if (job.Status == JobStatusCompleted || job.Status == JobStatusFailed) &&
			job.Created.Before(cutoff) {
			delete(jm.jobs, jobID)
		}
	}
}
