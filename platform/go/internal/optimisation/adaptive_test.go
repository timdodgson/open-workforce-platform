package optimisation

import (
	"testing"
)

// TestAdaptive_ImprovesOverInitial verifies adaptive mode finds better solutions.
func TestAdaptive_ImprovesOverInitial(t *testing.T) {
	problem := &mockProblem{}
	config := DefaultSearchConfig()
	config.Mode = "adaptive"
	config.Iterations = 20000
	config.Portfolio = []string{"sa", "lahc", "tabu"}
	config.AdaptiveWindow = 2000
	config.AdaptiveMinShare = 0.1

	result := RunSearch(problem, config)

	if result.BestPenalty >= result.InitialPenalty {
		t.Errorf("Adaptive should improve: best=%d, initial=%d", result.BestPenalty, result.InitialPenalty)
	}
	if result.Candidates == 0 {
		t.Error("No candidates evaluated")
	}
	t.Logf("Adaptive: initial=%d, best=%d, improved=%d, runtime=%dms",
		result.InitialPenalty, result.BestPenalty, result.Improved, result.DurationMs)
}

// TestAdaptive_Deterministic verifies same seed produces same result.
func TestAdaptive_Deterministic(t *testing.T) {
	problem := &mockProblem{}
	config := DefaultSearchConfig()
	config.Mode = "adaptive"
	config.Iterations = 10000
	config.Portfolio = []string{"sa", "lahc"}
	config.AdaptiveWindow = 1000
	config.Seed = 99

	r1 := RunSearch(problem, config)
	r2 := RunSearch(problem, config)

	if r1.BestPenalty != r2.BestPenalty {
		t.Errorf("Not deterministic: run1=%d, run2=%d", r1.BestPenalty, r2.BestPenalty)
	}
}

// TestAdaptive_BetterThanWorstStrategy verifies adaptive does at least as well as the worst individual.
func TestAdaptive_BetterThanWorstStrategy(t *testing.T) {
	problem := &mockProblem{}
	iterations := 10000

	// Run each strategy individually.
	saConfig := DefaultSearchConfig()
	saConfig.Mode = "sa"
	saConfig.Iterations = iterations
	saResult := RunSearch(problem, saConfig)

	lahcConfig := DefaultSearchConfig()
	lahcConfig.Mode = "lahc"
	lahcConfig.Iterations = iterations
	lahcResult := RunSearch(problem, lahcConfig)

	// Run adaptive with same total budget.
	adaptConfig := DefaultSearchConfig()
	adaptConfig.Mode = "adaptive"
	adaptConfig.Iterations = iterations
	adaptConfig.Portfolio = []string{"sa", "lahc"}
	adaptConfig.AdaptiveWindow = 1000
	adaptResult := RunSearch(problem, adaptConfig)

	worst := saResult.BestPenalty
	if lahcResult.BestPenalty > worst {
		worst = lahcResult.BestPenalty
	}

	if adaptResult.BestPenalty > worst {
		t.Errorf("Adaptive (%d) worse than worst individual (%d)", adaptResult.BestPenalty, worst)
	}

	t.Logf("SA=%d, LAHC=%d, Adaptive=%d", saResult.BestPenalty, lahcResult.BestPenalty, adaptResult.BestPenalty)
}
