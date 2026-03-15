. "$PSScriptRoot\fix-android-home.ps1"
. "$PSScriptRoot\shared-vars.ps1"

$adb = "$sdk\platform-tools\adb.exe"

Write-Host "Waiting for emulator to come online..."
& $adb wait-for-device

# wait-for-device returns as soon as ADB connects, not when Android is fully booted
$booted = ""
while ($booted -ne "1") {
    Start-Sleep 3
    $booted = (& $adb shell getprop sys.boot_completed 2>$null).Trim()
    Write-Host "Waiting for boot... ($booted)"
}

Write-Host "Emulator ready, installing Fully Kiosk..."
$apk = "$env:TEMP\fully-kiosk.apk"
Invoke-WebRequest -Uri "https://www.fully-kiosk.com/files/2026/02/Fully-Kiosk-Browser-v1.60.1.apk" -OutFile $apk
& $adb install $apk

Write-Host "Forwarding port 2323 for remote admin"
& $adb forward tcp:2323 tcp:2323