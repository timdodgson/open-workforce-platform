package cvrp

import (
	"testing"

	"github.com/timdodgson/open-workforce-platform/platform/go/internal/optimisation"
)

// TestSA_EndToEnd_CVRP proves SA can optimise a CVRP instance through the generic interface.
func TestSA_EndToEnd_CVRP(t *testing.T) {
	ds, err := LoadDataset("testdata/A-n10-k2.vrp")
	if err != nil {
		t.Fatalf("LoadDataset: %v", err)
	}

	problem := NewCVRPProblem(ds)

	// Get constructive baseline.
	baselineSol, _ := problem.CreateInitialSolution()
	baselineCost := problem.Evaluate(baselineSol)

	// Run SA via generic search engine.
	config := optimisation.SearchConfig{
		Mode:               "sa",
		Iterations:         100000,
		InitialTemperature: 100.0,
		MinTemperature:     0.0001,
		CoolingMode:        "adaptive",
		Seed:               42,
	}

	result := optimisation.RunSearch(problem, config)

	// Verify improvement.
	if result.BestPenalty >= baselineCost {
		t.Errorf("SA did not improve over baseline: best=%d, baseline=%d", result.BestPenalty, baselineCost)
	}

	// Verify feasibility: solution should remain feasible (Evaluate = pure distance for feasible).
	bestSol := result.BestSolution.(*cvrpSolution)
	feasible, violations := problem.ValidateFull(bestSol)
	if !feasible {
		t.Errorf("Final solution is infeasible: %+v", violations)
	}

	// Verify all customers served.
	assertAllCustomersVisitedOnce(t, problem, bestSol)

	improvement := float64(baselineCost-result.BestPenalty) / float64(baselineCost) * 100
	t.Logf("SA CVRP: baseline=%d, best=%d, improvement=%.1f%%, candidates=%d, runtime=%dms",
		baselineCost, result.BestPenalty, improvement, result.Candidates, result.DurationMs)
}

// TestSA_Deterministic_CVRP proves same seed produces same result.
func TestSA_Deterministic_CVRP(t *testing.T) {
	ds, _ := LoadDataset("testdata/A-n10-k2.vrp")
	problem := NewCVRPProblem(ds)

	config := optimisation.SearchConfig{
		Mode:               "sa",
		Iterations:         50000,
		InitialTemperature: 100.0,
		MinTemperature:     0.0001,
		CoolingMode:        "adaptive",
		Seed:               99,
	}

	r1 := optimisation.RunSearch(problem, config)
	r2 := optimisation.RunSearch(problem, config)

	if r1.BestPenalty != r2.BestPenalty {
		t.Errorf("Not deterministic: run1=%d, run2=%d", r1.BestPenalty, r2.BestPenalty)
	}
}

// TestSA_CVRP_FeasibilityMaintained verifies SA never produces infeasible best solution.
func TestSA_CVRP_FeasibilityMaintained(t *testing.T) {
	ds, _ := LoadDataset("testdata/A-n10-k2.vrp")
	problem := NewCVRPProblem(ds)

	// Run with multiple seeds to stress-test feasibility.
	seeds := []int64{1, 42, 99, 123, 777}
	for _, seed := range seeds {
		config := optimisation.SearchConfig{
			Mode:               "sa",
			Iterations:         50000,
			InitialTemperature: 100.0,
			MinTemperature:     0.0001,
			CoolingMode:        "adaptive",
			Seed:               seed,
		}

		result := optimisation.RunSearch(problem, config)
		bestSol := result.BestSolution.(*cvrpSolution)

		feasible, violations := problem.ValidateFull(bestSol)
		if !feasible {
			t.Errorf("Seed %d produced infeasible solution: %+v", seed, violations)
		}
		assertAllCustomersVisitedOnce(t, problem, bestSol)
	}
}

// --- LAHC Tests ---

