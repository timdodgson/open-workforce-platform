package ilp_test

import (
	"testing"
	"time"

	"github.com/timdodgson/open-workforce-platform/platform/go/internal/infrastructure/ilp"
)

func TestHighsSolver_Available(t *testing.T) {
	solver := &ilp.HighsSolver{}
	available := solver.Available()
	t.Logf("HiGHS solver available: %v", available)
}

func TestHighsSolver_Name(t *testing.T) {
	solver := &ilp.HighsSolver{}
	if solver.Name() != "HiGHS" {
		t.Errorf("expected name 'HiGHS', got '%s'", solver.Name())
	}
}

func TestHighsSolver_Parallel(t *testing.T) {
	solver := &ilp.HighsSolver{Parallel: true}
	if !solver.Parallel {
		t.Error("expected parallel to be true")
	}
}

func TestParseNativeOutput_Optimal(t *testing.T) {
	output := `Running HiGHS 1.15.1 (git hash: 04024d7)
Solving MIP model with:
   450 rows
   379 cols
   Thread count 16 (of 32 threads). Using 28 max workers. Parallel search on
Optimal - objective = 205.0
Solving report
  Status            Optimal
  Primal bound      205
  Dual bound        205
  Solution status   feasible
  Timing            2.1 (total)
`
	result := ilp.ParseNativeOutputForTest(output, 2100*time.Millisecond)

	if result.Status != "OPTIMAL" {
		t.Errorf("expected status OPTIMAL, got %s", result.Status)
	}
	if result.Objective != 205.0 {
		t.Errorf("expected objective 205, got %f", result.Objective)
	}
	if result.LowerBound != 205.0 {
		t.Errorf("expected lower bound 205, got %f", result.LowerBound)
	}
}

func TestParseNativeOutput_TimeLimitFeasible(t *testing.T) {
	output := `Running HiGHS 1.15.1
Solving report
  Status            Time limit reached
  Primal bound      3020
  Dual bound        1914
  Solution status   feasible
  Timing            14400.0 (total)
`
	result := ilp.ParseNativeOutputForTest(output, 14400*time.Second)

	if result.Status != "FEASIBLE" {
		t.Errorf("expected status FEASIBLE, got %s", result.Status)
	}
	if result.Objective != 3020.0 {
		t.Errorf("expected objective 3020, got %f", result.Objective)
	}
	if result.LowerBound != 1914.0 {
		t.Errorf("expected lower bound 1914, got %f", result.LowerBound)
	}
}

func TestParseProgressLine(t *testing.T) {
	line := " B       100      50       75  12.3%   1500.0          3000.0       50.00%    123   45   67    12345    60.5s"
	pt, ok := ilp.ParseProgressLineForTest(line)
	if !ok {
		t.Fatal("expected progress line to parse")
	}
	if pt.Nodes != 100 {
		t.Errorf("expected 100 nodes, got %d", pt.Nodes)
	}
	if pt.Elapsed != 60.5 {
		t.Errorf("expected elapsed 60.5, got %f", pt.Elapsed)
	}
}
