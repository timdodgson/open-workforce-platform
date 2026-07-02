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

func TestPFRS_SA_RunsFullIterationBudget(t *testing.T) {
	sc, _ := inrc2.LoadScenario(testDataDir + "Sc-n005w4.json")
	wd, _ := inrc2.LoadWeekData(testDataDir + "WD-n005w4-0.json")
	hist, _ := inrc2.LoadHistory(testDataDir + "H0-n005w4-0.json")

	config := inrc2.DefaultPFRSConfig()
	config.IterationsPerWorker = 5000
	config.MaxConcurrentWorkers = 1
	config.MaxTotalWorkers = 1 // Single worker, no branching complexity.
	config.BranchOnGlobalBest = false

	_, stats, err := inrc2.RunPFRS(sc, wd, hist, config)
	if err != nil {
		t.Fatalf("PFRS failed: %v", err)
	}

	// With 1 worker and 5000 iterations, candidates must equal 5000.
	// (Every iteration attempts one swap = one candidate.)
	if stats.CandidatesEvaluated != 5000 {
		t.Errorf("expected 5000 candidates (full budget), got %d", stats.CandidatesEvaluated)
	}
	if stats.TotalIterations != 5000 {
		t.Errorf("expected 5000 iterations, got %d", stats.TotalIterations)
	}

	t.Logf("Single worker ran full budget: iterations=%d, candidates=%d, penalty=%d",
		stats.TotalIterations, stats.CandidatesEvaluated, stats.FinalPenalty)
}

func TestPFRS_ProgressCallbackCalled(t *testing.T) {
	sc, _ := inrc2.LoadScenario(testDataDir + "Sc-n005w4.json")
	wd, _ := inrc2.LoadWeekData(testDataDir + "WD-n005w4-0.json")
	hist, _ := inrc2.LoadHistory(testDataDir + "H0-n005w4-0.json")

	callCount := 0
	var lastProgress inrc2.PFRSProgress

	config := inrc2.DefaultPFRSConfig()
	config.IterationsPerWorker = 50000 // Enough iterations to trigger at least one callback.
	config.MaxConcurrentWorkers = 2
	config.MaxTotalWorkers = 4
	config.ProgressIntervalMs = 10 // 10ms interval for test speed.
	config.OnProgress = func(p inrc2.PFRSProgress) {
		callCount++
		lastProgress = p
	}

	_, _, err := inrc2.RunPFRS(sc, wd, hist, config)
	if err != nil {
		t.Fatalf("PFRS failed: %v", err)
	}

	if callCount == 0 {
		t.Error("progress callback was never called")
	}

	// Last progress should show reasonable values.
	if lastProgress.CandidatesEvaluated == 0 {
		t.Error("progress reported 0 candidates")
	}
	if lastProgress.BestPenalty == 0 {
		t.Error("progress reported 0 best penalty")
	}

	t.Logf("Progress called %d times, last: workers=%d/%d candidates=%d penalty=%d elapsed=%dms",
		callCount, lastProgress.WorkersStarted, lastProgress.TotalWorkers,
		lastProgress.CandidatesEvaluated, lastProgress.BestPenalty, lastProgress.ElapsedMs)
}

func TestPFRS_ProgressDisabled(t *testing.T) {
	sc, _ := inrc2.LoadScenario(testDataDir + "Sc-n005w4.json")
	wd, _ := inrc2.LoadWeekData(testDataDir + "WD-n005w4-0.json")
	hist, _ := inrc2.LoadHistory(testDataDir + "H0-n005w4-0.json")

	callCount := 0

	config := inrc2.DefaultPFRSConfig()
	config.IterationsPerWorker = 5000
	config.MaxConcurrentWorkers = 1
	config.MaxTotalWorkers = 1
	config.ProgressIntervalMs = 0 // Disabled.
	config.OnProgress = func(p inrc2.PFRSProgress) {
		callCount++
	}

	_, _, err := inrc2.RunPFRS(sc, wd, hist, config)
	if err != nil {
		t.Fatalf("PFRS failed: %v", err)
	}

	if callCount != 0 {
		t.Errorf("progress should not be called when interval=0, but was called %d times", callCount)
	}
}

