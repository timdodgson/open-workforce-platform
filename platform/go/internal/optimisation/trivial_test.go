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

func makeCapacity(id string, cap int, available bool) optimisation.ResourceCapacity {
	return optimisation.ResourceCapacity{ResourceID: id, Capacity: cap, Available: available}
}

func makePriority(id string, priority int) optimisation.WorkItemPriority {
	return optimisation.WorkItemPriority{WorkItemID: id, Priority: priority}
}

// --- Capacity behaviour ---

func TestSolve_AssignsWithinCapacity(t *testing.T) {
	items := []workitem.WorkItem{makeItem("WI-001"), makeItem("WI-002")}
	capacities := []optimisation.ResourceCapacity{makeCapacity("RES-001", 3, true)}
	priorities := []optimisation.WorkItemPriority{makePriority("WI-001", 0), makePriority("WI-002", 0)}

	result, err := optimisation.Solve(items, capacities, priorities)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.Size() != 2 {
		t.Errorf("expected 2 assignments, got %d", result.Size())
	}
}

func TestSolve_ExactCapacity(t *testing.T) {
	items := []workitem.WorkItem{makeItem("WI-001"), makeItem("WI-002")}
	capacities := []optimisation.ResourceCapacity{makeCapacity("RES-001", 2, true)}
	priorities := []optimisation.WorkItemPriority{makePriority("WI-001", 0), makePriority("WI-002", 0)}

	result, err := optimisation.Solve(items, capacities, priorities)
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

func TestSolve_InsufficientCapacity(t *testing.T) {
	items := []workitem.WorkItem{makeItem("WI-001"), makeItem("WI-002"), makeItem("WI-003")}
	capacities := []optimisation.ResourceCapacity{makeCapacity("RES-001", 2, true)}
	priorities := []optimisation.WorkItemPriority{
		makePriority("WI-001", 0), makePriority("WI-002", 0), makePriority("WI-003", 0),
	}

	result, err := optimisation.Solve(items, capacities, priorities)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.Size() != 2 {
		t.Errorf("expected 2 assignments, got %d", result.Size())
	}
	if result.UnassignedCount() != 1 {
		t.Errorf("expected 1 unassigned, got %d", result.UnassignedCount())
	}
}

func TestSolve_SpillsToNextResource(t *testing.T) {
	items := []workitem.WorkItem{makeItem("WI-001"), makeItem("WI-002"), makeItem("WI-003")}
	capacities := []optimisation.ResourceCapacity{
		makeCapacity("RES-001", 2, true),
		makeCapacity("RES-002", 2, true),
	}
	priorities := []optimisation.WorkItemPriority{
		makePriority("WI-001", 0), makePriority("WI-002", 0), makePriority("WI-003", 0),
	}

	result, err := optimisation.Solve(items, capacities, priorities)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.Size() != 3 {
		t.Errorf("expected 3 assignments, got %d", result.Size())
	}
}

func TestSolve_ZeroCapacity(t *testing.T) {
	items := []workitem.WorkItem{makeItem("WI-001")}
	capacities := []optimisation.ResourceCapacity{makeCapacity("RES-001", 0, true)}
	priorities := []optimisation.WorkItemPriority{makePriority("WI-001", 0)}

	result, err := optimisation.Solve(items, capacities, priorities)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.Size() != 0 {
		t.Errorf("expected 0 assignments, got %d", result.Size())
	}
	if result.UnassignedCount() != 1 {
		t.Errorf("expected 1 unassigned, got %d", result.UnassignedCount())
	}
}

// --- Priority behaviour ---

func TestSolve_HigherPriorityAssignedFirst(t *testing.T) {
	items := []workitem.WorkItem{makeItem("WI-LOW"), makeItem("WI-HIGH")}
	capacities := []optimisation.ResourceCapacity{makeCapacity("RES-001", 1, true)}
	priorities := []optimisation.WorkItemPriority{
		makePriority("WI-LOW", 10), makePriority("WI-HIGH", 100),
	}

	result, err := optimisation.Solve(items, capacities, priorities)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	assigned := result.Assignments()[0]
	if assigned.WorkItemID() != "WI-HIGH" {
		t.Errorf("expected WI-HIGH assigned, got %s", assigned.WorkItemID())
	}
}

func TestSolve_EqualPriorityPreservesOrder(t *testing.T) {
	items := []workitem.WorkItem{makeItem("WI-FIRST"), makeItem("WI-SECOND"), makeItem("WI-THIRD")}
	capacities := []optimisation.ResourceCapacity{makeCapacity("RES-001", 2, true)}
	priorities := []optimisation.WorkItemPriority{
		makePriority("WI-FIRST", 50), makePriority("WI-SECOND", 50), makePriority("WI-THIRD", 50),
	}

	result, _ := optimisation.Solve(items, capacities, priorities)
	assignments := result.Assignments()
	if assignments[0].WorkItemID() != "WI-FIRST" {
		t.Errorf("expected WI-FIRST, got %s", assignments[0].WorkItemID())
	}
	if assignments[1].WorkItemID() != "WI-SECOND" {
		t.Errorf("expected WI-SECOND, got %s", assignments[1].WorkItemID())
	}
}

// --- Availability behaviour ---

func TestSolve_AvailableResourceReceivesAssignments(t *testing.T) {
	items := []workitem.WorkItem{makeItem("WI-001")}
	capacities := []optimisation.ResourceCapacity{makeCapacity("RES-001", 2, true)}
	priorities := []optimisation.WorkItemPriority{makePriority("WI-001", 0)}

	result, err := optimisation.Solve(items, capacities, priorities)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.Size() != 1 {
		t.Errorf("expected 1 assignment, got %d", result.Size())
	}
}

func TestSolve_UnavailableResourceSkipped(t *testing.T) {
	items := []workitem.WorkItem{makeItem("WI-001")}
	capacities := []optimisation.ResourceCapacity{
		makeCapacity("RES-UNAVAIL", 5, false),
		makeCapacity("RES-AVAIL", 2, true),
	}
	priorities := []optimisation.WorkItemPriority{makePriority("WI-001", 0)}

	result, err := optimisation.Solve(items, capacities, priorities)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	assigned := result.Assignments()[0]
	if assigned.ResourceID() != "RES-AVAIL" {
		t.Errorf("expected assignment to RES-AVAIL, got %s", assigned.ResourceID())
	}
}

func TestSolve_AllUnavailableResultsInAllUnassigned(t *testing.T) {
	items := []workitem.WorkItem{makeItem("WI-001"), makeItem("WI-002")}
	capacities := []optimisation.ResourceCapacity{
		makeCapacity("RES-001", 5, false),
		makeCapacity("RES-002", 5, false),
	}
	priorities := []optimisation.WorkItemPriority{
		makePriority("WI-001", 0), makePriority("WI-002", 0),
	}

	result, err := optimisation.Solve(items, capacities, priorities)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.Size() != 0 {
		t.Errorf("expected 0 assignments, got %d", result.Size())
	}
	if result.UnassignedCount() != 2 {
		t.Errorf("expected 2 unassigned, got %d", result.UnassignedCount())
	}
	if result.Score() != 0 {
		t.Errorf("expected score 0, got %d", result.Score())
	}
}

func TestSolve_UnavailableResourceCapacityNotCounted(t *testing.T) {
	items := []workitem.WorkItem{makeItem("WI-001")}
	capacities := []optimisation.ResourceCapacity{
		makeCapacity("RES-UNAVAIL", 10, false),
		makeCapacity("RES-AVAIL", 2, true),
	}
	priorities := []optimisation.WorkItemPriority{makePriority("WI-001", 0)}

	result, _ := optimisation.Solve(items, capacities, priorities)
	// Total capacity should only be from available resources (2, not 12).
	if result.TotalCapacity() != 2 {
		t.Errorf("expected total capacity 2, got %d", result.TotalCapacity())
	}
}

func TestSolve_MixedAvailabilityRespectsCapacity(t *testing.T) {
	items := []workitem.WorkItem{makeItem("WI-001"), makeItem("WI-002"), makeItem("WI-003")}
	capacities := []optimisation.ResourceCapacity{
		makeCapacity("RES-001", 5, false), // unavailable
		makeCapacity("RES-002", 2, true),  // available, capacity 2
	}
	priorities := []optimisation.WorkItemPriority{
		makePriority("WI-001", 0), makePriority("WI-002", 0), makePriority("WI-003", 0),
	}

	result, err := optimisation.Solve(items, capacities, priorities)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.Size() != 2 {
		t.Errorf("expected 2 assignments, got %d", result.Size())
	}
	if result.UnassignedCount() != 1 {
		t.Errorf("expected 1 unassigned, got %d", result.UnassignedCount())
	}

	for _, a := range result.Assignments() {
		if a.ResourceID() != "RES-002" {
			t.Errorf("expected all assignments to RES-002, got %s", a.ResourceID())
		}
	}
}

// --- Scoring and utilisation ---

func TestSolve_ScoreAllAssigned(t *testing.T) {
	items := []workitem.WorkItem{makeItem("WI-001")}
	capacities := []optimisation.ResourceCapacity{makeCapacity("RES-001", 5, true)}
	priorities := []optimisation.WorkItemPriority{makePriority("WI-001", 0)}

	result, _ := optimisation.Solve(items, capacities, priorities)
	if result.Score() != 100 {
		t.Errorf("expected score 100, got %d", result.Score())
	}
}

func TestSolve_Utilisation(t *testing.T) {
	items := []workitem.WorkItem{makeItem("WI-001"), makeItem("WI-002"), makeItem("WI-003")}
	capacities := []optimisation.ResourceCapacity{
		makeCapacity("RES-001", 2, true),
		makeCapacity("RES-002", 2, true),
	}
	priorities := []optimisation.WorkItemPriority{
		makePriority("WI-001", 0), makePriority("WI-002", 0), makePriority("WI-003", 0),
	}

	result, _ := optimisation.Solve(items, capacities, priorities)
	if result.Utilisation() != 75 {
		t.Errorf("expected utilisation 75, got %d", result.Utilisation())
	}
}

// --- Error cases ---

func TestSolve_EmptyItems(t *testing.T) {
	capacities := []optimisation.ResourceCapacity{makeCapacity("RES-001", 2, true)}
	_, err := optimisation.Solve([]workitem.WorkItem{}, capacities, nil)
	if err == nil {
		t.Fatal("expected error for empty items")
	}
}

func TestSolve_EmptyResources(t *testing.T) {
	items := []workitem.WorkItem{makeItem("WI-001")}
	_, err := optimisation.Solve(items, []optimisation.ResourceCapacity{}, nil)
	if err == nil {
		t.Fatal("expected error for empty resources")
	}
}

// --- Determinism ---

func TestSolve_Deterministic(t *testing.T) {
	items := []workitem.WorkItem{makeItem("WI-001"), makeItem("WI-002"), makeItem("WI-003")}
	capacities := []optimisation.ResourceCapacity{
		makeCapacity("RES-001", 2, true),
		makeCapacity("RES-002", 2, true),
	}
	priorities := []optimisation.WorkItemPriority{
		makePriority("WI-001", 30), makePriority("WI-002", 100), makePriority("WI-003", 50),
	}

	result1, _ := optimisation.Solve(items, capacities, priorities)
	result2, _ := optimisation.Solve(items, capacities, priorities)

	a1 := result1.Assignments()
	a2 := result2.Assignments()

	for i := range a1 {
		if a1[i].ResourceID() != a2[i].ResourceID() || a1[i].WorkItemID() != a2[i].WorkItemID() {
			t.Fatalf("optimiser is not deterministic at index %d", i)
		}
	}
}
