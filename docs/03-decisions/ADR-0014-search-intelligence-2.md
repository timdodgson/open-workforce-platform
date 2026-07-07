# ADR-0014: Search Intelligence 2.0

## Status

Proposed

## Context

Search Intelligence v1 (ADR-0013) validated that AI-guided compute allocation works. 320 runs across 4 domains confirmed 40–73% compute savings with zero quality loss and 19% quality improvement on VRPTW.

However, v1 relies heavily on fixed rule-based heuristics with a learned model only for portfolio budget allocation. The rules encode assumptions that are sometimes wrong (e.g. "SA is generally strong" causing JSS degradation). The system cannot improve its own decision-making over time without manual model retraining.

### Problems with v1

1. **Fixed rules cannot adapt.** Rule-based heuristics encode domain-general assumptions that are demonstrably wrong on specific domains.
2. **No counterfactual learning.** The system records what happened but not what would have happened under alternative decisions.
3. **No policy versioning.** There is no way to compare policy performance over time or roll back a degraded policy.
4. **Limited context.** Rules use simple feature thresholds. They cannot capture complex interactions between search state features.
5. **No online evaluation.** Decisions are not evaluated against expected value, making it impossible to detect policy drift.

### Current Architecture (v1)

```
Telemetry
    ↓
Rules (fixed heuristics)
    ↓
Decision
```

### Target Architecture (v2)

```
Telemetry
    ↓
Feature Extraction
    ↓
Policy Model (learned or rule-based)
    ↓
Decision
    ↓
Outcome Recording
    ↓
Counterfactual Learning
```

## Decision

Replace the fixed rule-based decision layer with a learned policy framework that:
- Supports multiple policy types (rule, learned, hybrid)
- Learns separately per domain
- Records counterfactual decisions for offline learning
- Evaluates policy performance online
- Falls back to rules when confidence is insufficient
- Maintains hard safety constraints as non-negotiable overrides

---

## Architecture

### Core Abstractions

#### Policy

```go
type Policy interface {
    // Recommend returns a decision for the given context.
    Recommend(ctx DecisionContext) Decision

    // Metadata returns version, type, and domain information.
    Metadata() PolicyMetadata
}
```

#### DecisionContext

```go
type DecisionContext struct {
    // Domain context
    Domain          string  // nrp, cvrp, jss, vrptw
    Instance        string  // e.g. A-n32-k5, la01, C101, n012w8
    Algorithm       string  // sa, lahc, tabu, portfolio

    // Search state features
    Features        FeatureVector

    // Historical context
    PreviousRuns    int     // how many runs exist for this config
    DomainExperience int    // total runs in this domain
}
```

#### FeatureVector

```go
type FeatureVector struct {
    // Progress features
    BudgetConsumed      float64 // fraction of iteration budget used
    ImprovementRate     float64 // improvements per 1000 candidates
    TimeSinceLastImprove int    // candidates since last global best
    StagnationRatio     float64 // plateau length / budget

    // Quality features
    CurrentObjective    int
    BestObjective       int
    DistanceFromBest    int
    GapToReference      float64 // gap to best-known if available

    // Algorithm-specific features
    Temperature         float64 // SA only
    AcceptanceRate      float64
    TabuHitRate         float64 // Tabu only
    LAHCLateAccepts     float64 // LAHC only

    // Beam search features (NRP only)
    BeamHealth          float64
    LineageEntropy      float64
    DiversitySlotUsage  float64
    WorkerDepth         int
    ParentImproved      bool

    // Portfolio features
    StrategyWinRate     float64 // historical win rate for this strategy
    StrategyBudgetShare float64 // current allocation fraction
}
```

#### Decision

```go
type Decision struct {
    Action          string          // run, skip, early_stop, extend, allocate, etc.
    Confidence      float64         // 0.0–1.0
    ExpectedValue   float64         // estimated improvement or savings
    PolicyVersion   string          // which policy version produced this
    Reason          string          // human-readable explanation
    Parameters      map[string]any  // action-specific params (budget, algorithm, etc.)
}
```

#### PolicyMetadata

```go
type PolicyMetadata struct {
    Name            string
    Version         string    // semver or timestamp
    Type            string    // "rule", "learned", "hybrid"
    Domain          string    // domain this policy is trained for ("*" for universal)
    TrainedOn       int       // number of training samples
    LastUpdated     time.Time
    ValidationScore float64   // offline validation accuracy
}
```

---

### Policy Types

#### 1. RulePolicy

The existing v1 rule-based system, preserved as a fallback.

```go
type RulePolicy struct {
    rules []Rule
}

func (p *RulePolicy) Recommend(ctx DecisionContext) Decision {
    for _, rule := range p.rules {
        if rule.Matches(ctx) {
            return rule.Decision(ctx)
        }
    }
    return Decision{Action: "continue", Confidence: 0.5}
}
```

