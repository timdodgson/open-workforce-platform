# Search Intelligence 2.0 — Live Validation Script
# Runs all 240 experiments (8 configs × 3 policies × 10 seeds)
# Execute from platform/go directory

$seeds = @(42, 123, 555, 777, 999, 1001, 2022, 3033, 4044, 5055)
$policies = @("rules", "hybrid", "learned")
$policyDir = "../ml/policies"

Write-Host "SI 2.0 Live Validation - 240 experiments" -ForegroundColor Cyan
Write-Host ""

# CVRP SA
foreach ($policy in $policies) {
    foreach ($seed in $seeds) {
        $label = "val-cvrp-a32k5-sa-${policy}-s${seed}"
        Write-Host "  $label" -ForegroundColor Gray
        go run ./cmd/owp solve-cvrp --instance ../../examples/cvrp/A-n32-k5.vrp --mode sa --iterations 500000 --policy-mode $policy --policy-dir $policyDir --seed $seed --run-label $label --storage s3 2>&1 | Out-Null
    }
}

# CVRP Portfolio
foreach ($policy in $policies) {
    foreach ($seed in $seeds) {
        $label = "val-cvrp-a32k5-portfolio-${policy}-s${seed}"
        Write-Host "  $label" -ForegroundColor Gray
        go run ./cmd/owp solve-cvrp --instance ../../examples/cvrp/A-n32-k5.vrp --mode portfolio --iterations 500000 --policy-mode $policy --policy-dir $policyDir --seed $seed --run-label $label --storage s3 2>&1 | Out-Null
    }
}

# JSS Tabu
foreach ($policy in $policies) {
    foreach ($seed in $seeds) {
        $label = "val-jss-la01-tabu-${policy}-s${seed}"
        Write-Host "  $label" -ForegroundColor Gray
        go run ./cmd/owp solve-jobshop --instance internal/infrastructure/jobshop/testdata/la01.txt --mode tabu --iterations 100000 --policy-mode $policy --policy-dir $policyDir --seed $seed --run-label $label --storage s3 2>&1 | Out-Null
    }
}

# JSS Portfolio
foreach ($policy in $policies) {
    foreach ($seed in $seeds) {
        $label = "val-jss-la01-portfolio-${policy}-s${seed}"
        Write-Host "  $label" -ForegroundColor Gray
        go run ./cmd/owp solve-jobshop --instance internal/infrastructure/jobshop/testdata/la01.txt --mode portfolio --iterations 100000 --policy-mode $policy --policy-dir $policyDir --seed $seed --run-label $label --storage s3 2>&1 | Out-Null
    }
}

# VRPTW SA
foreach ($policy in $policies) {
    foreach ($seed in $seeds) {
        $label = "val-vrptw-c101-sa-${policy}-s${seed}"
        Write-Host "  $label" -ForegroundColor Gray
        go run ./cmd/owp solve-vrptw --instance ../../examples/vrptw/C101.txt --mode sa --iterations 100000 --policy-mode $policy --policy-dir $policyDir --seed $seed --run-label $label --storage s3 2>&1 | Out-Null
    }
}

# VRPTW Portfolio
foreach ($policy in $policies) {
    foreach ($seed in $seeds) {
        $label = "val-vrptw-c101-portfolio-${policy}-s${seed}"
        Write-Host "  $label" -ForegroundColor Gray
        go run ./cmd/owp solve-vrptw --instance ../../examples/vrptw/C101.txt --mode portfolio --iterations 100000 --policy-mode $policy --policy-dir $policyDir --seed $seed --run-label $label --storage s3 2>&1 | Out-Null
    }
}

# NRP SA
foreach ($policy in $policies) {
    foreach ($seed in $seeds) {
        $label = "val-nrp-n012w8-sa-${policy}-s${seed}"
        Write-Host "  $label" -ForegroundColor Gray
        go run ./cmd/owp tune-pfrs --instance n012w8 --pfrs-mode sa --pfrs-iterations-per-worker 100000 --pfrs-max-total-workers 16 --seeds $seed --worker-decision-mode assist --policy-mode $policy --policy-dir $policyDir --pfrs-run-label $label --storage s3 2>&1 | Out-Null
    }
}

# NRP Portfolio
foreach ($policy in $policies) {
    foreach ($seed in $seeds) {
        $label = "val-nrp-n012w8-portfolio-${policy}-s${seed}"
        Write-Host "  $label" -ForegroundColor Gray
        go run ./cmd/owp tune-pfrs --instance n012w8 --pfrs-mode portfolio --pfrs-iterations-per-worker 100000 --pfrs-max-total-workers 16 --seeds $seed --worker-decision-mode assist --policy-mode $policy --policy-dir $policyDir --pfrs-run-label $label --storage s3 2>&1 | Out-Null
    }
}

Write-Host ""
Write-Host "Done. 240 experiments complete." -ForegroundColor Green
Write-Host "Run analysis: cd ..\ml; python validate_policies.py --data-dir ../web/pfrs-lab/data/runs --policy-dir policies --output ../../docs/reports/search-intelligence-v2-live-validation.md"
