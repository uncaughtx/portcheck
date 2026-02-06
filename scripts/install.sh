#!/bin/bash
# portcheck installer script
# Usage: curl -sSL https://raw.githubusercontent.com/uncaughtx/portcheck/main/scripts/install.sh | bash

set -e

VERSION="v0.1.0"
REPO="uncaughtx/portcheck"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

error() {
    echo -e "${RED}[ERROR]${NC} $1"
    exit 1
}

# Detect OS and architecture
detect_platform() {
    OS=$(uname -s | tr '[:upper:]' '[:lower:]')
    ARCH=$(uname -m)

    case "$OS" in
        linux)
            OS="linux"
            ;;
        darwin)
            OS="darwin"
            ;;
        mingw*|msys*|cygwin*)
            OS="windows"
            ;;
        *)
            error "Unsupported OS: $OS"
            ;;
    esac

    case "$ARCH" in
        x86_64|amd64)
            ARCH="amd64"
            ;;
        aarch64|arm64)
            ARCH="arm64"
            ;;
        *)
            error "Unsupported architecture: $ARCH"
            ;;
    esac

    # Windows doesn't support arm64 yet
    if [[ "$OS" == "windows" && "$ARCH" == "arm64" ]]; then
        error "Windows ARM64 is not supported yet"
    fi

    info "Detected platform: ${OS}/${ARCH}"
}

# Get the latest version if not specified
get_version() {
    if [[ "$VERSION" == "latest" ]]; then
        info "Fetching latest version..."
        VERSION=$(curl -sSL "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/')
        if [[ -z "$VERSION" ]]; then
            error "Failed to fetch latest version"
        fi
        info "Latest version: $VERSION"
    fi
}

# Download and install
install() {
    BINARY_NAME="portcheck"
    if [[ "$OS" == "windows" ]]; then
        BINARY_NAME="portcheck.exe"
    fi

    # Construct download URL
    # Format: portcheck_v0.1.0_linux_x86_64.tar.gz
    ARCHIVE_NAME="portcheck_${VERSION}_${OS}_"
    case "$ARCH" in
        amd64)
            ARCHIVE_NAME="${ARCHIVE_NAME}x86_64"
            ;;
        arm64)
            ARCHIVE_NAME="${ARCHIVE_NAME}arm64"
            ;;
    esac

    if [[ "$OS" == "windows" ]]; then
        ARCHIVE_NAME="${ARCHIVE_NAME}.zip"
    else
        ARCHIVE_NAME="${ARCHIVE_NAME}.tar.gz"
    fi

    DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${VERSION}/${ARCHIVE_NAME}"

    info "Downloading from: $DOWNLOAD_URL"

    # Create temp directory
    TMP_DIR=$(mktemp -d)
    trap "rm -rf $TMP_DIR" EXIT

    # Download
    if ! curl -sSL "$DOWNLOAD_URL" -o "$TMP_DIR/$ARCHIVE_NAME"; then
        error "Failed to download $DOWNLOAD_URL"
    fi

    # Extract
    cd "$TMP_DIR"
    if [[ "$ARCHIVE_NAME" == *.zip ]]; then
        unzip -q "$ARCHIVE_NAME"
    else
        tar -xzf "$ARCHIVE_NAME"
    fi

    # Install
    if [[ -w "$INSTALL_DIR" ]]; then
        mv "$BINARY_NAME" "$INSTALL_DIR/"
    else
        info "Requesting sudo to install to $INSTALL_DIR"
        sudo mv "$BINARY_NAME" "$INSTALL_DIR/"
    fi

    chmod +x "$INSTALL_DIR/$BINARY_NAME"

    info "Installed portcheck to $INSTALL_DIR/$BINARY_NAME"
}

# Verify installation
verify() {
    if command -v portcheck &> /dev/null; then
        info "Installation successful!"
        portcheck version
    else
        warn "portcheck installed but not in PATH"
        warn "Add $INSTALL_DIR to your PATH or run: $INSTALL_DIR/portcheck"
    fi
}

main() {
    echo ""
    echo "  portcheck installer"
    echo "  ==================="
    echo ""

    detect_platform
    get_version
    install
    verify

    echo ""
    info "Run 'portcheck --help' to get started"
    echo ""
}

main
