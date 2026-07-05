package cvrp

import (
	"math/rand"
	"testing"

	"github.com/timdodgson/open-workforce-platform/platform/go/internal/optimisation"
)

// --- Apply/Undo Correctness ---

// TestRelocate_ApplyUndo verifies relocate move is perfectly reversible.
func TestRelocate_ApplyUndo(t *testing.T) {
	p := loadTestProblem(t)
	sol, _ := p.BuildInitialSolution(NearestNeighbour)
	rng := rand.New(rand.NewSource(42))

	for i := 0; i < 100; i++ {
		fpBefore := p.SolutionFingerprint(sol)
		costBefore := p.Evaluate(sol)

		result := p.generateRelocate(sol, rng)
		if !result.Valid {
			continue
		}

		p.UndoMoveOnSolution(sol, result.Move.(Move))
		fpAfter := p.SolutionFingerprint(sol)
		costAfter := p.Evaluate(sol)

		if fpAfter != fpBefore {
			t.Fatalf("Relocate undo failed at iteration %d: fingerprint %s != %s", i, fpAfter, fpBefore)
		}
		if costAfter != costBefore {
			t.Fatalf("Relocate undo failed at iteration %d: cost %d != %d", i, costAfter, costBefore)
		}
	}
}

// TestInterSwap_ApplyUndo verifies inter-route swap is perfectly reversible.
func TestInterSwap_ApplyUndo(t *testing.T) {
	p := loadTestProblem(t)
	sol, _ := p.BuildInitialSolution(NearestNeighbour)
	rng := rand.New(rand.NewSource(99))

	for i := 0; i < 100; i++ {
		fpBefore := p.SolutionFingerprint(sol)
		costBefore := p.Evaluate(sol)

		result := p.generateInterSwap(sol, rng)
		if !result.Valid {
			continue
		}

		p.UndoMoveOnSolution(sol, result.Move.(Move))
		fpAfter := p.SolutionFingerprint(sol)
		costAfter := p.Evaluate(sol)

		if fpAfter != fpBefore {
			t.Fatalf("InterSwap undo failed at iteration %d: fingerprint %s != %s", i, fpAfter, fpBefore)
		}
		if costAfter != costBefore {
			t.Fatalf("InterSwap undo failed at iteration %d: cost %d != %d", i, costAfter, costBefore)
		}
	}
}

// TestIntraSwap_ApplyUndo verifies intra-route swap is perfectly reversible.
func TestIntraSwap_ApplyUndo(t *testing.T) {
	p := loadTestProblem(t)
	sol, _ := p.BuildInitialSolution(NearestNeighbour)
	rng := rand.New(rand.NewSource(77))

	for i := 0; i < 100; i++ {
		fpBefore := p.SolutionFingerprint(sol)
		costBefore := p.Evaluate(sol)

		result := p.generateIntraSwap(sol, rng)
		if !result.Valid {
			continue
		}

		p.UndoMoveOnSolution(sol, result.Move.(Move))
		fpAfter := p.SolutionFingerprint(sol)
		costAfter := p.Evaluate(sol)

		if fpAfter != fpBefore {
			t.Fatalf("IntraSwap undo failed at iteration %d: fingerprint %s != %s", i, fpAfter, fpBefore)
		}
		if costAfter != costBefore {
			t.Fatalf("IntraSwap undo failed at iteration %d: cost %d != %d", i, costAfter, costBefore)
		}
	}
}

// TestTwoOpt_ApplyUndo verifies 2-opt reversal is perfectly reversible.
func TestTwoOpt_ApplyUndo(t *testing.T) {
	p := loadTestProblem(t)
	sol, _ := p.BuildInitialSolution(NearestNeighbour)
	rng := rand.New(rand.NewSource(55))

	for i := 0; i < 100; i++ {
		fpBefore := p.SolutionFingerprint(sol)
		costBefore := p.Evaluate(sol)

		result := p.generateTwoOpt(sol, rng)
		if !result.Valid {
			continue
		}

		p.UndoMoveOnSolution(sol, result.Move.(Move))
		fpAfter := p.SolutionFingerprint(sol)
		costAfter := p.Evaluate(sol)

		if fpAfter != fpBefore {
			t.Fatalf("TwoOpt undo failed at iteration %d: fingerprint %s != %s", i, fpAfter, fpBefore)
		}
		if costAfter != costBefore {
			t.Fatalf("TwoOpt undo failed at iteration %d: cost %d != %d", i, costAfter, costBefore)
		}
	}
}

