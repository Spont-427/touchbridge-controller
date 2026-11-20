$ErrorActionPreference = 'Stop'
$Root = Split-Path -Parent $PSScriptRoot
Set-Location $Root
New-Item -ItemType Directory -Force -Path dist | Out-Null
$env:CGO_ENABLED = '0'
$env:GOOS = 'windows'
$env:GOARCH = 'amd64'
go build -trimpath -ldflags='-s -w' -o dist/touchbridge-windows-amd64.exe ./cmd/touchbridge
Write-Host "Built dist/touchbridge-windows-amd64.exe"
