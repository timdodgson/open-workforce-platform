package optimisation_test

import (
	"testing"

	"github.com/timdodgson/open-workforce-platform/platform/go/internal/domain/assignment"
	"github.com/timdodgson/open-workforce-platform/platform/go/internal/optimisation"
)

func makeAssign(resourceID, workItemID string) assignment.Assignment {
	a, _ := assignment.New(resourceID, workItemID)
	return a
}

func buildResourceIndex(capacities []optimisation.ResourceInput) map[string]int {
	idx := make(map[string]int, len(capacities))
	for i, rc := range capacities {
		idx[rc.ResourceID] = i
	}
	return idx
}

// --- Placement moves ---

func TestGenerateMoves_DirectPlacement(t *testing.T) {
	capacities := []optimisation.ResourceInput{makeCapacity("RES-001", 2, true, []string{"clinical"})}
	assignments := []assignment.Assignment{}
	resourceIndex := buildResourceIndex(capacities)
	requiredSkillOf := map[string]string{"WI-001": "clinical"}
	durationOf := map[string]int{"WI-001": 1}

	moves := optimisation.GenerateMoves("WI-001", "clinical", assignments, capacities, resourceIndex, requiredSkillOf, durationOf)

	if len(moves) != 1 {
		t.Fatalf("expected 1 move, got %d", len(moves))
	}
	if moves[0].IsDisplacement() {
		t.Error("expected direct placement, got displacement")
	}
	if moves[0].TargetResource != "RES-001" {
		t.Errorf("expected target RES-001, got %s", moves[0].TargetResource)
	}
}

func TestGenerateMoves_NoMovesWhenSkillMismatch(t *testing.T) {
	capacities := []optimisation.ResourceInput{makeCapacity("RES-001", 2, true, []string{"electrical"})}
	assignments := []assignment.Assignment{}
	resourceIndex := buildResourceIndex(capacities)
	requiredSkillOf := map[string]string{"WI-001": "clinical"}
	durationOf := map[string]int{"WI-001": 1}

	moves := optimisation.GenerateMoves("WI-001", "clinical", assignments, capacities, resourceIndex, requiredSkillOf, durationOf)

	if len(moves) != 0 {
		t.Errorf("expected 0 moves (skill mismatch), got %d", len(moves))
	}
}

func TestGenerateMoves_DisplacementMove(t *testing.T) {
	capacities := []optimisation.ResourceInput{
		makeCapacity("RES-CLINICAL", 1, true, []string{"clinical"}),
		makeCapacity("RES-GENERAL", 1, true, []string{"general"}),
	}
	assignments := []assignment.Assignment{makeAssign("RES-CLINICAL", "WI-B")}
	resourceIndex := buildResourceIndex(capacities)
	requiredSkillOf := map[string]string{"WI-A": "clinical", "WI-B": ""}
	durationOf := map[string]int{"WI-A": 1, "WI-B": 1}

	moves := optimisation.GenerateMoves("WI-A", "clinical", assignments, capacities, resourceIndex, requiredSkillOf, durationOf)

	found := false
	for _, m := range moves {
		if m.IsDisplacement() && m.DisplacedItemID == "WI-B" && m.DisplacedTarget == "RES-GENERAL" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected displacement move, got %v", moves)
	}
}

func TestGenerateMoves_UnavailableResourceSkipped(t *testing.T) {
	capacities := []optimisation.ResourceInput{makeCapacity("RES-001", 5, false, nil)}
	assignments := []assignment.Assignment{}
	resourceIndex := buildResourceIndex(capacities)
	requiredSkillOf := map[string]string{"WI-001": ""}
	durationOf := map[string]int{"WI-001": 1}

	moves := optimisation.GenerateMoves("WI-001", "", assignments, capacities, resourceIndex, requiredSkillOf, durationOf)

	if len(moves) != 0 {
		t.Errorf("expected 0 moves (unavailable), got %d", len(moves))
	}
}

func TestGenerateMoves_InsufficientDuration(t *testing.T) {
	// Resource has 30 min remaining, work item needs 60.
	capacities := []optimisation.ResourceInput{makeCapacity("RES-001", 60, true, nil)}
	assignments := []assignment.Assignment{makeAssign("RES-001", "WI-EXISTING")}
	resourceIndex := buildResourceIndex(capacities)
	requiredSkillOf := map[string]string{"WI-NEW": "", "WI-EXISTING": ""}
	durationOf := map[string]int{"WI-NEW": 60, "WI-EXISTING": 30}

	moves := optimisation.GenerateMoves("WI-NEW", "", assignments, capacities, resourceIndex, requiredSkillOf, durationOf)

	// 60 capacity - 30 used = 30 remaining. Need 60. Should not fit.
	if len(moves) != 0 {
		t.Errorf("expected 0 moves (insufficient duration), got %d", len(moves))
	}
}