**Use case:** Default fallback. Always available. Zero external dependencies.

#### 2. LearnedPolicy

A trained model that maps feature vectors to decisions.

```go
type LearnedPolicy struct {
    model       PolicyModel  // decision tree, gradient boost, etc.
    metadata    PolicyMetadata
    threshold   float64      // minimum confidence to apply
}

func (p *LearnedPolicy) Recommend(ctx DecisionContext) Decision {
    prediction := p.model.Predict(ctx.Features)
    if prediction.Confidence < p.threshold {
        return Decision{Action: "defer", Confidence: prediction.Confidence}
    }
    return Decision{
        Action:        prediction.Action,
        Confidence:    prediction.Confidence,
        ExpectedValue: prediction.ExpectedValue,
        PolicyVersion: p.metadata.Version,
    }
}
```

**Use case:** Per-domain trained models. Higher accuracy than rules on domains with sufficient data.

#### 3. HybridPolicy

Uses learned policy when confident, falls back to rules otherwise.

```go
type HybridPolicy struct {
    learned     *LearnedPolicy
    fallback    *RulePolicy
    threshold   float64
}

func (p *HybridPolicy) Recommend(ctx DecisionContext) Decision {
    decision := p.learned.Recommend(ctx)
    if decision.Action == "defer" || decision.Confidence < p.threshold {
        fallback := p.fallback.Recommend(ctx)
        fallback.Reason = "learned_low_confidence: " + fallback.Reason
        return fallback
    }
    return decision
}
```

**Use case:** Production default. Graceful degradation. Never worse than rules.

---

### Policy Provider

```go
type PolicyProvider struct {
    policies map[string]map[string]Policy // domain → decision_type → policy
    registry PolicyRegistry
}

func (pp *PolicyProvider) GetPolicy(domain string, decisionType string) Policy {
    if domainPolicies, ok := pp.policies[domain]; ok {
        if policy, ok := domainPolicies[decisionType]; ok {
            return policy
        }
    }
    // Fall back to universal rule policy.
    return pp.policies["*"][decisionType]
}
```

Policies are loaded from:
1. `--policy-dir <path>` CLI flag (directory of policy JSON files)
2. Embedded rule policies (compiled in, always available)
3. S3 policy store (optional, for cloud deployment)

---

### Candidate Policies

| Decision Type | Current (v1) | Target (v2) |
|---------------|-------------|-------------|
| Budget Allocation | Rule + learned model | HybridPolicy per domain |
| Early Stopping | Rule (stagnation threshold) | LearnedPolicy (progress curve model) |
| Restart Decision | Rule (budget %) | LearnedPolicy (improvement potential) |
| Stagnation Detection | Fixed threshold | LearnedPolicy (adaptive threshold per domain) |
| Algorithm Selection | Rule (SA default) | LearnedPolicy (domain-instance-aware) |
| Worker Spawning | Rule (depth + entropy) | LearnedPolicy (lineage value prediction) |
| Portfolio Allocation | Learned model + rule fallback | HybridPolicy with counterfactual evaluation |

---

### Context Awareness

Policies MUST learn separately per domain. Cross-domain transfer is not assumed.

```
policies/
├── nrp/
│   ├── worker_spawn.json
│   ├── budget_allocation.json
│   └── early_stop.json
├── cvrp/
│   ├── budget_allocation.json
│   ├── early_stop.json
│   └── algorithm_selection.json
├── jss/
│   ├── budget_allocation.json
│   └── early_stop.json
├── vrptw/
│   ├── budget_allocation.json
│   ├── early_stop.json
│   └── algorithm_selection.json
└── universal/
    └── rules.json  (fallback)
```

Each domain policy is trained on telemetry from that domain only. A universal rule policy provides the baseline fallback for all domains.

Instance-specific overrides may exist within domain policies (e.g. CVRP A-n32-k5 behaves differently from A-n80-k10).

---

### Counterfactual Learning

Every decision records both the actual action taken and the counterfactual alternative.

```go
type CounterfactualRecord struct {
    // Context
    Timestamp       time.Time
    RunID           string
    Domain          string
    Instance        string
    DecisionType    string

    // Actual decision
    ActualAction    string
    ActualConfidence float64
    PolicyVersion   string

    // Counterfactual
    CounterfactualAction string  // what would rule/alternative policy have done
    CounterfactualSource string  // "rule", "previous_version", "random"

    // Outcome (filled after run completes)
    ObservedOutcome     float64 // actual result (objective, runtime, etc.)
    EstimatedRegret     float64 // estimated loss vs counterfactual
    OutcomeMetric       string  // "objective", "runtime", "compute_saved"

    // Attribution
    ImprovementAmount   int
    ComputeSaved        int
    QualityDelta        float64
}
```

