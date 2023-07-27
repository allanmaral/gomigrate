<#
.SYNOPSIS
    gomigrate installer.
.DESCRIPTION
    The installer of gomigrate.
.PARAMETER GomigrateDir
    Specifies gomigrate root path.
    If not specified, gomigrate will be installed to '$env:USERPROFILE\gomigrate'.
.LINK
    https://github.com/allanmaral/gomigrate
#>
param(
    [String] $GomigrateDir,
)

# Disable StrictMode in this script
Set-StrictMode -Off

function Write-InstallInfo {
    param(
        [Parameter(Mandatory = $True, Position = 0)]
        [String] $String,
        [Parameter(Mandatory = $False, Position = 1)]
        [System.ConsoleColor] $ForegroundColor = $host.UI.RawUI.ForegroundColor
    )

    $backup = $host.UI.RawUI.ForegroundColor

    if ($ForegroundColor -ne $host.UI.RawUI.ForegroundColor) {
        $host.UI.RawUI.ForegroundColor = $ForegroundColor
    }

    Write-Output "$String"

    $host.UI.RawUI.ForegroundColor = $backup
}

function Deny-Install {
    param(
        [String] $message,
        [Int] $errorCode = 1
    )

    Write-InstallInfo -String $message -ForegroundColor DarkRed
    Write-InstallInfo "Abort."

    # Don't abort if invoked with iex that would close the PS session
    if ($IS_EXECUTED_FROM_IEX) {
        break
    } else {
        exit $errorCode
    }
}

function Test-IsAdministrator {
    return ([Security.Principal.WindowsPrincipal]`
            [Security.Principal.WindowsIdentity]::GetCurrent()`
    ).IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator) -and $env:USERNAME -ne 'WDAGUtilityAccount'
}

function Test-Prerequisite {
    # Detect if RunAsAdministrator, there is no need to run as administrator when installing gomigrate.
    if ((Test-IsAdministrator)) {
        Deny-Install "Running the installer as administrator is disabled."
    }

    # Show notification to change execution policy
    $allowedExecutionPolicy = @('Unrestricted', 'RemoteSigned', 'ByPass')
    if ((Get-ExecutionPolicy).ToString() -notin $allowedExecutionPolicy) {
        Deny-Install "PowerShell requires an execution policy in [$($allowedExecutionPolicy -join ", ")] to run gomigrate. For example, to set the execution policy to 'RemoteSigned' please run 'Set-ExecutionPolicy RemoteSigned -Scope CurrentUser'."
    }
}

function Test-isFileLocked {
    param(
        [String] $path
    )

    $file = New-Object System.IO.FileInfo $path

    if (!(Test-Path $path)) {
        return $false
    }

    try {
        $stream = $file.Open(
            [System.IO.FileMode]::Open,
            [System.IO.FileAccess]::ReadWrite,
            [System.IO.FileShare]::None
        )
        if ($stream) {
            $stream.Close()
        }
        return $false
    } catch {
        # The file is locked by a process.
        return $true
    }
}

function Expand-ZipArchive {
    param(
        [String] $path,
        [String] $to
    )

    if (!(Test-Path $path)) {
        Deny-Install "Unzip failed: can't find $path to unzip."
    }

    # Check if the zip file is locked, by antivirus software for example
    $retries = 0
    while ($retries -le 10) {
        if ($retries -eq 10) {
            Deny-Install "Unzip failed: can't unzip because a process is locking the file."
        }
        if (Test-isFileLocked $path) {
            Write-InstallInfo "Waiting for $path to be unlocked by another process... ($retries/10)"
            $retries++
            Start-Sleep -Seconds 2
        } else {
            break
        }
    }

    # Workaround to suspend Expand-Archive verbose output,
    # upstream issue: https://github.com/PowerShell/Microsoft.PowerShell.Archive/issues/98
    $oldVerbosePreference = $VerbosePreference
    $global:VerbosePreference = 'SilentlyContinue'

    # Disable progress bar to gain performance
    $oldProgressPreference = $ProgressPreference
    $global:ProgressPreference = 'SilentlyContinue'

    # PowerShell 5+: use Expand-Archive to extract zip files
    Microsoft.PowerShell.Archive\Expand-Archive -Path $path -DestinationPath $to -Force
    $global:VerbosePreference = $oldVerbosePreference
    $global:ProgressPreference = $oldProgressPreference
}

