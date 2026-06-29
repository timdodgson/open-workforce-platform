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

func buildResourceIndex(capacities []optimisation.ResourceCapacity) map[string]int {
	idx := make(map[string]int, len(capacities))
	for i, rc := range capacities {
		idx[rc.ResourceID] = i
	}
	return idx
}

func TestGenerateMoves_DirectPlacement(t *testing.T) {
	capacities := []optimisation.ResourceCapacity{
		makeCapacity("RES-001", 2, true, []string{"clinical"}),
	}
	assignments := []assignment.Assignment{}
	resourceIndex := buildResourceIndex(capacities)
	requiredSkillOf := map[string]string{"WI-001": "clinical"}

	moves := optimisation.GenerateMoves("WI-001", "clinical", assignments, capacities, resourceIndex, requiredSkillOf)

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
	capacities := []optimisation.ResourceCapacity{
		makeCapacity("RES-001", 2, true, []string{"electrical"}),
	}
	assignments := []assignment.Assignment{}
	resourceIndex := buildResourceIndex(capacities)
	requiredSkillOf := map[string]string{"WI-001": "clinical"}

	moves := optimisation.GenerateMoves("WI-001", "clinical", assignments, capacities, resourceIndex, requiredSkillOf)

	if len(moves) != 0 {
		t.Errorf("expected 0 moves (skill mismatch), got %d", len(moves))
	}
}

func TestGenerateMoves_DisplacementMove(t *testing.T) {
	capacities := []optimisation.ResourceCapacity{
		makeCapacity("RES-CLINICAL", 1, true, []string{"clinical"}),
		makeCapacity("RES-GENERAL", 1, true, []string{"general"}),
	}
	assignments := []assignment.Assignment{makeAssign("RES-CLINICAL", "WI-B")}
	resourceIndex := buildResourceIndex(capacities)
	requiredSkillOf := map[string]string{
		"WI-A": "clinical",
		"WI-B": "",
	}

	moves := optimisation.GenerateMoves("WI-A", "clinical", assignments, capacities, resourceIndex, requiredSkillOf)

	found := false
	for _, m := range moves {
		if m.IsDisplacement() && m.DisplacedItemID == "WI-B" && m.DisplacedTarget == "RES-GENERAL" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected displacement move (WI-B to RES-GENERAL), got %v", moves)
	}
}

func TestGenerateMoves_UnavailableResourceSkipped(t *testing.T) {
	capacities := []optimisation.ResourceCapacity{
		makeCapacity("RES-001", 5, false, nil),
	}
	assignments := []assignment.Assignment{}
	resourceIndex := buildResourceIndex(capacities)
	requiredSkillOf := map[string]string{"WI-001": ""}

	moves := optimisation.GenerateMoves("WI-001", "", assignments, capacities, resourceIndex, requiredSkillOf)

	if len(moves) != 0 {
		t.Errorf("expected 0 moves (unavailable), got %d", len(moves))
	}
}

func TestApplyMove_DirectPlacement(t *testing.T) {
	m := optimisation.CandidateMove{
		WorkItemID:     "WI-001",
		TargetResource: "RES-001",
	}
	assignments := []assignment.Assignment{}

	result, ok := optimisation.ApplyMove(m, assignments)
	if !ok {
		t.Fatal("expected move to succeed")
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 assignment, got %d", len(result))
	}
	if result[0].WorkItemID() != "WI-001" || result[0].ResourceID() != "RES-001" {
		t.Errorf("unexpected assignment: %s → %s", result[0].ResourceID(), result[0].WorkItemID())
	}
}

func TestApplyMove_Displacement(t *testing.T) {
	m := optimisation.CandidateMove{
		WorkItemID:      "WI-A",
		TargetResource:  "RES-CLINICAL",
		DisplacedItemID: "WI-B",
		DisplacedTarget: "RES-GENERAL",
	}
	assignments := []assignment.Assignment{makeAssign("RES-CLINICAL", "WI-B")}

	result, ok := optimisation.ApplyMove(m, assignments)
	if !ok {
		t.Fatal("expected move to succeed")
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 assignments, got %d", len(result))
	}
	if result[0].WorkItemID() != "WI-B" || result[0].ResourceID() != "RES-GENERAL" {
		t.Errorf("expected WI-B on RES-GENERAL, got %s on %s", result[0].WorkItemID(), result[0].ResourceID())
	}
	if result[1].WorkItemID() != "WI-A" || result[1].ResourceID() != "RES-CLINICAL" {
		t.Errorf("expected WI-A on RES-CLINICAL, got %s on %s", result[1].WorkItemID(), result[1].ResourceID())
	}
}
