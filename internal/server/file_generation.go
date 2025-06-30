package server

import (
	"fmt"
	"time"

	"github.com/mattjh1/psi-map/internal/types"
	"github.com/mattjh1/psi-map/internal/utils/validate"
)

// GenerateHTMLFile generates an HTML file from results without starting a server
func GenerateHTMLFile(results []*types.PageResult, filename string) error {
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
		Results   []*types.PageResult
		Summary   types.ReportSummary
		Generated time.Time
	}{
		Results:   results,
		Summary:   s.generateSummary(),
		Generated: time.Now(),
	}

	components := validate.SplitFilePath(filename)

	// Use secure file creation
	file, _, err := validate.SafeCreateFile(components.Dir, components.Name, components.Extension)
	if err != nil {
		return fmt.Errorf("failed to create file securely: %w", err)
	}
	defer file.Close()

	return tmpl.ExecuteTemplate(file, "report.html", data)
}
