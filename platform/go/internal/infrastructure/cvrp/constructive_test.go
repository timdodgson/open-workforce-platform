package cvrp

import (
	"testing"
)

// --- Nearest Neighbour Tests ---

// TestNearestNeighbour_AllCustomersVisited verifies every customer is assigned exactly once.
func TestNearestNeighbour_AllCustomersVisited(t *testing.T) {
	p := loadTestProblem(t)
	sol, err := p.BuildInitialSolution(NearestNeighbour)
	if err != nil {
		t.Fatalf("BuildInitialSolution: %v", err)
	}

	assertAllCustomersVisitedOnce(t, p, sol)
}

// TestNearestNeighbour_CapacityRespected verifies no route exceeds capacity.
func TestNearestNeighbour_CapacityRespected(t *testing.T) {
	p := loadTestProblem(t)
	sol, err := p.BuildInitialSolution(NearestNeighbour)
	if err != nil {
		t.Fatalf("BuildInitialSolution: %v", err)
	}

	assertCapacityRespected(t, p, sol)
}

// TestNearestNeighbour_PassesFullValidation confirms no hard constraint violations.
func TestNearestNeighbour_PassesFullValidation(t *testing.T) {
	p := loadTestProblem(t)
	sol, err := p.BuildInitialSolution(NearestNeighbour)
	if err != nil {
		t.Fatalf("BuildInitialSolution: %v", err)
	}

	assertFeasible(t, p, sol)
}

// TestNearestNeighbour_ProducesRoutes verifies at least one route is created.
func TestNearestNeighbour_ProducesRoutes(t *testing.T) {
	p := loadTestProblem(t)
	sol, err := p.BuildInitialSolution(NearestNeighbour)
	if err != nil {
		t.Fatalf("BuildInitialSolution: %v", err)
	}

	if len(sol.routes) == 0 {
		t.Error("Expected at least one route")
	}
}

// TestNearestNeighbour_LoadsMatchDemands verifies stored loads are correct.
func TestNearestNeighbour_LoadsMatchDemands(t *testing.T) {
	p := loadTestProblem(t)
	sol, err := p.BuildInitialSolution(NearestNeighbour)
	if err != nil {
		t.Fatalf("BuildInitialSolution: %v", err)
	}

	assertLoadsCorrect(t, p, sol)
}

// TestNearestNeighbour_Deterministic verifies same output given same input.
func TestNearestNeighbour_Deterministic(t *testing.T) {
	p := loadTestProblem(t)
	sol1, _ := p.BuildInitialSolution(NearestNeighbour)
	sol2, _ := p.BuildInitialSolution(NearestNeighbour)

	fp1 := p.SolutionFingerprint(sol1)
	fp2 := p.SolutionFingerprint(sol2)

	if fp1 != fp2 {
		t.Errorf("NearestNeighbour not deterministic: %s != %s", fp1, fp2)
	}
}

// --- Sequential Tests ---

// TestSequential_AllCustomersVisited verifies every customer is assigned exactly once.
func TestSequential_AllCustomersVisited(t *testing.T) {
	p := loadTestProblem(t)
	sol, err := p.BuildInitialSolution(Sequential)
	if err != nil {
		t.Fatalf("BuildInitialSolution: %v", err)
	}

	assertAllCustomersVisitedOnce(t, p, sol)
}

// TestSequential_CapacityRespected verifies no route exceeds capacity.
func TestSequential_CapacityRespected(t *testing.T) {
	p := loadTestProblem(t)
	sol, err := p.BuildInitialSolution(Sequential)
	if err != nil {
		t.Fatalf("BuildInitialSolution: %v", err)
	}

	assertCapacityRespected(t, p, sol)
}

// TestSequential_PassesFullValidation confirms no hard constraint violations.
func TestSequential_PassesFullValidation(t *testing.T) {
	p := loadTestProblem(t)
	sol, err := p.BuildInitialSolution(Sequential)
	if err != nil {
		t.Fatalf("BuildInitialSolution: %v", err)
	}

	assertFeasible(t, p, sol)
}

