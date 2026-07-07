package main

import (
	"strings"
	"testing"

	"github.com/timdodgson/open-workforce-platform/platform/go/internal/optimisation"
)

func TestCVRPResultsCSVHeaderUnchanged(t *testing.T) {
	if !strings.HasSuffix(cvrpResultsCSVHeader, "\n") {
		t.Fatal("results header must end with newline")
	}
	data := buildCVRPResultsCSV(cvrpResultsCSVParams{
		Instance: "A-n32-k5", Seed: 42, WinnerMode: "sa", Iterations: 5000,
		Temperature: 100.0,
		SearchResult: optimisation.SearchResult{
			InitialPenalty: 1000, BestPenalty: 800, Candidates: 100,
			Accepted: 50, Rejected: 50, DurationMs: 84,
		},
	})
	if !strings.HasPrefix(string(data), cvrpResultsCSVHeader) {
		t.Fatal("results CSV must start with canonical header")
	}
}

func TestCVRPDiscoveriesCSVHeaderUnchanged(t *testing.T) {
	data := buildCVRPDiscoveriesCSV(cvrpDiscoveriesCSVParams{
		RunLabel: "test-run", Instance: "A-n32-k5", Seed: 42,
		Iterations: 5000, Temperature: 100.0,
		Discoveries: []optimisation.Discovery{
			{ElapsedMs: 10, Candidate: 100, OldBest: 900, NewBest: 850, Improvement: 50},
		},
	})
	if !strings.HasPrefix(string(data), cvrpDiscoveriesCSVHeader) {
		t.Fatal("discoveries CSV must start with canonical header")
	}
}

func TestVRPTWDiscoveriesCSVHeaderUnchanged(t *testing.T) {
	data := buildVRPTWDiscoveriesCSV([]optimisation.Discovery{
		{ElapsedMs: 1, Candidate: 2, OldBest: 3, NewBest: 2, Improvement: 1},
	})
	if string(data) != "elapsed_ms,candidate,old_best,new_best,improvement\n1,2,3,2,1\n" {
		t.Fatalf("unexpected VRPTW discoveries CSV:\n%s", data)
	}
}

func TestPFRSStandardRunJSONFormat(t *testing.T) {
	got := formatPFRSStandardRunJSON(pfrsStandardRunJSONParams{
		InstanceName: "n012w8", WorkerMode: "sa", BestPenalty: 3465, RunLabel: "my-run",
	})
	want := `{
  "instance": "n012w8",
  "problemType": "nrp",
  "mode": "sa",
  "bestObjective": 3465,
  "totalPenalty": 3465,
  "runLabel": "my-run"
}`
	if got != want {
		t.Fatalf("PFRS standard run.json format changed:\ngot:\n%s\nwant:\n%s", got, want)
	}
}