// TestLAHC_EndToEnd_CVRP proves LAHC can optimise a CVRP instance.
func TestLAHC_EndToEnd_CVRP(t *testing.T) {
	ds, err := LoadDataset("testdata/A-n10-k2.vrp")
	if err != nil {
		t.Fatalf("LoadDataset: %v", err)
	}

	problem := NewCVRPProblem(ds)
	baselineSol, _ := problem.CreateInitialSolution()
	baselineCost := problem.Evaluate(baselineSol)

	config := optimisation.SearchConfig{
		Mode:                 "lahc",
		Iterations:           100000,
		LateAcceptanceLength: 1000,
		Seed:                 42,
	}

	result := optimisation.RunSearch(problem, config)

	if result.BestPenalty >= baselineCost {
		t.Errorf("LAHC did not improve over baseline: best=%d, baseline=%d", result.BestPenalty, baselineCost)
	}

	bestSol := result.BestSolution.(*cvrpSolution)
	feasible, violations := problem.ValidateFull(bestSol)
	if !feasible {
		t.Errorf("LAHC produced infeasible solution: %+v", violations)
	}
	assertAllCustomersVisitedOnce(t, problem, bestSol)

	improvement := float64(baselineCost-result.BestPenalty) / float64(baselineCost) * 100
	t.Logf("LAHC CVRP: baseline=%d, best=%d, improvement=%.1f%%, candidates=%d, runtime=%dms",
		baselineCost, result.BestPenalty, improvement, result.Candidates, result.DurationMs)
}

// TestLAHC_Deterministic_CVRP proves same seed produces same result.
func TestLAHC_Deterministic_CVRP(t *testing.T) {
	ds, _ := LoadDataset("testdata/A-n10-k2.vrp")
	problem := NewCVRPProblem(ds)

	config := optimisation.SearchConfig{
		Mode:                 "lahc",
		Iterations:           50000,
		LateAcceptanceLength: 500,
		Seed:                 99,
	}

	r1 := optimisation.RunSearch(problem, config)
	r2 := optimisation.RunSearch(problem, config)

	if r1.BestPenalty != r2.BestPenalty {
		t.Errorf("LAHC not deterministic: run1=%d, run2=%d", r1.BestPenalty, r2.BestPenalty)
	}
}

// TestLAHC_CVRP_FeasibilityMaintained verifies LAHC never produces infeasible best solution.
func TestLAHC_CVRP_FeasibilityMaintained(t *testing.T) {
	ds, _ := LoadDataset("testdata/A-n10-k2.vrp")
	problem := NewCVRPProblem(ds)

	seeds := []int64{1, 42, 99, 123, 777}
	for _, seed := range seeds {
		config := optimisation.SearchConfig{
			Mode:                 "lahc",
			Iterations:           50000,
			LateAcceptanceLength: 1000,
			Seed:                 seed,
		}

		result := optimisation.RunSearch(problem, config)
		bestSol := result.BestSolution.(*cvrpSolution)

		feasible, violations := problem.ValidateFull(bestSol)
		if !feasible {
			t.Errorf("LAHC seed %d produced infeasible solution: %+v", seed, violations)
		}
		assertAllCustomersVisitedOnce(t, problem, bestSol)
	}
}

// --- Comparison Test ---

// TestSA_vs_LAHC_BothImprove verifies both algorithms improve over constructive baseline.
func TestSA_vs_LAHC_BothImprove(t *testing.T) {
	ds, _ := LoadDataset("testdata/A-n10-k2.vrp")
	problem := NewCVRPProblem(ds)

	baselineSol, _ := problem.CreateInitialSolution()
	baselineCost := problem.Evaluate(baselineSol)

	saConfig := optimisation.SearchConfig{
		Mode:               "sa",
		Iterations:         100000,
		InitialTemperature: 100.0,
		MinTemperature:     0.0001,
		CoolingMode:        "adaptive",
		Seed:               42,
	}

	lahcConfig := optimisation.SearchConfig{
		Mode:                 "lahc",
		Iterations:           100000,
		LateAcceptanceLength: 1000,
		Seed:                 42,
	}

	saResult := optimisation.RunSearch(problem, saConfig)
	lahcResult := optimisation.RunSearch(problem, lahcConfig)

	if saResult.BestPenalty >= baselineCost {
		t.Errorf("SA did not improve: best=%d, baseline=%d", saResult.BestPenalty, baselineCost)
	}
	if lahcResult.BestPenalty >= baselineCost {
		t.Errorf("LAHC did not improve: best=%d, baseline=%d", lahcResult.BestPenalty, baselineCost)
	}

	t.Logf("Comparison (baseline=%d):", baselineCost)
	t.Logf("  SA:   best=%d (%.1f%% improvement)", saResult.BestPenalty,
		float64(baselineCost-saResult.BestPenalty)/float64(baselineCost)*100)
	t.Logf("  LAHC: best=%d (%.1f%% improvement)", lahcResult.BestPenalty,
		float64(baselineCost-lahcResult.BestPenalty)/float64(baselineCost)*100)
}


