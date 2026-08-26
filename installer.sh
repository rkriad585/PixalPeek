#!/usr/bin/env sh
set -e

REPO="rkriad585/PixalPeek"
API_URL="https://api.github.com/repos/$REPO/releases/latest"

echo "╔══════════════════════════════════════╗"
echo "║         PIXALPEEK INSTALLER          ║"
echo "╚══════════════════════════════════════╝"
echo ""

# Detect OS and arch
OS="$(uname -s)"
ARCH="$(uname -m)"

case "$OS" in
    Linux*)
        case "$ARCH" in
            x86_64|amd64)
                OS_NAME="linux"
                ARCH_NAME="amd64"
                ;;
            aarch64|arm64)
                OS_NAME="arm-linux"
                ARCH_NAME="arm64"
                ;;
            *)             echo "Unsupported architecture: $ARCH"; exit 1 ;;
        esac
        ;;
    Darwin*)
        OS_NAME="darwin"
        case "$ARCH" in
            x86_64|amd64)  ARCH_NAME="amd64" ;;
            arm64|aarch64) ARCH_NAME="arm64" ;;
            *)             echo "Unsupported architecture: $ARCH"; exit 1 ;;
        esac
        ;;
    *)
        echo "Unsupported OS: $OS"
        exit 1
        ;;
esac

echo "[1/4] Detecting platform: ${OS_NAME} ${ARCH_NAME}"

echo "[2/4] Fetching latest release..."
AUTH_HEADER=""
if [ -n "$GITHUB_TOKEN" ]; then
    AUTH_HEADER="-H \"Authorization: token $GITHUB_TOKEN\""
fi

RELEASE_JSON=$(curl -fsSL "$API_URL" ${GITHUB_TOKEN:+-H "Authorization: token $GITHUB_TOKEN"})
TAG=$(echo "$RELEASE_JSON" | grep '"tag_name"' | head -1 | sed 's/.*"tag_name": *"\([^"]*\)".*/\1/')
echo "  Latest version: $TAG"

# Find matching asset: pixalpeek-linux-amd64.deb, pixalpeek-darwin-arm64.dmg, etc.
ASSET_URL=""
ASSET_NAME=""

if [ "$OS_NAME" = "linux" ] || [ "$OS_NAME" = "arm-linux" ]; then
    # Try .deb first, then .rpm, then .tar.gz
    for ext in "deb" "rpm" "tar.gz" "tgz"; do
        MATCH=$(echo "$RELEASE_JSON" | grep '"browser_download_url"' | grep -i "pixalpeek-${OS_NAME}-${ARCH_NAME}.*\.${ext}" | head -1 | sed 's/.*"browser_download_url": *"\([^"]*\)".*/\1/')
        if [ -n "$MATCH" ]; then
            ASSET_URL="$MATCH"
            ASSET_NAME=$(basename "$MATCH")
            break
        fi
    done
elif [ "$OS_NAME" = "darwin" ]; then
    # Try .dmg first, then .pkg, then .tar.gz
    for ext in "dmg" "pkg" "tar.gz" "tgz"; do
        MATCH=$(echo "$RELEASE_JSON" | grep '"browser_download_url"' | grep -i "pixalpeek-darwin-${ARCH_NAME}.*\.${ext}" | head -1 | sed 's/.*"browser_download_url": *"\([^"]*\)".*/\1/')
        if [ -n "$MATCH" ]; then
            ASSET_URL="$MATCH"
            ASSET_NAME=$(basename "$MATCH")
            break
        fi
    done
fi

if [ -z "$ASSET_URL" ]; then
    echo "Error: No installer found for ${OS_NAME}-${ARCH_NAME}"
    echo ""
    echo "Expected pattern: pixalpeek-${OS_NAME}-${ARCH_NAME}.{deb,rpm,dmg}"
    echo ""
    echo "Available assets:"
    echo "$RELEASE_JSON" | grep '"name"' | grep -v '"name": "repo"' | head -20 | sed 's/.*"name": *"\([^"]*\)".*/  \1/'
    exit 1
fi

echo "  Found: $ASSET_NAME"

echo "[3/4] Downloading..."
TMPDIR=$(mktemp -d)
trap 'rm -rf "$TMPDIR"' EXIT

curl -fSL "$ASSET_URL" -o "$TMPDIR/$ASSET_NAME" \
    ${GITHUB_TOKEN:+-H "Authorization: token $GITHUB_TOKEN"}

FILE_SIZE=$(du -h "$TMPDIR/$ASSET_NAME" | cut -f1)
echo "  Downloaded: $FILE_SIZE"

echo "[4/4] Installing..."

case "$ASSET_NAME" in
    *.deb)
        echo "  Installing .deb package..."
        sudo dpkg -i "$TMPDIR/$ASSET_NAME"
        sudo apt-get install -f -y 2>/dev/null || true
        ;;
    *.rpm)
        echo "  Installing .rpm package..."
        if command -v dnf >/dev/null 2>&1; then
            sudo dnf install -y "$TMPDIR/$ASSET_NAME"
        elif command -v yum >/dev/null 2>&1; then
            sudo yum install -y "$TMPDIR/$ASSET_NAME"
        elif command -v rpm >/dev/null 2>&1; then
            sudo rpm -ivh "$TMPDIR/$ASSET_NAME"
        else
            echo "Error: No package manager found (need dnf, yum, or rpm)"
            exit 1
        fi
        ;;
    *.dmg)
        echo "  Mounting .dmg..."
        MOUNT_DIR="/Volumes/PixalPeek-Install"
        hdiutil attach "$TMPDIR/$ASSET_NAME" -nobrowse -quiet -mountpoint "$MOUNT_DIR"

        # Find .app or .pkg inside the dmg
        APP=$(find "$MOUNT_DIR" -name "*.app" -maxdepth 1 | head -1)
        PKG=$(find "$MOUNT_DIR" -name "*.pkg" -maxdepth 1 | head -1)

        if [ -n "$APP" ]; then
            echo "  Copying app to /Applications..."
            cp -R "$APP" /Applications/
            echo "  App installed to /Applications/$(basename "$APP")"
        elif [ -n "$PKG" ]; then
            echo "  Running package installer..."
            sudo installer -pkg "$PKG" -target /
        else
            echo "  Copying files..."
            cp -R "$MOUNT_DIR"/* /Applications/ 2>/dev/null || true
        fi

        hdiutil detach "$MOUNT_DIR" -quiet 2>/dev/null || true
        ;;
    *.pkg)
        echo "  Running package installer..."
        sudo installer -pkg "$TMPDIR/$ASSET_NAME" -target /
        ;;
    *.tar.gz|*.tgz)
        echo "  Extracting..."
        tar -xzf "$TMPDIR/$ASSET_NAME" -C "$TMPDIR"
        BIN=$(find "$TMPDIR" -name "pixalpeek" -type f | head -1)
        if [ -n "$BIN" ]; then
            chmod +x "$BIN"
            sudo cp "$BIN" /usr/local/bin/pixalpeek
            echo "  Installed to /usr/local/bin/pixalpeek"
        else
            echo "Error: pixalpeek binary not found in archive"
            exit 1
        fi
        ;;
    *)
        echo "Error: Unsupported file type: $ASSET_NAME"
        exit 1
        ;;
esac

echo ""
echo "✔ PixalPeek $TAG installed successfully"
echo ""
echo "Run 'pixalpeek --version' to verify."
