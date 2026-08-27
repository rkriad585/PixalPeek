#!/usr/bin/env bash
set -e

# ── Configuration ──────────────────────────────────────
APP_NAME="pixalpeek"
APP_DISPLAY="PixalPeek"
VERSION="0.1.5-beta"
REPO="rkriad585/PixalPeek"
DIST_DIR="dist"
OUT_DIR="dist/installers"
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"

RED='\033[0;31m'
GREEN='\033[0;32m'
CYAN='\033[0;36m'
YELLOW='\033[1;33m'
GRAY='\033[0;90m'
NC='\033[0m'

log()  { echo -e "${CYAN}$1${NC}"; }
ok()   { echo -e "${GREEN}$1${NC}"; }
warn() { echo -e "${YELLOW}$1${NC}"; }
err()  { echo -e "${RED}$1${NC}" >&2; exit 1; }

echo -e "${GREEN}╔══════════════════════════════════════════════╗${NC}"
echo -e "${GREEN}║     PIXALPEEK BUILDER (Cross-platform)      ║${NC}"
echo -e "${GREEN}╚══════════════════════════════════════════════╝${NC}"
echo ""

# ── Parse arguments ─────────────────────────────────────
TARGETS=""
ANDROID_ONLY=false
for arg in "$@"; do
    case $arg in
        --all)       TARGETS="linux macos android" ;;
        --linux)     TARGETS="$TARGETS linux" ;;
        --macos)     TARGETS="$TARGETS macos" ;;
        --android)   TARGETS="$TARGETS android" ;;
        --android-only) ANDROID_ONLY=true; TARGETS="android" ;;
        -h|--help)
            echo "Usage: ./build.sh [--all|--linux|--macos|--android]"
            exit 0
            ;;
    esac
done

# Auto-detect platform if no target specified
if [ -z "$TARGETS" ] && [ "$ANDROID_ONLY" = false ]; then
    case "$(uname -s)" in
        Linux*)  TARGETS="linux" ;;
        Darwin*) TARGETS="macos" ;;
        *)       err "Unsupported OS: $(uname -s)" ;;
    esac
fi

# ── Check prerequisites ────────────────────────────────
log "[0/6] Checking prerequisites..."

if ! command -v go &>/dev/null; then err "Go not found. Install from https://go.dev/dl/"; fi
if ! command -v node &>/dev/null; then err "Node.js not found. Install from https://nodejs.org/"; fi
if ! command -v npm &>/dev/null; then err "npm not found."; fi

if ! command -v wails3 &>/dev/null; then
    warn "  Wails v3 CLI not found. Installing..."
    go install github.com/wailsapp/wails/v3/cmd/wails3@v3.0.0-beta.14
    command -v wails3 &>/dev/null || err "Failed to install wails3 CLI"
fi

echo -e "  Go:     $(go version)"  "$GRAY"
echo -e "  Wails:  $(wails3 version 2>/dev/null)" "$GRAY"
echo -e "  Node:   $(node --version)" "$GRAY"
echo ""

mkdir -p "$DIST_DIR" "$OUT_DIR"

# ── Build with wails3 (handles frontend + bindings + icon embedding) ──
log "[1/6] Building with wails3 (host)..."
wails3 build
ok "  Built: bin/$APP_NAME ($(du -h "bin/$APP_NAME" | cut -f1))"
echo ""

# ══════════════════════════════════════════════════════
# LINUX
# ══════════════════════════════════════════════════════
build_linux() {
    log "[3/6] Building Linux packages..."
    local ARCH
    local OS_NAME
    ARCH="$(uname -m)"
    case "$ARCH" in
        x86_64)  ARCH="amd64"; OS_NAME="linux" ;;
        aarch64) ARCH="arm64"; OS_NAME="arm-linux" ;;
    esac

    # ── Cross-compile if on macOS building for Linux
    if [ "$(uname -s)" = "Linux" ]; then
        # Already built above by wails3
        BINARY="bin/$APP_NAME"
    else
        log "  Cross-compiling linux/$ARCH..."
        CGO_ENABLED=0 GOOS=linux GOARCH="$ARCH" \
            go build -tags desktop,production -ldflags "-s -w" \
            -o "$DIST_DIR/$APP_NAME-linux" .
        BINARY="$DIST_DIR/$APP_NAME-linux"
    fi

    # ── .deb package ──────────────────────────────────
    log "  Building .deb package..."
    local DEB_DIR="$DIST_DIR/deb-pkg"
    local DEB_OUT="$OUT_DIR/${APP_NAME}-${OS_NAME}-${ARCH}.deb"
    rm -rf "$DEB_DIR"
    mkdir -p "$DEB_DIR/DEBIAN" "$DEB_DIR/usr/local/bin"

    cp "$BINARY" "$DEB_DIR/usr/local/bin/$APP_NAME"
    chmod 755 "$DEB_DIR/usr/local/bin/$APP_NAME"

    # Generate control file
    local SIZE_KB
    SIZE_KB=$(du -k "$BINARY" | cut -f1)
    cat > "$DEB_DIR/DEBIAN/control" <<EOF
