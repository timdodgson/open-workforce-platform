package workerlearning

import (
	"os"
	"strings"
	"testing"
)

func TestEmitSingleWorkerLearning(t *testing.T) {
	dir := t.TempDir()
	err := EmitSingleWorkerLearning(dir, SingleWorkerConfig{
		ProblemType: "cvrp",
		Instance:    "A-n32-k5",
		Algorithm:   "sa",
		Seed:        42,
		Temperature: 100,
		Iterations:  500000,
	}, SearchOutcome{
		InitialPenalty: 1000,
		BestPenalty:    850,
		DurationMs:     120,
		Candidates:     500000,
		Accepted:       12000,
		Rejected:       480000,
	})
	if err != nil {
		t.Fatalf("EmitSingleWorkerLearning: %v", err)
	}

	data, err := os.ReadFile(dir + "/worker_learning.csv")
	if err != nil {
		t.Fatalf("read csv: %v", err)
	}
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != 2 {
		t.Fatalf("got %d lines, want header + 1 row", len(lines))
	}
	if !strings.Contains(lines[1], "cvrp") {
		t.Fatalf("row missing problem type: %s", lines[1])
	}
	if !strings.Contains(lines[1], ",150,") {
		t.Fatalf("row missing improvement_amount 150: %s", lines[1])
	}
}
