// hybrid_executor.go — HybridExecutor (SI 2.0 unified policy decision pipeline).
//
// Production solvers use PolicySearchHookRunner instead. This type remains for
// tests and a future consolidation sprint; do not wire both in the same hot loop.
//
// Decision flow: RulePolicy → LearnedPolicy → confidence check → HybridPolicy → safety → decision.
//
// Low confidence → fallback to RulePolicy.
// High confidence → learned policy wins.
// Safety always overrides.
//
// Used by SI 2.0 tests and lifecycle tooling today.
// Not yet called from production solver hot paths (search.go / pfrs_search.go).
package optimisation

import "time"

// ───────────────────────────────────────────────────────────────
// Execution Result
// ───────────────────────────────────────────────────────────────

// ExecutionResult is the final output of the hybrid executor.
// Contains the decision, full provenance, and explanation.
type ExecutionResult struct {
	// Final decision after all stages.
	Decision PolicyDecision

	// Which source produced the decision.
	Source ExecutionSource

	// Explanation of how the decision was reached.
	Explanation PolicyExplanation

	// Rule baseline (always computed for comparison).
	RuleDecision PolicyDecision

	// Learned decision (if available, even if not used).
	LearnedDecision *PolicyDecision

	// Safety override applied.
	SafetyOverride bool
	SafetyRule     string

	// Fallback details.
	FallbackOccurred bool
	FallbackReason   string

	// Hierarchy level used.
	HierarchyLevel PolicyLevel

	// Timing.
	ExecutedAt time.Time
}

// ExecutionSource identifies where the final decision came from.
type ExecutionSource string

const (
	SourceRule    ExecutionSource = "rule"
	SourceLearned ExecutionSource = "learned"
	SourceHybrid  ExecutionSource = "hybrid"
	SourceSafety  ExecutionSource = "safety"
)

// ───────────────────────────────────────────────────────────────
// Safety Gate
// ───────────────────────────────────────────────────────────────

// SafetyConstraint defines one non-negotiable safety rule.
type SafetyConstraint struct {
	Name     string
	Check    func(ctx PolicyContext, decision PolicyDecision) bool // returns true if violated
	Override PolicyDecision                                        // safe action to take when violated
}

// ───────────────────────────────────────────────────────────────
// Hybrid Executor
// ───────────────────────────────────────────────────────────────

// HybridExecutorConfig configures the executor.
type HybridExecutorConfig struct {
	// Confidence threshold: below this, fallback to rule.
	ConfidenceThreshold float64

	// Safety constraints (always override).
	SafetyConstraints []SafetyConstraint
}

// DefaultHybridExecutorConfig returns sensible defaults.
func DefaultHybridExecutorConfig() HybridExecutorConfig {
	return HybridExecutorConfig{
		ConfidenceThreshold: 0.60,
	}
}

// HybridExecutor orchestrates the full decision flow.
type HybridExecutor struct {
	hierarchy *PolicyHierarchy
	config    HybridExecutorConfig
	explainer *ExplanationBuilder
	evaluator *PolicyEvaluator
	recorder  *CounterfactualRecorder
}

// NewHybridExecutor creates the executor with all dependencies.
func NewHybridExecutor(
	hierarchy *PolicyHierarchy,
	config HybridExecutorConfig,
	evaluator *PolicyEvaluator,
	recorder *CounterfactualRecorder,
) *HybridExecutor {
	if config.ConfidenceThreshold <= 0 {
		config.ConfidenceThreshold = 0.60
	}
	return &HybridExecutor{
		hierarchy: hierarchy,
		config:    config,
		explainer: NewExplanationBuilder(),
		evaluator: evaluator,
		recorder:  recorder,
	}
}

