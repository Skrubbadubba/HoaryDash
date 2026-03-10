$sdk = if ($env:ANDROID_SDK_ROOT) { $env:ANDROID_SDK_ROOT } else { "F:\android" }
$wrongPath = "$env:USERPROFILE\.android"
$correctPath = "$sdk\data"

if (Test-Path $wrongPath) {
    Write-Host "Found .android in user home, moving to $correctPath..."
    if (-not (Test-Path $correctPath)) {
        New-Item -ItemType Directory -Path $correctPath | Out-Null
    }
    Get-ChildItem $wrongPath | ForEach-Object {
        $dest = Join-Path $correctPath $_.Name
        if (Test-Path $dest) {
            Remove-Item $dest -Recurse -Force
        }
        Move-Item $_.FullName $dest
    }
    Remove-Item $wrongPath -Force -ErrorAction SilentlyContinue
    Write-Host "Done."
}