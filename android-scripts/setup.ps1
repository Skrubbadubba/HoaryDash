. "$PSScriptRoot\fix-android-home.ps1"
. "$PSScriptRoot\shared-vars.ps1"

$sdkmanager = "$sdk\cmdline-tools\latest\bin\sdkmanager.bat"

& $sdkmanager "platform-tools" "emulator" "system-images;android-23;google_apis;x86"
