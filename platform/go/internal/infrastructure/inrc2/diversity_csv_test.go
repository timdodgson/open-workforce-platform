package inrc2_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/timdodgson/open-workforce-platform/platform/go/internal/infrastructure/inrc2"
)

func TestSolutionToRoster_RoundTrip(t *testing.T) {
	sc, _ := inrc2.LoadScenario(testDataDir + "Sc-n005w4.json")
	wd, _ := inrc2.LoadWeekData(testDataDir + "WD-n005w4-0.json")
	hist, _ := inrc2.LoadHistory(testDataDir + "H0-n005w4-0.json")

	roster, err := inrc2.BuildFeasibleRoster(sc, wd, hist)
	if err != nil {
		t.Fatalf("BuildFeasibleRoster failed: %v", err)
	}

	// Convert roster -> solution -> roster, verify fingerprints match.
	sol := inrc2.RosterToSolution(roster, sc, 0)
	reconstructed := inrc2.SolutionToRoster(sol, sc)

	fpOriginal := inrc2.RosterFingerprint(roster)
	fpReconstructed := inrc2.RosterFingerprint(reconstructed)

	if fpOriginal != fpReconstructed {
		t.Errorf("round-trip fingerprint mismatch: %s != %s", fpOriginal, fpReconstructed)
	}

	// Hamming distance should be 0 for identical rosters.
	dist := inrc2.RosterHammingDistance(roster, reconstructed)
	if dist != 0.0 {
		t.Errorf("round-trip Hamming distance should be 0, got %f", dist)
	}
}

func TestSolutionToRoster_Deterministic(t *testing.T) {
	sc, _ := inrc2.LoadScenario(testDataDir + "Sc-n005w4.json")
	wd, _ := inrc2.LoadWeekData(testDataDir + "WD-n005w4-0.json")
	hist, _ := inrc2.LoadHistory(testDataDir + "H0-n005w4-0.json")

	roster, _ := inrc2.BuildFeasibleRoster(sc, wd, hist)
	sol := inrc2.RosterToSolution(roster, sc, 0)

	// Call SolutionToRoster multiple times — should always produce same fingerprint.
	fp1 := inrc2.RosterFingerprint(inrc2.SolutionToRoster(sol, sc))
	fp2 := inrc2.RosterFingerprint(inrc2.SolutionToRoster(sol, sc))
	fp3 := inrc2.RosterFingerprint(inrc2.SolutionToRoster(sol, sc))

	if fp1 != fp2 || fp2 != fp3 {
		t.Errorf("SolutionToRoster not deterministic: %s, %s, %s", fp1, fp2, fp3)
	}
}

func TestBuildDiversityRows_BasicProperties(t *testing.T) {
	// Build a minimal BeamResult with known paths.
	result := inrc2.BeamResult{
		WinningPath: []inrc2.BeamPath{
			{ID: 1, ParentID: 0, Week: 1, Seed: 42, WeekPenalty: 100, CumulativePenalty: 100},
		},
		AllPaths: []inrc2.BeamPath{
			{ID: 1, ParentID: 0, Week: 1, Seed: 42, WeekPenalty: 100, CumulativePenalty: 100, Fingerprint: "aabbccddee00", HammingToBest: 0.0},
			{ID: 2, ParentID: 0, Week: 1, Seed: 101, WeekPenalty: 150, CumulativePenalty: 150, Fingerprint: "112233445566", HammingToBest: 0.15},
			{ID: 3, ParentID: 0, Week: 1, Seed: 202, WeekPenalty: 200, CumulativePenalty: 200, Fingerprint: "ffeeddccbb99", HammingToBest: 0.03},
		},
		WeekSummaries: []inrc2.BeamWeekSummary{
			{Week: 1, Candidates: 3, Retained: 2, BestCumulative: 100},
		},
	}

	sc := inrc2.Scenario{
		Nurses: []inrc2.Nurse{{ID: "A"}, {ID: "B"}},
	}

	ctx := inrc2.RunContext{
		RunID:    "test-run",
		Instance: "test",
		Seed:     42,
	}

	rows := inrc2.BuildDiversityRows(ctx, result, sc)

	if len(rows) != 3 {
		t.Fatalf("expected 3 diversity rows, got %d", len(rows))
	}

	// Verify beam spread = worst_retained - best_retained = 150 - 100 = 50.
	for _, r := range rows {
		if r.BeamSpread != 50 {
			t.Errorf("expected beam spread 50, got %d", r.BeamSpread)
		}
	}

	// Path 3 should be near-duplicate (HammingToBest < 0.05).
	found := false
	for _, r := range rows {
		if r.PathID == 3 && r.NearDuplicate {
			found = true
		}
	}
	if !found {
		t.Error("path 3 should be flagged as near-duplicate")
	}

	// Path 2 should NOT be near-duplicate.
	for _, r := range rows {
		if r.PathID == 2 && r.NearDuplicate {
			t.Error("path 2 should NOT be near-duplicate")
		}
	}

	// Winning path should be marked.
	for _, r := range rows {
		if r.PathID == 1 && !r.Winning {
			t.Error("path 1 should be marked as winning")
		}
		if r.PathID == 2 && r.Winning {
			t.Error("path 2 should NOT be marked as winning")
		}
	}
}

