package inrc2_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/timdodgson/open-workforce-platform/platform/go/internal/infrastructure/inrc2"
)

func TestBuildDiscoveryRows_DerivedMetrics(t *testing.T) {
	events := []inrc2.DiscoveryEvent{
		{TimestampMs: 100, WorkerID: 0, Candidate: 1000, Temperature: 90.0, CurrentPenalty: 450, PreviousBest: 500, NewBest: 450, Improvement: 50, EventType: "LOCAL_BEST"},
		{TimestampMs: 300, WorkerID: 0, Candidate: 5000, Temperature: 70.0, CurrentPenalty: 420, PreviousBest: 450, NewBest: 420, Improvement: 30, EventType: "LOCAL_BEST"},
		{TimestampMs: 600, WorkerID: 1, Candidate: 8000, Temperature: 50.0, CurrentPenalty: 400, PreviousBest: 420, NewBest: 400, Improvement: 20, EventType: "GLOBAL_BEST"},
	}

	ctx := inrc2.RunContext{RunID: "test", Instance: "test", Seed: 42, Temperature: 100.0}
	depthMap := map[int]int{0: 0, 1: 1}

	rows := inrc2.BuildDiscoveryRows(ctx, 1, 1, 42, events, depthMap, 1)

	if len(rows) != 3 {
		t.Fatalf("expected 3 rows, got %d", len(rows))
	}

	// First discovery: number=1, candsSincePrevious=1000 (from 0), timeSincePrev=100ms.
	if rows[0].DiscoveryNumber != 1 {
		t.Errorf("first discovery number should be 1, got %d", rows[0].DiscoveryNumber)
	}
	if rows[0].CandsSincePrevious != 1000 {
		t.Errorf("first cands_since should be 1000, got %d", rows[0].CandsSincePrevious)
	}

	// Second discovery: number=2, candsSincePrevious=4000 (5000-1000).
	if rows[1].DiscoveryNumber != 2 {
		t.Errorf("second discovery number should be 2, got %d", rows[1].DiscoveryNumber)
	}
	if rows[1].CandsSincePrevious != 4000 {
		t.Errorf("second cands_since should be 4000, got %d", rows[1].CandsSincePrevious)
	}
	if rows[1].TimeSincePreviousMs != 200 {
		t.Errorf("second time_since should be 200ms, got %d", rows[1].TimeSincePreviousMs)
	}

	// ImprovementPer10K for row 1: 30 / 4000 * 10000 = 75.
	expectedYield := 30.0 / 4000.0 * 10000.0
	if rows[1].ImprovementPer10K < expectedYield-0.01 || rows[1].ImprovementPer10K > expectedYield+0.01 {
		t.Errorf("yield should be ~%.2f, got %.2f", expectedYield, rows[1].ImprovementPer10K)
	}

	// Third row should be GLOBAL_BEST, depth 1.
	if rows[2].EventType != "GLOBAL_BEST" {
		t.Errorf("third event should be GLOBAL_BEST, got %s", rows[2].EventType)
	}
	if rows[2].BranchDepth != 1 {
		t.Errorf("third depth should be 1, got %d", rows[2].BranchDepth)
	}
}

func TestDiscoveriesCSVHeader_ColumnCount(t *testing.T) {
	header := inrc2.DiscoveriesCSVHeader()
	cols := strings.Split(header, ",")
	if len(cols) != 36 {
		t.Errorf("expected 36 columns, got %d", len(cols))
	}
}

