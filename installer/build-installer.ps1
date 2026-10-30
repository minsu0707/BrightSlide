$ErrorActionPreference = 'Stop'

$root = Split-Path -Parent $PSScriptRoot
$stage = Join-Path $PSScriptRoot 'stage'

$iscc = @(
    "$env:LOCALAPPDATA\Programs\Inno Setup 6\ISCC.exe",
    "${env:ProgramFiles(x86)}\Inno Setup 6\ISCC.exe",
    "$env:ProgramFiles\Inno Setup 6\ISCC.exe"
) | Where-Object { Test-Path $_ } | Select-Object -First 1
if (-not $iscc) { throw 'Inno Setup 6 not found. Install it with: winget install -e --id JRSoftware.InnoSetup --source winget' }

$version = ([xml](Get-Content (Join-Path $root 'BrightSlide.csproj'))).Project.PropertyGroup.Version | Where-Object { $_ } | Select-Object -First 1

if (Test-Path $stage) { Remove-Item $stage -Recurse -Force }
dotnet publish (Join-Path $root 'BrightSlide.csproj') -c Release -r win-x64 --self-contained true `
    -p:PublishSingleFile=true -p:IncludeNativeLibrariesForSelfExtract=true -o $stage
if ($LASTEXITCODE -ne 0) { throw 'dotnet publish failed' }

& $iscc "/DAppVersion=$version" "/DStageDir=$stage" (Join-Path $PSScriptRoot 'BrightSlide.iss')
if ($LASTEXITCODE -ne 0) { throw 'ISCC failed' }

Write-Host "Built: $(Join-Path $PSScriptRoot "Output\BrightSlide-Setup-$version.exe")"
