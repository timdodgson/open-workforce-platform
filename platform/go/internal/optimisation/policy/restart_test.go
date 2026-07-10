package policy

import (
	"testing"
)

func testRestartModel() *RestartModel {
	return &RestartModel{
		Version:   "1.0.0",
		TrainedOn: 80,
		Entries: []RestartModelEntry{
			{
				Domain: "cvrp", Algorithm: "sa",
				OptimalBudgetFraction: 0.60, OptimalPlateauRatio: 0.25,
				RestartSuccessRate: 0.55, MeanImprovAfterRestart: 30.0, MeanWasteIfFailed: 20000,
				BestRestartAlgorithm: "lahc", SameAlgoSuccessRate: 0.45, SwitchAlgoSuccessRate: 0.65,
				OptimalRestartBudget: 0.40,
				SampleCount:          40, Confidence: 0.80,
			},
			{
				Domain: "jss", Algorithm: "tabu",
				OptimalBudgetFraction: 0.50, OptimalPlateauRatio: 0.20,
				RestartSuccessRate: 0.70, MeanImprovAfterRestart: 15.0, MeanWasteIfFailed: 10000,
				BestRestartAlgorithm: "tabu", SameAlgoSuccessRate: 0.70, SwitchAlgoSuccessRate: 0.40,
				OptimalRestartBudget: 0.30,
				SampleCount:          25, Confidence: 0.75,
			},
			{
				Domain: "vrptw", Algorithm: "sa",
				OptimalBudgetFraction: 0.70, OptimalPlateauRatio: 0.30,
				RestartSuccessRate: 0.40, MeanImprovAfterRestart: 50.0, MeanWasteIfFailed: 30000,
				BestRestartAlgorithm: "sa", SameAlgoSuccessRate: 0.40, SwitchAlgoSuccessRate: 0.35,
				OptimalRestartBudget: 0.45,
				SampleCount:          8, Confidence: 0.55, // below threshold
			},
		},
	}
}

func TestRestartPolicy_NoModel_RuleBased(t *testing.T) {
	p := NewRestartPolicy(nil, DefaultRestartPolicyConfig())

	fv := FeatureVector{
		Problem: "cvrp", Algorithm: "sa",
		BudgetConsumed: 0.5, PlateauLength: 40000, IterationBudget: 100000,
		IterationsComplete: 50000,
	}
	d := p.Evaluate(fv)

	// Plateau ratio = 0.4 > 0.30 and budget > 0.25 → should restart (rule).
	if !d.ShouldRestart {
		t.Error("rule-based should restart with plateau_ratio 0.4")
	}
	if d.PolicyID != "restart-rules" {
		t.Errorf("PolicyID = %q, want restart-rules", d.PolicyID)
	}
	if d.Confidence != 0.45 {
		t.Errorf("Confidence = %f, want 0.45 (rule)", d.Confidence)
	}
}

func TestRestartPolicy_NoModel_LowPlateau_NoRestart(t *testing.T) {
	p := NewRestartPolicy(nil, DefaultRestartPolicyConfig())

	fv := FeatureVector{
		Problem: "cvrp", Algorithm: "sa",
		BudgetConsumed: 0.5, PlateauLength: 10000, IterationBudget: 100000,
		IterationsComplete: 50000,
	}
	d := p.Evaluate(fv)

	// Plateau ratio = 0.1 < 0.30 → should NOT restart.
	if d.ShouldRestart {
		t.Error("should not restart with plateau_ratio 0.1")
	}
}

func TestRestartPolicy_Learned_RecommendsRestart(t *testing.T) {
	p := NewRestartPolicy(testRestartModel(), DefaultRestartPolicyConfig())

	// CVRP SA: optimal budget fraction = 0.60, optimal plateau ratio = 0.25.
	// budget=0.60, plateau_ratio = 30000/100000 = 0.30.
	fv := FeatureVector{
		Problem: "cvrp", Algorithm: "sa",
		BudgetConsumed: 0.60, PlateauLength: 30000, IterationBudget: 100000,
		IterationsComplete: 60000,
	}
	d := p.Evaluate(fv)

	if !d.ShouldRestart {
		t.Errorf("learned policy should restart at optimal timing, got reason: %s", d.Reason)
	}
	// Should recommend LAHC (switch > same success rate).
	if d.RestartAlgorithm != "lahc" {
		t.Errorf("RestartAlgorithm = %q, want lahc", d.RestartAlgorithm)
	}
	if d.Confidence <= 0 {
		t.Errorf("Confidence = %f, should be > 0", d.Confidence)
	}
	if d.RestartBudget <= 0 {
		t.Errorf("RestartBudget = %d, should be > 0", d.RestartBudget)
	}
}

