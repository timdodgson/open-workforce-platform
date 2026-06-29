package optimisation_test

import (
	"testing"

	"github.com/timdodgson/open-workforce-platform/platform/go/internal/domain/workitem"
	"github.com/timdodgson/open-workforce-platform/platform/go/internal/optimisation"
)

// --- Selection ---

func TestSimulatedAnnealing_CanBeSelected(t *testing.T) {
	alg, err := optimisation.Get("simulated-annealing")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if alg.Name() != "simulated-annealing" {
		t.Errorf("expected name simulated-annealing, got %s", alg.Name())
	}
}

// --- Basic behaviour ---

func TestSimulatedAnnealing_ProducesValidPlan(t *testing.T) {
	items := []workitem.WorkItem{makeItem("WI-001"), makeItem("WI-002")}
	capacities := []optimisation.ResourceCapacity{makeCapacity("RES-001", 3, true, nil)}
	priorities := []optimisation.WorkItemPriority{makePriority("WI-001", 0, ""), makePriority("WI-002", 0, "")}

	alg, _ := optimisation.Get("simulated-annealing")
	result, err := alg.Solve(items, capacities, priorities)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.Size() != 2 {
		t.Errorf("expected 2 assignments, got %d", result.Size())
	}
	if result.Score() != 100 {
		t.Errorf("expected score 100, got %d", result.Score())
	}
}

// --- Respects constraints ---

func TestSimulatedAnnealing_RespectsCapacity(t *testing.T) {
	items := []workitem.WorkItem{makeItem("WI-001"), makeItem("WI-002"), makeItem("WI-003")}
	capacities := []optimisation.ResourceCapacity{makeCapacity("RES-001", 2, true, nil)}
	priorities := []optimisation.WorkItemPriority{
		makePriority("WI-001", 0, ""), makePriority("WI-002", 0, ""), makePriority("WI-003", 0, ""),
	}

	alg, _ := optimisation.Get("simulated-annealing")
	result, _ := alg.Solve(items, capacities, priorities)
	if result.Size() != 2 {
		t.Errorf("expected 2 assignments (capacity 2), got %d", result.Size())
	}
}

func TestSimulatedAnnealing_RespectsAvailability(t *testing.T) {
	items := []workitem.WorkItem{makeItem("WI-001")}
	capacities := []optimisation.ResourceCapacity{makeCapacity("RES-001", 5, false, nil)}
	priorities := []optimisation.WorkItemPriority{makePriority("WI-001", 0, "")}

	alg, _ := optimisation.Get("simulated-annealing")
	result, _ := alg.Solve(items, capacities, priorities)
	if result.Size() != 0 {
		t.Errorf("expected 0 assignments (unavailable), got %d", result.Size())
	}
}

func TestSimulatedAnnealing_RespectsSkills(t *testing.T) {
	items := []workitem.WorkItem{makeItem("WI-001")}
	capacities := []optimisation.ResourceCapacity{makeCapacity("RES-001", 5, true, []string{"electrical"})}
	priorities := []optimisation.WorkItemPriority{makePriority("WI-001", 0, "clinical")}

	alg, _ := optimisation.Get("simulated-annealing")
	result, _ := alg.Solve(items, capacities, priorities)
	if result.Size() != 0 {
		t.Errorf("expected 0 assignments (skill mismatch), got %d", result.Size())
	}
}

// --- Uses neighbourhood / improves like hill climbing ---

func TestSimulatedAnnealing_ImprovesOverConstructive(t *testing.T) {
	// Same scenario as hill climbing improvement test:
	// Constructive puts WI-B (higher priority, no skill) on RES-CLINICAL,
	// blocking WI-A (needs clinical). SA should fix this.
	items := []workitem.WorkItem{makeItem("WI-A"), makeItem("WI-B")}
	capacities := []optimisation.ResourceCapacity{
		makeCapacity("RES-CLINICAL", 1, true, []string{"clinical"}),
		makeCapacity("RES-GENERAL", 1, true, []string{"general"}),
	}
	priorities := []optimisation.WorkItemPriority{
		makePriority("WI-A", 100, "clinical"),
		makePriority("WI-B", 200, ""),
	}

	constructive, _ := optimisation.Solve(items, capacities, priorities)
	if constructive.Score() != 50 {
		t.Fatalf("expected constructive score 50, got %d", constructive.Score())
	}

	alg, _ := optimisation.Get("simulated-annealing")
	result, err := alg.Solve(items, capacities, priorities)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.Score() != 100 {
		t.Errorf("expected SA score 100, got %d", result.Score())
	}
}

// --- Determinism ---

func TestSimulatedAnnealing_Deterministic(t *testing.T) {
	items := []workitem.WorkItem{makeItem("WI-A"), makeItem("WI-B")}
	capacities := []optimisation.ResourceCapacity{
		makeCapacity("RES-CLINICAL", 1, true, []string{"clinical"}),
		makeCapacity("RES-GENERAL", 1, true, []string{"general"}),
	}
	priorities := []optimisation.WorkItemPriority{
		makePriority("WI-A", 100, "clinical"),
		makePriority("WI-B", 200, ""),
	}

	alg, _ := optimisation.Get("simulated-annealing")
	result1, _ := alg.Solve(items, capacities, priorities)
	result2, _ := alg.Solve(items, capacities, priorities)

	a1 := result1.Assignments()
	a2 := result2.Assignments()

	if len(a1) != len(a2) {
		t.Fatalf("different assignment counts: %d vs %d", len(a1), len(a2))
	}
	for i := range a1 {
		if a1[i].ResourceID() != a2[i].ResourceID() || a1[i].WorkItemID() != a2[i].WorkItemID() {
			t.Fatalf("not deterministic at index %d", i)
		}
	}
}

// --- Error cases ---

func TestSimulatedAnnealing_EmptyItems(t *testing.T) {
	capacities := []optimisation.ResourceCapacity{makeCapacity("RES-001", 2, true, nil)}
	alg, _ := optimisation.Get("simulated-annealing")
	_, err := alg.Solve(nil, capacities, nil)
	if err == nil {
		t.Fatal("expected error for empty items")
	}
}

func TestSimulatedAnnealing_EmptyResources(t *testing.T) {
	items := []workitem.WorkItem{makeItem("WI-001")}
	alg, _ := optimisation.Get("simulated-annealing")
	_, err := alg.Solve(items, nil, nil)
	if err == nil {
		t.Fatal("expected error for empty resources")
	}
}
