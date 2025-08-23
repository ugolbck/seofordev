# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

SEOForDev is a Go-based CLI tool for SEO analysis and optimization, featuring:
- Interactive TUI (Terminal User Interface) built with Bubble Tea
- Website crawling and SEO auditing capabilities using Playwright
- API integration with seofor.dev backend for analysis
- Keyword generation and content brief creation features
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
- `internal/api/`: HTTP client for seofor.dev API integration
- `internal/crawler/`: Playwright-based web crawler for site discovery
- `internal/playwright/`: Playwright setup and browser management
- `internal/tui/`: Complete Bubble Tea TUI implementation
- `internal/version/`: Version management and update checking

#### TUI Components
The TUI is modular with separate files for different screens:
- `main_menu.go`: Primary navigation interface
- `*_menu.go`: Various feature menus (audit, keyword, content brief)
- `*_details.go`: Detail views for results
- `*_history.go`: Historical data views
- `api_key_gatekeeper.go`: API key validation flow
- `config.go`: Configuration management
- `types.go`: Core data structures and Bubble Tea messages

### Key Data Structures

#### Audit System
- `AuditSession`: Complete audit workflow state
- `PageResult`: Individual page analysis results  
- `AuditConfig`: Audit parameters and settings
- Audit phases: SessionCreation → SiteDiscovery → CreditCheck → PageAnalysis → SessionCompletion

#### API Integration
- `Client`: HTTP client with structured request/response types
- Credit-based system with usage tracking
- Support for audits, keyword generation, and content briefs

#### Crawler Architecture
- Concurrent crawling with configurable workers
- Robots.txt compliance
- Link discovery with depth limits
- Page normalization and deduplication

## Development Notes

### Environment Variables
- `SEO_DEBUG` or `DEBUG`: Enable debug logging
- `SEO_BASE_URL`: Override API base URL (development only)

### Configuration
- Config stored in user's config directory
- API key validation on startup
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
- Structured error types for API responses
- Graceful degradation for missing features
- User-friendly error messages in TUI