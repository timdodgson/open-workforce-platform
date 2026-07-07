package optimisation

import (
	"testing"
)

func setupHybridExecutor(learnedConfidence float64) *HybridExecutor {
	h := NewPolicyHierarchy()

	// Global rule policy.
	h.RegisterGlobal("search", NewRulePolicy("global-rules", "1.0.0", "*", "search", []Rule{
		{Name: "rule_continue", Matches: func(_ PolicyContext) bool { return true },
			Decide: func(_ PolicyContext) PolicyDecision {
				return PolicyDecision{Action: "continue", Confidence: 1.0, Reason: "rule:default"}
			}},
	}))

	// Domain learned policy (mock).
	model := &mockModel{action: "early_stop", confidence: learnedConfidence}
	learned := NewLearnedPolicy(LearnedPolicyConfig{
		ID: "cvrp-learned", Version: "2.0.0", Domain: "cvrp",
		DecisionType: "search", Model: model, Threshold: 0.011, // low threshold: let executor handle confidence gate
	})
	h.RegisterDomain("cvrp", "search", learned)

	config := HybridExecutorConfig{ConfidenceThreshold: 0.60}
	evaluator := NewPolicyEvaluator()
	recorder := NewCounterfactualRecorder("")

	return NewHybridExecutor(h, config, evaluator, recorder)
}

func TestHybridExecutor_HighConfidence_PolicyWins(t *testing.T) {
	exec := setupHybridExecutor(0.85)

	ctx := PolicyContext{
		DecisionType: "search",
		Domain:       "cvrp",
		Instance:     "A-n32-k5",
		Features:     FeatureVector{Problem: "cvrp", Algorithm: "sa", BudgetConsumed: 0.7, PlateauLength: 50000, IterationBudget: 100000},
	}

	result := exec.Execute(ctx)

	if result.Decision.Action != "early_stop" {
		t.Errorf("Action = %q, want early_stop (learned)", result.Decision.Action)
	}
	if result.Source != SourceLearned {
		t.Errorf("Source = %q, want learned", result.Source)
	}
	if result.FallbackOccurred {
		t.Error("should not fallback with high confidence")
	}
}

func TestHybridExecutor_LowConfidence_FallbackToRule(t *testing.T) {
	h := NewPolicyHierarchy()
	h.RegisterGlobal("search", NewRulePolicy("global-rules", "1.0.0", "*", "search", []Rule{
		{Name: "rule_continue", Matches: func(_ PolicyContext) bool { return true },
			Decide: func(_ PolicyContext) PolicyDecision {
				return PolicyDecision{Action: "continue", Confidence: 1.0, Reason: "rule:default"}
			}},
	}))
	model := &mockModel{action: "early_stop", confidence: 0.40}
	learned := NewLearnedPolicy(LearnedPolicyConfig{
		ID: "cvrp-learned", Version: "2.0.0", Domain: "cvrp",
		DecisionType: "search", Model: model, Threshold: 0.011,
	})
	h.RegisterDomain("cvrp", "search", learned)

	ctx := PolicyContext{
		DecisionType: "search",
		Domain:       "cvrp",
		Instance:     "A-n32-k5",
		Features:     FeatureVector{Problem: "cvrp", Algorithm: "sa"},
	}

	// Verify hierarchy resolves at domain with confidence 0.40.
	hd := h.DecideWithHierarchy(ctx)
	if hd.ResolvedLevel != LevelDomain {
		t.Fatalf("hierarchy level = %q, want domain", hd.ResolvedLevel)
	}
	if hd.PolicyDecision.Confidence != 0.40 {
		t.Fatalf("hierarchy confidence = %f, want 0.40", hd.PolicyDecision.Confidence)
	}

	config := HybridExecutorConfig{ConfidenceThreshold: 0.60}
	exec := NewHybridExecutor(h, config, NewPolicyEvaluator(), NewCounterfactualRecorder(""))
	result := exec.Execute(ctx)

	if result.Source != SourceRule {
		t.Errorf("Source = %q, want rule", result.Source)
	}
	if !result.FallbackOccurred {
		t.Error("should fallback with confidence 0.40 < 0.60")
	}
	if result.FallbackReason != "low_confidence" {
		t.Errorf("FallbackReason = %q, want low_confidence", result.FallbackReason)
	}
}

func TestHybridExecutor_NoPolicyAvailable_FallbackToRule(t *testing.T) {
	h := NewPolicyHierarchy()
	h.RegisterGlobal("search", NewRulePolicy("rules", "1.0.0", "*", "search", []Rule{
		{Name: "default", Matches: func(_ PolicyContext) bool { return true },
			Decide: func(_ PolicyContext) PolicyDecision {
				return PolicyDecision{Action: "continue", Confidence: 1.0, Reason: "rule:only"}
			}},
	}))

	exec := NewHybridExecutor(h, DefaultHybridExecutorConfig(), nil, nil)

	ctx := PolicyContext{
		DecisionType: "search",
		Domain:       "jss", // no JSS domain policy registered — only global rule
		Instance:     "la01",
		Features:     FeatureVector{Problem: "jss", Algorithm: "tabu"},
	}

	result := exec.Execute(ctx)

	// Global rule provides the decision — resolves at global level.
	if result.Decision.Action != "continue" {
		t.Errorf("Action = %q, want continue", result.Decision.Action)
	}
	// Hierarchy resolved at global (rule) — confidence is 1.0 so no fallback needed.
	// The global rule IS the rule baseline — this is not a "fallback" scenario.
	if result.HierarchyLevel != LevelGlobal {
		t.Errorf("HierarchyLevel = %q, want global", result.HierarchyLevel)
	}
}

