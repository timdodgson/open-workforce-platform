package optimisation_test

import (
	"testing"

	"github.com/timdodgson/open-workforce-platform/platform/go/internal/domain/assignment"
	"github.com/timdodgson/open-workforce-platform/platform/go/internal/optimisation"
)

func makeTestAssignment(resourceID, workItemID string) assignment.Assignment {
	a, _ := assignment.New(resourceID, workItemID)
	return a
}

func TestObjectiveScore_MoreAssignmentsScoresHigher(t *testing.T) {
	capacities := []optimisation.ResourceCapacity{
		makeCapacity("RES-001", 5, true, nil),
	}

	one := []assignment.Assignment{makeTestAssignment("RES-001", "WI-001")}
	two := []assignment.Assignment{makeTestAssignment("RES-001", "WI-001"), makeTestAssignment("RES-001", "WI-002")}

	scoreOne := optimisation.ObjectiveScore(one, capacities)
	scoreTwo := optimisation.ObjectiveScore(two, capacities)

	if scoreTwo <= scoreOne {
		t.Errorf("expected more assignments to score higher: %d vs %d", scoreTwo, scoreOne)
	}
}

func TestObjectiveScore_AssignmentDominatesBalance(t *testing.T) {
	capacities := []optimisation.ResourceCapacity{
		makeCapacity("RES-001", 5, true, nil),
		makeCapacity("RES-002", 5, true, nil),
	}

	// 3 items, imbalanced (3+0): higher assignment count
	threeImbalanced := []assignment.Assignment{
		makeTestAssignment("RES-001", "WI-001"),
		makeTestAssignment("RES-001", "WI-002"),
		makeTestAssignment("RES-001", "WI-003"),
	}

	// 2 items, balanced (1+1): better balance but fewer assignments
	twoBalanced := []assignment.Assignment{
		makeTestAssignment("RES-001", "WI-001"),
		makeTestAssignment("RES-002", "WI-002"),
	}

	scoreThree := optimisation.ObjectiveScore(threeImbalanced, capacities)
	scoreTwo := optimisation.ObjectiveScore(twoBalanced, capacities)

	if scoreTwo >= scoreThree {
		t.Errorf("assignment should dominate balance: 3 imbalanced=%d, 2 balanced=%d", scoreThree, scoreTwo)
	}
}

func TestObjectiveScore_BalancedBetterThanImbalanced(t *testing.T) {
	capacities := []optimisation.ResourceCapacity{
		makeCapacity("RES-001", 5, true, nil),
		makeCapacity("RES-002", 5, true, nil),
	}

	// Same number of assignments, different distribution.
	balanced := []assignment.Assignment{
		makeTestAssignment("RES-001", "WI-001"),
		makeTestAssignment("RES-002", "WI-002"),
	}
	imbalanced := []assignment.Assignment{
		makeTestAssignment("RES-001", "WI-001"),
		makeTestAssignment("RES-001", "WI-002"),
	}

	scoreBalanced := optimisation.ObjectiveScore(balanced, capacities)
	scoreImbalanced := optimisation.ObjectiveScore(imbalanced, capacities)

	if scoreBalanced <= scoreImbalanced {
		t.Errorf("balanced should score higher: balanced=%d, imbalanced=%d", scoreBalanced, scoreImbalanced)
	}
}

func TestObjectiveScore_Deterministic(t *testing.T) {
	capacities := []optimisation.ResourceCapacity{
		makeCapacity("RES-001", 3, true, nil),
		makeCapacity("RES-002", 3, true, nil),
	}
	assignments := []assignment.Assignment{
		makeTestAssignment("RES-001", "WI-001"),
		makeTestAssignment("RES-002", "WI-002"),
	}

	score1 := optimisation.ObjectiveScore(assignments, capacities)
	score2 := optimisation.ObjectiveScore(assignments, capacities)

	if score1 != score2 {
		t.Errorf("scoring should be deterministic: %d vs %d", score1, score2)
	}
}

func TestObjectiveScore_EmptyAssignments(t *testing.T) {
	capacities := []optimisation.ResourceCapacity{
		makeCapacity("RES-001", 3, true, nil),
	}

	score := optimisation.ObjectiveScore([]assignment.Assignment{}, capacities)
	if score != 0 {
		t.Errorf("expected 0 for empty assignments, got %d", score)
	}
}

func TestObjectiveBreakdown_SumEqualsTotal(t *testing.T) {
	capacities := []optimisation.ResourceCapacity{
		makeCapacity("RES-001", 3, true, nil),
		makeCapacity("RES-002", 3, true, nil),
	}
	assignments := []assignment.Assignment{
		makeTestAssignment("RES-001", "WI-001"),
		makeTestAssignment("RES-002", "WI-002"),
		makeTestAssignment("RES-001", "WI-003"),
	}

	total := optimisation.ObjectiveScore(assignments, capacities)
	breakdown := optimisation.ObjectiveBreakdown(assignments, capacities)

	sum := 0
	for _, entry := range breakdown {
		sum += entry.Score
	}

	if sum != total {
		t.Errorf("breakdown sum %d does not equal total %d", sum, total)
	}
}

func TestObjectiveBreakdown_ContainsExpectedObjectives(t *testing.T) {
	capacities := []optimisation.ResourceCapacity{
		makeCapacity("RES-001", 3, true, nil),
	}
	assignments := []assignment.Assignment{
		makeTestAssignment("RES-001", "WI-001"),
	}

	breakdown := optimisation.ObjectiveBreakdown(assignments, capacities)

	if len(breakdown) != 2 {
		t.Fatalf("expected 2 objectives, got %d", len(breakdown))
	}
	if breakdown[0].Name != "Assignment" {
		t.Errorf("expected first objective 'Assignment', got %q", breakdown[0].Name)
	}
	if breakdown[1].Name != "Workload Balance" {
		t.Errorf("expected second objective 'Workload Balance', got %q", breakdown[1].Name)
	}
}

func TestObjectiveBreakdown_AssignmentContribution(t *testing.T) {
	capacities := []optimisation.ResourceCapacity{
		makeCapacity("RES-001", 5, true, nil),
	}
	assignments := []assignment.Assignment{
		makeTestAssignment("RES-001", "WI-001"),
		makeTestAssignment("RES-001", "WI-002"),
	}

	breakdown := optimisation.ObjectiveBreakdown(assignments, capacities)

	// 2 items × 1000 = 2000
	if breakdown[0].Score != 2000 {
		t.Errorf("expected assignment contribution 2000, got %d", breakdown[0].Score)
	}
}
