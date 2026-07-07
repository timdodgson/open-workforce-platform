# SI 2.0 quick validation — 6 runs (smoke test before full 240-run sweep)
# Execute from platform/go

$seeds = @(42)
$policies = @("rules", "hybrid", "learned")
$policyDir = "../ml/policies"

Write-Host "SI 2.0 Quick Validation - 6 experiments" -ForegroundColor Cyan

foreach ($policy in $policies) {
    $label = "val-quick-cvrp-sa-${policy}-s42"
    Write-Host "  $label" -ForegroundColor Gray
    go run ./cmd/owp solve-cvrp --instance ../../examples/cvrp/A-n32-k5.vrp --mode sa --iterations 50000 --policy-mode $policy --policy-dir $policyDir --seed 42 --run-label $label --pfrs-storage local 2>&1 | Out-Null
}

foreach ($policy in $policies) {
    $label = "val-quick-cvrp-portfolio-${policy}-s42"
    Write-Host "  $label" -ForegroundColor Gray
    go run ./cmd/owp solve-cvrp --instance ../../examples/cvrp/A-n32-k5.vrp --mode portfolio --iterations 50000 --policy-mode $policy --policy-dir $policyDir --seed 42 --run-label $label --pfrs-storage local 2>&1 | Out-Null
}

Write-Host ""
Write-Host "Done. Check data/runs/val-quick-* for policy_decisions.csv" -ForegroundColor Green