Package: $APP_NAME
Version: ${VERSION//-/.}
Section: utils
Priority: optional
Architecture: $ARCH
Maintainer: rkriad585 <rkriad585@gmail.com>
Homepage: https://rkriad585.github.io/PixalPeek
Installed-Size: $SIZE_KB
Description: QR code scanner & generator
 PixalPeek is a dual-mode QR code tool with a GUI (Wails v3)
 and a full-featured CLI. Supports scanning, generating, batch
 processing, camera scan, clipboard decode, and more.
EOF

    cat > "$DEB_DIR/DEBIAN/postinst" <<'POSTINST'
#!/bin/sh
echo "PixalPeek installed to /usr/local/bin/pixalpeek"
echo "Run 'pixalpeek --version' to verify."
POSTINST
    chmod 755 "$DEB_DIR/DEBIAN/postinst"

    cat > "$DEB_DIR/DEBIAN/postrm" <<'POSTRM'
#!/bin/sh
echo "PixalPeek removed."
POSTRM
    chmod 755 "$DEB_DIR/DEBIAN/postrm"

    dpkg-deb --build "$DEB_DIR" "$DEB_OUT" 2>/dev/null && \
        ok "  .deb: $DEB_OUT ($(du -h "$DEB_OUT" | cut -f1))" || \
        warn "  dpkg-deb not available, skipping .deb"

    # ── .rpm package ──────────────────────────────────
    log "  Building .rpm package..."
    local RPM_OUT="$OUT_DIR/${APP_NAME}-${OS_NAME}-${ARCH}.rpm"
    if command -v rpmbuild &>/dev/null; then
        local RPM_DIR="$DIST_DIR/rpm-build"
        mkdir -p "$RPM_DIR"/{BUILD,RPMS,SOURCES,SPECS,SRPMS}

        local VERSION_CLEAN="${VERSION//-/.}"
        cat > "$RPM_DIR/SPECS/${APP_NAME}.spec" <<EOF
Name:           $APP_NAME
Version:        $VERSION_CLEAN
Release:        1%{?dist}
Summary:        QR code scanner & generator
License:        MIT
URL:            https://rkriad585.github.io/PixalPeek
BuildArch:      $ARCH

%description
PixalPeek is a dual-mode QR code tool with a GUI and full CLI.

%install
mkdir -p %{buildroot}/usr/local/bin
cp $SCRIPT_DIR/$BINARY %{buildroot}/usr/local/bin/$APP_NAME
chmod 755 %{buildroot}/usr/local/bin/$APP_NAME

%files
/usr/local/bin/$APP_NAME

%changelog
* Mon Aug 25 2026 rkriad585 - $VERSION_CLEAN-1
- Initial RPM release
EOF

        rpmbuild -bb "$RPM_DIR/SPECS/${APP_NAME}.spec" --define "_topdir $RPM_DIR" 2>/dev/null && {
            local RPM_FILE
            RPM_FILE=$(find "$RPM_DIR/RPMS" -name "*.rpm" | head -1)
            if [ -n "$RPM_FILE" ]; then
                cp "$RPM_FILE" "$RPM_OUT"
                ok "  .rpm: $RPM_OUT ($(du -h "$RPM_OUT" | cut -f1))"
            fi
        } || warn "  rpmbuild failed, skipping .rpm"
    else
        warn "  rpmbuild not found, skipping .rpm (install: dnf install rpm-build)"
    fi

    # Cleanup cross-compile binary
    [ -f "$DIST_DIR/$APP_NAME-linux" ] && rm -f "$DIST_DIR/$APP_NAME-linux"
}

