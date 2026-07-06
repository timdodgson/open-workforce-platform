# Assist Mode Validation Experiment

## Setup

| Parameter | Value |
|-----------|-------|
| Instance | n012w8 |
| Algorithm | SA |
| Seeds | 42, 123, 555, 777, 999 |
| Modes | off, shadow, assist |
| Total Runs | 15 |

## How to Run

```powershell
cd platform/go
# Run all 15 experiments:
.\..\..\experiments\assist-validation.ps1
```

Or run individually:

```powershell
# Off (baseline)
go run ./cmd/owp tune-pfrs --instance n012w8 --pfrs-mode sa --seeds 42 --pfrs-run-label nrp-sa-off-s42 --pfrs-storage s3

# Shadow
go run ./cmd/owp tune-pfrs --instance n012w8 --pfrs-mode sa --seeds 42 --pfrs-run-label nrp-sa-shadow-s42 --pfrs-storage s3 --worker-decision-mode shadow

# Assist
go run ./cmd/owp tune-pfrs --instance n012w8 --pfrs-mode sa --seeds 42 --pfrs-run-label nrp-sa-assist-s42 --pfrs-storage s3 --worker-decision-mode assist
```

## Acceptance Criteria

- [ ] Assist must miss **zero** global bests
- [ ] Assist final objective must not be statistically worse than off
- [ ] Assist should save measurable CPU
- [ ] Shadow/assist dashboards must load correctly

## Results

*To be filled after running the experiment.*

### Raw Results

| Run Label | Mode | Seed | Final Obj | Workers | Skipped | CPU Saved | GB Missed | Time |
|-----------|------|------|-----------|---------|---------|-----------|-----------|------|
| nrp-sa-off-s42 | off | 42 | | | 0 | 0% | 0 | |
| nrp-sa-shadow-s42 | shadow | 42 | | | 0 | 0% | 0 | |
| nrp-sa-assist-s42 | assist | 42 | | | | | | |
| ... | ... | ... | ... | ... | ... | ... | ... | ... |

### Summary Statistics

| Mode | Mean Obj | Best Obj | Worst Obj | Std Dev | Avg CPU Saved | GB Missed |
|------|----------|----------|-----------|---------|---------------|-----------|
| off | | | | | 0% | 0 |
| shadow | | | | | 0% | 0 |
| assist | | | | | | |

### Conclusion

**Status:** PENDING

- [ ] SAFE — Assist mode produces equivalent objectives with measurable CPU savings and zero global best misses.
- [ ] NOT SAFE — Assist mode degrades solution quality or misses global bests.
- [ ] NEEDS MORE DATA — Insufficient evidence to conclude.

## Next Steps (if SAFE)

1. Run with Portfolio mode (SA + LAHC + Tabu)
2. Run on larger instances (n030w4, n060w4)
3. Run on CVRP domain
4. Consider enabling by default with safety guardrails
