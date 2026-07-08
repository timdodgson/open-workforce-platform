# Retrain SI 2.0 policies after new runs land in S3.
# Syncs S3 → local, trains, validates, uploads registry.
# Run from platform/go after validate-si2-deep.ps1 (or any new val-* runs).

$ErrorActionPreference = "Stop"
$runsDir = "..\web\pfrs-lab\data\runs"
$mlDir = "..\ml"
$bucket = "pfrs-research-lab-data"

Write-Host "SI 2.0 Policy Retrain" -ForegroundColor Cyan

Write-Host "  Syncing runs from S3..." -ForegroundColor Gray
aws s3 sync "s3://$bucket/runs" $runsDir

Push-Location $mlDir
try {
    if (Test-Path "policies\validation_results.json") {
        Copy-Item "policies\validation_results.json" "policies\validation_results.before.json" -Force
        Write-Host "  Saved previous validation → policies/validation_results.before.json" -ForegroundColor Gray
    }

    Write-Host "  Training policies..." -ForegroundColor Gray
    python train_policies.py --data-dir ../web/pfrs-lab/data/runs --output-dir policies

    Write-Host "  Validating..." -ForegroundColor Gray
    python validate_policies.py --data-dir ../web/pfrs-lab/data/runs --policy-dir policies --output ../../docs/reports/search-intelligence-v2-retrain.md

    Write-Host "  Uploading registry to S3..." -ForegroundColor Gray
    aws s3 cp policies\policy_registry.json "s3://$bucket/policy_registry.json"
} finally {
    Pop-Location
}

Write-Host ""
Write-Host "Done." -ForegroundColor Green
Write-Host "  Compare: policies/validation_results.before.json vs policies/validation_results.json" -ForegroundColor Cyan
Write-Host "  Report:  docs/reports/search-intelligence-v2-retrain.md" -ForegroundColor Cyan
