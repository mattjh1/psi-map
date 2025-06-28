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
		"div":            func(a, b int) int { return a / b },
		"float64":        func(i int) float64 { return float64(i) },
		"printf":         fmt.Sprintf,
		"percentage": func(part, total int) string {
			if total == 0 {
				return "0.0%"
			}
			return fmt.Sprintf("%.1f%%", float64(part)/float64(total)*constants.ScoreMultiplier)
		},
		"successRate": func(successful, total int) string {
			if total == 0 {
				return "0.0%"
			}
			return fmt.Sprintf("%.1f%%", float64(successful)/float64(total)*constants.ScoreMultiplier)
		},
		"dict":      dict,
		"getResult": getResult,
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
		return "bg-green-300 text-green-800 border-green-400"
	case "needs-improvement":
		return "bg-yellow-300 text-yellow-800 border-yellow-400"
	case "poor":
		return "bg-red-300 text-red-800 border-red-400"
	default:
		return "bg-gray-300 text-gray-800 border-gray-400"
	}
}

func getScoreClass(score float64) string {
	switch {
	case score >= constants.ScoreGoodThreshold:
		return "text-green-700"
	case score >= constants.ScorePoorThreshold:
		return "text-yellow-700"
	default:
		return "text-red-700"
	}
}

func hasMetrics(result *types.Result) bool {
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

func getResult(page *types.PageResult, strategy string) *types.Result {
	if strategy == "mobile" {
		return page.Mobile
	}
	return page.Desktop
}
