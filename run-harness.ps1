<#
================================================================================
  NATIONAL HIGHWAYS INFRA TRUST (NHIT) - ROOT HARNESS ENTRYPOINT
================================================================================
  Delegates directly to harness/run-harness.ps1
  Usage:
    .\run-harness.ps1 [-Fast] [-Bail] [-Phase <1-7>] [-Json] [-Html <file>] [-Markdown <file>]
================================================================================
#>

[CmdletBinding()]
param (
    [Parameter(Mandatory=$false)]
    [switch]$Fast,

    [Parameter(Mandatory=$false)]
    [switch]$Bail,

    [Parameter(Mandatory=$false)]
    [ValidateRange(1, 7)]
    [int]$Phase = 0,

    [Parameter(Mandatory=$false)]
    [switch]$Json,

    [Parameter(Mandatory=$false)]
    [string]$Html = "",

    [Parameter(Mandatory=$false)]
    [string]$Markdown = "",

    [Parameter(Mandatory=$false)]
    [switch]$VerboseOutput
)

$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$Runner = Join-Path $ScriptDir "harness\run-harness.ps1"

$PassParams = @{}
if ($Fast) { $PassParams["Fast"] = $true }
if ($Bail) { $PassParams["Bail"] = $true }
if ($Phase -gt 0) { $PassParams["Phase"] = $Phase }
if ($Json) { $PassParams["Json"] = $true }
if ($Html) { $PassParams["Html"] = $Html }
if ($Markdown) { $PassParams["Markdown"] = $Markdown }
if ($VerboseOutput) { $PassParams["VerboseOutput"] = $true }

& $Runner @PassParams
exit $LASTEXITCODE
