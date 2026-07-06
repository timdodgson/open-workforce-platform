# Assist Mode Validation Experiment
# Instance: n012w8 (NRP, SA only)
# Seeds: 42, 123, 555, 777, 999
# Modes: off, shadow, assist
# Total runs: 15
#
# Goal: Prove assist mode is safe before expanding to other domains.

$ErrorActionPreference = "Stop"
$GoDir = "$PSScriptRoot\..\platform\go"
$Seeds = @(42, 123, 555, 777, 999)
$Modes = @("off", "shadow", "assist")
$Instance = "n012w8"
$Algorithm = "sa"

Write-Host "=============================================="
Write-Host "  Assist Mode Validation Experiment"
Write-Host "  Instance: $Instance"
Write-Host "  Algorithm: $Algorithm"
Write-Host "  Seeds: $($Seeds -join ', ')"
Write-Host "  Modes: $($Modes -join ', ')"
Write-Host "  Total runs: $($Seeds.Count * $Modes.Count)"
Write-Host "=============================================="
Write-Host ""

$Results = @()

foreach ($mode in $Modes) {
    foreach ($seed in $Seeds) {
        $label = "nrp-$Algorithm-$mode-s$seed"
        Write-Host "--- Running: $label ---"

        $args = @(
            "tune-pfrs",
            "--instance", $Instance,
            "--pfrs-mode", $Algorithm,
            "--seeds", "$seed",
            "--pfrs-run-label", $label,
            "--pfrs-storage", "s3"
        )

        if ($mode -ne "off") {
            $args += @("--worker-decision-mode", $mode)
        }

        $startTime = Get-Date
        Push-Location $GoDir
        try {
            & go run ./cmd/owp @args 2>&1 | Tee-Object -Variable output
            $exitCode = $LASTEXITCODE
        } finally {
            Pop-Location
        }
        $duration = (Get-Date) - $startTime

        # Parse final objective from output.
        $finalObj = ($output | Select-String "Final Penalty:\s*(\d+)" | ForEach-Object { $_.Matches[0].Groups[1].Value })
        $workersStarted = ($output | Select-String "Workers Started:\s*(\d+)" | ForEach-Object { $_.Matches[0].Groups[1].Value })

        $result = [PSCustomObject]@{
            Label = $label
            Mode = $mode
            Seed = $seed
            FinalObjective = [int]$finalObj
            WorkersStarted = [int]$workersStarted
            DurationSec = [math]::Round($duration.TotalSeconds, 1)
            ExitCode = $exitCode
        }
        $Results += $result

        Write-Host "  Objective: $finalObj | Workers: $workersStarted | Time: $($result.DurationSec)s"
        Write-Host ""
    }
}

# Summary.
Write-Host "=============================================="
Write-Host "  RESULTS SUMMARY"
Write-Host "=============================================="
Write-Host ""

foreach ($mode in $Modes) {
    $modeResults = $Results | Where-Object { $_.Mode -eq $mode }
    $objectives = $modeResults | ForEach-Object { $_.FinalObjective }
    $mean = ($objectives | Measure-Object -Average).Average
    $min = ($objectives | Measure-Object -Minimum).Minimum
    $max = ($objectives | Measure-Object -Maximum).Maximum

    Write-Host "  $($mode.ToUpper()): mean=$([math]::Round($mean, 1)) best=$min worst=$max"
}

Write-Host ""
Write-Host "Done. Check S3 for uploaded results."
Write-Host "Run labels: nrp-sa-{off|shadow|assist}-s{42|123|555|777|999}"
