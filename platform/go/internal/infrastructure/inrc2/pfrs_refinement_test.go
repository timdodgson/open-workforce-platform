package inrc2_test

import (
	"testing"

	"github.com/timdodgson/open-workforce-platform/platform/go/internal/infrastructure/inrc2"
)

// TestRefinement_HillClimb_NeverIntroducesHardViolations runs the full pipeline:
// build feasible roster → PFRS optimise → refine → validate hard=0.
func TestRefinement_HillClimb_NeverIntroducesHardViolations(t *testing.T) {
	sc, _ := inrc2.LoadScenario(testDataDir + "Sc-n005w4.json")
	wd, _ := inrc2.LoadWeekData(testDataDir + "WD-n005w4-0.json")
	hist, _ := inrc2.LoadHistory(testDataDir + "H0-n005w4-0.json")

	// Run PFRS to get a solution.
	config := inrc2.DefaultPFRSConfig()
	config.IterationsPerWorker = 10000
	config.MaxTotalWorkers = 2
	config.Seed = 42

	sol, _, scoreResult, err := inrc2.SolveWeekPFRS(sc, wd, hist, config)
	if err != nil {
		t.Fatalf("SolveWeekPFRS failed: %v", err)
	}
	if scoreResult.HardViolations != 0 {
		t.Fatalf("PFRS produced hard violations before refinement: %d", scoreResult.HardViolations)
	}

	// Build a minimal BeamPath for refinement input.
	weekFiles := []string{testDataDir + "WD-n005w4-0.json"}
	winningPath := []inrc2.BeamPath{
		{Week: 1, Solution: sol, ScoreResult: scoreResult, WeekPenalty: scoreResult.SoftPenalty, CumulativePenalty: scoreResult.SoftPenalty},
	}

	// Run hill-climb refinement.
	refined, _ := inrc2.Refine(sc, weekFiles, winningPath, inrc2.RefinementConfig{
		Mode:       "hillclimb",
		Iterations: 50000,
		Seed:       42,
	}, hist)

	// Verify no hard violations after refinement.
	for _, wp := range refined {
		result := inrc2.Score(sc, wd, hist, wp.Solution)
		if result.HardViolations != 0 {
			t.Errorf("Week %d: refinement introduced %d hard violations", wp.Week, result.HardViolations)
		}
	}

	// Refinement with official validation gate should never make official score worse.
	// (It either improves or reverts to original.)
	originalScore := inrc2.Score(sc, wd, hist, sol)
	refinedScore := inrc2.Score(sc, wd, hist, refined[0].Solution)
	if refinedScore.SoftPenalty > originalScore.SoftPenalty {
		t.Errorf("refinement made official score worse: %d → %d", originalScore.SoftPenalty, refinedScore.SoftPenalty)
	}
}

// TestRefinement_SA_NeverIntroducesHardViolations tests SA refinement.
func TestRefinement_SA_NeverIntroducesHardViolations(t *testing.T) {
	sc, _ := inrc2.LoadScenario(testDataDir + "Sc-n005w4.json")
	wd, _ := inrc2.LoadWeekData(testDataDir + "WD-n005w4-0.json")
	hist, _ := inrc2.LoadHistory(testDataDir + "H0-n005w4-0.json")

	config := inrc2.DefaultPFRSConfig()
	config.IterationsPerWorker = 10000
	config.MaxTotalWorkers = 2
	config.Seed = 42

	sol, _, scoreResult, err := inrc2.SolveWeekPFRS(sc, wd, hist, config)
	if err != nil {
		t.Fatalf("SolveWeekPFRS failed: %v", err)
	}

	weekFiles := []string{testDataDir + "WD-n005w4-0.json"}
	winningPath := []inrc2.BeamPath{
		{Week: 1, Solution: sol, ScoreResult: scoreResult, WeekPenalty: scoreResult.SoftPenalty, CumulativePenalty: scoreResult.SoftPenalty},
	}

	refined, _ := inrc2.Refine(sc, weekFiles, winningPath, inrc2.RefinementConfig{
		Mode:               "sa",
		Iterations:         50000,
		Seed:               42,
		InitialTemperature: 10.0,
	}, hist)

	for _, wp := range refined {
		result := inrc2.Score(sc, wd, hist, wp.Solution)
		if result.HardViolations != 0 {
			t.Errorf("Week %d: SA refinement introduced %d hard violations", wp.Week, result.HardViolations)
		}
	}
}

