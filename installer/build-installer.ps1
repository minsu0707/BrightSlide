$ErrorActionPreference = 'Stop'

$root = Split-Path -Parent $PSScriptRoot
$stage = Join-Path $PSScriptRoot 'stage'

$iscc = @(
    "$env:LOCALAPPDATA\Programs\Inno Setup 6\ISCC.exe",
    "${env:ProgramFiles(x86)}\Inno Setup 6\ISCC.exe",
    "$env:ProgramFiles\Inno Setup 6\ISCC.exe"
) | Where-Object { Test-Path $_ } | Select-Object -First 1
if (-not $iscc) { throw 'Inno Setup 6 not found. Install it with: winget install -e --id JRSoftware.InnoSetup --source winget' }

$go = (Get-Command go -ErrorAction SilentlyContinue).Source
if (-not $go) { $go = "$env:ProgramFiles\Go\bin\go.exe" }
if (-not (Test-Path $go)) { throw 'Go not found. Install it with: winget install -e --id GoLang.Go --source winget' }

$version = (Get-Content (Join-Path $root 'VERSION') -Raw).Trim()

if (Test-Path $stage) { Remove-Item $stage -Recurse -Force }
New-Item -ItemType Directory -Path $stage | Out-Null
$env:CGO_ENABLED = '0'
# The icon resource is rsrc_windows_amd64.syso; any other target would silently build without it.
$env:GOOS = 'windows'
$env:GOARCH = 'amd64'
Push-Location $root
try {
    & $go build -trimpath -ldflags '-s -w -H=windowsgui' -o (Join-Path $stage 'BrightSlide.exe') .
    if ($LASTEXITCODE -ne 0) { throw 'go build failed' }
} finally { Pop-Location }

& $iscc "/DAppVersion=$version" "/DStageDir=$stage" (Join-Path $PSScriptRoot 'BrightSlide.iss')
if ($LASTEXITCODE -ne 0) { throw 'ISCC failed' }

Copy-Item (Join-Path $stage 'BrightSlide.exe') (Join-Path $PSScriptRoot "Output\BrightSlide-$version-win-x64.exe") -Force
Write-Host "Built: $(Join-Path $PSScriptRoot "Output\BrightSlide-Setup-$version.exe")"
Write-Host "Built: $(Join-Path $PSScriptRoot "Output\BrightSlide-$version-win-x64.exe")"