func TestWriteDiscoveriesCSV_WritesFile(t *testing.T) {
	rows := []inrc2.DiscoveryRow{
		{
			RunID: "run1", Instance: "test", Seed: 42, BeamWidth: 5,
			Iterations: 500000, Temperature: 100.0, CoolingMode: "adaptive", Timestamp: "2025-01-01T00:00:00Z",
			Week: 1, WorkerID: 0, BeamPath: 1, Candidate: 1000, ElapsedMs: 100,
			TemperatureAtEvent: 90.0, CurrentPenalty: 450, PreviousBest: 500, NewBest: 450,
			Improvement: 50, ImprovementPercent: 10.0, EventType: "GLOBAL_BEST",
			BranchDepth: 0, SeedUsed: 42, AcceptedWorseCount: 10, HardRejectCount: 500, SoftRejectCount: 100,
			DiscoveryNumber: 1, CandsSincePrevious: 1000, TimeSincePreviousMs: 100,
			ImprovementPer10K: 500.0, ImprovementPerSecond: 500.0,
		},
	}

	dir := t.TempDir()
	path := filepath.Join(dir, "discoveries.csv")

	err := inrc2.WriteDiscoveriesCSV(path, rows)
	if err != nil {
		t.Fatalf("WriteDiscoveriesCSV failed: %v", err)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read written file: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(content)), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines (header + 1 row), got %d", len(lines))
	}

	if !strings.Contains(lines[0], "discovery_number") {
		t.Error("header missing discovery_number column")
	}
	if !strings.Contains(lines[0], "improvement_per_10k") {
		t.Error("header missing improvement_per_10k column")
	}
}

func TestDiscoveryEvents_CapturedInPFRS(t *testing.T) {
	sc, _ := inrc2.LoadScenario(testDataDir + "Sc-n005w4.json")
	wd, _ := inrc2.LoadWeekData(testDataDir + "WD-n005w4-0.json")
	hist, _ := inrc2.LoadHistory(testDataDir + "H0-n005w4-0.json")

	var audit inrc2.PFRSAudit
	config := inrc2.DefaultPFRSConfig()
	config.IterationsPerWorker = 10000
	config.MaxTotalWorkers = 2
	config.Seed = 42
	config.OnAudit = func(a inrc2.PFRSAudit) {
		audit = a
	}

	_, _, err := inrc2.RunPFRS(sc, wd, hist, config)
	if err != nil {
		t.Fatalf("RunPFRS failed: %v", err)
	}

	// Should have at least one discovery (the initial local best improvement).
	if len(audit.Discoveries) == 0 {
		t.Fatal("expected at least one discovery event, got 0")
	}

	// All discoveries should have positive improvement.
	for i, d := range audit.Discoveries {
		if d.Improvement <= 0 {
			t.Errorf("discovery %d has non-positive improvement: %d", i, d.Improvement)
		}
		if d.EventType != "LOCAL_BEST" && d.EventType != "GLOBAL_BEST" {
			t.Errorf("discovery %d has invalid event type: %s", i, d.EventType)
		}
		if d.NewBest >= d.PreviousBest {
			t.Errorf("discovery %d: new best (%d) should be less than previous (%d)", i, d.NewBest, d.PreviousBest)
		}
	}

	// At least one should be GLOBAL_BEST (the first improvement is always global).
	hasGlobal := false
	for _, d := range audit.Discoveries {
		if d.EventType == "GLOBAL_BEST" {
			hasGlobal = true
			break
		}
	}
	if !hasGlobal {
		t.Error("expected at least one GLOBAL_BEST discovery")
	}
}

