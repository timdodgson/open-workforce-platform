# Search Intelligence 2.0 — Live Validation Script
# Runs all 240 experiments (8 configs × 3 policies × 10 seeds)
# Execute from platform/go directory

$seeds = @(42, 123, 555, 777, 999, 1001, 2022, 3033, 4044, 5055)
$policies = @("rules", "hybrid", "learned")
$policyDir = "../ml/policies"
$totalRuns = 240
$runIndex = 0
$failures = @()

$logDir = Join-Path $PSScriptRoot "regression-logs"
New-Item -ItemType Directory -Force -Path $logDir | Out-Null
$logFile = Join-Path $logDir ("validate-si2-{0:yyyyMMdd-HHmmss}.log" -f (Get-Date))

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

Write-Host "SI 2.0 Live Validation - $totalRuns experiments" -ForegroundColor Cyan
Write-Host "Log: $logFile" -ForegroundColor Cyan
Write-Host ""

# CVRP SA
foreach ($policy in $policies) {
    foreach ($seed in $seeds) {
        $label = "val-cvrp-a32k5-sa-${policy}-s${seed}"
        Invoke-OwpRun $label @(
            "solve-cvrp", "--instance", "../../examples/cvrp/A-n32-k5.vrp",
            "--mode", "sa", "--iterations", "500000",
            "--policy-mode", $policy, "--policy-dir", $policyDir,
            "--seed", "$seed", "--run-label", $label, "--storage", "s3"
        )
    }
}

# CVRP Portfolio
foreach ($policy in $policies) {
    foreach ($seed in $seeds) {
        $label = "val-cvrp-a32k5-portfolio-${policy}-s${seed}"
        Invoke-OwpRun $label @(
            "solve-cvrp", "--instance", "../../examples/cvrp/A-n32-k5.vrp",
            "--mode", "portfolio", "--iterations", "500000",
            "--policy-mode", $policy, "--policy-dir", $policyDir,
            "--seed", "$seed", "--run-label", $label, "--storage", "s3"
        )
    }
}

# JSS Tabu
foreach ($policy in $policies) {
    foreach ($seed in $seeds) {
        $label = "val-jss-la01-tabu-${policy}-s${seed}"
        Invoke-OwpRun $label @(
            "solve-jobshop", "--instance", "internal/infrastructure/jobshop/testdata/la01.txt",
            "--mode", "tabu", "--iterations", "100000",
            "--policy-mode", $policy, "--policy-dir", $policyDir,
            "--seed", "$seed", "--run-label", $label, "--storage", "s3"
        )
    }
}

# JSS Portfolio
foreach ($policy in $policies) {
    foreach ($seed in $seeds) {
        $label = "val-jss-la01-portfolio-${policy}-s${seed}"
        Invoke-OwpRun $label @(
            "solve-jobshop", "--instance", "internal/infrastructure/jobshop/testdata/la01.txt",
            "--mode", "portfolio", "--iterations", "100000",
            "--policy-mode", $policy, "--policy-dir", $policyDir,
            "--seed", "$seed", "--run-label", $label, "--storage", "s3"
        )
    }
}

# VRPTW SA
foreach ($policy in $policies) {
    foreach ($seed in $seeds) {
        $label = "val-vrptw-c101-sa-${policy}-s${seed}"
        Invoke-OwpRun $label @(
            "solve-vrptw", "--instance", "../../examples/vrptw/C101.txt",
            "--mode", "sa", "--iterations", "100000",
            "--policy-mode", $policy, "--policy-dir", $policyDir,
            "--seed", "$seed", "--run-label", $label, "--storage", "s3"
        )
    }
}

# VRPTW Portfolio
foreach ($policy in $policies) {
    foreach ($seed in $seeds) {
        $label = "val-vrptw-c101-portfolio-${policy}-s${seed}"
        Invoke-OwpRun $label @(
            "solve-vrptw", "--instance", "../../examples/vrptw/C101.txt",
            "--mode", "portfolio", "--iterations", "100000",
            "--policy-mode", $policy, "--policy-dir", $policyDir,
            "--seed", "$seed", "--run-label", $label, "--storage", "s3"
        )
    }
}

# NRP SA
foreach ($policy in $policies) {
    foreach ($seed in $seeds) {
        $label = "val-nrp-n012w8-sa-${policy}-s${seed}"
        Invoke-OwpRun $label @(
            "tune-pfrs", "--instance", "n012w8",
            "--pfrs-mode", "sa", "--pfrs-iterations-per-worker", "100000",
            "--pfrs-max-total-workers", "16", "--seeds", "$seed",
            "--worker-decision-mode", "assist",
            "--policy-mode", $policy, "--policy-dir", $policyDir,
            "--pfrs-run-label", $label, "--storage", "s3"
        )
    }
}

# NRP Portfolio
foreach ($policy in $policies) {
    foreach ($seed in $seeds) {
        $label = "val-nrp-n012w8-portfolio-${policy}-s${seed}"
        Invoke-OwpRun $label @(
            "tune-pfrs", "--instance", "n012w8",
            "--pfrs-mode", "portfolio", "--pfrs-iterations-per-worker", "100000",
            "--pfrs-max-total-workers", "16", "--seeds", "$seed",
            "--worker-decision-mode", "assist",
            "--policy-mode", $policy, "--policy-dir", $policyDir,
            "--pfrs-run-label", $label, "--storage", "s3"
        )
    }
}

Write-Host ""
if ($failures.Count -eq 0) {
    Write-Host "Done. $totalRuns experiments complete." -ForegroundColor Green
    Write-Log "Done. $totalRuns experiments complete."
} else {
    Write-Host "Done with $($failures.Count) failure(s) out of $totalRuns." -ForegroundColor Yellow
    Write-Log "Failures: $($failures -join ', ')"
}
Write-Host "Run analysis: cd ..\ml; python validate_policies.py --data-dir ../web/pfrs-lab/data/runs --policy-dir policies --output ../../docs/reports/search-intelligence-v2-live-validation.md"
