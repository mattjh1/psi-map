package psi

// LighthouseResult contains the Lighthouse audit results
type LighthouseResult struct {
	RequestedURL       string                    `json:"requestedUrl,omitempty"`
	FinalDisplayedURL  string                    `json:"finalDisplayedUrl,omitempty"`
	MainDocumentURL    string                    `json:"mainDocumentUrl,omitempty"`
	FinalURL           string                    `json:"finalUrl,omitempty"`
	LighthouseVersion  string                    `json:"lighthouseVersion,omitempty"`
	UserAgent          string                    `json:"userAgent,omitempty"`
	FetchTime          string                    `json:"fetchTime,omitempty"`
	Environment        *Environment              `json:"environment,omitempty"`
	RunWarnings        []string                  `json:"runWarnings,omitempty"`
	ConfigSettings     *ConfigSettings           `json:"configSettings,omitempty"`
	Categories         *Categories               `json:"categories,omitempty"`
	CategoryGroups     map[string]*CategoryGroup `json:"categoryGroups,omitempty"`
	Audits             map[string]*Audit         `json:"audits,omitempty"`
	Timing             *Timing                   `json:"timing,omitempty"`
	I18n               *I18n                     `json:"i18n,omitempty"`
	Entities           []Entity                  `json:"entities,omitempty"`
	FullPageScreenshot *FullPageScreenshot       `json:"fullPageScreenshot,omitempty"`
}

// Categories contains all Lighthouse category results
type Categories struct {
	Performance   *Category `json:"performance,omitempty"`
	Accessibility *Category `json:"accessibility,omitempty"`
	BestPractices *Category `json:"best-practices,omitempty"`
	SEO           *Category `json:"seo,omitempty"`
	PWA           *Category `json:"pwa,omitempty"`
}

// Category represents a Lighthouse category (Performance, Accessibility, etc.)
type Category struct {
	ID                string     `json:"id,omitempty"`
	Title             string     `json:"title,omitempty"`
	Description       string     `json:"description,omitempty"`
	Score             *float64   `json:"score,omitempty"`
	ManualDescription string     `json:"manualDescription,omitempty"`
	AuditRefs         []AuditRef `json:"auditRefs,omitempty"`
}

// AuditRef represents a reference to an audit within a category
type AuditRef struct {
	ID             string   `json:"id"`
	Weight         float64  `json:"weight,omitempty"`
	Group          string   `json:"group,omitempty"`
	Acronym        string   `json:"acronym,omitempty"`
	RelevantAudits []string `json:"relevantAudits,omitempty"`
}

// Audit represents a single Lighthouse audit
type Audit struct {
	ID               string         `json:"id,omitempty"`
	Title            string         `json:"title,omitempty"`
	Description      string         `json:"description,omitempty"`
	Score            *float64       `json:"score,omitempty"`
	ScoreDisplayMode string         `json:"scoreDisplayMode,omitempty"`
	NumericValue     *float64       `json:"numericValue,omitempty"`
	NumericUnit      string         `json:"numericUnit,omitempty"`
	DisplayValue     string         `json:"displayValue,omitempty"`
	Explanation      string         `json:"explanation,omitempty"`
	ErrorMessage     string         `json:"errorMessage,omitempty"`
	Warnings         []string       `json:"warnings,omitempty"`
	Details          map[string]any `json:"details,omitempty"`
	MetricSavings    *MetricSavings `json:"metricSavings,omitempty"`
}

// MetricSavings represents potential savings from fixing an audit
type MetricSavings struct {
	FCP *float64 `json:"FCP,omitempty"`
	LCP *float64 `json:"LCP,omitempty"`
}

// Environment contains information about the test environment
type Environment struct {
	NetworkUserAgent string  `json:"networkUserAgent,omitempty"`
	HostUserAgent    string  `json:"hostUserAgent,omitempty"`
	BenchmarkIndex   float64 `json:"benchmarkIndex,omitempty"`
}

// ConfigSettings contains Lighthouse configuration
type ConfigSettings struct {
	EmulatedFormFactor string   `json:"emulatedFormFactor,omitempty"`
	FormFactor         string   `json:"formFactor,omitempty"`
	Locale             string   `json:"locale,omitempty"`
	OnlyCategories     []string `json:"onlyCategories,omitempty"`
	Channel            string   `json:"channel,omitempty"`
}

// CategoryGroup represents a group of related audits
type CategoryGroup struct {
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
}

// Timing contains performance timing information
type Timing struct {
	Total float64 `json:"total,omitempty"`
}

// I18n contains internationalization data
type I18n struct {
	RendererFormattedStrings map[string]string `json:"rendererFormattedStrings,omitempty"`
}

// Entity represents a web entity (like a company or organization)
type Entity struct {
	Name         string   `json:"name,omitempty"`
	Origins      []string `json:"origins,omitempty"`
	Homepage     string   `json:"homepage,omitempty"`
	Categories   []string `json:"categories,omitempty"`
	IsFirstParty bool     `json:"isFirstParty,omitempty"`
}

// FullPageScreenshot contains screenshot data
type FullPageScreenshot struct {
	Screenshot *Screenshot      `json:"screenshot,omitempty"`
	Nodes      map[string]*Node `json:"nodes,omitempty"`
}

// Screenshot represents screenshot data
type Screenshot struct {
	Data   string `json:"data,omitempty"`
	Width  int    `json:"width,omitempty"`
	Height int    `json:"height,omitempty"`
}

// Node represents a DOM node in the screenshot
type Node struct {
	Left   float64 `json:"left,omitempty"`
	Top    float64 `json:"top,omitempty"`
	Width  float64 `json:"width,omitempty"`
	Height float64 `json:"height,omitempty"`
}
