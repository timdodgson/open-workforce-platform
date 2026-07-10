# SI 2.0 Deep Validation — 48 runs (8 configs × 3 policies × 2 seeds)
# Larger instances + longer budgets → richer checkpoint telemetry for policy training.
# Execute from platform/go:
#   powershell -ExecutionPolicy Bypass -File .\scripts\validate-si2-deep.ps1

$seeds = @(42, 123)
$policies = @("rules", "hybrid", "learned")
$policyDir = "../ml/policies"
$totalRuns = 48
$runIndex = 0
$failures = @()

$cvrpInstance = "../../examples/cvrp/A-n80-k10.vrp"
$cvrpSaIters = 5000000
$cvrpPortfolioIters = 2000000

$jssInstance = "internal/infrastructure/jobshop/testdata/ft10.txt"
$jssTabuIters = 1000000
$jssPortfolioIters = 500000

$vrptwInstance = "../../examples/vrptw/C101.txt"
$vrptwSaIters = 2000000
$vrptwPortfolioIters = 1000000

$nrpInstance = "n012w8"
$nrpItersPerWorker = 300000
$nrpMaxWorkers = 32

$logDir = Join-Path $PSScriptRoot "regression-logs"
New-Item -ItemType Directory -Force -Path $logDir | Out-Null
$logFile = Join-Path $logDir ("validate-si2-deep-{0:yyyyMMdd-HHmmss}.log" -f (Get-Date))

function Write-Log {
    param([string]$Message)
    Add-Content -Path $logFile -Value $Message
}

function Invoke-OwpRun {
    param([string]$Label, [string[]]$OwpArgs)
    $script:runIndex++
    $pct = [math]::Round(100 * $script:runIndex / $totalRuns, 1)
    $ts = Get-Date -Format "yyyy-MM-dd HH:mm:ss"
    $msg = "[$ts] ($runIndex/$totalRuns, $pct%) $Label"
    Write-Host $msg -ForegroundColor Gray
    Write-Log $msg

    if (Test-Path ".\owp.exe") {
        $out = & .\owp.exe @OwpArgs 2>&1
    } else {
        $out = go run ./cmd/owp @OwpArgs 2>&1
    }
    if ($LASTEXITCODE -ne 0) {
        $fail = "FAILED $Label (exit $LASTEXITCODE): $out"
        Write-Host $fail -ForegroundColor Red
        Write-Log $fail
        $script:failures += $Label
    }
}

Write-Host "SI 2.0 Deep Validation - $totalRuns experiments" -ForegroundColor Cyan
Write-Host "Log: $logFile" -ForegroundColor Cyan
Write-Host "  CVRP: A-n80-k10  SA=$cvrpSaIters  Portfolio=$cvrpPortfolioIters" -ForegroundColor DarkGray
Write-Host "  JSS:  ft10       Tabu=$jssTabuIters  Portfolio=$jssPortfolioIters" -ForegroundColor DarkGray
Write-Host "  VRPTW: C101      SA=$vrptwSaIters  Portfolio=$vrptwPortfolioIters" -ForegroundColor DarkGray
Write-Host "  NRP:  n012w8     $nrpItersPerWorker/worker x$nrpMaxWorkers max" -ForegroundColor DarkGray
Write-Host ""

foreach ($policy in $policies) {
    foreach ($seed in $seeds) {
        $label = "val-deep-cvrp-a80k10-sa-${policy}-s${seed}"
        Invoke-OwpRun $label @(
            "solve", "cvrp", "--instance", $cvrpInstance,
            "--mode", "sa", "--iterations", "$cvrpSaIters",
            "--policy-mode", $policy, "--policy-dir", $policyDir,
            "--seed", "$seed", "--run-label", $label, "--storage", "s3"
        )
    }
}

foreach ($policy in $policies) {
    foreach ($seed in $seeds) {
        $label = "val-deep-cvrp-a80k10-portfolio-${policy}-s${seed}"
        Invoke-OwpRun $label @(
            "solve", "cvrp", "--instance", $cvrpInstance,
            "--mode", "portfolio", "--iterations", "$cvrpPortfolioIters",
            "--policy-mode", $policy, "--policy-dir", $policyDir,
            "--seed", "$seed", "--run-label", $label, "--storage", "s3"
        )
    }
}

