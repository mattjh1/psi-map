# Roadmap

This document outlines the future direction of the `psi-map` project.

## ✅ Recently Completed

- **v0.3.1**: Improved logging and added more unit tests.
- **v0.3.0**: Enhanced CLI user experience and CI optimizations.
- **v0.2.0**: Added Docker support and interactive HTML report visualization.
- **v0.1.0**: Initial release with core analysis, caching, and JSON/stdout reporting.

## 🚧 In Progress

- **Test Coverage**: Increase unit and integration test coverage across the application to ensure stability and reliability.
- **Performance Optimizations**: Refactoring critical components to support performance-sensitive configuration (e.g., `hugeParam`, `rangeValCopy`). This work involves significant architectural changes and has temporarily invalidated existing tests, which are being updated alongside the refactor.

## 🗓️ Planned

### v0.4: Self-Contained Static Reports

- **Goal**: Ensure the static HTML report generated via the `analyze` command includes full interactivity and styling, matching the experience of the `serve` mode.
- **Chunks**:
  - [ ] **Embed Static Assets**: Use `go:embed` to include JavaScript and CSS files directly in the Go binary.
  - [ ] **Inline or Bundle Assets**: Update the HTML templates to inline scripts or reference embedded assets, ensuring no external dependencies are required.
  - [ ] **Unify Templates**: Refactor `report.html` to support both static generation and dynamic serving without feature disparity.
  - [ ] **Validate Offline Usage**: Test that the generated HTML report works fully offline with all interactive functionality.

### v0.5: Historical Performance Tracking

- **Goal**: Allow users to track PSI scores over time to identify trends and regressions.
- **Chunks**:
  - [ ] **New Command**: Introduce a `psi-map history <url>` command to display the performance history for a specific page.
  - [ ] **Report Integration**: Enhance the interactive HTML report with a "History" view, showing charts of how core metrics have changed over time.

### v0.6: Advanced URL Discovery

- **Goal**: Move beyond sitemaps by discovering URLs through crawling.
- **Chunks**:
  - [ ] **Crawler Integration**: Integrate a Go-based crawling library (e.g., `gocolly`).
  - [ ] **New Flag**: Add a `--crawl` flag to the `analyze` command to initiate a crawl from a base URL.
  - [ ] **Crawl Configuration**: Add flags to control crawl depth, concurrency, and respect for `robots.txt`.

## 💡 Ideas / Backlog

- **Comparative Analysis**:
  - A new command `psi-map compare <sitemap1> <sitemap2>` to generate a side-by-side performance comparison.
- **Additional Export Formats**:
  - Add support for `csv` output for easier analysis in spreadsheet software.
- **Alerting and Thresholds**:
  - Introduce flags to set performance thresholds (e.g., `--threshold-lcp 2.5`) that can fail a CI/CD pipeline.
- **Configuration File**:
  - Support a `.psi-map.yml` file to define default settings for projects.
- **Authenticated Page Analysis**:
  - Integrate a headless browser to allow analysis of pages behind a login.