// --- Tabu Tests ---

// TestTabu_EndToEnd_CVRP proves Tabu can optimise a CVRP instance.
func TestTabu_EndToEnd_CVRP(t *testing.T) {
	ds, err := LoadDataset("testdata/A-n10-k2.vrp")
	if err != nil {
		t.Fatalf("LoadDataset: %v", err)
	}

	problem := NewCVRPProblem(ds)
	baselineSol, _ := problem.CreateInitialSolution()
	baselineCost := problem.Evaluate(baselineSol)

	config := optimisation.SearchConfig{
		Mode:       "tabu",
		Iterations: 100000,
		TabuTenure: 7,
		Seed:       42,
	}

	result := optimisation.RunSearch(problem, config)

	if result.BestPenalty >= baselineCost {
		t.Errorf("Tabu did not improve over baseline: best=%d, baseline=%d", result.BestPenalty, baselineCost)
	}

	bestSol := result.BestSolution.(*cvrpSolution)
	feasible, violations := problem.ValidateFull(bestSol)
	if !feasible {
		t.Errorf("Tabu produced infeasible solution: %+v", violations)
	}
	assertAllCustomersVisitedOnce(t, problem, bestSol)

	improvement := float64(baselineCost-result.BestPenalty) / float64(baselineCost) * 100
	t.Logf("Tabu CVRP: baseline=%d, best=%d, improvement=%.1f%%, candidates=%d, runtime=%dms",
		baselineCost, result.BestPenalty, improvement, result.Candidates, result.DurationMs)
}

// TestTabu_Deterministic_CVRP proves same seed produces same result.
func TestTabu_Deterministic_CVRP(t *testing.T) {
	ds, _ := LoadDataset("testdata/A-n10-k2.vrp")
	problem := NewCVRPProblem(ds)

	config := optimisation.SearchConfig{
		Mode:       "tabu",
		Iterations: 50000,
		TabuTenure: 7,
		Seed:       99,
	}

	r1 := optimisation.RunSearch(problem, config)
	r2 := optimisation.RunSearch(problem, config)

	if r1.BestPenalty != r2.BestPenalty {
		t.Errorf("Tabu not deterministic: run1=%d, run2=%d", r1.BestPenalty, r2.BestPenalty)
	}
}

// TestTabu_CVRP_FeasibilityMaintained verifies Tabu never produces infeasible best solution.
func TestTabu_CVRP_FeasibilityMaintained(t *testing.T) {
	ds, _ := LoadDataset("testdata/A-n10-k2.vrp")
	problem := NewCVRPProblem(ds)

	seeds := []int64{1, 42, 99, 123, 777}
	for _, seed := range seeds {
		config := optimisation.SearchConfig{
			Mode:       "tabu",
			Iterations: 50000,
			TabuTenure: 7,
			Seed:       seed,
		}

		result := optimisation.RunSearch(problem, config)
		bestSol := result.BestSolution.(*cvrpSolution)

		feasible, violations := problem.ValidateFull(bestSol)
		if !feasible {
			t.Errorf("Tabu seed %d produced infeasible solution: %+v", seed, violations)
		}
		assertAllCustomersVisitedOnce(t, problem, bestSol)
	}
}

// --- All Three Algorithms Test ---

