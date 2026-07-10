# Post-refactor regression gate — run from platform/go:
#   powershell -ExecutionPolicy Bypass -File .\scripts\regression-post-refactor.ps1
#   powershell -ExecutionPolicy Bypass -File .\scripts\regression-post-refactor.ps1 -SkipGoTest
#
# Phase 1: Full Go test suite (no -short; includes CLI smoke + CVRP integration).
# Phase 2: Live CLI matrix — 24 runs mirroring validate-si2.ps1 (8 configs × 3 policies × 1 seed).
# Phase 3: Artifact checks on each run directory.

param([switch]$SkipGoTest)

$ErrorActionPreference = "Stop"
Set-Location $PSScriptRoot\..

$logDir = Join-Path $PSScriptRoot "regression-logs"
New-Item -ItemType Directory -Force -Path $logDir | Out-Null
$summaryPath = Join-Path $logDir ("regression-" + (Get-Date -Format "yyyyMMdd-HHmmss") + ".txt")
$failures = [System.Collections.Generic.List[string]]::new()

function Log($msg) {
    $line = "[{0}] {1}" -f (Get-Date -Format "HH:mm:ss"), $msg
    Write-Host $line
    Add-Content -Path $summaryPath -Value $line
}

function Fail($phase, $detail) {
    $msg = "FAIL [$phase] $detail"
    $failures.Add($msg)
    Log $msg
}

function Pass($phase, $detail) {
    Log "PASS [$phase] $detail"
}

Log "=== Post-refactor regression ==="
Log "Log file: $summaryPath"

# --- Phase 1: Go tests ---
if (-not $SkipGoTest) {
    Log "Phase 1: go test ./... -count=1 (full, no -short)"
    $testLog = Join-Path $logDir "go-test.log"
    go test ./... -count=1 -timeout 45m 2>&1 | Tee-Object -FilePath $testLog
    if ($LASTEXITCODE -ne 0) {
        Fail "go-test" "exit code $LASTEXITCODE (see $testLog)"
    } else {
        Pass "go-test" "all packages passed"
    }
} else {
    Log "Phase 1: skipped (-SkipGoTest)"
}

# Build binary once for live runs
Log "Building owp binary..."
$binDir = Join-Path $logDir "bin"
New-Item -ItemType Directory -Force -Path $binDir | Out-Null
$owp = Join-Path $binDir "owp"
if ($IsWindows -or $env:OS -match "Windows") { $owp += ".exe" }
go build -o $owp ./cmd/owp
if ($LASTEXITCODE -ne 0) {
    Fail "build" "go build ./cmd/owp failed"
    Log "=== REGRESSION FAILED ($($failures.Count) failures) ==="
    exit 1
}

$seeds = @(42)
$policies = @("rules", "hybrid", "learned")
$policyDir = "../ml/policies"
$runsRoot = Resolve-Path "..\web\pfrs-lab\data\runs"

# --- Phase 2: Live CLI matrix (24 runs) ---
Log "Phase 2: Live CLI matrix (24 runs, 50k iterations, local storage)"

