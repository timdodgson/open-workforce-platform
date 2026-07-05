package jobshop

import (
	"encoding/json"
	"math/rand"
	"testing"

	"github.com/timdodgson/open-workforce-platform/platform/go/internal/optimisation"
)

const testInstance = "testdata/ft06.txt"

func TestLoadDataset(t *testing.T) {
	ds, err := LoadDataset(testInstance)
	if err != nil {
		t.Fatalf("LoadDataset: %v", err)
	}
	if ds.Jobs != 6 {
		t.Errorf("Jobs = %d, want 6", ds.Jobs)
	}
	if ds.Machines != 6 {
		t.Errorf("Machines = %d, want 6", ds.Machines)
	}
	if len(ds.AllOps) != 36 {
		t.Errorf("AllOps = %d, want 36 (6x6)", len(ds.AllOps))
	}
	if len(ds.JobList) != 6 {
		t.Errorf("JobList = %d, want 6", len(ds.JobList))
	}
	// First job, first op: machine 2, duration 1.
	if ds.JobList[0].Operations[0].Machine != 2 || ds.JobList[0].Operations[0].Duration != 1 {
		t.Errorf("Job 0 Op 0: got machine=%d duration=%d, want machine=2 duration=1",
			ds.JobList[0].Operations[0].Machine, ds.JobList[0].Operations[0].Duration)
	}
}

func TestBuildInitialSchedule(t *testing.T) {
	ds, _ := LoadDataset(testInstance)
	sched := BuildInitialSchedule(ds)

	if sched.Makespan <= 0 {
		t.Error("Makespan should be positive")
	}
	if len(sched.Ops) != 36 {
		t.Errorf("Scheduled ops = %d, want 36", len(sched.Ops))
	}

	// Validate.
	violations := Validate(ds, sched)
	if len(violations) > 0 {
		for _, v := range violations {
			t.Errorf("Violation: [%s] %s", v.Code, v.Detail)
		}
	}

	t.Logf("ft06: constructive makespan = %d (optimal = 55)", sched.Makespan)
}

func TestJSSProblem_ImplementsInterface(t *testing.T) {
	var _ optimisation.Problem = (*JSSProblem)(nil)
}

func TestJSSProblem_CreateAndEvaluate(t *testing.T) {
	ds, _ := LoadDataset(testInstance)
	problem := NewJSSProblem(ds)

	sol, err := problem.CreateInitialSolution()
	if err != nil {
		t.Fatalf("CreateInitialSolution: %v", err)
	}

	makespan := problem.Evaluate(sol)
	if makespan <= 0 {
		t.Error("Makespan should be positive")
	}
	t.Logf("Initial makespan: %d", makespan)
}

func TestJSSProblem_TryMoveAndUndo(t *testing.T) {
	ds, _ := LoadDataset(testInstance)
	problem := NewJSSProblem(ds)
	sol, _ := problem.CreateInitialSolution()
	rng := rand.New(rand.NewSource(42))

	fpBefore := problem.SolutionFingerprint(sol)
	costBefore := problem.Evaluate(sol)

	var validMove optimisation.MoveResult
	for i := 0; i < 100; i++ {
		result := problem.TryMove(sol, rng)
		if result.Valid {
			validMove = result
			break
		}
	}
	if !validMove.Valid {
		t.Fatal("No valid move found")
	}

	// Undo.
	problem.UndoMove(sol, validMove.Move)

	fpAfter := problem.SolutionFingerprint(sol)
	costAfter := problem.Evaluate(sol)

	if fpAfter != fpBefore {
		t.Errorf("Fingerprint mismatch after undo: %s != %s", fpAfter, fpBefore)
	}
	if costAfter != costBefore {
		t.Errorf("Cost mismatch after undo: %d != %d", costAfter, costBefore)
	}
}

func TestJSSProblem_SearchLoop(t *testing.T) {
	ds, _ := LoadDataset(testInstance)
	problem := NewJSSProblem(ds)

	config := optimisation.SearchConfig{
		Mode:               "sa",
		Iterations:         50000,
		InitialTemperature: 100.0,
		MinTemperature:     0.001,
		CoolingMode:        "adaptive",
		Seed:               42,
	}

	result := optimisation.RunSearch(problem, config)

	if result.BestPenalty >= result.InitialPenalty {
		t.Errorf("SA should improve: best=%d, initial=%d", result.BestPenalty, result.InitialPenalty)
	}

	t.Logf("ft06: initial=%d, SA best=%d (optimal=55), gap=+%.1f%%",
		result.InitialPenalty, result.BestPenalty,
		float64(result.BestPenalty-55)/55*100)
}

func TestJSSProblem_AllModes(t *testing.T) {
	ds, _ := LoadDataset(testInstance)
	problem := NewJSSProblem(ds)

	modes := []string{"sa", "lahc", "adaptive"}
	for _, mode := range modes {
		config := optimisation.SearchConfig{
			Mode:                 mode,
			Iterations:           20000,
			InitialTemperature:   100.0,
			MinTemperature:       0.001,
			CoolingMode:          "adaptive",
			LateAcceptanceLength: 500,
			Seed:                 42,
		}
		result := optimisation.RunSearch(problem, config)
		if result.BestPenalty >= result.InitialPenalty {
			t.Errorf("%s: no improvement (best=%d, initial=%d)", mode, result.BestPenalty, result.InitialPenalty)
		}
		t.Logf("  %s: best=%d", mode, result.BestPenalty)
	}
}

func TestJSSProblem_Deterministic(t *testing.T) {
	ds, _ := LoadDataset(testInstance)
	problem := NewJSSProblem(ds)

	config := optimisation.SearchConfig{
		Mode: "sa", Iterations: 10000,
		InitialTemperature: 100, MinTemperature: 0.001, CoolingMode: "adaptive", Seed: 99,
	}

	r1 := optimisation.RunSearch(problem, config)
	r2 := optimisation.RunSearch(problem, config)

	if r1.BestPenalty != r2.BestPenalty {
		t.Errorf("Not deterministic: %d != %d", r1.BestPenalty, r2.BestPenalty)
	}
}

func TestJSSProblem_SerializeSolution(t *testing.T) {
	ds, _ := LoadDataset(testInstance)
	problem := NewJSSProblem(ds)
	sol, _ := problem.CreateInitialSolution()

	data, err := problem.SerializeSolution(sol)
	if err != nil {
		t.Fatalf("SerializeSolution: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("Empty serialization")
	}

	var parsed struct {
		Makespan int `json:"makespan"`
		Jobs     int `json:"jobs"`
		Machines int `json:"machines"`
	}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("Invalid JSON: %v", err)
	}
	if parsed.Makespan <= 0 {
		t.Error("Makespan should be positive in serialized output")
	}
	if parsed.Jobs != 6 || parsed.Machines != 6 {
		t.Errorf("Jobs=%d Machines=%d, want 6/6", parsed.Jobs, parsed.Machines)
	}
}
