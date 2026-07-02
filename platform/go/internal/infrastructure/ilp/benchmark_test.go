package ilp_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/timdodgson/open-workforce-platform/platform/go/internal/infrastructure/ilp"
)

func TestCompare_PositiveGap(t *testing.T) {
	ilpResult := ilp.BenchmarkResult{
		Instance:  "n005w4",
		Weeks:     1,
		Objective: 100,
		Status:    "OPTIMAL",
	}

	comparison := ilp.Compare(ilpResult, 110, 0.5)

	if comparison.AbsoluteGap != 10 {
		t.Errorf("expected absolute gap 10, got %d", comparison.AbsoluteGap)
	}
	if comparison.GapPercent != 10.0 {
		t.Errorf("expected gap 10%%, got %.1f%%", comparison.GapPercent)
	}
	if comparison.PFRSPenalty != 110 {
		t.Errorf("expected PFRS penalty 110, got %d", comparison.PFRSPenalty)
	}
	if comparison.PFRSRuntime != 0.5 {
		t.Errorf("expected PFRS runtime 0.5, got %f", comparison.PFRSRuntime)
	}
}

func TestCompare_ZeroGap(t *testing.T) {
	ilpResult := ilp.BenchmarkResult{
		Instance:  "n005w4",
		Weeks:     1,
		Objective: 200,
		Status:    "OPTIMAL",
	}

	comparison := ilp.Compare(ilpResult, 200, 1.0)

	if comparison.AbsoluteGap != 0 {
		t.Errorf("expected zero gap, got %d", comparison.AbsoluteGap)
	}
	if comparison.GapPercent != 0.0 {
		t.Errorf("expected 0%% gap, got %.1f%%", comparison.GapPercent)
	}
}

func TestCompare_NegativeGap(t *testing.T) {
	// PFRS found a better solution than ILP (possible if ILP timed out).
	ilpResult := ilp.BenchmarkResult{
		Instance:  "n005w4",
		Weeks:     1,
		Objective: 150,
		Status:    "FEASIBLE",
	}

	comparison := ilp.Compare(ilpResult, 130, 2.0)

	if comparison.AbsoluteGap != -20 {
		t.Errorf("expected gap -20, got %d", comparison.AbsoluteGap)
	}
}

func TestLoadBenchmarkResult_RoundTrip(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "test-result.json")

	data := `{
  "instance": "n005w4",
  "weeks": 2,
  "solver": "HiGHS",
  "status": "OPTIMAL",
  "objective": 625,
  "lowerBound": 625,
  "gapPercent": 0,
  "runtimeSeconds": 12.4,
  "timeLimit": 300,
  "hardViolations": 0
}`
	if err := os.WriteFile(path, []byte(data), 0644); err != nil {
		t.Fatal(err)
	}

	loaded, err := ilp.LoadBenchmarkResult(path)
	if err != nil {
		t.Fatalf("LoadBenchmarkResult failed: %v", err)
	}

	if loaded.Instance != "n005w4" {
		t.Errorf("Instance mismatch: got %s", loaded.Instance)
	}
	if loaded.Objective != 625 {
		t.Errorf("Objective mismatch: got %d", loaded.Objective)
	}
	if loaded.Status != "OPTIMAL" {
		t.Errorf("Status mismatch: got %s", loaded.Status)
	}
	if loaded.Weeks != 2 {
		t.Errorf("Weeks mismatch: got %d", loaded.Weeks)
	}
}
