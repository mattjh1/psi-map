# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- File validation for improved security
- Roadmap documentation
- Go install instructions in README

### Changed
- Optimized core logic for performance-sensitive config with related test fixes
- Broke up listCacheFiles function for better maintainability
- Updated roadmap documentation
- Removed unsupported output format from README
- Configured GoReleaser for releases

### Fixed
- Duplicate entry in makefile

## [v0.3.1] - 2025-01-01

### Added
- More unit tests for improved code coverage

### Fixed
- Use stderr for logs to ensure consistent -o stdout behavior
- Resolved lint issues

### Changed
- Updated documentation badges

## [v0.3.0] - 2024-12-24

### Added
- CLI user experience improvements
- Dockerfile for containerization
- MIT license
- Centralized logging system across application
- Advanced per-URL caching system for better performance

### Fixed
- Error in makefile
- CI configuration issues

### Changed
- Updated README documentation
- Optimized CI pipeline
- Updated security CI configuration
- Updated action versions

### Removed
- Pre-commit lint hook

## [v0.2.0] - 2024-12-17

### Added
- PageSpeed Insights API integration with basic caching
- HTML report generation with template system  
- CLI foundation and module structure
- Detailed report pages with improved UI and logging
- Data filtering capabilities with comprehensive tests
- Comprehensive testing framework
- Linting and improved Makefile with help system
- README documentation and version handling
- CI/CD pipeline and development environment setup
- Husky pre-commit hooks

### Changed
- Modernized frontend by replacing Bootstrap with Tailwind CSS
- Modularized assets for better organization
- Updated CI configuration

### Fixed  
- Pre-commit hooks and CI configuration issues

---

**Full Changelog**: https://github.com/mattjh1/psi-map/releases