// Execute runs the full hybrid decision flow.
func (e *HybridExecutor) Execute(ctx PolicyContext) ExecutionResult {
	now := time.Now()

	// Step 1: Always compute rule baseline.
	ruleDecision := e.computeRuleDecision(ctx)

	// Step 2: Resolve and execute learned policy via hierarchy.
	hierarchyDecision := e.hierarchy.DecideWithHierarchy(ctx)

	// Step 3: Confidence gate — decide which source wins.
	var finalDecision PolicyDecision
	var source ExecutionSource
	var fallback bool
	var fallbackReason string
	var learnedDec *PolicyDecision

	if hierarchyDecision.ResolvedLevel == LevelNone {
		// No policy at any level — use rules.
		finalDecision = ruleDecision
		source = SourceRule
		fallback = true
		fallbackReason = "no_learned_policy"
	} else if hierarchyDecision.PolicyDecision.Action == "defer" {
		// Policy explicitly deferred — use rules.
		finalDecision = ruleDecision
		source = SourceRule
		fallback = true
		fallbackReason = "policy_deferred"
		ld := hierarchyDecision.PolicyDecision
		learnedDec = &ld
	} else if hierarchyDecision.PolicyDecision.Confidence < e.config.ConfidenceThreshold {
		// Low confidence — fallback to rule.
		finalDecision = ruleDecision
		source = SourceRule
		fallback = true
		fallbackReason = "low_confidence"
		ld := hierarchyDecision.PolicyDecision
		learnedDec = &ld
	} else {
		// High confidence — policy wins.
		finalDecision = hierarchyDecision.PolicyDecision
		source = SourceLearned
		ld := hierarchyDecision.PolicyDecision
		learnedDec = &ld
	}

	// Step 4: Safety gate — always override if violated.
	safetyOverride := false
	safetyRule := ""
	for _, constraint := range e.config.SafetyConstraints {
		if constraint.Check(ctx, finalDecision) {
			finalDecision = constraint.Override
			finalDecision.PolicyID = "safety:" + constraint.Name
			finalDecision.PolicyVersion = "1.0.0"
			finalDecision.Reason = "safety_override:" + constraint.Name
			source = SourceSafety
			safetyOverride = true
			safetyRule = constraint.Name
			break
		}
	}

	// Step 5: Build explanation.
	explanation := e.explainer.Explain(ctx.Features, finalDecision)

	// Step 6: Record for evaluation.
	if e.evaluator != nil {
		e.evaluator.Record(PolicyEvaluationRecord{
			Timestamp:           now,
			RunID:               ctx.Features.RunID,
			DecisionType:        ctx.DecisionType,
			Domain:              ctx.Domain,
			Instance:            ctx.Instance,
			Algorithm:           ctx.Features.Algorithm,
			PolicyID:            finalDecision.PolicyID,
			PolicyVersion:       finalDecision.PolicyVersion,
			PolicyType:          string(source),
			Action:              finalDecision.Action,
			Confidence:          finalDecision.Confidence,
			ExpectedImprovement: 0, // filled by caller post-execution
		})
	}

	// Step 7: Record counterfactual.
	if e.recorder != nil {
		counterfactuals := []CounterfactualAction{}
		if source != SourceRule {
			counterfactuals = append(counterfactuals, CounterfactualAction{
				Action: ruleDecision.Action, Source: "rule",
			})
		}
		if learnedDec != nil && source != SourceLearned {
			counterfactuals = append(counterfactuals, CounterfactualAction{
				Action: learnedDec.Action, Source: "learned",
			})
		}

		e.recorder.Record(CounterfactualRecord{
			Timestamp:             now,
			RunID:                 ctx.Features.RunID,
			DecisionType:          ctx.DecisionType,
			Domain:                ctx.Domain,
			Instance:              ctx.Instance,
			Algorithm:             ctx.Features.Algorithm,
			ActualAction:          finalDecision.Action,
			ActualConfidence:      finalDecision.Confidence,
			PolicyID:              finalDecision.PolicyID,
			PolicyVersion:         finalDecision.PolicyVersion,
			PolicyType:            string(source),
			CounterfactualActions: counterfactuals,
		})
	}

	return ExecutionResult{
		Decision:         finalDecision,
		Source:           source,
		Explanation:      explanation,
		RuleDecision:     ruleDecision,
		LearnedDecision:  learnedDec,
		SafetyOverride:   safetyOverride,
		SafetyRule:       safetyRule,
		FallbackOccurred: fallback,
		FallbackReason:   fallbackReason,
		HierarchyLevel:   hierarchyDecision.ResolvedLevel,
		ExecutedAt:       now,
	}
}

func (e *HybridExecutor) computeRuleDecision(ctx PolicyContext) PolicyDecision {
	// Use a global rule policy from the hierarchy if available.
	globalPolicy := e.hierarchy.getGlobalPolicy(ctx)
	if globalPolicy != nil {
		d := globalPolicy.Decide(ctx)
		if d.Action != "defer" {
			return d
		}
	}

	// Absolute fallback.
	return PolicyDecision{
		Action:        "continue",
		Confidence:    0.5,
		Reason:        "default_rule",
		PolicyID:      "default-rules",
		PolicyVersion: "1.0.0",
	}
}
