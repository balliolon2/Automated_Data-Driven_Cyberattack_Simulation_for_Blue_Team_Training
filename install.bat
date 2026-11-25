@echo off
setlocal
set "SCRIPT_DIR=%~dp0"
set "PS1_PATH=%SCRIPT_DIR%install.ps1"

if not exist "%PS1_PATH%" (
    echo [ERROR] Cannot find install.ps1 next to install.bat ^("%PS1_PATH%"^).
    exit /b 1
)

where pwsh.exe >nul 2>&1
if %ERRORLEVEL%==0 (
    pwsh.exe -NoProfile -ExecutionPolicy Bypass -File "%PS1_PATH%" %*
    exit /b %ERRORLEVEL%
)

where powershell.exe >nul 2>&1
if %ERRORLEVEL%==0 (
    powershell.exe -NoProfile -ExecutionPolicy Bypass -File "%PS1_PATH%" %*
    exit /b %ERRORLEVEL%
)

echo [ERROR] Neither PowerShell 7 ^(pwsh.exe^) nor Windows PowerShell ^(powershell.exe^) was found on PATH.
exit /b 1
