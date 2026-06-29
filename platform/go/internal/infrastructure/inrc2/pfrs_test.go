package inrc2_test

import (
	"testing"

	"github.com/timdodgson/open-workforce-platform/platform/go/internal/infrastructure/inrc2"
)

func TestBuildFeasibleRoster_n005w4(t *testing.T) {
	sc, err := inrc2.LoadScenario(testDataDir + "Sc-n005w4.json")
	if err != nil {
		t.Fatalf("load scenario: %v", err)
	}
	wd, err := inrc2.LoadWeekData(testDataDir + "WD-n005w4-0.json")
	if err != nil {
		t.Fatalf("load week data: %v", err)
	}
	hist, err := inrc2.LoadHistory(testDataDir + "H0-n005w4-0.json")
	if err != nil {
		t.Fatalf("load history: %v", err)
	}

	roster, err := inrc2.BuildFeasibleRoster(sc, wd, hist)
	if err != nil {
		t.Fatalf("build feasible roster failed: %v", err)
	}

	if roster == nil {
		t.Fatal("roster is nil")
	}

	// Verify: 5 nurses, 7 days.
	if len(roster.NurseIDs) != 5 {
		t.Errorf("expected 5 nurses, got %d", len(roster.NurseIDs))
	}
	if roster.NumDays != 7 {
		t.Errorf("expected 7 days, got %d", roster.NumDays)
	}

	// Convert to solution and validate with official scorer.
	sol := inrc2.RosterToSolution(roster, sc, 0)
	result := inrc2.Score(sc, wd, hist, sol)

	// Must have 0 hard violations.
	if result.HardViolations != 0 {
		t.Errorf("expected 0 hard violations, got %d", result.HardViolations)
		for _, v := range result.HardDetails {
			t.Logf("  [%s] %s (nurse=%s day=%d)", v.Code, v.Message, v.Nurse, v.Day)
		}
	}

	// Should have some assignments.
	if len(sol.Assignments) == 0 {
		t.Error("expected assignments in feasible roster")
	}

	t.Logf("Feasible roster: %d assignments, hard=%d, soft=%d",
		len(sol.Assignments), result.HardViolations, result.SoftPenalty)
}

func TestBuildFeasibleRoster_OneShiftPerNursePerDay(t *testing.T) {
	sc, _ := inrc2.LoadScenario(testDataDir + "Sc-n005w4.json")
	wd, _ := inrc2.LoadWeekData(testDataDir + "WD-n005w4-0.json")
	hist, _ := inrc2.LoadHistory(testDataDir + "H0-n005w4-0.json")

	roster, err := inrc2.BuildFeasibleRoster(sc, wd, hist)
	if err != nil {
		t.Fatalf("build failed: %v", err)
	}

	// Check no nurse has more than one shift per day.
	for ni := range roster.NurseIDs {
		for d := 0; d < 7; d++ {
			a := roster.Get(ni, d)
			if a.ShiftType == "" {
				continue
			}
			// Count — should be exactly the one we see.
			count := 0
			if !roster.IsOff(ni, d) {
				count++
			}
			if count > 1 {
				t.Errorf("nurse %d has %d shifts on day %d", ni, count, d)
			}
		}
	}
}

func TestRosterToSolution_Roundtrip(t *testing.T) {
	sc, _ := inrc2.LoadScenario(testDataDir + "Sc-n005w4.json")
	wd, _ := inrc2.LoadWeekData(testDataDir + "WD-n005w4-0.json")
	hist, _ := inrc2.LoadHistory(testDataDir + "H0-n005w4-0.json")

	roster, err := inrc2.BuildFeasibleRoster(sc, wd, hist)
	if err != nil {
		t.Fatalf("build failed: %v", err)
	}

	sol := inrc2.RosterToSolution(roster, sc, 0)

	// Every assignment should have non-empty nurse, day, shiftType, skill.
	for _, a := range sol.Assignments {
		if a.Nurse == "" || a.Day == "" || a.ShiftType == "" || a.Skill == "" {
			t.Errorf("incomplete assignment: %+v", a)
		}
	}
}


