<div align="center">

```
                    ___               _
                   / __)             | |
  ___ _____  ___ _| |__ ___   ____ __| |_____ _   _
 /___) ___ |/ _ (_   __) _ \ / ___) _  | ___ | | | |
|___ | ____| |_| || | | |_| | |_ ( (_| | ____|\ V /
(___/|_____)\___/ |_|  \___/|_(_) \____|_____) \_/

```

# The Developer's SEO Toolkit

**Open-source CLI for localhost SEO audits**

[![MIT License](https://img.shields.io/badge/License-MIT-green.svg)](https://choosealicense.com/licenses/mit/)
[![Go Report Card](https://goreportcard.com/badge/github.com/ugolbck/seofordev)](https://goreportcard.com/report/github.com/ugolbck/seofordev)
[![GitHub release](https://img.shields.io/github/release/ugolbck/seofordev.svg)](https://github.com/ugolbck/seofordev/releases/latest)
[![GitHub stars](https://img.shields.io/github/stars/ugolbck/seofordev?style=social)](https://github.com/ugolbck/seofordev/stargazers)

**Stop losing traffic to SEO mistakes. Audit and optimize from your terminal.**

[Get Started](#-quick-start) • [Features](#-features) • [Docs](https://docs.seofor.dev)

</div>

---

## Demo

**Instant localhost SEO audits**
```bash
$ seo audit run
Crawling localhost:3000...
Found 12 pages, 0 errors
Generated optimization prompts - ready for AI!
```

## Features

- **Localhost SEO Audits** - Deep-dive analysis of your development sites
- **AI-Ready Exports** - Copy/paste prompts for Claude, Cursor, ChatGPT
- **Blazing Fast** - Built in Go, powered by Playwright
- **Smart Crawling** - JavaScript-rendered pages, respects robots.txt
- **IndexNow Integration** - Notify search engines about your changes
- **Zero Config** - Works out of the box, no setup required

## Quick Start

### 1. Install
```bash
# macOS/Linux
curl -sSfL https://seofor.dev/install.sh | bash
```

### 2. Audit Your Site
```bash
# Start your dev server (e.g., npm run dev, rails server, etc.)
$ cd your-project && npm run dev

# In another terminal, audit it
$ seo audit run --port 3000
# Get instant SEO insights + AI optimization prompts
```

## Commands Reference

```bash
seo audit run               # Audit localhost
seo audit list              # View audit history
seo audit show <id>         # Detailed results
seo config                  # Show settings
seo index submit <url>      # IndexNow submission
```

## Real-World Workflows

### Developer Workflow: Pre-Launch SEO Check
```bash
# Before deploying
seo audit run --port 3000
# Fix issues highlighted in AI prompts
# Deploy with confidence
```

### SEO Agency Workflow: Client Audits
```bash
# Quick client site analysis
seo audit run --port 8080
seo audit export <id>
# Professional SEO recommendations ready
```

## Why Choose seofor.dev?

| vs Manual SEO | vs Other Tools | vs Enterprise |
|:---:|:---:|:---:|
| **10x Faster** | **Developer-First** | **Free** |
| **AI-Integrated** | **Open Source** | **Zero Setup** |
| **Data-Driven** | **CLI-Native** | **Privacy-First** |

## For Contributors

```bash
# Get started with development
git clone https://github.com/ugolbck/seofordev.git
cd seofordev
go mod tidy
go build -o seo .

# Run tests
go test ./...

# See development docs
open CLAUDE.md
```

**Contributing**: We love contributions! Check out our [contribution guide](CONTRIBUTING.md) and help make SEO better for all developers.

## Built With

- **Go** - Performance & reliability
- **Cobra** - Beautiful CLI experience
- **Playwright** - Modern web crawling

## Roadmap

- [ ] **CI/CD integration** - GitHub Actions, etc.
- [ ] **Competitor analysis** - See how you stack up

*Vote on features in [GitHub Discussions](https://github.com/ugolbck/seofordev/discussions)*

## Support & Community

**Need Help?**

[Documentation](https://docs.seofor.dev) • [Issues](https://github.com/ugolbck/seofordev/issues) • [Discussions](https://github.com/ugolbck/seofordev/discussions)

[hey@seofor.dev](mailto:hey@seofor.dev) • [@ugo_builds](https://x.com/ugo_builds)

**Show Some Love**

Star this repo • Fork & contribute • Share with friends

---

<div align="center">

## Ready to Improve Your SEO?

**Start your SEO journey today - it's free and open source!**

```bash
curl -sSfL https://seofor.dev/install.sh | bash && seo
```

<sub>Made with love for developers who ship fast and rank high</sub>

**[Get Started](https://seofor.dev) • [Star on GitHub](https://github.com/ugolbck/seofordev)**

</div>

---

<details>
<summary>License</summary>

MIT License - see [LICENSE](LICENSE) file for details.

</details>
