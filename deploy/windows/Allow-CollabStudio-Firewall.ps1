param(
    [int]$HttpPort = 8080,
    [int]$DiscoveryPort = 9999
)

$ErrorActionPreference = "Stop"

$identity = [Security.Principal.WindowsIdentity]::GetCurrent()
$principal = New-Object Security.Principal.WindowsPrincipal($identity)
$isAdmin = $principal.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)

if (-not $isAdmin) {
    Write-Host "Please run this script as Administrator." -ForegroundColor Yellow
    Write-Host "Right-click the .bat file and choose Run as administrator."
    exit 1
}

$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$serverPath = Join-Path $scriptDir "CollabServer.exe"

if (-not (Test-Path -LiteralPath $serverPath)) {
    $serverPath = Join-Path (Split-Path -Parent $scriptDir) "CollabServer.exe"
}

if (-not (Test-Path -LiteralPath $serverPath)) {
    Write-Host "CollabServer.exe was not found next to this script." -ForegroundColor Red
    Write-Host "Please copy this script into the CollabStudio-Windows folder and run it again."
    exit 1
}

function Ensure-FirewallRule {
    param(
        [string]$Name,
        [string]$Protocol,
        [int]$Port
    )

    $existing = Get-NetFirewallRule -DisplayName $Name -ErrorAction SilentlyContinue
    if ($existing) {
        Write-Host "Firewall rule already exists: $Name"
        return
    }

    New-NetFirewallRule `
        -DisplayName $Name `
        -Direction Inbound `
        -Action Allow `
        -Program $serverPath `
        -Protocol $Protocol `
        -LocalPort $Port `
        -Profile Private,Domain `
        | Out-Null

    Write-Host "Firewall rule created: $Name"
}

Ensure-FirewallRule -Name "CollabStudio Server TCP $HttpPort" -Protocol TCP -Port $HttpPort
Ensure-FirewallRule -Name "CollabStudio Discovery UDP $DiscoveryPort" -Protocol UDP -Port $DiscoveryPort

Write-Host ""
Write-Host "Done. CollabStudio LAN access is allowed for this computer." -ForegroundColor Green
Write-Host "HTTP port: $HttpPort"
Write-Host "Discovery UDP port: $DiscoveryPort"