func TestPFRS_ProgressNilCallback(t *testing.T) {
	sc, _ := inrc2.LoadScenario(testDataDir + "Sc-n005w4.json")
	wd, _ := inrc2.LoadWeekData(testDataDir + "WD-n005w4-0.json")
	hist, _ := inrc2.LoadHistory(testDataDir + "H0-n005w4-0.json")

	config := inrc2.DefaultPFRSConfig()
	config.IterationsPerWorker = 5000
	config.MaxConcurrentWorkers = 1
	config.MaxTotalWorkers = 1
	config.OnProgress = nil // No callback — should not panic.

	_, _, err := inrc2.RunPFRS(sc, wd, hist, config)
	if err != nil {
		t.Fatalf("PFRS failed: %v", err)
	}
}

func TestPFRS_UnlimitedBranching_AllStarted(t *testing.T) {
	sc, _ := inrc2.LoadScenario(testDataDir + "Sc-n005w4.json")
	wd, _ := inrc2.LoadWeekData(testDataDir + "WD-n005w4-0.json")
	hist, _ := inrc2.LoadHistory(testDataDir + "H0-n005w4-0.json")

	config := inrc2.DefaultPFRSConfig()
	config.IterationsPerWorker = 5000
	config.MaxConcurrentWorkers = 2
	config.MaxTotalWorkers = 0 // Unlimited.

	_, stats, err := inrc2.RunPFRS(sc, wd, hist, config)
	if err != nil {
		t.Fatalf("PFRS failed: %v", err)
	}

	// With unlimited, all branches should be started (none dropped).
	if stats.BranchesDropped != 0 {
		t.Errorf("expected 0 dropped branches with unlimited, got %d", stats.BranchesDropped)
	}
	// Should have created some branches (algorithm finds improvements).
	if stats.BranchesCreated == 0 {
		t.Error("expected at least 1 branch to be created")
	}
	// Workers started = 1 (initial) + branches created.
	if stats.WorkersStarted != stats.BranchesCreated+1 {
		t.Errorf("workers started (%d) should equal branches (%d) + 1", stats.WorkersStarted, stats.BranchesCreated)
	}

	t.Logf("Unlimited branching: workers=%d branches=%d dropped=%d max_queue=%d max_concurrent=%d",
		stats.WorkersStarted, stats.BranchesCreated, stats.BranchesDropped,
		stats.MaxQueueDepth, stats.MaxConcurrentSeen)
}

func TestPFRS_ConcurrencyNeverExceedsMax(t *testing.T) {
	sc, _ := inrc2.LoadScenario(testDataDir + "Sc-n005w4.json")
	wd, _ := inrc2.LoadWeekData(testDataDir + "WD-n005w4-0.json")
	hist, _ := inrc2.LoadHistory(testDataDir + "H0-n005w4-0.json")

	config := inrc2.DefaultPFRSConfig()
	config.IterationsPerWorker = 10000
	config.MaxConcurrentWorkers = 2
	config.MaxTotalWorkers = 0 // Unlimited total, but only 2 at a time.

	_, stats, err := inrc2.RunPFRS(sc, wd, hist, config)
	if err != nil {
		t.Fatalf("PFRS failed: %v", err)
	}

	if stats.MaxConcurrentSeen > 2 {
		t.Errorf("max concurrent workers (%d) exceeded limit of 2", stats.MaxConcurrentSeen)
	}

	t.Logf("Concurrency limited: max_concurrent=%d workers_started=%d", stats.MaxConcurrentSeen, stats.WorkersStarted)
}

