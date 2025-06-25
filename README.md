# PSI-Map

![Go Version](https://img.shields.io/github/go-mod/go-version/mattjh1/psi-map)
![Build](https://github.com/mattjh1/psi-map/actions/workflows/ci.yml/badge.svg)
![Release](https://github.com/mattjh1/psi-map/actions/workflows/release.yml/badge.svg)
![Docker Build](https://github.com/mattjh1/psi-map/actions/workflows/docker.yml/badge.svg)
![Security Scan](https://github.com/mattjh1/psi-map/actions/workflows/security.yml/badge.svg)
![Codecov](https://codecov.io/gh/mattjh1/psi-map/branch/main/graph/badge.svg)
![License](https://img.shields.io/github/license/mattjh1/psi-map.svg)

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

```bash
# Start server with local sitemap file
psi-map serve --sitemap sitemap.xml

# Start server with remote sitemap URL
psi-map serve --sitemap https://example.com/sitemap.xml

# Or use the short flag
psi-map serve -s sitemap.xml

# Generate HTML report
psi-map serve --sitemap sitemap.xml --html report.html

# Generate JSON report  
psi-map serve -s https://example.com/sitemap.xml --json results.json

# Custom number of workers (default is half of available CPUs)
psi-map serve -s sitemap.xml --workers 10

# Set cache TTL (default 24 hours, 0 for no expiration)
psi-map serve -s sitemap.xml --cache-ttl 48
```

### Cache Management

```bash
# List cached results
psi-map cache list

# Clean expired cache files
psi-map cache clean

# Clear all cached results
psi-map cache clear
```

### Global Options

`--sitemap, -s`: Sitemap file path or URL (required)
``--html, -H``: Generate HTML report file
`--json, -j`: Generate JSON report file
`--workers, -w`: Maximum number of concurrent workers
`--cache-ttl`: Cache TTL in hours (0 = no expiration)

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
