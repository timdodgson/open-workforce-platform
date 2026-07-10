package runoutput_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/timdodgson/open-workforce-platform/platform/go/internal/infrastructure/runoutput"
)

func TestWriteILPBenchmark(t *testing.T) {
	dir := t.TempDir()
	meta := map[string]interface{}{
		"problemType": "cvrp",
		"mode":        "ilp",
		"runLabel":    "test-run",
	}
	bench := map[string]interface{}{"status": "OPTIMAL", "objective": 100}
	solution := []byte(`{"routes":[]}`)

	if err := runoutput.WriteILPBenchmark(dir, meta, bench, solution); err != nil {
		t.Fatalf("WriteILPBenchmark: %v", err)
	}

	for _, name := range []string{"run.json", "ilp-benchmark.json", "solution.json"} {
		if _, err := os.Stat(filepath.Join(dir, name)); err != nil {
			t.Fatalf("missing %s: %v", name, err)
		}
	}

	runJSON, _ := os.ReadFile(filepath.Join(dir, "run.json"))
	if !strings.Contains(string(runJSON), `"problemType": "cvrp"`) {
		t.Fatalf("unexpected run.json: %s", runJSON)
	}
}
