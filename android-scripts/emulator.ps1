# android-emulator.ps1
. "$PSScriptRoot\fix-android-home.ps1"
$sdk = $env:ANDROID_SDK_ROOT
if (-not $sdk) { $sdk = "F:\android" }

if (-not (Test-Path "$sdk\emulators")) {
    New-Item -ItemType Directory -Path "$sdk\emulators" | Out-Null
}

& "$sdk\emulator\emulator.exe" -avd chromium44 -no-snapshot-load