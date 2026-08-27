<p align="center">
  <img src="https://raw.githubusercontent.com/rkriad585/PixalPeek/refs/heads/main/logo/logo.svg" width="120" alt="PixalPeek Logo" />
</p>

<h1 align="center">PIXALPEEK</h1>

<p align="center">
  <strong>QR code scanner & generator</strong><br/>
  One Go binary, dual personality — GUI + CLI
</p>

<p align="center">
  <a href="https://github.com/rkriad585/PixalPeek/releases"><img src="https://img.shields.io/github/v/release/rkriad585/PixalPeek?include_prereleases&label=version&color=00ff99" alt="Version" /></a>
  <a href="https://github.com/rkriad585/PixalPeek/blob/main/LICENSE"><img src="https://img.shields.io/github/license/rkriad585/PixalPeek?color=00ff99" alt="License" /></a>
  <a href="https://github.com/rkriad585/PixalPeek/stargazers"><img src="https://img.shields.io/github/stars/rkriad585/PixalPeek?color=00ff99" alt="Stars" /></a>
  <a href="https://github.com/rkriad585/PixalPeek/issues"><img src="https://img.shields.io/github/issues/rkriad585/PixalPeek?color=ff5566" alt="Issues" /></a>
</p>

<p align="center">
  <a href="https://rkriad585.github.io/PixalPeek">Documentation</a> ·
  <a href="https://github.com/rkriad585/PixalPeek/releases">Releases</a> ·
  <a href="https://github.com/rkriad585/PixalPeek/issues">Report Issue</a>
</p>

---

## About

**PixalPeek** is a dual-mode QR code tool built with Go and React:

- **GUI** (Wails v3): Scan, generate, history, settings, camera scan, system tray — terminal-aesthetic UI with green-on-dark theme
- **CLI**: Full-featured command-line QR tool (scriptable, JSON output, batch processing)

## Installation

### One-line Installers

#### Windows (PowerShell)

```powershell
irm https://raw.githubusercontent.com/rkriad585/PixalPeek/main/installer.ps1 | iex
```

#### Linux / macOS

```bash
curl -fsSL https://raw.githubusercontent.com/rkriad585/PixalPeek/main/installer.sh | sh
```

### Download

Grab the latest binary from [Releases](https://github.com/rkriad585/PixalPeek/releases).

### Build from Source

```powershell
# Requires: Go 1.24+, Node 18+, Wails v3 CLI
task build
.\pixalpeek.exe
```

## Quick Start

```powershell
# Launch GUI (no args)
.\pixalpeek.exe

# CLI: decode a QR code
pixalpeek -qr image.png

# CLI: generate a QR code
pixalpeek -g "hello" -o out.png
```

## CLI Usage

```text
pixalpeek -qr <image> [options]        decode QR code(s) from an image
pixalpeek -g <content> [options]       generate a QR code image
pixalpeek --batch <csv|json> [options] batch-generate from a list
```

| Flag | Description |
|------|-------------|
| `-o, --output` | output path (`-` = stdout JSON for scans) |
| `--multi` | detect all codes in one image |
| `--size N`, `--ecc L\|M\|Q\|H` | generator sizing / error correction |
| `--fg #hex`, `--bg #hex` | colors |
| `--shape square\|rounded\|dot` | module style |
| `--logo file.png` | centered logo (forces ECC H) |
| `--format png\|jpg\|svg\|pdf` | output format |
| `--margin 0-8` | quiet zone |
| `--camera` | scan live from webcam |
| `--clipboard` | decode image on clipboard |
| `--scan-dir <dir>` | scan all images in a directory |
| `-s/--silent`, `-v/--version`, `-h/--help` | misc |

### Examples

```powershell
pixalpeek -g "https://example.com" -o link.png --shape dot --fg "#00ff99"
pixalpeek -qr screenshot.png --multi -o results.json
pixalpeek --clipboard
pixalpeek --batch codes.csv --zip -s
pixalpeek -g "WIFI:T:WPA;S:HomeNet;P:pass;;" -t wifi
pixalpeek --scan-dir ./images
```

Exit codes: `0` ok · `1` general · `2` args · `3` no QR found · `4` IO error · `5` unsupported image.

## Architecture

```text
main.go                 entry: CLI mode vs GUI mode
internal/qrengine       pure engine: detect/build/decode/encode/svg/pdf/batch
internal/cli            flags, validation, formatters, exit codes
internal/service        bound API: history, settings, presets, worker bridge
internal/headless       hidden-window workers (camera / clipboard)
internal/storage        SQLite + TOML config
internal/cache          in-memory TTL cache
internal/watcher        fsnotify directory watcher
frontend                React + TS UI (Vite), generated TS bindings
```

- Decoding: [makiuchi-d/gozxing](https://github.com/makiuchi-d/gozxing); encoding: [skip2/go-qrcode](https://github.com/skip2/go-qrcode)
- History auto-prunes to ~200 entries; pinned entries survive
- Settings stored via OS config dir (`~/.config/neostore/pixalpeek/config.toml`)

## Dev Loop

```powershell
task dev                          # wails3 dev (hot reload)
go test ./internal/qrengine       # engine tests
```

GUI smoke test: `$env:PIXALPEEK_AUTOTEST_MS="3000"; .\pixalpeek.exe`

## Author

**rkriad585** — [GitHub](https://github.com/rkriad585) · [Portfolio](https://rkriad585.github.io) · [rkriad585@gmail.com](mailto:rkriad585@gmail.com)

<p align="center">
  <a href="https://github.com/rkriad585">
    <img src="https://avatars.githubusercontent.com/u/107482047?v=4" width="64" style="border-radius:50%" alt="rkriad585" />
  </a>
</p>

## License

[MIT](LICENSE) © 2026 [rkriad585](https://github.com/rkriad585)
