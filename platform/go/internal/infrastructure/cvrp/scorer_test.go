package cvrp

import (
	"testing"
)

// helper: build a problem from the test instance.
func loadTestProblem(t *testing.T) *CVRPProblem {
	t.Helper()
	ds, err := LoadDataset("testdata/A-n10-k2.vrp")
	if err != nil {
		t.Fatalf("LoadDataset: %v", err)
	}
	return NewCVRPProblem(ds)
}

// --- Scoring Tests ---

// TestScore_InitialSolutionFeasible verifies the constructive solution is feasible.
func TestScore_InitialSolutionFeasible(t *testing.T) {
	p := loadTestProblem(t)
	sol, _ := p.CreateInitialSolution()
	cs := sol.(*cvrpSolution)

	result := p.Score(cs)

	if !result.Feasible {
		t.Errorf("Initial solution should be feasible, got violations: %+v", result.Violations)
	}
	if result.TotalDistance <= 0 {
		t.Errorf("TotalDistance should be positive, got %d", result.TotalDistance)
	}
	if result.RouteCount == 0 {
		t.Error("RouteCount should be > 0")
	}
	if result.AugmentedCost != result.TotalDistance {
		t.Errorf("Feasible solution: AugmentedCost (%d) should equal TotalDistance (%d)",
			result.AugmentedCost, result.TotalDistance)
	}

	t.Logf("Score: distance=%d, routes=%d, feasible=%v", result.TotalDistance, result.RouteCount, result.Feasible)
}

// TestScore_MatchesEvaluate verifies Score().AugmentedCost == Evaluate().
func TestScore_MatchesEvaluate(t *testing.T) {
	p := loadTestProblem(t)
	sol, _ := p.CreateInitialSolution()
	cs := sol.(*cvrpSolution)

	scoreResult := p.Score(cs)
	evalResult := p.Evaluate(sol)

	if scoreResult.AugmentedCost != evalResult {
		t.Errorf("Score.AugmentedCost=%d != Evaluate=%d", scoreResult.AugmentedCost, evalResult)
	}
}

// TestScore_PerRouteBreakdown verifies route distances sum to total.
func TestScore_PerRouteBreakdown(t *testing.T) {
	p := loadTestProblem(t)
	sol, _ := p.CreateInitialSolution()
	cs := sol.(*cvrpSolution)

	result := p.Score(cs)

	sumDistances := 0
	for _, d := range result.RouteDistances {
		sumDistances += d
	}
	if sumDistances != result.TotalDistance {
		t.Errorf("Sum of route distances (%d) != TotalDistance (%d)", sumDistances, result.TotalDistance)
	}

	if len(result.RouteDistances) != result.RouteCount {
		t.Errorf("RouteDistances length (%d) != RouteCount (%d)", len(result.RouteDistances), result.RouteCount)
	}
}

// --- Validation Tests ---

// TestValidate_FeasibleSolution passes with no violations.
func TestValidate_FeasibleSolution(t *testing.T) {
	p := loadTestProblem(t)
	sol, _ := p.CreateInitialSolution()
	cs := sol.(*cvrpSolution)

	feasible, violations := p.ValidateFull(cs)
	if !feasible {
		t.Errorf("Expected feasible, got violations: %+v", violations)
	}
}

// TestValidate_CapacityViolation detects overloaded routes.
func TestValidate_CapacityViolation(t *testing.T) {
	p := loadTestProblem(t)
	sol, _ := p.CreateInitialSolution()
	cs := sol.(*cvrpSolution)

	// Force all customers into one route — will exceed capacity.
	allCustomers := []int{}
	for _, route := range cs.routes {
		allCustomers = append(allCustomers, route...)
	}
	cs.routes = [][]int{allCustomers}
	cs.loads = []int{p.routeLoad(allCustomers)}

	violations := p.Validate(cs)

	hasCapacity := false
	for _, v := range violations {
		if v.Code == "CAPACITY" {
			hasCapacity = true
			break
		}
	}
	if !hasCapacity {
		t.Error("Expected CAPACITY violation when all customers in one route")
	}
}

// TestValidate_CoverageViolation detects unvisited customers.
func TestValidate_CoverageViolation(t *testing.T) {
	p := loadTestProblem(t)
	sol, _ := p.CreateInitialSolution()
	cs := sol.(*cvrpSolution)

	// Remove a customer from the solution.
	if len(cs.routes) > 0 && len(cs.routes[0]) > 1 {
		removed := cs.routes[0][0]
		cs.routes[0] = cs.routes[0][1:]
		cs.loads[0] -= p.dataset.Customers[removed].Demand
	}

	violations := p.Validate(cs)

	hasCoverage := false
	for _, v := range violations {
		if v.Code == "COVERAGE" {
			hasCoverage = true
			break
		}
	}
	if !hasCoverage {
		t.Error("Expected COVERAGE violation when a customer is removed")
	}
}

// TestValidate_DuplicateViolation detects customers visited more than once.
func TestValidate_DuplicateViolation(t *testing.T) {
	p := loadTestProblem(t)
	sol, _ := p.CreateInitialSolution()
	cs := sol.(*cvrpSolution)

	// Duplicate a customer.
	if len(cs.routes) >= 2 && len(cs.routes[0]) > 0 {
		custIdx := cs.routes[0][0]
		cs.routes[1] = append(cs.routes[1], custIdx)
		cs.loads[1] += p.dataset.Customers[custIdx].Demand
	}

	violations := p.Validate(cs)

	hasDuplicate := false
	for _, v := range violations {
		if v.Code == "DUPLICATE" {
			hasDuplicate = true
			break
		}
	}
	if !hasDuplicate {
		t.Error("Expected DUPLICATE violation when a customer appears in two routes")
	}
}

