#!/usr/bin/env bash
# DevHive Go Binary Installer
# Downloads the correct binary for the current platform and installs to ~/.devhive/bin

set -e

VERSION="${DEVHIVE_VERSION:-0.2.2}"
REPO="Oswald-Hao/devhive"
INSTALL_DIR="$HOME/.devhive/bin"
BINARY="devhive"

# ── Platform detection ────────────────────────────────────────

OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case "$ARCH" in
    x86_64)  ARCH="amd64" ;;
    aarch64) ARCH="arm64" ;;
    arm64)   ARCH="arm64" ;;
    *)       echo "Unsupported architecture: $ARCH" >&2; exit 1 ;;
esac

case "$OS" in
    linux)   OS="linux" ;;
    darwin)  OS="darwin" ;;
    *)       echo "Unsupported OS: $OS" >&2; exit 1 ;;
esac

TARBALL="devhive-${VERSION}-${OS}-${ARCH}.tar.gz"
DOWNLOAD_URL="https://github.com/${REPO}/releases/download/v${VERSION}/${TARBALL}"

# ── Install ───────────────────────────────────────────────────

mkdir -p "$INSTALL_DIR"

echo "DevHive v${VERSION} — installing for ${OS}/${ARCH}..."

if command -v curl &>/dev/null; then
    curl -sL "$DOWNLOAD_URL" -o "/tmp/${TARBALL}"
elif command -v wget &>/dev/null; then
    wget -q "$DOWNLOAD_URL" -O "/tmp/${TARBALL}"
else
    echo "Neither curl nor wget found. Please install one of them." >&2
    exit 1
fi

tar -xzf "/tmp/${TARBALL}" -C "$INSTALL_DIR"
chmod +x "$INSTALL_DIR/$BINARY"
rm "/tmp/${TARBALL}"

# ── PATH setup ────────────────────────────────────────────────

SHELL_NAME=$(basename "$SHELL")
RC_FILE=""

case "$SHELL_NAME" in
    bash) RC_FILE="$HOME/.bashrc" ;;
    zsh)  RC_FILE="$HOME/.zshrc" ;;
    fish) RC_FILE="$HOME/.config/fish/config.fish" ;;
esac

if [ -n "$RC_FILE" ] && ! grep -q "$INSTALL_DIR" "$RC_FILE" 2>/dev/null; then
    echo "export PATH=\"$INSTALL_DIR:\$PATH\"" >> "$RC_FILE"
    echo "Added $INSTALL_DIR to PATH in $RC_FILE"
fi

echo ""
echo "DevHive v${VERSION} installed to $INSTALL_DIR/$BINARY"
echo "Run 'dh' to start."
