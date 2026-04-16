@echo off
setlocal
cd /d "%~dp0"

if not exist "CollabClient.exe" (
  echo CollabClient.exe was not found in this folder.
  pause
  exit /b 1
)

if not exist "CollabServer.exe" (
  echo CollabServer.exe was not found in this folder.
  echo Please keep CollabClient.exe and CollabServer.exe in the same folder.
  pause
  exit /b 1
)

powershell -NoProfile -ExecutionPolicy Bypass -Command ^
  "$ok=$false; try { $r=Invoke-WebRequest -Uri 'http://127.0.0.1:8080/ping' -UseBasicParsing -TimeoutSec 1; $ok=($r.Content -match 'pong') } catch {}; if (-not $ok) { $used=(Get-NetTCPConnection -LocalPort 8080 -State Listen -ErrorAction SilentlyContinue); if ($used) { Write-Host 'Warning: TCP port 8080 is already in use. If CollabStudio cannot connect, close the old process or change the port.' -ForegroundColor Yellow; Start-Sleep -Seconds 3 } }"

start "" "%~dp0CollabClient.exe"
