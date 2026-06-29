package optimisation_test

import (
	"encoding/json"
	"testing"

	"github.com/timdodgson/open-workforce-platform/platform/go/internal/domain/workitem"
	"github.com/timdodgson/open-workforce-platform/platform/go/internal/optimisation"
)

func makeItem(id string) workitem.WorkItem {
	w, _ := workitem.New(id, "test.type", json.RawMessage(`{"key":"value"}`))
	return w
}

func makeCapacity(id string, cap int) optimisation.ResourceCapacity {
	return optimisation.ResourceCapacity{ResourceID: id, Capacity: cap}
}

// --- Assignment within capacity ---

func TestSolve_AssignsWithinCapacity(t *testing.T) {
	items := []workitem.WorkItem{makeItem("WI-001"), makeItem("WI-002")}
	capacities := []optimisation.ResourceCapacity{makeCapacity("RES-001", 3)}

	result, err := optimisation.Solve(items, capacities)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.Size() != 2 {
		t.Errorf("expected 2 assignments, got %d", result.Size())
	}
	if result.UnassignedCount() != 0 {
		t.Errorf("expected 0 unassigned, got %d", result.UnassignedCount())
	}
}

// --- Exact capacity ---

func TestSolve_ExactCapacity(t *testing.T) {
	items := []workitem.WorkItem{makeItem("WI-001"), makeItem("WI-002")}
	capacities := []optimisation.ResourceCapacity{makeCapacity("RES-001", 2)}

	result, err := optimisation.Solve(items, capacities)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.Size() != 2 {
		t.Errorf("expected 2 assignments, got %d", result.Size())
	}
	if result.Score() != 100 {
		t.Errorf("expected score 100, got %d", result.Score())
	}
	if result.Utilisation() != 100 {
		t.Errorf("expected utilisation 100, got %d", result.Utilisation())
	}
}

// --- Insufficient capacity ---

func TestSolve_InsufficientCapacity(t *testing.T) {
	items := []workitem.WorkItem{makeItem("WI-001"), makeItem("WI-002"), makeItem("WI-003")}
	capacities := []optimisation.ResourceCapacity{makeCapacity("RES-001", 2)}

	result, err := optimisation.Solve(items, capacities)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.Size() != 2 {
		t.Errorf("expected 2 assignments, got %d", result.Size())
	}
	if result.UnassignedCount() != 1 {
		t.Errorf("expected 1 unassigned, got %d", result.UnassignedCount())
	}

	unassigned := result.Unassigned()
	if unassigned[0] != "WI-003" {
		t.Errorf("expected WI-003 unassigned, got %s", unassigned[0])
	}
}

// --- Spills to next resource ---

func TestSolve_SpillsToNextResource(t *testing.T) {
	items := []workitem.WorkItem{makeItem("WI-001"), makeItem("WI-002"), makeItem("WI-003")}
	capacities := []optimisation.ResourceCapacity{
		makeCapacity("RES-001", 2),
		makeCapacity("RES-002", 2),
	}

	result, err := optimisation.Solve(items, capacities)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.Size() != 3 {
		t.Errorf("expected 3 assignments, got %d", result.Size())
	}

	assignments := result.Assignments()
	if assignments[0].ResourceID() != "RES-001" {
		t.Errorf("expected first to RES-001, got %s", assignments[0].ResourceID())
	}
	if assignments[1].ResourceID() != "RES-001" {
		t.Errorf("expected second to RES-001, got %s", assignments[1].ResourceID())
	}
	if assignments[2].ResourceID() != "RES-002" {
		t.Errorf("expected third to RES-002, got %s", assignments[2].ResourceID())
	}
}

// --- Zero capacity ---

func TestSolve_ZeroCapacity(t *testing.T) {
	items := []workitem.WorkItem{makeItem("WI-001")}
	capacities := []optimisation.ResourceCapacity{makeCapacity("RES-001", 0)}

	result, err := optimisation.Solve(items, capacities)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.Size() != 0 {
		t.Errorf("expected 0 assignments, got %d", result.Size())
	}
	if result.UnassignedCount() != 1 {
		t.Errorf("expected 1 unassigned, got %d", result.UnassignedCount())
	}
	if result.Score() != 0 {
		t.Errorf("expected score 0, got %d", result.Score())
	}
}

// --- Scoring ---

func TestSolve_ScoreAllAssigned(t *testing.T) {
	items := []workitem.WorkItem{makeItem("WI-001")}
	capacities := []optimisation.ResourceCapacity{makeCapacity("RES-001", 5)}

	result, _ := optimisation.Solve(items, capacities)
	if result.Score() != 100 {
		t.Errorf("expected score 100, got %d", result.Score())
	}
}

func TestSolve_ScorePartial(t *testing.T) {
	// 2 of 3 assigned = 67%
	items := []workitem.WorkItem{makeItem("WI-001"), makeItem("WI-002"), makeItem("WI-003")}
	capacities := []optimisation.ResourceCapacity{makeCapacity("RES-001", 2)}

	result, _ := optimisation.Solve(items, capacities)
	if result.Score() != 67 {
		t.Errorf("expected score 67, got %d", result.Score())
	}
}

// --- Utilisation ---

func TestSolve_Utilisation(t *testing.T) {
	// 3 assigned, total capacity 4 = 75%
	items := []workitem.WorkItem{makeItem("WI-001"), makeItem("WI-002"), makeItem("WI-003")}
	capacities := []optimisation.ResourceCapacity{
		makeCapacity("RES-001", 2),
		makeCapacity("RES-002", 2),
	}

	result, _ := optimisation.Solve(items, capacities)
	if result.Utilisation() != 75 {
		t.Errorf("expected utilisation 75, got %d", result.Utilisation())
	}
}

// --- Error cases ---

func TestSolve_EmptyItems(t *testing.T) {
	capacities := []optimisation.ResourceCapacity{makeCapacity("RES-001", 2)}
	_, err := optimisation.Solve([]workitem.WorkItem{}, capacities)
	if err == nil {
		t.Fatal("expected error for empty items")
	}
}

func TestSolve_EmptyResources(t *testing.T) {
	items := []workitem.WorkItem{makeItem("WI-001")}
	_, err := optimisation.Solve(items, []optimisation.ResourceCapacity{})
	if err == nil {
		t.Fatal("expected error for empty resources")
	}
}

// --- Determinism ---

func TestSolve_Deterministic(t *testing.T) {
	items := []workitem.WorkItem{makeItem("WI-001"), makeItem("WI-002"), makeItem("WI-003")}
	capacities := []optimisation.ResourceCapacity{
		makeCapacity("RES-001", 2),
		makeCapacity("RES-002", 2),
	}

	result1, _ := optimisation.Solve(items, capacities)
	result2, _ := optimisation.Solve(items, capacities)

	a1 := result1.Assignments()
	a2 := result2.Assignments()

	for i := range a1 {
		if a1[i].ResourceID() != a2[i].ResourceID() || a1[i].WorkItemID() != a2[i].WorkItemID() {
			t.Fatalf("optimiser is not deterministic at index %d", i)
		}
	}
}
