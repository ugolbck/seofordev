#!/bin/bash
set -e

REPO="ugolbck/seofordev"
BINARY="seo"
INSTALL_DIR="/usr/local/bin"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m'

error() {
  echo -e "${RED}error:${NC} $1" >&2
  exit 1
}

success() {
  echo -e "${GREEN}$1${NC}"
}

# Check dependencies
command -v curl >/dev/null 2>&1 || error "curl is required"
command -v tar >/dev/null 2>&1 || error "tar is required"

# Detect OS
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
case "$OS" in
  darwin) OS="darwin" ;;
  linux) OS="linux" ;;
  mingw*|msys*|cygwin*) error "Windows is not supported. Use WSL instead." ;;
  *) error "Unsupported OS: $OS" ;;
esac

# Detect architecture
ARCH=$(uname -m)
case "$ARCH" in
  x86_64|amd64) ARCH="amd64" ;;
  arm64|aarch64) ARCH="arm64" ;;
  *) error "Unsupported architecture: $ARCH" ;;
esac

# Get latest release tag
echo "Fetching latest release..."
LATEST=$(curl -sI "https://github.com/$REPO/releases/latest" | grep -i "^location:" | sed 's/.*tag\///' | tr -d '\r\n')
if [ -z "$LATEST" ]; then
  error "Failed to fetch latest release. Check your internet connection."
fi

echo "Installing seofordev $LATEST ($OS/$ARCH)..."

# Download
URL="https://github.com/$REPO/releases/download/$LATEST/seofordev_${OS}_${ARCH}.tar.gz"
TMP_DIR=$(mktemp -d)
trap "rm -rf $TMP_DIR" EXIT

echo "Downloading from $URL..."
HTTP_CODE=$(curl -sL -w "%{http_code}" "$URL" -o "$TMP_DIR/seofordev.tar.gz")
if [ "$HTTP_CODE" != "200" ]; then
  error "Download failed (HTTP $HTTP_CODE). Release may not exist for your platform."
fi

# Extract
echo "Extracting..."
tar -xzf "$TMP_DIR/seofordev.tar.gz" -C "$TMP_DIR"

# Verify binary exists
if [ ! -f "$TMP_DIR/$BINARY" ]; then
  error "Binary not found in archive"
fi

# Make executable
chmod +x "$TMP_DIR/$BINARY"

# Install
echo "Installing to $INSTALL_DIR..."
if [ -w "$INSTALL_DIR" ]; then
  mv "$TMP_DIR/$BINARY" "$INSTALL_DIR/$BINARY"
else
  sudo mv "$TMP_DIR/$BINARY" "$INSTALL_DIR/$BINARY"
fi

# Create config directory
mkdir -p ~/.seo

success "Installed seo to $INSTALL_DIR/$BINARY"
echo ""
echo "Get started:"
echo "  seo                     # Show help"
echo "  seo audit run           # Audit localhost:3000"
echo "  seo audit run -p 8080   # Audit localhost:8080"
echo ""
echo "Playwright browsers will be installed on first audit."
