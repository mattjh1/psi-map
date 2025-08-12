package types

// AnalysisConfig contains all configuration needed for an analysis
type AnalysisConfig struct {
	Sitemap       string                `json:"sitemap" example:"https://example.com/sitemap.xml"`
	OutputFormat  string                `json:"output_format" example:"json"`
	UseStdout     bool                  `json:"use_stdout" example:"false"`
	StartServer   bool                  `json:"start_server" example:"false"`
	MaxWorkers    int                   `json:"max_workers" example:"5"`
	CacheTTL      int                   `json:"cache_ttl" example:"3600"`
	Provider      string                `json:"provider" enums:"psi,lighthouse" example:"psi"`
	LighthouseURL string                `json:"lighthouse_url" example:"https://lighthouse-api.example.com"`
	Auth          *AuthenticationConfig `json:"auth,omitempty"`
}

// AuthenticationConfig holds authentication configuration for external services
type AuthenticationConfig struct {
	// Bearer token for Authorization header
	BearerToken string `json:"bearer_token,omitempty" example:"your-bearer-token-here"`
	// Cloudflare Access configuration
	CloudflareAccess *CloudflareAccessConfig `json:"cloudflare_access,omitempty"`
}

// CloudflareAccessConfig holds Cloudflare Access authentication details
type CloudflareAccessConfig struct {
	ClientID     string `json:"client_id,omitempty" example:"$CF_ACCESS_CLIENT_ID"`
	ClientSecret string `json:"client_secret,omitempty" example:"$CF_ACCESS_CLIENT_SECRET"`
}