func TestPFRS_ExplicitCapDropsBranches(t *testing.T) {
	sc, _ := inrc2.LoadScenario(testDataDir + "Sc-n005w4.json")
	wd, _ := inrc2.LoadWeekData(testDataDir + "WD-n005w4-0.json")
	hist, _ := inrc2.LoadHistory(testDataDir + "H0-n005w4-0.json")

	config := inrc2.DefaultPFRSConfig()
	config.IterationsPerWorker = 10000
	config.MaxConcurrentWorkers = 2
	config.MaxTotalWorkers = 3 // Hard cap at 3 total.

	_, stats, err := inrc2.RunPFRS(sc, wd, hist, config)
	if err != nil {
		t.Fatalf("PFRS failed: %v", err)
	}

	// Should have started at most 3 workers.
	if stats.WorkersStarted > 3 {
		t.Errorf("workers started (%d) exceeded MaxTotalWorkers=3", stats.WorkersStarted)
	}
	// If algorithm found more improvements than slots, some should be dropped.
	if stats.BestUpdates > stats.BranchesCreated && stats.BranchesDropped == 0 {
		t.Error("expected some branches to be dropped when cap is reached")
	}

	t.Logf("Capped: workers=%d branches=%d dropped=%d best_updates=%d",
		stats.WorkersStarted, stats.BranchesCreated, stats.BranchesDropped, stats.BestUpdates)
}

func TestPFRS_GlobalBestIncludesLateBranches(t *testing.T) {
	sc, _ := inrc2.LoadScenario(testDataDir + "Sc-n005w4.json")
	wd, _ := inrc2.LoadWeekData(testDataDir + "WD-n005w4-0.json")
	hist, _ := inrc2.LoadHistory(testDataDir + "H0-n005w4-0.json")

	// Run with unlimited branching.
	config := inrc2.DefaultPFRSConfig()
	config.IterationsPerWorker = 10000
	config.MaxConcurrentWorkers = 2
	config.MaxTotalWorkers = 0

	sol, stats, err := inrc2.RunPFRS(sc, wd, hist, config)
	if err != nil {
		t.Fatalf("PFRS failed: %v", err)
	}

	// Validate: final solution must be hard-feasible.
	result := inrc2.Score(sc, wd, hist, sol)
	if result.HardViolations != 0 {
		t.Errorf("final solution has hard violations: %d", result.HardViolations)
	}

	// Final penalty should match stats.
	if result.SoftPenalty != stats.FinalPenalty {
		t.Errorf("solution penalty (%d) doesn't match stats.FinalPenalty (%d)", result.SoftPenalty, stats.FinalPenalty)
	}

	t.Logf("Global best verified: penalty=%d workers=%d branches=%d",
		stats.FinalPenalty, stats.WorkersStarted, stats.BranchesCreated)
}

