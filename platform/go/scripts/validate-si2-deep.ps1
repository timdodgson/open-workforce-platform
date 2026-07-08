# SI 2.0 Deep Validation — 48 runs (8 configs × 3 policies × 2 seeds)
# Larger instances + longer budgets → richer checkpoint telemetry for policy training.
# Complements validate-si2.ps1 (240 fast runs for benchmarking).
# Execute from platform/go:
#   powershell -ExecutionPolicy Bypass -File .\scripts\validate-si2-deep.ps1
#
# Expected wall time: ~4–12 hours depending on hardware (minutes per run, not seconds).

$seeds = @(42, 123)
$policies = @("rules", "hybrid", "learned")
$policyDir = "../ml/policies"

# Instance / budget choices (largest available per domain in this repo).
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

Write-Host "SI 2.0 Deep Validation - 48 experiments" -ForegroundColor Cyan
Write-Host "  CVRP: A-n80-k10  SA=${cvrpSaIters}  Portfolio=${cvrpPortfolioIters}" -ForegroundColor DarkGray
Write-Host "  JSS:  ft10       Tabu=${jssTabuIters}  Portfolio=${jssPortfolioIters}" -ForegroundColor DarkGray
Write-Host "  VRPTW: C101      SA=${vrptwSaIters}  Portfolio=${vrptwPortfolioIters}" -ForegroundColor DarkGray
Write-Host "  NRP:  n012w8     ${nrpItersPerWorker}/worker x${nrpMaxWorkers} max" -ForegroundColor DarkGray
Write-Host ""

# CVRP SA
foreach ($policy in $policies) {
    foreach ($seed in $seeds) {
        $label = "val-deep-cvrp-a80k10-sa-${policy}-s${seed}"
        Write-Host "  $label" -ForegroundColor Gray
        go run ./cmd/owp solve-cvrp --instance $cvrpInstance --mode sa --iterations $cvrpSaIters --policy-mode $policy --policy-dir $policyDir --seed $seed --run-label $label --storage s3 2>&1 | Out-Null
    }
}

# CVRP Portfolio
foreach ($policy in $policies) {
    foreach ($seed in $seeds) {
        $label = "val-deep-cvrp-a80k10-portfolio-${policy}-s${seed}"
        Write-Host "  $label" -ForegroundColor Gray
        go run ./cmd/owp solve-cvrp --instance $cvrpInstance --mode portfolio --iterations $cvrpPortfolioIters --policy-mode $policy --policy-dir $policyDir --seed $seed --run-label $label --storage s3 2>&1 | Out-Null
    }
}

# JSS Tabu
foreach ($policy in $policies) {
    foreach ($seed in $seeds) {
        $label = "val-deep-jss-ft10-tabu-${policy}-s${seed}"
        Write-Host "  $label" -ForegroundColor Gray
        go run ./cmd/owp solve-jobshop --instance $jssInstance --mode tabu --iterations $jssTabuIters --policy-mode $policy --policy-dir $policyDir --seed $seed --run-label $label --storage s3 2>&1 | Out-Null
    }
}

# JSS Portfolio
foreach ($policy in $policies) {
    foreach ($seed in $seeds) {
        $label = "val-deep-jss-ft10-portfolio-${policy}-s${seed}"
        Write-Host "  $label" -ForegroundColor Gray
        go run ./cmd/owp solve-jobshop --instance $jssInstance --mode portfolio --iterations $jssPortfolioIters --policy-mode $policy --policy-dir $policyDir --seed $seed --run-label $label --storage s3 2>&1 | Out-Null
    }
}

# VRPTW SA
foreach ($policy in $policies) {
    foreach ($seed in $seeds) {
        $label = "val-deep-vrptw-c101-sa-${policy}-s${seed}"
        Write-Host "  $label" -ForegroundColor Gray
        go run ./cmd/owp solve-vrptw --instance $vrptwInstance --mode sa --iterations $vrptwSaIters --policy-mode $policy --policy-dir $policyDir --seed $seed --run-label $label --storage s3 2>&1 | Out-Null
    }
}

# VRPTW Portfolio
foreach ($policy in $policies) {
    foreach ($seed in $seeds) {
        $label = "val-deep-vrptw-c101-portfolio-${policy}-s${seed}"
        Write-Host "  $label" -ForegroundColor Gray
        go run ./cmd/owp solve-vrptw --instance $vrptwInstance --mode portfolio --iterations $vrptwPortfolioIters --policy-mode $policy --policy-dir $policyDir --seed $seed --run-label $label --storage s3 2>&1 | Out-Null
    }
}

# NRP SA
foreach ($policy in $policies) {
    foreach ($seed in $seeds) {
        $label = "val-deep-nrp-n012w8-sa-${policy}-s${seed}"
        Write-Host "  $label" -ForegroundColor Gray
        go run ./cmd/owp tune-pfrs --instance $nrpInstance --pfrs-mode sa --pfrs-iterations-per-worker $nrpItersPerWorker --pfrs-max-total-workers $nrpMaxWorkers --seeds $seed --worker-decision-mode assist --policy-mode $policy --policy-dir $policyDir --pfrs-run-label $label --pfrs-storage s3 2>&1 | Out-Null
    }
}

# NRP Portfolio
foreach ($policy in $policies) {
    foreach ($seed in $seeds) {
        $label = "val-deep-nrp-n012w8-portfolio-${policy}-s${seed}"
        Write-Host "  $label" -ForegroundColor Gray
        go run ./cmd/owp tune-pfrs --instance $nrpInstance --pfrs-mode portfolio --pfrs-iterations-per-worker $nrpItersPerWorker --pfrs-max-total-workers $nrpMaxWorkers --seeds $seed --worker-decision-mode assist --policy-mode $policy --policy-dir $policyDir --pfrs-run-label $label --pfrs-storage s3 2>&1 | Out-Null
    }
}

Write-Host ""
Write-Host "Done. 48 deep experiments complete." -ForegroundColor Green
Write-Host "Next: powershell -ExecutionPolicy Bypass -File .\scripts\retrain-si2-policies.ps1" -ForegroundColor Cyan
