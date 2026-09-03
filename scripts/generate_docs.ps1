# generate_docs.ps1 - Automated OpenAPI Documentation Generator for SOC Trainer
$ErrorActionPreference = "Stop"

Write-Host "Checking for swag CLI tool..." -ForegroundColor Cyan

$swagCmd = Get-Command swag -ErrorAction SilentlyContinue
if (-not $swagCmd) {
    $gopath = go env GOPATH
    if (-not $gopath) {
        $gopath = "$env:USERPROFILE\go"
    }
    $swagBinary = Join-Path $gopath "bin\swag.exe"

    if (-not (Test-Path $swagBinary)) {
        Write-Host "Installing swaggo/swag CLI..." -ForegroundColor Yellow
        go install github.com/swaggo/swag/cmd/swag@v1.16.4
    }
    $swagExec = $swagBinary
} else {
    $swagExec = "swag"
}

Write-Host "Generating Swagger/OpenAPI documentation using $swagExec..." -ForegroundColor Cyan
& $swagExec init -g main.go --parseDependency --parseInternal

if ($LASTEXITCODE -eq 0) {
    Write-Host "OpenAPI documentation successfully generated in docs/ (docs.go, swagger.json, swagger.yaml)" -ForegroundColor Green
} else {
    Write-Error "Failed to generate swagger documentation"
}