**How regret is estimated:**

For decisions where the counterfactual can be observed (e.g. shadow mode records what would have happened):
- `EstimatedRegret = CounterfactualOutcome - ActualOutcome`

For decisions where the counterfactual cannot be directly observed:
- Use historical data from similar contexts
- Use the policy's own expected value prediction vs actual outcome
- Record `EstimatedRegret = ExpectedValue - ActualOutcome`

**Training data generation:**

Counterfactual records are the primary training signal for policy improvement:
1. Shadow mode generates counterfactual data without risk
2. Assist mode generates counterfactual data with safety overrides
3. Adaptive mode generates counterfactual data with full policy control
4. Offline evaluation uses historical data to simulate alternative policies

---

### Online Policy Evaluation

Every decision includes metadata for real-time policy quality assessment.

```go
type PolicyEvaluation struct {
    DecisionID      string
    PolicyVersion   string
    Confidence      float64
    ExpectedValue   float64
    ActualValue     float64  // filled post-outcome
    Regret          float64  // filled post-outcome
    FallbackReason  string   // empty if learned policy was used
    SafetyOverride  bool     // true if safety rule overrode policy
}
```

**Policy quality metrics (computed per version, per domain):**

| Metric | Definition |
|--------|-----------|
| Accuracy | fraction of decisions where actual outcome matched prediction |
| Regret | cumulative regret vs rule-based baseline |
| Confidence Calibration | does 80% confidence mean 80% correct? |
| Safety Override Rate | how often safety rules override the policy |
| Fallback Rate | how often learned policy defers to rules |
| Improvement vs Rules | mean improvement over rule-based decisions |

---

### Safety

Safety constraints are non-negotiable. They override all policies regardless of confidence.

```go
type SafetyGate struct {
    constraints []SafetyConstraint
}

func (sg *SafetyGate) Apply(ctx DecisionContext, decision Decision) Decision {
    for _, constraint := range sg.constraints {
        if constraint.Violated(ctx, decision) {
            return Decision{
                Action:     constraint.SafeAction(),
                Confidence: 1.0,
                Reason:     "safety_override: " + constraint.Name(),
            }
        }
    }
    return decision
}
```

**Hard safety constraints (unchanged from v1):**

- Never skip global-best lineage workers
- Never early-stop before 20% budget consumed
- Never early-stop immediately after improvement
- Never allocate zero budget to all strategies
- Never remove all budget from historically strongest strategy
- Cap budget boost at 2× original
- Floor budget at 0.25× original
- Require minimum 2 strategies in portfolio mode

**Additional v2 safety constraints:**

- Policy confidence below threshold (0.60) → fallback to rules
- Policy version older than 30 days → log warning, consider refresh
- Policy trained on fewer than 50 samples → fallback to rules
- Regret exceeding 2× standard deviation → trigger policy review
- Three consecutive worse-than-rule outcomes → temporary fallback

---

### Policy Lifecycle

```
┌─────────────────────────────────────────────────────────────┐
│                    Policy Lifecycle                           │
│                                                              │
│  1. COLLECT     Telemetry → Feature Extraction → Training   │
│                 Data                                          │
│                                                              │
│  2. TRAIN       Training Data → Model → Offline Validation  │
│                                                              │
│  3. SHADOW      Deploy in shadow mode → Record predictions  │
│                 + outcomes → Measure accuracy                 │
│                                                              │
│  4. PROMOTE     Shadow accuracy above threshold →            │
│                 Promote to assist/adaptive                    │
│                                                              │
│  5. MONITOR     Track regret, confidence calibration,        │
│                 safety override rate                          │
│                                                              │
│  6. RETIRE      Performance degrades → Rollback to previous │
│                 version or rules                             │
└─────────────────────────────────────────────────────────────┘
```

---

### Policy Management

#### Policy Registry

```go
type PolicyRegistry struct {
    versions []PolicyVersion
}

type PolicyVersion struct {
    Version         string
    Domain          string
    DecisionType    string
    CreatedAt       time.Time
    TrainedSamples  int
    OfflineAccuracy float64
    ShadowAccuracy  float64  // -1 if not yet shadow-tested
    ProductionRuns  int
    Status          string   // "training", "shadow", "active", "retired"
    FilePath        string
}
```

#### CLI Integration

```bash
# Use specific policy directory
owp solve-cvrp --policy-dir ./policies --worker-decision-mode adaptive

# Use shadow mode to evaluate new policy without risk
owp solve-cvrp --policy-dir ./policies-v2 --worker-decision-mode shadow

# Compare two policy versions
owp policy-compare --v1 ./policies --v2 ./policies-v2 --domain cvrp
```