// --- Route Validity ---

// TestMoves_PreserveAllCustomers verifies no customers are lost or duplicated after many moves.
func TestMoves_PreserveAllCustomers(t *testing.T) {
	p := loadTestProblem(t)
	sol, _ := p.BuildInitialSolution(NearestNeighbour)
	rng := rand.New(rand.NewSource(123))

	// Apply 500 moves without undo (accept all).
	for i := 0; i < 500; i++ {
		result := p.GenerateMove(sol, rng)
		if result.Valid {
			// Accept unconditionally.
		}
	}

	assertAllCustomersVisitedOnce(t, p, sol)
}

// TestMoves_PreserveCapacityWhenFeasible verifies capacity is respected after feasible moves.
func TestMoves_PreserveCapacityWhenFeasible(t *testing.T) {
	p := loadTestProblem(t)
	sol, _ := p.BuildInitialSolution(NearestNeighbour)
	rng := rand.New(rand.NewSource(456))

	// GenerateMove only applies moves that pass capacity check.
	for i := 0; i < 500; i++ {
		p.GenerateMove(sol, rng)
	}

	assertCapacityRespected(t, p, sol)
}

// TestMoves_LoadsStayConsistent verifies stored loads match actual demands after moves.
func TestMoves_LoadsStayConsistent(t *testing.T) {
	p := loadTestProblem(t)
	sol, _ := p.BuildInitialSolution(NearestNeighbour)
	rng := rand.New(rand.NewSource(789))

	for i := 0; i < 500; i++ {
		p.GenerateMove(sol, rng)
	}

	assertLoadsCorrect(t, p, sol)
}

// --- Customer Preservation Through Apply/Undo Cycles ---

// TestMixedMoves_ApplyUndoCycle applies and undoes random moves, verifying restoration.
func TestMixedMoves_ApplyUndoCycle(t *testing.T) {
	p := loadTestProblem(t)
	sol, _ := p.BuildInitialSolution(NearestNeighbour)
	rng := rand.New(rand.NewSource(314))

	for i := 0; i < 200; i++ {
		fpBefore := p.SolutionFingerprint(sol)

		result := p.GenerateMove(sol, rng)
		if !result.Valid {
			continue
		}

		// Undo immediately.
		p.UndoMoveOnSolution(sol, result.Move.(Move))

		fpAfter := p.SolutionFingerprint(sol)
		if fpAfter != fpBefore {
			mv := result.Move.(Move)
			t.Fatalf("Apply/Undo cycle failed at iteration %d (move type: %s): %s != %s",
				i, mv.Type.String(), fpAfter, fpBefore)
		}
	}
}

// --- Capacity Handling ---

// TestRelocate_RejectsCapacityViolation verifies relocate won't create overloaded routes.
func TestRelocate_RejectsCapacityViolation(t *testing.T) {
	// Create a tight instance where inter-route relocates will often violate capacity.
	ds := &Dataset{
		Name:      "tight-cap",
		Dimension: 5,
		Capacity:  15,
		Depot:     Depot{ID: 1, X: 0, Y: 0},
		Customers: []Customer{
			{ID: 2, X: 10, Y: 0, Demand: 10},
			{ID: 3, X: 20, Y: 0, Demand: 10},
			{ID: 4, X: 30, Y: 0, Demand: 10},
			{ID: 5, X: 40, Y: 0, Demand: 10},
		},
	}
	p := NewCVRPProblem(ds)
	sol, _ := p.BuildInitialSolution(NearestNeighbour)
	rng := rand.New(rand.NewSource(42))

	// After many relocate attempts, capacity should never be exceeded.
	for i := 0; i < 500; i++ {
		p.generateRelocate(sol, rng)
	}

	assertCapacityRespected(t, p, sol)
}

