package cvrp

import (
	"strings"
	"testing"

	"github.com/timdodgson/open-workforce-platform/platform/go/internal/optimisation"
)

func TestResultsCSVHeaderUnchanged(t *testing.T) {
	if !strings.HasSuffix(ResultsCSVHeader, "\n") {
		t.Fatal("results header must end with newline")
	}
	data := BuildResultsCSV(ResultsCSVParams{
		Instance: "A-n32-k5", Seed: 42, WinnerMode: "sa", Iterations: 5000,
		Temperature: 100.0,
		SearchResult: optimisation.SearchResult{
			InitialPenalty: 1000, BestPenalty: 800, Candidates: 100,
			Accepted: 50, Rejected: 50, DurationMs: 84,
		},
	})
	if !strings.HasPrefix(string(data), ResultsCSVHeader) {
		t.Fatal("results CSV must start with canonical header")
	}
}

func TestDiscoveriesCSVHeaderUnchanged(t *testing.T) {
	data := BuildDiscoveriesCSV(DiscoveriesCSVParams{
		RunLabel: "test-run", Instance: "A-n32-k5", Seed: 42,
		Iterations: 5000, Temperature: 100.0,
		Discoveries: []optimisation.Discovery{
			{ElapsedMs: 10, Candidate: 100, OldBest: 900, NewBest: 850, Improvement: 50},
		},
	})
	if !strings.HasPrefix(string(data), DiscoveriesCSVHeader) {
		t.Fatal("discoveries CSV must start with canonical header")
	}
}