// --- Swap moves ---

func TestGenerateSwapMoves_ValidSwap(t *testing.T) {
	capacities := []optimisation.ResourceInput{
		makeCapacity("RES-A", 1, true, []string{"clinical", "general"}),
		makeCapacity("RES-B", 1, true, []string{"clinical", "general"}),
	}
	assignments := []assignment.Assignment{makeAssign("RES-A", "WI-001"), makeAssign("RES-B", "WI-002")}
	resourceIndex := buildResourceIndex(capacities)
	requiredSkillOf := map[string]string{"WI-001": "general", "WI-002": "general"}
	durationOf := map[string]int{"WI-001": 1, "WI-002": 1}

	moves := optimisation.GenerateSwapMoves(assignments, capacities, resourceIndex, requiredSkillOf, durationOf)

	if len(moves) != 1 {
		t.Fatalf("expected 1 swap move, got %d", len(moves))
	}
	if !moves[0].IsSwap() {
		t.Error("expected swap move")
	}
}

func TestGenerateSwapMoves_InvalidWhenSkillViolated(t *testing.T) {
	capacities := []optimisation.ResourceInput{
		makeCapacity("RES-CLINICAL", 1, true, []string{"clinical"}),
		makeCapacity("RES-ELECTRICAL", 1, true, []string{"electrical"}),
	}
	assignments := []assignment.Assignment{makeAssign("RES-CLINICAL", "WI-CLIN"), makeAssign("RES-ELECTRICAL", "WI-ELEC")}
	resourceIndex := buildResourceIndex(capacities)
	requiredSkillOf := map[string]string{"WI-CLIN": "clinical", "WI-ELEC": "electrical"}
	durationOf := map[string]int{"WI-CLIN": 1, "WI-ELEC": 1}

	moves := optimisation.GenerateSwapMoves(assignments, capacities, resourceIndex, requiredSkillOf, durationOf)

	if len(moves) != 0 {
		t.Errorf("expected 0 swap moves (skill violation), got %d", len(moves))
	}
}

func TestGenerateSwapMoves_InvalidWhenUnavailable(t *testing.T) {
	capacities := []optimisation.ResourceInput{
		makeCapacity("RES-A", 1, true, nil),
		makeCapacity("RES-B", 1, false, nil),
	}
	assignments := []assignment.Assignment{makeAssign("RES-A", "WI-001"), makeAssign("RES-B", "WI-002")}
	resourceIndex := buildResourceIndex(capacities)
	requiredSkillOf := map[string]string{"WI-001": "", "WI-002": ""}
	durationOf := map[string]int{"WI-001": 1, "WI-002": 1}

	moves := optimisation.GenerateSwapMoves(assignments, capacities, resourceIndex, requiredSkillOf, durationOf)

	if len(moves) != 0 {
		t.Errorf("expected 0 swap moves (unavailable), got %d", len(moves))
	}
}

// --- ApplyMove ---

func TestApplyMove_DirectPlacement(t *testing.T) {
	m := optimisation.CandidateMove{WorkItemID: "WI-001", TargetResource: "RES-001"}
	result, ok := optimisation.ApplyMove(m, []assignment.Assignment{})
	if !ok {
		t.Fatal("expected move to succeed")
	}
	if len(result) != 1 || result[0].WorkItemID() != "WI-001" {
		t.Errorf("unexpected result: %v", result)
	}
}

func TestApplyMove_Displacement(t *testing.T) {
	m := optimisation.CandidateMove{
		Type: optimisation.Displacement, WorkItemID: "WI-A", TargetResource: "RES-CLINICAL",
		DisplacedItemID: "WI-B", DisplacedTarget: "RES-GENERAL",
	}
	assignments := []assignment.Assignment{makeAssign("RES-CLINICAL", "WI-B")}
	result, ok := optimisation.ApplyMove(m, assignments)
	if !ok {
		t.Fatal("expected move to succeed")
	}
	if len(result) != 2 {
		t.Fatalf("expected 2, got %d", len(result))
	}
}

func TestApplyMove_Swap(t *testing.T) {
	m := optimisation.CandidateMove{
		Type: optimisation.SwapMove, WorkItemID: "WI-001", TargetResource: "RES-B",
		SwapItemID: "WI-002", SwapFrom: "RES-A",
	}
	assignments := []assignment.Assignment{makeAssign("RES-A", "WI-001"), makeAssign("RES-B", "WI-002")}
	result, ok := optimisation.ApplyMove(m, assignments)
	if !ok {
		t.Fatal("expected swap to succeed")
	}
	if result[0].ResourceID() != "RES-B" || result[1].ResourceID() != "RES-A" {
		t.Errorf("swap failed: %s/%s", result[0].ResourceID(), result[1].ResourceID())
	}
}
