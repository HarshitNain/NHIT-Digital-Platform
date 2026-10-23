param (

    [switch]$Fast,

    [switch]$Bail,

    [int]$Phase = 0,

    [switch]$Json,

    [string]$Html = "",

    [string]$Markdown = "",

    [switch]$VerboseOutput

)



$ErrorActionPreference = "Stop"



$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path

$RepoRoot = Split-Path -Parent $ScriptDir

$HarnessPy = Join-Path $RepoRoot "NHIT Backend\scripts\harness.py"

$EnvFile = Join-Path $ScriptDir "configs\test.env"

$ReportsDir = Join-Path $ScriptDir "reports"



if (-not (Test-Path $ReportsDir)) {

    New-Item -ItemType Directory -Path $ReportsDir -Force | Out-Null

}



if (-not $Html) {

    $Html = Join-Path $ReportsDir "harness_report.html"

}

if (-not $Markdown) {

    $Markdown = Join-Path $ReportsDir "harness_report.md"

}



if (Test-Path $EnvFile) {

    Get-Content $EnvFile | ForEach-Object {

        $line = $_.Trim()

        if ($line -and -not ($line.StartsWith("#")) -and ($line.Contains("="))) {

            $parts = $line.Split("=", 2)

            $k = $parts[0].Trim()

            $v = $parts[1].Trim().Trim('"').Trim("'")

            if ($k -and -not [System.Environment]::GetEnvironmentVariable($k)) {

                [System.Environment]::SetEnvironmentVariable($k, $v, [System.EnvironmentVariableTarget]::Process)

            }

        }

    }

}



if (-not $Json) {

    Write-Host ""

    Write-Host "======================================================================" -ForegroundColor Cyan

    Write-Host "  NHIT DIGITAL PLATFORM - ENTERPRISE VERIFICATION HARNESS" -ForegroundColor White

    Write-Host "======================================================================" -ForegroundColor Cyan

    Write-Host "  Harness Engine : $HarnessPy" -ForegroundColor Gray

    Write-Host "  HTML Report    : $Html" -ForegroundColor Gray

    Write-Host "  Markdown Report: $Markdown" -ForegroundColor Gray

    if ($Fast) { Write-Host "  Execution Mode : FAST" -ForegroundColor Yellow }

    if ($Bail) { Write-Host "  Failure Policy : FAIL-FAST" -ForegroundColor Red }

    if ($Phase -gt 0) { Write-Host "  Target Phase   : Phase $Phase" -ForegroundColor Magenta }

    Write-Host "----------------------------------------------------------------------" -ForegroundColor Cyan

}



$PythonExe = "python"



$HArgs = @()

$HArgs += $HarnessPy



if ($Fast) {

    $HArgs += "--fast"

}

if ($Bail) {

    $HArgs += "--bail"

}

if ($Phase -gt 0) {

    $HArgs += "--phase"

    $HArgs += "$Phase"

}

if ($Json) {

    $HArgs += "--json"

}

if ($Html) {

    $HArgs += "--html"

    $HArgs += $Html

}

if ($Markdown) {

    $HArgs += "--markdown"

    $HArgs += $Markdown

}

if ($VerboseOutput) {

    $HArgs += "--verbose"

}



& $PythonExe $HArgs

$ExitCode = $LASTEXITCODE



if (-not $Json) {

    Write-Host ""

    if ($ExitCode -eq 0) {

        Write-Host "======================================================================" -ForegroundColor Green

        Write-Host "  [SUCCESS] All platform checks, invariants and contracts verified!  " -ForegroundColor Green

        Write-Host "======================================================================" -ForegroundColor Green

    } else {

        Write-Host "======================================================================" -ForegroundColor Red

        Write-Host "  [FAILURE] Harness verification failed with exit code $ExitCode      " -ForegroundColor Red

        Write-Host "======================================================================" -ForegroundColor Red

    }

}



exit $ExitCode

