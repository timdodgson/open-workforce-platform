package optimisation

import (
	"testing"
)

func TestFeatureExtractor_FromWorkerContext(t *testing.T) {
	fe := NewFeatureExtractor()
	ctx := WorkerContext{
		Algorithm:        "sa",
		Week:             3,
		Depth:            2,
		ParentObjective:  5000,
		GlobalBest:       4500,
		DistanceFromBest: 500,
		BeamRank:         1,
		Entropy:          0.85,
		BeamHealth:       72.0,
		RecentImprovRate: 2.5,
		AllocatedIters:   500000,
		WorkerCount:      8,
	}

	fv := fe.FromWorkerContext(ctx, "run-001", "n012w8")

	if fv.SchemaVersion != FeatureSchemaVersion {
		t.Errorf("SchemaVersion = %q, want %q", fv.SchemaVersion, FeatureSchemaVersion)
	}
	if fv.Problem != "nrp" {
		t.Errorf("Problem = %q, want nrp", fv.Problem)
	}
	if fv.Instance != "n012w8" {
		t.Errorf("Instance = %q, want n012w8", fv.Instance)
	}
	if fv.Algorithm != "sa" {
		t.Errorf("Algorithm = %q, want sa", fv.Algorithm)
	}
	if fv.DecisionType != "worker" {
		t.Errorf("DecisionType = %q, want worker", fv.DecisionType)
	}
	if fv.Week != 3 {
		t.Errorf("Week = %d, want 3", fv.Week)
	}
	if fv.BranchDepth != 2 {
		t.Errorf("BranchDepth = %d, want 2", fv.BranchDepth)
	}
	if fv.DistanceFromBest != 500 {
		t.Errorf("DistanceFromBest = %d, want 500", fv.DistanceFromBest)
	}
	if fv.Entropy != 0.85 {
		t.Errorf("Entropy = %f, want 0.85", fv.Entropy)
	}
	if fv.WorkerCount != 8 {
		t.Errorf("WorkerCount = %d, want 8", fv.WorkerCount)
	}
}

func TestFeatureExtractor_FromSearchProgress(t *testing.T) {
	fe := NewFeatureExtractor()
	p := SearchProgress{
		Algorithm:          "tabu",
		IterationsComplete: 50000,
		IterationsTotal:    100000,
		CurrentPenalty:     700,
		BestPenalty:        666,
		InitialPenalty:     1200,
		ImprovementRate:    0.3,
		Temperature:        0,
		PlateauLength:      12000,
		Accepted:           800,
		Rejected:           49200,
		CandidatesEval:     50000,
	}

	fv := fe.FromSearchProgress(p, "run-002", "jss", "la01", 4500)

	if fv.BudgetConsumed != 0.5 {
		t.Errorf("BudgetConsumed = %f, want 0.5", fv.BudgetConsumed)
	}
	if fv.PlateauLength != 12000 {
		t.Errorf("PlateauLength = %d, want 12000", fv.PlateauLength)
	}
	if fv.DecisionType != "search" {
		t.Errorf("DecisionType = %q, want search", fv.DecisionType)
	}
	if fv.Problem != "jss" {
		t.Errorf("Problem = %q, want jss", fv.Problem)
	}
	if fv.AcceptanceRate < 0.015 || fv.AcceptanceRate > 0.017 {
		t.Errorf("AcceptanceRate = %f, want ~0.016", fv.AcceptanceRate)
	}
}

func TestFeatureExtractor_FromPortfolioContext(t *testing.T) {
	fe := NewFeatureExtractor()
	ctx := PortfolioContext{
		Strategies:  []string{"sa", "lahc", "tabu"},
		TotalBudget: 300000,
		ProblemType: "cvrp",
		Instance:    "A-n32-k5",
		PreviousResults: []PortfolioHistoryEntry{
			{Strategy: "sa", Improved: true},
			{Strategy: "sa", Improved: true},
			{Strategy: "sa", Improved: false},
			{Strategy: "lahc", Improved: true},
		},
	}

	fv := fe.FromPortfolioContext(ctx, "sa", "run-003")

	if fv.Problem != "cvrp" {
		t.Errorf("Problem = %q, want cvrp", fv.Problem)
	}
	if fv.DecisionType != "portfolio" {
		t.Errorf("DecisionType = %q, want portfolio", fv.DecisionType)
	}
	if fv.WorkerCount != 3 {
		t.Errorf("WorkerCount = %d, want 3 (strategy count)", fv.WorkerCount)
	}
	if fv.GapToReference < 0.66 || fv.GapToReference > 0.67 {
		t.Errorf("GapToReference (win rate) = %f, want ~0.667", fv.GapToReference)
	}
}