func TestHybridExecutor_SafetyOverrides(t *testing.T) {
	exec := setupHybridExecutor(0.95) // high confidence, but safety will override

	config := DefaultHybridExecutorConfig()
	config.SafetyConstraints = []SafetyConstraint{
		{
			Name: "never_stop_before_20pct",
			Check: func(ctx PolicyContext, d PolicyDecision) bool {
				return d.Action == "early_stop" && ctx.Features.BudgetConsumed < 0.20
			},
			Override: PolicyDecision{Action: "continue", Confidence: 1.0, Reason: "safety:min_budget"},
		},
	}

	h := NewPolicyHierarchy()
	model := &mockModel{action: "early_stop", confidence: 0.95}
	learned := NewLearnedPolicy(LearnedPolicyConfig{
		ID: "unsafe", Version: "1.0.0", Domain: "cvrp", DecisionType: "search",
		Model: model, Threshold: 0.01,
	})
	h.RegisterDomain("cvrp", "search", learned)
	h.RegisterGlobal("search", globalRulePolicy())

	exec = NewHybridExecutor(h, config, nil, nil)

	ctx := PolicyContext{
		DecisionType: "search",
		Domain:       "cvrp",
		Features:     FeatureVector{Problem: "cvrp", Algorithm: "sa", BudgetConsumed: 0.10}, // only 10%
	}

	result := exec.Execute(ctx)

	if result.Decision.Action != "continue" {
		t.Errorf("Action = %q, want continue (safety override)", result.Decision.Action)
	}
	if result.Source != SourceSafety {
		t.Errorf("Source = %q, want safety", result.Source)
	}
	if !result.SafetyOverride {
		t.Error("SafetyOverride should be true")
	}
	if result.SafetyRule != "never_stop_before_20pct" {
		t.Errorf("SafetyRule = %q", result.SafetyRule)
	}
}

func TestHybridExecutor_RecordsEvaluation(t *testing.T) {
	h := NewPolicyHierarchy()
	h.RegisterGlobal("search", globalRulePolicy())
	evaluator := NewPolicyEvaluator()
	recorder := NewCounterfactualRecorder("")

	exec := NewHybridExecutor(h, DefaultHybridExecutorConfig(), evaluator, recorder)

	ctx := PolicyContext{
		DecisionType: "search", Domain: "nrp",
		Features: FeatureVector{Problem: "nrp", Algorithm: "sa", RunID: "test-run"},
	}
	exec.Execute(ctx)

	m := evaluator.Metrics()
	if m.TotalDecisions != 1 {
		t.Errorf("evaluator decisions = %d, want 1", m.TotalDecisions)
	}
}

func TestHybridExecutor_RecordsCounterfactual(t *testing.T) {
	h := NewPolicyHierarchy()
	h.RegisterGlobal("search", globalRulePolicy())

	model := &mockModel{action: "early_stop", confidence: 0.40}
	learned := NewLearnedPolicy(LearnedPolicyConfig{
		ID: "test", Version: "1.0.0", Domain: "cvrp", DecisionType: "search",
		Model: model, Threshold: 0.01, // no internal defer
	})
	h.RegisterDomain("cvrp", "search", learned)

	evaluator := NewPolicyEvaluator()
	recorder := NewCounterfactualRecorder("") // disabled (no file)

	exec := NewHybridExecutor(h, DefaultHybridExecutorConfig(), evaluator, recorder)

	ctx := PolicyContext{
		DecisionType: "search", Domain: "cvrp",
		Features: FeatureVector{Problem: "cvrp", Algorithm: "sa"},
	}
	result := exec.Execute(ctx)

	// Low confidence → fallback to rule. Learned decision should be captured.
	if !result.FallbackOccurred {
		t.Error("should fallback with 0.40 confidence")
	}
	if result.LearnedDecision == nil {
		t.Fatal("LearnedDecision should be captured for counterfactual")
	}
	if result.LearnedDecision.Action != "early_stop" {
		t.Errorf("LearnedDecision.Action = %q, want early_stop", result.LearnedDecision.Action)
	}
}

func TestHybridExecutor_ExplanationGenerated(t *testing.T) {
	exec := setupHybridExecutor(0.80)

	ctx := PolicyContext{
		DecisionType: "search", Domain: "cvrp",
		Features: FeatureVector{
			Problem: "cvrp", Algorithm: "sa",
			BudgetConsumed: 0.8, PlateauLength: 60000, IterationBudget: 100000,
		},
	}
	result := exec.Execute(ctx)

	if result.Explanation.Action == "" {
		t.Error("Explanation should be generated")
	}
	if result.Explanation.Summary == "" {
		t.Error("Explanation.Summary should not be empty")
	}
}

func TestHybridExecutor_HierarchyLevel(t *testing.T) {
	exec := setupHybridExecutor(0.85)

	ctx := PolicyContext{
		DecisionType: "search", Domain: "cvrp",
		Features: FeatureVector{Problem: "cvrp", Algorithm: "sa"},
	}
	result := exec.Execute(ctx)

	if result.HierarchyLevel != LevelDomain {
		t.Errorf("HierarchyLevel = %q, want domain", result.HierarchyLevel)
	}
}
