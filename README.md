<div align="center">

```
                    ___               _             
                   / __)             | |            
  ___ _____  ___ _| |__ ___   ____ __| |_____ _   _ 
 /___) ___ |/ _ (_   __) _ \ / ___) _  | ___ | | | |
|___ | ____| |_| || | | |_| | |_ ( (_| | ____|\ V / 
(___/|_____)\___/ |_|  \___/|_(_) \____|_____) \_/  
                                                    
```

# The Developer's SEO Toolkit ✨

**Lightning-fast CLI for SEO audits, keyword research, and content optimization**

[![MIT License](https://img.shields.io/badge/License-MIT-green.svg)](https://choosealicense.com/licenses/mit/)
[![Go Report Card](https://goreportcard.com/badge/github.com/ugolbck/seofordev)](https://goreportcard.com/report/github.com/ugolbck/seofordev)
[![GitHub release](https://img.shields.io/github/release/ugolbck/seofordev.svg)](https://github.com/ugolbck/seofordev/releases/latest)
[![GitHub stars](https://img.shields.io/github/stars/ugolbck/seofordev?style=social)](https://github.com/ugolbck/seofordev/stargazers)

**Stop losing traffic to SEO mistakes. Audit, optimize, and dominate search results from your terminal.**

[✨ Get Started](#-quick-start) • [🔥 Features](#-features) • [📖 Docs](https://docs.seofor.dev) • [💎 Premium](https://seofor.dev/payments/pricing)

</div>

---

## 🎬 Demo

**🔍 Instant localhost SEO audits**
```bash
$ seo audit run
🚀 Crawling localhost:3000...
✅ Found 12 pages, 0 errors
📊 Generated optimization prompts - ready for AI!
```

**⚡ Lightning-fast keyword research**
```bash
$ seo keyword generate "saas analytics"
🔍 Generated 50 keywords for 'saas analytics'
💰 Credits used: 10 | Top keyword: "free saas analytics" (5000/mo, 23% difficulty)
```

## 🔥 Features

### 🆓 **Free & Always Available**
- **🔍 Localhost SEO Audits** - Deep-dive analysis of your development sites
- **🤖 AI-Ready Exports** - Copy/paste prompts for Claude, Cursor, ChatGPT
- **⚡ Blazing Fast** - Built in Go, powered by Playwright
- **📊 Smart Crawling** - JavaScript-rendered pages, respect robots.txt
- **🎯 Zero Config** - Works out of the box, no setup required

### 💎 **Premium Power Features**
- **🔬 Advanced Keyword Research** - Volume, difficulty, CPC data
- **📝 AI Content Briefs** - Competition analysis + optimization strategies
- **📈 Search Intent Analysis** - Know exactly what users want
- **⏰ Generation History** - Never lose your research again

## ⚡ Quick Start

### 1️⃣ Install (30 seconds)
```bash
# macOS/Linux (recommended)
curl -sSfL https://seofor.dev/install.sh | bash
```

### 2️⃣ Audit Your Site (2 minutes)
```bash
# Start your dev server (e.g., npm run dev, rails server, etc.)
$ cd your-project && npm run dev

# In another terminal, audit it
$ seo audit run --port 3000
# 🎉 Get instant SEO insights + AI optimization prompts
```

### 3️⃣ Level Up with Keyword Research
```bash
# Get your API key at https://seofor.dev/dashboard
$ seo config set-api-key your-api-key-here
$ seo keyword generate "your product keyword"
# 📈 Discover high-traffic, low-competition keywords
```

## 🛠️ Commands Reference

<table>
<tr>
<td>

**🆓 Free Features**
```bash
seo audit run               # Audit localhost
seo audit list              # View audit history  
seo audit show <id>         # Detailed results
seo config                  # Show settings
seo index submit <url>      # IndexNow submission
```

</td>
<td>

**💎 Premium Features**
```bash
seo keyword generate <term>    # Research keywords
seo keyword history            # View past research
seo keyword show <id>          # Detailed results
seo brief generate <keyword>   # Content brief
seo brief history              # View past briefs
seo pro status                 # Account info
```

</td>
</tr>
</table>

## 🎯 Real-World Workflows

### 🔧 **Developer Workflow**: Pre-Launch SEO Check
```bash
# Before deploying
seo audit run --port 3000
# Fix issues highlighted in AI prompts
# Deploy with confidence 🚀
```

### 📝 **Content Creator Workflow**: Research → Write → Optimize  
```bash
# 1. Research profitable keywords
seo keyword generate "AI tools"

# 2. Generate data-driven content brief  
seo brief generate "best AI tools for developers"

# 3. Export to your AI editor and create content
seo brief show <id> --copy
# Paste into Claude/Cursor/ChatGPT ✨
```

### 📊 **SEO Agency Workflow**: Client Audits
```bash
# Quick client site analysis
seo audit run --port 8080
seo audit export <id>
# Professional SEO recommendations ready 📈
```

## 🏆 Why Choose seofor.dev?

<div align="center">

| 🆚 **vs Manual SEO** | 🆚 **vs Other Tools** | 🆚 **vs Enterprise** |
|:---:|:---:|:---:|
| ⚡ **10x Faster** | 🎯 **Developer-First** | 💸 **95% Cheaper** |
| 🤖 **AI-Integrated** | 🔓 **Open Source** | 🚀 **Zero Setup** |
| 📊 **Data-Driven** | ⚡ **CLI-Native** | 🛡️ **Privacy-First** |

</div>

## 🏗️ For Contributors

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

**🤝 Contributing**: We ❤️ contributions! Check out our [contribution guide](CONTRIBUTING.md) and help make SEO better for all developers.

## 🎨 Built With

- **🔹 Go** - Performance & reliability
- **🔹 Cobra** - Beautiful CLI experience  
- **🔹 Playwright** - Modern web crawling

## 🔮 Roadmap

- [ ] **CI/CD integration** - GitHub Actions, etc.
- [ ] **Competitor analysis** - See how you stack up

*Vote on features in [GitHub Discussions](https://github.com/ugolbck/seofordev/discussions)*

## 📞 Support & Community

<div align="center">

**Need Help?**

[📖 Documentation](https://docs.seofor.dev) • [🐛 Issues](https://github.com/ugolbck/seofordev/issues) • [💬 Discussions](https://github.com/ugolbck/seofordev/discussions)

[✉️ hey@seofor.dev](mailto:hey@seofor.dev) • [🐦 @ugo_builds](https://x.com/ugo_builds)

**Show Some Love**

⭐ Star this repo • 🔀 Fork & contribute • 📢 Share with friends

</div>

---

<div align="center">

## 🎉 Ready to Dominate Search Results?

**Start your SEO journey today - it's free!**

```bash
curl -sSfL https://seofor.dev/install.sh | bash && seo
```

<sub>Made with ❤️ for developers who ship fast and rank high</sub>

**[🚀 Get Started](https://seofor.dev) • [💎 Go Premium](https://seofor.dev/payments/pricing) • [⭐ Star on GitHub](https://github.com/ugolbck/seofordev)**

</div>

---

<details>
<summary>📄 License</summary>

MIT License - see [LICENSE](LICENSE) file for details.

*Premium features require a seofor.dev subscription and are subject to additional terms of service.*

</details>