# ══════════════════════════════════════════════════════
# macOS
# ══════════════════════════════════════════════════════
build_macos() {
    log "[3/6] Building macOS package..."
    local ARCH
    ARCH="$(uname -m)"

    local BINARY="$DIST_DIR/${APP_NAME}-darwin-${ARCH}"
    local DMG_OUT="$OUT_DIR/${APP_NAME}-darwin-${ARCH}.dmg"

    log "  Building for darwin/$ARCH..."
    CGO_ENABLED=0 GOOS=darwin GOARCH="$ARCH" \
        go build -tags desktop,production -ldflags "-s -w" \
        -o "$BINARY" .

    ok "  Binary: $BINARY ($(du -h "$BINARY" | cut -f1))"

    if command -v hdiutil &>/dev/null; then
        log "  Creating .dmg..."

        local APP_DIR="$DIST_DIR/PixalPeek.app"
        local STAGING="$DIST_DIR/dmg-staging"
        rm -rf "$APP_DIR" "$STAGING"
        mkdir -p "$APP_DIR/Contents/MacOS" "$APP_DIR/Contents/Resources" "$STAGING"

        cp "$BINARY" "$APP_DIR/Contents/MacOS/$APP_NAME"
        chmod 755 "$APP_DIR/Contents/MacOS/$APP_NAME"
        cp "$SCRIPT_DIR/build/appicon.png" "$APP_DIR/Contents/Resources/appicon.png" 2>/dev/null || true

        # Copy Info.plist
        local PLIST="$SCRIPT_DIR/build/macos/Info.plist"
        if [ -f "$PLIST" ]; then
            cp "$PLIST" "$APP_DIR/Contents/Info.plist"
        else
            cat > "$APP_DIR/Contents/Info.plist" <<PLIST
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>CFBundleExecutable</key><string>$APP_NAME</string>
    <key>CFBundleName</key><string>PixalPeek</string>
    <key>CFBundleIdentifier</key><string>io.github.rkriad585.pixalpeek</string>
    <key>CFBundleVersion</key><string>$VERSION</string>
    <key>CFBundlePackageType</key><string>APPL</string>
</dict>
</plist>
PLIST
        fi

        cp -R "$APP_DIR" "$STAGING/"

        hdiutil create -volname "PixalPeek" -srcfolder "$STAGING" -ov -format UDZO "$DMG_OUT" 2>/dev/null
        ok "  .dmg: $DMG_OUT ($(du -h "$DMG_OUT" | cut -f1))"

        rm -rf "$APP_DIR" "$STAGING"
    else
        warn "  hdiutil not available (not on macOS), skipping .dmg"
        ok "  Binary saved: $BINARY"
    fi
}

# ══════════════════════════════════════════════════════
# Android
# ══════════════════════════════════════════════════════
build_android() {
    log "[3/6] Building Android APK..."

    if ! command -v wails &>/dev/null; then
        warn "  Wails CLI not found. Install: go install github.com/wailsapp/wails/v3/cmd/wails@latest"
        return
    fi

    local ANDROID_HOME="${ANDROID_HOME:-$ANDROID_SDK_ROOT}"
    if [ -z "$ANDROID_HOME" ]; then
        ANDROID_HOME="$HOME/Library/Android/sdk"
        [ ! -d "$ANDROID_HOME" ] && ANDROID_HOME="$HOME/Android/Sdk"
    fi

    if [ ! -d "$ANDROID_HOME" ]; then
        warn "  Android SDK not found at $ANDROID_HOME"
        warn "  Set ANDROID_HOME or install Android SDK"
        return
    fi
    ok "  Android SDK: $ANDROID_HOME"

    # Build for both arm64 and armeabi-v7a
    for ARCH in arm64 armeabi-v7a; do
        local APK_OUT="$OUT_DIR/${APP_NAME}-android-${ARCH}"
        log "  Building for android/$ARCH..."

        wails build -platform "android/$ARCH" -o "$APK_OUT" 2>/dev/null && {
            ok "  .apk: $APK_OUT ($(du -h "$APK_OUT" | cut -f1))"
        } || {
            warn "  android/$ARCH build failed (check Android SDK/NDK setup)"
        }
    done
}

# ── Run builds ──────────────────────────────────────────
for target in $TARGETS; do
    case $target in
        linux)   build_linux ;;
        macos)   build_macos ;;
        android) build_android ;;
    esac
done

# ── Summary ─────────────────────────────────────────────
echo ""
log "[4/6] Build complete!"
echo ""
echo -e "${GREEN}Output files:${NC}"
if [ -d "$OUT_DIR" ]; then
    find "$OUT_DIR" -maxdepth 1 -type f | sort | while read -r f; do
        echo -e "  $(basename "$f") ${GRAY}($(du -h "$f" | cut -f1))${NC}"
    done
fi
echo ""
echo -e "${CYAN}Upload to GitHub Releases:${NC}"
echo -e "  ${GRAY}gh release upload v$VERSION $OUT_DIR/*${NC}"
