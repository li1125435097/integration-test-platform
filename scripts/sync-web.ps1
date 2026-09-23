$ErrorActionPreference = "Stop"
$Root = Split-Path -Parent (Split-Path -Parent $MyInvocation.MyCommand.Path)
$Web = Join-Path $Root "web"
$Dest = Join-Path $Root "server/embedded/web"

Push-Location $Web
try {
  if (-not (Test-Path "package.json")) {
    throw "sync-web: missing web/package.json"
  }
  npm ci
  npm run build
} finally {
  Pop-Location
}

if (Test-Path $Dest) { Remove-Item -Recurse -Force $Dest }
New-Item -ItemType Directory -Force -Path $Dest | Out-Null
Copy-Item -Path (Join-Path $Web "dist\*") -Destination $Dest -Recurse -Force
Write-Host "sync-web: $Web/dist -> $Dest"