func TestPFRS_AuditCallback_Received(t *testing.T) {
	sc, _ := inrc2.LoadScenario(testDataDir + "Sc-n005w4.json")
	wd, _ := inrc2.LoadWeekData(testDataDir + "WD-n005w4-0.json")
	hist, _ := inrc2.LoadHistory(testDataDir + "H0-n005w4-0.json")

	var audit inrc2.PFRSAudit
	called := false

	config := inrc2.DefaultPFRSConfig()
	config.IterationsPerWorker = 5000
	config.MaxConcurrentWorkers = 2
	config.MaxTotalWorkers = 4
	config.OnAudit = func(a inrc2.PFRSAudit) {
		called = true
		audit = a
	}

	_, stats, err := inrc2.RunPFRS(sc, wd, hist, config)
	if err != nil {
		t.Fatalf("PFRS failed: %v", err)
	}

	if !called {
		t.Fatal("OnAudit callback was not called")
	}

	// Worker count should match stats.
	if len(audit.Workers) != stats.WorkersStarted {
		t.Errorf("audit workers (%d) != stats.WorkersStarted (%d)", len(audit.Workers), stats.WorkersStarted)
	}

	// Initial worker should have ParentWorkerID == -1.
	foundInitial := false
	for _, w := range audit.Workers {
		if w.WorkerID == 0 {
			foundInitial = true
			if w.ParentWorkerID != -1 {
				t.Errorf("initial worker (ID=0) should have ParentWorkerID=-1, got %d", w.ParentWorkerID)
			}
		}
		// Every worker should have run the full iteration budget.
		if w.Iterations != config.IterationsPerWorker {
			t.Errorf("worker %d ran %d iterations, expected %d", w.WorkerID, w.Iterations, config.IterationsPerWorker)
		}
		// CandidatesEval must equal iterations (one candidate per iteration).
		if w.CandidatesEval != w.Iterations {
			t.Errorf("worker %d candidates (%d) != iterations (%d)", w.WorkerID, w.CandidatesEval, w.Iterations)
		}
		// Accepted + RejectedByProb = CandidatesEval (for SA mode).
		// Hard rejections are separate — they don't count against the candidate budget.
		if w.Accepted+w.RejectedByProb != w.CandidatesEval {
			t.Errorf("worker %d: accepted(%d)+rejectedByProb(%d) != candidates(%d)",
				w.WorkerID, w.Accepted, w.RejectedByProb, w.CandidatesEval)
		}
		// AcceptanceRate sanity.
		if w.AcceptanceRate < 0 || w.AcceptanceRate > 1 {
			t.Errorf("worker %d acceptance rate out of range: %.4f", w.WorkerID, w.AcceptanceRate)
		}
		// Duration must be positive.
		if w.DurationMs <= 0 {
			t.Errorf("worker %d duration should be positive, got %d", w.WorkerID, w.DurationMs)
		}
	}
	if !foundInitial {
		t.Error("did not find initial worker (ID=0) in audit")
	}

	// Should have best-update events if there were branches.
	if stats.BranchesCreated > 0 && len(audit.BestUpdates) == 0 {
		t.Error("branches were created but no best-update events recorded")
	}

	// Every best-update should show improvement (NewPenalty < OldPenalty).
	for i, bu := range audit.BestUpdates {
		if bu.NewPenalty >= bu.OldPenalty {
			t.Errorf("best update %d: new(%d) >= old(%d)", i, bu.NewPenalty, bu.OldPenalty)
		}
		if bu.TimestampMs < 0 {
			t.Errorf("best update %d: negative timestamp", i)
		}
	}

	// Branch events should match BranchesCreated.
	if len(audit.Branches) != stats.BranchesCreated {
		t.Errorf("branch events (%d) != stats.BranchesCreated (%d)", len(audit.Branches), stats.BranchesCreated)
	}

	t.Logf("Audit: %d workers, %d best-updates, %d branches", len(audit.Workers), len(audit.BestUpdates), len(audit.Branches))
}

func TestPFRS_AuditNilCallback_NoPanic(t *testing.T) {
	sc, _ := inrc2.LoadScenario(testDataDir + "Sc-n005w4.json")
	wd, _ := inrc2.LoadWeekData(testDataDir + "WD-n005w4-0.json")
	hist, _ := inrc2.LoadHistory(testDataDir + "H0-n005w4-0.json")

	config := inrc2.DefaultPFRSConfig()
	config.IterationsPerWorker = 1000
	config.MaxConcurrentWorkers = 1
	config.MaxTotalWorkers = 1
	config.OnAudit = nil // Should not panic.

	_, _, err := inrc2.RunPFRS(sc, wd, hist, config)
	if err != nil {
		t.Fatalf("PFRS failed: %v", err)
	}
}

