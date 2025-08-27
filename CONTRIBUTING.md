# Contributing to seofor.dev

Thank you for your interest in contributing to seofor.dev! This guide will help you get started with contributing to our open source SEO tool.

## 🚀 Quick Start for Contributors

1. **Fork the repository** on GitHub
2. **Clone your fork** locally:
   ```bash
   git clone https://github.com/yourusername/seofordev.git
   cd seofordev
   ```
3. **Install dependencies**:
   ```bash
   go mod download
   ```
4. **Build and test**:
   ```bash
   go build -o seo .
   ./seo
   ```

## 🏗️ Development Setup

### Prerequisites
- **Go 1.24+** (check with `go version`)
- **Git**
- **Make** (optional, for build scripts)

### Environment Variables
For development, you can use these environment variables:
- `SEO_DEBUG=1` - Enable debug logging
- `SEO_BASE_URL=http://localhost:8000` - Override API base URL for testing

### Project Structure
```
├── cmd/                 # CLI entry point
├── internal/
│   ├── api/            # HTTP client for backend API
│   ├── crawler/        # Playwright-based web crawler
│   ├── playwright/     # Browser automation setup
│   ├── tui/           # Terminal UI components (Bubble Tea)
│   └── version/       # Version management
├── docs/              # Documentation
└── site/              # Generated documentation site
```

## 🎯 Areas for Contribution

### 🔓 **Open Source (Free Features)**
These areas welcome all contributions:
- **Website Auditing**: SEO analysis algorithms and checks
- **Crawler Improvements**: Better website discovery and crawling
- **UI/UX Enhancements**: Terminal interface improvements
- **Performance**: Optimization and efficiency improvements
- **Documentation**: Guides, examples, and API docs
- **Testing**: Unit tests, integration tests, and test coverage
- **Bug Fixes**: Any issues with free functionality

### 💎 **Premium Features** (Limited Contributions)
Premium features (keyword research, content briefs) are backend-protected but you can contribute to:
- **UI Components**: Frontend display of premium feature results
- **Error Handling**: Better user experience for premium feature errors
- **Documentation**: Usage guides for premium features

## 📋 Contribution Guidelines

### Code Style
- Follow standard Go conventions (`go fmt`, `go vet`)
- Use meaningful variable and function names
- Add comments for complex logic
- Keep functions focused and small

### Commit Guidelines
- Use conventional commits format:
  - `feat:` new features
  - `fix:` bug fixes
  - `docs:` documentation changes
  - `refactor:` code refactoring
  - `test:` adding tests
  - `chore:` maintenance tasks

Examples:
```
feat: add robots.txt parsing to crawler
fix: handle timeout errors in audit process
docs: update installation instructions
```

### Pull Request Process

1. **Create a feature branch** from `main`:
   ```bash
   git checkout -b feature/your-feature-name
   ```

2. **Make your changes** with clear, focused commits

3. **Test your changes**:
   ```bash
   go test ./...
   go build -o seo .
   ./seo  # Manual testing
   ```

4. **Update documentation** if needed (README, CLAUDE.md, code comments)

5. **Submit a pull request** with:
   - Clear title and description
   - Reference any related issues
   - Screenshots/demos for UI changes
   - Test results and performance impact

### Testing
- Write unit tests for new functionality
- Ensure existing tests pass: `go test ./...`
- Test with different operating systems if possible
- Manual testing with the TUI interface

## 🐛 Bug Reports

When reporting bugs, please include:
- **Environment**: OS, Go version, terminal type
- **Steps to reproduce**: Detailed reproduction steps
- **Expected vs actual behavior**
- **Logs**: Enable debug mode (`SEO_DEBUG=1`) and include relevant logs
- **Screenshots**: For UI issues

## 💡 Feature Requests

For new features:
- **Check existing issues** first to avoid duplicates
- **Describe the use case** and problem being solved
- **Propose implementation** if you have ideas
- **Consider scope**: Focus on features that benefit the open source community

## 🔒 Security

- **Do not commit secrets**: API keys, tokens, or credentials
- **Report security issues** privately to hey@seofor.dev
- **Follow secure coding practices**: Input validation, error handling
- **See [SECURITY.md](SECURITY.md)** for detailed security guidelines

## 🚫 What NOT to Contribute

Please avoid contributions that:
- Bypass premium feature authentication
- Include hardcoded API keys or secrets
- Break existing functionality without good reason
- Add unnecessary dependencies
- Violate the project's MIT license

## 📞 Getting Help

- **GitHub Issues**: For bugs and feature requests
- **GitHub Discussions**: For questions and ideas
- **Email**: hey@seofor.dev for private matters
- **Twitter**: [@ugo_builds](https://x.com/ugo_builds)

## 📄 License

By contributing to seofor.dev, you agree that your contributions will be licensed under the same MIT License that covers the project. See [LICENSE](LICENSE) for details.

---

Thank you for contributing to seofor.dev! 🚀