// TestInterSwap_RejectsCapacityViolation verifies swap won't create overloaded routes.
func TestInterSwap_RejectsCapacityViolation(t *testing.T) {
	ds := &Dataset{
		Name:      "tight-swap",
		Dimension: 5,
		Capacity:  12,
		Depot:     Depot{ID: 1, X: 0, Y: 0},
		Customers: []Customer{
			{ID: 2, X: 10, Y: 0, Demand: 5},
			{ID: 3, X: 20, Y: 0, Demand: 10},
			{ID: 4, X: 30, Y: 0, Demand: 8},
			{ID: 5, X: 40, Y: 0, Demand: 12},
		},
	}
	p := NewCVRPProblem(ds)
	sol, _ := p.BuildInitialSolution(NearestNeighbour)
	rng := rand.New(rand.NewSource(99))

	for i := 0; i < 500; i++ {
		p.generateInterSwap(sol, rng)
	}

	assertCapacityRespected(t, p, sol)
}

// --- Telemetry ---

// TestMove_Description verifies all move types produce meaningful descriptions.
func TestMove_Description(t *testing.T) {
	moves := []Move{
		{Type: Relocate, FromRoute: 0, FromPos: 1, ToRoute: 2, ToPos: 0, CustomerA: 5, CustomerB: -1},
		{Type: Swap, FromRoute: 0, FromPos: 1, ToRoute: 1, ToPos: 2, CustomerA: 3, CustomerB: 7},
		{Type: IntraSwap, FromRoute: 1, FromPos: 0, ToRoute: 1, ToPos: 3, CustomerA: 2, CustomerB: 4},
		{Type: TwoOpt, FromRoute: 0, FromPos: 1, ToRoute: 0, ToPos: 4, CustomerA: 1, CustomerB: -1},
	}

	for _, mv := range moves {
		desc := mv.Description()
		if desc == "" {
			t.Errorf("Move type %s produced empty description", mv.Type.String())
		}
		if desc == "unknown move" {
			t.Errorf("Move type %s produced 'unknown move'", mv.Type.String())
		}
	}
}

// TestMove_AffectedRoutes verifies route reporting.
func TestMove_AffectedRoutes(t *testing.T) {
	// Inter-route move affects 2 routes.
	mv := Move{Type: Swap, FromRoute: 0, ToRoute: 2}
	routes := mv.AffectedRoutes()
	if len(routes) != 2 {
		t.Errorf("Inter-route swap should affect 2 routes, got %d", len(routes))
	}

	// Intra-route move affects 1 route.
	mv2 := Move{Type: IntraSwap, FromRoute: 1, ToRoute: 1}
	routes2 := mv2.AffectedRoutes()
	if len(routes2) != 1 {
		t.Errorf("Intra-route swap should affect 1 route, got %d", len(routes2))
	}

	// 2-opt affects 1 route.
	mv3 := Move{Type: TwoOpt, FromRoute: 0, ToRoute: 0}
	routes3 := mv3.AffectedRoutes()
	if len(routes3) != 1 {
		t.Errorf("2-opt should affect 1 route, got %d", len(routes3))
	}
}

// TestMove_AffectedCustomers verifies customer reporting.
func TestMove_AffectedCustomers(t *testing.T) {
	mv := Move{Type: Swap, CustomerA: 3, CustomerB: 7}
	custs := mv.AffectedCustomers()
	if len(custs) != 2 || custs[0] != 3 || custs[1] != 7 {
		t.Errorf("Expected [3, 7], got %v", custs)
	}

	mv2 := Move{Type: Relocate, CustomerA: 5, CustomerB: -1}
	custs2 := mv2.AffectedCustomers()
	if len(custs2) != 1 || custs2[0] != 5 {
		t.Errorf("Expected [5], got %v", custs2)
	}
}

// --- Edge Cases ---

// TestMoves_SingleRouteInstance handles when all customers are in one route.
func TestMoves_SingleRouteInstance(t *testing.T) {
	ds := &Dataset{
		Name:      "single-route",
		Dimension: 4,
		Capacity:  100, // large enough for all
		Depot:     Depot{ID: 1, X: 0, Y: 0},
		Customers: []Customer{
			{ID: 2, X: 10, Y: 0, Demand: 5},
			{ID: 3, X: 20, Y: 0, Demand: 5},
			{ID: 4, X: 30, Y: 0, Demand: 5},
		},
	}
	p := NewCVRPProblem(ds)
	sol, _ := p.BuildInitialSolution(NearestNeighbour)
	rng := rand.New(rand.NewSource(42))

	// Should still generate valid moves (intra-swap, 2-opt, intra-route relocate).
	validCount := 0
	for i := 0; i < 200; i++ {
		result := p.GenerateMove(sol, rng)
		if result.Valid {
			validCount++
		}
	}

	if validCount == 0 {
		t.Error("Expected some valid moves for single-route instance")
	}
	assertAllCustomersVisitedOnce(t, p, sol)
}