func TestRestartPolicy_Learned_BelowMinBudget(t *testing.T) {
	p := NewRestartPolicy(testRestartModel(), DefaultRestartPolicyConfig())

	fv := FeatureVector{
		Problem: "cvrp", Algorithm: "sa",
		BudgetConsumed: 0.15, PlateauLength: 50000, IterationBudget: 100000,
		IterationsComplete: 15000,
	}
	d := p.Evaluate(fv)

	if d.ShouldRestart {
		t.Error("should never restart below min budget fraction")
	}
}

func TestRestartPolicy_Learned_LowConfidence_FallsBack(t *testing.T) {
	p := NewRestartPolicy(testRestartModel(), DefaultRestartPolicyConfig())

	// VRPTW SA has confidence 0.55 (below threshold 0.60).
	fv := FeatureVector{
		Problem: "vrptw", Algorithm: "sa",
		BudgetConsumed: 0.70, PlateauLength: 40000, IterationBudget: 100000,
		IterationsComplete: 70000,
	}
	d := p.Evaluate(fv)

	// Should fall back to rules (confidence check in learned path).
	if d.Confidence > 0.60 {
		t.Errorf("Confidence = %f, expected rule-level (low conf fallback)", d.Confidence)
	}
}

func TestRestartPolicy_Learned_SameAlgoWhenBetter(t *testing.T) {
	p := NewRestartPolicy(testRestartModel(), DefaultRestartPolicyConfig())

	// JSS Tabu: same_algo > switch_algo success rate → keep tabu.
	fv := FeatureVector{
		Problem: "jss", Algorithm: "tabu",
		BudgetConsumed: 0.50, PlateauLength: 25000, IterationBudget: 100000,
		IterationsComplete: 50000,
	}
	d := p.Evaluate(fv)

	if d.RestartAlgorithm != "tabu" {
		t.Errorf("RestartAlgorithm = %q, want tabu (same algo better)", d.RestartAlgorithm)
	}
}

func TestRestartPolicy_Learned_BadTiming_NoRestart(t *testing.T) {
	p := NewRestartPolicy(testRestartModel(), DefaultRestartPolicyConfig())

	// CVRP SA optimal = 0.60, but we're at 0.95 (bad timing, score ≈ 0).
	fv := FeatureVector{
		Problem: "cvrp", Algorithm: "sa",
		BudgetConsumed: 0.95, PlateauLength: 5000, IterationBudget: 100000,
		IterationsComplete: 95000,
	}
	d := p.Evaluate(fv)

	// Timing is bad + plateau is shallow → success prob should be below threshold.
	if d.ShouldRestart {
		t.Errorf("should not restart at bad timing, got reason: %s", d.Reason)
	}
}

func TestRestartPolicy_RestartBudget_Clamped(t *testing.T) {
	p := NewRestartPolicy(testRestartModel(), DefaultRestartPolicyConfig())

	fv := FeatureVector{
		Problem: "cvrp", Algorithm: "sa",
		BudgetConsumed: 0.60, PlateauLength: 30000, IterationBudget: 100000,
		IterationsComplete: 60000,
	}
	d := p.Evaluate(fv)

	// Remaining = 40000, optimal_restart_budget = 0.40, max = 0.50.
	// restart_budget = 40000 * 0.40 = 16000.
	if d.RestartBudget > 20000 {
		t.Errorf("RestartBudget = %d, should be capped by MaxRestartBudgetFraction", d.RestartBudget)
	}
	if d.RestartBudget < 1000 {
		t.Errorf("RestartBudget = %d, should be at least 1000", d.RestartBudget)
	}
}

func TestRestartPolicy_Metadata(t *testing.T) {
	p := NewRestartPolicy(testRestartModel(), DefaultRestartPolicyConfig())
	m := p.Metadata()

	if m.Type != "learned" {
		t.Errorf("Type = %q, want learned", m.Type)
	}
	if m.DecisionType != "restart" {
		t.Errorf("DecisionType = %q, want restart", m.DecisionType)
	}
	if m.TrainedSamples != 80 {
		t.Errorf("TrainedSamples = %d, want 80", m.TrainedSamples)
	}
}
