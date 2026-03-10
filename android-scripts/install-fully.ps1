. "$PSScriptRoot\fix-android-home.ps1"

$sdk = if ($env:ANDROID_SDK_ROOT) { $env:ANDROID_SDK_ROOT } else { "F:\android" }
$adb = "$sdk\platform-tools\adb.exe"

Write-Host "Waiting for emulator to come online..."
& $adb wait-for-device

# Even after wait-for-device the system isn't fully booted, wait for that too
$booted = ""
while ($booted -ne "1") {
    Start-Sleep 3
    $booted = & $adb shell getprop sys.boot_completed 2>$null
    $booted = $booted.Trim()
    Write-Host "Waiting for boot... ($booted)"
}

Write-Host "Emulator ready, installing Fully Kiosk..."
$apk = "$env:TEMP\fully-kiosk.apk"
Invoke-WebRequest -Uri "https://www.fully-kiosk.com/files/2026/02/Fully-Kiosk-Browser-v1.60.1.apk" -OutFile $apk
& $adb install $apk
Write-Host "Forwarding port 2323 for remote admin"
$adb forward tcp:2323 tcp:2323