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

func makePriority(id string, priority int) optimisation.WorkItemPriority {
	return optimisation.WorkItemPriority{WorkItemID: id, Priority: priority}
}

// --- Capacity behaviour (unchanged) ---

func TestSolve_AssignsWithinCapacity(t *testing.T) {
	items := []workitem.WorkItem{makeItem("WI-001"), makeItem("WI-002")}
	capacities := []optimisation.ResourceCapacity{makeCapacity("RES-001", 3)}
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
	capacities := []optimisation.ResourceCapacity{makeCapacity("RES-001", 2)}
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
	capacities := []optimisation.ResourceCapacity{makeCapacity("RES-001", 2)}
	priorities := []optimisation.WorkItemPriority{
		makePriority("WI-001", 0),
		makePriority("WI-002", 0),
		makePriority("WI-003", 0),
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
		makeCapacity("RES-001", 2),
		makeCapacity("RES-002", 2),
	}
	priorities := []optimisation.WorkItemPriority{
		makePriority("WI-001", 0),
		makePriority("WI-002", 0),
		makePriority("WI-003", 0),
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
	capacities := []optimisation.ResourceCapacity{makeCapacity("RES-001", 0)}
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
	capacities := []optimisation.ResourceCapacity{makeCapacity("RES-001", 1)}
	priorities := []optimisation.WorkItemPriority{
		makePriority("WI-LOW", 10),
		makePriority("WI-HIGH", 100),
	}

	result, err := optimisation.Solve(items, capacities, priorities)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.Size() != 1 {
		t.Fatalf("expected 1 assignment, got %d", result.Size())
	}

	assigned := result.Assignments()[0]
	if assigned.WorkItemID() != "WI-HIGH" {
		t.Errorf("expected WI-HIGH assigned (higher priority), got %s", assigned.WorkItemID())
	}

	unassigned := result.Unassigned()
	if unassigned[0] != "WI-LOW" {
		t.Errorf("expected WI-LOW unassigned, got %s", unassigned[0])
	}
}

func TestSolve_LowPriorityUnassignedWhenCapacityLimited(t *testing.T) {
	items := []workitem.WorkItem{makeItem("WI-A"), makeItem("WI-B"), makeItem("WI-C")}
	capacities := []optimisation.ResourceCapacity{makeCapacity("RES-001", 2)}
	priorities := []optimisation.WorkItemPriority{
		makePriority("WI-A", 50),
		makePriority("WI-B", 100),
		makePriority("WI-C", 10),
	}

	result, err := optimisation.Solve(items, capacities, priorities)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// WI-B (100) and WI-A (50) should be assigned; WI-C (10) unassigned.
	assignments := result.Assignments()
	if assignments[0].WorkItemID() != "WI-B" {
		t.Errorf("expected first assignment WI-B, got %s", assignments[0].WorkItemID())
	}
	if assignments[1].WorkItemID() != "WI-A" {
		t.Errorf("expected second assignment WI-A, got %s", assignments[1].WorkItemID())
	}

	unassigned := result.Unassigned()
	if unassigned[0] != "WI-C" {
		t.Errorf("expected WI-C unassigned, got %s", unassigned[0])
	}
}

func TestSolve_MissingPriorityDefaultsToZero(t *testing.T) {
	items := []workitem.WorkItem{makeItem("WI-NO-PRIO"), makeItem("WI-HAS-PRIO")}
	capacities := []optimisation.ResourceCapacity{makeCapacity("RES-001", 1)}
	priorities := []optimisation.WorkItemPriority{
		// WI-NO-PRIO is not in the priorities list — absent means 0.
		makePriority("WI-HAS-PRIO", 50),
	}

	result, err := optimisation.Solve(items, capacities, priorities)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	assigned := result.Assignments()[0]
	if assigned.WorkItemID() != "WI-HAS-PRIO" {
		t.Errorf("expected WI-HAS-PRIO assigned (priority 50 > 0), got %s", assigned.WorkItemID())
	}
}

func TestSolve_EqualPriorityPreservesOriginalOrder(t *testing.T) {
	items := []workitem.WorkItem{makeItem("WI-FIRST"), makeItem("WI-SECOND"), makeItem("WI-THIRD")}
	capacities := []optimisation.ResourceCapacity{makeCapacity("RES-001", 2)}
	priorities := []optimisation.WorkItemPriority{
		makePriority("WI-FIRST", 50),
		makePriority("WI-SECOND", 50),
		makePriority("WI-THIRD", 50),
	}

	result, err := optimisation.Solve(items, capacities, priorities)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Same priority: original order preserved, so WI-FIRST and WI-SECOND assigned.
	assignments := result.Assignments()
	if assignments[0].WorkItemID() != "WI-FIRST" {
		t.Errorf("expected first assignment WI-FIRST, got %s", assignments[0].WorkItemID())
	}
	if assignments[1].WorkItemID() != "WI-SECOND" {
		t.Errorf("expected second assignment WI-SECOND, got %s", assignments[1].WorkItemID())
	}

	unassigned := result.Unassigned()
	if unassigned[0] != "WI-THIRD" {
		t.Errorf("expected WI-THIRD unassigned, got %s", unassigned[0])
	}
}

// --- Scoring and utilisation ---

func TestSolve_ScoreAllAssigned(t *testing.T) {
	items := []workitem.WorkItem{makeItem("WI-001")}
	capacities := []optimisation.ResourceCapacity{makeCapacity("RES-001", 5)}
	priorities := []optimisation.WorkItemPriority{makePriority("WI-001", 0)}

	result, _ := optimisation.Solve(items, capacities, priorities)
	if result.Score() != 100 {
		t.Errorf("expected score 100, got %d", result.Score())
	}
}

func TestSolve_ScorePartial(t *testing.T) {
	items := []workitem.WorkItem{makeItem("WI-001"), makeItem("WI-002"), makeItem("WI-003")}
	capacities := []optimisation.ResourceCapacity{makeCapacity("RES-001", 2)}
	priorities := []optimisation.WorkItemPriority{
		makePriority("WI-001", 0),
		makePriority("WI-002", 0),
		makePriority("WI-003", 0),
	}

	result, _ := optimisation.Solve(items, capacities, priorities)
	if result.Score() != 67 {
		t.Errorf("expected score 67, got %d", result.Score())
	}
}

func TestSolve_Utilisation(t *testing.T) {
	items := []workitem.WorkItem{makeItem("WI-001"), makeItem("WI-002"), makeItem("WI-003")}
	capacities := []optimisation.ResourceCapacity{
		makeCapacity("RES-001", 2),
		makeCapacity("RES-002", 2),
	}
	priorities := []optimisation.WorkItemPriority{
		makePriority("WI-001", 0),
		makePriority("WI-002", 0),
		makePriority("WI-003", 0),
	}

	result, _ := optimisation.Solve(items, capacities, priorities)
	if result.Utilisation() != 75 {
		t.Errorf("expected utilisation 75, got %d", result.Utilisation())
	}
}

// --- Error cases ---

func TestSolve_EmptyItems(t *testing.T) {
	capacities := []optimisation.ResourceCapacity{makeCapacity("RES-001", 2)}
	_, err := optimisation.Solve([]workitem.WorkItem{}, capacities, nil)
	if err == nil {
		t.Fatal("expected error for empty items")
	}
}

func TestSolve_EmptyResources(t *testing.T) {
	items := []workitem.WorkItem{makeItem("WI-001")}
	priorities := []optimisation.WorkItemPriority{makePriority("WI-001", 0)}
	_, err := optimisation.Solve(items, []optimisation.ResourceCapacity{}, priorities)
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
	priorities := []optimisation.WorkItemPriority{
		makePriority("WI-001", 30),
		makePriority("WI-002", 100),
		makePriority("WI-003", 50),
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
