param(
    [switch]$RunGui,
    [switch]$Clean
)

$ErrorActionPreference = "Stop"

$repoRoot = Split-Path -Parent $PSScriptRoot
$binDir = Join-Path $PSScriptRoot "bin"
$examples = @("install", "uninstall", "dialog", "query", "gui")

if ($Clean -and (Test-Path $binDir)) {
    Remove-Item -Path $binDir -Recurse -Force
}

if (-not (Test-Path $binDir)) {
    New-Item -ItemType Directory -Path $binDir | Out-Null
}

Push-Location $repoRoot
try {
    foreach ($example in $examples) {
        $output = Join-Path $binDir ("{0}.exe" -f $example)
        Write-Host "[build] examples/$example -> $output"
        go build -o $output ("./examples/{0}" -f $example)
    }
}
finally {
    Pop-Location
}

Write-Host "Build concluido. Binarios em: $binDir"

if ($RunGui) {
    $guiExe = Join-Path $binDir "gui.exe"
    if (-not (Test-Path $guiExe)) {
        throw "Binario gui.exe nao encontrado em $guiExe"
    }

    Write-Host "Executando GUI Lab..."
    Start-Process -FilePath $guiExe | Out-Null
    Write-Host "Abra: http://127.0.0.1:17841"
}
