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

func makeCapacity(id string, cap int, available bool, skills []string) optimisation.ResourceCapacity {
	return optimisation.ResourceCapacity{ResourceID: id, Capacity: cap, Available: available, Skills: skills}
}

func makePriority(id string, priority int, requiredSkill string) optimisation.WorkItemPriority {
	return optimisation.WorkItemPriority{WorkItemID: id, Priority: priority, RequiredSkill: requiredSkill}
}

// --- Capacity behaviour ---

func TestSolve_AssignsWithinCapacity(t *testing.T) {
	items := []workitem.WorkItem{makeItem("WI-001"), makeItem("WI-002")}
	capacities := []optimisation.ResourceCapacity{makeCapacity("RES-001", 3, true, nil)}
	priorities := []optimisation.WorkItemPriority{makePriority("WI-001", 0, ""), makePriority("WI-002", 0, "")}

	result, err := optimisation.Solve(items, capacities, priorities)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.Size() != 2 {
		t.Errorf("expected 2 assignments, got %d", result.Size())
	}
}

func TestSolve_InsufficientCapacity(t *testing.T) {
	items := []workitem.WorkItem{makeItem("WI-001"), makeItem("WI-002"), makeItem("WI-003")}
	capacities := []optimisation.ResourceCapacity{makeCapacity("RES-001", 2, true, nil)}
	priorities := []optimisation.WorkItemPriority{
		makePriority("WI-001", 0, ""), makePriority("WI-002", 0, ""), makePriority("WI-003", 0, ""),
	}

	result, _ := optimisation.Solve(items, capacities, priorities)
	if result.Size() != 2 {
		t.Errorf("expected 2 assignments, got %d", result.Size())
	}
	if result.UnassignedCount() != 1 {
		t.Errorf("expected 1 unassigned, got %d", result.UnassignedCount())
	}
}

func TestSolve_UnavailableResourceSkipped(t *testing.T) {
	items := []workitem.WorkItem{makeItem("WI-001")}
	capacities := []optimisation.ResourceCapacity{
		makeCapacity("RES-UNAVAIL", 5, false, nil),
		makeCapacity("RES-AVAIL", 2, true, nil),
	}
	priorities := []optimisation.WorkItemPriority{makePriority("WI-001", 0, "")}

	result, _ := optimisation.Solve(items, capacities, priorities)
	if result.Assignments()[0].ResourceID() != "RES-AVAIL" {
		t.Errorf("expected assignment to RES-AVAIL, got %s", result.Assignments()[0].ResourceID())
	}
}

func TestSolve_HigherPriorityAssignedFirst(t *testing.T) {
	items := []workitem.WorkItem{makeItem("WI-LOW"), makeItem("WI-HIGH")}
	capacities := []optimisation.ResourceCapacity{makeCapacity("RES-001", 1, true, nil)}
	priorities := []optimisation.WorkItemPriority{
		makePriority("WI-LOW", 10, ""), makePriority("WI-HIGH", 100, ""),
	}

	result, _ := optimisation.Solve(items, capacities, priorities)
	if result.Assignments()[0].WorkItemID() != "WI-HIGH" {
		t.Errorf("expected WI-HIGH assigned, got %s", result.Assignments()[0].WorkItemID())
	}
}

func TestSolve_EqualPriorityPreservesOrder(t *testing.T) {
	items := []workitem.WorkItem{makeItem("WI-FIRST"), makeItem("WI-SECOND")}
	capacities := []optimisation.ResourceCapacity{makeCapacity("RES-001", 1, true, nil)}
	priorities := []optimisation.WorkItemPriority{
		makePriority("WI-FIRST", 50, ""), makePriority("WI-SECOND", 50, ""),
	}

	result, _ := optimisation.Solve(items, capacities, priorities)
	if result.Assignments()[0].WorkItemID() != "WI-FIRST" {
		t.Errorf("expected WI-FIRST (original order), got %s", result.Assignments()[0].WorkItemID())
	}
}

