. "$PSScriptRoot\fix-android-home.ps1"
. "$PSScriptRoot\shared-vars.ps1"

# ANDROID_AVD_HOME is the most specific override for AVD file location
if (-not $env:ANDROID_AVD_HOME) {
    $env:ANDROID_AVD_HOME = Join-Path $androidUserHome "avd"
}

if (-not (Test-Path $env:ANDROID_AVD_HOME)) {
    New-Item -ItemType Directory -Path $env:ANDROID_AVD_HOME | Out-Null
}

& "$sdk\emulator\emulator.exe" -avd chromium44 -no-snapshot-load