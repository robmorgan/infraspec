#!/bin/sh
# InfraSpec installer script
# Usage: curl -fsSL https://infraspec.sh/install.sh | bash
#
# This script detects the OS and architecture, then downloads and installs
# the appropriate InfraSpec binary from GitHub releases.

set -e

REPO="robmorgan/infraspec"
BINARY_NAME="infraspec"
INSTALL_DIR="/usr/local/bin"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

info() {
    printf "${BLUE}==>${NC} %s\n" "$1"
}

success() {
    printf "${GREEN}==>${NC} %s\n" "$1"
}

warn() {
    printf "${YELLOW}Warning:${NC} %s\n" "$1"
}

error() {
    printf "${RED}Error:${NC} %s\n" "$1" >&2
    exit 1
}

# Detect OS
detect_os() {
    case "$(uname -s)" in
        Linux*)     echo "linux" ;;
        Darwin*)    echo "darwin" ;;
        MINGW*|MSYS*|CYGWIN*)  echo "windows" ;;
        *)          error "Unsupported operating system: $(uname -s)" ;;
    esac
}

# Detect architecture
detect_arch() {
    case "$(uname -m)" in
        x86_64|amd64)   echo "amd64" ;;
        arm64|aarch64)  echo "arm64" ;;
        *)              error "Unsupported architecture: $(uname -m)" ;;
    esac
}

# Get the latest release version from GitHub
get_latest_version() {
    if command -v curl >/dev/null 2>&1; then
        curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/'
    elif command -v wget >/dev/null 2>&1; then
        wget -qO- "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/'
    else
        error "Neither curl nor wget found. Please install one of them."
    fi
}

# Download file
download() {
    local url="$1"
    local output="$2"

    if command -v curl >/dev/null 2>&1; then
        curl -fsSL "$url" -o "$output"
    elif command -v wget >/dev/null 2>&1; then
        wget -q "$url" -O "$output"
    else
        error "Neither curl nor wget found. Please install one of them."
    fi
}

# Main installation function
main() {
    echo ""
    printf "${BLUE}"
    echo "  _____        __           ____                   "
    echo " |_   _|      / _|         / ___|_ __   ___  ___   "
    echo "   | |  _ __ | |_ _ __ __ _\\___ \\ '_ \\ / _ \\/ __|  "
    echo "   | | | '_ \\|  _| '__/ _\` |___) | |_) |  __/ (__   "
    echo "  |___||_| |_|_| |_| \\__,_|____/| .__/ \\___|\\___|  "
    echo "                                |_|                "
    printf "${NC}"
    echo ""
    info "Installing InfraSpec..."
    echo ""

    # Detect system
    OS=$(detect_os)
    ARCH=$(detect_arch)
    info "Detected OS: $OS, Architecture: $ARCH"

    # Get latest version
    info "Fetching latest version..."
    VERSION=$(get_latest_version)
    if [ -z "$VERSION" ]; then
        error "Failed to determine latest version"
    fi
    # Remove 'v' prefix if present for the download URL
    VERSION_NUM="${VERSION#v}"
    info "Latest version: $VERSION"

    # Determine archive extension
    if [ "$OS" = "windows" ]; then
        EXT="zip"
    else
        EXT="tar.gz"
    fi

    # Construct download URL
    ARCHIVE_NAME="${BINARY_NAME}_${VERSION_NUM}_${OS}_${ARCH}.${EXT}"
    DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${VERSION}/${ARCHIVE_NAME}"

    info "Downloading $ARCHIVE_NAME..."

    # Create temp directory
    TMP_DIR=$(mktemp -d)
    trap "rm -rf $TMP_DIR" EXIT

    # Download archive
    ARCHIVE_PATH="${TMP_DIR}/${ARCHIVE_NAME}"
    download "$DOWNLOAD_URL" "$ARCHIVE_PATH" || error "Failed to download from $DOWNLOAD_URL"

    # Extract archive
    info "Extracting archive..."
    cd "$TMP_DIR"
    if [ "$EXT" = "zip" ]; then
        if command -v unzip >/dev/null 2>&1; then
            unzip -q "$ARCHIVE_PATH"
        else
            error "unzip is required to extract the archive"
        fi
    else
        tar -xzf "$ARCHIVE_PATH"
    fi

    # Find the binary
    if [ ! -f "$TMP_DIR/$BINARY_NAME" ]; then
        error "Binary not found in archive"
    fi

    # Determine install location
    if [ -w "$INSTALL_DIR" ]; then
        DEST="$INSTALL_DIR"
    elif [ -w "$HOME/.local/bin" ]; then
        DEST="$HOME/.local/bin"
        warn "$INSTALL_DIR is not writable, installing to $DEST instead"
    else
        # Try with sudo
        if command -v sudo >/dev/null 2>&1; then
            info "Requesting sudo access to install to $INSTALL_DIR..."
            DEST="$INSTALL_DIR"
            sudo mkdir -p "$DEST"
            sudo mv "$TMP_DIR/$BINARY_NAME" "$DEST/$BINARY_NAME"
            sudo chmod +x "$DEST/$BINARY_NAME"
        else
            mkdir -p "$HOME/.local/bin"
            DEST="$HOME/.local/bin"
            warn "Installing to $DEST (add this to your PATH)"
        fi
    fi

    # Install binary (if not already done with sudo)
    if [ -f "$TMP_DIR/$BINARY_NAME" ]; then
        mkdir -p "$DEST"
        mv "$TMP_DIR/$BINARY_NAME" "$DEST/$BINARY_NAME"
        chmod +x "$DEST/$BINARY_NAME"
    fi

    # Verify installation
    if [ -x "$DEST/$BINARY_NAME" ]; then
        echo ""
        success "InfraSpec $VERSION installed successfully!"
        echo ""

        # Check if destination is in PATH
        case ":$PATH:" in
            *":$DEST:"*) ;;
            *)
                warn "$DEST is not in your PATH"
                echo "  Add it to your shell profile:"
                echo ""
                echo "    export PATH=\"\$PATH:$DEST\""
                echo ""
                ;;
        esac

        echo "  Get started:"
        echo ""
        echo "    ${GREEN}infraspec init${NC}           # Initialize a new project"
        echo "    ${GREEN}infraspec new test.feature${NC}  # Create a new test"
        echo "    ${GREEN}infraspec features/${NC}      # Run your tests"
        echo ""
        echo "  Documentation: https://infraspec.sh/docs"
        echo "  Virtual Cloud: https://infraspec.sh/virtual-cloud"
        echo ""
    else
        error "Installation failed - binary not found at $DEST/$BINARY_NAME"
    fi
}

main "$@"
