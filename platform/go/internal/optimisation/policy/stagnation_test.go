package policy

import (
	"testing"
)

func testCurveModel() *ImprovementCurveModel {
	return &ImprovementCurveModel{
		Version:   "1.0.0",
		TrainedOn: 100,
		Curves: []ImprovementCurveEntry{
			{
				Domain: "cvrp", Algorithm: "sa",
				DecayRate: 10.0, Amplitude: 0.95, HalfLife: 0.07,
				MeanImprovements: 5.0, MeanLastImproveAt: 0.4, StdLastImproveAt: 0.15,
				SampleCount: 50, Confidence: 0.85,
			},
			{
				Domain: "jss", Algorithm: "tabu",
				DecayRate: 8.0, Amplitude: 0.90, HalfLife: 0.09,
				MeanImprovements: 3.0, MeanLastImproveAt: 0.3, StdLastImproveAt: 0.1,
				SampleCount: 30, Confidence: 0.75,
			},
			{
				Domain: "cvrp", Algorithm: "sa", Instance: "A-n32-k5",
				DecayRate: 12.0, Amplitude: 0.98, HalfLife: 0.06,
				MeanImprovements: 8.0, MeanLastImproveAt: 0.35, StdLastImproveAt: 0.12,
				SampleCount: 20, Confidence: 0.80,
			},
			{
				Domain: "vrptw", Algorithm: "sa",
				DecayRate: 5.0, Amplitude: 0.80, HalfLife: 0.14,
				MeanImprovements: 10.0, MeanLastImproveAt: 0.6, StdLastImproveAt: 0.2,
				SampleCount: 5, Confidence: 0.50,
			},
		},
	}
}

func TestLearnedStagnationDetector_NoModel(t *testing.T) {
	d := NewLearnedStagnationDetector(nil, DefaultStagnationPolicyConfig())

	fv := FeatureVector{Problem: "cvrp", Algorithm: "sa", BudgetConsumed: 0.5, PlateauLength: 50000, IterationBudget: 100000}
	a := d.Assess(fv)

	if a.PolicyConfidence != 0.0 {
		t.Errorf("PolicyConfidence = %f, want 0.0 (no model)", a.PolicyConfidence)
	}
	if a.RecommendEarlyStop {
		t.Error("should not recommend stop without model")
	}
}

func TestLearnedStagnationDetector_NoCurveForConfig(t *testing.T) {
	d := NewLearnedStagnationDetector(testCurveModel(), DefaultStagnationPolicyConfig())

	fv := FeatureVector{Problem: "nrp", Algorithm: "sa", BudgetConsumed: 0.8, PlateauLength: 80000, IterationBudget: 100000}
	a := d.Assess(fv)

	if a.PolicyConfidence != 0.0 {
		t.Errorf("PolicyConfidence = %f, want 0.0 (no curve)", a.PolicyConfidence)
	}
}

func TestLearnedStagnationDetector_HighPlateau_RecommendsStop(t *testing.T) {
	d := NewLearnedStagnationDetector(testCurveModel(), DefaultStagnationPolicyConfig())

	// CVRP SA: decay_rate=10, amplitude=0.95
	// plateau_ratio = 80000/100000 = 0.8
	// P(improve) = 0.95 * exp(-10 * 0.8) = 0.95 * exp(-8) ≈ 0.00032
	// Well below threshold 0.10 → should recommend stop.
	fv := FeatureVector{
		Problem:         "cvrp",
		Algorithm:       "sa",
		BudgetConsumed:  0.8,
		PlateauLength:   80000,
		IterationBudget: 100000,
		AcceptanceRate:  0.02,
	}
	a := d.Assess(fv)

	if !a.RecommendEarlyStop {
		t.Errorf("should recommend stop, P(improve)=%f", a.ProbImprove)
	}
	if a.ProbImprove > 0.01 {
		t.Errorf("ProbImprove = %f, expected < 0.01", a.ProbImprove)
	}
	if a.StagnationConfidence < 0.5 {
		t.Errorf("StagnationConfidence = %f, expected > 0.5", a.StagnationConfidence)
	}
}

