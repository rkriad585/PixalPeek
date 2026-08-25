$ErrorActionPreference = "Stop"

$repo = "rkriad585/PixalPeek"
$apiUrl = "https://api.github.com/repos/$repo/releases/latest"

Write-Host "╔══════════════════════════════════════╗" -ForegroundColor Green
Write-Host "║         PIXALPEEK INSTALLER          ║" -ForegroundColor Green
Write-Host "╚══════════════════════════════════════╝" -ForegroundColor Green
Write-Host ""

Write-Host "[1/3] Fetching latest release..." -ForegroundColor Cyan
$headers = @{}
if ($env:GITHUB_TOKEN) {
    $headers["Authorization"] = "token $env:GITHUB_TOKEN"
}
$release = Invoke-RestMethod -Uri $apiUrl -Headers $headers
$tag = $release.tag_name
Write-Host "  Latest version: $tag" -ForegroundColor Gray

# Match pixalpeek-Windows-amd64.exe or pixalpeek-windows-x64.exe etc.
$asset = $release.assets | Where-Object {
    $_.name -match "pixalpeek-.*[Ww]indows.*\.exe$"
} | Select-Object -First 1

if (-not $asset) {
    Write-Error "No Windows installer found. Expected pattern: pixalpeek-Windows-<arch>.exe"
    Write-Host "  Available assets:" -ForegroundColor Yellow
    $release.assets | ForEach-Object { Write-Host "    $($_.name)" -ForegroundColor Gray }
    exit 1
}

Write-Host "[2/3] Downloading $($asset.name)..." -ForegroundColor Cyan
$tempFile = Join-Path $env:TEMP $asset.name
$dlHeaders = @{}
if ($env:GITHUB_TOKEN) {
    $dlHeaders["Authorization"] = "token $env:GITHUB_TOKEN"
}

$progressPreference = 'SilentlyContinue'
Invoke-WebRequest -Uri $asset.browser_download_url -OutFile $tempFile -Headers $dlHeaders -UseBasicParsing
$progressPreference = 'Continue'

$fileSize = [math]::Round((Get-Item $tempFile).Length / 1MB, 2)
Write-Host "  Downloaded: $fileSize MB" -ForegroundColor Gray

Write-Host "[3/3] Launching installer..." -ForegroundColor Cyan
Write-Host ""
Write-Host "  Running: $tempFile" -ForegroundColor Gray
Write-Host ""

Start-Process -FilePath $tempFile -Wait

Write-Host ""
Write-Host "✔ PixalPeek $tag installed successfully" -ForegroundColor Green
Write-Host ""
Write-Host "Run 'pixalpeek --version' to verify." -ForegroundColor Cyan
