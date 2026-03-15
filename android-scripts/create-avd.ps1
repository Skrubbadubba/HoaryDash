. "$PSScriptRoot\fix-android-home.ps1"
. "$PSScriptRoot\shared-vars.ps1"

& "$sdk\cmdline-tools\latest\bin\avdmanager.bat" `
    create avd `
    --name chromium44 `
    --package "system-images;android-23;google_apis;x86" `
    --device 21 `
    --force