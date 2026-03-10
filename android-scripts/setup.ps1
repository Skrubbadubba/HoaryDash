# android-setup.ps1
$sdk = $env:ANDROID_SDK_ROOT
if (-not $sdk) { $sdk = "F:\android" }

& "$sdk\cmdline-tools\latest\bin\sdkmanager.bat" "platform-tools" "emulator" "system-images;android-23;google_apis;x86"
& "$sdk\cmdline-tools\latest\bin\sdkmanager.bat" --licenses