func TestPostReheatEnrichment(t *testing.T) {
	// Simulate: worker 0 does LOCAL_BEST, then REHEAT at candidate 100K,
	// then LOCAL_BEST that beats the pre-reheat best.
	events := []inrc2.DiscoveryEvent{
		{TimestampMs: 50, WorkerID: 0, Candidate: 5000, Temperature: 80.0, CurrentPenalty: 450, PreviousBest: 500, NewBest: 450, Improvement: 50, EventType: "LOCAL_BEST"},
		{TimestampMs: 500, WorkerID: 0, Candidate: 100000, Temperature: 0.01, CurrentPenalty: 460, PreviousBest: 450, NewBest: 450, Improvement: 0, EventType: "REHEAT"},
		{TimestampMs: 600, WorkerID: 0, Candidate: 120000, Temperature: 50.0, CurrentPenalty: 440, PreviousBest: 450, NewBest: 440, Improvement: 10, EventType: "LOCAL_BEST"},
		{TimestampMs: 700, WorkerID: 0, Candidate: 150000, Temperature: 40.0, CurrentPenalty: 430, PreviousBest: 440, NewBest: 430, Improvement: 10, EventType: "GLOBAL_BEST"},
	}

	ctx := inrc2.RunContext{RunID: "test", Instance: "test", Seed: 42, Temperature: 100.0}
	depthMap := map[int]int{0: 0}

	rows := inrc2.BuildDiscoveryRows(ctx, 1, 1, 42, events, depthMap, 0)

	if len(rows) != 4 {
		t.Fatalf("expected 4 rows, got %d", len(rows))
	}

	// The REHEAT row is index 1 (sorted by candidate).
	reheatRow := rows[1]
	if reheatRow.EventType != "REHEAT" {
		t.Fatalf("expected row 1 to be REHEAT, got %s", reheatRow.EventType)
	}

	// Post-reheat should show improvement: 450 -> 430 = delta 20.
	if !reheatRow.PostReheatImproved {
		t.Error("post_reheat_improved should be true")
	}
	if reheatRow.PostReheatBestDelta != 20 {
		t.Errorf("post_reheat_best_delta should be 20, got %d", reheatRow.PostReheatBestDelta)
	}
	if reheatRow.PostReheatCandidatesToImprove != 20000 {
		t.Errorf("post_reheat_candidates_to_improve should be 20000, got %d", reheatRow.PostReheatCandidatesToImprove)
	}
	if !reheatRow.PostReheatSpawnedBranch {
		t.Error("post_reheat_spawned_branch should be true (GLOBAL_BEST occurred)")
	}
	if !reheatRow.PostReheatBeatGlobal {
		t.Error("post_reheat_beat_global should be true")
	}
	// Worker 0 is the winning worker.
	if !reheatRow.PostReheatOnWinningLineage {
		t.Error("post_reheat_on_winning_lineage should be true (worker 0 is winning)")
	}
}

func TestPostReheatEnrichment_NoImprovement(t *testing.T) {
	// Worker reheats but never beats its pre-reheat best.
	events := []inrc2.DiscoveryEvent{
		{TimestampMs: 50, WorkerID: 0, Candidate: 5000, Temperature: 80.0, CurrentPenalty: 450, PreviousBest: 500, NewBest: 450, Improvement: 50, EventType: "LOCAL_BEST"},
		{TimestampMs: 500, WorkerID: 0, Candidate: 100000, Temperature: 0.01, CurrentPenalty: 460, PreviousBest: 450, NewBest: 450, Improvement: 0, EventType: "REHEAT"},
		// Next LOCAL_BEST is 455 — worse than the 450 at reheat time.
		{TimestampMs: 600, WorkerID: 0, Candidate: 120000, Temperature: 50.0, CurrentPenalty: 455, PreviousBest: 460, NewBest: 455, Improvement: 5, EventType: "LOCAL_BEST"},
	}

	ctx := inrc2.RunContext{RunID: "test", Instance: "test", Seed: 42, Temperature: 100.0}
	rows := inrc2.BuildDiscoveryRows(ctx, 1, 1, 42, events, nil, -1)

	reheatRow := rows[1]
	if reheatRow.EventType != "REHEAT" {
		t.Fatalf("expected REHEAT row, got %s", reheatRow.EventType)
	}
	if reheatRow.PostReheatImproved {
		t.Error("post_reheat_improved should be false — 455 > 450")
	}
	if reheatRow.PostReheatBestDelta != 0 {
		t.Errorf("post_reheat_best_delta should be 0, got %d", reheatRow.PostReheatBestDelta)
	}
	if reheatRow.PostReheatOnWinningLineage {
		t.Error("should be false when winningWorkerID is -1")
	}
}
