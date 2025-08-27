# Security Policy

## Supported Versions

We actively support the following versions of seofor.dev:

| Version | Supported          |
| ------- | ------------------ |
| Latest  | ✅ Yes            |
| < Latest| ❌ No             |

We recommend always using the latest version for security and feature updates.

## Reporting a Vulnerability

**Please do not report security vulnerabilities through public GitHub issues.**

### How to Report

1. **Email**: Send details to **hey@seofor.dev**
2. **Subject Line**: Include "SECURITY:" prefix
3. **Include**: 
   - Description of the vulnerability
   - Steps to reproduce
   - Potential impact
   - Any suggested fixes

### What to Expect

- **Initial Response**: Within 48 hours
- **Status Updates**: Every 7 days until resolution
- **Timeline**: We aim to resolve critical issues within 7 days
- **Credit**: Security researchers will be credited (unless they prefer anonymity)

## Security Features

### API Key Protection
- API keys are never logged or exposed in error messages
- Keys are masked in UI display (`MaskAPIKey` function)
- Keys excluded from JSON serialization (`json:"-"` tags)
- Server-side validation prevents unauthorized access to premium features

### Network Security
- HTTP clients use appropriate timeouts (30s for API calls)
- TLS/HTTPS enforced for all external communications
- User-Agent headers set for identification
- No hardcoded credentials in source code

### Input Validation
- URL parsing with comprehensive error handling
- Regex patterns properly escaped to prevent ReDoS
- File paths validated and normalized
- Input size limits to prevent resource exhaustion

### File System Security
- Directories created with appropriate permissions (0755)
- Log files created with secure permissions (0644)
- Playwright browser data isolated to user directory
- No sensitive data written to temporary files

## Security Best Practices for Contributors

### Code Review Checklist
- ✅ No hardcoded secrets or API keys
- ✅ Input validation for all user inputs
- ✅ Proper error handling without information disclosure
- ✅ Safe file operations with permission checks
- ✅ Timeout handling for network operations
- ✅ SQL injection prevention (if applicable)
- ✅ XSS prevention for any web interfaces

### Secure Development
```bash
# Before committing, check for secrets
git log --grep="api" --grep="key" --grep="secret" --grep="token" -i

# Scan dependencies for vulnerabilities (if using tools like nancy)
go list -json -deps ./... | nancy sleuth

# Run security linter (if using gosec)
gosec ./...
```

### Environment Variables
Use environment variables for sensitive configuration:
- `SEO_BASE_URL` - Override API base URL for development
- `SEO_DEBUG` - Enable debug logging (development only)

**Never commit `.env` files or hardcode production secrets.**

## Vulnerability Disclosure Policy

### Scope
Security vulnerabilities in:
- ✅ Core SEO auditing functionality  
- ✅ Web crawler and Playwright integration
- ✅ API client and authentication
- ✅ Terminal UI and user input handling
- ✅ File system operations
- ✅ Third-party dependencies

### Out of Scope
- ❌ Backend API vulnerabilities (report to seofor.dev directly)
- ❌ Social engineering attacks
- ❌ Physical security issues
- ❌ Denial of service attacks against public services
- ❌ Issues in third-party services we don't control

### Responsible Disclosure
We follow responsible disclosure practices:
1. **Report privately** to hey@seofor.dev
2. **Allow time** for us to investigate and fix
3. **Coordinate** public disclosure timing
4. **Avoid** accessing user data or disrupting services

## Security Updates

### Release Process
- Security fixes are released as patch versions immediately
- Critical vulnerabilities trigger emergency releases
- All security updates are documented in release notes
- Users are notified through GitHub releases and documentation

### Staying Updated
- ⭐ **Star the repository** for release notifications
- 📧 **Subscribe to releases** on GitHub  
- 🐦 **Follow [@ugo_builds](https://x.com/ugo_builds)** for updates
- 📚 **Check the documentation** at docs.seofor.dev

## Security Tools and Dependencies

### Dependency Management
- We regularly update dependencies for security patches
- Dependencies are vetted for known vulnerabilities
- Minimal dependency approach to reduce attack surface

### Recommended Security Tools
Contributors and users can enhance security with:
- `gosec` - Go security analyzer
- `nancy` - Dependency vulnerability scanner  
- `staticcheck` - Go static analysis
- `go mod audit` - Module vulnerability checking

## Contact

For security-related questions or concerns:
- **Email**: hey@seofor.dev
- **Subject**: Include "SECURITY:" prefix
- **Response Time**: Within 48 hours

---

**Security is a shared responsibility. Thank you for helping keep seofor.dev secure! 🔒**