# See https://stackoverflow.com/a/69239861
function Add-Path {
  param(
    [Parameter(Mandatory, Position=0)]
    [string] $LiteralPath
  )

  $regPath = 'registry::HKEY_CURRENT_USER\Environment'

  # Note the use of the .GetValue() method to ensure that the *unexpanded* value is returned.
  $currDirs = (Get-Item -LiteralPath $regPath).GetValue('Path', '', 'DoNotExpandEnvironmentNames') -split ';' -ne ''

  if ($LiteralPath -in $currDirs) {
    Write-InstallInfo "Already present in the persistent user-level Path: $LiteralPath."
    return
  }

  $newValue = ($currDirs + $LiteralPath) -join ';'

  # Update the registry.
  Set-ItemProperty -Type ExpandString -LiteralPath $regPath Path $newValue

  # Broadcast WM_SETTINGCHANGE to get the Windows shell to reload the
  # updated environment, via a dummy [Environment]::SetEnvironmentVariable() operation.
  $dummyName = [guid]::NewGuid().ToString()
  [Environment]::SetEnvironmentVariable($dummyName, 'foo', 'User')
  [Environment]::SetEnvironmentVariable($dummyName, [NullString]::value, 'User')

  # Finally, also update the current session's `$env:Path` definition.
  # Note: For simplicity, we always append to the in-process *composite* value,
  #        even though for a -Scope Machine update this isn't strictly the same.
  $env:Path = ($env:Path -replace ';$') + ';' + $LiteralPath

  Write-InstallInfo "`"$LiteralPath`" successfully appended to the persistent user-level Path and also the current-process value."
}

function Install-Gomigrate {
    Write-InstallInfo "Initializing..."
    # Check prerequisites
    Test-Prerequisite

    # Download gomigrate from GitHub
    Write-InstallInfo "Determining latest release..."
    $releases = "https://api.github.com/repos/$GOMIGRATE_REPO/releases"
    [Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12
    $tag = (Invoke-WebRequest -Uri $releases -UseBasicParsing | ConvertFrom-Json)[0].tag_name

    $downloadUrl = "https://github.com/$GOMIGRATE_REPO/releases/download/$tag/$GOMIGRATE_PACKAGE_FILE"
    $gomigrateZipfile = "$GOMIGRATE_DIR\gomigrate.zip"
    if (!(Test-Path $GOMIGRATE_DIR)) {
        New-Item -Type Directory $GOMIGRATE_DIR | Out-Null
    }

    Write-InstallInfo "Downloading $downloadUrl to $gomigrateZipfile"
    [Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12
    Invoke-WebRequest $download -Out $gomigrateZipfile

    Write-InstallInfo "Extracting..."
    $gomigrateUnzipTempDir = "$GOMIGRATE_DIR\_tmp"
    Expand-ZipArchive $gomigrateZipfile $gomigrateUnzipTempDir
    Copy-Item "$gomigrateUnzipTempDir\*" $GOMIGRATE_DIR -Recurse -Force

    # Cleanup
    Remove-Item $gomigrateUnzipTempDir -Recurse -Force
    Remove-Item $gomigrateZipfile

    # Finially ensure gomigrate is in the PATH
    # Add-Path $GOMIGRATE_DIR

    Write-InstallInfo "gomigrate was installed successfully!" -ForegroundColor DarkGreen
    Write-InstallInfo "Type 'gomigrate --help' for instructions."
}

# Prepare variables
$IS_EXECUTED_FROM_IEX = ($null -eq $MyInvocation.MyCommand.Path)

# gomigrate root directory
$GOMIGRATE_DIR = $GomigrateDir, $env:GOMIGRATE, "$env:USERPROFILE\gomigrate" | Where-Object { -not [String]::IsNullOrEmpty($_) } | Select-Object -First 1

$GOMIGRATE_REPO = "allanmaral/gomigrate"
$GOMIGRATE_PACKAGE_FILE = "gomigrate_Windows_x86_64.zip"

# Quit if anything goes wrong
$oldErrorActionPreference = $ErrorActionPreference
$ErrorActionPreference = 'Stop'

# Bootstrap function
Install-Gomigrate

# Reset $ErrorActionPreference to original value
$ErrorActionPreference = $oldErrorActionPreference