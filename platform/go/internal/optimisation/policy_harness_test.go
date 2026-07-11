package optimisation

import "testing"

func TestBuildPolicyHarnessReport(t *testing.T) {
	results := []ExperimentResult{
		{Domain: "cvrp", Instance: "A-n32-k5", Algorithm: "sa", PolicyMode: "rules", Objective: 800, RuntimeMs: 100, Feasible: true, Seed: 42},
		{Domain: "cvrp", Instance: "A-n32-k5", Algorithm: "sa", PolicyMode: "hybrid", Objective: 798, RuntimeMs: 60, Feasible: true, Seed: 42},
		{Domain: "cvrp", Instance: "A-n32-k5", Algorithm: "sa", PolicyMode: "rules", Objective: 810, RuntimeMs: 110, Feasible: true, Seed: 123},
		{Domain: "cvrp", Instance: "A-n32-k5", Algorithm: "sa", PolicyMode: "hybrid", Objective: 808, RuntimeMs: 65, Feasible: true, Seed: 123},
	}
	report := BuildPolicyHarnessReport(results, "val-", 4)
	if report.TotalRuns != 4 {
		t.Fatalf("total runs: got %d", report.TotalRuns)
	}
	if !report.Gates.Step1HarnessOK {
		t.Fatal("expected harness ok")
	}
	if len(report.Comparisons) == 0 {
		t.Fatal("expected comparisons")
	}
}
