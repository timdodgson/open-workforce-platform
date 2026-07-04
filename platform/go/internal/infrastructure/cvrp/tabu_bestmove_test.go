package cvrp

import (
	"testing"

	"github.com/timdodgson/open-workforce-platform/platform/go/internal/optimisation"
)

// TestTabu_BestMove_A_n32_k5 tests proper Tabu with best-move on a larger instance.
func TestTabu_BestMove_A_n32_k5(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping large instance test in short mode")
	}

	ds, err := LoadDataset("../../../../../examples/cvrp/A-n32-k5.vrp")
	if err != nil {
		t.Fatalf("LoadDataset: %v", err)
	}

	problem := NewCVRPProblem(ds)
	baselineSol, _ := problem.CreateInitialSolution()
	baselineCost := problem.Evaluate(baselineSol)

	// SA for comparison.
	saConfig := optimisation.SearchConfig{
		Mode:               "sa",
		Iterations:         100000,
		InitialTemperature: 200.0,
		MinTemperature:     0.0001,
		CoolingMode:        "adaptive",
		Seed:               42,
	}
	saResult := optimisation.RunSearch(problem, saConfig)

	// Tabu with best-move neighbourhood (100 samples per iteration).
	// Use fewer iterations since each is 100× more expensive.
	tabuConfig := optimisation.SearchConfig{
		Mode:              "tabu",
		Iterations:        5000, // 5K iterations × 100 neighbourhood = 500K evaluations
		TabuTenure:        7,
		TabuNeighbourhood: 100,
		Seed:              42,
	}
	tabuResult := optimisation.RunSearch(problem, tabuConfig)

	// Verify feasibility.
	tabuSol := tabuResult.BestSolution.(*cvrpSolution)
	feasible, violations := problem.ValidateFull(tabuSol)
	if !feasible {
		t.Errorf("Tabu produced infeasible solution: %+v", violations)
	}

	t.Logf("A-n32-k5 (31 customers, optimal=784):")
	t.Logf("  Baseline (NN):     %d", baselineCost)
	t.Logf("  SA (100K iter):    %d (improvement: %.1f%%)", saResult.BestPenalty,
		float64(baselineCost-saResult.BestPenalty)/float64(baselineCost)*100)
	t.Logf("  Tabu (5K×100):     %d (improvement: %.1f%%)", tabuResult.BestPenalty,
		float64(baselineCost-tabuResult.BestPenalty)/float64(baselineCost)*100)
	t.Logf("  SA runtime:        %dms", saResult.DurationMs)
	t.Logf("  Tabu runtime:      %dms", tabuResult.DurationMs)

	// Tabu should improve significantly from baseline.
	if tabuResult.BestPenalty >= baselineCost {
		t.Errorf("Tabu did not improve: best=%d, baseline=%d", tabuResult.BestPenalty, baselineCost)
	}
}
