# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**seofor.dev** is a CLI-first SEO tool built in Go that helps indie hackers and developers audit their localhost applications and rank from day 1. The project includes:

- **Go CLI application** (`main.go`, `cmd/`, `internal/`) - Main SEO auditing tool with Terminal UI (TUI)
- **MkDocs documentation** (`docs/`, `mkdocs.yml`) - Project documentation website
- **Playwright integration** - Web crawling and browser automation for SEO analysis

⚠️ **Important**: This repository contains work-in-progress open-source code. The production code is in a private repository and is being transitioned here. Users should install from releases, not build from source.

## Key Commands

### Go Development
```bash
# Build the application
go build -o seo .

# Run the application directly
go run .

# Run with debug logging
SEO_DEBUG=1 go run .
DEBUG=1 go run .

# Test the application
go test ./...

# Format Go code
go fmt ./...

# Vet for potential issues
go vet ./...

# Clean module cache
go clean -modcache

# Update dependencies
go mod tidy
```

### Documentation Development
```bash
# Install Python dependencies
pip install -r requirements.txt

# Serve documentation locally (auto-reload)
mkdocs serve

# Build documentation
mkdocs build

# Deploy documentation (handled by GitHub Actions)
mkdocs gh-deploy
```

## Architecture

### Core Components

**Entry Point**
- `main.go` - Imports and calls `cmd.Execute()`
- `cmd/root.go` - Root Cobra command with TUI initialization

**TUI System** (`internal/tui/`)
- `config/config.go` - Configuration management, API key validation, user settings
- `logger/logger.go` - Debug logging system (activated with `SEO_DEBUG=1`)
- `version.go` - Version checking and update notifications

**Web Crawling** (`internal/crawler/`)
- `crawler.go` - Multi-threaded Playwright-based web crawler with robots.txt support
- Configurable concurrency, depth limits, ignore patterns
- Respects robots.txt rules for ethical crawling

**Playwright Integration** (`internal/playwright/`)
- `setup.go` - Playwright installation and browser management
- Handles Chromium browser setup in `~/.seo/playwright/`
- Version-aware installation with automatic cleanup

**API Integration** (`internal/api/`)
- Server communication for SEO analysis
- API key authentication with seofor.dev service

### Configuration

User configuration stored in `~/.seo/config.yml`:
- API key for seofor.dev service
- Default crawling settings (port, concurrency, max pages/depth)
- Ignore patterns for URLs

Environment variables:
- `SEO_DEBUG=1` or `DEBUG=1` - Enable debug logging
- `SEO_BASE_URL` - Override API base URL (development only)
- `TEST_PLAYWRIGHT_DIR` - Override Playwright directory for testing

### Debugging

The application includes comprehensive debug logging. Enable with:
```bash
SEO_DEBUG=1 ./seo
```

Logs are written to `~/.seo/debug.log` when debug mode is active.

### Dependencies

Key external dependencies:
- **Cobra** - CLI framework and command structure
- **Bubble Tea** - Terminal UI framework for interactive interface
- **Playwright Go** - Browser automation for web crawling
- **gopkg.in/yaml.v3** - YAML parsing for configuration files

## Development Workflow

1. **Building**: Use `go build -o seo .` to create executable
2. **Testing**: Use `go test ./...` for unit tests
3. **Formatting**: Always run `go fmt ./...` before committing
4. **Documentation**: Use `mkdocs serve` for local docs development
5. **Dependencies**: Run `go mod tidy` when adding/removing dependencies

## Installation Flow

Users install via:
```bash
curl -sSfL https://seofor.dev/install.sh | bash
```

First run triggers Playwright setup (~150MB download) with user-visible progress.

## TUI Application Flow

1. **Startup** - Check Playwright installation, load config, validate API key
2. **API Key Gatekeeper** - Prompt for API key if not set/invalid
3. **Main Menu** - Navigate between audit tools, settings, keyword research
4. **Configuration** - Manage crawling settings, API key, app preferences
5. **Audit Results** - Display SEO findings with export options

## Important Notes

- The CLI binary name is `seo` (not `oss_seo` as in code)
- Playwright browsers are installed to `~/.seo/playwright/browsers/`
- Configuration and logs stored in `~/.seo/` directory
- Application requires Unix environment (WSL, macOS, Linux) with root access
- API key required for full functionality (free tier available)