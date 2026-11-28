<#
.SYNOPSIS
    SOC Trainer - Windows Automated Deployment & Operations Harness (install.ps1)
.DESCRIPTION
    Symmetric Windows counterpart to install.sh implementing Spec #64 (Q1-Q17).
    PowerShell 5.1+ and 7+ compatible. Requires zero host Python/Go/Node toolchains.
#>

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

$Script:CallerPwd = (Get-Location).Path
$Script:RootDir = if ($PSScriptRoot) { $PSScriptRoot } else { $Script:CallerPwd }
Set-Location $Script:RootDir

$Script:ResetTruncateSql = "TRUNCATE TABLE session_actions, simulation_sessions, exam_session_questions, exam_sessions, user_skill_profiles, scenario_snapshots, evaluation_metrics, user_theory_progress, user_theory_subtopic_progress, user_theory_sample_questions, comment_upvotes, thread_upvotes, thread_comments, analysis_threads, user_notifications, specialist_applications, nfr_evaluation_runs CASCADE;"

# CLI Options & Defaults
$Script:Subcommand = ""
$Script:RestoreFile = ""
$Script:LogsService = ""
$Script:Mode = ""
$Script:NonInteractive = $false
$Script:Domain = ""
$Script:AcmeEmail = ""
$Script:AutoDeps = $false
$Script:ConfigureFirewall = $false
$Script:ImageTag = ""
$Script:VerifyE2E = $false
$Script:WithPgAdmin = $false
$Script:SkipDnsCheck = $false
$Script:NoPull = $false
$Script:NoUploads = $false
$Script:RemoveVolumes = $false

function Write-LogInfo([string]$Msg)    { Write-Host "[INFO] $Msg" -ForegroundColor Cyan }
function Write-LogOk([string]$Msg)      { Write-Host "[SUCCESS] $Msg" -ForegroundColor Green }
function Write-LogWarn([string]$Msg)    { Write-Host "[WARN] $Msg" -ForegroundColor Yellow }
function Write-LogErr([string]$Msg)     { Write-Host "[ERROR] $Msg" -ForegroundColor Red }

function Fail-Stage {
    param(
        [string]$Stage,
        [int]$ExitCode = 1,
        [string]$Cause,
        [string]$Diagnostic,
        [string]$Resume
    )
    Write-Host ""
    Write-Host "========================================================" -ForegroundColor Red
    Write-LogErr "Failure Stage : $Stage"
    Write-LogErr "Exit Code     : $ExitCode"
    Write-LogErr "Likely Cause  : $Cause"
    if ($Diagnostic) { Write-Host "Diagnostic    : $Diagnostic" -ForegroundColor Yellow }
    if ($Resume)     { Write-Host "Resume Step   : $Resume" -ForegroundColor Cyan }
    Write-Host "========================================================" -ForegroundColor Red
    exit $ExitCode
}

