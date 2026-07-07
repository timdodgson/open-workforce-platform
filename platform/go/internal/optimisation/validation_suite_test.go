package optimisation

import (
	"testing"
)

func TestDefaultValidationConfig(t *testing.T) {
	c := DefaultValidationConfig()
	if c.TotalExperiments() != 240 {
		t.Errorf("TotalExperiments = %d, want 240 (8 configs × 3 modes × 10 seeds)", c.TotalExperiments())
	}
}

func TestComputeStatistics(t *testing.T) {
	results := []ExperimentResult{
		{Domain: "cvrp", Instance: "A-n32-k5", Algorithm: "sa", PolicyMode: "rules", Objective: 784, RuntimeMs: 78, CandidatesEval: 500000},
		{Domain: "cvrp", Instance: "A-n32-k5", Algorithm: "sa", PolicyMode: "rules", Objective: 796, RuntimeMs: 80, CandidatesEval: 500000},
		{Domain: "cvrp", Instance: "A-n32-k5", Algorithm: "sa", PolicyMode: "rules", Objective: 784, RuntimeMs: 75, CandidatesEval: 500000},
		{Domain: "cvrp", Instance: "A-n32-k5", Algorithm: "sa", PolicyMode: "rules", Objective: 829, RuntimeMs: 82, CandidatesEval: 500000},
		{Domain: "cvrp", Instance: "A-n32-k5", Algorithm: "sa", PolicyMode: "rules", Objective: 784, RuntimeMs: 77, CandidatesEval: 500000},
	}

	stats := ComputeStatistics(results)

	if stats.N != 5 {
		t.Errorf("N = %d, want 5", stats.N)
	}
	// Mean = (784+796+784+829+784)/5 = 795.4
	if stats.Mean < 795 || stats.Mean > 796 {
		t.Errorf("Mean = %f, want ~795.4", stats.Mean)
	}
	if stats.Median != 784 {
		t.Errorf("Median = %f, want 784", stats.Median)
	}
	if stats.Min != 784 {
		t.Errorf("Min = %d, want 784", stats.Min)
	}
	if stats.Max != 829 {
		t.Errorf("Max = %d, want 829", stats.Max)
	}
	if stats.StdDev <= 0 {
		t.Error("StdDev should be > 0")
	}
	if stats.CI95Lower >= stats.Mean || stats.CI95Upper <= stats.Mean {
		t.Errorf("CI95 [%f, %f] should surround mean %f", stats.CI95Lower, stats.CI95Upper, stats.Mean)
	}
}

func TestComputeStatistics_Empty(t *testing.T) {
	stats := ComputeStatistics(nil)
	if stats.N != 0 {
		t.Errorf("N = %d, want 0 for empty", stats.N)
	}
}

func TestCompareGroups_Equivalent(t *testing.T) {
	a := []ExperimentResult{
		{Domain: "cvrp", Objective: 784, Seed: 42},
		{Domain: "cvrp", Objective: 784, Seed: 123},
		{Domain: "cvrp", Objective: 784, Seed: 555},
	}
	b := []ExperimentResult{
		{Domain: "cvrp", Objective: 784, Seed: 42},
		{Domain: "cvrp", Objective: 784, Seed: 123},
		{Domain: "cvrp", Objective: 784, Seed: 555},
	}

	comp := CompareGroups(a, b, "rules", "hybrid")

	if comp.Verdict != "equivalent" {
		t.Errorf("Verdict = %q, want equivalent", comp.Verdict)
	}
	if comp.Ties != 3 {
		t.Errorf("Ties = %d, want 3", comp.Ties)
	}
}

func TestCompareGroups_Better(t *testing.T) {
	a := []ExperimentResult{
		{Domain: "vrptw", Objective: 900, Seed: 42},
		{Domain: "vrptw", Objective: 910, Seed: 123},
		{Domain: "vrptw", Objective: 920, Seed: 555},
		{Domain: "vrptw", Objective: 890, Seed: 777},
		{Domain: "vrptw", Objective: 905, Seed: 999},
	}
	b := []ExperimentResult{
		{Domain: "vrptw", Objective: 830, Seed: 42},
		{Domain: "vrptw", Objective: 835, Seed: 123},
		{Domain: "vrptw", Objective: 828, Seed: 555},
		{Domain: "vrptw", Objective: 840, Seed: 777},
		{Domain: "vrptw", Objective: 832, Seed: 999},
	}

	comp := CompareGroups(a, b, "rules", "learned")

	if comp.Verdict != "better" {
		t.Errorf("Verdict = %q, want better (B has lower objectives)", comp.Verdict)
	}
	if comp.Losses != 5 {
		t.Errorf("Losses = %d, want 5 (A loses all)", comp.Losses)
	}
	if !comp.Significant {
		t.Error("should be significant")
	}
}

func TestCompareGroups_Empty(t *testing.T) {
	comp := CompareGroups(nil, nil, "a", "b")
	if comp.Verdict != "not_evaluated" {
		t.Errorf("Verdict = %q, want not_evaluated", comp.Verdict)
	}
}
