package inrc2_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/timdodgson/open-workforce-platform/platform/go/internal/infrastructure/inrc2"
)

func TestBuildBeamTreeRows_MarksWinningAndRetained(t *testing.T) {
	result := inrc2.BeamResult{
		WinningPath: []inrc2.BeamPath{
			{ID: 1, ParentID: 0, Week: 1, Seed: 42, WeekPenalty: 100, CumulativePenalty: 100},
			{ID: 5, ParentID: 1, Week: 2, Seed: 101, WeekPenalty: 80, CumulativePenalty: 180},
		},
		AllPaths: []inrc2.BeamPath{
			{ID: 1, ParentID: 0, Week: 1, Seed: 42, WeekPenalty: 100, CumulativePenalty: 100,
				Stats: inrc2.PFRSStats{WorkersStarted: 5, CandidatesEvaluated: 1000}},
			{ID: 2, ParentID: 0, Week: 1, Seed: 101, WeekPenalty: 120, CumulativePenalty: 120,
				Stats: inrc2.PFRSStats{WorkersStarted: 5, CandidatesEvaluated: 1000}},
			{ID: 3, ParentID: 0, Week: 1, Seed: 202, WeekPenalty: 150, CumulativePenalty: 150,
				Stats: inrc2.PFRSStats{WorkersStarted: 5, CandidatesEvaluated: 1000}},
			{ID: 5, ParentID: 1, Week: 2, Seed: 101, WeekPenalty: 80, CumulativePenalty: 180,
				Stats: inrc2.PFRSStats{WorkersStarted: 5, CandidatesEvaluated: 1000}},
			{ID: 6, ParentID: 2, Week: 2, Seed: 42, WeekPenalty: 110, CumulativePenalty: 230,
				Stats: inrc2.PFRSStats{WorkersStarted: 5, CandidatesEvaluated: 1000}},
		},
		WeekSummaries: []inrc2.BeamWeekSummary{
			{Week: 1, Candidates: 3, Retained: 2, BestCumulative: 100},
			{Week: 2, Candidates: 2, Retained: 2, BestCumulative: 180},
		},
	}

	rows := inrc2.BuildBeamTreeRows(result)
	if len(rows) != 5 {
		t.Fatalf("expected 5 rows, got %d", len(rows))
	}

	// Path 1: winning + retained rank 1.
	r1 := findRow(rows, 1)
	if !r1.Winning {
		t.Error("path 1 should be winning")
	}
	if !r1.Retained || r1.RetainedRank != 1 {
		t.Errorf("path 1: retained=%v rank=%d, want retained=true rank=1", r1.Retained, r1.RetainedRank)
	}

	// Path 3: NOT retained (beam width 2, it's rank 3).
	r3 := findRow(rows, 3)
	if r3.Retained {
		t.Error("path 3 should NOT be retained")
	}

	// Path 5: winning + retained.
	r5 := findRow(rows, 5)
	if !r5.Winning {
		t.Error("path 5 should be winning")
	}
	if !r5.Retained {
		t.Error("path 5 should be retained")
	}
}

func TestWriteBeamTreeCSV_WritesFile(t *testing.T) {
	result := inrc2.BeamResult{
		WinningPath: []inrc2.BeamPath{{ID: 1, Week: 1}},
		AllPaths:    []inrc2.BeamPath{{ID: 1, ParentID: 0, Week: 1, Seed: 42, WeekPenalty: 100, CumulativePenalty: 100}},
		WeekSummaries: []inrc2.BeamWeekSummary{{Week: 1, Candidates: 1, Retained: 1, BestCumulative: 100}},
	}

	dir := t.TempDir()
	path := filepath.Join(dir, "tree.csv")

	err := inrc2.WriteBeamTreeCSV(path, result)
	if err != nil {
		t.Fatalf("WriteBeamTreeCSV failed: %v", err)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read tree.csv: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(content)), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines (header + 1 row), got %d", len(lines))
	}

	// Header should start with path_id.
	if !strings.HasPrefix(lines[0], "path_id,") {
		t.Errorf("header doesn't start with path_id: %s", lines[0])
	}

	// Row should start with "1,0,1,42" (pathID, parentID, week, seed).
	if !strings.HasPrefix(lines[1], "1,0,1,42,") {
		t.Errorf("unexpected row: %s", lines[1])
	}
}

func TestBeamSearch_ProducesTreeData(t *testing.T) {
	// Integration test: run a real beam search and verify tree CSV output.
	const n005Dir = "../../../../../examples/inrc2/testdatasets_json/n005w4/"

	sc, err := inrc2.LoadScenario(n005Dir + "Sc-n005w4.json")
	if err != nil {
		t.Skipf("test data not available: %v", err)
	}

	weekFiles := []string{
		n005Dir + "WD-n005w4-0.json",
		n005Dir + "WD-n005w4-1.json",
	}
	hist, _ := inrc2.LoadHistory(n005Dir + "H0-n005w4-0.json")

	config := inrc2.DefaultPFRSConfig()
	config.IterationsPerWorker = 5000
	config.MaxConcurrentWorkers = 2
	config.MaxTotalWorkers = 2
	config.BranchOnGlobalBest = false
	config.CoolingMode = "adaptive"

	beam := inrc2.BeamConfig{BeamWidth: 2, Seeds: []int64{42, 101}}

	result, err := inrc2.RunBeamSearch(sc, weekFiles, hist, config, beam, nil)
	if err != nil {
		t.Fatalf("beam search failed: %v", err)
	}

	// Must have paths for both weeks.
	if len(result.AllPaths) == 0 {
		t.Fatal("beam search produced no paths")
	}
	if len(result.WinningPath) == 0 {
		t.Fatal("beam search has no winning path")
	}

	// Write tree CSV and verify it's parseable.
	dir := t.TempDir()
	path := filepath.Join(dir, "tree.csv")
	if err := inrc2.WriteBeamTreeCSV(path, result); err != nil {
		t.Fatalf("WriteBeamTreeCSV failed: %v", err)
	}

	rows := inrc2.BuildBeamTreeRows(result)

	// At least one row should be winning.
	hasWinner := false
	for _, r := range rows {
		if r.Winning {
			hasWinner = true
			break
		}
	}
	if !hasWinner {
		t.Error("no winning path marked in tree rows")
	}

	t.Logf("Beam search: %d total paths, %d winning lineage, tree.csv has %d rows",
		len(result.AllPaths), len(result.WinningPath), len(rows))
}

func findRow(rows []inrc2.BeamTreeRow, pathID int) inrc2.BeamTreeRow {
	for _, r := range rows {
		if r.PathID == pathID {
			return r
		}
	}
	return inrc2.BeamTreeRow{}
}