// TestMoves_TwoCustomerRoute handles minimum-size routes.
func TestMoves_TwoCustomerRoute(t *testing.T) {
	ds := &Dataset{
		Name:      "two-cust",
		Dimension: 3,
		Capacity:  100,
		Depot:     Depot{ID: 1, X: 0, Y: 0},
		Customers: []Customer{
			{ID: 2, X: 10, Y: 0, Demand: 5},
			{ID: 3, X: 20, Y: 0, Demand: 5},
		},
	}
	p := NewCVRPProblem(ds)
	sol, _ := p.BuildInitialSolution(NearestNeighbour)
	rng := rand.New(rand.NewSource(42))

	for i := 0; i < 100; i++ {
		result := p.GenerateMove(sol, rng)
		if result.Valid {
			p.UndoMoveOnSolution(sol, result.Move.(Move))
		}
	}

	assertAllCustomersVisitedOnce(t, p, sol)
}

// --- Integration with Problem Interface ---

// TestOrOpt_ApplyUndo verifies or-opt move is perfectly reversible.
func TestOrOpt_ApplyUndo(t *testing.T) {
	p := loadTestProblem(t)
	sol, _ := p.BuildInitialSolution(NearestNeighbour)
	rng := rand.New(rand.NewSource(33))

	for i := 0; i < 100; i++ {
		fpBefore := p.SolutionFingerprint(sol)
		costBefore := p.Evaluate(sol)

		result := p.generateOrOpt(sol, rng)
		if !result.Valid {
			continue
		}

		p.UndoMoveOnSolution(sol, result.Move.(Move))
		fpAfter := p.SolutionFingerprint(sol)
		costAfter := p.Evaluate(sol)

		if fpAfter != fpBefore {
			t.Fatalf("OrOpt undo failed at iteration %d: fingerprint %s != %s", i, fpAfter, fpBefore)
		}
		if costAfter != costBefore {
			t.Fatalf("OrOpt undo failed at iteration %d: cost %d != %d", i, costAfter, costBefore)
		}
	}
}

// TestTryMove_ViaInterface verifies the Problem interface TryMove delegates correctly.
func TestTryMove_ViaInterface(t *testing.T) {
	p := loadTestProblem(t)
	var prob optimisation.Problem = p

	sol, _ := prob.CreateInitialSolution()
	rng := rand.New(rand.NewSource(42))

	costBefore := prob.Evaluate(sol)
	fpBefore := p.SolutionFingerprint(sol)

	var validResult optimisation.MoveResult
	for i := 0; i < 100; i++ {
		result := prob.TryMove(sol, rng)
		if result.Valid {
			validResult = result
			break
		}
	}

	if !validResult.Valid {
		t.Fatal("Could not generate a valid move via interface")
	}

	// Undo via interface.
	prob.UndoMove(sol, validResult.Move)

	costAfter := prob.Evaluate(sol)
	fpAfter := p.SolutionFingerprint(sol)

	if costAfter != costBefore {
		t.Errorf("Interface TryMove/UndoMove cost mismatch: %d != %d", costAfter, costBefore)
	}
	if fpAfter != fpBefore {
		t.Errorf("Interface TryMove/UndoMove fingerprint mismatch: %s != %s", fpAfter, fpBefore)
	}
}

// TestGenerateMove_AllTypesReachable verifies all 4 move types can be generated.
func TestGenerateMove_AllTypesReachable(t *testing.T) {
	p := loadTestProblem(t)
	sol, _ := p.BuildInitialSolution(NearestNeighbour)
	rng := rand.New(rand.NewSource(42))

	seen := map[MoveType]bool{}
	for i := 0; i < 5000 && len(seen) < 5; i++ {
		result := p.GenerateMove(sol, rng)
		if result.Valid {
			mv := result.Move.(Move)
			seen[mv.Type] = true
			// Undo to keep solution stable.
			p.UndoMoveOnSolution(sol, mv)
		}
	}

	for _, mt := range []MoveType{Relocate, Swap, IntraSwap, TwoOpt, OrOpt} {
		if !seen[mt] {
			t.Errorf("Move type %s was never generated", mt.String())
		}
	}
}