func TestWriteDiversityCSV_WritesFile(t *testing.T) {
	rows := []inrc2.DiversityRow{
		{
			RunID: "run1", Instance: "test", Seed: 42, BeamWidth: 5,
			Iterations: 500000, Temperature: 100.0, CoolingMode: "adaptive", Timestamp: "2025-01-01T00:00:00Z",
			Week: 1, PathID: 1, Fingerprint: "aabbccddee00",
			HammingToBest: 0.0, HammingToParent: 0.12, BeamSpread: 50,
			NearDuplicate: false, Retained: true, RetainedRank: 1, Winning: true,
			CumulativePenalty: 100, WeekPenalty: 100,
		},
		{
			RunID: "run1", Instance: "test", Seed: 42, BeamWidth: 5,
			Iterations: 500000, Temperature: 100.0, CoolingMode: "adaptive", Timestamp: "2025-01-01T00:00:00Z",
			Week: 1, PathID: 2, Fingerprint: "112233445566",
			HammingToBest: 0.03, HammingToParent: 0.08, BeamSpread: 50,
			NearDuplicate: true, Retained: true, RetainedRank: 2, Winning: false,
			CumulativePenalty: 150, WeekPenalty: 150,
		},
	}

	dir := t.TempDir()
	path := filepath.Join(dir, "diversity.csv")

	err := inrc2.WriteDiversityCSV(path, rows)
	if err != nil {
		t.Fatalf("WriteDiversityCSV failed: %v", err)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read written file: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(content)), "\n")
	if len(lines) != 3 { // header + 2 data rows
		t.Fatalf("expected 3 lines, got %d", len(lines))
	}

	// Verify header.
	header := lines[0]
	if !strings.Contains(header, "fingerprint") {
		t.Error("header missing 'fingerprint' column")
	}
	if !strings.Contains(header, "hamming_to_best") {
		t.Error("header missing 'hamming_to_best' column")
	}
	if !strings.Contains(header, "beam_spread") {
		t.Error("header missing 'beam_spread' column")
	}
	if !strings.Contains(header, "near_duplicate") {
		t.Error("header missing 'near_duplicate' column")
	}

	// Verify near_duplicate flag in row 2 (path 2, NearDuplicate=true => 1).
	fields := strings.Split(lines[2], ",")
	// near_duplicate is at index 14 (0-indexed from header).
	headerFields := strings.Split(header, ",")
	ndIdx := -1
	for i, h := range headerFields {
		if h == "near_duplicate" {
			ndIdx = i
			break
		}
	}
	if ndIdx < 0 {
		t.Fatal("could not find near_duplicate column index")
	}
	if fields[ndIdx] != "1" {
		t.Errorf("expected near_duplicate=1 for row 2, got %s", fields[ndIdx])
	}
}

func TestDiversityCSVHeader_ColumnCount(t *testing.T) {
	header := inrc2.DiversityCSVHeader()
	cols := strings.Split(header, ",")
	if len(cols) != 20 {
		t.Errorf("expected 20 columns in diversity CSV header, got %d", len(cols))
	}
}

func TestBeamSpread_SingleRetained(t *testing.T) {
	// With only 1 retained path, beam spread should be 0.
	result := inrc2.BeamResult{
		WinningPath: []inrc2.BeamPath{
			{ID: 1, ParentID: 0, Week: 1},
		},
		AllPaths: []inrc2.BeamPath{
			{ID: 1, ParentID: 0, Week: 1, Seed: 42, WeekPenalty: 100, CumulativePenalty: 100, Fingerprint: "aabb"},
		},
		WeekSummaries: []inrc2.BeamWeekSummary{
			{Week: 1, Candidates: 1, Retained: 1, BestCumulative: 100},
		},
	}

	sc := inrc2.Scenario{Nurses: []inrc2.Nurse{{ID: "A"}}}
	ctx := inrc2.RunContext{RunID: "test", Instance: "test", Seed: 42}

	rows := inrc2.BuildDiversityRows(ctx, result, sc)
	if len(rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(rows))
	}
	if rows[0].BeamSpread != 0 {
		t.Errorf("single retained path should have beam spread 0, got %d", rows[0].BeamSpread)
	}
}
