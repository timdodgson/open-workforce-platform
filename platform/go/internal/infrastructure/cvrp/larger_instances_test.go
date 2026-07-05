package cvrp

import (
	"testing"

	"github.com/timdodgson/open-workforce-platform/platform/go/internal/optimisation"
)

// TestLargerInstances_AlgorithmComparison runs all algorithms on progressively harder instances.
func TestLargerInstances_AlgorithmComparison(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping larger instance comparison in short mode")
	}

	instances := []struct {
		path    string
		name    string
		optimal int
	}{
		{"../../../../../examples/cvrp/A-n32-k5.vrp", "A-n32-k5", 784},
		{"../../../../../examples/cvrp/A-n45-k6.vrp", "A-n45-k6", 944},
		{"../../../../../examples/cvrp/A-n60-k9.vrp", "A-n60-k9", 1354},
		{"../../../../../examples/cvrp/A-n80-k10.vrp", "A-n80-k10", 1763},
	}

	modes := []struct {
		name   string
		config optimisation.SearchConfig
	}{
		{"SA", optimisation.SearchConfig{
			Mode: "sa", Iterations: 500000,
			InitialTemperature: 200, MinTemperature: 0.0001, CoolingMode: "adaptive", Seed: 42,
		}},
		{"LAHC", optimisation.SearchConfig{
			Mode: "lahc", Iterations: 500000, LateAcceptanceLength: 2000, Seed: 42,
		}},
		{"Tabu", optimisation.SearchConfig{
			Mode: "tabu", Iterations: 10000, TabuTenure: 10, TabuNeighbourhood: 50, Seed: 42,
		}},
	}

	t.Log("")
	t.Log("┌────────────┬──────────┬──────┬──────┬──────┬─────────┬────────┐")
	t.Log("│ Instance   │ Optimal  │  SA  │ LAHC │ Tabu │ Best    │ Gap%   │")
	t.Log("├────────────┼──────────┼──────┼──────┼──────┼─────────┼────────┤")

	for _, inst := range instances {
		ds, err := LoadDataset(inst.path)
		if err != nil {
			t.Errorf("%s: %v", inst.name, err)
			continue
		}
		problem := NewCVRPProblem(ds)
		baselineSol, _ := problem.CreateInitialSolution()
		baseline := problem.Evaluate(baselineSol)

		results := make(map[string]int)
		for _, mode := range modes {
			result := optimisation.RunSearch(problem, mode.config)
			results[mode.name] = result.BestPenalty
		}

		best := baseline
		bestAlgo := "NN"
		for algo, val := range results {
			if val < best {
				best = val
				bestAlgo = algo
			}
		}

		gap := float64(best-inst.optimal) / float64(inst.optimal) * 100

		t.Logf("│ %-10s │ %6d   │ %4d │ %4d │ %4d │ %4d %s │ %+5.1f%% │",
			inst.name, inst.optimal,
			results["SA"], results["LAHC"], results["Tabu"],
			best, bestAlgo, gap)

		_ = baseline
	}

	t.Log("└────────────┴──────────┴──────┴──────┴──────┴─────────┴────────┘")
}