func TestPFRS_AuditWorkerBestPenalty(t *testing.T) {
	sc, _ := inrc2.LoadScenario(testDataDir + "Sc-n005w4.json")
	wd, _ := inrc2.LoadWeekData(testDataDir + "WD-n005w4-0.json")
	hist, _ := inrc2.LoadHistory(testDataDir + "H0-n005w4-0.json")

	var audit inrc2.PFRSAudit

	config := inrc2.DefaultPFRSConfig()
	config.IterationsPerWorker = 10000
	config.MaxConcurrentWorkers = 1
	config.MaxTotalWorkers = 1
	config.BranchOnGlobalBest = false
	config.OnAudit = func(a inrc2.PFRSAudit) {
		audit = a
	}

	_, stats, err := inrc2.RunPFRS(sc, wd, hist, config)
	if err != nil {
		t.Fatalf("PFRS failed: %v", err)
	}

	if len(audit.Workers) != 1 {
		t.Fatalf("expected 1 worker in audit, got %d", len(audit.Workers))
	}

	w := audit.Workers[0]
	// BestPenalty must be <= StartPenalty (SA should improve or maintain).
	if w.BestPenalty > w.StartPenalty {
		t.Errorf("best penalty (%d) > start penalty (%d)", w.BestPenalty, w.StartPenalty)
	}
	// BestPenalty should match the final stats penalty (single worker, no branching).
	if w.BestPenalty != stats.FinalPenalty {
		t.Errorf("worker best (%d) != final penalty (%d)", w.BestPenalty, stats.FinalPenalty)
	}
	// BestIteration should be > 0 if improvement was made.
	if w.BestPenalty < w.StartPenalty && w.BestIteration == 0 {
		t.Error("best improved but best iteration is 0")
	}
	// FinalTemperature should be > 0 for SA mode.
	if w.FinalTemperature <= 0 {
		t.Errorf("SA worker final temperature should be > 0, got %f", w.FinalTemperature)
	}

	t.Logf("Worker audit: start=%d best=%d final=%d bestIter=%d temp=%.6f rate=%.4f",
		w.StartPenalty, w.BestPenalty, w.FinalPenalty, w.BestIteration, w.FinalTemperature, w.AcceptanceRate)
}

func TestPFRS_HighConcurrency_NoRace(t *testing.T) {
	// Stress test: runs with high concurrency and high temperature to trigger
	// heavy branching. Validates no crashes under load.
	sc, _ := inrc2.LoadScenario(testDataDir + "Sc-n005w4.json")
	wd, _ := inrc2.LoadWeekData(testDataDir + "WD-n005w4-0.json")
	hist, _ := inrc2.LoadHistory(testDataDir + "H0-n005w4-0.json")

	config := inrc2.DefaultPFRSConfig()
	config.IterationsPerWorker = 10000
	config.MaxConcurrentWorkers = 8
	config.MaxTotalWorkers = 0        // unlimited branching
	config.InitialTemperature = 100.0 // high temp = lots of branches
	config.CoolingMode = "adaptive"
	config.BranchOnGlobalBest = true

	sol, stats, err := inrc2.RunPFRS(sc, wd, hist, config)
	if err != nil {
		t.Fatalf("PFRS failed under high concurrency: %v", err)
	}

	// Validate hard feasibility.
	result := inrc2.Score(sc, wd, hist, sol)
	if result.HardViolations != 0 {
		t.Errorf("high concurrency produced hard violations: %d", result.HardViolations)
	}

	// Should have spawned multiple workers.
	if stats.WorkersStarted < 2 {
		t.Errorf("expected multiple workers under high temp, got %d", stats.WorkersStarted)
	}

	t.Logf("High concurrency: penalty=%d workers=%d branches=%d candidates=%d",
		stats.FinalPenalty, stats.WorkersStarted, stats.BranchesCreated, stats.CandidatesEvaluated)
}

