. (Join-Path $PSScriptRoot "v4\common.ps1")

Write-Step "Stopping isolated V4 containers and deleting only mindcare-v4 volumes"
Invoke-Compose down -v --remove-orphans
Write-Host "V4 cleanup complete. Unrelated Docker projects and volumes were not touched." -ForegroundColor Green