// TestAllModes_RunThroughSameInterface verifies SA, LAHC, and Tabu all work via RunSearch.
func TestAllModes_RunThroughSameInterface(t *testing.T) {
	ds, _ := LoadDataset("testdata/A-n10-k2.vrp")
	problem := NewCVRPProblem(ds)

	baselineSol, _ := problem.CreateInitialSolution()
	baselineCost := problem.Evaluate(baselineSol)

	modes := []struct {
		name   string
		config optimisation.SearchConfig
	}{
		{"SA", optimisation.SearchConfig{
			Mode: "sa", Iterations: 100000,
			InitialTemperature: 100, MinTemperature: 0.0001, CoolingMode: "adaptive", Seed: 42,
		}},
		{"LAHC", optimisation.SearchConfig{
			Mode: "lahc", Iterations: 100000, LateAcceptanceLength: 1000, Seed: 42,
		}},
		{"Tabu", optimisation.SearchConfig{
			Mode: "tabu", Iterations: 100000, TabuTenure: 7, Seed: 42,
		}},
	}

	for _, m := range modes {
		result := optimisation.RunSearch(problem, m.config)

		if result.BestPenalty >= baselineCost {
			t.Errorf("%s did not improve: best=%d, baseline=%d", m.name, result.BestPenalty, baselineCost)
		}

		bestSol := result.BestSolution.(*cvrpSolution)
		feasible, _ := problem.ValidateFull(bestSol)
		if !feasible {
			t.Errorf("%s produced infeasible solution", m.name)
		}

		improvement := float64(baselineCost-result.BestPenalty) / float64(baselineCost) * 100
		t.Logf("  %s: best=%d (%.1f%% improvement), candidates=%d", m.name, result.BestPenalty, improvement, result.Candidates)
	}
}


// --- Portfolio Tests ---

// TestPortfolio_EndToEnd_CVRP proves portfolio runs all strategies and picks the best.
func TestPortfolio_EndToEnd_CVRP(t *testing.T) {
	ds, _ := LoadDataset("testdata/A-n10-k2.vrp")
	problem := NewCVRPProblem(ds)

	baselineSol, _ := problem.CreateInitialSolution()
	baselineCost := problem.Evaluate(baselineSol)

	config := optimisation.SearchConfig{
		Mode:                 "portfolio",
		Iterations:           100000,
		InitialTemperature:   100.0,
		MinTemperature:       0.0001,
		CoolingMode:          "adaptive",
		LateAcceptanceLength: 1000,
		TabuTenure:           7,
		Portfolio:            []string{"sa", "lahc", "tabu"},
		Seed:                 42,
	}

	pr := optimisation.RunPortfolio(problem, config)

	// Should have 3 entries.
	if len(pr.Entries) != 3 {
		t.Errorf("Expected 3 entries, got %d", len(pr.Entries))
	}

	// Winner should improve over baseline.
	if pr.BestResult.BestPenalty >= baselineCost {
		t.Errorf("Portfolio did not improve: best=%d, baseline=%d", pr.BestResult.BestPenalty, baselineCost)
	}

	// All entries should have run.
	for _, e := range pr.Entries {
		if e.Result.Candidates == 0 {
			t.Errorf("Strategy %s had 0 candidates", e.Mode)
		}
	}

	// Verify feasibility of winner.
	bestSol := pr.BestResult.BestSolution.(*cvrpSolution)
	feasible, violations := problem.ValidateFull(bestSol)
	if !feasible {
		t.Errorf("Portfolio winner infeasible: %+v", violations)
	}

	t.Logf("Portfolio winner: %s (best=%d, baseline=%d, improvement=%.1f%%)",
		pr.Winner, pr.BestResult.BestPenalty, baselineCost,
		float64(baselineCost-pr.BestResult.BestPenalty)/float64(baselineCost)*100)
	for _, e := range pr.Entries {
		t.Logf("  %s: best=%d, candidates=%d, runtime=%dms",
			e.Mode, e.Result.BestPenalty, e.Result.Candidates, e.Result.DurationMs)
	}
}