// TestRefinement_LAHC_NeverIntroducesHardViolations tests LAHC refinement.
func TestRefinement_LAHC_NeverIntroducesHardViolations(t *testing.T) {
	sc, _ := inrc2.LoadScenario(testDataDir + "Sc-n005w4.json")
	wd, _ := inrc2.LoadWeekData(testDataDir + "WD-n005w4-0.json")
	hist, _ := inrc2.LoadHistory(testDataDir + "H0-n005w4-0.json")

	config := inrc2.DefaultPFRSConfig()
	config.IterationsPerWorker = 10000
	config.MaxTotalWorkers = 2
	config.Seed = 42

	sol, _, scoreResult, err := inrc2.SolveWeekPFRS(sc, wd, hist, config)
	if err != nil {
		t.Fatalf("SolveWeekPFRS failed: %v", err)
	}

	weekFiles := []string{testDataDir + "WD-n005w4-0.json"}
	winningPath := []inrc2.BeamPath{
		{Week: 1, Solution: sol, ScoreResult: scoreResult, WeekPenalty: scoreResult.SoftPenalty, CumulativePenalty: scoreResult.SoftPenalty},
	}

	refined, _ := inrc2.Refine(sc, weekFiles, winningPath, inrc2.RefinementConfig{
		Mode:       "lahc",
		Iterations: 50000,
		Seed:       42,
	}, hist)

	for _, wp := range refined {
		result := inrc2.Score(sc, wd, hist, wp.Solution)
		if result.HardViolations != 0 {
			t.Errorf("Week %d: LAHC refinement introduced %d hard violations", wp.Week, result.HardViolations)
		}
	}
}

// TestRefinement_MultiWeek_HistoryPropagation tests that history is correctly
// passed between weeks during refinement — forbidden succession across week boundary.
func TestRefinement_MultiWeek_HistoryPropagation(t *testing.T) {
	sc, _ := inrc2.LoadScenario(testDataDir + "Sc-n005w4.json")
	hist, _ := inrc2.LoadHistory(testDataDir + "H0-n005w4-0.json")

	// Use multiple week files.
	weekFiles := []string{
		testDataDir + "WD-n005w4-0.json",
		testDataDir + "WD-n005w4-1.json",
	}

	// Run PFRS for both weeks.
	config := inrc2.DefaultPFRSConfig()
	config.IterationsPerWorker = 5000
	config.MaxTotalWorkers = 1
	config.Seed = 42

	wd0, _ := inrc2.LoadWeekData(weekFiles[0])
	sol0, _, sr0, _ := inrc2.SolveWeekPFRS(sc, wd0, hist, config)
	hist1 := inrc2.UpdateHistory(sc, hist, sol0)

	wd1, _ := inrc2.LoadWeekData(weekFiles[1])
	sol1, _, sr1, _ := inrc2.SolveWeekPFRS(sc, wd1, hist1, config)

	winningPath := []inrc2.BeamPath{
		{Week: 1, Solution: sol0, ScoreResult: sr0, WeekPenalty: sr0.SoftPenalty, CumulativePenalty: sr0.SoftPenalty},
		{Week: 2, Solution: sol1, ScoreResult: sr1, WeekPenalty: sr1.SoftPenalty, CumulativePenalty: sr0.SoftPenalty + sr1.SoftPenalty},
	}

	// Refine both weeks.
	refined, _ := inrc2.Refine(sc, weekFiles, winningPath, inrc2.RefinementConfig{
		Mode:               "sa",
		Iterations:         20000,
		Seed:               42,
		InitialTemperature: 10.0,
	}, hist)

	// Validate both weeks with proper rolling history.
	valHist := hist
	for i, wp := range refined {
		wd, _ := inrc2.LoadWeekData(weekFiles[i])
		result := inrc2.Score(sc, wd, valHist, wp.Solution)
		if result.HardViolations != 0 {
			t.Errorf("Week %d: multi-week refinement introduced %d hard violations (history propagation bug)", wp.Week, result.HardViolations)
		}
		valHist = inrc2.UpdateHistory(sc, valHist, wp.Solution)
	}
}

// TestRefinement_None_NoChange verifies that mode "none" returns input unchanged.
func TestRefinement_None_NoChange(t *testing.T) {
	sc, _ := inrc2.LoadScenario(testDataDir + "Sc-n005w4.json")
	wd, _ := inrc2.LoadWeekData(testDataDir + "WD-n005w4-0.json")
	hist, _ := inrc2.LoadHistory(testDataDir + "H0-n005w4-0.json")

	config := inrc2.DefaultPFRSConfig()
	config.IterationsPerWorker = 5000
	config.MaxTotalWorkers = 1
	config.Seed = 42

	sol, _, sr, _ := inrc2.SolveWeekPFRS(sc, wd, hist, config)

	weekFiles := []string{testDataDir + "WD-n005w4-0.json"}
	winningPath := []inrc2.BeamPath{
		{Week: 1, Solution: sol, ScoreResult: sr, WeekPenalty: sr.SoftPenalty, CumulativePenalty: sr.SoftPenalty},
	}

	refined, summary := inrc2.Refine(sc, weekFiles, winningPath, inrc2.RefinementConfig{
		Mode: "none",
	}, hist)

	if len(refined) != len(winningPath) {
		t.Errorf("expected same length, got %d vs %d", len(refined), len(winningPath))
	}
	if summary.TotalImprovement != 0 {
		t.Errorf("expected 0 improvement with mode=none, got %d", summary.TotalImprovement)
	}
	// Solution should be identical.
	if len(refined[0].Solution.Assignments) != len(sol.Assignments) {
		t.Error("solution was modified despite mode=none")
	}
}
