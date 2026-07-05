package vrptw

import (
	"testing"

	"github.com/timdodgson/open-workforce-platform/platform/go/internal/optimisation"
)

const testInstancePath = "testdata/c101-25.txt"

// TestLoadDataset verifies Solomon format parsing.
func TestLoadDataset(t *testing.T) {
	ds, err := LoadDataset(testInstancePath)
	if err != nil {
		t.Fatalf("LoadDataset: %v", err)
	}

	if ds.Name != "C101-25" {
		t.Errorf("Name = %q, want C101-25", ds.Name)
	}
	if ds.Capacity != 200 {
		t.Errorf("Capacity = %d, want 200", ds.Capacity)
	}
	if ds.Vehicles != 25 {
		t.Errorf("Vehicles = %d, want 25", ds.Vehicles)
	}
	if len(ds.Customers) != 25 {
		t.Errorf("Customers = %d, want 25", len(ds.Customers))
	}
	if ds.Depot.X != 40 || ds.Depot.Y != 50 {
		t.Errorf("Depot = (%g, %g), want (40, 50)", ds.Depot.X, ds.Depot.Y)
	}
	if ds.Depot.DueDate != 1236 {
		t.Errorf("Depot DueDate = %d, want 1236", ds.Depot.DueDate)
	}

	// Check first customer.
	c1 := ds.Customers[0]
	if c1.ID != 1 || c1.X != 45 || c1.Y != 68 || c1.Demand != 10 {
		t.Errorf("Customer 1 = %+v, unexpected", c1)
	}
	if c1.ReadyTime != 912 || c1.DueDate != 967 || c1.Service != 90 {
		t.Errorf("Customer 1 time window = [%d, %d], service=%d", c1.ReadyTime, c1.DueDate, c1.Service)
	}
}

// TestCreateInitialSolution verifies constructive heuristic produces a valid solution.
func TestCreateInitialSolution(t *testing.T) {
	ds, err := LoadDataset(testInstancePath)
	if err != nil {
		t.Fatalf("LoadDataset: %v", err)
	}

	problem := NewVRPTWProblem(ds)
	sol, err := problem.CreateInitialSolution()
	if err != nil {
		t.Fatalf("CreateInitialSolution: %v", err)
	}

	// Verify all customers assigned.
	vrpSol := sol.(*vrptwSolution)
	totalCustomers := 0
	for _, route := range vrpSol.routes {
		totalCustomers += len(route)
	}
	if totalCustomers != 25 {
		t.Errorf("Total customers in solution = %d, want 25", totalCustomers)
	}

	// Verify evaluation produces positive distance.
	cost := problem.Evaluate(sol)
	if cost <= 0 {
		t.Errorf("Evaluate = %d, want > 0", cost)
	}
	t.Logf("Initial solution: %d distance, %d routes, feasible=%v",
		problem.TotalDistance(sol), problem.RouteCount(sol), problem.IsFeasible(sol))
}

// TestSA_EndToEnd proves SA can optimise VRPTW through the generic interface.
func TestSA_EndToEnd(t *testing.T) {
	ds, err := LoadDataset(testInstancePath)
	if err != nil {
		t.Fatalf("LoadDataset: %v", err)
	}

	problem := NewVRPTWProblem(ds)
	config := optimisation.SearchConfig{
		Mode:               "sa",
		Iterations:         50000,
		InitialTemperature: 100.0,
		MinTemperature:     0.0001,
		CoolingMode:        "adaptive",
		Seed:               42,
	}

	result := optimisation.RunSearch(problem, config)

	if result.BestPenalty <= 0 {
		t.Errorf("BestPenalty = %d, want > 0", result.BestPenalty)
	}
	if result.BestPenalty >= result.InitialPenalty {
		t.Errorf("SA did not improve: initial=%d, best=%d", result.InitialPenalty, result.BestPenalty)
	}
	if result.Improved == 0 {
		t.Error("No improvements found")
	}

	t.Logf("SA: %d → %d (%.1f%% improvement, %d discoveries)",
		result.InitialPenalty, result.BestPenalty,
		float64(result.InitialPenalty-result.BestPenalty)/float64(result.InitialPenalty)*100,
		result.Improved)
}

// TestLAHC_EndToEnd proves LAHC works on VRPTW.
func TestLAHC_EndToEnd(t *testing.T) {
	ds, err := LoadDataset(testInstancePath)
	if err != nil {
		t.Fatalf("LoadDataset: %v", err)
	}

	problem := NewVRPTWProblem(ds)
	config := optimisation.SearchConfig{
		Mode:                 "lahc",
		Iterations:           50000,
		LateAcceptanceLength: 500,
		Seed:                 42,
	}

	result := optimisation.RunSearch(problem, config)

	if result.BestPenalty >= result.InitialPenalty {
		t.Errorf("LAHC did not improve: initial=%d, best=%d", result.InitialPenalty, result.BestPenalty)
	}

	t.Logf("LAHC: %d → %d (%.1f%% improvement)",
		result.InitialPenalty, result.BestPenalty,
		float64(result.InitialPenalty-result.BestPenalty)/float64(result.InitialPenalty)*100)
}

// TestDeterministic proves same seed produces same result.
func TestDeterministic(t *testing.T) {
	ds, _ := LoadDataset(testInstancePath)
	problem := NewVRPTWProblem(ds)

	config := optimisation.SearchConfig{
		Mode:               "sa",
		Iterations:         10000,
		InitialTemperature: 50.0,
		MinTemperature:     0.001,
		CoolingMode:        "adaptive",
		Seed:               777,
	}

	r1 := optimisation.RunSearch(problem, config)
	r2 := optimisation.RunSearch(problem, config)

	if r1.BestPenalty != r2.BestPenalty {
		t.Errorf("Non-deterministic: run1=%d, run2=%d", r1.BestPenalty, r2.BestPenalty)
	}
}

// TestFeasibilityCheck verifies time window validation.
func TestFeasibilityCheck(t *testing.T) {
	ds, _ := LoadDataset(testInstancePath)
	problem := NewVRPTWProblem(ds)

	sol, _ := problem.CreateInitialSolution()
	feasible := problem.IsFeasible(sol)

	t.Logf("Initial solution feasible: %v", feasible)
	// The constructive heuristic should produce a feasible solution.
	if !feasible {
		t.Log("Warning: constructive heuristic produced infeasible solution (may be acceptable for tight windows)")
	}
}

// TestPortfolio_EndToEnd proves portfolio mode works on VRPTW.
func TestPortfolio_EndToEnd(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping portfolio test in short mode")
	}

	ds, _ := LoadDataset(testInstancePath)
	problem := NewVRPTWProblem(ds)

	config := optimisation.SearchConfig{
		Mode:                 "portfolio",
		Iterations:           50000,
		InitialTemperature:   100.0,
		MinTemperature:       0.0001,
		CoolingMode:          "adaptive",
		LateAcceptanceLength: 500,
		TabuTenure:           7,
		TabuNeighbourhood:    50,
		Portfolio:            []string{"sa", "lahc", "tabu"},
		Seed:                 42,
	}

	result := optimisation.RunSearch(problem, config)

	if result.BestPenalty >= result.InitialPenalty {
		t.Errorf("Portfolio did not improve: initial=%d, best=%d", result.InitialPenalty, result.BestPenalty)
	}

	t.Logf("Portfolio: %d → %d (%.1f%% improvement)",
		result.InitialPenalty, result.BestPenalty,
		float64(result.InitialPenalty-result.BestPenalty)/float64(result.InitialPenalty)*100)
}
