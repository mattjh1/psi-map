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
docker run --rm -v $(pwd):/workspace ghcr.io/mattjh1/psi-map:latest sitemap.xml
```

## Usage

```bash
# Basic usage
psi-map sitemap.xml

# With custom output format
psi-map --format html --output report.html sitemap.xml

# Skip cache and force fresh analysis
psi-map --no-cache sitemap.xml

# List cached results
psi-map --list-cache

# Clear all cache
psi-map --clear-cache
```

## Configuration

| Flag | Description | Default |
|------|-------------|---------|
| `--workers` | Number of concurrent workers | 5 |
| `--format` | Output format (html, json, csv) | html |
| `--output` | Output file path | stdout |
| `--no-cache` | Skip cache, force fresh analysis | false |
| `--max-cache-age` | Maximum cache age | 24h |

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