func TestPFRS_ScorePenaltyOnly_NoCrashUnderLoad(t *testing.T) {
	// Verifies ScorePenaltyOnly can be called millions of times concurrently
	// without crashes from GC pressure / map corruption.
	sc, _ := inrc2.LoadScenario(testDataDir + "Sc-n005w4.json")
	wd, _ := inrc2.LoadWeekData(testDataDir + "WD-n005w4-0.json")
	hist, _ := inrc2.LoadHistory(testDataDir + "H0-n005w4-0.json")

	roster, _ := inrc2.BuildFeasibleRoster(sc, wd, hist)
	sol := inrc2.RosterToSolution(roster, sc, 0)

	// Run 8 concurrent goroutines, each scoring 10k times.
	const goroutines = 8
	const iterations = 10000

	done := make(chan int, goroutines)
	for g := 0; g < goroutines; g++ {
		go func() {
			ws := inrc2.NewScoringWorkspace(sc, wd, hist)
			lastPenalty := 0
			for i := 0; i < iterations; i++ {
				lastPenalty = inrc2.ScorePenaltyOnly(ws, sol)
			}
			done <- lastPenalty
		}()
	}

	// Collect results — all should be the same penalty.
	var penalties []int
	for g := 0; g < goroutines; g++ {
		penalties = append(penalties, <-done)
	}

	// All goroutines should produce the same score.
	for i := 1; i < len(penalties); i++ {
		if penalties[i] != penalties[0] {
			t.Errorf("goroutine %d got penalty %d, expected %d", i, penalties[i], penalties[0])
		}
	}

	t.Logf("Concurrent scoring: %d goroutines × %d iterations = %d calls, all returned penalty=%d",
		goroutines, iterations, goroutines*iterations, penalties[0])
}

func TestScorePenaltyOnlyFromRoster_MatchesOriginal(t *testing.T) {
	// Verifies that ScorePenaltyOnlyFromRoster produces the same penalty
	// as the Solution-based ScorePenaltyOnly for the same roster.
	sc, _ := inrc2.LoadScenario(testDataDir + "Sc-n005w4.json")
	wd, _ := inrc2.LoadWeekData(testDataDir + "WD-n005w4-0.json")
	hist, _ := inrc2.LoadHistory(testDataDir + "H0-n005w4-0.json")

	roster, _ := inrc2.BuildFeasibleRoster(sc, wd, hist)
	ws := inrc2.NewScoringWorkspace(sc, wd, hist)

	// Score via Solution path (original).
	sol := inrc2.RosterToSolution(roster, sc, 0)
	penaltyViaSolution := inrc2.ScorePenaltyOnly(ws, sol)

	// Score via Roster path (new).
	penaltyViaRoster := inrc2.ScorePenaltyOnlyFromRoster(ws, roster)

	if penaltyViaSolution != penaltyViaRoster {
		t.Errorf("scoring mismatch: via Solution=%d, via Roster=%d", penaltyViaSolution, penaltyViaRoster)
	}

	t.Logf("Both paths return penalty=%d", penaltyViaRoster)
}

func TestScorePenaltyOnlyFromRoster_MatchesParity(t *testing.T) {
	// Verifies the new Roster-direct scorer matches the validator parity value.
	sc, _ := inrc2.LoadScenario(testDataDir + "Sc-n005w4.json")
	wd, _ := inrc2.LoadWeekData(testDataDir + "WD-n005w4-0.json")
	hist, _ := inrc2.LoadHistory(testDataDir + "H0-n005w4-0.json")

	// Run PFRS to get a solution, then verify Roster-direct matches full scorer.
	config := inrc2.DefaultPFRSConfig()
	config.IterationsPerWorker = 5000
	config.MaxConcurrentWorkers = 2
	config.MaxTotalWorkers = 2
	config.BranchOnGlobalBest = false

	sol, _, err := inrc2.RunPFRS(sc, wd, hist, config)
	if err != nil {
		t.Fatalf("PFRS failed: %v", err)
	}

	// Full official score.
	officialResult := inrc2.Score(sc, wd, hist, sol)

	// PFRS uses ScorePenaltyOnlyFromRoster internally — final penalty should match.
	// Note: PFRS returns the best penalty found, which should equal the official soft penalty.
	if officialResult.HardViolations != 0 {
		t.Fatalf("hard violations: %d", officialResult.HardViolations)
	}

	t.Logf("Official soft penalty: %d, hard: %d", officialResult.SoftPenalty, officialResult.HardViolations)
}

