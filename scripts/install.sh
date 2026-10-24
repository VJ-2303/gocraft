#!/bin/sh
# gocraft installer for macOS and Linux
# Usage: curl -fsSL https://raw.githubusercontent.com/VJ-2303/gocraft/main/scripts/install.sh | sh

set -eu

REPO="VJ-2303/gocraft"
BINARY="gocraft"

# Detect Operating System
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
case "$OS" in
  darwin) OS="darwin" ;;
  linux) OS="linux" ;;
  *)
    echo "Error: unsupported operating system '$OS'. gocraft supports macOS and Linux." >&2
    exit 1
    ;;
esac

# Detect Architecture
ARCH=$(uname -m)
case "$ARCH" in
  x86_64|amd64) ARCH="amd64" ;;
  arm64|aarch64) ARCH="arm64" ;;
  *)
    echo "Error: unsupported architecture '$ARCH'. gocraft supports amd64 and arm64." >&2
    exit 1
    ;;
esac

# Resolve Version
VERSION="${VERSION:-}"
if [ -z "$VERSION" ]; then
  # Query GitHub API for latest release tag
  VERSION=$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" 2>/dev/null | grep '"tag_name":' | head -n1 | sed -E 's/.*"tag_name": *"([^"]+)".*/\1/') || true
  if [ -z "$VERSION" ]; then
    # Fallback to HTTP redirect location if API is rate-limited
    VERSION=$(curl -fssI "https://github.com/${REPO}/releases/latest" | tr -d '\r' | grep -i '^location:' | sed -E 's|.*/tag/([^/]+).*|\1|') || true
  fi
fi

if [ -z "$VERSION" ]; then
  echo "Error: unable to determine latest gocraft release. Specify VERSION=v1.0.0 explicitly." >&2
  exit 1
fi

TAG="$VERSION"
CLEAN_VERSION="${VERSION#v}"
TARBALL="${BINARY}_${CLEAN_VERSION}_${OS}_${ARCH}.tar.gz"
DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${TAG}/${TARBALL}"
CHECKSUMS_URL="https://github.com/${REPO}/releases/download/${TAG}/checksums.txt"

echo "Downloading gocraft ${TAG} for ${OS}/${ARCH}..."

TMP_DIR=$(mktemp -d)
trap 'rm -rf "$TMP_DIR"' EXIT INT TERM

# Download archive and checksums
curl -fsSL "$DOWNLOAD_URL" -o "$TMP_DIR/$TARBALL"
curl -fsSL "$CHECKSUMS_URL" -o "$TMP_DIR/checksums.txt" 2>/dev/null || true

# Verify checksum if checksums.txt was retrieved
if [ -f "$TMP_DIR/checksums.txt" ]; then
  EXPECTED_SHA=$(grep -E "[[:space:]]${TARBALL}$" "$TMP_DIR/checksums.txt" | awk '{print $1}') || true
  if [ -n "$EXPECTED_SHA" ]; then
    if command -v sha256sum >/dev/null 2>&1; then
      ACTUAL_SHA=$(sha256sum "$TMP_DIR/$TARBALL" | awk '{print $1}')
    elif command -v shasum >/dev/null 2>&1; then
      ACTUAL_SHA=$(shasum -a 256 "$TMP_DIR/$TARBALL" | awk '{print $1}')
    else
      ACTUAL_SHA=""
    fi

    if [ -n "$ACTUAL_SHA" ] && [ "$EXPECTED_SHA" != "$ACTUAL_SHA" ]; then
      echo "Error: checksum verification failed for $TARBALL" >&2
      echo "Expected: $EXPECTED_SHA" >&2
      echo "Actual:   $ACTUAL_SHA" >&2
      exit 1
    fi
  fi
fi

# Extract binary
tar -xzf "$TMP_DIR/$TARBALL" -C "$TMP_DIR" "$BINARY"
chmod +x "$TMP_DIR/$BINARY"

# Determine target install directory
if [ -z "${INSTALL_DIR:-}" ]; then
  if [ -w "/usr/local/bin" ]; then
    INSTALL_DIR="/usr/local/bin"
  elif command -v sudo >/dev/null 2>&1 && sudo -n true 2>/dev/null; then
    INSTALL_DIR="/usr/local/bin"
  else
    INSTALL_DIR="${HOME}/.local/bin"
  fi
fi

NEED_SUDO=0
if [ -d "$INSTALL_DIR" ]; then
  if [ ! -w "$INSTALL_DIR" ]; then
    NEED_SUDO=1
  fi
else
  if ! mkdir -p "$INSTALL_DIR" 2>/dev/null; then
    NEED_SUDO=1
  fi
fi

echo "Installing gocraft to ${INSTALL_DIR}..."

if [ "$NEED_SUDO" = "1" ]; then
  if ! command -v sudo >/dev/null 2>&1; then
    echo "Error: cannot write to $INSTALL_DIR and sudo is not available." >&2
    exit 1
  fi
  sudo mkdir -p "$INSTALL_DIR"
  sudo mv "$TMP_DIR/$BINARY" "$INSTALL_DIR/$BINARY"
else
  mkdir -p "$INSTALL_DIR"
  mv "$TMP_DIR/$BINARY" "$INSTALL_DIR/$BINARY"
fi

echo "Successfully installed gocraft ${TAG} to ${INSTALL_DIR}/gocraft"

# Check PATH
case ":$PATH:" in
  *":$INSTALL_DIR:"*) ;;
  *)
    echo ""
    echo "Notice: ${INSTALL_DIR} is not in your PATH."
    echo "Add it by adding this line to your ~/.bashrc or ~/.zshrc:"
    echo "  export PATH=\"${INSTALL_DIR}:\$PATH\""
    ;;
esac

echo ""
echo "Run 'gocraft' to launch the interactive project wizard."