---

### Dashboard: Policy Evolution

New dashboard page: `/intelligence/policies`

**Sections:**

1. **Policy Overview** — Active policies per domain, version, status
2. **Performance Timeline** — Regret over time, accuracy over time, per policy version
3. **Confidence Calibration** — Reliability diagram (predicted vs actual)
4. **Rollout History** — When each version was promoted/retired and why
5. **Counterfactual Analysis** — Estimated regret vs rule baseline
6. **Policy Comparison** — Side-by-side performance of two versions
7. **Safety Dashboard** — Override rates, fallback rates, constraint triggers

**Key metrics displayed:**

| Metric | Visualisation |
|--------|--------------|
| Policy accuracy over time | Line chart |
| Cumulative regret vs rules | Area chart |
| Confidence calibration | Reliability diagram |
| Decision breakdown | Stacked bar (learned/rule/safety) |
| Domain performance | Table with per-domain stats |
| Version comparison | Paired bar charts |

---

### Telemetry Output

New CSV files per run:

**`policy_decisions.csv`**
```
timestamp,run_id,domain,instance,decision_type,action,confidence,
expected_value,policy_version,policy_type,fallback_reason,safety_override,
counterfactual_action,counterfactual_source
```

**`policy_outcomes.csv`**
```
timestamp,run_id,domain,instance,decision_type,policy_version,
actual_value,expected_value,regret,improvement_amount,compute_saved
```

**`policy_registry.json`**
```json
{
  "policies": [
    {
      "version": "cvrp-budget-v3",
      "domain": "cvrp",
      "decisionType": "budget_allocation",
      "status": "active",
      "trainedSamples": 240,
      "offlineAccuracy": 0.82,
      "shadowAccuracy": 0.79,
      "productionRuns": 45,
      "createdAt": "2026-07-01T00:00:00Z"
    }
  ]
}
```

---

## Implementation Plan

### Phase 1: Framework (no behaviour change)

1. Define `Policy`, `DecisionContext`, `Decision`, `PolicyMetadata` interfaces
2. Wrap existing rule engines in `RulePolicy` implementations
3. Introduce `PolicyProvider` with domain-aware routing
4. Add `CounterfactualRecord` to all decision outputs
5. Add `PolicyEvaluation` fields to existing CSV outputs
6. No behaviour change — all decisions still use rules

### Phase 2: Learned Policies

1. Train domain-specific models from historical telemetry
2. Implement `LearnedPolicy` with confidence thresholds
3. Implement `HybridPolicy` combining learned + rules
4. Deploy in shadow mode for validation
5. Compare learned vs rule accuracy on historical data

### Phase 3: Online Evaluation

1. Add policy quality metrics computation
2. Add automatic fallback on poor performance
3. Add policy version tracking and comparison
4. Build dashboard visualisations

### Phase 4: Continuous Learning

1. Implement offline policy retraining pipeline
2. Add counterfactual regret estimation
3. Add policy promotion/retirement automation
4. Add cross-version comparison tooling

---

## Consequences

### Positive

- Policies improve over time with more data
- Domain-specific learning eliminates false priors
- Counterfactual records provide rich training signal
- Policy versioning enables safe experimentation
- Hybrid approach guarantees never worse than v1 rules
- Online evaluation detects degradation before it accumulates
- Framework supports arbitrary future policy types

### Negative

- Increased complexity in the decision layer
- Requires sufficient telemetry data per domain before policies are useful
- Policy training introduces an offline pipeline dependency
- More CSV output per run
- Dashboard complexity increases

### Risks

- Learned policies may overfit to specific instances
- Policy confidence may be poorly calibrated initially
- Counterfactual estimation may be inaccurate for complex decisions
- Policy drift could occur if domain characteristics change

### Mitigations

- Mandatory shadow validation before production promotion
- Hard safety constraints override all policy decisions
- Fallback to rules on low confidence or poor recent performance
- Minimum training sample requirements (50+) before learned policy activates
- Regular offline validation against held-out data

---

## Relationship to v1

v2 is a superset of v1. All v1 behaviour is preserved:

- `off` mode: unchanged
- `shadow` mode: records policy decisions + counterfactuals (more data)
- `assist` mode: uses HybridPolicy (learned when confident, rules otherwise)
- `adaptive` mode: uses LearnedPolicy with online evaluation

Existing rule engines become `RulePolicy` implementations within the framework. No breaking changes to CLI, telemetry format, or dashboard.

---

## Decision

Adopt Search Intelligence 2.0 as the long-term architecture for adaptive optimisation decisions. Implement incrementally starting with the framework layer (Phase 1) to avoid risk.
