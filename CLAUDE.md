# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

seofor.dev is a Go-based CLI tool for SEO analysis and optimization, featuring:
- Interactive TUI (Terminal User Interface) built with Bubble Tea
- Website crawling and SEO auditing capabilities using Playwright
- Local-only analysis (no external API required)
- IndexNow integration for search engine notifications
- MkDocs-based documentation system

## Build and Development Commands

### Go Commands
```bash
# Build the application
go build -o seo .

# Run the application
go run .

# Install dependencies
go mod tidy

# Run tests (if any exist)
go test ./...
```

### Documentation
```bash
# Serve documentation locally
mkdocs serve

# Build documentation
mkdocs build

# Install Python dependencies for docs
pip install -r requirements.txt
```

## Architecture Overview

### Entry Point
- `main.go`: Simple entry point that calls `cmd.Execute()`
- `cmd/root.go`: Cobra-based CLI setup with main application logic

### Core Packages

#### Internal Structure
- `internal/audit/`: Local audit processing and storage
- `internal/crawler/`: Playwright-based web crawler for site discovery
- `internal/playwright/`: Playwright setup and browser management
- `internal/config/`: Configuration management
- `internal/export/`: Export formatting for AI prompts
- `internal/services/`: Service layer for audit operations
- `internal/version/`: Version management and update checking

### Key Data Structures

#### Audit System
- `AuditSession`: Complete audit workflow state
- `PageResult`: Individual page analysis results
- `AuditConfig`: Audit parameters and settings
- Audit phases: SiteDiscovery → PageAnalysis → SessionCompletion

#### Crawler Architecture
- Concurrent crawling with configurable workers
- Robots.txt compliance
- Link discovery with depth limits
- Page normalization and deduplication

## Development Notes

### Environment Variables
- `SEO_DEBUG` or `DEBUG`: Enable debug logging

### Configuration
- Config stored in user's config directory (~/.seo/config.yml)
- Automatic update checking

### Dependencies
- Bubble Tea for TUI framework
- Cobra for CLI structure
- Playwright for web crawling
- YAML for configuration
- Clipboard integration for export features

### Browser Setup
- Playwright browsers installed to custom directory
- Headless Chromium for crawling
- Timeout handling for page loads

### Error Handling
- Graceful degradation for missing features
- User-friendly error messages in TUI