func TestLearnedStagnationDetector_LowPlateau_Continues(t *testing.T) {
	d := NewLearnedStagnationDetector(testCurveModel(), DefaultStagnationPolicyConfig())

	// CVRP SA: plateau_ratio = 5000/100000 = 0.05
	// P(improve) = 0.95 * exp(-10 * 0.05) = 0.95 * exp(-0.5) ≈ 0.576
	// Above threshold → should continue.
	fv := FeatureVector{
		Problem:         "cvrp",
		Algorithm:       "sa",
		BudgetConsumed:  0.3,
		PlateauLength:   5000,
		IterationBudget: 100000,
	}
	a := d.Assess(fv)

	if a.RecommendEarlyStop {
		t.Errorf("should continue, P(improve)=%f", a.ProbImprove)
	}
	if a.ProbImprove < 0.5 {
		t.Errorf("ProbImprove = %f, expected > 0.5", a.ProbImprove)
	}
}

func TestLearnedStagnationDetector_BelowMinBudget_NeverStops(t *testing.T) {
	d := NewLearnedStagnationDetector(testCurveModel(), DefaultStagnationPolicyConfig())

	// Budget consumed = 10% (below 20% minimum).
	// Even with extreme plateau, should not recommend stop.
	fv := FeatureVector{
		Problem:         "cvrp",
		Algorithm:       "sa",
		BudgetConsumed:  0.10,
		PlateauLength:   90000,
		IterationBudget: 100000,
	}
	a := d.Assess(fv)

	if a.RecommendEarlyStop {
		t.Error("should never stop below min budget fraction")
	}
}

func TestLearnedStagnationDetector_LowConfidence_NeverStops(t *testing.T) {
	d := NewLearnedStagnationDetector(testCurveModel(), DefaultStagnationPolicyConfig())

	// VRPTW SA has confidence 0.50 (below threshold 0.60).
	// Even with stagnation, should not recommend stop.
	fv := FeatureVector{
		Problem:         "vrptw",
		Algorithm:       "sa",
		BudgetConsumed:  0.8,
		PlateauLength:   80000,
		IterationBudget: 100000,
	}
	a := d.Assess(fv)

	if a.RecommendEarlyStop {
		t.Error("should not stop when model confidence is below threshold")
	}
	if a.PolicyConfidence >= 0.60 {
		t.Errorf("PolicyConfidence = %f, expected < 0.60 (5 samples)", a.PolicyConfidence)
	}
}

func TestLearnedStagnationDetector_InstanceSpecific(t *testing.T) {
	d := NewLearnedStagnationDetector(testCurveModel(), DefaultStagnationPolicyConfig())

	// Should use instance-specific curve for A-n32-k5 (decay=12) not domain-wide (decay=10).
	// plateau_ratio = 50000/100000 = 0.5
	// Instance: P = 0.98 * exp(-12 * 0.5) = 0.98 * exp(-6) ≈ 0.0024
	// Domain:   P = 0.95 * exp(-10 * 0.5) = 0.95 * exp(-5) ≈ 0.0064
	fv := FeatureVector{
		Problem:         "cvrp",
		Instance:        "A-n32-k5",
		Algorithm:       "sa",
		BudgetConsumed:  0.5,
		PlateauLength:   50000,
		IterationBudget: 100000,
	}
	a := d.Assess(fv)

	// Instance-specific should give even lower P(improve).
	if a.ProbImprove > 0.005 {
		t.Errorf("ProbImprove = %f, expected < 0.005 (instance-specific curve)", a.ProbImprove)
	}
	if !a.RecommendEarlyStop {
		t.Error("should recommend stop for instance-specific long plateau")
	}
}

func TestLearnedStagnationDetector_ExpectedRemainingValue(t *testing.T) {
	d := NewLearnedStagnationDetector(testCurveModel(), DefaultStagnationPolicyConfig())

	// Low plateau: P(improve) is high → expected remaining value should be positive.
	fv := FeatureVector{
		Problem:         "cvrp",
		Algorithm:       "sa",
		BudgetConsumed:  0.2,
		PlateauLength:   1000,
		IterationBudget: 100000,
		AcceptanceRate:  0.02,
	}
	a := d.Assess(fv)

	if a.ExpectedRemainingValue <= 0 {
		t.Errorf("ExpectedRemainingValue = %f, should be positive for low plateau", a.ExpectedRemainingValue)
	}
}
