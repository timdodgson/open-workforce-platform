package optimisation_test

import (
	"testing"

	"github.com/timdodgson/open-workforce-platform/platform/go/internal/domain/workitem"
	"github.com/timdodgson/open-workforce-platform/platform/go/internal/optimisation"
)

// --- Hill climbing starts from constructive solution ---

func TestHillClimbing_StartsFromConstructive(t *testing.T) {
	// Simple case: all items fit. Hill climbing should produce same result as constructive.
	items := []workitem.WorkItem{makeItem("WI-001"), makeItem("WI-002")}
	capacities := []optimisation.ResourceCapacity{makeCapacity("RES-001", 3, true, nil)}
	priorities := []optimisation.WorkItemPriority{makePriority("WI-001", 0, ""), makePriority("WI-002", 0, "")}

	constructive, _ := optimisation.Solve(items, capacities, priorities)
	hillClimb, err := optimisation.SolveHillClimbing(items, capacities, priorities)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if hillClimb.Size() != constructive.Size() {
		t.Errorf("expected same size %d, got %d", constructive.Size(), hillClimb.Size())
	}
	if hillClimb.Score() != constructive.Score() {
		t.Errorf("expected same score %d, got %d", constructive.Score(), hillClimb.Score())
	}
}

// --- Hill climbing improves when constructive is suboptimal ---

func TestHillClimbing_ImprovesOverConstructive(t *testing.T) {
	// Setup where constructive makes a suboptimal choice:
	// WI-A (priority 100, requires "clinical")
	// WI-B (priority 50, no required skill)
	// RES-CLINICAL: capacity 1, skills ["clinical"]
	// RES-GENERAL: capacity 1, skills ["general"]
	//
	// Constructive (by priority): assigns WI-A to RES-CLINICAL, then WI-B needs
	// a slot — RES-CLINICAL is full, RES-GENERAL has capacity but WI-B has no
	// skill requirement so it fits. Score = 100.
	//
	// Actually this works fine constructively. Let me create a scenario where
	// constructive fails:
	//
	// WI-A (priority 100, requires "clinical")
	// WI-B (priority 200, no required skill)
	// RES-CLINICAL: capacity 1, skills ["clinical"]
	//
	// Constructive: WI-B (higher priority, no skill req) gets RES-CLINICAL first.
	// WI-A (requires clinical) — RES-CLINICAL is full → unassigned.
	// Score = 50.
	//
	// Hill climbing: move WI-B off RES-CLINICAL... but where? No other resource.
	// Need a second resource.
	//
	// Add RES-GENERAL: capacity 1, skills ["general"]
	//
	// Constructive: WI-B (priority 200) → RES-CLINICAL (first available, no skill req).
	// WI-A (priority 100, requires clinical) → RES-CLINICAL full, RES-GENERAL doesn't have clinical → unassigned.
	// Score = 50.
	//
	// Hill climbing: move WI-B from RES-CLINICAL to RES-GENERAL (WI-B has no skill req).
	// Now RES-CLINICAL has a slot. Place WI-A on RES-CLINICAL. Score = 100.

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

	hillClimb, err := optimisation.SolveHillClimbing(items, capacities, priorities)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if hillClimb.Score() != 100 {
		t.Errorf("expected hill climbing score 100, got %d", hillClimb.Score())
	}
	if hillClimb.Size() != 2 {
		t.Errorf("expected 2 assignments, got %d", hillClimb.Size())
	}
}

// --- Invalid moves are rejected ---

func TestHillClimbing_RejectsInvalidMoves(t *testing.T) {
	// WI-A requires "clinical", only RES-CLINICAL has it but is full with WI-B.
	// WI-B requires "clinical" too — cannot be moved to RES-GENERAL.
	// Hill climbing should not improve.

	items := []workitem.WorkItem{makeItem("WI-A"), makeItem("WI-B")}
	capacities := []optimisation.ResourceCapacity{
		makeCapacity("RES-CLINICAL", 1, true, []string{"clinical"}),
		makeCapacity("RES-GENERAL", 1, true, []string{"general"}),
	}
	priorities := []optimisation.WorkItemPriority{
		makePriority("WI-A", 100, "clinical"),
		makePriority("WI-B", 200, "clinical"), // both need clinical
	}

	constructive, _ := optimisation.Solve(items, capacities, priorities)
	hillClimb, _ := optimisation.SolveHillClimbing(items, capacities, priorities)

	// Cannot improve — WI-B can't be moved (needs clinical).
	if hillClimb.Score() != constructive.Score() {
		t.Errorf("expected no improvement, constructive=%d, hillClimb=%d", constructive.Score(), hillClimb.Score())
	}
}