// TestValidate_EmptySolution detects completely empty solution.
func TestValidate_EmptySolution(t *testing.T) {
	p := loadTestProblem(t)
	cs := &cvrpSolution{routes: [][]int{}, loads: []int{}}

	violations := p.Validate(cs)

	hasCoverage := false
	for _, v := range violations {
		if v.Code == "COVERAGE" {
			hasCoverage = true
			break
		}
	}
	if !hasCoverage {
		t.Error("Expected COVERAGE violations for empty solution")
	}
	if len(violations) != len(p.dataset.Customers) {
		t.Errorf("Expected %d COVERAGE violations, got %d", len(p.dataset.Customers), len(violations))
	}
}

// --- Distance Tests ---

// TestTotalDistance_MatchesEvaluateForFeasible verifies pure distance matches.
func TestTotalDistance_MatchesEvaluateForFeasible(t *testing.T) {
	p := loadTestProblem(t)
	sol, _ := p.CreateInitialSolution()
	cs := sol.(*cvrpSolution)

	totalDist := p.TotalDistance(cs)
	evalCost := p.Evaluate(sol)

	// For a feasible solution, Evaluate = TotalDistance (no penalty).
	if totalDist != evalCost {
		t.Errorf("TotalDistance=%d != Evaluate=%d for feasible solution", totalDist, evalCost)
	}
}

// TestRouteDistance_SingleCustomer verifies depot→customer→depot distance.
func TestRouteDistance_SingleCustomer(t *testing.T) {
	// Build a minimal dataset: depot at (0,0), one customer at (3,4).
	ds := &Dataset{
		Name:      "test-single",
		Dimension: 2,
		Capacity:  100,
		Depot:     Depot{ID: 1, X: 0, Y: 0},
		Customers: []Customer{{ID: 2, X: 3, Y: 4, Demand: 10}},
	}
	p := NewCVRPProblem(ds)

	// Route with single customer (index 0).
	route := []int{0}
	dist := p.routeDistance(route)

	// Distance = depot→customer + customer→depot = 5 + 5 = 10.
	if dist != 10 {
		t.Errorf("Single customer route distance = %d, want 10", dist)
	}
}

// TestRouteDistance_TwoCustomers verifies depot→A→B→depot distance.
func TestRouteDistance_TwoCustomers(t *testing.T) {
	// Depot at (0,0), customer A at (3,0), customer B at (3,4).
	ds := &Dataset{
		Name:      "test-two",
		Dimension: 3,
		Capacity:  100,
		Depot:     Depot{ID: 1, X: 0, Y: 0},
		Customers: []Customer{
			{ID: 2, X: 3, Y: 0, Demand: 5},
			{ID: 3, X: 3, Y: 4, Demand: 5},
		},
	}
	p := NewCVRPProblem(ds)

	// Route: customer 0 then customer 1.
	route := []int{0, 1}
	dist := p.routeDistance(route)

	// depot(0,0)→A(3,0): 3
	// A(3,0)→B(3,4): 4
	// B(3,4)→depot(0,0): 5
	// Total: 3 + 4 + 5 = 12
	if dist != 12 {
		t.Errorf("Two customer route distance = %d, want 12", dist)
	}
}

// TestRouteDistance_EmptyRoute returns zero.
func TestRouteDistance_EmptyRoute(t *testing.T) {
	p := loadTestProblem(t)
	dist := p.routeDistance([]int{})
	if dist != 0 {
		t.Errorf("Empty route distance = %d, want 0", dist)
	}
}

// --- Augmented Cost Tests ---

// TestAugmentedCost_InfeasibleHigherThanFeasible verifies penalty increases cost.
func TestAugmentedCost_InfeasibleHigherThanFeasible(t *testing.T) {
	p := loadTestProblem(t)
	sol, _ := p.CreateInitialSolution()
	cs := sol.(*cvrpSolution)

	feasibleScore := p.Score(cs)

	// Force infeasibility: put all customers in one route.
	allCustomers := []int{}
	for _, route := range cs.routes {
		allCustomers = append(allCustomers, route...)
	}
	cs.routes = [][]int{allCustomers}
	cs.loads = []int{p.routeLoad(allCustomers)}

	infeasibleScore := p.Score(cs)

	if infeasibleScore.AugmentedCost <= feasibleScore.AugmentedCost {
		t.Errorf("Infeasible augmented cost (%d) should be > feasible (%d)",
			infeasibleScore.AugmentedCost, feasibleScore.AugmentedCost)
	}
}

// TestScore_ViolationDetails verifies violation messages are informative.
func TestScore_ViolationDetails(t *testing.T) {
	p := loadTestProblem(t)
	cs := &cvrpSolution{routes: [][]int{}, loads: []int{}}

	result := p.Score(cs)

	for _, v := range result.Violations {
		if v.Detail == "" {
			t.Errorf("Violation %q has empty detail", v.Code)
		}
		if v.Code == "" {
			t.Error("Violation has empty code")
		}
	}
}