// TestPortfolio_AtLeastAsBestAsIndividual verifies portfolio >= best individual.
func TestPortfolio_AtLeastAsBestAsIndividual(t *testing.T) {
	ds, _ := LoadDataset("testdata/A-n10-k2.vrp")
	problem := NewCVRPProblem(ds)

	config := optimisation.SearchConfig{
		Mode:                 "portfolio",
		Iterations:           50000,
		InitialTemperature:   100.0,
		MinTemperature:       0.0001,
		CoolingMode:          "adaptive",
		LateAcceptanceLength: 1000,
		TabuTenure:           7,
		Portfolio:            []string{"sa", "lahc", "tabu"},
		Seed:                 42,
	}

	pr := optimisation.RunPortfolio(problem, config)

	// Portfolio best should equal the best individual entry.
	bestIndividual := int(^uint(0) >> 1)
	for _, e := range pr.Entries {
		if e.Result.BestPenalty < bestIndividual {
			bestIndividual = e.Result.BestPenalty
		}
	}

	if pr.BestResult.BestPenalty != bestIndividual {
		t.Errorf("Portfolio best (%d) != best individual (%d)", pr.BestResult.BestPenalty, bestIndividual)
	}
}

// TestPortfolio_Deterministic proves same seed produces same result.
func TestPortfolio_Deterministic(t *testing.T) {
	ds, _ := LoadDataset("testdata/A-n10-k2.vrp")
	problem := NewCVRPProblem(ds)

	config := optimisation.SearchConfig{
		Mode:                 "portfolio",
		Iterations:           50000,
		InitialTemperature:   100.0,
		MinTemperature:       0.0001,
		CoolingMode:          "adaptive",
		LateAcceptanceLength: 1000,
		TabuTenure:           7,
		Portfolio:            []string{"sa", "lahc", "tabu"},
		Seed:                 99,
	}

	pr1 := optimisation.RunPortfolio(problem, config)
	pr2 := optimisation.RunPortfolio(problem, config)

	if pr1.BestResult.BestPenalty != pr2.BestResult.BestPenalty {
		t.Errorf("Portfolio not deterministic: run1=%d, run2=%d",
			pr1.BestResult.BestPenalty, pr2.BestResult.BestPenalty)
	}
	if pr1.Winner != pr2.Winner {
		t.Errorf("Portfolio winner not deterministic: %s vs %s", pr1.Winner, pr2.Winner)
	}
}

// TestPortfolio_FeasibilityMaintained verifies all seeds produce feasible results.
func TestPortfolio_FeasibilityMaintained(t *testing.T) {
	ds, _ := LoadDataset("testdata/A-n10-k2.vrp")
	problem := NewCVRPProblem(ds)

	seeds := []int64{1, 42, 99, 123, 777}
	for _, seed := range seeds {
		config := optimisation.SearchConfig{
			Mode:                 "portfolio",
			Iterations:           50000,
			InitialTemperature:   100.0,
			MinTemperature:       0.0001,
			CoolingMode:          "adaptive",
			LateAcceptanceLength: 1000,
			TabuTenure:           7,
			Portfolio:            []string{"sa", "lahc", "tabu"},
			Seed:                 seed,
		}

		pr := optimisation.RunPortfolio(problem, config)
		bestSol := pr.BestResult.BestSolution.(*cvrpSolution)

		feasible, violations := problem.ValidateFull(bestSol)
		if !feasible {
			t.Errorf("Seed %d: portfolio winner infeasible: %+v", seed, violations)
		}
	}
}

// TestPortfolio_ViaRunSearch verifies portfolio works through the standard RunSearch interface.
func TestPortfolio_ViaRunSearch(t *testing.T) {
	ds, _ := LoadDataset("testdata/A-n10-k2.vrp")
	problem := NewCVRPProblem(ds)

	baselineSol, _ := problem.CreateInitialSolution()
	baselineCost := problem.Evaluate(baselineSol)

	config := optimisation.SearchConfig{
		Mode:                 "portfolio",
		Iterations:           50000,
		InitialTemperature:   100.0,
		MinTemperature:       0.0001,
		CoolingMode:          "adaptive",
		LateAcceptanceLength: 1000,
		TabuTenure:           7,
		Portfolio:            []string{"sa", "lahc"},
		Seed:                 42,
	}

	// RunSearch should work for portfolio mode too (returns best result).
	result := optimisation.RunSearch(problem, config)

	if result.BestPenalty >= baselineCost {
		t.Errorf("Portfolio via RunSearch didn't improve: best=%d, baseline=%d", result.BestPenalty, baselineCost)
	}
}


