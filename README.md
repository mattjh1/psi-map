# PSI-Map

[![Go Version](https://img.shields.io/github/go-mod/go-version/mattjh1/psi-map)](https://github.com/mattjh1/psi-map/blob/main/go.mod)
[![Go Report Card](https://goreportcard.com/badge/github.com/mattjh1/psi-map)](https://goreportcard.com/report/github.com/mattjh1/psi-map)
[![Build](https://github.com/mattjh1/psi-map/actions/workflows/ci.yml/badge.svg)](https://github.com/mattjh1/psi-map/actions/workflows/ci.yml)
[![Release](https://github.com/mattjh1/psi-map/actions/workflows/release.yml/badge.svg)](https://github.com/mattjh1/psi-map/actions/workflows/release.yml)
[![Docker Build](https://github.com/mattjh1/psi-map/actions/workflows/docker.yml/badge.svg)](https://github.com/mattjh1/psi-map/actions/workflows/docker.yml)
[![Docker Image](https://img.shields.io/badge/docker-ghcr.io-blue)](https://github.com/mattjh1/psi-map/pkgs/container/psi-map)
[![Security Scan](https://github.com/mattjh1/psi-map/actions/workflows/security.yml/badge.svg)](https://github.com/mattjh1/psi-map/security/code-scanning)
<!-- one day... -->
<!-- [![Codecov](https://codecov.io/gh/mattjh1/psi-map/branch/main/graph/badge.svg)](https://codecov.io/gh/mattjh1/psi-map) -->
[![License](https://img.shields.io/github/license/mattjh1/psi-map.svg)](https://github.com/mattjh1/psi-map/blob/main/LICENSE)

A command-line tool for batch PageSpeed Insights analysis using sitemap.xml files.

## Features

- Parse sitemap.xml files to extract URLs
- Concurrent PageSpeed Insights analysis
- Intelligent caching system
- Multiple output formats (HTML, JSON, CSV)
- Cross-platform support

## Installation

### From Release

Download the latest binary from the [releases page](https://github.com/mattjh1/psi-map/releases).

### From Source

```bash
git clone https://github.com/mattjh1/psi-map.git
cd psi-map
make install
```

### Using Docker

```bash
docker run --rm -v $(pwd):/workspace ghcr.io/mattjh1/psi-map:latest serve --sitemap sitemap.xml
```

## Usage

### Web Server

Start an interactive web server to view PageSpeed Insights results in your browser.

```bash
# Start server with local sitemap
psi-map server sitemap.xml

# Start server with remote sitemap
psi-map server https://example.com/sitemap.xml

# Custom port
psi-map server --port 3000 sitemap.xml
```

### Analyze Command

Analyze a sitemap and generate PageSpeed Insights reports in various formats.

```bash
# Basic JSON output (default)
psi-map analyze sitemap.xml

# Generate HTML report
psi-map analyze -o html sitemap.xml

# Custom output directory and filename
psi-map analyze -o json --output-dir ./reports --name my-report sitemap.xml

# Output to stdout
psi-map analyze -o stdout https://example.com/sitemap.xml
```


### Cache Management

Manage cached PageSpeed Insights results.

```bash
# List cached results
psi-map cache list

# Clean expired cache files
psi-map cache clean

# Clear all cached results
psi-map cache clear
```

### Command Aliases

- `analyze` = `run`
- `server` = `serve`

### Example Use Cases

- **Web Server**: `psi-map serve --port 3000 sitemap.xml`
- **Quick Analysis**: `psi-map analyze sitemap.xml`
- **HTML Report**: `psi-map analyze -o html --name site-performance sitemap.xml`
- **CI/CD Pipeline**: `psi-map analyze -o json --output-dir ./reports --name build-${BUILD_ID} sitemap.xml`
- **CI/CD Performance Gate**:

  ```bash
  psi-map analyze -o stdout sitemap.xml | \
  jq -e 'all(.[]; .Mobile.scores.performance >= 80 and .Desktop.scores.performance >= 80)' && \
  echo "✅ All pages meet performance threshold" || \
  (echo "❌ Performance check failed" && exit 1)
  ```


## Development

```bash
# Run tests
make test

# Build for current platform
make build

# Build for all platforms
make build-all

# Run linting
make lint

# Clean build artifacts
make clean
```

## License

MIT License - see LICENSE file for details.
