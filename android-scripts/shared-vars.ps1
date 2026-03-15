$sdk = if ($env:ANDROID_HOME) { $env:ANDROID_HOME } else { "F:\android" }

$androidUserHome = if ($env:ANDROID_USER_HOME) { 
    $env:ANDROID_USER_HOME 
} elseif ($env:ANDROID_SDK_HOME) { 
    Join-Path $env:ANDROID_SDK_HOME ".android" 
} else { 
    "$sdk\data\.android" 
}

$env:ANDROID_HOME      = $sdk
$env:ANDROID_USER_HOME = $androidUserHome
$env:ANDROID_SDK_HOME  = Split-Path $androidUserHome -Parent