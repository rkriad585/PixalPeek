$ErrorActionPreference = "Stop"

$repo       = "rkriad585/PixalPeek"
$appName    = "PixalPeek"
$version    = "0.1.5-beta"
$issScript  = "build\windows\installer.iss"
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

$iscc = Get-Command ISCC -ErrorAction SilentlyContinue
if (-not $iscc) {
    # Check common Inno Setup install paths
    $isccPaths = @(
        "${env:ProgramFiles(x86)}\Inno Setup 6\ISCC.exe",
        "${env:ProgramFiles(x86)}\Inno Setup 5\ISCC.exe",
        "$env:ProgramFiles\Inno Setup 6\ISCC.exe",
        "$env:LOCALAPPDATA\Programs\Inno Setup 6\ISCC.exe"
    )
    $isccFound = $false
    foreach ($p in $isccPaths) {
        if (Test-Path $p) {
            $iscc = $p
            $isccFound = $true
            break
        }
    }
    if (-not $isccFound) {
        Write-Host "  Inno Setup not found. Install it (e.g. winget install JRSoftware.InnoSetup) and re-run." -ForegroundColor Yellow
        Write-Host "  https://jrsoftware.org/isdl.php" -ForegroundColor Gray
        exit 1
    }
} else {
    $iscc = $iscc.Source
}

Write-Host "  Go:     $(go version)" -ForegroundColor Gray
Write-Host "  Wails:  $(wails3 version 2>$null)" -ForegroundColor Gray
Write-Host "  Inno:   $iscc" -ForegroundColor Gray

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

# ── Build Inno Setup installer ─────────────────────────
Write-Host "[3/5] Building Inno Setup installer..." -ForegroundColor Cyan
if (Test-Path $issScript) {
    & $iscc /O"$outDir" /DAPP_VERSION=$version /DAPP_ARCH=amd64 $issScript
    if ($LASTEXITCODE -ne 0) { Write-Error "Inno Setup build failed"; exit 1 }

    $installer = Get-ChildItem -Path $outDir -Filter "pixalpeek-windows-*.exe" | Select-Object -First 1
    if ($installer) {
        $installerSize = [math]::Round($installer.Length / 1MB, 2)
        Write-Host "  Installer: $($installer.Name) ($installerSize MB)" -ForegroundColor Gray
    }
} else {
    Write-Host "  Inno Setup script not found, skipping installer" -ForegroundColor Yellow
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
