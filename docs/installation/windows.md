# Windows Installation

This guide covers installing SEOForDev on Windows systems, including Windows 10, Windows 11, and Windows Server.

## Prerequisites

Before installing SEOForDev, ensure your system meets these requirements:

- **Operating System**: Windows 10 (version 1903) or later, Windows 11, or Windows Server 2019+
- **Architecture**: x64 (64-bit)
- **PowerShell**: PowerShell 5.1 or later (included with Windows 10/11)
- **Administrator Access**: Required for system-wide installation

## Installation Methods

### Method 1: Direct Download (Recommended)

This is the most straightforward installation method for Windows users.

#### Step 1: Download the Binary

Open PowerShell as Administrator and run:

```powershell
# Create a temporary directory
New-Item -ItemType Directory -Path "$env:TEMP\seofordev" -Force
Set-Location "$env:TEMP\seofordev"

# Download the Windows binary
Invoke-WebRequest -Uri "https://github.com/ugolbck/seofordev/releases/latest/download/seofordev-windows-x64.exe" -OutFile "seofordev.exe"
```

#### Step 2: Install to System Path

**Option A: Install for all users (requires Administrator)**
```powershell
# Move to Program Files
Move-Item -Path "seofordev.exe" -Destination "C:\Program Files\seofordev\seofordev.exe" -Force

# Add to PATH
$currentPath = [Environment]::GetEnvironmentVariable("PATH", "Machine")
$newPath = $currentPath + ";C:\Program Files\seofordev"
[Environment]::SetEnvironmentVariable("PATH", $newPath, "Machine")

# Refresh environment variables
$env:PATH = [System.Environment]::GetEnvironmentVariable("PATH","Machine") + ";" + [System.Environment]::GetEnvironmentVariable("PATH","User")
```

**Option B: Install for current user only**
```powershell
# Create user bin directory
New-Item -ItemType Directory -Path "$env:USERPROFILE\bin" -Force

# Move executable
Move-Item -Path "seofordev.exe" -Destination "$env:USERPROFILE\bin\seofordev.exe" -Force

# Add to user PATH
$currentPath = [Environment]::GetEnvironmentVariable("PATH", "User")
$newPath = $currentPath + ";$env:USERPROFILE\bin"
[Environment]::SetEnvironmentVariable("PATH", $newPath, "User")

# Refresh environment variables
$env:PATH = [System.Environment]::GetEnvironmentVariable("PATH","Machine") + ";" + [System.Environment]::GetEnvironmentVariable("PATH","User")
```

### Method 2: Package Managers

#### Scoop (Recommended for Developers)

If you have Scoop installed:

```powershell
# Add the bucket (if not already added)
scoop bucket add seofordev https://github.com/ugolbck/seofordev-scoop

# Install SEOForDev
scoop install seofordev
```

#### Chocolatey

If you have Chocolatey installed:

```powershell
choco install seofordev
```

#### Winget

Using Windows Package Manager:

```powershell
winget install ugolbck.seofordev
```

### Method 3: Manual Installation

For users who prefer manual control:

```powershell
# Create installation directory
New-Item -ItemType Directory -Path "C:\seofordev" -Force

# Download and extract
Set-Location "C:\seofordev"
Invoke-WebRequest -Uri "https://github.com/ugolbck/seofordev/releases/latest/download/seofordev-windows-x64.zip" -OutFile "seofordev.zip"
Expand-Archive -Path "seofordev.zip" -DestinationPath "." -Force

# Add to PATH
$currentPath = [Environment]::GetEnvironmentVariable("PATH", "Machine")
$newPath = $currentPath + ";C:\seofordev"
[Environment]::SetEnvironmentVariable("PATH", $newPath, "Machine")
```

## Verification

After installation, verify that SEOForDev is working:

```powershell
seofordev --version
```

Expected output:
```
SEOForDev v0.1.0
```

## Testing the Installation

Run a quick test to ensure everything is working:

```powershell
seofordev --help
```

This should display the help menu with available commands and options.

## Troubleshooting

### Command Not Found

If `seofordev` command is not found:

1. **Check if it's in your PATH:**
   ```powershell
   Get-Command seofordev -ErrorAction SilentlyContinue
   $env:PATH -split ';'
   ```

2. **Refresh your terminal session:**
   - Close and reopen PowerShell/Command Prompt
   - Or run: `refreshenv` (if Chocolatey is installed)

### Execution Policy Issues

If you encounter execution policy errors:

```powershell
# Check current execution policy
Get-ExecutionPolicy

# Set execution policy (requires Administrator)
Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser
```

### Antivirus False Positives

Some antivirus software may flag the binary:

1. **Add an exception** for the seofordev executable
2. **Submit the file** to your antivirus vendor as a false positive
3. **Use Windows Defender** exclusion if needed:
   ```powershell
   Add-MpPreference -ExclusionPath "C:\Program Files\seofordev"
   ```

### Permission Denied

If you get permission errors:

1. **Run PowerShell as Administrator**
2. **Check file permissions:**
   ```powershell
   Get-Acl "C:\Program Files\seofordev\seofordev.exe"
   ```

## Using SEOForDev on Windows

### PowerShell Integration

SEOForDev works seamlessly with PowerShell:

```powershell
# Run analysis
seofordev analyze https://localhost:3000

# Save output to file
seofordev analyze https://localhost:3000 --output report.json

# Use with PowerShell variables
$url = "https://localhost:3000"
seofordev analyze $url
```

### Command Prompt

SEOForDev also works in Command Prompt:

```cmd
seofordev --version
seofordev analyze https://localhost:3000
```

### Windows Terminal

For the best experience, use Windows Terminal:

1. **Install Windows Terminal** from Microsoft Store
2. **Open a new tab** with PowerShell
3. **Run SEOForDev commands** as normal
