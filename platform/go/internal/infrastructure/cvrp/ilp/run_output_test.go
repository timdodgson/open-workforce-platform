package ilp_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/timdodgson/open-workforce-platform/platform/go/internal/infrastructure/cvrp"
	cvrpilp "github.com/timdodgson/open-workforce-platform/platform/go/internal/infrastructure/cvrp/ilp"
	"github.com/timdodgson/open-workforce-platform/platform/go/internal/infrastructure/runoutput"
)

func TestFinalizeBenchmark_writesArtifacts(t *testing.T) {
	ds := &cvrp.Dataset{Name: "smoke", Capacity: 100}
	result := cvrpilp.BenchmarkResult{
		Status:         "OPTIMAL",
		Objective:      784,
		LowerBound:     784,
		Vehicles:       5,
		Variables:      100,
		Constraints:    50,
		SolutionJSON:   []byte(`{"routes":[]}`),
		RuntimeSeconds: 1.5,
	}

	dir := t.TempDir()
	// FinalizeBenchmark uses runoutput.EnsureDir which writes under DefaultRunsRoot;
	// call WriteILPBenchmark directly via FinalizeBenchmark by using a temp label
	// and patching is hard — test the helper path with explicit dir via runoutput.
	label := filepath.Base(dir)
	_ = label

	outDir, err := cvrpilp.FinalizeBenchmark(ds, result, cvrpilp.FinalizeOptions{
		RunLabel:     "cvrp-ilp-smoke-test",
		TimeLimitSec: 60,
	})
	if err != nil {
		t.Fatalf("FinalizeBenchmark: %v", err)
	}
	if outDir == "" {
		t.Fatal("expected output dir")
	}
	t.Cleanup(func() { os.RemoveAll(outDir) })

	for _, name := range []string{"run.json", "ilp-benchmark.json", "solution.json"} {
		if _, err := os.Stat(filepath.Join(outDir, name)); err != nil {
			t.Fatalf("missing %s: %v", name, err)
		}
	}

	// Sanity: runoutput package writes the same artifact set.
	if err := runoutput.WriteILPBenchmark(t.TempDir(), map[string]interface{}{"mode": "ilp"}, result, result.SolutionJSON); err != nil {
		t.Fatalf("WriteILPBenchmark: %v", err)
	}
}