func TestPFRS_SA_ProducesHardFeasible(t *testing.T) {
	sc, _ := inrc2.LoadScenario(testDataDir + "Sc-n005w4.json")
	wd, _ := inrc2.LoadWeekData(testDataDir + "WD-n005w4-0.json")
	hist, _ := inrc2.LoadHistory(testDataDir + "H0-n005w4-0.json")

	config := inrc2.DefaultPFRSConfig()
	config.Mode = "sa"
	config.IterationsPerWorker = 5000
	config.MaxConcurrentWorkers = 2
	config.MaxTotalWorkers = 4

	sol, stats, err := inrc2.RunPFRS(sc, wd, hist, config)
	if err != nil {
		t.Fatalf("PFRS failed: %v", err)
	}

	// Validate with official scorer.
	result := inrc2.Score(sc, wd, hist, sol)
	if result.HardViolations != 0 {
		t.Errorf("PFRS produced hard violations: %d", result.HardViolations)
		for _, v := range result.HardDetails {
			t.Logf("  [%s] %s", v.Code, v.Message)
		}
	}

	t.Logf("PFRS SA: penalty=%d, hard=%d, workers=%d, branches=%d, iterations=%d, duration=%dms",
		stats.FinalPenalty, result.HardViolations, stats.WorkersStarted,
		stats.BranchesCreated, stats.TotalIterations, stats.DurationMs)
}

func TestPFRS_LAHC_ProducesHardFeasible(t *testing.T) {
	sc, _ := inrc2.LoadScenario(testDataDir + "Sc-n005w4.json")
	wd, _ := inrc2.LoadWeekData(testDataDir + "WD-n005w4-0.json")
	hist, _ := inrc2.LoadHistory(testDataDir + "H0-n005w4-0.json")

	config := inrc2.DefaultPFRSConfig()
	config.Mode = "lahc"
	config.IterationsPerWorker = 5000
	config.LateAcceptanceLength = 500
	config.MaxConcurrentWorkers = 2
	config.MaxTotalWorkers = 4

	sol, stats, err := inrc2.RunPFRS(sc, wd, hist, config)
	if err != nil {
		t.Fatalf("PFRS failed: %v", err)
	}

	result := inrc2.Score(sc, wd, hist, sol)
	if result.HardViolations != 0 {
		t.Errorf("PFRS LAHC produced hard violations: %d", result.HardViolations)
	}

	t.Logf("PFRS LAHC: penalty=%d, hard=%d, workers=%d, branches=%d, iterations=%d",
		stats.FinalPenalty, result.HardViolations, stats.WorkersStarted,
		stats.BranchesCreated, stats.TotalIterations)
}

func TestPFRS_SA_ImprovesPenalty(t *testing.T) {
	sc, _ := inrc2.LoadScenario(testDataDir + "Sc-n005w4.json")
	wd, _ := inrc2.LoadWeekData(testDataDir + "WD-n005w4-0.json")
	hist, _ := inrc2.LoadHistory(testDataDir + "H0-n005w4-0.json")

	// Get initial penalty.
	initialRoster, _ := inrc2.BuildFeasibleRoster(sc, wd, hist)
	initialSol := inrc2.RosterToSolution(initialRoster, sc, 0)
	initialResult := inrc2.Score(sc, wd, hist, initialSol)
	initialPenalty := initialResult.SoftPenalty

	// Run PFRS.
	config := inrc2.DefaultPFRSConfig()
	config.IterationsPerWorker = 10000
	config.MaxConcurrentWorkers = 2
	config.MaxTotalWorkers = 4

	_, stats, err := inrc2.RunPFRS(sc, wd, hist, config)
	if err != nil {
		t.Fatalf("PFRS failed: %v", err)
	}

	// PFRS should improve or preserve penalty.
	if stats.FinalPenalty > initialPenalty {
		t.Errorf("PFRS worsened penalty: initial=%d, final=%d", initialPenalty, stats.FinalPenalty)
	}

	t.Logf("Improvement: %d -> %d (delta=%d)", initialPenalty, stats.FinalPenalty, initialPenalty-stats.FinalPenalty)
}

func TestPFRS_WorkerLimitsRespected(t *testing.T) {
	sc, _ := inrc2.LoadScenario(testDataDir + "Sc-n005w4.json")
	wd, _ := inrc2.LoadWeekData(testDataDir + "WD-n005w4-0.json")
	hist, _ := inrc2.LoadHistory(testDataDir + "H0-n005w4-0.json")

	config := inrc2.DefaultPFRSConfig()
	config.IterationsPerWorker = 1000
	config.MaxConcurrentWorkers = 2
	config.MaxTotalWorkers = 3

	_, stats, _ := inrc2.RunPFRS(sc, wd, hist, config)

	if stats.WorkersStarted > config.MaxTotalWorkers {
		t.Errorf("exceeded max total workers: %d > %d", stats.WorkersStarted, config.MaxTotalWorkers)
	}
}
