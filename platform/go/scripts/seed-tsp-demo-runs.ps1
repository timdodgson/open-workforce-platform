# Seed TSP BYOD demo runs into local lab storage (no S3 required).
# Execute from platform/go:
#   powershell -ExecutionPolicy Bypass -File .\scripts\seed-tsp-demo-runs.ps1

$instance = "..\..\examples\byod-tsp\instances\tsp-5city.json"
$runs = @(
    @{ Label = "demo-tsp-5city-sa-s42"; Mode = "sa"; Iters = 50000; Seed = 42 },
    @{ Label = "demo-tsp-5city-lahc-s42"; Mode = "lahc"; Iters = 50000; Seed = 42 },
    @{ Label = "demo-tsp-5city-greedy-s42"; Mode = "greedy"; Iters = 20000; Seed = 42 }
)

function Invoke-Owp {
    param([string]$Label, [string[]]$OwpArgs)
    Write-Host "Running $Label..." -ForegroundColor Gray
    go run ./cmd/owp @OwpArgs
    if ($LASTEXITCODE -ne 0) {
        Write-Host "FAILED $Label (exit $LASTEXITCODE)" -ForegroundColor Red
        exit $LASTEXITCODE
    }
}

Write-Host "Seeding TSP BYOD demo runs (local manifest)" -ForegroundColor Cyan

foreach ($r in $runs) {
    Invoke-Owp $r.Label @(
        "solve", "tsp",
        "--instance", $instance,
        "--mode", $r.Mode,
        "--iterations", "$($r.Iters)",
        "--seed", "$($r.Seed)",
        "--run-label", $r.Label
    )
}

Write-Host ""
Write-Host "Done. Runs under ../web/pfrs-lab/data/runs/ and manifest.json updated." -ForegroundColor Green
Write-Host "Open /runs or /lab/byod in the lab UI." -ForegroundColor Green
