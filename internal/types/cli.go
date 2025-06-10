package types

// AnalysisConfig holds the configuration for analysis
type AnalysisConfig struct {
	Sitemap     string
	OutputHTML  string
	OutputJSON  string
	StartServer bool
	ServerPort  string
	MaxWorkers  int
	CacheTTL    int
}
