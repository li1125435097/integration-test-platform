param(
    [ValidateSet("build", "release")]
    [string]$Mode = "build"
)

$ErrorActionPreference = "Stop"
$Root = Split-Path -Parent (Split-Path -Parent $MyInvocation.MyCommand.Path)
Set-Location $Root

$Version = "0.1.0-dev"
if (Get-Command git -ErrorAction SilentlyContinue) {
    $gitVer = git describe --tags --always 2>$null
    if ($gitVer) { $Version = $gitVer }
}

& "$Root/scripts/sync-web.ps1"

$LDFLAGS = "-s -w -X main.version=$Version"
$Platforms = @(
    "linux/amd64", "linux/arm64", "windows/amd64", "darwin/amd64", "darwin/arm64"
)

function Build-One($Platform) {
    $parts = $Platform -split "/"
    $goos = $parts[0]
    $goarch = $parts[1]
    $name = "itp-$goos-$goarch"
    $out = if ($goos -eq "windows") { "dist/$name.exe" } else { "dist/$name" }
    Write-Host "building $Platform -> $out"
    $env:GOOS = $goos
    $env:GOARCH = $goarch
    go build -tags release -trimpath -ldflags $LDFLAGS -o $out ./server
}

New-Item -ItemType Directory -Force -Path dist, dist/release | Out-Null

if ($Mode -eq "release") {
    foreach ($p in $Platforms) {
        Build-One $p
        $goos, $goarch = $p -split "/"
        $zipname = "itp-$Version-$goos-$goarch"
        $staged = "dist/release/$zipname"
        if (Test-Path $staged) { Remove-Item -Recurse -Force $staged }
        New-Item -ItemType Directory -Force -Path "$staged/config" | Out-Null
        if ($goos -eq "windows") {
            Copy-Item "dist/itp-$goos-$goarch.exe" "$staged/integration-test-platform.exe"
        } else {
            Copy-Item "dist/itp-$goos-$goarch" "$staged/integration-test-platform"
        }
        Copy-Item config/menu.json "$staged/config/"
        @"
集成测试平台

启动: 运行 integration-test-platform
默认访问: http://127.0.0.1:8080
菜单配置: config/menu.json
"@ | Set-Content -Encoding UTF8 "$staged/RUN.txt"
        if (Test-Path "dist/release/$zipname.zip") { Remove-Item "dist/release/$zipname.zip" }
        Compress-Archive -Path $staged -DestinationPath "dist/release/$zipname.zip"
        Write-Host "release zip: dist/release/$zipname.zip"
    }
} else {
    $goos = go env GOOS
    $goarch = go env GOARCH
    Build-One "$goos/$goarch"
}