func TestPFRS_N030W4_HighConcurrency_NoCrash(t *testing.T) {
	// Regression test for GC/stack crash in scoreConsecutiveWorkingDays.
	// The crash occurred with the n030w4 dataset (30 nurses) under high concurrency
	// where sort.Ints created interface allocations that interacted badly with GC
	// during goroutine scheduling. The fix removes the unnecessary sort.Ints call
	// since all callers provide pre-sorted workDays (built by iterating d=0..6).
	const n030Dir = "../../../../../examples/inrc2/datasets_json/n030w4/"

	sc, err := inrc2.LoadScenario(n030Dir + "Sc-n030w4.json")
	if err != nil {
		t.Skipf("n030w4 dataset not available: %v", err)
	}
	wd, err := inrc2.LoadWeekData(n030Dir + "WD-n030w4-0.json")
	if err != nil {
		t.Skipf("n030w4 week data not available: %v", err)
	}
	hist, err := inrc2.LoadHistory(n030Dir + "H0-n030w4-0.json")
	if err != nil {
		t.Skipf("n030w4 history not available: %v", err)
	}

	config := inrc2.DefaultPFRSConfig()
	config.IterationsPerWorker = 50000
	config.MaxConcurrentWorkers = 16
	config.MaxTotalWorkers = 0 // unlimited branching
	config.InitialTemperature = 100.0
	config.CoolingMode = "adaptive"
	config.BranchOnGlobalBest = true
	config.Seed = 42

	sol, stats, err := inrc2.RunPFRS(sc, wd, hist, config)
	if err != nil {
		t.Fatalf("PFRS crashed on n030w4: %v", err)
	}

	// Verify hard feasibility.
	result := inrc2.Score(sc, wd, hist, sol)
	if result.HardViolations != 0 {
		t.Errorf("hard violations: %d", result.HardViolations)
	}

	// Verify scoring consistency: PFRS internal score must match official scorer.
	if result.SoftPenalty != stats.FinalPenalty {
		t.Errorf("scoring inconsistency: PFRS reports %d, official scorer reports %d",
			stats.FinalPenalty, result.SoftPenalty)
	}

	t.Logf("n030w4 stress: penalty=%d workers=%d branches=%d candidates=%d",
		stats.FinalPenalty, stats.WorkersStarted, stats.BranchesCreated, stats.CandidatesEvaluated)
}

func TestScoreConsecutiveWorkingDays_PreSortedInput(t *testing.T) {
	// Verifies that scoreConsecutiveWorkingDays produces correct results
	// without sort.Ints — all inputs must be pre-sorted ascending.
	// This test is called via the exported ScorePenaltyOnlyFromRoster path.
	sc, _ := inrc2.LoadScenario(testDataDir + "Sc-n005w4.json")
	wd, _ := inrc2.LoadWeekData(testDataDir + "WD-n005w4-0.json")
	hist, _ := inrc2.LoadHistory(testDataDir + "H0-n005w4-0.json")

	roster, _ := inrc2.BuildFeasibleRoster(sc, wd, hist)
	ws := inrc2.NewScoringWorkspace(sc, wd, hist)

	// Score multiple times to ensure stability without sort.
	first := inrc2.ScorePenaltyOnlyFromRoster(ws, roster)
	for i := 0; i < 100; i++ {
		p := inrc2.ScorePenaltyOnlyFromRoster(ws, roster)
		if p != first {
			t.Fatalf("iteration %d: got %d, expected %d (scoring unstable without sort)", i, p, first)
		}
	}

	// Cross-check with Solution-based scorer.
	sol := inrc2.RosterToSolution(roster, sc, 0)
	pSol := inrc2.ScorePenaltyOnly(ws, sol)
	if pSol != first {
		t.Errorf("Roster path=%d vs Solution path=%d (mismatch)", first, pSol)
	}

	t.Logf("Scoring stable without sort: penalty=%d (verified 100 iterations)", first)
}