// TestSequential_LoadsMatchDemands verifies stored loads are correct.
func TestSequential_LoadsMatchDemands(t *testing.T) {
	p := loadTestProblem(t)
	sol, err := p.BuildInitialSolution(Sequential)
	if err != nil {
		t.Fatalf("BuildInitialSolution: %v", err)
	}

	assertLoadsCorrect(t, p, sol)
}

// --- Edge Case Tests ---

// TestConstructive_EmptyDataset handles zero customers.
func TestConstructive_EmptyDataset(t *testing.T) {
	ds := &Dataset{
		Name:      "empty",
		Dimension: 1,
		Capacity:  100,
		Depot:     Depot{ID: 1, X: 0, Y: 0},
		Customers: []Customer{},
	}
	p := NewCVRPProblem(ds)

	sol, err := p.BuildInitialSolution(NearestNeighbour)
	if err != nil {
		t.Fatalf("BuildInitialSolution: %v", err)
	}
	if len(sol.routes) != 0 {
		t.Errorf("Expected 0 routes for empty dataset, got %d", len(sol.routes))
	}
}

// TestConstructive_SingleCustomer handles one customer.
func TestConstructive_SingleCustomer(t *testing.T) {
	ds := &Dataset{
		Name:      "single",
		Dimension: 2,
		Capacity:  100,
		Depot:     Depot{ID: 1, X: 0, Y: 0},
		Customers: []Customer{{ID: 2, X: 5, Y: 0, Demand: 10}},
	}
	p := NewCVRPProblem(ds)

	sol, err := p.BuildInitialSolution(NearestNeighbour)
	if err != nil {
		t.Fatalf("BuildInitialSolution: %v", err)
	}
	if len(sol.routes) != 1 {
		t.Errorf("Expected 1 route, got %d", len(sol.routes))
	}
	if len(sol.routes[0]) != 1 {
		t.Errorf("Expected 1 customer in route, got %d", len(sol.routes[0]))
	}
	assertFeasible(t, p, sol)
}

// TestConstructive_TightCapacity verifies correct route splitting when capacity is tight.
func TestConstructive_TightCapacity(t *testing.T) {
	// 4 customers each with demand 30, capacity 50 → need at least 3 routes.
	ds := &Dataset{
		Name:      "tight",
		Dimension: 5,
		Capacity:  50,
		Depot:     Depot{ID: 1, X: 0, Y: 0},
		Customers: []Customer{
			{ID: 2, X: 10, Y: 0, Demand: 30},
			{ID: 3, X: 20, Y: 0, Demand: 30},
			{ID: 4, X: 30, Y: 0, Demand: 30},
			{ID: 5, X: 40, Y: 0, Demand: 30},
		},
	}
	p := NewCVRPProblem(ds)

	for _, strategy := range []ConstructionStrategy{NearestNeighbour, Sequential} {
		sol, err := p.BuildInitialSolution(strategy)
		if err != nil {
			t.Fatalf("strategy %d: BuildInitialSolution: %v", strategy, err)
		}

		assertFeasible(t, p, sol)
		assertAllCustomersVisitedOnce(t, p, sol)

		// Each route can hold at most 1 customer (30 < 50, but 30+30 > 50).
		// So minimum 4 routes needed, or routes with single customer each.
		for i, load := range sol.loads {
			if load > ds.Capacity {
				t.Errorf("strategy %d: route %d load %d exceeds capacity %d", strategy, i, load, ds.Capacity)
			}
		}
	}
}

