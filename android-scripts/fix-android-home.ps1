# Move .android data if the user has explicitly set a custom location for it
$customUserHome = if ($env:ANDROID_USER_HOME) {
    $env:ANDROID_USER_HOME
} elseif ($env:ANDROID_SDK_HOME) {
    Join-Path $env:ANDROID_SDK_HOME ".android"
} else {
    $null
}

if ($customUserHome) {
    $defaultPath = "$env:USERPROFILE\.android"

    if ((Test-Path $defaultPath) -and ($defaultPath -ne $customUserHome)) {
        Write-Host "Custom ANDROID_USER_HOME is set but data is still in default location."
        Write-Host "Moving $defaultPath -> $customUserHome..."

        if (-not (Test-Path $customUserHome)) {
            New-Item -ItemType Directory -Path $customUserHome | Out-Null
        }

        Get-ChildItem $defaultPath | ForEach-Object {
            $dest = Join-Path $customUserHome $_.Name
            if (Test-Path $dest) { Remove-Item $dest -Recurse -Force }
            Move-Item $_.FullName $dest
        }

        Remove-Item $defaultPath -Force -ErrorAction SilentlyContinue
        Write-Host "Done."
    }

    $env:ANDROID_USER_HOME = $customUserHome
    $env:ANDROID_SDK_HOME  = Split-Path $customUserHome -Parent
}

if (-not $env:ANDROID_HOME) {
    $env:ANDROID_HOME = "F:\android"
}