$liveRuns = @(
    @{ Kind = "cvrp-sa";       Args = { param($p,$s) @("solve-cvrp","--instance","../../examples/cvrp/A-n32-k5.vrp","--mode","sa","--iterations","50000","--policy-mode",$p,"--policy-dir",$policyDir,"--seed",$s) } }
    @{ Kind = "cvrp-portfolio"; Args = { param($p,$s) @("solve-cvrp","--instance","../../examples/cvrp/A-n32-k5.vrp","--mode","portfolio","--iterations","50000","--policy-mode",$p,"--policy-dir",$policyDir,"--seed",$s) } }
    @{ Kind = "jss-tabu";       Args = { param($p,$s) @("solve-jobshop","--instance","internal/infrastructure/jobshop/testdata/la01.txt","--mode","tabu","--iterations","50000","--policy-mode",$p,"--policy-dir",$policyDir,"--seed",$s) } }
    @{ Kind = "jss-portfolio";  Args = { param($p,$s) @("solve-jobshop","--instance","internal/infrastructure/jobshop/testdata/la01.txt","--mode","portfolio","--iterations","50000","--policy-mode",$p,"--policy-dir",$policyDir,"--seed",$s) } }
    @{ Kind = "vrptw-sa";       Args = { param($p,$s) @("solve-vrptw","--instance","../../examples/vrptw/C101.txt","--mode","sa","--iterations","50000","--policy-mode",$p,"--policy-dir",$policyDir,"--seed",$s) } }
    @{ Kind = "vrptw-portfolio"; Args = { param($p,$s) @("solve-vrptw","--instance","../../examples/vrptw/C101.txt","--mode","portfolio","--iterations","50000","--policy-mode",$p,"--policy-dir",$policyDir,"--seed",$s) } }
    @{ Kind = "nrp-sa";         Args = { param($p,$s) @("tune-pfrs","--instance","n005w4","--pfrs-mode","sa","--pfrs-iterations-per-worker","10000","--pfrs-max-total-workers","4","--seeds",$s,"--worker-decision-mode","assist","--policy-mode",$p,"--policy-dir",$policyDir) } }
    @{ Kind = "nrp-portfolio";  Args = { param($p,$s) @("tune-pfrs","--instance","n005w4","--pfrs-mode","portfolio","--pfrs-iterations-per-worker","10000","--pfrs-max-total-workers","4","--seeds",$s,"--worker-decision-mode","assist","--policy-mode",$p,"--policy-dir",$policyDir) } }
)

$runCount = 0
foreach ($cfg in $liveRuns) {
    foreach ($policy in $policies) {
        foreach ($seed in $seeds) {
            $runCount++
            $label = "regress-$($cfg.Kind)-${policy}-s${seed}"
            $runLog = Join-Path $logDir "$label.log"
            Log "  [$runCount/24] $label"

            $baseArgs = & $cfg.Args $policy $seed
            if ($cfg.Kind -like "nrp-*") {
                $allArgs = $baseArgs + @("--pfrs-run-label", $label)
            } else {
                $allArgs = $baseArgs + @("--run-label", $label)
            }

            $prevEAP = $ErrorActionPreference
            $ErrorActionPreference = "Continue"
            & $owp @allArgs 2>&1 | Tee-Object -FilePath $runLog
            $exitCode = $LASTEXITCODE
            $ErrorActionPreference = $prevEAP
            if ($exitCode -ne 0) {
                Fail "live" "$label exit $exitCode (see $runLog)"
                continue
            }

            $runDir = Join-Path $runsRoot $label
            if (-not (Test-Path (Join-Path $runDir "run.json"))) {
                Fail "artifact" "$label missing run.json"
                continue
            }
            $policyCSV = Join-Path $runDir "policy_decisions.csv"
            if (-not (Test-Path $policyCSV)) {
                Fail "artifact" "$label missing policy_decisions.csv"
                continue
            }
            Pass "live" $label
        }
    }
}

# --- Phase 3: INRC2 validator (no run dir) ---
Log "Phase 3: validate-inrc2"
$inrc2Base = Resolve-Path "..\..\examples\inrc2\testdatasets_json\n005w4"
$valLog = Join-Path $logDir "validate-inrc2.log"
$prevEAP = $ErrorActionPreference
$ErrorActionPreference = "Continue"
& $owp @(
    "validate-inrc2",
    (Join-Path $inrc2Base "Sc-n005w4.json"),
    (Join-Path $inrc2Base "WD-n005w4-1.json"),
    (Join-Path $inrc2Base "H0-n005w4-0.json"),
    (Join-Path $inrc2Base "Solution_H_0-WD_1-2-3-3\Sol-n005w4-1-0.json")
) 2>&1 | Tee-Object -FilePath $valLog
$valExit = $LASTEXITCODE
$ErrorActionPreference = $prevEAP
if ($valExit -ne 0) {
    Fail "validate-inrc2" "exit $valExit"
} else {
    Pass "validate-inrc2" "ok"
}

# --- Summary ---
Log ""
if ($failures.Count -eq 0) {
    Log "=== REGRESSION PASSED (go tests + $runCount live runs + INRC2) ==="
    exit 0
}

Log "=== REGRESSION FAILED ($($failures.Count) failures) ==="
foreach ($f in $failures) { Log "  $f" }
exit 1