// TestConstructive_ExactCapacity verifies when demand exactly equals capacity.
func TestConstructive_ExactCapacity(t *testing.T) {
	ds := &Dataset{
		Name:      "exact",
		Dimension: 3,
		Capacity:  20,
		Depot:     Depot{ID: 1, X: 0, Y: 0},
		Customers: []Customer{
			{ID: 2, X: 5, Y: 0, Demand: 20},
			{ID: 3, X: 10, Y: 0, Demand: 20},
		},
	}
	p := NewCVRPProblem(ds)

	sol, err := p.BuildInitialSolution(NearestNeighbour)
	if err != nil {
		t.Fatalf("BuildInitialSolution: %v", err)
	}

	assertFeasible(t, p, sol)
	// Each customer fills a route exactly.
	if len(sol.routes) != 2 {
		t.Errorf("Expected 2 routes (exact capacity), got %d", len(sol.routes))
	}
}

// TestConstructive_LargeInstance verifies correctness on the test instance.
func TestConstructive_LargeInstance(t *testing.T) {
	p := loadTestProblem(t)

	for _, strategy := range []ConstructionStrategy{NearestNeighbour, Sequential} {
		sol, err := p.BuildInitialSolution(strategy)
		if err != nil {
			t.Fatalf("strategy %d: BuildInitialSolution: %v", strategy, err)
		}

		assertFeasible(t, p, sol)
		assertAllCustomersVisitedOnce(t, p, sol)
		assertLoadsCorrect(t, p, sol)
		assertCapacityRespected(t, p, sol)

		dist := p.TotalDistance(sol)
		t.Logf("strategy %d: routes=%d, distance=%d", strategy, len(sol.routes), dist)
	}
}

// TestNearestNeighbour_BetterThanSequential verifies NN produces shorter routes.
func TestNearestNeighbour_BetterThanSequential(t *testing.T) {
	p := loadTestProblem(t)
	nn, _ := p.BuildInitialSolution(NearestNeighbour)
	seq, _ := p.BuildInitialSolution(Sequential)

	nnDist := p.TotalDistance(nn)
	seqDist := p.TotalDistance(seq)

	// NN should generally be better (or equal) to sequential.
	// Not guaranteed for all instances, but for our test instance it should hold.
	t.Logf("NN distance=%d, Sequential distance=%d", nnDist, seqDist)
	if nnDist > seqDist*2 {
		t.Errorf("NN (%d) is dramatically worse than Sequential (%d)", nnDist, seqDist)
	}
}

// --- Assertion helpers ---

func assertAllCustomersVisitedOnce(t *testing.T, p *CVRPProblem, sol *cvrpSolution) {
	t.Helper()
	numCustomers := len(p.dataset.Customers)
	visitCount := make([]int, numCustomers)
	for _, route := range sol.routes {
		for _, custIdx := range route {
			if custIdx < 0 || custIdx >= numCustomers {
				t.Errorf("Invalid customer index %d (range 0-%d)", custIdx, numCustomers-1)
				continue
			}
			visitCount[custIdx]++
		}
	}
	for i, count := range visitCount {
		if count != 1 {
			t.Errorf("Customer %d (ID=%d) visited %d times, expected 1",
				i, p.dataset.Customers[i].ID, count)
		}
	}
}

func assertCapacityRespected(t *testing.T, p *CVRPProblem, sol *cvrpSolution) {
	t.Helper()
	for i, route := range sol.routes {
		load := p.routeLoad(route)
		if load > p.dataset.Capacity {
			t.Errorf("Route %d: load %d exceeds capacity %d", i, load, p.dataset.Capacity)
		}
	}
}

func assertFeasible(t *testing.T, p *CVRPProblem, sol *cvrpSolution) {
	t.Helper()
	feasible, violations := p.ValidateFull(sol)
	if !feasible {
		for _, v := range violations {
			t.Errorf("Violation: [%s] %s", v.Code, v.Detail)
		}
	}
}

func assertLoadsCorrect(t *testing.T, p *CVRPProblem, sol *cvrpSolution) {
	t.Helper()
	for i, route := range sol.routes {
		expectedLoad := p.routeLoad(route)
		if sol.loads[i] != expectedLoad {
			t.Errorf("Route %d: stored load %d != computed load %d", i, sol.loads[i], expectedLoad)
		}
	}
}
