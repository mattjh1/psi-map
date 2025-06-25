package types

// AnalysisConfig holds the configuration for analysis
type AnalysisConfig struct {
	Sitemap      string
	OutputFile   string
	OutputFormat string
	UseStdout    bool
	StartServer  bool
	ServerPort   string
	MaxWorkers   int
	CacheTTL     int
}
