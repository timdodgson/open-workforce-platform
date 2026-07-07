package optimisation

import (
	"testing"
)

func TestPortfolioBudgetRulePolicy_SABoost(t *testing.T) {
	policy := NewPortfolioBudgetRulePolicy("cvrp")

	ctx := PolicyContext{
		DecisionType: "portfolio",
		Domain:       "cvrp",
		Features: FeatureVector{
			Algorithm:   "sa",
			WorkerCount: 3,
		},
	}

	d := policy.Decide(ctx)

	if d.Action != "allocate" {
		t.Errorf("Action = %q, want allocate", d.Action)
	}
	mult, ok := d.Parameters["budget_mult"].(float64)
	if !ok || mult != 1.1 {
		t.Errorf("budget_mult = %v, want 1.1", d.Parameters["budget_mult"])
	}
}

func TestPortfolioBudgetRulePolicy_TabuOnJSS(t *testing.T) {
	policy := NewPortfolioBudgetRulePolicy("jss")

	ctx := PolicyContext{
		DecisionType: "portfolio",
		Domain:       "jss",
		Features: FeatureVector{
			Algorithm:   "tabu",
			WorkerCount: 3,
		},
	}

	d := policy.Decide(ctx)

	mult, ok := d.Parameters["budget_mult"].(float64)
	if !ok || mult != 1.15 {
		t.Errorf("budget_mult = %v, want 1.15", d.Parameters["budget_mult"])
	}
}

func TestPortfolioBudgetRulePolicy_LAHCReduce(t *testing.T) {
	policy := NewPortfolioBudgetRulePolicy("cvrp")

	ctx := PolicyContext{
		DecisionType: "portfolio",
		Domain:       "cvrp",
		Features: FeatureVector{
			Algorithm:   "lahc",
			WorkerCount: 3,
		},
	}

	d := policy.Decide(ctx)

	mult, ok := d.Parameters["budget_mult"].(float64)
	if !ok || mult != 0.9 {
		t.Errorf("budget_mult = %v, want 0.9", d.Parameters["budget_mult"])
	}
}

func TestPortfolioBudgetModelV2_Predict(t *testing.T) {
	model := &PortfolioBudgetModel{
		Version:   "2.0.0",
		TrainedOn: 100,
		Entries: []StrategyPerformanceEntry{
			{Domain: "cvrp", Strategy: "sa", WinRate: 0.7, MeanImprovement: 50.0, SampleCount: 20, RecommendedMult: 1.3, Confidence: 0.80},
			{Domain: "cvrp", Strategy: "lahc", WinRate: 0.2, MeanImprovement: 10.0, SampleCount: 20, RecommendedMult: 0.7, Confidence: 0.75},
		},
	}

	m := NewPortfolioBudgetModelV2(model)

	// SA should predict with high confidence.
	pred := m.Predict(FeatureVector{Problem: "cvrp", Algorithm: "sa"})
	if pred.Action != "allocate" {
		t.Errorf("SA action = %q, want allocate", pred.Action)
	}
	if pred.Confidence != 0.80 {
		t.Errorf("SA confidence = %f, want 0.80", pred.Confidence)
	}

	// Unknown strategy should defer.
	pred = m.Predict(FeatureVector{Problem: "cvrp", Algorithm: "unknown"})
	if pred.Action != "defer" {
		t.Errorf("unknown action = %q, want defer", pred.Action)
	}
}

func TestPortfolioBudgetModelV2_InsufficientSamples(t *testing.T) {
	model := &PortfolioBudgetModel{
		Version:   "2.0.0",
		TrainedOn: 5,
		Entries: []StrategyPerformanceEntry{
			{Domain: "jss", Strategy: "tabu", WinRate: 0.9, SampleCount: 2, RecommendedMult: 1.5, Confidence: 0.70},
		},
	}

	m := NewPortfolioBudgetModelV2(model)
	pred := m.Predict(FeatureVector{Problem: "jss", Algorithm: "tabu"})

	if pred.Action != "defer" {
		t.Errorf("low sample action = %q, want defer", pred.Action)
	}
}

func TestPortfolioBudgetModelV2_BudgetMultiplier_Clamped(t *testing.T) {
	model := &PortfolioBudgetModel{
		Version:   "2.0.0",
		TrainedOn: 50,
		Entries: []StrategyPerformanceEntry{
			{Domain: "cvrp", Strategy: "sa", RecommendedMult: 5.0, SampleCount: 10, Confidence: 0.9},
			{Domain: "cvrp", Strategy: "lahc", RecommendedMult: 0.1, SampleCount: 10, Confidence: 0.9},
		},
	}

	m := NewPortfolioBudgetModelV2(model)

	// Should clamp to 2.0 max.
	if mult := m.BudgetMultiplier("cvrp", "", "sa"); mult != 2.0 {
		t.Errorf("SA mult = %f, want 2.0 (clamped)", mult)
	}
	// Should clamp to 0.25 min.
	if mult := m.BudgetMultiplier("cvrp", "", "lahc"); mult != 0.25 {
		t.Errorf("LAHC mult = %f, want 0.25 (clamped)", mult)
	}
}

func TestNewPortfolioBudgetPolicy_NoModel(t *testing.T) {
	policy := NewPortfolioBudgetPolicy(PortfolioBudgetPolicyConfig{
		Domain:    "cvrp",
		ModelPath: "",
	})

	m := policy.Metadata()
	if m.Type != "rule" {
		t.Errorf("no-model policy type = %q, want rule", m.Type)
	}
}

func TestNewPortfolioBudgetPolicy_InvalidModel(t *testing.T) {
	policy := NewPortfolioBudgetPolicy(PortfolioBudgetPolicyConfig{
		Domain:    "cvrp",
		ModelPath: "/nonexistent/path.json",
	})

	m := policy.Metadata()
	if m.Type != "rule" {
		t.Errorf("invalid model policy type = %q, want rule (fallback)", m.Type)
	}
}

func TestAllocateBudgetsViaPolicy(t *testing.T) {
	policy := NewPortfolioBudgetRulePolicy("cvrp")
	extractor := NewFeatureExtractor()

	allocs := AllocateBudgetsViaPolicy(
		policy,
		[]string{"sa", "lahc", "tabu"},
		100000,
		"cvrp", "A-n32-k5", "test-run",
		extractor,
	)

	if len(allocs) != 3 {
		t.Fatalf("expected 3 allocations, got %d", len(allocs))
	}

	// SA should get boost.
	if allocs[0].BudgetMult != 1.1 {
		t.Errorf("SA mult = %f, want 1.1", allocs[0].BudgetMult)
	}
	if allocs[0].FinalBudget != 110000 {
		t.Errorf("SA budget = %d, want 110000", allocs[0].FinalBudget)
	}

	// LAHC should get reduction.
	if allocs[1].BudgetMult != 0.9 {
		t.Errorf("LAHC mult = %f, want 0.9", allocs[1].BudgetMult)
	}

	// All decisions should have policy metadata.
	for i, a := range allocs {
		if a.Decision.PolicyID == "" {
			t.Errorf("alloc[%d] missing PolicyID", i)
		}
		if a.Decision.PolicyVersion == "" {
			t.Errorf("alloc[%d] missing PolicyVersion", i)
		}
	}
}