// --- Adaptive Hyper-Heuristic Tests ---

// TestAdaptive_EndToEnd_CVRP proves adaptive mode works on CVRP.
func TestAdaptive_EndToEnd_CVRP(t *testing.T) {
	ds, _ := LoadDataset("testdata/A-n10-k2.vrp")
	problem := NewCVRPProblem(ds)

	baselineSol, _ := problem.CreateInitialSolution()
	baselineCost := problem.Evaluate(baselineSol)

	config := optimisation.SearchConfig{
		Mode:                 "adaptive",
		Iterations:           100000,
		InitialTemperature:   100.0,
		MinTemperature:       0.0001,
		CoolingMode:          "adaptive",
		LateAcceptanceLength: 1000,
		TabuTenure:           7,
		Portfolio:            []string{"sa", "lahc", "tabu"},
		AdaptiveWindow:       5000,
		AdaptiveMinShare:     0.1,
		Seed:                 42,
	}

	result := optimisation.RunSearch(problem, config)

	if result.BestPenalty >= baselineCost {
		t.Errorf("Adaptive did not improve: best=%d, baseline=%d", result.BestPenalty, baselineCost)
	}

	bestSol := result.BestSolution.(*cvrpSolution)
	feasible, violations := problem.ValidateFull(bestSol)
	if !feasible {
		t.Errorf("Adaptive produced infeasible solution: %+v", violations)
	}

	improvement := float64(baselineCost-result.BestPenalty) / float64(baselineCost) * 100
	t.Logf("Adaptive CVRP: baseline=%d, best=%d, improvement=%.1f%%, candidates=%d, runtime=%dms",
		baselineCost, result.BestPenalty, improvement, result.Candidates, result.DurationMs)
}

// TestAdaptive_vs_Others_A_n32_k5 compares adaptive against individual strategies on larger instance.
func TestAdaptive_vs_Others_A_n32_k5(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping large instance adaptive test in short mode")
	}

	ds, err := LoadDataset("../../../../../examples/cvrp/A-n32-k5.vrp")
	if err != nil {
		t.Skipf("Instance not found: %v", err)
	}

	problem := NewCVRPProblem(ds)
	baselineSol, _ := problem.CreateInitialSolution()
	baselineCost := problem.Evaluate(baselineSol)

	iterations := 200000

	configs := []struct {
		name   string
		config optimisation.SearchConfig
	}{
		{"SA", optimisation.SearchConfig{Mode: "sa", Iterations: iterations, InitialTemperature: 200, MinTemperature: 0.0001, CoolingMode: "adaptive", Seed: 42}},
		{"LAHC", optimisation.SearchConfig{Mode: "lahc", Iterations: iterations, LateAcceptanceLength: 2000, Seed: 42}},
		{"Adaptive", optimisation.SearchConfig{
			Mode: "adaptive", Iterations: iterations,
			InitialTemperature: 200, MinTemperature: 0.0001, CoolingMode: "adaptive",
			LateAcceptanceLength: 2000,
			Portfolio: []string{"sa", "lahc"}, AdaptiveWindow: 10000, AdaptiveMinShare: 0.1, Seed: 42,
		}},
	}

	t.Logf("A-n32-k5 (baseline=%d, optimal=784):", baselineCost)
	for _, c := range configs {
		result := optimisation.RunSearch(problem, c.config)
		gap := float64(result.BestPenalty-784) / float64(784) * 100
		t.Logf("  %-10s: best=%d, gap=+%.1f%%, runtime=%dms, improved=%d",
			c.name, result.BestPenalty, gap, result.DurationMs, result.Improved)
	}
}