function Show-Usage {
@"
SOC Trainer - Windows Setup & Operations CLI (install.ps1 / install.bat)

Usage:
  .\install.bat [command] [flags]
  .\install.ps1 [command] [flags]

Commands (opens interactive menu when omitted in interactive terminal):
  install            Preflight, env setup, containerized seed, startup & non-polluting smoke verify
  check              Read-only preflight checks (Windows, WSL2, Docker/Compose, ports, RAM/disk, .env ACL)
  status             Show container health, endpoints, active mode, redacted LLM/TLS state
  seed               Run idempotent containerized seed (domains, 540 questions, theory, 15 scenarios)
  check-seed         Read-only PostgreSQL count check across core tables
  update             Auto-backup -> optional git pull (--no-pull) -> rebuild/pull & re-seed -> smoke verify
  backup             Save gzipped SQL + uploads archive in backups/ (keeps 10 newest; --no-uploads)
  restore <file>     Guarded DB (+ matching uploads) restore after safety backup and RESTORE confirmation
  logs [service]     Stream Docker Compose logs
  uninstall          Stop/remove containers cleanly (preserves data volumes; --volumes needs DELETE DATA)
  reset-dev-data     Local-only wipe of learner/session runtime tables & re-seed (BLOCKED in Cloud mode)

Flags:
  --mode <local|cloud>       Deployment mode (local or cloud)
  --non-interactive          Run without interactive prompts using ENV / *_FILE / generated secrets
  --domain <hostname>        Public domain for Cloud Auto-HTTPS via Caddy (Let's Encrypt 80/443)
  --acme-email <email>       Optional ACME contact email for TLS notifications
  --auto-deps                Auto-install missing prerequisites via winget / wsl
  --configure-firewall       Explicitly add Windows Defender Firewall inbound rules (80/443)
  --image-tag <tag>          Pull & deploy pinned versioned images instead of building from source
  --verify-e2e               Also execute scripts/test_integration.py (Local mode ONLY)
  --with-pgadmin             Start opt-in pgAdmin tools profile on 127.0.0.1:5050
  --skip-dns-check           Skip advisory DNS A-record check before Cloud HTTPS startup
  --no-pull                  Skip git pull during update
  --no-uploads               Skip uploads volume archive during backup
  --volumes                  With uninstall: delete Docker volumes after typed DELETE DATA confirmation
  --help, -h                 Show this help message
"@ | Write-Host
}

# Safe wrappers so native command stderr never triggers NativeCommandError in PowerShell 5.1
function Invoke-NativeQuiet([string]$Exe, [string[]]$CmdArgs) {
    $prevEap = $ErrorActionPreference
    $ErrorActionPreference = "Continue"
    try {
        & $Exe @CmdArgs *> $null
        return [int]$LASTEXITCODE
    } finally {
        $ErrorActionPreference = $prevEap
    }
}

function Invoke-NativeCapture([string]$Exe, [string[]]$CmdArgs) {
    $prevEap = $ErrorActionPreference
    $ErrorActionPreference = "Continue"
    try {
        $lines = & $Exe @CmdArgs 2>$null
        $rc = [int]$LASTEXITCODE
        return @{ ExitCode = $rc; Output = (($lines | Out-String).Trim()) }
    } finally {
        $ErrorActionPreference = $prevEap
    }
}

function Resolve-CallerPath([string]$RawPath) {
    if ([string]::IsNullOrWhiteSpace($RawPath)) { return "" }
    if ([System.IO.Path]::IsPathRooted($RawPath)) { return $RawPath }
    $callerCandidate = Join-Path $Script:CallerPwd $RawPath
    if (Test-Path $callerCandidate) { return $callerCandidate }
    return (Join-Path $Script:RootDir $RawPath)
}

# Parse incoming CLI arguments (supports both GNU --flag and PS -Flag styles)
$rawArgs = @($args)
$i = 0
$positionals = New-Object System.Collections.Generic.List[string]
while ($i -lt $rawArgs.Count) {
    $arg = [string]$rawArgs[$i]
    switch -Regex ($arg) {
        '^(--help|-h|-\?)$' { Show-Usage; exit 0 }
        '^--mode=(.+)$'     { $Script:Mode = $Matches[1].ToLowerInvariant(); break }
        '^(--mode|-Mode)$' {
            if ($i + 1 -ge $rawArgs.Count) { Fail-Stage "Parse Arguments" 2 "Missing value for $arg" ".\install.bat --help" "" }
            $Script:Mode = ([string]$rawArgs[$i + 1]).ToLowerInvariant(); $i++; break
        }
        '^(--non-interactive|-NonInteractive)$' { $Script:NonInteractive = $true; break }
        '^--domain=(.+)$'   { $Script:Domain = $Matches[1].Trim(); break }
        '^(--domain|-Domain)$' {
            if ($i + 1 -ge $rawArgs.Count) { Fail-Stage "Parse Arguments" 2 "Missing value for $arg" ".\install.bat --help" "" }
            $Script:Domain = ([string]$rawArgs[$i + 1]).Trim(); $i++; break
        }
        '^--acme-email=(.+)$' { $Script:AcmeEmail = $Matches[1].Trim(); break }
        '^(--acme-email|-AcmeEmail)$' {
            if ($i + 1 -ge $rawArgs.Count) { Fail-Stage "Parse Arguments" 2 "Missing value for $arg" ".\install.bat --help" "" }
            $Script:AcmeEmail = ([string]$rawArgs[$i + 1]).Trim(); $i++; break
        }
        '^(--auto-deps|-AutoDeps)$' { $Script:AutoDeps = $true; break }
        '^(--configure-firewall|-ConfigureFirewall)$' { $Script:ConfigureFirewall = $true; break }
        '^--image-tag=(.+)$' { $Script:ImageTag = $Matches[1].Trim(); break }
        '^(--image-tag|-ImageTag)$' {
            if ($i + 1 -ge $rawArgs.Count) { Fail-Stage "Parse Arguments" 2 "Missing value for $arg" ".\install.bat --help" "" }
            $Script:ImageTag = ([string]$rawArgs[$i + 1]).Trim(); $i++; break
        }
        '^(--verify-e2e|-VerifyE2E)$' { $Script:VerifyE2E = $true; break }
        '^(--with-pgadmin|-WithPgAdmin)$' { $Script:WithPgAdmin = $true; break }
        '^(--skip-dns-check|-SkipDnsCheck)$' { $Script:SkipDnsCheck = $true; break }
        '^(--no-pull|-NoPull)$' { $Script:NoPull = $true; break }
        '^(--no-uploads|-NoUploads)$' { $Script:NoUploads = $true; break }
        '^(--volumes|-Volumes)$' { $Script:RemoveVolumes = $true; break }
        '^-' {
            Fail-Stage "Parse Arguments" 2 "Unknown flag: $arg" ".\install.bat --help" "Run '.\install.bat --help' for valid options."
        }
        default {
            $positionals.Add($arg)
        }
    }
    $i++
}

if ($positionals.Count -gt 0) {
    $Script:Subcommand = $positionals[0].ToLowerInvariant()
    if ($positionals.Count -gt 1) {
        if ($Script:Subcommand -eq "restore") { $Script:RestoreFile = $positionals[1] }
        elseif ($Script:Subcommand -eq "logs") { $Script:LogsService = $positionals[1] }
    }
}

function Write-Utf8NoBom([string]$Path, [string]$Text) {
    $enc = New-Object System.Text.UTF8Encoding($false)
    [System.IO.File]::WriteAllText($Path, $Text, $enc)
}

function New-CryptoHex([int]$ByteCount) {
    $bytes = New-Object byte[] $ByteCount
    $rng = [System.Security.Cryptography.RandomNumberGenerator]::Create()
    try { $rng.GetBytes($bytes) } finally { $rng.Dispose() }
    return ([System.BitConverter]::ToString($bytes)).Replace("-", "").ToLowerInvariant()
}

function New-CryptoSafePassword([int]$Length = 24) {
    $chars = "ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz23456789-_@!"
    $bytes = New-Object byte[] $Length
    $rng = [System.Security.Cryptography.RandomNumberGenerator]::Create()
    try { $rng.GetBytes($bytes) } finally { $rng.Dispose() }
    $sb = New-Object System.Text.StringBuilder
    foreach ($b in $bytes) { [void]$sb.Append($chars[$b % $chars.Length]) }
    return $sb.ToString()
}

function Read-DotEnvMap([string]$Path) {
    $map = [ordered]@{}
    if (-not (Test-Path $Path -PathType Leaf)) { return $map }
    foreach ($line in [System.IO.File]::ReadAllLines($Path)) {
        $trim = $line.Trim()
        if ($trim -eq "" -or $trim.StartsWith("#")) { continue }
        $eqIdx = $trim.IndexOf("=")
        if ($eqIdx -lt 1) { continue }
        $k = $trim.Substring(0, $eqIdx).Trim()
        $v = $trim.Substring($eqIdx + 1).Trim()
        if (($v.StartsWith('"') -and $v.EndsWith('"') -and $v.Length -ge 2) -or ($v.StartsWith("'") -and $v.EndsWith("'") -and $v.Length -ge 2)) {
            $v = $v.Substring(1, $v.Length - 2)
        }
        $map[$k] = $v
    }
    return $map
}

function Get-DbConfig {
    $envMap = Ensure-ComposeProcessEnv
    $pgUser = if ($envMap.Contains("POSTGRES_USER") -and $envMap["POSTGRES_USER"]) { [string]$envMap["POSTGRES_USER"] } else { "admin" }
    $pgDb = if ($envMap.Contains("POSTGRES_DB") -and $envMap["POSTGRES_DB"]) { [string]$envMap["POSTGRES_DB"] } else { "cyber_sim" }
    if ($pgUser -notmatch '^[A-Za-z0-9_]+$' -or $pgDb -notmatch '^[A-Za-z0-9_]+$') {
        Fail-Stage "Database Config" 2 "POSTGRES_USER/POSTGRES_DB contain unsafe identifier characters." "" "Use alphanumeric/underscore database identifiers."
    }
    return @{ User = $pgUser; Db = $pgDb; EnvMap = $envMap }
}

function Read-SecretPrompt([string]$PromptText) {
    $sec = Read-Host -Prompt $PromptText -AsSecureString
    $bstr = [System.Runtime.InteropServices.Marshal]::SecureStringToBSTR($sec)
    try {
        return [System.Runtime.InteropServices.Marshal]::PtrToStringAuto($bstr)
    } finally {
        [System.Runtime.InteropServices.Marshal]::ZeroFreeBSTR($bstr)
    }
}

function Test-SecretFileOrEnv([string]$Key, [System.Collections.IDictionary]$ExistingEnv) {
    $fromEnv = [Environment]::GetEnvironmentVariable($Key)
    if (-not [string]::IsNullOrWhiteSpace($fromEnv)) { return $fromEnv.Trim() }

    $fileEnvVar = [Environment]::GetEnvironmentVariable("${Key}_FILE")
    if (-not [string]::IsNullOrWhiteSpace($fileEnvVar)) {
        $resolvedFile = Resolve-CallerPath $fileEnvVar.Trim()
        if (-not (Test-Path $resolvedFile -PathType Leaf)) {
            Fail-Stage "Environment Setup" 1 "${Key}_FILE points to missing file: $resolvedFile" "Test-Path '$resolvedFile'" "Provide a readable secret file path."
        }
        $val = ([System.IO.File]::ReadAllText($resolvedFile)).Trim("`r", "`n")
        if ([string]::IsNullOrWhiteSpace($val)) {
            Fail-Stage "Environment Setup" 1 "${Key}_FILE is empty: $resolvedFile" "Check contents of '$resolvedFile'" "Populate the secret file and retry."
        }
        return $val
    }

    if ($ExistingEnv.Contains($Key) -and -not [string]::IsNullOrWhiteSpace([string]$ExistingEnv[$Key])) {
        return ([string]$ExistingEnv[$Key]).Trim()
    }
    return ""
}

function Resolve-ExistingOrDefault([string]$Key, [System.Collections.IDictionary]$ExistingEnv, [string]$DefaultVal) {
    $fromEnv = [Environment]::GetEnvironmentVariable($Key)
    if (-not [string]::IsNullOrWhiteSpace($fromEnv)) { return $fromEnv.Trim() }
    if ($ExistingEnv.Contains($Key) -and -not [string]::IsNullOrWhiteSpace([string]$ExistingEnv[$Key])) {
        return ([string]$ExistingEnv[$Key]).Trim()
    }
    return $DefaultVal
}

function Set-EnvFileAcl([string]$Path, [bool]$Fatal = $true) {
    $currentSid = ([System.Security.Principal.WindowsIdentity]::GetCurrent()).User.Value
    $userAce = "*${currentSid}:(F)"
    # Grant Full Control (F) strictly to file owner, SYSTEM (*S-1-5-18), and Built-in Administrators (*S-1-5-32-544)
    # so Move/Write/Elevated-Admin/OneDrive sync work while blocking Users/Everyone/Authenticated Users.
    try {
        $attrs = [System.IO.File]::GetAttributes($Path)
        if (($attrs -band [System.IO.FileAttributes]::ReadOnly) -ne 0) {
            [System.IO.File]::SetAttributes($Path, ($attrs -band (-bnot [System.IO.FileAttributes]::ReadOnly)))
        }
    } catch {}

    $rc = Invoke-NativeQuiet "icacls" @($Path, "/inheritance:r", "/grant:r", $userAce, "*S-1-5-18:(F)", "*S-1-5-32-544:(F)")
    if ($rc -ne 0) {
        # Fallback if an earlier broken ACL stripped WRITE_DAC; reclaim ownership and re-lock
        $null = Invoke-NativeQuiet "takeown" @("/F", $Path)
        $rc = Invoke-NativeQuiet "icacls" @($Path, "/inheritance:r", "/grant:r", $userAce, "*S-1-5-18:(F)", "*S-1-5-32-544:(F)")
    }
    if ($rc -ne 0) {
        if ($Fatal) {
            Fail-Stage "Secret File Protection" $rc "Failed to lock ACL on $Path via icacls." "icacls '$Path'" "Ensure current user owns $Path (and OneDrive is not blocking access) and retry."
        }
        Write-LogWarn "icacls returned exit code $rc for $Path."
    }
}

function Test-EnvFileAcl([string]$Path) {
    $res = Invoke-NativeCapture "icacls" @($Path)
    if ($res.ExitCode -ne 0 -or -not $res.Output) {
        Write-LogWarn "Could not inspect .env ACL via icacls."
        return
    }
    if ($res.Output -match '(?i)(BUILTIN\\Users|Everyone|NT AUTHORITY\\Authenticated Users)') {
        Write-LogWarn ".env ACL allows broad read/access ($($Matches[1])); re-run install to tighten via icacls."
    } else {
        Write-LogOk ".env File ACL: Owner/Admin-restricted"
    }
}

function Write-MergedEnvFile([System.Collections.IDictionary]$Updates) {
    $envPath = Join-Path $Script:RootDir ".env"
    $templatePath = Join-Path $Script:RootDir ".env.example"
    $sourceLines = if (Test-Path $envPath -PathType Leaf) {
        [System.IO.File]::ReadAllLines($envPath)
    } elseif (Test-Path $templatePath -PathType Leaf) {
        [System.IO.File]::ReadAllLines($templatePath)
    } else {
        @()
    }

    $seen = New-Object 'System.Collections.Generic.HashSet[string]'
    $out = New-Object 'System.Collections.Generic.List[string]'
    foreach ($line in $sourceLines) {
        $trim = $line.Trim()
        if ($trim -eq "" -or $trim.StartsWith("#") -or ($trim.IndexOf("=") -lt 1)) {
            $out.Add($line)
            continue
        }
        $k = $trim.Substring(0, $trim.IndexOf("=")).Trim()
        if ($Updates.Contains($k)) {
            $out.Add("$k=$($Updates[$k])")
            [void]$seen.Add($k)
        } else {
            $out.Add($line)
        }
    }
    foreach ($k in $Updates.Keys) {
        if (-not $seen.Contains($k)) {
            $out.Add("$k=$($Updates[$k])")
        }
    }

    # Pre-create empty .env (zero secrets) and lock ACL BEFORE writing secrets directly
    # Avoids Move-Item rename races inside OneDrive/cloud-synced folders and rescues older (R,W) locks.
    if (-not (Test-Path $envPath -PathType Leaf)) {
        Write-Utf8NoBom $envPath ""
    }
    Set-EnvFileAcl $envPath $true

    $payload = ($out -join "`n") + "`n"
    $written = $false
    $lastErr = $null
    for ($attempt = 1; $attempt -le 4; $attempt++) {
        try {
            Write-Utf8NoBom $envPath $payload
            $written = $true
            break
        } catch {
            $lastErr = $_
            Start-Sleep -Milliseconds 250
        }
    }
    if (-not $written) {
        Fail-Stage "Environment Setup" 1 "Could not write secrets to ${envPath}: $lastErr" "icacls '$envPath'" "Pause OneDrive sync briefly or check folder permissions, then re-run."
    }

    # Clean up any stale .env.tmp.* files left from earlier failed attempts (unlock ACL first if needed)
    foreach ($stale in @(Get-ChildItem -Path $Script:RootDir -Filter ".env.tmp.*" -Force -ErrorAction SilentlyContinue)) {
        Set-EnvFileAcl $stale.FullName $false
        Remove-Item -LiteralPath $stale.FullName -Force -ErrorAction SilentlyContinue
    }
    Write-LogOk "Locked .env file permissions (owner-only Full Control via icacls)."
}

function Ensure-ComposeProcessEnv {
    $envPath = Join-Path $Script:RootDir ".env"
    $env:COMPOSE_PATH_SEPARATOR = ":"
    if ($Script:ImageTag) { $env:IMAGE_TAG = $Script:ImageTag }
    else { Remove-Item Env:IMAGE_TAG -ErrorAction SilentlyContinue }

    if (Test-Path $envPath -PathType Leaf) {
        $map = Read-DotEnvMap $envPath
        if ($map.Contains("COMPOSE_PATH_SEPARATOR") -and $map["COMPOSE_PATH_SEPARATOR"]) {
            $env:COMPOSE_PATH_SEPARATOR = [string]$map["COMPOSE_PATH_SEPARATOR"]
        }
        if ($map.Contains("COMPOSE_FILE") -and $map["COMPOSE_FILE"]) {
            $env:COMPOSE_FILE = [string]$map["COMPOSE_FILE"]
        } else {
            Remove-Item Env:COMPOSE_FILE -ErrorAction SilentlyContinue
        }
        return $map
    }
    Remove-Item Env:COMPOSE_FILE -ErrorAction SilentlyContinue
    return [ordered]@{}
}

function Invoke-DockerCompose([string[]]$ComposeArgs) {
    Ensure-ComposeProcessEnv | Out-Null
    $baseArgs = New-Object 'System.Collections.Generic.List[string]'
    $baseArgs.Add("compose")
    if ($Script:WithPgAdmin) {
        $baseArgs.Add("--profile")
        $baseArgs.Add("tools")
    }
    foreach ($a in $ComposeArgs) { $baseArgs.Add($a) }
    $prevEap = $ErrorActionPreference
    $ErrorActionPreference = "Continue"
    try {
        & docker @baseArgs
    } finally {
        $ErrorActionPreference = $prevEap
    }
}

function Test-PortInUse([int]$Port) {
    try {
        if (Get-Command Get-NetTCPConnection -ErrorAction SilentlyContinue) {
            $conn = Get-NetTCPConnection -LocalPort $Port -State Listen -ErrorAction SilentlyContinue
            return ($null -ne $conn)
        }
    } catch {}
    $client = New-Object System.Net.Sockets.TcpClient
    try {
        $ar = $client.BeginConnect("127.0.0.1", $Port, $null, $null)
        if ($ar.AsyncWaitHandle.WaitOne(250, $false)) {
            $client.EndConnect($ar)
            return $true
        }
    } catch {} finally {
        $client.Close()
    }
    return $false
}

function Invoke-Preflight([bool]$AllowAutoInstall = $false) {
    Write-LogInfo "Running Windows Preflight Checks..."
    $errors = 0

    $osCaption = [System.Environment]::OSVersion.VersionString
    $arch = if ($env:PROCESSOR_ARCHITECTURE) { $env:PROCESSOR_ARCHITECTURE } else { "unknown" }
    Write-LogInfo "OS: Windows ($osCaption) | Arch: $arch | PowerShell: $($PSVersionTable.PSVersion)"

    try {
        $cs = Get-CimInstance -ClassName Win32_ComputerSystem -ErrorAction SilentlyContinue
        if ($cs) {
            $ramGb = [Math]::Round($cs.TotalPhysicalMemory / 1GB, 2)
            Write-LogInfo "System RAM: ${ramGb} GB"
            if ($ramGb -lt 2.0) { Write-LogWarn "System RAM is below 2 GB; Docker build/runtime may run slowly." }
        }
    } catch {}

    try {
        $driveName = (Get-Item $Script:RootDir).PSDrive.Name
        $psDrive = Get-PSDrive -Name $driveName -ErrorAction SilentlyContinue
        if ($psDrive -and $psDrive.Free) {
            $freeGb = [Math]::Round($psDrive.Free / 1GB, 2)
            Write-LogInfo "Free Disk on ${driveName}: : ${freeGb} GB"
            if ($freeGb -lt 4.0) { Write-LogWarn "Low disk space (< 4 GB); container images/builds require several gigabytes." }
        }
    } catch {}

    # Check WSL2 status (with optional install under --auto-deps)
    if (Get-Command wsl -ErrorAction SilentlyContinue) {
        $wslRc = Invoke-NativeQuiet "wsl" @("--status")
        if ($wslRc -eq 0) {
            Write-LogOk "WSL2 subsystem: Ready"
        } else {
            Write-LogWarn "WSL2 subsystem returned non-zero status ($wslRc)."
            if ($AllowAutoInstall) {
                $doWsl = $Script:AutoDeps
                if (-not $doWsl -and -not $Script:NonInteractive) {
                    $ansW = Read-Host "WSL2 is not configured cleanly. Run 'wsl --install --no-distribution' now? [y/N]"
                    $doWsl = ($ansW -match '^[Yy]')
                }
                if ($doWsl) {
                    & wsl --install --no-distribution
                    Fail-Stage "Preflight (WSL2 Setup)" 2 "WSL2 feature setup ran and usually requires a Windows reboot." "wsl --status" "Reboot Windows, start Docker Desktop, then re-run '.\install.bat install'."
                }
            }
        }
    } else {
        Write-LogWarn "wsl.exe not detected (expected on Docker Desktop WSL2 backend)."
    }

    # Check Git
    if (Get-Command git -ErrorAction SilentlyContinue) {
        $gitRes = Invoke-NativeCapture "git" @("--version")
        Write-LogOk "Git: $($gitRes.Output)"
    } else {
        Write-LogWarn "Git not found on PATH."
        if ($AllowAutoInstall -and (Get-Command winget -ErrorAction SilentlyContinue)) {
            $doGit = $Script:AutoDeps
            if (-not $doGit -and -not $Script:NonInteractive) {
                $ans = Read-Host "Install Git for Windows via winget? [Y/n]"
                $doGit = ([string]::IsNullOrWhiteSpace($ans) -or $ans -match '^[Yy]')
            }
            if ($doGit) {
                & winget install --id Git.Git -e --source winget --accept-package-agreements --accept-source-agreements
                $gitBin = Join-Path $env:ProgramFiles "Git\cmd"
                if (Test-Path $gitBin) { $env:Path = "$gitBin;$env:Path" }
            }
        }
    }

    # Check Docker CLI & Daemon
    $dockerCmd = Get-Command docker -ErrorAction SilentlyContinue
    if (-not $dockerCmd) {
        if ($AllowAutoInstall -and (Get-Command winget -ErrorAction SilentlyContinue)) {
            $doDocker = $Script:AutoDeps
            if (-not $doDocker -and -not $Script:NonInteractive) {
                $ans = Read-Host "Docker CLI not found. Install Docker Desktop via winget? [Y/n]"
                $doDocker = ([string]::IsNullOrWhiteSpace($ans) -or $ans -match '^[Yy]')
            }
            if ($doDocker) {
                Write-LogInfo "Installing Docker Desktop via winget..."
                & winget install --id Docker.DockerDesktop -e --source winget --accept-package-agreements --accept-source-agreements
                if ($LASTEXITCODE -ne 0) {
                    Fail-Stage "Preflight (Docker Desktop Install)" $LASTEXITCODE "winget failed to install Docker Desktop." "winget install --id Docker.DockerDesktop -e" "Install Docker Desktop manually, restart if WSL2 requires reboot, then re-run '.\install.bat install'."
                }
                Fail-Stage "Preflight (Reboot/Start Required)" 2 "Docker Desktop was just installed and needs WSL2 activation/startup." "Open Docker Desktop from Start Menu and verify 'docker info'." "Restart Windows if prompted, launch Docker Desktop until Engine is running, then re-run '.\install.bat install'."
            }
        }
        Write-LogErr "Docker CLI: Not found."
        $errors++
    } else {
        $dVer = (Invoke-NativeCapture "docker" @("--version")).Output
        Write-LogOk "Docker CLI: $dVer"
        $compRes = Invoke-NativeCapture "docker" @("compose", "version")
        if ($compRes.ExitCode -eq 0) {
            Write-LogOk "Docker Compose: $($compRes.Output)"
        } else {
            Write-LogErr "Docker Compose plugin v2 not found."
            $errors++
        }

        $infoRc = Invoke-NativeQuiet "docker" @("info")
        if ($infoRc -ne 0) {
            if ($AllowAutoInstall) {
                $ddExe = Join-Path $env:ProgramFiles "Docker\Docker\Docker Desktop.exe"
                if (Test-Path $ddExe) {
                    Write-LogWarn "Docker daemon is stopped. Launching Docker Desktop and waiting up to 90s..."
                    Start-Process -FilePath $ddExe | Out-Null
                    $waited = 0
                    $ready = $false
                    while ($waited -lt 90) {
                        Start-Sleep -Seconds 3
                        $waited += 3
                        if ((Invoke-NativeQuiet "docker" @("info")) -eq 0) { $ready = $true; break }
                        Write-Host "." -NoNewline
                    }
                    Write-Host ""
                    if ($ready) {
                        Write-LogOk "Docker daemon is running!"
                    } else {
                        Write-LogErr "Docker daemon did not become ready within 90s."
                        $errors++
                    }
                } else {
                    Write-LogErr "Docker daemon is not running and Docker Desktop executable was not found."
                    $errors++
                }
            } else {
                Write-LogErr "Docker daemon is not running."
                $errors++
            }
        } else {
            Write-LogOk "Docker daemon: Running"
        }
    }

    $envPath = Join-Path $Script:RootDir ".env"
    if (Test-Path $envPath -PathType Container) {
        Write-LogErr ".env is a directory instead of a file!"
        $errors++
    } elseif (Test-Path $envPath -PathType Leaf) {
        Write-LogOk "Environment File (.env): Present"
        Test-EnvFileAcl $envPath
    } else {
        Write-LogInfo "Environment File (.env): Not created yet (will generate during install)"
    }

    foreach ($req in @("docker-compose.yml", "docker-compose.cloud.yml", "Caddyfile.template", ".env.example", "init.sql")) {
        if (-not (Test-Path (Join-Path $Script:RootDir $req) -PathType Leaf)) {
            Write-LogErr "Required file missing: $req"
            $errors++
        }
    }

    foreach ($p in @(80, 443, 8080, 5432)) {
        if (Test-PortInUse $p) {
            Write-LogWarn "Port $p is currently listening on host (expected if stack is already running)."
        } else {
            Write-LogInfo "Port $p : Available"
        }
    }

    $shPath = Join-Path $Script:RootDir "install.sh"
    if (Test-Path $shPath -PathType Leaf) {
        $shBytes = [System.IO.File]::ReadAllBytes($shPath)
        for ($idx = 0; $idx -lt ($shBytes.Length - 1); $idx++) {
            if ($shBytes[$idx] -eq 13 -and $shBytes[$idx + 1] -eq 10) {
                Write-LogWarn "install.sh contains CRLF bytes on disk; check .gitattributes before copying to Linux."
                break
            }
        }
    }

    if ($errors -gt 0) {
        Fail-Stage "Preflight Checks" 1 "$errors preflight requirement(s) failed." "docker info; docker compose version" "Ensure Docker Desktop (WSL2) is running and re-run."
    }
    Write-LogOk "All Windows Preflight Checks Passed!"
}

function Render-Caddyfile([string]$DomainName, [string]$Email) {
    $tplPath = Join-Path $Script:RootDir "Caddyfile.template"
    if (-not (Test-Path $tplPath -PathType Leaf)) {
        Fail-Stage "Cloud HTTPS Config" 1 "Missing Caddyfile.template" "git status" "Restore Caddyfile.template from repository."
    }
    $tpl = [System.IO.File]::ReadAllText($tplPath)
    $emailBlock = if (-not [string]::IsNullOrWhiteSpace($Email)) {
        "{`n    email $($Email.Trim())`n}`n`n"
    } else {
        ""
    }
    $rendered = $tpl.Replace("__ACME_EMAIL_BLOCK__", $emailBlock).Replace("__DOMAIN__", $DomainName.Trim())
    Write-Utf8NoBom (Join-Path $Script:RootDir "Caddyfile") $rendered
    Write-LogOk "Rendered Caddyfile for $DomainName (HTTP -> HTTPS Auto-TLS)."
}

function Check-CloudDns([string]$DomainName) {
    if ($Script:SkipDnsCheck) {
        Write-LogInfo "Skipping DNS A-record check (--skip-dns-check)."
        return
    }
    Write-LogInfo "Checking DNS resolution for domain '$DomainName'..."
    $resolvedIps = @()
    try {
        $resolvedIps = @([System.Net.Dns]::GetHostAddresses($DomainName) | ForEach-Object { $_.IPAddressToString })
    } catch {
        $resolvedIps = @()
    }
    if ($resolvedIps.Count -eq 0) {
        Write-LogWarn "Domain '$DomainName' did not resolve via DNS yet. Let's Encrypt TLS requires public DNS pointing to this server."
        if (-not $Script:NonInteractive) {
            $cont = Read-Host "Continue anyway? [y/N]"
            if ($cont -notmatch '^[Yy]') {
                Fail-Stage "DNS Verification" 1 "Operator paused waiting for DNS propagation on $DomainName." "Resolve-DnsName $DomainName" "Point your DNS A-record to this server's public IP and re-run."
            }
        }
        return
    }

    try {
        $pubIp = (Invoke-RestMethod -Uri "https://api.ipify.org" -TimeoutSec 4 -ErrorAction Stop).ToString().Trim()
        if ($pubIp -and ($resolvedIps -notcontains $pubIp)) {
            Write-LogWarn "DNS for $DomainName resolves to [$($resolvedIps -join ', ')], whereas detected host outbound IP is $pubIp."
            if (-not $Script:NonInteractive) {
                $cont = Read-Host "Continue with Caddy Auto-HTTPS anyway? [y/N]"
                if ($cont -notmatch '^[Yy]') {
                    Fail-Stage "DNS Verification" 1 "DNS A-record mismatch for $DomainName." "Resolve-DnsName $DomainName" "Update DNS or rerun with --skip-dns-check."
                }
            }
        } else {
            Write-LogOk "DNS A-record matches host public IP ($pubIp)."
        }
    } catch {
        Write-LogInfo "Resolved $DomainName -> $($resolvedIps -join ', ') (public IP lookup unavailable)."
    }
}

function Audit-WindowsFirewall([bool]$IsCloud, [bool]$AllowTls) {
    Write-Host ""
    Write-LogInfo "Firewall & Network Exposure Check:"
    if ($IsCloud) {
        Write-Host "  Required Cloud Inbound Rules : TCP 80 (HTTP), TCP 443 (HTTPS when domain TLS active)" -ForegroundColor Gray
        Write-Host "  Must Stay Private / Loopback : TCP 5432 (PostgreSQL), TCP 8080 (Go API), TCP 5050 (pgAdmin)" -ForegroundColor Gray
    } else {
        Write-Host "  Local Stack Ports            : Web http://localhost:80 | API http://127.0.0.1:8080" -ForegroundColor Gray
    }

    if (-not $Script:ConfigureFirewall -and $IsCloud -and -not $Script:NonInteractive) {
        $ans = Read-Host "Configure Windows Defender Firewall inbound rules for HTTP/HTTPS now? [y/N]"
        if ($ans -match '^[Yy]') { $Script:ConfigureFirewall = $true }
    }

    if ($Script:ConfigureFirewall) {
        $rc80 = Invoke-NativeQuiet "netsh" @("advfirewall", "firewall", "show", "rule", "name=SOC-Trainer-HTTP-80")
        if ($rc80 -ne 0) {
            $add80 = Invoke-NativeQuiet "netsh" @("advfirewall", "firewall", "add", "rule", "name=SOC-Trainer-HTTP-80", "dir=in", "action=allow", "protocol=TCP", "localport=80")
            if ($add80 -ne 0) { Write-LogWarn "Failed to add HTTP 80 firewall rule (Administrator rights may be required)." }
        }
        if ($AllowTls) {
            $rc443 = Invoke-NativeQuiet "netsh" @("advfirewall", "firewall", "show", "rule", "name=SOC-Trainer-HTTPS-443")
            if ($rc443 -ne 0) {
                $add443 = Invoke-NativeQuiet "netsh" @("advfirewall", "firewall", "add", "rule", "name=SOC-Trainer-HTTPS-443", "dir=in", "action=allow", "protocol=TCP", "localport=443")
                if ($add443 -ne 0) { Write-LogWarn "Failed to add HTTPS 443 firewall rule (Administrator rights may be required)." }
            }
        }
        Write-LogOk "Windows Defender Firewall check/rule configuration finished."
    } else {
        Write-LogInfo "Host firewall untouched (run with --configure-firewall as Administrator to add inbound 80/443 rules)."
    }
}

function Initialize-EnvConfig {
    $envPath = Join-Path $Script:RootDir ".env"
    $existing = Read-DotEnvMap $envPath

    $modeVal = $Script:Mode
    if (-not $modeVal) {
        if ($existing.Contains("APP_ENVIRONMENT") -and ($existing["APP_ENVIRONMENT"] -in @("local", "cloud"))) {
            $modeVal = [string]$existing["APP_ENVIRONMENT"]
        } elseif ($Script:NonInteractive) {
            $modeVal = "local"
        } else {
            Write-Host ""
            Write-Host "Select Deployment Mode:" -ForegroundColor White
            Write-Host "  1) local  (Workstation / localhost HTTP :80) [Default]"
            Write-Host "  2) cloud  (Public Cloud VM with Domain Auto-HTTPS or Public IP)"
            $choice = Read-Host "Choice [1/2]"
            $modeVal = if ($choice -eq "2" -or $choice -eq "cloud") { "cloud" } else { "local" }
        }
    }
    if ($modeVal -notin @("local", "cloud")) {
        Fail-Stage "Environment Setup" 2 "Invalid mode '$modeVal'. Must be 'local' or 'cloud'." ".\install.bat --help" ""
    }

    if ($Script:VerifyE2E -and $modeVal -eq "cloud") {
        Fail-Stage "Safety Guard" 1 "--verify-e2e is strictly forbidden in Cloud mode to avoid polluting research tables." "Use smoke checks in Cloud mode." "Run without --verify-e2e."
    }

    $existingDomain = if ($existing.Contains("DOMAIN")) { ([string]$existing["DOMAIN"]).Trim() } else { "" }
    $domainVal = $Script:Domain
    if (-not $domainVal -and $modeVal -eq "cloud") { $domainVal = $existingDomain }
    $acmeEmailVal = $Script:AcmeEmail
    if (-not $acmeEmailVal -and $existing.Contains("ACME_EMAIL")) { $acmeEmailVal = ([string]$existing["ACME_EMAIL"]).Trim() }
    $frontendUrl = if ($existing.Contains("FRONTEND_URL") -and $existing["FRONTEND_URL"]) { ([string]$existing["FRONTEND_URL"]).Trim() } else { "http://localhost" }
    $composeFile = "docker-compose.yml"

    if ($modeVal -eq "cloud") {
        if (-not $domainVal -and -not $Script:NonInteractive) {
            $domainInput = Read-Host "Cloud Domain Name for Auto-HTTPS (leave blank for IP-only HTTP :80)"
            $domainVal = $domainInput.Trim()
        }
        if ($domainVal) {
            if ($domainVal -match '^\d{1,3}(\.\d{1,3}){3}$' -or $domainVal.Contains(":")) {
                Fail-Stage "Cloud Domain Validation" 2 "DOMAIN ('$domainVal') appears to be an IP address. Omit --domain for IP-only HTTP mode." ".\install.bat install --mode cloud" "Supply a DNS hostname or omit --domain."
            }
            if (-not $acmeEmailVal -and -not $Script:NonInteractive) {
                $acmeEmailVal = (Read-Host "ACME email for Let's Encrypt alerts (optional, press Enter to skip)").Trim()
            }
            Check-CloudDns $domainVal
            Render-Caddyfile $domainVal $acmeEmailVal
            $frontendUrl = "https://$domainVal"
            $composeFile = "docker-compose.yml:docker-compose.cloud.yml"
        } else {
            Write-LogWarn "Cloud mode without Domain: running HTTP :80 without TLS. Add --domain later for HTTPS."
            if (-not $Script:NonInteractive -and ($frontendUrl -eq "http://localhost")) {
                $pubIp = ""
                try { $pubIp = (Invoke-RestMethod -Uri "https://api.ipify.org" -TimeoutSec 3 -ErrorAction Stop).ToString().Trim() } catch {}
                $suggested = if ($pubIp) { "http://$pubIp" } else { "http://localhost" }
                $ipPrompt = Read-Host "Public IP / Origin for CORS FRONTEND_URL [$suggested]"
                $frontendUrl = if ([string]::IsNullOrWhiteSpace($ipPrompt)) { $suggested } else { $ipPrompt.Trim() }
            }
            $composeFile = "docker-compose.yml"
        }
    } else {
        # Cloud -> Local mode switch cleans stale Domain / Caddy / HTTPS URL automatically
        if ($existingDomain -and $frontendUrl -eq "https://$existingDomain") {
            $frontendUrl = "http://localhost"
        } elseif ($frontendUrl.StartsWith("https://")) {
            $frontendUrl = "http://localhost"
        }
        $domainVal = ""
        $acmeEmailVal = ""
        $composeFile = "docker-compose.yml"
    }

    $pgUser = Test-SecretFileOrEnv "POSTGRES_USER" $existing
    if (-not $pgUser) { $pgUser = "admin" }
    $pgDb = Test-SecretFileOrEnv "POSTGRES_DB" $existing
    if (-not $pgDb) { $pgDb = "cyber_sim" }
    if ($pgUser -notmatch '^[A-Za-z0-9_]+$' -or $pgDb -notmatch '^[A-Za-z0-9_]+$') {
        Fail-Stage "Environment Setup" 2 "POSTGRES_USER/POSTGRES_DB must match [A-Za-z0-9_]+." "" ""
    }

    $pgPass = Test-SecretFileOrEnv "POSTGRES_PASSWORD" $existing
    if (-not $pgPass) {
        if (-not $Script:NonInteractive) {
            $entered = Read-SecretPrompt "PostgreSQL Password (press Enter to auto-generate strong secret)"
            $pgPass = if ([string]::IsNullOrWhiteSpace($entered)) { New-CryptoSafePassword 24 } else { $entered }
        } else {
            $pgPass = New-CryptoSafePassword 24
        }
    }

    $jwtSecret = Test-SecretFileOrEnv "JWT_SECRET" $existing
    if (-not $jwtSecret) {
        $jwtSecret = New-CryptoHex 32
    }

    $adminEmail = Test-SecretFileOrEnv "ADMIN_EMAIL" $existing
    if (-not $adminEmail) {
        if (-not $Script:NonInteractive) {
            $inEmail = Read-Host "Initial Admin Email [admin@soc.local]"
            $adminEmail = if ([string]::IsNullOrWhiteSpace($inEmail)) { "admin@soc.local" } else { $inEmail.Trim() }
        } else {
            $adminEmail = "admin@soc.local"
        }
    }

    $adminNickname = Test-SecretFileOrEnv "ADMIN_NICKNAME" $existing
    if (-not $adminNickname) { $adminNickname = "admin" }

    $adminPass = Test-SecretFileOrEnv "ADMIN_PASSWORD" $existing
    $generatedAdminPassword = ""
    if (-not $adminPass) {
        if (-not $Script:NonInteractive) {
            $enteredAdm = Read-SecretPrompt "Initial Admin Password (press Enter to auto-generate strong password)"
            if ([string]::IsNullOrWhiteSpace($enteredAdm)) {
                $adminPass = New-CryptoSafePassword 18
                $generatedAdminPassword = $adminPass
            } else {
                $adminPass = $enteredAdm
            }
        } else {
            $adminPass = New-CryptoSafePassword 18
            $generatedAdminPassword = $adminPass
        }
    }

    $llmKey = Test-SecretFileOrEnv "LLM_API_KEY" $existing
    $llmUrl = Resolve-ExistingOrDefault "LLM_BASE_URL" $existing "https://api.openai.com/v1"
    $llmModel = Resolve-ExistingOrDefault "LLM_MODEL" $existing "gpt-4o-mini"

    if (-not $Script:NonInteractive -and -not $llmKey) {
        $llmChoice = Read-Host "Configure Live OpenAI-compatible LLM API Key? [y/N - Enter skips into Static Fallback mode]"
        if ($llmChoice -match '^[Yy]') {
            $enteredUrl = Read-Host "LLM Base URL [$llmUrl]"
            if (-not [string]::IsNullOrWhiteSpace($enteredUrl)) { $llmUrl = $enteredUrl.Trim() }
            $enteredModel = Read-Host "LLM Model [$llmModel]"
            if (-not [string]::IsNullOrWhiteSpace($enteredModel)) { $llmModel = $enteredModel.Trim() }
            $enteredKey = Read-SecretPrompt "LLM API Key (hidden input)"
            if (-not [string]::IsNullOrWhiteSpace($enteredKey)) { $llmKey = $enteredKey.Trim() }
        }
    }

    $updates = [ordered]@{
        POSTGRES_USER              = $pgUser
        POSTGRES_PASSWORD          = $pgPass
        POSTGRES_DB                = $pgDb
        JWT_SECRET                 = $jwtSecret
        ADMIN_EMAIL                = $adminEmail
        ADMIN_PASSWORD             = $adminPass
        ADMIN_NICKNAME             = $adminNickname
        APP_ENVIRONMENT            = $modeVal
        GIN_MODE                   = (Resolve-ExistingOrDefault "GIN_MODE" $existing "release")
        ENABLE_SWAGGER             = (Resolve-ExistingOrDefault "ENABLE_SWAGGER" $existing "false")
        BYPASS_POST_TEST_THRESHOLD = (Resolve-ExistingOrDefault "BYPASS_POST_TEST_THRESHOLD" $existing "false")
        BYPASS_THEORY_GATE         = (Resolve-ExistingOrDefault "BYPASS_THEORY_GATE" $existing "false")
        REQUIRE_EMAIL_VERIFICATION = (Resolve-ExistingOrDefault "REQUIRE_EMAIL_VERIFICATION" $existing "true")
        APP_BASE_URL               = $frontendUrl
        FRONTEND_URL               = $frontendUrl
        DOMAIN                     = $domainVal
        ACME_EMAIL                 = $acmeEmailVal
        COMPOSE_PATH_SEPARATOR     = ":"
        COMPOSE_FILE               = $composeFile
        LLM_API_KEY                = $llmKey
        LLM_BASE_URL               = $llmUrl
        LLM_MODEL                  = $llmModel
    }
    Write-MergedEnvFile $updates

    return @{
        Mode               = $modeVal
        Domain             = $domainVal
        FrontendUrl        = $frontendUrl
        PostgresUser       = $pgUser
        PostgresDb         = $pgDb
        AdminEmail         = $adminEmail
        AdminPassword      = $adminPass
        GeneratedAdminPass = $generatedAdminPassword
        LlmConfigured      = (-not [string]::IsNullOrWhiteSpace($llmKey))
        TlsActive          = ($modeVal -eq "cloud" -and -not [string]::IsNullOrWhiteSpace($domainVal))
    }
}

function Wait-ContainerHealthy([string]$ContainerName, [int]$TimeoutSeconds = 120) {
    Write-LogInfo "Waiting for $ContainerName to report healthy..."
    $elapsed = 0
    while ($elapsed -lt $TimeoutSeconds) {
        $res = Invoke-NativeCapture "docker" @("inspect", "--format", "{{if .State.Health}}{{.State.Health.Status}}{{else}}{{.State.Status}}{{end}}", $ContainerName)
        if ($res.ExitCode -eq 0 -and $res.Output -eq "healthy") {
            Write-LogOk "$ContainerName is healthy."
            return
        }
        Start-Sleep -Seconds 2
        $elapsed += 2
        Write-Host "." -NoNewline
    }
    Write-Host ""
    Fail-Stage "Container Readiness ($ContainerName)" 1 "Timed out after ${TimeoutSeconds}s waiting for $ContainerName." "docker compose logs --tail=80" "Inspect logs and fix configuration, then re-run install."
}

function Invoke-SeedPipeline {
    Write-LogInfo "Executing idempotent containerized database seeding..."
    if ($Script:ImageTag) {
        Invoke-DockerCompose @("pull", "seed")
        if ($LASTEXITCODE -ne 0) {
            Fail-Stage "Pull Seeder Image" ([int]$LASTEXITCODE) "Failed to pull tagged seed image ($($Script:ImageTag))." "docker compose pull seed" "Verify image tag exists or omit --image-tag to build from source."
        }
        Invoke-DockerCompose @("up", "--no-build", "--no-deps", "--force-recreate", "--exit-code-from", "seed", "seed")
    } else {
        Invoke-DockerCompose @("up", "--build", "--no-deps", "--force-recreate", "--exit-code-from", "seed", "seed")
    }
    $rc = [int]$LASTEXITCODE
    if ($rc -ne 0) {
        Fail-Stage "Database Seeding" $rc "Containerized seed job failed cleanly before starting backend." "docker compose logs seed" "Resolve the content/DB error above and run '.\install.bat seed'."
    }
    Write-LogOk "Containerized seeding completed cleanly."
}

function Get-DbTableCount([string]$PgUser, [string]$PgDb, [string]$Query) {
    $res = Invoke-NativeCapture "docker" @("compose", "exec", "-T", "db", "psql", "-U", $PgUser, "-d", $PgDb, "-t", "-A", "-c", $Query)
    if ($res.ExitCode -ne 0) { return -1 }
    $val = 0
    if ([int]::TryParse($res.Output, [ref]$val)) { return $val }
    return -1
}

function Invoke-CheckSeed {
    $dbCfg = Get-DbConfig
    $pgUser = $dbCfg.User
    $pgDb = $dbCfg.Db

    Write-LogInfo "Checking PostgreSQL seeded counts and table integrity..."
    $domains   = Get-DbTableCount $pgUser $pgDb "SELECT COUNT(*) FROM security_domains;"
    $questions = Get-DbTableCount $pgUser $pgDb "SELECT COUNT(*) FROM questions;"
    $subtopics = Get-DbTableCount $pgUser $pgDb "SELECT COUNT(*) FROM theory_subtopics;"
    $thTheory  = Get-DbTableCount $pgUser $pgDb "SELECT COUNT(*) FROM theory_translations WHERE locale = 'th';"
    $scenarios = Get-DbTableCount $pgUser $pgDb "SELECT COUNT(*) FROM scenarios;"
    $users     = Get-DbTableCount $pgUser $pgDb "SELECT COUNT(*) FROM users;"
    $sessions  = Get-DbTableCount $pgUser $pgDb "SELECT COUNT(*) FROM simulation_sessions;"

    Write-Host "  security_domains    : $domains (expected >= 5)"
    Write-Host "  questions           : $questions (expected >= 540)"
    Write-Host "  theory_subtopics    : $subtopics (expected >= 28)"
    Write-Host "  theory_translations : $thTheory (expected >= 33)"
    Write-Host "  scenarios           : $scenarios (expected >= 15)"
    Write-Host "  users (runtime)     : $users"
    Write-Host "  simulation_sessions : $sessions"

    if ($domains -lt 5 -or $questions -lt 540 -or $subtopics -lt 28 -or $thTheory -lt 33 -or $scenarios -lt 15) {
        Fail-Stage "Seed Verification" 1 "Database seed verification failed (missing domains/questions/subtopics/theory/scenarios)." ".\install.bat seed" "Run '.\install.bat seed' and recheck."
    }
    Write-LogOk "All seed integrity assertions passed!"
}

function Invoke-SmokeVerify([System.Collections.IDictionary]$Cfg) {
    Write-LogInfo "Running non-polluting post-install smoke checks..."
    Invoke-CheckSeed

    $healthOk = $false
    for ($try = 0; $try -lt 15; $try++) {
        try {
            $health = Invoke-RestMethod -Uri "http://127.0.0.1:8080/api/health" -TimeoutSec 5 -ErrorAction Stop
            if ($health.status -eq "healthy" -and $health.db_status -eq "up") {
                $healthOk = $true
                break
            }
        } catch {}
        Start-Sleep -Seconds 2
    }
    if (-not $healthOk) {
        Fail-Stage "Backend Health Smoke Check" 1 "GET http://127.0.0.1:8080/api/health did not report healthy + db_status=up." "docker compose logs backend" "Inspect backend logs."
    }
    Write-LogOk "Backend Health (/api/health): HEALTHY (db_status=up)"

    $feRes = Invoke-NativeQuiet "docker" @("compose", "exec", "-T", "frontend", "wget", "-q", "-O", "/dev/null", "http://127.0.0.1/")
    if ($feRes -ne 0) {
        Fail-Stage "Frontend Smoke Check" 1 "Frontend internal web check failed." "docker compose logs frontend" "Inspect frontend logs."
    }
    if ($Cfg.TlsActive -and $Cfg.Domain) {
        $caddyRes = Invoke-NativeQuiet "docker" @("exec", "cyber_sim_caddy", "caddy", "version")
        if ($caddyRes -ne 0) {
            Fail-Stage "Caddy TLS Proxy Check" 1 "cyber_sim_caddy container is not responding cleanly." "docker compose logs caddy" "Inspect Caddy logs."
        }
        try {
            $httpsResp = Invoke-WebRequest -Uri "https://$($Cfg.Domain)/" -TimeoutSec 6 -UseBasicParsing -ErrorAction Stop
            Write-LogOk "Cloud HTTPS Endpoint (https://$($Cfg.Domain)/): HTTP $($httpsResp.StatusCode)"
        } catch {
            Write-LogWarn "HTTPS endpoint https://$($Cfg.Domain)/ is still issuing/awaiting DNS propagation ($($_.Exception.Message)). Internal Caddy & Frontend are running."
        }
    } else {
        Write-LogOk "Frontend Web UI: READY"
    }

    if ($Cfg.AdminEmail -and $Cfg.AdminPassword) {
        try {
            $body = @{ email = $Cfg.AdminEmail; password = $Cfg.AdminPassword } | ConvertTo-Json -Compress
            $loginResp = Invoke-RestMethod -Uri "http://127.0.0.1:8080/api/login" -Method Post -ContentType "application/json" -Body $body -TimeoutSec 8 -ErrorAction Stop
            if (-not $loginResp.token) {
                Fail-Stage "Admin Auth Smoke Check" 1 "Login returned 200 without JWT token." "docker compose logs backend" "Verify JWT_SECRET and ADMIN_* settings."
            }
            Write-LogOk "Admin Authentication Smoke Check: VERIFIED ($($Cfg.AdminEmail))"
        } catch {
            Fail-Stage "Admin Auth Smoke Check" 1 "Admin login failed during verification: $_" "docker compose logs backend" "Verify ADMIN_EMAIL and ADMIN_PASSWORD in .env."
        }
    }

    if ($Script:VerifyE2E) {
        if ($Cfg.Mode -eq "cloud") {
            Fail-Stage "Safety Guard" 1 "--verify-e2e cannot run in Cloud mode." "" ""
        }
        $py = Get-Command python -ErrorAction SilentlyContinue
        if (-not $py) {
            Fail-Stage "E2E Verification" 1 "python is required for optional --verify-e2e integration harness." "python --version" "Install Python + psycopg2-binary/requests or omit --verify-e2e."
        }
        Write-LogInfo "Executing scripts/test_integration.py against http://127.0.0.1:8080/api ..."
        & python -u (Join-Path $Script:RootDir "scripts\test_integration.py") "http://127.0.0.1:8080/api"
        if ($LASTEXITCODE -ne 0) {
            Fail-Stage "E2E Verification" ([int]$LASTEXITCODE) "test_integration.py failed." "python scripts\test_integration.py http://127.0.0.1:8080/api" "Fix reported endpoint issue and retry."
        }
    }
}

function Invoke-Backup {
    $dbCfg = Get-DbConfig
    $pgUser = $dbCfg.User
    $pgDb = $dbCfg.Db

    Wait-ContainerHealthy "cyber_sim_db" 60

    $backupDir = Join-Path $Script:RootDir "backups"
    if (-not (Test-Path $backupDir)) { New-Item -ItemType Directory -Path $backupDir | Out-Null }
    $ts = (Get-Date).ToString("yyyyMMdd_HHmmss")
    $sqlGzPath = Join-Path $backupDir "cyber_sim_${ts}.sql.gz"

    Write-LogInfo "Backing up PostgreSQL database '$pgDb' to $sqlGzPath ..."
    $psi = New-Object System.Diagnostics.ProcessStartInfo
    $psi.FileName = "docker"
    $psi.Arguments = "compose exec -T db pg_dump --no-owner --no-privileges -U `"$pgUser`" `"$pgDb`""
    $psi.WorkingDirectory = $Script:RootDir
    $psi.RedirectStandardOutput = $true
    $psi.RedirectStandardError = $true
    $psi.UseShellExecute = $false
    $proc = [System.Diagnostics.Process]::Start($psi)

    $errTask = $proc.StandardError.ReadToEndAsync()
    $fs = [System.IO.File]::Create($sqlGzPath)
    $gz = New-Object System.IO.Compression.GzipStream($fs, [System.IO.Compression.CompressionMode]::Compress)
    try {
        $proc.StandardOutput.BaseStream.CopyTo($gz)
    } finally {
        $gz.Dispose()
        $fs.Dispose()
    }
    $proc.WaitForExit()
    $stderrOut = $errTask.GetAwaiter().GetResult()
    if ($proc.ExitCode -ne 0 -or ((Get-Item $sqlGzPath).Length -lt 32)) {
        Remove-Item $sqlGzPath -Force -ErrorAction SilentlyContinue
        Fail-Stage "Database Backup" $proc.ExitCode "pg_dump failed: $stderrOut" "docker compose exec db pg_isready -U $pgUser" "Check DB container status."
    }
    $sqlKb = [Math]::Round((Get-Item $sqlGzPath).Length / 1KB, 1)
    Write-LogOk "Created DB backup: $sqlGzPath ($sqlKb KB)"

    if (-not $Script:NoUploads -and ((Invoke-NativeQuiet "docker" @("volume", "inspect", "soc-trainer_uploads")) -eq 0)) {
        $upGzPath = Join-Path $backupDir "cyber_sim_uploads_${ts}.tar.gz"
        $upRes = Invoke-NativeQuiet "docker" @("run", "--rm", "-v", "soc-trainer_uploads:/data:ro", "-v", "${backupDir}:/backup", "alpine:latest", "tar", "-czf", "/backup/cyber_sim_uploads_${ts}.tar.gz", "-C", "/data", ".")
        if ($upRes -eq 0) {
            Write-LogOk "Created uploads archive: $upGzPath"
        } else {
            Write-LogWarn "Uploads volume archive failed cleanly; SQL dump remains valid."
        }
    }

    $allSql = @(Get-ChildItem -Path $backupDir -Filter "cyber_sim_*.sql.gz" -File | Sort-Object Name -Descending)
    if ($allSql.Count -gt 10) {
        for ($k = 10; $k -lt $allSql.Count; $k++) {
            $oldSql = $allSql[$k]
            if ($oldSql.Name -match '^cyber_sim_(\d{8}_\d{6})\.sql\.gz$') {
                $oldUp = Join-Path $backupDir ("cyber_sim_uploads_" + $Matches[1] + ".tar.gz")
                if (Test-Path $oldUp) { Remove-Item $oldUp -Force -ErrorAction SilentlyContinue }
            }
            Remove-Item $oldSql.FullName -Force -ErrorAction SilentlyContinue
        }
        Write-LogInfo "Pruned old backups beyond newest 10 sets."
    }
}

function Invoke-Restore([string]$BackupFile) {
    $resolvedInput = Resolve-CallerPath $BackupFile
    if ([string]::IsNullOrWhiteSpace($resolvedInput) -or -not (Test-Path $resolvedInput -PathType Leaf)) {
        Fail-Stage "Restore Validation" 2 "Specify a valid existing .sql.gz or .sql backup file path." ".\install.bat restore backups\cyber_sim_<timestamp>.sql.gz" ""
    }
    $fullFile = (Resolve-Path $resolvedInput).Path
    if ((Get-Item $fullFile).Length -lt 16) {
        Fail-Stage "Restore Validation" 2 "Backup file appears empty: $fullFile" "" ""
    }

    $dbCfg = Get-DbConfig
    $pgUser = $dbCfg.User
    $pgDb = $dbCfg.Db
    $envMap = $dbCfg.EnvMap

    if ($Script:NonInteractive) {
        Fail-Stage "Restore Guard" 1 "Destructive restore requires typing 'RESTORE' interactively (blocked in --non-interactive)." "" ""
    }
    Write-LogWarn "Restore will replace all database state in '$pgDb' with contents of $fullFile"
    $typed = Read-Host "Type RESTORE to confirm"
    if ($typed -ne "RESTORE") {
        Write-LogInfo "Restore cancelled."
        return
    }

    Write-LogInfo "Taking pre-restore safety backup first..."
    Invoke-Backup

    Write-LogInfo "Stopping application services prior to restore..."
    $stopServices = @("backend", "frontend")
    if ($envMap.Contains("COMPOSE_FILE") -and ([string]$envMap["COMPOSE_FILE"]).Contains("cloud")) { $stopServices += "caddy" }
    Invoke-DockerCompose (@("stop") + $stopServices)

    Wait-ContainerHealthy "cyber_sim_db" 60

    Write-LogInfo "Recreating clean target database '$pgDb'..."
    $recreateSql = "SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname = '$pgDb' AND pid <> pg_backend_pid(); DROP DATABASE IF EXISTS `"$pgDb`"; CREATE DATABASE `"$pgDb`" OWNER `"$pgUser`";"
    $recRc = Invoke-NativeQuiet "docker" @("compose", "exec", "-T", "db", "psql", "-U", $pgUser, "-d", "postgres", "-v", "ON_ERROR_STOP=1", "-c", $recreateSql)
    if ($recRc -ne 0) {
        Fail-Stage "Database Recreate" $recRc "Could not drop/recreate '$pgDb' prior to restore." "docker compose logs db" "Check active connections and re-run restore."
    }

    Write-LogInfo "Streaming backup archive into PostgreSQL..."
    $psi = New-Object System.Diagnostics.ProcessStartInfo
    $psi.FileName = "docker"
    $psi.Arguments = "compose exec -T db psql -v ON_ERROR_STOP=1 -U `"$pgUser`" -d `"$pgDb`""
    $psi.WorkingDirectory = $Script:RootDir
    $psi.RedirectStandardInput = $true
    $psi.RedirectStandardError = $true
    $psi.UseShellExecute = $false
    $proc = [System.Diagnostics.Process]::Start($psi)
    $errTask = $proc.StandardError.ReadToEndAsync()

    $inFs = [System.IO.File]::OpenRead($fullFile)
    try {
        if ($fullFile.EndsWith(".gz", [System.StringComparison]::OrdinalIgnoreCase)) {
            $gz = New-Object System.IO.Compression.GzipStream($inFs, [System.IO.Compression.CompressionMode]::Decompress)
            try { $gz.CopyTo($proc.StandardInput.BaseStream) } finally { $gz.Dispose() }
        } else {
            $inFs.CopyTo($proc.StandardInput.BaseStream)
        }
    } finally {
        $proc.StandardInput.BaseStream.Close()
        $inFs.Dispose()
    }
    $proc.WaitForExit()
    $stderrOut = $errTask.GetAwaiter().GetResult()
    if ($proc.ExitCode -ne 0) {
        Fail-Stage "Database Restore" $proc.ExitCode "psql restore failed: $stderrOut" "docker compose logs db" "Safety backup created just prior in backups/."
    }
    Write-LogOk "Database restored cleanly."

    if ([System.IO.Path]::GetFileName($fullFile) -match '^cyber_sim_(\d{8}_\d{6})\.sql\.gz$') {
        $candidate = Join-Path ([System.IO.Path]::GetDirectoryName($fullFile)) ("cyber_sim_uploads_" + $Matches[1] + ".tar.gz")
        if (Test-Path $candidate -PathType Leaf) {
            $upFull = (Resolve-Path $candidate).Path
            $upDir = [System.IO.Path]::GetDirectoryName($upFull)
            $upName = [System.IO.Path]::GetFileName($upFull)
            Write-LogInfo "Restoring matching specialist uploads archive: $upFull ..."
            $upRc = Invoke-NativeQuiet "docker" @("run", "--rm", "-e", "UP_NAME=$upName", "-v", "soc-trainer_uploads:/data", "-v", "${upDir}:/backup:ro", "alpine:latest", "sh", "-c", 'find /data -mindepth 1 -delete && tar -xzf "/backup/$UP_NAME" -C /data && chown -R 65534:65534 /data')
            if ($upRc -eq 0) { Write-LogOk "Uploads volume restored cleanly." }
            else { Write-LogWarn "Uploads volume restore exited non-zero; inspect $upFull manually." }
        }
    }

    Invoke-SeedPipeline
    Invoke-DockerCompose @("up", "-d", "--remove-orphans")
    Wait-ContainerHealthy "cyber_sim_backend" 120
    Write-LogOk "Restore sequence completed and application stack restarted."
}

function Invoke-InstallPipeline {
    Invoke-Preflight $true
    $cfg = Initialize-EnvConfig

    if ($cfg.Mode -eq "cloud") {
        Audit-WindowsFirewall $true $cfg.TlsActive
    } else {
        Audit-WindowsFirewall $false $false
    }

    Write-LogInfo "Starting PostgreSQL database..."
    Invoke-DockerCompose @("up", "-d", "db")
    $rcDb = [int]$LASTEXITCODE
    if ($rcDb -ne 0) {
        Fail-Stage "Start Database" $rcDb "docker compose up -d db failed." "docker compose logs db" "Verify Docker daemon & .env secrets."
    }
    Wait-ContainerHealthy "cyber_sim_db" 120

    Invoke-SeedPipeline

    Write-LogInfo "Starting backend, uploads-init, frontend$(if ($cfg.TlsActive) { ', caddy' })..."
    if ($Script:ImageTag) {
        Invoke-DockerCompose @("pull", "backend", "frontend")
        if ($LASTEXITCODE -ne 0) {
            Fail-Stage "Pull Application Images" ([int]$LASTEXITCODE) "Failed to pull tagged images ($($Script:ImageTag))." "docker compose pull" "Verify image tag exists or omit --image-tag."
        }
        Invoke-DockerCompose @("up", "-d", "--no-build", "--remove-orphans")
    } else {
        Invoke-DockerCompose @("up", "-d", "--build", "--remove-orphans")
    }
    $rcApp = [int]$LASTEXITCODE
    if ($rcApp -ne 0) {
        Fail-Stage "Start Application Stack" $rcApp "docker compose up -d failed." "docker compose logs" "Fix reported build/runtime issue and re-run install."
    }

    Wait-ContainerHealthy "cyber_sim_backend" 120
    Wait-ContainerHealthy "cyber_sim_frontend" 90
    if ($cfg.TlsActive) {
        Write-LogInfo "Caddy TLS container is active on :80/:443 for $($cfg.Domain)."
    }

    Invoke-SmokeVerify $cfg

    Write-Host ""
    Write-Host "========================================================" -ForegroundColor Green
    Write-LogOk "SOC Trainer Installed & Verified ($($cfg.Mode) mode)"
    Write-Host "========================================================" -ForegroundColor Green
    Write-Host "  Web Frontend     : $($cfg.FrontendUrl)"
    Write-Host "  Backend API      : http://127.0.0.1:8080/api"
    Write-Host "  Health Endpoint  : http://127.0.0.1:8080/api/health (HEALTHY / DB UP)"
    Write-Host "  LLM Mode         : $(if ($cfg.LlmConfigured) { 'Live AI Configured (key redacted)' } else { 'Static Fallback Mode (key omitted)' })"
    Write-Host "  Secrets (.env)   : Locked owner-only (DB/JWT/LLM secrets redacted)"
    if ($cfg.GeneratedAdminPass -and -not $Script:NonInteractive) {
        Write-Host ""
        Write-Host "  Initial Admin Login (shown ONCE — please save & change password):" -ForegroundColor Yellow
        Write-Host "    Email          : $($cfg.AdminEmail)" -ForegroundColor Yellow
        Write-Host "    Password       : $($cfg.GeneratedAdminPass)" -ForegroundColor Yellow
    }
    Write-Host "========================================================" -ForegroundColor Green
}

function Invoke-UpdatePipeline {
    Invoke-Preflight $false
    Write-LogInfo "Starting safe update pipeline (auto-backup -> optional pull -> rebuild/pull -> seed -> health verify)..."
    Invoke-Backup

    if (-not $Script:NoPull -and (Test-Path (Join-Path $Script:RootDir ".git")) -and (Get-Command git -ErrorAction SilentlyContinue)) {
        $dirty = (Invoke-NativeCapture "git" @("status", "--porcelain")).Output
        if ($dirty) {
            Write-LogWarn "Working tree has local changes; skipping git pull (commit/stash or pass --no-pull)."
        } else {
            Write-LogInfo "Pulling latest updates (git pull --ff-only)..."
            $pullRc = Invoke-NativeQuiet "git" @("pull", "--ff-only")
            if ($pullRc -ne 0) {
                Write-LogWarn "git pull --ff-only did not advance; continuing with current working tree."
            }
        }
    }

    Invoke-DockerCompose @("up", "-d", "db")
    Wait-ContainerHealthy "cyber_sim_db" 120
    Invoke-SeedPipeline

    if ($Script:ImageTag) {
        Invoke-DockerCompose @("pull", "backend", "frontend")
        if ($LASTEXITCODE -ne 0) {
            Fail-Stage "Pull Application Images" ([int]$LASTEXITCODE) "Image pull failed for tag '$($Script:ImageTag)'." "docker compose pull" "Pre-update backup saved in backups/."
        }
        Invoke-DockerCompose @("up", "-d", "--no-build", "--remove-orphans")
    } else {
        Invoke-DockerCompose @("up", "-d", "--build", "--remove-orphans")
    }
    $rc = [int]$LASTEXITCODE
    if ($rc -ne 0) {
        Fail-Stage "Update Application Stack" $rc "Container update failed." "docker compose logs" "Inspect error or restore pre-update backup in backups/."
    }

    Wait-ContainerHealthy "cyber_sim_backend" 120
    Wait-ContainerHealthy "cyber_sim_frontend" 90
    $envMap = Ensure-ComposeProcessEnv
    $dom = if ($envMap.Contains("DOMAIN")) { [string]$envMap["DOMAIN"] } else { "" }
    $modeVal = if ($envMap.Contains("APP_ENVIRONMENT")) { [string]$envMap["APP_ENVIRONMENT"] } else { "local" }
    $cfg = @{
        Mode          = $modeVal
        Domain        = $dom
        TlsActive     = ($modeVal -eq "cloud" -and -not [string]::IsNullOrWhiteSpace($dom))
        AdminEmail    = if ($envMap.Contains("ADMIN_EMAIL")) { [string]$envMap["ADMIN_EMAIL"] } else { "" }
        AdminPassword = if ($envMap.Contains("ADMIN_PASSWORD")) { [string]$envMap["ADMIN_PASSWORD"] } else { "" }
    }
    Invoke-SmokeVerify $cfg
    Write-LogOk "Stack updated and verified cleanly!"
}

function Invoke-Status {
    $envMap = Ensure-ComposeProcessEnv
    $modeVal = if ($envMap.Contains("APP_ENVIRONMENT") -and $envMap["APP_ENVIRONMENT"]) { [string]$envMap["APP_ENVIRONMENT"] } else { "unconfigured" }
    $urlVal = if ($envMap.Contains("FRONTEND_URL") -and $envMap["FRONTEND_URL"]) { [string]$envMap["FRONTEND_URL"] } else { "http://localhost" }
    $llmSet = ($envMap.Contains("LLM_API_KEY") -and -not [string]::IsNullOrWhiteSpace([string]$envMap["LLM_API_KEY"]))
    $tlsSet = ($envMap.Contains("DOMAIN") -and -not [string]::IsNullOrWhiteSpace([string]$envMap["DOMAIN"]))

    Write-Host "SOC Trainer Deployment Status:" -ForegroundColor White
    Write-Host "  Mode / Env   : $modeVal"
    Write-Host "  Frontend URL : $urlVal"
    Write-Host "  Cloud TLS    : $(if ($tlsSet) { "Enabled ($($envMap['DOMAIN']))" } else { 'Disabled (HTTP)' })"
    Write-Host "  LLM Mode     : $(if ($llmSet) { 'Configured (key redacted)' } else { 'Static Fallback' })"
    Write-Host ""
    Invoke-DockerCompose @("ps")
}

function Invoke-Uninstall {
    $envMap = Ensure-ComposeProcessEnv
    $isCloud = ($envMap.Contains("APP_ENVIRONMENT") -and [string]$envMap["APP_ENVIRONMENT"] -eq "cloud")
    if ($Script:RemoveVolumes) {
        if ($isCloud) {
            Write-LogWarn "CRITICAL: Cloud environment detected! Deleting Docker volumes removes PostgreSQL research datasets & uploads permanently."
        } else {
            Write-LogWarn "Deleting Docker volumes permanently removes PostgreSQL data & uploads."
        }
        if ($Script:NonInteractive) {
            Fail-Stage "Uninstall Guard" 1 "Deleting Docker volumes requires interactive typed confirmation ('DELETE DATA')." "" ""
        }
        $typed = Read-Host "Type 'DELETE DATA' to confirm permanent volume removal"
        if ($typed -ne "DELETE DATA") {
            Write-LogInfo "Volume deletion cancelled. Running safe 'docker compose down' (keeping volumes)..."
            Invoke-DockerCompose @("down", "--remove-orphans")
            return
        }
        Invoke-DockerCompose @("down", "-v", "--remove-orphans")
        Write-LogOk "Containers and persistent volumes removed."
        return
    }

    Invoke-DockerCompose @("down", "--remove-orphans")
    Write-LogOk "Containers stopped and removed cleanly (persistent volumes preserved)."
}

function Invoke-ResetDevData {
    $dbCfg = Get-DbConfig
    $envMap = $dbCfg.EnvMap
    $modeVal = if ($envMap.Contains("APP_ENVIRONMENT")) { ([string]$envMap["APP_ENVIRONMENT"]).ToLowerInvariant() } else { "local" }
    $depMode = if ($envMap.Contains("DEPLOYMENT_MODE")) { ([string]$envMap["DEPLOYMENT_MODE"]).ToLowerInvariant() } else { "" }

    if ($Script:Mode -eq "cloud" -or $modeVal -eq "cloud" -or $depMode -eq "cloud") {
        Fail-Stage "Reset Dev Data Guard" 1 "ABORTED: reset-dev-data is strictly prohibited in CLOUD environment." "" "Switch to a local development workspace for destructive resets."
    }
    if ($Script:NonInteractive) {
        Fail-Stage "Reset Dev Data Guard" 1 "Destructive reset requires interactive confirmation ('RESET')." "" ""
    }

    Write-LogWarn "=== DANGER ZONE: RESET DEV DATA ==="
    Write-LogWarn "This truncates learner sessions, questionnaires, notifications, threads, and NFR runs."
    $confirm = Read-Host "Type RESET to confirm"
    if ($confirm -ne "RESET") {
        Write-LogInfo "Operation cancelled."
        return
    }

    $rc = Invoke-NativeQuiet "docker" @("compose", "exec", "-T", "db", "psql", "-U", $dbCfg.User, "-d", $dbCfg.Db, "-v", "ON_ERROR_STOP=1", "-c", $Script:ResetTruncateSql)
    if ($rc -ne 0) {
        Fail-Stage "Reset Dev Data" $rc "Could not truncate runtime tables." "docker compose logs db" ""
    }
    Write-LogOk "Runtime tables truncated. Re-running idempotent seed..."
    Invoke-SeedPipeline
}

function Show-InteractiveMenu {
    while ($true) {
        Write-Host ""
        Write-Host "===============================================" -ForegroundColor Cyan
        Write-Host " SOC Trainer — Windows Setup & Operations Menu" -ForegroundColor Cyan
        Write-Host "===============================================" -ForegroundColor Cyan
        Write-Host "  1) Install / Setup (Local Mode)"
        Write-Host "  2) Install / Setup (Cloud Mode + Auto-HTTPS)"
        Write-Host "  3) Preflight Check (check)"
        Write-Host "  4) Service Status (status)"
        Write-Host "  5) Verify Seed Data (check-seed)"
        Write-Host "  6) Backup Database + Uploads (backup)"
        Write-Host "  7) Restore from Backup (restore)"
        Write-Host "  8) Safe Update (update)"
        Write-Host "  9) View Container Logs (logs)"
        Write-Host " 10) Stop / Uninstall Stack (uninstall)"
        Write-Host " 11) Reset Dev Runtime Data (local only)"
        Write-Host "  0) Exit"
        $sel = (Read-Host "Select option [0-11]").Trim()
        switch ($sel) {
            "1" { $Script:Mode = "local"; Invoke-InstallPipeline }
            "2" { $Script:Mode = "cloud"; Invoke-InstallPipeline }
            "3" { Invoke-Preflight $false }
            "4" { Invoke-Status }
            "5" { Invoke-CheckSeed }
            "6" { Invoke-Backup }
            "7" {
                $pathIn = Read-Host "Path to .sql.gz backup file"
                Invoke-Restore $pathIn
            }
            "8" { Invoke-UpdatePipeline }
            "9" {
                $svc = (Read-Host "Service name (Enter for all)").Trim()
                if ($svc) { Invoke-DockerCompose @("logs", "--tail=100", $svc) }
                else { Invoke-DockerCompose @("logs", "--tail=80") }
            }
            "10" { Invoke-Uninstall }
            "11" { Invoke-ResetDevData }
            "0" { return }
            default { Write-LogWarn "Invalid selection: $sel" }
        }
    }
}

# Dispatcher
if (-not $Script:Subcommand) {
    if ($Script:NonInteractive -or -not [Environment]::UserInteractive) {
        $Script:Subcommand = "install"
    } else {
        Show-InteractiveMenu
        exit 0
    }
}

switch ($Script:Subcommand) {
    "install"        { Invoke-InstallPipeline }
    "check"          { Invoke-Preflight $false }
    "status"         { Invoke-Status }
    "seed"           { Ensure-ComposeProcessEnv | Out-Null; Invoke-SeedPipeline }
    "check-seed"     { Invoke-CheckSeed }
    "update"         { Invoke-UpdatePipeline }
    "backup"         { Invoke-Backup }
    "restore"        { Invoke-Restore $Script:RestoreFile }
    "logs"           {
        if ($Script:LogsService) { Invoke-DockerCompose @("logs", "-f", $Script:LogsService) }
        else { Invoke-DockerCompose @("logs", "-f") }
        exit $LASTEXITCODE
    }
    "uninstall"      { Invoke-Uninstall }
    "reset-dev-data" { Invoke-ResetDevData }
    default {
        Show-Usage
        Fail-Stage "Command Dispatch" 2 "Unknown subcommand: $($Script:Subcommand)" ".\install.bat --help" ""
    }
}
