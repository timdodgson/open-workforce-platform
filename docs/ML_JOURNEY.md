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

## Step 3 — Per-context policies ✓

**Goal:** Train separate classifiers per `domain × algorithm × instance` with fallback to broader contexts.

**Implementation:**
- `train_context_classifiers()` in `policy_training_utils.py`
- Go `findClassifier(domain, algorithm, instance)` prefers instance-specific trees
- Retrained policies: **15 stagnation classifiers** (14 instance-specific)

**Retrain after pull:**

```powershell
cd platform/ml
python train_policies.py --data-dir ../web/pfrs-lab/data/runs --output-dir policies
npm run compare-policy-modes  # from pfrs-lab with S3 env
```

**Gate:** Classifier count increases; per-context CV ≥ 0.5; harness unchanged or better on quality/runtime.

## Step 4 — Counterfactual offline eval ✓

**Goal:** Simulate learned decisions on historical checkpoints before promotion; block deploy when false-stop rate is too high.

**Implementation:**
- `counterfactual_eval.py` — ex-post false-stop rate + `counterfactual_learning.csv` telemetry
- Promotion gate: `false_stop_rate ≤ 5%` per domain (learned stopped when improvement remained)
- Harness reads `validation_results.json` for `step4PromoteOk`

**Run:**

```powershell
cd platform/web/pfrs-lab
npm run evaluate-counterfactual
npm run compare-policy-modes
```

**Gate:** `step4_promotion_ready: true` and harness `step4PromoteOk: true`.

## Step 5 — Contextual bandits ✓

**Goal:** Sequential portfolio/worker budget decisions via offline contextual bandits.

**Implementation:**
- `bandit_training.py` — per-context arm stats from `portfolio_assist.csv` / `worker_assist.csv`
- Exported in `budget_policy.json` / `worker_policy.json` as `bandit` blocks
- Go `bandit_policy.go` — instance/context lookup at solve time

**Retrain + eval:**

```powershell
cd platform/ml
python train_policies.py --data-dir ../web/pfrs-lab/data/runs --output-dir policies
cd ../web/pfrs-lab
npm run evaluate-counterfactual
npm run compare-policy-modes
```

**Gate:** `episode_regret ≤ 10%`; harness `step5PromoteOk: true`.

## Step 6 — Trajectory sequence models ✓

**Goal:** Train stagnation policies on full per-run checkpoint traces, not isolated rows.

**Implementation:**
- `trajectory_training.py` — sequence features (`trace_progress`, `recent_slope`, `plateau_streak_ratio`, …)
- Exported in `stagnation_policy.json` as `trajectory` block with distilled trees
- Go stagnation assessor prefers trajectory classifiers when `promotion_ready`

**Gate:** `gain_vs_checkpoint ≥ 0` (non-regression vs checkpoint-only) and `episode_accuracy ≥ 0.55`; harness `step6PromoteOk: true`.

## Step 7 — Neural (plateau contexts) ✓

**Goal:** Apply small MLPs only where Step 6 trajectory gain plateaued; distill to Go trees.

**Implementation:**
- `neural_training.py` — skips if global trajectory gain ≥ 1%; else MLP per context
- Promotes only when `gain_vs_trajectory ≥ 0.3%` vs trajectory classifier
- Exported in `stagnation_policy.json` as `neural` block; Go prefers neural on promoted contexts

**Gate:** `step7_promotion_ready: true` (at least one context promoted); harness `step7PromoteOk`.

## Step 8 — Closed-loop research ✓

**Goal:** Auto-propose ML experiments from validation/registry/harness gaps; human approves before execution.

**Implementation:**
- `research_loop.py` — reads `validation_results.json`, `policy_registry.json`, harness comparisons, neural promotions → `research_queue.json`
- npm `propose-ml-experiments` / `run-ml-experiments` (dry-run default; `--approve-ids` to execute `owp` commands from `platform/go`)
- Harness `step8LoopOk` + `step8PromoteOk`; maturity **10/10** when both pass

**Run:**

```powershell
cd platform/web/pfrs-lab
npm run propose-ml-experiments
npm run run-ml-experiments
npm run compare-policy-modes

# Execute one approved proposal (requires S3 env + Go toolchain)
npm run run-ml-experiments -- --approve-ids=<proposal-id>
```

**Gate:** `step8_loop_ok: true` (queue with `requires_approval` on every proposal); `step8_promotion_ready: true` (≥2 signal sources: registry, harness, neural, retrain).

---

## Cost realism

| Step | Typical eng effort | Compute | Expected gain |
|------|-------------------|---------|---------------|
| 1 | days | negligible | visibility only |
| 2 | 1–2 weeks | retrain minutes | low–medium |
| 3–4 | weeks | medium | medium |
| 5+ | months+ | high | diminishing |

Step 9→10 often yields **<1% quality** for **50×** cost — acceptable only if SI is the core product.
