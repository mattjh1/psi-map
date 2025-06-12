package server

import (
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"time"

	"github.com/mattjh1/psi-map/internal/constants"
	"github.com/mattjh1/psi-map/internal/types"
)

//go:embed templates/*.html templates/partials/*.html
var templateFS embed.FS

// loadReportTemplateFromFS loads and parses the report template from embedded filesystem
func loadReportTemplateFromFS() (*template.Template, error) {
	tmpl, err := template.New("report.html").Funcs(template.FuncMap{
		"formatDuration": formatDuration,
		"formatScore":    formatScore,
		"getGradeClass":  getGradeClass,
		"getScoreClass":  getScoreClass,
		"hasMetrics":     hasMetrics,
		"formatBytes":    formatBytes,
		"toSafeJSON":     toSafeJSON,
		"add":            func(a, b int) int { return a + b },
		"mul":            func(a, b int) int { return a * b },
		"dict":           dict,
		"getResult":      getResult,
	}).ParseFS(templateFS, "templates/report.html", "templates/layout.html", "templates/partials/*.html")
	if err != nil {
		return nil, fmt.Errorf("failed to parse templates: %v", err)
	}
	return tmpl, nil
}

// Template utility functions

func formatDuration(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%.0fms", float64(d)/float64(time.Millisecond))
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
	case score >= constants.ScoreGoodThreshold:
		return "text-success"
	case score >= constants.ScorePoorThreshold:
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

func toSafeJSON(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		return "null"
	}
	return string(b)
}

func dict(values ...any) map[string]any {
	if len(values)%2 != 0 {
		panic("dict requires an even number of arguments")
	}
	dict := make(map[string]any, len(values)/constants.MapSizeDivisor)
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
