# CVRP Validation Experiment Pack
# Instance: A-n10-k2 (10 customers, capacity 50)
# Modes: sa, lahc, tabu, portfolio
# Seeds: 42, 123, 555, 777, 999
# Iterations: 500000

$instance = "../../examples/inrc2/../../platform/go/internal/infrastructure/cvrp/testdata/A-n10-k2.vrp"
# Correct relative path from platform/go:
$instance = "internal/infrastructure/cvrp/testdata/A-n10-k2.vrp"
$iterations = 500000

$modes = @("sa", "lahc", "tabu", "portfolio")
$seeds = @(42, 123, 555, 777, 999)
$portfolioArgs = "sa,lahc,tabu"

Write-Host "=== CVRP Validation Experiment ===" -ForegroundColor Cyan
Write-Host "Instance: A-n10-k2"
Write-Host "Iterations: $iterations"
Write-Host "Modes: $($modes -join ', ')"
Write-Host "Seeds: $($seeds -join ', ')"
Write-Host "Total runs: $($modes.Count * $seeds.Count)"
Write-Host ""

foreach ($mode in $modes) {
    foreach ($seed in $seeds) {
        $label = "cvrp-$mode-s$seed"
        Write-Host "Running: $label" -ForegroundColor Green

        $args = @(
            "run", "./cmd/owp", "solve-cvrp",
            "--instance", $instance,
            "--mode", $mode,
            "--iterations", $iterations,
            "--seed", $seed,
            "--run-label", $label
        )

        if ($mode -eq "portfolio") {
            $args += "--portfolio"
            $args += $portfolioArgs
        }

        & go @args

        Write-Host ""
    }
}

Write-Host "=== All runs complete ===" -ForegroundColor Cyan
Write-Host "Check dashboard at http://localhost:3000 or deployed URL"
Write-Host "Runs should appear as: cvrp-sa-s42, cvrp-lahc-s42, etc."
