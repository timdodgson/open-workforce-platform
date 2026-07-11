# ML Journey — Search Intelligence maturity ladder

Measured path from **~3–4/10** (offline decision trees) toward **10/10** (autonomous research loop). Each step has a **go/no-go gate** on benefit vs cost.

**Harness report:** `docs/reports/ml-harness/latest.json` (regenerate with `npm run compare-policy-modes`)  
**Private learning page:** `/admin/ml-journey` (login required)

---

## Maturity scale

| Level | Description |
|-------|-------------|
| 1–2 | Rules only |
| **3–4** | **Baseline (Step 0)** — offline trees, hybrid fallback, 288 val-* runs validated |
| **5** | Step 1–2 — measurement harness + gradient boosting (distilled to Go-runnable trees) |
| 6 | Per-context policies + counterfactual offline eval |
| 7 | Sequential decisions (bandits) |
| 8 | Trajectory / episode models |
| 9 | Deep policies where trees plateau |
| 10 | Closed-loop experiment + retrain automation |

---

## Scorecard (fill after each step)

| Step | Quality Δ | Compute Δ | Safety | Train cost | Eng weeks | ROI | Go? |
|------|-----------|-----------|--------|------------|-----------|-----|-----|
| 0 baseline | — | — | 0 violations | low | — | — | ✓ |
| 1 harness | | | | low | 0.5 | | |
| 2 boosting | | | | low | 1 | | |

**Stop rule:** Do not climb if next step needs **>2× cost** for **<2%** combined gain.

---

## Step 0 — Baseline (current)

- Policies: `stagnation`, `restart`, `budget`, `worker` JSON in `platform/ml/policies/`
- Runtime: Go walks exported sklearn trees (`SklearnTree`)
- Mode: `hybrid` = learned first, rules if low confidence
- Evidence: 288/288 val-* on S3, EXP-007 assist validation

---

## Step 1 — Measurement harness ✓

**Goal:** One command compares `rules` vs `hybrid` vs `learned` on the same val-* labels.

**Commands:**

```powershell
# Production bucket (from platform/web/pfrs-lab)
$env:STORAGE_PROVIDER='s3'
$env:PFRS_S3_BUCKET='pfrs-research-lab-data'
npm run compare-policy-modes

# Local runs directory
cd platform/go
go run ./cmd/owp validate-si2 compare --prefix val- --runs-dir ../web/pfrs-lab/data/runs

# Python (local data dir, writes markdown + JSON)
cd platform/ml
python compare_policy_outcomes.py --data-dir ../web/pfrs-lab/data/runs --prefix val-
```

**Metrics per domain × algorithm:**
- Mean objective (lower is better)
- Mean runtime ms
- Welch-style comparison vs rules
- Paired wins/losses/ties per seed
- **ROI score** = `(compute_saved_pct + quality_gain_pct) / complexity_cost`

**Gate:** Harness runs cleanly and produces `docs/reports/ml-harness/latest.json` before Step 2 promotion.

---

## Step 2 — Stronger classical ML ✓

**Goal:** Train with **gradient boosting** when it beats a plain tree on grouped CV; export a **distilled decision tree** for Go inference (no runtime Python).

**Changes:**
- `policy_training_utils.train_domain_classifier` — boosting on by default; distill winner to deployable tree
- Classifier metadata: `trainer`, `cv_tree`, `cv_boost`
- Portfolio / stagnation / restart / worker training use boosted path

**Retrain:**

```powershell
cd platform/ml
python train_policies.py --data-dir ../web/pfrs-lab/data/runs --output-dir policies
```

**Gate:** Re-run Step 1 harness. Promote only if hybrid/learned beats rules on **≥2 domains** for runtime or quality without safety regression.

---

## Steps 3–8 (planned, not implemented)

| Step | Target level | Summary |
|------|--------------|---------|
| 3 | 5.5 | Per domain×algorithm×instance policies; promotion per context |
| 4 | 6 | Counterfactual offline eval before deploy |
| 5 | 7 | Contextual bandits for portfolio / worker budgets |
| 6 | 8 | Sequence models on full search traces |
| 7 | 9 | Neural policies only where Step 6 plateaus |
| 8 | 10 | Human-in-the-loop autonomous experiment loop |

---

## Cost realism

| Step | Typical eng effort | Compute | Expected gain |
|------|-------------------|---------|---------------|
| 1 | days | negligible | visibility only |
| 2 | 1–2 weeks | retrain minutes | low–medium |
| 3–4 | weeks | medium | medium |
| 5+ | months+ | high | diminishing |

Step 9→10 often yields **<1% quality** for **50×** cost — acceptable only if SI is the core product.
