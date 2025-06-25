package constants

import "time"

// Valid output formats.
const (
	HTML   = "html"
	JSON   = "json"
	STDOUT = "stdout"
)

// Runner constants
const (
	WaitGroupWorkers = 2
)

// CLI App constants
const (
	CPUDivisor      = 2
	DefaultTTLHours = 24
)

// CLI Cache constants
const (
	SeparatorLength   = 90
	MaxSitemapDisplay = 45
)

// Performance Metrics Thresholds - Core Web Vitals
const (
	// First Contentful Paint thresholds (milliseconds)
	FCPGoodThreshold = 1800
	FCPPoorThreshold = 3000

	// Largest Contentful Paint thresholds (milliseconds)
	LCPGoodThreshold = 2500
	LCPPoorThreshold = 4000

	// Cumulative Layout Shift thresholds
	CLSGoodThreshold = 0.1
	CLSPoorThreshold = 0.25

	// First Input Delay thresholds (milliseconds)
	FIDGoodThreshold = 100
	FIDPoorThreshold = 300
)

// Core Web Vitals Grades
const (
	GradeGood             = "good"
	GradeNeedsImprovement = "needs-improvement"
	GradePoor             = "poor"
)

// Performance Score Thresholds
const (
	ScoreGoodThreshold = 90
	ScorePoorThreshold = 50
)

// Server Configuration
const (
	// HTTP Server Timeouts
	ReadHeaderTimeout = 30 * time.Second
	ReadTimeout       = 60 * time.Second
	WriteTimeout      = 60 * time.Second
	IdleTimeout       = 120 * time.Second

	// Context Timeout
	ShutdownTimeout = 5 * time.Second

	// Map allocation optimization
	MapSizeDivisor = 2
)

// File System Permissions
const (
	DefaultDirPermissions = 0o755
)

// Time Calculations
const (
	Day24H = 24
)

// Audit Score Thresholds (Lighthouse)
const (
	AuditScorePoorThreshold = 0.5
	AuditScoreGoodThreshold = 0.9
	ScoreMultiplier         = 100
)
