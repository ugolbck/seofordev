# Installation Overview

SEOForDev is available for multiple platforms and can be installed in several ways. Choose the method that best suits your needs and operating system.

## Installation Methods

### 1. Binary Downloads (Recommended)

The easiest way to install SEOForDev is by downloading pre-compiled binaries from our [GitHub Releases](https://github.com/ugolbck/seofordev/releases). This method works for all supported platforms and doesn't require any additional dependencies.

**Supported Platforms:**
- Windows (x64)
- Linux (x64)
- macOS (x64, ARM64)


## System Requirements

- **Operating System**: Windows 10+, macOS 10.15+, or Linux (glibc 2.17+)
- **Architecture**: x64 or ARM64 (macOS)

## Quick Installation

### Windows
```powershell
# Download and run the installer
Invoke-WebRequest -Uri "https://github.com/ugolbck/seofordev/releases/latest/download/seofordev-windows-x64.exe" -OutFile "seofordev.exe"
```

### Linux/macOS
```bash
# Download and make executable
curl -L -o seo https://github.com/ugolbck/seofordev/releases/latest/download/seofordev-linux-x64
chmod +x seo
sudo mv seo /usr/local/bin/
```

## Verification

After installation, verify that SEOForDev is working correctly:

```bash
seo --version
```

You should see output similar to:
```
SEOForDev v0.1.0
```

## Next Steps

- [Unix/Linux/macOS Installation](unix.md) - Detailed instructions for Unix-based systems
- [Windows Installation](windows.md) - Step-by-step Windows installation guide
- [Troubleshooting](troubleshooting.md) - Common installation issues and solutions

## Need Help?

If you encounter any issues during installation, please:

1. Check the [troubleshooting guide](troubleshooting.md)
2. Review the [GitHub Issues](https://github.com/ugolbck/seofordev/issues)
3. Join our [Discord community](https://discord.gg/seofordev) for support 