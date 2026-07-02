package inrc2_test

import (
	"testing"

	"github.com/timdodgson/open-workforce-platform/platform/go/internal/infrastructure/inrc2"
)

// TestPFRS_RaceCondition_HighParallelism runs PFRS with high concurrency and
// unlimited branching to detect data races. Run with: go test -race -run TestPFRS_RaceCondition
// This reproduces the conditions that cause the nil pointer panic in successionRejectReason.
func TestPFRS_RaceCondition_HighParallelism(t *testing.T) {
	sc, _ := inrc2.LoadScenario(testDataDir + "Sc-n005w4.json")
	wd, _ := inrc2.LoadWeekData(testDataDir + "WD-n005w4-0.json")
	hist, _ := inrc2.LoadHistory(testDataDir + "H0-n005w4-0.json")

	// High parallelism config that triggers lots of branching.
	config := inrc2.DefaultPFRSConfig()
	config.IterationsPerWorker = 50000
	config.MaxConcurrentWorkers = 8
	config.MaxTotalWorkers = 0 // unlimited
	config.BranchOnGlobalBest = true
	config.Seed = 42
	config.InitialTemperature = 100.0

	// Run multiple times to increase chance of triggering race.
	for run := 0; run < 5; run++ {
		config.Seed = int64(42 + run)
		sol, stats, err := inrc2.RunPFRS(sc, wd, hist, config)
		if err != nil {
			t.Fatalf("run %d: RunPFRS failed: %v", run, err)
		}
		if stats.FinalPenalty <= 0 {
			t.Errorf("run %d: invalid final penalty: %d", run, stats.FinalPenalty)
		}
		if len(sol.Assignments) == 0 {
			t.Errorf("run %d: empty solution", run)
		}
		// Verify the solution is hard-valid.
		result := inrc2.Score(sc, wd, hist, sol)
		if result.HardViolations != 0 {
			t.Errorf("run %d: hard violations: %d", run, result.HardViolations)
		}
	}
}

// TestPFRS_RaceCondition_BeamSearch runs a full beam search with high parallelism
// to detect races in the multi-week orchestration layer.
func TestPFRS_RaceCondition_BeamSearch(t *testing.T) {
	sc, _ := inrc2.LoadScenario(testDataDir + "Sc-n005w4.json")
	hist, _ := inrc2.LoadHistory(testDataDir + "H0-n005w4-0.json")

	weekFiles := []string{
		testDataDir + "WD-n005w4-0.json",
		testDataDir + "WD-n005w4-1.json",
		testDataDir + "WD-n005w4-2.json",
		testDataDir + "WD-n005w4-3.json",
	}

	config := inrc2.DefaultPFRSConfig()
	config.IterationsPerWorker = 20000
	config.MaxConcurrentWorkers = 8
	config.MaxTotalWorkers = 0
	config.BranchOnGlobalBest = true
	config.Seed = 42

	var audit inrc2.PFRSAudit
	config.OnAudit = func(a inrc2.PFRSAudit) {
		audit = a
	}

	beam := inrc2.BeamConfig{
		BeamWidth:         5,
		Seeds:             []int64{42, 101, 202},
		DiversitySlotsPct: 30,
		BeamStrategy:      "budget",
		LookaheadWeight:   1.0,
	}

	result, err := inrc2.RunBeamSearch(sc, weekFiles, hist, config, beam, nil)
	if err != nil {
		t.Fatalf("RunBeamSearch failed: %v", err)
	}
	if !result.AllValid {
		t.Error("beam search produced invalid results")
	}
	if result.TotalPenalty <= 0 {
		t.Error("invalid total penalty")
	}
	if len(result.WinningPath) == 0 {
		t.Error("empty winning path")
	}
	_ = audit
}
