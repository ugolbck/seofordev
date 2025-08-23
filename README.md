# SEOForDev 🚀

[![MIT License](https://img.shields.io/badge/License-MIT-green.svg)](https://choosealicense.com/licenses/mit/)
[![Go Report Card](https://goreportcard.com/badge/github.com/ugolbck/seofordev)](https://goreportcard.com/report/github.com/ugolbck/seofordev)
[![GitHub release](https://img.shields.io/github/release/ugolbck/seofordev.svg)](https://github.com/ugolbck/seofordev/releases/latest)

A powerful CLI-first SEO tool that helps developers audit their websites and optimize for search engines from day one. Built with Go and featuring an interactive terminal interface.

## ✨ Features

### 🔓 **Free & Open Source**
- **Website Auditing**: Comprehensive SEO analysis of your localhost
- **Interactive TUI**: Beautiful terminal user interface built with Bubble Tea
- **Playwright Integration**: Reliable web crawling with JavaScript support
- **Export Results**: Generate AI-ready optimization prompts

### 💎 **Premium Features** (API Key Required)
- **Keyword Research**: Generate high-volume, low-competition keywords
- **Content Brief Generation**: Create SEO-optimized content briefs to create quality content

## 🚀 Installation

### Option 1: Install Script (Recommended)
```bash
curl -sSfL https://seofor.dev/install.sh | bash
```

### Option 2: Download Binary
Visit our [releases page](https://github.com/ugolbck/seofordev/releases/latest) and download the appropriate binary for your platform.

### Option 3: Build from Source
```bash
git clone https://github.com/ugolbck/seofordev.git
cd seofordev
go build -o seo .
```

## 📋 Requirements

- **Operating System**: macOS, Linux, or Windows (WSL)
- **Network**: Internet connection for Playwright setup and premium features
- **Permissions**: May require sudo/admin rights for initial Playwright installation

## 🎯 Quick Start

1. **Launch the application**:
   ```bash
   seo
   ```

2. **First-time setup**: The app will automatically install Playwright dependencies (~150MB, one-time download)

3. **Start auditing**: Navigate to `Localhost Audit > New Audit` and audit your development server

4. **Export results**: Press `e` to copy SEO optimization prompts to your clipboard

5. **Integrate with AI**: Paste the prompts into Claude Code, Cursor, or your preferred AI coding assistant

## 🔑 Premium Features Setup

For keyword research and content brief generation:

1. Sign up at [seofor.dev/dashboard](https://seofor.dev/dashboard/)
2. Generate an API key
3. Enter it when prompted in the application
4. Access premium features from the main menu

## Premium features

These features require a paid plan (monthly subscription, or lifetime one-time purchase). You don't need them to optimize your existing pages (site audit), but **they are great to generate new, relevant content** to improve your app's visibility and rankings online. This is eventually **necessary to drive more organic traffic**, and potential customers to your app.

The sooner you generate interesting, high-value content that people are looking for, the more early traffic you will get.

### Keyword suggestions

A great way to start your SEO journey is to pick a set of keywords that you want to _rank on_. You want to write content that echoes with those keywords.

Optimally, you want to target keywords that a lot of people are looking for (high Volume), but that few websites are already covering (low Difficulty). The Keyword Suggestions tool is here for that.

If you don't know where to start, or if you're looking for easy keywords with high volume, this tool will do that for you.

### Content Brief generation

Once you have targeted a few keywords you want to rank on, you will need to generate content.

An easy way to start is to set up a simple, markdown-based blog on your site, and **create a new article per target keyword**.

The brief generator synergizes with your favorite AI code editors. Since we don't know what web framework you're using, we leave it up to your AI to write and integrate the actual article into your website. **We provide the data and instructions that lead to high ranking on your chosen keyword**, your AI tool does the rest. Don't hesitate to tweak the prompt we give you if you wish a specific tone.

We strongly encourage you to proofread any AI-generated content, as it may make mistakes.

#### How to use

1. Navigate to the `Content Brief Generation` tool.
2. Enter the target keyword and run the generation (this may take up to a minute, you may leave the app and check your content brief history later).
3. We analyze your competitors and many other data points, in order to generate a content brief that is tailored for your audience and optimized to compete with Google's top results.
4. Press `e` on the results page to export the content brief to your clipboard.
5. Paste the content brief prompt into your favorite AI code editor (Cursor, Claude Code, etc.) and let the magic happen.
6. Once you've pushed your new content online, we recommend you log in to your Google Search Console, and ask Google to index your new page URL. This should speed up the indexation process.


## 🏗️ Development

### Building from Source
```bash
git clone https://github.com/ugolbck/seofordev.git
cd seofordev
go mod download
go build -o seo .
```

### Running Tests
```bash
go test ./...
```

### Contributing
We welcome contributions! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

## 📚 Documentation

- **API Documentation**: [docs.seofor.dev](https://docs.seofor.dev)
- **Architecture**: See [CLAUDE.md](CLAUDE.md) for development guidance
- **Security Policy**: [SECURITY.md](SECURITY.md)

## 🆘 Support

- **Issues**: [GitHub Issues](https://github.com/ugolbck/seofordev/issues)
- **Discussions**: [GitHub Discussions](https://github.com/ugolbck/seofordev/discussions)  
- **Email**: hey@seofor.dev
- **Twitter**: [@ugo_builds](https://x.com/ugo_builds)

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

**Note**: Premium features require a subscription to seofor.dev and are subject to additional terms of service.

---

<div align="center">
  <p>Made with ❤️ for developers who care about SEO</p>
  <p><a href="https://seofor.dev">seofor.dev</a> • <a href="https://github.com/ugolbck/seofordev">GitHub</a></p>
</div>
