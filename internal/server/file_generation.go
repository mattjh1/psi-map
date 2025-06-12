package server

import (
	"fmt"
	"os"
	"time"

	"github.com/mattjh1/psi-map/internal/types"
)

// GenerateHTMLFile generates an HTML file from results without starting a server
func GenerateHTMLFile(results []types.PageResult, filename string) error {
	tmpl, err := loadReportTemplateFromFS()
	if err != nil {
		return fmt.Errorf("template parsing error: %v", err)
	}

	// Verify the report template exists
	if tmpl.Lookup("report.html") == nil {
		return fmt.Errorf("report.html template not found")
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

	return tmpl.ExecuteTemplate(file, "report.html", data)
}
