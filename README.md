# PIXALPEEK

Dot-matrix QR code scanner & generator — one Go binary, dual personality:

- **GUI** (Wails v3): scan / generate / history / settings, system tray, camera scan
- **CLI**: full-featured command-line QR tool (scriptable, JSON output)

## Quick start

```powershell
task build          # icon -> bindings -> frontend -> pixalpeek.exe
.\pixalpeek.exe     # launch GUI (no args)
```

Requires: Go 1.24+, Node 18+, [wails3 v3.0.0-beta.2](https://github.com/wailsapp/wails) CLI.

## CLI usage

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
| `-s/--silent`, `-v/--version`, `-h/--help` | misc |

### Examples

```powershell
pixalpeek -g "https://example.com" -o link.png --shape dot --fg "#00ff99"
pixalpeek -qr screenshot.png --multi -o results.json
pixalpeek --clipboard
pixalpeek --batch codes.csv --zip -s
pixalpeek -g "WIFI:T:WPA;S:HomeNet;P:pass;;" -t wifi
```

Exit codes: `0` ok · `1` general · `2` args · `3` no QR found · `4` IO error · `5` unsupported image.

## Architecture

```text
main.go                 entry: CLI mode vs GUI mode
internal/qrengine       pure engine: detect/build/decode/encode/svg/pdf/batch
internal/cli            flags, validation, formatters, exit codes
internal/service        bound API: history, settings, presets, worker bridge
internal/headless       hidden-window workers (camera / clipboard)
frontend                React + TS UI (vite), generated TS bindings
tools/icongen           generates build/appicon.png
```

- Decoding: `makiuchi-d/gozxing`; encoding: `skip2/go-qrcode`
- History auto-prunes to ~200 entries; pinned entries survive (`★`)
- Settings/history stored via OS config dir (`%APPDATA%/pixalpeek` on Windows)

## Dev loop

```powershell
task dev        # wails3 dev (hot reload)
go test ./internal/qrengine   # engine tests incl. logo/multi-code round-trips
```

GUI smoke test without display interaction: `$env:PIXALPEEK_AUTOTEST_MS="3000"; .\pixalpeek.exe`
