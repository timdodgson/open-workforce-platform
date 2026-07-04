package cvrp

import (
	"encoding/json"
	"math/rand"
	"testing"

	"github.com/timdodgson/open-workforce-platform/platform/go/internal/optimisation"
)

const testInstancePath = "testdata/A-n10-k2.vrp"

// TestLoadDataset verifies CVRPLIB format parsing.
func TestLoadDataset(t *testing.T) {
	ds, err := LoadDataset(testInstancePath)
	if err != nil {
		t.Fatalf("LoadDataset: %v", err)
	}

	if ds.Name != "A-n10-k2" {
		t.Errorf("Name = %q, want A-n10-k2", ds.Name)
	}
	if ds.Dimension != 11 {
		t.Errorf("Dimension = %d, want 11", ds.Dimension)
	}
	if ds.Capacity != 50 {
		t.Errorf("Capacity = %d, want 50", ds.Capacity)
	}
	if len(ds.Customers) != 10 {
		t.Errorf("Customers = %d, want 10", len(ds.Customers))
	}
	if ds.Depot.ID != 1 {
		t.Errorf("Depot.ID = %d, want 1", ds.Depot.ID)
	}
	if ds.Depot.X != 40 || ds.Depot.Y != 40 {
		t.Errorf("Depot coordinates = (%v, %v), want (40, 40)", ds.Depot.X, ds.Depot.Y)
	}

	// Verify first customer.
	if ds.Customers[0].ID != 2 {
		t.Errorf("First customer ID = %d, want 2", ds.Customers[0].ID)
	}
	if ds.Customers[0].Demand != 10 {
		t.Errorf("First customer demand = %d, want 10", ds.Customers[0].Demand)
	}

	// Total demand should be 10+7+13+19+15+8+11+12+14+9 = 118.
	totalDemand := 0
	for _, c := range ds.Customers {
		totalDemand += c.Demand
	}
	if totalDemand != 118 {
		t.Errorf("Total demand = %d, want 118", totalDemand)
	}
}

// TestCVRPProblem_ImplementsInterface verifies compile-time satisfaction.
func TestCVRPProblem_ImplementsInterface(t *testing.T) {
	var _ optimisation.Problem = (*CVRPProblem)(nil)
}

// TestCVRPProblem_CreateInitialSolution verifies the initial solution is feasible.
func TestCVRPProblem_CreateInitialSolution(t *testing.T) {
	ds, err := LoadDataset(testInstancePath)
	if err != nil {
		t.Fatalf("LoadDataset: %v", err)
	}

	problem := NewCVRPProblem(ds)
	sol, err := problem.CreateInitialSolution()
	if err != nil {
		t.Fatalf("CreateInitialSolution: %v", err)
	}

	cs := sol.(*cvrpSolution)

	// All customers must be visited exactly once.
	visited := make(map[int]bool)
	for _, route := range cs.routes {
		for _, custIdx := range route {
			if visited[custIdx] {
				t.Errorf("Customer %d visited more than once", custIdx)
			}
			visited[custIdx] = true
		}
	}
	if len(visited) != len(ds.Customers) {
		t.Errorf("Visited %d customers, want %d", len(visited), len(ds.Customers))
	}

	// No route should exceed capacity.
	for i, load := range cs.loads {
		if load > ds.Capacity {
			t.Errorf("Route %d load %d exceeds capacity %d", i, load, ds.Capacity)
		}
	}

	// Should need at least ceil(118/50) = 3 routes.
	if len(cs.routes) < 3 {
		t.Logf("Using %d routes (minimum theoretical: 3)", len(cs.routes))
	}
}

// TestCVRPProblem_CloneSolution verifies independence.
func TestCVRPProblem_CloneSolution(t *testing.T) {
	ds, _ := LoadDataset(testInstancePath)
	problem := NewCVRPProblem(ds)
	sol, _ := problem.CreateInitialSolution()

	clone := problem.CloneSolution(sol)
	cs := sol.(*cvrpSolution)
	cc := clone.(*cvrpSolution)

	// Mutate original.
	if len(cs.routes) > 0 && len(cs.routes[0]) > 0 {
		cs.routes[0][0] = 999
	}

	// Clone should be unaffected.
	if len(cc.routes) > 0 && len(cc.routes[0]) > 0 && cc.routes[0][0] == 999 {
		t.Error("Clone was affected by original mutation")
	}
}

// TestCVRPProblem_Evaluate verifies evaluation produces positive cost.
func TestCVRPProblem_Evaluate(t *testing.T) {
	ds, _ := LoadDataset(testInstancePath)
	problem := NewCVRPProblem(ds)
	sol, _ := problem.CreateInitialSolution()

	cost := problem.Evaluate(sol)
	if cost <= 0 {
		t.Errorf("Evaluate returned %d, expected positive cost", cost)
	}
	t.Logf("Initial solution cost: %d", cost)
}

