# gocraft installer for Windows PowerShell
# Usage: irm https://raw.githubusercontent.com/VJ-2303/gocraft/main/scripts/install.ps1 | iex

param(
    [string]$Version = ""
)

$ErrorActionPreference = "Stop"

$repo = "VJ-2303/gocraft"
$binary = "gocraft.exe"

# Detect Architecture
$arch = if ($env:PROCESSOR_ARCHITECTURE -eq "ARM64") { "arm64" } else { "amd64" }

# Resolve Version
if (-not $Version) {
    if ($env:VERSION) {
        $Version = $env:VERSION
    } else {
        try {
            $release = Invoke-RestMethod -Uri "https://api.github.com/repos/$repo/releases/latest" -UseBasicParsing -Headers @{ "User-Agent" = "gocraft-installer" }
            $Version = $release.tag_name
        } catch {
            Write-Error "Failed to fetch latest gocraft release: $_"
            exit 1
        }
    }
}

if (-not $Version) {
    Write-Error "Could not determine gocraft release version."
    exit 1
}

$tag = $Version
$cleanVersion = $tag.TrimStart("v")
$zipName = "gocraft_${cleanVersion}_windows_${arch}.zip"
$downloadUrl = "https://github.com/$repo/releases/download/$tag/$zipName"

Write-Host "Downloading gocraft $tag for Windows ($arch)..." -ForegroundColor Cyan

$tempDir = Join-Path ([System.IO.Path]::GetTempPath()) ([System.Guid]::NewGuid().ToString())
New-Item -ItemType Directory -Path $tempDir -Force | Out-Null

try {
    $zipPath = Join-Path $tempDir $zipName
    Invoke-WebRequest -Uri $downloadUrl -OutFile $zipPath -UseBasicParsing

    $installDir = Join-Path $env:LOCALAPPDATA "Programs\gocraft\bin"
    if (-not (Test-Path $installDir)) {
        New-Item -ItemType Directory -Path $installDir -Force | Out-Null
    }

    Write-Host "Extracting to $installDir..." -ForegroundColor Cyan
    Expand-Archive -Path $zipPath -DestinationPath $tempDir -Force
    Copy-Item -Path (Join-Path $tempDir $binary) -Destination (Join-Path $installDir $binary) -Force

    # Ensure install directory is in User PATH
    $userPath = [Environment]::GetEnvironmentVariable("Path", "User")
    if ($userPath -split ";" -notcontains $installDir) {
        Write-Host "Adding $installDir to User PATH..." -ForegroundColor Cyan
        $newPath = if ($userPath) { "$userPath;$installDir" } else { $installDir }
        [Environment]::SetEnvironmentVariable("Path", $newPath, "User")
    }

    # Refresh current session PATH
    if ($env:Path -split ";" -notcontains $installDir) {
        $env:Path = "$env:Path;$installDir"
    }

    Write-Host ""
    Write-Host "Successfully installed gocraft $tag to $installDir\$binary" -ForegroundColor Green
    Write-Host ""
    Write-Host "Run 'gocraft' to launch the interactive project wizard." -ForegroundColor Yellow
} finally {
    Remove-Item -Path $tempDir -Recurse -Force -ErrorAction SilentlyContinue
}
