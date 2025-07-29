# Troubleshooting Installation Issues

This guide helps you resolve common installation problems with SEOForDev across different platforms.

## Common Issues

### Command Not Found

**Symptoms:** `seofordev: command not found` or `'seofordev' is not recognized`

**Solutions:**

#### Linux/macOS
```bash
# Check if the binary exists
which seofordev
ls -la /usr/local/bin/seofordev

# Verify PATH
echo $PATH

# Reinstall to correct location
sudo mv seofordev /usr/local/bin/
chmod +x /usr/local/bin/seofordev
```

#### Windows
```powershell
# Check if executable exists
Get-Command seofordev -ErrorAction SilentlyContinue

# Verify PATH
$env:PATH -split ';'

# Reinstall to Program Files
Move-Item -Path "seofordev.exe" -Destination "C:\Program Files\seofordev\seofordev.exe" -Force
```

### Permission Denied

**Symptoms:** `Permission denied` when trying to run the binary

**Solutions:**

#### Linux/macOS
```bash
# Fix executable permissions
chmod +x seofordev

# Check file permissions
ls -la seofordev

# If installed system-wide, check ownership
sudo chown root:root /usr/local/bin/seofordev
sudo chmod 755 /usr/local/bin/seofordev
```

#### Windows
```powershell
# Run PowerShell as Administrator
# Check file permissions
Get-Acl "C:\Program Files\seofordev\seofordev.exe"

# Fix permissions if needed
icacls "C:\Program Files\seofordev\seofordev.exe" /grant Everyone:F
```

### Binary Not Executable

**Symptoms:** `cannot execute binary file: Exec format error`

**Solutions:**

1. **Download the correct architecture:**
   - For Intel Macs: `seofordev-macos-x64`
   - For Apple Silicon: `seofordev-macos-arm64`
   - For Linux: `seofordev-linux-x64`
   - For Windows: `seofordev-windows-x64.exe`

2. **Check your system architecture:**
   ```bash
   # Linux/macOS
   uname -m
   
   # Windows
   echo $env:PROCESSOR_ARCHITECTURE
   ```

### Download Failures

**Symptoms:** Network errors or incomplete downloads

**Solutions:**

#### Using curl (Linux/macOS)
```bash
# Retry with different options
curl -L -o seofordev https://github.com/ugolbck/seofordev/releases/latest/download/seofordev-linux-x64

# Or use wget
wget -O seofordev https://github.com/ugolbck/seofordev/releases/latest/download/seofordev-linux-x64
```

#### Using PowerShell (Windows)
```powershell
# Retry with different options
Invoke-WebRequest -Uri "https://github.com/ugolbck/seofordev/releases/latest/download/seofordev-windows-x64.exe" -OutFile "seofordev.exe" -UseBasicParsing

# Or use curl.exe
curl.exe -L -o seofordev.exe https://github.com/ugolbck/seofordev/releases/latest/download/seofordev-windows-x64.exe
```

### Antivirus Interference

**Symptoms:** Binary is quarantined or blocked

**Solutions:**

#### Windows
1. **Add exclusion to Windows Defender:**
   ```powershell
   Add-MpPreference -ExclusionPath "C:\Program Files\seofordev"
   ```

2. **Check Windows Security:**
   - Open Windows Security
   - Go to Virus & threat protection
   - Click "Manage settings"
   - Add exclusion for the seofordev folder

#### macOS
```bash
# Remove quarantine attribute
xattr -d com.apple.quarantine seofordev

# Or allow in System Preferences
# System Preferences > Security & Privacy > General > Allow Anyway
```

### Package Manager Issues

#### Homebrew Issues
```bash
# Update Homebrew
brew update

# Clean up
brew cleanup

# Reinstall
brew uninstall seofordev
brew install ugolbck/seofordev/seofordev
```

#### Scoop Issues
```powershell
# Update Scoop
scoop update

# Clean up
scoop cleanup *

# Reinstall
scoop uninstall seofordev
scoop install seofordev
```

### Environment Variable Problems

**Symptoms:** Command works in some terminals but not others

**Solutions:**

#### Linux/macOS
```bash
# Check current shell
echo $SHELL

# For bash
echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.bashrc
source ~/.bashrc

# For zsh
echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.zshrc
source ~/.zshrc

# For fish
echo 'set -gx PATH $HOME/.local/bin $PATH' >> ~/.config/fish/config.fish
source ~/.config/fish/config.fish
```

#### Windows
```powershell
# Check current PATH
$env:PATH -split ';'

# Refresh environment variables
refreshenv

# Or restart your terminal session
```

### Corrupted Downloads

**Symptoms:** Binary runs but crashes or shows unexpected behavior

**Solutions:**

1. **Verify file integrity:**
   ```bash
   # Linux/macOS
   sha256sum seofordev
   
   # Windows
   Get-FileHash seofordev.exe -Algorithm SHA256
   ```

2. **Redownload the file:**
   ```bash
   # Remove corrupted file
   rm seofordev
   
   # Download again
   curl -L -o seofordev https://github.com/ugolbck/seofordev/releases/latest/download/seofordev-linux-x64
   ```

## Platform-Specific Issues

### Linux Issues

#### Missing Dependencies
```bash
# Ubuntu/Debian
sudo apt update
sudo apt install libc6

# CentOS/RHEL
sudo yum install glibc

# Check glibc version
ldd --version
```

#### SELinux Issues
```bash
# Check SELinux status
sestatus

# If enabled, add context
sudo semanage fcontext -a -t bin_t "/usr/local/bin/seofordev"
sudo restorecon -v /usr/local/bin/seofordev
```

### macOS Issues

#### Gatekeeper Issues
```bash
# Remove quarantine attribute
xattr -d com.apple.quarantine seofordev

# Or allow in System Preferences
# System Preferences > Security & Privacy > General > Allow Anyway
```

#### Rosetta Issues (Apple Silicon)
```bash
# Install Rosetta if needed
softwareupdate --install-rosetta

# Check if binary is universal
file seofordev
```

### Windows Issues

#### Execution Policy
```powershell
# Check execution policy
Get-ExecutionPolicy

# Set for current user
Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser

# Or bypass for single command
PowerShell -ExecutionPolicy Bypass -Command "seofordev --version"
```

#### UAC Issues
1. **Run PowerShell as Administrator**
2. **Or install for current user only:**
   ```powershell
   New-Item -ItemType Directory -Path "$env:USERPROFILE\bin" -Force
   Move-Item -Path "seofordev.exe" -Destination "$env:USERPROFILE\bin\seofordev.exe"
   ```

## Getting Help

If you're still experiencing issues:

1. **Check the GitHub Issues:** [https://github.com/ugolbck/seofordev/issues](https://github.com/ugolbck/seofordev/issues)

2. **Create a new issue** with:
   - Your operating system and version
   - Installation method used
   - Exact error message
   - Steps to reproduce the issue

3. **Include system information:**
   ```bash
   # Linux/macOS
   uname -a
   lsb_release -a  # Linux only
   
   # Windows
   systeminfo
   ```
