package cvrp

import (
	"fmt"
	"math"
	"sort"
	"testing"

	"github.com/timdodgson/open-workforce-platform/platform/go/internal/optimisation"
)

// TestCVRP_ValidationExperiment runs the full validation experiment pack.
// 4 modes × 5 seeds = 20 runs on A-n10-k2.
// Produces a summary report.
func TestCVRP_ValidationExperiment(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping validation experiment in short mode")
	}

	ds, err := LoadDataset("testdata/A-n10-k2.vrp")
	if err != nil {
		t.Fatalf("LoadDataset: %v", err)
	}

	problem := NewCVRPProblem(ds)
	baselineSol, _ := problem.CreateInitialSolution()
	baselineCost := problem.Evaluate(baselineSol)

	modes := []string{"sa", "lahc", "tabu", "portfolio"}
	seeds := []int64{42, 123, 555, 777, 999}
	iterations := 500000

	type RunResult struct {
		Mode       string
		Seed       int64
		Best       int
		Initial    int
		Candidates int
		RuntimeMs  int64
		Feasible   bool
	}

	var results []RunResult

	for _, mode := range modes {
		for _, seed := range seeds {
			config := optimisation.SearchConfig{
				Mode:                 mode,
				Iterations:           iterations,
				InitialTemperature:   100.0,
				MinTemperature:       0.0001,
				CoolingMode:          "adaptive",
				LateAcceptanceLength: 1000,
				TabuTenure:           7,
				Portfolio:            []string{"sa", "lahc", "tabu"},
				Seed:                 seed,
			}

			result := optimisation.RunSearch(problem, config)

			// Validate feasibility.
			bestSol := result.BestSolution.(*cvrpSolution)
			feasible, _ := problem.ValidateFull(bestSol)

			results = append(results, RunResult{
				Mode:       mode,
				Seed:       seed,
				Best:       result.BestPenalty,
				Initial:    result.InitialPenalty,
				Candidates: result.Candidates,
				RuntimeMs:  result.DurationMs,
				Feasible:   feasible,
			})
		}
	}

	// Generate report.
	t.Log("")
	t.Log("╔══════════════════════════════════════════════════════════════╗")
	t.Log("║          CVRP VALIDATION EXPERIMENT REPORT                  ║")
	t.Log("╠══════════════════════════════════════════════════════════════╣")
	t.Logf("║  Instance:    A-n10-k2 (10 customers, capacity 50)         ║")
	t.Logf("║  Baseline:    %d (nearest-neighbour constructive)           ║", baselineCost)
	t.Logf("║  Iterations:  %dK per run                                  ║", iterations/1000)
	t.Logf("║  Seeds:       42, 123, 555, 777, 999                       ║")
	t.Log("╚══════════════════════════════════════════════════════════════╝")
	t.Log("")

	// Per-mode statistics.
	t.Log("┌──────────┬──────┬──────┬──────┬───────┬──────────┬──────────┐")
	t.Log("│ Mode     │ Best │ Mean │Worst │ StdDev│ Avg(ms)  │ Feasible │")
	t.Log("├──────────┼──────┼──────┼──────┼───────┼──────────┼──────────┤")

	type ModeStats struct {
		Mode     string
		Best     int
		Mean     float64
		Worst    int
		StdDev   float64
		AvgMs    int64
		AllFeas  bool
		Values   []int
	}

	var modeStats []ModeStats

	for _, mode := range modes {
		var penalties []int
		var runtimes []int64
		allFeasible := true
		for _, r := range results {
			if r.Mode == mode {
				penalties = append(penalties, r.Best)
				runtimes = append(runtimes, r.RuntimeMs)
				if !r.Feasible {
					allFeasible = false
				}
			}
		}

		sort.Ints(penalties)
		best := penalties[0]
		worst := penalties[len(penalties)-1]
		sum := 0
		for _, p := range penalties {
			sum += p
		}
		mean := float64(sum) / float64(len(penalties))

		variance := 0.0
		for _, p := range penalties {
			variance += (float64(p) - mean) * (float64(p) - mean)
		}
		stddev := math.Sqrt(variance / float64(len(penalties)-1))

		var totalMs int64
		for _, ms := range runtimes {
			totalMs += ms
		}
		avgMs := totalMs / int64(len(runtimes))

		feasStr := "✓ ALL"
		if !allFeasible {
			feasStr = "✗ FAIL"
		}

		t.Logf("│ %-8s │ %4d │ %4.0f │ %4d │ %5.1f │ %6dms │ %8s │",
			mode, best, mean, worst, stddev, avgMs, feasStr)

		modeStats = append(modeStats, ModeStats{
			Mode: mode, Best: best, Mean: mean, Worst: worst,
			StdDev: stddev, AvgMs: avgMs, AllFeas: allFeasible, Values: penalties,
		})
	}

	t.Log("└──────────┴──────┴──────┴──────┴───────┴──────────┴──────────┘")
	t.Log("")

	// Overall winner.
	sort.Slice(modeStats, func(i, j int) bool { return modeStats[i].Mean < modeStats[j].Mean })
	t.Logf("🏆 Winner (lowest mean): %s (mean=%.0f, best=%d)",
		modeStats[0].Mode, modeStats[0].Mean, modeStats[0].Best)
	t.Logf("   Improvement over baseline: %.1f%%",
		float64(baselineCost-modeStats[0].Best)/float64(baselineCost)*100)
	t.Log("")

	// Per-run detail.
	t.Log("Per-run detail:")
	t.Log("┌──────────┬──────┬──────┬────────┬──────────┬──────────┐")
	t.Log("│ Mode     │ Seed │ Best │ Cands  │ Runtime  │ Feasible │")
	t.Log("├──────────┼──────┼──────┼────────┼──────────┼──────────┤")
	for _, r := range results {
		feasStr := "✓"
		if !r.Feasible {
			feasStr = "✗"
		}
		t.Logf("│ %-8s │ %4d │ %4d │ %5dK │ %6dms │    %s     │",
			r.Mode, r.Seed, r.Best, r.Candidates/1000, r.RuntimeMs, feasStr)
	}
	t.Log("└──────────┴──────┴──────┴────────┴──────────┴──────────┘")

	// Assertions.
	for _, r := range results {
		if !r.Feasible {
			t.Errorf("%s seed %d: INFEASIBLE", r.Mode, r.Seed)
		}
		if r.Best >= baselineCost {
			t.Errorf("%s seed %d: no improvement (best=%d, baseline=%d)", r.Mode, r.Seed, r.Best, baselineCost)
		}
	}

	// Write markdown report.
	t.Log("")
	t.Log("--- MARKDOWN REPORT ---")
	t.Log("")
	t.Log("## CVRP Validation Experiment Results")
	t.Log("")
	t.Logf("**Instance:** A-n10-k2 (10 customers, capacity 50)")
	t.Logf("**Baseline:** %d (nearest-neighbour)", baselineCost)
	t.Logf("**Iterations:** %dK × 5 seeds × 4 modes = 20 runs", iterations/1000)
	t.Log("")
	t.Log("| Mode | Best | Mean | Worst | StdDev | Avg Runtime | Feasible |")
	t.Log("|------|------|------|-------|--------|-------------|----------|")
	for _, ms := range modeStats {
		feas := "✓"
		if !ms.AllFeas {
			feas = "✗"
		}
		t.Logf("| %s | %d | %.0f | %d | %.1f | %dms | %s |",
			ms.Mode, ms.Best, ms.Mean, ms.Worst, ms.StdDev, ms.AvgMs, feas)
	}
	t.Log("")
	t.Logf("**Winner:** %s (mean=%.0f, best=%d, improvement=%.1f%%)",
		modeStats[0].Mode, modeStats[0].Mean, modeStats[0].Best,
		float64(baselineCost-modeStats[0].Best)/float64(baselineCost)*100)
	_ = fmt.Sprint("") // suppress unused import
}
