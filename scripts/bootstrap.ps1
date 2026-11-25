<#
.SYNOPSIS
    Clean-machine Windows bootstrapper for SOC Trainer (Blue Team Training Platform).
.DESCRIPTION
    Ensures Git is installed (using winget when confirmed or -AutoDeps is set),
    clones or fast-forwards the repository, and launches install.ps1 with all
    remaining flags forwarded cleanly.
    Example one-liner:
      irm https://raw.githubusercontent.com/balliolon2/Automated_Data-Driven_Cyberattack_Simulation_for_Blue_Team_Training/main/scripts/bootstrap.ps1 | iex
#>

$ErrorActionPreference = "Stop"

$DefaultRepoUrl = if ($env:SOC_REPO_URL) { $env:SOC_REPO_URL } else { "https://github.com/balliolon2/Automated_Data-Driven_Cyberattack_Simulation_for_Blue_Team_Training.git" }
$Branch = if ($env:SOC_REPO_BRANCH) { $env:SOC_REPO_BRANCH } else { "main" }
$TargetDir = if ($env:SOC_INSTALL_DIR) { $env:SOC_INSTALL_DIR } else { Join-Path (Get-Location).Path "Automated_Data-Driven_Cyberattack_Simulation_for_Blue_Team_Training" }

$ForwardArgs = New-Object System.Collections.Generic.List[string]
$AutoDeps = $false
$NonInteractive = $false

$idx = 0
while ($idx -lt $args.Count) {
    $tok = [string]$args[$idx]
    switch -Regex ($tok) {
        '^(--repo-url|-RepoUrl)$' {
            if ($idx + 1 -ge $args.Count) { throw "Missing value after $tok" }
            $DefaultRepoUrl = [string]$args[$idx + 1]; $idx += 2; continue
        }
        '^(--branch|-Branch)$' {
            if ($idx + 1 -ge $args.Count) { throw "Missing value after $tok" }
            $Branch = [string]$args[$idx + 1]; $idx += 2; continue
        }
        '^(--dir|-Dir)$' {
            if ($idx + 1 -ge $args.Count) { throw "Missing value after $tok" }
            $TargetDir = [string]$args[$idx + 1]; $idx += 2; continue
        }
        '^(--auto-deps|-AutoDeps)$' {
            $AutoDeps = $true
            $ForwardArgs.Add("--auto-deps")
        }
        '^(--non-interactive|-NonInteractive)$' {
            $NonInteractive = $true
            $ForwardArgs.Add("--non-interactive")
        }
        default {
            $ForwardArgs.Add($tok)
        }
    }
    $idx++
}

if ($ForwardArgs.Count -eq 0) {
    $ForwardArgs.Add("install")
}

Write-Host "[INFO] SOC Trainer Windows Clean-Machine Bootstrapper" -ForegroundColor Cyan

# Ensure Git is available
$gitCmd = Get-Command git -ErrorAction SilentlyContinue
if (-not $gitCmd) {
    $wingetCmd = Get-Command winget -ErrorAction SilentlyContinue
    if (-not $wingetCmd) {
        Write-Error "[ERROR] Git is required and winget was not found. Install Git for Windows from https://git-scm.com/download/win and re-run."
        exit 1
    }

    $doInstall = $AutoDeps
    if (-not $doInstall -and -not $NonInteractive) {
        $ans = Read-Host "Git is not installed. Install Git via winget now? [Y/n]"
        $doInstall = ([string]::IsNullOrWhiteSpace($ans) -or $ans -match '^[Yy]')
    }
    if (-not $doInstall) {
        Write-Error "[ERROR] Git is missing. Install Git (`winget install --id Git.Git -e`) or re-run with --auto-deps."
        exit 1
    }

    Write-Host "[INFO] Installing Git via winget..." -ForegroundColor Cyan
    & winget install --id Git.Git -e --source winget --accept-package-agreements --accept-source-agreements
    if ($LASTEXITCODE -ne 0) {
        Write-Error "[ERROR] winget Git installation failed with exit code $LASTEXITCODE."
        exit $LASTEXITCODE
    }

    # Refresh PATH so newly installed Git is visible in this session
    $gitBin = Join-Path $env:ProgramFiles "Git\cmd"
    if (Test-Path $gitBin) { $env:Path = "$gitBin;$env:Path" }
    $gitCmd = Get-Command git -ErrorAction SilentlyContinue
    if (-not $gitCmd) {
        Write-Host "[WARN] Git was installed, please restart PowerShell and re-run bootstrapper." -ForegroundColor Yellow
        exit 2
    }
}

# Support running inside an existing clone OR cloning into TargetDir
if (Test-Path (Join-Path (Get-Location).Path "docker-compose.yml")) {
    $TargetDir = (Get-Location).Path
    Write-Host "[INFO] Detected existing repository at $TargetDir" -ForegroundColor Cyan
} elseif (Test-Path (Join-Path $TargetDir ".git")) {
    Write-Host "[INFO] Updating existing clone at $TargetDir (branch $Branch)..." -ForegroundColor Cyan
    Push-Location $TargetDir
    try {
        & git pull --ff-only origin $Branch
        if ($LASTEXITCODE -ne 0) {
            Write-Host "[WARN] git pull --ff-only could not advance automatically; continuing with current working tree." -ForegroundColor Yellow
        }
    } finally {
        Pop-Location
    }
} else {
    Write-Host "[INFO] Cloning repository ($Branch) into $TargetDir..." -ForegroundColor Cyan
    & git clone --branch $Branch $DefaultRepoUrl $TargetDir
    if ($LASTEXITCODE -ne 0) {
        Write-Error "[ERROR] Failed to clone repository from $DefaultRepoUrl"
        exit $LASTEXITCODE
    }
}

$installerPath = Join-Path $TargetDir "install.ps1"
if (-not (Test-Path $installerPath)) {
    Write-Error "[ERROR] install.ps1 not found at $installerPath"
    exit 1
}

Push-Location $TargetDir
try {
    & powershell.exe -NoProfile -ExecutionPolicy Bypass -File $installerPath @ForwardArgs
    exit $LASTEXITCODE
} finally {
    Pop-Location
}
