# android-create-avd.ps1
. "$PSScriptRoot\fix-android-home.ps1"
$sdk = $env:ANDROID_SDK_ROOT
if (-not $sdk) { $sdk = "F:\android" }

& "$sdk\cmdline-tools\latest\bin\avdmanager.bat" create avd --name chromium44 --package "system-images;android-23;google_apis;x86" --device 21 --force
. "$PSScriptRoot\fix-android-home.ps1"