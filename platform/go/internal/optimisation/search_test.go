package optimisation

import (
	"math/rand"
	"testing"
)

// mockProblem is a trivial Problem implementation for testing the search engine.
// The "solution" is a single integer. Moves increment or decrement by 1.
// Optimal solution is 0.
type mockProblem struct{}
type mockSolution struct{ value int }
type mockMove struct{ delta int }

func (p *mockProblem) CreateInitialSolution() (Solution, error) {
	return &mockSolution{value: 100}, nil
}

func (p *mockProblem) CloneSolution(s Solution) Solution {
	ms := s.(*mockSolution)
	return &mockSolution{value: ms.value}
}

func (p *mockProblem) Evaluate(s Solution) int {
	ms := s.(*mockSolution)
	if ms.value < 0 {
		return -ms.value // penalty for negative
	}
	return ms.value
}

func (p *mockProblem) TryMove(s Solution, rng *rand.Rand) MoveResult {
	ms := s.(*mockSolution)
	delta := rng.Intn(11) - 5 // -5 to +5
	ms.value += delta
	return MoveResult{Valid: true, Move: mockMove{delta: delta}}
}

func (p *mockProblem) UndoMove(s Solution, m Move) {
	ms := s.(*mockSolution)
	mm := m.(mockMove)
	ms.value -= mm.delta
}

func (p *mockProblem) SolutionFingerprint(s Solution) string {
	return ""
}

func (p *mockProblem) SerializeSolution(s Solution) ([]byte, error) {
	return nil, nil
}

// TestRunSearch_SA_ImprovesFromInitial verifies SA finds better solutions.
func TestRunSearch_SA_ImprovesFromInitial(t *testing.T) {
	problem := &mockProblem{}
	config := DefaultSearchConfig()
	config.Iterations = 10000

	result := RunSearch(problem, config)

	if result.InitialPenalty != 100 {
		t.Errorf("InitialPenalty = %d, want 100", result.InitialPenalty)
	}
	if result.BestPenalty >= result.InitialPenalty {
		t.Errorf("SA should improve: best=%d, initial=%d", result.BestPenalty, result.InitialPenalty)
	}
	if result.Candidates != config.Iterations {
		t.Errorf("Candidates = %d, want %d", result.Candidates, config.Iterations)
	}
	t.Logf("SA: initial=%d, best=%d, improved %d times", result.InitialPenalty, result.BestPenalty, result.Improved)
}

// TestRunSearch_SA_Deterministic verifies same seed produces same result.
func TestRunSearch_SA_Deterministic(t *testing.T) {
	problem := &mockProblem{}
	config := DefaultSearchConfig()
	config.Iterations = 5000
	config.Seed = 99

	r1 := RunSearch(problem, config)
	r2 := RunSearch(problem, config)

	if r1.BestPenalty != r2.BestPenalty {
		t.Errorf("Not deterministic: run1=%d, run2=%d", r1.BestPenalty, r2.BestPenalty)
	}
	if r1.Accepted != r2.Accepted {
		t.Errorf("Not deterministic: accepted run1=%d, run2=%d", r1.Accepted, r2.Accepted)
	}
}

// TestRunSearch_SA_ReachesOptimum verifies SA can reach 0 on trivial problem.
func TestRunSearch_SA_ReachesOptimum(t *testing.T) {
	problem := &mockProblem{}
	config := DefaultSearchConfig()
	config.Iterations = 100000
	config.InitialTemperature = 50.0

	result := RunSearch(problem, config)

	if result.BestPenalty > 5 {
		t.Errorf("SA should reach near-optimal: best=%d (expected ≤5)", result.BestPenalty)
	}
}


// TestRunSearch_LAHC_ImprovesFromInitial verifies LAHC finds better solutions.
func TestRunSearch_LAHC_ImprovesFromInitial(t *testing.T) {
	problem := &mockProblem{}
	config := DefaultSearchConfig()
	config.Mode = "lahc"
	config.Iterations = 10000
	config.LateAcceptanceLength = 100

	result := RunSearch(problem, config)

	if result.BestPenalty >= result.InitialPenalty {
		t.Errorf("LAHC should improve: best=%d, initial=%d", result.BestPenalty, result.InitialPenalty)
	}
	t.Logf("LAHC: initial=%d, best=%d, improved %d times", result.InitialPenalty, result.BestPenalty, result.Improved)
}

// TestRunSearch_LAHC_Deterministic verifies same seed produces same result.
func TestRunSearch_LAHC_Deterministic(t *testing.T) {
	problem := &mockProblem{}
	config := DefaultSearchConfig()
	config.Mode = "lahc"
	config.Iterations = 5000
	config.LateAcceptanceLength = 50
	config.Seed = 99

	r1 := RunSearch(problem, config)
	r2 := RunSearch(problem, config)

	if r1.BestPenalty != r2.BestPenalty {
		t.Errorf("LAHC not deterministic: run1=%d, run2=%d", r1.BestPenalty, r2.BestPenalty)
	}
}


// TestRunSearch_Tabu_ImprovesFromInitial verifies Tabu finds better solutions.
func TestRunSearch_Tabu_ImprovesFromInitial(t *testing.T) {
	problem := &mockProblem{}
	config := DefaultSearchConfig()
	config.Mode = "tabu"
	config.Iterations = 10000
	config.TabuTenure = 5

	result := RunSearch(problem, config)

	if result.BestPenalty >= result.InitialPenalty {
		t.Errorf("Tabu should improve: best=%d, initial=%d", result.BestPenalty, result.InitialPenalty)
	}
	t.Logf("Tabu: initial=%d, best=%d, improved %d times", result.InitialPenalty, result.BestPenalty, result.Improved)
}

// TestRunSearch_Tabu_Deterministic verifies same seed produces same result.
func TestRunSearch_Tabu_Deterministic(t *testing.T) {
	problem := &mockProblem{}
	config := DefaultSearchConfig()
	config.Mode = "tabu"
	config.Iterations = 5000
	config.TabuTenure = 5
	config.Seed = 99

	r1 := RunSearch(problem, config)
	r2 := RunSearch(problem, config)

	if r1.BestPenalty != r2.BestPenalty {
		t.Errorf("Tabu not deterministic: run1=%d, run2=%d", r1.BestPenalty, r2.BestPenalty)
	}
}
