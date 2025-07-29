# Unix/Linux/macOS Installation

This guide covers installing SEOForDev on Unix-based systems including Linux distributions and macOS.

## Prerequisites

Before installing SEOForDev, ensure your system meets these requirements:

- **Linux**: glibc 2.17 or later
- **macOS**: macOS 10.15 (Catalina) or later
- **Architecture**: x64 or ARM64 (Apple Silicon)
- **Terminal access**: Ability to run commands in terminal/command line

## Installation Methods

### Method 1: Direct Download (Recommended)

This is the fastest and most reliable installation method.

#### Step 1: Download the Binary

Choose the appropriate binary for your system:

**For Linux (x64):**
```bash
curl -L -o seofordev https://github.com/ugolbck/seofordev/releases/latest/download/seofordev-linux-x64
```

**For macOS (Intel):**
```bash
curl -L -o seofordev https://github.com/ugolbck/seofordev/releases/latest/download/seofordev-macos-x64
```

**For macOS (Apple Silicon):**
```bash
curl -L -o seofordev https://github.com/ugolbck/seofordev/releases/latest/download/seofordev-macos-arm64
```

#### Step 2: Make Executable

```bash
chmod +x seofordev
```

#### Step 3: Install to System Path

**Option A: Install for all users (requires sudo)**
```bash
sudo mv seofordev /usr/local/bin/
```

**Option B: Install for current user only**
```bash
mkdir -p ~/.local/bin
mv seofordev ~/.local/bin/
echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.bashrc
source ~/.bashrc
```

### Method 2: Package Managers

#### Homebrew (macOS/Linux)

If you have Homebrew installed:

```bash
brew install ugolbck/seofordev/seofordev
```

#### Snap (Linux)

```bash
sudo snap install seofordev
```

#### AUR (Arch Linux)

```bash
yay -S seofordev
```

### Method 3: Manual Installation

For advanced users who prefer manual control:

```bash
# Create installation directory
sudo mkdir -p /opt/seofordev

# Download and extract
cd /tmp
curl -L -O https://github.com/ugolbck/seofordev/releases/latest/download/seofordev-linux-x64.tar.gz
tar -xzf seofordev-linux-x64.tar.gz

# Move to installation directory
sudo mv seofordev /opt/seofordev/

# Create symlink
sudo ln -s /opt/seofordev/seofordev /usr/local/bin/seofordev
```

## Verification

After installation, verify that SEOForDev is working:

```bash
seofordev --version
```

Expected output:
```
SEOForDev v0.1.0
```

## Testing the Installation

Run a quick test to ensure everything is working:

```bash
seofordev --help
```

This should display the help menu with available commands and options.

## Troubleshooting

### Permission Denied Error

If you encounter a "Permission denied" error:

```bash
# Check file permissions
ls -la seofordev

# Fix permissions if needed
chmod +x seofordev
```

### Command Not Found

If `seofordev` command is not found:

1. **Check if it's in your PATH:**
   ```bash
   which seofordev
   echo $PATH
   ```

2. **Add to PATH if needed:**
   ```bash
   # For bash
   echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.bashrc
   source ~/.bashrc
   
   # For zsh
   echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.zshrc
   source ~/.zshrc
   ```

### macOS Security Issues

On macOS, you might encounter security warnings:

1. **Go to System Preferences > Security & Privacy**
2. **Click "Allow Anyway" for the seofordev binary**
3. **Or run with explicit permission:**
   ```bash
   xattr -d com.apple.quarantine seofordev
   ```

## Next Steps

- [Windows Installation](windows.md) - Installation guide for Windows users