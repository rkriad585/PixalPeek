$ErrorActionPreference = "Stop"

$repo       = "rkriad585/PixalPeek"
$appName    = "PixalPeek"
$version    = "0.1.5-beta"
$nsisScript = "build\windows\installer.nsi"
$distDir    = "dist"
$outDir     = "dist\installers"

Write-Host "╔══════════════════════════════════════════════╗" -ForegroundColor Green
Write-Host "║       PIXALPEEK BUILDER (Windows)           ║" -ForegroundColor Green
Write-Host "╚══════════════════════════════════════════════╝" -ForegroundColor Green
Write-Host ""

# ── Check prerequisites ──────────────────────────────────
Write-Host "[0/5] Checking prerequisites..." -ForegroundColor Cyan

$go = Get-Command go -ErrorAction SilentlyContinue
if (-not $go) { Write-Error "Go not found. Install from https://go.dev/dl/"; exit 1 }

$wails = Get-Command wails3 -ErrorAction SilentlyContinue
if (-not $wails) {
    Write-Host "  Wails v3 CLI not found. Installing..." -ForegroundColor Yellow
    go install github.com/wailsapp/wails/v3/cmd/wails3@v3.0.0-beta.14
    if ($LASTEXITCODE -ne 0) { Write-Error "Failed to install wails3 CLI"; exit 1 }
}

$nsis = Get-Command makensis -ErrorAction SilentlyContinue
if (-not $nsis) {
    # Check common NSIS install paths
    $nsisPaths = @(
        "$env:ProgramFiles\NSIS\makensis.exe",
        "${env:ProgramFiles(x86)}\NSIS\makensis.exe",
        "$env:LOCALAPPDATA\Programs\NSIS\makensis.exe"
    )
    $nsisFound = $false
    foreach ($p in $nsisPaths) {
        if (Test-Path $p) {
            $nsis = $p
            $nsisFound = $true
            break
        }
    }
    if (-not $nsisFound) {
        Write-Host "  NSIS not found. Install from https://nsis.sourceforge.io/Download" -ForegroundColor Yellow
        Write-Host "  Downloading NSIS installer..." -ForegroundColor Cyan

        $nsisUrl = "https://sourceforge.net/projects/nsis/files/NSIS%203/3.10/nsis-3.10-setup.exe/download"
        $nsisSetup = Join-Path $env:TEMP "nsis-setup.exe"
        Invoke-WebRequest -Uri $nsisUrl -OutFile $nsisSetup -UseBasicParsing
        Write-Host "  Run NSIS installer, then re-run this script." -ForegroundColor Yellow
        Start-Process $nsisSetup
        exit 1
    }
} else {
    $nsis = $nsis.Source
}

Write-Host "  Go:     $(go version)" -ForegroundColor Gray
Write-Host "  Wails:  $(wails3 version 2>$null)" -ForegroundColor Gray
Write-Host "  NSIS:   $nsis" -ForegroundColor Gray

# ── Ensure directories ──────────────────────────────────
if (-not (Test-Path $distDir))  { New-Item -ItemType Directory -Path $distDir  -Force | Out-Null }
if (-not (Test-Path $outDir))   { New-Item -ItemType Directory -Path $outDir   -Force | Out-Null }

# ── Build with wails3 (handles frontend + bindings + icon embedding) ──
Write-Host "[1/5] Building with wails3 (windows/amd64)..." -ForegroundColor Cyan
$env:CGO_ENABLED = "1"
wails3 build
if ($LASTEXITCODE -ne 0) { Write-Error "wails3 build failed"; exit 1 }

# wails3 outputs to bin/
if (-not (Test-Path "$distDir")) { New-Item -ItemType Directory -Path $distDir -Force | Out-Null }
Copy-Item "bin\pixalpeek.exe" "$distDir\pixalpeek.exe" -Force

$exeSize = [math]::Round((Get-Item "$distDir\pixalpeek.exe").Length / 1MB, 2)
Write-Host "  Built: pixalpeek.exe ($exeSize MB)" -ForegroundColor Gray

# ── Build NSIS installer ────────────────────────────────
Write-Host "[3/5] Building NSIS installer..." -ForegroundColor Cyan
if (Test-Path $nsisScript) {
    & $nsis /DAPP_VERSION=$version $nsisScript
    if ($LASTEXITCODE -ne 0) { Write-Error "NSIS build failed"; exit 1 }

    $installer = Get-ChildItem -Path "." -Filter "pixalpeek-windows-*.exe" | Select-Object -First 1
    if ($installer) {
        Move-Item $installer.FullName -Destination $outDir -Force
        $installerSize = [math]::Round($installer.Length / 1MB, 2)
        Write-Host "  Installer: $($installer.Name) ($installerSize MB)" -ForegroundColor Gray
    }
} else {
    Write-Host "  NSIS script not found, skipping installer" -ForegroundColor Yellow
}

# ── Build Android APK ──────────────────────────────────
Write-Host "[4/5] Checking Android SDK..." -ForegroundColor Cyan
$androidHome = $env:ANDROID_HOME
if (-not $androidHome) { $androidHome = $env:ANDROID_SDK_ROOT }
if (-not $androidHome) {
    $androidHome = "$env:LOCALAPPDATA\Android\Sdk"
    if (-not (Test-Path $androidHome)) {
        $androidHome = "$env:USERPROFILE\AppData\Local\Android\Sdk"
    }
}

if (Test-Path $androidHome) {
    Write-Host "  Android SDK: $androidHome" -ForegroundColor Gray
    Write-Host "  To build APK, run: wails build -platform android/arm64 -o $outDir\pixalpeek-android-arm64" -ForegroundColor Yellow
} else {
    Write-Host "  Android SDK not found, skipping APK build" -ForegroundColor Yellow
    Write-Host "  Install Android SDK and run: wails build -platform android/arm64" -ForegroundColor Gray
}

# ── Summary ─────────────────────────────────────────────
Write-Host "[5/5] Build complete!" -ForegroundColor Cyan
Write-Host ""
Write-Host "Output files:" -ForegroundColor Green
Get-ChildItem -Path $outDir -File | ForEach-Object {
    $size = [math]::Round($_.Length / 1MB, 2)
    Write-Host "  $($_.Name) ($size MB)" -ForegroundColor Gray
}
Write-Host ""
Write-Host "Upload to GitHub Releases:" -ForegroundColor Cyan
Write-Host "  gh release upload v$version $outDir\*" -ForegroundColor Gray
