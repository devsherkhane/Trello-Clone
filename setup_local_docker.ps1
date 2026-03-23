# Setup Local Docker Helper Script for Drift

Write-Host "--- Drift Docker Setup Helper ---" -ForegroundColor Cyan

# 1. Check for Docker
$dockerCheck = Get-Command docker -ErrorAction SilentlyContinue
if ($dockerCheck) {
    Write-Host "[OK] Docker is already installed!" -ForegroundColor Green
    docker --version
} else {
    Write-Host "[MISSING] Docker is not installed or not in your PATH." -ForegroundColor Yellow
    Write-Host "Please download and install Docker Desktop for Windows here:"
    Write-Host "https://www.docker.com/products/docker-desktop/" -ForegroundColor Blue
    Write-Host ""
}

# 2. Check for Docker Compose
$composeCheck = Get-Command docker-compose -ErrorAction SilentlyContinue
if ($composeCheck) {
    Write-Host "[OK] Docker Compose is already installed!" -ForegroundColor Green
    docker-compose --version
} else {
    Write-Host "[INFO] Modern Docker includes 'docker compose' as a plugin."
    Write-Host "Try running: docker compose version" -ForegroundColor Gray
}

Write-Host ""
Write-Host "Once Docker is installed and running, you can deploy Drift locally with:"
Write-Host "docker-compose up -d --build" -ForegroundColor Cyan