// TestCVRPProblem_TryMoveAndUndo verifies move/undo roundtrip.
func TestCVRPProblem_TryMoveAndUndo(t *testing.T) {
	ds, _ := LoadDataset(testInstancePath)
	problem := NewCVRPProblem(ds)
	sol, _ := problem.CreateInitialSolution()
	rng := rand.New(rand.NewSource(42))

	costBefore := problem.Evaluate(sol)
	fpBefore := problem.SolutionFingerprint(sol)

	// Find a valid move.
	var result optimisation.MoveResult
	for i := 0; i < 1000; i++ {
		result = problem.TryMove(sol, rng)
		if result.Valid {
			break
		}
	}
	if !result.Valid {
		t.Fatal("Could not find a valid move in 1000 attempts")
	}

	// Undo it.
	problem.UndoMove(sol, result.Move)

	costAfter := problem.Evaluate(sol)
	fpAfter := problem.SolutionFingerprint(sol)

	if costAfter != costBefore {
		t.Errorf("After undo: cost=%d, expected %d", costAfter, costBefore)
	}
	if fpAfter != fpBefore {
		t.Errorf("After undo: fingerprint=%s, expected %s", fpAfter, fpBefore)
	}
}

// TestCVRPProblem_SearchLoop verifies hill-climbing improves the solution.
func TestCVRPProblem_SearchLoop(t *testing.T) {
	ds, _ := LoadDataset(testInstancePath)
	problem := NewCVRPProblem(ds)
	sol, _ := problem.CreateInitialSolution()
	rng := rand.New(rand.NewSource(42))

	initialCost := problem.Evaluate(sol)
	currentCost := initialCost
	bestCost := initialCost
	iterations := 50000
	candidates := 0

	for i := 0; i < iterations; i++ {
		result := problem.TryMove(sol, rng)
		if !result.Valid {
			continue
		}
		candidates++

		newCost := problem.Evaluate(sol)
		if newCost < currentCost {
			currentCost = newCost
			if currentCost < bestCost {
				bestCost = currentCost
			}
		} else {
			problem.UndoMove(sol, result.Move)
		}
	}

	if candidates == 0 {
		t.Fatal("No valid candidates in 50000 iterations")
	}
	if bestCost >= initialCost {
		t.Errorf("Hill-climbing did not improve: initial=%d, best=%d", initialCost, bestCost)
	}

	improvement := float64(initialCost-bestCost) / float64(initialCost) * 100
	t.Logf("SearchLoop: initial=%d, best=%d, improvement=%.1f%%, candidates=%d", initialCost, bestCost, improvement, candidates)
}

// TestCVRPProblem_SerializeSolution verifies JSON output.
func TestCVRPProblem_SerializeSolution(t *testing.T) {
	ds, _ := LoadDataset(testInstancePath)
	problem := NewCVRPProblem(ds)
	sol, _ := problem.CreateInitialSolution()

	data, err := problem.SerializeSolution(sol)
	if err != nil {
		t.Fatalf("SerializeSolution: %v", err)
	}

	var output struct {
		Routes    []json.RawMessage `json:"routes"`
		TotalCost int               `json:"totalCost"`
		Vehicles  int               `json:"vehicles"`
		Feasible  bool              `json:"feasible"`
	}
	if err := json.Unmarshal(data, &output); err != nil {
		t.Fatalf("Invalid JSON: %v", err)
	}
	if output.Vehicles == 0 {
		t.Error("No vehicles in serialized output")
	}
	if output.TotalCost <= 0 {
		t.Error("TotalCost should be positive")
	}
	if !output.Feasible {
		t.Error("Initial solution should be feasible")
	}
}

// TestCVRPProblem_SolutionFingerprint verifies determinism.
func TestCVRPProblem_SolutionFingerprint(t *testing.T) {
	ds, _ := LoadDataset(testInstancePath)
	problem := NewCVRPProblem(ds)
	sol, _ := problem.CreateInitialSolution()

	fp1 := problem.SolutionFingerprint(sol)
	fp2 := problem.SolutionFingerprint(sol)
	if fp1 != fp2 {
		t.Errorf("Fingerprint not deterministic: %s vs %s", fp1, fp2)
	}
	if len(fp1) == 0 {
		t.Error("Fingerprint is empty")
	}
}

// TestCVRPProblem_AllCustomersServedAfterMoves verifies solution integrity after many moves.
func TestCVRPProblem_AllCustomersServedAfterMoves(t *testing.T) {
	ds, _ := LoadDataset(testInstancePath)
	problem := NewCVRPProblem(ds)
	sol, _ := problem.CreateInitialSolution()
	rng := rand.New(rand.NewSource(123))

	// Apply 1000 moves (accept all valid ones).
	for i := 0; i < 1000; i++ {
		result := problem.TryMove(sol, rng)
		if result.Valid {
			// Accept unconditionally for this integrity test.
		}
	}

	// Verify all customers still visited exactly once.
	cs := sol.(*cvrpSolution)
	visited := make(map[int]int) // custIdx -> count
	for _, route := range cs.routes {
		for _, custIdx := range route {
			visited[custIdx]++
		}
	}

	for i := 0; i < len(ds.Customers); i++ {
		count := visited[i]
		if count != 1 {
			t.Errorf("Customer %d visited %d times (expected 1)", i, count)
		}
	}
}

// TestDistanceRounded verifies Euclidean distance calculation.
func TestDistanceRounded(t *testing.T) {
	// 3-4-5 triangle.
	d := DistanceRounded(0, 0, 3, 4)
	if d != 5 {
		t.Errorf("DistanceRounded(0,0,3,4) = %d, want 5", d)
	}
}