// --- Equal-score moves are rejected ---

func TestHillClimbing_RejectsEqualScoreMoves(t *testing.T) {
	// All items assigned in constructive. Score is 100. No improvement possible.
	items := []workitem.WorkItem{makeItem("WI-001"), makeItem("WI-002")}
	capacities := []optimisation.ResourceCapacity{
		makeCapacity("RES-001", 2, true, nil),
		makeCapacity("RES-002", 2, true, nil),
	}
	priorities := []optimisation.WorkItemPriority{
		makePriority("WI-001", 50, ""), makePriority("WI-002", 50, ""),
	}

	hillClimb, _ := optimisation.SolveHillClimbing(items, capacities, priorities)
	if hillClimb.Score() != 100 {
		t.Errorf("expected score 100, got %d", hillClimb.Score())
	}
}

// --- Deterministic ---

func TestHillClimbing_Deterministic(t *testing.T) {
	items := []workitem.WorkItem{makeItem("WI-A"), makeItem("WI-B")}
	capacities := []optimisation.ResourceCapacity{
		makeCapacity("RES-CLINICAL", 1, true, []string{"clinical"}),
		makeCapacity("RES-GENERAL", 1, true, []string{"general"}),
	}
	priorities := []optimisation.WorkItemPriority{
		makePriority("WI-A", 100, "clinical"),
		makePriority("WI-B", 200, ""),
	}

	result1, _ := optimisation.SolveHillClimbing(items, capacities, priorities)
	result2, _ := optimisation.SolveHillClimbing(items, capacities, priorities)

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

// --- Respects constraints ---

func TestHillClimbing_RespectsAvailability(t *testing.T) {
	items := []workitem.WorkItem{makeItem("WI-001")}
	capacities := []optimisation.ResourceCapacity{
		makeCapacity("RES-001", 5, false, nil), // unavailable
	}
	priorities := []optimisation.WorkItemPriority{makePriority("WI-001", 0, "")}

	result, _ := optimisation.SolveHillClimbing(items, capacities, priorities)
	if result.Size() != 0 {
		t.Errorf("expected 0 assignments (unavailable), got %d", result.Size())
	}
}

func TestHillClimbing_RespectsCapacity(t *testing.T) {
	items := []workitem.WorkItem{makeItem("WI-001"), makeItem("WI-002"), makeItem("WI-003")}
	capacities := []optimisation.ResourceCapacity{makeCapacity("RES-001", 2, true, nil)}
	priorities := []optimisation.WorkItemPriority{
		makePriority("WI-001", 0, ""), makePriority("WI-002", 0, ""), makePriority("WI-003", 0, ""),
	}

	result, _ := optimisation.SolveHillClimbing(items, capacities, priorities)
	if result.Size() != 2 {
		t.Errorf("expected 2 assignments (capacity 2), got %d", result.Size())
	}
}

func TestHillClimbing_RespectsSkills(t *testing.T) {
	items := []workitem.WorkItem{makeItem("WI-001")}
	capacities := []optimisation.ResourceCapacity{
		makeCapacity("RES-001", 5, true, []string{"electrical"}),
	}
	priorities := []optimisation.WorkItemPriority{makePriority("WI-001", 0, "clinical")}

	result, _ := optimisation.SolveHillClimbing(items, capacities, priorities)
	if result.Size() != 0 {
		t.Errorf("expected 0 assignments (skill mismatch), got %d", result.Size())
	}
}

// --- Error cases ---

func TestHillClimbing_EmptyItems(t *testing.T) {
	capacities := []optimisation.ResourceCapacity{makeCapacity("RES-001", 2, true, nil)}
	_, err := optimisation.SolveHillClimbing(nil, capacities, nil)
	if err == nil {
		t.Fatal("expected error for empty items")
	}
}

func TestHillClimbing_EmptyResources(t *testing.T) {
	items := []workitem.WorkItem{makeItem("WI-001")}
	_, err := optimisation.SolveHillClimbing(items, nil, nil)
	if err == nil {
		t.Fatal("expected error for empty resources")
	}
}