foreach ($policy in $policies) {
    foreach ($seed in $seeds) {
        $label = "val-deep-jss-ft10-tabu-${policy}-s${seed}"
        Invoke-OwpRun $label @(
            "solve", "jobshop", "--instance", $jssInstance,
            "--mode", "tabu", "--iterations", "$jssTabuIters",
            "--policy-mode", $policy, "--policy-dir", $policyDir,
            "--seed", "$seed", "--run-label", $label, "--storage", "s3"
        )
    }
}

foreach ($policy in $policies) {
    foreach ($seed in $seeds) {
        $label = "val-deep-jss-ft10-portfolio-${policy}-s${seed}"
        Invoke-OwpRun $label @(
            "solve", "jobshop", "--instance", $jssInstance,
            "--mode", "portfolio", "--iterations", "$jssPortfolioIters",
            "--policy-mode", $policy, "--policy-dir", $policyDir,
            "--seed", "$seed", "--run-label", $label, "--storage", "s3"
        )
    }
}

foreach ($policy in $policies) {
    foreach ($seed in $seeds) {
        $label = "val-deep-vrptw-c101-sa-${policy}-s${seed}"
        Invoke-OwpRun $label @(
            "solve", "vrptw", "--instance", $vrptwInstance,
            "--mode", "sa", "--iterations", "$vrptwSaIters",
            "--policy-mode", $policy, "--policy-dir", $policyDir,
            "--seed", "$seed", "--run-label", $label, "--storage", "s3"
        )
    }
}

foreach ($policy in $policies) {
    foreach ($seed in $seeds) {
        $label = "val-deep-vrptw-c101-portfolio-${policy}-s${seed}"
        Invoke-OwpRun $label @(
            "solve", "vrptw", "--instance", $vrptwInstance,
            "--mode", "portfolio", "--iterations", "$vrptwPortfolioIters",
            "--policy-mode", $policy, "--policy-dir", $policyDir,
            "--seed", "$seed", "--run-label", $label, "--storage", "s3"
        )
    }
}

foreach ($policy in $policies) {
    foreach ($seed in $seeds) {
        $label = "val-deep-nrp-n012w8-sa-${policy}-s${seed}"
        Invoke-OwpRun $label @(
            "tune-pfrs", "--instance", $nrpInstance,
            "--pfrs-mode", "sa", "--pfrs-iterations-per-worker", "$nrpItersPerWorker",
            "--pfrs-max-total-workers", "$nrpMaxWorkers", "--seeds", "$seed",
            "--worker-decision-mode", "assist",
            "--policy-mode", $policy, "--policy-dir", $policyDir,
            "--pfrs-run-label", $label, "--storage", "s3"
        )
    }
}

foreach ($policy in $policies) {
    foreach ($seed in $seeds) {
        $label = "val-deep-nrp-n012w8-portfolio-${policy}-s${seed}"
        Invoke-OwpRun $label @(
            "tune-pfrs", "--instance", $nrpInstance,
            "--pfrs-mode", "portfolio", "--pfrs-iterations-per-worker", "$nrpItersPerWorker",
            "--pfrs-max-total-workers", "$nrpMaxWorkers", "--seeds", "$seed",
            "--worker-decision-mode", "assist",
            "--policy-mode", $policy, "--policy-dir", $policyDir,
            "--pfrs-run-label", $label, "--storage", "s3"
        )
    }
}

Write-Host ""
if ($failures.Count -eq 0) {
    Write-Host "Done. $totalRuns deep experiments complete." -ForegroundColor Green
    Write-Log "Done. $totalRuns deep experiments complete."
} else {
    Write-Host "Done with $($failures.Count) failure(s) out of $totalRuns." -ForegroundColor Yellow
    Write-Log "Failures: $($failures -join ', ')"
}
Write-Host "Next: cd ..\ml; python train_policies.py --data-dir ../web/pfrs-lab/data/runs --output-dir policies"