func TestSolve_AssignedWhenResourceHasRequiredSkill(t *testing.T) {
	items := []workitem.WorkItem{makeItem("WI-001")}
	capacities := []optimisation.ResourceCapacity{
		makeCapacity("RES-001", 2, true, []string{"clinical", "assessment"}),
	}
	priorities := []optimisation.WorkItemPriority{makePriority("WI-001", 0, "clinical")}

	result, err := optimisation.Solve(items, capacities, priorities)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.Size() != 1 {
		t.Errorf("expected 1 assignment, got %d", result.Size())
	}
}

func TestSolve_UnassignedWhenNoResourceHasRequiredSkill(t *testing.T) {
	items := []workitem.WorkItem{makeItem("WI-001")}
	capacities := []optimisation.ResourceCapacity{
		makeCapacity("RES-001", 5, true, []string{"electrical", "plumbing"}),
	}
	priorities := []optimisation.WorkItemPriority{makePriority("WI-001", 0, "clinical")}

	result, _ := optimisation.Solve(items, capacities, priorities)
	if result.Size() != 0 {
		t.Errorf("expected 0 assignments, got %d", result.Size())
	}
}

func TestSolve_NoRequiredSkillAssignedToAnyResource(t *testing.T) {
	items := []workitem.WorkItem{makeItem("WI-001")}
	capacities := []optimisation.ResourceCapacity{
		makeCapacity("RES-001", 2, true, []string{"clinical"}),
	}
	priorities := []optimisation.WorkItemPriority{makePriority("WI-001", 0, "")}

	result, _ := optimisation.Solve(items, capacities, priorities)
	if result.Size() != 1 {
		t.Errorf("expected 1 assignment, got %d", result.Size())
	}
}

func TestSolve_ResourceWithNoSkillsCannotSatisfyRequirement(t *testing.T) {
	items := []workitem.WorkItem{makeItem("WI-001")}
	capacities := []optimisation.ResourceCapacity{
		makeCapacity("RES-NO-SKILLS", 5, true, nil),
	}
	priorities := []optimisation.WorkItemPriority{makePriority("WI-001", 0, "clinical")}

	result, _ := optimisation.Solve(items, capacities, priorities)
	if result.Size() != 0 {
		t.Errorf("expected 0 assignments, got %d", result.Size())
	}
}

func TestSolve_SkillMatchIsCaseSensitive(t *testing.T) {
	items := []workitem.WorkItem{makeItem("WI-001")}
	capacities := []optimisation.ResourceCapacity{
		makeCapacity("RES-001", 5, true, []string{"Clinical"}),
	}
	priorities := []optimisation.WorkItemPriority{makePriority("WI-001", 0, "clinical")}

	result, _ := optimisation.Solve(items, capacities, priorities)
	if result.Size() != 0 {
		t.Errorf("expected 0 assignments (case-sensitive mismatch), got %d", result.Size())
	}
}

func TestSolve_EmptyItems(t *testing.T) {
	capacities := []optimisation.ResourceCapacity{makeCapacity("RES-001", 2, true, nil)}
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

func TestSolve_Deterministic(t *testing.T) {
	items := []workitem.WorkItem{makeItem("WI-001"), makeItem("WI-002"), makeItem("WI-003")}
	capacities := []optimisation.ResourceCapacity{
		makeCapacity("RES-001", 2, true, []string{"clinical", "general"}),
		makeCapacity("RES-002", 2, true, []string{"general"}),
	}
	priorities := []optimisation.WorkItemPriority{
		makePriority("WI-001", 30, "general"),
		makePriority("WI-002", 100, "clinical"),
		makePriority("WI-003", 50, "general"),
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

// --- Algorithm registry ---

func TestGet_Constructive(t *testing.T) {
	alg, err := optimisation.Get("constructive")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if alg.Name() != "constructive" {
		t.Errorf("expected name constructive, got %s", alg.Name())
	}
}

func TestGet_HillClimbing(t *testing.T) {
	alg, err := optimisation.Get("hill-climbing")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if alg.Name() != "hill-climbing" {
		t.Errorf("expected name hill-climbing, got %s", alg.Name())
	}
}

func TestGet_UnknownAlgorithm(t *testing.T) {
	_, err := optimisation.Get("unknown-algo")
	if err == nil {
		t.Fatal("expected error for unknown algorithm")
	}
}
