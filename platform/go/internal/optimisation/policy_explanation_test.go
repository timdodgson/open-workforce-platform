package optimisation

import (
	"strings"
	"testing"
)

func TestExplanationBuilder_EarlyStop(t *testing.T) {
	eb := NewExplanationBuilder()

	features := FeatureVector{
		Problem:         "cvrp",
		Algorithm:       "sa",
		BudgetConsumed:  0.80,
		PlateauLength:   60000,
		IterationBudget: 100000,
		ImprovementRate: 0.0,
		Temperature:     0.001,
		AcceptanceRate:  0.001,
	}
	decision := PolicyDecision{
		Action:     "early_stop",
		Confidence: 0.82,
	}

	exp := eb.Explain(features, decision)

	if exp.Action != "early_stop" {
		t.Errorf("Action = %q, want early_stop", exp.Action)
	}
	if len(exp.Contributions) == 0 {
		t.Fatal("expected contributions")
	}

	// Plateau should be a strong "for" contributor.
	found := false
	for _, c := range exp.Contributions {
		if c.Feature == "plateau_length" && c.Direction == "for" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected plateau_length as 'for' contributor to early_stop")
	}

	// Summary should be non-empty and contain "stopped".
	if !strings.Contains(exp.Summary, "stopped") {
		t.Errorf("Summary = %q, expected to contain 'stopped'", exp.Summary)
	}
	if exp.ReasonCode == "" {
		t.Error("ReasonCode should not be empty")
	}
}

func TestExplanationBuilder_RunWorker(t *testing.T) {
	eb := NewExplanationBuilder()

	features := FeatureVector{
		Problem:          "nrp",
		Algorithm:        "sa",
		BudgetConsumed:   0.0,
		DistanceFromBest: 200,
		Entropy:          0.85,
		BeamHealth:       75.0,
		ImprovementRate:  3.5,
	}
	decision := PolicyDecision{
		Action:     "run",
		Confidence: 0.90,
	}

	exp := eb.Explain(features, decision)

	if exp.Action != "run" {
		t.Errorf("Action = %q, want run", exp.Action)
	}

	// Beam health should contribute "for" a run decision.
	found := false
	for _, c := range exp.Contributions {
		if c.Feature == "beam_health" && c.Direction == "for" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected beam_health as 'for' contributor to run")
	}

	if !strings.Contains(exp.Summary, "submitted") {
		t.Errorf("Summary = %q, expected 'submitted'", exp.Summary)
	}
}

func TestExplanationBuilder_SkipWorker(t *testing.T) {
	eb := NewExplanationBuilder()

	features := FeatureVector{
		Problem:          "nrp",
		Algorithm:        "sa",
		DistanceFromBest: 2000,
		Entropy:          0.2,
		BeamHealth:       30.0,
		ImprovementRate:  0.0,
	}
	decision := PolicyDecision{
		Action:     "skip",
		Confidence: 0.75,
	}

	exp := eb.Explain(features, decision)

	// Distance from best should contribute "for" skip.
	found := false
	for _, c := range exp.Contributions {
		if c.Feature == "distance_from_best" && c.Direction == "for" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected distance_from_best as 'for' contributor to skip")
	}

	// Entropy should also contribute "for" skip (low entropy).
	foundEntropy := false
	for _, c := range exp.Contributions {
		if c.Feature == "entropy" && c.Direction == "for" {
			foundEntropy = true
			break
		}
	}
	if !foundEntropy {
		t.Error("expected entropy as 'for' contributor to skip (low diversity)")
	}
}

func TestExplanationBuilder_MaxContributions(t *testing.T) {
	eb := NewExplanationBuilder()

	// Provide many features.
	features := FeatureVector{
		Problem: "cvrp", Algorithm: "sa",
		BudgetConsumed: 0.7, PlateauLength: 30000, IterationBudget: 100000,
		ImprovementRate: 0.5, DistanceFromBest: 500, Temperature: 0.5,
		Entropy: 0.6, BeamHealth: 60.0, AcceptanceRate: 0.02,
	}
	decision := PolicyDecision{Action: "continue", Confidence: 0.65}

	exp := eb.Explain(features, decision)

	if len(exp.Contributions) > 6 {
		t.Errorf("contributions = %d, should be capped at 6", len(exp.Contributions))
	}
}

func TestExplanationBuilder_ContributionsSorted(t *testing.T) {
	eb := NewExplanationBuilder()

	features := FeatureVector{
		Problem: "jss", Algorithm: "tabu",
		BudgetConsumed: 0.5, PlateauLength: 50000, IterationBudget: 100000,
		ImprovementRate: 0.1, AcceptanceRate: 0.01,
	}
	decision := PolicyDecision{Action: "early_stop", Confidence: 0.70}

	exp := eb.Explain(features, decision)

	if len(exp.Contributions) < 2 {
		t.Fatal("expected at least 2 contributions")
	}

	// Should be sorted by absolute contribution descending.
	for i := 1; i < len(exp.Contributions); i++ {
		prev := exp.Contributions[i-1].Contribution
		curr := exp.Contributions[i].Contribution
		if abs(prev) < abs(curr) {
			t.Errorf("contributions not sorted: [%d]=%f < [%d]=%f",
				i-1, abs(prev), i, abs(curr))
		}
	}
}

func TestExplanationBuilder_EmptyFeatures(t *testing.T) {
	eb := NewExplanationBuilder()

	features := FeatureVector{Problem: "cvrp", Algorithm: "sa"}
	decision := PolicyDecision{Action: "continue", Confidence: 0.5}

	exp := eb.Explain(features, decision)

	// Should produce a valid explanation even with zero features.
	if exp.Summary == "" {
		t.Error("Summary should not be empty")
	}
}

func TestExplanationBuilder_ReasonCode(t *testing.T) {
	eb := NewExplanationBuilder()

	features := FeatureVector{
		Problem: "cvrp", Algorithm: "sa",
		BudgetConsumed: 0.9, PlateauLength: 80000, IterationBudget: 100000,
	}
	decision := PolicyDecision{Action: "early_stop", Confidence: 0.85}

	exp := eb.Explain(features, decision)

	if exp.ReasonCode == "" {
		t.Error("ReasonCode should not be empty")
	}
	// Should contain feature names with direction.
	if !strings.Contains(exp.ReasonCode, ":") {
		t.Errorf("ReasonCode = %q, expected feature:direction format", exp.ReasonCode)
	}
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
