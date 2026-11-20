<#
.SYNOPSIS
    One-time Windows setup for TouchBridge.

.DESCRIPTION
    Optionally installs ViGEmBus 1.22.0 so TouchBridge can create an Xbox 360
    compatible virtual gamepad, and adds firewall rules that allow the server
    on private networks only.

    The ViGEmBus installer is downloaded from the official Nefarius GitHub
    release and verified against a pinned SHA-256 before it is executed.
    TouchBridge itself never elevates; this setup script requests administrator
    permission only for the driver and firewall changes.

.PARAMETER NoDriver
    Skip the ViGEmBus install. Mouse and keyboard input still work.

.PARAMETER NoFirewall
    Skip the firewall rules. Windows may show a firewall prompt on first run.
#>
[CmdletBinding()]
param(
    [switch]$Elevated,
    [switch]$NoDriver,
    [switch]$NoFirewall
)

$ErrorActionPreference = 'Stop'

$DriverUrl = 'https://github.com/nefarius/ViGEmBus/releases/download/v1.22.0/ViGEmBus_1.22.0_x64_x86_arm64.exe'
$DriverSha256 = '89220a7865076b342892f98865f3499fb7c4cfd673159e89d352c360fd014c6a'
$DriverSize = 6278576
$TcpRuleName = 'TouchBridge server (TCP 8765)'
$UdpRuleName = 'TouchBridge discovery (UDP 8766)'

function Test-Administrator {
    $identity = [Security.Principal.WindowsIdentity]::GetCurrent()
    $principal = New-Object Security.Principal.WindowsPrincipal($identity)
    return $principal.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)
}

function Test-ViGEmBus {
    if (Test-Path 'HKLM:\SYSTEM\CurrentControlSet\Services\ViGEmBus') { return $true }
    if (Test-Path 'HKLM:\SOFTWARE\Nefarius Software Solutions e.U.\ViGEm Bus Driver') { return $true }
    return $null -ne (Get-Service -Name 'ViGEmBus' -ErrorAction SilentlyContinue)
}

function Test-FirewallRule([string]$Name) {
    return $null -ne (Get-NetFirewallRule -DisplayName $Name -ErrorAction SilentlyContinue)
}

function Add-PrivateFirewallRule([string]$Name, [string]$Protocol, [int]$Port) {
    if (Test-FirewallRule $Name) { return }
    if (Get-Command New-NetFirewallRule -ErrorAction SilentlyContinue) {
        New-NetFirewallRule -DisplayName $Name -Direction Inbound -Action Allow `
            -Protocol $Protocol -LocalPort $Port -Profile Private | Out-Null
        return
    }
    netsh advfirewall firewall add rule name="$Name" dir=in action=allow `
        protocol=$Protocol localport=$Port profile=private | Out-Null
}

$needDriver = (-not $NoDriver) -and (-not (Test-ViGEmBus))
$needFirewall = (-not $NoFirewall) -and (
    (-not (Test-FirewallRule $TcpRuleName)) -or (-not (Test-FirewallRule $UdpRuleName))
)

if (-not $needDriver -and -not $needFirewall) {
    Write-Host 'TouchBridge is already set up.'
    Write-Host 'ViGEmBus and the private-network firewall rules are present.'
    exit 0
}

if (-not (Test-Administrator)) {
    Write-Host 'TouchBridge setup needs administrator permission for:'
    if ($needDriver) { Write-Host '  - installing the ViGEmBus driver' }
    if ($needFirewall) { Write-Host '  - adding private-network firewall rules' }
    Write-Host 'Windows will now show a UAC prompt.'
    $arguments = "-NoProfile -ExecutionPolicy Bypass -File `"$PSCommandPath`" -Elevated"
    if ($NoDriver) { $arguments += ' -NoDriver' }
    if ($NoFirewall) { $arguments += ' -NoFirewall' }
    try {
        $process = Start-Process -FilePath 'powershell.exe' -Verb RunAs `
            -ArgumentList $arguments -Wait -PassThru
        if ($process.ExitCode -ne 0) {
            Write-Host 'Setup did not complete. Review the messages in the administrator window.'
        }
        exit $process.ExitCode
    } catch {
        Write-Warning 'Setup was cancelled or administrator permission was refused.'
        exit 1
    }
}

$restartNeeded = $false
$driverInstalled = $false
$setupFailed = $false

if ($needDriver) {
    Write-Host 'Downloading ViGEmBus 1.22.0 from the official release...'
    [Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12
    $tempDir = Join-Path $env:TEMP ('touchbridge-vigem-' + [guid]::NewGuid().ToString('N'))
    New-Item -ItemType Directory -Path $tempDir | Out-Null
    $installer = Join-Path $tempDir 'ViGEmBus_1.22.0_x64_x86_arm64.exe'
    try {
        Invoke-WebRequest -Uri $DriverUrl -OutFile $installer -UseBasicParsing
        if ((Get-Item $installer).Length -ne $DriverSize) {
            throw 'The downloaded installer has an unexpected size.'
        }
        $hash = (Get-FileHash -Algorithm SHA256 -Path $installer).Hash.ToLowerInvariant()
        if ($hash -ne $DriverSha256) {
            throw 'The downloaded installer failed SHA-256 verification.'
        }
        Write-Host 'Checksum verified. Installing ViGEmBus...'
        $install = Start-Process -FilePath $installer `
            -ArgumentList '/exenoui', '/qn', '/norestart' -Wait -PassThru
        if ($install.ExitCode -eq 3010) {
            $restartNeeded = $true
        } elseif ($install.ExitCode -ne 0) {
            throw ("The ViGEmBus installer exited with code {0}." -f $install.ExitCode)
        }
        $driverInstalled = $true
        Write-Host 'ViGEmBus installed.'
    } catch {
        Write-Warning $_.Exception.Message
        Write-Host "The verified installer was kept at: $installer"
        Write-Host 'You can run it manually, then run this setup again if needed.'
        $setupFailed = $true
    }
}

if ($needFirewall) {
    Write-Host 'Adding private-network firewall rules for TCP 8765 and UDP 8766...'
    try {
        Add-PrivateFirewallRule $TcpRuleName 'TCP' 8765
        Add-PrivateFirewallRule $UdpRuleName 'UDP' 8766
    } catch {
        Write-Warning ("Could not add the firewall rules: {0}" -f $_.Exception.Message)
        Write-Host 'Allow TouchBridge on Private networks if Windows asks.'
    }
}

if ($restartNeeded) {
    Write-Host ''
    Write-Host 'Windows may need a restart before the virtual gamepad appears.'
}

Write-Host ''
if ($driverInstalled -or (Test-ViGEmBus)) {
    Write-Host 'Gamepad setup is ready.'
} else {
    Write-Host 'Mouse and keyboard are ready. Gamepad output is unavailable until ViGEmBus is installed.'
}
Write-Host 'Start the server with Start-TouchBridge.bat.'
if ($setupFailed) { exit 1 }
exit 0
